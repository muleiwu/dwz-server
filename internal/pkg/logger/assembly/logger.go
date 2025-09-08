package assembly

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/logger/impl"
)

type Logger struct {
	Helper interfaces.HelperInterface
}

func (receiver *Logger) Assembly() error {
	receiver.Helper.SetLogger(impl.NewLogger())
	return nil
}
