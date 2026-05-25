package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/domain_validate"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrSecurityPasswordRequired = errors.New("需要访问密码")
	ErrSecurityAccessDenied     = errors.New("访问受限")
	ErrSecurityURLBlocked       = errors.New("目标 URL 存在安全风险")
)

type RedirectDecision struct {
	TargetURL     string
	ShortLink     *model.ShortLink
	Security      *model.LinkSecuritySetting
	Reason        string
	PasswordURL   string
	ReportEnabled bool
}

type urlSafetyResult struct {
	Safe   bool
	Reason string
}

type LinkSecurityService struct {
	helper interfaces.HelperInterface
}

func NewLinkSecurityService(helper interfaces.HelperInterface) *LinkSecurityService {
	return &LinkSecurityService{helper: helper}
}

func (s *LinkSecurityService) GetSecurity(shortLinkID, workspaceID uint64) (*dto.LinkSecurityResponse, error) {
	if err := s.ensureShortLinkInWorkspace(shortLinkID, workspaceID); err != nil {
		return nil, err
	}
	setting, err := s.findSetting(shortLinkID, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.defaultSecurityResponse(shortLinkID, workspaceID), nil
		}
		return nil, err
	}
	return s.settingToResponse(setting), nil
}

func (s *LinkSecurityService) UpsertSecurity(shortLinkID, workspaceID, userID uint64, req *dto.LinkSecurityRequest) (*dto.LinkSecurityResponse, error) {
	if req == nil {
		return s.GetSecurity(shortLinkID, workspaceID)
	}
	if err := s.ensureShortLinkInWorkspace(shortLinkID, workspaceID); err != nil {
		return nil, err
	}

	setting, err := s.findSetting(shortLinkID, workspaceID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		setting = &model.LinkSecuritySetting{
			WorkspaceID: workspaceID,
			ShortLinkID: shortLinkID,
			IPPolicy:    model.LinkIPPolicyOff,
			BotPolicy:   model.LinkBotPolicyRecordOnly,
			CreatedBy:   actorPtr(userID),
			UpdatedBy:   actorPtr(userID),
		}
	}

	if req.AccessWindowStart != nil && req.AccessWindowEnd != nil && req.AccessWindowEnd.Before(*req.AccessWindowStart) {
		return nil, errors.New("有效访问结束时间不能早于开始时间")
	}
	if req.MaxClicks != nil && *req.MaxClicks < 1 {
		return nil, errors.New("最大访问次数必须大于 0")
	}

	if req.Password != nil {
		password := strings.TrimSpace(*req.Password)
		if password == "" {
			setting.PasswordHash = ""
			setting.PasswordEnabled = false
		} else {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return nil, err
			}
			setting.PasswordHash = string(hash)
			setting.PasswordEnabled = true
		}
	}
	if req.PasswordEnabled != nil {
		setting.PasswordEnabled = *req.PasswordEnabled
	}
	if setting.PasswordEnabled && setting.PasswordHash == "" {
		return nil, errors.New("启用访问密码时必须设置密码")
	}

	setting.AccessWindowStart = req.AccessWindowStart
	setting.AccessWindowEnd = req.AccessWindowEnd
	setting.MaxClicks = req.MaxClicks
	if req.IPPolicy != "" {
		setting.IPPolicy = req.IPPolicy
	}
	if setting.IPPolicy == "" {
		setting.IPPolicy = model.LinkIPPolicyOff
	}
	if req.BotPolicy != "" {
		setting.BotPolicy = req.BotPolicy
	}
	if setting.BotPolicy == "" {
		setting.BotPolicy = model.LinkBotPolicyRecordOnly
	}
	if req.ReportEnabled != nil {
		setting.ReportEnabled = *req.ReportEnabled
	}
	if userID > 0 {
		setting.UpdatedBy = &userID
	}

	db := s.helper.GetDatabase()
	if err := db.Save(setting).Error; err != nil {
		return nil, err
	}

	if req.IPRules != nil {
		if err := s.replaceIPRules(workspaceID, shortLinkID, req.IPRules); err != nil {
			return nil, err
		}
	}
	return s.settingToResponse(setting), nil
}

func (s *LinkSecurityService) ApplyCreateSecurity(shortLink *model.ShortLink, userID uint64, req *dto.LinkSecurityRequest) error {
	if req == nil {
		return nil
	}
	_, err := s.UpsertSecurity(shortLink.ID, shortLink.WorkspaceID, userID, req)
	return err
}

