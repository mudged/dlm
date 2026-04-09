// Package routineengine supervises scene routine runs on the server (REQ-021, REQ-038 / architecture §3.16–§3.17.2).
package routineengine

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"example.com/dlm/backend/internal/routineengine/shapeanim"
	"example.com/dlm/backend/internal/store"
)

//go:embed bootstrap.py
var bootstrapScript []byte

// ScenePatchNotifier is invoked after successful batch scene light patches (SSE + WLED).
type ScenePatchNotifier func(ctx context.Context, sceneID string, states []store.ScenePatchedState)

// Engine supervises active routine_runs rows.
type Engine struct {
	store   *store.Store
	notify  ScenePatchNotifier
	log     *slog.Logger
	apiBase string
	python  string

	mu     sync.Mutex
	active map[string]context.CancelFunc
}

// New creates an engine. listenAddr is the same string passed to http.Server (e.g. ":8080").
func New(st *store.Store, listenAddr string, notify ScenePatchNotifier, log *slog.Logger) *Engine {
	if log == nil {
		log = slog.Default()
	}
	py := os.Getenv("DLM_PYTHON3")
	if py == "" {
		py = "python3"
	}
	return &Engine{
		store:   st,
		notify:  notify,
		log:     log,
		apiBase: APIBaseURL(listenAddr),
		python:  py,
		active:  make(map[string]context.CancelFunc),
	}
}

// Start begins supervision for a new run (call after the run row is committed).
func (e *Engine) Start(ctx context.Context, runID, sceneID, routineID string) error {
	r, err := e.store.GetRoutine(ctx, routineID)
	if err != nil {
		return err
	}
	switch r.Type {
	case store.RoutineTypePythonSceneScript:
		if _, err := exec.LookPath(e.python); err != nil {
			return errors.New("python3 not found on PATH (install python3 or set DLM_PYTHON3)")
		}
		e.startPython(ctx, runID, sceneID, r.PythonSource)
		return nil
	case store.RoutineTypeShapeAnimation:
		def := string(r.DefinitionJSON)
		if def == "" {
			return errors.New("shape_animation routine missing definition_json")
		}
		e.startShape(ctx, runID, sceneID, def)
		return nil
	default:
		return errors.New("unsupported routine type for server engine")
	}
}

// Stop cancels the supervisor for runID.
func (e *Engine) Stop(runID string) {
	e.mu.Lock()
	cancel := e.active[runID]
	delete(e.active, runID)
	e.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (e *Engine) registerCancel(runID string, cancel context.CancelFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if prev, ok := e.active[runID]; ok {
		prev()
	}
	e.active[runID] = cancel
}

func (e *Engine) unregister(runID string) {
	e.mu.Lock()
	delete(e.active, runID)
	e.mu.Unlock()
}

func (e *Engine) startPython(parent context.Context, runID, sceneID, source string) {
	ctx, cancel := context.WithCancel(parent)
	e.registerCancel(runID, cancel)

	go func() {
		defer e.unregister(runID)
		tmp, err := os.CreateTemp("", "dlm-user-*.py")
		if err != nil {
			e.log.Error("routineengine python temp", "err", err, "run_id", runID)
			return
		}
		path := tmp.Name()
		if _, err := tmp.WriteString(source); err != nil {
			_ = tmp.Close()
			_ = os.Remove(path)
			e.log.Error("routineengine python write", "err", err, "run_id", runID)
			return
		}
		_ = tmp.Close()
		defer func() { _ = os.Remove(path) }()

		cmd := exec.CommandContext(ctx, e.python, "-")
		cmd.Stdin = bytes.NewReader(bootstrapScript)
		cmd.Env = append(os.Environ(),
			"DLM_API_BASE="+e.apiBase,
			"DLM_SCENE_ID="+sceneID,
			"DLM_USER_SOURCE_PATH="+path,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil && ctx.Err() == nil {
			e.log.Warn("routineengine python exited", "err", err, "run_id", runID)
		}
	}()
}

func (e *Engine) startShape(parent context.Context, runID, sceneID, definitionJSON string) {
	ctx, cancel := context.WithCancel(parent)
	e.registerCancel(runID, cancel)
	rng := shapeanim.NewRng(uint32(time.Now().UnixNano() & 0xffffffff))

	dimsStore, err := e.store.GetSceneDimensions(ctx, sceneID)
	if err != nil {
		e.log.Error("routineengine shape init dimensions", "err", err, "run_id", runID)
		cancel()
		e.unregister(runID)
		return
	}
	dims := shapeanim.FromStore(dimsStore)
	sim, err := shapeanim.ParseAndInit(definitionJSON, dims, rng)
	if err != nil {
		e.log.Error("routineengine shape init", "err", err, "run_id", runID)
		cancel()
		e.unregister(runID)
		return
	}

	go func() {
		defer e.unregister(runID)
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dimsStore, err := e.store.GetSceneDimensions(ctx, sceneID)
				if err != nil {
					continue
				}
				d := shapeanim.FromStore(dimsStore)
				allStop := shapeanim.Tick(sim, d, rng)
				lights, err := e.store.ListSceneLights(ctx, sceneID)
				if err != nil {
					e.log.Warn("routineengine shape lights", "err", err)
					continue
				}
				updates := shapeanim.BuildBatchUpdates(sim, lights)
				if len(updates) > 0 {
					res, err := e.store.PatchSceneLightsBatch(ctx, sceneID, updates)
					if err != nil {
						e.log.Warn("routineengine shape batch", "err", err, "run_id", runID)
					} else if res != nil && res.UpdatedCount > 0 && e.notify != nil {
						e.notify(ctx, sceneID, res.States)
					}
				}
				if allStop {
					if err := e.store.StopRoutineRun(ctx, sceneID, runID); err != nil {
						e.log.Warn("routineengine shape auto-stop", "err", err, "run_id", runID)
					}
					return
				}
			}
		}
	}()
}
