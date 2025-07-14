package dao

import (
	"cnb.cool/mliev/open/dwz-server/helper/logger"
	"time"

	"cnb.cool/mliev/open/dwz-server/helper/database"

	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/util"
	"gorm.io/gorm"
)

type ShortLinkDao struct{}

// Create 创建短网址
func (d *ShortLinkDao) Create(shortLink *model.ShortLink) error {
	return database.GetDB().Create(shortLink).Error
}

// FindByShortCode 根据短代码和域名查找短网址（兼容旧方式）
func (d *ShortLinkDao) FindByShortCode(domain, shortCode string) (*model.ShortLink, error) {
	var shortLink model.ShortLink
	err := database.GetDB().Where("domain = ? AND short_code = ? AND deleted_at IS NULL", domain, shortCode).First(&shortLink).Error
	if err != nil {
		return nil, err
	}
	return &shortLink, nil
}

// FindByShortCodeDecoded 根据短代码解码后的ID查找短网址（新方式）
func (d *ShortLinkDao) FindByShortCodeDecoded(domain, shortCode string) (*model.ShortLink, error) {
	// 先尝试解码短代码为ID
	converter := util.NewBase62Converter()
	id, err := converter.Decode(shortCode)
	if err != nil {
		// 如果解码失败，回退到旧的查找方式（用于兼容自定义短代码）
		return d.FindByShortCode(domain, shortCode)
	}

	// 使用解码后的ID直接查询
	var shortLink model.ShortLink
	err = database.GetDB().Where("id = ? AND domain = ? AND deleted_at IS NULL", id, domain).First(&shortLink).Error
	if err != nil {
		return nil, err
	}
	return &shortLink, nil
}

// FindByID 根据ID查找短网址
func (d *ShortLinkDao) FindByID(id uint64) (*model.ShortLink, error) {
	var shortLink model.ShortLink
	err := database.GetDB().Where("id = ? AND deleted_at IS NULL", id).First(&shortLink).Error
	if err != nil {
		return nil, err
	}
	return &shortLink, nil
}

// Update 更新短网址
func (d *ShortLinkDao) Update(shortLink *model.ShortLink) error {
	return database.GetDB().Save(shortLink).Error
}

// Delete 删除短网址（软删除）
func (d *ShortLinkDao) Delete(id uint64) error {
	return database.GetDB().Delete(&model.ShortLink{}, id).Error
}

// List 获取短网址列表
func (d *ShortLinkDao) List(offset, limit int, domain, keyword string) ([]model.ShortLink, int64, error) {
	var shortLinks []model.ShortLink
	var total int64

	query := database.GetDB().Model(&model.ShortLink{}).Where("deleted_at IS NULL")

	if domain != "" {
		query = query.Where("domain = ?", domain)
	}

	if keyword != "" {
		query = query.Where("original_url LIKE ? OR title LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取列表
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&shortLinks).Error
	return shortLinks, total, err
}

// IncrementClickCount 增加点击次数
func (d *ShortLinkDao) IncrementClickCount(id uint64) error {
	return database.GetDB().Model(&model.ShortLink{}).Where("id = ?", id).UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).Error
}

// GetClickStatistics 获取点击统计
func (d *ShortLinkDao) GetClickStatistics(shortLinkID uint64, startDate, endDate time.Time) ([]model.ClickStatistic, error) {
	var statistics []model.ClickStatistic
	err := database.GetDB().Where("short_link_id = ? AND click_date >= ? AND click_date <= ?",
		shortLinkID, startDate, endDate).Order("click_date DESC").Find(&statistics).Error
	return statistics, err
}

// GetDailyClickCount 获取每日点击统计
func (d *ShortLinkDao) GetDailyClickCount(shortLinkID uint64, days int) (map[string]int64, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	var results []struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	err := database.GetDB().Model(&model.ClickStatistic{}).
		Select("DATE(click_date) as date, COUNT(*) as count").
		Where("short_link_id = ? AND DATE(click_date) >= ? AND DATE(click_date) <= ?",
			shortLinkID, startDate, endDate).
		Group("DATE(click_date)").
		Order("date").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为map并确保所有日期都有值
	countMap := make(map[string]int64)

	// 先初始化所有日期的点击数为0
	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)
	for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		countMap[dateStr] = 0
	}

	// 用实际查询结果更新点击数
	for _, result := range results {
		tempTime, err := time.Parse(time.RFC3339, result.Date)
		if err != nil {
			logger.Logger().Error(err.Error())
			continue
		}
		dateStr := tempTime.Format("2006-01-02")
		countMap[dateStr] = result.Count
	}

	return countMap, nil
}

// GetClickCountByDateRange 获取指定时间范围内的点击数
func (d *ShortLinkDao) GetClickCountByDateRange(shortLinkID uint64, startDate, endDate time.Time) (int64, error) {
	var count int64
	err := database.GetDB().Model(&model.ClickStatistic{}).
		Where("short_link_id = ? AND click_date >= ? AND click_date < ?",
			shortLinkID, startDate, endDate).Count(&count).Error
	return count, err
}

