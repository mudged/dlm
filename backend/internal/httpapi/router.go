package httpapi

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"example.com/dlm/backend/internal/config"
)

// NewSiteHandler wires /health, /api/v1/, and optional static UI from content (Next export).
// API routes are registered before the static file server. If content is nil, only API routes exist.
func NewSiteHandler(cfg *config.Config, content fs.FS) http.Handler {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	api := http.NewServeMux()
	api.HandleFunc("GET /status", statusHandler)
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
