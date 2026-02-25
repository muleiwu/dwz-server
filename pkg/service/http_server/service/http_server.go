package service

import (
	"cnb.cool/mliev/open/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/service/http_server/impl"
)

type HttpServer struct {
	Helper interfaces.HelperInterface
}

func (receiver *HttpServer) Run() error {

	newHttpServer := impl.NewHttpServer(receiver.Helper)

	newHttpServer.RunHttp()
	return nil
}