func (s *LinkSecurityService) EvaluateRedirect(shortLink *model.ShortLink, domain, shortCode, clientIP, userAgent, referer, accessToken string) (*model.LinkSecuritySetting, error) {
	setting, err := s.findSetting(shortLink.ID, shortLink.WorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	now := time.Now()

	if setting.AccessWindowStart != nil && now.Before(*setting.AccessWindowStart) {
		s.recordEvent(shortLink, model.SecurityEventAccessDenied, "不在有效访问时间内", clientIP, userAgent, referer)
		return setting, ErrSecurityAccessDenied
	}
	if setting.AccessWindowEnd != nil && now.After(*setting.AccessWindowEnd) {
		s.recordEvent(shortLink, model.SecurityEventAccessDenied, "不在有效访问时间内", clientIP, userAgent, referer)
		return setting, ErrSecurityAccessDenied
	}
	if setting.MaxClicks != nil && shortLink.ClickCount >= *setting.MaxClicks {
		s.recordEvent(shortLink, model.SecurityEventAccessDenied, "已达到最大访问次数", clientIP, userAgent, referer)
		return setting, ErrSecurityAccessDenied
	}
	if denied, reason := s.evaluateIPPolicy(setting, clientIP); denied {
		s.recordEvent(shortLink, model.SecurityEventAccessDenied, reason, clientIP, userAgent, referer)
		return setting, ErrSecurityAccessDenied
	}
	if setting.BotPolicy == model.LinkBotPolicyBlockKnownBots && parseTrafficMetadata(userAgent).IsBot {
		s.recordEvent(shortLink, model.SecurityEventBotBlocked, "已识别 Bot 访问", clientIP, userAgent, referer)
		return setting, ErrSecurityAccessDenied
	}
	if setting.URLBlocked {
		reason := setting.URLBlockedReason
		if reason == "" {
			reason = "目标 URL 命中安全规则"
		}
		s.recordEvent(shortLink, model.SecurityEventURLBlocked, reason, clientIP, userAgent, referer)
		return setting, ErrSecurityURLBlocked
	}
	if setting.PasswordEnabled && !s.verifyAccessToken(domain, shortCode, accessToken, setting.PasswordHash) {
		s.recordEvent(shortLink, model.SecurityEventPasswordRequired, "需要访问密码", clientIP, userAgent, referer)
		return setting, ErrSecurityPasswordRequired
	}
	return setting, nil
}

func (s *LinkSecurityService) VerifyPassword(domain, shortCode, password, clientIP, userAgent string) (string, string, int, error) {
	var shortLink model.ShortLink
	if err := s.helper.GetDatabase().
		Where("domain = ? AND short_code = ? AND deleted_at IS NULL", domain, shortCode).
		First(&shortLink).Error; err != nil {
		return "", "", 0, errors.New("短网址不存在")
	}
	setting, err := s.findSetting(shortLink.ID, shortLink.WorkspaceID)
	if err != nil || !setting.PasswordEnabled || setting.PasswordHash == "" {
		return "", "", 0, errors.New("该短网址未启用访问密码")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(setting.PasswordHash), []byte(password)); err != nil {
		s.recordEvent(&shortLink, model.SecurityEventPasswordFailed, "访问密码错误", clientIP, userAgent, "")
		return "", "", 0, errors.New("访问密码错误")
	}
	maxAge := int((12 * time.Hour).Seconds())
	return s.accessCookieName(domain, shortCode), s.signAccessToken(domain, shortCode, setting.PasswordHash, time.Now().Add(12*time.Hour)), maxAge, nil
}

func (s *LinkSecurityService) ScanURL(workspaceID uint64, rawURL string) urlSafetyResult {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return urlSafetyResult{Safe: false, Reason: "URL 格式无效"}
	}
	host := strings.ToLower(parsed.Hostname())
	rawLower := strings.ToLower(rawURL)

	var rules []model.SecurityURLRule
	if err := s.helper.GetDatabase().
		Where("workspace_id = ? AND enabled = ? AND deleted_at IS NULL", workspaceID, true).
		Order("action ASC").
		Find(&rules).Error; err != nil {
		if isMissingSecurityTableError(err) {
			return urlSafetyResult{Safe: true}
		}
		return urlSafetyResult{Safe: true}
	}

	for _, rule := range rules {
		if rule.Action == model.SecurityRuleActionAllow && s.matchURLRule(rule, host, rawLower) {
			return urlSafetyResult{Safe: true}
		}
	}
	for _, rule := range rules {
		if rule.Action == model.SecurityRuleActionBlock && s.matchURLRule(rule, host, rawLower) {
			return urlSafetyResult{Safe: false, Reason: fmt.Sprintf("命中%s安全规则: %s", securityRuleTypeLabel(rule.RuleType), rule.Pattern)}
		}
	}
	return urlSafetyResult{Safe: true}
}

