package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	CampaignStatusActive   = "active"
	CampaignStatusArchived = "archived"
)

type Campaign struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64         `gorm:"not null;index" json:"workspace_id"`
	Name        string         `gorm:"size:150;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	StartAt     *time.Time     `json:"start_at"`
	EndAt       *time.Time     `json:"end_at"`
	Status      string         `gorm:"size:20;default:'active';not null" json:"status"`
	CreatedBy   *uint64        `gorm:"index" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Campaign) TableName() string {
	return "campaigns"
}
