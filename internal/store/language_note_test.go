package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestLanguageNoteLifecycleAndEngagementFiltering(t *testing.T) {
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
		Title:            "Language Note Test Source",
		Medium:           string(model.SourceMediumBook),
		OriginalLanguage: "zh",
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	engagement, err := store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:           source.ID,
		EngagedAt:          "2026-03-06T12:00:00Z",
		Reflection:         "A reflection where wording and register matter.",
		SourceLanguage:     "zh",
		ReflectionLanguage: "en",
		AccessMode:         string(model.AccessModeLookupHeavy),
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	note, err := store.CreateLanguageNote(ctx, model.LanguageNoteInput{
		EngagementID: engagement.ID,
		Term:         "势",
		Language:     "zh",
		NoteType:     string(model.LanguageNoteTypeCulturalNuance),
		Content:      "The term carries strategic momentum, not just static power.",
	})
	if err != nil {
		t.Fatalf("CreateLanguageNote() error = %v", err)
	}

	if note.EngagementID != engagement.ID {
		t.Fatalf("expected engagement id %q, got %q", engagement.ID, note.EngagementID)
	}

	notes, err := store.ListEngagementLanguageNotes(ctx, engagement.ID)
	if err != nil {
		t.Fatalf("ListEngagementLanguageNotes() error = %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 language note, got %d", len(notes))
	}

	listed, err := store.ListEngagements(ctx, model.EngagementFilters{Limit: 10, HasLanguageNotes: true})
	if err != nil {
		t.Fatalf("ListEngagements() error = %v", err)
	}
	if len(listed) != 1 || listed[0].LanguageNoteCount != 1 {
		t.Fatalf("expected engagement with language note count 1, got %+v", listed)
	}

	updated, err := store.UpdateLanguageNote(ctx, note.ID, model.LanguageNoteInput{
		Term:     "势",
		Language: "zh",
		NoteType: string(model.LanguageNoteTypeTranslation),
		Content:  "Closer to force-of-situation than simple power.",
	})
	if err != nil {
		t.Fatalf("UpdateLanguageNote() error = %v", err)
	}
	if updated.NoteType == nil || *updated.NoteType != string(model.LanguageNoteTypeTranslation) {
		t.Fatalf("expected updated note type %q, got %+v", model.LanguageNoteTypeTranslation, updated.NoteType)
	}

	if err := store.ArchiveLanguageNote(ctx, note.ID); err != nil {
		t.Fatalf("ArchiveLanguageNote() error = %v", err)
	}
	if _, err := store.GetLanguageNote(ctx, note.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after archive, got %v", err)
	}

	listed, err = store.ListEngagements(ctx, model.EngagementFilters{Limit: 10, HasLanguageNotes: true})
	if err != nil {
		t.Fatalf("ListEngagements() error = %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected no engagements with active language notes, got %d", len(listed))
	}
}
