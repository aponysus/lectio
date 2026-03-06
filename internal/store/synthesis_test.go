package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestSynthesisLifecycleAndEligibility(t *testing.T) {
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
		Title:  "Synthesis Test Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	inquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "What has the inquiry actually become?",
		Question: "How should the accumulated material now be compressed?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}

	for i := 0; i < 3; i++ {
		engagement, err := store.CreateEngagement(ctx, model.EngagementInput{
			SourceID:   source.ID,
			EngagedAt:  "2026-03-06T12:00:00Z",
			Reflection: "A reflection that should push the inquiry toward synthesis.",
		})
		if err != nil {
			t.Fatalf("CreateEngagement() error = %v", err)
		}
		if err := store.ReplaceEngagementInquiries(ctx, engagement.ID, []string{inquiry.ID}); err != nil {
			t.Fatalf("ReplaceEngagementInquiries() error = %v", err)
		}
	}

	eligible, err := store.ListEligibleForSynthesisInquiries(ctx, 10)
	if err != nil {
		t.Fatalf("ListEligibleForSynthesisInquiries() error = %v", err)
	}
	if len(eligible) != 1 {
		t.Fatalf("expected 1 eligible inquiry, got %d", len(eligible))
	}

	created, err := store.CreateSynthesis(ctx, model.SynthesisInput{
		Title:     "First checkpoint",
		Body:      "The inquiry now has enough material to state a current view and name the remaining tension.",
		Type:      string(model.SynthesisTypeCheckpoint),
		InquiryID: inquiry.ID,
		Notes:     "Drafted after three linked engagements.",
	})
	if err != nil {
		t.Fatalf("CreateSynthesis() error = %v", err)
	}

	if created.Inquiry == nil || created.Inquiry.ID != inquiry.ID {
		t.Fatalf("expected linked inquiry summary for created synthesis")
	}

	listed, err := store.ListInquirySyntheses(ctx, inquiry.ID, 10)
	if err != nil {
		t.Fatalf("ListInquirySyntheses() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 inquiry synthesis, got %d", len(listed))
	}

	inquiryDetail, err := store.GetInquiry(ctx, inquiry.ID)
	if err != nil {
		t.Fatalf("GetInquiry() error = %v", err)
	}
	if inquiryDetail.SynthesisCount != 1 {
		t.Fatalf("expected synthesis count 1, got %d", inquiryDetail.SynthesisCount)
	}

	updated, err := store.UpdateSynthesis(ctx, created.ID, model.SynthesisInput{
		Title:     "Revised checkpoint",
		Body:      "A clearer compression of the current inquiry state.",
		Type:      string(model.SynthesisTypePosition),
		InquiryID: inquiry.ID,
		Notes:     "Tightened after rereading the claims.",
	})
	if err != nil {
		t.Fatalf("UpdateSynthesis() error = %v", err)
	}
	if updated.Type != string(model.SynthesisTypePosition) {
		t.Fatalf("expected updated type %q, got %q", model.SynthesisTypePosition, updated.Type)
	}

	eligible, err = store.ListEligibleForSynthesisInquiries(ctx, 10)
	if err != nil {
		t.Fatalf("ListEligibleForSynthesisInquiries() error = %v", err)
	}
	if len(eligible) != 0 {
		t.Fatalf("expected no eligible inquiries after synthesis, got %d", len(eligible))
	}

	if err := store.ArchiveSynthesis(ctx, created.ID); err != nil {
		t.Fatalf("ArchiveSynthesis() error = %v", err)
	}
	if _, err := store.GetSynthesis(ctx, created.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after archive, got %v", err)
	}
}
