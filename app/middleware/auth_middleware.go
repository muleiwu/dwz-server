package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/constants"
	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/helper"
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

// 签名认证相关的请求头
const (
	HeaderAppID     = "X-App-Id"
	HeaderSignature = "X-Signature"
	HeaderTimestamp = "X-Timestamp"
	HeaderNonce     = "X-Nonce"
)

// AuthMiddleware 用户认证中间件
// 支持双模式认证：签名认证 > Bearer Token
func AuthMiddleware(helperInstance envInterface.HelperInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 优先尝试签名认证
		if hasSignatureHeaders(c) {
			if err := validateSignatureAuth(c, helperInstance); err != nil {
				respondUnauthorized(c, err.Error())
				return
			}
			c.Next()
			return
		}

		// 2. 尝试 Bearer Token 认证
		if hasBearerToken(c) {
			token, err := extractBearerToken(c)
			if err != nil {
				respondUnauthorized(c, err.Error())
				return
			}

			user, err := validateToken(token, helperInstance)
			if err != nil {
				respondUnauthorized(c, err.Error())
				return
			}

			c.Set("current_user", user)
			c.Next()
			return
		}

		// 3. 无认证信息
		respondUnauthorized(c, "缺少认证信息")
	}
}

// hasSignatureHeaders 检查请求是否包含签名认证所需的所有头
func hasSignatureHeaders(c *gin.Context) bool {
	appID := c.GetHeader(HeaderAppID)
	signature := c.GetHeader(HeaderSignature)
	timestamp := c.GetHeader(HeaderTimestamp)
	nonce := c.GetHeader(HeaderNonce)

	return appID != "" && signature != "" && timestamp != "" && nonce != ""
}

// hasBearerToken 检查请求是否包含 Bearer Token
func hasBearerToken(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	return authHeader != "" && strings.HasPrefix(authHeader, "Bearer ")
}

// validateSignatureAuth 验证签名认证
func validateSignatureAuth(c *gin.Context, helperInstance envInterface.HelperInterface) error {
	// 获取签名认证头
	appID := c.GetHeader(HeaderAppID)
	signature := c.GetHeader(HeaderSignature)
	timestampStr := c.GetHeader(HeaderTimestamp)
	nonce := c.GetHeader(HeaderNonce)

	// 验证 nonce 非空
	signatureHelper := helper.GetSignatureHelper()
	if !signatureHelper.ValidateNonce(nonce) {
		return errors.New("缺少认证信息")
	}

	// 解析时间戳
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return errors.New("时间戳无效")
	}

	// 验证时间戳窗口（±5分钟）
	currentTime := time.Now().Unix()
	if !signatureHelper.ValidateTimestamp(timestamp, currentTime) {
		return errors.New("时间戳无效")
	}

	// 根据 AppID 查询 Token
	tokenDAO := dao.NewUserTokenDAO(helperInstance)
	token, err := tokenDAO.GetByAppID(appID)
	if err != nil {
		return errors.New("无效的AppID")
	}

	// 检查 Token 状态
	if !token.IsActive() {
		return errors.New("Token已禁用")
	}

	// 检查用户状态
	if !token.User.IsActive() {
		return errors.New("用户已被禁用")
	}

	// 解密 App Secret
	decryptedSecret, err := signatureHelper.DecryptAppSecret(token.AppSecret)
	if err != nil {
		return errors.New("签名验证失败")
	}

	// 获取请求参数
	params := extractRequestParams(c)

	// 验证签名
	method := c.Request.Method
	path := c.Request.URL.Path
	if !signatureHelper.VerifySignature(decryptedSecret, method, path, params, timestamp, nonce, signature) {
		return errors.New("签名验证失败")
	}

	// 更新最后使用时间
	token.UpdateLastUsed()
	tokenDAO.Update(token)

	// 设置用户上下文
	c.Set("current_user", &token.User)

	return nil
}

// extractRequestParams 从请求中提取参数用于签名验证
func extractRequestParams(c *gin.Context) map[string]interface{} {
	params := make(map[string]interface{})

	// 1. 提取 URL 查询参数
	for key, values := range c.Request.URL.Query() {
		if len(values) == 1 {
			params[key] = values[0]
		} else {
			params[key] = values
		}
	}

	// 2. 提取 JSON Body 参数（仅对 POST/PUT/PATCH 请求）
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// 读取 body
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// 重新设置 body 以便后续处理
				c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

				// 解析 JSON
				var bodyParams map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyParams); err == nil {
					for key, value := range bodyParams {
						params[key] = value
					}
				}
			}
		}
	}

	return params
}

// extractBearerToken 从请求中提取 Bearer Token
func extractBearerToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("缺少认证信息")
	}

	// 解析 Bearer Token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	return "", errors.New("Token格式错误")
}

// validateToken 验证Token并返回用户信息
func validateToken(tokenString string, helperInstance envInterface.HelperInterface) (*model.User, error) {
	// 尝试验证API Token
	if user, err := validateAPIToken(tokenString, helperInstance); err == nil {
		return user, nil
	}

	// 尝试验证登录Token
	return validateLoginToken(tokenString, helperInstance)
}

// validateAPIToken 验证API Token
func validateAPIToken(tokenString string, helperInstance envInterface.HelperInterface) (*model.User, error) {
	tokenDAO := dao.NewUserTokenDAO(helperInstance)
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

	// 更新最后使用时间
	token.UpdateLastUsed()
	tokenDAO.Update(token)

	return &token.User, nil
}

// validateLoginToken 验证登录Token（简化版）
func validateLoginToken(tokenString string, helperInstance envInterface.HelperInterface) (*model.User, error) {
	if !strings.HasPrefix(tokenString, "user_") {
		return nil, errors.New("无效的Token格式")
	}

	parts := strings.Split(tokenString, "_")
	if len(parts) < 2 {
		return nil, errors.New("Token格式错误")
	}

	userID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	userDAO := dao.NewUserDAO(helperInstance)
	user, err := userDAO.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		return nil, errors.New("用户已被禁用")
	}

	return user, nil
}

// respondUnauthorized 返回未授权响应
func respondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    constants.ErrCodeUnauthorized,
		"message": message,
	})
	c.Abort()
}

// GetCurrentUser 获取当前用户
func GetCurrentUser(c *gin.Context) *model.User {
	if user, exists := c.Get("current_user"); exists {
		if u, ok := user.(*model.User); ok {
			return u
		}
	}
	return nil
}

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID(c *gin.Context) uint64 {
	if user := GetCurrentUser(c); user != nil {
		return user.ID
	}
	return 0
}
