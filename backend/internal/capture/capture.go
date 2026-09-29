// Package capture implements a server-side LED sweep controller (REQ-047).
// When started for a device it turns one LED on at a time in ascending index
// order (0 … n-1), each for a configurable dwell period (default 1 s), then
// turns all LEDs off on completion or stop.  It operates directly by LED index
// and does not go through LightStateStore, so it works even when the device is
// not assigned to a model.
package capture

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"example.com/dlm/backend/internal/store"
)

// Sentinel errors returned by Controller.
var (
	// ErrCaptureNoLights is returned when the device has light_count == 0.
	ErrCaptureNoLights = errors.New("device has no lights configured")
	// ErrCaptureConflict is returned when a sweep is already running for the
	// device, or when an active routine run exists for its assigned model.
	ErrCaptureConflict = errors.New("capture sweep already running for this device")
	// ErrCaptureRoutineCheck is returned when the active-routine conflict guard
	// cannot be evaluated (e.g. store query failure).  Start must fail closed.
	ErrCaptureRoutineCheck = errors.New("could not check for active routine run")
	// ErrCaptureDwellTooShort is returned when dwell is shorter than the bookend allows.
	ErrCaptureDwellTooShort = errors.New("capture dwell is shorter than the bookend signal allows")
)

const (
	defaultCueOn      = 200 * time.Millisecond
	defaultCueGap     = 200 * time.Millisecond
	defaultSettle     = 500 * time.Millisecond
	minDefaultDwell   = 500 * time.Millisecond
)

// driver drives raw LED frames on a WLED device.
type driver interface {
	DriveSingleLED(ctx context.Context, d store.Device, litIdx, n int) error
	DriveAllOff(ctx context.Context, d store.Device, n int) error
	DriveAllColor(ctx context.Context, d store.Device, n, r, g, b int) error
}

// deviceGetter retrieves a device from persistent storage.
type deviceGetter interface {
	GetDevice(ctx context.Context, id string) (store.Device, error)
}

// RoutineChecker checks whether a model has an active routine run.
// Pass nil to skip the routine-conflict guard.
type RoutineChecker interface {
	ModelHasActiveRoutineRun(ctx context.Context, modelID string) (bool, error)
}

// Status is the observable state of a capture sweep for one device.
type Status struct {
	State        string `json:"state"`
	LightCount   int    `json:"light_count"`
	CurrentIndex int    `json:"current_index"`
	Phase        string `json:"phase"`
}

const (
	stateIdle     = "idle"
	stateRunning  = "running"
	stateStopping = "stopping"
)

// sweepEntry holds per-device sweep state and its cancellation channel.
type sweepEntry struct {
	mu           sync.Mutex
	state        string
	lightCount   int
	currentIndex int
	phase        string
	stopOnce     sync.Once
	stop         chan struct{}
}

func (e *sweepEntry) doStop() {
	e.stopOnce.Do(func() { close(e.stop) })
}

// ControllerOpts configures optional Controller parameters.
type ControllerOpts struct {
	Dwell  time.Duration
	CueOn  time.Duration
	CueGap time.Duration
	Settle time.Duration
}

// Controller manages at most one active capture sweep per device.
// It is safe for concurrent use.
type Controller struct {
	getter  deviceGetter
	drv     driver
	checker RoutineChecker // may be nil
	dwell   time.Duration
	cueOn   time.Duration
	cueGap  time.Duration
	settle  time.Duration

	mu     sync.Mutex
	sweeps map[string]*sweepEntry
}

// New creates a Controller.  checker may be nil to skip routine-conflict
// checking.  The dwell period is resolved in priority order: opts.Dwell →
// DLM_CAPTURE_DWELL_MS env var → default 1000 ms.
func New(getter deviceGetter, drv driver, checker RoutineChecker, opts *ControllerOpts) *Controller {
	dwell := 1000 * time.Millisecond
	cueOn := defaultCueOn
	cueGap := defaultCueGap
	settle := defaultSettle

	if opts != nil {
		if opts.Dwell > 0 {
			dwell = opts.Dwell
		}
		if opts.CueOn > 0 {
			cueOn = opts.CueOn
		}
		if opts.CueGap > 0 {
			cueGap = opts.CueGap
		}
		if opts.Settle > 0 {
			settle = opts.Settle
		}
	}
	if opts == nil || opts.Dwell <= 0 {
		if ms, err := strconv.Atoi(os.Getenv("DLM_CAPTURE_DWELL_MS")); err == nil && ms > 0 {
			dwell = time.Duration(ms) * time.Millisecond
		}
	}

	return &Controller{
		getter:  getter,
		drv:     drv,
		checker: checker,
		dwell:   dwell,
		cueOn:   cueOn,
		cueGap:  cueGap,
		settle:  settle,
		sweeps:  make(map[string]*sweepEntry),
	}
}

func (c *Controller) dwellTooShort() bool {
	usingDefaultBookend := c.cueOn == defaultCueOn && c.cueGap == defaultCueGap && c.settle == defaultSettle
	if !usingDefaultBookend {
		return false
	}
	if c.dwell < minDefaultDwell {
		return true
	}
	return c.dwell <= c.cueOn
}

