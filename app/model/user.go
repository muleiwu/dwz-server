package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"size:50;not null;uniqueIndex" json:"username"` // 用户名，唯一
	Password  string         `gorm:"size:255;not null" json:"-"`                   // 密码，不返回给前端
	RealName  string         `gorm:"size:100" json:"real_name"`                    // 真实姓名
	Email     string         `gorm:"size:255;uniqueIndex" json:"email"`            // 邮箱，唯一
	Phone     string         `gorm:"size:20" json:"phone"`                         // 手机号
	Status    int8           `gorm:"default:1" json:"status"`                      // 状态：1-正常，0-禁用
	LastLogin *time.Time     `json:"last_login"`                                   // 最后登录时间
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) TableName() string {
	return "users"
}

// SetPassword 设置密码（加密存储）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsActive 是否激活状态
func (u *User) IsActive() bool {
	return u.Status == 1
}
