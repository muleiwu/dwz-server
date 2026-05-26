package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
)

type ClickStatisticService struct {
	helper            interfaces.HelperInterface
	clickStatisticDao *dao.ClickStatisticDao
	shortLinkDao      *dao.ShortLinkDao
}

func NewClickStatisticService(helper interfaces.HelperInterface) *ClickStatisticService {
	return &ClickStatisticService{
		helper:            helper,
		clickStatisticDao: dao.NewClickStatisticDao(helper),
		shortLinkDao:      dao.NewShortLinkDao(helper),
	}
}

const (
	clickStatisticAnalysisCachePrefix  = "click_statistics:analysis"
	clickStatisticAnalysisCacheVersion = "v1"
	clickStatisticAnalysisCacheTTL     = 5 * time.Minute
)

// GetClickStatisticList 获取点击统计列表
func (s *ClickStatisticService) GetClickStatisticList(req *dto.ClickStatisticListRequest) (*dto.ClickStatisticListResponse, error) {
	return s.GetClickStatisticListInWorkspace(req, 1)
}

func (s *ClickStatisticService) GetClickStatisticListInWorkspace(req *dto.ClickStatisticListRequest, workspaceID uint64) (*dto.ClickStatisticListResponse, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	statistics, total, err := s.clickStatisticDao.ListInWorkspace(workspaceID, req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	responses := make([]dto.ClickStatisticDetailResponse, 0, len(statistics))
	for _, statistic := range statistics {
		response := s.modelToResponse(&statistic)
		responses = append(responses, *response)
	}

	return &dto.ClickStatisticListResponse{
		List:  responses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// GetClickStatisticAnalysis 获取点击统计分析
func (s *ClickStatisticService) GetClickStatisticAnalysis(shortLinkID uint64, days int) (*dto.ClickStatisticAnalysisResponse, error) {
	// 计算时间范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	return s.clickStatisticDao.GetAnalysis(shortLinkID, startDate, endDate)
}

func (s *ClickStatisticService) GetClickStatisticAnalysisInWorkspace(workspaceID uint64, req *dto.ClickStatisticListRequest, days int) (*dto.ClickStatisticAnalysisResponse, error) {
	if req.StartDate.IsZero() && req.EndDate.IsZero() {
		req.StartDate, req.EndDate = defaultClickStatisticDateRange(days)
	}

	var cached dto.ClickStatisticAnalysisResponse
	cacheKey := s.analysisCacheKey("summary", workspaceID, req, "")
	if s.getCache(cacheKey, &cached) == nil {
		return &cached, nil
	}

	analysis, err := s.clickStatisticDao.GetAnalysisInWorkspace(workspaceID, req)
	if err != nil {
		return nil, err
	}
	s.setCache(cacheKey, analysis)
	return analysis, nil
}

func (s *ClickStatisticService) GetClickStatisticGeoAnalysisInWorkspace(workspaceID uint64, req *dto.ClickStatisticListRequest, level string, days int) (*dto.ClickStatisticGeoAnalysisResponse, error) {
	level, err := normalizeGeoAnalysisLevel(level)
	if err != nil {
		return nil, err
	}
	if req.StartDate.IsZero() && req.EndDate.IsZero() {
		req.StartDate, req.EndDate = defaultClickStatisticDateRange(days)
	}

	var cached dto.ClickStatisticGeoAnalysisResponse
	cacheKey := s.analysisCacheKey("geo", workspaceID, req, level)
	if s.getCache(cacheKey, &cached) == nil {
		return &cached, nil
	}

	analysis, err := s.clickStatisticDao.GetGeoAnalysisInWorkspace(workspaceID, req, level)
	if err != nil {
		return nil, err
	}
	s.setCache(cacheKey, analysis)
	return analysis, nil
}

func defaultClickStatisticDateRange(days int) (time.Time, time.Time) {
	if days < 1 || days > 365 {
		days = 7
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return today.AddDate(0, 0, -(days - 1)), today.AddDate(0, 0, 1)
}

func normalizeGeoAnalysisLevel(level string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "country":
		return "country", nil
	case "province":
		return "province", nil
	case "city":
		return "city", nil
	default:
		return "", errors.New("地理统计级别必须为 country、province 或 city")
	}
}

func (s *ClickStatisticService) analysisCacheKey(kind string, workspaceID uint64, req *dto.ClickStatisticListRequest, level string) string {
	isBot := ""
	if req.IsBot != nil {
		isBot = strconv.FormatBool(*req.IsBot)
	}
	raw := strings.Join([]string{
		clickStatisticAnalysisCacheVersion,
		kind,
		"workspace_id=" + strconv.FormatUint(workspaceID, 10),
		"level=" + level,
		"short_link_id=" + strconv.FormatUint(req.ShortLinkID, 10),
		"campaign_id=" + strconv.FormatUint(req.CampaignID, 10),
		"route_id=" + strconv.FormatUint(req.RouteID, 10),
		"tag_id=" + strconv.FormatUint(req.TagID, 10),
		"device_type=" + req.DeviceType,
		"is_bot=" + isBot,
		"ip=" + req.IP,
		"country=" + req.Country,
		"province=" + req.Province,
		"city=" + req.City,
		"isp=" + req.ISP,
		"start_date=" + req.StartDate.Format(time.RFC3339Nano),
		"end_date=" + req.EndDate.Format(time.RFC3339Nano),
	}, "|")
	sum := sha1.Sum([]byte(raw))
	return fmt.Sprintf("%s:%s:%s", clickStatisticAnalysisCachePrefix, kind, hex.EncodeToString(sum[:]))
}

func (s *ClickStatisticService) getCache(key string, dest any) error {
	cache := s.helper.GetCache()
	if cache == nil {
		return errors.New("cache unavailable")
	}
	return cache.Get(context.Background(), key, dest)
}

func (s *ClickStatisticService) setCache(key string, value any) {
	cache := s.helper.GetCache()
	if cache == nil {
		return
	}
	if err := cache.Set(context.Background(), key, value, clickStatisticAnalysisCacheTTL); err != nil && s.helper.GetLogger() != nil {
		s.helper.GetLogger().Warn("写入点击统计分析缓存失败: " + err.Error())
	}
}

// GetClickStatisticAnalysisByDateRange 按日期范围获取点击统计分析
func (s *ClickStatisticService) GetClickStatisticAnalysisByDateRange(shortLinkID uint64, startDate, endDate time.Time) (*dto.ClickStatisticAnalysisResponse, error) {
	return s.clickStatisticDao.GetAnalysis(shortLinkID, startDate, endDate)
}

func (s *ClickStatisticService) ExportCSV(workspaceID uint64, req *dto.ClickStatisticListRequest) ([]byte, error) {
	const maxRows = 50000
	statistics, err := s.clickStatisticDao.ExportInWorkspace(workspaceID, req, maxRows+1)
	if err != nil {
		return nil, err
	}
	if len(statistics) > maxRows {
		return nil, errors.New("导出数据超过 50000 行，请缩小筛选范围")
	}

	buffer := &bytes.Buffer{}
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(buffer)
	header := []string{
		"id", "workspace_id", "campaign_id", "route_id", "route_name", "short_link_id", "ip", "user_agent", "referer", "query_params",
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"device_type", "browser", "os", "is_bot", "bot_name",
		"country", "province", "city", "isp", "click_date", "created_at",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}
	for _, stat := range statistics {
		campaignID := ""
		if stat.CampaignID != nil {
			campaignID = strconv.FormatUint(*stat.CampaignID, 10)
		}
		routeID := ""
		if stat.RouteID != nil {
			routeID = strconv.FormatUint(*stat.RouteID, 10)
		}
		record := []string{
			strconv.FormatUint(stat.ID, 10),
			strconv.FormatUint(stat.WorkspaceID, 10),
			campaignID,
			routeID,
			stat.RouteName,
			strconv.FormatUint(stat.ShortLinkID, 10),
			stat.IP,
			stat.UserAgent,
			stat.Referer,
			stat.QueryParams,
			stat.UTMSource,
			stat.UTMMedium,
			stat.UTMCampaign,
			stat.UTMTerm,
			stat.UTMContent,
			stat.DeviceType,
			stat.Browser,
			stat.OS,
			strconv.FormatBool(stat.IsBot),
			stat.BotName,
			stat.Country,
			stat.Province,
			stat.City,
			stat.ISP,
			stat.ClickDate.Format(time.RFC3339),
			stat.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// modelToResponse 将模型转换为响应格式
func (s *ClickStatisticService) modelToResponse(statistic *model.ClickStatistic) *dto.ClickStatisticDetailResponse {
	response := &dto.ClickStatisticDetailResponse{
		ID:          statistic.ID,
		WorkspaceID: statistic.WorkspaceID,
		CampaignID:  statistic.CampaignID,
		RouteID:     statistic.RouteID,
		RouteName:   statistic.RouteName,
		ShortLinkID: statistic.ShortLinkID,
		IP:          statistic.IP,
		UserAgent:   statistic.UserAgent,
		Referer:     statistic.Referer,
		QueryParams: statistic.QueryParams,
		UTMSource:   statistic.UTMSource,
		UTMMedium:   statistic.UTMMedium,
		UTMCampaign: statistic.UTMCampaign,
		UTMTerm:     statistic.UTMTerm,
		UTMContent:  statistic.UTMContent,
		DeviceType:  statistic.DeviceType,
		Browser:     statistic.Browser,
		OS:          statistic.OS,
		IsBot:       statistic.IsBot,
		BotName:     statistic.BotName,
		Country:     statistic.Country,
		Province:    statistic.Province,
		City:        statistic.City,
		ISP:         statistic.ISP,
		ClickDate:   statistic.ClickDate,
		CreatedAt:   statistic.CreatedAt,
	}

	// 获取短链接信息
	if shortLink, err := s.shortLinkDao.FindByID(statistic.ShortLinkID); err == nil {
		response.ShortCode = shortLink.ShortCode
		response.Domain = shortLink.Domain
		response.OriginalURL = shortLink.OriginalURL
	}

	return response
}
