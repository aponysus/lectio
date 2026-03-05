package store_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aponysus/lectio/internal/store"
)

func TestOpenAutoMigrateAppliesPendingMigrations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "migrations")
	writeMigration(t, migrationsDir, "0001_init.up.sql", `
CREATE TABLE test_entries (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL
);
`)

	dbPath := filepath.Join(dir, "lectio.db")
	st, err := store.Open(ctx, store.OpenConfig{
		Path:          dbPath,
		MigrationsDir: migrationsDir,
		AutoMigrate:   true,
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	state, err := st.MigrationState(ctx)
	if err != nil {
		t.Fatalf("migration state: %v", err)
	}
	if state.CurrentVersion != 1 || state.LatestVersion != 1 {
		t.Fatalf("unexpected migration state: %+v", state)
	}

	var journalMode string
	if err := st.DB().QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("journal_mode pragma: %v", err)
	}
	if strings.ToLower(journalMode) != "wal" {
		t.Fatalf("unexpected journal mode: %q", journalMode)
	}
}

func TestOpenFailsWhenSchemaBehindAndAutoMigrateDisabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "migrations")
	writeMigration(t, migrationsDir, "0001_init.up.sql", `
CREATE TABLE test_entries (
	id INTEGER PRIMARY KEY
);
`)

	dbPath := filepath.Join(dir, "lectio.db")
	st, err := store.Open(ctx, store.OpenConfig{
		Path:          dbPath,
		MigrationsDir: migrationsDir,
		AutoMigrate:   true,
	})
	if err != nil {
		t.Fatalf("seed store: %v", err)
	}
	_ = st.Close()

	writeMigration(t, migrationsDir, "0002_add_notes.up.sql", `
ALTER TABLE test_entries ADD COLUMN note TEXT;
`)

	_, err = store.Open(ctx, store.OpenConfig{
		Path:          dbPath,
		MigrationsDir: migrationsDir,
		AutoMigrate:   false,
	})
	if err == nil {
		t.Fatal("expected schema behind error")
	}
	if !strings.Contains(err.Error(), "sqlite schema behind") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeMigration(t *testing.T, dir, name, body string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir migrations dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(body)+"\n"), 0o644); err != nil {
		t.Fatalf("write migration: %v", err)
	}
}
