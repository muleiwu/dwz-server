package impl

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"math/bits"

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
	// 使用分布式发号器
	id, err := g.GenerateID(domainID, ctx)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("分布式发号器故障: %v", err))
	}

	// 如果启用XOR混淆，对ID进行混淆
	encodedID := id
	if config.EnableXorObfuscation {
		encodedID = obfuscateIDRedis(id, config.XorSecret, config.XorRot)
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

// obfuscateIDRedis 使用XOR和位旋转混淆ID，保持结果的base62长度与原ID一致
func obfuscateIDRedis(id uint64, secret uint64, rot int) uint64 {
	// 计算ID需要的位数
	bitLen := bits.Len64(id)
	if bitLen == 0 {
		bitLen = 1 // 至少1位
	}

	// 创建位掩码，确保结果在相同范围内
	mask := (uint64(1) << bitLen) - 1

	// 先限制ID在范围内（理论上已经在范围内，但保险起见）
	id = id & mask

	// 对secret也应用掩码，避免XOR后超出范围
	maskedSecret := secret & mask

	// 位旋转需要在有效位数内旋转
	rotInRange := rot % bitLen
	if rotInRange < 0 {
		rotInRange += bitLen
	}

	// 在bitLen位内进行旋转
	rotated := ((id << rotInRange) | (id >> (bitLen - rotInRange))) & mask

	// XOR并应用掩码
	result := (rotated ^ maskedSecret) & mask

	return result
}
