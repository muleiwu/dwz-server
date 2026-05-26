package dto

import "time"

type LinkSecurityIPRuleRequest struct {
	CIDR        string `json:"cidr" binding:"required"`
	Description string `json:"description"`
}

type LinkSecurityIPRuleResponse struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	ShortLinkID uint64    `json:"short_link_id"`
	CIDR        string    `json:"cidr"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LinkSecurityRequest struct {
	Password          *string                     `json:"password"`
	PasswordEnabled   *bool                       `json:"password_enabled"`
	AccessWindowStart *time.Time                  `json:"access_window_start"`
	AccessWindowEnd   *time.Time                  `json:"access_window_end"`
	MaxClicks         *int64                      `json:"max_clicks" binding:"omitempty,min=1"`
	IPPolicy          string                      `json:"ip_policy" binding:"omitempty,oneof=off allowlist blocklist"`
	IPRules           []LinkSecurityIPRuleRequest `json:"ip_rules"`
	BotPolicy         string                      `json:"bot_policy" binding:"omitempty,oneof=record_only allow block_known_bots"`
	ReportEnabled     *bool                       `json:"report_enabled"`
}

type LinkSecurityResponse struct {
	ID                uint64                       `json:"id"`
	WorkspaceID       uint64                       `json:"workspace_id"`
	ShortLinkID       uint64                       `json:"short_link_id"`
	PasswordEnabled   bool                         `json:"password_enabled"`
	AccessWindowStart *time.Time                   `json:"access_window_start"`
	AccessWindowEnd   *time.Time                   `json:"access_window_end"`
	MaxClicks         *int64                       `json:"max_clicks"`
	IPPolicy          string                       `json:"ip_policy"`
	IPRules           []LinkSecurityIPRuleResponse `json:"ip_rules"`
	BotPolicy         string                       `json:"bot_policy"`
	ReportEnabled     bool                         `json:"report_enabled"`
	URLBlocked        bool                         `json:"url_blocked"`
	URLBlockedReason  string                       `json:"url_blocked_reason"`
	SecurityEnabled   bool                         `json:"security_enabled"`
	SecuritySummary   string                       `json:"security_summary"`
	CreatedBy         *uint64                      `json:"created_by"`
	UpdatedBy         *uint64                      `json:"updated_by"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

type SecurityURLRuleRequest struct {
	RuleType string `json:"rule_type" binding:"required,oneof=domain keyword"`
	Action   string `json:"action" binding:"required,oneof=block allow"`
	Pattern  string `json:"pattern" binding:"required,max=500"`
	Enabled  *bool  `json:"enabled"`
}

type SecurityURLRuleListRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	RuleType string `form:"rule_type" binding:"omitempty,oneof=domain keyword"`
	Action   string `form:"action" binding:"omitempty,oneof=block allow"`
	Keyword  string `form:"keyword"`
	Enabled  *bool  `form:"enabled"`
}

type SecurityURLRuleResponse struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	RuleType    string    `json:"rule_type"`
	Action      string    `json:"action"`
	Pattern     string    `json:"pattern"`
	Enabled     bool      `json:"enabled"`
	CreatedBy   *uint64   `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SecurityURLRuleListResponse struct {
	List  []SecurityURLRuleResponse `json:"list"`
	Total int64                     `json:"total"`
	Page  int                       `json:"page"`
	Size  int                       `json:"size"`
}

type SecurityEventListRequest struct {
	Page        int        `form:"page" binding:"min=1"`
	PageSize    int        `form:"page_size" binding:"min=1,max=100"`
	ShortLinkID uint64     `form:"short_link_id"`
	EventType   string     `form:"event_type"`
	StartDate   *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate     *time.Time `form:"end_date" time_format:"2006-01-02"`
}

type SecurityEventResponse struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	ShortLinkID uint64    `json:"short_link_id"`
	EventType   string    `json:"event_type"`
	Reason      string    `json:"reason"`
	ClientIP    string    `json:"client_ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	CreatedAt   time.Time `json:"created_at"`
}

type SecurityEventListResponse struct {
	List  []SecurityEventResponse `json:"list"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

type PublicPasswordRequest struct {
	Domain    string `json:"domain" form:"domain" binding:"required"`
	ShortCode string `json:"short_code" form:"short_code" binding:"required"`
	Password  string `json:"password" form:"password" binding:"required"`
	Next      string `json:"next" form:"next"`
}

type AbuseReportCreateRequest struct {
	Domain        string `json:"domain" form:"domain"`
	ShortCode     string `json:"short_code" form:"short_code"`
	ShortLinkID   uint64 `json:"short_link_id" form:"short_link_id"`
	ReportType    string `json:"report_type" form:"report_type" binding:"required,oneof=malware phishing spam illegal other"`
	Description   string `json:"description" form:"description" binding:"max=1000"`
	ReporterEmail string `json:"reporter_email" form:"reporter_email" binding:"omitempty,email,max=255"`
}

type AbuseReportUpdateRequest struct {
	Status         string `json:"status" binding:"required,oneof=pending reviewing resolved rejected"`
	ResolutionNote string `json:"resolution_note" binding:"max=1000"`
	DisableLink    bool   `json:"disable_link"`
}

type AbuseReportListRequest struct {
	Page        int        `form:"page" binding:"min=1"`
	PageSize    int        `form:"page_size" binding:"min=1,max=100"`
	Status      string     `form:"status" binding:"omitempty,oneof=pending reviewing resolved rejected"`
	ReportType  string     `form:"report_type" binding:"omitempty,oneof=malware phishing spam illegal other"`
	ShortLinkID uint64     `form:"short_link_id"`
	StartDate   *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate     *time.Time `form:"end_date" time_format:"2006-01-02"`
}

type AbuseReportResponse struct {
	ID             uint64     `json:"id"`
	WorkspaceID    uint64     `json:"workspace_id"`
	ShortLinkID    uint64     `json:"short_link_id"`
	ReportType     string     `json:"report_type"`
	Description    string     `json:"description"`
	ReporterEmail  string     `json:"reporter_email"`
	ReporterIP     string     `json:"reporter_ip"`
	UserAgent      string     `json:"user_agent"`
	Status         string     `json:"status"`
	ResolutionNote string     `json:"resolution_note"`
	HandledBy      *uint64    `json:"handled_by"`
	HandledAt      *time.Time `json:"handled_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type AbuseReportListResponse struct {
	List  []AbuseReportResponse `json:"list"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}
