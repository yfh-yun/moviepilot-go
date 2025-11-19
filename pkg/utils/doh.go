package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DoHHelper DNS over HTTPS助手
type DoHHelper struct {
	enabled     bool
	domains     []string
	dohServers  []string
	cache       map[string]*DNSCacheEntry
	cacheMutex  sync.RWMutex
	httpClient  *http.Client
	timeout     time.Duration
	originalResolver net.Resolver
}

// DNSCacheEntry DNS缓存条目
type DNSCacheEntry struct {
	IP       string
	ExpireAt time.Time
}

// DoHRequest DoH请求结构
type DoHRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// DoHResponse DoH响应结构
type DoHResponse struct {
	Status int `json:"Status"`
	TC     bool `json:"TC"`
	RD     bool `json:"RD"`
	RA     bool `json:"RA"`
	AD     bool `json:"AD"`
	CD     bool `json:"CD"`
	Question []DoHQuestion `json:"Question"`
	Answer   []DoHAnswer   `json:"Answer"`
}

// DoHQuestion DoH问题
type DoHQuestion struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

// DoHAnswer DoH答案
type DoHAnswer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

// NewDoHHelper 创建DoH助手实例
func NewDoHHelper() *DoHHelper {
	return &DoHHelper{
		enabled:    false,
		domains:    []string{},
		dohServers: []string{
			"https://cloudflare-dns.com/dns-query",
			"https://dns.google/resolve",
			"https://1.1.1.1/dns-query",
		},
		cache:       make(map[string]*DNSCacheEntry),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		timeout: 5 * time.Second,
	}
}

// Enable 启用DoH
func (doh *DoHHelper) Enable(domains []string) error {
	if domains == nil || len(domains) == 0 {
		return fmt.Errorf("domains list cannot be empty")
	}

	doh.enabled = true
	doh.domains = domains

	// 保存原始解析器
	doh.originalResolver = *net.Resolver{}

	// 替换网络解析器
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: doh.timeout,
			}
			
			// 对于DoH域名，使用自定义解析
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				host = address
				port = "53"
			}

			if doh.shouldUseDoH(host) {
				if ip := doh.getCachedIP(host); ip != "" {
					address = net.JoinHostPort(ip, port)
				}
			}

			return d.DialContext(ctx, network, address)
		},
	}

	return nil
}

// Disable 禁用DoH
func (doh *DoHHelper) Disable() {
	doh.enabled = false
	doh.domains = []string{}
	
	// 恢复原始解析器
	net.DefaultResolver = &doh.originalResolver
}

// IsEnabled 检查DoH是否启用
func (doh *DoHHelper) IsEnabled() bool {
	return doh.enabled
}

// SetDomains 设置需要使用DoH的域名
func (doh *DoHHelper) SetDomains(domains []string) {
	doh.domains = domains
}

// GetDomains 获取需要使用DoH的域名列表
func (doh *DoHHelper) GetDomains() []string {
	return doh.domains
}

// SetDoHServers 设置DoH服务器列表
func (doh *DoHHelper) SetDoHServers(servers []string) {
	doh.dohServers = servers
}

// GetDoHServers 获取DoH服务器列表
func (doh *DoHHelper) GetDoHServers() []string {
	return doh.dohServers
}

// SetTimeout 设置超时时间
func (doh *DoHHelper) SetTimeout(timeout time.Duration) {
	doh.timeout = timeout
	doh.httpClient.Timeout = timeout
}

// GetTimeout 获取超时时间
func (doh *DoHHelper) GetTimeout() time.Duration {
	return doh.timeout
}

// shouldUseDoH 检查是否应该使用DoH
func (doh *DoHHelper) shouldUseDoH(domain string) bool {
	if !doh.enabled {
		return false
	}

	for _, d := range doh.domains {
		if strings.HasSuffix(domain, d) {
			return true
		}
	}

	return false
}

