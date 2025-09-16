package autoload

import (
	"cnb.cool/mliev/open/dwz-server/app/controller"
	"cnb.cool/mliev/open/dwz-server/app/middleware"
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/internal/pkg/http_server/impl"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/ginregex"
)

type Router struct {
}

func (receiver Router) InitConfig(helper envInterface.HelperInterface) map[string]any {
	return map[string]any{
		"http.router": func(router *gin.Engine, deps *impl.HttpDeps) {

			regexRouter := ginregex.New(router, nil)

			// 添加安装检查中间件
			router.Use(middleware.InstallMiddleware(helper))

			router.Any("/favicon.ico", func(c *gin.Context) {
				c.Status(204) // 返回204 No Content
			})

			// 健康检查接口
			router.GET("/health", deps.WrapHandler(controller.HealthController{}.GetHealth))
			router.GET("/health/simple", deps.WrapHandler(controller.HealthController{}.GetHealthSimple))

			router.GET("/install.cgi", deps.WrapHandler(controller.InstallController{}.GetInstall))

			// 安装接口路由（不需要认证）
			installRoutes := router.Group("/api/v1/install")
			{
				installRoutes.POST("/test-db", deps.WrapHandler(controller.InstallController{}.TestConnection)) // 测试数据库连接
				installRoutes.POST("", deps.WrapHandler(controller.InstallController{}.Install))                // 执行安装
			}

			// 首页
			router.GET("/", deps.WrapHandler(controller.IndexController{}.GetIndex))

			// 登录路由（不需要认证）
			router.POST("/api/v1/login", middleware.OperationLogMiddleware(helper), deps.WrapHandler(controller.UserController{}.Login)) // 用户登录

			// API路由组
			v1 := router.Group("/api/v1", middleware.OperationLogMiddleware(helper), middleware.AuthMiddleware(helper))
			{
				// 短网址管理路由
				shortLinks := v1.Group("/short_links")
				{
					shortLinks.POST("", deps.WrapHandler(controller.ShortLinkController{}.CreateShortLink))                      // 创建短网址
					shortLinks.GET("", deps.WrapHandler(controller.ShortLinkController{}.GetShortLinkList))                      // 获取短网址列表
					shortLinks.GET("/:id", deps.WrapHandler(controller.ShortLinkController{}.GetShortLink))                      // 获取短网址详情
					shortLinks.PUT("/:id", deps.WrapHandler(controller.ShortLinkController{}.UpdateShortLink))                   // 更新短网址
					shortLinks.DELETE("/:id", deps.WrapHandler(controller.ShortLinkController{}.DeleteShortLink))                // 删除短网址
					shortLinks.GET("/:id/statistics", deps.WrapHandler(controller.ShortLinkController{}.GetShortLinkStatistics)) // 获取统计信息
					shortLinks.POST("/batch", deps.WrapHandler(controller.ShortLinkController{}.BatchCreateShortLinks))          // 批量创建短网址
				}

				// 域名管理路由
				domains := v1.Group("/domains")
				{
					domains.POST("", deps.WrapHandler(controller.DomainController{}.CreateDomain))                 // 创建域名配置
					domains.GET("", deps.WrapHandler(controller.DomainController{}.GetDomainList))                 // 获取域名列表
					domains.GET("/active", deps.WrapHandler(controller.DomainController{}.GetActiveDomains))       // 获取活跃域名列表
					domains.PUT("/:id", deps.WrapHandler(controller.DomainController{}.UpdateDomain))              // 更新域名配置
					domains.PUT("/:id/status", deps.WrapHandler(controller.DomainController{}.UpdateStatusDomain)) // 更新域名状态
					domains.DELETE("/:id", deps.WrapHandler(controller.DomainController{}.DeleteDomain))           // 删除域名配置
				}

				// AB测试管理路由
				abTests := v1.Group("/ab_tests")
				{
					abTests.POST("", deps.WrapHandler(controller.ABTestController{}.CreateABTest))                      // 创建AB测试
					abTests.GET("", deps.WrapHandler(controller.ABTestController{}.GetABTestList))                      // 获取AB测试列表
					abTests.GET("/:id", deps.WrapHandler(controller.ABTestController{}.GetABTest))                      // 获取AB测试详情
					abTests.PUT("/:id", deps.WrapHandler(controller.ABTestController{}.UpdateABTest))                   // 更新AB测试
					abTests.DELETE("/:id", deps.WrapHandler(controller.ABTestController{}.DeleteABTest))                // 删除AB测试
					abTests.POST("/:id/start", deps.WrapHandler(controller.ABTestController{}.StartABTest))             // 启动AB测试
					abTests.POST("/:id/stop", deps.WrapHandler(controller.ABTestController{}.StopABTest))               // 停止AB测试
					abTests.GET("/:id/statistics", deps.WrapHandler(controller.ABTestController{}.GetABTestStatistics)) // 获取AB测试统计
				}

				// 用户管理路由（需要认证）
				users := v1.Group("/users")
				{
					users.POST("", deps.WrapHandler(controller.UserController{}.CreateUser))                       // 创建用户
					users.GET("", deps.WrapHandler(controller.UserController{}.GetUserList))                       // 获取用户列表
					users.GET("/:id", deps.WrapHandler(controller.UserController{}.GetUser))                       // 获取用户详情
					users.PUT("/:id", deps.WrapHandler(controller.UserController{}.UpdateUser))                    // 更新用户
					users.DELETE("/:id", deps.WrapHandler(controller.UserController{}.DeleteUser))                 // 删除用户
					users.POST("/:id/reset-password", deps.WrapHandler(controller.UserController{}.ResetPassword)) // 重置密码
				}

				// 当前用户相关路由（需要认证）
				profile := v1.Group("/profile")
				{
					profile.GET("", deps.WrapHandler(controller.UserController{}.GetCurrentUser))                  // 获取当前用户信息
					profile.POST("/change-password", deps.WrapHandler(controller.UserController{}.ChangePassword)) // 修改密码
				}

				// Token管理路由（需要认证）
				tokens := v1.Group("/tokens", middleware.AuthMiddleware(helper))
				{
					tokens.POST("", deps.WrapHandler(controller.UserController{}.CreateToken))             // 创建Token
					tokens.GET("", deps.WrapHandler(controller.UserController{}.GetTokenList))             // 获取Token列表
					tokens.DELETE("/:token_id", deps.WrapHandler(controller.UserController{}.DeleteToken)) // 删除Token
				}

				// 操作日志路由（需要认证）
				logs := v1.Group("/logs", middleware.AuthMiddleware(helper))
				{
					logs.GET("", deps.WrapHandler(controller.UserController{}.GetOperationLogs)) // 获取操作日志
				}

				// 点击统计路由
				clickStats := v1.Group("/click_statistics")
				{
					clickStats.GET("", deps.WrapHandler(controller.ClickStatisticController{}.GetClickStatisticList))              // 获取点击统计列表
					clickStats.GET("/analysis", deps.WrapHandler(controller.ClickStatisticController{}.GetClickStatisticAnalysis)) // 获取点击统计分析
				}

				// AB测试点击统计路由
				abTestClickStats := v1.Group("/ab_test_click_statistics")
				{
					abTestClickStats.GET("", deps.WrapHandler(controller.ABTestClickStatisticController{}.GetABTestClickStatisticList))              // 获取AB测试点击统计列表
					abTestClickStats.GET("/analysis", deps.WrapHandler(controller.ABTestClickStatisticController{}.GetABTestClickStatisticAnalysis)) // 获取AB测试点击统计分析
					abTestClickStats.GET("/:id/variants", deps.WrapHandler(controller.ABTestClickStatisticController{}.GetABTestVariantStatistics))  // 获取AB测试版本统计
				}
			}

			// 短网址跳转路由
			regexRouter.GET(`^/(?P<code>[a-zA-Z0-9]+)$`, deps.WrapHandler(controller.ShortLinkController{}.RedirectShortLink))        // 短网址跳转
			regexRouter.GET(`^/preview/(?P<code>[a-zA-Z0-9]+)$`, deps.WrapHandler(controller.ShortLinkController{}.PreviewShortLink)) // 预览短网址

		},
	}
}
