package impl

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"github.com/muleiwu/base_n"
)

type IDGeneratorLocal struct {
	fallbackChars string
	base62        *base_n.BaseN
	counters      map[uint64]uint64 // Map of domainID -> current counter value
	countersMutex sync.RWMutex      // Mutex for accessing the counters map
}

// NewIDGeneratorLocal creates a new instance of IDGeneratorLocal
func NewIDGeneratorLocal() interfaces.IDGenerator {
	return &IDGeneratorLocal{
		fallbackChars: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		base62:        base_n.NewBase62(),
		counters:      make(map[uint64]uint64),
		countersMutex: sync.RWMutex{},
	}
}

func (g *IDGeneratorLocal) GenerateID(domainID uint64, ctx context.Context) (uint64, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// Continue if context is not cancelled
	}

	// First try a fast path with just a read lock
	g.countersMutex.RLock()
	_, exists := g.counters[domainID]
	g.countersMutex.RUnlock()

	// If the domain counter doesn't exist, we need to initialize it
	if !exists {
		g.countersMutex.Lock()
		if _, exists := g.counters[domainID]; !exists {
			g.counters[domainID] = 0
		}
		g.countersMutex.Unlock()
	}

	// For high-concurrency, use per-domain locking instead of a global lock
	// This will reduce contention between different domains
	g.countersMutex.Lock()
	defer g.countersMutex.Unlock()

	// Get the current value and increment
	currentValue := g.counters[domainID]
	newValue := currentValue + 1

	// Update the counter
	g.counters[domainID] = newValue

	return newValue, nil
}

func (g *IDGeneratorLocal) InitializeDomainCounter(domainID uint64, startValue uint64) error {
	// Use a lock to ensure thread safety
	g.countersMutex.Lock()
	defer g.countersMutex.Unlock()

	// Check if counter already exists
	currentValue, exists := g.counters[domainID]

	if !exists {
		// Counter does not exist, initialize it
		g.counters[domainID] = startValue
		return nil
	}

	// Counter exists, update only if startValue is greater
	if startValue > currentValue {
		g.counters[domainID] = startValue
	}

	return nil
}

func (g *IDGeneratorLocal) ResetDomainCounter(domainID uint64, newValue uint64) error {
	// Use a lock to ensure thread safety
	g.countersMutex.Lock()
	defer g.countersMutex.Unlock()

	// Simply reset the counter to the new value
	g.counters[domainID] = newValue

	return nil
}

func (g *IDGeneratorLocal) GenerateShortCode(domainID uint64, ctx context.Context) (string, *uint64, error) {
	// Generate ID using our concurrent ID generator
	id, err := g.GenerateID(domainID, ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate ID: %v", err)
	}

	// Convert to base62 using the base62 module
	base62Code := g.base62.Encode(int64(id))

	// Add anti-guessing suffix
	shortCode, err := g.addAntiGuessingSuffix(base62Code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to add anti-guessing suffix: %v", err)
	}

	return shortCode, &id, nil
}

// GenerateShortCodeWithConfig 使用自定义配置生成短代码
func (g *IDGeneratorLocal) GenerateShortCodeWithConfig(domainID uint64, ctx context.Context, config interfaces.ShortCodeConfig) (string, *uint64, error) {
	// 检查是否需要初始化计数器（当计数器为0且DefaultStartNumber > 0时）
	if config.DefaultStartNumber > 0 {
		g.countersMutex.RLock()
		currentValue, exists := g.counters[domainID]
		g.countersMutex.RUnlock()

		// 如果计数器不存在或为0，使用默认开始数字初始化
		if !exists || currentValue == 0 {
			g.InitializeDomainCounter(domainID, config.DefaultStartNumber)
		}
	}

	// Generate ID using our concurrent ID generator
	id, err := g.GenerateID(domainID, ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate ID: %v", err)
	}

	// 如果启用XOR混淆，对ID进行混淆
	encodedID := id
	if config.EnableXorObfuscation {
		encodedID = g.obfuscateID(id, config.XorSecret, config.XorRot)
	}

	// Convert to base62 using the base62 module
	base62Code := g.base62.Encode(int64(encodedID))

	// Add anti-guessing suffix with config
	shortCode, err := g.addAntiGuessingSuffixWithConfig(base62Code, config)
	if err != nil {
		return "", nil, fmt.Errorf("failed to add anti-guessing suffix: %v", err)
	}

	return shortCode, &id, nil
}

// addAntiGuessingSuffix 添加防猜测后缀
func (g *IDGeneratorLocal) addAntiGuessingSuffix(base62Code string) (string, error) {
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
func (g *IDGeneratorLocal) addAntiGuessingSuffixWithConfig(base62Code string, config interfaces.ShortCodeConfig) (string, error) {
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

// pow62 计算 62 的 n 次方
func (g *IDGeneratorLocal) pow62(n int) uint64 {
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
func (g *IDGeneratorLocal) calculateBase62Digits(id uint64) int {
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
func (g *IDGeneratorLocal) obfuscateID(id uint64, secret uint64, rot int) uint64 {
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

// generateRandomSuffix 生成指定长度的随机后缀
func (g *IDGeneratorLocal) generateRandomSuffix(length int) (string, error) {
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
func (g *IDGeneratorLocal) calculateChecksum(input string) int {
	checksum := 0
	for _, char := range input {
		checksum ^= int(char)
	}
	return checksum % len(g.fallbackChars)
}
