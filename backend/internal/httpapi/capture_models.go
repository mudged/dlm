package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"example.com/dlm/backend/internal/cvruntime"
	"example.com/dlm/backend/internal/reconstruct"
	"example.com/dlm/backend/internal/store"
	"example.com/dlm/backend/internal/wiremodel"
)

// maxCaptureUploadBytes is the per-request body limit for video uploads.
// Deliberately much larger than the 1 MiB CSV model limit (REQ-048).
// maxCaptureConfirmBodyBytes limits the JSON confirm payload (name only).
// Allowed video container extensions for reconstruction feeds (REQ-048).
const (
	maxCaptureUploadBytes       = 512 << 20 // 512 MiB
	maxCaptureConfirmBodyBytes  = 4096
)

var allowedVideoExts = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".mkv":  true,
	".webm": true,
}

// POST /models/capture
func (a *apiDeps) postModelsCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	if a.reconstruct == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", "reconstruction not available")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCaptureUploadBytes)
	// ParseMultipartForm spills files to temp-disk beyond the memory threshold,
	// keeping per-request heap usage bounded (REQ-048 security note).
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid multipart form or payload too large")
		return
	}

	fhs := r.MultipartForm.File["files"]
	if len(fhs) < 2 {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "at least 2 video files are required")
		return
	}

	var fileReaders []io.Reader
	var fileNames []string
	for _, fh := range fhs {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !allowedVideoExts[ext] {
			writeAPIError(w, http.StatusBadRequest, "bad_request",
				"unsupported video container; allowed extensions: .mp4, .mov, .mkv, .webm")
			return
		}
		f, err := fh.Open()
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not read uploaded file")
			return
		}
		defer func() { _ = f.Close() }()
		fileReaders = append(fileReaders, f)
		fileNames = append(fileNames, fh.Filename)
	}

	params := reconstruct.CreateParams{}
	if m := strings.TrimSpace(r.FormValue("marker")); m == "true" || m == "1" {
		params.Marker = &cvruntime.Marker{Dictionary: "DICT_4X4_50", EdgeLengthM: 0.1}
	}
	if sh := strings.TrimSpace(r.FormValue("scale_hint")); sh != "" {
		v, err := strconv.ParseFloat(sh, 64)
		if err != nil || !(v > 0) || math.IsNaN(v) || math.IsInf(v, 0) {
			writeAPIError(w, http.StatusBadRequest, "bad_request", "scale_hint must be a positive finite number")
			return
		}
		params.ScaleHint = &v
	}

	jobID, err := a.reconstruct.Create(r.Context(), fileReaders, fileNames, params)
	if errors.Is(err, reconstruct.ErrCapExceeded) {
		writeAPIError(w, http.StatusServiceUnavailable, "capacity_exceeded", "a reconstruction job is already in progress; try again later")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"job_id": jobID,
		"status": "pending",
	})
}

// GET /models/capture/{jobId}
func (a *apiDeps) getModelsCaptureJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	if a.reconstruct == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", "reconstruction not available")
		return
	}

	jobID := r.PathValue("jobId")
	job, ok := a.reconstruct.Get(jobID)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}

	resp := map[string]any{
		"status":   job.Status,
		"progress": job.Progress,
	}
	if job.Result != nil {
		resp["result"] = job.Result
	}
	if job.Err != "" {
		resp["error"] = job.Err
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /models/capture/{jobId}/confirm
func (a *apiDeps) postModelsCaptureConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	if a.reconstruct == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", "reconstruction not available")
		return
	}

	jobID := r.PathValue("jobId")

	r.Body = http.MaxBytesReader(w, r.Body, maxCaptureConfirmBodyBytes)
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body or payload too large")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "name is required")
		return
	}

	sum, err := a.reconstruct.Confirm(r.Context(), jobID, name)
	if errors.Is(err, reconstruct.ErrJobNotFound) || errors.Is(err, reconstruct.ErrJobNotSucceeded) {
		writeAPIError(w, http.StatusNotFound, "not_found", "job not found or not in succeeded state")
		return
	}
	var pe *wiremodel.ParseError
	if errors.As(err, &pe) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", pe.Message)
		return
	}
	if errors.Is(err, store.ErrDuplicateName) {
		writeAPIError(w, http.StatusConflict, "conflict", "a model with this name already exists")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not confirm model")
		return
	}

	writeJSON(w, http.StatusCreated, sum)
}

// DELETE /models/capture/{jobId}
func (a *apiDeps) deleteModelsCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	if a.reconstruct == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", "reconstruction not available")
		return
	}

	jobID := r.PathValue("jobId")
	if err := a.reconstruct.Discard(jobID); errors.Is(err, reconstruct.ErrJobNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
