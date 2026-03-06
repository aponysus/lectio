package validation

import (
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

func NormalizeInquiryInput(input model.InquiryInput) (model.InquiryInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Question = strings.TrimSpace(input.Question)
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.WhyItMatters = strings.TrimSpace(input.WhyItMatters)
	input.CurrentView = strings.TrimSpace(input.CurrentView)
	input.OpenTensions = strings.TrimSpace(input.OpenTensions)

	switch {
	case input.Title == "":
		return model.InquiryInput{}, FieldError{Field: "title", Message: "is required"}
	case len(input.Title) > 240:
		return model.InquiryInput{}, FieldError{Field: "title", Message: "must be 240 characters or fewer"}
	case input.Question == "":
		return model.InquiryInput{}, FieldError{Field: "question", Message: "is required"}
	case len(input.Question) > 4000:
		return model.InquiryInput{}, FieldError{Field: "question", Message: "must be 4000 characters or fewer"}
	case input.Status == "":
		return model.InquiryInput{}, FieldError{Field: "status", Message: "is required"}
	case !model.IsValidInquiryStatus(input.Status):
		return model.InquiryInput{}, FieldError{Field: "status", Message: "is invalid"}
	case len(input.WhyItMatters) > 4000:
		return model.InquiryInput{}, FieldError{Field: "why_it_matters", Message: "must be 4000 characters or fewer"}
	case len(input.CurrentView) > 4000:
		return model.InquiryInput{}, FieldError{Field: "current_view", Message: "must be 4000 characters or fewer"}
	case len(input.OpenTensions) > 4000:
		return model.InquiryInput{}, FieldError{Field: "open_tensions", Message: "must be 4000 characters or fewer"}
	}

	return input, nil
}

func NormalizeInquiryFilters(filters model.InquiryFilters) (model.InquiryFilters, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Status = strings.ToUpper(strings.TrimSpace(filters.Status))

	if filters.Status != "" && !model.IsValidInquiryStatus(filters.Status) {
		return model.InquiryFilters{}, FieldError{Field: "status", Message: "is invalid"}
	}

	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	return filters, nil
}

func NormalizeInquiryIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	normalized := make([]string, 0, len(ids))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}

	return normalized
}
