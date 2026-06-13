package capture_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"example.com/dlm/backend/internal/capture"
	"example.com/dlm/backend/internal/store"
)

// fakeDriver records all drive calls and can be inspected after a sweep.
type fakeDriver struct {
	mu    sync.Mutex
	calls []driveCall
}

type driveCall struct {
	kind string // "single" or "alloff"
	idx  int
}

func (f *fakeDriver) DriveSingleLED(_ context.Context, _ store.Device, litIdx, _ int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, driveCall{kind: "single", idx: litIdx})
	return nil
}

func (f *fakeDriver) DriveAllOff(_ context.Context, _ store.Device, _ int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, driveCall{kind: "alloff"})
	return nil
}

func (f *fakeDriver) snapshot() []driveCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]driveCall, len(f.calls))
	copy(out, f.calls)
	return out
}

// fakeGetter returns a fixed device.
type fakeGetter struct {
	device store.Device
	err    error
}

func (g *fakeGetter) GetDevice(_ context.Context, _ string) (store.Device, error) {
	return g.device, g.err
}

func newController(t *testing.T, d store.Device, drv *fakeDriver) *capture.Controller {
	t.Helper()
	getter := &fakeGetter{device: d}
	return capture.New(getter, drv, nil, &capture.ControllerOpts{Dwell: 20 * time.Millisecond})
}

func TestCapture_sweepCallsInOrder_thenAllOff(t *testing.T) {
	drv := &fakeDriver{}
	dev := store.Device{ID: "d1", LightCount: 3}
	ctrl := newController(t, dev, drv)

	st, err := ctrl.Start(context.Background(), "d1")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if st.State != "running" {
		t.Fatalf("initial state = %q want running", st.State)
	}

	// Wait generously for the 3-LED sweep + alloff to complete.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		calls := drv.snapshot()
		if len(calls) >= 4 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	calls := drv.snapshot()
	if len(calls) < 4 {
		t.Fatalf("expected ≥4 calls (3 single + alloff), got %d: %+v", len(calls), calls)
	}

	// Verify sequence: single(0), single(1), single(2), alloff
	for i, want := range []driveCall{
		{kind: "single", idx: 0},
		{kind: "single", idx: 1},
		{kind: "single", idx: 2},
		{kind: "alloff"},
	} {
		if calls[i] != want {
			t.Errorf("calls[%d] = %+v, want %+v", i, calls[i], want)
		}
	}

	// Controller should be idle after completion.
	deadline2 := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline2) {
		if ctrl.GetStatus("d1").State == "idle" {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if ctrl.GetStatus("d1").State != "idle" {
		t.Fatalf("state after completion = %q want idle", ctrl.GetStatus("d1").State)
	}
}

func TestCapture_stopMidSweep_callsAllOff(t *testing.T) {
	drv := &fakeDriver{}
	dev := store.Device{ID: "d2", LightCount: 10}
	// Use a longer dwell so we can stop mid-sweep reliably.
	ctrl := capture.New(&fakeGetter{device: dev}, drv, nil, &capture.ControllerOpts{Dwell: 100 * time.Millisecond})

	if _, err := ctrl.Start(context.Background(), "d2"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Allow one LED to be driven, then stop.
	time.Sleep(30 * time.Millisecond)
	ctrl.Stop("d2")

	// Wait for the goroutine to finish.
	deadline := time.Now().Add(400 * time.Millisecond)
	for time.Now().Before(deadline) {
		if ctrl.GetStatus("d2").State == "idle" {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if ctrl.GetStatus("d2").State != "idle" {
		t.Fatalf("state after Stop = %q want idle", ctrl.GetStatus("d2").State)
	}

	calls := drv.snapshot()
	if len(calls) == 0 {
		t.Fatal("expected at least one call")
	}
	last := calls[len(calls)-1]
	if last.kind != "alloff" {
		t.Fatalf("last call = %+v, want alloff", last)
	}

	// No further single calls after alloff.
	for i := range calls {
		if calls[i].kind == "alloff" {
			if i < len(calls)-1 {
				t.Fatalf("calls after alloff: %+v", calls[i+1:])
			}
			break
		}
	}
}

func TestCapture_startTwice_returnsConflict(t *testing.T) {
	drv := &fakeDriver{}
	dev := store.Device{ID: "d3", LightCount: 5}
	ctrl := capture.New(&fakeGetter{device: dev}, drv, nil, &capture.ControllerOpts{Dwell: 200 * time.Millisecond})

	if _, err := ctrl.Start(context.Background(), "d3"); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	_, err := ctrl.Start(context.Background(), "d3")
	if !errors.Is(err, capture.ErrCaptureConflict) {
		t.Fatalf("second Start err = %v, want ErrCaptureConflict", err)
	}
	ctrl.Stop("d3")
}

func TestCapture_zeroLightCount_returnsNoLights(t *testing.T) {
	drv := &fakeDriver{}
	dev := store.Device{ID: "d4", LightCount: 0}
	ctrl := newController(t, dev, drv)

	_, err := ctrl.Start(context.Background(), "d4")
	if !errors.Is(err, capture.ErrCaptureNoLights) {
		t.Fatalf("err = %v, want ErrCaptureNoLights", err)
	}
}

func TestCapture_stopIdleDevice_isNoop(t *testing.T) {
	drv := &fakeDriver{}
	dev := store.Device{ID: "d5", LightCount: 5}
	ctrl := newController(t, dev, drv)
	// Should not panic or error.
	ctrl.Stop("d5")
}

func TestCapture_deviceNotFound(t *testing.T) {
	drv := &fakeDriver{}
	getter := &fakeGetter{err: store.ErrDeviceNotFound}
	ctrl := capture.New(getter, drv, nil, &capture.ControllerOpts{Dwell: 20 * time.Millisecond})

	_, err := ctrl.Start(context.Background(), "missing")
	if !errors.Is(err, store.ErrDeviceNotFound) {
		t.Fatalf("err = %v, want ErrDeviceNotFound", err)
	}
}
