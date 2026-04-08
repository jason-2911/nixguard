// Package main is the entry point for the NixGuard API server.
// The server runs as an unprivileged user and communicates with
// the privileged agent via Unix domain socket for system operations.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nixguard/nixguard/internal/config"
	"github.com/nixguard/nixguard/internal/event"
	"github.com/nixguard/nixguard/pkg/logger"
	"github.com/nixguard/nixguard/pkg/version"
)

func main() {
	// ── Bootstrap ──────────────────────────────────────────────
	cfg, err := config.Load("configs/defaults/server.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("starting nixguard-server",
		slog.String("version", version.Version),
		slog.String("build_time", version.BuildTime),
	)

	// ── Event Bus ──────────────────────────────────────────────
	bus := event.NewBus(log)

	// ── Module Registry ────────────────────────────────────────
	// Each module registers itself with the application.
	// Modules are initialized in dependency order.
	app, err := initApp(cfg, bus, log)
	if err != nil {
		log.Error("failed to initialize application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ── HTTP Server ────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Server.ListenAddr,
		Handler:      app.Router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ── Graceful Shutdown ──────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("server listening", slog.String("addr", cfg.Server.ListenAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced shutdown", slog.String("error", err.Error()))
	}

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error("app shutdown error", slog.String("error", err.Error()))
	}

	log.Info("server stopped")
}
