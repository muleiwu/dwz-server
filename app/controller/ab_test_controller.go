package controller

import (
	"errors"
	"strconv"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type ABTestController struct {
	BaseResponse
}

// CreateABTestFeedback 公开转化反馈接口
func (ctrl ABTestController) CreateABTestFeedback(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	var req dto.ABTestFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	response, err := service.NewABTestService(helper).RecordABTestFeedback(
		&req,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		c.GetHeader("Referer"),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrABTestFeedbackBadRequest):
			ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		case errors.Is(err, service.ErrABTestFeedbackInvalidToken), errors.Is(err, service.ErrABTestFeedbackExpiredToken):
			ctrl.Error(c, constants.ErrCodeUnauthorized, err.Error())
		default:
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.Success(c, response)
}

// CreateABTest 创建AB测试
func (ctrl ABTestController) CreateABTest(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理AB测试")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.CreateABTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	abTestService := service.NewABTestService(helper)
	response, err := abTestService.CreateABTestInWorkspace(&req, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetABTest 获取AB测试详情
func (ctrl ABTestController) GetABTest(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	abTestService := service.NewABTestService(helper)
	response, err := abTestService.GetABTestInWorkspace(id, middleware.GetCurrentWorkspaceID(c))
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

// UpdateABTest 更新AB测试
func (ctrl ABTestController) UpdateABTest(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理AB测试")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	var req dto.UpdateABTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	abTestService := service.NewABTestService(helper)
	response, err := abTestService.UpdateABTestInWorkspace(id, &req, middleware.GetCurrentWorkspaceID(c))
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

// DeleteABTest 删除AB测试
func (ctrl ABTestController) DeleteABTest(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理AB测试")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	abTestService := service.NewABTestService(helper)
	err = abTestService.DeleteABTestInWorkspace(id, middleware.GetCurrentWorkspaceID(c))
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

// GetABTestList 获取AB测试列表
func (ctrl ABTestController) GetABTestList(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.ABTestListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	abTestService := service.NewABTestService(helper)
	response, err := abTestService.GetABTestListInWorkspace(&req, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// StartABTest 启动AB测试
func (ctrl ABTestController) StartABTest(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理AB测试")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	var req dto.StartABTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有传递时间参数，使用空请求
		req = dto.StartABTestRequest{}
	}

	abTestService := service.NewABTestService(helper)
	response, err := abTestService.StartABTestInWorkspace(id, &req, middleware.GetCurrentWorkspaceID(c))
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

// StopABTest 停止AB测试
func (ctrl ABTestController) StopABTest(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理AB测试")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	var req dto.StopABTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有传递时间参数，使用空请求
		req = dto.StopABTestRequest{}
	}

	abTestService := service.NewABTestService(helper)
	response, err := abTestService.StopABTestInWorkspace(id, &req, middleware.GetCurrentWorkspaceID(c))
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

// GetABTestStatistics 获取AB测试统计信息
func (ctrl ABTestController) GetABTestStatistics(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 7
	}

	abTestService := service.NewABTestService(helper)
	response, err := abTestService.GetABTestStatisticsInWorkspace(id, days, middleware.GetCurrentWorkspaceID(c))
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
