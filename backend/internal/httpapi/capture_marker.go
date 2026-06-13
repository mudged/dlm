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

	format := r.URL.Query().Get("type")
	switch format {
	case "", "pdf", "aruco":
		serveMarkerAsset(w, fiducialMarkerPDF, "application/pdf", "fiducial_marker_aruco4x4_50_id0_100mm.pdf")
	case "png":
		serveMarkerAsset(w, fiducialMarkerPNG, "image/png", "fiducial_marker_aruco4x4_50_id0_100mm.png")
	default:
		// Unknown type/size: serve the sensible PDF default (REQ-049 optional params).
		serveMarkerAsset(w, fiducialMarkerPDF, "application/pdf", "fiducial_marker_aruco4x4_50_id0_100mm.pdf")
	}
}

func serveMarkerAsset(w http.ResponseWriter, data []byte, contentType, filename string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename="+filename)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
