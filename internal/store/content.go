package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

type sqliteEntryRepository struct {
	db      *sql.DB
	sources SourceRepository
	tags    TagRepository
}

type sqliteSourceRepository struct {
	db *sql.DB
}

type sqliteTagRepository struct {
	db *sql.DB
}

func newEntryRepository(db *sql.DB) EntryRepository {
	sources := newSourceRepository(db)
	tags := newTagRepository(db)

	return &sqliteEntryRepository{
		db:      db,
		sources: sources,
		tags:    tags,
	}
}

func newSourceRepository(db *sql.DB) SourceRepository {
	return &sqliteSourceRepository{db: db}
}

func newTagRepository(db *sql.DB) TagRepository {
	return &sqliteTagRepository{db: db}
}

func (r *sqliteEntryRepository) Create(ctx context.Context, input CreateEntryInput) (Entry, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Entry{}, fmt.Errorf("begin create entry: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	source, err := resolveSourceTx(ctx, tx, input.SourceID, input.Source)
	if err != nil {
		return Entry{}, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO entries (source_id, passage, reflection, mood, energy)
		VALUES (?, ?, ?, ?, ?)
	`, source.ID, nullableString(strings.TrimSpace(input.Passage)), input.Reflection, nullableString(strings.TrimSpace(input.Mood)), input.Energy)
	if err != nil {
		return Entry{}, fmt.Errorf("insert entry: %w", err)
	}

	entryID, err := result.LastInsertId()
	if err != nil {
		return Entry{}, fmt.Errorf("read entry id: %w", err)
	}

	tags, err := ensureTagsTx(ctx, tx, input.Tags)
	if err != nil {
		return Entry{}, err
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO entry_tags (entry_id, tag_id)
			VALUES (?, ?)
		`, entryID, tag.ID); err != nil {
			return Entry{}, fmt.Errorf("link entry tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Entry{}, fmt.Errorf("commit create entry: %w", err)
	}

	return r.GetByID(ctx, entryID)
}

func (r *sqliteEntryRepository) GetByID(ctx context.Context, id int64) (Entry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, source_id, COALESCE(passage, ''), reflection, COALESCE(mood, ''), energy, created_at, updated_at
		FROM entries
		WHERE id = ?
	`, id)

	entry, err := scanEntry(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Entry{}, ErrNotFound
	}
	if err != nil {
		return Entry{}, fmt.Errorf("get entry: %w", err)
	}

	entry.Tags, err = listEntryTags(ctx, r.db, entry.ID)
	if err != nil {
		return Entry{}, err
	}

	return entry, nil
}

func (r *sqliteEntryRepository) List(ctx context.Context, filter EntryListFilter) (EntryListResult, error) {
	conditions := []string{}
	args := []any{}

	if filter.SourceID != nil {
		conditions = append(conditions, "e.source_id = ?")
		args = append(args, *filter.SourceID)
	}
	if tag := normalizeSlug(filter.Tag); tag != "" {
		conditions = append(conditions, `EXISTS (
			SELECT 1
			FROM entry_tags et
			JOIN tags t ON t.id = et.tag_id
			WHERE et.entry_id = e.id AND t.slug = ?
		)`)
		args = append(args, tag)
	}
	if filter.From != nil {
		conditions = append(conditions, "e.created_at >= ?")
		args = append(args, filter.From.UTC())
	}
	if filter.To != nil {
		conditions = append(conditions, "e.created_at <= ?")
		args = append(args, filter.To.UTC())
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM entries e ` + whereClause
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return EntryListResult{}, fmt.Errorf("count entries: %w", err)
	}

	queryArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.source_id, COALESCE(e.passage, ''), e.reflection, COALESCE(e.mood, ''), e.energy, e.created_at, e.updated_at
		FROM entries e
		`+whereClause+`
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return EntryListResult{}, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	entries := make([]Entry, 0, pageSize)
	entryIDs := make([]int64, 0, pageSize)
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return EntryListResult{}, fmt.Errorf("scan list entry: %w", err)
		}
		entries = append(entries, entry)
		entryIDs = append(entryIDs, entry.ID)
	}
	if err := rows.Err(); err != nil {
		return EntryListResult{}, fmt.Errorf("iterate entries: %w", err)
	}

	tagMap, err := listTagsForEntries(ctx, r.db, entryIDs)
	if err != nil {
		return EntryListResult{}, err
	}
	for i := range entries {
		entries[i].Tags = tagMap[entries[i].ID]
	}

	return EntryListResult{
		Entries: entries,
		Total:   total,
	}, nil
}

