package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	maxCaptureUploadBytes      = 2 << 30 // 2 GiB combined
	maxCaptureConfirmBodyBytes = 4096
	// captureUploadDeadline is how long a video upload may take to arrive and
	// be stored. It replaces the server-wide read/write timeouts for this
	// route only, so two clips can finish over a slow link.
	captureUploadDeadline = 15 * time.Minute
)

// captureUploadMaxBytes is the enforced body limit. Tests lower it to
// exercise the too-large path without sending a multi-gigabyte body.
var captureUploadMaxBytes int64 = maxCaptureUploadBytes

func classifyCaptureUploadError(err error) (status int, code, message string) {
	var maxBytes *http.MaxBytesError
	if errors.As(err, &maxBytes) || strings.Contains(err.Error(), "request body too large") {
		return http.StatusBadRequest, "payload_too_large", captureUploadLimitMessage()
	}
	if isUploadTimeout(err) {
		return http.StatusBadRequest, "upload_timeout",
			"The upload was interrupted before the server finished receiving the videos."
	}
	return http.StatusBadRequest, "bad_request",
		"The upload could not be read. Check that the files are videos and try again."
}

func captureUploadLimitMessage() string {
	const gib int64 = 1 << 30
	if maxCaptureUploadBytes%gib == 0 {
		return fmt.Sprintf("The videos are larger than %d GB combined. Use shorter clips.", maxCaptureUploadBytes/gib)
	}
	mb := maxCaptureUploadBytes >> 20
	return fmt.Sprintf("The videos are larger than %d MB combined. Use shorter clips.", mb)
}

func isUploadTimeout(err error) bool {
	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

func isCaptureStorageError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "reconstruct: create work dir") ||
		strings.Contains(msg, "reconstruct: create feed file") ||
		strings.Contains(msg, "reconstruct: write feed file") ||
		strings.Contains(msg, "reconstruct: close feed file")
}

func extendCaptureUploadDeadline(w http.ResponseWriter) {
	deadline := time.Now().Add(captureUploadDeadline)
	rc := http.NewResponseController(w)
	_ = rc.SetReadDeadline(deadline)
	_ = rc.SetWriteDeadline(deadline)
}

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

	extendCaptureUploadDeadline(w)
	log := requestLogger(r)

	// Reject from the Content-Length header before reading the body. A browser
	// that is still uploading does not reliably surface a response written
	// halfway through the body, so the page also checks the size first.
	if r.ContentLength > captureUploadMaxBytes {
		log.Warn("capture upload rejected",
			"err", "content length exceeds upload limit",
			"code", "payload_too_large",
			"content_length", r.ContentLength,
		)
		writeAPIError(w, http.StatusBadRequest, "payload_too_large", captureUploadLimitMessage())
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, captureUploadMaxBytes)
	// ParseMultipartForm spills files to temp-disk beyond the memory threshold,
	// keeping per-request heap usage bounded (REQ-048 security note).
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		status, code, msg := classifyCaptureUploadError(err)
		log.Warn("capture upload rejected",
			"err", err,
			"code", code,
			"content_length", r.ContentLength,
		)
		writeAPIError(w, status, code, msg)
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
			log.Warn("capture upload rejected",
				"code", "bad_request",
				"file", fh.Filename,
				"content_length", r.ContentLength,
			)
			writeAPIError(w, http.StatusBadRequest, "bad_request",
				"unsupported video container; allowed extensions: .mp4, .mov, .mkv, .webm")
			return
		}
		f, err := fh.Open()
		if err != nil {
			log.Error("capture upload file open failed",
				"err", err,
				"file", fh.Filename,
				"content_length", r.ContentLength,
			)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not read uploaded file")
			return
		}
		defer func() { _ = f.Close() }()
		fileReaders = append(fileReaders, f)
		fileNames = append(fileNames, fh.Filename)
	}

	params := reconstruct.CreateParams{}
	if m := strings.TrimSpace(r.FormValue("marker")); m == "true" || m == "1" {
		params.Marker = &cvruntime.Marker{
			Dictionary:  "DICT_4X4_50",
			EdgeLengthM: 0.1,
			IDs:         []int{0, 1, 2},
		}
	}
	if sh := strings.TrimSpace(r.FormValue("scale_hint")); sh != "" {
		v, err := strconv.ParseFloat(sh, 64)
		if err != nil || !(v > 0) || math.IsNaN(v) || math.IsInf(v, 0) {
			writeAPIError(w, http.StatusBadRequest, "bad_request", "scale_hint must be a positive finite number")
			return
		}
		params.ScaleHint = &v
	}
	if lc := strings.TrimSpace(r.FormValue("light_count")); lc != "" {
		v, err := strconv.Atoi(lc)
		if err != nil || v < 1 || v > 1000 {
			writeAPIError(w, http.StatusBadRequest, "bad_request", "light_count must be a whole number from 1 to 1000")
			return
		}
		params.LightCount = &v
	}

	jobID, err := a.reconstruct.Create(r.Context(), fileReaders, fileNames, params)
	if errors.Is(err, reconstruct.ErrCapExceeded) {
		log.Warn("capture upload rejected",
			"err", err,
			"code", "capacity_exceeded",
			"files", fileNames,
			"content_length", r.ContentLength,
		)
		writeAPIError(w, http.StatusServiceUnavailable, "capacity_exceeded", "a reconstruction job is already in progress; try again later")
		return
	}
	if err != nil {
		log.Error("capture upload failed",
			"err", err,
			"files", fileNames,
			"content_length", r.ContentLength,
		)
		msg := err.Error()
		if isCaptureStorageError(err) {
			msg = "The server could not store the videos. Check that the disk is not full."
		}
		writeAPIError(w, http.StatusBadRequest, "bad_request", msg)
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
