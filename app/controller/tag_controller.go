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

type TagController struct {
	BaseResponse
}

func (ctrl TagController) Create(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限创建标签")
		return
	}
	var req dto.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewTagService(helperPkg.GetHelper()).Create(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			ctrl.Error(c, constants.ErrCodeConflict, "标签已存在")
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}
	ctrl.Success(c, response)
}

func (ctrl TagController) List(c httpInterfaces.RouterContextInterface) {
	var req dto.TagListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewTagService(helperPkg.GetHelper()).List(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl TagController) Get(c httpInterfaces.RouterContextInterface) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID")
		return
	}
	response, err := service.NewTagService(helperPkg.GetHelper()).Get(id, middleware.GetCurrentWorkspaceID(c))
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

func (ctrl TagController) Update(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限更新标签")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID")
		return
	}
	var req dto.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewTagService(helperPkg.GetHelper()).Update(id, middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl TagController) Delete(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限删除标签")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID")
		return
	}
	if err := service.NewTagService(helperPkg.GetHelper()).Delete(id, middleware.GetCurrentWorkspaceID(c)); err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.SuccessWithMessage(c, "删除成功", nil)
}
