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

	// 域名正则
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

	// 包含冒号时，解析为 host:port 格式
	if strings.Contains(domain, ":") {
		parts := strings.Split(domain, ":")
		if len(parts) != 2 {
			return errors.New("无效的域名格式")
		}
		host, portStr := parts[0], parts[1]
		// 验证端口
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 0 || port > 65535 {
			return errors.New("无效的端口号")
		}
		// host 部分可以是 IP 或域名
		if net.ParseIP(host) != nil {
			return nil
		}
		if domainRegex.MatchString(host) {
			return nil
		}
		return errors.New("无效的域名格式")
	}

	// 纯 IP
	if net.ParseIP(domain) != nil {
		return nil
	}

	// 纯域名
	if !domainRegex.MatchString(domain) {
		return errors.New("无效的域名格式")
	}

	return nil
}
