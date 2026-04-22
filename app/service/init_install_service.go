package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/app/model"
	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	_ "github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
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
	if err := os.WriteFile(lockFile, []byte(lockContent), 0644); err != nil {
		return fmt.Errorf("安装标记文件创建失败: %v", err)
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

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("配置文件写入失败: %v", err)
	}
	return nil
}

func (s *InitInstallService) TestDatabaseConnection(cfg InstallDatabaseConfig, maxRetries int) error {
	var dsn, driver string
	switch cfg.Driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		driver = "mysql"
	case "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName)
		driver = "postgres"
	case "sqlite":
		dsn = cfg.Filepath
		driver = "sqlite"
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Driver)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Second * 10)

	retryInterval := 2 * time.Second
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			lastErr = err
			s.helper.GetLogger().Warn(fmt.Sprintf("数据库连接测试失败 (第%d次重试): %v", i+1, err))
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
			}
			continue
		}
		s.helper.GetLogger().Info("数据库连接测试成功")
		return nil
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
