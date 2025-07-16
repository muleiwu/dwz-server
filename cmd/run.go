package cmd

import (
	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/config/middleware"
	"cnb.cool/mliev/open/dwz-server/helper/database"
	"cnb.cool/mliev/open/dwz-server/helper/env"
	"cnb.cool/mliev/open/dwz-server/helper/install"
	"cnb.cool/mliev/open/dwz-server/helper/logger"
	"cnb.cool/mliev/open/dwz-server/helper/redis"
	"cnb.cool/mliev/open/dwz-server/router"
	"cnb.cool/mliev/open/dwz-server/util"
	"context"
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var templateFS embed.FS

// Start 启动应用程序
func Start(fs embed.FS) {
	templateFS = fs
	initializeServices()
	go RunHttp()
	// 添加阻塞以保持主程序运行
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

// initializeServices 初始化所有服务
func initializeServices() {
	// 检查安装状态

	if err := env.InitViper(); err != nil {
		logger.Logger().Warn(fmt.Sprintf("配置初始化警告: %v", err))
	}

	if !install.CheckInstallStatus() && env.EnvBool("AUTO_INIT", false) {
		installService := service.NewInitInstallService()
		installService.AutoInstall()
	}

	if install.CheckInstallStatus() {
		logger.Logger().Info("系统已安装，正在初始化服务...")

		// 自动迁移数据库表结构
		if err := autoMigrate(); err != nil {
			logger.Logger().Error(fmt.Sprintf("数据库迁移失败: %v", err))
			os.Exit(1)
		}

		// 初始化Redis连接
		redis.GetRedis()

		// 初始化分布式发号器的域名计数器
		if err := initializeDomainCounters(); err != nil {
			logger.Logger().Error(fmt.Sprintf("分布式发号器初始化失败: %v", err))
			os.Exit(1)
		}
	} else {
		logger.Logger().Info("系统未安装，请访问 /install.cgi 进行安装")
		// 未安装时，只初始化基本的日志服务
		return
	}
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate() error {
	return database.AutoMigrate()
}

// zapLogWriter 实现io.Writer接口，将gin的日志输出重定向到zap
type zapLogWriter struct {
	zapLogger *zap.Logger
	isError   bool
}

// Write 实现io.Writer接口
func (z *zapLogWriter) Write(p []byte) (n int, err error) {
	if z.isError {
		z.zapLogger.Error(string(p))
	} else {
		z.zapLogger.Info(string(p))
	}
	return len(p), nil
}

// RunHttp 启动HTTP服务器并注册路由和中间件
func RunHttp() {

	// 设置Gin模式
	if env.EnvString("mode", "") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 完全替换gin的默认Logger
	gin.DisableConsoleColor()
	zapLogger := logger.Logger()
	gin.DefaultWriter = &zapLogWriter{zapLogger: zapLogger}
	gin.DefaultErrorWriter = &zapLogWriter{zapLogger: zapLogger, isError: true}

	// 配置Gin引擎
	// 配置Gin引擎并替换默认logger
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(GinZapLogger())

	// 加载HTML模板
	if err := loadTemplates(engine); err != nil {
		logger.Logger().Error(fmt.Sprintf("加载模板失败: %v", err))
		return
	}

	// 注册中间件
	handlerFuncs := middleware.MiddlewareConfig{}.Get()
	for i, handlerFunc := range handlerFuncs {
		if handlerFunc == nil {
			continue
		}
		engine.Use(handlerFunc)
		logger.Logger().Info(fmt.Sprintf("注册中间件: %d", i))
	}

	router.InitRouter(engine)

	// 创建一个HTTP服务器，以便能够优雅关闭
	addr := env.EnvString("addr", ":8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// 创建一个通道来接收中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在单独的goroutine中启动服务器
	go func() {
		logger.Logger().Info(fmt.Sprintf("服务器启动于 %s", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger().Error(fmt.Sprintf("启动服务器失败: %v", err))
		}
	}()

	// 等待中断信号
	<-quit
	logger.Logger().Info("正在关闭服务器...")

	// 创建一个5秒的上下文用于超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅地关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger().Error(fmt.Sprintf("服务器强制关闭: %v", err))
	}

	logger.Logger().Info("服务器已优雅关闭")
}

// GinZapLogger 返回一个Gin中间件，使用zap记录HTTP请求
func GinZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// 请求处理完成后记录日志
		cost := time.Since(start)
		zapLogger := logger.Logger()
		statusCode := c.Writer.Status()

		// 通用的日志字段
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", cost),
			zap.String("user-agent", c.Request.UserAgent()),
		}

		// 根据状态码决定日志级别
		switch {
		case statusCode >= 500:
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
			zapLogger.Error("请求处理", fields...)
		case statusCode >= 400:
			zapLogger.Warn("请求处理", fields...)
		default:
			zapLogger.Info("请求处理", fields...)
		}
	}
}

// initializeDomainCounters 初始化域名计数器
func initializeDomainCounters() error {
	domainDao := &dao.DomainDao{}
	shortLinkDao := &dao.ShortLinkDao{}
	idGenerator := util.NewDistributedIDGenerator(redis.GetRedis())

	// 获取所有活跃域名
	domains, err := domainDao.GetActiveDomains()
	if err != nil {
		return fmt.Errorf("获取活跃域名失败: %v", err)
	}

	logger.Logger().Info(fmt.Sprintf("开始初始化%d个域名的计数器", len(domains)))

	// 为每个域名初始化计数器
	for _, domain := range domains {
		// 查询该域名下的最大short_link ID
		maxID, err := shortLinkDao.GetMaxIDByDomain(domain.Domain)
		if err != nil {
			return fmt.Errorf("查询域名%s最大ID失败: %v", domain.Domain, err)
		}

		// 设置起始值为maxID + 1，确保新生成的ID不会冲突
		startValue := maxID + 1

		// 初始化Redis计数器
		if err := idGenerator.InitializeDomainCounter(domain.ID, startValue); err != nil {
			return fmt.Errorf("初始化域名%s Redis计数器失败: %v", domain.Domain, err)
		}

		logger.Logger().Info(fmt.Sprintf("域名%s(ID:%d)计数器初始化完成，起始值:%d", domain.Domain, domain.ID, startValue))
	}

	logger.Logger().Info("所有域名计数器初始化完成")
	return nil
}

// loadTemplates 加载模板到Gin引擎
func loadTemplates(engine *gin.Engine) error {
	// 从嵌入的文件系统创建子文件系统
	subFS, err := fs.Sub(templateFS, "templates")
	if err != nil {
		return err
	}

	// 创建模板并解析所有模板文件
	tmpl := template.Must(template.New("").ParseFS(subFS, "*.html"))

	// 设置HTML模板
	engine.SetHTMLTemplate(tmpl)

	return nil
}
