package service

import (
	"errors"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

// AuthError 认证错误类型
type AuthError struct {
	Code              int
	Message           string
	LimitType         string // "ip" 或 "username"
	RemainingAttempts int
	LockoutSeconds    int
	IsRateLimitError  bool
}

func (e *AuthError) Error() string {
	return e.Message
}

// AuthService 认证服务
type AuthService struct {
	helper      interfaces.HelperInterface
	userService *UserService
	rateLimiter *RateLimiterService
	tokenDAO    *dao.UserTokenDAO
}

// NewAuthService 创建认证服务
func NewAuthService(helper interfaces.HelperInterface) *AuthService {
	return &AuthService{
		helper:      helper,
		userService: NewUserService(helper),
		rateLimiter: NewRateLimiterService(helper),
		tokenDAO:    dao.NewUserTokenDAO(helper),
	}
}

// Login 用户登录（带速率限制）
func (s *AuthService) Login(req *dto.LoginRequest, clientIP string) (*dto.LoginResponse, error) {
	// 1. 检查IP限制
	ipResult, err := s.rateLimiter.CheckIPLimit(clientIP)
	if err != nil {
		// 缓存错误，记录日志但允许继续（优雅降级）
		s.helper.GetLogger().Warn("检查IP限制失败: " + err.Error())
	} else if !ipResult.Allowed {
		return nil, &AuthError{
			Code:              429,
			Message:           "登录尝试次数过多，请稍后再试",
			LimitType:         "ip",
			RemainingAttempts: ipResult.RemainingAttempts,
			LockoutSeconds:    ipResult.LockoutSeconds,
			IsRateLimitError:  true,
		}
	}

	// 2. 检查用户名限制
	usernameResult, err := s.rateLimiter.CheckUsernameLimit(req.Username)
	if err != nil {
		// 缓存错误，记录日志但允许继续（优雅降级）
		s.helper.GetLogger().Warn("检查用户名限制失败: " + err.Error())
	} else if !usernameResult.Allowed {
		return nil, &AuthError{
			Code:              429,
			Message:           "账户已被锁定，请稍后再试",
			LimitType:         "username",
			RemainingAttempts: usernameResult.RemainingAttempts,
			LockoutSeconds:    usernameResult.LockoutSeconds,
			IsRateLimitError:  true,
		}
	}

	// 3. 调用 UserService 验证凭据
	response, err := s.userService.Login(req)
	if err != nil {
		// 登录失败，记录失败尝试
		if recordErr := s.rateLimiter.RecordFailedAttempt(clientIP, req.Username); recordErr != nil {
			s.helper.GetLogger().Warn("记录失败尝试失败: " + recordErr.Error())
		}

		// 获取剩余尝试次数用于错误响应
		remainingAttempts := 0
		if ipResult != nil && usernameResult != nil {
			// 取两者中较小的剩余次数
			if ipResult.RemainingAttempts < usernameResult.RemainingAttempts {
				remainingAttempts = ipResult.RemainingAttempts - 1
			} else {
				remainingAttempts = usernameResult.RemainingAttempts - 1
			}
			if remainingAttempts < 0 {
				remainingAttempts = 0
			}
		}

		// 检查是否因为用户被禁用
		if err.Error() == "用户已被禁用" {
			return nil, &AuthError{
				Code:             403,
				Message:          err.Error(),
				IsRateLimitError: false,
			}
		}

		return nil, &AuthError{
			Code:              401,
			Message:           err.Error(),
			RemainingAttempts: remainingAttempts,
			IsRateLimitError:  false,
		}
	}

	// 4. 登录成功，重置尝试计数
	if resetErr := s.rateLimiter.ResetAttempts(clientIP, req.Username); resetErr != nil {
		s.helper.GetLogger().Warn("重置尝试计数失败: " + resetErr.Error())
	}

	return response, nil
}

// Logout 用户登出
func (s *AuthService) Logout(userID uint64) error {
	if userID == 0 {
		return errors.New("无效的用户ID")
	}

	// 删除用户的所有Token，使会话失效
	if err := s.tokenDAO.DeleteByUserID(userID); err != nil {
		s.helper.GetLogger().Error("删除用户Token失败: " + err.Error())
		return errors.New("登出失败")
	}

	return nil
}