func (s *LinkSecurityService) RescanShortLink(shortLinkID, workspaceID uint64) (*dto.LinkSecurityResponse, error) {
	var shortLink model.ShortLink
	if err := s.helper.GetDatabase().
		Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", shortLinkID, workspaceID).
		First(&shortLink).Error; err != nil {
		return nil, errors.New("短网址不存在")
	}

	result := s.ScanURL(workspaceID, shortLink.OriginalURL)
	setting, err := s.findSetting(shortLinkID, workspaceID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		setting = &model.LinkSecuritySetting{
			WorkspaceID: workspaceID,
			ShortLinkID: shortLinkID,
			IPPolicy:    model.LinkIPPolicyOff,
			BotPolicy:   model.LinkBotPolicyRecordOnly,
		}
	}
	setting.URLBlocked = !result.Safe
	setting.URLBlockedReason = result.Reason
	if err := s.helper.GetDatabase().Save(setting).Error; err != nil {
		return nil, err
	}
	if !result.Safe {
		s.recordEvent(&shortLink, model.SecurityEventURLBlocked, result.Reason, "", "", "")
	}
	return s.settingToResponse(setting), nil
}

func (s *LinkSecurityService) CreateURLRule(workspaceID, userID uint64, req *dto.SecurityURLRuleRequest) (*dto.SecurityURLRuleResponse, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	pattern := normalizeSecurityRulePattern(req.RuleType, req.Pattern)
	if pattern == "" {
		return nil, errors.New("规则内容不能为空")
	}
	rule := &model.SecurityURLRule{
		WorkspaceID: workspaceID,
		RuleType:    req.RuleType,
		Action:      req.Action,
		Pattern:     pattern,
		Enabled:     enabled,
		CreatedBy:   actorPtr(userID),
	}
	if err := s.helper.GetDatabase().Create(rule).Error; err != nil {
		return nil, err
	}
	resp := securityURLRuleToResponse(rule)
	return &resp, nil
}

func (s *LinkSecurityService) UpdateURLRule(id, workspaceID uint64, req *dto.SecurityURLRuleRequest) (*dto.SecurityURLRuleResponse, error) {
	var rule model.SecurityURLRule
	if err := s.helper.GetDatabase().
		Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID).
		First(&rule).Error; err != nil {
		return nil, errors.New("安全规则不存在")
	}
	rule.RuleType = req.RuleType
	rule.Action = req.Action
	rule.Pattern = normalizeSecurityRulePattern(req.RuleType, req.Pattern)
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if rule.Pattern == "" {
		return nil, errors.New("规则内容不能为空")
	}
	if err := s.helper.GetDatabase().Save(&rule).Error; err != nil {
		return nil, err
	}
	resp := securityURLRuleToResponse(&rule)
	return &resp, nil
}

func (s *LinkSecurityService) DeleteURLRule(id, workspaceID uint64) error {
	return s.helper.GetDatabase().
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		Delete(&model.SecurityURLRule{}).Error
}

