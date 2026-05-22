package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	helper2 "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/domain_validate"
	"gorm.io/gorm"
)

type ShortLinkService struct {
	helper            interfaces.HelperInterface
	context           context.Context
	shortLinkDao      *dao.ShortLinkDao
	clickStatisticDao *dao.ClickStatisticDao
	domainDao         *dao.DomainDao
	campaignDao       *dao.CampaignDao
	tagDao            *dao.TagDao
	idGenerator       interfaces.IDGenerator // 新的分布式发号器
	abTestService     *ABTestService         // AB测试服务
}

// LoggerAdapter 适配器，让zap日志符合util.Logger接口
type LoggerAdapter struct{}

func NewShortLinkService(helper interfaces.HelperInterface, context context.Context) *ShortLinkService {
	return &ShortLinkService{
		helper:            helper,
		context:           context,
		shortLinkDao:      dao.NewShortLinkDao(helper),
		clickStatisticDao: dao.NewClickStatisticDao(helper),
		domainDao:         dao.NewDomainDao(helper),
		campaignDao:       dao.NewCampaignDao(helper),
		tagDao:            dao.NewTagDao(helper),
		idGenerator:       helper2.GetIdGenerator(),
		abTestService:     NewABTestService(helper),
	}
}

// CreateShortLink 创建短网址
func (s *ShortLinkService) CreateShortLink(req *dto.CreateShortLinkRequest, creatorIP string) (*dto.ShortLinkResponse, error) {
	return s.CreateShortLinkInWorkspace(req, creatorIP, 1, 0)
}

