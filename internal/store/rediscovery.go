package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) ListRediscoveryItems(ctx context.Context, limit int) ([]model.RediscoveryItem, error) {
	if limit <= 0 {
		limit = 6
	}
	if limit > 24 {
		limit = 24
	}

	if err := s.syncRediscoveryItems(ctx); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			kind,
			target_type,
			target_id,
			reason,
			status,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at,
			strftime('%Y-%m-%dT%H:%M:%SZ', updated_at) AS updated_at
		FROM rediscovery_items
		WHERE status IN ('NEW', 'SEEN')
		ORDER BY
			CASE kind
				WHEN 'recent_reactivation' THEN 0
				WHEN 'unsynthesized_inquiry' THEN 1
				WHEN 'stale_tentative_claim' THEN 2
				ELSE 3
			END,
			datetime(created_at) DESC,
			id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.RediscoveryItem{}
	for rows.Next() {
		item, err := scanRediscoveryItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	filtered := make([]model.RediscoveryItem, 0, len(items))
	newIDs := []string{}
	for _, item := range items {
		if err := s.populateRediscoveryTarget(ctx, &item); err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}
			return nil, err
		}

		if item.Status == string(model.RediscoveryStatusNew) {
			newIDs = append(newIDs, item.ID)
			item.Status = string(model.RediscoveryStatusSeen)
		}

		filtered = append(filtered, item)
	}

	if len(newIDs) > 0 {
		if err := s.markRediscoveryItemsSeen(ctx, newIDs); err != nil {
			return nil, err
		}
	}

	return filtered, nil
}

func (s *Store) DismissRediscoveryItem(ctx context.Context, id string) error {
	return s.updateRediscoveryItemStatus(ctx, id, string(model.RediscoveryStatusDismissed))
}

func (s *Store) MarkRediscoveryItemActedOn(ctx context.Context, id string) error {
	return s.updateRediscoveryItemStatus(ctx, id, string(model.RediscoveryStatusActedOn))
}

