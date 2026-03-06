package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aponysus/lectio/internal/model"
	"github.com/aponysus/lectio/internal/store"
)

type ImportV2Result struct {
	SourcesCreated     int
	EngagementsCreated int
}

type legacyV2Payload struct {
	Sources []legacyV2Source `json:"sources"`
	Entries []legacyV2Entry  `json:"entries"`
}

type legacyV2Source struct {
	ID        legacyID `json:"id"`
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	Year      *int     `json:"year"`
	Tradition string   `json:"tradition"`
	Language  string   `json:"language"`
}

type legacyV2Entry struct {
	ID        legacyID             `json:"id"`
	SourceID  legacyID             `json:"source_id"`
	Source    *legacyV2EntrySource `json:"source"`
	Passage   string               `json:"passage"`
	Reflection string              `json:"reflection"`
	CreatedAt string               `json:"created_at"`
	UpdatedAt string               `json:"updated_at"`
}

type legacyV2EntrySource struct {
	ID    legacyID `json:"id"`
	Title string   `json:"title"`
}

type legacyID string

func (id *legacyID) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		*id = ""
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(trimmed, &stringValue); err == nil {
		*id = legacyID(strings.TrimSpace(stringValue))
		return nil
	}

	var numberValue json.Number
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&numberValue); err == nil {
		*id = legacyID(numberValue.String())
		return nil
	}

	return fmt.Errorf("legacy id must be string, number, or null")
}

func ImportV2File(ctx context.Context, repo *store.Store, path string) (ImportV2Result, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ImportV2Result{}, err
	}

	var payload legacyV2Payload
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return ImportV2Result{}, err
	}

	result := ImportV2Result{}
	sourceIndex := map[string]string{}

	for _, legacySource := range payload.Sources {
		if strings.TrimSpace(legacySource.Title) == "" {
			continue
		}

		key := sourceKey(string(legacySource.ID), legacySource.Title)
		if key != "" {
			if _, exists := sourceIndex[key]; exists {
				continue
			}
		}

		created, err := repo.CreateSource(ctx, model.SourceInput{
			Title:            strings.TrimSpace(legacySource.Title),
			Medium:           string(model.SourceMediumOther),
			Creator:          strings.TrimSpace(legacySource.Author),
			Year:             legacySource.Year,
			OriginalLanguage: strings.TrimSpace(legacySource.Language),
			CultureOrContext: strings.TrimSpace(legacySource.Tradition),
		})
		if err != nil {
			return ImportV2Result{}, err
		}

		if key != "" {
			sourceIndex[key] = created.ID
		}
		result.SourcesCreated++
	}

	for _, entry := range payload.Entries {
		sourceID, createdSource, err := ensureLegacySource(ctx, repo, sourceIndex, entry)
		if err != nil {
			return ImportV2Result{}, err
		}
		if createdSource {
			result.SourcesCreated++
		}

		reflection := strings.TrimSpace(entry.Reflection)
		passage := strings.TrimSpace(entry.Passage)
		switch {
		case reflection == "" && passage == "":
			return ImportV2Result{}, fmt.Errorf("legacy entry %q has neither reflection nor passage", entry.ID)
		case reflection == "":
			reflection = passage
		case passage != "":
			reflection = reflection + "\n\nLegacy passage:\n" + passage
		}

		if _, err := repo.CreateEngagement(ctx, model.EngagementInput{
			SourceID:   sourceID,
			EngagedAt:  normalizeLegacyTimestamp(entry.CreatedAt, entry.UpdatedAt),
			Reflection: reflection,
		}); err != nil {
			return ImportV2Result{}, err
		}

		result.EngagementsCreated++
	}

	return result, nil
}

func ensureLegacySource(
	ctx context.Context,
	repo *store.Store,
	sourceIndex map[string]string,
	entry legacyV2Entry,
) (string, bool, error) {
	legacySourceID := string(entry.SourceID)
	legacyTitle := ""
	if entry.Source != nil {
		if legacySourceID == "" {
			legacySourceID = string(entry.Source.ID)
		}
		legacyTitle = entry.Source.Title
	}

	key := sourceKey(legacySourceID, legacyTitle)
	if key != "" {
		if currentID, exists := sourceIndex[key]; exists {
			return currentID, false, nil
		}
	}

	title := strings.TrimSpace(legacyTitle)
	if title == "" {
		if legacySourceID != "" {
			title = "Legacy Source " + legacySourceID
		} else {
			title = "Legacy Imported Source"
		}
	}

	created, err := repo.CreateSource(ctx, model.SourceInput{
		Title:  title,
		Medium: string(model.SourceMediumOther),
	})
	if err != nil {
		return "", false, err
	}

	if key == "" {
		key = sourceKey(legacySourceID, title)
	}
	sourceIndex[key] = created.ID

	return created.ID, true, nil
}

func sourceKey(legacyIDValue, title string) string {
	switch {
	case strings.TrimSpace(legacyIDValue) != "":
		return "id:" + strings.TrimSpace(legacyIDValue)
	case strings.TrimSpace(title) != "":
		return "title:" + strings.ToLower(strings.TrimSpace(title))
	default:
		return ""
	}
}

func normalizeLegacyTimestamp(createdAt, updatedAt string) string {
	candidates := []string{
		strings.TrimSpace(createdAt),
		strings.TrimSpace(updatedAt),
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if parsed, err := time.Parse(time.RFC3339, candidate); err == nil {
			return parsed.UTC().Format(time.RFC3339)
		}
	}

	return time.Now().UTC().Format(time.RFC3339)
}
