package model

import (
	"time"

	"gorm.io/gorm"
)

// OIDCProvider 保存单个 OpenID Connect 提供商的配置。
// client_secret 字段以对称加密后的密文形式入库,运行时通过
// helper.DecryptOIDCSecret 解密。
type OIDCProvider struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:50;not null;uniqueIndex" json:"name"`
	DisplayName  string         `gorm:"size:100" json:"display_name"`
	Issuer       string         `gorm:"size:255;not null" json:"issuer"`
	ClientID     string         `gorm:"size:255;not null" json:"client_id"`
	ClientSecret string         `gorm:"size:1024;not null;column:client_secret" json:"-"`
	Scopes       string         `gorm:"size:255" json:"scopes"`
	RedirectURI  string         `gorm:"size:255;column:redirect_uri" json:"redirect_uri"`
	Enabled      int8           `gorm:"default:0" json:"enabled"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (*OIDCProvider) TableName() string { return "oidc_providers" }

// IsEnabled 判断是否启用。
func (p *OIDCProvider) IsEnabled() bool { return p != nil && p.Enabled == 1 }
