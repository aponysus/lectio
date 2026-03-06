package validation

import (
	"strings"
	"time"

	"github.com/aponysus/lectio/internal/model"
)

func NormalizeEngagementInput(input model.EngagementInput) (model.EngagementInput, error) {
	input.SourceID = strings.TrimSpace(input.SourceID)
	input.PortionLabel = strings.TrimSpace(input.PortionLabel)
	input.Reflection = strings.TrimSpace(input.Reflection)
	input.WhyItMatters = strings.TrimSpace(input.WhyItMatters)
	input.SourceLanguage = strings.TrimSpace(input.SourceLanguage)
	input.ReflectionLanguage = strings.TrimSpace(input.ReflectionLanguage)
	input.AccessMode = strings.ToUpper(strings.TrimSpace(input.AccessMode))

	switch {
	case input.SourceID == "":
		return model.EngagementInput{}, FieldError{Field: "source_id", Message: "is required"}
	case input.EngagedAt == "":
		return model.EngagementInput{}, FieldError{Field: "engaged_at", Message: "is required"}
	case input.Reflection == "":
		return model.EngagementInput{}, FieldError{Field: "reflection", Message: "is required"}
	case len(input.Reflection) > 10000:
		return model.EngagementInput{}, FieldError{Field: "reflection", Message: "must be 10000 characters or fewer"}
	case len(input.PortionLabel) > 240:
		return model.EngagementInput{}, FieldError{Field: "portion_label", Message: "must be 240 characters or fewer"}
	case len(input.WhyItMatters) > 4000:
		return model.EngagementInput{}, FieldError{Field: "why_it_matters", Message: "must be 4000 characters or fewer"}
	case len(input.SourceLanguage) > 64:
		return model.EngagementInput{}, FieldError{Field: "source_language", Message: "must be 64 characters or fewer"}
	case len(input.ReflectionLanguage) > 64:
		return model.EngagementInput{}, FieldError{Field: "reflection_language", Message: "must be 64 characters or fewer"}
	case input.AccessMode != "" && !model.IsValidAccessMode(input.AccessMode):
		return model.EngagementInput{}, FieldError{Field: "access_mode", Message: "is invalid"}
	}

	if input.RevisitPriority != nil && (*input.RevisitPriority < 1 || *input.RevisitPriority > 5) {
		return model.EngagementInput{}, FieldError{Field: "revisit_priority", Message: "must be between 1 and 5"}
	}

	engagedAt, err := time.Parse(time.RFC3339, input.EngagedAt)
	if err != nil {
		return model.EngagementInput{}, FieldError{Field: "engaged_at", Message: "must be a valid RFC3339 timestamp"}
	}

	input.EngagedAt = engagedAt.UTC().Format(time.RFC3339)
	return input, nil
}

func NormalizeEngagementFilters(filters model.EngagementFilters) (model.EngagementFilters, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.SourceID = strings.TrimSpace(filters.SourceID)
	filters.AccessMode = strings.ToUpper(strings.TrimSpace(filters.AccessMode))

	if filters.AccessMode != "" && !model.IsValidAccessMode(filters.AccessMode) {
		return model.EngagementFilters{}, FieldError{Field: "access_mode", Message: "is invalid"}
	}

	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	return filters, nil
}
