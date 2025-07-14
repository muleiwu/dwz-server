package database

import (
	"cnb.cool/mliev/open/dwz-server/config/database"
	"cnb.cool/mliev/open/dwz-server/config/migration"
	"cnb.cool/mliev/open/dwz-server/helper/logger"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"sync"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

// initDB initializes the database connection (private function)
func initDB() {
	var err error
	dbConfig := database.GetDatabaseConfig()

	var driver gorm.Dialector
	if dbConfig.Driver == "postgresql" {
		driver = postgres.New(
			postgres.Config{
				DSN:                  dbConfig.GetPostgreSQLDSN(),
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			})
	} else {
		// Default to MySQL
		driver = mysql.Open(dbConfig.GetMySQLDSN())
	}

	db, err = gorm.Open(driver, &gorm.Config{})

	if err != nil {
		logger.Logger().Error(fmt.Sprintf("[db connect err:%s]", err.Error()))
		os.Exit(1)
		return
	}
}

func AutoMigrate() error {
	// Auto migrate the database schema using migration config
	migrationConfig := migration.MigrationConfig{}
	migrationModels := migrationConfig.Get()

	if len(migrationModels) > 0 {
		err := GetDB().AutoMigrate(migrationModels...)
		if err != nil {
			return err
		}

		logger.Logger().Info(fmt.Sprintf("[db migration success: %d models migrated]", len(migrationModels)))
	}

	return nil
}

// Database returns the singleton database instance
func Database() *gorm.DB {
	dbOnce.Do(initDB)
	return db
}

// GetDB returns the singleton database instance (alias for Database)
func GetDB() *gorm.DB {
	return Database()
}
