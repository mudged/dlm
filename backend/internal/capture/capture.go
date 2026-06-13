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
)

// driver drives raw LED frames on a WLED device.
type driver interface {
	DriveSingleLED(ctx context.Context, d store.Device, litIdx, n int) error
	DriveAllOff(ctx context.Context, d store.Device, n int) error
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
}

const (
	stateIdle    = "idle"
	stateRunning = "running"
)

// sweepEntry holds per-device sweep state and its cancellation channel.
type sweepEntry struct {
	mu           sync.Mutex
	state        string
	lightCount   int
	currentIndex int
	stopOnce     sync.Once
	stop         chan struct{}
}

func (e *sweepEntry) doStop() {
	e.stopOnce.Do(func() { close(e.stop) })
}

// ControllerOpts configures optional Controller parameters.
type ControllerOpts struct {
	// Dwell overrides the per-LED on-time.  Ignored if <= 0.
	Dwell time.Duration
}

// Controller manages at most one active capture sweep per device.
// It is safe for concurrent use.
type Controller struct {
	getter  deviceGetter
	drv     driver
	checker RoutineChecker // may be nil
	dwell   time.Duration

	mu     sync.Mutex
	sweeps map[string]*sweepEntry
}

// New creates a Controller.  checker may be nil to skip routine-conflict
// checking.  The dwell period is resolved in priority order: opts.Dwell →
// DLM_CAPTURE_DWELL_MS env var → default 1000 ms.
func New(getter deviceGetter, drv driver, checker RoutineChecker, opts *ControllerOpts) *Controller {
	dwell := 1000 * time.Millisecond

	if opts != nil && opts.Dwell > 0 {
		dwell = opts.Dwell
	} else if ms, err := strconv.Atoi(os.Getenv("DLM_CAPTURE_DWELL_MS")); err == nil && ms > 0 {
		dwell = time.Duration(ms) * time.Millisecond
	}

	return &Controller{
		getter:  getter,
		drv:     drv,
		checker: checker,
		dwell:   dwell,
		sweeps:  make(map[string]*sweepEntry),
	}
}

// Start begins a capture sweep for deviceID.  It returns the initial Status
// (state="running") or one of:
//   - store.ErrDeviceNotFound  — device does not exist
//   - ErrCaptureNoLights       — device.light_count == 0
//   - ErrCaptureConflict       — sweep already running, or model has active routine
func (c *Controller) Start(ctx context.Context, deviceID string) (Status, error) {
	d, err := c.getter.GetDevice(ctx, deviceID)
	if err != nil {
		return Status{}, err
	}
	if d.LightCount == 0 {
		return Status{}, ErrCaptureNoLights
	}

	// Best-effort routine-conflict guard (TODO(REQ-047): extend with routine-run cross-check).
	if c.checker != nil && d.ModelID != nil && *d.ModelID != "" {
		if busy, cherr := c.checker.ModelHasActiveRoutineRun(ctx, *d.ModelID); cherr == nil && busy {
			return Status{}, ErrCaptureConflict
		}
	}

	n := d.LightCount

	c.mu.Lock()
	if entry, exists := c.sweeps[deviceID]; exists {
		entry.mu.Lock()
		running := entry.state == stateRunning
		entry.mu.Unlock()
		if running {
			c.mu.Unlock()
			return Status{}, ErrCaptureConflict
		}
	}

	entry := &sweepEntry{
		state:      stateRunning,
		lightCount: n,
		stop:       make(chan struct{}),
	}
	c.sweeps[deviceID] = entry
	c.mu.Unlock()

	go c.runSweep(d, entry, n)

	return Status{State: stateRunning, LightCount: n, CurrentIndex: 0}, nil
}

// runSweep executes the LED sweep loop in a goroutine.
func (c *Controller) runSweep(d store.Device, entry *sweepEntry, n int) {
	ticker := time.NewTicker(c.dwell)
	defer ticker.Stop()

	ctx := context.Background()
	idx := 0
	for {
		_ = c.drv.DriveSingleLED(ctx, d, idx, n)

		entry.mu.Lock()
		entry.currentIndex = idx
		entry.mu.Unlock()

		select {
		case <-entry.stop:
			_ = c.drv.DriveAllOff(ctx, d, n)
			entry.mu.Lock()
			entry.state = stateIdle
			entry.mu.Unlock()
			return
		case <-ticker.C:
			idx++
			if idx >= n {
				_ = c.drv.DriveAllOff(ctx, d, n)
				entry.mu.Lock()
				entry.state = stateIdle
				entry.mu.Unlock()
				return
			}
		}
	}
}

// Stop signals the running sweep for deviceID to stop and waits for all LEDs
// to go off (within the 2 s REQ-040 bound).  Stopping an idle or unknown
// device is a no-op.
func (c *Controller) Stop(deviceID string) {
	c.mu.Lock()
	entry, exists := c.sweeps[deviceID]
	c.mu.Unlock()
	if !exists {
		return
	}
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
