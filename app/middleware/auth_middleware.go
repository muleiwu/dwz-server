package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cnb.cool/mliev/open/dwz-server/app/constants"
	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/model"
	envInterface "cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 用户认证中间件
func AuthMiddleware(helper envInterface.GetHelperInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c, helper)
		if err != nil {
			respondUnauthorized(c, err.Error())
			return
		}

		user, err := validateToken(token, helper)
		if err != nil {
			respondUnauthorized(c, "Token验证失败: "+err.Error())
			return
		}

		c.Set("current_user", user)
		c.Next()
	}
}

// extractToken 从请求中提取Token
func extractToken(c *gin.Context, helper envInterface.GetHelperInterface) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("缺少认证信息")
	}

	// 解析Bearer Token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	// 兼容直接传Token的情况
	if authHeader == "" {
		return "", errors.New("Token格式错误")
	}

	return authHeader, nil
}

// validateToken 验证Token并返回用户信息
func validateToken(tokenString string, helper envInterface.GetHelperInterface) (*model.User, error) {
	// 尝试验证API Token
	if user, err := validateAPIToken(tokenString, helper); err == nil {
		return user, nil
	}

	// 尝试验证登录Token
	return validateLoginToken(tokenString, helper)
}

// validateAPIToken 验证API Token
func validateAPIToken(tokenString string, helper envInterface.GetHelperInterface) (*model.User, error) {
	tokenDAO := dao.NewUserTokenDAO(helper)
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
func validateLoginToken(tokenString string, helper envInterface.GetHelperInterface) (*model.User, error) {
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

	userDAO := dao.NewUserDAO(helper)
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
