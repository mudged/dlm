package routineengine

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"example.com/dlm/backend/internal/store"
)

// fakeEngineStore is a test double for engineStore.
type fakeEngineStore struct {
	mu sync.Mutex

	routine       *store.RoutineDTO
	getRoutineErr error
	getDimsErr    error

	stopCalls []stopCall
}

type stopCall struct {
	sceneID, runID string
}

func (f *fakeEngineStore) GetRoutine(_ context.Context, _ string) (*store.RoutineDTO, error) {
	return f.routine, f.getRoutineErr
}

func (f *fakeEngineStore) GetSceneDimensions(_ context.Context, _ string) (*store.SceneDimensions, error) {
	if f.getDimsErr != nil {
		return nil, f.getDimsErr
	}
	return &store.SceneDimensions{}, nil
}

func (f *fakeEngineStore) StopRoutineRun(_ context.Context, sceneID, runID string) error {
	f.mu.Lock()
	f.stopCalls = append(f.stopCalls, stopCall{sceneID, runID})
	f.mu.Unlock()
	return nil
}

func (f *fakeEngineStore) ListSceneLights(_ context.Context, _ string) ([]store.SceneLightFlat, error) {
	return nil, nil
}

func (f *fakeEngineStore) PatchSceneLightsBatch(_ context.Context, _ string, _ []store.SceneBatchLightUpdate) (*store.SceneBulkPatchResult, error) {
	return nil, nil
}

func (f *fakeEngineStore) stopCallsSnapshot() []stopCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]stopCall, len(f.stopCalls))
	copy(out, f.stopCalls)
	return out
}

// newTestEngine builds an Engine backed by fk, with python set to the given
// executable (use "true" for a fast-exiting no-op).
func newTestEngine(fk *fakeEngineStore, python string) *Engine {
	return &Engine{
		store:   fk,
		notify:  nil,
		log:     slog.Default(),
		apiBase: "http://127.0.0.1:8080",
		python:  python,
		active:  make(map[string]context.CancelFunc),
	}
}

