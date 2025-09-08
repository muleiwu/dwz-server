package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"gorm.io/gorm"
)

type ABTestService struct {
	abTestDao    *dao.ABTestDao
	shortLinkDao *dao.ShortLinkDao
}

func NewABTestService(helper interfaces.GetHelperInterface) *ABTestService {
	return &ABTestService{
		abTestDao:    &dao.ABTestDao{Helper: helper},
		shortLinkDao: &dao.ShortLinkDao{Helper: helper},
	}
}

// CreateABTest 创建AB测试
func (s *ABTestService) CreateABTest(req *dto.CreateABTestRequest) (*dto.ABTestResponse, error) {
	// 验证短链接是否存在
	_, err := s.shortLinkDao.FindByID(req.ShortLinkID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("短链接不存在")
		}
		return nil, err
	}

	// 检查是否已有正在运行的AB测试
	existingTest, err := s.abTestDao.FindActiveABTestByShortLinkID(req.ShortLinkID)
	if err == nil && existingTest != nil {
		return nil, errors.New("该短链接已有正在运行的AB测试")
	}

	// 验证变体权重
	if err := s.validateVariantWeights(req.Variants, req.TrafficSplit); err != nil {
		return nil, err
	}

	// 创建AB测试
	abTest := &model.ABTest{
		ShortLinkID:  req.ShortLinkID,
		Name:         req.Name,
		Description:  req.Description,
		Status:       "draft",
		TrafficSplit: req.TrafficSplit,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		IsActive:     true,
	}

	if err := s.abTestDao.CreateABTest(abTest); err != nil {
		return nil, err
	}

	// 创建变体
	for i, variantReq := range req.Variants {
		weight := variantReq.Weight
		// 如果是平均分配模式，自动计算权重
		if req.TrafficSplit == "equal" {
			weight = 100 / len(req.Variants)
			// 处理余数，最后一个变体获得剩余权重
			if i == len(req.Variants)-1 {
				weight = 100 - (weight * (len(req.Variants) - 1))
			}
		}

		variant := &model.ABTestVariant{
			ABTestID:    abTest.ID,
			Name:        variantReq.Name,
			TargetURL:   variantReq.TargetURL,
			Weight:      weight,
			IsControl:   variantReq.IsControl,
			Description: variantReq.Description,
			IsActive:    true,
		}

		if err := s.abTestDao.CreateABTestVariant(variant); err != nil {
			return nil, err
		}

		abTest.Variants = append(abTest.Variants, *variant)
	}

	return s.modelToResponse(abTest), nil
}

// GetABTest 获取AB测试详情
func (s *ABTestService) GetABTest(id uint64) (*dto.ABTestResponse, error) {
	abTest, err := s.abTestDao.FindABTestByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("AB测试不存在")
		}
		return nil, err
	}

	return s.modelToResponse(abTest), nil
}

// UpdateABTest 更新AB测试
func (s *ABTestService) UpdateABTest(id uint64, req *dto.UpdateABTestRequest) (*dto.ABTestResponse, error) {
	abTest, err := s.abTestDao.FindABTestByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("AB测试不存在")
		}
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		abTest.Name = req.Name
	}
	if req.Description != "" {
		abTest.Description = req.Description
	}
	if req.Status != "" {
		// 验证状态转换
		if err := s.validateStatusTransition(abTest.Status, req.Status); err != nil {
			return nil, err
		}
		abTest.Status = req.Status
	}
	if req.TrafficSplit != "" {
		abTest.TrafficSplit = req.TrafficSplit
	}
	if req.StartTime != nil {
		abTest.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		abTest.EndTime = req.EndTime
	}
	if req.IsActive != nil {
		abTest.IsActive = *req.IsActive
	}

	if err := s.abTestDao.UpdateABTest(abTest); err != nil {
		return nil, err
	}

	return s.modelToResponse(abTest), nil
}

