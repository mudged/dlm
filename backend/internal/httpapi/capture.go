package httpapi

import (
	"errors"
	"net/http"

	"example.com/dlm/backend/internal/capture"
	"example.com/dlm/backend/internal/store"
)

func (a *apiDeps) postCaptureStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing device id")
		return
	}
	if a.capture == nil {
		// capture controller unavailable (API-only build without a pusher).
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", "capture controller not available")
		return
	}
	st, err := a.capture.Start(r.Context(), id)
	if errors.Is(err, store.ErrDeviceNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "device not found")
		return
	}
	if errors.Is(err, capture.ErrCaptureNoLights) {
		writeAPIError(w, http.StatusUnprocessableEntity, "capture_no_lights", "device has no lights configured")
		return
	}
	if errors.Is(err, capture.ErrCaptureConflict) {
		writeAPIError(w, http.StatusConflict, "capture_conflict", "a capture sweep is already running for this device")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not start capture")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"device_id":     id,
		"state":         st.State,
		"light_count":   st.LightCount,
		"current_index": st.CurrentIndex,
	})
}

func (a *apiDeps) postCaptureStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing device id")
		return
	}
	if _, err := a.store.GetDevice(r.Context(), id); errors.Is(err, store.ErrDeviceNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "device not found")
		return
	} else if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load device")
		return
	}
	if a.capture != nil {
		a.capture.Stop(id)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"device_id": id,
		"state":     "idle",
	})
}

func (a *apiDeps) getCaptureStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing device id")
		return
	}
	d, err := a.store.GetDevice(r.Context(), id)
	if errors.Is(err, store.ErrDeviceNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "device not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load device")
		return
	}

	var st capture.Status
	if a.capture != nil {
		st = a.capture.GetStatus(id)
	} else {
		st = capture.Status{State: "idle"}
	}

	// Use the device's configured light_count when the controller hasn't
	// recorded one (e.g. sweep never started on this device).
	if st.LightCount == 0 {
		st.LightCount = d.LightCount
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"state":         st.State,
		"light_count":   st.LightCount,
		"current_index": st.CurrentIndex,
	})
}
