package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env               string
	HTTPAddr          string
	LogLevel          slog.Level
	DBPath            string
	MigrationsDir     string
	AutoMigrate       bool
	SessionCookieName string
	CSRFCookieName    string
	CSRFHeaderName    string

	BootstrapEmail    string
	BootstrapPassword string
	SessionSecret     string
	CSRFSecret        string
}

func Load() (Config, error) {
	cfg := Config{
		Env:               getEnv("LECTIO_ENV", "development"),
		HTTPAddr:          getEnv("LECTIO_ADDR", ":8080"),
		LogLevel:          parseLogLevel(getEnv("LECTIO_LOG_LEVEL", "info")),
		DBPath:            getEnv("LECTIO_DB_PATH", "./devdata/lectio.db"),
		MigrationsDir:     getEnv("LECTIO_MIGRATIONS_DIR", "./migrations"),
		SessionCookieName: "lectio_session",
		CSRFCookieName:    "lectio_csrf",
		CSRFHeaderName:    "X-CSRF-Token",
		BootstrapEmail:    strings.TrimSpace(os.Getenv("LECTIO_BOOTSTRAP_EMAIL")),
		BootstrapPassword: strings.TrimSpace(os.Getenv("LECTIO_BOOTSTRAP_PASSWORD")),
		SessionSecret:     strings.TrimSpace(os.Getenv("LECTIO_SESSION_SECRET")),
		CSRFSecret:        strings.TrimSpace(os.Getenv("LECTIO_CSRF_SECRET")),
	}
	cfg.AutoMigrate = getEnvBool("LECTIO_AUTO_MIGRATE", cfg.Env != "production")

	if cfg.BootstrapPassword == "" {
		cfg.BootstrapPassword = "change-me"
	}
	if cfg.SessionSecret == "" {
		cfg.SessionSecret = "dev-session-secret"
	}
	if cfg.CSRFSecret == "" {
		cfg.CSRFSecret = "dev-csrf-secret"
	}

	if cfg.Env == "production" {
		if cfg.BootstrapPassword == "" || cfg.BootstrapPassword == "change-me" {
			return Config{}, fmt.Errorf("LECTIO_BOOTSTRAP_PASSWORD must be set in production")
		}
		if cfg.SessionSecret == "" || cfg.CSRFSecret == "" {
			return Config{}, fmt.Errorf("LECTIO_SESSION_SECRET and LECTIO_CSRF_SECRET must be set in production")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