// DeleteABTest 删除AB测试
func (s *ABTestService) DeleteABTest(id uint64) error {
	abTest, err := s.abTestDao.FindABTestByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("AB测试不存在")
		}
		return err
	}

	// 只有草稿状态的测试可以删除
	if abTest.Status != "draft" {
		return errors.New("只有草稿状态的AB测试可以删除")
	}

	return s.abTestDao.DeleteABTest(id)
}

// GetABTestList 获取AB测试列表
func (s *ABTestService) GetABTestList(req *dto.ABTestListRequest) (*dto.ABTestListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	abTests, total, err := s.abTestDao.ListABTests(offset, req.PageSize, req.ShortLinkID, req.Status)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ABTestResponse, 0, len(abTests))
	for _, abTest := range abTests {
		responses = append(responses, *s.modelToResponse(&abTest))
	}

	return &dto.ABTestListResponse{
		List:  responses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// StartABTest 启动AB测试
func (s *ABTestService) StartABTest(id uint64, req *dto.StartABTestRequest) (*dto.ABTestResponse, error) {
	abTest, err := s.abTestDao.FindABTestByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("AB测试不存在")
		}
		return nil, err
	}

	if abTest.Status != "draft" {
		return nil, errors.New("只有草稿状态的AB测试可以启动")
	}

	// 检查是否有活跃的变体
	activeVariants := abTest.GetActiveVariants()
	if len(activeVariants) < 2 {
		return nil, errors.New("至少需要2个活跃的变体才能启动测试")
	}

	abTest.Status = "running"
	if req.StartTime != nil {
		abTest.StartTime = req.StartTime
	} else {
		now := time.Now()
		abTest.StartTime = &now
	}

	if err := s.abTestDao.UpdateABTest(abTest); err != nil {
		return nil, err
	}

	return s.modelToResponse(abTest), nil
}

// StopABTest 停止AB测试
func (s *ABTestService) StopABTest(id uint64, req *dto.StopABTestRequest) (*dto.ABTestResponse, error) {
	abTest, err := s.abTestDao.FindABTestByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("AB测试不存在")
		}
		return nil, err
	}

	if abTest.Status != "running" {
		return nil, errors.New("只有运行中的AB测试可以停止")
	}

	abTest.Status = "completed"
	if req.EndTime != nil {
		abTest.EndTime = req.EndTime
	} else {
		now := time.Now()
		abTest.EndTime = &now
	}

	if err := s.abTestDao.UpdateABTest(abTest); err != nil {
		return nil, err
	}

	return s.modelToResponse(abTest), nil
}

// GetABTestRedirectInfo 获取AB测试重定向信息
func (s *ABTestService) GetABTestRedirectInfo(shortLinkID uint64, userIP, userAgent string) (*dto.ABTestRedirectInfo, error) {
	// 查找正在运行的AB测试
	abTest, err := s.abTestDao.FindActiveABTestByShortLinkID(shortLinkID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有AB测试，返回nil
		}
		return nil, err
	}

	// 检查测试是否正在运行
	if !abTest.IsRunning() {
		return nil, nil
	}

	// 获取活跃的变体
	activeVariants := abTest.GetActiveVariants()
	if len(activeVariants) == 0 {
		return nil, nil
	}

	// 生成会话ID（基于IP和UserAgent）
	sessionID := s.generateSessionID(userIP, userAgent, abTest.ID)

	// 选择变体
	selectedVariant := s.selectVariant(activeVariants, sessionID, abTest.TrafficSplit)

	return &dto.ABTestRedirectInfo{
		ABTestID:    abTest.ID,
		VariantID:   selectedVariant.ID,
		TargetURL:   selectedVariant.TargetURL,
		VariantName: selectedVariant.Name,
		SessionID:   sessionID,
	}, nil
}

