package store

import (
	"context"
	"database/sql"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) SystemStatus(ctx context.Context, env string) (model.SystemStatus, error) {
	status := model.SystemStatus{
		AppName:     "Lectio",
		Environment: env,
	}

	var bootstrappedAt sql.NullString
	if err := s.db.QueryRowContext(ctx, `
		SELECT value
		FROM app_meta
		WHERE key = 'bootstrapped_at'
	`).Scan(&bootstrappedAt); err != nil {
		return model.SystemStatus{}, err
	}

	if bootstrappedAt.Valid {
		status.BootstrappedAt = bootstrappedAt.String
	}

	if err := s.db.QueryRowContext(ctx, `SELECT datetime('now')`).Scan(&status.DatabaseTime); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&status.AppliedMigrations); err != nil {
		return model.SystemStatus{}, err
	}

	return status, nil
}
