package validation

import (
	"fmt"
	"strings"
	"time"

	"github.com/aponysus/lectio/internal/model"
)

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Message)
}

func NormalizeSourceInput(input model.SourceInput) (model.SourceInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Medium = strings.ToUpper(strings.TrimSpace(input.Medium))
	input.Creator = strings.TrimSpace(input.Creator)
	input.OriginalLanguage = strings.TrimSpace(input.OriginalLanguage)
	input.CultureOrContext = strings.TrimSpace(input.CultureOrContext)
	input.Notes = strings.TrimSpace(input.Notes)

	switch {
	case input.Title == "":
		return model.SourceInput{}, FieldError{Field: "title", Message: "is required"}
	case len(input.Title) > 240:
		return model.SourceInput{}, FieldError{Field: "title", Message: "must be 240 characters or fewer"}
	case input.Medium == "":
		return model.SourceInput{}, FieldError{Field: "medium", Message: "is required"}
	case !model.IsValidSourceMedium(input.Medium):
		return model.SourceInput{}, FieldError{Field: "medium", Message: "is invalid"}
	case len(input.Creator) > 240:
		return model.SourceInput{}, FieldError{Field: "creator", Message: "must be 240 characters or fewer"}
	case len(input.OriginalLanguage) > 64:
		return model.SourceInput{}, FieldError{Field: "original_language", Message: "must be 64 characters or fewer"}
	case len(input.CultureOrContext) > 160:
		return model.SourceInput{}, FieldError{Field: "culture_or_context", Message: "must be 160 characters or fewer"}
	case len(input.Notes) > 4000:
		return model.SourceInput{}, FieldError{Field: "notes", Message: "must be 4000 characters or fewer"}
	}

	if input.Year != nil {
		currentYear := time.Now().UTC().Year() + 5
		if *input.Year < -4000 || *input.Year > currentYear {
			return model.SourceInput{}, FieldError{Field: "year", Message: "must be a reasonable integer"}
		}
	}

	return input, nil
}

func NormalizeSourceFilters(filters model.SourceFilters) (model.SourceFilters, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Medium = strings.ToUpper(strings.TrimSpace(filters.Medium))
	filters.OriginalLanguage = strings.TrimSpace(filters.OriginalLanguage)

	switch filters.Sort {
	case "", model.SourceSortRecent:
		filters.Sort = model.SourceSortRecent
	case model.SourceSortTitle:
	default:
		return model.SourceFilters{}, FieldError{Field: "sort", Message: "is invalid"}
	}

	if filters.Medium != "" && !model.IsValidSourceMedium(filters.Medium) {
		return model.SourceFilters{}, FieldError{Field: "medium", Message: "is invalid"}
	}

	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	return filters, nil
}
