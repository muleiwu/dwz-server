package migration

import (
	"testing"

	ceMigrations "cnb.cool/mliev/dwz/dwz-server/v2/migrations"
)

func TestResolveMigrationDirSupportsPackageRootFS(t *testing.T) {
	dir, ok := resolveMigrationDir(ceMigrations.FS, "migrations/mysql")
	if !ok {
		t.Fatal("expected package-root migrations FS to be resolved")
	}
	if dir != "mysql" {
		t.Fatalf("expected mysql dir, got %s", dir)
	}
}
