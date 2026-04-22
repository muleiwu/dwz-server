package model

import "time"

// OIDCBinding 记录本地用户与远端 OIDC 身份(provider + sub)的绑定关系。
type OIDCBinding struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	UserID      uint64     `gorm:"not null;index;column:user_id" json:"user_id"`
	Provider    string     `gorm:"size:50;not null" json:"provider"`
	Sub         string     `gorm:"size:255;not null" json:"sub"`
	Email       string     `gorm:"size:255" json:"email"`
	LastLoginAt *time.Time `gorm:"column:last_login_at" json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (*OIDCBinding) TableName() string { return "oidc_bindings" }
