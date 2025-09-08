package autoload

import envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"

type Database struct {
}

func (receiver Database) InitConfig(helper envInterface.GetHelperInterface) map[string]any {
	return map[string]any{
		"database.driver":   helper.GetEnv().GetString("database.driver", "mysql"),
		"database.host":     helper.GetEnv().GetString("database.host", "127.0.0.1"),
		"database.port":     helper.GetEnv().GetInt("database.port", 3306),
		"database.dbname":   helper.GetEnv().GetString("database.dbname", "test"),
		"database.username": helper.GetEnv().GetString("database.username", "test"),
		"database.password": helper.GetEnv().GetString("database.password", "123456"),
	}
}
