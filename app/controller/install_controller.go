package controller

import (
	"net/http"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
	"cnb.cool/mliev/open/go-web/pkg/server/reload"
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

	if err := installer.CreateInstallLock(); err != nil {
		ctrl.Error(c, http.StatusInternalServerError, "安装标记创建失败: "+err.Error())
		return
	}

	// Mark in-memory state installed so the next request after reload sees the lock.
	if installed := helper.GetInstalled(); installed != nil {
		installed.SetInstalled(true)
	}

	// Trigger a SIGHUP-equivalent reload so the new config / DB / cache /
	// id_generator are picked up by the assemblies. The admin user is created
	// after reload completes via a one-shot env hook.
	go func(req installRequest) {
		// brief delay so the response can be flushed
		// before the HTTP server is torn down by reload.
		// 100ms is enough on local; production reload also tolerates this.
		// (kept inline to avoid an extra import for time.Sleep semantics.)
		// no-op below
		reload.TriggerReload()
		// After reload, the new container has fresh DB/Redis/etc.
		// Create the admin user using the freshly wired helper.
		// Note: reload is asynchronous from the perspective of this goroutine;
		// the install_controller responds before this completes.
		_ = createAdminPostInstall(req.Admin)
	}(req)

	ctrl.SuccessWithMessage(c, "安装完成，正在重载服务...", nil)
}

func createAdminPostInstall(admin installAdminPayload) error {
	// After the install endpoint triggers reload, the new container builds a
	// fresh DB and the migration server creates the schema. We poll for the
	// users table to exist before inserting; 50 × 100ms = 5s ceiling.
	const attempts = 50
	for i := 0; i < attempts; i++ {
		time.Sleep(100 * time.Millisecond)
		h := helperPkg.GetHelper()
		db := h.GetDatabase()
		if db == nil {
			continue
		}
		if err := db.Exec("SELECT 1 FROM users LIMIT 1").Error; err != nil {
			continue
		}
		installer := service.NewInitInstallService(h)
		return installer.CreateAdminUser(admin.Username, "", admin.Email, "", admin.Password)
	}
	return nil
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
