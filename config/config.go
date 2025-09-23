package config

import (
	"cnb.cool/mliev/open/dwz-server/config/autoload"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type Config struct {
}

func (receiver Config) Get() []interfaces.InitConfig {
	return []interfaces.InitConfig{
		autoload.Base{},
		autoload.Http{},
		autoload.Config{},
		autoload.Migration{},
		autoload.StaticFs{},
		autoload.Database{},
		autoload.Redis{},
		autoload.Middleware{},
		autoload.Router{},
	}
}
