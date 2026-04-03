package httpapi

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"example.com/dlm/backend/internal/store"
	"example.com/dlm/backend/internal/wiremodel"
)

const maxModelUploadBytes = 1 << 20 // 1 MiB

func (a *apiDeps) listModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	list, err := a.store.List(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not list models")
		return
	}
	if list == nil {
		list = []store.Summary{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *apiDeps) getModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id")
		return
	}
	d, err := a.store.Get(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load model")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (a *apiDeps) deleteModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id")
		return
	}
	if err := a.store.Delete(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	} else if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not delete model")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiDeps) createModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxModelUploadBytes)
	if err := r.ParseMultipartForm(maxModelUploadBytes); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid multipart form or file too large")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "name is required")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "csv file is required under field \"file\"")
		return
	}
	defer func() { _ = file.Close() }()

	body, err := io.ReadAll(file)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "could not read uploaded file")
		return
	}
	lights, err := wiremodel.ParseLightsCSV(bytes.NewReader(body))
	if err != nil {
		var pe *wiremodel.ParseError
		if errors.As(err, &pe) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", pe.Message)
			return
		}
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	sum, err := a.store.Create(r.Context(), name, lights)
	if errors.Is(err, store.ErrDuplicateName) {
		writeAPIError(w, http.StatusConflict, "conflict", "a model with this name already exists")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not save model")
		return
	}
	writeJSON(w, http.StatusCreated, sum)
}
