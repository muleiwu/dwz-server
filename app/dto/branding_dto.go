package dto

import "time"

type BrandingResponse struct {
	LogoURL          string `json:"logo_url"`
	BrandName        string `json:"brand_name"`
	CopyrightEnabled bool   `json:"copyright_enabled"`
	CopyrightText    string `json:"copyright_text"`
	CopyrightLink    string `json:"copyright_link"`
	Source           string `json:"source"`
}

type SystemBrandingRequest struct {
	LogoURL   string `json:"logo_url" binding:"omitempty,max=500"`
	BrandName string `json:"brand_name" binding:"omitempty,max=80"`
}

type SystemBrandingResponse struct {
	ID               uint8     `json:"id"`
	LogoURL          string    `json:"logo_url"`
	BrandName        string    `json:"brand_name"`
	CopyrightEnabled bool      `json:"copyright_enabled"`
	CopyrightText    string    `json:"copyright_text"`
	CopyrightLink    string    `json:"copyright_link"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}

type LogoUploadResponse struct {
	URL string `json:"url"`
}
