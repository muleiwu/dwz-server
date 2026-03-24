package config

import (
	"cnb.cool/mliev/dwz/dwz-server/config/autoload"
	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/http_server/service"
	idGeneratorService "cnb.cool/mliev/dwz/dwz-server/pkg/service/id_generator/service"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/migration"
)

type Server struct {
	Helper interfaces.HelperInterface
}

func (receiver Server) Get() []interfaces.ServerInterface {
	return []interfaces.ServerInterface{
		&migration.Migration{
			Helper:    receiver.Helper,
			Migration: autoload.Migration{}.Get(),
		},
		&idGeneratorService.IDGenerator{
			Helper: receiver.Helper,
		},
		&service.HttpServer{
			Helper: receiver.Helper,
		},
	}
}
