package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) CreateLanguageNote(ctx context.Context, input model.LanguageNoteInput) (model.LanguageNote, error) {
	if err := s.ensureEngagementExists(ctx, input.EngagementID); err != nil {
		return model.LanguageNote{}, err
	}

	id, err := newID()
	if err != nil {
		return model.LanguageNote{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO language_notes (
			id,
			engagement_id,
			term,
			language,
			note_type,
			content,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		id,
		input.EngagementID,
		nullString(input.Term),
		nullString(input.Language),
		nullString(input.NoteType),
		input.Content,
	)
	if err != nil {
		return model.LanguageNote{}, err
	}

	return s.GetLanguageNote(ctx, id)
}

func (s *Store) GetLanguageNote(ctx context.Context, id string) (model.LanguageNote, error) {
	row := s.db.QueryRowContext(ctx,
		languageNoteSelectQuery+" WHERE ln.id = ? AND ln.archived_at IS NULL",
		id,
	)

	note, err := scanLanguageNote(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LanguageNote{}, ErrNotFound
		}
		return model.LanguageNote{}, err
	}

	return note, nil
}

func (s *Store) UpdateLanguageNote(ctx context.Context, id string, input model.LanguageNoteInput) (model.LanguageNote, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE language_notes
		SET
			term = ?,
			language = ?,
			note_type = ?,
			content = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND archived_at IS NULL
	`,
		nullString(input.Term),
		nullString(input.Language),
		nullString(input.NoteType),
		input.Content,
		id,
	)
	if err != nil {
		return model.LanguageNote{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.LanguageNote{}, err
	}
	if rowsAffected == 0 {
		return model.LanguageNote{}, ErrNotFound
	}

	return s.GetLanguageNote(ctx, id)
}

func (s *Store) ListEngagementLanguageNotes(ctx context.Context, engagementID string) ([]model.LanguageNote, error) {
	if err := s.ensureEngagementExists(ctx, engagementID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		`+languageNoteSelectQuery+`
		WHERE ln.engagement_id = ? AND ln.archived_at IS NULL
		ORDER BY datetime(ln.created_at) DESC, datetime(ln.updated_at) DESC, ln.id DESC
	`, engagementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []model.LanguageNote{}
	for rows.Next() {
		note, err := scanLanguageNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

func (s *Store) ArchiveLanguageNote(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE language_notes
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

const languageNoteSelectQuery = `
	SELECT
		ln.id,
		ln.engagement_id,
		ln.term,
		ln.language,
		ln.note_type,
		ln.content,
		strftime('%Y-%m-%dT%H:%M:%SZ', ln.created_at) AS created_at,
		strftime('%Y-%m-%dT%H:%M:%SZ', ln.updated_at) AS updated_at,
		CASE
			WHEN ln.archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', ln.archived_at)
		END AS archived_at
	FROM language_notes ln
`

type languageNoteScanner interface {
	Scan(dest ...any) error
}

func scanLanguageNote(scanner languageNoteScanner) (model.LanguageNote, error) {
	var note model.LanguageNote
	var term sql.NullString
	var language sql.NullString
	var noteType sql.NullString
	var archivedAt sql.NullString

	if err := scanner.Scan(
		&note.ID,
		&note.EngagementID,
		&term,
		&language,
		&noteType,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
		&archivedAt,
	); err != nil {
		return model.LanguageNote{}, err
	}

	note.Term = stringPointer(term)
	note.Language = stringPointer(language)
	note.NoteType = stringPointer(noteType)
	note.ArchivedAt = stringPointer(archivedAt)

	return note, nil
}
