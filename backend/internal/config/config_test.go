package config

import (
	"testing"
	"time"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("HTTP_LISTEN", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "")
	t.Setenv("HTTP_WRITE_TIMEOUT_SEC", "")

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
	if len(c.CORSAllowedOrigins) != 0 {
		t.Fatalf("CORSAllowedOrigins = %v", c.CORSAllowedOrigins)
	}
}

func TestLoad_customListenAndCORS(t *testing.T) {
	t.Setenv("HTTP_LISTEN", ":9090")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, http://127.0.0.1:3000")

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
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}
