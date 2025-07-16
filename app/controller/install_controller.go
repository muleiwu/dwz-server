package controller

import (
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/config"
	database2 "cnb.cool/mliev/open/dwz-server/config/database"
	redis3 "cnb.cool/mliev/open/dwz-server/config/redis"
	"cnb.cool/mliev/open/dwz-server/helper/database"
	"cnb.cool/mliev/open/dwz-server/helper/env"
	"cnb.cool/mliev/open/dwz-server/helper/logger"
	redis2 "cnb.cool/mliev/open/dwz-server/helper/redis"
	"net/http"

	"cnb.cool/mliev/open/dwz-server/helper/install"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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
	installService := service.NewInitInstallService()

	databaseConfig := database2.DatabaseConfig{
		Driver:   config.Type,
		Username: config.User,
		Password: config.Password,
		Host:     config.Host,
		Port:     config.Port,
		DBName:   config.Name,
	}

	return installService.TestDatabaseConnection(databaseConfig)
}

// testRedisConnection 测试Redis连接
func (receiver InstallController) testRedisConnection(config RedisConfig) error {
	installService := service.NewInitInstallService()

	redisConfig := redis3.RedisConfig{
		Host:     config.Host,
		Port:     config.Port,
		Password: config.Password,
		DB:       config.DB,
	}
	return installService.TestRedisConnection(redisConfig)
}

// createConfigFile 创建配置文件
func (receiver InstallController) createConfigFile(req InstallRequest) error {

	databaseConfig := database2.DatabaseConfig{
		Driver:   req.Database.Type,
		Username: req.Database.User,
		Password: req.Database.Password,
		Host:     req.Database.Host,
		Port:     req.Database.Port,
		DBName:   req.Database.Name,
	}

	redisConfig := redis3.RedisConfig{
		Host:     req.Redis.Host,
		Port:     req.Redis.Port,
		Password: req.Redis.Password,
		DB:       req.Redis.DB,
	}
	installService := service.NewInitInstallService()
	return installService.CreateConfigFile(databaseConfig, redisConfig)
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
	installService := service.NewInitInstallService()
	return installService.CreateAdminUser(config.Username, "", config.Email, "", config.Password)
}

// createInstallLock 创建安装标记文件
func (receiver InstallController) createInstallLock() error {
	installService := service.NewInitInstallService()
	return installService.CreateInstallLock()
}
