package service

import (
	"errors"
	"net/url"
	"sort"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

type RouteResolveInput struct {
	ClientIP       string
	UserAgent      string
	AcceptLanguage string
	Referer        string
	QueryString    string
}

type RouteResolveResult struct {
	RoutingEnabled bool
	Matched        bool
	FallbackUsed   bool
	Route          *model.LinkRoute
	TargetURL      string
	Reason         string
}

type LinkRouteService struct {
	helper          interfaces.HelperInterface
	securityService *LinkSecurityService
}

func NewLinkRouteService(helper interfaces.HelperInterface) *LinkRouteService {
	return &LinkRouteService{
		helper:          helper,
		securityService: NewLinkSecurityService(helper),
	}
}

func (s *LinkRouteService) ListRoutes(shortLinkID, workspaceID uint64) (*dto.LinkRouteListResponse, error) {
	if _, err := s.ensureShortLink(shortLinkID, workspaceID); err != nil {
		return nil, err
	}
	routes, err := s.loadRoutes(shortLinkID, workspaceID, false)
	if err != nil {
		return nil, err
	}
	list := make([]dto.LinkRouteResponse, 0, len(routes))
	for _, route := range routes {
		list = append(list, linkRouteToResponse(&route))
	}
	return &dto.LinkRouteListResponse{List: list}, nil
}

func (s *LinkRouteService) CreateRoute(shortLinkID, workspaceID, userID uint64, req *dto.LinkRouteRequest) (*dto.LinkRouteResponse, error) {
	if _, err := s.ensureShortLink(shortLinkID, workspaceID); err != nil {
		return nil, err
	}
	if err := s.validateRouteRequest(workspaceID, req); err != nil {
		return nil, err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	route := &model.LinkRoute{
		WorkspaceID: workspaceID,
		ShortLinkID: shortLinkID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Priority:    req.Priority,
		TargetURL:   strings.TrimSpace(req.TargetURL),
		IsActive:    isActive,
		CreatedBy:   actorPtr(userID),
		UpdatedBy:   actorPtr(userID),
	}
	if route.Priority == 0 {
		route.Priority = 100
	}
	if err := s.helper.GetDatabase().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(route).Error; err != nil {
			return err
		}
		return s.replaceConditionGroups(tx, route.ID, req.ConditionGroups)
	}); err != nil {
		return nil, err
	}
	created, err := s.findRoute(route.ID, shortLinkID, workspaceID)
	if err != nil {
		return nil, err
	}
	resp := linkRouteToResponse(created)
	return &resp, nil
}

func (s *LinkRouteService) UpdateRoute(routeID, shortLinkID, workspaceID, userID uint64, req *dto.LinkRouteRequest) (*dto.LinkRouteResponse, error) {
	if err := s.validateRouteRequest(workspaceID, req); err != nil {
		return nil, err
	}
	route, err := s.findRoute(routeID, shortLinkID, workspaceID)
	if err != nil {
		return nil, err
	}
	route.Name = strings.TrimSpace(req.Name)
	route.Description = strings.TrimSpace(req.Description)
	route.Priority = req.Priority
	if route.Priority == 0 {
		route.Priority = 100
	}
	route.TargetURL = strings.TrimSpace(req.TargetURL)
	if req.IsActive != nil {
		route.IsActive = *req.IsActive
	}
	if userID > 0 {
		route.UpdatedBy = &userID
	}
	if err := s.helper.GetDatabase().Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(route).Error; err != nil {
			return err
		}
		return s.replaceConditionGroups(tx, route.ID, req.ConditionGroups)
	}); err != nil {
		return nil, err
	}
	updated, err := s.findRoute(routeID, shortLinkID, workspaceID)
	if err != nil {
		return nil, err
	}
	resp := linkRouteToResponse(updated)
	return &resp, nil
}

