package autoload

import (
	"cnb.cool/mliev/open/go-web/pkg/helper"
)

type IdGenerator struct{}

func (IdGenerator) InitConfig() map[string]any {
	return map[string]any{
		"id_generator.driver": helper.GetEnv().GetString("id_generator.driver", "local"),
	}
}
