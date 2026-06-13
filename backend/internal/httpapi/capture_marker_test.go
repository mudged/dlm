package httpapi

import (
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
