package service

import (
	"context"
	"time"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

// Rate limit constants
const (
	// IP限制配置
	IPMaxAttempts   = 10               // IP最大尝试次数
	IPWindowPeriod  = 15 * time.Minute // IP时间窗口
	IPLockoutPeriod = 15 * time.Minute // IP锁定时间

	// 用户名限制配置
	UserMaxAttempts   = 5                // 用户名最大尝试次数
	UserWindowPeriod  = 15 * time.Minute // 用户名时间窗口
	UserLockoutPeriod = 15 * time.Minute // 用户名锁定时间

	// 缓存键前缀
	CacheKeyIPAttempt   = "login_attempt:ip:"
	CacheKeyUserAttempt = "login_attempt:user:"
)

// LoginAttemptRecord 登录尝试记录
type LoginAttemptRecord struct {
	Count    int   `json:"count"`     // 尝试次数
	FirstAt  int64 `json:"first_at"`  // 首次尝试时间戳
	LastAt   int64 `json:"last_at"`   // 最后尝试时间戳
	LockedAt int64 `json:"locked_at"` // 锁定时间戳（0表示未锁定）
}

// RateLimitResult 速率限制检查结果
type RateLimitResult struct {
	Allowed           bool  // 是否允许
	RemainingAttempts int   // 剩余尝试次数
	LockedUntil       int64 // 锁定截止时间戳（秒）
	LockoutSeconds    int   // 剩余锁定秒数
}

// RateLimiterService 速率限制服务
type RateLimiterService struct {
	helper interfaces.HelperInterface
}

// NewRateLimiterService 创建速率限制服务
func NewRateLimiterService(helper interfaces.HelperInterface) *RateLimiterService {
	return &RateLimiterService{
		helper: helper,
	}
}

// CheckIPLimit 检查IP限制
func (s *RateLimiterService) CheckIPLimit(ip string) (*RateLimitResult, error) {
	return s.checkLimit(CacheKeyIPAttempt+ip, IPMaxAttempts, IPLockoutPeriod)
}

// checkLimit 通用限制检查方法
func (s *RateLimiterService) checkLimit(cacheKey string, maxAttempts int, lockoutPeriod time.Duration) (*RateLimitResult, error) {
	cache := s.helper.GetCache()
	if cache == nil {
		// 缓存不可用，允许请求继续（优雅降级）
		s.helper.GetLogger().Warn("缓存服务不可用，跳过速率限制检查")
		return &RateLimitResult{
			Allowed:           true,
			RemainingAttempts: maxAttempts,
		}, nil
	}

	ctx := context.Background()
	var record LoginAttemptRecord

	// 尝试从缓存获取记录
	err := cache.Get(ctx, cacheKey, &record)
	if err != nil {
		// 缓存中没有记录，允许请求
		return &RateLimitResult{
			Allowed:           true,
			RemainingAttempts: maxAttempts,
		}, nil
	}

	now := time.Now().Unix()

	// 检查是否已锁定
	if record.LockedAt > 0 {
		lockoutEnd := record.LockedAt + int64(lockoutPeriod.Seconds())
		if now < lockoutEnd {
			// 仍在锁定期内
			return &RateLimitResult{
				Allowed:           false,
				RemainingAttempts: 0,
				LockedUntil:       lockoutEnd,
				LockoutSeconds:    int(lockoutEnd - now),
			}, nil
		}
		// 锁定已过期，允许请求
		return &RateLimitResult{
			Allowed:           true,
			RemainingAttempts: maxAttempts,
		}, nil
	}

	// 检查尝试次数是否超过阈值
	if record.Count >= maxAttempts {
		// 超过阈值，应该被锁定
		return &RateLimitResult{
			Allowed:           false,
			RemainingAttempts: 0,
		}, nil
	}

	// 允许请求，返回剩余尝试次数
	return &RateLimitResult{
		Allowed:           true,
		RemainingAttempts: maxAttempts - record.Count,
	}, nil
}

// CheckUsernameLimit 检查用户名限制
func (s *RateLimiterService) CheckUsernameLimit(username string) (*RateLimitResult, error) {
	return s.checkLimit(CacheKeyUserAttempt+username, UserMaxAttempts, UserLockoutPeriod)
}

// RecordFailedAttempt 记录失败尝试
func (s *RateLimiterService) RecordFailedAttempt(ip string, username string) error {
	cache := s.helper.GetCache()
	if cache == nil {
		// 缓存不可用，记录警告日志
		s.helper.GetLogger().Warn("缓存服务不可用，无法记录登录尝试")
		return nil
	}

	// 记录IP尝试
	if ip != "" {
		if err := s.recordAttempt(CacheKeyIPAttempt+ip, IPMaxAttempts, IPWindowPeriod); err != nil {
			s.helper.GetLogger().Error("记录IP登录尝试失败: " + err.Error())
		}
	}

	// 记录用户名尝试
	if username != "" {
		if err := s.recordAttempt(CacheKeyUserAttempt+username, UserMaxAttempts, UserWindowPeriod); err != nil {
			s.helper.GetLogger().Error("记录用户名登录尝试失败: " + err.Error())
		}
	}

	return nil
}

// recordAttempt 记录单个尝试
func (s *RateLimiterService) recordAttempt(cacheKey string, maxAttempts int, windowPeriod time.Duration) error {
	cache := s.helper.GetCache()
	ctx := context.Background()
	now := time.Now().Unix()

	var record LoginAttemptRecord

	// 尝试获取现有记录
	err := cache.Get(ctx, cacheKey, &record)
	if err != nil {
		// 没有现有记录，创建新记录
		record = LoginAttemptRecord{
			Count:   1,
			FirstAt: now,
			LastAt:  now,
		}
	} else {
		// 更新现有记录
		record.Count++
		record.LastAt = now

		// 检查是否需要触发锁定
		if record.Count >= maxAttempts && record.LockedAt == 0 {
			record.LockedAt = now
		}
	}

	// 设置TTL为窗口期
	return cache.Set(ctx, cacheKey, record, windowPeriod)
}

// ResetAttempts 重置尝试次数
func (s *RateLimiterService) ResetAttempts(ip string, username string) error {
	cache := s.helper.GetCache()
	if cache == nil {
		// 缓存不可用，记录警告日志
		s.helper.GetLogger().Warn("缓存服务不可用，无法重置登录尝试")
		return nil
	}

	ctx := context.Background()

	// 删除IP尝试记录
	if ip != "" {
		if err := cache.Del(ctx, CacheKeyIPAttempt+ip); err != nil {
			s.helper.GetLogger().Error("删除IP登录尝试记录失败: " + err.Error())
		}
	}

	// 删除用户名尝试记录
	if username != "" {
		if err := cache.Del(ctx, CacheKeyUserAttempt+username); err != nil {
			s.helper.GetLogger().Error("删除用户名登录尝试记录失败: " + err.Error())
		}
	}

	return nil
}
