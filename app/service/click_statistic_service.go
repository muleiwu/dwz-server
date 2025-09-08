package service

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type ClickStatisticService struct {
	clickStatisticDao *dao.ClickStatisticDao
	shortLinkDao      *dao.ShortLinkDao
}

func NewClickStatisticService(helper interfaces.GetHelperInterface) *ClickStatisticService {
	return &ClickStatisticService{
		clickStatisticDao: dao.NewClickStatisticDao(helper),
		shortLinkDao:      dao.NewShortLinkDao(helper),
	}
}

// GetClickStatisticList 获取点击统计列表
func (s *ClickStatisticService) GetClickStatisticList(req *dto.ClickStatisticListRequest) (*dto.ClickStatisticListResponse, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	statistics, total, err := s.clickStatisticDao.List(req)
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

// GetClickStatisticAnalysisByDateRange 按日期范围获取点击统计分析
func (s *ClickStatisticService) GetClickStatisticAnalysisByDateRange(shortLinkID uint64, startDate, endDate time.Time) (*dto.ClickStatisticAnalysisResponse, error) {
	return s.clickStatisticDao.GetAnalysis(shortLinkID, startDate, endDate)
}

// modelToResponse 将模型转换为响应格式
func (s *ClickStatisticService) modelToResponse(statistic *model.ClickStatistic) *dto.ClickStatisticDetailResponse {
	response := &dto.ClickStatisticDetailResponse{
		ID:          statistic.ID,
		ShortLinkID: statistic.ShortLinkID,
		IP:          statistic.IP,
		UserAgent:   statistic.UserAgent,
		Referer:     statistic.Referer,
		QueryParams: statistic.QueryParams,
		Country:     statistic.Country,
		City:        statistic.City,
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
