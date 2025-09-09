package model

import (
	"time"

	"gorm.io/gorm"
)

// ABTestVariant AB测试版本模型
type ABTestVariant struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	ABTestID    uint64         `gorm:"not null;index" json:"ab_test_id"`     // 关联的AB测试ID
	Name        string         `gorm:"size:100;not null" json:"name"`        // 版本名称 (如: A, B, Control)
	TargetURL   string         `gorm:"size:2000;not null" json:"target_url"` // 目标URL
	Weight      int            `gorm:"default:50" json:"weight"`             // 权重 (百分比)
	IsControl   bool           `gorm:"default:false" json:"is_control"`      // 是否为对照组
	Description string         `gorm:"size:500" json:"description"`          // 版本描述
	IsActive    bool           `gorm:"default:true" json:"is_active"`        // 是否激活
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	ABTest ABTest `gorm:"foreignKey:ABTestID" json:"ab_test,omitempty"`
}

func (ABTestVariant) TableName() string {
	return "ab_test_variants"
}
