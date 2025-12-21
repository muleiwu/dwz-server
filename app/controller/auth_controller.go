package controller

import (
	"net/http"

	"cnb.cool/mliev/open/dwz-server/app/constants"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/middleware"
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	BaseResponse
}

// Login 用户登录
func (ctrl AuthController) Login(c *gin.Context, helper interfaces.HelperInterface) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
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
			ctrl.ErrorWithData(c, constants.ErrCodeUnauthorized, authErr.Message, gin.H{
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

// Logout 用户登出
func (ctrl AuthController) Logout(c *gin.Context, helper interfaces.HelperInterface) {
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
