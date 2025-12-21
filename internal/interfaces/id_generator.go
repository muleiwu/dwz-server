package interfaces

import "context"

// ShortCodeConfig 短码生成配置
type ShortCodeConfig struct {
	RandomSuffixLength int  // 随机后缀位数 (0-10)
	EnableChecksum     bool // 是否启用校验位
}

type IDGenerator interface {

	// InitializeDomainCounter 初始化域名计数器
	InitializeDomainCounter(domainID uint64, startValue uint64) error

	// GenerateID 为指定域名生成下一个ID
	GenerateID(domainID uint64, ctx context.Context) (uint64, error)

	// GenerateShortCode 生成短代码（包含防猜测措施）
	GenerateShortCode(domainID uint64, ctx context.Context) (string, *uint64, error)

	// GenerateShortCodeWithConfig 使用自定义配置生成短代码
	GenerateShortCodeWithConfig(domainID uint64, ctx context.Context, config ShortCodeConfig) (string, *uint64, error)

	// ResetDomainCounter 重置域名计数器（谨慎使用）
	ResetDomainCounter(domainID uint64, newValue uint64) error
}