func (s *ShortLinkService) CreateShortLinkInWorkspace(req *dto.CreateShortLinkRequest, creatorIP string, workspaceID, userID uint64) (*dto.ShortLinkResponse, error) {
	// 验证原始URL
	if _, err := url.ParseRequestURI(req.OriginalURL); err != nil {
		return nil, errors.New("无效的URL格式")
	}
	finalURL, err := mergeUTMToURL(req.OriginalURL, req.UTMSource, req.UTMMedium, req.UTMCampaign, req.UTMTerm, req.UTMContent)
	if err != nil {
		return nil, errors.New("无效的URL格式")
	}

	// 获取默认域名（如果没有指定）
	domain := req.Domain
	if domain == "" {
		return nil, errors.New("域名不能为空")
	}

	// 验证域名格式,域名不能带有协议头
	if err := s.validateDomain(domain); err != nil {
		return nil, err
	}

	// 验证域名是否存在且活跃并获取域名信息
	domainInfo, err := s.domainDao.FindByDomain(domain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("域名不存在")
		}
		return nil, err
	}
	if !domainInfo.IsActive {
		return nil, errors.New("域名未激活")
	}
	if domainInfo.WorkspaceID != workspaceID {
		return nil, errors.New("域名不存在")
	}

	if err := s.validateCampaignAndTags(workspaceID, req.CampaignID, req.TagIDs); err != nil {
		return nil, err
	}

	var actor *uint64
	if userID > 0 {
		actor = &userID
	}

	// 创建短网址记录
	shortLink := &model.ShortLink{
		WorkspaceID: workspaceID,
		CampaignID:  req.CampaignID,
		Domain:      domain,
		DomainID:    domainInfo.ID,
		Protocol:    domainInfo.Protocol,
		OriginalURL: finalURL,
		Title:       req.Title,
		Description: req.Description,
		UTMSource:   req.UTMSource,
		UTMMedium:   req.UTMMedium,
		UTMCampaign: req.UTMCampaign,
		UTMTerm:     req.UTMTerm,
		UTMContent:  req.UTMContent,
		Notes:       req.Notes,
		ExpireAt:    req.ExpireAt,
		IsActive:    true,
		CreatorIP:   creatorIP,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}

	// 处理自定义短代码
	if req.CustomCode != "" {

		// 检查自定义短代码是否已存在
		exists, err := s.shortLinkDao.ExistsByDomainAndCode(domain, req.CustomCode)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("自定义短代码已存在")
		}

		shortLink.ShortCode = req.CustomCode
		shortLink.IsCustomCode = true
	} else {
		// 使用分布式发号器生成短代码，使用域名配置
		// 处理指针类型，提供默认值
		randomSuffixLength := 2
		if domainInfo.RandomSuffixLength != nil {
			randomSuffixLength = *domainInfo.RandomSuffixLength
		}
		enableChecksum := true
		if domainInfo.EnableChecksum != nil {
			enableChecksum = *domainInfo.EnableChecksum
		}
		enableXorObfuscation := false
		if domainInfo.EnableXorObfuscation != nil {
			enableXorObfuscation = *domainInfo.EnableXorObfuscation
		}
		xorSecret := uint64(0)
		if domainInfo.XorSecret != nil {
			xorSecret = *domainInfo.XorSecret
		}
		xorRot := 0
		if domainInfo.XorRot != nil {
			xorRot = *domainInfo.XorRot
		}
		defaultStartNumber := uint64(0)
		if domainInfo.DefaultStartNumber != nil {
			defaultStartNumber = *domainInfo.DefaultStartNumber
		}
		config := interfaces.ShortCodeConfig{
			RandomSuffixLength:   randomSuffixLength,
			EnableChecksum:       enableChecksum,
			EnableXorObfuscation: enableXorObfuscation,
			XorSecret:            xorSecret,
			XorRot:               xorRot,
			DefaultStartNumber:   defaultStartNumber,
		}
		generatedCode, issuerNumber, err := s.idGenerator.GenerateShortCodeWithConfig(domainInfo.ID, s.context, config)
		if err != nil {
			return nil, fmt.Errorf("生成短代码失败: %v", err)
		}
		shortLink.ShortCode = generatedCode
		shortLink.IsCustomCode = false
		shortLink.IssuerNumber = issuerNumber
	}

	// 保存到数据库（使用自定义ID避免GORM自动生成）
	if err := s.shortLinkDao.Create(shortLink); err != nil {
		return nil, err
	}
	if len(req.TagIDs) > 0 {
		if err := s.tagDao.ReplaceShortLinkTags(shortLink.ID, req.TagIDs); err != nil {
			return nil, err
		}
	}

	// 缓存到Redis
	s.cacheShortLink(shortLink)

	return s.modelToResponse(shortLink), nil
}

// GetShortLink 根据ID获取短网址
func (s *ShortLinkService) GetShortLink(id uint64) (*dto.ShortLinkResponse, error) {
	return s.GetShortLinkInWorkspace(id, 1)
}

func (s *ShortLinkService) GetShortLinkInWorkspace(id, workspaceID uint64) (*dto.ShortLinkResponse, error) {
	shortLink, err := s.shortLinkDao.FindByID(id)
	if workspaceID > 0 {
		shortLink, err = s.shortLinkDao.FindByIDInWorkspace(id, workspaceID)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("短网址不存在")
		}
		return nil, err
	}

	return s.modelToResponse(shortLink), nil
}

// UpdateShortLink 更新短网址
func (s *ShortLinkService) UpdateShortLink(id uint64, req *dto.UpdateShortLinkRequest) (*dto.ShortLinkResponse, error) {
	return s.UpdateShortLinkInWorkspace(id, req, 1, 0)
}

