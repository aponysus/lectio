package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aponysus/lectio/internal/auth"
	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/importer"
	"github.com/aponysus/lectio/internal/server/routes"
	"github.com/aponysus/lectio/internal/store"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"serve"}
	}

	switch args[0] {
	case "serve":
		if err := runServer(cfg, logger); err != nil {
			logger.Error("server exited with error", "error", err)
			os.Exit(1)
		}
	case "migrate":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: lectio migrate [up|status]")
			os.Exit(1)
		}
		if err := runMigrations(cfg, logger, args[1]); err != nil {
			logger.Error("migration command failed", "error", err)
			os.Exit(1)
		}
	case "import-v2":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: lectio import-v2 <path-to-legacy-json>")
			os.Exit(1)
		}
		if err := runImportV2(cfg, logger, args[1]); err != nil {
			logger.Error("v2 import failed", "error", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		os.Exit(1)
	}
}

func runServer(cfg config.Config, logger *slog.Logger) error {
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	if err := store.ApplyMigrations(ctx, db); err != nil {
		return err
	}

	repo := store.New(db)
	authManager := auth.NewManager(cfg.SessionSecret, cfg.CSRFSecret, cfg.SessionTTL, cfg.IsProduction())

	router := routes.New(routes.Dependencies{
		Logger: logger,
		Config: cfg,
		Store:  repo,
		Auth:   authManager,
	})

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", cfg.Addr, "env", cfg.Env)
		errCh <- server.ListenAndServe()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signalCh:
		logger.Info("received shutdown signal", "signal", sig.String())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func runMigrations(cfg config.Config, logger *slog.Logger, mode string) error {
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()

	switch mode {
	case "up":
		return store.ApplyMigrations(ctx, db)
	case "status":
		statuses, err := store.ListMigrations(ctx, db)
		if err != nil {
			return err
		}
		for _, status := range statuses {
			appliedAt := "pending"
			if status.AppliedAt != nil {
				appliedAt = status.AppliedAt.Format(time.RFC3339)
			}
			logger.Info("migration status",
				"name", status.Name,
				"applied", status.Applied,
				"applied_at", appliedAt,
			)
		}
		return nil
	default:
		return fmt.Errorf("unknown migration mode %q", mode)
	}
}

func runImportV2(cfg config.Config, logger *slog.Logger, path string) error {
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	if err := store.ApplyMigrations(ctx, db); err != nil {
		return err
	}

	result, err := importer.ImportV2File(ctx, store.New(db), path)
	if err != nil {
		return err
	}

	logger.Info("v2 import complete",
		"path", path,
		"sources_created", result.SourcesCreated,
		"engagements_created", result.EngagementsCreated,
	)
	return nil
}
