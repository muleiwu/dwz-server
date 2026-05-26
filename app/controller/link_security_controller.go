package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type LinkSecurityController struct {
	BaseResponse
}

func (ctrl LinkSecurityController) GetShortLinkSecurity(c httpInterfaces.RouterContextInterface) {
	id, ok := parseUintParam(c, "id", ctrl.BaseResponse)
	if !ok {
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).GetSecurity(id, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.writeSecurityError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) UpdateShortLinkSecurity(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限更新链接安全配置")
		return
	}
	id, ok := parseUintParam(c, "id", ctrl.BaseResponse)
	if !ok {
		return
	}
	var req dto.LinkSecurityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).
		UpsertSecurity(id, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c), &req)
	if err != nil {
		ctrl.writeSecurityError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) RescanShortLink(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限重扫链接")
		return
	}
	id, ok := parseUintParam(c, "id", ctrl.BaseResponse)
	if !ok {
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).
		RescanShortLink(id, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.writeSecurityError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) ListURLRules(c httpInterfaces.RouterContextInterface) {
	var req dto.SecurityURLRuleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).ListURLRules(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) CreateURLRule(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限创建安全规则")
		return
	}
	var req dto.SecurityURLRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).
		CreateURLRule(middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c), &req)
	if err != nil {
		ctrl.writeSecurityError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) UpdateURLRule(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限更新安全规则")
		return
	}
	id, ok := parseUintParam(c, "id", ctrl.BaseResponse)
	if !ok {
		return
	}
	var req dto.SecurityURLRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).
		UpdateURLRule(id, middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.writeSecurityError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) DeleteURLRule(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限删除安全规则")
		return
	}
	id, ok := parseUintParam(c, "id", ctrl.BaseResponse)
	if !ok {
		return
	}
	if err := service.NewLinkSecurityService(helperPkg.GetHelper()).DeleteURLRule(id, middleware.GetCurrentWorkspaceID(c)); err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.SuccessWithMessage(c, "删除成功", nil)
}

func (ctrl LinkSecurityController) ListEvents(c httpInterfaces.RouterContextInterface) {
	var req dto.SecurityEventListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).ListEvents(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) ListAbuseReports(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限查看举报")
		return
	}
	var req dto.AbuseReportListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).ListAbuseReports(middleware.GetCurrentWorkspaceID(c), &req)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) UpdateAbuseReport(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageAdminResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限处理举报")
		return
	}
	id, ok := parseUintParam(c, "id", ctrl.BaseResponse)
	if !ok {
		return
	}
	var req dto.AbuseReportUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).
		UpdateAbuseReport(id, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c), &req)
	if err != nil {
		ctrl.writeSecurityError(c, err)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) SubmitPassword(c httpInterfaces.RouterContextInterface) {
	var req dto.PublicPasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	cookieName, cookieValue, maxAge, err := service.NewLinkSecurityService(helperPkg.GetHelper()).
		VerifyPassword(req.Domain, req.ShortCode, req.Password, c.ClientIP(), c.UserAgent())
	if err != nil {
		if isBrowserForm(c) {
			ShortLinkController{}.renderPasswordPage(c, req.Domain, req.ShortCode, "访问密码错误")
			return
		}
		ctrl.Error(c, constants.ErrCodeForbidden, err.Error())
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(cookieName, cookieValue, maxAge, "/", "", c.Request().TLS != nil, true)
	if isBrowserForm(c) {
		next := req.Next
		if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
			next = "/" + req.ShortCode
		}
		c.Redirect(http.StatusFound, next)
		return
	}
	ctrl.Success(c, map[string]any{"verified": true})
}

func (ctrl LinkSecurityController) CreatePublicAbuseReport(c httpInterfaces.RouterContextInterface) {
	var req dto.AbuseReportCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	resp, err := service.NewLinkSecurityService(helperPkg.GetHelper()).CreateAbuseReport(&req, c.ClientIP(), c.UserAgent())
	if err != nil {
		if isBrowserForm(c) {
			ShortLinkController{}.renderSecurityMessagePage(c, req.Domain, "举报未提交", err.Error(), http.StatusBadRequest)
			return
		}
		ctrl.writeSecurityError(c, err)
		return
	}
	if isBrowserForm(c) {
		ShortLinkController{}.renderSecurityMessagePage(c, req.Domain, "举报已提交", "我们已收到您的反馈，将由管理员人工审核。", http.StatusOK)
		return
	}
	ctrl.Success(c, resp)
}

func (ctrl LinkSecurityController) writeSecurityError(c httpInterfaces.RouterContextInterface, err error) {
	if strings.Contains(err.Error(), "不存在") {
		ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		return
	}
	if strings.Contains(err.Error(), "无权限") || strings.Contains(err.Error(), "未开启") || errors.Is(err, service.ErrSecurityAccessDenied) {
		ctrl.Error(c, constants.ErrCodeForbidden, err.Error())
		return
	}
	ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
}

func parseUintParam(c httpInterfaces.RouterContextInterface, name string, responder BaseResponse) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		responder.Error(c, constants.ErrCodeBadRequest, "无效的ID")
		return 0, false
	}
	return id, true
}

func isBrowserForm(c httpInterfaces.RouterContextInterface) bool {
	ct := strings.ToLower(c.ContentType())
	return strings.Contains(ct, "form")
}
