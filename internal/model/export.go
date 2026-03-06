package model

type ExportPayload struct {
	FormatVersion       int                     `json:"format_version"`
	ExportedAt          string                  `json:"exported_at"`
	Sources             []Source                `json:"sources"`
	Engagements         []ExportEngagement      `json:"engagements"`
	Inquiries           []ExportInquiry         `json:"inquiries"`
	EngagementInquiries []EngagementInquiryLink `json:"engagement_inquiries"`
	Claims              []ExportClaim           `json:"claims"`
	ClaimInquiries      []ClaimInquiryLink      `json:"claim_inquiries"`
	LanguageNotes       []LanguageNote          `json:"language_notes"`
	Syntheses           []ExportSynthesis       `json:"syntheses"`
	RediscoveryItems    []ExportRediscoveryItem `json:"rediscovery_items,omitempty"`
}

type ExportEngagement struct {
	ID                 string  `json:"id"`
	SourceID           string  `json:"source_id"`
	EngagedAt          string  `json:"engaged_at"`
	PortionLabel       *string `json:"portion_label,omitempty"`
	Reflection         string  `json:"reflection"`
	WhyItMatters       *string `json:"why_it_matters,omitempty"`
	SourceLanguage     *string `json:"source_language,omitempty"`
	ReflectionLanguage *string `json:"reflection_language,omitempty"`
	AccessMode         *string `json:"access_mode,omitempty"`
	RevisitPriority    *int    `json:"revisit_priority,omitempty"`
	IsRereadOrRewatch  bool    `json:"is_reread_or_rewatch"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	ArchivedAt         *string `json:"archived_at,omitempty"`
}

type ExportInquiry struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Question      string  `json:"question"`
	Status        string  `json:"status"`
	WhyItMatters  *string `json:"why_it_matters,omitempty"`
	CurrentView   *string `json:"current_view,omitempty"`
	OpenTensions  *string `json:"open_tensions,omitempty"`
	ReactivatedAt *string `json:"reactivated_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	ArchivedAt    *string `json:"archived_at,omitempty"`
}

type EngagementInquiryLink struct {
	EngagementID string `json:"engagement_id"`
	InquiryID    string `json:"inquiry_id"`
	CreatedAt    string `json:"created_at"`
}

type ExportClaim struct {
	ID                 string  `json:"id"`
	Text               string  `json:"text"`
	ClaimType          string  `json:"claim_type"`
	Confidence         *int    `json:"confidence,omitempty"`
	Status             string  `json:"status"`
	OriginEngagementID *string `json:"origin_engagement_id,omitempty"`
	Notes              *string `json:"notes,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	ArchivedAt         *string `json:"archived_at,omitempty"`
}

type ClaimInquiryLink struct {
	ClaimID   string `json:"claim_id"`
	InquiryID string `json:"inquiry_id"`
	CreatedAt string `json:"created_at"`
}

type ExportSynthesis struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	Type       string  `json:"type"`
	InquiryID  string  `json:"inquiry_id"`
	Notes      *string `json:"notes,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	ArchivedAt *string `json:"archived_at,omitempty"`
}

type ExportRediscoveryItem struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
