package model

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64         `gorm:"not null;uniqueIndex:uk_tags_workspace_name" json:"workspace_id"`
	Name        string         `gorm:"size:100;not null;uniqueIndex:uk_tags_workspace_name" json:"name"`
	Color       string         `gorm:"size:20" json:"color"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Tag) TableName() string {
	return "tags"
}

type ShortLinkTag struct {
	ShortLinkID uint64    `gorm:"primaryKey" json:"short_link_id"`
	TagID       uint64    `gorm:"primaryKey" json:"tag_id"`
	CreatedAt   time.Time `json:"created_at"`

	Tag       Tag       `gorm:"foreignKey:TagID" json:"tag,omitempty"`
	ShortLink ShortLink `gorm:"foreignKey:ShortLinkID" json:"short_link,omitempty"`
}

func (ShortLinkTag) TableName() string {
	return "short_link_tags"
}
