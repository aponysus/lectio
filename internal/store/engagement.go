package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

func (s *Store) CreateEngagement(ctx context.Context, input model.EngagementInput) (model.Engagement, error) {
	source, err := s.sourceSummaryByID(ctx, input.SourceID)
	if err != nil {
		return model.Engagement{}, err
	}

	id, err := newID()
	if err != nil {
		return model.Engagement{}, err
	}

	_, err = s.db.ExecContext(ctx, `
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
		source.ID,
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
		return model.Engagement{}, err
	}

	return s.GetEngagement(ctx, id)
}

func (s *Store) ListEngagements(ctx context.Context, filters model.EngagementFilters) ([]model.Engagement, error) {
	args := []any{}
	var where []string

	if !filters.IncludeArchived {
		where = append(where, "e.archived_at IS NULL")
	}
	if filters.Query != "" {
		where = append(where, "LOWER(e.reflection) LIKE ?")
		args = append(args, "%"+strings.ToLower(filters.Query)+"%")
	}
	if filters.SourceID != "" {
		where = append(where, "e.source_id = ?")
		args = append(args, filters.SourceID)
	}
	if filters.AccessMode != "" {
		where = append(where, "e.access_mode = ?")
		args = append(args, filters.AccessMode)
	}
	if filters.HasLanguageNotes {
		where = append(where, "COALESCE(note_stats.language_note_count, 0) > 0")
	}

	query := engagementSelectQuery
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY datetime(e.engaged_at) DESC, e.updated_at DESC, e.id DESC LIMIT ?"
	args = append(args, filters.Limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
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

func (s *Store) GetEngagement(ctx context.Context, id string) (model.Engagement, error) {
	row := s.db.QueryRowContext(ctx,
		engagementSelectQuery+" WHERE e.id = ? AND e.archived_at IS NULL",
		id,
	)

	engagement, err := scanEngagement(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Engagement{}, ErrNotFound
		}
		return model.Engagement{}, err
	}

	return engagement, nil
}

func (s *Store) UpdateEngagement(ctx context.Context, id string, input model.EngagementInput) (model.Engagement, error) {
	if _, err := s.sourceSummaryByID(ctx, input.SourceID); err != nil {
		return model.Engagement{}, err
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE engagements
		SET
			source_id = ?,
			engaged_at = ?,
			portion_label = ?,
			reflection = ?,
			why_it_matters = ?,
			source_language = ?,
			reflection_language = ?,
			access_mode = ?,
			revisit_priority = ?,
			is_reread_or_rewatch = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND archived_at IS NULL
	`,
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
		id,
	)
	if err != nil {
		return model.Engagement{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Engagement{}, err
	}
	if rowsAffected == 0 {
		return model.Engagement{}, ErrNotFound
	}

	return s.GetEngagement(ctx, id)
}

func (s *Store) ArchiveEngagement(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE engagements
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

func (s *Store) sourceSummaryByID(ctx context.Context, id string) (model.SourceSummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, title, medium, creator
		FROM sources
		WHERE id = ? AND archived_at IS NULL
	`, id)

	var summary model.SourceSummary
	var creator sql.NullString
	if err := row.Scan(&summary.ID, &summary.Title, &summary.Medium, &creator); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SourceSummary{}, ErrNotFound
		}
		return model.SourceSummary{}, err
	}

	summary.Creator = stringPointer(creator)
	return summary, nil
}

const engagementSelectQuery = `
	SELECT
		e.id,
		e.source_id,
		strftime('%Y-%m-%dT%H:%M:%SZ', e.engaged_at) AS engaged_at,
		e.portion_label,
		e.reflection,
		e.why_it_matters,
		e.source_language,
		e.reflection_language,
		e.access_mode,
		e.revisit_priority,
		e.is_reread_or_rewatch,
		strftime('%Y-%m-%dT%H:%M:%SZ', e.created_at) AS created_at,
	strftime('%Y-%m-%dT%H:%M:%SZ', e.updated_at) AS updated_at,
	CASE
		WHEN e.archived_at IS NOT NULL THEN strftime('%Y-%m-%dT%H:%M:%SZ', e.archived_at)
	END AS archived_at,
	COALESCE(note_stats.language_note_count, 0) AS language_note_count,
	s.id,
	s.title,
	s.medium,
	s.creator
	FROM engagements e
	LEFT JOIN (
		SELECT
			ln.engagement_id,
			COUNT(ln.id) AS language_note_count
		FROM language_notes ln
		WHERE ln.archived_at IS NULL
		GROUP BY ln.engagement_id
	) note_stats ON note_stats.engagement_id = e.id
	JOIN sources s ON s.id = e.source_id
`

type engagementScanner interface {
	Scan(dest ...any) error
}

func scanEngagement(scanner engagementScanner) (model.Engagement, error) {
	var engagement model.Engagement
	var portionLabel sql.NullString
	var whyItMatters sql.NullString
	var sourceLanguage sql.NullString
	var reflectionLanguage sql.NullString
	var accessMode sql.NullString
	var revisitPriority sql.NullInt64
	var archivedAt sql.NullString
	var sourceCreator sql.NullString

	if err := scanner.Scan(
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
		&engagement.LanguageNoteCount,
		&engagement.Source.ID,
		&engagement.Source.Title,
		&engagement.Source.Medium,
		&sourceCreator,
	); err != nil {
		return model.Engagement{}, err
	}

	engagement.PortionLabel = stringPointer(portionLabel)
	engagement.WhyItMatters = stringPointer(whyItMatters)
	engagement.SourceLanguage = stringPointer(sourceLanguage)
	engagement.ReflectionLanguage = stringPointer(reflectionLanguage)
	engagement.AccessMode = stringPointer(accessMode)
	engagement.RevisitPriority = intPointer(revisitPriority)
	engagement.ArchivedAt = stringPointer(archivedAt)
	engagement.Source.Creator = stringPointer(sourceCreator)

	return engagement, nil
}
