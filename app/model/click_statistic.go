package model

import "time"

// ClickStatistic 点击统计模型
type ClickStatistic struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64    `gorm:"not null;default:1;index" json:"workspace_id"`
	CampaignID  *uint64   `gorm:"index" json:"campaign_id"`
	ShortLinkID uint64    `gorm:"index:idx_short_link_date;not null" json:"short_link_id"`
	IP          string    `gorm:"size:45" json:"ip"`
	UserAgent   string    `gorm:"size:1024" json:"user_agent"`
	Referer     string    `gorm:"size:2048" json:"referer"`
	QueryParams string    `gorm:"size:2048" json:"query_params"`
	UTMSource   string    `gorm:"size:255" json:"utm_source"`
	UTMMedium   string    `gorm:"size:255" json:"utm_medium"`
	UTMCampaign string    `gorm:"size:255" json:"utm_campaign"`
	UTMTerm     string    `gorm:"size:255" json:"utm_term"`
	UTMContent  string    `gorm:"size:255" json:"utm_content"`
	DeviceType  string    `gorm:"size:50" json:"device_type"`
	Browser     string    `gorm:"size:100" json:"browser"`
	OS          string    `gorm:"size:100" json:"os"`
	IsBot       bool      `gorm:"default:false;index" json:"is_bot"`
	BotName     string    `gorm:"size:100" json:"bot_name"`
	Country     string    `gorm:"size:100" json:"country"`
	Province    string    `gorm:"size:100" json:"province"`
	City        string    `gorm:"size:100" json:"city"`
	ISP         string    `gorm:"size:100" json:"isp"`
	ClickDate   time.Time `gorm:"index:idx_short_link_date" json:"click_date"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ClickStatistic) TableName() string {
	return "click_statistics"
}
