package autoload

import (
	"cnb.cool/mliev/open/dwz-server/app/model"
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type Migration struct {
}

func (receiver Migration) Get() []any {
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

func (receiver Migration) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"database.migration": receiver.Get(),
	}
}
