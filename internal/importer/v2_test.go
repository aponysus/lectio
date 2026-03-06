package importer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aponysus/lectio/internal/model"
	"github.com/aponysus/lectio/internal/store"
)

func TestImportV2File(t *testing.T) {
	t.Parallel()

	db, err := store.Open(filepath.Join(t.TempDir(), "lectio.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := store.ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	repo := store.New(db)
	importPath := filepath.Join(t.TempDir(), "legacy-v2.json")
	if err := os.WriteFile(importPath, []byte(`{
  "sources": [
    {
      "id": 12,
      "title": "Legacy Source",
      "author": "V2 Author",
      "year": 2021,
      "tradition": "Patristic",
      "language": "en"
    }
  ],
  "entries": [
    {
      "id": 91,
      "source_id": 12,
      "passage": "A remembered passage.",
      "reflection": "A carried-forward reflection.",
      "created_at": "2025-03-01T12:00:00Z"
    }
  ]
}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := ImportV2File(ctx, repo, importPath)
	if err != nil {
		t.Fatalf("ImportV2File() error = %v", err)
	}

	if result.SourcesCreated != 1 {
		t.Fatalf("expected 1 source created, got %d", result.SourcesCreated)
	}
	if result.EngagementsCreated != 1 {
		t.Fatalf("expected 1 engagement created, got %d", result.EngagementsCreated)
	}

	sources, err := repo.ListSources(ctx, model.SourceFilters{Limit: 10})
	if err != nil {
		t.Fatalf("ListSources() error = %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].Creator == nil || *sources[0].Creator != "V2 Author" {
		t.Fatalf("expected imported creator, got %+v", sources[0].Creator)
	}

	engagements, err := repo.ListEngagements(ctx, model.EngagementFilters{Limit: 10})
	if err != nil {
		t.Fatalf("ListEngagements() error = %v", err)
	}
	if len(engagements) != 1 {
		t.Fatalf("expected 1 engagement, got %d", len(engagements))
	}
	if !strings.Contains(engagements[0].Reflection, "Legacy passage:") {
		t.Fatalf("expected passage folded into reflection, got %q", engagements[0].Reflection)
	}
	if engagements[0].EngagedAt != "2025-03-01T12:00:00Z" {
		t.Fatalf("expected legacy timestamp preserved, got %q", engagements[0].EngagedAt)
	}
}
