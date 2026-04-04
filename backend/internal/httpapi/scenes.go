package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"example.com/dlm/backend/internal/store"
)

const maxSceneJSONBytes = 1 << 18 // 256 KiB

type createSceneBody struct {
	Name   string            `json:"name"`
	Models []json.RawMessage `json:"models"`
}

type addSceneModelBody struct {
	ModelID string `json:"model_id"`
	OffsetX *int   `json:"offset_x"`
	OffsetY *int   `json:"offset_y"`
	OffsetZ *int   `json:"offset_z"`
}

type patchSceneModelBody struct {
	OffsetX int `json:"offset_x"`
	OffsetY int `json:"offset_y"`
	OffsetZ int `json:"offset_z"`
}

func (a *apiDeps) listScenes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	list, err := a.store.ListScenes(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not list scenes")
		return
	}
	if list == nil {
		list = []store.SceneSummary{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *apiDeps) createScene(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSceneJSONBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "could not read body")
		return
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var body createSceneBody
	if err := dec.Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if dec.More() {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "name is required")
		return
	}
	if len(body.Models) == 0 {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "at least one model is required")
		return
	}
	modelIDs := make([]string, 0, len(body.Models))
	for _, rm := range body.Models {
		if len(strings.TrimSpace(string(rm))) == 0 {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "each model entry must be an object")
			return
		}
		itemDec := json.NewDecoder(bytes.NewReader(rm))
		itemDec.DisallowUnknownFields()
		var item struct {
			ModelID string `json:"model_id"`
		}
		if err := itemDec.Decode(&item); err != nil {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "each model object must contain only model_id (offsets are computed on create)")
			return
		}
		if itemDec.More() {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "invalid model entry")
			return
		}
		mid := strings.TrimSpace(item.ModelID)
		if mid == "" {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "each model requires model_id")
			return
		}
		modelIDs = append(modelIDs, mid)
	}

	sum, err := a.store.CreateScene(r.Context(), body.Name, modelIDs)
	if errors.Is(err, store.ErrDuplicateSceneName) {
		writeAPIError(w, http.StatusConflict, "conflict", "a scene with this name already exists")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "unknown model_id in request")
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "outside the non-negative") || strings.Contains(msg, "offsets must") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		if strings.Contains(msg, "duplicate model_id") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not create scene")
		return
	}
	writeJSON(w, http.StatusCreated, sum)
}

func (a *apiDeps) getScene(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	d, err := a.store.GetScene(r.Context(), id)
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load scene")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (a *apiDeps) deleteScene(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	if err := a.store.DeleteScene(r.Context(), id); errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	} else if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not delete scene")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiDeps) postSceneModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	if sceneID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSceneJSONBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "could not read body")
		return
	}
	var body addSceneModelBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	mid := strings.TrimSpace(body.ModelID)
	if mid == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "model_id is required")
		return
	}

	pl, err := a.store.AddSceneModel(r.Context(), sceneID, mid, body.OffsetX, body.OffsetY, body.OffsetZ)
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "model not found")
		return
	}
	if errors.Is(err, store.ErrModelAlreadyInScene) {
		writeAPIError(w, http.StatusConflict, "conflict", "model is already in this scene")
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "outside the non-negative") || strings.Contains(msg, "offsets must") || strings.Contains(msg, "either provide") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not add model to scene")
		return
	}
	writeJSON(w, http.StatusCreated, pl)
}

func (a *apiDeps) patchSceneModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	modelID := r.PathValue("modelId")
	if sceneID == "" || modelID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene or model id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSceneJSONBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "could not read body")
		return
	}
	var body patchSceneModelBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	pl, err := a.store.PatchSceneModelOffsets(r.Context(), sceneID, modelID, body.OffsetX, body.OffsetY, body.OffsetZ)
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found or model not in scene")
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "outside the non-negative") || strings.Contains(msg, "offsets must") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update placement")
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (a *apiDeps) deleteSceneModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	modelID := r.PathValue("modelId")
	if sceneID == "" || modelID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene or model id")
		return
	}
	err := a.store.RemoveSceneModel(r.Context(), sceneID, modelID)
	if errors.Is(err, store.ErrSceneLastModel) {
		writeAPIErrorDetail(w, http.StatusConflict, "scene_last_model",
			"Removing the last model deletes the entire scene. Confirm in the UI, then delete the scene.",
			map[string]string{"scene_id": sceneID})
		return
	}
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not in scene")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not remove model from scene")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
