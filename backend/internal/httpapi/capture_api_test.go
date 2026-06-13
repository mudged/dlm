package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/dlm/backend/internal/capture"
	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/lightstate"
	"example.com/dlm/backend/internal/store"
)

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// noopDriver satisfies the capture driver interface without sending HTTP requests.
type noopDriver struct{}

func (n *noopDriver) DriveSingleLED(_ context.Context, _ store.Device, _, _ int) error {
	return nil
}
func (n *noopDriver) DriveAllOff(_ context.Context, _ store.Device, _ int) error { return nil }

// newCaptureTestServer creates an httptest.Server with a real capture controller
// backed by noopDriver so no WLED network calls happen.
func newCaptureTestServer(t *testing.T, st *store.Store) *httptest.Server {
	t.Helper()
	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       filepath.Join(t.TempDir(), "unused.db"),
	}
	ctrl := capture.New(st, &noopDriver{}, nil, &capture.ControllerOpts{Dwell: 20 * time.Millisecond})
	t.Cleanup(ctrl.Shutdown)

	log := noopLogger()
	deps := &apiDeps{
		store:   st,
		rev:     NewRevisionHubWithLogger(log),
		// *capture.Controller satisfies captureCtrl directly.
		capture: ctrl,
	}
	h := buildSiteHandler(cfg, nil, deps, log)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func newCaptureStore(t *testing.T) *store.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "cap.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	st.SetLightState(lightstate.New())
	if err := st.LoadLightStateFromDB(context.Background()); err != nil {
		t.Fatal(err)
	}
	return st
}

func TestAPIv1Capture_startStatusStop(t *testing.T) {
	st := newCaptureStore(t)
	d, err := st.CreateDevice(context.Background(), store.DeviceCreate{
		Name:       "test-strip",
		BaseURL:    "http://wled.test",
		LightCount: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	srv := newCaptureTestServer(t, st)
	base := srv.URL + "/api/v1/devices/" + d.ID + "/capture"

	// POST start → 200 running
	res, err := http.Post(base+"/start", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("start status=%d body=%s", res.StatusCode, b)
	}
	var startBody map[string]any
	if err := json.Unmarshal(b, &startBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if startBody["state"] != "running" {
		t.Fatalf("state = %v want running", startBody["state"])
	}
	if startBody["device_id"] != d.ID {
		t.Fatalf("device_id = %v want %s", startBody["device_id"], d.ID)
	}
	if lc, _ := startBody["light_count"].(float64); lc != 100 {
		t.Fatalf("light_count = %v want 100", startBody["light_count"])
	}

	// GET status → 200 with state field
	res2, err := http.Get(base)
	if err != nil {
		t.Fatal(err)
	}
	b2, _ := io.ReadAll(res2.Body)
	_ = res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("status status=%d body=%s", res2.StatusCode, b2)
	}
	var statusBody map[string]any
	if err := json.Unmarshal(b2, &statusBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := statusBody["state"]; !ok {
		t.Fatalf("missing state in %s", b2)
	}

	// POST stop mid-sweep → 200 stopping (sweep still turning LEDs off)
	res3, err := http.Post(base+"/stop", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	b3, _ := io.ReadAll(res3.Body)
	_ = res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("stop status=%d body=%s", res3.StatusCode, b3)
	}
	var stopBody map[string]any
	if err := json.Unmarshal(b3, &stopBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stopBody["state"] != "stopping" {
		t.Fatalf("stop state = %v want stopping", stopBody["state"])
	}
	if stopBody["device_id"] != d.ID {
		t.Fatalf("device_id = %v want %s", stopBody["device_id"], d.ID)
	}
}

func TestAPIv1Capture_unknownDevice_returns404(t *testing.T) {
	st := newCaptureStore(t)
	srv := newCaptureTestServer(t, st)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/devices/no-such/capture/start"},
		{http.MethodPost, "/api/v1/devices/no-such/capture/stop"},
		{http.MethodGet, "/api/v1/devices/no-such/capture"},
	}
	for _, tc := range cases {
		req, _ := http.NewRequest(tc.method, srv.URL+tc.path, strings.NewReader(""))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("method=%s path=%s status=%d want 404", tc.method, tc.path, res.StatusCode)
		}
	}
}

func TestAPIv1Capture_zeroLightCount_returns422(t *testing.T) {
	st := newCaptureStore(t)
	d, err := st.CreateDevice(context.Background(), store.DeviceCreate{
		Name:       "zero",
		BaseURL:    "http://wled.test",
		LightCount: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	srv := newCaptureTestServer(t, st)
	res, err := http.Post(srv.URL+"/api/v1/devices/"+d.ID+"/capture/start", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnprocessableEntity {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status=%d want 422, body=%s", res.StatusCode, b)
	}
	var env struct {
		Error struct{ Code string `json:"code"` } `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error.Code != "capture_no_lights" {
		t.Fatalf("code = %q want capture_no_lights", env.Error.Code)
	}
}
