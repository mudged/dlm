package httpapi

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"example.com/dlm/backend/internal/capture"
	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/devices"
	"example.com/dlm/backend/internal/reconstruct"
	"example.com/dlm/backend/internal/routineengine"
	"example.com/dlm/backend/internal/store"
)

// engineCtrl is the interface the HTTP layer uses for the routine engine.
type engineCtrl interface {
	Start(ctx context.Context, runID, sceneID, routineID string) error
	Stop(runID string)
	Shutdown()
}

// captureCtrl is the interface the HTTP layer uses to drive capture sweeps.
type captureCtrl interface {
	Start(ctx context.Context, deviceID string) (capture.Status, error)
	Stop(deviceID string)
	GetStatus(deviceID string) capture.Status
	Shutdown()
}

// reconstructCtrl is the interface the HTTP layer uses to manage reconstruction jobs.
type reconstructCtrl interface {
	Create(ctx context.Context, files []io.Reader, names []string, params reconstruct.CreateParams) (string, error)
	Get(id string) (*reconstruct.Job, bool)
	Confirm(ctx context.Context, jobID, name string) (store.Summary, error)
	Discard(jobID string) error
	Shutdown()
}

// apiDeps holds API handlers' shared dependencies.
type apiDeps struct {
	store       *store.Store
	rev         *RevisionHub
	pusher      devicePusher
	engine      engineCtrl
	capture     captureCtrl
	reconstruct reconstructCtrl
}

// SiteHandler wraps the HTTP handler and holds references to background
// services so the caller can shut them all down on process exit or factory
// reset.  It implements http.Handler so it can be passed directly to
// http.Server and httptest.NewServer.
type SiteHandler struct {
	http.Handler
	engine      engineCtrl
	capture     captureCtrl
	reconstruct reconstructCtrl
}

// Shutdown stops the routine engine, capture sweeps, and reconstruction jobs.
// It should be called after the HTTP server has drained (srv.Shutdown) so that
// no new work is started while the workers are stopping.
//
// Ordering: engine first (may mutate light state), then capture (raw LED
// frames), then reconstruct (Python CV children).  The engine's Shutdown
// waits up to 5 s for goroutines to drain; the others are fire-and-cancel.
func (s *SiteHandler) Shutdown(ctx context.Context) {
	if s.engine != nil {
		s.engine.Shutdown()
	}
	if s.capture != nil {
		s.capture.Shutdown()
	}
	if s.reconstruct != nil {
		s.reconstruct.Shutdown()
	}
}