func (s *ShortLinkService) UpdateShortLinkInWorkspace(id uint64, req *dto.UpdateShortLinkRequest, workspaceID, userID uint64) (*dto.ShortLinkResponse, error) {
	shortLink, err := s.shortLinkDao.FindByID(id)
	if workspaceID > 0 {
		shortLink, err = s.shortLinkDao.FindByIDInWorkspace(id, workspaceID)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("短网址不存在")
		}
		return nil, err
	}

	// 更新字段
	if req.OriginalURL != "" {
		if _, err := url.ParseRequestURI(req.OriginalURL); err != nil {
			return nil, errors.New("无效的URL格式")
		}
		shortLink.OriginalURL = req.OriginalURL
	}
	if err := s.validateCampaignAndTags(workspaceID, req.CampaignID, req.TagIDs); err != nil {
		return nil, err
	}
	if req.CampaignID != nil {
		shortLink.CampaignID = req.CampaignID
	}

	if req.Title != "" {
		shortLink.Title = req.Title
	}

	if req.Description != "" {
		shortLink.Description = req.Description
	}
	if req.UTMSource != "" {
		shortLink.UTMSource = req.UTMSource
	}
	if req.UTMMedium != "" {
		shortLink.UTMMedium = req.UTMMedium
	}
	if req.UTMCampaign != "" {
		shortLink.UTMCampaign = req.UTMCampaign
	}
	if req.UTMTerm != "" {
		shortLink.UTMTerm = req.UTMTerm
	}
	if req.UTMContent != "" {
		shortLink.UTMContent = req.UTMContent
	}
	if req.Notes != "" {
		shortLink.Notes = req.Notes
	}
	finalURL, err := mergeUTMToURL(shortLink.OriginalURL, shortLink.UTMSource, shortLink.UTMMedium, shortLink.UTMCampaign, shortLink.UTMTerm, shortLink.UTMContent)
	if err != nil {
		return nil, errors.New("无效的URL格式")
	}
	shortLink.OriginalURL = finalURL

	shortLink.ExpireAt = req.ExpireAt

	if req.IsActive != nil {
		shortLink.IsActive = *req.IsActive
	}
	if userID > 0 {
		shortLink.UpdatedBy = &userID
	}

	if err := s.shortLinkDao.Update(shortLink); err != nil {
		return nil, err
	}
	if req.TagIDs != nil {
		if err := s.tagDao.ReplaceShortLinkTags(shortLink.ID, req.TagIDs); err != nil {
			return nil, err
		}
	}

	// 更新缓存
	s.cacheShortLink(shortLink)

	return s.modelToResponse(shortLink), nil
}

// UpdateShortLinkStatus 更新短网址状态
func (s *ShortLinkService) UpdateShortLinkStatus(id uint64, isActive bool) (*dto.ShortLinkResponse, error) {
	return s.UpdateShortLinkStatusInWorkspace(id, isActive, 1, 0)
}

func (s *ShortLinkService) UpdateShortLinkStatusInWorkspace(id uint64, isActive bool, workspaceID, userID uint64) (*dto.ShortLinkResponse, error) {
	shortLink, err := s.shortLinkDao.FindByID(id)
	if workspaceID > 0 {
		shortLink, err = s.shortLinkDao.FindByIDInWorkspace(id, workspaceID)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("短网址不存在")
		}
		return nil, err
	}

	// 更新状态
	shortLink.IsActive = isActive
	if userID > 0 {
		shortLink.UpdatedBy = &userID
	}

	if err := s.shortLinkDao.Update(shortLink); err != nil {
		return nil, err
	}

	// 更新缓存
	s.cacheShortLink(shortLink)

	return s.modelToResponse(shortLink), nil
}

// DeleteShortLink 删除短网址
func (s *ShortLinkService) DeleteShortLink(id uint64) error {
	return s.DeleteShortLinkInWorkspace(id, 1)
}

func (s *ShortLinkService) DeleteShortLinkInWorkspace(id, workspaceID uint64) error {
	shortLink, err := s.shortLinkDao.FindByID(id)
	if workspaceID > 0 {
		shortLink, err = s.shortLinkDao.FindByIDInWorkspace(id, workspaceID)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("短网址不存在")
		}
		return err
	}

	// 检查是否已禁用，只有禁用状态才能删除
	if shortLink.IsActive {
		return errors.New("请先禁用短网址后再删除")
	}

	if err := s.shortLinkDao.Delete(id); err != nil {
		return err
	}

	// 从缓存中删除
	s.removeCacheShortLink(shortLink.Domain, shortLink.GetShortCode())

	return nil
}

