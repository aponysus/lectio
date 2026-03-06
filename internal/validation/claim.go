package validation

import (
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

func NormalizeClaimInput(input model.ClaimInput) (model.ClaimInput, error) {
	input.Text = strings.TrimSpace(input.Text)
	input.ClaimType = strings.ToUpper(strings.TrimSpace(input.ClaimType))
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.OriginEngagementID = strings.TrimSpace(input.OriginEngagementID)
	input.Notes = strings.TrimSpace(input.Notes)

	switch {
	case input.Text == "":
		return model.ClaimInput{}, FieldError{Field: "text", Message: "is required"}
	case len(input.Text) > 4000:
		return model.ClaimInput{}, FieldError{Field: "text", Message: "must be 4000 characters or fewer"}
	case input.ClaimType == "":
		return model.ClaimInput{}, FieldError{Field: "claim_type", Message: "is required"}
	case !model.IsValidClaimType(input.ClaimType):
		return model.ClaimInput{}, FieldError{Field: "claim_type", Message: "is invalid"}
	case input.Status == "":
		return model.ClaimInput{}, FieldError{Field: "status", Message: "is required"}
	case !model.IsValidClaimStatus(input.Status):
		return model.ClaimInput{}, FieldError{Field: "status", Message: "is invalid"}
	case len(input.Notes) > 4000:
		return model.ClaimInput{}, FieldError{Field: "notes", Message: "must be 4000 characters or fewer"}
	}

	if input.Confidence != nil && (*input.Confidence < 1 || *input.Confidence > 5) {
		return model.ClaimInput{}, FieldError{Field: "confidence", Message: "must be between 1 and 5"}
	}

	return input, nil
}
