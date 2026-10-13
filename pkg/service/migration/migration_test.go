package migration

import (
	"testing"

	ceMigrations "cnb.cool/mliev/dwz/dwz-server/v2/migrations"
)

func TestMariaDBUsesMySQLMigrations(t *testing.T) {
	if dialectMap["mariadb"] != "mysql" {
		t.Fatal("MariaDB must use the MySQL dialect")
	}
	dir, ok := resolveMigrationDir(ceMigrations.FS, dirByDriver["mariadb"])
	if !ok || dir != "mysql" {
		t.Fatalf("MariaDB must resolve to the existing MySQL migrations, got %q (found=%t)", dir, ok)
	}
}

func TestResolveMigrationDirSupportsPackageRootFS(t *testing.T) {
	dir, ok := resolveMigrationDir(ceMigrations.FS, "migrations/mysql")
	if !ok {
		t.Fatal("expected package-root migrations FS to be resolved")
	}
	if dir != "mysql" {
		t.Fatalf("expected mysql dir, got %s", dir)
	}
}