func (s *Store) updateRediscoveryItemStatus(ctx context.Context, id, status string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE rediscovery_items
		SET
			status = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status IN ('NEW', 'SEEN')
	`, status, id)
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

func (s *Store) syncRediscoveryItems(ctx context.Context) error {
	if err := s.resolveInactiveRediscoveryItems(ctx); err != nil {
		return err
	}
	if err := s.generateRediscoveryItems(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Store) resolveInactiveRediscoveryItems(ctx context.Context) error {
	statements := []string{
		`
		UPDATE rediscovery_items
		SET
			status = 'ACTED_ON',
			updated_at = CURRENT_TIMESTAMP
		WHERE
			status IN ('NEW', 'SEEN')
			AND kind = 'stale_tentative_claim'
			AND NOT EXISTS (
				SELECT 1
				FROM claims c
				WHERE
					c.id = rediscovery_items.target_id
					AND c.archived_at IS NULL
					AND c.status = 'TENTATIVE'
					AND datetime(COALESCE(c.updated_at, c.created_at)) <= datetime('now', '-14 days')
			)
		`,
		`
		UPDATE rediscovery_items
		SET
			status = 'ACTED_ON',
			updated_at = CURRENT_TIMESTAMP
		WHERE
			status IN ('NEW', 'SEEN')
			AND kind = 'active_inquiry_old_engagement'
			AND NOT EXISTS (
				SELECT 1
				FROM engagements e
				JOIN engagement_inquiries ei ON ei.engagement_id = e.id
				JOIN inquiries i ON i.id = ei.inquiry_id
				WHERE
					e.id = rediscovery_items.target_id
					AND e.archived_at IS NULL
					AND i.archived_at IS NULL
					AND i.status = 'ACTIVE'
					AND datetime(e.engaged_at) <= datetime('now', '-30 days')
			)
		`,
		`
		UPDATE rediscovery_items
		SET
			status = 'ACTED_ON',
			updated_at = CURRENT_TIMESTAMP
		WHERE
			status IN ('NEW', 'SEEN')
			AND kind = 'unsynthesized_inquiry'
			AND NOT EXISTS (
				SELECT 1
				FROM inquiries i
				LEFT JOIN (
					SELECT ei.inquiry_id, COUNT(e.id) AS engagement_count
					FROM engagement_inquiries ei
					JOIN engagements e ON e.id = ei.engagement_id AND e.archived_at IS NULL
					GROUP BY ei.inquiry_id
				) engagement_stats ON engagement_stats.inquiry_id = i.id
				LEFT JOIN (
					SELECT inquiry_id, COUNT(*) AS synthesis_count
					FROM syntheses
					WHERE archived_at IS NULL
					GROUP BY inquiry_id
				) synthesis_stats ON synthesis_stats.inquiry_id = i.id
				WHERE
					i.id = rediscovery_items.target_id
					AND i.archived_at IS NULL
					AND i.status != 'ABANDONED'
					AND COALESCE(engagement_stats.engagement_count, 0) >= 4
					AND COALESCE(synthesis_stats.synthesis_count, 0) = 0
			)
		`,
		`
		UPDATE rediscovery_items
		SET
			status = 'ACTED_ON',
			updated_at = CURRENT_TIMESTAMP
		WHERE
			status IN ('NEW', 'SEEN')
			AND kind = 'recent_reactivation'
			AND NOT EXISTS (
				SELECT 1
				FROM inquiries i
				WHERE
					i.id = rediscovery_items.target_id
					AND i.archived_at IS NULL
					AND i.status = 'ACTIVE'
					AND i.reactivated_at IS NOT NULL
					AND datetime(i.reactivated_at) >= datetime('now', '-7 days')
			)
		`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) generateRediscoveryItems(ctx context.Context) error {
	if err := s.generateStaleTentativeClaimRediscoveryItems(ctx); err != nil {
		return err
	}
	if err := s.generateOldActiveInquiryEngagementRediscoveryItems(ctx); err != nil {
		return err
	}
	if err := s.generateUnsynthesizedInquiryRediscoveryItems(ctx); err != nil {
		return err
	}
	if err := s.generateRecentReactivationRediscoveryItems(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Store) generateStaleTentativeClaimRediscoveryItems(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id
		FROM claims c
		WHERE
			c.archived_at IS NULL
			AND c.status = 'TENTATIVE'
			AND datetime(COALESCE(c.updated_at, c.created_at)) <= datetime('now', '-14 days')
		ORDER BY datetime(COALESCE(c.updated_at, c.created_at)) ASC, c.id ASC
		LIMIT 12
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	claimIDs := []string{}
	for rows.Next() {
		var claimID string
		if err := rows.Scan(&claimID); err != nil {
			return err
		}
		claimIDs = append(claimIDs, claimID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, claimID := range claimIDs {
		if err := s.createRediscoveryItem(ctx,
			string(model.RediscoveryKindStaleTentativeClaim),
			string(model.RediscoveryTargetTypeClaim),
			claimID,
			"This tentative claim has been sitting for more than two weeks.",
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) generateOldActiveInquiryEngagementRediscoveryItems(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		WITH ranked_engagements AS (
			SELECT
				e.id,
				ei.inquiry_id,
				ROW_NUMBER() OVER (
					PARTITION BY ei.inquiry_id
					ORDER BY
						COALESCE(e.revisit_priority, 0) DESC,
						datetime(e.engaged_at) ASC,
						e.id DESC
				) AS row_num
			FROM engagements e
			JOIN engagement_inquiries ei ON ei.engagement_id = e.id
			JOIN inquiries i ON i.id = ei.inquiry_id
			WHERE
				e.archived_at IS NULL
				AND i.archived_at IS NULL
				AND i.status = 'ACTIVE'
				AND datetime(e.engaged_at) <= datetime('now', '-30 days')
		)
		SELECT id
		FROM ranked_engagements
		WHERE row_num = 1
		ORDER BY id ASC
		LIMIT 12
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	engagementIDs := []string{}
	for rows.Next() {
		var engagementID string
		if err := rows.Scan(&engagementID); err != nil {
			return err
		}
		engagementIDs = append(engagementIDs, engagementID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, engagementID := range engagementIDs {
		if err := s.createRediscoveryItem(ctx,
			string(model.RediscoveryKindActiveInquiryOldEntry),
			string(model.RediscoveryTargetTypeEngagement),
			engagementID,
			"An older engagement is still tied to an active inquiry.",
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) generateUnsynthesizedInquiryRediscoveryItems(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT i.id
		FROM inquiries i
		LEFT JOIN (
			SELECT ei.inquiry_id, COUNT(e.id) AS engagement_count
			FROM engagement_inquiries ei
			JOIN engagements e ON e.id = ei.engagement_id AND e.archived_at IS NULL
			GROUP BY ei.inquiry_id
		) engagement_stats ON engagement_stats.inquiry_id = i.id
		LEFT JOIN (
			SELECT inquiry_id, COUNT(*) AS synthesis_count
			FROM syntheses
			WHERE archived_at IS NULL
			GROUP BY inquiry_id
		) synthesis_stats ON synthesis_stats.inquiry_id = i.id
		WHERE
			i.archived_at IS NULL
			AND i.status != 'ABANDONED'
			AND COALESCE(engagement_stats.engagement_count, 0) >= 4
			AND COALESCE(synthesis_stats.synthesis_count, 0) = 0
		ORDER BY datetime(i.updated_at) DESC, i.id DESC
		LIMIT 12
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	inquiryIDs := []string{}
	for rows.Next() {
		var inquiryID string
		if err := rows.Scan(&inquiryID); err != nil {
			return err
		}
		inquiryIDs = append(inquiryIDs, inquiryID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, inquiryID := range inquiryIDs {
		if err := s.createRediscoveryItem(ctx,
			string(model.RediscoveryKindUnsynthesizedInquiry),
			string(model.RediscoveryTargetTypeInquiry),
			inquiryID,
			"This inquiry has enough material to compress into a synthesis.",
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) generateRecentReactivationRediscoveryItems(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT i.id
		FROM inquiries i
		WHERE
			i.archived_at IS NULL
			AND i.status = 'ACTIVE'
			AND i.reactivated_at IS NOT NULL
			AND datetime(i.reactivated_at) >= datetime('now', '-7 days')
		ORDER BY datetime(i.reactivated_at) DESC, i.id DESC
		LIMIT 12
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	inquiryIDs := []string{}
	for rows.Next() {
		var inquiryID string
		if err := rows.Scan(&inquiryID); err != nil {
			return err
		}
		inquiryIDs = append(inquiryIDs, inquiryID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, inquiryID := range inquiryIDs {
		if err := s.createRediscoveryItem(ctx,
			string(model.RediscoveryKindRecentReactivation),
			string(model.RediscoveryTargetTypeInquiry),
			inquiryID,
			"This inquiry was reactivated recently after dormancy.",
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) createRediscoveryItem(ctx context.Context, kind, targetType, targetID, reason string) error {
	id, err := newID()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO rediscovery_items (
			id,
			kind,
			target_type,
			target_id,
			reason,
			status,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, 'NEW', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, kind, targetType, targetID, reason)
	return err
}

func (s *Store) markRediscoveryItemsSeen(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	query, args := inClause("UPDATE rediscovery_items SET status = 'SEEN', updated_at = CURRENT_TIMESTAMP WHERE id IN (", ids)
	query += ") AND status = 'NEW'"
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) populateRediscoveryTarget(ctx context.Context, item *model.RediscoveryItem) error {
	switch item.TargetType {
	case string(model.RediscoveryTargetTypeClaim):
		claim, err := s.GetClaim(ctx, item.TargetID)
		if err != nil {
			return err
		}
		item.Claim = &claim

		inquiry, err := s.firstClaimInquirySummary(ctx, claim.ID)
		if err == nil {
			item.LinkedInquiry = inquiry
		} else if !errors.Is(err, ErrNotFound) {
			return err
		}
	case string(model.RediscoveryTargetTypeEngagement):
		engagement, err := s.GetEngagement(ctx, item.TargetID)
		if err != nil {
			return err
		}
		item.Engagement = &engagement

		inquiries, err := s.ListEngagementInquiries(ctx, engagement.ID)
		if err == nil && len(inquiries) > 0 {
			item.LinkedInquiry = &inquiries[0]
		} else if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
	case string(model.RediscoveryTargetTypeInquiry):
		inquiry, err := s.GetInquiry(ctx, item.TargetID)
		if err != nil {
			return err
		}
		item.Inquiry = &inquiry
	default:
		return ErrNotFound
	}

	return nil
}

func (s *Store) firstClaimInquirySummary(ctx context.Context, claimID string) (*model.InquirySummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT
			i.id,
			i.title,
			i.question,
			i.status
		FROM inquiries i
		JOIN claim_inquiries ci ON ci.inquiry_id = i.id
		WHERE ci.claim_id = ? AND i.archived_at IS NULL
		ORDER BY
			CASE i.status
				WHEN 'ACTIVE' THEN 0
				WHEN 'DORMANT' THEN 1
				WHEN 'SYNTHESIZED' THEN 2
				ELSE 3
			END,
			LOWER(i.title) ASC
		LIMIT 1
	`, claimID)

	var inquiry model.InquirySummary
	if err := row.Scan(&inquiry.ID, &inquiry.Title, &inquiry.Question, &inquiry.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &inquiry, nil
}

func scanRediscoveryItem(scanner interface {
	Scan(dest ...any) error
}) (model.RediscoveryItem, error) {
	var item model.RediscoveryItem
	if err := scanner.Scan(
		&item.ID,
		&item.Kind,
		&item.TargetType,
		&item.TargetID,
		&item.Reason,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return model.RediscoveryItem{}, err
	}

	return item, nil
}

func inClause(prefix string, ids []string) (string, []any) {
	args := make([]any, 0, len(ids))
	query := prefix
	for index, id := range ids {
		if index > 0 {
			query += ", "
		}
		query += "?"
		args = append(args, id)
	}
	return query, args
}
