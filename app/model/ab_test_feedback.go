package model

import "time"

// ABTestFeedback records business outcomes reported by landing pages or
// downstream systems after an A/B redirect.
type ABTestFeedback struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64    `gorm:"not null;default:1;index" json:"workspace_id"`
	ABTestID    uint64    `gorm:"not null;uniqueIndex:uk_ab_test_feedback_event;index" json:"ab_test_id"`
	VariantID   uint64    `gorm:"not null;index" json:"variant_id"`
	ShortLinkID uint64    `gorm:"not null;index" json:"short_link_id"`
	SessionID   string    `gorm:"size:128;not null;index" json:"session_id"`
	EventID     string    `gorm:"size:128;not null;uniqueIndex:uk_ab_test_feedback_event" json:"event_id"`
	Value       *float64  `gorm:"type:decimal(18,4)" json:"value"`
	Currency    string    `gorm:"size:16" json:"currency"`
	Metadata    string    `gorm:"type:text" json:"metadata"`
	IP          string    `gorm:"size:45" json:"ip"`
	UserAgent   string    `gorm:"size:1024" json:"user_agent"`
	Referer     string    `gorm:"size:2048" json:"referer"`
	OccurredAt  time.Time `gorm:"not null;index" json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`

	ABTest    ABTest        `gorm:"foreignKey:ABTestID" json:"ab_test,omitempty"`
	Variant   ABTestVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	ShortLink ShortLink     `gorm:"foreignKey:ShortLinkID" json:"short_link,omitempty"`
}

func (ABTestFeedback) TableName() string {
	return "ab_test_feedbacks"
}
