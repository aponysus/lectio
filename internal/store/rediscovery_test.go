package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestRediscoveryGenerationAndActions(t *testing.T) {
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
		Title:  "Rediscovery Test Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	activeInquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "What remains unresolved here?",
		Question: "Which claim still needs pressure?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry(active) error = %v", err)
	}

	revisitPriority := 5
	oldEngagement, err := store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:        source.ID,
		EngagedAt:       "2025-01-05T12:00:00Z",
		PortionLabel:    "Chapter 2",
		Reflection:      "An older engagement still worth revisiting.",
		RevisitPriority: &revisitPriority,
	})
	if err != nil {
		t.Fatalf("CreateEngagement(old) error = %v", err)
	}
	if err := store.ReplaceEngagementInquiries(ctx, oldEngagement.ID, []string{activeInquiry.ID}); err != nil {
		t.Fatalf("ReplaceEngagementInquiries(old) error = %v", err)
	}

	staleClaim, err := store.CreateClaim(ctx, model.ClaimInput{
		Text:               "The structure may be doing more than the explicit thesis admits.",
		ClaimType:          string(model.ClaimTypeInterpretation),
		Status:             string(model.ClaimStatusTentative),
		OriginEngagementID: oldEngagement.ID,
	}, []string{activeInquiry.ID})
	if err != nil {
		t.Fatalf("CreateClaim() error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE claims
		SET
			created_at = datetime('now', '-20 days'),
			updated_at = datetime('now', '-20 days')
		WHERE id = ?
	`, staleClaim.ID); err != nil {
		t.Fatalf("seed stale claim timestamps: %v", err)
	}

	unsynthesizedInquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "What pattern is the corpus converging on?",
		Question: "Across four encounters, what view is becoming stable?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry(unsynthesized) error = %v", err)
	}

	engagementDates := []string{
		"2025-01-01T10:00:00Z",
		"2025-01-12T10:00:00Z",
		"2025-01-23T10:00:00Z",
		"2025-02-03T10:00:00Z",
	}
	for index, engagedAt := range engagementDates {
		engagement, err := store.CreateEngagement(ctx, model.EngagementInput{
			SourceID:     source.ID,
			EngagedAt:    engagedAt,
			PortionLabel: "Unsynthesized pass",
			Reflection:   "Enough material now exists to warrant synthesis.",
		})
		if err != nil {
			t.Fatalf("CreateEngagement(unsynthesized %d) error = %v", index, err)
		}
		if err := store.ReplaceEngagementInquiries(ctx, engagement.ID, []string{unsynthesizedInquiry.ID}); err != nil {
			t.Fatalf("ReplaceEngagementInquiries(unsynthesized %d) error = %v", index, err)
		}
	}

	reactivatedInquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "Can this dormant question be re-opened?",
		Question: "What changed enough to make this worth active attention again?",
		Status:   string(model.InquiryStatusDormant),
	})
	if err != nil {
		t.Fatalf("CreateInquiry(reactivated) error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE inquiries
		SET
			created_at = datetime('now', '-40 days'),
			updated_at = datetime('now', '-40 days')
		WHERE id = ?
	`, reactivatedInquiry.ID); err != nil {
		t.Fatalf("seed reactivated inquiry timestamps: %v", err)
	}
	if _, err := store.UpdateInquiry(ctx, reactivatedInquiry.ID, model.InquiryInput{
		Title:    reactivatedInquiry.Title,
		Question: reactivatedInquiry.Question,
		Status:   string(model.InquiryStatusActive),
	}); err != nil {
		t.Fatalf("UpdateInquiry(reactivate) error = %v", err)
	}

	items, err := store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems() error = %v", err)
	}

	if len(items) < 4 {
		t.Fatalf("expected at least 4 rediscovery items, got %d", len(items))
	}

	kinds := map[string]bool{}
	var staleItem model.RediscoveryItem
	var staleFound bool
	var unsynthesizedItem model.RediscoveryItem
	var unsynthesizedFound bool

	for _, item := range items {
		kinds[item.Kind] = true
		switch item.Kind {
		case string(model.RediscoveryKindStaleTentativeClaim):
			if item.TargetID == staleClaim.ID {
				staleItem = item
				staleFound = true
			}
		case string(model.RediscoveryKindUnsynthesizedInquiry):
			if item.TargetID == unsynthesizedInquiry.ID {
				unsynthesizedItem = item
				unsynthesizedFound = true
			}
		}
	}

	expectedKinds := []string{
		string(model.RediscoveryKindStaleTentativeClaim),
		string(model.RediscoveryKindActiveInquiryOldEntry),
		string(model.RediscoveryKindUnsynthesizedInquiry),
		string(model.RediscoveryKindRecentReactivation),
	}
	for _, kind := range expectedKinds {
		if !kinds[kind] {
			t.Fatalf("expected rediscovery kind %q to be present", kind)
		}
	}

	if !staleFound {
		t.Fatalf("expected stale claim item for claim %s", staleClaim.ID)
	}
	if staleItem.Claim == nil {
		t.Fatalf("expected stale claim item to include claim target payload")
	}
	if staleItem.LinkedInquiry == nil || staleItem.LinkedInquiry.ID != activeInquiry.ID {
		t.Fatalf("expected stale claim item to include linked inquiry %s", activeInquiry.ID)
	}

	if err := store.DismissRediscoveryItem(ctx, staleItem.ID); err != nil {
		t.Fatalf("DismissRediscoveryItem() error = %v", err)
	}

	items, err = store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems() after dismiss error = %v", err)
	}
	for _, item := range items {
		if item.Kind == string(model.RediscoveryKindStaleTentativeClaim) && item.TargetID == staleClaim.ID {
			t.Fatalf("dismissed stale claim item reappeared immediately")
		}
	}

	if !unsynthesizedFound {
		t.Fatalf("expected unsynthesized inquiry item for inquiry %s", unsynthesizedInquiry.ID)
	}

	if _, err := store.CreateSynthesis(ctx, model.SynthesisInput{
		Title:     "Checkpoint synthesis",
		Body:      "The material is now compressed into a first pass.",
		Type:      string(model.SynthesisTypeCheckpoint),
		InquiryID: unsynthesizedInquiry.ID,
	}); err != nil {
		t.Fatalf("CreateSynthesis() error = %v", err)
	}

	items, err = store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems() after synthesis error = %v", err)
	}
	for _, item := range items {
		if item.Kind == string(model.RediscoveryKindUnsynthesizedInquiry) && item.TargetID == unsynthesizedInquiry.ID {
			t.Fatalf("unsynthesized inquiry item remained after synthesis creation")
		}
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM rediscovery_items WHERE id = ?`, unsynthesizedItem.ID).Scan(&status); err != nil {
		t.Fatalf("load rediscovery item status: %v", err)
	}
	if status != string(model.RediscoveryStatusActedOn) {
		t.Fatalf("expected resolved rediscovery item status %q, got %q", model.RediscoveryStatusActedOn, status)
	}
}
