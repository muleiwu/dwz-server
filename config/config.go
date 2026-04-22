package config

import (
	"cnb.cool/mliev/dwz/dwz-server/config/autoload"
	"cnb.cool/mliev/open/go-web/pkg/interfaces"
)

type Config struct{}

func (Config) Get() []interfaces.InitConfig {
	return []interfaces.InitConfig{
		autoload.App{},
		autoload.Http{},
		autoload.StaticFs{},
		autoload.Middleware{},
		autoload.Router{},
		autoload.Database{},
		autoload.Redis{},
		autoload.Cache{},
		autoload.IdGenerator{},
		autoload.Jwt{},
	}
}
