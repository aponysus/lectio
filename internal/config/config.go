package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Env               string
	Addr              string
	DBPath            string
	WebDistDir        string
	BootstrapPassword string
	SessionSecret     string
	CSRFSecret        string
	SessionTTL        time.Duration
}

func Load() Config {
	return Config{
		Env:               getenv("LECTIO_ENV", "development"),
		Addr:              getenv("LECTIO_ADDR", ":8080"),
		DBPath:            getenv("LECTIO_DB_PATH", "./devdata/lectio.db"),
		WebDistDir:        getenv("LECTIO_WEB_DIST", "./web/dist"),
		BootstrapPassword: getenv("LECTIO_BOOTSTRAP_PASSWORD", "changeme"),
		SessionSecret:     getenv("LECTIO_SESSION_SECRET", "development-session-secret-change-me"),
		CSRFSecret:        getenv("LECTIO_CSRF_SECRET", "development-csrf-secret-change-me"),
		SessionTTL:        12 * time.Hour,
	}
}

func (c Config) Validate() error {
	switch {
	case c.Addr == "":
		return errors.New("LECTIO_ADDR must not be empty")
	case c.DBPath == "":
		return errors.New("LECTIO_DB_PATH must not be empty")
	case c.BootstrapPassword == "":
		return errors.New("LECTIO_BOOTSTRAP_PASSWORD must not be empty")
	case c.SessionSecret == "":
		return errors.New("LECTIO_SESSION_SECRET must not be empty")
	case c.CSRFSecret == "":
		return errors.New("LECTIO_CSRF_SECRET must not be empty")
	default:
		return nil
	}
}

func (c Config) IsProduction() bool {
	return c.Env == "production"
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