// waitForStopCall polls until fk has at least one StopRoutineRun entry or the
// deadline passes.
func waitForStopCall(t *testing.T, fk *fakeEngineStore, timeout time.Duration) []stopCall {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if calls := fk.stopCallsSnapshot(); len(calls) > 0 {
			return calls
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fk.stopCallsSnapshot()
}

// validShapeDefJSON is a minimal but valid shape_animation definition accepted
// by both store.ValidateShapeAnimationDefinitionJSON and shapeanim.ParseAndInit.
var validShapeDefJSON = json.RawMessage(`{
	"version": 1,
	"background": {"mode": "lights_off"},
	"shapes": [{
		"kind": "sphere",
		"size": {"mode": "fixed", "radius_m": 0.1},
		"color": {"mode": "fixed", "color": "#ff0000"},
		"brightness_pct": 100,
		"placement": {"mode": "fixed", "center_m": {"x": 0, "y": 0, "z": 0}},
		"motion": {"direction": {"dx": 1, "dy": 0, "dz": 0}, "speed": {"mode": "fixed", "m_s": 0.1}},
		"edge_behavior": "wrap"
	}]
}`)

// TestEngine_pythonProcessExit_stopsRunInStore verifies that when the supervised
// Python process exits (success or failure), StopRoutineRun is called on the
// store so the routine_runs row leaves the "running" state.
// Uses "true" as the python executable so the "process" exits immediately.
func TestEngine_pythonProcessExit_stopsRunInStore(t *testing.T) {
	fk := &fakeEngineStore{
		routine: &store.RoutineDTO{
			Type:         store.RoutineTypePythonSceneScript,
			PythonSource: "pass",
		},
	}
	e := newTestEngine(fk, "true")

	if err := e.Start(context.Background(), "run-1", "scene-1", "routine-1"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	calls := waitForStopCall(t, fk, 5*time.Second)
	if len(calls) != 1 {
		t.Fatalf("want exactly 1 StopRoutineRun call, got %d: %+v", len(calls), calls)
	}
	if calls[0].sceneID != "scene-1" || calls[0].runID != "run-1" {
		t.Fatalf("StopRoutineRun called with wrong args: %+v", calls[0])
	}
}

// TestEngine_shapeInitDimsFailure_stopsRunInStore verifies that when
// GetSceneDimensions returns an error during shape animation initialisation,
// StopRoutineRun is called so the run does not remain stuck as "running".
func TestEngine_shapeInitDimsFailure_stopsRunInStore(t *testing.T) {
	fk := &fakeEngineStore{
		routine: &store.RoutineDTO{
			Type:           store.RoutineTypeShapeAnimation,
			DefinitionJSON: validShapeDefJSON,
		},
		getDimsErr: errors.New("scene has no dimensions"),
	}
	e := newTestEngine(fk, "python3")

	// startShape init runs synchronously; Start returns before the goroutine
	// is spawned (init error path never reaches the goroutine).
	if err := e.Start(context.Background(), "run-2", "scene-2", "routine-2"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	calls := fk.stopCallsSnapshot()
	if len(calls) != 1 {
		t.Fatalf("want exactly 1 StopRoutineRun call, got %d: %+v", len(calls), calls)
	}
	if calls[0].sceneID != "scene-2" || calls[0].runID != "run-2" {
		t.Fatalf("StopRoutineRun called with wrong args: %+v", calls[0])
	}
}

// TestEngine_Shutdown_cancelsActiveRunsAndDrains verifies that Shutdown cancels
// every active run, waits for goroutines to exit, and that StopRoutineRun is
// called for each — bringing the DB rows to a terminal state.
func TestEngine_Shutdown_cancelsActiveRunsAndDrains(t *testing.T) {
	fk := &fakeEngineStore{
		routine: &store.RoutineDTO{
			Type:           store.RoutineTypeShapeAnimation,
			DefinitionJSON: validShapeDefJSON,
		},
	}
	e := newTestEngine(fk, "python3")

	// Start two runs so we can verify both are stopped.
	for _, id := range []string{"run-s1", "run-s2"} {
		if err := e.Start(context.Background(), id, "scene-s", "routine-s"); err != nil {
			t.Fatalf("Start %s: %v", id, err)
		}
	}

	e.Shutdown()

	// After Shutdown the active map must be empty.
	e.mu.Lock()
	remaining := len(e.active)
	e.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("active map has %d entries after Shutdown, want 0", remaining)
	}

	// Both runs must have reached a terminal store state.
	calls := fk.stopCallsSnapshot()
	stopped := make(map[string]bool)
	for _, c := range calls {
		stopped[c.runID] = true
	}
	for _, id := range []string{"run-s1", "run-s2"} {
		if !stopped[id] {
			t.Errorf("StopRoutineRun not called for %s after Shutdown", id)
		}
	}
}

// TestEngine_shapeContextCancelled_stopsRunInStore verifies that when the shape
// animation goroutine's context is cancelled (e.g. stop endpoint cancels it),
// StopRoutineRun is called to reach a terminal DB state even if the stop
// endpoint also calls it (idempotent).
func TestEngine_shapeContextCancelled_stopsRunInStore(t *testing.T) {
	fk := &fakeEngineStore{
		routine: &store.RoutineDTO{
			Type:           store.RoutineTypeShapeAnimation,
			DefinitionJSON: validShapeDefJSON,
		},
	}
	e := newTestEngine(fk, "python3")

	if err := e.Start(context.Background(), "run-3", "scene-3", "routine-3"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Simulate the stop endpoint: cancel via engine.Stop.
	e.Stop("run-3")

	calls := waitForStopCall(t, fk, 5*time.Second)
	if len(calls) == 0 {
		t.Fatal("want at least 1 StopRoutineRun call after engine.Stop, got 0")
	}
	for _, c := range calls {
		if c.sceneID != "scene-3" || c.runID != "run-3" {
			t.Errorf("unexpected StopRoutineRun args: %+v", c)
		}
	}
}
