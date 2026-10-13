package database

import (
	"errors"
	"testing"

	driverManager "cnb.cool/mliev/open/go-web/pkg/driver"
	"cnb.cool/mliev/open/go-web/pkg/server/database/config"
	"cnb.cool/mliev/open/go-web/pkg/server/database/driver"
	"gorm.io/gorm"
)

func TestOpenSelectsDriverAndPreservesConfig(t *testing.T) {
	for _, name := range []string{"mariadb", "mysql", "postgresql", "sqlite", "memory"} {
		t.Run(name, func(t *testing.T) {
			originalManager := driver.DatabaseDriverManager
			driver.DatabaseDriverManager = driverManager.NewManager[*gorm.DB]()
			t.Cleanup(func() { driver.DatabaseDriverManager = originalManager })

			cfg := &config.DatabaseConfig{
				Driver: name, Host: "localhost", Port: 3307, DBName: "dwz",
				Username: "dwz", Password: "secret", PreferSimpleProtocol: true,
			}
			original := *cfg
			want := original
			if name == "mariadb" {
				want.Driver = "mysql"
			}
			connection := &gorm.DB{}
			factoryCalls := 0
			driver.DatabaseDriverManager.Extend(want.Driver, func(value any) (*gorm.DB, error) {
				factoryCalls++
				got, ok := value.(*config.DatabaseConfig)
				if !ok || got == cfg || *got != want {
					t.Fatal("driver must receive a copy with only the connection driver mapped")
				}
				return connection, nil
			})
			got, err := Open(cfg)
			if err != nil || got != connection || factoryCalls != 1 {
				t.Fatalf("connection factory not used correctly: calls=%d err=%v", factoryCalls, err)
			}
			if *cfg != original {
				t.Fatal("application database config was modified")
			}
		})
	}
}

func TestOpenReturnsConnectionError(t *testing.T) {
	originalManager := driver.DatabaseDriverManager
	driver.DatabaseDriverManager = driverManager.NewManager[*gorm.DB]()
	t.Cleanup(func() { driver.DatabaseDriverManager = originalManager })
	want := errors.New("connection failed")
	driver.DatabaseDriverManager.Extend("mysql", func(any) (*gorm.DB, error) {
		return nil, want
	})
	db, err := Open(&config.DatabaseConfig{Driver: "mariadb"})
	if db != nil || !errors.Is(err, want) {
		t.Fatalf("connection error was not preserved: %v", err)
	}
}
