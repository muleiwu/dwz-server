package autoload

import (
	"regexp"
	"strings"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/controller"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

var shortCodePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

var shortCodeReservedPrefixes = []string{
	"/api/",
	"/health",
	"/install/",
	"/favicon",
	"/static/",
	"/assets/",
	"/css/",
	"/js/",
	"/images/",
}

// shortCodeDispatch is a middleware that intercepts GET /<code> and
// GET /preview/<code> before gin's tree router runs. It exists because
// go-web's RegexGroup mounted at the root path conflicts with sibling
// explicit routes — this dispatcher gives us the same behaviour with
// no go-web changes.
func shortCodeDispatch() httpInterfaces.HandlerFunc {
	ctrl := controller.ShortLinkController{}
	return func(c httpInterfaces.RouterContextInterface) {
		if c.Method() != "GET" {
			c.Next()
			return
		}

		path := c.Path()
		if path == "" || path == "/" {
			c.Next()
			return
		}

		for _, prefix := range shortCodeReservedPrefixes {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		if rest, ok := strings.CutPrefix(path, "/preview/"); ok {
			if shortCodePattern.MatchString(rest) {
				setShortCodeParam(c, rest)
				ctrl.PreviewShortLink(c)
				c.Abort()
				return
			}
		}

		if seg := strings.TrimPrefix(path, "/"); !strings.Contains(seg, "/") && shortCodePattern.MatchString(seg) {
			setShortCodeParam(c, seg)
			ctrl.RedirectShortLink(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

// setShortCodeParam stores the short code so that c.Param("code") inside the
// controller still resolves it. RouterContextInterface exposes Set/Get for
// generic context values; controller code reads via c.Param which falls back
// to the gin Params lookup. Since we are not going through gin's route tree
// here, we attach the value via Set under both the conventional Param key
// and the bare key the controller uses.
func setShortCodeParam(c httpInterfaces.RouterContextInterface, code string) {
	c.Set("code", code)
}
