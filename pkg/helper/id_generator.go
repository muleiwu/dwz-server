package helper

import (
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"cnb.cool/mliev/open/go-web/pkg/container"
)

// GetIdGenerator returns the distributed ID generator. Returns nil while the
// install flow has not finished — callers must guard against that.
func GetIdGenerator() interfaces.IDGenerator {
	g, err := container.Get[interfaces.IDGenerator]()
	if err != nil {
		return nil
	}
	return g
}

// SetIdGenerator is retained for source-compat with code that previously
// initialized the generator manually. The id_generator Server now registers
// itself into the container, so this helper just forwards to it.
func SetIdGenerator(g interfaces.IDGenerator) {
	if g == nil {
		return
	}
	container.Register(container.NewSimpleProvider(idGeneratorType(), g))
}