func (r *sqliteEntryRepository) Update(context.Context, Entry) (Entry, error) {
	return Entry{}, errors.New("store: entry update not implemented")
}

func (r *sqliteEntryRepository) Delete(context.Context, int64) error {
	return errors.New("store: entry delete not implemented")
}

func (r *sqliteSourceRepository) Create(ctx context.Context, source Source) (Source, error) {
	if strings.TrimSpace(source.Title) == "" {
		return Source{}, fmt.Errorf("create source: title required")
	}

	language := strings.TrimSpace(source.Language)
	if language == "" {
		language = "en"
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO sources (title, author, year, tradition, language)
		VALUES (?, ?, ?, ?, ?)
	`, source.Title, nullableString(strings.TrimSpace(source.Author)), source.Year, nullableString(strings.TrimSpace(source.Tradition)), language)
	if err != nil {
		return Source{}, fmt.Errorf("create source: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Source{}, fmt.Errorf("read source id: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *sqliteSourceRepository) GetByID(ctx context.Context, id int64) (Source, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, COALESCE(author, ''), year, COALESCE(tradition, ''), language, created_at, updated_at,
		       (SELECT COUNT(*) FROM entries WHERE source_id = sources.id) AS entry_count
		FROM sources
		WHERE id = ?
	`, id)

	source, err := scanSource(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Source{}, ErrNotFound
	}
	if err != nil {
		return Source{}, fmt.Errorf("get source: %w", err)
	}

	return source, nil
}

func (r *sqliteSourceRepository) List(ctx context.Context, filter SourceListFilter) ([]Source, error) {
	args := []any{}
	whereClause := ""
	if q := strings.TrimSpace(filter.Query); q != "" {
		whereClause = "WHERE lower(title) LIKE lower(?)"
		args = append(args, q+"%")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, COALESCE(author, ''), year, COALESCE(tradition, ''), language, created_at, updated_at,
		       (SELECT COUNT(*) FROM entries WHERE source_id = sources.id) AS entry_count
		FROM sources
		`+whereClause+`
		ORDER BY title ASC, id ASC
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		source, err := scanSource(rows)
		if err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sources: %w", err)
	}

	return sources, nil
}

func (r *sqliteSourceRepository) Resolve(ctx context.Context, input SourceInput) (Source, error) {
	return resolveSourceTx(ctx, r.db, input.ID, input)
}

func (r *sqliteSourceRepository) Update(context.Context, Source) (Source, error) {
	return Source{}, errors.New("store: source update not implemented")
}

func (r *sqliteTagRepository) Ensure(ctx context.Context, labels []string) ([]Tag, error) {
	return ensureTagsTx(ctx, r.db, labels)
}

