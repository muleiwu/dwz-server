package util

import (
	"errors"
	"math"
	"strings"
)

// Base62Converter 62进制转换器
type Base62Converter struct {
	charset string
	base    int
}

// NewBase62Converter 创建62进制转换器
func NewBase62Converter() *Base62Converter {
	return &Base62Converter{
		charset: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		base:    62,
	}
}

// Encode 将数字编码为62进制字符串
func (c *Base62Converter) Encode(num uint64) string {
	if num == 0 {
		return "0"
	}

	result := ""
	for num > 0 {
		remainder := num % uint64(c.base)
		result = string(c.charset[remainder]) + result
		num = num / uint64(c.base)
	}

	return result
}

// Decode 将62进制字符串解码为数字
func (c *Base62Converter) Decode(str string) (uint64, error) {
	if str == "" {
		return 0, errors.New("empty string")
	}

	var result uint64 = 0
	strLen := len(str)

	for i, char := range str {
		// 查找字符在字符集中的位置
		pos := strings.IndexRune(c.charset, char)
		if pos == -1 {
			return 0, errors.New("invalid character in string")
		}

		// 计算该位的值
		power := strLen - i - 1
		if power > 10 { // 防止溢出，uint64最大约18位十进制
			return 0, errors.New("string too long, would cause overflow")
		}

		value := uint64(pos) * uint64(math.Pow(float64(c.base), float64(power)))

		// 检查是否会溢出
		if result > math.MaxUint64-value {
			return 0, errors.New("overflow detected")
		}

		result += value
	}

	return result, nil
}

// ValidateCode 验证短代码是否有效
func (c *Base62Converter) ValidateCode(code string) bool {
	if len(code) == 0 || len(code) > 11 { // 62^11 > uint64最大值，所以限制11位
		return false
	}

	// 检查是否只包含允许的字符
	for _, char := range code {
		if !strings.ContainsRune(c.charset, char) {
			return false
		}
	}

	// 尝试解码，看是否会溢出
	_, err := c.Decode(code)
	return err == nil
}

// GetMaxSafeLength 获取安全的最大长度（不会溢出uint64）
func (c *Base62Converter) GetMaxSafeLength() int {
	// 62^10 = 839299365868340224 < uint64最大值
	// 62^11 = 52036560803398893888 > uint64最大值
	return 10
}

// EstimateLength 估算给定数字编码后的长度
func (c *Base62Converter) EstimateLength(num uint64) int {
	if num == 0 {
		return 1
	}

	length := 0
	for num > 0 {
		num = num / uint64(c.base)
		length++
	}
	return length
}

// GetRange 获取指定长度的数字范围
func (c *Base62Converter) GetRange(length int) (min, max uint64) {
	if length <= 0 {
		return 0, 0
	}
	if length == 1 {
		return 0, uint64(c.base) - 1
	}

	min = uint64(math.Pow(float64(c.base), float64(length-1)))
	max = uint64(math.Pow(float64(c.base), float64(length))) - 1

	// 确保不超过uint64最大值
	if max > math.MaxUint64 {
		max = math.MaxUint64
	}

	return min, max
}

// ShortCodeGenerator 短网址代码生成器（保持兼容性）
type ShortCodeGenerator struct {
	converter *Base62Converter
}

// NewShortCodeGenerator 创建短网址代码生成器
func NewShortCodeGenerator(length int) *ShortCodeGenerator {
	return &ShortCodeGenerator{
		converter: NewBase62Converter(),
	}
}

// EncodeID 将数据库ID编码为短代码
func (g *ShortCodeGenerator) EncodeID(id uint64) string {
	return g.converter.Encode(id)
}

// DecodeID 将短代码解码为数据库ID
func (g *ShortCodeGenerator) DecodeID(code string) (uint64, error) {
	return g.converter.Decode(code)
}

// ValidateCustomCode 验证自定义短代码是否有效
func (g *ShortCodeGenerator) ValidateCustomCode(code string) bool {
	return g.converter.ValidateCode(code)
}

// 以下为保持向后兼容的函数（已废弃，建议使用EncodeID）

// Generate 生成随机短代码（已废弃，建议使用EncodeID）
func (g *ShortCodeGenerator) Generate() (string, error) {
	return "", errors.New("deprecated: use EncodeID instead")
}

// GenerateCustom 生成自定义长度的短代码（已废弃）
func (g *ShortCodeGenerator) GenerateCustom(length int) (string, error) {
	return "", errors.New("deprecated: use EncodeID instead")
}

// GenerateWithTimestamp 生成带时间戳的短代码（已废弃）
func (g *ShortCodeGenerator) GenerateWithTimestamp() (string, error) {
	return "", errors.New("deprecated: use EncodeID instead")
}

// GenerateSequential 生成顺序短代码（已废弃）
func (g *ShortCodeGenerator) GenerateSequential(count int) ([]string, error) {
	return nil, errors.New("deprecated: use EncodeID instead")
}
