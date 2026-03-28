package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds HTTP server and CORS settings (12-factor env).
type Config struct {
	HTTPListen         string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	CORSAllowedOrigins []string
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

	var origins []string
	if raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			if o := strings.TrimSpace(p); o != "" {
				origins = append(origins, o)
			}
		}
	}

	return &Config{
		HTTPListen:         listen,
		ReadTimeout:        time.Duration(readSec) * time.Second,
		WriteTimeout:       time.Duration(writeSec) * time.Second,
		CORSAllowedOrigins: origins,
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
