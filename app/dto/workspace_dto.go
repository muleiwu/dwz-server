package dto

import "time"

type WorkspaceResponse struct {
	ID          uint64    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerUserID *uint64   `json:"owner_user_id"`
	Status      int8      `json:"status"`
	Role        string    `json:"role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateWorkspaceRequest struct {
	Slug        string `json:"slug" binding:"required,min=2,max=100"`
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type WorkspaceListResponse struct {
	List []WorkspaceResponse `json:"list"`
}

type WorkspaceMemberResponse struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	UserID      uint64    `json:"user_id"`
	Username    string    `json:"username"`
	RealName    string    `json:"real_name"`
	Role        string    `json:"role"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AddWorkspaceMemberRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member viewer"`
}

type UpdateWorkspaceMemberRequest struct {
	Role   string `json:"role" binding:"required,oneof=admin member viewer"`
	Status *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

type WorkspaceMemberListResponse struct {
	List []WorkspaceMemberResponse `json:"list"`
}
