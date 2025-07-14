package database

import (
	"cnb.cool/mliev/open/dwz-server/config"
	"fmt"
)

type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:   config.GetString("database.driver", "mysql"),
		Host:     config.GetString("database.host", "localhost"),
		Port:     config.GetInt("database.port", 3306),
		DBName:   config.GetString("database.dbname", "dwz"),
		Username: config.GetString("database.username", "dwz"),
		Password: config.GetString("database.password", "dwz"),
	}
}

func (dc DatabaseConfig) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dc.Username,
		dc.Password,
		dc.Host,
		dc.Port,
		dc.DBName)
}

func (dc DatabaseConfig) GetPostgreSQLDSN() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		dc.Username,
		dc.Password,
		dc.Host,
		dc.Port,
		dc.DBName)
}
