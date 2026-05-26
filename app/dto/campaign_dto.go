package dto

import "time"

type CampaignRequest struct {
	Name        string     `json:"name" binding:"required,max=150"`
	Description string     `json:"description" binding:"max=500"`
	StartAt     *time.Time `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	Status      string     `json:"status" binding:"omitempty,oneof=active archived"`
}

type CampaignResponse struct {
	ID          uint64     `json:"id"`
	WorkspaceID uint64     `json:"workspace_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	StartAt     *time.Time `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	Status      string     `json:"status"`
	CreatedBy   *uint64    `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CampaignListRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status" binding:"omitempty,oneof=active archived"`
}

type CampaignListResponse struct {
	List  []CampaignResponse `json:"list"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

type CampaignReportRequest struct {
	CampaignID uint64     `form:"campaign_id"`
	StartDate  *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate    *time.Time `form:"end_date" time_format:"2006-01-02"`
}

type CampaignReportItem struct {
	CampaignID     uint64 `json:"campaign_id"`
	CampaignName   string `json:"campaign_name"`
	ShortLinkCount int64  `json:"short_link_count"`
	ClickCount     int64  `json:"click_count"`
	UniqueIPs      int64  `json:"unique_ips"`
}

type CampaignReportResponse struct {
	List []CampaignReportItem `json:"list"`
}
