package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) CreateSynthesis(ctx context.Context, input model.SynthesisInput) (model.Synthesis, error) {
	if err := s.ensureInquiryExists(ctx, input.InquiryID); err != nil {
		return model.Synthesis{}, err
	}

	id, err := newID()
	if err != nil {
		return model.Synthesis{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Synthesis{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO syntheses (
			id,
			title,
			body,
			type,
			inquiry_id,
			notes,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		id,
		input.Title,
		input.Body,
		input.Type,
		input.InquiryID,
		nullString(input.Notes),
	); err != nil {
		return model.Synthesis{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Synthesis{}, err
	}

	return s.GetSynthesis(ctx, id)
}

func (s *Store) ListSyntheses(ctx context.Context, filters model.SynthesisFilters) ([]model.Synthesis, error) {
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	if filters.InquiryID != "" {
		if err := s.ensureInquiryExists(ctx, filters.InquiryID); err != nil {
			return nil, err
		}
	}

	args := []any{}
	where := []string{"sy.archived_at IS NULL"}

	if filters.InquiryID != "" {
		where = append(where, "sy.inquiry_id = ?")
		args = append(args, filters.InquiryID)
	}

	query := synthesisSelectQuery + " WHERE " + strings.Join(where, " AND ") + `
		ORDER BY datetime(sy.created_at) DESC, datetime(sy.updated_at) DESC, sy.id DESC
		LIMIT ?
	`
	args = append(args, filters.Limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	syntheses := []model.Synthesis{}
	for rows.Next() {
		synthesis, err := scanSynthesis(rows)
		if err != nil {
			return nil, err
		}
		syntheses = append(syntheses, synthesis)
	}

	return syntheses, rows.Err()
}

func (s *Store) ListInquirySyntheses(ctx context.Context, inquiryID string, limit int) ([]model.Synthesis, error) {
	return s.ListSyntheses(ctx, model.SynthesisFilters{
		InquiryID: inquiryID,
		Limit:     limit,
	})
}

func (s *Store) GetSynthesis(ctx context.Context, id string) (model.Synthesis, error) {
	row := s.db.QueryRowContext(ctx,
		synthesisSelectQuery+" WHERE sy.id = ? AND sy.archived_at IS NULL",
		id,
	)

	synthesis, err := scanSynthesis(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Synthesis{}, ErrNotFound
		}
		return model.Synthesis{}, err
	}

	return synthesis, nil
}

func (s *Store) UpdateSynthesis(ctx context.Context, id string, input model.SynthesisInput) (model.Synthesis, error) {
	if err := s.ensureInquiryExists(ctx, input.InquiryID); err != nil {
		return model.Synthesis{}, err
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE syntheses
		SET
			title = ?,
			body = ?,
			type = ?,
			inquiry_id = ?,
			notes = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND archived_at IS NULL
	`,
		input.Title,
		input.Body,
		input.Type,
		input.InquiryID,
		nullString(input.Notes),
		id,
	)
	if err != nil {
		return model.Synthesis{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Synthesis{}, err
	}
	if rowsAffected == 0 {
		return model.Synthesis{}, ErrNotFound
	}

	return s.GetSynthesis(ctx, id)
}

func (s *Store) ArchiveSynthesis(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE syntheses
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

const synthesisSelectQuery = `
	SELECT
		sy.id,
		sy.title,
		sy.body,
		sy.type,
		sy.inquiry_id,
		sy.notes,
		strftime('%Y-%m-%dT%H:%M:%SZ', sy.created_at) AS created_at,
		strftime('%Y-%m-%dT%H:%M:%SZ', sy.updated_at) AS updated_at,
		CASE
			WHEN sy.archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', sy.archived_at)
		END AS archived_at,
		i.id,
		i.title,
		i.question,
		i.status
	FROM syntheses sy
	LEFT JOIN inquiries i ON i.id = sy.inquiry_id
`

type synthesisScanner interface {
	Scan(dest ...any) error
}

func scanSynthesis(scanner synthesisScanner) (model.Synthesis, error) {
	var synthesis model.Synthesis
	var notes sql.NullString
	var archivedAt sql.NullString
	var inquiryID sql.NullString
	var inquiryTitle sql.NullString
	var inquiryQuestion sql.NullString
	var inquiryStatus sql.NullString

	if err := scanner.Scan(
		&synthesis.ID,
		&synthesis.Title,
		&synthesis.Body,
		&synthesis.Type,
		&synthesis.InquiryID,
		&notes,
		&synthesis.CreatedAt,
		&synthesis.UpdatedAt,
		&archivedAt,
		&inquiryID,
		&inquiryTitle,
		&inquiryQuestion,
		&inquiryStatus,
	); err != nil {
		return model.Synthesis{}, err
	}

	synthesis.Notes = stringPointer(notes)
	synthesis.ArchivedAt = stringPointer(archivedAt)
	if inquiryID.Valid && inquiryTitle.Valid && inquiryQuestion.Valid && inquiryStatus.Valid {
		synthesis.Inquiry = &model.InquirySummary{
			ID:       inquiryID.String,
			Title:    inquiryTitle.String,
			Question: inquiryQuestion.String,
			Status:   inquiryStatus.String,
		}
	}

	return synthesis, nil
}