func (s *LinkSecurityService) ListURLRules(workspaceID uint64, req *dto.SecurityURLRuleListRequest) (*dto.SecurityURLRuleListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	query := s.helper.GetDatabase().Model(&model.SecurityURLRule{}).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID)
	if req.RuleType != "" {
		query = query.Where("rule_type = ?", req.RuleType)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Keyword != "" {
		query = query.Where("pattern LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rules []model.SecurityURLRule
	if err := query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rules).Error; err != nil {
		return nil, err
	}
	list := make([]dto.SecurityURLRuleResponse, 0, len(rules))
	for _, rule := range rules {
		list = append(list, securityURLRuleToResponse(&rule))
	}
	return &dto.SecurityURLRuleListResponse{List: list, Total: total, Page: req.Page, Size: req.PageSize}, nil
}

func (s *LinkSecurityService) ListEvents(workspaceID uint64, req *dto.SecurityEventListRequest) (*dto.SecurityEventListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	query := s.helper.GetDatabase().Model(&model.LinkSecurityEvent{}).Where("workspace_id = ?", workspaceID)
	if req.ShortLinkID > 0 {
		query = query.Where("short_link_id = ?", req.ShortLinkID)
	}
	if req.EventType != "" {
		query = query.Where("event_type = ?", req.EventType)
	}
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		end := req.EndDate.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", end)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var events []model.LinkSecurityEvent
	if err := query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&events).Error; err != nil {
		return nil, err
	}
	list := make([]dto.SecurityEventResponse, 0, len(events))
	for _, event := range events {
		list = append(list, dto.SecurityEventResponse{
			ID:          event.ID,
			WorkspaceID: event.WorkspaceID,
			ShortLinkID: event.ShortLinkID,
			EventType:   event.EventType,
			Reason:      event.Reason,
			ClientIP:    event.ClientIP,
			UserAgent:   event.UserAgent,
			Referer:     event.Referer,
			CreatedAt:   event.CreatedAt,
		})
	}
	return &dto.SecurityEventListResponse{List: list, Total: total, Page: req.Page, Size: req.PageSize}, nil
}

func (s *LinkSecurityService) CreateAbuseReport(req *dto.AbuseReportCreateRequest, clientIP, userAgent string) (*dto.AbuseReportResponse, error) {
	shortLink, err := s.resolveReportedShortLink(req)
	if err != nil {
		return nil, err
	}
	setting, err := s.findSetting(shortLink.ID, shortLink.WorkspaceID)
	if err != nil || !setting.ReportEnabled {
		return nil, errors.New("该短网址未开启举报入口")
	}
	var duplicates int64
	if err := s.helper.GetDatabase().Model(&model.AbuseReport{}).
		Where("short_link_id = ? AND reporter_ip = ? AND status IN ? AND created_at >= ?",
			shortLink.ID, clientIP, []string{model.AbuseReportStatusPending, model.AbuseReportStatusReviewing}, time.Now().Add(-time.Hour)).
		Count(&duplicates).Error; err != nil {
		return nil, err
	}
	if duplicates > 0 {
		return nil, errors.New("已收到举报，请勿重复提交")
	}
	report := &model.AbuseReport{
		WorkspaceID:   shortLink.WorkspaceID,
		ShortLinkID:   shortLink.ID,
		ReportType:    req.ReportType,
		Description:   req.Description,
		ReporterEmail: req.ReporterEmail,
		ReporterIP:    domain_validate.TruncateString(clientIP, 45),
		UserAgent:     domain_validate.TruncateString(userAgent, 1024),
		Status:        model.AbuseReportStatusPending,
	}
	if err := s.helper.GetDatabase().Create(report).Error; err != nil {
		return nil, err
	}
	s.recordEvent(shortLink, model.SecurityEventAbuseReported, "收到滥用举报", clientIP, userAgent, "")
	resp := abuseReportToResponse(report)
	return &resp, nil
}

func (s *LinkSecurityService) ListAbuseReports(workspaceID uint64, req *dto.AbuseReportListRequest) (*dto.AbuseReportListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	query := s.helper.GetDatabase().Model(&model.AbuseReport{}).Where("workspace_id = ?", workspaceID)
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.ReportType != "" {
		query = query.Where("report_type = ?", req.ReportType)
	}
	if req.ShortLinkID > 0 {
		query = query.Where("short_link_id = ?", req.ShortLinkID)
	}
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		end := req.EndDate.AddDate(0, 0, 1)
		query = query.Where("created_at < ?", end)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var reports []model.AbuseReport
	if err := query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&reports).Error; err != nil {
		return nil, err
	}
	list := make([]dto.AbuseReportResponse, 0, len(reports))
	for _, report := range reports {
		list = append(list, abuseReportToResponse(&report))
	}
	return &dto.AbuseReportListResponse{List: list, Total: total, Page: req.Page, Size: req.PageSize}, nil
}

func (s *LinkSecurityService) UpdateAbuseReport(id, workspaceID, userID uint64, req *dto.AbuseReportUpdateRequest) (*dto.AbuseReportResponse, error) {
	var report model.AbuseReport
	if err := s.helper.GetDatabase().
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		First(&report).Error; err != nil {
		return nil, errors.New("举报不存在")
	}
	now := time.Now()
	report.Status = req.Status
	report.ResolutionNote = req.ResolutionNote
	report.HandledBy = actorPtr(userID)
	report.HandledAt = &now
	if err := s.helper.GetDatabase().Save(&report).Error; err != nil {
		return nil, err
	}
	if req.DisableLink {
		s.disableReportedShortLink(report.ShortLinkID, workspaceID)
	}
	resp := abuseReportToResponse(&report)
	return &resp, nil
}

