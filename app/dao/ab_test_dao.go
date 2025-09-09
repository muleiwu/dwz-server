package dao

import (
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"gorm.io/gorm"
)

type ABTestDao struct {
	helper interfaces.HelperInterface
}

func NewABTestDao(helper interfaces.HelperInterface) *ABTestDao {
	return &ABTestDao{helper: helper}
}

// CreateABTest 创建AB测试
func (dao *ABTestDao) CreateABTest(abTest *model.ABTest) error {
	return dao.helper.GetDatabase().Create(abTest).Error
}

// CreateABTestVariant 创建AB测试变体
func (dao *ABTestDao) CreateABTestVariant(variant *model.ABTestVariant) error {
	db := dao.helper.GetDatabase()
	return db.Create(variant).Error
}

// FindABTestByID 根据ID查找AB测试
func (dao *ABTestDao) FindABTestByID(id uint64) (*model.ABTest, error) {
	db := dao.helper.GetDatabase()
	var abTest model.ABTest
	err := db.Preload("Variants").Where("id = ?", id).First(&abTest).Error
	return &abTest, err
}

// FindABTestByShortLinkID 根据短链接ID查找正在运行的AB测试
func (dao *ABTestDao) FindActiveABTestByShortLinkID(shortLinkID uint64) (*model.ABTest, error) {
	db := dao.helper.GetDatabase()
	var abTest model.ABTest
	err := db.Preload("Variants", "is_active = ?", true).
		Where("short_link_id = ? AND is_active = ? AND status = ?", shortLinkID, true, "running").
		First(&abTest).Error
	return &abTest, err
}

// FindABTestVariantByID 根据ID查找AB测试变体
func (dao *ABTestDao) FindABTestVariantByID(id uint64) (*model.ABTestVariant, error) {
	db := dao.helper.GetDatabase()
	var variant model.ABTestVariant
	err := db.Where("id = ?", id).First(&variant).Error
	return &variant, err
}

// UpdateABTest 更新AB测试
func (dao *ABTestDao) UpdateABTest(abTest *model.ABTest) error {
	db := dao.helper.GetDatabase()
	return db.Save(abTest).Error
}

// UpdateABTestVariant 更新AB测试变体
func (dao *ABTestDao) UpdateABTestVariant(variant *model.ABTestVariant) error {
	db := dao.helper.GetDatabase()
	return db.Save(variant).Error
}

// DeleteABTest 删除AB测试
func (dao *ABTestDao) DeleteABTest(id uint64) error {
	db := dao.helper.GetDatabase()
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除相关的变体
		if err := tx.Where("ab_test_id = ?", id).Delete(&model.ABTestVariant{}).Error; err != nil {
			return err
		}
		// 删除AB测试
		return tx.Delete(&model.ABTest{}, id).Error
	})
}

// DeleteABTestVariant 删除AB测试变体
func (dao *ABTestDao) DeleteABTestVariant(id uint64) error {
	db := dao.helper.GetDatabase()
	return db.Delete(&model.ABTestVariant{}, id).Error
}

// ListABTests 获取AB测试列表
func (dao *ABTestDao) ListABTests(offset, limit int, shortLinkID uint64, status string) ([]model.ABTest, int64, error) {
	db := dao.helper.GetDatabase()
	var abTests []model.ABTest
	var total int64

	query := db.Model(&model.ABTest{}).Preload("Variants")

	if shortLinkID > 0 {
		query = query.Where("short_link_id = ?", shortLinkID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取列表
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&abTests).Error
	return abTests, total, err
}

// CreateABTestClickStatistic 创建AB测试点击统计
func (dao *ABTestDao) CreateABTestClickStatistic(stat *model.ABTestClickStatistic) error {
	db := dao.helper.GetDatabase()
	return db.Create(stat).Error
}

// GetABTestStatistics 获取AB测试统计数据
func (dao *ABTestDao) GetABTestStatistics(abTestID uint64, days int) (map[uint64]int64, error) {
	db := dao.helper.GetDatabase()
	var results []struct {
		VariantID  uint64 `json:"variant_id"`
		ClickCount int64  `json:"click_count"`
	}

	query := db.Model(&model.ABTestClickStatistic{}).
		Select("variant_id, COUNT(*) as click_count").
		Where("ab_test_id = ?", abTestID).
		Group("variant_id")

	if days > 0 {
		query = query.Where("click_date >= DATE_SUB(NOW(), INTERVAL ? DAY)", days)
	}

	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	statistics := make(map[uint64]int64)
	for _, result := range results {
		statistics[result.VariantID] = result.ClickCount
	}

	return statistics, nil
}

// GetDailyABTestStatistics 获取AB测试每日统计数据
func (dao *ABTestDao) GetDailyABTestStatistics(abTestID uint64, days int) ([]map[string]interface{}, error) {
	db := dao.helper.GetDatabase()
	var results []struct {
		Date       string `json:"date"`
		VariantID  uint64 `json:"variant_id"`
		ClickCount int64  `json:"click_count"`
	}

	query := db.Model(&model.ABTestClickStatistic{}).
		Select("DATE(click_date) as date, variant_id, COUNT(*) as click_count").
		Where("ab_test_id = ?", abTestID).
		Group("DATE(click_date), variant_id").
		Order("date DESC, variant_id ASC")

	if days > 0 {
		query = query.Where("click_date >= DATE_SUB(NOW(), INTERVAL ? DAY)", days)
	}

	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	statisticsMap := make(map[string]map[uint64]int64)
	for _, result := range results {
		if statisticsMap[result.Date] == nil {
			statisticsMap[result.Date] = make(map[uint64]int64)
		}
		statisticsMap[result.Date][result.VariantID] = result.ClickCount
	}

	var dailyStats []map[string]interface{}
	for date, variantStats := range statisticsMap {
		dailyStats = append(dailyStats, map[string]interface{}{
			"date":     date,
			"variants": variantStats,
		})
	}

	return dailyStats, nil
}

// CheckSessionExists 检查会话是否存在（用于去重）
func (dao *ABTestDao) CheckSessionExists(abTestID, variantID uint64, sessionID string) (bool, error) {
	db := dao.helper.GetDatabase()
	var count int64
	err := db.Model(&model.ABTestClickStatistic{}).
		Where("ab_test_id = ? AND variant_id = ? AND session_id = ?", abTestID, variantID, sessionID).
		Count(&count).Error
	return count > 0, err
}
