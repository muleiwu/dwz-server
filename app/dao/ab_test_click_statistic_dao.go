package dao

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type ABTestClickStatisticDao struct {
	Helper interfaces.GetHelperInterface
}

// Create 创建AB测试点击统计记录
func (d *ABTestClickStatisticDao) Create(statistic *model.ABTestClickStatistic) error {
	return d.Helper.GetDatabase().Create(statistic).Error
}

// List 获取AB测试点击统计列表
func (d *ABTestClickStatisticDao) List(req *dto.ABTestClickStatisticListRequest) ([]model.ABTestClickStatistic, int64, error) {
	var statistics []model.ABTestClickStatistic
	var total int64

	query := d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{}).
		Preload("ABTest").
		Preload("Variant").
		Preload("ShortLink")

	// 条件筛选
	if req.ABTestID > 0 {
		query = query.Where("ab_test_id = ?", req.ABTestID)
	}

	if req.VariantID > 0 {
		query = query.Where("variant_id = ?", req.VariantID)
	}

	if req.ShortLinkID > 0 {
		query = query.Where("short_link_id = ?", req.ShortLinkID)
	}

	if req.IP != "" {
		query = query.Where("ip = ?", req.IP)
	}

	if req.Country != "" {
		query = query.Where("country = ?", req.Country)
	}

	if req.City != "" {
		query = query.Where("city = ?", req.City)
	}

	if !req.StartDate.IsZero() {
		query = query.Where("click_date >= ?", req.StartDate)
	}

	if !req.EndDate.IsZero() {
		query = query.Where("click_date <= ?", req.EndDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取列表
	offset := (req.Page - 1) * req.PageSize
	err := query.Order("click_date DESC").Offset(offset).Limit(req.PageSize).Find(&statistics).Error
	return statistics, total, err
}

// GetAnalysis 获取AB测试点击统计分析数据
func (d *ABTestClickStatisticDao) GetAnalysis(abTestID uint64, startDate, endDate time.Time) (*dto.ABTestClickStatisticAnalysisResponse, error) {
	query := d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{})

	if abTestID > 0 {
		query = query.Where("ab_test_id = ?", abTestID)
	}

	if !startDate.IsZero() {
		query = query.Where("click_date >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("click_date <= ?", endDate)
	}

	analysis := &dto.ABTestClickStatisticAnalysisResponse{}

	// 总点击数
	query.Count(&analysis.TotalClicks)

	// 独立IP数
	query.Distinct("ip").Count(&analysis.UniqueIPs)

	// 独立会话数
	query.Distinct("session_id").Count(&analysis.UniqueSessions)

	// 版本统计
	var variantStats []dto.ABTestVariantStatistic
	d.Helper.GetDatabase().Table("ab_test_click_statistics").
		Select(`variant_id, 
				(SELECT name FROM ab_test_variants WHERE id = variant_id) as variant_name,
				(SELECT target_url FROM ab_test_variants WHERE id = variant_id) as target_url,
				COUNT(*) as click_count,
				COUNT(DISTINCT session_id) as unique_clicks`).
		Where("ab_test_id = ?", abTestID).
		Group("variant_id").
		Find(&variantStats)

	// 计算流量占比
	if analysis.TotalClicks > 0 {
		for i := range variantStats {
			variantStats[i].TrafficPercent = float64(variantStats[i].ClickCount) / float64(analysis.TotalClicks) * 100
		}
	}
	analysis.VariantStats = variantStats

	// 热门国家统计
	var countryStats []dto.CountryStatistic
	d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{}).
		Select("country, COUNT(*) as count").
		Where("ab_test_id = ? AND country != ''", abTestID).
		Group("country").
		Order("count DESC").
		Limit(10).
		Find(&countryStats)
	analysis.TopCountries = countryStats

	// 热门城市统计
	var cityStats []dto.CityStatistic
	d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{}).
		Select("city, COUNT(*) as count").
		Where("ab_test_id = ? AND city != ''", abTestID).
		Group("city").
		Order("count DESC").
		Limit(10).
		Find(&cityStats)
	analysis.TopCities = cityStats

	// 热门来源统计
	var refererStats []dto.RefererStatistic
	d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{}).
		Select("referer, COUNT(*) as count").
		Where("ab_test_id = ? AND referer != ''", abTestID).
		Group("referer").
		Order("count DESC").
		Limit(10).
		Find(&refererStats)
	analysis.TopReferers = refererStats

	// 小时统计
	var hourlyStats []dto.HourlyStatistic
	d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{}).
		Select("HOUR(click_date) as hour, COUNT(*) as count").
		Where("ab_test_id = ?", abTestID).
		Group("HOUR(click_date)").
		Order("hour").
		Find(&hourlyStats)
	analysis.HourlyStats = hourlyStats

	// 日统计
	var dailyStats []dto.DailyStatistic
	d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{}).
		Select("DATE(click_date) as date, COUNT(*) as count").
		Where("ab_test_id = ?", abTestID).
		Group("DATE(click_date)").
		Order("date").
		Find(&dailyStats)
	analysis.DailyStats = dailyStats

	// 转化率统计（按版本）
	analysis.ConversionRate = make(map[string]dto.ConversionRateStats)
	for _, variant := range variantStats {
		// 这里可以根据需要计算具体的转化率
		// 目前先用点击数作为基础数据
		analysis.ConversionRate[variant.VariantName] = dto.ConversionRateStats{
			Impressions:    variant.ClickCount, // 展示数可以从其他地方获取
			Clicks:         variant.ClickCount,
			ConversionRate: 100.0, // 这里需要根据实际业务逻辑计算
		}
	}

	return analysis, nil
}