// getCachedIP 获取缓存的IP地址
func (doh *DoHHelper) getCachedIP(domain string) string {
	doh.cacheMutex.RLock()
	defer doh.cacheMutex.RUnlock()

	if entry, exists := doh.cache[domain]; exists {
		if time.Now().Before(entry.ExpireAt) {
			return entry.IP
		}
	}

	return ""
}

// setCachedIP 设置缓存的IP地址
func (doh *DoHHelper) setCachedIP(domain, ip string, ttl int) {
	doh.cacheMutex.Lock()
	defer doh.cacheMutex.Unlock()

	expireAt := time.Now().Add(time.Duration(ttl) * time.Second)
	doh.cache[domain] = &DNSCacheEntry{
		IP:       ip,
		ExpireAt: expireAt,
	}
}

// ResolveDomain 使用DoH解析域名
func (doh *DoHHelper) ResolveDomain(domain string) ([]string, error) {
	if !doh.shouldUseDoH(domain) {
		// 使用标准DNS解析
		return net.LookupHost(domain)
	}

	// 检查缓存
	if cachedIP := doh.getCachedIP(domain); cachedIP != "" {
		return []string{cachedIP}, nil
	}

	// 使用DoH解析
	ips, err := doh.resolveWithDoH(domain)
	if err != nil {
		// DoH失败，回退到标准DNS
		return net.LookupHost(domain)
	}

	return ips, nil
}

// resolveWithDoH 使用DoH解析域名
func (doh *DoHHelper) resolveWithDoH(domain string) ([]string, error) {
	var lastError error

	// 尝试所有DoH服务器
	for _, server := range doh.dohServers {
		ips, err := doh.queryDoHServer(server, domain)
		if err != nil {
			lastError = err
			continue
		}

		if len(ips) > 0 {
			return ips, nil
		}
	}

	return nil, fmt.Errorf("all DoH servers failed, last error: %v", lastError)
}

