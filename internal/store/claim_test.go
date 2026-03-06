package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestClaimLifecycleAndLinking(t *testing.T) {
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
		Title:  "Claim Test Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	engagement, err := store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2026-03-06T12:00:00Z",
		Reflection: "A reflection that will sharpen into claims.",
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	inquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "What exactly follows from this passage?",
		Question: "Which claim survives closer scrutiny?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}

	confidence := 4
	claim, err := store.CreateClaim(ctx, model.ClaimInput{
		Text:               "The author treats uncertainty as a method rather than a defect.",
		ClaimType:          string(model.ClaimTypeInterpretation),
		Confidence:         &confidence,
		Status:             string(model.ClaimStatusTentative),
		OriginEngagementID: engagement.ID,
		Notes:              "Needs comparison against the closing section.",
	}, []string{inquiry.ID})
	if err != nil {
		t.Fatalf("CreateClaim() error = %v", err)
	}

	if claim.Origin == nil || claim.Origin.EngagementID != engagement.ID {
		t.Fatalf("expected origin engagement %q, got %+v", engagement.ID, claim.Origin)
	}

	inquiryClaims, err := store.ListInquiryClaims(ctx, inquiry.ID)
	if err != nil {
		t.Fatalf("ListInquiryClaims() error = %v", err)
	}
	if len(inquiryClaims) != 1 {
		t.Fatalf("expected 1 inquiry claim, got %d", len(inquiryClaims))
	}

	engagementClaims, err := store.ListEngagementClaims(ctx, engagement.ID)
	if err != nil {
		t.Fatalf("ListEngagementClaims() error = %v", err)
	}
	if len(engagementClaims) != 1 {
		t.Fatalf("expected 1 engagement claim, got %d", len(engagementClaims))
	}

	listedInquiries, err := store.ListInquiries(ctx, model.InquiryFilters{Limit: 10})
	if err != nil {
		t.Fatalf("ListInquiries() error = %v", err)
	}
	if len(listedInquiries) != 1 || listedInquiries[0].ClaimCount != 1 {
		t.Fatalf("expected claim count 1, got %+v", listedInquiries)
	}

	updated, err := store.UpdateClaim(ctx, claim.ID, model.ClaimInput{
		Text:               "The author treats uncertainty as a disciplined method.",
		ClaimType:          string(model.ClaimTypeInterpretation),
		Confidence:         &confidence,
		Status:             string(model.ClaimStatusRevised),
		OriginEngagementID: engagement.ID,
		Notes:              "Revised after rereading.",
	})
	if err != nil {
		t.Fatalf("UpdateClaim() error = %v", err)
	}
	if updated.Status != string(model.ClaimStatusRevised) {
		t.Fatalf("expected updated status %q, got %q", model.ClaimStatusRevised, updated.Status)
	}

	if err := store.ReplaceClaimInquiries(ctx, claim.ID, nil); err != nil {
		t.Fatalf("ReplaceClaimInquiries() error = %v", err)
	}
	inquiryClaims, err = store.ListInquiryClaims(ctx, inquiry.ID)
	if err != nil {
		t.Fatalf("ListInquiryClaims() error = %v", err)
	}
	if len(inquiryClaims) != 0 {
		t.Fatalf("expected 0 inquiry claims after unlink, got %d", len(inquiryClaims))
	}

	if err := store.ArchiveClaim(ctx, claim.ID); err != nil {
		t.Fatalf("ArchiveClaim() error = %v", err)
	}
	if _, err := store.GetClaim(ctx, claim.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after archive, got %v", err)
	}
}
