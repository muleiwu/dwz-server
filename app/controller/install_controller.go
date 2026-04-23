package controller

import (
	"net/http"
	"os"
	"syscall"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/install_bootstrap"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type InstallController struct {
	BaseResponse
}

type installPageData struct {
	SiteName       string
	ICPNumber      string
	PoliceNumber   string
	Domain         string
	Copyright      string
	DatabaseConfig service.InstallDatabaseConfig
	RedisConfig    service.InstallRedisConfig
}

type installDatabasePayload struct {
	Type     string `json:"type" binding:"required"`
	Filepath string `json:"filepath"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type installRedisPayload struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type installAdminPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

type installRequest struct {
	Database          installDatabasePayload `json:"database" binding:"required"`
	Redis             *installRedisPayload   `json:"redis,omitempty"`
	Admin             installAdminPayload    `json:"admin" binding:"required"`
	CacheDriver       string                 `json:"cacheDriver" binding:"required,oneof=local redis none memory"`
	IDGeneratorDriver string                 `json:"idGeneratorDriver" binding:"required,oneof=local redis"`
}

type installTestRequest struct {
	Database          installDatabasePayload `json:"database" binding:"required"`
	Redis             *installRedisPayload   `json:"redis,omitempty"`
	CacheDriver       string                 `json:"cacheDriver" binding:"required,oneof=local redis none memory"`
	IDGeneratorDriver string                 `json:"idGeneratorDriver" binding:"required,oneof=local redis"`
}

func (ctrl InstallController) GetInstall(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	if installed := helper.GetInstalled(); installed != nil && installed.IsInstalled() {
		c.Redirect(http.StatusFound, "/")
		return
	}

	host := c.Host()
	env := helper.GetEnv()
	cfg := helper.GetConfig()

	pageData := installPageData{
		SiteName:     cfg.GetString("website.name", "短网址服务"),
		Domain:       host,
		Copyright:    cfg.GetString("website.copyright", ""),
		ICPNumber:    "",
		PoliceNumber: "",
		DatabaseConfig: service.InstallDatabaseConfig{
			Driver:   env.GetString("database.driver", cfg.GetString("database.driver", "mysql")),
			Filepath: env.GetString("database.filepath", cfg.GetString("database.filepath", "./config/sqlite.db")),
			Host:     env.GetString("database.host", cfg.GetString("database.host", "localhost")),
			Port:     env.GetInt("database.port", cfg.GetInt("database.port", 3306)),
			DBName:   env.GetString("database.dbname", cfg.GetString("database.dbname", "dwz")),
			Username: env.GetString("database.username", cfg.GetString("database.username", "dwz")),
			Password: env.GetString("database.password", cfg.GetString("database.password", "dwz")),
		},
		RedisConfig: service.InstallRedisConfig{
			Host:     env.GetString("redis.host", cfg.GetString("redis.host", "localhost")),
			Port:     env.GetInt("redis.port", cfg.GetInt("redis.port", 6379)),
			Password: env.GetString("redis.password", cfg.GetString("redis.password", "")),
			DB:       env.GetInt("redis.db", cfg.GetInt("redis.db", 0)),
		},
	}

	c.HTML(http.StatusOK, "install.html", pageData)
}

func (ctrl InstallController) TestConnection(c httpInterfaces.RouterContextInterface) {
	var req installTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	installer := service.NewInitInstallService(helperPkg.GetHelper())
	if err := installer.TestDatabaseConnection(toInstallDB(req.Database), 1); err != nil {
		ctrl.Error(c, http.StatusBadRequest, "数据库连接失败: "+err.Error())
		return
	}

	needsRedis := req.CacheDriver == "redis" || req.IDGeneratorDriver == "redis"
	if needsRedis && req.Redis != nil {
		if err := installer.TestRedisConnection(toInstallRedis(*req.Redis), 2); err != nil {
			ctrl.Error(c, http.StatusBadRequest, "Redis连接失败: "+err.Error())
			return
		}
	}

	ctrl.SuccessWithMessage(c, "连接测试成功", nil)
}

func (ctrl InstallController) Install(c httpInterfaces.RouterContextInterface) {
	var req installRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	helper := helperPkg.GetHelper()
	logger := helper.GetLogger()
	if installed := helper.GetInstalled(); installed != nil && installed.IsInstalled() {
		ctrl.Error(c, http.StatusBadRequest, "系统已经安装")
		return
	}

	installer := service.NewInitInstallService(helper)

	dbCfg := toInstallDB(req.Database)
	if err := installer.TestDatabaseConnection(dbCfg, 1); err != nil {
		ctrl.Error(c, http.StatusBadRequest, "数据库连接失败: "+err.Error())
		return
	}

	var redisCfg service.InstallRedisConfig
	needsRedis := req.CacheDriver == "redis" || req.IDGeneratorDriver == "redis"
	if needsRedis {
		if req.Redis == nil {
			ctrl.Error(c, http.StatusBadRequest, "Redis 驱动需要 Redis 配置")
			return
		}
		redisCfg = toInstallRedis(*req.Redis)
		if err := installer.TestRedisConnection(redisCfg, 2); err != nil {
			ctrl.Error(c, http.StatusBadRequest, "Redis连接失败: "+err.Error())
			return
		}
	}

	if err := installer.CreateConfigFile(dbCfg, redisCfg, req.CacheDriver, req.IDGeneratorDriver); err != nil {
		ctrl.Error(c, http.StatusInternalServerError, "配置文件创建失败: "+err.Error())
		return
	}

	// 同步在当前请求里跑迁移并创建管理员。失败会沿 HTTP 响应直接返回给前端,
	// 避免原先"推迟到 syscall.Exec 之后 Migration server 跑"的路径:go-web 的
	// initializeServices 会静默吞掉 server.Run() 的错误,一旦重启 / 重装失败,
	// 用户会看到 HTTP 正常但数据库空无一表。install.lock 放到迁移成功之后才写,
	// 保证 lock 与 schema 状态一致,失败可以直接重试安装向导。
	adminPayload := install_bootstrap.AdminPayload{
		Username: req.Admin.Username,
		Password: req.Admin.Password,
		Email:    req.Admin.Email,
	}
	if err := installer.RunMigrationsAndSeed(dbCfg, adminPayload); err != nil {
		ctrl.Error(c, http.StatusInternalServerError, "数据库初始化失败: "+err.Error())
		return
	}

	if err := installer.CreateInstallLock(); err != nil {
		ctrl.Error(c, http.StatusInternalServerError, "安装标记创建失败: "+err.Error())
		return
	}

	// 翻转当前进程的 installed 标记,让在 exec 完成前到达的请求不再走安装向导。
	if installed := helper.GetInstalled(); installed != nil {
		installed.SetInstalled(true)
	}

	ctrl.SuccessWithMessage(c, "安装完成，服务即将重启...", nil)

	// go-web's reload reuses the existing container providers (same priority
	// rejects replacement), so the freshly written config.yaml would be
	// ignored. Re-exec the binary instead — that gives us a guaranteed clean
	// container, and gomander preserves the daemon lifecycle.
	go func() {
		time.Sleep(500 * time.Millisecond)
		binary, err := os.Executable()
		if err != nil {
			logger.Error("[install] 无法定位可执行文件，请手动重启服务: " + err.Error())
			return
		}
		if err := syscall.Exec(binary, os.Args, os.Environ()); err != nil {
			logger.Error("[install] 重启服务失败，请手动重启: " + err.Error())
		}
	}()
}

func toInstallDB(p installDatabasePayload) service.InstallDatabaseConfig {
	return service.InstallDatabaseConfig{
		Driver:   p.Type,
		Filepath: p.Filepath,
		Host:     p.Host,
		Port:     p.Port,
		DBName:   p.Name,
		Username: p.User,
		Password: p.Password,
	}
}

func toInstallRedis(p installRedisPayload) service.InstallRedisConfig {
	return service.InstallRedisConfig{
		Host:     p.Host,
		Port:     p.Port,
		Password: p.Password,
		DB:       p.DB,
	}
}
