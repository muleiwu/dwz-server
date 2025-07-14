package model

import (
	"time"

	"gorm.io/gorm"
)

// Domain 域名配置模型
type Domain struct {
	ID              uint64         `gorm:"primaryKey" json:"id"`                             // 自增主键
	Protocol        string         `gorm:"size:10;default:'https';not null" json:"protocol"` // 协议头 http或https
	Domain          string         `gorm:"uniqueIndex;size:100;not null" json:"domain"`      // 域名  例如 n3.ink
	SiteName        string         `gorm:"size:100;default:''" json:"site_name"`             // 网站名称
	ICPNumber       string         `gorm:"size:50;default:''" json:"icp_number"`             // ICP备案号码
	PoliceNumber    string         `gorm:"size:50;default:''" json:"police_number"`          // 公安备案号码
	PassQueryParams bool           `gorm:"default:false" json:"pass_query_params"`           // 是否透传GET参数
	Description     string         `gorm:"type:text" json:"description"`                     // 描述
	IsActive        bool           `gorm:"default:true" json:"is_active"`                    // 是否激活
	CreatedAt       time.Time      `json:"created_at"`                                       // 创建时间
	UpdatedAt       time.Time      `json:"updated_at"`                                       // 更新时间
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`                                   // 删除时间
}

func (Domain) TableName() string {
	return "domains"
}
