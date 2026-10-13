package database

import (
	"cnb.cool/mliev/open/go-web/pkg/server/database/config"
	"cnb.cool/mliev/open/go-web/pkg/server/database/driver"
	"gorm.io/gorm"
)

// Open creates a connection without changing the application's database config.
func Open(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	connectionConfig := *cfg
	// MariaDB shares the MySQL protocol, but keeps its own persisted driver name.
	if connectionConfig.Driver == "mariadb" {
		connectionConfig.Driver = "mysql"
	}
	return driver.DatabaseDriverManager.Make(connectionConfig.Driver, &connectionConfig)
}
