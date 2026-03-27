package config

import (
	"cnb.cool/mliev/dwz/dwz-server/config/autoload"
	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
)

type Config struct {
}

func (receiver Config) Get() []interfaces.InitConfig {
	return []interfaces.InitConfig{
		autoload.Base{},
		autoload.Http{},
		autoload.Config{},
		autoload.IdGenerator{},
		autoload.Migration{},
		autoload.StaticFs{},
		autoload.Database{},
		autoload.Redis{},
		autoload.Jwt{},
		autoload.Middleware{},
		autoload.Router{},
	}
}
