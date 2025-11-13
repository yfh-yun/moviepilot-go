// Package utils 提供IP地址处理相关的工具函�?package utils

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// IpUtils IP地址处理工具�?type IpUtils struct{}

// IsIPv4 判断是不是IPv4地址
// ip: 待检查的IP地址
// 返回是否为IPv4地址
func (i *IpUtils) IsIPv4(ip string) bool {
	// 使用Go标准库的net.ParseIP来解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// IPv4地址在Go中会被表示为IPv4格式（长度为4字节�?	// 或者兼容IPv6格式（前12字节�?，后4字节为IPv4地址�?	return parsedIP.To4() != nil
}

// IsIPv6 判断是不是IPv6地址
// ip: 待检查的IP地址
// 返回是否为IPv6地址
func (i *IpUtils) IsIPv6(ip string) bool {
	// 使用Go标准库的net.ParseIP来解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// 如果是IPv4地址，To4()会返回非nil�?	// 如果是IPv6地址，To4()会返回nil，而IP本身不为nil
	return parsedIP.To4() == nil && parsedIP.To16() != nil
}

// IsInternal 判断一个host是内网还是外�?// hostname: 待检查的主机名或URL
// 返回是否为内网地址
func (i *IpUtils) IsInternal(hostname string) bool {
	// 解析URL获取主机�?	parsedURL, err := url.Parse(hostname)
	if err == nil && parsedURL.Hostname() != "" {
		hostname = parsedURL.Hostname()
	}
	
	// 判断是否为IP地址
	if i.IsIP(hostname) {
		return i.IsPrivateIP(hostname)
	} else {
		return i.IsInternalDomain(hostname)
	}
}

// IsIP 判断是不是IP地址
// addr: 待检查的地址
// 返回是否为IP地址
func (i *IpUtils) IsIP(addr string) bool {
	// 使用Go标准库的net.ParseIP来解析IP地址
	return net.ParseIP(addr) != nil
}

// IsInternalDomain 判断域名是否为内部域�?// domain: 待检查的域名
// 返回是否为内部域�?func (i *IpUtils) IsInternalDomain(domain string) bool {
	// 获取域名对应的IP地址
	ips, err := net.LookupIP(domain)
	if err != nil {
		return false
	}
	
	// 判断任何一个IP地址是否属于内网IP地址范围
	for _, ip := range ips {
		if i.IsPrivateIP(ip.String()) {
			return true
		}
	}
	
	return false
}

// IsPrivateIP 判断是不是内网IP
// ipStr: 待检查的IP地址字符�?// 返回是否为内网IP
func (i *IpUtils) IsPrivateIP(ipStr string) bool {
	// 去除首尾空格
	ipStr = strings.TrimSpace(ipStr)
	
	// 使用Go标准库的net.ParseIP来解析IP地址
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	
	// 使用IP地址的IsPrivate方法判断是否为私有地址
	// 这包括以下范围：
	// - IPv4: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8
	// - IPv6: fd00::/8, ::1
	return ip.IsPrivate()
}

// GetIPAddresses 获取主机名对应的IP地址列表
// hostname: 主机�?// 返回IP地址列表和错误信�?func (i *IpUtils) GetIPAddresses(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, fmt.Errorf("无法解析主机�?%s: %v", hostname, err)
	}
	
	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}
	
	return ipStrings, nil
}

// IsValidIPPort 检查IP地址和端口组合是否有�?// ip: IP地址
// port: 端口�?// 返回是否有效
func (i *IpUtils) IsValidIPPort(ip string, port string) bool {
	// 组合IP和端�?	addr := net.JoinHostPort(ip, port)
	
	// 使用TCP地址解析来验�?	_, err := net.ResolveTCPAddr("tcp", addr)
	return err == nil
}

// IsLoopback 判断是否为回环地址
// ip: IP地址
// 返回是否为回环地址
func (i *IpUtils) IsLoopback(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsLoopback()
}

// IsGlobalUnicast 判断是否为全局单播地址
// ip: IP地址
// 返回是否为全局单播地址
func (i *IpUtils) IsGlobalUnicast(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsGlobalUnicast()
}

// GetIPVersion 获取IP地址版本
// ip: IP地址
// 返回IP版本�?4�?)和是否有效的布尔�?func (i *IpUtils) GetIPVersion(ip string) (int, bool) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0, false
	}
	
	if parsedIP.To4() != nil {
		return 4, true
	}
	return 6, true
}
