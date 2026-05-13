package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/lightstate"
	"example.com/dlm/backend/internal/store"
	"example.com/dlm/backend/internal/wiremodel"
)

func TestAPIv1Devices_discover_returns501(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/api/v1/devices/discover", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotImplemented {
		t.Fatalf("status = %d", res.StatusCode)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "not_implemented" {
		t.Fatalf("code = %q", body.Error.Code)
	}
}

func TestAPIv1Devices_createListAssignUnassign(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "devices_api.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	st.SetLightState(lightstate.New())
	if err := st.LoadLightStateFromDB(ctx); err != nil {
		t.Fatal(err)
	}
	sum, err := st.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             path,
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
	t.Cleanup(srv.Close)
	base := srv.URL + "/api/v1/devices"

	createBody := `{"name":"Strip","base_url":"http://wled.test"}`
	res, err := http.Post(base, "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", res.StatusCode, b)
	}
	var created map[string]any
	if err := json.Unmarshal(b, &created); err != nil {
		t.Fatal(err)
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatalf("missing id in %s", b)
	}

	res2, err := http.Get(base)
	if err != nil {
		t.Fatal(err)
	}
	b2, _ := io.ReadAll(res2.Body)
	_ = res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("list status=%d body=%s", res2.StatusCode, b2)
	}
	var listWrap struct {
		Devices []map[string]any `json:"devices"`
	}
	if err := json.Unmarshal(b2, &listWrap); err != nil || len(listWrap.Devices) != 1 {
		t.Fatalf("list decode: err=%v n=%d", err, len(listWrap.Devices))
	}

	assignURL := base + "/" + id + "/assign"
	assignPayload := map[string]string{"model_id": sum.ID}
	aj, _ := json.Marshal(assignPayload)
	res3, err := http.Post(assignURL, "application/json", bytes.NewReader(aj))
	if err != nil {
		t.Fatal(err)
	}
	b3, _ := io.ReadAll(res3.Body)
	_ = res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("assign status=%d body=%s", res3.StatusCode, b3)
	}

	res4, err := http.Get(base + "/" + id)
	if err != nil {
		t.Fatal(err)
	}
	b4, _ := io.ReadAll(res4.Body)
	_ = res4.Body.Close()
	if res4.StatusCode != http.StatusOK {
		t.Fatalf("get status=%d body=%s", res4.StatusCode, b4)
	}
	var one map[string]any
	if err := json.Unmarshal(b4, &one); err != nil {
		t.Fatal(err)
	}
	if one["model_id"] != sum.ID {
		t.Fatalf("model_id = %v", one["model_id"])
	}

	unURL := base + "/" + id + "/unassign"
	res5, err := http.Post(unURL, "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	b5, _ := io.ReadAll(res5.Body)
	_ = res5.Body.Close()
	if res5.StatusCode != http.StatusOK {
		t.Fatalf("unassign status=%d body=%s", res5.StatusCode, b5)
	}
	var un map[string]any
	if err := json.Unmarshal(b5, &un); err != nil {
		t.Fatal(err)
	}
	if _, has := un["model_id"]; has {
		t.Fatalf("expected model_id omitted after unassign, got %+v", un)
	}
}

func TestAPIv1Devices_postRejectsInvalidBaseURL(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)
	base := srv.URL + "/api/v1/devices"

	cases := []struct {
		name string
		body string
	}{
		{"file_scheme", `{"name":"x","base_url":"file:///etc/passwd"}`},
		{"ftp_scheme", `{"name":"x","base_url":"ftp://example.com/"}`},
		{"empty_host", `{"name":"x","base_url":"http://"}`},
		{"http_no_host", `{"name":"x","base_url":"http:"}`},
		{"whitespace", `{"name":"x","base_url":"   "}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := http.Post(base, "application/json", strings.NewReader(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status=%d want 400", res.StatusCode)
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
			if env.Error.Code != "invalid_base_url" {
				t.Fatalf("code = %q want invalid_base_url (msg=%q)", env.Error.Code, env.Error.Message)
			}
		})
	}
}

func TestAPIv1Devices_patchRejectsInvalidBaseURL(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "devices_patch_invalid.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	st.SetLightState(lightstate.New())
	if err := st.LoadLightStateFromDB(ctx); err != nil {
		t.Fatal(err)
	}
	d, err := st.CreateDevice(ctx, store.DeviceCreate{Name: "p", BaseURL: "http://wled.local"})
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		HTTPListen:   ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		DBPath:       path,
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
	t.Cleanup(srv.Close)

	patchURL := srv.URL + "/api/v1/devices/" + d.ID
	req, _ := http.NewRequest(http.MethodPatch, patchURL, strings.NewReader(`{"base_url":"file:///etc/passwd"}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400", res.StatusCode)
	}
	var env struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Error.Code != "invalid_base_url" {
		t.Fatalf("code = %q want invalid_base_url", env.Error.Code)
	}
}

func TestAPIv1Devices_postAcceptsHTTPAndNormalizes(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)
	base := srv.URL + "/api/v1/devices"

	body := `{"name":"trim","base_url":"https://1.2.3.4:8080/"}`
	res, err := http.Post(base, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status=%d body=%s", res.StatusCode, b)
	}
	var created map[string]any
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if got, _ := created["base_url"].(string); got != "https://1.2.3.4:8080" {
		t.Fatalf("base_url = %q want trailing slash trimmed", got)
	}
}

func TestAPIv1Devices_assignConflict_returns409(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "devices_409.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	st.SetLightState(lightstate.New())
	if err := st.LoadLightStateFromDB(ctx); err != nil {
		t.Fatal(err)
	}
	sum, err := st.Create(ctx, "m1", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	d1, err := st.CreateDevice(ctx, store.DeviceCreate{Name: "a", BaseURL: "http://a"})
	if err != nil {
		t.Fatal(err)
	}
	d2, err := st.CreateDevice(ctx, store.DeviceCreate{Name: "b", BaseURL: "http://b"})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AssignDevice(ctx, d1.ID, sum.ID); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             path,
	}
	srv := httptest.NewServer(NewSiteHandler(cfg, nil, st, nil, nil))
	t.Cleanup(srv.Close)

	assignURL := srv.URL + "/api/v1/devices/" + d2.ID + "/assign"
	aj, _ := json.Marshal(map[string]string{"model_id": sum.ID})
	res, err := http.Post(assignURL, "application/json", bytes.NewReader(aj))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status=%d", res.StatusCode)
	}
}
