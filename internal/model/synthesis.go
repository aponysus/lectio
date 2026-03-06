package model

type SynthesisType string

const (
	SynthesisTypeCheckpoint SynthesisType = "CHECKPOINT"
	SynthesisTypeComparison SynthesisType = "COMPARISON"
	SynthesisTypePosition   SynthesisType = "POSITION"
)

var SynthesisTypes = []SynthesisType{
	SynthesisTypeCheckpoint,
	SynthesisTypeComparison,
	SynthesisTypePosition,
}

type Synthesis struct {
	ID         string          `json:"id"`
	Title      string          `json:"title"`
	Body       string          `json:"body"`
	Type       string          `json:"type"`
	InquiryID  string          `json:"inquiry_id"`
	Notes      *string         `json:"notes,omitempty"`
	CreatedAt  string          `json:"created_at"`
	UpdatedAt  string          `json:"updated_at"`
	ArchivedAt *string         `json:"archived_at,omitempty"`
	Inquiry    *InquirySummary `json:"inquiry,omitempty"`
}

type SynthesisInput struct {
	Title     string
	Body      string
	Type      string
	InquiryID string
	Notes     string
}

type SynthesisFilters struct {
	InquiryID string
	Limit     int
}

func IsValidSynthesisType(value string) bool {
	for _, synthesisType := range SynthesisTypes {
		if string(synthesisType) == value {
			return true
		}
	}
	return false
}
