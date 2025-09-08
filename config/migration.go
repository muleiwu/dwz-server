package config

import "cnb.cool/mliev/open/dwz-server/app/model"

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
