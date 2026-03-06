package validation

import (
	"strings"

	"github.com/aponysus/lectio/internal/model"
)

func NormalizeLanguageNoteInput(input model.LanguageNoteInput) (model.LanguageNoteInput, error) {
	input.EngagementID = strings.TrimSpace(input.EngagementID)
	input.Term = strings.TrimSpace(input.Term)
	input.Language = strings.TrimSpace(input.Language)
	input.NoteType = strings.ToUpper(strings.TrimSpace(input.NoteType))
	input.Content = strings.TrimSpace(input.Content)

	switch {
	case len(input.Term) > 240:
		return model.LanguageNoteInput{}, FieldError{Field: "term", Message: "must be 240 characters or fewer"}
	case len(input.Language) > 64:
		return model.LanguageNoteInput{}, FieldError{Field: "language", Message: "must be 64 characters or fewer"}
	case input.NoteType != "" && !model.IsValidLanguageNoteType(input.NoteType):
		return model.LanguageNoteInput{}, FieldError{Field: "note_type", Message: "is invalid"}
	case input.Content == "":
		return model.LanguageNoteInput{}, FieldError{Field: "content", Message: "is required"}
	case len(input.Content) > 4000:
		return model.LanguageNoteInput{}, FieldError{Field: "content", Message: "must be 4000 characters or fewer"}
	}

	return input, nil
}
