package service

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/http_server/impl"
)

type HttpServer struct {
	Helper interfaces.HelperInterface
}

func (receiver *HttpServer) Run() error {

	newHttpServer := impl.NewHttpServer(receiver.Helper)

	newHttpServer.RunHttp()
	return nil
}
