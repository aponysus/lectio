package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestEngagementLifecycle(t *testing.T) {
	t.Parallel()

	db, err := Open(filepath.Join(t.TempDir(), "lectio.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	store := New(db)
	source, err := store.CreateSource(ctx, model.SourceInput{
		Title:  "Grounding Text",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	priority := 4
	created, err := store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:           source.ID,
		EngagedAt:          "2026-03-06T12:00:00Z",
		PortionLabel:       "Chapter 1",
		Reflection:         "This is the first reflection.",
		WhyItMatters:       "It sets up the problem space.",
		SourceLanguage:     "en",
		ReflectionLanguage: "en",
		AccessMode:         string(model.AccessModeOriginal),
		RevisitPriority:    &priority,
		IsRereadOrRewatch:  true,
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	if created.Source.Title != source.Title {
		t.Fatalf("expected source title %q, got %q", source.Title, created.Source.Title)
	}

	listed, err := store.ListEngagements(ctx, model.EngagementFilters{Limit: 10})
	if err != nil {
		t.Fatalf("ListEngagements() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 engagement, got %d", len(listed))
	}

	updated, err := store.UpdateEngagement(ctx, created.ID, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2026-03-06T13:30:00Z",
		Reflection: "Updated reflection.",
	})
	if err != nil {
		t.Fatalf("UpdateEngagement() error = %v", err)
	}
	if updated.Reflection != "Updated reflection." {
		t.Fatalf("expected updated reflection, got %q", updated.Reflection)
	}

	if err := store.ArchiveEngagement(ctx, created.ID); err != nil {
		t.Fatalf("ArchiveEngagement() error = %v", err)
	}
	if _, err := store.GetEngagement(ctx, created.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after archive, got %v", err)
	}
}

func TestListEngagementsFiltersByQuery(t *testing.T) {
	t.Parallel()

	db, err := Open(filepath.Join(t.TempDir(), "lectio.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	store := New(db)
	source, err := store.CreateSource(ctx, model.SourceInput{
		Title:  "Searchable Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	_, err = store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2026-03-06T10:00:00Z",
		Reflection: "A note about tragic structure and dramatic irony.",
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	_, err = store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2026-03-06T11:00:00Z",
		Reflection: "A separate note about syntax and cadence.",
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	results, err := store.ListEngagements(ctx, model.EngagementFilters{
		Query: "dramatic irony",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEngagements() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 engagement, got %d", len(results))
	}
	if results[0].Reflection != "A note about tragic structure and dramatic irony." {
		t.Fatalf("unexpected engagement returned: %q", results[0].Reflection)
	}
}
