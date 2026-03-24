package assembly

import (
	"sync"

	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/database/config"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/database/impl"
)

type Database struct {
	Helper interfaces.HelperInterface
}

var (
	databaseOnce sync.Once
)

func (receiver *Database) Assembly() error {
	databaseConfig := config.NewConfig(receiver.Helper.GetConfig())
	database, err := impl.NewDatabase(receiver.Helper, databaseConfig.Driver, databaseConfig.Host, databaseConfig.Port, databaseConfig.DBName, databaseConfig.Username, databaseConfig.Password, databaseConfig.Filepath)
	if err != nil {
		return err
	}
	receiver.Helper.SetDatabase(database)

	return err
}
