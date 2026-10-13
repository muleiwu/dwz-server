package service

import (
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

func TestNormalizeInstallDatabaseConfigSupportsMariaDB(t *testing.T) {
	cfg, err := NormalizeInstallDatabaseConfig(InstallDatabaseConfig{
		Driver: "  MariaDB  ", Host: " 127.0.0.1 ", Port: 3307,
		DBName: "dwz", Username: "dwz", Password: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Driver != "mariadb" || cfg.Host != "127.0.0.1" || cfg.Port != 3307 {
		t.Fatalf("unexpected normalized config: driver=%q host=%q port=%d", cfg.Driver, cfg.Host, cfg.Port)
	}
	content, err := renderInstallConfigFile(cfg, InstallRedisConfig{}, "memory", "local", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var saved installConfigFile
	if err := yaml.Unmarshal(content, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Database.Driver != "mariadb" || saved.Database.Host != cfg.Host ||
		saved.Database.Port != cfg.Port || saved.Database.DBName != cfg.DBName ||
		saved.Database.Username != cfg.Username || saved.Database.Password != cfg.Password {
		t.Fatal("MariaDB config was not preserved in generated YAML")
	}
}

func TestNormalizeInstallDatabaseConfigValidatesNetworkDatabases(t *testing.T) {
	cases := []struct {
		name   string
		change func(*InstallDatabaseConfig)
	}{
		{"host", func(c *InstallDatabaseConfig) { c.Host = "127.0.0.1)/dwz?allowAllFiles=true" }},
		{"zero_port", func(c *InstallDatabaseConfig) { c.Port = 0 }},
		{"large_port", func(c *InstallDatabaseConfig) { c.Port = 65536 }},
		{"empty_database", func(c *InstallDatabaseConfig) { c.DBName = " " }},
		{"database_parameters", func(c *InstallDatabaseConfig) { c.DBName = "dwz?allowAllFiles=true" }},
		{"database_control_character", func(c *InstallDatabaseConfig) { c.DBName = "dwz\n" }},
		{"empty_user", func(c *InstallDatabaseConfig) { c.Username = " " }},
	}
	for _, driver := range []string{"mysql", "mariadb", "postgresql"} {
		t.Run(driver, func(t *testing.T) {
			cfg := InstallDatabaseConfig{Driver: driver, Host: "localhost", Port: 3306, DBName: "dwz", Username: "dwz"}
			if _, err := NormalizeInstallDatabaseConfig(cfg); err != nil {
				t.Fatalf("valid config rejected: %v", err)
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					invalid := cfg
					tc.change(&invalid)
					if _, err := NormalizeInstallDatabaseConfig(invalid); err == nil {
						t.Fatal("expected invalid config to be rejected")
					}
				})
			}
		})
	}
}

func TestInstallMariaDBUsesMySQLMigrations(t *testing.T) {
	if installDialectMap["mariadb"] != "mysql" || installMigrationDirMap["mariadb"] != "mysql" {
		t.Fatal("MariaDB must use the MySQL dialect and migration directory during installation")
	}
}

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
