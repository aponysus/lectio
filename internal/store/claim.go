package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) CreateClaim(ctx context.Context, input model.ClaimInput, inquiryIDs []string) (model.Claim, error) {
	if input.OriginEngagementID != "" {
		if err := s.ensureEngagementExists(ctx, input.OriginEngagementID); err != nil {
			return model.Claim{}, err
		}
	}
	for _, inquiryID := range inquiryIDs {
		if err := s.ensureInquiryExists(ctx, inquiryID); err != nil {
			return model.Claim{}, err
		}
	}

	id, err := newID()
	if err != nil {
		return model.Claim{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Claim{}, err
	}
	defer tx.Rollback()

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
		nullString(input.OriginEngagementID),
		nullString(input.Notes),
	)
	if err != nil {
		return model.Claim{}, err
	}

	for _, inquiryID := range inquiryIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO claim_inquiries (claim_id, inquiry_id, created_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
		`, id, inquiryID); err != nil {
			return model.Claim{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Claim{}, err
	}

	return s.GetClaim(ctx, id)
}

func (s *Store) GetClaim(ctx context.Context, id string) (model.Claim, error) {
	row := s.db.QueryRowContext(ctx,
		claimSelectQuery+" WHERE c.id = ? AND c.archived_at IS NULL",
		id,
	)

	claim, err := scanClaim(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Claim{}, ErrNotFound
		}
		return model.Claim{}, err
	}

	return claim, nil
}

func (s *Store) UpdateClaim(ctx context.Context, id string, input model.ClaimInput) (model.Claim, error) {
	if input.OriginEngagementID != "" {
		if err := s.ensureEngagementExists(ctx, input.OriginEngagementID); err != nil {
			return model.Claim{}, err
		}
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE claims
		SET
			text = ?,
			claim_type = ?,
			confidence = ?,
			status = ?,
			origin_engagement_id = ?,
			notes = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND archived_at IS NULL
	`,
		input.Text,
		input.ClaimType,
		nullInt(input.Confidence),
		input.Status,
		nullString(input.OriginEngagementID),
		nullString(input.Notes),
		id,
	)
	if err != nil {
		return model.Claim{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Claim{}, err
	}
	if rowsAffected == 0 {
		return model.Claim{}, ErrNotFound
	}

	return s.GetClaim(ctx, id)
}

func (s *Store) ReplaceClaimInquiries(ctx context.Context, claimID string, inquiryIDs []string) error {
	if err := s.ensureClaimExists(ctx, claimID); err != nil {
		return err
	}

	for _, inquiryID := range inquiryIDs {
		if err := s.ensureInquiryExists(ctx, inquiryID); err != nil {
			return err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM claim_inquiries WHERE claim_id = ?`, claimID); err != nil {
		return err
	}

	for _, inquiryID := range inquiryIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO claim_inquiries (claim_id, inquiry_id, created_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
		`, claimID, inquiryID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) ListInquiryClaims(ctx context.Context, inquiryID string) ([]model.Claim, error) {
	if err := s.ensureInquiryExists(ctx, inquiryID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		`+claimSelectQuery+`
		JOIN claim_inquiries ci ON ci.claim_id = c.id
		WHERE ci.inquiry_id = ? AND c.archived_at IS NULL
		ORDER BY c.updated_at DESC, c.id DESC
	`, inquiryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := []model.Claim{}
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	return claims, rows.Err()
}

func (s *Store) ListEngagementClaims(ctx context.Context, engagementID string) ([]model.Claim, error) {
	if err := s.ensureEngagementExists(ctx, engagementID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		`+claimSelectQuery+`
		WHERE c.origin_engagement_id = ? AND c.archived_at IS NULL
		ORDER BY c.updated_at DESC, c.id DESC
	`, engagementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := []model.Claim{}
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	return claims, rows.Err()
}

func (s *Store) ArchiveClaim(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE claims
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

func (s *Store) ensureClaimExists(ctx context.Context, id string) error {
	row := s.db.QueryRowContext(ctx, `
		SELECT 1
		FROM claims
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

const claimSelectQuery = `
	SELECT
		c.id,
		c.text,
		c.claim_type,
		c.confidence,
		c.status,
		c.origin_engagement_id,
		c.notes,
		strftime('%Y-%m-%dT%H:%M:%SZ', c.created_at) AS created_at,
		strftime('%Y-%m-%dT%H:%M:%SZ', c.updated_at) AS updated_at,
		CASE
			WHEN c.archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', c.archived_at)
		END AS archived_at,
		e.id,
		e.source_id,
		s.title,
		s.medium,
		e.portion_label
	FROM claims c
	LEFT JOIN engagements e ON e.id = c.origin_engagement_id AND e.archived_at IS NULL
	LEFT JOIN sources s ON s.id = e.source_id
`

type claimScanner interface {
	Scan(dest ...any) error
}

func scanClaim(scanner claimScanner) (model.Claim, error) {
	var claim model.Claim
	var confidence sql.NullInt64
	var originEngagementID sql.NullString
	var notes sql.NullString
	var archivedAt sql.NullString
	var originID sql.NullString
	var originSourceID sql.NullString
	var originSourceTitle sql.NullString
	var originSourceMedium sql.NullString
	var originPortionLabel sql.NullString

	if err := scanner.Scan(
		&claim.ID,
		&claim.Text,
		&claim.ClaimType,
		&confidence,
		&claim.Status,
		&originEngagementID,
		&notes,
		&claim.CreatedAt,
		&claim.UpdatedAt,
		&archivedAt,
		&originID,
		&originSourceID,
		&originSourceTitle,
		&originSourceMedium,
		&originPortionLabel,
	); err != nil {
		return model.Claim{}, err
	}

	claim.Confidence = intPointer(confidence)
	claim.OriginEngagementID = stringPointer(originEngagementID)
	claim.Notes = stringPointer(notes)
	claim.ArchivedAt = stringPointer(archivedAt)

	if originID.Valid && originSourceID.Valid && originSourceTitle.Valid && originSourceMedium.Valid {
		claim.Origin = &model.ClaimOrigin{
			EngagementID: originID.String,
			SourceID:     originSourceID.String,
			SourceTitle:  originSourceTitle.String,
			SourceMedium: originSourceMedium.String,
			PortionLabel: stringPointer(originPortionLabel),
		}
	}

	return claim, nil
}
