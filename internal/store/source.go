package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

var ErrNotFound = errors.New("not found")

func (s *Store) CreateSource(ctx context.Context, input model.SourceInput) (model.Source, error) {
	id, err := newID()
	if err != nil {
		return model.Source{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO sources (
			id,
			title,
			medium,
			creator,
			year,
			original_language,
			culture_or_context,
			notes,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		id,
		input.Title,
		input.Medium,
		nullString(input.Creator),
		nullInt(input.Year),
		nullString(input.OriginalLanguage),
		nullString(input.CultureOrContext),
		nullString(input.Notes),
	)
	if err != nil {
		return model.Source{}, err
	}

	return s.GetSource(ctx, id)
}

func (s *Store) ListSources(ctx context.Context, filters model.SourceFilters) ([]model.Source, error) {
	args := []any{}
	var where []string

	if !filters.IncludeArchived {
		where = append(where, "archived_at IS NULL")
	}
	if filters.Query != "" {
		where = append(where, "(LOWER(title) LIKE ? OR LOWER(COALESCE(creator, '')) LIKE ?)")
		pattern := "%" + strings.ToLower(filters.Query) + "%"
		args = append(args, pattern, pattern)
	}
	if filters.Medium != "" {
		where = append(where, "medium = ?")
		args = append(args, filters.Medium)
	}
	if filters.OriginalLanguage != "" {
		where = append(where, "LOWER(COALESCE(original_language, '')) = ?")
		args = append(args, strings.ToLower(filters.OriginalLanguage))
	}

	query := sourceSelectQuery
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	switch filters.Sort {
	case model.SourceSortTitle:
		query += " ORDER BY LOWER(title) ASC, updated_at DESC"
	default:
		query += " ORDER BY updated_at DESC, LOWER(title) ASC"
	}
	query += " LIMIT ?"
	args = append(args, filters.Limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []model.Source{}
	for rows.Next() {
		source, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}

	return sources, rows.Err()
}

func (s *Store) GetSource(ctx context.Context, id string) (model.Source, error) {
	row := s.db.QueryRowContext(ctx,
		sourceSelectQuery+" WHERE id = ? AND archived_at IS NULL",
		id,
	)

	source, err := scanSource(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Source{}, ErrNotFound
		}
		return model.Source{}, err
	}

	return source, nil
}

func (s *Store) UpdateSource(ctx context.Context, id string, input model.SourceInput) (model.Source, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE sources
		SET
			title = ?,
			medium = ?,
			creator = ?,
			year = ?,
			original_language = ?,
			culture_or_context = ?,
			notes = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND archived_at IS NULL
	`,
		input.Title,
		input.Medium,
		nullString(input.Creator),
		nullInt(input.Year),
		nullString(input.OriginalLanguage),
		nullString(input.CultureOrContext),
		nullString(input.Notes),
		id,
	)
	if err != nil {
		return model.Source{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Source{}, err
	}
	if rowsAffected == 0 {
		return model.Source{}, ErrNotFound
	}

	return s.GetSource(ctx, id)
}

func (s *Store) ArchiveSource(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE sources
		SET
			archived_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND archived_at IS NULL
	`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

const sourceSelectQuery = `
	SELECT
		id,
		title,
		medium,
		creator,
		year,
		original_language,
		culture_or_context,
		notes,
		strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at,
		strftime('%Y-%m-%dT%H:%M:%SZ', updated_at) AS updated_at,
		CASE
			WHEN archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', archived_at)
		END AS archived_at
	FROM sources
`

type sourceScanner interface {
	Scan(dest ...any) error
}

func scanSource(scanner sourceScanner) (model.Source, error) {
	var source model.Source
	var creator sql.NullString
	var year sql.NullInt64
	var originalLanguage sql.NullString
	var cultureOrContext sql.NullString
	var notes sql.NullString
	var archivedAt sql.NullString

	if err := scanner.Scan(
		&source.ID,
		&source.Title,
		&source.Medium,
		&creator,
		&year,
		&originalLanguage,
		&cultureOrContext,
		&notes,
		&source.CreatedAt,
		&source.UpdatedAt,
		&archivedAt,
	); err != nil {
		return model.Source{}, err
	}

	source.Creator = stringPointer(creator)
	source.Year = intPointer(year)
	source.OriginalLanguage = stringPointer(originalLanguage)
	source.CultureOrContext = stringPointer(cultureOrContext)
	source.Notes = stringPointer(notes)
	source.ArchivedAt = stringPointer(archivedAt)

	return source, nil
}

func newID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func intPointer(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int64)
	return &v
}
