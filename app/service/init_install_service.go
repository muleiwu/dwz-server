package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/migrations"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/install_bootstrap"
	dbconfig "cnb.cool/mliev/open/go-web/pkg/server/database/config"
	dbDriver "cnb.cool/mliev/open/go-web/pkg/server/database/driver"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"go.yaml.in/yaml/v3"
	"gorm.io/gorm"
)

// InstallDatabaseConfig is the install-time DB descriptor. Mirrored locally so
// the install flow doesn't depend on go-web's narrower DatabaseConfig struct
// (which lacks the SQLite Filepath field).
type InstallDatabaseConfig struct {
	Driver   string
	Filepath string
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
}

type InstallRedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type InitInstallService struct {
	helper interfaces.HelperInterface
}

const (
	lockFile   = "./config/install.lock"
	configFile = "./config/config.yaml"
)

func NewInitInstallService(helper interfaces.HelperInterface) *InitInstallService {
	return &InitInstallService{helper: helper}
}

func (s *InitInstallService) CreateAdminUser(username, realName, email, phone, password string) error {
	user := &model.User{
		Username:      username,
		RealName:      realName,
		Email:         email,
		Phone:         phone,
		Status:        1,
		IsSystemAdmin: true,
	}
	if err := user.SetPassword(password); err != nil {
		return err
	}
	userDAO := dao.NewUserDAO(s.helper)
	if err := userDAO.Create(user); err != nil {
		return err
	}
	if err := dao.NewWorkspaceDao(s.helper).CreateMember(&model.WorkspaceMember{
		WorkspaceID: 1,
		UserID:      user.ID,
		Role:        model.WorkspaceRoleOwner,
		Status:      1,
	}); err != nil {
		return err
	}
	s.helper.GetLogger().Info(fmt.Sprintf("管理员账户创建完成: %s\n", username))
	return nil
}

func (s *InitInstallService) CreateInstallLock() error {
	lockContent := fmt.Sprintf("系统已安装\n安装时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	if err := os.MkdirAll(filepath.Dir(lockFile), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 (%s): %v", filepath.Dir(lockFile), err)
	}
	if err := os.WriteFile(lockFile, []byte(lockContent), 0644); err != nil {
		return fmt.Errorf("安装标记文件创建失败 (%s): %v", lockFile, err)
	}
	if abs, err := filepath.Abs(lockFile); err == nil {
		s.helper.GetLogger().Info("[install] 安装锁已写入: " + abs)
	}
	return nil
}

func (s *InitInstallService) CreateConfigFile(db InstallDatabaseConfig, rd InstallRedisConfig, cacheDriver, idGeneratorDriver string) error {
	db, err := NormalizeInstallDatabaseConfig(db)
	if err != nil {
		return err
	}
	content, err := renderInstallConfigFile(db, rd, cacheDriver, idGeneratorDriver, time.Now())
	if err != nil {
		return err
	}

	// Ensure parent directory exists — common cause of silent failures when the
	// binary is launched from a fresh checkout without a config/ folder.
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 (%s): %v", filepath.Dir(configFile), err)
	}
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		return fmt.Errorf("配置文件写入失败 (%s): %v", configFile, err)
	}
	if abs, err := filepath.Abs(configFile); err == nil {
		s.helper.GetLogger().Info("[install] 配置文件已写入: " + abs)
	}
	return nil
}

func (s *InitInstallService) TestDatabaseConnection(cfg InstallDatabaseConfig, maxRetries int) error {
	var err error
	cfg, err = NormalizeInstallDatabaseConfig(cfg)
	if err != nil {
		return err
	}

	// 复用 openInstallGormDB —— 走 GORM 各 driver 的 init 注册路径，避免直接
	// 调 sql.Open("postgres", ...) 时因没有 _ "github.com/lib/pq" /
	// _ "github.com/jackc/pgx/v5/stdlib" blank import 触发
	// `sql: unknown driver "postgres" (forgotten import?)`。GORM 的 postgres
	// driver 直接用 pgx，不会自动注册 database/sql 名下的 "postgres" 驱动。

	retryInterval := 2 * time.Second
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := pingInstallDB(cfg)
		if err == nil {
			s.helper.GetLogger().Info("数据库连接测试成功")
			return nil
		}
		lastErr = err
		s.helper.GetLogger().Warn(fmt.Sprintf("数据库连接测试失败 (第%d次重试): %v", i+1, err))
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}
	return fmt.Errorf("数据库连接测试失败 (已重试%d次): %v", maxRetries, lastErr)
}

func (s *InitInstallService) TestRedisConnection(cfg InstallRedisConfig, maxRetries int) error {
	if cfg.Host == "" {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	defer rdb.Close()

	ctx := context.Background()
	retryInterval := 2 * time.Second
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if _, err := rdb.Ping(ctx).Result(); err != nil {
			lastErr = err
			s.helper.GetLogger().Warn(fmt.Sprintf("Redis连接测试失败 (第%d次重试): %v", i+1, err))
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
			}
			continue
		}
		s.helper.GetLogger().Info("Redis连接测试成功")
		return nil
	}
	return fmt.Errorf("redis连接测试失败 (已重试%d次): %v", maxRetries, lastErr)
}