func (s *LinkRouteService) DeleteRoute(routeID, shortLinkID, workspaceID uint64) error {
	route, err := s.findRoute(routeID, shortLinkID, workspaceID)
	if err != nil {
		return err
	}
	return s.helper.GetDatabase().Transaction(func(tx *gorm.DB) error {
		var groups []model.LinkRouteConditionGroup
		if err := tx.Where("route_id = ?", route.ID).Find(&groups).Error; err != nil {
			return err
		}
		for _, group := range groups {
			if err := tx.Where("group_id = ?", group.ID).Delete(&model.LinkRouteCondition{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("route_id = ?", route.ID).Delete(&model.LinkRouteConditionGroup{}).Error; err != nil {
			return err
		}
		return tx.Delete(route).Error
	})
}

func (s *LinkRouteService) ReorderRoutes(shortLinkID, workspaceID, userID uint64, req *dto.LinkRouteReorderRequest) error {
	if _, err := s.ensureShortLink(shortLinkID, workspaceID); err != nil {
		return err
	}
	return s.helper.GetDatabase().Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Routes {
			updates := map[string]any{"priority": item.Priority}
			if userID > 0 {
				updates["updated_by"] = userID
			}
			result := tx.Model(&model.LinkRoute{}).
				Where("id = ? AND short_link_id = ? AND workspace_id = ? AND deleted_at IS NULL", item.ID, shortLinkID, workspaceID).
				Updates(updates)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("路由规则不存在")
			}
		}
		return nil
	})
}

func (s *LinkRouteService) TestRoute(shortLinkID, workspaceID uint64, req *dto.LinkRouteTestRequest) (*dto.LinkRouteTestResponse, error) {
	shortLink, err := s.ensureShortLink(shortLinkID, workspaceID)
	if err != nil {
		return nil, err
	}
	result, err := s.Resolve(shortLink, RouteResolveInput{
		ClientIP:       req.ClientIP,
		UserAgent:      req.UserAgent,
		AcceptLanguage: req.AcceptLanguage,
		Referer:        req.Referer,
		QueryString:    strings.TrimPrefix(req.Query, "?"),
	})
	if err != nil {
		return nil, err
	}
	resp := &dto.LinkRouteTestResponse{
		Matched:      result.Matched,
		TargetURL:    result.TargetURL,
		FallbackUsed: result.FallbackUsed,
		Reason:       result.Reason,
	}
	if result.Route != nil {
		resp.RouteID = result.Route.ID
		resp.RouteName = result.Route.Name
	}
	return resp, nil
}

func (s *LinkRouteService) Resolve(shortLink *model.ShortLink, input RouteResolveInput) (*RouteResolveResult, error) {
	routes, err := s.loadRoutes(shortLink.ID, shortLink.WorkspaceID, true)
	if err != nil {
		if isMissingSecurityTableError(err) {
			return &RouteResolveResult{TargetURL: shortLink.OriginalURL, Reason: "未配置高级路由"}, nil
		}
		return nil, err
	}
	if len(routes) == 0 {
		return &RouteResolveResult{TargetURL: shortLink.OriginalURL, Reason: "未配置高级路由"}, nil
	}

	context := s.buildMatchContext(input)
	for i := range routes {
		route := &routes[i]
		if s.matchRoute(route, context) {
			return &RouteResolveResult{
				RoutingEnabled: true,
				Matched:        true,
				Route:          route,
				TargetURL:      route.TargetURL,
				Reason:         "命中高级路由",
			}, nil
		}
	}

	if shortLink.FallbackURL != "" {
		return &RouteResolveResult{
			RoutingEnabled: true,
			FallbackUsed:   true,
			TargetURL:      shortLink.FallbackURL,
			Reason:         "未命中路由，使用兜底地址",
		}, nil
	}
	return &RouteResolveResult{
		RoutingEnabled: true,
		TargetURL:      shortLink.OriginalURL,
		Reason:         "未命中路由，使用原始 URL",
	}, nil
}

