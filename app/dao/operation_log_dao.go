package dao

import (
	"time"

	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type OperationLogDAO struct {
	helper interfaces.HelperInterface
}

func NewOperationLogDAO(helper interfaces.HelperInterface) *OperationLogDAO {
	return &OperationLogDAO{
		helper: helper,
	}
}

// Create 创建操作日志
func (dao *OperationLogDAO) Create(log *model.OperationLog) error {
	return dao.helper.GetDatabase().Create(log).Error
}

// GetByID 根据ID获取日志
func (dao *OperationLogDAO) GetByID(id uint64) (*model.OperationLog, error) {
	var log model.OperationLog
	err := dao.helper.GetDatabase().Preload("User").Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetList 获取操作日志列表
func (dao *OperationLogDAO) GetList(offset, limit int, userID *uint64, username, operation, resource, method string, status *int8, startTime, endTime *time.Time) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := dao.helper.GetDatabase().Model(&model.OperationLog{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if operation != "" {
		query = query.Where("operation LIKE ?", "%"+operation+"%")
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if method != "" {
		query = query.Where("method = ?", method)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取列表
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

// DeleteOldLogs 删除过期日志（物理删除）
func (dao *OperationLogDAO) DeleteOldLogs(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return dao.helper.GetDatabase().Unscoped().Where("created_at < ?", cutoffTime).Delete(&model.OperationLog{}).Error
}