// NewSiteHandler wires /health, /api/v1/, and optional static UI from content (Next export).
// API routes are registered before the static file server. If content is nil, only API routes exist.
// st must be non-nil (models API requires persistence).
// rev may be nil; a private RevisionHub is used so in-process notifications still work (REQ-029 SSE).
// pusher may be nil; when set, logical light changes are pushed to assigned WLED devices (REQ-035–REQ-039).
func NewSiteHandler(cfg *config.Config, content fs.FS, st *store.Store, rev *RevisionHub, pusher *devices.Pusher) *SiteHandler {
	if st == nil {
		panic("httpapi.NewSiteHandler: store is nil")
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	if rev == nil {
		rev = NewRevisionHubWithLogger(log)
	}

	deps := &apiDeps{store: st, rev: rev, pusher: pusher}
	eng := routineengine.New(st, cfg.HTTPListen, func(ctx context.Context, sceneID string, states []store.ScenePatchedState) {
		deps.notifyAfterSceneLightPatch(ctx, sceneID, states)
	}, log)
	deps.engine = eng

	// When a real pusher is available, wire up a capture controller so the
	// sweep endpoints have a live driver.  When pusher is nil (API-only
	// construction, e.g. tests that pass nil) deps.capture stays nil and the
	// start handler returns 503 — see capture.go comment.
	var capCtrl captureCtrl
	if pusher != nil {
		capCtrl = capture.New(st, pusher, st, nil)
		deps.capture = capCtrl
	}

	// Wire the reconstruction manager. The work-dir base lives under DataDir so
	// it survives across restarts alongside the database.
	captureWorkDir := filepath.Join(cfg.DataDir, "runtime", "capture")
	recMgr := reconstruct.New(reconstruct.RealRunner{}, st, captureWorkDir)
	deps.reconstruct = recMgr

	return &SiteHandler{
		Handler:     buildSiteHandler(cfg, content, deps, log),
		engine:      eng,
		capture:     capCtrl,
		reconstruct: recMgr,
	}
}

// buildSiteHandler wires all routes and middleware from a pre-built apiDeps.
// It is also used by tests that need to inject a custom captureCtrl without a
// real WLED pusher.
func buildSiteHandler(cfg *config.Config, content fs.FS, deps *apiDeps, log *slog.Logger) http.Handler {
	if log == nil {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	api := http.NewServeMux()
	// Register every method on this path to the same handler so the
	// internal POST check can return the §3.2 JSON envelope on 405,
	// rather than falling through to the `GET /{path...}` 404 catch-all.
	for _, m := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions,
		http.MethodHead,
	} {
		api.HandleFunc(m+" /system/factory-reset", deps.postFactoryReset)
	}
	api.HandleFunc("GET /devices", deps.listDevices)
	api.HandleFunc("POST /devices", deps.postDevice)
	api.HandleFunc("POST /devices/discover", deps.postDevicesDiscover)
	api.HandleFunc("GET /devices/{id}", deps.getDevice)
	api.HandleFunc("PATCH /devices/{id}", deps.patchDevice)
	api.HandleFunc("DELETE /devices/{id}", deps.deleteDevice)
	api.HandleFunc("POST /devices/{id}/assign", deps.postDeviceAssign)
	api.HandleFunc("POST /devices/{id}/unassign", deps.postDeviceUnassign)
	api.HandleFunc("POST /devices/{id}/capture/start", deps.postCaptureStart)
	api.HandleFunc("POST /devices/{id}/capture/stop", deps.postCaptureStop)
	api.HandleFunc("GET /devices/{id}/capture", deps.getCaptureStatus)
	api.HandleFunc("GET /status", statusHandler)
	api.HandleFunc("GET /routines", deps.listRoutines)
	api.HandleFunc("POST /routines", deps.createRoutine)
	api.HandleFunc("GET /routines/{id}", deps.getRoutine)
	api.HandleFunc("PATCH /routines/{id}", deps.patchRoutine)
	api.HandleFunc("DELETE /routines/{id}", deps.deleteRoutine)
	api.HandleFunc("GET /models", deps.listModels)
	api.HandleFunc("POST /models", deps.createModel)
	// Capture/reconstruction routes must be registered before /models/{id} so
	// that "capture" is not mis-parsed as a model id (REQ-048, REQ-049).
	api.HandleFunc("GET /capture/marker", deps.getCaptureMarker)
	api.HandleFunc("POST /models/capture", deps.postModelsCapture)
	api.HandleFunc("GET /models/capture/{jobId}", deps.getModelsCaptureJob)
	api.HandleFunc("POST /models/capture/{jobId}/confirm", deps.postModelsCaptureConfirm)
	api.HandleFunc("DELETE /models/capture/{jobId}", deps.deleteModelsCapture)
	api.HandleFunc("DELETE /scenes/{id}/models/{modelId}", deps.deleteSceneModel)
	api.HandleFunc("PATCH /scenes/{id}/models/{modelId}", deps.patchSceneModel)
	api.HandleFunc("POST /scenes/{id}/models", deps.postSceneModel)
	api.HandleFunc("GET /scenes/{id}/dimensions", deps.getSceneDimensions)
	api.HandleFunc("GET /scenes/{id}/lights/events", deps.getSceneLightsEvents)
	api.HandleFunc("GET /scenes/{id}/lights", deps.listSceneLights)
	api.HandleFunc("POST /scenes/{id}/lights/query/cuboid", deps.postSceneLightsQueryCuboid)
	api.HandleFunc("POST /scenes/{id}/lights/query/sphere", deps.postSceneLightsQuerySphere)
	api.HandleFunc("PATCH /scenes/{id}/lights/state/cuboid", deps.patchSceneLightsStateCuboid)
	api.HandleFunc("PATCH /scenes/{id}/lights/state/sphere", deps.patchSceneLightsStateSphere)
	api.HandleFunc("PATCH /scenes/{id}/lights/state/scene", deps.patchSceneLightsStateScene)
	api.HandleFunc("PATCH /scenes/{id}/lights/state/batch", deps.patchSceneLightsStateBatch)
	api.HandleFunc("GET /scenes/{id}/routines/runs", deps.listSceneRoutineRuns)
	api.HandleFunc("POST /scenes/{id}/routines/runs/{runId}/stop", deps.postSceneRoutineRunStop)
	api.HandleFunc("POST /scenes/{id}/routines/{routineId}/start", deps.postSceneRoutineStart)
	api.HandleFunc("DELETE /scenes/{id}", deps.deleteScene)
	api.HandleFunc("PATCH /scenes/{id}", deps.patchScene)
	api.HandleFunc("GET /scenes/{id}", deps.getScene)
	api.HandleFunc("GET /scenes", deps.listScenes)
	api.HandleFunc("POST /scenes", deps.createScene)
	api.HandleFunc("GET /models/{id}/lights/state", deps.listLightStates)
	api.HandleFunc("GET /models/{id}/lights/events", deps.getModelLightsEvents)
	api.HandleFunc("POST /models/{id}/lights/state/reset", deps.postResetLightStates)
	api.HandleFunc("PATCH /models/{id}/lights/state/batch", deps.patchLightStatesBatch)
	api.HandleFunc("GET /models/{id}/lights/{lightId}/state", deps.getLightState)
	api.HandleFunc("PATCH /models/{id}/lights/{lightId}/state", deps.patchLightState)
	api.HandleFunc("GET /models/{id}", deps.getModel)
	api.HandleFunc("DELETE /models/{id}", deps.deleteModel)
	api.HandleFunc("GET /{path...}", apiNotFoundHandler)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	if content != nil {
		mux.Handle("/", staticExportHandler(content))
	}

	var h http.Handler = mux
	h = corsMiddleware(cfg.CORSAllowedOrigins, h)
	h = requestIDMiddleware(h)
	h = loggingMiddleware(log, h)
	h = recoverMiddleware(log, h)
	return h
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "dlm-api",
		"version": "0.1.0",
	})
}

func apiNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, http.StatusNotFound, "not_found", "no handler for this path")
}

// staticExportHandler serves the Next static export tree. Unmatched GET paths that look
// like client routes (no file extension) fall back to index.html; missing _next assets 404.
func staticExportHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		upath := path.Clean(r.URL.Path)
		rel := strings.TrimPrefix(upath, "/")
		if rel == "." || rel == "" {
			http.ServeFileFS(w, r, fsys, "index.html")
			return
		}

		f, err := fsys.Open(rel)
		if err == nil {
			info, ierr := f.Stat()
			_ = f.Close()
			if ierr == nil && info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		if shouldSPAFallback(rel) {
			// Use ServeFileFS so r.URL.Path is not rewritten to …/index.html (that triggers
			// net/http's canonical redirect and can loop with GET /).
			http.ServeFileFS(w, r, fsys, "index.html")
			return
		}
		http.NotFound(w, r)
	})
}

func shouldSPAFallback(rel string) bool {
	if strings.HasPrefix(rel, "_next/") {
		return false
	}
	return !strings.Contains(rel, ".")
}
