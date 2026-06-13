package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/cvruntime"
	"example.com/dlm/backend/internal/reconstruct"
	"example.com/dlm/backend/internal/store"
)

// fakeReconstructCtrl is a minimal in-process reconstructCtrl for HTTP tests.
type fakeReconstructCtrl struct {
	jobs map[string]*reconstruct.Job
	next int
}

func newFakeReconstructCtrl() *fakeReconstructCtrl {
	return &fakeReconstructCtrl{jobs: make(map[string]*reconstruct.Job)}
}

func (f *fakeReconstructCtrl) Create(_ context.Context, files []io.Reader, _ []string, _ reconstruct.CreateParams) (string, error) {
	if len(files) < 2 {
		return "", errors.New("at least 2 video files are required")
	}
	id := "test-job-id"
	f.jobs[id] = &reconstruct.Job{
		ID:        id,
		Status:    reconstruct.StatusSucceeded,
		Progress:  1.0,
		CreatedAt: time.Now(),
		Result: &cvruntime.Result{
			Status:     "succeeded",
			LightCount: 2,
			Lights: []cvruntime.LightPoint{
				{ID: 0, X: 0, Y: 0, Z: 0},
				{ID: 1, X: 1, Y: 1, Z: 1},
			},
		},
	}
	return id, nil
}

func (f *fakeReconstructCtrl) Get(id string) (*reconstruct.Job, bool) {
	j, ok := f.jobs[id]
	return j, ok
}

func (f *fakeReconstructCtrl) Confirm(_ context.Context, jobID, name string) (store.Summary, error) {
	j, ok := f.jobs[jobID]
	if !ok {
		return store.Summary{}, reconstruct.ErrJobNotFound
	}
	if j.Status != reconstruct.StatusSucceeded {
		return store.Summary{}, reconstruct.ErrJobNotSucceeded
	}
	return store.Summary{
		ID:         "model-123",
		Name:       name,
		LightCount: 2,
		CreatedAt:  time.Now(),
	}, nil
}

func (f *fakeReconstructCtrl) Discard(jobID string) error {
	if _, ok := f.jobs[jobID]; !ok {
		return reconstruct.ErrJobNotFound
	}
	delete(f.jobs, jobID)
	return nil
}

func newCaptureModelsTestServer(t *testing.T) (*httptest.Server, *fakeReconstructCtrl) {
	t.Helper()
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
		DataDir:      t.TempDir(),
	}
	fake := newFakeReconstructCtrl()
	log := noopLogger()
	st := testStore(t)
	deps := &apiDeps{
		store:       st,
		rev:         NewRevisionHubWithLogger(log),
		reconstruct: fake,
	}
	h := buildSiteHandler(cfg, nil, deps, log)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, fake
}

// multipartBody builds a multipart/form-data body with one or more "files" fields.
func multipartBody(t *testing.T, files map[string]string) (body *bytes.Buffer, contentType string) {
	t.Helper()
	body = &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for name, content := range files {
		fw, err := w.CreateFormFile("files", name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(fw, content); err != nil {
			t.Fatal(err)
		}
	}
	_ = w.Close()
	return body, w.FormDataContentType()
}

func TestAPIv1CaptureModels_postWith2Files_returns202(t *testing.T) {
	srv, _ := newCaptureModelsTestServer(t)

	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	})
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}
	var resp struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.JobID == "" {
		t.Fatal("expected non-empty job_id")
	}
	if resp.Status != "pending" {
		t.Fatalf("status = %q, want pending", resp.Status)
	}
}

func TestAPIv1CaptureModels_postWith1File_returns400(t *testing.T) {
	srv, _ := newCaptureModelsTestServer(t)

	body, ct := multipartBody(t, map[string]string{"only.mp4": "data"})
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestAPIv1CaptureModels_postUnsupportedExtension_returns400(t *testing.T) {
	srv, _ := newCaptureModelsTestServer(t)

	body, ct := multipartBody(t, map[string]string{
		"feed_a.avi": "data",
		"feed_b.mp4": "data",
	})
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestAPIv1CaptureModels_getStatus_returns200(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	// Seed a job directly.
	fake.jobs["known-job"] = &reconstruct.Job{
		ID:       "known-job",
		Status:   reconstruct.StatusRunning,
		Progress: 0.5,
	}

	res, err := http.Get(srv.URL + "/api/v1/models/capture/known-job")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}
	var resp map[string]any
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "running" {
		t.Fatalf("status = %v, want running", resp["status"])
	}
}

func TestAPIv1CaptureModels_getUnknownJob_returns404(t *testing.T) {
	srv, _ := newCaptureModelsTestServer(t)

	res, err := http.Get(srv.URL + "/api/v1/models/capture/no-such-job")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

func TestAPIv1CaptureModels_confirm_returns201(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	fake.jobs["job-to-confirm"] = &reconstruct.Job{
		ID:     "job-to-confirm",
		Status: reconstruct.StatusSucceeded,
		Result: &cvruntime.Result{Status: "succeeded"},
	}

	body := strings.NewReader(`{"name":"my-reconstruction"}`)
	res, err := http.Post(srv.URL+"/api/v1/models/capture/job-to-confirm/confirm", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}
	var sum store.Summary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if sum.Name != "my-reconstruction" {
		t.Fatalf("name = %q, want my-reconstruction", sum.Name)
	}
}

func TestAPIv1CaptureModels_confirmDuplicateName_returns409(t *testing.T) {
	// Create a server with a fake that returns ErrDuplicateName.
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
		DataDir:      t.TempDir(),
	}
	dupFake := &dupNameFake{}
	log := noopLogger()
	deps := &apiDeps{
		store:       testStore(t),
		rev:         NewRevisionHubWithLogger(log),
		reconstruct: dupFake,
	}
	srv2 := httptest.NewServer(buildSiteHandler(cfg, nil, deps, log))
	t.Cleanup(srv2.Close)

	body := strings.NewReader(`{"name":"taken"}`)
	res, err := http.Post(srv2.URL+"/api/v1/models/capture/any-job/confirm", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want 409, body = %s", res.StatusCode, b)
	}
}

// dupNameFake is a reconstructCtrl whose Confirm always returns ErrDuplicateName.
type dupNameFake struct{}

func (dupNameFake) Create(_ context.Context, _ []io.Reader, _ []string, _ reconstruct.CreateParams) (string, error) {
	return "job", nil
}
func (dupNameFake) Get(id string) (*reconstruct.Job, bool) {
	return &reconstruct.Job{ID: id, Status: reconstruct.StatusSucceeded}, true
}
func (dupNameFake) Confirm(_ context.Context, _, _ string) (store.Summary, error) {
	return store.Summary{}, store.ErrDuplicateName
}
func (dupNameFake) Discard(_ string) error { return nil }

func TestAPIv1CaptureModels_confirmUnknownJob_returns404(t *testing.T) {
	srv, _ := newCaptureModelsTestServer(t)

	body := strings.NewReader(`{"name":"x"}`)
	res, err := http.Post(srv.URL+"/api/v1/models/capture/no-such-job/confirm", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

func TestAPIv1CaptureModels_delete_returns204(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	fake.jobs["to-delete"] = &reconstruct.Job{ID: "to-delete", Status: reconstruct.StatusSucceeded}

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/models/capture/to-delete", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want 204, body = %s", res.StatusCode, b)
	}
}

func TestAPIv1CaptureModels_deleteUnknownJob_returns404(t *testing.T) {
	srv, _ := newCaptureModelsTestServer(t)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/models/capture/no-such-job", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}
