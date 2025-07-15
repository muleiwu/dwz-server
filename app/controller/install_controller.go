package controller

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/config"
	database2 "cnb.cool/mliev/open/dwz-server/config/database"
	redis3 "cnb.cool/mliev/open/dwz-server/config/redis"
	"cnb.cool/mliev/open/dwz-server/helper/database"
	"cnb.cool/mliev/open/dwz-server/helper/env"
	"cnb.cool/mliev/open/dwz-server/helper/logger"
	redis2 "cnb.cool/mliev/open/dwz-server/helper/redis"

	"cnb.cool/mliev/open/dwz-server/helper/install"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type InstallController struct {
	BaseResponse
}

// InstallPageData 安装页面数据结构
type InstallPageData struct {
	SiteName       string
	ICPNumber      string
	PoliceNumber   string
	Domain         string
	DatabaseConfig database2.DatabaseConfig
	RedisConfig    redis3.RedisConfig
}

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Type     string `json:"type" binding:"required"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Name     string `json:"name" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RedisConfig Redis配置结构
type RedisConfig struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// AdminConfig 管理员配置结构
type AdminConfig struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

// InstallRequest 安装请求结构
type InstallRequest struct {
	Database DatabaseConfig `json:"database" binding:"required"`
	Redis    RedisConfig    `json:"redis" binding:"required"`
	Admin    AdminConfig    `json:"admin" binding:"required"`
}

// TestConnectionRequest 测试连接请求结构
type TestConnectionRequest struct {
	Database DatabaseConfig `json:"database" binding:"required"`
	Redis    RedisConfig    `json:"redis" binding:"required"`
}

// GetDefaultDatabaseConfig 从环境变量获取默认数据库配置
func (receiver InstallController) GetDefaultDatabaseConfig() DatabaseConfig {
	// 初始化配置以支持环境变量读取
	err := config.InitViper()

	if err != nil {
		logger.Logger().Error(err.Error())
	}

	// 从环境变量获取数据库类型，默认mysql
	dbType := config.GetString("database.type", "mysql")
	if dbType == "postgres" {
		dbType = "postgresql"
	}

	// 根据数据库类型设置默认端口
	defaultPort := 3306
	if dbType == "postgresql" {
		defaultPort = 5432
	}

	return DatabaseConfig{
		Type:     dbType,
		Host:     config.GetString("database.host", "localhost"),
		Port:     config.GetInt("database.port", defaultPort),
		Name:     config.GetString("database.dbname", "dwz"),
		User:     config.GetString("database.username", "dwz"),
		Password: config.GetString("database.password", "dwz"),
	}
}

// GetDefaultRedisConfig 从环境变量获取默认Redis配置
func (receiver InstallController) GetDefaultRedisConfig() RedisConfig {
	// 初始化配置以支持环境变量读取
	err := config.InitViper()

	if err != nil {
		logger.Logger().Error(err.Error())
	}

	return RedisConfig{
		Host:     config.GetString("redis.host", "localhost"),
		Port:     config.GetInt("redis.port", 6379),
		Password: config.GetString("redis.password", ""),
		DB:       config.GetInt("redis.db", 0),
	}
}

// GetInstall 显示安装页面
func (receiver InstallController) GetInstall(c *gin.Context) {
	// 检查是否已经安装
	if install.IsInstalled() {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// 获取当前访问的域名
	host := c.Request.Host

	// 从环境变量获取默认配置
	defaultDbConfig := receiver.GetDefaultDatabaseConfig()
	defaultRedisConfig := receiver.GetDefaultRedisConfig()

	// 构造页面数据，将配置结构映射到页面需要的结构
	pageData := InstallPageData{
		SiteName:     "短网址服务",
		ICPNumber:    "",
		PoliceNumber: "",
		Domain:       host,
		DatabaseConfig: database2.DatabaseConfig{
			Driver:   defaultDbConfig.Type,
			Host:     defaultDbConfig.Host,
			Port:     defaultDbConfig.Port,
			DBName:   defaultDbConfig.Name,
			Username: defaultDbConfig.User,
			Password: defaultDbConfig.Password,
		},
		RedisConfig: redis3.RedisConfig{
			Host:     defaultRedisConfig.Host,
			Port:     defaultRedisConfig.Port,
			Password: defaultRedisConfig.Password,
			DB:       defaultRedisConfig.DB,
		},
	}

	c.HTML(http.StatusOK, "install.html", pageData)
}

// TestConnection 测试数据库和Redis连接
func (receiver InstallController) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		receiver.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 测试数据库连接
	if err := receiver.testDatabaseConnection(req.Database); err != nil {
		receiver.Error(c, http.StatusBadRequest, "数据库连接失败: "+err.Error())
		return
	}

	// 测试Redis连接
	if err := receiver.testRedisConnection(req.Redis); err != nil {
		receiver.Error(c, http.StatusBadRequest, "Redis连接失败: "+err.Error())
		return
	}

	receiver.SuccessWithMessage(c, "连接测试成功", nil)
}