func (s *LinkSecurityService) findSetting(shortLinkID, workspaceID uint64) (*model.LinkSecuritySetting, error) {
	var setting model.LinkSecuritySetting
	err := s.helper.GetDatabase().
		Where("short_link_id = ? AND workspace_id = ? AND deleted_at IS NULL", shortLinkID, workspaceID).
		First(&setting).Error
	if isMissingSecurityTableError(err) {
		return nil, gorm.ErrRecordNotFound
	}
	return &setting, err
}

func (s *LinkSecurityService) ensureShortLinkInWorkspace(shortLinkID, workspaceID uint64) error {
	var count int64
	if err := s.helper.GetDatabase().Model(&model.ShortLink{}).
		Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", shortLinkID, workspaceID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("短网址不存在")
	}
	return nil
}

func (s *LinkSecurityService) replaceIPRules(workspaceID, shortLinkID uint64, rules []dto.LinkSecurityIPRuleRequest) error {
	db := s.helper.GetDatabase()
	if err := db.Where("short_link_id = ? AND workspace_id = ?", shortLinkID, workspaceID).
		Delete(&model.LinkSecurityIPRule{}).Error; err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}
	entities := make([]model.LinkSecurityIPRule, 0, len(rules))
	for _, item := range rules {
		normalized, err := normalizeCIDR(item.CIDR)
		if err != nil {
			return fmt.Errorf("无效的 IP 规则 %s", item.CIDR)
		}
		entities = append(entities, model.LinkSecurityIPRule{
			WorkspaceID: workspaceID,
			ShortLinkID: shortLinkID,
			CIDR:        normalized,
			Description: item.Description,
		})
	}
	return db.Create(&entities).Error
}

func (s *LinkSecurityService) evaluateIPPolicy(setting *model.LinkSecuritySetting, clientIP string) (bool, string) {
	if setting.IPPolicy == "" || setting.IPPolicy == model.LinkIPPolicyOff || clientIP == "" {
		return false, ""
	}
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return true, "无法识别访问 IP"
	}
	var rules []model.LinkSecurityIPRule
	if err := s.helper.GetDatabase().
		Where("short_link_id = ? AND workspace_id = ? AND deleted_at IS NULL", setting.ShortLinkID, setting.WorkspaceID).
		Find(&rules).Error; err != nil {
		return true, "IP 规则读取失败"
	}
	matched := false
	for _, rule := range rules {
		_, network, err := net.ParseCIDR(rule.CIDR)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			matched = true
			break
		}
	}
	switch setting.IPPolicy {
	case model.LinkIPPolicyAllowlist:
		if !matched {
			return true, "IP 不在允许范围内"
		}
	case model.LinkIPPolicyBlocklist:
		if matched {
			return true, "IP 在阻止范围内"
		}
	}
	return false, ""
}

func (s *LinkSecurityService) recordEvent(shortLink *model.ShortLink, eventType, reason, clientIP, userAgent, referer string) {
	if shortLink == nil {
		return
	}
	event := &model.LinkSecurityEvent{
		WorkspaceID: shortLink.WorkspaceID,
		ShortLinkID: shortLink.ID,
		EventType:   eventType,
		Reason:      domain_validate.TruncateString(reason, 500),
		ClientIP:    domain_validate.TruncateString(clientIP, 45),
		UserAgent:   domain_validate.TruncateString(userAgent, 1024),
		Referer:     domain_validate.TruncateString(referer, 2048),
	}
	if err := s.helper.GetDatabase().Create(event).Error; err != nil {
		if isMissingSecurityTableError(err) {
			return
		}
		s.helper.GetLogger().Warn("[link-security] 记录安全事件失败: " + err.Error())
	}
}

