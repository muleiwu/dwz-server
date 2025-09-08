package config

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/http_server/service"
	"cnb.cool/mliev/open/dwz-server/internal/service/migration"
)

type Server struct {
	Helper interfaces.HelperInterface
}

func (receiver Server) Get() []interfaces.ServerInterface {
	return []interfaces.ServerInterface{
		&migration.Migration{
			Helper:    receiver.Helper,
			Migration: Migration{}.Get(),
		},
		&service.HttpServer{
			Helper: receiver.Helper,
		},
	}
}