func (r *sqliteTagRepository) GetBySlug(ctx context.Context, slug string) (Tag, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, slug, label, created_at,
		       (SELECT COUNT(*) FROM entry_tags WHERE tag_id = tags.id) AS entry_count
		FROM tags
		WHERE slug = ?
	`, normalizeSlug(slug))

	tag, err := scanTag(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Tag{}, ErrNotFound
	}
	if err != nil {
		return Tag{}, fmt.Errorf("get tag: %w", err)
	}
	return tag, nil
}

func (r *sqliteTagRepository) List(ctx context.Context, filter TagListFilter) ([]Tag, error) {
	args := []any{}
	whereClause := ""
	if q := strings.TrimSpace(filter.Query); q != "" {
		whereClause = "WHERE lower(label) LIKE lower(?) OR slug LIKE ?"
		args = append(args, q+"%", normalizeSlug(q)+"%")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, slug, label, created_at,
		       (SELECT COUNT(*) FROM entry_tags WHERE tag_id = tags.id) AS entry_count
		FROM tags
		`+whereClause+`
		ORDER BY label ASC, id ASC
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}

func (r *sqliteTagRepository) CoOccurrence(context.Context) ([]TagPair, error) {
	return nil, errors.New("store: tag co-occurrence not implemented")
}

type execQuerier interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func resolveSourceTx(ctx context.Context, db execQuerier, sourceID *int64, input SourceInput) (Source, error) {
	if input.ID != nil {
		sourceID = input.ID
	}
	if sourceID != nil {
		row := db.QueryRowContext(ctx, `
			SELECT id, title, COALESCE(author, ''), year, COALESCE(tradition, ''), language, created_at, updated_at,
			       (SELECT COUNT(*) FROM entries WHERE source_id = sources.id) AS entry_count
			FROM sources
			WHERE id = ?
		`, *sourceID)
		source, err := scanSource(row)
		if errors.Is(err, sql.ErrNoRows) {
			return Source{}, ErrNotFound
		}
		if err != nil {
			return Source{}, fmt.Errorf("resolve source by id: %w", err)
		}
		return source, nil
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Source{}, fmt.Errorf("create entry: source title required when source id is absent")
	}

	language := strings.TrimSpace(input.Language)
	if language == "" {
		language = "en"
	}
	author := strings.TrimSpace(input.Author)
	tradition := strings.TrimSpace(input.Tradition)

	row := db.QueryRowContext(ctx, `
		SELECT id, title, COALESCE(author, ''), year, COALESCE(tradition, ''), language, created_at, updated_at,
		       (SELECT COUNT(*) FROM entries WHERE source_id = sources.id) AS entry_count
		FROM sources
		WHERE lower(title) = lower(?)
		  AND ifnull(lower(author), '') = ifnull(lower(?), '')
		  AND ifnull(year, 0) = ifnull(?, 0)
		  AND lower(language) = lower(?)
	`, title, nullableString(author), input.Year, language)
	source, err := scanSource(row)
	if err == nil {
		return source, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Source{}, fmt.Errorf("lookup source: %w", err)
	}

	result, err := db.ExecContext(ctx, `
		INSERT INTO sources (title, author, year, tradition, language)
		VALUES (?, ?, ?, ?, ?)
	`, title, nullableString(author), input.Year, nullableString(tradition), language)
	if err != nil {
		return Source{}, fmt.Errorf("create source: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Source{}, fmt.Errorf("read created source id: %w", err)
	}

	row = db.QueryRowContext(ctx, `
		SELECT id, title, COALESCE(author, ''), year, COALESCE(tradition, ''), language, created_at, updated_at,
		       (SELECT COUNT(*) FROM entries WHERE source_id = sources.id) AS entry_count
		FROM sources
		WHERE id = ?
	`, id)
	source, err = scanSource(row)
	if err != nil {
		return Source{}, fmt.Errorf("read created source: %w", err)
	}
	return source, nil
}

func ensureTagsTx(ctx context.Context, db execQuerier, labels []string) ([]Tag, error) {
	normalized := normalizeTagLabels(labels)
	tags := make([]Tag, 0, len(normalized))
	for _, label := range normalized {
		slug := normalizeSlug(label)
		if slug == "" {
			continue
		}

		row := db.QueryRowContext(ctx, `
			SELECT id, slug, label, created_at,
			       (SELECT COUNT(*) FROM entry_tags WHERE tag_id = tags.id) AS entry_count
			FROM tags
			WHERE slug = ?
		`, slug)
		tag, err := scanTag(row)
		if err == nil {
			tags = append(tags, tag)
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lookup tag: %w", err)
		}

		result, err := db.ExecContext(ctx, `
			INSERT INTO tags (slug, label)
			VALUES (?, ?)
		`, slug, label)
		if err != nil {
			return nil, fmt.Errorf("create tag: %w", err)
		}
		tagID, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("read tag id: %w", err)
		}

		row = db.QueryRowContext(ctx, `
			SELECT id, slug, label, created_at,
			       (SELECT COUNT(*) FROM entry_tags WHERE tag_id = tags.id) AS entry_count
			FROM tags
			WHERE id = ?
		`, tagID)
		tag, err = scanTag(row)
		if err != nil {
			return nil, fmt.Errorf("read created tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func listEntryTags(ctx context.Context, db *sql.DB, entryID int64) ([]Tag, error) {
	tagMap, err := listTagsForEntries(ctx, db, []int64{entryID})
	if err != nil {
		return nil, err
	}
	return tagMap[entryID], nil
}

func listTagsForEntries(ctx context.Context, db *sql.DB, entryIDs []int64) (map[int64][]Tag, error) {
	result := make(map[int64][]Tag, len(entryIDs))
	if len(entryIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(entryIDs))
	args := make([]any, 0, len(entryIDs))
	for _, id := range entryIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT et.entry_id, t.id, t.slug, t.label, t.created_at,
		       (SELECT COUNT(*) FROM entry_tags WHERE tag_id = t.id) AS entry_count
		FROM entry_tags et
		JOIN tags t ON t.id = et.tag_id
		WHERE et.entry_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY t.label ASC, t.id ASC
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list entry tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entryID int64
		var tag Tag
		if err := rows.Scan(&entryID, &tag.ID, &tag.Slug, &tag.Label, &tag.CreatedAt, &tag.Count); err != nil {
			return nil, fmt.Errorf("scan entry tag: %w", err)
		}
		result[entryID] = append(result[entryID], tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entry tags: %w", err)
	}

	return result, nil
}

func scanEntry(scanner interface{ Scan(...any) error }) (Entry, error) {
	var entry Entry
	var energy sql.NullInt64
	if err := scanner.Scan(
		&entry.ID,
		&entry.SourceID,
		&entry.Passage,
		&entry.Reflection,
		&entry.Mood,
		&energy,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return Entry{}, err
	}
	if energy.Valid {
		value := int(energy.Int64)
		entry.Energy = &value
	}
	return entry, nil
}

func scanSource(scanner interface{ Scan(...any) error }) (Source, error) {
	var source Source
	var year sql.NullInt64
	if err := scanner.Scan(
		&source.ID,
		&source.Title,
		&source.Author,
		&year,
		&source.Tradition,
		&source.Language,
		&source.CreatedAt,
		&source.UpdatedAt,
		&source.EntryCount,
	); err != nil {
		return Source{}, err
	}
	if year.Valid {
		value := int(year.Int64)
		source.Year = &value
	}
	return source, nil
}

func scanTag(scanner interface{ Scan(...any) error }) (Tag, error) {
	var tag Tag
	if err := scanner.Scan(&tag.ID, &tag.Slug, &tag.Label, &tag.CreatedAt, &tag.Count); err != nil {
		return Tag{}, err
	}
	return tag, nil
}

func normalizeTagLabels(labels []string) []string {
	seen := make(map[string]struct{}, len(labels))
	normalized := make([]string, 0, len(labels))
	for _, label := range labels {
		display := strings.TrimSpace(label)
		slug := normalizeSlug(display)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		normalized = append(normalized, display)
	}
	return normalized
}

func normalizeSlug(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = slugSanitizer.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return normalized
}
