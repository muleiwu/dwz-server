package assembly

import (
	"sync"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/env/impl"
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