// queryDoHServer 查询DoH服务器
func (doh *DoHHelper) queryDoHServer(server, domain string) ([]string, error) {
	// 构建DoH请求
	request := DoHRequest{
		Name: domain,
		Type: "A", // IPv4地址记录
	}

	// 转换为DNS消息格式（简化实现）
	dnsMessage := doh.buildDNSMessage(request)
	
	// Base64编码
	encodedMessage := base64.RawURLEncoding.EncodeToString(dnsMessage)

	// 构建URL
	var url string
	if strings.Contains(server, "dns.google") {
		url = fmt.Sprintf("%s?name=%s&type=A", server, domain)
	} else {
		url = fmt.Sprintf("%s?dns=%s", server, encodedMessage)
	}

	// 发送HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("User-Agent", "MoviePilot-DoH/1.0")

	resp, err := doh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	// 解析响应
	var dohResponse DoHResponse
	if err := json.NewDecoder(resp.Body).Decode(&dohResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// 提取IP地址
	var ips []string
	for _, answer := range dohResponse.Answer {
		if answer.Type == 1 { // A记录
			ips = append(ips, answer.Data)
			// 缓存结果
			doh.setCachedIP(domain, answer.Data, answer.TTL)
		}
	}

	return ips, nil
}

// buildDNSMessage 构建DNS消息（简化实现）
func (doh *DoHHelper) buildDNSMessage(request DoHRequest) []byte {
	// 这里应该实现完整的DNS消息格式
	// 简化实现，返回一个基本的DNS查询消息
	message := make([]byte, 12) // DNS头部
	
	// 设置ID
	message[0] = 0x12
	message[1] = 0x34
	
	// 设置标志（标准查询）
	message[2] = 0x01 // QR=0, Opcode=0, AA=0, TC=0, RD=1
	message[3] = 0x00 // RA=0, Z=0, RCODE=0
	
	// 设置问题数量
	message[4] = 0x00
	message[5] = 0x01
	
	// 设置答案、授权、附加记录数量为0
	message[6] = 0x00
	message[7] = 0x00
	message[8] = 0x00
	message[9] = 0x00
	message[10] = 0x00
	message[11] = 0x00
	
	// 添加查询部分
	parts := strings.Split(request.Name, ".")
	for _, part := range parts {
		if len(part) > 0 {
			message = append(message, byte(len(part)))
			message = append(message, []byte(part)...)
		}
	}
	message = append(message, 0x00) // 结束标签
	
	// 添加查询类型和类别
	message = append(message, 0x00, 0x01) // A记录
	message = append(message, 0x00, 0x01) // IN类别
	
	return message
}

// ClearCache 清空缓存
func (doh *DoHHelper) ClearCache() {
	doh.cacheMutex.Lock()
	defer doh.cacheMutex.Unlock()

	doh.cache = make(map[string]*DNSCacheEntry)
}

// GetCacheSize 获取缓存大小
func (doh *DoHHelper) GetCacheSize() int {
	doh.cacheMutex.RLock()
	defer doh.cacheMutex.RUnlock()

	return len(doh.cache)
}

// GetCacheEntries 获取缓存条目
func (doh *DoHHelper) GetCacheEntries() map[string]*DNSCacheEntry {
	doh.cacheMutex.RLock()
	defer doh.cacheMutex.RUnlock()

	// 返回副本
	entries := make(map[string]*DNSCacheEntry)
	for domain, entry := range doh.cache {
		entries[domain] = &DNSCacheEntry{
			IP:       entry.IP,
			ExpireAt: entry.ExpireAt,
		}
	}

	return entries
}

// RemoveExpiredEntries 移除过期的缓存条目
func (doh *DoHHelper) RemoveExpiredEntries() {
	doh.cacheMutex.Lock()
	defer doh.cacheMutex.Unlock()

	now := time.Now()
	for domain, entry := range doh.cache {
		if now.After(entry.ExpireAt) {
			delete(doh.cache, domain)
		}
	}
}

// TestDoHServer 测试DoH服务器
func (doh *DoHHelper) TestDoHServer(server string) error {
	// 使用一个常见的域名进行测试
	testDomain := "google.com"
	
	_, err := doh.queryDoHServer(server, testDomain)
	if err != nil {
		return fmt.Errorf("DoH server test failed: %v", err)
	}

	return nil
}

// TestAllDoHServers 测试所有DoH服务器
func (doh *DoHHelper) TestAllDoHServers() map[string]error {
	results := make(map[string]error)

	for _, server := range doh.dohServers {
		err := doh.TestDoHServer(server)
		results[server] = err
	}

	return results
}

// GetStats 获取统计信息
func (doh *DoHHelper) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":     doh.enabled,
		"domains":     len(doh.domains),
		"servers":     len(doh.dohServers),
		"cache_size":  doh.GetCacheSize(),
		"timeout":     doh.timeout.String(),
	}

	return stats
}

// ExportConfig 导出配置
func (doh *DoHHelper) ExportConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    doh.enabled,
		"domains":    doh.domains,
		"servers":    doh.dohServers,
		"timeout":    doh.timeout.String(),
	}
}

// ImportConfig 导入配置
func (doh *DoHHelper) ImportConfig(config map[string]interface{}) error {
	if enabled, ok := config["enabled"].(bool); ok {
		if enabled {
			if domains, ok := config["domains"].([]string); ok {
				if err := doh.Enable(domains); err != nil {
					return err
				}
			}
		} else {
			doh.Disable()
		}
	}

	if domains, ok := config["domains"].([]string); ok {
		doh.SetDomains(domains)
	}

	if servers, ok := config["servers"].([]string); ok {
		doh.SetDoHServers(servers)
	}

	if timeoutStr, ok := config["timeout"].(string); ok {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			doh.SetTimeout(timeout)
		}
	}

	return nil
}