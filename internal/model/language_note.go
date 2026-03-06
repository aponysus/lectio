package model

type LanguageNoteType string

const (
	LanguageNoteTypeTranslation    LanguageNoteType = "TRANSLATION"
	LanguageNoteTypeRegister       LanguageNoteType = "REGISTER"
	LanguageNoteTypeIdiom          LanguageNoteType = "IDIOM"
	LanguageNoteTypeCollocation    LanguageNoteType = "COLLOCATION"
	LanguageNoteTypeCulturalNuance LanguageNoteType = "CULTURAL_NUANCE"
	LanguageNoteTypeOther          LanguageNoteType = "OTHER"
)

var LanguageNoteTypes = []LanguageNoteType{
	LanguageNoteTypeTranslation,
	LanguageNoteTypeRegister,
	LanguageNoteTypeIdiom,
	LanguageNoteTypeCollocation,
	LanguageNoteTypeCulturalNuance,
	LanguageNoteTypeOther,
}

type LanguageNote struct {
	ID           string  `json:"id"`
	EngagementID string  `json:"engagement_id"`
	Term         *string `json:"term,omitempty"`
	Language     *string `json:"language,omitempty"`
	NoteType     *string `json:"note_type,omitempty"`
	Content      string  `json:"content"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	ArchivedAt   *string `json:"archived_at,omitempty"`
}

type LanguageNoteInput struct {
	EngagementID string
	Term         string
	Language     string
	NoteType     string
	Content      string
}

func IsValidLanguageNoteType(value string) bool {
	for _, noteType := range LanguageNoteTypes {
		if string(noteType) == value {
			return true
		}
	}
	return false
}
