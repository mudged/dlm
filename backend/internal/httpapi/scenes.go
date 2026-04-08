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

const maxSceneJSONBytes = 1 << 18      // 256 KiB
const maxSceneLightsBatchBytes = 4 << 20 // 4 MiB (large per-light batch payloads)

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

type patchSceneBody struct {
	BoundaryMarginM float64 `json:"boundary_margin_m"`
}

type sceneCuboidBody struct {
	Position   store.ScenePoint          `json:"position"`
	Dimensions store.SceneDimensionsSize `json:"dimensions"`
}

type sceneSphereBody struct {
	Center store.ScenePoint `json:"center"`
	Radius float64          `json:"radius"`
}

type sceneCuboidPatchBody struct {
	Position      store.ScenePoint          `json:"position"`
	Dimensions    store.SceneDimensionsSize `json:"dimensions"`
	On            *bool                     `json:"on"`
	Color         *string                   `json:"color"`
	BrightnessPct *float64                  `json:"brightness_pct"`
}

type sceneSpherePatchBody struct {
	Center        store.ScenePoint `json:"center"`
	Radius        float64          `json:"radius"`
	On            *bool            `json:"on"`
	Color         *string          `json:"color"`
	BrightnessPct *float64         `json:"brightness_pct"`
}

type sceneWholePatchBody struct {
	On            *bool    `json:"on"`
	Color         *string  `json:"color"`
	BrightnessPct *float64 `json:"brightness_pct"`
}

type sceneLightsBatchPatchBody struct {
	Updates []sceneBatchUpdateItem `json:"updates"`
}

type sceneBatchUpdateItem struct {
	ModelID       string   `json:"model_id"`
	LightID       int      `json:"light_id"`
	On            *bool    `json:"on"`
	Color         *string  `json:"color"`
	BrightnessPct *float64 `json:"brightness_pct"`
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

func (a *apiDeps) patchScene(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
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
	var body patchSceneBody
	if err := dec.Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if dec.More() {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if err := a.store.PatchSceneBoundaryMarginM(r.Context(), id, body.BoundaryMarginM); errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	} else if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "boundary_margin_m") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update scene")
		return
	}
	d, err := a.store.GetScene(r.Context(), id)
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

func (a *apiDeps) getSceneDimensions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	if sceneID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	dims, err := a.store.GetSceneDimensions(r.Context(), sceneID)
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load scene dimensions")
		return
	}
	writeJSON(w, http.StatusOK, dims)
}

func (a *apiDeps) listSceneLights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	if sceneID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	lights, err := a.store.ListSceneLights(r.Context(), sceneID)
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load scene lights")
		return
	}
	if lights == nil {
		lights = []store.SceneLightFlat{}
	}
	writeJSON(w, http.StatusOK, lights)
}

func (a *apiDeps) postSceneLightsQueryCuboid(w http.ResponseWriter, r *http.Request) {
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
	var body sceneCuboidBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	lights, err := a.store.QuerySceneLightsCuboid(r.Context(), sceneID, store.SceneCuboid{
		Position:   body.Position,
		Dimensions: body.Dimensions,
	})
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if errors.Is(err, store.ErrSceneInvalidGeometry) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not query scene lights")
		return
	}
	if lights == nil {
		lights = []store.SceneLightFlat{}
	}
	writeJSON(w, http.StatusOK, lights)
}

func (a *apiDeps) postSceneLightsQuerySphere(w http.ResponseWriter, r *http.Request) {
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
	var body sceneSphereBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	lights, err := a.store.QuerySceneLightsSphere(r.Context(), sceneID, store.SceneSphere{
		Center: body.Center,
		Radius: body.Radius,
	})
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if errors.Is(err, store.ErrSceneInvalidGeometry) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not query scene lights")
		return
	}
	if lights == nil {
		lights = []store.SceneLightFlat{}
	}
	writeJSON(w, http.StatusOK, lights)
}

func isSceneStateValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "at least one of on, color, brightness_pct") ||
		strings.Contains(msg, "color must") ||
		strings.Contains(msg, "brightness_pct")
}

func (a *apiDeps) patchSceneLightsStateCuboid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
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
	var body sceneCuboidPatchBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	out, err := a.store.PatchSceneLightsCuboid(r.Context(), sceneID, store.SceneCuboid{
		Position:   body.Position,
		Dimensions: body.Dimensions,
	}, store.LightStatePatch{
		On:            body.On,
		Color:         body.Color,
		BrightnessPct: body.BrightnessPct,
	})
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if errors.Is(err, store.ErrSceneInvalidGeometry) || isSceneStateValidationError(err) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update scene lights")
		return
	}
	if out.States == nil {
		out.States = []store.ScenePatchedState{}
	}
	if out.UpdatedCount > 0 {
		a.rev.NotifyAfterSceneLightPatch(r.Context(), a.store, out.States)
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *apiDeps) patchSceneLightsStateSphere(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
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
	var body sceneSpherePatchBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	out, err := a.store.PatchSceneLightsSphere(r.Context(), sceneID, store.SceneSphere{
		Center: body.Center,
		Radius: body.Radius,
	}, store.LightStatePatch{
		On:            body.On,
		Color:         body.Color,
		BrightnessPct: body.BrightnessPct,
	})
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if errors.Is(err, store.ErrSceneInvalidGeometry) || isSceneStateValidationError(err) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update scene lights")
		return
	}
	if out.States == nil {
		out.States = []store.ScenePatchedState{}
	}
	if out.UpdatedCount > 0 {
		a.rev.NotifyAfterSceneLightPatch(r.Context(), a.store, out.States)
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *apiDeps) patchSceneLightsStateScene(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
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
	var body sceneWholePatchBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	out, err := a.store.PatchSceneLightsScene(r.Context(), sceneID, store.LightStatePatch{
		On:            body.On,
		Color:         body.Color,
		BrightnessPct: body.BrightnessPct,
	})
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if isSceneStateValidationError(err) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update scene lights")
		return
	}
	if out.States == nil {
		out.States = []store.ScenePatchedState{}
	}
	if out.UpdatedCount > 0 {
		a.rev.NotifyAfterSceneLightPatch(r.Context(), a.store, out.States)
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *apiDeps) patchSceneLightsStateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	if sceneID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSceneLightsBatchBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "could not read body")
		return
	}
	var body sceneLightsBatchPatchBody
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	updates := make([]store.SceneBatchLightUpdate, 0, len(body.Updates))
	for _, u := range body.Updates {
		updates = append(updates, store.SceneBatchLightUpdate{
			ModelID: strings.TrimSpace(u.ModelID),
			LightID: u.LightID,
			Patch: store.LightStatePatch{
				On:            u.On,
				Color:         u.Color,
				BrightnessPct: u.BrightnessPct,
			},
		})
	}
	out, err := a.store.PatchSceneLightsBatch(r.Context(), sceneID, updates)
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if errors.Is(err, store.ErrSceneLightNotInScene) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "one or more lights are not in this scene")
		return
	}
	if isSceneStateValidationError(err) || (err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not update scene lights")
		return
	}
	if out.States == nil {
		out.States = []store.ScenePatchedState{}
	}
	if out.UpdatedCount > 0 {
		a.rev.NotifyAfterSceneLightPatch(r.Context(), a.store, out.States)
	}
	writeJSON(w, http.StatusOK, out)
}
