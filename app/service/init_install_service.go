package service

import (
	"cnb.cool/mliev/open/dwz-server/helper/install"
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/config"
	"cnb.cool/mliev/open/dwz-server/config/database"
	database2 "cnb.cool/mliev/open/dwz-server/config/database"
	redis3 "cnb.cool/mliev/open/dwz-server/config/redis"
	database1 "cnb.cool/mliev/open/dwz-server/helper/database"
	"cnb.cool/mliev/open/dwz-server/helper/logger"
	"github.com/redis/go-redis/v9"
)

type InitInstallService struct {
}

const lockFile = "./config/install.lock"
const configFile = "./config/config.yaml"

func NewInitInstallService() *InitInstallService {
	return &InitInstallService{}
}

func (receiver *InitInstallService) AutoInstall() {
	// 自动安装流程
	databaseConfig := database2.DatabaseConfig{
		Driver:   config.GetString("database.driver", ""),
		Host:     config.GetString("database.host", ""),
		Port:     config.GetInt("database.port", 0),
		DBName:   config.GetString("database.dbname", ""),
		Username: config.GetString("database.username", ""),
		Password: config.GetString("database.password", ""),
	}

	redisConfig := redis3.RedisConfig{
		Host:     config.GetString("redis.host", ""),
		Port:     config.GetInt("redis.port", 0),
		Password: config.GetString("redis.password", ""),
		DB:       config.GetInt("redis.db", 0),
	}

	err := receiver.TestDatabaseConnection(databaseConfig)
	if err != nil {
		logger.Logger().Error(fmt.Sprintf("[自动安装] 检查数据库连接失败, 原因: %s", err.Error()))
		os.Exit(1)
	}

	err = receiver.TestRedisConnection(redisConfig)
	if err != nil {
		logger.Logger().Error(fmt.Sprintf("[自动安装] 检查Redis连接失败, 原因: %s", err.Error()))
		os.Exit(1)
	}

	err = receiver.CreateConfigFile(databaseConfig, redisConfig)
	if err != nil {
		logger.Logger().Error(fmt.Sprintf("[自动安装] 写入配置文件失败, 原因: %s", err.Error()))
		os.Exit(1)
	}

	// 未安装，且配置自动初始化
	err = database1.AutoMigrate()
	if err != nil {
		logger.Logger().Error(fmt.Sprintf("[自动安装] 数据库迁移失败, 原因: %s", err.Error()))
		os.Exit(1)
	}

	err = receiver.CreateInstallLock()

	if err != nil {
		logger.Logger().Error(fmt.Sprintf("[自动安装] 写入安装成功锁定文件失败, 原因: %s", err.Error()))
		os.Exit(1)
	}

	err = receiver.CreateAdminUser("admin", "admin", "system@system.local", "", "admin")

	if err != nil {
		logger.Logger().Error(fmt.Sprintf("[自动安装] 自动添默认用户失败, 原因: %s", err.Error()))
		os.Exit(1)
	}

	// 标记系统为已安装
	install.MarkAsInstalled()

	logger.Logger().Info(fmt.Sprintf("【自动安装】成功， 用户名：admin 密码：admin"))
	logger.Logger().Info(fmt.Sprintf("【自动安装】成功， 请打开系统后，立刻修改密码！！！"))
	logger.Logger().Info(fmt.Sprintf("【自动安装】成功， 请打开系统后，立刻修改密码！！！"))
	logger.Logger().Info(fmt.Sprintf("【自动安装】成功， 请打开系统后，立刻修改密码！！！"))
}

func (receiver *InitInstallService) CreateAdminUser(username, realName, email, phone, password string) error {
	// 由于安装过程中数据库连接可能尚未初始化，这里简化处理
	// 实际实现中应该直接使用数据库连接创建用户

	// 创建用户
	user := &model.User{
		Username: username,
		RealName: realName,
		Email:    email,
		Phone:    phone,
		Status:   1, // 默认启用
	}

	// 设置密码
	if err := user.SetPassword(password); err != nil {
		return err
	}

	userDAO := dao.NewUserDAO()
	// 保存到数据库
	if err := userDAO.Create(user); err != nil {
		return err
	}

	logger.Logger().Info(fmt.Sprintf("管理员账户创建完成: %s\n", username))

	return nil
}

func (receiver *InitInstallService) CreateInstallLock() error {

	lockContent := fmt.Sprintf("系统已安装\n安装时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(lockFile, []byte(lockContent), 0644); err != nil {
		return fmt.Errorf("安装标记文件创建失败: %v", err)
	}

	return nil
}

func (receiver *InitInstallService) CreateConfigFile(databaseInfo database.DatabaseConfig, redisInfo redis3.RedisConfig) error {

	configContent := fmt.Sprintf(`# 短网址服务配置文件
# 由安装向导自动生成于 %s

# 服务配置
server:
  port: 8080
  mode: release

# 数据库配置
database:
  type: %s
  host: %s
  port: %d
  dbname: %s
  username: %s
  password: %s
  charset: utf8mb4
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 300s

# Redis配置
redis:
  host: %s
  port: %d
  password: %s
  db: %d
  max_idle: 10
  max_active: 100
  idle_timeout: 300s

# 日志配置
logger:
  level: info
  file: logs/app.log
  max_size: 100
  max_age: 7
  max_backups: 10

# 中间件配置
middleware:
  cors:
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["*"]
    expose_headers: ["*"]
    allow_credentials: true
    max_age: 12h

# 操作日志配置
operation_log:
  enabled: true
  max_age: 30d
  batch_size: 100
  flush_interval: 10s

# 迁移配置
migration:
  enabled: true
  auto_migrate: true
`,
		time.Now().Format("2006-01-02 15:04:05"),
		databaseInfo.Driver,
		databaseInfo.Host,
		databaseInfo.Port,
		databaseInfo.DBName,
		databaseInfo.Username,
		databaseInfo.Password,
		redisInfo.Host,
		redisInfo.Port,
		redisInfo.Password,
		redisInfo.DB,
	)

	// 写入配置文件
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("配置文件写入失败: %v", err)
	}

	return nil
}

func (receiver *InitInstallService) TestDatabaseConnection(config database.DatabaseConfig) error {
	var dsn string
	var driver string

	switch config.Driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.DBName)
		driver = "mysql"
	case "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.Username, config.Password, config.DBName)
		driver = "postgres"
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.Driver)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 设置连接池参数
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Second * 10)

	// 测试连接 - 添加重试机制
	maxRetries := 60
	retryInterval := 2 * time.Second
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			lastErr = err
			logger.Logger().Warn(fmt.Sprintf("数据库连接测试失败 (第%d次重试): %v", i+1, err))

			// 如果不是最后一次重试，等待后重试
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
			}
		} else {
			// 连接成功
			logger.Logger().Info("数据库连接测试成功")
			return nil
		}
	}

	return fmt.Errorf("数据库连接测试失败 (已重试%d次): %v", maxRetries, lastErr)
}

func (receiver *InitInstallService) TestRedisConnection(config redis3.RedisConfig) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	defer rdb.Close()

	ctx := context.Background()

	// 添加重试机制
	maxRetries := 60
	retryInterval := 2 * time.Second
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			lastErr = err
			logger.Logger().Warn(fmt.Sprintf("Redis连接测试失败 (第%d次重试): %v", i+1, err))

			// 如果不是最后一次重试，等待后重试
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
			}
		} else {
			// 连接成功
			logger.Logger().Info("Redis连接测试成功")
			return nil
		}
	}

	return fmt.Errorf("redis连接测试失败 (已重试%d次): %v", maxRetries, lastErr)
}
