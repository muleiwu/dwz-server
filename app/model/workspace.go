package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	WorkspaceRoleOwner  = "owner"
	WorkspaceRoleAdmin  = "admin"
	WorkspaceRoleMember = "member"
	WorkspaceRoleViewer = "viewer"
)

type Workspace struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Slug        string         `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	OwnerUserID *uint64        `gorm:"index" json:"owner_user_id"`
	Status      int8           `gorm:"default:1" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Members []WorkspaceMember `gorm:"foreignKey:WorkspaceID" json:"members,omitempty"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

func (w *Workspace) IsActive() bool {
	return w.Status == 1
}

type WorkspaceMember struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64         `gorm:"not null;uniqueIndex:uk_workspace_members_workspace_user" json:"workspace_id"`
	UserID      uint64         `gorm:"not null;uniqueIndex:uk_workspace_members_workspace_user;index" json:"user_id"`
	Role        string         `gorm:"size:20;not null" json:"role"`
	Status      int8           `gorm:"default:1" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

func (m *WorkspaceMember) IsActive() bool {
	return m.Status == 1
}