// GetShortLinkList 获取短网址列表
func (s *ShortLinkService) GetShortLinkList(req *dto.ShortLinkListRequest) (*dto.ShortLinkListResponse, error) {
	return s.GetShortLinkListInWorkspace(req, 1)
}

func (s *ShortLinkService) GetShortLinkListInWorkspace(req *dto.ShortLinkListRequest, workspaceID uint64) (*dto.ShortLinkListResponse, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	shortLinks, total, err := s.shortLinkDao.ListInWorkspace(workspaceID, offset, req.PageSize, req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	responses := make([]dto.ShortLinkResponse, 0, len(shortLinks))
	for _, shortLink := range shortLinks {
		responses = append(responses, *s.modelToResponse(&shortLink))
	}

	return &dto.ShortLinkListResponse{
		List:  responses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// RedirectShortLink 短网址跳转并记录统计
func (s *ShortLinkService) RedirectShortLink(domain, shortCode, clientIP, userAgent, referer, queryParams string) (string, error) {

	// 先从缓存查找
	shortLink, err := s.getShortLinkFromCache(domain, shortCode)
	if err != nil || shortLink == nil {
		// 缓存未命中，尝试多种方式从数据库查找
		shortLink, err = s.findShortLinkByCode(domain, shortCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", errors.New("短网址不存在")
			}
			return "", err
		}

		// 缓存到Redis
		s.cacheShortLink(shortLink)
	}

	// 检查是否激活
	if !shortLink.IsActive {
		return "", errors.New("短网址已被禁用")
	}

	// 检查是否过期
	if shortLink.IsExpired() {
		return "", errors.New("短网址已过期")
	}

	// 异步记录点击统计
	if clientIP != "" { // 只有非预览请求才记录统计
		go s.recordClickStatistic(shortLink, clientIP, userAgent, referer, queryParams)
		go s.incrementClickCount(shortLink.ID)
	}

	return shortLink.OriginalURL, nil
}

// RedirectShortLinkWithQuery 短网址跳转并记录统计（支持GET参数透传）
func (s *ShortLinkService) RedirectShortLinkWithQuery(domain, shortCode, clientIP, userAgent, referer, queryString string) (string, error) {

	// 先从缓存查找
	shortLink, err := s.getShortLinkFromCache(domain, shortCode)
	if err != nil || shortLink == nil {
		s.helper.GetLogger().Warn(fmt.Sprintf("缓存未命中，尝试多种方式从数据库查找-> domain: %s, shortCode: %s, %+v", domain, shortCode, err))
		// 缓存未命中，尝试多种方式从数据库查找
		shortLink, err = s.findShortLinkByCode(domain, shortCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", errors.New("短网址不存在")
			}
			return "", err
		}

		// 缓存到Redis
		s.cacheShortLink(shortLink)
	}

	// 检查是否激活
	if !shortLink.IsActive {
		return "", errors.New("短网址已被禁用")
	}

	// 检查是否过期
	if shortLink.IsExpired() {
		return "", errors.New("短网址已过期")
	}

	// 检查是否有AB测试
	var targetURL string
	if clientIP != "" { // 只有非预览请求才记录统计和检查AB测试
		// 检查AB测试
		if abTestInfo, err := s.abTestService.GetABTestRedirectInfo(shortLink.ID, clientIP, userAgent); err == nil && abTestInfo != nil {
			// 有AB测试，使用AB测试的目标URL
			targetURL = abTestInfo.TargetURL
			go s.recordClickStatistic(shortLink, clientIP, userAgent, referer, queryString)
			go s.incrementClickCount(shortLink.ID)
			go s.abTestService.RecordABTestClick(abTestInfo, clientIP, userAgent, referer, queryString)
		} else {
			// 没有AB测试，使用原始URL
			targetURL = shortLink.OriginalURL
			go s.recordClickStatistic(shortLink, clientIP, userAgent, referer, queryString)
			go s.incrementClickCount(shortLink.ID)
		}
	} else {
		// 预览请求，直接使用原始URL
		targetURL = shortLink.OriginalURL
	}

	// 获取域名配置以确定是否透传GET参数
	domainInfo, err := s.domainDao.FindByDomain(domain)
	if err != nil {
		// 如果查找域名配置失败，默认不透传参数，直接返回目标URL
		return targetURL, nil
	}

	// 构建最终的跳转URL
	finalURL := targetURL

	// 如果域名配置允许透传GET参数且存在查询参数
	if domainInfo.PassQueryParams && queryString != "" {
		// 解析目标URL
		origURL, err := url.Parse(targetURL)
		if err != nil {
			// 如果解析失败，返回目标URL
			return targetURL, nil
		}

		// 解析查询参数
		query := origURL.Query()

		// 解析新的查询参数并合并
		newQuery, err := url.ParseQuery(queryString)
		if err == nil {
			for key, values := range newQuery {
				for _, value := range values {
					query.Add(key, value)
				}
			}
		}

		// 重新构建URL
		origURL.RawQuery = query.Encode()
		finalURL = origURL.String()
	}

	return finalURL, nil
}

// findShortLinkByCode 通过短代码查找短链
func (s *ShortLinkService) findShortLinkByCode(domain, shortCode string) (*model.ShortLink, error) {
	// 策略1：直接通过custom_code字段查找（适用于自定义代码和新的分布式发号器代码）
	var shortLink model.ShortLink
	link := model.ShortLink{}
	err := s.helper.GetDatabase().Table(link.TableName()).Where("domain = ? AND short_code = ? AND deleted_at IS NULL", domain, shortCode).First(&shortLink).Error
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &shortLink, nil
}

// GetShortLinkStatistics 获取短网址统计信息
func (s *ShortLinkService) GetShortLinkStatistics(id uint64, days int) (*dto.ShortLinkStatisticResponse, error) {
	return s.GetShortLinkStatisticsInWorkspace(id, days, 1)
}

func (s *ShortLinkService) GetShortLinkStatisticsInWorkspace(id uint64, days int, workspaceID uint64) (*dto.ShortLinkStatisticResponse, error) {
	shortLink, err := s.shortLinkDao.FindByID(id)
	if workspaceID > 0 {
		shortLink, err = s.shortLinkDao.FindByIDInWorkspace(id, workspaceID)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("短网址不存在")
		}
		return nil, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	// 获取各时间段的点击数
	todayClicks, _ := s.shortLinkDao.GetClickCountByDateRange(id, today, today.AddDate(0, 0, 1))
	weekClicks, _ := s.shortLinkDao.GetClickCountByDateRange(id, weekAgo, now)
	monthClicks, _ := s.shortLinkDao.GetClickCountByDateRange(id, monthAgo, now)

	// 获取每日统计
	dailyStats, _ := s.shortLinkDao.GetDailyClickCount(id, days)
	dailyStatistics := make([]dto.ClickStatisticResponse, 0)

	// 填充每日统计数据
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		count := dailyStats[date]
		dailyStatistics = append(dailyStatistics, dto.ClickStatisticResponse{
			Date:       date,
			ClickCount: count,
		})
	}

	return &dto.ShortLinkStatisticResponse{
		TotalClicks:     shortLink.ClickCount,
		TodayClicks:     todayClicks,
		WeekClicks:      weekClicks,
		MonthClicks:     monthClicks,
		DailyStatistics: dailyStatistics,
	}, nil
}

