package model

import (
	"fmt"
	"time"

	"cnb.cool/mliev/open/dwz-server/pkg/base62"
	"gorm.io/gorm"
)

// ShortLink 短网址模型
type ShortLink struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	IssuerNumber *uint64        `gorm:"index" json:"issuer_number"`                       // 发号器分配的号码
	DomainID     uint64         `gorm:"not null;index" json:"domain_id"`                  // 关联域名表ID
	Protocol     string         `gorm:"size:10;default:'https';not null" json:"protocol"` // 协议头 http或https
	Domain       string         `gorm:"size:100;not null;index;" json:"domain"`           // 域名
	OriginalURL  string         `gorm:"size:2000;not null" json:"original_url"`           // 原始URL
	Title        string         `gorm:"size:255" json:"title"`                            // 网页标题
	IsCustomCode bool           `gorm:"default:false;" json:"is_custom_code"`             // 是否使用自定义短代码
	ShortCode    string         `gorm:"size:20;index" json:"short_code"`                  // 短代码(可自定义)
	ClickCount   int64          `gorm:"default:0" json:"click_count"`                     // 点击次数
	CreatorIP    string         `gorm:"size:45" json:"creator_ip"`                        // 创建者IP
	Description  string         `gorm:"size:500" json:"description"`                      // 描述
	ExpireAt     *time.Time     `json:"expire_at"`                                        // 过期时间，null表示永不过期
	IsActive     bool           `gorm:"default:true" json:"is_active"`                    // 是否激活
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *ShortLink) TableName() string {
	return "short_links"
}

// GetShortCode 获取短代码（新版本优先使用CustomCode字段）
func (s *ShortLink) GetShortCode() string {
	if s.ShortCode != "" {
		return s.ShortCode
	}
	// 兼容旧数据：如果没有CustomCode，使用ID转换（仅用于向后兼容）
	converter := base62.NewBase62()
	return converter.Encode(s.ID)
}

// IsExpired 检查是否过期
func (s *ShortLink) IsExpired() bool {
	if s.ExpireAt == nil {
		return false
	}
	return time.Now().After(*s.ExpireAt)
}

// GetFullURL 获取完整的短网址
func (s *ShortLink) GetFullURL() string {
	return fmt.Sprintf("%s://%s/%s", s.Protocol, s.Domain, s.GetShortCode())
}
