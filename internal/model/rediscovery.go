package model

type RediscoveryKind string

const (
	RediscoveryKindStaleTentativeClaim   RediscoveryKind = "stale_tentative_claim"
	RediscoveryKindActiveInquiryOldEntry RediscoveryKind = "active_inquiry_old_engagement"
	RediscoveryKindUnsynthesizedInquiry  RediscoveryKind = "unsynthesized_inquiry"
	RediscoveryKindRecentReactivation    RediscoveryKind = "recent_reactivation"
)

var RediscoveryKinds = []RediscoveryKind{
	RediscoveryKindStaleTentativeClaim,
	RediscoveryKindActiveInquiryOldEntry,
	RediscoveryKindUnsynthesizedInquiry,
	RediscoveryKindRecentReactivation,
}

type RediscoveryTargetType string

const (
	RediscoveryTargetTypeClaim      RediscoveryTargetType = "CLAIM"
	RediscoveryTargetTypeEngagement RediscoveryTargetType = "ENGAGEMENT"
	RediscoveryTargetTypeInquiry    RediscoveryTargetType = "INQUIRY"
)

type RediscoveryStatus string

const (
	RediscoveryStatusNew       RediscoveryStatus = "NEW"
	RediscoveryStatusSeen      RediscoveryStatus = "SEEN"
	RediscoveryStatusDismissed RediscoveryStatus = "DISMISSED"
	RediscoveryStatusActedOn   RediscoveryStatus = "ACTED_ON"
)

type RediscoveryItem struct {
	ID            string          `json:"id"`
	Kind          string          `json:"kind"`
	TargetType    string          `json:"target_type"`
	TargetID      string          `json:"target_id"`
	Reason        string          `json:"reason"`
	Status        string          `json:"status"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
	Claim         *Claim          `json:"claim,omitempty"`
	Engagement    *Engagement     `json:"engagement,omitempty"`
	Inquiry       *Inquiry        `json:"inquiry,omitempty"`
	LinkedInquiry *InquirySummary `json:"linked_inquiry,omitempty"`
}

func IsValidRediscoveryKind(value string) bool {
	for _, kind := range RediscoveryKinds {
		if string(kind) == value {
			return true
		}
	}
	return false
}
