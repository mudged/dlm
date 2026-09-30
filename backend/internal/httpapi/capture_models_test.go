package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
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
	jobs         map[string]*reconstruct.Job
	last         reconstruct.CreateParams
	createCalled bool
	next         int
}

func newFakeReconstructCtrl() *fakeReconstructCtrl {
	return &fakeReconstructCtrl{jobs: make(map[string]*reconstruct.Job)}
}

func (f *fakeReconstructCtrl) Create(_ context.Context, files []io.Reader, _ []string, p reconstruct.CreateParams) (string, error) {
	f.createCalled = true
	f.last = p
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

func (f *fakeReconstructCtrl) Shutdown() {}

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

// multipartBody builds a multipart/form-data body with one or more "files" fields
// and optional non-file form fields.
func multipartBody(t *testing.T, files map[string]string, fields ...string) (body *bytes.Buffer, contentType string) {
	t.Helper()
	if len(fields)%2 != 0 {
		t.Fatal("fields must be key/value pairs")
	}
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
	for i := 0; i < len(fields); i += 2 {
		if err := w.WriteField(fields[i], fields[i+1]); err != nil {
			t.Fatal(err)
		}
	}
	_ = w.Close()
	return body, w.FormDataContentType()
}

func TestAPIv1CaptureModels_postWithMarkerAndScaleHint_forwardsCreateParams(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	}, "marker", "true", "scale_hint", "1.5")
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}

	if fake.last.Marker == nil {
		t.Fatal("expected Marker to be set")
	}
	if fake.last.Marker.Dictionary != "DICT_4X4_50" {
		t.Fatalf("Marker.Dictionary = %q, want DICT_4X4_50", fake.last.Marker.Dictionary)
	}
	if fake.last.Marker.EdgeLengthM != 0.1 {
		t.Fatalf("Marker.EdgeLengthM = %v, want 0.1", fake.last.Marker.EdgeLengthM)
	}
	if len(fake.last.Marker.IDs) != 3 || fake.last.Marker.IDs[0] != 0 || fake.last.Marker.IDs[1] != 1 || fake.last.Marker.IDs[2] != 2 {
		t.Fatalf("Marker.IDs = %v, want [0 1 2]", fake.last.Marker.IDs)
	}
	if fake.last.ScaleHint == nil {
		t.Fatal("expected ScaleHint to be set")
	}
	if *fake.last.ScaleHint != 1.5 {
		t.Fatalf("ScaleHint = %v, want 1.5", *fake.last.ScaleHint)
	}
}

func TestAPIv1CaptureModels_postWithInvalidScaleHint_returns400(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	}, "scale_hint", "not-a-number")
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want 400, body = %s", res.StatusCode, b)
	}
	if fake.last.ScaleHint != nil {
		t.Fatal("Create should not be called with invalid scale_hint")
	}
}

func TestAPIv1CaptureModels_postWithLightCount_forwardsCreateParams(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	}, "light_count", "12")
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}
	if fake.last.LightCount == nil {
		t.Fatal("expected LightCount to be set")
	}
	if *fake.last.LightCount != 12 {
		t.Fatalf("LightCount = %v, want 12", *fake.last.LightCount)
	}
}

func TestAPIv1CaptureModels_postWithEmptyLightCount_omitsCreateParam(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	}, "light_count", "")
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, body = %s", res.StatusCode, b)
	}
	if fake.last.LightCount != nil {
		t.Fatalf("LightCount = %v, want nil", *fake.last.LightCount)
	}
}

func TestAPIv1CaptureModels_postWithInvalidLightCount_returns400(t *testing.T) {
	const wantMsg = "light_count must be a whole number from 1 to 1000"
	for _, val := range []string{"0", "nope", "1001"} {
		t.Run(val, func(t *testing.T) {
			srv, fake := newCaptureModelsTestServer(t)

			body, ct := multipartBody(t, map[string]string{
				"feed_a.mp4": "fake-video-a",
				"feed_b.mp4": "fake-video-b",
			}, "light_count", val)
			res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusBadRequest {
				b, _ := io.ReadAll(res.Body)
				t.Fatalf("status = %d, want 400, body = %s", res.StatusCode, b)
			}
			var env struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if env.Error.Code != "bad_request" {
				t.Fatalf("code = %q, want bad_request", env.Error.Code)
			}
			if env.Error.Message != wantMsg {
				t.Fatalf("message = %q, want %q", env.Error.Message, wantMsg)
			}
			if fake.createCalled {
				t.Fatal("Create should not be called with invalid light_count")
			}
		})
	}
}

