package domain_validate

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// TruncateString 将字符串截断到指定的最大长度
// 如果字符串长度小于等于 maxLength，则返回原字符串
// 否则返回截断后的字符串
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}

func ValidateDomain(domain string) error {
	// 检查是否包含协议头
	if strings.Contains(domain, "://") {
		return errors.New("域名不应包含协议头(http://或https://)")
	}

	// IP的格式也可以，支持IP:端口
	if strings.Contains(domain, ":") {
		// 可能是IP:端口格式
		parts := strings.Split(domain, ":")
		if len(parts) == 2 {
			// 检查IP部分
			if netIP := net.ParseIP(parts[0]); netIP != nil {
				// 检查端口部分
				port, err := strconv.Atoi(parts[1])
				if err == nil && port >= 0 && port <= 65535 {
					return nil
				}
			}
		}
	} else if netIP := net.ParseIP(domain); netIP != nil {
		// 如果是有效的纯IP地址，直接返回
		return nil
	}

	// 使用正则表达式验证域名格式
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	if !domainRegex.MatchString(domain) {
		return errors.New("无效的域名格式")
	}

	return nil
}
