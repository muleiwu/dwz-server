package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
)

type ClickStatisticService struct {
	clickStatisticDao *dao.ClickStatisticDao
	shortLinkDao      *dao.ShortLinkDao
}

func NewClickStatisticService(helper interfaces.HelperInterface) *ClickStatisticService {
	return &ClickStatisticService{
		clickStatisticDao: dao.NewClickStatisticDao(helper),
		shortLinkDao:      dao.NewShortLinkDao(helper),
	}
}

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
		req.EndDate = time.Now()
		req.StartDate = req.EndDate.AddDate(0, 0, -days)
	}
	return s.clickStatisticDao.GetAnalysisInWorkspace(workspaceID, req)
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
		"id", "workspace_id", "campaign_id", "short_link_id", "ip", "user_agent", "referer", "query_params",
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
		record := []string{
			strconv.FormatUint(stat.ID, 10),
			strconv.FormatUint(stat.WorkspaceID, 10),
			campaignID,
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
