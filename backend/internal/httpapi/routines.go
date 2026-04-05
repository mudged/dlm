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

const maxRoutineJSONBytes = 1 << 18 // 256 KiB

type createRoutineBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type routineRunsResponse struct {
	Runs []store.RoutineRunDTO `json:"runs"`
}

type startRoutineResponse struct {
	RunID    string `json:"run_id"`
	SceneID  string `json:"scene_id"`
	RoutineID string `json:"routine_id"`
	Status   string `json:"status"`
}

type stopRoutineResponse struct {
	RunID  string `json:"run_id"`
	Status string `json:"status"`
}

func (a *apiDeps) listRoutines(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	list, err := a.store.ListRoutines(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not list routines")
		return
	}
	if list == nil {
		list = []store.RoutineDTO{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *apiDeps) createRoutine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRoutineJSONBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "could not read body")
		return
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var body createRoutineBody
	if err := dec.Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if dec.More() {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	out, err := a.store.CreateRoutine(r.Context(), body.Name, body.Description, body.Type)
	if errors.Is(err, store.ErrRoutineUnknownType) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "name is required") || strings.Contains(msg, "type is required") {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", msg)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not create routine")
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (a *apiDeps) deleteRoutine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing routine id")
		return
	}
	err := a.store.DeleteRoutine(r.Context(), id)
	if errors.Is(err, store.ErrRoutineNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "routine not found")
		return
	}
	if errors.Is(err, store.ErrRoutineRunActive) {
		writeAPIError(w, http.StatusConflict, "routine_run_active", "stop the running instance before deleting this routine")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not delete routine")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiDeps) listSceneRoutineRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	if sceneID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	runs, err := a.store.ListRunningRoutineRunsForScene(r.Context(), sceneID)
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not list routine runs")
		return
	}
	if runs == nil {
		runs = []store.RoutineRunDTO{}
	}
	writeJSON(w, http.StatusOK, routineRunsResponse{Runs: runs})
}

func (a *apiDeps) postSceneRoutineStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	routineID := r.PathValue("routineId")
	if sceneID == "" || routineID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene or routine id")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRoutineJSONBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "could not read body")
		return
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) > 0 && !bytes.Equal(raw, []byte("{}")) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "request body must be empty or {}")
		return
	}

	runID, already, err := a.store.StartRoutineRun(r.Context(), sceneID, routineID)
	if errors.Is(err, store.ErrRoutineNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "routine not found")
		return
	}
	if errors.Is(err, store.ErrSceneNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}
	var conflict *store.SceneRoutineConflictError
	if errors.As(err, &conflict) {
		writeAPIErrorDetail(w, http.StatusConflict, "scene_routine_conflict", err.Error(), map[string]string{
			"run_id":     conflict.RunID,
			"routine_id": conflict.RoutineID,
		})
		return
	}
	if err != nil {
		if isSceneStateValidationError(err) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not start routine")
		return
	}

	resp := startRoutineResponse{
		RunID: runID, SceneID: sceneID, RoutineID: routineID, Status: store.RoutineStatusRunning,
	}
	if already {
		writeJSON(w, http.StatusOK, resp)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (a *apiDeps) postSceneRoutineRunStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	runID := r.PathValue("runId")
	if sceneID == "" || runID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene or run id")
		return
	}
	err := a.store.StopRoutineRun(r.Context(), sceneID, runID)
	if errors.Is(err, store.ErrRoutineRunNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "routine run not found for this scene")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not stop routine")
		return
	}
	writeJSON(w, http.StatusOK, stopRoutineResponse{RunID: runID, Status: store.RoutineStatusStopped})
}
