package store

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	embeddedmigrations "github.com/aponysus/lectio/migrations"
)

type MigrationStatus struct {
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	if err := ensureMigrationTable(ctx, db); err != nil {
		return err
	}

	applied, err := appliedMigrationMap(ctx, db)
	if err != nil {
		return err
	}

	entries, err := migrationEntries()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if applied[entry] {
			continue
		}

		rawMigration, err := embeddedmigrations.FS.ReadFile(entry)
		if err != nil {
			return err
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, string(rawMigration)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry, err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (name, applied_at) VALUES (?, CURRENT_TIMESTAMP)`,
			entry,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func ListMigrations(ctx context.Context, db *sql.DB) ([]MigrationStatus, error) {
	if err := ensureMigrationTable(ctx, db); err != nil {
		return nil, err
	}

	entries, err := migrationEntries()
	if err != nil {
		return nil, err
	}

	appliedAtByName := map[string]time.Time{}
	rows, err := db.QueryContext(ctx, `SELECT name, applied_at FROM schema_migrations ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var appliedAt time.Time
		if err := rows.Scan(&name, &appliedAt); err != nil {
			return nil, err
		}
		appliedAtByName[name] = appliedAt
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	statuses := make([]MigrationStatus, 0, len(entries))
	for _, entry := range entries {
		status := MigrationStatus{Name: entry}
		if appliedAt, ok := appliedAtByName[entry]; ok {
			status.Applied = true
			status.AppliedAt = &appliedAt
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL
		)
	`)
	return err
}

func appliedMigrationMap(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, rows.Err()
}

func migrationEntries() ([]string, error) {
	entries, err := fs.ReadDir(embeddedmigrations.FS, ".")
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		names = append(names, entry.Name())
	}

	sort.Strings(names)
	return names, nil
}
