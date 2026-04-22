package controller

import (
	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"

	"net/http"
	"strconv"
	"strings"

	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
)

type ShortLinkController struct {
	BaseResponse
}

// CreateShortLink 创建短网址
func (ctrl ShortLinkController) CreateShortLink(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	// 获取客户端IP
	clientIP := c.ClientIP()

	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.CreateShortLink(&req, clientIP)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
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
	response, err := shortLinkService.GetShortLink(id)
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
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.UpdateShortLink(id, &req)
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

// UpdateShortLinkStatus 更新短网址状态
func (ctrl ShortLinkController) UpdateShortLinkStatus(c httpInterfaces.RouterContextInterface) {
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
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.UpdateShortLinkStatus(id, req.IsActive)
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
	helper := helperPkg.GetHelper()
	_ = helper
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "无效的ID格式")
		return
	}
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	err = shortLinkService.DeleteShortLink(id)
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

// GetShortLinkList 获取短网址列表
func (ctrl ShortLinkController) GetShortLinkList(c httpInterfaces.RouterContextInterface) {
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.ShortLinkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.GetShortLinkList(&req)
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
	response, err := shortLinkService.GetShortLinkStatistics(id, days)
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
	helper := helperPkg.GetHelper()
	_ = helper
	var req dto.BatchCreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 获取客户端IP
	clientIP := c.ClientIP()
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	response, err := shortLinkService.BatchCreateShortLinks(&req, clientIP, helper)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}

	ctrl.Success(c, response)
}

// ErrorPageData 错误页面模板数据结构
type ErrorPageData struct {
	SiteName     string
	ICPNumber    string
	PoliceNumber string
	Domain       string
	Copyright    string
}

// AntiRedPageData 防红页面模板数据结构
type AntiRedPageData struct {
	SiteName     string
	ICPNumber    string
	PoliceNumber string
	Domain       string
	Copyright    string
	TargetURL    string
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
		ctrl.render404Page(c,c.Host())
		return
	}

	// 获取域名，从请求的Host头获取
	domain := c.Host()

	// 获取客户端信息
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// 获取查询参数字符串
	queryString := c.Request().URL.RawQuery
	shortLinkService := service.NewShortLinkService(helper, c.Request().Context())
	// 进行跳转（使用新的支持GET参数透传的方法）
	originalURL, err := shortLinkService.RedirectShortLinkWithQuery(domain, shortCode, clientIP, userAgent, referer, queryString)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			// 渲染404页面而不是返回JSON错误
			ctrl.render404Page(c,domain)
		} else if strings.Contains(err.Error(), "无效") {
			// 渲染过期页面而不是返回JSON错误
			ctrl.render404Page(c,domain)
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

	// 防红检查：如果域名启用了防红且为微信/QQ内置浏览器，则显示引导页
	if isWeChatOrQQBrowser(userAgent) {
		domainService := service.NewDomainService(helper)
		domainInfo, domainErr := domainService.GetDomainByName(domain)
		if domainErr == nil && domainInfo.EnableAntiRed != nil && *domainInfo.EnableAntiRed {
			ctrl.renderAntiRedPage(c, domainInfo, originalURL)
			return
		}
	}

	// 302重定向到原始URL
	c.Redirect(http.StatusFound, originalURL)
}

// renderAntiRedPage 渲染防红引导页面
func (ctrl ShortLinkController) renderAntiRedPage(c httpInterfaces.RouterContextInterface, domainInfo *model.Domain, targetURL string) {
	helper := helperPkg.GetHelper()
	siteName := helper.GetEnv().GetString("website.name", "短网址服务")
	if domainInfo.SiteName != "" {
		siteName = domainInfo.SiteName
	}

	pageData := AntiRedPageData{
		SiteName:     siteName,
		ICPNumber:    domainInfo.ICPNumber,
		PoliceNumber: domainInfo.PoliceNumber,
		Domain:       domainInfo.Domain,
		Copyright:    helper.GetEnv().GetString("website.copyright", ""),
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
	// 默认数据
	pageData := ErrorPageData{
		SiteName:     siteName,
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
		if strings.Contains(err.Error(), "不存在") {
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
