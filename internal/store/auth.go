package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type sqliteUserRepository struct {
	db *sql.DB
}

type sqliteSessionRepository struct {
	db *sql.DB
}

func newUserRepository(db *sql.DB) UserRepository {
	return &sqliteUserRepository{db: db}
}

func newSessionRepository(db *sql.DB) SessionRepository {
	return &sqliteSessionRepository{db: db}
}

func (r *sqliteUserRepository) EnsureBootstrapUser(ctx context.Context, email, password string) (User, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))

	existing, err := r.GetFirst(ctx)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return User{}, fmt.Errorf("hash bootstrap password: %w", err)
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES (?, ?)
	`, nullableString(normalizedEmail), string(passwordHash))
	if err != nil {
		return User{}, fmt.Errorf("insert bootstrap user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("read bootstrap user id: %w", err)
	}

	return r.getByID(ctx, id)
}

func (r *sqliteUserRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" {
		return User{}, ErrNotFound
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(email, ''), password_hash, created_at
		FROM users
		WHERE lower(email) = lower(?)
	`, normalizedEmail)

	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *sqliteUserRepository) GetFirst(ctx context.Context) (User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(email, ''), password_hash, created_at
		FROM users
		ORDER BY id ASC
		LIMIT 1
	`)

	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get first user: %w", err)
	}

	return user, nil
}

func (r *sqliteUserRepository) getByID(ctx context.Context, id int64) (User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(email, ''), password_hash, created_at
		FROM users
		WHERE id = ?
	`, id)

	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *sqliteSessionRepository) Create(ctx context.Context, session Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, expires_at, last_seen_at)
		VALUES (?, ?, ?, ?)
	`, session.ID, session.UserID, session.ExpiresAt.UTC(), session.LastSeenAt.UTC())
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (r *sqliteSessionRepository) GetByID(ctx context.Context, id string) (Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, created_at, expires_at, last_seen_at
		FROM sessions
		WHERE id = ?
	`, id)

	var session Session
	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastSeenAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	return session, nil
}

func (r *sqliteSessionRepository) Touch(ctx context.Context, id string, lastSeen time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE sessions
		SET last_seen_at = ?
		WHERE id = ?
	`, lastSeen.UTC(), id)
	if err != nil {
		return fmt.Errorf("touch session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("touch session rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *sqliteSessionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete session rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func scanUser(scanner interface{ Scan(...any) error }) (User, error) {
	var user User
	err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
