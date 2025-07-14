package model

import "time"

// ClickStatistic 点击统计模型
type ClickStatistic struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	ShortLinkID uint64    `gorm:"index:idx_short_link_date;not null" json:"short_link_id"`
	IP          string    `gorm:"size:45" json:"ip"`
	UserAgent   string    `gorm:"size:1024" json:"user_agent"`
	Referer     string    `gorm:"size:2048" json:"referer"`
	QueryParams string    `gorm:"size:2048" json:"query_params"`
	Country     string    `gorm:"size:100" json:"country"`
	City        string    `gorm:"size:100" json:"city"`
	ClickDate   time.Time `gorm:"index:idx_short_link_date" json:"click_date"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ClickStatistic) TableName() string {
	return "click_statistics"
}
