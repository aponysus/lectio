package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestCreateEngagementCaptureCreatesNestedRecords(t *testing.T) {
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
		Title:  "Transactional Capture Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	existingInquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "What does the first inquiry already hold?",
		Question: "How should this engagement deepen an existing question?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}

	engagement, err := store.CreateEngagementCapture(ctx, model.EngagementCaptureInput{
		Engagement: model.EngagementInput{
			SourceID:           source.ID,
			EngagedAt:          "2026-03-06T15:00:00Z",
			PortionLabel:       "Chapter 4",
			Reflection:         "The transactional create flow should leave behind a coherent inquiry graph.",
			WhyItMatters:       "It is the main MVP capture loop.",
			SourceLanguage:     "en",
			ReflectionLanguage: "en",
			AccessMode:         string(model.AccessModeOriginal),
		},
		InquiryIDs: []string{existingInquiry.ID},
		InlineInquiries: []model.InquiryInput{
			{
				Title:        "What new pressure did the text introduce?",
				Question:     "Which unresolved tension only became visible in this engagement?",
				Status:       string(model.InquiryStatusActive),
				WhyItMatters: "This is the inline-created inquiry case.",
			},
		},
		Claims: []model.ClaimInput{
			{
				Text:      "The engagement flow needs one command boundary, not five independent mutations.",
				ClaimType: string(model.ClaimTypeInterpretation),
				Status:    string(model.ClaimStatusActive),
			},
			{
				Text:      "Inline capture is only trustworthy if linked records survive as one unit.",
				ClaimType: string(model.ClaimTypeObservation),
				Status:    string(model.ClaimStatusTentative),
			},
		},
		LanguageNotes: []model.LanguageNoteInput{
			{
				Term:     "praxis",
				Language: "grc",
				NoteType: string(model.LanguageNoteTypeTranslation),
				Content:  "Used as a transactional regression marker.",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateEngagementCapture() error = %v", err)
	}

	if engagement.SourceID != source.ID {
		t.Fatalf("expected engagement source %q, got %q", source.ID, engagement.SourceID)
	}

	linkedInquiries, err := store.ListEngagementInquiries(ctx, engagement.ID)
	if err != nil {
		t.Fatalf("ListEngagementInquiries() error = %v", err)
	}
	if len(linkedInquiries) != 2 {
		t.Fatalf("expected 2 linked inquiries, got %d", len(linkedInquiries))
	}

	engagementClaims, err := store.ListEngagementClaims(ctx, engagement.ID)
	if err != nil {
		t.Fatalf("ListEngagementClaims() error = %v", err)
	}
	if len(engagementClaims) != 2 {
		t.Fatalf("expected 2 linked claims, got %d", len(engagementClaims))
	}

	engagementNotes, err := store.ListEngagementLanguageNotes(ctx, engagement.ID)
	if err != nil {
		t.Fatalf("ListEngagementLanguageNotes() error = %v", err)
	}
	if len(engagementNotes) != 1 {
		t.Fatalf("expected 1 language note, got %d", len(engagementNotes))
	}

	existingInquiryClaims, err := store.ListInquiryClaims(ctx, existingInquiry.ID)
	if err != nil {
		t.Fatalf("ListInquiryClaims(existing) error = %v", err)
	}
	if len(existingInquiryClaims) != 2 {
		t.Fatalf("expected 2 claims on existing inquiry, got %d", len(existingInquiryClaims))
	}

	inquiries, err := store.ListInquiries(ctx, model.InquiryFilters{Limit: 10})
	if err != nil {
		t.Fatalf("ListInquiries() error = %v", err)
	}
	if len(inquiries) != 2 {
		t.Fatalf("expected 2 inquiries after inline create, got %d", len(inquiries))
	}
}

func TestCreateEngagementCaptureRollsBackOnDuplicateInquiryLink(t *testing.T) {
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
		Title:  "Rollback Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	inquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "Will the transaction roll back?",
		Question: "If the join insert fails, does the engagement disappear too?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}

	_, err = store.CreateEngagementCapture(ctx, model.EngagementCaptureInput{
		Engagement: model.EngagementInput{
			SourceID:   source.ID,
			EngagedAt:  "2026-03-06T17:00:00Z",
			Reflection: "This write should fail on the duplicate inquiry link.",
		},
		InquiryIDs: []string{inquiry.ID, inquiry.ID},
	})
	if err == nil {
		t.Fatal("expected CreateEngagementCapture() to fail on duplicate inquiry link")
	}

	engagements, err := store.ListEngagements(ctx, model.EngagementFilters{Limit: 10})
	if err != nil {
		t.Fatalf("ListEngagements() error = %v", err)
	}
	if len(engagements) != 0 {
		t.Fatalf("expected rollback to leave 0 engagements, got %d", len(engagements))
	}
}