func (s *LinkSecurityService) settingToResponse(setting *model.LinkSecuritySetting) *dto.LinkSecurityResponse {
	var rules []model.LinkSecurityIPRule
	_ = s.helper.GetDatabase().
		Where("short_link_id = ? AND workspace_id = ? AND deleted_at IS NULL", setting.ShortLinkID, setting.WorkspaceID).
		Order("created_at ASC").Find(&rules).Error
	ipRules := make([]dto.LinkSecurityIPRuleResponse, 0, len(rules))
	for _, rule := range rules {
		ipRules = append(ipRules, dto.LinkSecurityIPRuleResponse{
			ID:          rule.ID,
			WorkspaceID: rule.WorkspaceID,
			ShortLinkID: rule.ShortLinkID,
			CIDR:        rule.CIDR,
			Description: rule.Description,
			CreatedAt:   rule.CreatedAt,
			UpdatedAt:   rule.UpdatedAt,
		})
	}
	summary, enabled := securitySummary(setting)
	return &dto.LinkSecurityResponse{
		ID:                setting.ID,
		WorkspaceID:       setting.WorkspaceID,
		ShortLinkID:       setting.ShortLinkID,
		PasswordEnabled:   setting.PasswordEnabled,
		AccessWindowStart: setting.AccessWindowStart,
		AccessWindowEnd:   setting.AccessWindowEnd,
		MaxClicks:         setting.MaxClicks,
		IPPolicy:          fallbackString(setting.IPPolicy, model.LinkIPPolicyOff),
		IPRules:           ipRules,
		BotPolicy:         fallbackString(setting.BotPolicy, model.LinkBotPolicyRecordOnly),
		ReportEnabled:     setting.ReportEnabled,
		URLBlocked:        setting.URLBlocked,
		URLBlockedReason:  setting.URLBlockedReason,
		SecurityEnabled:   enabled,
		SecuritySummary:   summary,
		CreatedBy:         setting.CreatedBy,
		UpdatedBy:         setting.UpdatedBy,
		CreatedAt:         setting.CreatedAt,
		UpdatedAt:         setting.UpdatedAt,
	}
}

func (s *LinkSecurityService) defaultSecurityResponse(shortLinkID, workspaceID uint64) *dto.LinkSecurityResponse {
	return &dto.LinkSecurityResponse{
		WorkspaceID:     workspaceID,
		ShortLinkID:     shortLinkID,
		IPPolicy:        model.LinkIPPolicyOff,
		BotPolicy:       model.LinkBotPolicyRecordOnly,
		SecurityEnabled: false,
		SecuritySummary: "未启用",
		IPRules:         []dto.LinkSecurityIPRuleResponse{},
	}
}

func (s *LinkSecurityService) accessCookieName(domain, shortCode string) string {
	sum := sha1.Sum([]byte(domain + "|" + shortCode))
	return "dwz_link_access_" + hex.EncodeToString(sum[:])[:16]
}

func (s *LinkSecurityService) AccessCookieName(domain, shortCode string) string {
	return s.accessCookieName(domain, shortCode)
}

func (s *LinkSecurityService) signAccessToken(domain, shortCode, passwordHash string, expiresAt time.Time) string {
	exp := strconv.FormatInt(expiresAt.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(s.cookieSecret()))
	_, _ = mac.Write([]byte(domain + "|" + shortCode + "|" + passwordHash + "|" + exp))
	return exp + "." + hex.EncodeToString(mac.Sum(nil))
}

func (s *LinkSecurityService) verifyAccessToken(domain, shortCode, token, passwordHash string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	expUnix, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().Unix() > expUnix {
		return false
	}
	expected := s.signAccessToken(domain, shortCode, passwordHash, time.Unix(expUnix, 0))
	return hmac.Equal([]byte(expected), []byte(token))
}

func (s *LinkSecurityService) cookieSecret() string {
	secret := s.helper.GetConfig().GetString("jwt.secret", "")
	if secret == "" {
		secret = "dwz-link-access"
	}
	return secret
}

func (s *LinkSecurityService) matchURLRule(rule model.SecurityURLRule, host, rawLower string) bool {
	pattern := normalizeSecurityRulePattern(rule.RuleType, rule.Pattern)
	switch rule.RuleType {
	case model.SecurityRuleTypeDomain:
		return host == pattern || strings.HasSuffix(host, "."+pattern)
	case model.SecurityRuleTypeKeyword:
		return strings.Contains(rawLower, strings.ToLower(pattern))
	default:
		return false
	}
}

