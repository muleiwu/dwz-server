package assembly

import (
	"sync"

	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/env/impl"
)

type Env struct {
	Helper interfaces.HelperInterface
}

var (
	envOnce sync.Once
)

func (receiver *Env) Assembly() error {

	receiver.Helper.SetEnv(impl.NewEnv())
	return nil
}
