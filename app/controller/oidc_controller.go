package controller

import (
	"net/http"
	"net/url"
	"strconv"
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
// - login 流:302 到 admin SPA 的 SSO 接管页(#/auth/oidc-redirect),带 token + final_to
// - bind 流:302 到 {return_to 或 admin profile},带 oidc_bind=ok
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
		params := map[string]string{
			"token":      result.Token,
			"expires_at": strconv.FormatInt(result.ExpiresAt.Unix(), 10),
		}
		if result.ReturnTo != "" {
			params["final_to"] = result.ReturnTo
		}
		c.Redirect(http.StatusFound, appendHashAwareQuery(defaultLoginReturn(), params))
	case service.OIDCFlowBind:
		target := result.ReturnTo
		if target == "" {
			target = defaultBindReturn()
		}
		c.Redirect(http.StatusFound, appendHashAwareQuery(target, map[string]string{"oidc_bind": "ok"}))
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
	// 错误跳登录页 (不是 oidc-redirect),让用户可以立即重试。
	cfg := helperPkg.GetHelper().GetConfig()
	target := cfg.GetString("oidc.error_return_uri", "/admin/#/auth/login")
	c.Redirect(http.StatusFound, appendHashAwareQuery(target, map[string]string{"oidc_error": msg}))
}

// defaultLoginReturn 成功登录后 302 的目标。必须指向 SPA 的 SSO 接管页,
// 由其消费 token 并跳转到业务首页(或 final_to 指定的路径)。
func defaultLoginReturn() string {
	cfg := helperPkg.GetHelper().GetConfig()
	if v := cfg.GetString("oidc.login_return_uri", ""); v != "" {
		return v
	}
	return "/admin/#/auth/oidc-redirect"
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

// appendHashAwareQuery 在支持 hash 模式 SPA 路由的前提下追加 query 参数。
// - target 形如 "/admin/#/auth/oidc-redirect":把参数放进 hash 的 query
//   ("/admin/#/auth/oidc-redirect?token=xxx"),这样 Vue Router 读得到
// - target 无 "#":按标准 URL query 处理
// 早期实现用 url.Parse + RawQuery 放在真 query 里(`?...#...`),
// 对 hash 模式路由来说那段 query 读不到,token 实际上就丢了。
func appendHashAwareQuery(target string, params map[string]string) string {
	if target == "" || len(params) == 0 {
		return target
	}
	if idx := strings.Index(target, "#"); idx >= 0 {
		base := target[:idx]
		hash := target[idx+1:]
		hashPath := hash
		hashQuery := ""
		if q := strings.Index(hash, "?"); q >= 0 {
			hashPath = hash[:q]
			hashQuery = hash[q+1:]
		}
		values, _ := url.ParseQuery(hashQuery)
		for k, v := range params {
			values.Set(k, v)
		}
		encoded := values.Encode()
		if encoded == "" {
			return base + "#" + hashPath
		}
		return base + "#" + hashPath + "?" + encoded
	}

	u, err := url.Parse(target)
	if err != nil {
		return target
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
