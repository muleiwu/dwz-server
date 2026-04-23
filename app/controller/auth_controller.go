package controller

import (
	"fmt"
	"net/http"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

// AuthController 认证控制器
type AuthController struct {
	BaseResponse
}

// Login 用户登录
func (ctrl AuthController) Login(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// OIDC 唯一登录模式:除非通过 OIDC_EXCLUSIVE_BYPASS 显式放行,
	// 否则密码登录直接 403。破开 glass 后仍然记录一次 WARN 便于审计。
	oidcSvc := service.NewOIDCService(helper)
	if exclusive, providerName, err := oidcSvc.IsExclusiveMode(); err != nil {
		helper.GetLogger().Warn("[oidc] 检查 exclusive 状态失败: " + err.Error())
	} else if exclusive {
		if oidcSvc.IsExclusiveBypass() {
			helper.GetLogger().Warn(fmt.Sprintf(
				"[oidc] 密码登录因 OIDC_EXCLUSIVE_BYPASS 放行, username=%s provider=%s",
				req.Username, providerName,
			))
		} else {
			ctrl.Error(c, constants.ErrCodeForbidden, "系统已启用 OIDC 唯一登录,请使用 SSO 登录")
			return
		}
	}

	// 获取客户端IP地址
	clientIP := c.ClientIP()

	// 调用AuthService进行登录
	authService := service.NewAuthService(helper)
	response, err := authService.Login(&req, clientIP)
	if err != nil {
		// 检查是否为认证错误
		if authErr, ok := err.(*service.AuthError); ok {
			if authErr.IsRateLimitError {
				// 速率限制错误，返回429状态码和详细信息
				c.JSON(http.StatusTooManyRequests, dto.RateLimitErrorResponse{
					Code:              authErr.Code,
					Message:           authErr.Message,
					LimitType:         authErr.LimitType,
					RemainingAttempts: authErr.RemainingAttempts,
					LockoutSeconds:    authErr.LockoutSeconds,
				})
				return
			}

			// 用户被禁用
			if authErr.Code == 403 {
				ctrl.Error(c, constants.ErrCodeForbidden, authErr.Message)
				return
			}

			// 认证失败（用户名或密码错误）
			ctrl.ErrorWithData(c, constants.ErrCodeUnauthorized, authErr.Message, map[string]any{
				"remaining_attempts": authErr.RemainingAttempts,
			})
			return
		}

		// 其他错误
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetLoginOptions 免认证接口,返回登录页渲染所需的开关(目前用于 OIDC SSO 按钮)。
func (ctrl AuthController) GetLoginOptions(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	oidcSvc := service.NewOIDCService(helper)
	opts, err := oidcSvc.GetLoginOptions()
	if err != nil {
		// 未配置或读取失败时返回保守默认,避免阻塞登录页。
		ctrl.Success(c, map[string]any{"oidc_enabled": false})
		return
	}
	ctrl.Success(c, opts)
}

// Logout 用户登出
func (ctrl AuthController) Logout(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	// 获取当前用户信息
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}

	// 调用AuthService进行登出
	authService := service.NewAuthService(helper)
	err := authService.Logout(currentUser.ID)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.SuccessWithMessage(c, "登出成功", nil)
}