func generateInstallSecret(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// 迁移方言名(goose 识别的)与对应 migrations/<dir> 子目录。与
// pkg/service/migration/migration.go 的映射保持一致,但这里独立维护,避免
// 包循环;memory 驱动不在安装向导里走此路径。
var installDialectMap = map[string]string{
	"mysql":      "mysql",
	"postgresql": "postgres",
	"sqlite":     "sqlite3",
}

var installMigrationDirMap = map[string]string{
	"mysql":      "mysql",
	"postgresql": "postgresql",
	"sqlite":     "sqlite",
}

// RunMigrationsAndSeed 在安装请求当前进程内直接完成建表 + 初始管理员创建。
//
// 设计动机:原流程把实际建表推迟到 syscall.Exec 之后由 Migration server 接手,
// 但 go-web 的 initializeServices 会静默吞掉 server.Run() 的错误,一旦重启或
// assembly 重装失败,用户看到 HTTP 正常但数据库空空;改为在这里同步执行后,
// 失败可以直接通过 HTTP 响应返回给前端,状态与 install.lock 的写入保持一致。
func (s *InitInstallService) RunMigrationsAndSeed(cfg InstallDatabaseConfig, admin install_bootstrap.AdminPayload) error {
	var err error
	cfg, err = NormalizeInstallDatabaseConfig(cfg)
	if err != nil {
		return err
	}
	db, err := openInstallGormDB(cfg)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	defer func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取 sql.DB 失败: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库 ping 失败: %w", err)
	}

	dialect, ok := installDialectMap[cfg.Driver]
	if !ok {
		return fmt.Errorf("不支持的迁移方言: %s", cfg.Driver)
	}
	dir, ok := installMigrationDirMap[cfg.Driver]
	if !ok {
		return fmt.Errorf("缺少 %s 驱动的迁移目录映射", cfg.Driver)
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("设置方言失败: %w", err)
	}
	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(sqlDB, dir); err != nil {
		return fmt.Errorf("执行迁移失败: %w", err)
	}
	s.helper.GetLogger().Info(fmt.Sprintf("[install] 数据库迁移完成 (driver=%s, dir=%s)", cfg.Driver, dir))

	if err := seedInstallAdmin(db, admin); err != nil {
		return fmt.Errorf("创建管理员失败: %w", err)
	}
	s.helper.GetLogger().Info("[install] 管理员账户已创建: " + admin.Username)

	// bootstrap 文件若存在则清理,避免下次启动时 Migration.Consume() 二次创建
	// 引起唯一键冲突告警。Install 控制器已不再写入该文件,这里只是防御。
	if err := os.Remove(install_bootstrap.AdminFile); err != nil && !os.IsNotExist(err) {
		s.helper.GetLogger().Warn("[install] 清理 admin bootstrap 文件失败: " + err.Error())
	}
	return nil
}

// pingInstallDB 用 GORM 打开一次连接，立即 ping，再关掉。给 TestDatabaseConnection
// 复用，避开 sql.Open("postgres", ...) 找不到 driver 的坑。
func pingInstallDB(cfg InstallDatabaseConfig) error {
	gormDB, err := openInstallGormDB(cfg)
	if err != nil {
		return err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Second * 10)
	return sqlDB.Ping()
}

func openInstallGormDB(cfg InstallDatabaseConfig) (*gorm.DB, error) {
	var err error
	cfg, err = NormalizeInstallDatabaseConfig(cfg)
	if err != nil {
		return nil, err
	}
	driverCfg := toGoWebDatabaseConfig(cfg)
	db, err := dbDriver.DatabaseDriverManager.Make(driverCfg.Driver, driverCfg)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func toGoWebDatabaseConfig(cfg InstallDatabaseConfig) *dbconfig.DatabaseConfig {
	host := cfg.Host
	if cfg.Driver == "sqlite" {
		host = cfg.Filepath
	}
	return &dbconfig.DatabaseConfig{
		Driver:   cfg.Driver,
		Host:     host,
		Port:     cfg.Port,
		DBName:   cfg.DBName,
		Username: cfg.Username,
		Password: cfg.Password,
	}
}

func NormalizeInstallDatabaseConfig(cfg InstallDatabaseConfig) (InstallDatabaseConfig, error) {
	cfg.Driver = strings.ToLower(strings.TrimSpace(cfg.Driver))
	switch cfg.Driver {
	case "mysql", "postgresql":
		host, err := dbconfig.NormalizeTCPHost(cfg.Host)
		if err != nil {
			return cfg, err
		}
		if err := dbconfig.ValidateTCPPort(cfg.Port); err != nil {
			return cfg, err
		}
		if err := validateInstallIdentifier("数据库名", cfg.DBName); err != nil {
			return cfg, err
		}
		if strings.TrimSpace(cfg.Username) == "" {
			return cfg, fmt.Errorf("数据库用户名不能为空")
		}
		cfg.Host = host
	case "sqlite":
		cfg.Filepath = strings.TrimSpace(cfg.Filepath)
		if cfg.Filepath == "" {
			return cfg, fmt.Errorf("SQLite 数据库文件不能为空")
		}
		if hasControlChar(cfg.Filepath) {
			return cfg, fmt.Errorf("SQLite 数据库文件路径包含控制字符")
		}
	default:
		return cfg, fmt.Errorf("不支持的数据库类型: %s", cfg.Driver)
	}
	return cfg, nil
}

func validateInstallIdentifier(label, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s不能为空", label)
	}
	if hasControlChar(value) {
		return fmt.Errorf("%s包含控制字符", label)
	}
	if strings.ContainsAny(value, "?&=#/\\") {
		return fmt.Errorf("%s不能包含连接串分隔符", label)
	}
	return nil
}

