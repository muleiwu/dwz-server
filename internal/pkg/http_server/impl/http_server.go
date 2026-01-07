package impl

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HttpServer struct {
	Helper     interfaces.HelperInterface
	templateFS *embed.FS
	routerFunc func(router *gin.Engine)
}

func NewHttpServer(helper interfaces.HelperInterface) *HttpServer {
	return &HttpServer{
		Helper: helper,
	}
}

type zapLogWriter struct {
	logger  interfaces.LoggerInterface
	isError bool
}

// Write 实现io.Writer接口
func (z *zapLogWriter) Write(p []byte) (n int, err error) {
	if z.isError {
		z.logger.Error(string(p))
	} else {
		z.logger.Info(string(p))
	}
	return len(p), nil
}

func (receiver *HttpServer) InitLoggerFields() []interfaces.LoggerFieldInterface {
	return []interfaces.LoggerFieldInterface{}
}

// GinZapLogger 返回一个Gin中间件，使用zap记录HTTP请求
func (receiver *HttpServer) ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// 请求处理完成后记录日志
		cost := time.Since(start)
		statusCode := c.Writer.Status()
		traceId := c.GetString("traceId")

		// 通用的日志字段
		fields := []interfaces.LoggerFieldInterface{
			NewLoggerField("trace-id", traceId),
			NewLoggerField("method", c.Request.Method),
			NewLoggerField("path", path),
			NewLoggerField("query", query),
			NewLoggerField("status", statusCode),
			NewLoggerField("ip", c.ClientIP()),
			NewLoggerField("latency", cost),
			NewLoggerField("user-agent", c.Request.UserAgent()),
		}

		// 根据状态码决定日志级别
		switch {
		case statusCode >= 500:
			fields = append(fields, NewLoggerField("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
			receiver.Helper.GetLogger().Error("请求处理", fields...)
		case statusCode >= 400:
			receiver.Helper.GetLogger().Warn("请求处理", fields...)
		default:
			receiver.Helper.GetLogger().Info("请求处理", fields...)
		}
	}
}

// RunHttp 启动HTTP服务器并注册路由和中间件
func (receiver *HttpServer) RunHttp() {
	// 设置Gin模式
	if receiver.Helper.GetConfig().GetString("http.mode", "") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 完全替换gin的默认Logger
	gin.DisableConsoleColor()

	//gin.DefaultWriter = &zapLogWriter{logger: receiver.Helper.GetLogger()}
	//gin.DefaultErrorWriter = &zapLogWriter{logger: receiver.Helper.GetLogger(), isError: true}

	// 配置Gin引擎
	// 配置Gin引擎并替换默认logger
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(receiver.traceIdMiddleware())
	engine.Use(receiver.ginLogger())

	// 加载HTML模板
	if err := receiver.loadTemplates(engine); err != nil {
		receiver.Helper.GetLogger().Error(fmt.Sprintf("加载模板失败: %v", err))
	}

	// 加载网站静态资源
	receiver.loadWebStatic(engine)

	// 注册中间件
	//handlerFuncs := config.MiddlewareConfig{}.Get()
	middlewareFuncList := receiver.Helper.GetConfig().Get("http.middleware", []gin.HandlerFunc{}).([]gin.HandlerFunc)
	for i, handlerFunc := range middlewareFuncList {
		if handlerFunc == nil {
			continue
		}
		engine.Use(handlerFunc)
		receiver.Helper.GetLogger().Info(fmt.Sprintf("注册中间件: %d", i))
	}

	deps := NewHttpDeps(receiver.Helper, engine)
	header := receiver.Helper.GetConfig().Get("http.router", func(router *gin.Engine, deps *HttpDeps) {

	}).(func(*gin.Engine, *HttpDeps))

	header(engine, deps)

	//receiver.routerFunc(engine)

	// 创建一个HTTP服务器，以便能够优雅关闭
	addr := receiver.Helper.GetConfig().GetString("http.addr", ":8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// 创建一个通道来接收中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在单独的goroutine中启动服务器
	go func() {
		receiver.Helper.GetLogger().Info(fmt.Sprintf("服务器启动于 %s", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			receiver.Helper.GetLogger().Error(fmt.Sprintf("启动服务器失败: %v", err))
			panic(err)
		}
	}()

	// 在单独的goroutine中等待中断信号以便优雅关闭
	go func() {
		// 等待中断信号
		<-quit
		receiver.Helper.GetLogger().Info("正在关闭服务器...")

		// 创建一个5秒的上下文用于超时控制
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 优雅地关闭服务器
		if err := srv.Shutdown(ctx); err != nil {
			receiver.Helper.GetLogger().Error(fmt.Sprintf("服务器强制关闭: %v", err))
		}

		receiver.Helper.GetLogger().Info("服务器已优雅关闭")
	}()
}

// loadTemplates 加载模板到Gin引擎
func (receiver *HttpServer) loadTemplates(engine *gin.Engine) error {
	templatesMode := receiver.Helper.GetConfig().GetString("http.templates_mode", "embed")
	templatesDir := receiver.Helper.GetConfig().GetString("http.templates_dir", "templates")

	receiver.Helper.GetLogger().Debug(fmt.Sprintf("当前模板文件模式：%s", templatesMode))

	var tmpl *template.Template
	var err error

	if templatesMode == "disk" {
		// disk 模式：从磁盘加载模板，支持热更新
		receiver.Helper.GetLogger().Info("Disk 模式：从磁盘加载模板文件，支持热更新")
		tmpl, err = template.ParseGlob(fmt.Sprintf("%s/*.html", templatesDir))
		if err != nil {
			return fmt.Errorf("从磁盘加载模板失败: %v", err)
		}
	} else {
		// embed 模式：从嵌入的文件系统加载模板
		receiver.Helper.GetLogger().Info("Embed 模式：从嵌入文件系统加载模板")
		staticFs := receiver.Helper.GetConfig().Get("static.fs", map[string]embed.FS{}).(map[string]embed.FS)

		templates, ok := staticFs["templates"]
		if !ok {
			return errors.New("没有模板目录需要初始化")
		}

		// 从嵌入的文件系统创建子文件系统
		subFS, err := fs.Sub(templates, templatesDir)
		if err != nil {
			return fmt.Errorf("创建模板子文件系统失败: %v", err)
		}

		parseFS, err := template.New("").ParseFS(subFS, "*.html")
		if err != nil {
			return fmt.Errorf("解析嵌入模板失败: %v", err)
		}

		// 创建模板并解析所有模板文件
		tmpl = template.Must(parseFS, err)
	}

	// 设置HTML模板
	engine.SetHTMLTemplate(tmpl)

	return nil
}

func (receiver *HttpServer) loadWebStatic(engine *gin.Engine) {
	staticHandler := NewStaticHandler(receiver.Helper, engine)
	staticHandler.setupStaticFileServers()
}

func (receiver *HttpServer) traceIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		newUUID, err := uuid.NewV7()
		if err != nil {
			newUUID = uuid.New()
		}
		c.Set("traceId", newUUID.String())
		c.Writer.Header().Set("trace-id", newUUID.String())
		c.Writer.Header().Set("version-number", receiver.Helper.GetVersion().GetVersion())
		c.Writer.Header().Set("version-git-commit", receiver.Helper.GetVersion().GetGitCommit())
		c.Next()
	}
}
