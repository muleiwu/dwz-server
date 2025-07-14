package model

import "time"

// ABTestClickStatistic AB测试点击统计模型
type ABTestClickStatistic struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	ABTestID    uint64    `gorm:"not null;index:idx_ab_test_click" json:"ab_test_id"`          // AB测试ID
	VariantID   uint64    `gorm:"not null;index:idx_variant_click" json:"variant_id"`          // 版本ID
	ShortLinkID uint64    `gorm:"not null;index" json:"short_link_id"`                         // 短链接ID
	IP          string    `gorm:"size:45" json:"ip"`                                           // 访客IP
	UserAgent   string    `gorm:"size:1024" json:"user_agent"`                                 // 用户代理
	Referer     string    `gorm:"size:2048" json:"referer"`                                    // 来源
	QueryParams string    `gorm:"size:2048" json:"query_params"`                               // 查询参数
	Country     string    `gorm:"size:100" json:"country"`                                     // 国家
	City        string    `gorm:"size:100" json:"city"`                                        // 城市
	SessionID   string    `gorm:"size:128;index" json:"session_id"`                            // 会话ID，用于去重
	ClickDate   time.Time `gorm:"index:idx_ab_test_click,idx_variant_click" json:"click_date"` // 点击日期
	CreatedAt   time.Time `json:"created_at"`

	// 关联关系
	ABTest    ABTest        `gorm:"foreignKey:ABTestID" json:"ab_test,omitempty"`
	Variant   ABTestVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	ShortLink ShortLink     `gorm:"foreignKey:ShortLinkID" json:"short_link,omitempty"`
}

func (ABTestClickStatistic) TableName() string {
	return "ab_test_click_statistics"
}
