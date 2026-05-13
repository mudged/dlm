package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"example.com/dlm/backend/internal/store"
)

const maxDeviceBodyBytes = 16384

func deviceToJSON(d store.Device) map[string]any {
	m := map[string]any{
		"id":         d.ID,
		"type":       d.Type,
		"name":       d.Name,
		"base_url":   d.BaseURL,
		"created_at": d.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if d.ModelID != nil && *d.ModelID != "" {
		m["model_id"] = *d.ModelID
	}
	return m
}

func (a *apiDeps) listDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	list, err := a.store.ListDevices(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not list devices")
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, d := range list {
		out = append(out, deviceToJSON(d))
	}
	writeJSON(w, http.StatusOK, map[string]any{"devices": out})
}

type jsonDeviceCreate struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	BaseURL      string `json:"base_url"`
	WLEDPassword string `json:"wled_password"`
}

func (a *apiDeps) postDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxDeviceBodyBytes)
	var body jsonDeviceCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if body.Type == "" {
		body.Type = store.DeviceTypeWLED
	}
	d, err := a.store.CreateDevice(r.Context(), store.DeviceCreate{
		Type:         body.Type,
		Name:         body.Name,
		BaseURL:      body.BaseURL,
		WLEDPassword: body.WLEDPassword,
	})
	if errors.Is(err, store.ErrInvalidBaseURL) {
		writeAPIError(w, http.StatusBadRequest, "invalid_base_url", err.Error())
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unsupported") || strings.Contains(msg, "required") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not create device")
		return
	}
	writeJSON(w, http.StatusCreated, deviceToJSON(d))
}

func (a *apiDeps) postDevicesDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	_, _ = io.Copy(io.Discard, r.Body)
	writeAPIError(w, http.StatusNotImplemented, "not_implemented", "device discovery is not implemented yet; add devices manually per REQ-035")
}

func (a *apiDeps) getDevice(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, deviceToJSON(d))
}

type jsonDevicePatch struct {
	Name         *string `json:"name"`
	BaseURL      *string `json:"base_url"`
	WLEDPassword *string `json:"wled_password"`
}

func (a *apiDeps) patchDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing device id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxDeviceBodyBytes)
	var body jsonDevicePatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	d, err := a.store.PatchDevice(r.Context(), id, store.DevicePatch{
		Name:         body.Name,
		BaseURL:      body.BaseURL,
		WLEDPassword: body.WLEDPassword,
	})
	if errors.Is(err, store.ErrDeviceNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "device not found")
		return
	}
	if errors.Is(err, store.ErrInvalidBaseURL) {
		writeAPIError(w, http.StatusBadRequest, "invalid_base_url", err.Error())
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "cannot be empty") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update device")
		return
	}
	writeJSON(w, http.StatusOK, deviceToJSON(d))
}

func (a *apiDeps) deleteDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing device id")
		return
	}
	if err := a.store.DeleteDevice(r.Context(), id); errors.Is(err, store.ErrDeviceNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "device not found")
		return
	} else if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not delete device")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type jsonDeviceAssign struct {
	ModelID string `json:"model_id"`
}

func (a *apiDeps) postDeviceAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing device id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var body jsonDeviceAssign
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	body.ModelID = strings.TrimSpace(body.ModelID)
	if body.ModelID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "model_id is required")
		return
	}
	err := a.store.AssignDevice(r.Context(), id, body.ModelID)
	if errors.Is(err, store.ErrDeviceNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "device not found")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}
	if errors.Is(err, store.ErrDeviceAssignmentConflict) {
		writeAPIError(w, http.StatusConflict, "assignment_conflict", err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not assign device")
		return
	}
	d, err := a.store.GetDevice(r.Context(), id)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load device")
		return
	}
	if a.pusher != nil {
		_ = a.pusher.PushModel(r.Context(), body.ModelID)
	}
	writeJSON(w, http.StatusOK, deviceToJSON(d))
}

func (a *apiDeps) postDeviceUnassign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing device id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 256)
	_, _ = io.Copy(io.Discard, r.Body)
	if err := a.store.UnassignDevice(r.Context(), id); errors.Is(err, store.ErrDeviceNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "device not found")
		return
	} else if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not unassign device")
		return
	}
	d, err := a.store.GetDevice(r.Context(), id)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load device")
		return
	}
	writeJSON(w, http.StatusOK, deviceToJSON(d))
}