// BatchCreateShortLinks 批量创建短网址
func (s *ShortLinkService) BatchCreateShortLinks(req *dto.BatchCreateShortLinkRequest, creatorIP string, helper interfaces.HelperInterface) (*dto.BatchCreateShortLinkResponse, error) {
	return s.BatchCreateShortLinksInWorkspace(req, creatorIP, helper, 1, 0)
}

func (s *ShortLinkService) BatchCreateShortLinksInWorkspace(req *dto.BatchCreateShortLinkRequest, creatorIP string, helper interfaces.HelperInterface, workspaceID, userID uint64) (*dto.BatchCreateShortLinkResponse, error) {
	success := make([]dto.ShortLinkResponse, 0)
	failed := make([]dto.BatchFailedItem, 0)

	domain := req.Domain
	if domain == "" {
		domain = s.helper.GetEnv().GetString("shortlink_domain", "http://localhost:8080")
	}

	for _, originalURL := range req.URLs {
		createReq := &dto.CreateShortLinkRequest{
			OriginalURL: originalURL,
			Domain:      domain,
		}

		response, err := s.CreateShortLinkInWorkspace(createReq, creatorIP, workspaceID, userID)
		if err != nil {
			failed = append(failed, dto.BatchFailedItem{
				URL:   originalURL,
				Error: err.Error(),
			})
		} else {
			success = append(success, *response)
		}
	}

	return &dto.BatchCreateShortLinkResponse{
		Success: success,
		Failed:  failed,
	}, nil
}

