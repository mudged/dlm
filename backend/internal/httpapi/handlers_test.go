package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/lightstate"
	"example.com/dlm/backend/internal/store"
	"example.com/dlm/backend/internal/wiremodel"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	ls := lightstate.New()
	s.SetLightState(ls)
	ctx := context.Background()
	if err := s.LoadLightStateFromDB(ctx); err != nil {
		t.Fatal(err)
	}
	return s
}

func newTestHandler(t *testing.T, cfg *config.Config) http.Handler {
	t.Helper()
	if cfg == nil {
		cfg = &config.Config{
			HTTPListen:         ":8080",
			ReadTimeout:        15 * time.Second,
			WriteTimeout:       15 * time.Second,
			CORSAllowedOrigins: nil,
			DBPath:             filepath.Join(t.TempDir(), "unused.db"),
		}
	}
	return NewSiteHandler(cfg, nil, testStore(t), nil, nil)
}

func TestHealth_returnsOKJSON(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status field = %q", body.Status)
	}
}

func TestAPIv1Status_returnsJSON(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/status")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	var body struct {
		Service string `json:"service"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Service != "dlm-api" || body.Version == "" {
		t.Fatalf("body = %+v", body)
	}
}

func TestAPIv1UnknownRoute_returnsErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/no-such-resource")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d", res.StatusCode)
	}
	b, _ := io.ReadAll(res.Body)
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("body %s: %v", b, err)
	}
	if env.Error.Code == "" || env.Error.Message == "" {
		t.Fatalf("error envelope = %+v", env)
	}
}

func TestCORSPreflight_allowsConfiguredOrigin(t *testing.T) {
	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, testStore(t), nil, nil))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodOptions, srv.URL+"/api/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestStatic_servesEmbeddableExport(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!doctype html><html><body>ui-ok</body></html>"),
		},
		"_next/static/chunk.js": &fstest.MapFile{Data: []byte("//x")},
	}
	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, fsys, testStore(t), nil, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "ui-ok") {
		t.Fatalf("body = %q", body)
	}

	res2, err := http.Get(srv.URL + "/_next/static/chunk.js")
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("GET chunk status = %d", res2.StatusCode)
	}
}

func TestAPI_precedenceOverStaticPrefix(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("html")},
	}
	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, fsys, testStore(t), nil, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestStatic_unknownClientRoute_fallsBackToIndexHTML(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!doctype html><html><body>spa-fallback</body></html>"),
		},
	}
	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, fsys, testStore(t), nil, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/settings/profile")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "spa-fallback") {
		t.Fatalf("body = %q", body)
	}
}

func TestStatic_missingNextAsset_returnsNotFound(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("ok")},
	}
	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, fsys, testStore(t), nil, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/_next/static/missing.js")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestAPIv1FactoryReset_wrongMethod_returnsJSONEnvelope(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/system/factory-reset")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d want 405", res.StatusCode)
	}
	ct := res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q want application/json", ct)
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
	if env.Error.Code != "method_not_allowed" {
		t.Fatalf("code = %q want method_not_allowed", env.Error.Code)
	}
	if env.Error.Message == "" {
		t.Fatalf("message empty")
	}
}

func TestAPIv1FactoryReset_resetsToThreeSamples(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "factory.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ls := lightstate.New()
	st.SetLightState(ls)
	if err := st.SeedDefaultSamples(ctx); err != nil {
		t.Fatal(err)
	}
	if err := st.LoadLightStateFromDB(ctx); err != nil {
		t.Fatal(err)
	}
	userSum, err := st.Create(ctx, "user-model", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	dev, err := st.CreateDevice(ctx, store.DeviceCreate{Name: "wled-strip", BaseURL: "http://192.168.1.50"})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AssignDevice(ctx, dev.ID, userSum.ID); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             path,
	}
	h := NewSiteHandler(cfg, nil, st, nil, nil)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	resDev, err := http.Get(srv.URL + "/api/v1/devices")
	if err != nil {
		t.Fatal(err)
	}
	defer resDev.Body.Close()
	if resDev.StatusCode != http.StatusOK {
		t.Fatalf("devices list before reset: status %d", resDev.StatusCode)
	}
	var devList struct {
		Devices []any `json:"devices"`
	}
	if err := json.NewDecoder(resDev.Body).Decode(&devList); err != nil {
		t.Fatal(err)
	}
	if len(devList.Devices) != 1 {
		t.Fatalf("before reset want 1 device, got %d", len(devList.Devices))
	}

	res, err := http.Post(srv.URL+"/api/v1/system/factory-reset", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d body = %s", res.StatusCode, b)
	}
	var okBody struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(res.Body).Decode(&okBody); err != nil || !okBody.OK {
		t.Fatalf("decode ok: err=%v ok=%v", err, okBody.OK)
	}

	resDev2, err := http.Get(srv.URL + "/api/v1/devices")
	if err != nil {
		t.Fatal(err)
	}
	defer resDev2.Body.Close()
	if resDev2.StatusCode != http.StatusOK {
		t.Fatalf("devices list after reset: status %d", resDev2.StatusCode)
	}
	var devList2 struct {
		Devices []any `json:"devices"`
	}
	if err := json.NewDecoder(resDev2.Body).Decode(&devList2); err != nil {
		t.Fatal(err)
	}
	if len(devList2.Devices) != 0 {
		t.Fatalf("after reset want 0 devices, got %d", len(devList2.Devices))
	}

	res2, err := http.Get(srv.URL + "/api/v1/models")
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d", res2.StatusCode)
	}
	var models []store.Summary
	if err := json.NewDecoder(res2.Body).Decode(&models); err != nil {
		t.Fatal(err)
	}
	if len(models) != 3 {
		t.Fatalf("after factory reset want 3 models, got %d", len(models))
	}

	res3, err := http.Get(srv.URL + "/api/v1/routines")
	if err != nil {
		t.Fatal(err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("routines list status = %d", res3.StatusCode)
	}
	var routines []store.RoutineDTO
	if err := json.NewDecoder(res3.Body).Decode(&routines); err != nil {
		t.Fatal(err)
	}
	if len(routines) != 3 {
		t.Fatalf("after factory reset want 3 routines, got %d", len(routines))
	}
	for _, r := range routines {
		if r.Type != store.RoutineTypePythonSceneScript || r.PythonSource == "" {
			t.Fatalf("want python_scene_script with source, got %+v", r)
		}
	}
}
