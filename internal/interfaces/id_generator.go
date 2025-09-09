package interfaces

import "context"

type IDGenerator interface {

	// InitializeDomainCounter 初始化域名计数器
	InitializeDomainCounter(domainID uint64, startValue uint64) error

	// GenerateID 为指定域名生成下一个ID
	GenerateID(domainID uint64, ctx context.Context) (uint64, error)

	// GenerateShortCode 生成短代码（包含防猜测措施）
	GenerateShortCode(domainID uint64, ctx context.Context) (string, *uint64, error)

	// ResetDomainCounter 重置域名计数器（谨慎使用）
	ResetDomainCounter(domainID uint64, newValue uint64) error
}
