package service

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type ABTestClickStatisticService struct {
	abTestClickStatisticDao *dao.ABTestClickStatisticDao
	shortLinkDao            *dao.ShortLinkDao
	abTestDao               *dao.ABTestDao
}

func NewABTestClickStatisticService(helper interfaces.GetHelperInterface) *ABTestClickStatisticService {
	return &ABTestClickStatisticService{
		abTestClickStatisticDao: &dao.ABTestClickStatisticDao{Helper: helper},
		shortLinkDao:            &dao.ShortLinkDao{Helper: helper},
		abTestDao:               &dao.ABTestDao{Helper: helper},
	}
}

// GetABTestClickStatisticList 获取AB测试点击统计列表
func (s *ABTestClickStatisticService) GetABTestClickStatisticList(req *dto.ABTestClickStatisticListRequest) (*dto.ABTestClickStatisticListResponse, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	statistics, total, err := s.abTestClickStatisticDao.List(req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	responses := make([]dto.ABTestClickStatisticDetailResponse, 0, len(statistics))
	for _, statistic := range statistics {
		response := s.modelToResponse(&statistic)
		responses = append(responses, *response)
	}

	return &dto.ABTestClickStatisticListResponse{
		List:  responses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// GetABTestClickStatisticAnalysis 获取AB测试点击统计分析
func (s *ABTestClickStatisticService) GetABTestClickStatisticAnalysis(abTestID uint64, days int) (*dto.ABTestClickStatisticAnalysisResponse, error) {
	// 计算时间范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	return s.abTestClickStatisticDao.GetAnalysis(abTestID, startDate, endDate)
}

// GetABTestClickStatisticAnalysisByDateRange 按日期范围获取AB测试点击统计分析
func (s *ABTestClickStatisticService) GetABTestClickStatisticAnalysisByDateRange(abTestID uint64, startDate, endDate time.Time) (*dto.ABTestClickStatisticAnalysisResponse, error) {
	return s.abTestClickStatisticDao.GetAnalysis(abTestID, startDate, endDate)
}

// GetVariantStatistics 获取版本统计
func (s *ABTestClickStatisticService) GetVariantStatistics(abTestID uint64, days int) ([]dto.ABTestVariantStatistic, error) {
	// 计算时间范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	return s.abTestClickStatisticDao.GetVariantStatistics(abTestID, startDate, endDate)
}

// modelToResponse 将模型转换为响应格式
func (s *ABTestClickStatisticService) modelToResponse(statistic *model.ABTestClickStatistic) *dto.ABTestClickStatisticDetailResponse {
	response := &dto.ABTestClickStatisticDetailResponse{
		ID:          statistic.ID,
		ABTestID:    statistic.ABTestID,
		VariantID:   statistic.VariantID,
		ShortLinkID: statistic.ShortLinkID,
		IP:          statistic.IP,
		UserAgent:   statistic.UserAgent,
		Referer:     statistic.Referer,
		QueryParams: statistic.QueryParams,
		Country:     statistic.Country,
		City:        statistic.City,
		SessionID:   statistic.SessionID,
		ClickDate:   statistic.ClickDate,
		CreatedAt:   statistic.CreatedAt,
	}

	// 获取AB测试信息
	if statistic.ABTest.ID > 0 {
		response.ABTestName = statistic.ABTest.Name
	}

	// 获取版本信息
	if statistic.Variant.ID > 0 {
		response.VariantName = statistic.Variant.Name
		response.TargetURL = statistic.Variant.TargetURL
	}

	// 获取短链接信息
	if statistic.ShortLink.ID > 0 {
		response.ShortCode = statistic.ShortLink.ShortCode
		response.Domain = statistic.ShortLink.Domain
	} else if shortLink, err := s.shortLinkDao.FindByID(statistic.ShortLinkID); err == nil {
		response.ShortCode = shortLink.ShortCode
		response.Domain = shortLink.Domain
	}

	return response
}
