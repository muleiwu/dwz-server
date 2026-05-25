package dao

import (
	"fmt"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"gorm.io/gorm"
)

type ABTestClickAggregate struct {
	VariantID    uint64
	ClickCount   int64
	UniqueClicks int64
}

type ABTestFeedbackAggregate struct {
	VariantID       uint64
	ConversionCount int64
	ConversionValue float64
}

type ABTestDao struct {
	helper interfaces.HelperInterface
}

func NewABTestDao(helper interfaces.HelperInterface) *ABTestDao {
	return &ABTestDao{helper: helper}
}

// getDBDriver 获取数据库驱动类型
func (dao *ABTestDao) getDBDriver() string {
	return dao.helper.GetConfig().GetString("database.driver", "mysql")
}

// getDateSubSQL 获取日期减法SQL表达式（兼容MySQL和SQLite）
func (dao *ABTestDao) getDateSubSQL(column string) string {
	driver := dao.getDBDriver()
	switch driver {
	case "sqlite", "memory":
		return fmt.Sprintf("%s >= datetime('now', '-' || ? || ' days')", column)
	case "postgresql":
		return fmt.Sprintf("%s >= NOW() - (? * INTERVAL '1 day')", column)
	default:
		return fmt.Sprintf("%s >= DATE_SUB(NOW(), INTERVAL ? DAY)", column)
	}
}

// getDateSQL 获取日期提取SQL表达式（兼容MySQL和SQLite）
func (dao *ABTestDao) getDateSQL(column string) string {
	driver := dao.getDBDriver()
	if driver == "sqlite" {
		return "date(" + column + ")"
	}
	return "DATE(" + column + ")"
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

func (dao *ABTestDao) FindABTestByIDInWorkspace(id, workspaceID uint64) (*model.ABTest, error) {
	db := dao.helper.GetDatabase()
	var abTest model.ABTest
	err := db.Preload("Variants").Where("id = ? AND workspace_id = ?", id, workspaceID).First(&abTest).Error
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

func (dao *ABTestDao) ListABTestsInWorkspace(workspaceID uint64, offset, limit int, shortLinkID uint64, status string) ([]model.ABTest, int64, error) {
	db := dao.helper.GetDatabase()
	var abTests []model.ABTest
	var total int64

	query := db.Model(&model.ABTest{}).Preload("Variants").Where("workspace_id = ?", workspaceID)
	if shortLinkID > 0 {
		query = query.Where("short_link_id = ?", shortLinkID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&abTests).Error
	return abTests, total, err
}

// CreateABTestClickStatistic 创建AB测试点击统计
func (dao *ABTestDao) CreateABTestClickStatistic(stat *model.ABTestClickStatistic) error {
	db := dao.helper.GetDatabase()
	return db.Create(stat).Error
}

// CreateABTestFeedback 创建AB测试转化反馈
func (dao *ABTestDao) CreateABTestFeedback(feedback *model.ABTestFeedback) error {
	return dao.helper.GetDatabase().Create(feedback).Error
}

func (dao *ABTestDao) FindABTestFeedbackByEventID(abTestID uint64, eventID string) (*model.ABTestFeedback, error) {
	var feedback model.ABTestFeedback
	err := dao.helper.GetDatabase().
		Where("ab_test_id = ? AND event_id = ?", abTestID, eventID).
		First(&feedback).Error
	return &feedback, err
}

// GetABTestClickAggregates 获取AB测试点击与去重点击统计
func (dao *ABTestDao) GetABTestClickAggregates(abTestID uint64, days int) (map[uint64]ABTestClickAggregate, error) {
	db := dao.helper.GetDatabase()
	var results []struct {
		VariantID    uint64 `json:"variant_id"`
		ClickCount   int64  `json:"click_count"`
		UniqueClicks int64  `json:"unique_clicks"`
	}

	query := db.Model(&model.ABTestClickStatistic{}).
		Select("variant_id, COUNT(*) as click_count, COUNT(DISTINCT session_id) as unique_clicks").
		Where("ab_test_id = ?", abTestID).
		Group("variant_id")

	if days > 0 {
		query = query.Where(dao.getDateSubSQL("click_date"), days)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	aggregates := make(map[uint64]ABTestClickAggregate)
	for _, result := range results {
		aggregates[result.VariantID] = ABTestClickAggregate{
			VariantID:    result.VariantID,
			ClickCount:   result.ClickCount,
			UniqueClicks: result.UniqueClicks,
		}
	}
	return aggregates, nil
}

// GetABTestFeedbackAggregates 获取AB测试转化反馈统计
func (dao *ABTestDao) GetABTestFeedbackAggregates(abTestID uint64, days int) (map[uint64]ABTestFeedbackAggregate, error) {
	db := dao.helper.GetDatabase()
	var results []struct {
		VariantID       uint64  `json:"variant_id"`
		ConversionCount int64   `json:"conversion_count"`
		ConversionValue float64 `json:"conversion_value"`
	}

	query := db.Model(&model.ABTestFeedback{}).
		Select("variant_id, COUNT(*) as conversion_count, COALESCE(SUM(value), 0) as conversion_value").
		Where("ab_test_id = ?", abTestID).
		Group("variant_id")

	if days > 0 {
		query = query.Where(dao.getDateSubSQL("occurred_at"), days)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	aggregates := make(map[uint64]ABTestFeedbackAggregate)
	for _, result := range results {
		aggregates[result.VariantID] = ABTestFeedbackAggregate{
			VariantID:       result.VariantID,
			ConversionCount: result.ConversionCount,
			ConversionValue: result.ConversionValue,
		}
	}
	return aggregates, nil
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
		query = query.Where(dao.getDateSubSQL("click_date"), days)
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

	dateSQL := dao.getDateSQL("click_date")
	query := db.Model(&model.ABTestClickStatistic{}).
		Select(dateSQL+" as date, variant_id, COUNT(*) as click_count").
		Where("ab_test_id = ?", abTestID).
		Group(dateSQL + ", variant_id").
		Order("date DESC, variant_id ASC")

	if days > 0 {
		query = query.Where(dao.getDateSubSQL("click_date"), days)
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