// RecordABTestClick 记录AB测试点击
func (s *ABTestService) RecordABTestClick(redirectInfo *dto.ABTestRedirectInfo, clientIP, userAgent, referer, queryParams string) error {
	// 检查会话是否已存在（防重复）
	exists, err := s.abTestDao.CheckSessionExists(redirectInfo.ABTestID, redirectInfo.VariantID, redirectInfo.SessionID)
	if err != nil {
		return err
	}
	if exists {
		return nil // 已存在，不重复记录
	}

	// 获取AB测试信息以获取ShortLinkID
	abTest, err := s.abTestDao.FindABTestByID(redirectInfo.ABTestID)
	if err != nil {
		return err
	}

	// 创建统计记录
	stat := &model.ABTestClickStatistic{
		ABTestID:    redirectInfo.ABTestID,
		VariantID:   redirectInfo.VariantID,
		ShortLinkID: abTest.ShortLinkID,
		IP:          clientIP,
		UserAgent:   userAgent,
		Referer:     referer,
		QueryParams: queryParams,
		SessionID:   redirectInfo.SessionID,
		ClickDate:   time.Now(),
	}

	return s.abTestDao.CreateABTestClickStatistic(stat)
}

// GetABTestStatistics 获取AB测试统计
func (s *ABTestService) GetABTestStatistics(id uint64, days int) (*dto.ABTestStatisticResponse, error) {
	abTest, err := s.abTestDao.FindABTestByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("AB测试不存在")
		}
		return nil, err
	}

	// 获取统计数据
	variantStats, err := s.abTestDao.GetABTestStatistics(id, days)
	if err != nil {
		return nil, err
	}

	// 获取每日统计
	dailyStats, err := s.abTestDao.GetDailyABTestStatistics(id, days)
	if err != nil {
		return nil, err
	}

	// 计算总点击数
	var totalClicks int64
	for _, count := range variantStats {
		totalClicks += count
	}

	// 构建变体统计
	variantStatsList := make([]dto.ABTestVariantStatResponse, 0, len(abTest.Variants))
	var winningVariant *dto.ABTestVariantResponse
	maxClicks := int64(0)

	for _, variant := range abTest.Variants {
		clickCount := variantStats[variant.ID]
		percentage := float64(0)
		if totalClicks > 0 {
			percentage = float64(clickCount) / float64(totalClicks) * 100
		}

		variantStat := dto.ABTestVariantStatResponse{
			Variant:        *s.variantModelToResponse(&variant),
			ClickCount:     clickCount,
			ConversionRate: percentage,
			Percentage:     percentage,
		}

		variantStatsList = append(variantStatsList, variantStat)

		// 找出表现最好的变体
		if clickCount > maxClicks {
			maxClicks = clickCount
			variantResp := s.variantModelToResponse(&variant)
			winningVariant = variantResp
		}
	}

	// 转换每日统计格式
	dailyStatsList := make([]dto.ABTestDailyStatResponse, 0, len(dailyStats))
	for _, dailyStat := range dailyStats {
		dailyStatsList = append(dailyStatsList, dto.ABTestDailyStatResponse{
			Date:     dailyStat["date"].(string),
			Variants: dailyStat["variants"].(map[uint64]int64),
		})
	}

	return &dto.ABTestStatisticResponse{
		ABTestID:       id,
		TotalClicks:    totalClicks,
		VariantStats:   variantStatsList,
		DailyStats:     dailyStatsList,
		ConversionRate: 0, // 可以根据需要计算实际转化率
		WinningVariant: winningVariant,
	}, nil
}

// 私有方法

// validateVariantWeights 验证变体权重
func (s *ABTestService) validateVariantWeights(variants []dto.CreateABTestVariantRequest, trafficSplit string) error {
	if len(variants) < 2 {
		return errors.New("至少需要2个变体")
	}

	if trafficSplit == "weighted" || trafficSplit == "custom" {
		totalWeight := 0
		for _, variant := range variants {
			if variant.Weight <= 0 || variant.Weight > 100 {
				return errors.New("变体权重必须在1-100之间")
			}
			totalWeight += variant.Weight
		}
		if totalWeight != 100 {
			return errors.New("所有变体权重之和必须等于100")
		}
	}

	return nil
}

