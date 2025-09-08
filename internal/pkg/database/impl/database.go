package impl

import (
	"fmt"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getMySQLDSN(host string, port int, username string, password string, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username,
		password,
		host,
		port,
		dbName)
}

func getPostgreSQLDSN(host string, port int, username string, password string, dbName string) string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		username,
		password,
		host,
		port,
		dbName)
}
func getSqliteSQLDSN(host string, port int, username string, password string, dbName string) string {
	return host
}

func NewDatabase(helper interfaces.GetHelperInterface, driver string, host string, port int, dbName string, username string, password string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if driver == "postgresql" {
		dialector = postgres.New(postgres.Config{
			DSN:                  getPostgreSQLDSN(host, port, username, password, dbName),
			PreferSimpleProtocol: true,
		})
	} else if driver == "mysql" {
		dialector = mysql.Open(getMySQLDSN(host, port, username, password, dbName))
	} else if driver == "sqlite" {
		dialector = sqlite.Open(getSqliteSQLDSN(host, port, username, password, dbName))
	} else if driver == "memory" {
		dialector = sqlite.Open(":memory:")
	} else {
		return nil, fmt.Errorf("Unsupported driver: %s", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
