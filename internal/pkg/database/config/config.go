package config

import (
	"fmt"

	configInterface "cnb.cool/mliev/open/dwz-server/internal/pkg/config/interfaces"
)

type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

func NewConfig(config configInterface.ConfigInterface) *DatabaseConfig {
	return &DatabaseConfig{
		Driver:   config.GetString("database.driver", "postgresql"),
		Host:     config.GetString("database.host", "127.0.0.1"),
		Port:     config.GetInt("database.port", 5432),
		DBName:   config.GetString("database.dbname", "test"),
		Username: config.GetString("database.username", "test"),
		Password: config.GetString("database.password", "123456"),
	}
}

func (dc *DatabaseConfig) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dc.Username,
		dc.Password,
		dc.Host,
		dc.Port,
		dc.DBName)
}

func (dc *DatabaseConfig) GetPostgreSQLDSN() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		dc.Username,
		dc.Password,
		dc.Host,
		dc.Port,
		dc.DBName)
}
