package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

const (
	HeaderAppID     = "X-App-Id"
	HeaderSignature = "X-Signature"
	HeaderTimestamp = "X-Timestamp"
	HeaderNonce     = "X-Nonce"
)

// AuthMiddleware 三模式：签名认证 > JWT 登录 Token > API Bearer Token
func AuthMiddleware() httpInterfaces.HandlerFunc {
	return func(c httpInterfaces.RouterContextInterface) {
		if hasSignatureHeaders(c) {
			if err := validateSignatureAuth(c); err != nil {
				respondUnauthorized(c, err.Error())
				return
			}
			c.Next()
			return
		}

		if hasBearerToken(c) {
			token, err := extractBearerToken(c)
			if err != nil {
				respondUnauthorized(c, err.Error())
				return
			}

			if user, err := validateJWTToken(token); err == nil {
				c.Set("current_user", user)
				c.Next()
				return
			}

			user, err := validateAPIToken(token)
			if err != nil {
				respondUnauthorized(c, err.Error())
				return
			}
			c.Set("current_user", user)
			c.Next()
			return
		}

		respondUnauthorized(c, "缺少认证信息")
	}
}

func hasSignatureHeaders(c httpInterfaces.RouterContextInterface) bool {
	return c.GetHeader(HeaderAppID) != "" &&
		c.GetHeader(HeaderSignature) != "" &&
		c.GetHeader(HeaderTimestamp) != "" &&
		c.GetHeader(HeaderNonce) != ""
}

func hasBearerToken(c httpInterfaces.RouterContextInterface) bool {
	authHeader := c.GetHeader("Authorization")
	return authHeader != "" && strings.HasPrefix(authHeader, "Bearer ")
}

func validateSignatureAuth(c httpInterfaces.RouterContextInterface) error {
	appID := c.GetHeader(HeaderAppID)
	signature := c.GetHeader(HeaderSignature)
	timestampStr := c.GetHeader(HeaderTimestamp)
	nonce := c.GetHeader(HeaderNonce)

	signatureHelper := helper.GetSignatureHelper()
	if !signatureHelper.ValidateNonce(nonce) {
		return errors.New("缺少认证信息")
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return errors.New("时间戳无效")
	}
	if !signatureHelper.ValidateTimestamp(timestamp, time.Now().Unix()) {
		return errors.New("时间戳无效")
	}

	tokenDAO := dao.NewUserTokenDAO(helper.GetHelper())
	token, err := tokenDAO.GetByAppID(appID)
	if err != nil {
		return errors.New("无效的AppID")
	}
	if !token.IsActive() {
		return errors.New("Token已禁用")
	}
	if !token.User.IsActive() {
		return errors.New("用户已被禁用")
	}

	decryptedSecret, err := signatureHelper.DecryptAppSecret(token.AppSecret)
	if err != nil {
		return errors.New("签名验证失败")
	}

	params := extractRequestParams(c)
	if !signatureHelper.VerifySignature(decryptedSecret, c.Method(), c.Path(), params, timestamp, nonce, signature) {
		return errors.New("签名验证失败")
	}

	token.UpdateLastUsed()
	_ = tokenDAO.Update(token)
	c.Set("current_user", &token.User)
	return nil
}

func extractRequestParams(c httpInterfaces.RouterContextInterface) map[string]any {
	params := make(map[string]any)
	for key, values := range c.Request().URL.Query() {
		if len(values) == 1 {
			params[key] = values[0]
		} else {
			params[key] = values
		}
	}

	method := c.Method()
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			if body := c.Request().Body; body != nil {
				bodyBytes, err := io.ReadAll(body)
				if err == nil && len(bodyBytes) > 0 {
					c.Request().Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
					var bodyParams map[string]any
					if err := json.Unmarshal(bodyBytes, &bodyParams); err == nil {
						for key, value := range bodyParams {
							params[key] = value
						}
					}
				}
			}
		}
	}
	return params
}

func extractBearerToken(c httpInterfaces.RouterContextInterface) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("缺少认证信息")
	}
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}
	return "", errors.New("Token格式错误")
}

func validateAPIToken(tokenString string) (*model.User, error) {
	tokenDAO := dao.NewUserTokenDAO(helper.GetHelper())
	token, err := tokenDAO.GetByToken(tokenString)
	if err != nil {
		return nil, err
	}
	if !token.IsActive() {
		return nil, errors.New("Token已失效")
	}
	if !token.User.IsActive() {
		return nil, errors.New("用户已被禁用")
	}
	token.UpdateLastUsed()
	_ = tokenDAO.Update(token)
	return &token.User, nil
}

func validateJWTToken(tokenString string) (*model.User, error) {
	cfg := helper.GetHelper().GetConfig()
	jwtSecret := cfg.GetString("jwt.secret", "")
	if jwtSecret == "" {
		return nil, errors.New("JWT not configured")
	}
	helper.InitJWTHelper(jwtSecret, cfg.GetInt("jwt.expire_hours", 24))

	claims, err := helper.GetJWTHelper().ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	userDAO := dao.NewUserDAO(helper.GetHelper())
	user, err := userDAO.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if !user.IsActive() {
		return nil, errors.New("用户已被禁用")
	}
	return user, nil
}

func respondUnauthorized(c httpInterfaces.RouterContextInterface, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
		"code":    constants.ErrCodeUnauthorized,
		"message": message,
	})
}

// GetCurrentUser returns the authenticated user previously stored in the
// request context. Returns nil if absent.
func GetCurrentUser(c httpInterfaces.RouterContextInterface) *model.User {
	if v := c.Get("current_user"); v != nil {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return nil
}

func GetCurrentUserID(c httpInterfaces.RouterContextInterface) uint64 {
	if u := GetCurrentUser(c); u != nil {
		return u.ID
	}
	return 0
}
