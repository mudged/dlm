package httpapi

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/store"
)

// apiDeps holds API handlers' shared dependencies.
type apiDeps struct {
	store *store.Store
}

// NewSiteHandler wires /health, /api/v1/, and optional static UI from content (Next export).
// API routes are registered before the static file server. If content is nil, only API routes exist.
// st must be non-nil (models API requires persistence).
func NewSiteHandler(cfg *config.Config, content fs.FS, st *store.Store) http.Handler {
	if st == nil {
		panic("httpapi.NewSiteHandler: store is nil")
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	api := http.NewServeMux()
	deps := &apiDeps{store: st}
	api.HandleFunc("POST /system/factory-reset", deps.postFactoryReset)
	api.HandleFunc("GET /status", statusHandler)
	api.HandleFunc("GET /routines", deps.listRoutines)
	api.HandleFunc("POST /routines", deps.createRoutine)
	api.HandleFunc("DELETE /routines/{id}", deps.deleteRoutine)
	api.HandleFunc("GET /models", deps.listModels)
	api.HandleFunc("POST /models", deps.createModel)
	api.HandleFunc("DELETE /scenes/{id}/models/{modelId}", deps.deleteSceneModel)
	api.HandleFunc("PATCH /scenes/{id}/models/{modelId}", deps.patchSceneModel)
	api.HandleFunc("POST /scenes/{id}/models", deps.postSceneModel)
	api.HandleFunc("GET /scenes/{id}/dimensions", deps.getSceneDimensions)
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
	api.HandleFunc("GET /scenes/{id}", deps.getScene)
	api.HandleFunc("GET /scenes", deps.listScenes)
	api.HandleFunc("POST /scenes", deps.createScene)
	api.HandleFunc("GET /models/{id}/lights/state", deps.listLightStates)
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
