package dao

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

// ClickStatisticDao 点击统计DAO
type ClickStatisticDao struct {
	helper interfaces.GetHelperInterface
}

func NewClickStatisticDao(helper interfaces.GetHelperInterface) *ClickStatisticDao {
	return &ClickStatisticDao{helper: helper}
}

// Create 创建点击统计记录
func (d *ClickStatisticDao) Create(statistic *model.ClickStatistic) error {
	return d.helper.GetDatabase().Create(statistic).Error
}

// List 获取点击统计列表
func (d *ClickStatisticDao) List(req *dto.ClickStatisticListRequest) ([]model.ClickStatistic, int64, error) {
	var statistics []model.ClickStatistic
	var total int64

	query := d.helper.GetDatabase().Model(&model.ClickStatistic{})

	// 条件筛选
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

// GetAnalysis 获取点击统计分析数据
func (d *ClickStatisticDao) GetAnalysis(shortLinkID uint64, startDate, endDate time.Time) (*dto.ClickStatisticAnalysisResponse, error) {
	query := d.helper.GetDatabase().Model(&model.ClickStatistic{})

	if shortLinkID > 0 {
		query = query.Where("short_link_id = ?", shortLinkID)
	}

	if !startDate.IsZero() {
		query = query.Where("click_date >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("click_date <= ?", endDate)
	}

	analysis := &dto.ClickStatisticAnalysisResponse{}

	// 总点击数
	var totalClicks int64
	query.Count(&totalClicks)
	analysis.TotalClicks = totalClicks

	// 独立IP数
	var uniqueIPs int64
	query.Distinct("ip").Count(&uniqueIPs)
	analysis.UniqueIPs = uniqueIPs

	// 构建基本查询条件
	whereConditions := []interface{}{}
	whereSQL := "1=1"

	if shortLinkID > 0 {
		whereSQL += " AND short_link_id = ?"
		whereConditions = append(whereConditions, shortLinkID)
	}

	if !startDate.IsZero() {
		whereSQL += " AND click_date >= ?"
		whereConditions = append(whereConditions, startDate)
	}

	if !endDate.IsZero() {
		whereSQL += " AND click_date <= ?"
		whereConditions = append(whereConditions, endDate)
	}

	// 热门国家统计
	var countryStats []dto.CountryStatistic
	d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Select("country, COUNT(*) as count").
		Where(whereSQL+" AND country != ''", whereConditions...).
		Group("country").
		Order("count DESC").
		Limit(10).
		Find(&countryStats)
	analysis.TopCountries = countryStats

	// 热门城市统计
	var cityStats []dto.CityStatistic
	d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Select("city, COUNT(*) as count").
		Where(whereSQL+" AND city != ''", whereConditions...).
		Group("city").
		Order("count DESC").
		Limit(10).
		Find(&cityStats)
	analysis.TopCities = cityStats

	// 热门来源统计
	var refererStats []dto.RefererStatistic
	d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Select("referer, COUNT(*) as count").
		Where(whereSQL+" AND referer != ''", whereConditions...).
		Group("referer").
		Order("count DESC").
		Limit(10).
		Find(&refererStats)
	analysis.TopReferers = refererStats

	// 小时统计
	var hourlyStats []dto.HourlyStatistic
	d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Select("HOUR(click_date) as hour, COUNT(*) as count").
		Where(whereSQL, whereConditions...).
		Group("HOUR(click_date)").
		Order("hour").
		Find(&hourlyStats)
	analysis.HourlyStats = hourlyStats

	// 日统计
	var dailyStats []dto.DailyStatistic
	d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Select("DATE(click_date) as date, COUNT(*) as count").
		Where(whereSQL, whereConditions...).
		Group("DATE(click_date)").
		Order("date").
		Find(&dailyStats)
	analysis.DailyStats = dailyStats

	return analysis, nil
}
