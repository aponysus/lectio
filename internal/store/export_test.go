package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestExportDataIncludesCoreTables(t *testing.T) {
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
		Title:  "Export Seed Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	engagement, err := store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2026-03-06T12:00:00Z",
		Reflection: "A seeded engagement for export coverage.",
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	inquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "What survives export?",
		Question: "Which linked records are preserved in the backup payload?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}

	if err := store.ReplaceEngagementInquiries(ctx, engagement.ID, []string{inquiry.ID}); err != nil {
		t.Fatalf("ReplaceEngagementInquiries() error = %v", err)
	}

	claim, err := store.CreateClaim(ctx, model.ClaimInput{
		Text:               "The export should preserve the graph, not just isolated rows.",
		ClaimType:          string(model.ClaimTypeInterpretation),
		Status:             string(model.ClaimStatusTentative),
		OriginEngagementID: engagement.ID,
	}, []string{inquiry.ID})
	if err != nil {
		t.Fatalf("CreateClaim() error = %v", err)
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE claims
		SET updated_at = datetime('now', '-20 days')
		WHERE id = ?
	`, claim.ID); err != nil {
		t.Fatalf("aging claim for rediscovery error = %v", err)
	}

	if _, err := store.CreateLanguageNote(ctx, model.LanguageNoteInput{
		EngagementID: engagement.ID,
		Term:         "mneme",
		Language:     "grc",
		NoteType:     string(model.LanguageNoteTypeTranslation),
		Content:      "A language note that should travel with the export.",
	}); err != nil {
		t.Fatalf("CreateLanguageNote() error = %v", err)
	}

	if _, err := store.CreateSynthesis(ctx, model.SynthesisInput{
		Title:     "Checkpoint export synthesis",
		Body:      "The payload includes both object tables and join tables.",
		Type:      string(model.SynthesisTypeCheckpoint),
		InquiryID: inquiry.ID,
	}); err != nil {
		t.Fatalf("CreateSynthesis() error = %v", err)
	}

	if _, err := store.ListRediscoveryItems(ctx, 10); err != nil {
		t.Fatalf("ListRediscoveryItems() error = %v", err)
	}

	payload, err := store.ExportData(ctx)
	if err != nil {
		t.Fatalf("ExportData() error = %v", err)
	}

	if payload.FormatVersion != 1 {
		t.Fatalf("expected format version 1, got %d", payload.FormatVersion)
	}
	if len(payload.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(payload.Sources))
	}
	if len(payload.Engagements) != 1 {
		t.Fatalf("expected 1 engagement, got %d", len(payload.Engagements))
	}
	if len(payload.Inquiries) != 1 {
		t.Fatalf("expected 1 inquiry, got %d", len(payload.Inquiries))
	}
	if len(payload.EngagementInquiries) != 1 {
		t.Fatalf("expected 1 engagement inquiry link, got %d", len(payload.EngagementInquiries))
	}
	if len(payload.Claims) != 1 {
		t.Fatalf("expected 1 claim, got %d", len(payload.Claims))
	}
	if len(payload.ClaimInquiries) != 1 {
		t.Fatalf("expected 1 claim inquiry link, got %d", len(payload.ClaimInquiries))
	}
	if len(payload.LanguageNotes) != 1 {
		t.Fatalf("expected 1 language note, got %d", len(payload.LanguageNotes))
	}
	if len(payload.Syntheses) != 1 {
		t.Fatalf("expected 1 synthesis, got %d", len(payload.Syntheses))
	}
	if len(payload.RediscoveryItems) != 1 {
		t.Fatalf("expected 1 rediscovery item, got %d", len(payload.RediscoveryItems))
	}

	if payload.EngagementInquiries[0].EngagementID != engagement.ID || payload.EngagementInquiries[0].InquiryID != inquiry.ID {
		t.Fatalf("unexpected engagement link payload: %+v", payload.EngagementInquiries[0])
	}
	if payload.ClaimInquiries[0].ClaimID != claim.ID || payload.ClaimInquiries[0].InquiryID != inquiry.ID {
		t.Fatalf("unexpected claim link payload: %+v", payload.ClaimInquiries[0])
	}
	if payload.RediscoveryItems[0].TargetID != claim.ID {
		t.Fatalf("expected rediscovery target %q, got %q", claim.ID, payload.RediscoveryItems[0].TargetID)
	}
}
