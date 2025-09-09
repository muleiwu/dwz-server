package controller

import (
	"net/http"

	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	database2 "cnb.cool/mliev/open/dwz-server/internal/pkg/database/config"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/database/impl"
	redis2 "cnb.cool/mliev/open/dwz-server/internal/pkg/redis/config"
	impl2 "cnb.cool/mliev/open/dwz-server/internal/pkg/redis/impl"
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
	Copyright      string
	DatabaseConfig database2.DatabaseConfig
	RedisConfig    redis2.RedisConfig
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
func (receiver InstallController) GetDefaultDatabaseConfig(helper interfaces.HelperInterface) DatabaseConfig {

	// 从环境变量获取数据库类型，默认mysql
	dbType := helper.GetConfig().GetString("database.type", "mysql")
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
		Host:     helper.GetConfig().GetString("database.host", "localhost"),
		Port:     helper.GetConfig().GetInt("database.port", defaultPort),
		Name:     helper.GetConfig().GetString("database.dbname", "dwz"),
		User:     helper.GetConfig().GetString("database.username", "dwz"),
		Password: helper.GetConfig().GetString("database.password", "dwz"),
	}
}

// GetDefaultRedisConfig 从环境变量获取默认Redis配置
func (receiver InstallController) GetDefaultRedisConfig(helper interfaces.HelperInterface) RedisConfig {

	return RedisConfig{
		Host:     helper.GetConfig().GetString("redis.host", "localhost"),
		Port:     helper.GetConfig().GetInt("redis.port", 6379),
		Password: helper.GetConfig().GetString("redis.password", ""),
		DB:       helper.GetConfig().GetInt("redis.db", 0),
	}
}

// GetInstall 显示安装页面
func (receiver InstallController) GetInstall(c *gin.Context, helper interfaces.HelperInterface) {
	// 检查是否已经安装
	if helper.GetInstalled().IsInstalled() {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// 获取当前访问的域名
	host := c.Request.Host

	// 从环境变量获取默认配置
	defaultDbConfig := receiver.GetDefaultDatabaseConfig(helper)
	defaultRedisConfig := receiver.GetDefaultRedisConfig(helper)
	siteName := helper.GetConfig().GetString("website.name", "短网址服务")
	copyright := helper.GetConfig().GetString("website.copyright", "")
	// 构造页面数据，将配置结构映射到页面需要的结构
	pageData := InstallPageData{
		SiteName:     siteName,
		ICPNumber:    "",
		PoliceNumber: "",
		Domain:       host,
		Copyright:    copyright,
		DatabaseConfig: database2.DatabaseConfig{
			Driver:   defaultDbConfig.Type,
			Host:     defaultDbConfig.Host,
			Port:     defaultDbConfig.Port,
			DBName:   defaultDbConfig.Name,
			Username: defaultDbConfig.User,
			Password: defaultDbConfig.Password,
		},
		RedisConfig: redis2.RedisConfig{
			Host:     defaultRedisConfig.Host,
			Port:     defaultRedisConfig.Port,
			Password: defaultRedisConfig.Password,
			DB:       defaultRedisConfig.DB,
		},
	}

	c.HTML(http.StatusOK, "install.html", pageData)
}

// TestConnection 测试数据库和Redis连接
func (receiver InstallController) TestConnection(c *gin.Context, helper interfaces.HelperInterface) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		receiver.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 测试数据库连接
	if err := receiver.testDatabaseConnection(req.Database, helper); err != nil {
		receiver.Error(c, http.StatusBadRequest, "数据库连接失败: "+err.Error())
		return
	}

	// 测试Redis连接
	if err := receiver.testRedisConnection(req.Redis, helper); err != nil {
		receiver.Error(c, http.StatusBadRequest, "Redis连接失败: "+err.Error())
		return
	}

	receiver.SuccessWithMessage(c, "连接测试成功", nil)
}