func (s *LinkRouteService) RoutingSummary(shortLinkID, workspaceID uint64, fallbackURL string) (bool, string) {
	var total int64
	var active int64
	db := s.helper.GetDatabase().Model(&model.LinkRoute{}).
		Where("short_link_id = ? AND workspace_id = ? AND deleted_at IS NULL", shortLinkID, workspaceID)
	if err := db.Count(&total).Error; err != nil {
		if isMissingSecurityTableError(err) {
			return false, "未配置"
		}
		return false, "未配置"
	}
	if total == 0 {
		return false, "未配置"
	}
	db.Where("is_active = ?", true).Count(&active)
	parts := make([]string, 0, 2)
	if active > 0 {
		parts = append(parts, "已启用")
	} else {
		parts = append(parts, "已停用")
	}
	if fallbackURL != "" {
		parts = append(parts, "有兜底")
	}
	return active > 0, strings.Join(parts, " / ")
}

func (s *LinkRouteService) validateRouteRequest(workspaceID uint64, req *dto.LinkRouteRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("路由名称不能为空")
	}
	if strings.TrimSpace(req.TargetURL) == "" {
		return errors.New("目标 URL 不能为空")
	}
	if _, err := parseTargetURL(req.TargetURL); err != nil {
		return errors.New("目标 URL 格式无效")
	}
	if result := s.securityService.ScanURL(workspaceID, req.TargetURL); !result.Safe {
		return errors.New("目标 URL 命中安全规则: " + result.Reason)
	}
	if len(req.ConditionGroups) == 0 {
		return errors.New("至少需要一个条件组")
	}
	for _, group := range req.ConditionGroups {
		if len(group.Conditions) == 0 {
			return errors.New("条件组不能为空")
		}
		for _, condition := range group.Conditions {
			if err := validateRouteCondition(condition); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *LinkRouteService) replaceConditionGroups(tx *gorm.DB, routeID uint64, groups []dto.LinkRouteConditionGroupRequest) error {
	var existing []model.LinkRouteConditionGroup
	if err := tx.Where("route_id = ?", routeID).Find(&existing).Error; err != nil {
		return err
	}
	for _, group := range existing {
		if err := tx.Where("group_id = ?", group.ID).Delete(&model.LinkRouteCondition{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("route_id = ?", routeID).Delete(&model.LinkRouteConditionGroup{}).Error; err != nil {
		return err
	}
	for groupIndex, groupReq := range groups {
		group := model.LinkRouteConditionGroup{
			RouteID:  routeID,
			Position: groupIndex,
		}
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		conditions := make([]model.LinkRouteCondition, 0, len(groupReq.Conditions))
		for conditionIndex, conditionReq := range groupReq.Conditions {
			conditions = append(conditions, model.LinkRouteCondition{
				GroupID:        group.ID,
				ConditionType:  conditionReq.ConditionType,
				Operator:       conditionReq.Operator,
				ConditionKey:   strings.TrimSpace(conditionReq.ConditionKey),
				ConditionValue: strings.TrimSpace(conditionReq.ConditionValue),
				Position:       conditionIndex,
			})
		}
		if len(conditions) > 0 {
			if err := tx.Create(&conditions).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *LinkRouteService) ensureShortLink(shortLinkID, workspaceID uint64) (*model.ShortLink, error) {
	var shortLink model.ShortLink
	if err := s.helper.GetDatabase().
		Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", shortLinkID, workspaceID).
		First(&shortLink).Error; err != nil {
		return nil, errors.New("短网址不存在")
	}
	return &shortLink, nil
}

func (s *LinkRouteService) findRoute(routeID, shortLinkID, workspaceID uint64) (*model.LinkRoute, error) {
	var route model.LinkRoute
	err := s.helper.GetDatabase().
		Preload("ConditionGroups", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC, id ASC") }).
		Preload("ConditionGroups.Conditions", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC, id ASC") }).
		Where("id = ? AND short_link_id = ? AND workspace_id = ? AND deleted_at IS NULL", routeID, shortLinkID, workspaceID).
		First(&route).Error
	if err != nil {
		return nil, errors.New("路由规则不存在")
	}
	return &route, nil
}

func (s *LinkRouteService) loadRoutes(shortLinkID, workspaceID uint64, activeOnly bool) ([]model.LinkRoute, error) {
	var routes []model.LinkRoute
	query := s.helper.GetDatabase().
		Preload("ConditionGroups", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC, id ASC") }).
		Preload("ConditionGroups.Conditions", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC, id ASC") }).
		Where("short_link_id = ? AND workspace_id = ? AND deleted_at IS NULL", shortLinkID, workspaceID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("priority ASC, id ASC").Find(&routes).Error
	return routes, err
}

type routeMatchContext struct {
	Country     string
	Province    string
	City        string
	DeviceType  string
	Browser     string
	OS          string
	Language    string
	Referer     string
	QueryValues url.Values
}

func (s *LinkRouteService) buildMatchContext(input RouteResolveInput) routeMatchContext {
	region := s.helper.GetIPRegion().Lookup(input.ClientIP)
	metadata := parseTrafficMetadata(input.UserAgent)
	queryValues, _ := url.ParseQuery(strings.TrimPrefix(input.QueryString, "?"))
	return routeMatchContext{
		Country:     strings.TrimSpace(region.Country),
		Province:    strings.TrimSpace(region.Province),
		City:        strings.TrimSpace(region.City),
		DeviceType:  strings.ToLower(strings.TrimSpace(metadata.DeviceType)),
		Browser:     strings.ToLower(strings.TrimSpace(metadata.Browser)),
		OS:          strings.ToLower(strings.TrimSpace(metadata.OS)),
		Language:    firstLanguage(input.AcceptLanguage),
		Referer:     strings.ToLower(strings.TrimSpace(input.Referer)),
		QueryValues: queryValues,
	}
}

func (s *LinkRouteService) matchRoute(route *model.LinkRoute, context routeMatchContext) bool {
	if len(route.ConditionGroups) == 0 {
		return false
	}
	for _, group := range route.ConditionGroups {
		if len(group.Conditions) == 0 {
			continue
		}
		allMatched := true
		for _, condition := range group.Conditions {
			if !matchCondition(condition, context) {
				allMatched = false
				break
			}
		}
		if allMatched {
			return true
		}
	}
	return false
}

func validateRouteCondition(condition dto.LinkRouteConditionRequest) error {
	conditionType := condition.ConditionType
	operator := condition.Operator
	switch conditionType {
	case model.RouteConditionCountry, model.RouteConditionProvince, model.RouteConditionCity,
		model.RouteConditionDeviceType, model.RouteConditionBrowser, model.RouteConditionOS:
		if operator != model.RouteOperatorEq && operator != model.RouteOperatorIn {
			return errors.New("该条件类型仅支持 eq 或 in")
		}
	case model.RouteConditionLanguage:
		if operator != model.RouteOperatorEq && operator != model.RouteOperatorIn && operator != model.RouteOperatorPrefix {
			return errors.New("语言条件仅支持 eq、in 或 prefix")
		}
	case model.RouteConditionReferer:
		if operator != model.RouteOperatorEq && operator != model.RouteOperatorContains && operator != model.RouteOperatorPrefix && operator != model.RouteOperatorSuffix {
			return errors.New("来源 Referer 条件仅支持 eq、contains、prefix 或 suffix")
		}
	case model.RouteConditionQueryParam:
		if strings.TrimSpace(condition.ConditionKey) == "" {
			return errors.New("Query 参数条件必须填写参数名")
		}
		if operator != model.RouteOperatorExists && operator != model.RouteOperatorEq && operator != model.RouteOperatorIn &&
			operator != model.RouteOperatorContains && operator != model.RouteOperatorPrefix && operator != model.RouteOperatorSuffix {
			return errors.New("Query 参数条件操作符不支持")
		}
	default:
		return errors.New("不支持的路由条件类型")
	}
	if operator != model.RouteOperatorExists && strings.TrimSpace(condition.ConditionValue) == "" {
		return errors.New("条件值不能为空")
	}
	return nil
}

func matchCondition(condition model.LinkRouteCondition, context routeMatchContext) bool {
	actual := ""
	switch condition.ConditionType {
	case model.RouteConditionCountry:
		actual = context.Country
	case model.RouteConditionProvince:
		actual = context.Province
	case model.RouteConditionCity:
		actual = context.City
	case model.RouteConditionDeviceType:
		actual = context.DeviceType
	case model.RouteConditionBrowser:
		actual = context.Browser
	case model.RouteConditionOS:
		actual = context.OS
	case model.RouteConditionLanguage:
		actual = context.Language
	case model.RouteConditionReferer:
		actual = context.Referer
	case model.RouteConditionQueryParam:
		values, exists := context.QueryValues[condition.ConditionKey]
		if condition.Operator == model.RouteOperatorExists {
			return exists
		}
		if !exists || len(values) == 0 {
			return false
		}
		for _, value := range values {
			if matchScalar(value, condition.Operator, condition.ConditionValue) {
				return true
			}
		}
		return false
	default:
		return false
	}
	return matchScalar(actual, condition.Operator, condition.ConditionValue)
}

func matchScalar(actual, operator, expected string) bool {
	actual = strings.ToLower(strings.TrimSpace(actual))
	expected = strings.ToLower(strings.TrimSpace(expected))
	values := splitRouteValues(expected)
	switch operator {
	case model.RouteOperatorEq:
		return actual != "" && actual == expected
	case model.RouteOperatorIn:
		for _, value := range values {
			if actual != "" && actual == value {
				return true
			}
		}
		return false
	case model.RouteOperatorContains:
		return actual != "" && expected != "" && strings.Contains(actual, expected)
	case model.RouteOperatorPrefix:
		return actual != "" && expected != "" && strings.HasPrefix(actual, expected)
	case model.RouteOperatorSuffix:
		return actual != "" && expected != "" && strings.HasSuffix(actual, expected)
	default:
		return false
	}
}

func splitRouteValues(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			values = append(values, part)
		}
	}
	sort.Strings(values)
	return values
}

func firstLanguage(header string) string {
	if strings.TrimSpace(header) == "" {
		return ""
	}
	first := strings.Split(header, ",")[0]
	first = strings.Split(first, ";")[0]
	return strings.ToLower(strings.TrimSpace(first))
}

func linkRouteToResponse(route *model.LinkRoute) dto.LinkRouteResponse {
	groups := make([]dto.LinkRouteConditionGroupResponse, 0, len(route.ConditionGroups))
	for _, group := range route.ConditionGroups {
		conditions := make([]dto.LinkRouteConditionResponse, 0, len(group.Conditions))
		for _, condition := range group.Conditions {
			conditions = append(conditions, dto.LinkRouteConditionResponse{
				ID:             condition.ID,
				ConditionType:  condition.ConditionType,
				Operator:       condition.Operator,
				ConditionKey:   condition.ConditionKey,
				ConditionValue: condition.ConditionValue,
				Position:       condition.Position,
			})
		}
		groups = append(groups, dto.LinkRouteConditionGroupResponse{
			ID:         group.ID,
			Position:   group.Position,
			Conditions: conditions,
		})
	}
	return dto.LinkRouteResponse{
		ID:              route.ID,
		WorkspaceID:     route.WorkspaceID,
		ShortLinkID:     route.ShortLinkID,
		Name:            route.Name,
		Description:     route.Description,
		Priority:        route.Priority,
		TargetURL:       route.TargetURL,
		IsActive:        route.IsActive,
		ConditionGroups: groups,
		CreatedBy:       route.CreatedBy,
		UpdatedBy:       route.UpdatedBy,
		CreatedAt:       route.CreatedAt,
		UpdatedAt:       route.UpdatedAt,
	}
}
