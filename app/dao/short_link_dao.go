package dao

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/base62"

	"gorm.io/gorm"
)

type ShortLinkDao struct {
	helper interfaces.HelperInterface
}

func NewShortLinkDao(helper interfaces.HelperInterface) *ShortLinkDao {
	return &ShortLinkDao{helper: helper}
}

// Create 创建短网址
func (d *ShortLinkDao) Create(shortLink *model.ShortLink) error {
	return d.helper.GetDatabase().Create(shortLink).Error
}

// FindByShortCode 根据短代码和域名查找短网址（兼容旧方式）
func (d *ShortLinkDao) FindByShortCode(domain, shortCode string) (*model.ShortLink, error) {
	var shortLink model.ShortLink
	err := d.helper.GetDatabase().Where("domain = ? AND short_code = ? AND deleted_at IS NULL", domain, shortCode).First(&shortLink).Error
	if err != nil {
		return nil, err
	}
	return &shortLink, nil
}

// FindByShortCodeDecoded 根据短代码解码后的ID查找短网址（新方式）
func (d *ShortLinkDao) FindByShortCodeDecoded(domain, shortCode string) (*model.ShortLink, error) {
	// 先尝试解码短代码为ID
	converter := base62.NewBase62()
	id, err := converter.Decode(shortCode)
	if err != nil {
		// 如果解码失败，回退到旧的查找方式（用于兼容自定义短代码）
		return d.FindByShortCode(domain, shortCode)
	}

	// 使用解码后的ID直接查询
	var shortLink model.ShortLink
	err = d.helper.GetDatabase().Where("id = ? AND domain = ? AND deleted_at IS NULL", id, domain).First(&shortLink).Error
	if err != nil {
		return nil, err
	}
	return &shortLink, nil
}

// FindByID 根据ID查找短网址
func (d *ShortLinkDao) FindByID(id uint64) (*model.ShortLink, error) {
	var shortLink model.ShortLink
	err := d.helper.GetDatabase().Where("id = ? AND deleted_at IS NULL", id).First(&shortLink).Error
	if err != nil {
		return nil, err
	}
	return &shortLink, nil
}

// Update 更新短网址
func (d *ShortLinkDao) Update(shortLink *model.ShortLink) error {
	return d.helper.GetDatabase().Save(shortLink).Error
}

// Delete 删除短网址（软删除）
func (d *ShortLinkDao) Delete(id uint64) error {
	return d.helper.GetDatabase().Delete(&model.ShortLink{}, id).Error
}

// List 获取短网址列表
func (d *ShortLinkDao) List(offset, limit int, domain, keyword string) ([]model.ShortLink, int64, error) {
	var shortLinks []model.ShortLink
	var total int64

	query := d.helper.GetDatabase().Model(&model.ShortLink{}).Where("deleted_at IS NULL")

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
	return d.helper.GetDatabase().Model(&model.ShortLink{}).Where("id = ?", id).UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).Error
}

// GetClickStatistics 获取点击统计
func (d *ShortLinkDao) GetClickStatistics(shortLinkID uint64, startDate, endDate time.Time) ([]model.ClickStatistic, error) {
	var statistics []model.ClickStatistic
	err := d.helper.GetDatabase().Where("short_link_id = ? AND click_date >= ? AND click_date <= ?",
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

	err := d.helper.GetDatabase().Model(&model.ClickStatistic{}).
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
			d.helper.GetLogger().Error(err.Error())
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
	err := d.helper.GetDatabase().Model(&model.ClickStatistic{}).
		Where("short_link_id = ? AND click_date >= ? AND click_date < ?",
			shortLinkID, startDate, endDate).Count(&count).Error
	return count, err
}

// ExistsByDomainAndCode 检查域名和短代码是否已存在
func (d *ShortLinkDao) ExistsByDomainAndCode(domain, shortCode string) (bool, error) {
	var count int64
	err := d.helper.GetDatabase().Model(&model.ShortLink{}).
		Where("domain = ? AND short_code = ? AND deleted_at IS NULL", domain, shortCode).
		Count(&count).Error
	return count > 0, err
}

// ExistsByID 检查ID是否已存在
func (d *ShortLinkDao) ExistsByID(id uint64) (bool, error) {
	var count int64
	err := d.helper.GetDatabase().Model(&model.ShortLink{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Count(&count).Error
	return count > 0, err
}

// GetMaxIDByDomain 获取指定域名下的最大ID
func (d *ShortLinkDao) GetMaxIDByDomain(domain string) (uint64, error) {
	var maxID uint64
	err := d.helper.GetDatabase().Model(&model.ShortLink{}).
		Where("domain = ? AND deleted_at IS NULL AND issuer_number IS NOT NULL", domain).
		Select("COALESCE(MAX(issuer_number), 0)").
		Row().Scan(&maxID)
	if err != nil {
		return 0, err
	}
	return maxID, nil
}
