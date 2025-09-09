package base62

import (
	"errors"
	"math"
	"strings"
)

type Base62 struct {
	charset string
	base    int
}

// NewBase62 创建62进制转换器
func NewBase62() *Base62 {
	return &Base62{
		charset: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		base:    62,
	}
}

// Encode 将数字编码为62进制字符串
func (c *Base62) Encode(num uint64) string {
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
func (c *Base62) Decode(str string) (uint64, error) {
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
func (c *Base62) ValidateCode(code string) bool {
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
func (c *Base62) GetMaxSafeLength() int {
	// 62^10 = 839299365868340224 < uint64最大值
	// 62^11 = 52036560803398893888 > uint64最大值
	return 10
}

// EstimateLength 估算给定数字编码后的长度
func (c *Base62) EstimateLength(num uint64) int {
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
func (c *Base62) GetRange(length int) (min, max uint64) {
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
