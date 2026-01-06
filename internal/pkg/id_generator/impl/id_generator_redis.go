package impl

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/muleiwu/base_n"
	"github.com/redis/go-redis/v9"
)

type IDGeneratorRedis struct {
	redis         *redis.Client
	base62        *base_n.BaseN
	logger        interfaces.LoggerInterface
	fallbackChars string
}

func NewIDGeneratorRedis(helper interfaces.HelperInterface) interfaces.IDGenerator {
	return &IDGeneratorRedis{
		logger:        helper.GetLogger(),
		redis:         helper.GetRedis(),
		base62:        base_n.NewBase62(),
		fallbackChars: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}
}

// GenerateID 为指定域名生成下一个ID
func (g *IDGeneratorRedis) GenerateID(domainID uint64, ctx context.Context) (uint64, error) {
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
func (g *IDGeneratorRedis) InitializeDomainCounter(domainID uint64, startValue uint64) error {
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

		if uint64(current) <= startValue {
			err = g.redis.Set(ctx, key, startValue, 0).Err()
			if err != nil {
				return err
			}
			g.logger.Info(fmt.Sprintf("更新域名%d计数器从%d到%d", domainID, current, startValue))
		}
	}

	return nil
}

// ResetDomainCounter 重置域名计数器（谨慎使用）
func (g *IDGeneratorRedis) ResetDomainCounter(domainID uint64, newValue uint64) error {
	ctx := context.Background()
	key := fmt.Sprintf("domain_counter:%d", domainID)

	err := g.redis.Set(ctx, key, newValue, 0).Err()
	if err != nil {
		return err
	}

	g.logger.Warn(fmt.Sprintf("重置域名%d计数器为%d", domainID, newValue))
	return nil
}

// GenerateShortCode 生成短代码（包含防猜测措施）
func (g *IDGeneratorRedis) GenerateShortCode(domainID uint64, ctx context.Context) (string, *uint64, error) {
	// 主方案：使用分布式发号器
	id, err := g.GenerateID(domainID, ctx)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("分布式发号器故障，使用降级方案: %v", err))
	}

	// 将ID转换为62进制
	base62Code := g.base62.Encode(int64(id))

	// 添加防猜测措施：两位随机后缀 + 校验码
	shortCode, err := g.addAntiGuessingSuffix(base62Code)

	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("添加防猜测后缀失败: %v", err))
	}

	return shortCode, &id, nil
}

// GenerateShortCodeWithConfig 使用自定义配置生成短代码
func (g *IDGeneratorRedis) GenerateShortCodeWithConfig(domainID uint64, ctx context.Context, config interfaces.ShortCodeConfig) (string, *uint64, error) {
	// 检查是否需要初始化计数器（当计数器为0且DefaultStartNumber > 0时）
	if config.DefaultStartNumber > 0 {
		key := fmt.Sprintf("domain_counter:%d", domainID)

		// 检查当前计数器值
		current, err := g.redis.Get(ctx, key).Int64()
		if err == redis.Nil || current == 0 {
			// 计数器不存在或为0，使用默认开始数字初始化
			g.InitializeDomainCounter(domainID, config.DefaultStartNumber)
		} else if err != nil {
			g.logger.Error(fmt.Sprintf("检查域名%d计数器失败: %v", domainID, err))
		}
	}

	// 使用分布式发号器
	id, err := g.GenerateID(domainID, ctx)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("分布式发号器故障: %v", err))
	}

	// 如果启用XOR混淆，对ID进行混淆
	encodedID := id
	if config.EnableXorObfuscation {
		encodedID = g.obfuscateID(id, config.XorSecret, config.XorRot)
	}

	// 将ID转换为62进制
	base62Code := g.base62.Encode(int64(encodedID))

	// 添加防猜测措施（使用配置）
	shortCode, err := g.addAntiGuessingSuffixWithConfig(base62Code, config)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("添加防猜测后缀失败: %v", err))
	}

	return shortCode, &id, nil
}

// addAntiGuessingSuffix 添加防猜测后缀
func (g *IDGeneratorRedis) addAntiGuessingSuffix(base62Code string) (string, error) {
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

// addAntiGuessingSuffixWithConfig 使用配置添加防猜测后缀
func (g *IDGeneratorRedis) addAntiGuessingSuffixWithConfig(base62Code string, config interfaces.ShortCodeConfig) (string, error) {
	result := base62Code

	// 添加随机后缀
	if config.RandomSuffixLength > 0 {
		randomSuffix, err := g.generateRandomSuffix(config.RandomSuffixLength)
		if err != nil {
			return "", err
		}
		result += randomSuffix
	}

	// 添加校验位
	if config.EnableChecksum {
		checksum := g.calculateChecksum(result)
		result += string(g.fallbackChars[checksum])
	}

	return result, nil
}

// generateRandomSuffix 生成指定长度的随机后缀
func (g *IDGeneratorRedis) generateRandomSuffix(length int) (string, error) {
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
func (g *IDGeneratorRedis) calculateChecksum(input string) int {
	checksum := 0
	for _, char := range input {
		checksum ^= int(char)
	}
	return checksum % len(g.fallbackChars)
}

// pow62 计算 62 的 n 次方
func (g *IDGeneratorRedis) pow62(n int) uint64 {
	if n <= 0 {
		return 1
	}
	result := uint64(1)
	for i := 0; i < n; i++ {
		result *= 62
	}
	return result
}

// calculateBase62Digits 计算 ID 对应的 Base62 位数
func (g *IDGeneratorRedis) calculateBase62Digits(id uint64) int {
	if id == 0 {
		return 1
	}
	digits := 1
	threshold := uint64(62)
	for id >= threshold {
		digits++
		threshold *= 62
	}
	return digits
}

// obfuscateID 使用XOR和位旋转混淆ID，保持结果的base62长度与原ID一致
func (g *IDGeneratorRedis) obfuscateID(id uint64, secret uint64, rot int) uint64 {
	// 计算 ID 对应的 Base62 位数
	digits := g.calculateBase62Digits(id)

	// 获取该位数的范围边界
	minVal := g.pow62(digits - 1)
	maxVal := g.pow62(digits) - 1
	rangeSize := maxVal - minVal + 1

	// 归一化到 [0, rangeSize-1]
	normalized := id - minVal

	// 在范围内进行位旋转（如果 rot > 0）
	if rot > 0 && rangeSize > 1 {
		rotAmount := uint64(rot) % rangeSize
		normalized = (normalized + rotAmount) % rangeSize
	}

	// XOR 混淆（保证双射）
	obfuscated := normalized ^ (secret % rangeSize)

	// 映射回原范围
	return obfuscated + minVal
}
