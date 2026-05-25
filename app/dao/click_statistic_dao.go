package dao

import (
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

// ClickStatisticDao 点击统计DAO
type ClickStatisticDao struct {
	helper interfaces.HelperInterface
}

func NewClickStatisticDao(helper interfaces.HelperInterface) *ClickStatisticDao {
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

	if req.Province != "" {
		query = query.Where("province = ?", req.Province)
	}

	if req.City != "" {
		query = query.Where("city = ?", req.City)
	}

	if req.ISP != "" {
		query = query.Where("isp = ?", req.ISP)
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

func (d *ClickStatisticDao) ListInWorkspace(workspaceID uint64, req *dto.ClickStatisticListRequest) ([]model.ClickStatistic, int64, error) {
	var statistics []model.ClickStatistic
	var total int64
	query := d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (req.Page - 1) * req.PageSize
	err := query.Order("click_date DESC").Offset(offset).Limit(req.PageSize).Find(&statistics).Error
	return statistics, total, err
}

func (d *ClickStatisticDao) ExportInWorkspace(workspaceID uint64, req *dto.ClickStatisticListRequest, limit int) ([]model.ClickStatistic, error) {
	var statistics []model.ClickStatistic
	query := d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req)
	err := query.Order("click_date DESC").Limit(limit).Find(&statistics).Error
	return statistics, err
}

func (d *ClickStatisticDao) applyFilters(query *gorm.DB, workspaceID uint64, req *dto.ClickStatisticListRequest) *gorm.DB {
	query = query.Where("click_statistics.workspace_id = ?", workspaceID)
	if req.ShortLinkID > 0 {
		query = query.Where("click_statistics.short_link_id = ?", req.ShortLinkID)
	}
	if req.CampaignID > 0 {
		query = query.Where("click_statistics.campaign_id = ?", req.CampaignID)
	}
	if req.RouteID > 0 {
		query = query.Where("click_statistics.route_id = ?", req.RouteID)
	}
	if req.TagID > 0 {
		query = query.Joins("JOIN short_link_tags slt ON slt.short_link_id = click_statistics.short_link_id AND slt.tag_id = ?", req.TagID)
	}
	if req.DeviceType != "" {
		query = query.Where("click_statistics.device_type = ?", req.DeviceType)
	}
	if req.IsBot != nil {
		query = query.Where("click_statistics.is_bot = ?", *req.IsBot)
	}
	if req.IP != "" {
		query = query.Where("click_statistics.ip = ?", req.IP)
	}
	if req.Country != "" {
		query = query.Where("click_statistics.country = ?", req.Country)
	}
	if req.Province != "" {
		query = query.Where("click_statistics.province = ?", req.Province)
	}
	if req.City != "" {
		query = query.Where("click_statistics.city = ?", req.City)
	}
	if req.ISP != "" {
		query = query.Where("click_statistics.isp = ?", req.ISP)
	}
	if !req.StartDate.IsZero() {
		query = query.Where("click_statistics.click_date >= ?", req.StartDate)
	}
	if !req.EndDate.IsZero() {
		query = query.Where("click_statistics.click_date < ?", req.EndDate)
	}
	return query
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

	// 热门省份统计
	var provinceStats []dto.ProvinceStatistic
	d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Select("province, COUNT(*) as count").
		Where(whereSQL+" AND province != ''", whereConditions...).
		Group("province").
		Order("count DESC").
		Limit(10).
		Find(&provinceStats)
	analysis.TopProvinces = provinceStats

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

	// 热门运营商统计
	var ispStats []dto.ISPStatistic
	d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Select("isp, COUNT(*) as count").
		Where(whereSQL+" AND isp != ''", whereConditions...).
		Group("isp").
		Order("count DESC").
		Limit(10).
		Find(&ispStats)
	analysis.TopISPs = ispStats

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

func (d *ClickStatisticDao) GetAnalysisInWorkspace(workspaceID uint64, req *dto.ClickStatisticListRequest) (*dto.ClickStatisticAnalysisResponse, error) {
	analysis := &dto.ClickStatisticAnalysisResponse{}

	d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Count(&analysis.TotalClicks)
	d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Distinct("ip").
		Count(&analysis.UniqueIPs)

	group := func(selectSQL, whereSQL, groupSQL, orderSQL string, dest any) {
		q := d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
			Select(selectSQL).
			Where(whereSQL).
			Group(groupSQL).
			Order(orderSQL).
			Limit(10)
		q.Find(dest)
	}

	group("country, COUNT(*) as count", "country != ''", "country", "count DESC", &analysis.TopCountries)
	group("province, COUNT(*) as count", "province != ''", "province", "count DESC", &analysis.TopProvinces)
	group("city, COUNT(*) as count", "city != ''", "city", "count DESC", &analysis.TopCities)
	group("isp, COUNT(*) as count", "isp != ''", "isp", "count DESC", &analysis.TopISPs)
	group("referer, COUNT(*) as count", "referer != ''", "referer", "count DESC", &analysis.TopReferers)
	group("device_type, COUNT(*) as count", "device_type != ''", "device_type", "count DESC", &analysis.TopDevices)
	group("browser, COUNT(*) as count", "browser != ''", "browser", "count DESC", &analysis.TopBrowsers)
	group("os, COUNT(*) as count", "os != ''", "os", "count DESC", &analysis.TopOS)
	group("utm_source AS value, COUNT(*) as count", "utm_source != ''", "utm_source", "count DESC", &analysis.TopUTMSources)
	group("utm_campaign AS value, COUNT(*) as count", "utm_campaign != ''", "utm_campaign", "count DESC", &analysis.TopUTMCampaigns)
	group("route_id, route_name, COUNT(*) as count", "route_id IS NOT NULL", "route_id, route_name", "count DESC", &analysis.TopRoutes)

	type botRow struct {
		IsBot bool
		Count int64
	}
	var botRows []botRow
	d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Select("is_bot, COUNT(*) as count").
		Group("is_bot").
		Find(&botRows)
	for _, row := range botRows {
		if row.IsBot {
			analysis.BotStats.BotClicks = row.Count
		} else {
			analysis.BotStats.HumanClicks = row.Count
		}
	}

	hourSQL := d.getHourSQL("click_date")
	dateSQL := d.getDateSQL("click_date")
	d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Select(hourSQL + " as hour, COUNT(*) as count").
		Group(hourSQL).
		Order("hour").
		Find(&analysis.HourlyStats)
	d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Select(dateSQL + " as date, COUNT(*) as count").
		Group(dateSQL).
		Order("date").
		Find(&analysis.DailyStats)

	return analysis, nil
}

func (d *ClickStatisticDao) GetGeoAnalysisInWorkspace(workspaceID uint64, req *dto.ClickStatisticListRequest, level string) (*dto.ClickStatisticGeoAnalysisResponse, error) {
	analysis := &dto.ClickStatisticGeoAnalysisResponse{
		Level:    level,
		Country:  req.Country,
		Province: req.Province,
		Regions:  []dto.GeoRegionStatistic{},
	}

	if err := d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Count(&analysis.TotalClicks).Error; err != nil {
		return nil, err
	}
	if err := d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Distinct("ip").
		Count(&analysis.UniqueIPs).Error; err != nil {
		return nil, err
	}

	column := "country"
	switch level {
	case "province":
		column = "province"
	case "city":
		column = "city"
	}
	columnSQL := "click_statistics." + column

	err := d.applyFilters(d.helper.GetDatabase().Model(&model.ClickStatistic{}), workspaceID, req).
		Select(columnSQL + " AS name, COUNT(*) as count").
		Where(columnSQL + " != ''").
		Group(columnSQL).
		Order("count DESC").
		Find(&analysis.Regions).Error
	if err != nil {
		return nil, err
	}

	return analysis, nil
}

func (d *ClickStatisticDao) getDBDriver() string {
	return d.helper.GetConfig().GetString("database.driver", "mysql")
}

func (d *ClickStatisticDao) getHourSQL(column string) string {
	switch d.getDBDriver() {
	case "sqlite":
		return "CAST(strftime('%H', " + column + ") AS INTEGER)"
	case "postgres", "postgresql":
		return "EXTRACT(HOUR FROM " + column + ")::INT"
	default:
		return "HOUR(" + column + ")"
	}
}

func (d *ClickStatisticDao) getDateSQL(column string) string {
	switch d.getDBDriver() {
	case "sqlite":
		return "date(" + column + ")"
	case "postgres", "postgresql":
		return "TO_CHAR(" + column + ", 'YYYY-MM-DD')"
	default:
		return "DATE(" + column + ")"
	}
}

// CountAll 获取所有点击统计数量
func (d *ClickStatisticDao) CountAll() (int64, error) {
	var count int64
	err := d.helper.GetDatabase().Model(&model.ClickStatistic{}).Count(&count).Error
	return count, err
}

// CountByDateRange 获取指定时间范围内的点击统计数量
func (d *ClickStatisticDao) CountByDateRange(startDate, endDate time.Time) (int64, error) {
	var count int64
	err := d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Where("click_date >= ? AND click_date < ?", startDate, endDate).
		Count(&count).Error
	return count, err
}
