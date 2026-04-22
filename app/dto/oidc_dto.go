package dto

import "time"

// LoginOptionsResponse 登录页渲染所需的公开选项,免认证接口返回。
type LoginOptionsResponse struct {
	OIDCEnabled     bool   `json:"oidc_enabled"`
	OIDCDisplayName string `json:"oidc_display_name,omitempty"`
	OIDCProvider    string `json:"oidc_provider,omitempty"`
}

// OIDCConfigResponse 后台查询 OIDC 配置的返回。client_secret 永远不回显,
// 改为返回 secret_set 让前端判断是否已经填过。
type OIDCConfigResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Issuer      string    `json:"issuer"`
	ClientID    string    `json:"client_id"`
	SecretSet   bool      `json:"secret_set"`
	Scopes      string    `json:"scopes"`
	RedirectURI string    `json:"redirect_uri"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SaveOIDCConfigRequest 后台保存配置;client_secret 留空则保留现有值。
type SaveOIDCConfigRequest struct {
	Name         string `json:"name" binding:"required,max=50"`
	DisplayName  string `json:"display_name" binding:"max=100"`
	Issuer       string `json:"issuer" binding:"required,url,max=255"`
	ClientID     string `json:"client_id" binding:"required,max=255"`
	ClientSecret string `json:"client_secret" binding:"max=1024"`
	Scopes       string `json:"scopes" binding:"max=255"`
	RedirectURI  string `json:"redirect_uri" binding:"omitempty,url,max=255"`
	Enabled      bool   `json:"enabled"`
}

// TestOIDCConnectionRequest 用于「测试连接」按钮;字段语义与保存一致但不落库。
// client_secret 留空时用已有 provider 的密文解密后校验。
type TestOIDCConnectionRequest struct {
	Issuer       string `json:"issuer" binding:"required,url"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret"`
}

// OIDCBindingInfo 当前用户在某 provider 的绑定状态。
type OIDCBindingInfo struct {
	Provider    string     `json:"provider"`
	DisplayName string     `json:"display_name"`
	Sub         string     `json:"sub"`
	Email       string     `json:"email"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