// Start begins a capture sweep for deviceID.  It returns the initial Status
// (state="running") or one of:
//   - store.ErrDeviceNotFound  — device does not exist
//   - ErrCaptureNoLights       — device.light_count == 0
//   - ErrCaptureConflict       — sweep already running, or model has active routine
//   - ErrCaptureRoutineCheck   — active-routine guard could not be evaluated
func (c *Controller) Start(ctx context.Context, deviceID string) (Status, error) {
	d, err := c.getter.GetDevice(ctx, deviceID)
	if err != nil {
		return Status{}, err
	}
	if d.LightCount == 0 {
		return Status{}, ErrCaptureNoLights
	}

	if c.checker != nil && d.ModelID != nil && *d.ModelID != "" {
		busy, cherr := c.checker.ModelHasActiveRoutineRun(ctx, *d.ModelID)
		if cherr != nil {
			return Status{}, errors.Join(ErrCaptureRoutineCheck, cherr)
		}
		if busy {
			return Status{}, ErrCaptureConflict
		}
	}

	n := d.LightCount

	if c.dwellTooShort() {
		return Status{}, ErrCaptureDwellTooShort
	}

	c.mu.Lock()
	if entry, exists := c.sweeps[deviceID]; exists {
		entry.mu.Lock()
		active := entry.state == stateRunning || entry.state == stateStopping
		entry.mu.Unlock()
		if active {
			c.mu.Unlock()
			return Status{}, ErrCaptureConflict
		}
	}

	entry := &sweepEntry{
		state:      stateRunning,
		lightCount: n,
		phase:      "preamble",
		stop:       make(chan struct{}),
	}
	c.sweeps[deviceID] = entry
	c.mu.Unlock()

	go c.runSweep(d, entry, n)

	return Status{State: stateRunning, LightCount: n, Phase: "preamble"}, nil
}

func (c *Controller) waitOrStop(entry *sweepEntry, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-entry.stop:
		return false
	case <-timer.C:
		return true
	}
}

func (c *Controller) setPhase(entry *sweepEntry, phase string, idx int, haveIdx bool) {
	entry.mu.Lock()
	entry.phase = phase
	if haveIdx {
		entry.currentIndex = idx
	}
	entry.mu.Unlock()
}

// playBookend paints red, blue, green. settleAfter waits c.settle after the
// green pulse's all-off (opening signal). The closing signal passes false.
// Returns false when Stop fires. On Stop the strip is all off.
func (c *Controller) playBookend(ctx context.Context, d store.Device, entry *sweepEntry, n int, settleAfter bool) bool {
	colors := [][3]int{{255, 0, 0}, {0, 0, 255}, {0, 255, 0}}
	for i, rgb := range colors {
		_ = c.drv.DriveAllColor(ctx, d, n, rgb[0], rgb[1], rgb[2])
		if !c.waitOrStop(entry, c.cueOn) {
			_ = c.drv.DriveAllOff(ctx, d, n)
			return false
		}
		_ = c.drv.DriveAllOff(ctx, d, n)
		if i == len(colors)-1 {
			if !settleAfter {
				return true
			}
			return c.waitOrStop(entry, c.settle)
		}
		if !c.waitOrStop(entry, c.cueGap) {
			return false
		}
	}
	return true
}

// runSweep executes the LED sweep loop in a goroutine.
func (c *Controller) runSweep(d store.Device, entry *sweepEntry, n int) {
	ctx := context.Background()
	defer c.removeSweep(d.ID)

	c.setPhase(entry, "preamble", 0, false)
	if !c.playBookend(ctx, d, entry, n, true) {
		_ = c.drv.DriveAllOff(ctx, d, n)
		return
	}

	for idx := 0; idx < n; idx++ {
		c.setPhase(entry, "sweep", idx, true)
		_ = c.drv.DriveSingleLED(ctx, d, idx, n)
		if !c.waitOrStop(entry, c.dwell) {
			_ = c.drv.DriveAllOff(ctx, d, n)
			return
		}
	}

	c.setPhase(entry, "postamble", 0, false)
	_ = c.drv.DriveAllOff(ctx, d, n)
	if !c.waitOrStop(entry, c.settle) {
		return
	}
	if !c.playBookend(ctx, d, entry, n, false) {
		_ = c.drv.DriveAllOff(ctx, d, n)
		return
	}
}

func (c *Controller) removeSweep(deviceID string) {
	c.mu.Lock()
	delete(c.sweeps, deviceID)
	c.mu.Unlock()
}

// Stop signals the running sweep for deviceID to stop and waits for all LEDs
// to go off (within the 2 s REQ-040 bound).  Stopping an idle or unknown
// device is a no-op.
func (c *Controller) Stop(deviceID string) {
	c.mu.Lock()
	entry, exists := c.sweeps[deviceID]
	if !exists {
		c.mu.Unlock()
		return
	}
	entry.mu.Lock()
	if entry.state == stateRunning {
		entry.state = stateStopping
	}
	entry.mu.Unlock()
	c.mu.Unlock()
	entry.doStop()
}

// GetStatus returns the current sweep status for deviceID.  If no sweep entry
// exists the device is considered idle with LightCount=0 (callers may overlay
// the device's configured light_count).
func (c *Controller) GetStatus(deviceID string) Status {
	c.mu.Lock()
	entry, exists := c.sweeps[deviceID]
	c.mu.Unlock()
	if !exists {
		return Status{State: stateIdle}
	}
	entry.mu.Lock()
	defer entry.mu.Unlock()
	return Status{
		State:        entry.state,
		LightCount:   entry.lightCount,
		CurrentIndex: entry.currentIndex,
		Phase:        entry.phase,
	}
}

// Shutdown stops all active sweeps.  It should be called on process exit.
func (c *Controller) Shutdown() {
	c.mu.Lock()
	entries := make([]*sweepEntry, 0, len(c.sweeps))
	for _, e := range c.sweeps {
		entries = append(entries, e)
	}
	c.mu.Unlock()
	for _, e := range entries {
		e.doStop()
	}
}
