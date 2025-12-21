package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Token 类型常量
const (
	TokenTypeBearer    = "bearer"    // 传统 Bearer Token 认证
	TokenTypeSignature = "signature" // HMAC-SHA256 签名认证
)

// UserToken 用户Token模型
type UserToken struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	UserID     uint64         `gorm:"not null;index" json:"user_id"`                       // 关联用户ID
	TokenName  string         `gorm:"size:100;not null" json:"token_name"`                 // Token名称
	TokenType  string         `gorm:"size:20;not null;default:'bearer'" json:"token_type"` // Token类型：bearer 或 signature
	Token      *string        `gorm:"size:190;uniqueIndex" json:"token"`                   // Bearer Token值，唯一（signature类型为null）
	AppID      *string        `gorm:"size:64;uniqueIndex" json:"app_id"`                   // 签名认证的 App ID（bearer类型为null）
	AppSecret  string         `gorm:"size:256" json:"-"`                                   // 加密存储的 App Secret（不在JSON中返回）
	LastUsedAt *time.Time     `json:"last_used_at"`                                        // 最后使用时间
	ExpireAt   *time.Time     `json:"expire_at"`                                           // 过期时间，null表示永不过期
	Status     int8           `gorm:"default:1" json:"status"`                             // 状态：1-正常，0-禁用
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
	token := hex.EncodeToString(bytes)
	ut.Token = &token
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

// GenerateAppID 生成唯一的 App ID
// 格式: app_ + 16字节随机hex字符串 (共36字符)
func (ut *UserToken) GenerateAppID() error {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	appID := fmt.Sprintf("app_%s", hex.EncodeToString(bytes))
	ut.AppID = &appID
	return nil
}

// GenerateAppSecret 生成安全的 App Secret
// 生成32字节随机hex字符串 (64字符)
func (ut *UserToken) GenerateAppSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	secret := hex.EncodeToString(bytes)
	return secret, nil
}

// IsSignatureType 检查是否为签名认证类型
func (ut *UserToken) IsSignatureType() bool {
	return ut.TokenType == TokenTypeSignature
}

// IsBearerType 检查是否为 Bearer Token 类型
func (ut *UserToken) IsBearerType() bool {
	return ut.TokenType == TokenTypeBearer
}