// GetVariantStatistics 获取版本统计
func (d *ABTestClickStatisticDao) GetVariantStatistics(abTestID uint64, startDate, endDate time.Time) ([]dto.ABTestVariantStatistic, error) {
	var stats []dto.ABTestVariantStatistic

	query := d.Helper.GetDatabase().Table("ab_test_click_statistics").
		Select(`variant_id, 
				(SELECT name FROM ab_test_variants WHERE id = variant_id) as variant_name,
				(SELECT target_url FROM ab_test_variants WHERE id = variant_id) as target_url,
				COUNT(*) as click_count,
				COUNT(DISTINCT session_id) as unique_clicks`).
		Where("ab_test_id = ?", abTestID)

	if !startDate.IsZero() {
		query = query.Where("click_date >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("click_date <= ?", endDate)
	}

	err := query.Group("variant_id").Find(&stats).Error

	// 计算流量占比
	var totalClicks int64
	for _, stat := range stats {
		totalClicks += stat.ClickCount
	}

	if totalClicks > 0 {
		for i := range stats {
			stats[i].TrafficPercent = float64(stats[i].ClickCount) / float64(totalClicks) * 100
		}
	}

	return stats, err
}

// GetCountryStatistics 获取国家统计
func (d *ABTestClickStatisticDao) GetCountryStatistics(abTestID uint64, startDate, endDate time.Time, limit int) ([]dto.CountryStatistic, error) {
	var stats []dto.CountryStatistic
	query := d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{}).
		Select("country, COUNT(*) as count").
		Where("country != ''")

	if abTestID > 0 {
		query = query.Where("ab_test_id = ?", abTestID)
	}

	if !startDate.IsZero() {
		query = query.Where("click_date >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("click_date <= ?", endDate)
	}

	err := query.Group("country").Order("count DESC").Limit(limit).Find(&stats).Error
	return stats, err
}

// GetTotalCount 获取总点击数
func (d *ABTestClickStatisticDao) GetTotalCount(abTestID uint64, startDate, endDate time.Time) (int64, error) {
	var count int64
	query := d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{})

	if abTestID > 0 {
		query = query.Where("ab_test_id = ?", abTestID)
	}

	if !startDate.IsZero() {
		query = query.Where("click_date >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("click_date <= ?", endDate)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetUniqueIPCount 获取独立IP数
func (d *ABTestClickStatisticDao) GetUniqueIPCount(abTestID uint64, startDate, endDate time.Time) (int64, error) {
	var count int64
	query := d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{})

	if abTestID > 0 {
		query = query.Where("ab_test_id = ?", abTestID)
	}

	if !startDate.IsZero() {
		query = query.Where("click_date >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("click_date <= ?", endDate)
	}

	err := query.Distinct("ip").Count(&count).Error
	return count, err
}

// GetUniqueSessionCount 获取独立会话数
func (d *ABTestClickStatisticDao) GetUniqueSessionCount(abTestID uint64, startDate, endDate time.Time) (int64, error) {
	var count int64
	query := d.Helper.GetDatabase().Model(&model.ABTestClickStatistic{})

	if abTestID > 0 {
		query = query.Where("ab_test_id = ?", abTestID)
	}

	if !startDate.IsZero() {
		query = query.Where("click_date >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("click_date <= ?", endDate)
	}

	err := query.Distinct("session_id").Count(&count).Error
	return count, err
}
