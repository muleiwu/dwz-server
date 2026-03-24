package assembly

import (
	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/logger/impl"
)

type Logger struct {
	Helper interfaces.HelperInterface
}

func (receiver *Logger) Assembly() error {
	receiver.Helper.SetLogger(impl.NewLogger())
	return nil
}
