package store_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/aponysus/lectio/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func TestEnsureBootstrapUserCreatesHashedUserOnce(t *testing.T) {
	t.Parallel()

	st := openTestStore(t)
	defer st.Close()

	ctx := context.Background()
	user, err := st.Users().EnsureBootstrapUser(ctx, "reader@example.com", "sufficiently-secret")
	if err != nil {
		t.Fatalf("ensure bootstrap user: %v", err)
	}
	if user.Email != "reader@example.com" {
		t.Fatalf("unexpected email: %q", user.Email)
	}
	if user.PasswordHash == "sufficiently-secret" {
		t.Fatal("password stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("sufficiently-secret")); err != nil {
		t.Fatalf("password hash mismatch: %v", err)
	}

	again, err := st.Users().EnsureBootstrapUser(ctx, "reader@example.com", "new-password-ignored")
	if err != nil {
		t.Fatalf("ensure bootstrap user second time: %v", err)
	}
	if again.ID != user.ID {
		t.Fatalf("expected same bootstrap user, got %d and %d", user.ID, again.ID)
	}
}

func TestSessionRepositoryLifecycle(t *testing.T) {
	t.Parallel()

	st := openTestStore(t)
	defer st.Close()

	ctx := context.Background()
	user, err := st.Users().EnsureBootstrapUser(ctx, "reader@example.com", "sufficiently-secret")
	if err != nil {
		t.Fatalf("ensure bootstrap user: %v", err)
	}

	now := time.Now().UTC().Round(time.Second)
	session := store.Session{
		ID:         "session-token",
		UserID:     user.ID,
		ExpiresAt:  now.Add(14 * 24 * time.Hour),
		LastSeenAt: now,
	}
	if err := st.Sessions().Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	got, err := st.Sessions().GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.UserID != user.ID {
		t.Fatalf("unexpected user id: %d", got.UserID)
	}

	nextSeen := now.Add(5 * time.Minute)
	if err := st.Sessions().Touch(ctx, session.ID, nextSeen); err != nil {
		t.Fatalf("touch session: %v", err)
	}

	got, err = st.Sessions().GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("get touched session: %v", err)
	}
	if !got.LastSeenAt.Equal(nextSeen) {
		t.Fatalf("expected last seen %s, got %s", nextSeen, got.LastSeenAt)
	}

	if err := st.Sessions().Delete(ctx, session.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	_, err = st.Sessions().GetByID(ctx, session.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func openTestStore(t *testing.T) *store.Store {
	t.Helper()

	ctx := context.Background()
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "migrations")
	writeMigration(t, migrationsDir, "0001_init.up.sql", `
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)

	st, err := store.Open(ctx, store.OpenConfig{
		Path:          filepath.Join(dir, "lectio.db"),
		MigrationsDir: migrationsDir,
		AutoMigrate:   true,
	})
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	return st
}
