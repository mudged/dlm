package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds HTTP server, CORS, and SQLite path (12-factor env).
type Config struct {
	HTTPListen         string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	CORSAllowedOrigins []string
	DBPath             string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	listen := os.Getenv("HTTP_LISTEN")
	if listen == "" {
		listen = ":8080"
	}

	readSec, err := getenvIntOrDefault("HTTP_READ_TIMEOUT_SEC", 15)
	if err != nil {
		return nil, err
	}
	writeSec, err := getenvIntOrDefault("HTTP_WRITE_TIMEOUT_SEC", 15)
	if err != nil {
		return nil, err
	}
	if readSec < 1 || writeSec < 1 {
		return nil, fmt.Errorf("HTTP_READ_TIMEOUT_SEC and HTTP_WRITE_TIMEOUT_SEC must be >= 1")
	}

	// Cross-origin EventSource from `next dev` (:3000) to Go (:8080) requires CORS (see web/lib/sseUrl.ts).
	// When unset, allow typical Next dev origins so SSE works without extra env; production embedded UI is same-origin and ignores this.
	var origins []string
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	switch {
	case raw == "-":
		origins = nil
	case raw != "":
		for _, p := range strings.Split(raw, ",") {
			if o := strings.TrimSpace(p); o != "" {
				origins = append(origins, o)
			}
		}
	default:
		// Common `next dev` ports; EventSource uses NEXT_PUBLIC_DLM_API_ORIGIN → Go (often :8080) cross-origin.
		origins = []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:8000",
			"http://127.0.0.1:8000",
		}
	}

	dbPath := strings.TrimSpace(os.Getenv("DLM_DB_PATH"))
	if dbPath == "" {
		dataDir := strings.TrimSpace(os.Getenv("DLM_DATA_DIR"))
		if dataDir == "" {
			dataDir = "data"
		}
		dbPath = filepath.Join(dataDir, "dlm.db")
	}

	return &Config{
		HTTPListen:         listen,
		ReadTimeout:        time.Duration(readSec) * time.Second,
		WriteTimeout:       time.Duration(writeSec) * time.Second,
		CORSAllowedOrigins: origins,
		DBPath:             dbPath,
	}, nil
}

func getenvIntOrDefault(key string, def int) (int, error) {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer %q", key, s)
	}
	return n, nil
}
