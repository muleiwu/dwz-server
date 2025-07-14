package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// UserToken 用户Token模型
type UserToken struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	UserID     uint64         `gorm:"not null;index" json:"user_id"`              // 关联用户ID
	TokenName  string         `gorm:"size:100;not null" json:"token_name"`        // Token名称
	Token      string         `gorm:"size:255;not null;uniqueIndex" json:"token"` // Token值，唯一
	LastUsedAt *time.Time     `json:"last_used_at"`                               // 最后使用时间
	ExpireAt   *time.Time     `json:"expire_at"`                                  // 过期时间，null表示永不过期
	Status     int8           `gorm:"default:1" json:"status"`                    // 状态：1-正常，0-禁用
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联查询
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ut *UserToken) TableName() string {
	return "user_tokens"
}

// GenerateToken 生成随机Token
func (ut *UserToken) GenerateToken() error {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	ut.Token = hex.EncodeToString(bytes)
	return nil
}

// IsExpired 检查是否过期
func (ut *UserToken) IsExpired() bool {
	if ut.ExpireAt == nil {
		return false
	}
	return time.Now().After(*ut.ExpireAt)
}

// IsActive 是否激活状态
func (ut *UserToken) IsActive() bool {
	return ut.Status == 1 && !ut.IsExpired()
}

// UpdateLastUsed 更新最后使用时间
func (ut *UserToken) UpdateLastUsed() {
	now := time.Now()
	ut.LastUsedAt = &now
}