// validateStatusTransition 验证状态转换
func (s *ABTestService) validateStatusTransition(from, to string) error {
	validTransitions := map[string][]string{
		"draft":     {"draft", "running"},
		"running":   {"paused", "running", "completed"},
		"paused":    {"running", "paused", "completed"},
		"completed": {}, // 完成状态不能转换
	}

	if allowedStates, exists := validTransitions[from]; exists {
		for _, state := range allowedStates {
			if state == to {
				return nil
			}
		}
	}

	return fmt.Errorf("不能从状态 %s 转换到 %s", from, to)
}

// generateSessionID 生成会话ID
func (s *ABTestService) generateSessionID(userIP, userAgent string, abTestID uint64) string {
	data := fmt.Sprintf("%s:%s:%d:%s", userIP, userAgent, abTestID, time.Now().Format("2006-01-02"))
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// selectVariant 选择变体
func (s *ABTestService) selectVariant(variants []model.ABTestVariant, sessionID, trafficSplit string) model.ABTestVariant {
	if len(variants) == 0 {
		return model.ABTestVariant{}
	}

	// 基于会话ID生成随机数种子，确保同一用户总是看到相同版本
	seed := int64(0)
	for _, char := range sessionID {
		seed += int64(char)
	}

	rng := rand.New(rand.NewSource(seed))

	switch trafficSplit {
	case "equal":
		// 平均分配
		index := rng.Intn(len(variants))
		return variants[index]
	case "weighted", "custom":
		// 按权重分配
		return s.selectVariantByWeight(variants, rng.Intn(100))
	default:
		// 默认平均分配
		index := rng.Intn(len(variants))
		return variants[index]
	}
}

// selectVariantByWeight 按权重选择变体
func (s *ABTestService) selectVariantByWeight(variants []model.ABTestVariant, random int) model.ABTestVariant {
	currentWeight := 0
	for _, variant := range variants {
		currentWeight += variant.Weight
		if random < currentWeight {
			return variant
		}
	}
	// 如果没有匹配到，返回第一个
	return variants[0]
}

// modelToResponse 转换模型到响应
func (s *ABTestService) modelToResponse(abTest *model.ABTest) *dto.ABTestResponse {
	variants := make([]dto.ABTestVariantResponse, 0, len(abTest.Variants))
	for _, variant := range abTest.Variants {
		variants = append(variants, *s.variantModelToResponse(&variant))
	}

	return &dto.ABTestResponse{
		ID:           abTest.ID,
		ShortLinkID:  abTest.ShortLinkID,
		Name:         abTest.Name,
		Description:  abTest.Description,
		Status:       abTest.Status,
		TrafficSplit: abTest.TrafficSplit,
		StartTime:    abTest.StartTime,
		EndTime:      abTest.EndTime,
		IsActive:     abTest.IsActive,
		CreatedAt:    abTest.CreatedAt,
		UpdatedAt:    abTest.UpdatedAt,
		Variants:     variants,
	}
}

// variantModelToResponse 转换变体模型到响应
func (s *ABTestService) variantModelToResponse(variant *model.ABTestVariant) *dto.ABTestVariantResponse {
	return &dto.ABTestVariantResponse{
		ID:          variant.ID,
		ABTestID:    variant.ABTestID,
		Name:        variant.Name,
		TargetURL:   variant.TargetURL,
		Weight:      variant.Weight,
		IsControl:   variant.IsControl,
		Description: variant.Description,
		IsActive:    variant.IsActive,
		CreatedAt:   variant.CreatedAt,
		UpdatedAt:   variant.UpdatedAt,
	}
}
