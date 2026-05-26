package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	LinkIPPolicyOff       = "off"
	LinkIPPolicyAllowlist = "allowlist"
	LinkIPPolicyBlocklist = "blocklist"

	LinkBotPolicyRecordOnly     = "record_only"
	LinkBotPolicyAllow          = "allow"
	LinkBotPolicyBlockKnownBots = "block_known_bots"

	SecurityRuleTypeDomain  = "domain"
	SecurityRuleTypeKeyword = "keyword"

	SecurityRuleActionBlock = "block"
	SecurityRuleActionAllow = "allow"

	AbuseReportTypeMalware  = "malware"
	AbuseReportTypePhishing = "phishing"
	AbuseReportTypeSpam     = "spam"
	AbuseReportTypeIllegal  = "illegal"
	AbuseReportTypeOther    = "other"

	AbuseReportStatusPending   = "pending"
	AbuseReportStatusReviewing = "reviewing"
	AbuseReportStatusResolved  = "resolved"
	AbuseReportStatusRejected  = "rejected"

	SecurityEventPasswordRequired = "password_required"
	SecurityEventPasswordFailed   = "password_failed"
	SecurityEventAccessDenied     = "access_denied"
	SecurityEventBotBlocked       = "bot_blocked"
	SecurityEventURLBlocked       = "url_blocked"
	SecurityEventAbuseReported    = "abuse_reported"
)

type LinkSecuritySetting struct {
	ID                uint64         `gorm:"primaryKey" json:"id"`
	WorkspaceID       uint64         `gorm:"not null;index" json:"workspace_id"`
	ShortLinkID       uint64         `gorm:"not null;uniqueIndex" json:"short_link_id"`
	PasswordEnabled   bool           `gorm:"not null;default:false" json:"password_enabled"`
	PasswordHash      string         `gorm:"size:255" json:"-"`
	AccessWindowStart *time.Time     `json:"access_window_start"`
	AccessWindowEnd   *time.Time     `json:"access_window_end"`
	MaxClicks         *int64         `json:"max_clicks"`
	IPPolicy          string         `gorm:"size:20;not null;default:'off'" json:"ip_policy"`
	BotPolicy         string         `gorm:"size:30;not null;default:'record_only'" json:"bot_policy"`
	ReportEnabled     bool           `gorm:"not null;default:false" json:"report_enabled"`
	URLBlocked        bool           `gorm:"not null;default:false" json:"url_blocked"`
	URLBlockedReason  string         `gorm:"size:500" json:"url_blocked_reason"`
	CreatedBy         *uint64        `gorm:"index" json:"created_by"`
	UpdatedBy         *uint64        `json:"updated_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LinkSecuritySetting) TableName() string {
	return "link_security_settings"
}

type LinkSecurityIPRule struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64         `gorm:"not null;index" json:"workspace_id"`
	ShortLinkID uint64         `gorm:"not null;index" json:"short_link_id"`
	CIDR        string         `gorm:"size:64;not null" json:"cidr"`
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LinkSecurityIPRule) TableName() string {
	return "link_security_ip_rules"
}

type SecurityURLRule struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64         `gorm:"not null;index" json:"workspace_id"`
	RuleType    string         `gorm:"size:20;not null;index" json:"rule_type"`
	Action      string         `gorm:"size:20;not null;index" json:"action"`
	Pattern     string         `gorm:"size:500;not null" json:"pattern"`
	Enabled     bool           `gorm:"not null;default:true;index" json:"enabled"`
	CreatedBy   *uint64        `gorm:"index" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SecurityURLRule) TableName() string {
	return "security_url_rules"
}

type AbuseReport struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	WorkspaceID    uint64     `gorm:"not null;index" json:"workspace_id"`
	ShortLinkID    uint64     `gorm:"not null;index" json:"short_link_id"`
	ReportType     string     `gorm:"size:30;not null;index" json:"report_type"`
	Description    string     `gorm:"size:1000" json:"description"`
	ReporterEmail  string     `gorm:"size:255" json:"reporter_email"`
	ReporterIP     string     `gorm:"size:45;index" json:"reporter_ip"`
	UserAgent      string     `gorm:"size:1024" json:"user_agent"`
	Status         string     `gorm:"size:30;not null;default:'pending';index" json:"status"`
	ResolutionNote string     `gorm:"size:1000" json:"resolution_note"`
	HandledBy      *uint64    `gorm:"index" json:"handled_by"`
	HandledAt      *time.Time `json:"handled_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (AbuseReport) TableName() string {
	return "abuse_reports"
}

type LinkSecurityEvent struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64    `gorm:"not null;index" json:"workspace_id"`
	ShortLinkID uint64    `gorm:"not null;index" json:"short_link_id"`
	EventType   string    `gorm:"size:50;not null;index" json:"event_type"`
	Reason      string    `gorm:"size:500" json:"reason"`
	ClientIP    string    `gorm:"size:45;index" json:"client_ip"`
	UserAgent   string    `gorm:"size:1024" json:"user_agent"`
	Referer     string    `gorm:"size:2048" json:"referer"`
	CreatedAt   time.Time `json:"created_at"`
}

func (LinkSecurityEvent) TableName() string {
	return "link_security_events"
}
