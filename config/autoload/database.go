package autoload

import (
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type Database struct {
}

func (receiver Database) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"database.driver":   helper.GetEnv().GetString("database.driver", "mysql"),
		"database.host":     helper.GetEnv().GetString("database.host", "localhost"),
		"database.port":     helper.GetEnv().GetInt("database.port", 3306),
		"database.dbname":   helper.GetEnv().GetString("database.dbname", "dwz"),
		"database.username": helper.GetEnv().GetString("database.username", "dwz"),
		"database.password": helper.GetEnv().GetString("database.password", "dwz"),
	}
}
