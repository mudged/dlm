package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("HTTP_LISTEN", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "")
	t.Setenv("HTTP_WRITE_TIMEOUT_SEC", "")
	t.Setenv("DLM_DB_PATH", "")
	t.Setenv("DLM_DATA_DIR", "")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPListen != ":8080" {
		t.Fatalf("HTTPListen = %q", c.HTTPListen)
	}
	if c.ReadTimeout != 15*time.Second {
		t.Fatalf("ReadTimeout = %v", c.ReadTimeout)
	}
	if len(c.CORSAllowedOrigins) != 4 {
		t.Fatalf("CORSAllowedOrigins = %v (want four default Next dev origins)", c.CORSAllowedOrigins)
	}
	want := []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost:8000",
		"http://127.0.0.1:8000",
	}
	for i := range want {
		if c.CORSAllowedOrigins[i] != want[i] {
			t.Fatalf("CORSAllowedOrigins[%d] = %q want %q", i, c.CORSAllowedOrigins[i], want[i])
		}
	}
	wantDB := filepath.Join("data", "dlm.db")
	if c.DBPath != wantDB {
		t.Fatalf("DBPath = %q want %q", c.DBPath, wantDB)
	}
}

func TestLoad_corsExplicitOptOut(t *testing.T) {
	t.Setenv("HTTP_LISTEN", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "-")
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "")
	t.Setenv("HTTP_WRITE_TIMEOUT_SEC", "")
	t.Setenv("DLM_DB_PATH", "")
	t.Setenv("DLM_DATA_DIR", "")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.CORSAllowedOrigins) != 0 {
		t.Fatalf("CORSAllowedOrigins = %v want empty", c.CORSAllowedOrigins)
	}
}

func TestLoad_customListenAndCORS(t *testing.T) {
	t.Setenv("HTTP_LISTEN", ":9090")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, http://127.0.0.1:3000")
	t.Setenv("DLM_DB_PATH", "")
	t.Setenv("DLM_DATA_DIR", "")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPListen != ":9090" {
		t.Fatalf("HTTPListen = %q", c.HTTPListen)
	}
	if len(c.CORSAllowedOrigins) != 2 {
		t.Fatalf("got %d origins", len(c.CORSAllowedOrigins))
	}
	if c.CORSAllowedOrigins[0] != "http://localhost:3000" || c.CORSAllowedOrigins[1] != "http://127.0.0.1:3000" {
		t.Fatalf("origins = %v", c.CORSAllowedOrigins)
	}
}

func TestLoad_invalidTimeout(t *testing.T) {
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "not-a-number")
	t.Setenv("DLM_DB_PATH", "")
	t.Setenv("DLM_DATA_DIR", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_dlmDBPath(t *testing.T) {
	t.Setenv("HTTP_LISTEN", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "")
	t.Setenv("HTTP_WRITE_TIMEOUT_SEC", "")
	t.Setenv("DLM_DB_PATH", "/tmp/custom.db")
	t.Setenv("DLM_DATA_DIR", "")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.DBPath != "/tmp/custom.db" {
		t.Fatalf("DBPath = %q", c.DBPath)
	}
}
