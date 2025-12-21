package controller

import (
	"strconv"
	"strings"

	"cnb.cool/mliev/open/dwz-server/app/constants"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/middleware"
	"cnb.cool/mliev/open/dwz-server/app/service"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	BaseResponse
}

// CreateUser 创建用户
func (ctrl UserController) CreateUser(c *gin.Context, helper interfaces.HelperInterface) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userService := service.NewUserService(helper)
	response, err := userService.CreateUser(&req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetUser 获取用户详情
func (ctrl UserController) GetUser(c *gin.Context, helper interfaces.HelperInterface) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	userService := service.NewUserService(helper)
	response, err := userService.GetUser(id)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.Success(c, response)
}

// UpdateUser 更新用户
func (ctrl UserController) UpdateUser(c *gin.Context, helper interfaces.HelperInterface) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userService := service.NewUserService(helper)
	response, err := userService.UpdateUser(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.Success(c, response)
}

// DeleteUser 删除用户
func (ctrl UserController) DeleteUser(c *gin.Context, helper interfaces.HelperInterface) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	userService := service.NewUserService(helper)
	err = userService.DeleteUser(id)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.SuccessWithMessage(c, "删除成功", nil)
}

// GetUserList 获取用户列表
func (ctrl UserController) GetUserList(c *gin.Context, helper interfaces.HelperInterface) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userService := service.NewUserService(helper)
	response, err := userService.GetUserList(&req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// ChangePassword 修改密码
func (ctrl UserController) ChangePassword(c *gin.Context, helper interfaces.HelperInterface) {
	// 获取当前用户ID
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userService := service.NewUserService(helper)
	err := userService.ChangePassword(currentUser.ID, &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.SuccessWithMessage(c, "密码修改成功", nil)
}

// ResetPassword 重置密码
func (ctrl UserController) ResetPassword(c *gin.Context, helper interfaces.HelperInterface) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userService := service.NewUserService(helper)
	err = userService.ResetPassword(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.SuccessWithMessage(c, "密码重置成功", nil)
}

// GetCurrentUser 获取当前用户信息
func (ctrl UserController) GetCurrentUser(c *gin.Context, helper interfaces.HelperInterface) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}

	userService := service.NewUserService(helper)
	userInfo := userService.ConvertToUserInfo(currentUser)
	ctrl.Success(c, userInfo)
}

// CreateToken 创建Token
func (ctrl UserController) CreateToken(c *gin.Context, helper interfaces.HelperInterface) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}

	var req dto.CreateUserTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	tokenService := service.NewUserTokenService(helper)
	response, err := tokenService.CreateToken(currentUser.ID, &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetTokenList 获取Token列表
func (ctrl UserController) GetTokenList(c *gin.Context, helper interfaces.HelperInterface) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}

	var req dto.UserTokenListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	tokenService := service.NewUserTokenService(helper)
	response, err := tokenService.GetTokenList(currentUser.ID, &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// DeleteToken 删除Token
func (ctrl UserController) DeleteToken(c *gin.Context, helper interfaces.HelperInterface) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}

	tokenIdStr := c.Param("token_id")
	tokenId, err := strconv.ParseUint(tokenIdStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的Token ID格式")
		return
	}

	tokenService := service.NewUserTokenService(helper)
	err = tokenService.DeleteToken(currentUser.ID, tokenId)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "无权限") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.SuccessWithMessage(c, "Token删除成功", nil)
}

// GetOperationLogs 获取操作日志
func (ctrl UserController) GetOperationLogs(c *gin.Context, helper interfaces.HelperInterface) {
	var req dto.OperationLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	logService := service.NewOperationLogService(helper)
	response, err := logService.GetLogList(&req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}
