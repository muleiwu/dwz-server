package service

import (
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

func TestNormalizeInstallDatabaseConfigRejectsMySQLParamInjection(t *testing.T) {
	_, err := NormalizeInstallDatabaseConfig(InstallDatabaseConfig{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		DBName:   "dwz?allowAllFiles=true",
		Username: "dwz",
		Password: "secret",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "连接串分隔符") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeInstallDatabaseConfigRejectsHostBreakout(t *testing.T) {
	_, err := NormalizeInstallDatabaseConfig(InstallDatabaseConfig{
		Driver:   "mysql",
		Host:     "127.0.0.1)/dwz?allowAllFiles=true",
		Port:     3306,
		DBName:   "dwz",
		Username: "dwz",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRenderInstallConfigFileEscapesYAMLValues(t *testing.T) {
	password := "pw\nlogger:\n  level: debug"
	content, err := renderInstallConfigFile(InstallDatabaseConfig{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		DBName:   "dwz",
		Username: "dwz",
		Password: password,
	}, InstallRedisConfig{}, "local", "local", time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("renderInstallConfigFile returned error: %v", err)
	}

	var parsed installConfigFile
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("generated YAML did not parse: %v\n%s", err, content)
	}
	if parsed.Database.Password != password {
		t.Fatalf("password was not preserved: %q", parsed.Database.Password)
	}
	if parsed.Logger.Level != "info" {
		t.Fatalf("logger level was overwritten: %q", parsed.Logger.Level)
	}
}
