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

type LinkRouteController struct {
	BaseResponse
}

func (ctrl LinkRouteController) ListRoutes(c httpInterfaces.RouterContextInterface) {
	shortLinkID, ok := ctrl.parseShortLinkID(c)
	if !ok {
		return
	}
	resp, err := service.NewLinkRouteService(helperPkg.GetHelper()).
		ListRoutes(shortLinkID, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.writeRouteError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkRouteController) CreateRoute(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理高级路由")
		return
	}
	shortLinkID, ok := ctrl.parseShortLinkID(c)
	if !ok {
		return
	}
	var req dto.LinkRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkRouteService(helperPkg.GetHelper()).
		CreateRoute(shortLinkID, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c), &req)
	if err != nil {
		ctrl.writeRouteError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkRouteController) UpdateRoute(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理高级路由")
		return
	}
	shortLinkID, ok := ctrl.parseShortLinkID(c)
	if !ok {
		return
	}
	routeID, ok := ctrl.parseRouteID(c)
	if !ok {
		return
	}
	var req dto.LinkRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkRouteService(helperPkg.GetHelper()).
		UpdateRoute(routeID, shortLinkID, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c), &req)
	if err != nil {
		ctrl.writeRouteError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkRouteController) DeleteRoute(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理高级路由")
		return
	}
	shortLinkID, ok := ctrl.parseShortLinkID(c)
	if !ok {
		return
	}
	routeID, ok := ctrl.parseRouteID(c)
	if !ok {
		return
	}
	if err := service.NewLinkRouteService(helperPkg.GetHelper()).
		DeleteRoute(routeID, shortLinkID, middleware.GetCurrentWorkspaceID(c)); err != nil {
		ctrl.writeRouteError(c, err)
		return
	}
	ctrl.SuccessWithMessage(c, "删除成功", nil)
}

func (ctrl LinkRouteController) ReorderRoutes(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限管理高级路由")
		return
	}
	shortLinkID, ok := ctrl.parseShortLinkID(c)
	if !ok {
		return
	}
	var req dto.LinkRouteReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	if err := service.NewLinkRouteService(helperPkg.GetHelper()).
		ReorderRoutes(shortLinkID, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c), &req); err != nil {
		ctrl.writeRouteError(c, err)
		return
	}
	ctrl.SuccessWithMessage(c, "排序已更新", nil)
}

func (ctrl LinkRouteController) TestRoute(c httpInterfaces.RouterContextInterface) {
	shortLinkID, ok := ctrl.parseShortLinkID(c)
	if !ok {
		return
	}
	var req dto.LinkRouteTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkRouteService(helperPkg.GetHelper()).
		TestRoute(shortLinkID, middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.writeRouteError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkRouteController) parseShortLinkID(c httpInterfaces.RouterContextInterface) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的短网址 ID")
		return 0, false
	}
	return id, true
}

func (ctrl LinkRouteController) parseRouteID(c httpInterfaces.RouterContextInterface) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("route_id"), 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的路由 ID")
		return 0, false
	}
	return id, true
}

func (ctrl LinkRouteController) writeRouteError(c httpInterfaces.RouterContextInterface, err error) {
	if strings.Contains(err.Error(), "不存在") {
		ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		return
	}
	ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
}
