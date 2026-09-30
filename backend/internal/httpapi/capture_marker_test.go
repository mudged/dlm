package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/dlm/backend/internal/config"
)

func TestGetCaptureMarker_returnsPDF(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, &config.Config{
		HTTPListen:         ":8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		CORSAllowedOrigins: nil,
		DBPath:             filepath.Join(t.TempDir(), "unused.db"),
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/capture/marker")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/pdf" {
		t.Fatalf("Content-Type = %q, want application/pdf", ct)
	}
	cd := res.Header.Get("Content-Disposition")
	if !strings.Contains(cd, "inline") || !strings.Contains(cd, "fiducial_marker_aruco4x4_50_id0_100mm.pdf") {
		t.Fatalf("Content-Disposition = %q", cd)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("empty body")
	}
	if !strings.HasPrefix(string(body), "%PDF") {
		t.Fatalf("body does not look like PDF (prefix %q)", string(body[:min(8, len(body))]))
	}
}

func TestGetCaptureMarker_typePNG(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/capture/marker?type=png")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "image/png" {
		t.Fatalf("Content-Type = %q, want image/png", ct)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("empty body")
	}
	// PNG magic
	if body[0] != 0x89 || string(body[1:4]) != "PNG" {
		t.Fatal("body does not look like PNG")
	}
}

func TestGetCaptureMarker_idSelectsDistinctPDF(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	bodies := map[string][]byte{}
	for _, id := range []string{"0", "1", "2"} {
		res, err := http.Get(srv.URL + "/api/v1/capture/marker?id=" + id)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("id %s status = %d", id, res.StatusCode)
		}
		name := "fiducial_marker_aruco4x4_50_id" + id + "_100mm.pdf"
		if cd := res.Header.Get("Content-Disposition"); !strings.Contains(cd, name) {
			t.Fatalf("id %s Content-Disposition = %q", id, cd)
		}
		if !bytes.Contains(body, []byte("stick them on different sides")) {
			t.Fatalf("id %s pdf missing placement note", id)
		}
		bodies[id] = body
	}
	if bytes.Equal(bodies["0"], bodies["1"]) || bytes.Equal(bodies["1"], bodies["2"]) {
		t.Fatal("marker pdfs are not distinct")
	}
}

func TestGetCaptureMarker_idPNGDistinct(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	var prev []byte
	for _, id := range []string{"0", "1", "2"} {
		res, err := http.Get(srv.URL + "/api/v1/capture/marker?id=" + id + "&type=png")
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("id %s status = %d", id, res.StatusCode)
		}
		if res.Header.Get("Content-Type") != "image/png" {
			t.Fatalf("id %s Content-Type = %q", id, res.Header.Get("Content-Type"))
		}
		if len(body) < 8 || body[0] != 0x89 || string(body[1:4]) != "PNG" {
			t.Fatalf("id %s is not a png", id)
		}
		if prev != nil && bytes.Equal(prev, body) {
			t.Fatalf("id %s png matches the previous id", id)
		}
		prev = body
	}
}

func TestGetCaptureMarker_badID(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t, nil))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/capture/marker?id=3")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", res.StatusCode)
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != "bad_request" || payload.Error.Message != "marker id must be 0, 1, or 2" {
		t.Fatalf("error = %#v", payload.Error)
	}
}
