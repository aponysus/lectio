package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aponysus/lectio/internal/model"
)

func TestRediscoveryArchiveRegression(t *testing.T) {
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
		Title:  "Archive Rediscovery Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	activeInquiry, err := store.CreateInquiry(ctx, model.InquiryInput{
		Title:    "Which prompts should survive archival?",
		Question: "Once a target is archived, does rediscovery stop treating it as active material?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry(active) error = %v", err)
	}

	oldEngagement, err := store.CreateEngagement(ctx, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2025-01-05T12:00:00Z",
		Reflection: "An older engagement that should produce an active inquiry / old entry prompt.",
	})
	if err != nil {
		t.Fatalf("CreateEngagement(old) error = %v", err)
	}
	if err := store.ReplaceEngagementInquiries(ctx, oldEngagement.ID, []string{activeInquiry.ID}); err != nil {
		t.Fatalf("ReplaceEngagementInquiries(old) error = %v", err)
	}

	staleClaim, err := store.CreateClaim(ctx, model.ClaimInput{
		Text:               "A stale tentative claim should stop surfacing once archived.",
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
		Title:    "Which synthesis prompt should return after archival?",
		Question: "If the only synthesis is archived, should the inquiry become synthesis-ready again?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry(unsynthesized) error = %v", err)
	}

	for _, engagedAt := range []string{
		"2025-01-01T10:00:00Z",
		"2025-01-10T10:00:00Z",
		"2025-01-19T10:00:00Z",
		"2025-01-28T10:00:00Z",
	} {
		engagement, err := store.CreateEngagement(ctx, model.EngagementInput{
			SourceID:   source.ID,
			EngagedAt:  engagedAt,
			Reflection: "This engagement exists to make the inquiry synthesis-ready.",
		})
		if err != nil {
			t.Fatalf("CreateEngagement(unsynthesized) error = %v", err)
		}
		if err := store.ReplaceEngagementInquiries(ctx, engagement.ID, []string{unsynthesizedInquiry.ID}); err != nil {
			t.Fatalf("ReplaceEngagementInquiries(unsynthesized) error = %v", err)
		}
	}

	items, err := store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems(initial) error = %v", err)
	}

	assertRediscoveryItemPresent(t, items, string(model.RediscoveryKindStaleTentativeClaim), staleClaim.ID)
	assertRediscoveryItemPresent(t, items, string(model.RediscoveryKindActiveInquiryOldEntry), oldEngagement.ID)
	assertRediscoveryItemPresent(t, items, string(model.RediscoveryKindUnsynthesizedInquiry), unsynthesizedInquiry.ID)

	if err := store.ArchiveClaim(ctx, staleClaim.ID); err != nil {
		t.Fatalf("ArchiveClaim() error = %v", err)
	}

	items, err = store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems(after claim archive) error = %v", err)
	}
	assertRediscoveryItemAbsent(t, items, string(model.RediscoveryKindStaleTentativeClaim), staleClaim.ID)

	if err := store.ArchiveInquiry(ctx, activeInquiry.ID); err != nil {
		t.Fatalf("ArchiveInquiry() error = %v", err)
	}

	items, err = store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems(after inquiry archive) error = %v", err)
	}
	assertRediscoveryItemAbsent(t, items, string(model.RediscoveryKindActiveInquiryOldEntry), oldEngagement.ID)

	synthesis, err := store.CreateSynthesis(ctx, model.SynthesisInput{
		Title:     "Archive regression synthesis",
		Body:      "The inquiry is temporarily compressed, which should suppress the unsynthesized prompt.",
		Type:      string(model.SynthesisTypeCheckpoint),
		InquiryID: unsynthesizedInquiry.ID,
	})
	if err != nil {
		t.Fatalf("CreateSynthesis() error = %v", err)
	}

	items, err = store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems(after synthesis create) error = %v", err)
	}
	assertRediscoveryItemAbsent(t, items, string(model.RediscoveryKindUnsynthesizedInquiry), unsynthesizedInquiry.ID)

	if err := store.ArchiveSynthesis(ctx, synthesis.ID); err != nil {
		t.Fatalf("ArchiveSynthesis() error = %v", err)
	}

	items, err = store.ListRediscoveryItems(ctx, 12)
	if err != nil {
		t.Fatalf("ListRediscoveryItems(after synthesis archive) error = %v", err)
	}
	assertRediscoveryItemPresent(t, items, string(model.RediscoveryKindUnsynthesizedInquiry), unsynthesizedInquiry.ID)
}

func assertRediscoveryItemPresent(t *testing.T, items []model.RediscoveryItem, kind, targetID string) {
	t.Helper()
	for _, item := range items {
		if item.Kind == kind && item.TargetID == targetID {
			return
		}
	}

	t.Fatalf("expected rediscovery item kind=%q target=%q to be present", kind, targetID)
}

func assertRediscoveryItemAbsent(t *testing.T, items []model.RediscoveryItem, kind, targetID string) {
	t.Helper()
	for _, item := range items {
		if item.Kind == kind && item.TargetID == targetID {
			t.Fatalf("expected rediscovery item kind=%q target=%q to be absent", kind, targetID)
		}
	}
}
