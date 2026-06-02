package model

import "time"

type SystemBranding struct {
	ID               uint8     `gorm:"primaryKey" json:"id"`
	BrandName        string    `gorm:"size:80;not null;default:''" json:"brand_name"`
	LogoURL          string    `gorm:"size:500;not null;default:''" json:"logo_url"`
	CopyrightEnabled bool      `gorm:"not null" json:"copyright_enabled"`
	CopyrightText    string    `gorm:"size:200;not null;default:''" json:"copyright_text"`
	CopyrightLink    string    `gorm:"size:500;not null;default:''" json:"copyright_link"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (SystemBranding) TableName() string {
	return "ee_system_brandings"
}

type WorkspaceBranding struct {
	WorkspaceID      uint64    `gorm:"primaryKey" json:"workspace_id"`
	BrandName        string    `gorm:"size:80;not null;default:''" json:"brand_name"`
	LogoURL          string    `gorm:"size:500;not null;default:''" json:"logo_url"`
	CopyrightEnabled bool      `gorm:"not null" json:"copyright_enabled"`
	CopyrightText    string    `gorm:"size:200;not null;default:''" json:"copyright_text"`
	CopyrightLink    string    `gorm:"size:500;not null;default:''" json:"copyright_link"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (WorkspaceBranding) TableName() string {
	return "ee_workspace_brandings"
}

type DomainBranding struct {
	DomainID              uint64    `gorm:"primaryKey" json:"domain_id"`
	WorkspaceID           uint64    `gorm:"not null;index" json:"workspace_id"`
	OverrideBrandName     bool      `gorm:"not null" json:"override_brand_name"`
	BrandName             string    `gorm:"size:80;not null;default:''" json:"brand_name"`
	OverrideLogo          bool      `gorm:"not null" json:"override_logo"`
	LogoURL               string    `gorm:"size:500;not null;default:''" json:"logo_url"`
	OverrideCopyright     bool      `gorm:"not null" json:"override_copyright"`
	CopyrightEnabled      bool      `gorm:"not null" json:"copyright_enabled"`
	OverrideCopyrightText bool      `gorm:"not null" json:"override_copyright_text"`
	CopyrightText         string    `gorm:"size:200;not null;default:''" json:"copyright_text"`
	CopyrightLink         string    `gorm:"size:500;not null;default:''" json:"copyright_link"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (DomainBranding) TableName() string {
	return "ee_domain_brandings"
}
