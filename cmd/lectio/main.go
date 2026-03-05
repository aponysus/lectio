package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/server"
	"github.com/aponysus/lectio/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	db, err := store.Open(context.Background(), store.OpenConfig{
		Path:          cfg.DBPath,
		MigrationsDir: cfg.MigrationsDir,
		AutoMigrate:   cfg.AutoMigrate,
	})
	if err != nil {
		logger.Error("failed to open sqlite store", "error", err, "path", cfg.DBPath)
		os.Exit(1)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("failed to close sqlite store", "error", closeErr)
		}
	}()
	bootstrapUser, err := db.Users().EnsureBootstrapUser(context.Background(), cfg.BootstrapEmail, cfg.BootstrapPassword)
	if err != nil {
		logger.Error("failed to ensure bootstrap user", "error", err)
		os.Exit(1)
	}
	logger.Info("bootstrap user ready", "user_id", bootstrapUser.ID, "email", bootstrapUser.Email)

	srv := server.New(cfg, logger, db)

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("starting http server", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("http server exited with error", "error", serveErr)
			os.Exit(1)
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-sigCtx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
