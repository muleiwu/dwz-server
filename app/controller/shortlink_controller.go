package controller

import (
	"errors"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"

	"net/http"
	"strconv"
	"strings"

	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type ShortLinkController struct {
	BaseResponse
}

// CreateShortLink 创建短网址
func (ctrl ShortLinkController) CreateShortLink(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限创建短网址")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, bindErrorMessage(err))
		return
	}
	// 获取客户端IP
	clientIP := c.ClientIP()

	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.CreateShortLinkInWorkspace(&req, clientIP, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c))
	if err != nil {
		ctrl.writeShortLinkError(c, err)
		return
	}

	ctrl.Success(c, response)
}

// GetShortLink 获取短网址详情
func (ctrl ShortLinkController) GetShortLink(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}

	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.GetShortLinkInWorkspace(id, middleware.GetCurrentWorkspaceID(c))
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

// UpdateShortLink 更新短网址
func (ctrl ShortLinkController) UpdateShortLink(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限更新短网址")
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

	var req dto.UpdateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, bindErrorMessage(err))
		return
	}
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.UpdateShortLinkInWorkspace(id, &req, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c))
	if err != nil {
		ctrl.writeShortLinkError(c, err)
		return
	}

	ctrl.Success(c, response)
}

// UpdateShortLinkStatus 更新短网址状态
func (ctrl ShortLinkController) UpdateShortLinkStatus(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限更新短网址")
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

	var req dto.UpdateShortLinkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, bindErrorMessage(err))
		return
	}

	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.UpdateShortLinkStatusInWorkspace(id, req.IsActive, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c))
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

// DeleteShortLink 删除短网址
func (ctrl ShortLinkController) DeleteShortLink(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限删除短网址")
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
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	err = shortLinkService.DeleteShortLinkInWorkspace(id, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, err.Error())
		} else if strings.Contains(err.Error(), "请先禁用") {
			ctrl.Error(c, constants.ErrCodeForbidden, err.Error())
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	ctrl.SuccessWithMessage(c, "删除成功", nil)
}

