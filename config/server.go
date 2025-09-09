package config

import (
	"cnb.cool/mliev/open/dwz-server/config/autoload"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/http_server/service"
	idGeneratorService "cnb.cool/mliev/open/dwz-server/internal/pkg/id_generator/service"
	"cnb.cool/mliev/open/dwz-server/internal/service/migration"
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
