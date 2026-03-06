package model

type InquiryStatus string

const (
	InquiryStatusActive      InquiryStatus = "ACTIVE"
	InquiryStatusDormant     InquiryStatus = "DORMANT"
	InquiryStatusSynthesized InquiryStatus = "SYNTHESIZED"
	InquiryStatusAbandoned   InquiryStatus = "ABANDONED"
)

var InquiryStatuses = []InquiryStatus{
	InquiryStatusActive,
	InquiryStatusDormant,
	InquiryStatusSynthesized,
	InquiryStatusAbandoned,
}

type Inquiry struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Question        string  `json:"question"`
	Status          string  `json:"status"`
	WhyItMatters    *string `json:"why_it_matters,omitempty"`
	CurrentView     *string `json:"current_view,omitempty"`
	OpenTensions    *string `json:"open_tensions,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	ArchivedAt      *string `json:"archived_at,omitempty"`
	EngagementCount int     `json:"engagement_count"`
	ClaimCount      int     `json:"claim_count"`
	LatestActivity  *string `json:"latest_activity,omitempty"`
}

type InquirySummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Question string `json:"question"`
	Status   string `json:"status"`
}

type InquiryInput struct {
	Title        string
	Question     string
	Status       string
	WhyItMatters string
	CurrentView  string
	OpenTensions string
}

type InquiryFilters struct {
	Query           string
	Status          string
	Limit           int
	IncludeArchived bool
}

func IsValidInquiryStatus(value string) bool {
	for _, status := range InquiryStatuses {
		if string(status) == value {
			return true
		}
	}
	return false
}
