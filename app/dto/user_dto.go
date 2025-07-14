package dto

import "time"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint64     `json:"id"`
	Username  string     `json:"username"`
	RealName  string     `json:"real_name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Status    int8       `json:"status"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"newuser"`
	Password string `json:"password" binding:"required,min=6,max=50" example:"password123"`
	RealName string `json:"real_name" binding:"max=100" example:"张三"`
	Email    string `json:"email" binding:"omitempty,email,max=255" example:"user@example.com"`
	Phone    string `json:"phone" binding:"omitempty,max=20" example:"13800138000"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	RealName string `json:"real_name" binding:"omitempty,max=100" example:"李四"`
	Email    string `json:"email" binding:"omitempty,email,max=255" example:"user@example.com"`
	Phone    string `json:"phone" binding:"omitempty,max=20" example:"13800138000"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1" example:"1"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50" example:"newpassword"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=50" example:"newpassword"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1" example:"1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100" example:"10"`
	Username string `form:"username" example:"admin"`
	RealName string `form:"real_name" example:"管理员"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1" example:"1"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	List       []UserInfo `json:"list"`
	Pagination Pagination `json:"pagination"`
}

// CreateUserTokenRequest 创建用户Token请求
type CreateUserTokenRequest struct {
	TokenName string     `json:"token_name" binding:"required,max=100" example:"API Token"`
	ExpireAt  *time.Time `json:"expire_at" example:"2024-12-31T23:59:59Z"`
}

// CreateUserTokenResponse 创建用户Token响应
type CreateUserTokenResponse struct {
	ID        uint64     `json:"id"`
	TokenName string     `json:"token_name"`
	Token     string     `json:"token"`
	ExpireAt  *time.Time `json:"expire_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// UserTokenInfo Token信息
type UserTokenInfo struct {
	ID         uint64     `json:"id"`
	TokenName  string     `json:"token_name"`
	Token      string     `json:"token"` // 列表中只显示前8位
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpireAt   *time.Time `json:"expire_at"`
	Status     int8       `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

// UserTokenListRequest Token列表请求
type UserTokenListRequest struct {
	Page      int    `form:"page,default=1" binding:"min=1" example:"1"`
	PageSize  int    `form:"page_size,default=10" binding:"min=1,max=100" example:"10"`
	TokenName string `form:"token_name" example:"API"`
	Status    *int8  `form:"status" binding:"omitempty,oneof=0 1" example:"1"`
}

// UserTokenListResponse Token列表响应
type UserTokenListResponse struct {
	List       []UserTokenInfo `json:"list"`
	Pagination Pagination      `json:"pagination"`
}
