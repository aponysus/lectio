package model

type ClaimType string

const (
	ClaimTypeObservation    ClaimType = "OBSERVATION"
	ClaimTypeInterpretation ClaimType = "INTERPRETATION"
	ClaimTypePersonalView   ClaimType = "PERSONAL_VIEW"
	ClaimTypeQuestion       ClaimType = "QUESTION"
	ClaimTypeHypothesis     ClaimType = "HYPOTHESIS"
)

var ClaimTypes = []ClaimType{
	ClaimTypeObservation,
	ClaimTypeInterpretation,
	ClaimTypePersonalView,
	ClaimTypeQuestion,
	ClaimTypeHypothesis,
}

type ClaimStatus string

const (
	ClaimStatusActive    ClaimStatus = "ACTIVE"
	ClaimStatusTentative ClaimStatus = "TENTATIVE"
	ClaimStatusRevised   ClaimStatus = "REVISED"
	ClaimStatusAbandoned ClaimStatus = "ABANDONED"
)

var ClaimStatuses = []ClaimStatus{
	ClaimStatusActive,
	ClaimStatusTentative,
	ClaimStatusRevised,
	ClaimStatusAbandoned,
}

type ClaimOrigin struct {
	EngagementID string  `json:"engagement_id"`
	SourceID     string  `json:"source_id"`
	SourceTitle  string  `json:"source_title"`
	SourceMedium string  `json:"source_medium"`
	PortionLabel *string `json:"portion_label,omitempty"`
}

type Claim struct {
	ID                 string       `json:"id"`
	Text               string       `json:"text"`
	ClaimType          string       `json:"claim_type"`
	Confidence         *int         `json:"confidence,omitempty"`
	Status             string       `json:"status"`
	OriginEngagementID *string      `json:"origin_engagement_id,omitempty"`
	Notes              *string      `json:"notes,omitempty"`
	CreatedAt          string       `json:"created_at"`
	UpdatedAt          string       `json:"updated_at"`
	ArchivedAt         *string      `json:"archived_at,omitempty"`
	Origin             *ClaimOrigin `json:"origin,omitempty"`
}

type ClaimInput struct {
	Text               string
	ClaimType          string
	Confidence         *int
	Status             string
	OriginEngagementID string
	Notes              string
}

func IsValidClaimType(value string) bool {
	for _, claimType := range ClaimTypes {
		if string(claimType) == value {
			return true
		}
	}
	return false
}

func IsValidClaimStatus(value string) bool {
	for _, status := range ClaimStatuses {
		if string(status) == value {
			return true
		}
	}
	return false
}