// 私有方法

// validateDomain 验证域名
func (s *ShortLinkService) validateDomain(domain string) error {
	return domain_validate.ValidateDomain(domain)
}

// cacheShortLink 缓存短网址到Redis
func (s *ShortLinkService) cacheShortLink(shortLink *model.ShortLink) {
	key := fmt.Sprintf("shortlink:%s:%s", shortLink.Domain, shortLink.GetShortCode())

	err := s.helper.GetCache().Set(s.context, key, &shortLink, 24*time.Hour)
	if err != nil {
		s.helper.GetLogger().Error(err.Error())
	}
}

// getShortLinkFromCache 从Redis缓存获取短网址
func (s *ShortLinkService) getShortLinkFromCache(domain, shortCode string) (*model.ShortLink, error) {
	key := fmt.Sprintf("shortlink:%s:%s", domain, shortCode)

	var shortLink model.ShortLink

	err := s.helper.GetCache().Get(s.context, key, &shortLink)

	return &shortLink, err
}

// removeCacheShortLink 从Redis缓存删除短网址
func (s *ShortLinkService) removeCacheShortLink(domain, shortCode string) {
	key := fmt.Sprintf("shortlink:%s:%s", domain, shortCode)
	err := s.helper.GetCache().Del(s.context, key)
	if err != nil {
		s.helper.GetLogger().Error(err.Error())
	}
}

// recordClickStatistic 记录点击统计
func (s *ShortLinkService) recordClickStatistic(shortLink *model.ShortLink, clientIP, userAgent, referer string, queryParams string) {
	region := s.helper.GetIPRegion().Lookup(clientIP)
	metadata := parseTrafficMetadata(userAgent)
	statistic := &model.ClickStatistic{
		WorkspaceID: shortLink.WorkspaceID,
		CampaignID:  shortLink.CampaignID,
		ShortLinkID: shortLink.ID,
		IP:          domain_validate.TruncateString(clientIP, 45),
		UserAgent:   domain_validate.TruncateString(userAgent, 1024),
		Referer:     domain_validate.TruncateString(referer, 2048),
		QueryParams: domain_validate.TruncateString(queryParams, 2048), // 截断过长的参数
		UTMSource:   domain_validate.TruncateString(shortLink.UTMSource, 255),
		UTMMedium:   domain_validate.TruncateString(shortLink.UTMMedium, 255),
		UTMCampaign: domain_validate.TruncateString(shortLink.UTMCampaign, 255),
		UTMTerm:     domain_validate.TruncateString(shortLink.UTMTerm, 255),
		UTMContent:  domain_validate.TruncateString(shortLink.UTMContent, 255),
		DeviceType:  domain_validate.TruncateString(metadata.DeviceType, 50),
		Browser:     domain_validate.TruncateString(metadata.Browser, 100),
		OS:          domain_validate.TruncateString(metadata.OS, 100),
		IsBot:       metadata.IsBot,
		BotName:     domain_validate.TruncateString(metadata.BotName, 100),
		Country:     domain_validate.TruncateString(region.Country, 100),
		Province:    domain_validate.TruncateString(region.Province, 100),
		City:        domain_validate.TruncateString(region.City, 100),
		ISP:         domain_validate.TruncateString(region.ISP, 100),
		ClickDate:   time.Now(),
	}

	s.clickStatisticDao.Create(statistic)
}

