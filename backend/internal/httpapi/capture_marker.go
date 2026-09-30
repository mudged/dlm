package httpapi

import (
	"net/http"
	"strconv"
)

// getCaptureMarker serves the optional printable fiducial marker (REQ-049 BR 5).
// GET /capture/marker
func (a *apiDeps) getCaptureMarker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id, ok := markerQueryID(r.URL.Query().Get("id"))
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "marker id must be 0, 1, or 2")
		return
	}
	format := r.URL.Query().Get("type")
	png := format == "png"
	switch format {
	case "", "pdf", "aruco", "png":
	default:
		png = false
	}
	data, contentType, filename := markerAsset(id, png)
	serveMarkerAsset(w, data, contentType, filename)
}

func markerQueryID(raw string) (int, bool) {
	if raw == "" {
		return 0, true
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id < 0 || id > 2 {
		return 0, false
	}
	return id, true
}

func serveMarkerAsset(w http.ResponseWriter, data []byte, contentType, filename string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename="+filename)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
