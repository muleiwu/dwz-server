package base

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
	"cnb.cool/mliev/open/dwz-server/pkg/base62"
)

type IdGeneratorBase struct {
	fallbackChars string
	base62        *base62.Base62
	counters      map[uint64]uint64 // Map of domainID -> current counter value
	countersMutex sync.RWMutex      // Mutex for accessing the counters map
}

// NewIdGeneratorBase creates a new instance of IdGeneratorBase
func NewIdGeneratorBase() interfaces.IDGenerator {
	return &IdGeneratorBase{
		fallbackChars: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		base62:        base62.NewBase62(),
		counters:      make(map[uint64]uint64),
		countersMutex: sync.RWMutex{},
	}
}

func (g *IdGeneratorBase) GenerateID(domainID uint64, ctx context.Context) (uint64, error) {
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

func (g *IdGeneratorBase) InitializeDomainCounter(domainID uint64, startValue uint64) error {
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

func (g *IdGeneratorBase) ResetDomainCounter(domainID uint64, newValue uint64) error {
	// Use a lock to ensure thread safety
	g.countersMutex.Lock()
	defer g.countersMutex.Unlock()

	// Simply reset the counter to the new value
	g.counters[domainID] = newValue

	return nil
}

func (g *IdGeneratorBase) GenerateShortCode(domainID uint64, ctx context.Context) (string, *uint64, error) {
	// Generate ID using our concurrent ID generator
	id, err := g.GenerateID(domainID, ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate ID: %v", err)
	}

	// Convert to base62 using the base62 module
	base62Code := g.base62.Encode(id)

	// Add anti-guessing suffix
	shortCode, err := g.addAntiGuessingSuffix(base62Code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to add anti-guessing suffix: %v", err)
	}

	return shortCode, &id, nil
}

// addAntiGuessingSuffix 添加防猜测后缀
func (g *IdGeneratorBase) addAntiGuessingSuffix(base62Code string) (string, error) {
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
func (g *IdGeneratorBase) generateRandomSuffix(length int) (string, error) {
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
func (g *IdGeneratorBase) calculateChecksum(input string) int {
	checksum := 0
	for _, char := range input {
		checksum ^= int(char)
	}
	return checksum % len(g.fallbackChars)
}
