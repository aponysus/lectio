package validation

import (
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

func NormalizeSynthesisInput(input model.SynthesisInput) (model.SynthesisInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	input.Type = strings.ToUpper(strings.TrimSpace(input.Type))
	input.InquiryID = strings.TrimSpace(input.InquiryID)
	input.Notes = strings.TrimSpace(input.Notes)

	switch {
	case input.Title == "":
		return model.SynthesisInput{}, FieldError{Field: "title", Message: "is required"}
	case len(input.Title) > 240:
		return model.SynthesisInput{}, FieldError{Field: "title", Message: "must be 240 characters or fewer"}
	case input.Body == "":
		return model.SynthesisInput{}, FieldError{Field: "body", Message: "is required"}
	case len(input.Body) > 20000:
		return model.SynthesisInput{}, FieldError{Field: "body", Message: "must be 20000 characters or fewer"}
	case input.Type == "":
		return model.SynthesisInput{}, FieldError{Field: "type", Message: "is required"}
	case !model.IsValidSynthesisType(input.Type):
		return model.SynthesisInput{}, FieldError{Field: "type", Message: "is invalid"}
	case input.InquiryID == "":
		return model.SynthesisInput{}, FieldError{Field: "inquiry_id", Message: "is required"}
	case len(input.Notes) > 4000:
		return model.SynthesisInput{}, FieldError{Field: "notes", Message: "must be 4000 characters or fewer"}
	}

	return input, nil
}

func NormalizeSynthesisFilters(filters model.SynthesisFilters) (model.SynthesisFilters, error) {
	filters.InquiryID = strings.TrimSpace(filters.InquiryID)

	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	return filters, nil
}
