package controller

import (
	"strconv"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type CampaignController struct {
	BaseResponse
}

func (ctrl CampaignController) Create(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限创建活动")
		return
	}
	var req dto.CampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetCurrentUserID(c)
	response, err := service.NewCampaignService(helperPkg.GetHelper()).Create(middleware.GetCurrentWorkspaceID(c), userID, &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl CampaignController) List(c httpInterfaces.RouterContextInterface) {
	var req dto.CampaignListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewCampaignService(helperPkg.GetHelper()).List(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl CampaignController) Get(c httpInterfaces.RouterContextInterface) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID")
		return
	}
	response, err := service.NewCampaignService(helperPkg.GetHelper()).Get(id, middleware.GetCurrentWorkspaceID(c))
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

func (ctrl CampaignController) Update(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限更新活动")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID")
		return
	}
	var req dto.CampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewCampaignService(helperPkg.GetHelper()).Update(id, middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl CampaignController) Delete(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限删除活动")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID")
		return
	}
	if err := service.NewCampaignService(helperPkg.GetHelper()).Delete(id, middleware.GetCurrentWorkspaceID(c)); err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.SuccessWithMessage(c, "删除成功", nil)
}

func (ctrl CampaignController) Reports(c httpInterfaces.RouterContextInterface) {
	var req dto.CampaignReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewCampaignService(helperPkg.GetHelper()).Reports(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}
