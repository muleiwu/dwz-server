package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RouteConditionCountry    = "country"
	RouteConditionProvince   = "province"
	RouteConditionCity       = "city"
	RouteConditionDeviceType = "device_type"
	RouteConditionBrowser    = "browser"
	RouteConditionOS         = "os"
	RouteConditionLanguage   = "language"
	RouteConditionReferer    = "referer"
	RouteConditionQueryParam = "query_param"

	RouteOperatorExists   = "exists"
	RouteOperatorEq       = "eq"
	RouteOperatorIn       = "in"
	RouteOperatorContains = "contains"
	RouteOperatorPrefix   = "prefix"
	RouteOperatorSuffix   = "suffix"
)

type LinkRoute struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	WorkspaceID uint64         `gorm:"not null;index" json:"workspace_id"`
	ShortLinkID uint64         `gorm:"not null;index" json:"short_link_id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Priority    int            `gorm:"not null;default:100;index" json:"priority"`
	TargetURL   string         `gorm:"size:2000;not null" json:"target_url"`
	IsActive    bool           `gorm:"not null;default:true;index" json:"is_active"`
	CreatedBy   *uint64        `gorm:"index" json:"created_by"`
	UpdatedBy   *uint64        `json:"updated_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	ConditionGroups []LinkRouteConditionGroup `gorm:"foreignKey:RouteID" json:"condition_groups,omitempty"`
}

func (LinkRoute) TableName() string {
	return "link_routes"
}

type LinkRouteConditionGroup struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	RouteID   uint64    `gorm:"not null;index" json:"route_id"`
	Position  int       `gorm:"not null;default:0" json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Conditions []LinkRouteCondition `gorm:"foreignKey:GroupID" json:"conditions,omitempty"`
}

func (LinkRouteConditionGroup) TableName() string {
	return "link_route_condition_groups"
}

type LinkRouteCondition struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	GroupID        uint64    `gorm:"not null;index" json:"group_id"`
	ConditionType  string    `gorm:"size:30;not null;index" json:"condition_type"`
	Operator       string    `gorm:"size:20;not null" json:"operator"`
	ConditionKey   string    `gorm:"size:255" json:"condition_key"`
	ConditionValue string    `gorm:"size:1000" json:"condition_value"`
	Position       int       `gorm:"not null;default:0" json:"position"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (LinkRouteCondition) TableName() string {
	return "link_route_conditions"
}
