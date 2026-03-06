package model

type AccessMode string

const (
	AccessModeOriginal    AccessMode = "ORIGINAL"
	AccessModeTranslation AccessMode = "TRANSLATION"
	AccessModeBilingual   AccessMode = "BILINGUAL"
	AccessModeSubtitled   AccessMode = "SUBTITLED"
	AccessModeLookupHeavy AccessMode = "LOOKUP_HEAVY"
	AccessModeOther       AccessMode = "OTHER"
)

var AccessModes = []AccessMode{
	AccessModeOriginal,
	AccessModeTranslation,
	AccessModeBilingual,
	AccessModeSubtitled,
	AccessModeLookupHeavy,
	AccessModeOther,
}

type SourceSummary struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Medium  string  `json:"medium"`
	Creator *string `json:"creator,omitempty"`
}

type Engagement struct {
	ID                 string        `json:"id"`
	SourceID           string        `json:"source_id"`
	EngagedAt          string        `json:"engaged_at"`
	PortionLabel       *string       `json:"portion_label,omitempty"`
	Reflection         string        `json:"reflection"`
	WhyItMatters       *string       `json:"why_it_matters,omitempty"`
	SourceLanguage     *string       `json:"source_language,omitempty"`
	ReflectionLanguage *string       `json:"reflection_language,omitempty"`
	AccessMode         *string       `json:"access_mode,omitempty"`
	RevisitPriority    *int          `json:"revisit_priority,omitempty"`
	IsRereadOrRewatch  bool          `json:"is_reread_or_rewatch"`
	CreatedAt          string        `json:"created_at"`
	UpdatedAt          string        `json:"updated_at"`
	ArchivedAt         *string       `json:"archived_at,omitempty"`
	LanguageNoteCount  int           `json:"language_note_count"`
	Source             SourceSummary `json:"source"`
}

type EngagementInput struct {
	SourceID           string
	EngagedAt          string
	PortionLabel       string
	Reflection         string
	WhyItMatters       string
	SourceLanguage     string
	ReflectionLanguage string
	AccessMode         string
	RevisitPriority    *int
	IsRereadOrRewatch  bool
}

type EngagementCaptureInput struct {
	Engagement      EngagementInput
	InquiryIDs      []string
	InlineInquiries []InquiryInput
	Claims          []ClaimInput
	LanguageNotes   []LanguageNoteInput
}

type EngagementFilters struct {
	Query            string
	SourceID         string
	AccessMode       string
	HasLanguageNotes bool
	Limit            int
	IncludeArchived  bool
}

func IsValidAccessMode(value string) bool {
	for _, mode := range AccessModes {
		if string(mode) == value {
			return true
		}
	}
	return false
}
