package migration

import (
	"cnb.cool/mliev/open/dwz-server/app/model"
)

type MigrationConfig struct {
	// 是否启用自动迁移
	AutoMigrate bool
	// 需要迁移的模型列表
	Models []any
}

func (receiver MigrationConfig) Get() []any {
	if receiver.AutoMigrate {
		return receiver.Models
	}
	return []any{
		&model.ShortLink{},
		&model.ClickStatistic{},
		&model.Domain{},
		&model.ABTest{},
		&model.ABTestVariant{},
		&model.ABTestClickStatistic{},
		&model.User{},
		&model.UserToken{},
		&model.OperationLog{},
	}
}
