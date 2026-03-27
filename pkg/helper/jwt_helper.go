package helper

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// LoginClaims JWT登录Token的Claims
type LoginClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTHelper JWT助手
type JWTHelper struct {
	secret      []byte
	expireHours int
}

var (
	jwtHelper     *JWTHelper
	jwtHelperOnce sync.Once
)

// InitJWTHelper 初始化JWT助手单例（幂等，仅首次调用生效）
func InitJWTHelper(secret string, expireHours int) {
	jwtHelperOnce.Do(func() {
		jwtHelper = &JWTHelper{
			secret:      []byte(secret),
			expireHours: expireHours,
		}
	})
}

// GetJWTHelper 获取JWT助手单例
func GetJWTHelper() *JWTHelper {
	return jwtHelper
}

// GenerateToken 生成JWT Token
func (h *JWTHelper) GenerateToken(userID uint64, username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(h.expireHours) * time.Hour)

	claims := LoginClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "dwz-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken 验证JWT Token并返回Claims
func (h *JWTHelper) ValidateToken(tokenString string) (*LoginClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LoginClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return h.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*LoginClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
