package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestInquiryLifecycleAndLinking(t *testing.T) {
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
		Title:  "Inquiry Test Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	engagement, err := store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2026-03-06T12:00:00Z",
		Reflection: "A reflection that should land on an inquiry.",
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	inquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "What is really being argued here?",
		Question: "How does the text structure its central claim?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}

	if err := store.ReplaceEngagementInquiries(ctx, engagement.ID, []string{inquiry.ID}); err != nil {
		t.Fatalf("ReplaceEngagementInquiries() error = %v", err)
	}

	linkedInquiries, err := store.ListEngagementInquiries(ctx, engagement.ID)
	if err != nil {
		t.Fatalf("ListEngagementInquiries() error = %v", err)
	}
	if len(linkedInquiries) != 1 {
		t.Fatalf("expected 1 linked inquiry, got %d", len(linkedInquiries))
	}

	inquiryEngagements, err := store.ListInquiryEngagements(ctx, inquiry.ID, 10)
	if err != nil {
		t.Fatalf("ListInquiryEngagements() error = %v", err)
	}
	if len(inquiryEngagements) != 1 {
		t.Fatalf("expected 1 linked engagement, got %d", len(inquiryEngagements))
	}

	listed, err := store.ListInquiries(ctx, model.InquiryFilters{Limit: 10})
	if err != nil {
		t.Fatalf("ListInquiries() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 inquiry, got %d", len(listed))
	}
	if listed[0].EngagementCount != 1 {
		t.Fatalf("expected engagement count 1, got %d", listed[0].EngagementCount)
	}

	updated, err := store.UpdateInquiry(ctx, inquiry.ID, model.InquiryInput{
		Title:        "What is really being argued here now?",
		Question:     "How does the text structure its central claim after revision?",
		Status:       string(model.InquiryStatusDormant),
		WhyItMatters: "This question governs later synthesis.",
	})
	if err != nil {
		t.Fatalf("UpdateInquiry() error = %v", err)
	}
	if updated.Status != string(model.InquiryStatusDormant) {
		t.Fatalf("expected updated status %q, got %q", model.InquiryStatusDormant, updated.Status)
	}

	if err := store.ArchiveInquiry(ctx, inquiry.ID); err != nil {
		t.Fatalf("ArchiveInquiry() error = %v", err)
	}
	if _, err := store.GetInquiry(ctx, inquiry.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after archive, got %v", err)
	}
}
