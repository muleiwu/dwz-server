package model

import (
	"time"

	"gorm.io/gorm"
)

// ABTest AB测试实验模型
type ABTest struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	ShortLinkID  uint64         `gorm:"not null;index" json:"short_link_id"`          // 关联的短链接ID
	Name         string         `gorm:"size:255;not null" json:"name"`                // 实验名称
	Description  string         `gorm:"size:500" json:"description"`                  // 实验描述
	Status       string         `gorm:"size:20;default:'draft'" json:"status"`        // 实验状态: draft, running, paused, completed
	TrafficSplit string         `gorm:"size:20;default:'equal'" json:"traffic_split"` // 流量分配策略: equal, weighted, custom
	StartTime    *time.Time     `json:"start_time"`                                   // 开始时间
	EndTime      *time.Time     `json:"end_time"`                                     // 结束时间
	IsActive     bool           `gorm:"default:true" json:"is_active"`                // 是否激活
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	ShortLink ShortLink       `gorm:"foreignKey:ShortLinkID" json:"short_link,omitempty"`
	Variants  []ABTestVariant `gorm:"foreignKey:ABTestID" json:"variants,omitempty"`
}

func (ABTest) TableName() string {
	return "ab_tests"
}

// IsRunning 检查AB测试是否正在运行
func (a *ABTest) IsRunning() bool {
	if !a.IsActive || a.Status != "running" {
		return false
	}

	now := time.Now()

	// 检查开始时间
	if a.StartTime != nil && now.Before(*a.StartTime) {
		return false
	}

	// 检查结束时间
	if a.EndTime != nil && now.After(*a.EndTime) {
		return false
	}

	return true
}

// GetActiveVariants 获取激活的变体列表
func (a *ABTest) GetActiveVariants() []ABTestVariant {
	var activeVariants []ABTestVariant
	for _, variant := range a.Variants {
		if variant.IsActive {
			activeVariants = append(activeVariants, variant)
		}
	}
	return activeVariants
}