// ExistsByDomainAndCode 检查域名和短代码是否已存在
func (d *ShortLinkDao) ExistsByDomainAndCode(domain, shortCode string) (bool, error) {
	var count int64
	err := database.GetDB().Model(&model.ShortLink{}).
		Where("domain = ? AND short_code = ? AND deleted_at IS NULL", domain, shortCode).
		Count(&count).Error
	return count > 0, err
}

// ExistsByID 检查ID是否已存在
func (d *ShortLinkDao) ExistsByID(id uint64) (bool, error) {
	var count int64
	err := database.GetDB().Model(&model.ShortLink{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Count(&count).Error
	return count > 0, err
}

// GetMaxIDByDomain 获取指定域名下的最大ID
func (d *ShortLinkDao) GetMaxIDByDomain(domain string) (uint64, error) {
	var maxID uint64
	err := database.GetDB().Model(&model.ShortLink{}).
		Where("domain = ? AND deleted_at IS NULL AND issuer_number IS NOT NULL", domain).
		Select("COALESCE(MAX(issuer_number), 0)").
		Row().Scan(&maxID)
	if err != nil {
		return 0, err
	}
	return maxID, nil
}

// ClickStatisticDao 点击统计DAO
type ClickStatisticDao struct{}

// Create 创建点击统计记录
func (d *ClickStatisticDao) Create(statistic *model.ClickStatistic) error {
	return database.GetDB().Create(statistic).Error
}

// List 获取点击统计列表
func (d *ClickStatisticDao) List(req *dto.ClickStatisticListRequest) ([]model.ClickStatistic, int64, error) {
	var statistics []model.ClickStatistic
	var total int64

	query := database.GetDB().Model(&model.ClickStatistic{})

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
	query := database.GetDB().Model(&model.ClickStatistic{})

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
	database.GetDB().Model(&model.ClickStatistic{}).
		Select("country, COUNT(*) as count").
		Where(whereSQL+" AND country != ''", whereConditions...).
		Group("country").
		Order("count DESC").
		Limit(10).
		Find(&countryStats)
	analysis.TopCountries = countryStats

	// 热门城市统计
	var cityStats []dto.CityStatistic
	database.GetDB().Model(&model.ClickStatistic{}).
		Select("city, COUNT(*) as count").
		Where(whereSQL+" AND city != ''", whereConditions...).
		Group("city").
		Order("count DESC").
		Limit(10).
		Find(&cityStats)
	analysis.TopCities = cityStats

	// 热门来源统计
	var refererStats []dto.RefererStatistic
	database.GetDB().Model(&model.ClickStatistic{}).
		Select("referer, COUNT(*) as count").
		Where(whereSQL+" AND referer != ''", whereConditions...).
		Group("referer").
		Order("count DESC").
		Limit(10).
		Find(&refererStats)
	analysis.TopReferers = refererStats

	// 小时统计
	var hourlyStats []dto.HourlyStatistic
	database.GetDB().Model(&model.ClickStatistic{}).
		Select("HOUR(click_date) as hour, COUNT(*) as count").
		Where(whereSQL, whereConditions...).
		Group("HOUR(click_date)").
		Order("hour").
		Find(&hourlyStats)
	analysis.HourlyStats = hourlyStats

	// 日统计
	var dailyStats []dto.DailyStatistic
	database.GetDB().Model(&model.ClickStatistic{}).
		Select("DATE(click_date) as date, COUNT(*) as count").
		Where(whereSQL, whereConditions...).
		Group("DATE(click_date)").
		Order("date").
		Find(&dailyStats)
	analysis.DailyStats = dailyStats

	return analysis, nil
}

// DomainDao 域名DAO
type DomainDao struct{}

// Create 创建域名
func (d *DomainDao) Create(domain *model.Domain) error {
	return database.GetDB().Create(domain).Error
}

// FindByDomain 根据域名查找
func (d *DomainDao) FindByDomain(domain string) (*model.Domain, error) {
	var domainModel model.Domain
	err := database.GetDB().Where("domain = ? AND deleted_at IS NULL", domain).First(&domainModel).Error
	if err != nil {
		return nil, err
	}
	return &domainModel, nil
}

// List 获取域名列表
func (d *DomainDao) List() ([]model.Domain, error) {
	var domains []model.Domain
	err := database.GetDB().Where("deleted_at IS NULL").Order("created_at DESC").Find(&domains).Error
	return domains, err
}

// Update 更新域名
func (d *DomainDao) Update(domain *model.Domain) error {
	return database.GetDB().Save(domain).Error
}

// Delete 删除域名（软删除）
func (d *DomainDao) Delete(id uint64) error {
	return database.GetDB().Delete(&model.Domain{}, id).Error
}

// GetActiveDomains 获取活跃域名列表
func (d *DomainDao) GetActiveDomains() ([]model.Domain, error) {
	var domains []model.Domain
	err := database.GetDB().Where("is_active = ? AND deleted_at IS NULL", true).Find(&domains).Error
	return domains, err
}

// ExistsByDomain 检查域名是否已存在
func (d *DomainDao) ExistsByDomain(domain string) (bool, error) {
	var count int64
	err := database.GetDB().Model(&model.Domain{}).
		Where("domain = ? AND deleted_at IS NULL", domain).Count(&count).Error
	return count > 0, err
}
