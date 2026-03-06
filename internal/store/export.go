package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) ExportData(ctx context.Context) (model.ExportPayload, error) {
	sources, err := s.exportSources(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	engagements, err := s.exportEngagements(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	inquiries, err := s.exportInquiries(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	engagementInquiries, err := s.exportEngagementInquiries(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	claims, err := s.exportClaims(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	claimInquiries, err := s.exportClaimInquiries(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	languageNotes, err := s.exportLanguageNotes(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	syntheses, err := s.exportSyntheses(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	rediscoveryItems, err := s.exportRediscoveryItems(ctx)
	if err != nil {
		return model.ExportPayload{}, err
	}

	return model.ExportPayload{
		FormatVersion:       1,
		ExportedAt:          time.Now().UTC().Format(time.RFC3339),
		Sources:             sources,
		Engagements:         engagements,
		Inquiries:           inquiries,
		EngagementInquiries: engagementInquiries,
		Claims:              claims,
		ClaimInquiries:      claimInquiries,
		LanguageNotes:       languageNotes,
		Syntheses:           syntheses,
		RediscoveryItems:    rediscoveryItems,
	}, nil
}

func (s *Store) exportSources(ctx context.Context) ([]model.Source, error) {
	rows, err := s.db.QueryContext(ctx, sourceSelectQuery+`
		ORDER BY datetime(created_at) ASC, id ASC
	`)
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

func (s *Store) exportEngagements(ctx context.Context) ([]model.ExportEngagement, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			source_id,
			strftime('%Y-%m-%dT%H:%M:%SZ', engaged_at) AS engaged_at,
			portion_label,
			reflection,
			why_it_matters,
			source_language,
			reflection_language,
			access_mode,
			revisit_priority,
			is_reread_or_rewatch,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at,
			strftime('%Y-%m-%dT%H:%M:%SZ', updated_at) AS updated_at,
			CASE
				WHEN archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', archived_at)
			END AS archived_at
		FROM engagements
		ORDER BY datetime(engaged_at) ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	engagements := []model.ExportEngagement{}
	for rows.Next() {
		var engagement model.ExportEngagement
		var portionLabel sql.NullString
		var whyItMatters sql.NullString
		var sourceLanguage sql.NullString
		var reflectionLanguage sql.NullString
		var accessMode sql.NullString
		var revisitPriority sql.NullInt64
		var archivedAt sql.NullString

		if err := rows.Scan(
			&engagement.ID,
			&engagement.SourceID,
			&engagement.EngagedAt,
			&portionLabel,
			&engagement.Reflection,
			&whyItMatters,
			&sourceLanguage,
			&reflectionLanguage,
			&accessMode,
			&revisitPriority,
			&engagement.IsRereadOrRewatch,
			&engagement.CreatedAt,
			&engagement.UpdatedAt,
			&archivedAt,
		); err != nil {
			return nil, err
		}

		engagement.PortionLabel = stringPointer(portionLabel)
		engagement.WhyItMatters = stringPointer(whyItMatters)
		engagement.SourceLanguage = stringPointer(sourceLanguage)
		engagement.ReflectionLanguage = stringPointer(reflectionLanguage)
		engagement.AccessMode = stringPointer(accessMode)
		engagement.RevisitPriority = intPointer(revisitPriority)
		engagement.ArchivedAt = stringPointer(archivedAt)
		engagements = append(engagements, engagement)
	}

	return engagements, rows.Err()
}

func (s *Store) exportInquiries(ctx context.Context) ([]model.ExportInquiry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			title,
			question,
			status,
			why_it_matters,
			current_view,
			open_tensions,
			CASE
				WHEN reactivated_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', reactivated_at)
			END AS reactivated_at,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at,
			strftime('%Y-%m-%dT%H:%M:%SZ', updated_at) AS updated_at,
			CASE
				WHEN archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', archived_at)
			END AS archived_at
		FROM inquiries
		ORDER BY datetime(created_at) ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inquiries := []model.ExportInquiry{}
	for rows.Next() {
		var inquiry model.ExportInquiry
		var whyItMatters sql.NullString
		var currentView sql.NullString
		var openTensions sql.NullString
		var reactivatedAt sql.NullString
		var archivedAt sql.NullString

		if err := rows.Scan(
			&inquiry.ID,
			&inquiry.Title,
			&inquiry.Question,
			&inquiry.Status,
			&whyItMatters,
			&currentView,
			&openTensions,
			&reactivatedAt,
			&inquiry.CreatedAt,
			&inquiry.UpdatedAt,
			&archivedAt,
		); err != nil {
			return nil, err
		}

		inquiry.WhyItMatters = stringPointer(whyItMatters)
		inquiry.CurrentView = stringPointer(currentView)
		inquiry.OpenTensions = stringPointer(openTensions)
		inquiry.ReactivatedAt = stringPointer(reactivatedAt)
		inquiry.ArchivedAt = stringPointer(archivedAt)
		inquiries = append(inquiries, inquiry)
	}

	return inquiries, rows.Err()
}

func (s *Store) exportEngagementInquiries(ctx context.Context) ([]model.EngagementInquiryLink, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			engagement_id,
			inquiry_id,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at
		FROM engagement_inquiries
		ORDER BY datetime(created_at) ASC, engagement_id ASC, inquiry_id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []model.EngagementInquiryLink{}
	for rows.Next() {
		var link model.EngagementInquiryLink
		if err := rows.Scan(&link.EngagementID, &link.InquiryID, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

func (s *Store) exportClaims(ctx context.Context) ([]model.ExportClaim, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			text,
			claim_type,
			confidence,
			status,
			origin_engagement_id,
			notes,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at,
			strftime('%Y-%m-%dT%H:%M:%SZ', updated_at) AS updated_at,
			CASE
				WHEN archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', archived_at)
			END AS archived_at
		FROM claims
		ORDER BY datetime(created_at) ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := []model.ExportClaim{}
	for rows.Next() {
		var claim model.ExportClaim
		var confidence sql.NullInt64
		var originEngagementID sql.NullString
		var notes sql.NullString
		var archivedAt sql.NullString

		if err := rows.Scan(
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
		); err != nil {
			return nil, err
		}

		claim.Confidence = intPointer(confidence)
		claim.OriginEngagementID = stringPointer(originEngagementID)
		claim.Notes = stringPointer(notes)
		claim.ArchivedAt = stringPointer(archivedAt)
		claims = append(claims, claim)
	}

	return claims, rows.Err()
}

func (s *Store) exportClaimInquiries(ctx context.Context) ([]model.ClaimInquiryLink, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			claim_id,
			inquiry_id,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at
		FROM claim_inquiries
		ORDER BY datetime(created_at) ASC, claim_id ASC, inquiry_id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []model.ClaimInquiryLink{}
	for rows.Next() {
		var link model.ClaimInquiryLink
		if err := rows.Scan(&link.ClaimID, &link.InquiryID, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

func (s *Store) exportLanguageNotes(ctx context.Context) ([]model.LanguageNote, error) {
	rows, err := s.db.QueryContext(ctx, languageNoteSelectQuery+`
		ORDER BY datetime(ln.created_at) ASC, ln.id ASC
	`)
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

func (s *Store) exportSyntheses(ctx context.Context) ([]model.ExportSynthesis, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			title,
			body,
			type,
			inquiry_id,
			notes,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) AS created_at,
			strftime('%Y-%m-%dT%H:%M:%SZ', updated_at) AS updated_at,
			CASE
				WHEN archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', archived_at)
			END AS archived_at
		FROM syntheses
		ORDER BY datetime(created_at) ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	syntheses := []model.ExportSynthesis{}
	for rows.Next() {
		var synthesis model.ExportSynthesis
		var notes sql.NullString
		var archivedAt sql.NullString

		if err := rows.Scan(
			&synthesis.ID,
			&synthesis.Title,
			&synthesis.Body,
			&synthesis.Type,
			&synthesis.InquiryID,
			&notes,
			&synthesis.CreatedAt,
			&synthesis.UpdatedAt,
			&archivedAt,
		); err != nil {
			return nil, err
		}

		synthesis.Notes = stringPointer(notes)
		synthesis.ArchivedAt = stringPointer(archivedAt)
		syntheses = append(syntheses, synthesis)
	}

	return syntheses, rows.Err()
}

func (s *Store) exportRediscoveryItems(ctx context.Context) ([]model.ExportRediscoveryItem, error) {
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
		ORDER BY datetime(created_at) ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ExportRediscoveryItem{}
	for rows.Next() {
		var item model.ExportRediscoveryItem
		if err := rows.Scan(
			&item.ID,
			&item.Kind,
			&item.TargetType,
			&item.TargetID,
			&item.Reason,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