func hasControlChar(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

type installConfigFile struct {
	Server      installServerConfig `yaml:"server"`
	Database    installDatabaseFile `yaml:"database"`
	Redis       *installRedisFile   `yaml:"redis,omitempty"`
	JWT         installJWTConfig    `yaml:"jwt"`
	Logger      installLoggerConfig `yaml:"logger"`
	Cache       installDriverConfig `yaml:"cache"`
	IDGenerator installDriverConfig `yaml:"id_generator"`
}

type installServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type installDatabaseFile struct {
	Driver          string `yaml:"driver"`
	Filepath        string `yaml:"filepath,omitempty"`
	Host            string `yaml:"host,omitempty"`
	Port            int    `yaml:"port,omitempty"`
	DBName          string `yaml:"dbname,omitempty"`
	Username        string `yaml:"username,omitempty"`
	Password        string `yaml:"password,omitempty"`
	Charset         string `yaml:"charset"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

type installRedisFile struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Password    string `yaml:"password"`
	DB          int    `yaml:"db"`
	MaxIdle     int    `yaml:"max_idle"`
	MaxActive   int    `yaml:"max_active"`
	IdleTimeout string `yaml:"idle_timeout"`
}

type installJWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type installLoggerConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
}

type installDriverConfig struct {
	Driver string `yaml:"driver"`
}

func renderInstallConfigFile(db InstallDatabaseConfig, rd InstallRedisConfig, cacheDriver, idGeneratorDriver string, generatedAt time.Time) ([]byte, error) {
	cfg := installConfigFile{
		Server: installServerConfig{
			Port: 8080,
			Mode: "release",
		},
		Database: installDatabaseFile{
			Driver:          db.Driver,
			Charset:         "utf8mb4",
			MaxOpenConns:    100,
			MaxIdleConns:    20,
			ConnMaxLifetime: "300s",
		},
		JWT: installJWTConfig{
			Secret:      generateInstallSecret(32),
			ExpireHours: 24,
		},
		Logger: installLoggerConfig{
			Level:      "info",
			File:       "logs/app.log",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 10,
		},
		Cache:       installDriverConfig{Driver: cacheDriver},
		IDGenerator: installDriverConfig{Driver: idGeneratorDriver},
	}
	if db.Driver == "sqlite" {
		cfg.Database.Filepath = db.Filepath
	} else {
		cfg.Database.Host = db.Host
		cfg.Database.Port = db.Port
		cfg.Database.DBName = db.DBName
		cfg.Database.Username = db.Username
		cfg.Database.Password = db.Password
	}
	if cacheDriver == "redis" || idGeneratorDriver == "redis" {
		cfg.Redis = &installRedisFile{
			Host:        rd.Host,
			Port:        rd.Port,
			Password:    rd.Password,
			DB:          rd.DB,
			MaxIdle:     10,
			MaxActive:   100,
			IdleTimeout: "300s",
		}
	}
	body, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("生成配置文件失败: %w", err)
	}
	header := fmt.Sprintf("# 短网址服务配置文件\n# 由安装向导自动生成于 %s\n\n", generatedAt.Format("2006-01-02 15:04:05"))
	return append([]byte(header), body...), nil
}

func seedInstallAdmin(db *gorm.DB, admin install_bootstrap.AdminPayload) error {
	user := &model.User{
		Username:      admin.Username,
		Email:         admin.Email,
		Status:        1,
		IsSystemAdmin: true,
	}
	if err := user.SetPassword(admin.Password); err != nil {
		return err
	}
	if err := db.Create(user).Error; err != nil {
		return err
	}
	if err := db.Model(&model.Workspace{}).Where("id = ?", 1).Update("owner_user_id", user.ID).Error; err != nil {
		return err
	}
	return db.Create(&model.WorkspaceMember{
		WorkspaceID: 1,
		UserID:      user.ID,
		Role:        model.WorkspaceRoleOwner,
		Status:      1,
	}).Error
}
