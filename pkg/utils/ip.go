package utils

import (
	"net"
	"net/url"
	"strings"
)

// IsIPv4 判断是否为 IPv4 地址
func IsIPv4(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil && parsed.To4() != nil
}

// IsIPv6 判断是否为 IPv6 地址
func IsIPv6(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil && parsed.To4() == nil
}

// IsIP 判断字符串是否为合法 IP（仅IPv4，与Python is_ip函数行为一致）
func IsIP(addr string) bool {
	addr = strings.TrimSpace(addr)
	// 使用ParseIP检查是否为IP，然后检查是否为IPv4
	parsed := net.ParseIP(addr)
	return parsed != nil && parsed.To4() != nil
}

// IsInternalDomain 判断域名是否解析为内网 IP
func IsInternalDomain(domain string) bool {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return false
	}

	ips, err := net.LookupHost(domain)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if IsPrivateIP(ip) {
			return true
		}
	}
	return false
}

// IsInternal 判断一个 host/url 是否为内网地址
// 行为对应 Python IpUtils.is_internal：
//   - 先从 URL 中解析 hostname
//   - 若本身是 IP，则判断是否内网 IP
//   - 否则按域名解析再判断
func IsInternal(hostname string) bool {
	parsed, err := url.Parse(hostname)
	if err == nil && parsed.Hostname() != "" {
		hostname = parsed.Hostname()
	}

	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return false
	}

	if IsIP(hostname) {
		return IsPrivateIP(hostname)
	}

	return IsInternalDomain(hostname)
}

// IsPrivateIP 判断 IP 是否为内网或本地地址
func IsPrivateIP(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}

	// 处理回环地址和链路本地地址
	if parsed.IsLoopback() || parsed.IsLinkLocalUnicast() || parsed.IsLinkLocalMulticast() {
		return true
	}

	// 私有 IPv4 网段：
	// 10.0.0.0/8
	// 172.16.0.0/12
	// 192.168.0.0/16
	// 169.254.0.0/16（链路本地）
	privateCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
	}

	// 常见私有 IPv6 网段：
	// fc00::/7（唯一本地地址）
	// fe80::/10（链路本地）
	privateCIDRs = append(privateCIDRs,
		"fc00::/7",
		"fe80::/10",
	)

	for _, cidr := range privateCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsed) {
			return true
		}
	}

	return false
}