// BatchUpdateShortLinkStatus 批量更新短网址状态
func (ctrl ShortLinkController) BatchUpdateShortLinkStatus(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限更新短网址")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper

	var req dto.BatchUpdateShortLinkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, bindErrorMessage(err))
		return
	}

	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.BatchUpdateShortLinkStatusInWorkspace(req.IDs, *req.IsActive, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// BatchDeleteShortLinks 批量删除短网址
func (ctrl ShortLinkController) BatchDeleteShortLinks(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限删除短网址")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper

	var req dto.BatchShortLinkIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, bindErrorMessage(err))
		return
	}

	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.BatchDeleteShortLinksInWorkspace(req.IDs, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetShortLinkList 获取短网址列表
func (ctrl ShortLinkController) GetShortLinkList(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.ShortLinkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, bindErrorMessage(err))
		return
	}
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.GetShortLinkListInWorkspace(&req, middleware.GetCurrentWorkspaceID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// GetShortLinkStatistics 获取短网址统计信息
func (ctrl ShortLinkController) GetShortLinkStatistics(c httpInterfaces.RouterContextInterface) {
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
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.GetShortLinkStatisticsInWorkspace(id, days, middleware.GetCurrentWorkspaceID(c))
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

// BatchCreateShortLinks 批量创建短网址
func (ctrl ShortLinkController) BatchCreateShortLinks(c httpInterfaces.RouterContextInterface) {
	if !middleware.CanManageBusinessResource(c) {
		ctrl.Error(c, constants.ErrCodeForbidden, "无权限创建短网址")
		return
	}
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.BatchCreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, bindErrorMessage(err))
		return
	}

	// 获取客户端IP
	clientIP := c.ClientIP()
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.BatchCreateShortLinksInWorkspace(&req, clientIP, helper, middleware.GetCurrentWorkspaceID(c), middleware.GetCurrentUserID(c))
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// ErrorPageData 错误页面模板数据结构
type ErrorPageData struct {
	SiteName     string
	LogoURL      string
	ICPNumber    string
	PoliceNumber string
	Domain       string
	Copyright    string
}

// AntiRedPageData 防红页面模板数据结构
type AntiRedPageData struct {
	SiteName     string
	LogoURL      string
	ICPNumber    string
	PoliceNumber string
	Domain       string
	Copyright    string
	TargetURL    string
}

type SecurityPageData struct {
	SiteName     string
	LogoURL      string
	ICPNumber    string
	PoliceNumber string
	Domain       string
	Copyright    string
	ShortCode    string
	Message      string
	Title        string
	Next         string
	ReportURL    string
}

// paramOrCtx fetches a path parameter from the router, falling back to the
// request-scoped context Set by middleware-driven dispatch (used by the
// short-code interceptor in config/autoload/short_code_dispatch.go).
func paramOrCtx(c httpInterfaces.RouterContextInterface, key string) string {
	if v := c.Param(key); v != "" {
		return v
	}
	return c.GetString(key)
}

// isWeChatOrQQBrowser 检测是否为微信或QQ内置浏览器
func isWeChatOrQQBrowser(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	return strings.Contains(ua, "micromessenger") || strings.Contains(ua, "qq/")
}

// RedirectShortLink 短网址跳转
func (ctrl ShortLinkController) RedirectShortLink(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	shortCode := paramOrCtx(c, "code")
	if shortCode == "" {
		// 当shortCode为空时，渲染404页面
		ctrl.render404Page(c, c.Host())
		return
	}

	// 获取域名，从请求的Host头获取
	domain := c.Host()

	// 获取客户端信息
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")
	acceptLanguage := c.GetHeader("Accept-Language")

	// 获取查询参数字符串
	queryString := c.Request().URL.RawQuery
	linkSecurityService := service.NewLinkSecurityService(helper)
	accessToken := ""
	if cookie, cookieErr := c.Cookie(linkSecurityService.AccessCookieName(domain, shortCode)); cookieErr == nil {
		accessToken = cookie
	}
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	decision, err := shortLinkService.ResolveRedirectWithSecurityAndLanguage(domain, shortCode, clientIP, userAgent, referer, queryString, accessToken, acceptLanguage)
	if err != nil {
		if errors.Is(err, service.ErrSecurityPasswordRequired) {
			ctrl.renderPasswordPage(c, domain, shortCode, "")
		} else if errors.Is(err, service.ErrSecurityURLBlocked) {
			ctrl.renderSecurityBlockedPage(c, domain, "链接存在安全风险", "该短链接暂时无法访问，请联系链接管理员确认。", http.StatusForbidden)
		} else if errors.Is(err, service.ErrSecurityAccessDenied) {
			ctrl.renderSecurityBlockedPage(c, domain, "访问受限", "当前访问不符合该短链接的安全策略。", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "不存在") {
			// 渲染404页面而不是返回JSON错误
			ctrl.render404Page(c, domain)
		} else if strings.Contains(err.Error(), "无效") {
			// 渲染过期页面而不是返回JSON错误
			ctrl.render404Page(c, domain)
		} else if strings.Contains(err.Error(), "过期") {
			// 渲染过期页面而不是返回JSON错误
			ctrl.renderExpiredPage(c, domain)
		} else if strings.Contains(err.Error(), "禁用") {
			// 渲染禁用页面而不是返回JSON错误
			ctrl.renderDisabledPage(c, domain)
		} else {
			// 渲染通用错误页面而不是返回JSON错误
			ctrl.renderInternalErrorPage(c, domain)
		}
		return
	}
	originalURL := decision.TargetURL

	// 防红检查：如果域名启用了防红且为微信/QQ内置浏览器，则显示引导页
	if isWeChatOrQQBrowser(userAgent) {
		domainService := service.NewDomainService(helper)
		domainInfo, domainErr := domainService.GetDomainByName(domain)
		if domainErr == nil && domainInfo.EnableAntiRed != nil && *domainInfo.EnableAntiRed {
			ctrl.renderAntiRedPage(c, domainInfo, originalURL)
			return
		}
	}

	statusCode := decision.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusFound
	}
	c.Redirect(statusCode, originalURL)
}

// renderAntiRedPage 渲染防红引导页面
func (ctrl ShortLinkController) renderAntiRedPage(c httpInterfaces.RouterContextInterface, domainInfo *model.Domain, targetURL string) {
	helper := helperPkg.GetHelper()
	siteName := helper.GetEnv().GetString("website.name", "短网址服务")
	copyright := helper.GetEnv().GetString("website.copyright", "")
	logoURL := ""
	if branding, brandingErr := service.NewBrandingService(helper).GetPublicBranding(domainInfo.Domain); brandingErr == nil {
		if branding.BrandName != "" {
			siteName = branding.BrandName
		}
		logoURL = branding.LogoURL
		copyright = branding.CopyrightText
	}
	if domainInfo.SiteName != "" {
		siteName = domainInfo.SiteName
	}

	pageData := AntiRedPageData{
		SiteName:     siteName,
		LogoURL:      logoURL,
		ICPNumber:    domainInfo.ICPNumber,
		PoliceNumber: domainInfo.PoliceNumber,
		Domain:       domainInfo.Domain,
		Copyright:    copyright,
		TargetURL:    targetURL,
	}

	c.HTML(http.StatusOK, "anti_red.html", pageData)
}

// render404Page 渲染404页面
func (ctrl ShortLinkController) render404Page(c httpInterfaces.RouterContextInterface, domain string) {
	ctrl.renderErrorPage(c, domain, "404.html", http.StatusNotFound)
}

// renderExpiredPage 渲染过期页面
func (ctrl ShortLinkController) renderExpiredPage(c httpInterfaces.RouterContextInterface, domain string) {
	ctrl.renderErrorPage(c, domain, "expired.html", 410) // 410 Gone
}

// renderDisabledPage 渲染禁用页面
func (ctrl ShortLinkController) renderDisabledPage(c httpInterfaces.RouterContextInterface, domain string) {
	ctrl.renderErrorPage(c, domain, "disabled.html", http.StatusForbidden)
}

// renderInternalErrorPage 渲染通用错误页面
func (ctrl ShortLinkController) renderInternalErrorPage(c httpInterfaces.RouterContextInterface, domain string) {
	ctrl.renderErrorPage(c, domain, "error.html", http.StatusInternalServerError)
}

// renderErrorPage 通用错误页面渲染方法
func (ctrl ShortLinkController) renderErrorPage(c httpInterfaces.RouterContextInterface, domain, template string, statusCode int) {
	helper := helperPkg.GetHelper()
	// 获取域名信息
	domainService := service.NewDomainService(helper)
	domainInfo, err := domainService.GetDomainByName(domain)

	siteName := helper.GetEnv().GetString("website.name", "短网址服务")
	copyright := helper.GetEnv().GetString("website.copyright", "")
	logoURL := ""
	if branding, brandingErr := service.NewBrandingService(helper).GetPublicBranding(domain); brandingErr == nil {
		if branding.BrandName != "" {
			siteName = branding.BrandName
		}
		logoURL = branding.LogoURL
		copyright = branding.CopyrightText
	}
	// 默认数据
	pageData := ErrorPageData{
		SiteName:     siteName,
		LogoURL:      logoURL,
		ICPNumber:    "",
		PoliceNumber: "",
		Domain:       domain,
		Copyright:    copyright,
	}

	if err == nil {
		pageData.SiteName = domainInfo.SiteName
		if pageData.SiteName == "" {
			pageData.SiteName = "短网址服务"
		}
		pageData.ICPNumber = domainInfo.ICPNumber
		pageData.PoliceNumber = domainInfo.PoliceNumber
	}

	// 渲染错误页面
	c.HTML(statusCode, template, pageData)
}

func (ctrl ShortLinkController) renderPasswordPage(c httpInterfaces.RouterContextInterface, domain, shortCode, message string) {
	ctrl.renderSecurityPage(c, domain, "password.html", http.StatusOK, SecurityPageData{
		Title:     "需要访问密码",
		ShortCode: shortCode,
		Message:   message,
		Next:      c.Request().URL.RequestURI(),
	})
}

func (ctrl ShortLinkController) renderSecurityBlockedPage(c httpInterfaces.RouterContextInterface, domain, title, message string, statusCode int) {
	ctrl.renderSecurityPage(c, domain, "security_blocked.html", statusCode, SecurityPageData{
		Title:   title,
		Message: message,
	})
}

func (ctrl ShortLinkController) renderSecurityMessagePage(c httpInterfaces.RouterContextInterface, domain, title, message string, statusCode int) {
	ctrl.renderSecurityPage(c, domain, "security_message.html", statusCode, SecurityPageData{
		Title:   title,
		Message: message,
	})
}

func (ctrl ShortLinkController) renderSecurityPage(c httpInterfaces.RouterContextInterface, domain, template string, statusCode int, pageData SecurityPageData) {
	helper := helperPkg.GetHelper()
	domainService := service.NewDomainService(helper)
	domainInfo, err := domainService.GetDomainByName(domain)
	pageData.SiteName = helper.GetEnv().GetString("website.name", "短网址服务")
	pageData.Copyright = helper.GetEnv().GetString("website.copyright", "")
	pageData.Domain = domain
	if branding, brandingErr := service.NewBrandingService(helper).GetPublicBranding(domain); brandingErr == nil {
		if branding.BrandName != "" {
			pageData.SiteName = branding.BrandName
		}
		pageData.LogoURL = branding.LogoURL
		pageData.Copyright = branding.CopyrightText
	}
	if err == nil {
		if domainInfo.SiteName != "" {
			pageData.SiteName = domainInfo.SiteName
		}
		pageData.ICPNumber = domainInfo.ICPNumber
		pageData.PoliceNumber = domainInfo.PoliceNumber
	}
	c.HTML(statusCode, template, pageData)
}

// PreviewShortLink 预览短网址信息（不计入统计）
func (ctrl ShortLinkController) PreviewShortLink(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	shortCode := paramOrCtx(c, "code")
	if shortCode == "" {
		ctrl.Error(c, constants.ErrCodeBadRequest, "短网址代码不能为空")
		return
	}

	// 获取域名
	domain := "http://" + c.Host()
	if c.Request().TLS != nil {
		domain = "https://" + c.Host()
	}
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	// 预览时不传递IP信息，这样就不会记录统计，也不传递查询参数
	originalURL, err := shortLinkService.RedirectShortLinkWithQuery(domain, shortCode, "", "", "", "")
	if err != nil {
		if errors.Is(err, service.ErrSecurityPasswordRequired) {
			ctrl.Error(c, constants.ErrCodeForbidden, "短网址受访问密码保护")
		} else if errors.Is(err, service.ErrSecurityAccessDenied) || errors.Is(err, service.ErrSecurityURLBlocked) {
			ctrl.Error(c, constants.ErrCodeForbidden, "短网址访问受限")
		} else if strings.Contains(err.Error(), "不存在") {
			ctrl.Error(c, constants.ErrCodeNotFound, "短网址不存在或已失效")
		} else if strings.Contains(err.Error(), "过期") {
			ctrl.Error(c, 410, "短网址已过期")
		} else if strings.Contains(err.Error(), "禁用") {
			ctrl.Error(c, constants.ErrCodeForbidden, "短网址已被禁用")
		} else {
			ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		}
		return
	}

	// 返回预览信息
	previewInfo := map[string]interface{}{
		"short_code":   shortCode,
		"domain":       domain,
		"short_url":    domain + "/" + shortCode,
		"original_url": originalURL,
	}

	ctrl.Success(c, previewInfo)
}

func (ctrl ShortLinkController) writeShortLinkError(c httpInterfaces.RouterContextInterface, err error) {
	message := err.Error()
	if strings.Contains(message, "短网址不存在") {
		ctrl.Error(c, constants.ErrCodeNotFound, message)
		return
	}
	if isShortLinkBadRequestError(message) {
		ctrl.Error(c, constants.ErrCodeBadRequest, message)
		return
	}
	ctrl.Error(c, constants.ErrCodeInternal, message)
}

func isShortLinkBadRequestError(message string) bool {
	badRequestPhrases := []string{
		"无效",
		"不能为空",
		"不存在",
		"未激活",
		"命中安全规则",
		"仅支持",
		"已存在",
	}
	for _, phrase := range badRequestPhrases {
		if strings.Contains(message, phrase) {
			return true
		}
	}
	return false
}
