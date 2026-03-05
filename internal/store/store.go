package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultBusyTimeout = 5 * time.Second
)

var ErrNotFound = errors.New("store: not found")

type OpenConfig struct {
	Path          string
	MigrationsDir string
	AutoMigrate   bool
}

type MigrationState struct {
	CurrentVersion int64
	LatestVersion  int64
	Dirty          bool
}

type Store struct {
	db              *sql.DB
	latestMigration int64
	entries         EntryRepository
	sources         SourceRepository
	tags            TagRepository
	threads         ThreadRepository
	users           UserRepository
	sessions        SessionRepository
	resonances      ResonanceRepository
}

func Open(ctx context.Context, cfg OpenConfig) (*Store, error) {
	if strings.TrimSpace(cfg.Path) == "" {
		return nil, errors.New("store: sqlite path is required")
	}

	if err := ensureParentDir(cfg.Path); err != nil {
		return nil, fmt.Errorf("prepare sqlite path: %w", err)
	}

	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := applyPragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite pragmas: %w", err)
	}

	latest, err := latestMigrationVersion(cfg.MigrationsDir)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := ensureMigrations(ctx, db, cfg.MigrationsDir, cfg.AutoMigrate); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{
		db:              db,
		latestMigration: latest,
		users:           newUserRepository(db),
		sessions:        newSessionRepository(db),
	}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("store: database not initialized")
	}
	return s.db.PingContext(ctx)
}

func (s *Store) MigrationState(ctx context.Context) (MigrationState, error) {
	if s == nil || s.db == nil {
		return MigrationState{}, errors.New("store: database not initialized")
	}

	current, dirty, err := currentMigrationVersion(ctx, s.db)
	if err != nil {
		return MigrationState{}, err
	}

	return MigrationState{
		CurrentVersion: current,
		LatestVersion:  s.latestMigration,
		Dirty:          dirty,
	}, nil
}

func (s *Store) Entries() EntryRepository {
	return s.entries
}

func (s *Store) Sources() SourceRepository {
	return s.sources
}

func (s *Store) Tags() TagRepository {
	return s.tags
}

func (s *Store) Threads() ThreadRepository {
	return s.threads
}

func (s *Store) Users() UserRepository {
	return s.users
}

func (s *Store) Sessions() SessionRepository {
	return s.sessions
}

func (s *Store) Resonances() ResonanceRepository {
	return s.resonances
}

func applyPragmas(ctx context.Context, db *sql.DB) error {
	statements := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		fmt.Sprintf("PRAGMA busy_timeout = %d", defaultBusyTimeout.Milliseconds()),
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return db.PingContext(ctx)
}

func ensureParentDir(path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	return os.MkdirAll(dir, 0o755)
}

func ensureMigrations(ctx context.Context, db *sql.DB, dir string, autoMigrate bool) error {
	migrations, err := loadMigrations(dir)
	if err != nil {
		return err
	}

	current, dirty, err := currentMigrationVersion(ctx, db)
	if err != nil {
		return err
	}
	if dirty {
		return errors.New("sqlite schema is marked dirty; manual intervention required before startup")
	}

	latest := int64(0)
	if len(migrations) > 0 {
		latest = migrations[len(migrations)-1].Version
	}

	if current > latest {
		return fmt.Errorf("sqlite schema version %d is ahead of known migrations %d", current, latest)
	}
	if current == latest {
		return nil
	}
	if !autoMigrate {
		return fmt.Errorf("sqlite schema behind: current=%d latest=%d; run make migrate-up or set LECTIO_AUTO_MIGRATE=true", current, latest)
	}

	for _, migration := range migrations {
		if migration.Version <= current {
			continue
		}
		if err := applyMigration(ctx, db, migration); err != nil {
			return err
		}
	}

	return nil
}

type migrationFile struct {
	Version int64
	Name    string
	Body    string
}

func latestMigrationVersion(dir string) (int64, error) {
	migrations, err := loadMigrations(dir)
	if err != nil {
		return 0, err
	}
	if len(migrations) == 0 {
		return 0, nil
	}
	return migrations[len(migrations)-1].Version, nil
}

func loadMigrations(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	migrations := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		version, err := parseMigrationVersion(name)
		if err != nil {
			return nil, err
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}

		migrations = append(migrations, migrationFile{
			Version: version,
			Name:    name,
			Body:    string(body),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationVersion(name string) (int64, error) {
	prefix, _, found := strings.Cut(name, "_")
	if !found {
		return 0, fmt.Errorf("invalid migration name %q", name)
	}

	version, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid migration version in %q: %w", name, err)
	}
	if version <= 0 {
		return 0, fmt.Errorf("migration version must be positive in %q", name)
	}

	return version, nil
}

func currentMigrationVersion(ctx context.Context, db *sql.DB) (int64, bool, error) {
	var tableExists int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'").Scan(&tableExists); err != nil {
		return 0, false, fmt.Errorf("query schema_migrations table: %w", err)
	}
	if tableExists == 0 {
		return 0, false, nil
	}

	rows, err := db.QueryContext(ctx, "SELECT version, dirty FROM schema_migrations")
	if err != nil {
		return 0, false, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()

	var (
		version int64
		dirty   bool
		count   int
	)
	for rows.Next() {
		count++
		if err := rows.Scan(&version, &dirty); err != nil {
			return 0, false, fmt.Errorf("scan schema_migrations: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, false, fmt.Errorf("iterate schema_migrations: %w", err)
	}
	if count == 0 {
		return 0, false, nil
	}
	if count > 1 {
		return 0, false, errors.New("schema_migrations has multiple rows")
	}

	return version, dirty, nil
}

func applyMigration(ctx context.Context, db *sql.DB, migration migrationFile) error {
	if err := setMigrationVersion(ctx, db, migration.Version, true); err != nil {
		return fmt.Errorf("mark migration %s dirty: %w", migration.Name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.Name, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, migration.Body); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.Name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Name, err)
	}

	if err := setMigrationVersion(ctx, db, migration.Version, false); err != nil {
		return fmt.Errorf("finalize migration %s: %w", migration.Name, err)
	}

	return nil
}

type migrationVersionWriter interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func setMigrationVersion(ctx context.Context, writer migrationVersionWriter, version int64, dirty bool) error {
	if _, err := writer.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER NOT NULL,
			dirty BOOLEAN NOT NULL
		)
	`); err != nil {
		return err
	}

	if _, err := writer.ExecContext(ctx, "DELETE FROM schema_migrations"); err != nil {
		return err
	}

	_, err := writer.ExecContext(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES (?, ?)", version, dirty)
	return err
}