// Install 执行安装
func (receiver InstallController) Install(c *gin.Context, helper interfaces.HelperInterface) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		receiver.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 检查是否已经安装
	if helper.GetInstalled().IsInstalled() {
		receiver.Error(c, http.StatusBadRequest, "系统已经安装")
		return
	}

	// 再次测试连接
	if err := receiver.testDatabaseConnection(req.Database, helper); err != nil {
		receiver.Error(c, http.StatusBadRequest, "数据库连接失败: "+err.Error())
		return
	}

	if err := receiver.testRedisConnection(req.Redis, helper); err != nil {
		receiver.Error(c, http.StatusBadRequest, "Redis连接失败: "+err.Error())
		return
	}

	// 创建配置文件
	if err := receiver.createConfigFile(req, helper); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "配置文件创建失败: "+err.Error())
		return
	}

	// 初始化数据库
	if err := receiver.initializeDatabase(req.Database, req.Redis, helper); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "数据库初始化失败: "+err.Error())
		return
	}

	// 创建管理员账户
	if err := receiver.createAdminUser(req.Admin, helper); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "管理员账户创建失败: "+err.Error())
		return
	}

	// 创建安装标记文件
	if err := receiver.createInstallLock(helper); err != nil {
		receiver.Error(c, http.StatusInternalServerError, "安装标记创建失败: "+err.Error())
		return
	}

	// 标记系统为已安装
	helper.GetInstalled().SetInstalled(true)

	receiver.SuccessWithMessage(c, "安装完成", nil)
}

// testDatabaseConnection 测试数据库连接
func (receiver InstallController) testDatabaseConnection(config DatabaseConfig, helper interfaces.HelperInterface) error {
	installService := service.NewInitInstallService(helper)

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
func (receiver InstallController) testRedisConnection(config RedisConfig, helper interfaces.HelperInterface) error {
	installService := service.NewInitInstallService(helper)

	redisConfig := redis2.RedisConfig{
		Host:     config.Host,
		Port:     config.Port,
		Password: config.Password,
		DB:       config.DB,
	}
	return installService.TestRedisConnection(redisConfig)
}

// createConfigFile 创建配置文件
func (receiver InstallController) createConfigFile(req InstallRequest, helper interfaces.HelperInterface) error {

	databaseConfig := database2.DatabaseConfig{
		Driver:   req.Database.Type,
		Username: req.Database.User,
		Password: req.Database.Password,
		Host:     req.Database.Host,
		Port:     req.Database.Port,
		DBName:   req.Database.Name,
	}

	redisConfig := redis2.RedisConfig{
		Host:     req.Redis.Host,
		Port:     req.Redis.Port,
		Password: req.Redis.Password,
		DB:       req.Redis.DB,
	}
	installService := service.NewInitInstallService(helper)
	return installService.CreateConfigFile(databaseConfig, redisConfig)
}

// initializeDatabase 初始化数据库
func (receiver InstallController) initializeDatabase(databaseConfig DatabaseConfig, redisConfig RedisConfig, helper interfaces.HelperInterface) error {

	database, err := impl.NewDatabase(helper, databaseConfig.Type, databaseConfig.Host, databaseConfig.Port, databaseConfig.Name, databaseConfig.User, databaseConfig.Password)
	if err != nil {
		return err
	}

	helper.SetDatabase(database)

	redis, err := impl2.NewRedis(helper, redisConfig.Host, redisConfig.Port, redisConfig.DB, redisConfig.Password)

	if err != nil {
		return err
	}
	helper.SetRedis(redis)

	migration := helper.GetConfig().Get("database.migration", []any{}).([]any)

	err = helper.GetDatabase().AutoMigrate(migration...)
	if err != nil {
		return err
	}

	return nil
}

// createAdminUser 创建管理员账户
func (receiver InstallController) createAdminUser(config AdminConfig, helper interfaces.HelperInterface) error {
	installService := service.NewInitInstallService(helper)
	return installService.CreateAdminUser(config.Username, "", config.Email, "", config.Password)
}

// createInstallLock 创建安装标记文件
func (receiver InstallController) createInstallLock(helper interfaces.HelperInterface) error {
	installService := service.NewInitInstallService(helper)
	return installService.CreateInstallLock()
}