func TestReconstructCreate_copiesLightCountOntoJobSpec(t *testing.T) {
	runner := &specCaptureRunner{done: make(chan struct{})}
	m := reconstruct.New(runner, testStore(t), t.TempDir())
	n := 12
	_, err := m.Create(context.Background(), []io.Reader{
		strings.NewReader("v1"),
		strings.NewReader("v2"),
	}, []string{"a.mp4", "b.mp4"}, reconstruct.CreateParams{LightCount: &n})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	select {
	case <-runner.done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for CV runner")
	}
	if runner.got.LightCount == nil {
		t.Fatal("expected JobSpec.LightCount to be set")
	}
	if *runner.got.LightCount != 12 {
		t.Fatalf("JobSpec.LightCount = %v, want 12", *runner.got.LightCount)
	}
}

type specCaptureRunner struct {
	got  cvruntime.JobSpec
	done chan struct{}
}

func (s *specCaptureRunner) Run(_ context.Context, spec cvruntime.JobSpec) (cvruntime.Result, error) {
	s.got = spec
	close(s.done)
	return cvruntime.Result{Status: "succeeded"}, nil
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

func TestAPIv1CaptureModels_getStatus_includesRejectedFeeds(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	fake.jobs["known-job"] = &reconstruct.Job{
		ID:       "known-job",
		Status:   reconstruct.StatusSucceeded,
		Progress: 1.0,
		Result: &cvruntime.Result{
			Status: "succeeded",
			RejectedFeeds: []cvruntime.RejectedFeed{
				{File: "side.mp4", Reason: "no start or end signal found"},
			},
		},
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
	var resp struct {
		Result struct {
			RejectedFeeds []struct {
				File   string `json:"file"`
				Reason string `json:"reason"`
			} `json:"rejected_feeds"`
		} `json:"result"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Result.RejectedFeeds) != 1 {
		t.Fatalf("rejected_feeds = %#v, want 1 entry", resp.Result.RejectedFeeds)
	}
	if resp.Result.RejectedFeeds[0].File != "side.mp4" {
		t.Fatalf("file = %q, want side.mp4", resp.Result.RejectedFeeds[0].File)
	}
	if resp.Result.RejectedFeeds[0].Reason != "no start or end signal found" {
		t.Fatalf("reason = %q", resp.Result.RejectedFeeds[0].Reason)
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

func TestAPIv1CaptureModels_confirmOversizedBody_returns400(t *testing.T) {
	srv, fake := newCaptureModelsTestServer(t)

	fake.jobs["job-to-confirm"] = &reconstruct.Job{
		ID:     "job-to-confirm",
		Status: reconstruct.StatusSucceeded,
		Result: &cvruntime.Result{Status: "succeeded"},
	}

	oversizedName := strings.Repeat("x", maxCaptureConfirmBodyBytes)
	body := strings.NewReader(`{"name":"` + oversizedName + `"}`)
	res, err := http.Post(srv.URL+"/api/v1/models/capture/job-to-confirm/confirm", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want 400, body = %s", res.StatusCode, b)
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
func (dupNameFake) Shutdown()              {}

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

func TestAPIv1CaptureModels_badMultipart_explainsUnreadableUpload(t *testing.T) {
	srv, _ := newCaptureModelsTestServer(t)

	res, err := http.Post(srv.URL+"/api/v1/models/capture", "multipart/form-data; boundary=not-a-real-boundary", strings.NewReader("this is not a multipart body"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
	msg := apiErrorMessage(t, res)
	if !strings.Contains(msg, "could not be read") {
		t.Fatalf("message = %q, want an explanation that the upload could not be read", msg)
	}
}

func TestAPIv1CaptureModels_payloadTooLarge_explainsLimit(t *testing.T) {
	orig := captureUploadMaxBytes
	captureUploadMaxBytes = 64
	t.Cleanup(func() { captureUploadMaxBytes = orig })

	srv, _ := newCaptureModelsTestServer(t)
	body := strings.NewReader(strings.Repeat("x", 128))
	res, err := http.Post(srv.URL+"/api/v1/models/capture", "multipart/form-data; boundary=x", body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
	msg := apiErrorMessage(t, res)
	if !strings.Contains(msg, "2 GB") {
		t.Fatalf("message = %q, want the 2 GB limit", msg)
	}
}

func TestClassifyCaptureUploadError_timeout(t *testing.T) {
	_, code, msg := classifyCaptureUploadError(os.ErrDeadlineExceeded)
	if code != "upload_timeout" {
		t.Fatalf("code = %q, want upload_timeout", code)
	}
	if !strings.Contains(msg, "interrupted") {
		t.Fatalf("message = %q, want an interrupted-upload explanation", msg)
	}
}

func TestAPIv1CaptureModels_storageFailure_explainsDiskAndLogsCause(t *testing.T) {
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
		DataDir:      t.TempDir(),
	}
	capLog := &captureHandler{}
	log := slog.New(capLog)
	deps := &apiDeps{
		store:       testStore(t),
		rev:         NewRevisionHubWithLogger(log),
		reconstruct: storageFailFake{},
	}
	srv := httptest.NewServer(buildSiteHandler(cfg, nil, deps, log))
	t.Cleanup(srv.Close)

	body, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	})
	res, err := http.Post(srv.URL+"/api/v1/models/capture", ct, body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want 400, body = %s", res.StatusCode, b)
	}
	msg := apiErrorMessage(t, res)
	if !strings.Contains(msg, "could not store the videos") {
		t.Fatalf("message = %q, want a disk-storage explanation", msg)
	}
	if strings.Contains(msg, "no space left") {
		t.Fatalf("message = %q, want the raw disk error kept out of the response", msg)
	}
	if !logHasAttr(capLog.snapshot(), "capture upload failed", "err", "no space left on device") {
		t.Fatalf("expected a capture upload log containing the disk error, got %s", formatLogs(capLog.snapshot()))
	}
}

func TestAPIv1CaptureModels_badMultipart_logsUnderlyingError(t *testing.T) {
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
		DataDir:      t.TempDir(),
	}
	capLog := &captureHandler{}
	log := slog.New(capLog)
	deps := &apiDeps{
		store:       testStore(t),
		rev:         NewRevisionHubWithLogger(log),
		reconstruct: newFakeReconstructCtrl(),
	}
	srv := httptest.NewServer(buildSiteHandler(cfg, nil, deps, log))
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/api/v1/models/capture", "multipart/form-data; boundary=not-a-real-boundary", strings.NewReader("this is not a multipart body"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if !logHasAttr(capLog.snapshot(), "capture upload rejected", "err", "") {
		t.Fatalf("expected a capture upload rejection log with err, got %s", formatLogs(capLog.snapshot()))
	}
}

func TestAPIv1CaptureModels_slowUploadPastServerReadTimeout_returns202(t *testing.T) {
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
		DataDir:      t.TempDir(),
	}
	log := noopLogger()
	deps := &apiDeps{
		store:       testStore(t),
		rev:         NewRevisionHubWithLogger(log),
		reconstruct: newFakeReconstructCtrl(),
	}
	h := buildSiteHandler(cfg, nil, deps, log)
	srv := httptest.NewUnstartedServer(h)
	srv.Config.ReadHeaderTimeout = time.Second
	srv.Config.ReadTimeout = 200 * time.Millisecond
	srv.Config.WriteTimeout = 200 * time.Millisecond
	srv.Start()
	t.Cleanup(srv.Close)

	payload, ct := multipartBody(t, map[string]string{
		"feed_a.mp4": "fake-video-a",
		"feed_b.mp4": "fake-video-b",
	})
	pr, pw := io.Pipe()
	go func() {
		data := payload.Bytes()
		mid := len(data) / 2
		if _, err := pw.Write(data[:mid]); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		time.Sleep(500 * time.Millisecond)
		if _, err := pw.Write(data[mid:]); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = pw.Close()
	}()

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/models/capture", pr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", ct)
	req.ContentLength = int64(payload.Len())
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("slow upload: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want 202, body = %s", res.StatusCode, b)
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

func apiErrorMessage(t *testing.T, res *http.Response) string {
	t.Helper()
	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body.Error.Message
}

type storageFailFake struct{}

func (storageFailFake) Create(context.Context, []io.Reader, []string, reconstruct.CreateParams) (string, error) {
	return "", fmt.Errorf("reconstruct: write feed file: %w", errors.New("no space left on device"))
}
func (storageFailFake) Get(string) (*reconstruct.Job, bool) { return nil, false }
func (storageFailFake) Confirm(context.Context, string, string) (store.Summary, error) {
	return store.Summary{}, reconstruct.ErrJobNotFound
}
func (storageFailFake) Discard(string) error { return nil }
func (storageFailFake) Shutdown()            {}

func logHasAttr(recs []slog.Record, message, key, substr string) bool {
	for _, r := range recs {
		if r.Message != message {
			continue
		}
		found := false
		r.Attrs(func(a slog.Attr) bool {
			if a.Key != key {
				return true
			}
			val := a.Value.String()
			if substr == "" {
				found = val != ""
			} else if strings.Contains(val, substr) {
				found = true
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}

func formatLogs(recs []slog.Record) string {
	var buf bytes.Buffer
	for _, r := range recs {
		buf.WriteString(r.Level.String())
		buf.WriteByte(' ')
		buf.WriteString(r.Message)
		r.Attrs(func(a slog.Attr) bool {
			buf.WriteByte(' ')
			buf.WriteString(a.Key)
			buf.WriteByte('=')
			buf.WriteString(a.Value.String())
			return true
		})
		buf.WriteByte('\n')
	}
	return buf.String()
}
