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
	"example.com/dlm/backend/internal/devices"
	"example.com/dlm/backend/internal/httpapi"
	"example.com/dlm/backend/internal/lightstate"
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

	ls := lightstate.New()
	st.SetLightState(ls)

	ctx := context.Background()
	if err := st.SeedDefaultSamples(ctx); err != nil {
		log.Error("seed default samples", "err", err)
		os.Exit(1)
	}
	if err := st.SeedDefaultPythonRoutines(ctx); err != nil {
		log.Error("seed default python routines", "err", err)
		os.Exit(1)
	}
	if err := st.LoadLightStateFromDB(ctx); err != nil {
		log.Error("load light state from db", "err", err)
		os.Exit(1)
	}

	ui, err := webdist.StaticFS()
	if err != nil {
		log.Error("webdist", "err", err)
		os.Exit(1)
	}

	revHub := httpapi.NewRevisionHub()
	pusher := devices.NewPusher(st, nil)
	if err := pusher.SyncAllAssignedModels(ctx); err != nil {
		log.Warn("wled sync on startup", "err", err)
	}
	handler := httpapi.NewSiteHandler(cfg, ui, st, revHub, pusher)
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

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
		os.Exit(1)
	}
	log.Info("stopped")
}
