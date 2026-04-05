package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"example.com/dlm/backend/internal/config"
	"example.com/dlm/backend/internal/httpapi"
	"example.com/dlm/backend/internal/store"
	"example.com/dlm/backend/internal/webdist"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Error("mkdir db dir", "path", cfg.DBPath, "err", err)
		os.Exit(1)
	}
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Error("store", "path", cfg.DBPath, "err", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	if err := st.SeedDefaultSamples(context.Background()); err != nil {
		log.Error("seed default samples", "err", err)
		os.Exit(1)
	}

	ui, err := webdist.StaticFS()
	if err != nil {
		log.Error("webdist", "err", err)
		os.Exit(1)
	}

	handler := httpapi.NewSiteHandler(cfg, ui, st)
	srv := &http.Server{
		Addr:              cfg.HTTPListen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}

	go func() {
		log.Info("listening", "addr", cfg.HTTPListen)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()
	go httpapi.StartRoutineScheduler(schedCtx, log, st)

	<-ctx.Done()
	schedCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
		os.Exit(1)
	}
	log.Info("stopped")
}
