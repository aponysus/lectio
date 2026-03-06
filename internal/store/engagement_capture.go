package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) CreateEngagementCapture(ctx context.Context, input model.EngagementCaptureInput) (model.Engagement, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Engagement{}, err
	}
	defer tx.Rollback()

	if err := ensureSourceExistsTx(ctx, tx, input.Engagement.SourceID); err != nil {
		return model.Engagement{}, err
	}

	inquiryIDs := append([]string{}, input.InquiryIDs...)
	for _, inquiry := range input.InlineInquiries {
		inquiryID, err := createInquiryTx(ctx, tx, inquiry)
		if err != nil {
			return model.Engagement{}, err
		}
		inquiryIDs = append(inquiryIDs, inquiryID)
	}

	for _, inquiryID := range inquiryIDs {
		if err := ensureInquiryExistsTx(ctx, tx, inquiryID); err != nil {
			return model.Engagement{}, err
		}
	}

	engagementID, err := createEngagementTx(ctx, tx, input.Engagement)
	if err != nil {
		return model.Engagement{}, err
	}

	for _, inquiryID := range inquiryIDs {
		if err := linkEngagementInquiryTx(ctx, tx, engagementID, inquiryID); err != nil {
			return model.Engagement{}, err
		}
	}

	for _, claim := range input.Claims {
		claimID, err := createClaimTx(ctx, tx, claim, engagementID)
		if err != nil {
			return model.Engagement{}, err
		}

		for _, inquiryID := range inquiryIDs {
			if err := linkClaimInquiryTx(ctx, tx, claimID, inquiryID); err != nil {
				return model.Engagement{}, err
			}
		}
	}

	for _, note := range input.LanguageNotes {
		if err := createLanguageNoteTx(ctx, tx, note, engagementID); err != nil {
			return model.Engagement{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Engagement{}, err
	}

	return s.GetEngagement(ctx, engagementID)
}

func ensureSourceExistsTx(ctx context.Context, tx *sql.Tx, id string) error {
	row := tx.QueryRowContext(ctx, `
		SELECT 1
		FROM sources
		WHERE id = ? AND archived_at IS NULL
	`, id)

	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func ensureInquiryExistsTx(ctx context.Context, tx *sql.Tx, id string) error {
	row := tx.QueryRowContext(ctx, `
		SELECT 1
		FROM inquiries
		WHERE id = ? AND archived_at IS NULL
	`, id)

	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func createInquiryTx(ctx context.Context, tx *sql.Tx, input model.InquiryInput) (string, error) {
	id, err := newID()
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO inquiries (
			id,
			title,
			question,
			status,
			why_it_matters,
			current_view,
			open_tensions,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		id,
		input.Title,
		input.Question,
		input.Status,
		nullString(input.WhyItMatters),
		nullString(input.CurrentView),
		nullString(input.OpenTensions),
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func createEngagementTx(ctx context.Context, tx *sql.Tx, input model.EngagementInput) (string, error) {
	id, err := newID()
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO engagements (
			id,
			source_id,
			engaged_at,
			portion_label,
			reflection,
			why_it_matters,
			source_language,
			reflection_language,
			access_mode,
			revisit_priority,
			is_reread_or_rewatch,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		id,
		input.SourceID,
		input.EngagedAt,
		nullString(input.PortionLabel),
		input.Reflection,
		nullString(input.WhyItMatters),
		nullString(input.SourceLanguage),
		nullString(input.ReflectionLanguage),
		nullString(input.AccessMode),
		nullInt(input.RevisitPriority),
		input.IsRereadOrRewatch,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func linkEngagementInquiryTx(ctx context.Context, tx *sql.Tx, engagementID, inquiryID string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO engagement_inquiries (engagement_id, inquiry_id, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`, engagementID, inquiryID)
	return err
}

func createClaimTx(ctx context.Context, tx *sql.Tx, input model.ClaimInput, engagementID string) (string, error) {
	id, err := newID()
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO claims (
			id,
			text,
			claim_type,
			confidence,
			status,
			origin_engagement_id,
			notes,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		id,
		input.Text,
		input.ClaimType,
		nullInt(input.Confidence),
		input.Status,
		engagementID,
		nullString(input.Notes),
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func linkClaimInquiryTx(ctx context.Context, tx *sql.Tx, claimID, inquiryID string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO claim_inquiries (claim_id, inquiry_id, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`, claimID, inquiryID)
	return err
}

func createLanguageNoteTx(ctx context.Context, tx *sql.Tx, input model.LanguageNoteInput, engagementID string) error {
	id, err := newID()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
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
		engagementID,
		nullString(input.Term),
		nullString(input.Language),
		nullString(input.NoteType),
		input.Content,
	)

	return err
}