// incrementClickCount 增加点击次数
func (s *ShortLinkService) incrementClickCount(shortLinkID uint64) {
	s.shortLinkDao.IncrementClickCount(shortLinkID)
}

// modelToResponse 将模型转换为响应格式
func (s *ShortLinkService) modelToResponse(shortLink *model.ShortLink) *dto.ShortLinkResponse {
	tags, _ := s.tagDao.GetTagsByShortLinkID(shortLink.ID)
	tagResponses := make([]dto.TagResponse, 0, len(tags))
	for _, tag := range tags {
		tagResponses = append(tagResponses, dto.TagResponse{
			ID:          tag.ID,
			WorkspaceID: tag.WorkspaceID,
			Name:        tag.Name,
			Color:       tag.Color,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
		})
	}
	campaignName := ""
	if shortLink.Campaign != nil {
		campaignName = shortLink.Campaign.Name
	}
	return &dto.ShortLinkResponse{
		ID:           shortLink.ID,
		WorkspaceID:  shortLink.WorkspaceID,
		CampaignID:   shortLink.CampaignID,
		CampaignName: campaignName,
		Tags:         tagResponses,
		ShortCode:    shortLink.GetShortCode(),
		Domain:       shortLink.Domain,
		ShortURL:     shortLink.GetFullURL(),
		OriginalURL:  shortLink.OriginalURL,
		Title:        shortLink.Title,
		Description:  shortLink.Description,
		UTMSource:    shortLink.UTMSource,
		UTMMedium:    shortLink.UTMMedium,
		UTMCampaign:  shortLink.UTMCampaign,
		UTMTerm:      shortLink.UTMTerm,
		UTMContent:   shortLink.UTMContent,
		Notes:        shortLink.Notes,
		ExpireAt:     shortLink.ExpireAt,
		IsActive:     shortLink.IsActive,
		ClickCount:   shortLink.ClickCount,
		CreatedBy:    shortLink.CreatedBy,
		UpdatedBy:    shortLink.UpdatedBy,
		CreatedAt:    shortLink.CreatedAt,
		UpdatedAt:    shortLink.UpdatedAt,
	}
}

func (s *ShortLinkService) validateCampaignAndTags(workspaceID uint64, campaignID *uint64, tagIDs []uint64) error {
	if campaignID != nil && *campaignID > 0 {
		if _, err := s.campaignDao.FindByID(*campaignID, workspaceID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("活动不存在")
			}
			return err
		}
	}
	if len(tagIDs) > 0 {
		tags, err := s.tagDao.FindMany(tagIDs, workspaceID)
		if err != nil {
			return err
		}
		if len(tags) != len(tagIDs) {
			return errors.New("标签不存在")
		}
	}
	return nil
}

func mergeUTMToURL(rawURL, source, medium, campaign, term, content string) (string, error) {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}
	query := parsedURL.Query()
	if source != "" {
		query.Set("utm_source", source)
	}
	if medium != "" {
		query.Set("utm_medium", medium)
	}
	if campaign != "" {
		query.Set("utm_campaign", campaign)
	}
	if term != "" {
		query.Set("utm_term", term)
	}
	if content != "" {
		query.Set("utm_content", content)
	}
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}
