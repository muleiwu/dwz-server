package controller

import (
	"net/http"
	"net/url"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/middleware"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

// OIDCController 面向终端用户的 OIDC 入口:登录授权跳转、回调、绑定、解绑。
type OIDCController struct {
	BaseResponse
}

// Authorize 发起登录授权流,重定向到 IdP。
// 免认证接口,可选 query 参数:return_to(登录成功后前端跳回的路径)。
func (ctrl OIDCController) Authorize(c httpInterfaces.RouterContextInterface) {
	svc := service.NewOIDCService(helperPkg.GetHelper())
	returnTo := sanitizeReturnTo(c.Query("return_to"))
	authURL, err := svc.BuildAuthURL(c.Request().Context(), service.OIDCFlowLogin, 0, returnTo)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// Bind 已登录用户发起绑定授权流。需要认证,返回 JSON { auth_url } 由前端跳转。
func (ctrl OIDCController) Bind(c httpInterfaces.RouterContextInterface) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}
	svc := service.NewOIDCService(helperPkg.GetHelper())
	returnTo := sanitizeReturnTo(c.Query("return_to"))
	authURL, err := svc.BuildAuthURL(c.Request().Context(), service.OIDCFlowBind, userID, returnTo)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeBadRequest, err.Error())
		return
	}
	ctrl.Success(c, map[string]any{"auth_url": authURL})
}

// Callback IdP 授权码回调,免认证。完成后:
// - login 流:302 到 {return_to 或 admin 首页}?token=...
// - bind 流:302 到 {return_to 或 admin profile}?oidc_bind=ok
func (ctrl OIDCController) Callback(c httpInterfaces.RouterContextInterface) {
	code := c.Query("code")
	state := c.Query("state")
	if errParam := c.Query("error"); errParam != "" {
		ctrl.redirectWithError(c, c.Query("error_description"))
		return
	}

	svc := service.NewOIDCService(helperPkg.GetHelper())
	result, err := svc.HandleCallback(c.Request().Context(), code, state)
	if err != nil {
		ctrl.redirectWithError(c, err.Error())
		return
	}

	switch result.Flow {
	case service.OIDCFlowLogin:
		target := result.ReturnTo
		if target == "" {
			target = defaultLoginReturn()
		}
		c.Redirect(http.StatusFound, service.AppendReturnToken(target, result.Token, result.ExpiresAt))
	case service.OIDCFlowBind:
		target := result.ReturnTo
		if target == "" {
			target = defaultBindReturn()
		}
		c.Redirect(http.StatusFound, appendQuery(target, "oidc_bind", "ok"))
	default:
		ctrl.Error(c, constants.ErrCodeInternal, "未知授权流")
	}
}

// Unbind 解除当前用户与指定 provider 的绑定。
func (ctrl OIDCController) Unbind(c httpInterfaces.RouterContextInterface) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}
	provider := c.Param("provider")
	if provider == "" {
		provider = c.Query("provider")
	}
	svc := service.NewOIDCService(helperPkg.GetHelper())
	if err := svc.Unbind(userID, provider); err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, nil)
}

// GetMyBindings 返回当前用户的所有 OIDC 绑定关系。
func (ctrl OIDCController) GetMyBindings(c httpInterfaces.RouterContextInterface) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		ctrl.Error(c, constants.ErrCodeUnauthorized, "用户未登录")
		return
	}
	svc := service.NewOIDCService(helperPkg.GetHelper())
	list, err := svc.ListUserBindings(userID)
	if err != nil {
		ctrl.Error(c, constants.ErrCodeInternal, err.Error())
		return
	}
	ctrl.Success(c, list)
}

// redirectWithError 登录失败时跳回登录页,带上可读 error 参数供前端 Toast。
func (ctrl OIDCController) redirectWithError(c httpInterfaces.RouterContextInterface, msg string) {
	target := defaultLoginReturn()
	c.Redirect(http.StatusFound, appendQuery(target, "oidc_error", msg))
}

func defaultLoginReturn() string {
	cfg := helperPkg.GetHelper().GetConfig()
	if v := cfg.GetString("oidc.login_return_uri", ""); v != "" {
		return v
	}
	return "/admin/#/login"
}

func defaultBindReturn() string {
	cfg := helperPkg.GetHelper().GetConfig()
	if v := cfg.GetString("oidc.bind_return_uri", ""); v != "" {
		return v
	}
	return "/admin/#/profile"
}

// sanitizeReturnTo 仅接受相对路径(以 / 开头且不以 // 开头),
// 防止开放重定向到外部站点。
func sanitizeReturnTo(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "//") || !strings.HasPrefix(raw, "/") {
		return ""
	}
	return raw
}

func appendQuery(target, key, value string) string {
	if target == "" {
		return target
	}
	u, err := url.Parse(target)
	if err != nil {
		return target
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}