func (s *LinkSecurityService) resolveReportedShortLink(req *dto.AbuseReportCreateRequest) (*model.ShortLink, error) {
	var shortLink model.ShortLink
	query := s.helper.GetDatabase().Where("deleted_at IS NULL")
	if req.ShortLinkID > 0 {
		query = query.Where("id = ?", req.ShortLinkID)
	} else if req.Domain != "" && req.ShortCode != "" {
		query = query.Where("domain = ? AND short_code = ?", req.Domain, req.ShortCode)
	} else {
		return nil, errors.New("缺少举报短网址")
	}
	if err := query.First(&shortLink).Error; err != nil {
		return nil, errors.New("短网址不存在")
	}
	return &shortLink, nil
}

func (s *LinkSecurityService) disableReportedShortLink(shortLinkID, workspaceID uint64) {
	var shortLink model.ShortLink
	if err := s.helper.GetDatabase().
		Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", shortLinkID, workspaceID).
		First(&shortLink).Error; err != nil {
		return
	}
	shortLink.IsActive = false
	if err := s.helper.GetDatabase().Save(&shortLink).Error; err == nil {
		key := fmt.Sprintf("shortlink:%s:%s", shortLink.Domain, shortLink.GetShortCode())
		_ = s.helper.GetCache().Del(context.Background(), key)
	}
}

func securitySummary(setting *model.LinkSecuritySetting) (string, bool) {
	if setting == nil {
		return "未启用", false
	}
	parts := make([]string, 0, 5)
	if setting.URLBlocked {
		parts = append(parts, "URL 风险")
	}
	if setting.PasswordEnabled {
		parts = append(parts, "访问密码")
	}
	if setting.AccessWindowStart != nil || setting.AccessWindowEnd != nil || setting.MaxClicks != nil || setting.IPPolicy != "" && setting.IPPolicy != model.LinkIPPolicyOff || setting.BotPolicy == model.LinkBotPolicyBlockKnownBots {
		parts = append(parts, "访问受限")
	}
	if setting.ReportEnabled {
		parts = append(parts, "举报入口")
	}
	if len(parts) == 0 {
		return "未启用", false
	}
	return strings.Join(parts, " / "), true
}

func securityURLRuleToResponse(rule *model.SecurityURLRule) dto.SecurityURLRuleResponse {
	return dto.SecurityURLRuleResponse{
		ID:          rule.ID,
		WorkspaceID: rule.WorkspaceID,
		RuleType:    rule.RuleType,
		Action:      rule.Action,
		Pattern:     rule.Pattern,
		Enabled:     rule.Enabled,
		CreatedBy:   rule.CreatedBy,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

func abuseReportToResponse(report *model.AbuseReport) dto.AbuseReportResponse {
	return dto.AbuseReportResponse{
		ID:             report.ID,
		WorkspaceID:    report.WorkspaceID,
		ShortLinkID:    report.ShortLinkID,
		ReportType:     report.ReportType,
		Description:    report.Description,
		ReporterEmail:  report.ReporterEmail,
		ReporterIP:     report.ReporterIP,
		UserAgent:      report.UserAgent,
		Status:         report.Status,
		ResolutionNote: report.ResolutionNote,
		HandledBy:      report.HandledBy,
		HandledAt:      report.HandledAt,
		CreatedAt:      report.CreatedAt,
		UpdatedAt:      report.UpdatedAt,
	}
}

func normalizeCIDR(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("empty cidr")
	}
	if strings.Contains(value, "/") {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return "", err
		}
		return prefix.Masked().String(), nil
	}
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return "", err
	}
	if addr.Is4() {
		return addr.String() + "/32", nil
	}
	return addr.String() + "/128", nil
}

func normalizeSecurityRulePattern(ruleType, pattern string) string {
	pattern = strings.TrimSpace(strings.ToLower(pattern))
	if ruleType == model.SecurityRuleTypeDomain {
		pattern = strings.TrimPrefix(pattern, "http://")
		pattern = strings.TrimPrefix(pattern, "https://")
		if i := strings.Index(pattern, "/"); i >= 0 {
			pattern = pattern[:i]
		}
		pattern = strings.TrimSuffix(pattern, ".")
	}
	return pattern
}

func securityRuleTypeLabel(ruleType string) string {
	if ruleType == model.SecurityRuleTypeDomain {
		return "域名"
	}
	return "关键词"
}

func fallbackString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func isMissingSecurityTableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table") ||
		strings.Contains(msg, "doesn't exist") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "sqlstate 42p01")
}

func actorPtr(userID uint64) *uint64 {
	if userID == 0 {
		return nil
	}
	return &userID
}
