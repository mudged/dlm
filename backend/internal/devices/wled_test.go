package devices

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
