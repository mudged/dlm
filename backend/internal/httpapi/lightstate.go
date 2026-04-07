package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"example.com/dlm/backend/internal/store"
)

const maxLightPatchBytes = 8192
const maxLightBatchPatchBytes = 65536

type jsonLightStatePatch struct {
	On            *bool    `json:"on"`
	Color         *string  `json:"color"`
	BrightnessPct *float64 `json:"brightness_pct"`
}

func (a *apiDeps) postResetLightStates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	modelID := r.PathValue("id")
	if modelID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	_, _ = io.Copy(io.Discard, r.Body)

	states, unchangedAll, err := a.store.ResetAllLightStates(r.Context(), modelID)
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not reset light states")
		return
	}
	if states == nil {
		states = []store.LightStateDTO{}
	}
	if !unchangedAll {
		a.rev.NotifyModelLightsChanged(r.Context(), a.store, modelID)
	}
	resp := map[string]any{"states": states}
	if unchangedAll {
		resp["unchanged_all"] = true
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *apiDeps) listLightStates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	modelID := r.PathValue("id")
	if modelID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id")
		return
	}
	states, err := a.store.ListLightStates(r.Context(), modelID)
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load light states")
		return
	}
	if states == nil {
		states = []store.LightStateDTO{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"states": states})
}

func (a *apiDeps) getLightState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	modelID := r.PathValue("id")
	lightID, ok := parseLightPathID(r.PathValue("lightId"))
	if modelID == "" || !ok {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id or light id")
		return
	}
	st, err := a.store.GetLightState(r.Context(), modelID, lightID)
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}
	if errors.Is(err, store.ErrInvalidLightIndex) {
		writeAPIError(w, http.StatusNotFound, "not_found", "light not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load light state")
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (a *apiDeps) patchLightState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	modelID := r.PathValue("id")
	lightID, ok := parseLightPathID(r.PathValue("lightId"))
	if modelID == "" || !ok {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id or light id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxLightPatchBytes)
	var body jsonLightStatePatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	patch := store.LightStatePatch{
		On:            body.On,
		Color:         body.Color,
		BrightnessPct: body.BrightnessPct,
	}
	if patch.On == nil && patch.Color == nil && patch.BrightnessPct == nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "at least one of on, color, brightness_pct is required")
		return
	}

	st, unchanged, err := a.store.PatchLightState(r.Context(), modelID, lightID, patch)
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}
	if errors.Is(err, store.ErrInvalidLightIndex) {
		writeAPIError(w, http.StatusNotFound, "not_found", "light not found")
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "color") || strings.Contains(strings.ToLower(msg), "brightness") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update light state")
		return
	}
	if !unchanged {
		a.rev.NotifyModelLightsChanged(r.Context(), a.store, modelID)
	}
	out := map[string]any{
		"id":             st.ID,
		"on":             st.On,
		"color":          st.Color,
		"brightness_pct": st.BrightnessPct,
	}
	if unchanged {
		out["unchanged"] = true
	}
	writeJSON(w, http.StatusOK, out)
}

func parseLightPathID(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

type jsonBatchLightStatePatch struct {
	Ids             []int    `json:"ids"`
	On              *bool    `json:"on"`
	Color           *string  `json:"color"`
	BrightnessPct   *float64 `json:"brightness_pct"`
}

func (a *apiDeps) patchLightStatesBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	modelID := r.PathValue("id")
	if modelID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxLightBatchPatchBytes)
	var body jsonBatchLightStatePatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if body.Ids == nil || len(body.Ids) == 0 {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "ids must be a non-empty array")
		return
	}

	patch := store.LightStatePatch{
		On:            body.On,
		Color:         body.Color,
		BrightnessPct: body.BrightnessPct,
	}
	if patch.On == nil && patch.Color == nil && patch.BrightnessPct == nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "at least one of on, color, brightness_pct is required")
		return
	}

	states, unchangedAll, err := a.store.BatchPatchLightStates(r.Context(), modelID, body.Ids, patch)
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}
	if errors.Is(err, store.ErrBatchEmptyIDs) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if errors.Is(err, store.ErrBatchDuplicateIDs) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "duplicate light ids in batch")
		return
	}
	if errors.Is(err, store.ErrInvalidLightIndex) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "one or more light ids are out of range for this model")
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "color") || strings.Contains(strings.ToLower(msg), "brightness") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update light states")
		return
	}
	if states == nil {
		states = []store.LightStateDTO{}
	}
	if !unchangedAll {
		a.rev.NotifyModelLightsChanged(r.Context(), a.store, modelID)
	}
	resp := map[string]any{"states": states}
	if unchangedAll {
		resp["unchanged_all"] = true
	}
	writeJSON(w, http.StatusOK, resp)
}
