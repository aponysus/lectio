package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) CreateInquiry(ctx context.Context, input model.InquiryInput) (model.Inquiry, error) {
	id, err := newID()
	if err != nil {
		return model.Inquiry{}, err
	}

	_, err = s.db.ExecContext(ctx, `
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
		return model.Inquiry{}, err
	}

	return s.GetInquiry(ctx, id)
}

func (s *Store) ListInquiries(ctx context.Context, filters model.InquiryFilters) ([]model.Inquiry, error) {
	args := []any{}
	var where []string

	if !filters.IncludeArchived {
		where = append(where, "i.archived_at IS NULL")
	}
	if filters.Query != "" {
		where = append(where, "(LOWER(i.title) LIKE ? OR LOWER(i.question) LIKE ?)")
		pattern := "%" + strings.ToLower(filters.Query) + "%"
		args = append(args, pattern, pattern)
	}
	if filters.Status != "" {
		where = append(where, "i.status = ?")
		args = append(args, filters.Status)
	}

	query := inquirySelectQuery
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += `
		ORDER BY
			CASE i.status
				WHEN 'ACTIVE' THEN 0
				WHEN 'DORMANT' THEN 1
				WHEN 'SYNTHESIZED' THEN 2
				ELSE 3
			END,
			COALESCE(stats.latest_activity, i.updated_at) DESC,
			LOWER(i.title) ASC
		LIMIT ?
	`
	args = append(args, filters.Limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inquiries := []model.Inquiry{}
	for rows.Next() {
		inquiry, err := scanInquiry(rows)
		if err != nil {
			return nil, err
		}
		inquiries = append(inquiries, inquiry)
	}

	return inquiries, rows.Err()
}

func (s *Store) GetInquiry(ctx context.Context, id string) (model.Inquiry, error) {
	row := s.db.QueryRowContext(ctx,
		inquirySelectQuery+" WHERE i.id = ? AND i.archived_at IS NULL",
		id,
	)

	inquiry, err := scanInquiry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Inquiry{}, ErrNotFound
		}
		return model.Inquiry{}, err
	}

	return inquiry, nil
}

func (s *Store) UpdateInquiry(ctx context.Context, id string, input model.InquiryInput) (model.Inquiry, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE inquiries
		SET
			title = ?,
			question = ?,
			status = ?,
			why_it_matters = ?,
			current_view = ?,
			open_tensions = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND archived_at IS NULL
	`,
		input.Title,
		input.Question,
		input.Status,
		nullString(input.WhyItMatters),
		nullString(input.CurrentView),
		nullString(input.OpenTensions),
		id,
	)
	if err != nil {
		return model.Inquiry{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Inquiry{}, err
	}
	if rowsAffected == 0 {
		return model.Inquiry{}, ErrNotFound
	}

	return s.GetInquiry(ctx, id)
}

func (s *Store) ArchiveInquiry(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE inquiries
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

func (s *Store) ListInquiryEngagements(ctx context.Context, inquiryID string, limit int) ([]model.Engagement, error) {
	if err := s.ensureInquiryExists(ctx, inquiryID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, `
		`+engagementSelectQuery+`
		JOIN engagement_inquiries ei ON ei.engagement_id = e.id
		WHERE ei.inquiry_id = ? AND e.archived_at IS NULL
		ORDER BY datetime(e.engaged_at) DESC, e.updated_at DESC, e.id DESC
		LIMIT ?
	`, inquiryID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	engagements := []model.Engagement{}
	for rows.Next() {
		engagement, err := scanEngagement(rows)
		if err != nil {
			return nil, err
		}
		engagements = append(engagements, engagement)
	}

	return engagements, rows.Err()
}

func (s *Store) ListEngagementInquiries(ctx context.Context, engagementID string) ([]model.InquirySummary, error) {
	if err := s.ensureEngagementExists(ctx, engagementID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			i.id,
			i.title,
			i.question,
			i.status
		FROM inquiries i
		JOIN engagement_inquiries ei ON ei.inquiry_id = i.id
		WHERE ei.engagement_id = ? AND i.archived_at IS NULL
		ORDER BY
			CASE i.status
				WHEN 'ACTIVE' THEN 0
				WHEN 'DORMANT' THEN 1
				WHEN 'SYNTHESIZED' THEN 2
				ELSE 3
			END,
			LOWER(i.title) ASC
	`, engagementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inquiries := []model.InquirySummary{}
	for rows.Next() {
		var inquiry model.InquirySummary
		if err := rows.Scan(&inquiry.ID, &inquiry.Title, &inquiry.Question, &inquiry.Status); err != nil {
			return nil, err
		}
		inquiries = append(inquiries, inquiry)
	}

	return inquiries, rows.Err()
}

func (s *Store) ReplaceEngagementInquiries(ctx context.Context, engagementID string, inquiryIDs []string) error {
	if err := s.ensureEngagementExists(ctx, engagementID); err != nil {
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

	if _, err := tx.ExecContext(ctx, `DELETE FROM engagement_inquiries WHERE engagement_id = ?`, engagementID); err != nil {
		return err
	}

	for _, inquiryID := range inquiryIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO engagement_inquiries (engagement_id, inquiry_id, created_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
		`, engagementID, inquiryID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) ensureInquiryExists(ctx context.Context, id string) error {
	row := s.db.QueryRowContext(ctx, `
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

func (s *Store) ensureEngagementExists(ctx context.Context, id string) error {
	row := s.db.QueryRowContext(ctx, `
		SELECT 1
		FROM engagements
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

const inquirySelectQuery = `
	SELECT
		i.id,
		i.title,
		i.question,
		i.status,
		i.why_it_matters,
		i.current_view,
		i.open_tensions,
		strftime('%Y-%m-%dT%H:%M:%SZ', i.created_at) AS created_at,
		strftime('%Y-%m-%dT%H:%M:%SZ', i.updated_at) AS updated_at,
		CASE
			WHEN i.archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', i.archived_at)
		END AS archived_at,
		COALESCE(stats.engagement_count, 0) AS engagement_count,
		0 AS claim_count,
		CASE
			WHEN stats.latest_activity IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', stats.latest_activity)
		END AS latest_activity
	FROM inquiries i
	LEFT JOIN (
		SELECT
			ei.inquiry_id,
			COUNT(e.id) AS engagement_count,
			MAX(datetime(e.engaged_at)) AS latest_activity
		FROM engagement_inquiries ei
		JOIN engagements e ON e.id = ei.engagement_id AND e.archived_at IS NULL
		GROUP BY ei.inquiry_id
	) stats ON stats.inquiry_id = i.id
`

type inquiryScanner interface {
	Scan(dest ...any) error
}

func scanInquiry(scanner inquiryScanner) (model.Inquiry, error) {
	var inquiry model.Inquiry
	var whyItMatters sql.NullString
	var currentView sql.NullString
	var openTensions sql.NullString
	var archivedAt sql.NullString
	var latestActivity sql.NullString

	if err := scanner.Scan(
		&inquiry.ID,
		&inquiry.Title,
		&inquiry.Question,
		&inquiry.Status,
		&whyItMatters,
		&currentView,
		&openTensions,
		&inquiry.CreatedAt,
		&inquiry.UpdatedAt,
		&archivedAt,
		&inquiry.EngagementCount,
		&inquiry.ClaimCount,
		&latestActivity,
	); err != nil {
		return model.Inquiry{}, err
	}

	inquiry.WhyItMatters = stringPointer(whyItMatters)
	inquiry.CurrentView = stringPointer(currentView)
	inquiry.OpenTensions = stringPointer(openTensions)
	inquiry.ArchivedAt = stringPointer(archivedAt)
	inquiry.LatestActivity = stringPointer(latestActivity)

	return inquiry, nil
}
