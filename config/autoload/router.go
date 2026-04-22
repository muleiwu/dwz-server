package autoload

import (
	"net/http"

	"cnb.cool/mliev/dwz/dwz-server/app/controller"
	"cnb.cool/mliev/dwz/dwz-server/app/middleware"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type Router struct{}

func (Router) InitConfig() map[string]any {
	return map[string]any{
		"http.router": func(router httpInterfaces.RouterInterface) {
			router.GET("/favicon.ico", func(c httpInterfaces.RouterContextInterface) {
				c.Status(http.StatusNoContent)
			})

			// 健康检查
			health := router.Group("/health")
			{
				health.GET("", controller.HealthController{}.GetHealth)
				health.GET("/simple", controller.HealthController{}.GetHealthSimple)
			}

			// 安装页面 + 安装 API（无认证）
			router.GET("/install/index", controller.InstallController{}.GetInstall)
			install := router.Group("/api/v1/install")
			{
				install.POST("/test-db", controller.InstallController{}.TestConnection)
				install.POST("", controller.InstallController{}.Install)
			}

			// 首页
			router.GET("/", controller.IndexController{}.GetIndex)

			// 登录路由（无认证，但记录操作日志）
			authPublic := router.Group("/api/v1/auth")
			authPublic.Use(middleware.OperationLogMiddleware())
			{
				authPublic.POST("/login", controller.AuthController{}.Login)
			}
			// 兼容老登录路径
			compatLogin := router.Group("/api/v1")
			compatLogin.Use(middleware.OperationLogMiddleware())
			{
				compatLogin.POST("/login", controller.AuthController{}.Login)
			}

			// 受保护的 API：操作日志 + 鉴权
			v1 := router.Group("/api/v1")
			v1.Use(middleware.OperationLogMiddleware())
			v1.Use(middleware.AuthMiddleware())
			{
				auth := v1.Group("/auth")
				{
					auth.POST("/logout", controller.AuthController{}.Logout)
				}

				short := v1.Group("/short_links")
				{
					short.POST("", controller.ShortLinkController{}.CreateShortLink)
					short.GET("", controller.ShortLinkController{}.GetShortLinkList)
					short.GET("/:id", controller.ShortLinkController{}.GetShortLink)
					short.PUT("/:id", controller.ShortLinkController{}.UpdateShortLink)
					short.PUT("/:id/status", controller.ShortLinkController{}.UpdateShortLinkStatus)
					short.DELETE("/:id", controller.ShortLinkController{}.DeleteShortLink)
					short.GET("/:id/statistics", controller.ShortLinkController{}.GetShortLinkStatistics)
					short.POST("/batch", controller.ShortLinkController{}.BatchCreateShortLinks)
				}

				domains := v1.Group("/domains")
				{
					domains.POST("", controller.DomainController{}.CreateDomain)
					domains.GET("", controller.DomainController{}.GetDomainList)
					domains.GET("/active", controller.DomainController{}.GetActiveDomains)
					domains.PUT("/:id", controller.DomainController{}.UpdateDomain)
					domains.PUT("/:id/status", controller.DomainController{}.UpdateStatusDomain)
					domains.DELETE("/:id", controller.DomainController{}.DeleteDomain)
				}

				ab := v1.Group("/ab_tests")
				{
					ab.POST("", controller.ABTestController{}.CreateABTest)
					ab.GET("", controller.ABTestController{}.GetABTestList)
					ab.GET("/:id", controller.ABTestController{}.GetABTest)
					ab.PUT("/:id", controller.ABTestController{}.UpdateABTest)
					ab.DELETE("/:id", controller.ABTestController{}.DeleteABTest)
					ab.POST("/:id/start", controller.ABTestController{}.StartABTest)
					ab.POST("/:id/stop", controller.ABTestController{}.StopABTest)
					ab.GET("/:id/statistics", controller.ABTestController{}.GetABTestStatistics)
				}

				users := v1.Group("/users")
				{
					users.POST("", controller.UserController{}.CreateUser)
					users.GET("", controller.UserController{}.GetUserList)
					users.GET("/:id", controller.UserController{}.GetUser)
					users.PUT("/:id", controller.UserController{}.UpdateUser)
					users.DELETE("/:id", controller.UserController{}.DeleteUser)
					users.POST("/:id/reset-password", controller.UserController{}.ResetPassword)
				}

				profile := v1.Group("/profile")
				{
					profile.GET("", controller.UserController{}.GetCurrentUser)
					profile.POST("/change-password", controller.UserController{}.ChangePassword)
				}

				tokens := v1.Group("/tokens")
				{
					tokens.POST("", controller.UserController{}.CreateToken)
					tokens.GET("", controller.UserController{}.GetTokenList)
					tokens.DELETE("/:token_id", controller.UserController{}.DeleteToken)
				}

				logs := v1.Group("/logs")
				{
					logs.GET("", controller.UserController{}.GetOperationLogs)
				}

				clickStats := v1.Group("/click_statistics")
				{
					clickStats.GET("", controller.ClickStatisticController{}.GetClickStatisticList)
					clickStats.GET("/analysis", controller.ClickStatisticController{}.GetClickStatisticAnalysis)
				}

				stats := v1.Group("/statistics")
				{
					stats.GET("/system", controller.StatisticsController{}.GetSystem)
					stats.GET("/dashboard", controller.StatisticsController{}.GetDashboard)
					stats.GET("/short-links", controller.StatisticsController{}.GetShortLinks)
				}

				abClick := v1.Group("/ab_test_click_statistics")
				{
					abClick.GET("", controller.ABTestClickStatisticController{}.GetABTestClickStatisticList)
					abClick.GET("/analysis", controller.ABTestClickStatisticController{}.GetABTestClickStatisticAnalysis)
					abClick.GET("/:id/variants", controller.ABTestClickStatisticController{}.GetABTestVariantStatistics)
				}
			}

			// 短码跳转 (`/<code>` 与 `/preview/<code>`) 由
			// config/autoload/short_code_dispatch.go 注册的中间件处理，避免
			// go-web RegexGroup 在根路径与具体路由的 catch-all 冲突。
		},
	}
}
