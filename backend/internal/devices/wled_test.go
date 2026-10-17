package devices

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/dlm/backend/internal/store"
)

func TestPostJSONState_success(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s want POST", r.Method)
		}
		if r.URL.Path != "/json/state" {
			t.Errorf("path: got %s want /json/state", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	payload := map[string]any{"on": true, "bri": 255}
	if err := postJSONState(context.Background(), deviceHTTPClient(), srv.URL, "", payload); err != nil {
		t.Fatalf("postJSONState: %v", err)
	}
	if gotBody == nil || gotBody["on"] != true {
		t.Fatalf("server body: %#v", gotBody)
	}
}

func TestPostJSONState_doesNotFollowRedirects(t *testing.T) {
	t.Parallel()

	var attackerHit bool
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attackerHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer attacker.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/state" {
			t.Errorf("path: got %s want /json/state", r.URL.Path)
		}
		http.Redirect(w, r, attacker.URL, http.StatusFound)
	}))
	defer redirector.Close()

	err := postJSONState(context.Background(), deviceHTTPClient(), redirector.URL, "", map[string]any{"on": true})
	if err == nil {
		t.Fatal("expected error for redirect response")
	}
	if !strings.Contains(err.Error(), "302") {
		t.Fatalf("expected 302 in error, got: %v", err)
	}
	if attackerHit {
		t.Fatal("client followed redirect to attacker URL")
	}
}

func TestBuildWLEDState_colourIsRGBWithoutBulbNumber(t *testing.T) {
	t.Parallel()
	body := buildWLEDState([]store.LightStateDTO{
		{ID: 0, On: false, Color: "#ffffff", BrightnessPct: 100},
		{ID: 1, On: true, Color: "#ff0000", BrightnessPct: 100},
		{ID: 2, On: true, Color: "#00ff00", BrightnessPct: 50},
	})
	seg, ok := body["seg"].([]map[string]any)
	if !ok || len(seg) != 1 {
		t.Fatalf("seg: %#v", body["seg"])
	}
	got, ok := seg[0]["i"].([][]int)
	if !ok {
		t.Fatalf("i: %#v", seg[0]["i"])
	}
	want := [][]int{{0, 0, 0}, {255, 0, 0}, {0, 128, 0}}
	if len(got) != len(want) {
		t.Fatalf("colors: got %#v want %#v", got, want)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("colors: got %#v want %#v", got, want)
		}
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("colors: got %#v want %#v", got, want)
			}
		}
	}
}
