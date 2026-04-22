package autoload

import (
	"cnb.cool/mliev/open/go-web/pkg/helper"
)

type Database struct{}

func (Database) InitConfig() map[string]any {
	env := helper.GetEnv()
	driver := env.GetString("database.driver", "mysql")
	host := env.GetString("database.host", "localhost")
	// SQLite uses the host field as the file path (go-web convention).
	if driver == "sqlite" {
		host = env.GetString("database.filepath", "./config/sqlite.db")
	}
	return map[string]any{
		"database.driver":   driver,
		"database.host":     host,
		"database.port":     env.GetInt("database.port", 3306),
		"database.dbname":   env.GetString("database.dbname", "dwz"),
		"database.username": env.GetString("database.username", "dwz"),
		"database.password": env.GetString("database.password", "dwz"),
		// retained for backwards compatibility with any direct readers
		"database.filepath": env.GetString("database.filepath", "./config/sqlite.db"),
	}
}
