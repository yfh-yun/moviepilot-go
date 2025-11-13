// Package main 提供IP工具使用示例
package main

import (
	"fmt"
	
	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== IP工具使用示例 ===")
	
	ipUtils := &utils.IpUtils{}
	
	// 示例1: 判断IPv4地址
	fmt.Println("\n1. 判断IPv4地址:")
	ipv4Addresses := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"invalid.ip",
		"2001:db8::1", // IPv6
	}
	
	for _, ip := range ipv4Addresses {
		isIPv4 := ipUtils.IsIPv4(ip)
		fmt.Printf("  %s: %t\n", ip, isIPv4)
	}
	
	// 示例2: 判断IPv6地址
	fmt.Println("\n2. 判断IPv6地址:")
	ipv6Addresses := []string{
		"2001:db8::1",
		"::1",
		"fe80::1",
		"192.168.1.1", // IPv4
		"invalid.ip",
	}
	
	for _, ip := range ipv6Addresses {
		isIPv6 := ipUtils.IsIPv6(ip)
		fmt.Printf("  %s: %t\n", ip, isIPv6)
	}
	
	// 示例3: 判断是否为IP地址
	fmt.Println("\n3. 判断是否为IP地址:")
	addresses := []string{
		"192.168.1.1",
		"2001:db8::1",
		"::1",
		"example.com",
		"invalid.ip",
	}
	
	for _, addr := range addresses {
		isIP := ipUtils.IsIP(addr)
		fmt.Printf("  %s: %t\n", addr, isIP)
	}
	
	// 示例4: 判断是否为内网IP
	fmt.Println("\n4. 判断是否为内网IP:")
	privateIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"127.0.0.1",
		"::1",
		"8.8.8.8",
		"2001:db8::1",
		" 192.168.1.1 ", // 带空格的IP
	}
	
	for _, ip := range privateIPs {
		isPrivate := ipUtils.IsPrivateIP(ip)
		fmt.Printf("  '%s': %t\n", ip, isPrivate)
	}
	
	// 示例5: 判断主机是否为内�?	fmt.Println("\n5. 判断主机是否为内�?")
	hosts := []string{
		"http://192.168.1.1:8080",
		"https://10.0.0.1:3000",
		"http://example.com",
		"172.16.0.1:8000",
		"localhost",
		"127.0.0.1",
	}
	
	for _, host := range hosts {
		isInternal := ipUtils.IsInternal(host)
		fmt.Printf("  %s: %t\n", host, isInternal)
	}
	
	// 示例6: 判断域名是否为内部域�?	fmt.Println("\n6. 判断域名是否为内部域�?")
	domains := []string{
		"localhost",
		"example.com",
		"google.com",
	}
	
	for _, domain := range domains {
		isInternalDomain := ipUtils.IsInternalDomain(domain)
		fmt.Printf("  %s: %t\n", domain, isInternalDomain)
	}
	
	// 示例7: 获取域名对应的IP地址
	fmt.Println("\n7. 获取域名对应的IP地址:")
	domain := "google.com"
	ips, err := ipUtils.GetIPAddresses(domain)
	if err != nil {
		fmt.Printf("  获取 %s 的IP地址失败: %v\n", domain, err)
	} else {
		fmt.Printf("  %s 的IP地址: %v\n", domain, ips)
	}
	
	// 示例8: 验证IP地址和端口组�?	fmt.Println("\n8. 验证IP地址和端口组�?")
	ipPorts := []struct {
		ip   string
		port string
	}{
		{"192.168.1.1", "8080"},
		{"2001:db8::1", "3000"},
		{"invalid.ip", "80"},
		{"192.168.1.1", "99999"}, // 无效端口
	}
	
	for _, item := range ipPorts {
		isValid := ipUtils.IsValidIPPort(item.ip, item.port)
		fmt.Printf("  %s:%s: %t\n", item.ip, item.port, isValid)
	}
	
	// 示例9: 判断回环地址
	fmt.Println("\n9. 判断回环地址:")
	loopbackIPs := []string{
		"127.0.0.1",
		"::1",
		"192.168.1.1",
		"8.8.8.8",
	}
	
	for _, ip := range loopbackIPs {
		isLoopback := ipUtils.IsLoopback(ip)
		fmt.Printf("  %s: %t\n", ip, isLoopback)
	}
	
	// 示例10: 判断全局单播地址
	fmt.Println("\n10. 判断全局单播地址:")
	unicastIPs := []string{
		"8.8.8.8",
		"2001:db8::1",
		"127.0.0.1",
		"::1",
	}
	
	for _, ip := range unicastIPs {
		isUnicast := ipUtils.IsGlobalUnicast(ip)
		fmt.Printf("  %s: %t\n", ip, isUnicast)
	}
	
	// 示例11: 获取IP版本
	fmt.Println("\n11. 获取IP版本:")
	versionIPs := []string{
		"192.168.1.1",
		"2001:db8::1",
		"::1",
		"invalid.ip",
	}
	
	for _, ip := range versionIPs {
		version, valid := ipUtils.GetIPVersion(ip)
		if valid {
			fmt.Printf("  %s: IPv%d\n", ip, version)
		} else {
			fmt.Printf("  %s: 无效IP\n", ip)
		}
	}
}
