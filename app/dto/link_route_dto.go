package dto

import "time"

type LinkRouteConditionRequest struct {
	ConditionType  string `json:"condition_type" binding:"required"`
	Operator       string `json:"operator" binding:"required"`
	ConditionKey   string `json:"condition_key"`
	ConditionValue string `json:"condition_value"`
}

type LinkRouteConditionGroupRequest struct {
	Conditions []LinkRouteConditionRequest `json:"conditions" binding:"required,min=1"`
}

type LinkRouteRequest struct {
	Name            string                           `json:"name" binding:"required"`
	Description     string                           `json:"description"`
	Priority        int                              `json:"priority"`
	TargetURL       string                           `json:"target_url" binding:"required,url"`
	IsActive        *bool                            `json:"is_active"`
	ConditionGroups []LinkRouteConditionGroupRequest `json:"condition_groups" binding:"required,min=1"`
}

type LinkRouteReorderItem struct {
	ID       uint64 `json:"id" binding:"required"`
	Priority int    `json:"priority"`
}

type LinkRouteReorderRequest struct {
	Routes []LinkRouteReorderItem `json:"routes" binding:"required,min=1"`
}

type LinkRouteTestRequest struct {
	ClientIP       string `json:"client_ip"`
	UserAgent      string `json:"user_agent"`
	AcceptLanguage string `json:"accept_language"`
	Referer        string `json:"referer"`
	Query          string `json:"query"`
}

type LinkRouteConditionResponse struct {
	ID             uint64 `json:"id"`
	ConditionType  string `json:"condition_type"`
	Operator       string `json:"operator"`
	ConditionKey   string `json:"condition_key"`
	ConditionValue string `json:"condition_value"`
	Position       int    `json:"position"`
}

type LinkRouteConditionGroupResponse struct {
	ID         uint64                       `json:"id"`
	Position   int                          `json:"position"`
	Conditions []LinkRouteConditionResponse `json:"conditions"`
}

type LinkRouteResponse struct {
	ID              uint64                            `json:"id"`
	WorkspaceID     uint64                            `json:"workspace_id"`
	ShortLinkID     uint64                            `json:"short_link_id"`
	Name            string                            `json:"name"`
	Description     string                            `json:"description"`
	Priority        int                               `json:"priority"`
	TargetURL       string                            `json:"target_url"`
	IsActive        bool                              `json:"is_active"`
	ConditionGroups []LinkRouteConditionGroupResponse `json:"condition_groups"`
	CreatedBy       *uint64                           `json:"created_by"`
	UpdatedBy       *uint64                           `json:"updated_by"`
	CreatedAt       time.Time                         `json:"created_at"`
	UpdatedAt       time.Time                         `json:"updated_at"`
}

type LinkRouteListResponse struct {
	List []LinkRouteResponse `json:"list"`
}

type LinkRouteTestResponse struct {
	Matched      bool   `json:"matched"`
	RouteID      uint64 `json:"route_id,omitempty"`
	RouteName    string `json:"route_name,omitempty"`
	TargetURL    string `json:"target_url"`
	FallbackUsed bool   `json:"fallback_used"`
	Reason       string `json:"reason"`
}
