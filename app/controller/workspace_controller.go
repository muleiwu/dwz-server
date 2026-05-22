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

type WorkspaceController struct {
	BaseResponse
}

func (ctrl WorkspaceController) ListWorkspaces(c httpInterfaces.RouterContextInterface) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}
	response, err := service.NewWorkspaceService(helperPkg.GetHelper()).ListWorkspaces(user.ID)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl WorkspaceController) CreateWorkspace(c httpInterfaces.RouterContextInterface) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}
	var req dto.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewWorkspaceService(helperPkg.GetHelper()).CreateWorkspace(user.ID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "已存在") {
			ctrl.Error(c, constants.ErrCodeConflict, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}
	ctrl.Success(c, response)
}

func (ctrl WorkspaceController) UpdateWorkspace(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageWorkspace(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理工作区")
		return
	}
	var req dto.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewWorkspaceService(helperPkg.GetHelper()).UpdateWorkspace(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl WorkspaceController) ListMembers(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageWorkspace(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理成员")
		return
	}
	response, err := service.NewWorkspaceService(helperPkg.GetHelper()).ListMembers(middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl WorkspaceController) AddMember(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageWorkspace(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理成员")
		return
	}
	var req dto.AddWorkspaceMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewWorkspaceService(helperPkg.GetHelper()).AddMember(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		if strings.Contains(err.Error(), "已在") {
			ctrl.Error(c, constants.ErrCodeConflict, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}
	ctrl.Success(c, response)
}

func (ctrl WorkspaceController) UpdateMember(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageWorkspace(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理成员")
		return
	}
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的用户ID")
		return
	}
	var req dto.UpdateWorkspaceMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	response, err := service.NewWorkspaceService(helperPkg.GetHelper()).UpdateMember(middleware.GetCurrentWorkspaceID(c), userID, &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, response)
}

func (ctrl WorkspaceController) RemoveMember(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageWorkspace(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理成员")
		return
	}
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的用户ID")
		return
	}
	if err := service.NewWorkspaceService(helperPkg.GetHelper()).RemoveMember(middleware.GetCurrentWorkspaceID(c), userID); err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.SuccessWithMessage(c, "移除成功", nil)
}
