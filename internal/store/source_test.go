package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestSourceLifecycle(t *testing.T) {
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
	year := 2024
	created, err := store.CreateSource(ctx, model.SourceInput{
		Title:            "The Shape of Thought",
		Medium:           string(model.SourceMediumBook),
		Creator:          "A. Writer",
		Year:             &year,
		OriginalLanguage: "en",
		CultureOrContext: "Test corpus",
		Notes:            "Seed note.",
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	if created.Title != "The Shape of Thought" {
		t.Fatalf("expected title to round-trip, got %q", created.Title)
	}

	listed, err := store.ListSources(ctx, model.SourceFilters{Sort: model.SourceSortRecent, Limit: 20})
	if err != nil {
		t.Fatalf("ListSources() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 source, got %d", len(listed))
	}

	updated, err := store.UpdateSource(ctx, created.ID, model.SourceInput{
		Title:            "The Shape of Thought Revised",
		Medium:           string(model.SourceMediumEssay),
		Creator:          "A. Writer",
		OriginalLanguage: "fr",
		CultureOrContext: "Updated corpus",
	})
	if err != nil {
		t.Fatalf("UpdateSource() error = %v", err)
	}
	if updated.Title != "The Shape of Thought Revised" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	if err := store.ArchiveSource(ctx, created.ID); err != nil {
		t.Fatalf("ArchiveSource() error = %v", err)
	}

	if _, err := store.GetSource(ctx, created.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after archive, got %v", err)
	}
}
