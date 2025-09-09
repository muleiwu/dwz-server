package distributed_id_generator

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/base62"
	"github.com/redis/go-redis/v9"
)

// DistributedIDGenerator 分布式ID生成器
type DistributedIDGenerator struct {
	redis         *redis.Client
	base62        *base62.Base62
	logger        interfaces.LoggerInterface
	fallbackChars string
}

func NewDistributedIDGenerator(redisClient *redis.Client) *DistributedIDGenerator {
	return &DistributedIDGenerator{
		redis:         redisClient,
		base62:        base62.NewBase62(),
		fallbackChars: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}
}

// GenerateID 为指定域名生成下一个ID
func (g *DistributedIDGenerator) GenerateID(domainID uint64, ctx context.Context) (uint64, error) {
	key := fmt.Sprintf("domain_counter:%d", domainID)

	// 使用Redis INCR命令获取下一个ID
	result, err := g.redis.Incr(ctx, key).Result()
	if err != nil {
		g.logger.Error(fmt.Sprintf("Redis INCR失败: %v", err))
		return 0, err
	}

	return uint64(result), nil
}

// InitializeDomainCounter 初始化域名计数器
func (g *DistributedIDGenerator) InitializeDomainCounter(domainID uint64, startValue uint64) error {
	ctx := context.Background()
	key := fmt.Sprintf("domain_counter:%d", domainID)

	// 检查键是否已存在
	exists, err := g.redis.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	// 如果不存在或者当前值小于startValue，则设置新值
	if exists == 0 {
		err = g.redis.Set(ctx, key, startValue, 0).Err()
		if err != nil {
			return err
		}
		g.logger.Info(fmt.Sprintf("初始化域名%d计数器为%d", domainID, startValue))
	} else {
		// 获取当前值，如果小于startValue则更新
		current, err := g.redis.Get(ctx, key).Int64()
		if err != nil {
			return err
		}

		if uint64(current) < startValue {
			err = g.redis.Set(ctx, key, startValue, 0).Err()
			if err != nil {
				return err
			}
			g.logger.Info(fmt.Sprintf("更新域名%d计数器从%d到%d", domainID, current, startValue))
		}
	}

	return nil
}

// GenerateShortCode 生成短代码（包含防猜测措施）
func (g *DistributedIDGenerator) GenerateShortCode(domainID uint64, ctx context.Context) (string, *uint64, error) {
	// 主方案：使用分布式发号器
	id, err := g.GenerateID(domainID, ctx)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("分布式发号器故障，使用降级方案: %v", err))
	}

	// 将ID转换为62进制
	base62Code := g.base62.Encode(id)

	// 添加防猜测措施：两位随机后缀 + 校验码
	shortCode, err := g.addAntiGuessingSuffix(base62Code)

	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("添加防猜测后缀失败: %v", err))
	}

	return shortCode, &id, nil
}

// addAntiGuessingSuffix 添加防猜测后缀
func (g *DistributedIDGenerator) addAntiGuessingSuffix(base62Code string) (string, error) {
	// 生成两位随机后缀
	randomSuffix, err := g.generateRandomSuffix(2)
	if err != nil {
		return "", err
	}

	// 计算校验码（异或）
	checksum := g.calculateChecksum(base62Code + randomSuffix)

	// 返回格式：base62Code + 随机后缀 + 校验码
	return base62Code + randomSuffix + string(g.fallbackChars[checksum]), nil
}

// generateRandomSuffix 生成指定长度的随机后缀
func (g *DistributedIDGenerator) generateRandomSuffix(length int) (string, error) {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(g.fallbackChars)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = g.fallbackChars[randomIndex.Int64()]
	}

	return string(result), nil
}

// calculateChecksum 计算校验码
func (g *DistributedIDGenerator) calculateChecksum(input string) int {
	checksum := 0
	for _, char := range input {
		checksum ^= int(char)
	}
	return checksum % len(g.fallbackChars)
}

// ValidateShortCode 验证短代码
func (g *DistributedIDGenerator) ValidateShortCode(shortCode string) bool {
	if len(shortCode) < 4 { // 至少包含基础码 + 2位随机 + 1位校验
		return false
	}

	// 分离各部分
	codeLen := len(shortCode)
	baseCode := shortCode[:codeLen-3]
	randomSuffix := shortCode[codeLen-3 : codeLen-1]
	checksum := shortCode[codeLen-1:]

	// 验证校验码
	expectedChecksum := g.calculateChecksum(baseCode + randomSuffix)
	expectedChecksumChar := string(g.fallbackChars[expectedChecksum])

	return checksum == expectedChecksumChar
}

// ExtractBaseCode 从短代码中提取基础代码（去除防猜测后缀）
func (g *DistributedIDGenerator) ExtractBaseCode(shortCode string) (string, error) {
	if !g.ValidateShortCode(shortCode) {
		return "", errors.New("invalid short code")
	}

	codeLen := len(shortCode)
	baseCode := shortCode[:codeLen-3]

	return baseCode, nil
}

// DecodeToID 将基础代码解码为ID
func (g *DistributedIDGenerator) DecodeToID(baseCode string) (uint64, error) {
	return g.base62.Decode(baseCode)
}

// GetCurrentCounter 获取域名当前计数器值
func (g *DistributedIDGenerator) GetCurrentCounter(domainID uint64) (uint64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("domain_counter:%d", domainID)

	result, err := g.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil // 键不存在，返回0
		}
		return 0, err
	}

	counter, err := strconv.ParseUint(result, 10, 64)
	if err != nil {
		return 0, err
	}

	return counter, nil
}

// ResetDomainCounter 重置域名计数器（谨慎使用）
func (g *DistributedIDGenerator) ResetDomainCounter(domainID uint64, newValue uint64) error {
	ctx := context.Background()
	key := fmt.Sprintf("domain_counter:%d", domainID)

	err := g.redis.Set(ctx, key, newValue, 0).Err()
	if err != nil {
		return err
	}

	g.logger.Warn(fmt.Sprintf("重置域名%d计数器为%d", domainID, newValue))
	return nil
}
