package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/migrations"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/install_bootstrap"
	"github.com/glebarez/sqlite"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
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
		Username: username,
		RealName: realName,
		Email:    email,
		Phone:    phone,
		Status:   1,
	}
	if err := user.SetPassword(password); err != nil {
		return err
	}
	userDAO := dao.NewUserDAO(s.helper)
	if err := userDAO.Create(user); err != nil {
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
	content := fmt.Sprintf(`# 短网址服务配置文件
# 由安装向导自动生成于 %s

# 服务配置
server:
  port: 8080
  mode: release

# 数据库配置
database:
  driver: %s`,
		time.Now().Format("2006-01-02 15:04:05"),
		db.Driver)

	if db.Driver == "sqlite" {
		content += fmt.Sprintf(`
  filepath: %s`, db.Filepath)
	} else {
		content += fmt.Sprintf(`
  host: %s
  port: %d
  dbname: %s
  username: %s
  password: %s`, db.Host, db.Port, db.DBName, db.Username, db.Password)
	}

	content += `
  charset: utf8mb4
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 300s`

	needsRedis := cacheDriver == "redis" || idGeneratorDriver == "redis"
	if needsRedis {
		content += fmt.Sprintf(`

# Redis配置
redis:
  host: %s
  port: %d
  password: %s
  db: %d
  max_idle: 10
  max_active: 100
  idle_timeout: 300s`, rd.Host, rd.Port, rd.Password, rd.DB)
	}

	jwtSecret := generateInstallSecret(32)
	content += fmt.Sprintf(`

# JWT配置
jwt:
  secret: %s
  expire_hours: 24

# 日志配置
logger:
  level: info
  file: logs/app.log
  max_size: 100
  max_age: 7
  max_backups: 10

# 缓存配置
cache:
  driver: %s

# ID生成器配置
id_generator:
  driver: %s
`, jwtSecret, cacheDriver, idGeneratorDriver)

	// Ensure parent directory exists — common cause of silent failures when the
	// binary is launched from a fresh checkout without a config/ folder.
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 (%s): %v", filepath.Dir(configFile), err)
	}
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("配置文件写入失败 (%s): %v", configFile, err)
	}
	if abs, err := filepath.Abs(configFile); err == nil {
		s.helper.GetLogger().Info("[install] 配置文件已写入: " + abs)
	}
	return nil
}

func (s *InitInstallService) TestDatabaseConnection(cfg InstallDatabaseConfig, maxRetries int) error {
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
	switch cfg.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgresql":
		dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	case "sqlite":
		return gorm.Open(sqlite.Open(cfg.Filepath), &gorm.Config{})
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Driver)
	}
}

func seedInstallAdmin(db *gorm.DB, admin install_bootstrap.AdminPayload) error {
	user := &model.User{
		Username: admin.Username,
		Email:    admin.Email,
		Status:   1,
	}
	if err := user.SetPassword(admin.Password); err != nil {
		return err
	}
	return db.Create(user).Error
}