// Install 执行安装
func (receiver InstallController) Install(c *gin.Context) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		receiver.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 检查是否已经安装
	if install.IsInstalled() {
		receiver.Error(c, http.StatusBadRequest, "系统已经安装")
		return
	}

	// 再次测试连接
	if err := receiver.testDatabaseConnection(req.Database); err != nil {
		receiver.Error(c, http.StatusBadRequest, "数据库连接失败: "+err.Error())
		return
	}

	if err := receiver.testRedisConnection(req.Redis); err != nil {
		receiver.Error(c, http.StatusBadRequest, "Redis连接失败: "+err.Error())
		return
	}

	// 创建配置文件
	if err := receiver.createConfigFile(req); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "配置文件创建失败: "+err.Error())
		return
	}

	// 初始化数据库
	if err := receiver.initializeDatabase(req.Database); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "数据库初始化失败: "+err.Error())
		return
	}

	// 创建管理员账户
	if err := receiver.createAdminUser(req.Admin); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "管理员账户创建失败: "+err.Error())
		return
	}

	// 创建安装标记文件
	if err := receiver.createInstallLock(); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "安装标记创建失败: "+err.Error())
		return
	}

	// 标记系统为已安装
	install.MarkAsInstalled()

	receiver.SuccessWithMessage(c, "安装完成", nil)
}

// testDatabaseConnection 测试数据库连接
func (receiver InstallController) testDatabaseConnection(config DatabaseConfig) error {
	var dsn string
	var driver string

	switch config.Type {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.Name)
		driver = "mysql"
	case "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.Name)
		driver = "postgres"
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.Type)
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

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	return nil
}

// testRedisConnection 测试Redis连接
func (receiver InstallController) testRedisConnection(config RedisConfig) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	defer rdb.Close()

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis连接测试失败: %s", err.Error())
	}

	return nil
}

// createConfigFile 创建配置文件
func (receiver InstallController) createConfigFile(req InstallRequest) error {
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
		req.Database.Type,
		req.Database.Host,
		req.Database.Port,
		req.Database.Name,
		req.Database.User,
		req.Database.Password,
		req.Redis.Host,
		req.Redis.Port,
		req.Redis.Password,
		req.Redis.DB,
	)

	// 写入配置文件
	if err := os.WriteFile("./config/config.yaml", []byte(configContent), 0644); err != nil {
		return fmt.Errorf("配置文件写入失败: %v", err)
	}

	return nil
}

// initializeDatabase 初始化数据库
func (receiver InstallController) initializeDatabase(config DatabaseConfig) error {
	err := env.ReloadViper()
	if err != nil {
		return err
	}

	database.GetDB()
	err = database.AutoMigrate()

	if err != nil {
		return err
	}

	// 初始化Redis连接
	redis2.GetRedis()

	return nil
}

// createAdminUser 创建管理员账户
func (receiver InstallController) createAdminUser(config AdminConfig) error {
	// 由于安装过程中数据库连接可能尚未初始化，这里简化处理
	// 实际实现中应该直接使用数据库连接创建用户

	// 创建用户
	user := &model.User{
		Username: config.Username,
		RealName: config.Username,
		Email:    config.Email,
		Phone:    "",
		Status:   1, // 默认启用
	}

	// 设置密码
	if err := user.SetPassword(config.Password); err != nil {
		return err
	}

	userDAO := dao.NewUserDAO()
	// 保存到数据库
	if err := userDAO.Create(user); err != nil {
		return err
	}

	logger.Logger().Info(fmt.Sprintf("管理员账户创建完成: %s\n", config.Username))

	return nil
}

// createInstallLock 创建安装标记文件
func (receiver InstallController) createInstallLock() error {
	lockFile := "./config/install.lock"
	lockContent := fmt.Sprintf("系统已安装\n安装时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(lockFile, []byte(lockContent), 0644); err != nil {
		return fmt.Errorf("安装标记文件创建失败: %v", err)
	}

	return nil
}
