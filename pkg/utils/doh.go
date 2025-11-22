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

	"moviepilot-go/pkg/errors"
	"moviepilot-go/pkg/logger"
	"go.uber.org/zap"
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
	logger.Debug("Creating new DoHHelper instance", zap.String("func", "NewDoHHelper"))
	
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
	logger.Debug("Enabling DoH", zap.String("func", "Enable"), zap.Strings("domains", domains))
	
	if domains == nil || len(domains) == 0 {
		logger.Error("Domains list cannot be empty", zap.String("func", "Enable"))
		return errors.NewAppError(http.StatusBadRequest, "domains list cannot be empty", "")
	}

	// 验证域名格式
	for _, domain := range domains {
		if domain == "" {
			logger.Error("Domain cannot be empty", zap.String("func", "Enable"))
			return errors.NewAppError(http.StatusBadRequest, "domain cannot be empty", "")
		}
	}

	doh.enabled = true
	doh.domains = domains

	// 保存原始解析器
	doh.originalResolver = net.Resolver{}

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
					logger.Debug("Using cached IP for DoH domain", 
						zap.String("func", "Enable.Dial"),
						zap.String("domain", host),
						zap.String("ip", ip))
				}
			}

			return d.DialContext(ctx, network, address)
		},
	}

	logger.Info("DoH enabled successfully", zap.String("func", "Enable"), zap.Strings("domains", domains))
	return nil
}

// Disable 禁用DoH
func (doh *DoHHelper) Disable() {
	logger.Debug("Disabling DoH", zap.String("func", "Disable"))
	
	doh.enabled = false
	doh.domains = []string{}
	
	// 恢复原始解析器
	net.DefaultResolver = &doh.originalResolver
	
	logger.Info("DoH disabled successfully", zap.String("func", "Disable"))
}

// IsEnabled 检查DoH是否启用
func (doh *DoHHelper) IsEnabled() bool {
	return doh.enabled
}

// SetDomains 设置需要使用DoH的域名
func (doh *DoHHelper) SetDomains(domains []string) {
	logger.Debug("Setting DoH domains", zap.String("func", "SetDomains"), zap.Strings("domains", domains))
	
	if domains == nil {
		logger.Warn("Domains list is nil, setting to empty", zap.String("func", "SetDomains"))
		doh.domains = []string{}
		return
	}
	
	// 验证域名格式
	for _, domain := range domains {
		if domain == "" {
			logger.Error("Domain cannot be empty", zap.String("func", "SetDomains"))
			return
		}
	}
	
	doh.domains = domains
	logger.Info("DoH domains updated", zap.String("func", "SetDomains"), zap.Strings("domains", domains))
}

// GetDomains 获取需要使用DoH的域名列表
func (doh *DoHHelper) GetDomains() []string {
	return doh.domains
}

// SetDoHServers 设置DoH服务器列表
func (doh *DoHHelper) SetDoHServers(servers []string) {
	logger.Debug("Setting DoH servers", zap.String("func", "SetDoHServers"), zap.Strings("servers", servers))
	
	if servers == nil {
		logger.Warn("Servers list is nil, setting to empty", zap.String("func", "SetDoHServers"))
		doh.dohServers = []string{}
		return
	}
	
	// 验证服务器URL格式
	for _, server := range servers {
		if server == "" {
			logger.Error("Server URL cannot be empty", zap.String("func", "SetDoHServers"))
			return
		}
		if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
			logger.Error("Server URL must start with http:// or https://", 
				zap.String("func", "SetDoHServers"), zap.String("server", server))
			return
		}
	}
	
	doh.dohServers = servers
	logger.Info("DoH servers updated", zap.String("func", "SetDoHServers"), zap.Strings("servers", servers))
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
	logger.Debug("Caching IP address", 
		zap.String("func", "setCachedIP"), 
		zap.String("domain", domain), 
		zap.String("ip", ip), 
		zap.Int("ttl", ttl))
	
	if domain == "" {
		logger.Error("Domain cannot be empty", zap.String("func", "setCachedIP"))
		return
	}
	
	if ip == "" {
		logger.Error("IP cannot be empty", zap.String("func", "setCachedIP"))
		return
	}
	
	if ttl <= 0 {
		logger.Warn("Invalid TTL, using default 300 seconds", 
			zap.String("func", "setCachedIP"), 
			zap.Int("ttl", ttl))
		ttl = 300 // 默认5分钟
	}

	doh.cacheMutex.Lock()
	defer doh.cacheMutex.Unlock()

	expireAt := time.Now().Add(time.Duration(ttl) * time.Second)
	doh.cache[domain] = &DNSCacheEntry{
		IP:       ip,
		ExpireAt: expireAt,
	}
	
	logger.Debug("IP cached successfully", 
		zap.String("func", "setCachedIP"), 
		zap.String("domain", domain), 
		zap.String("ip", ip), 
		zap.Time("expire_at", expireAt))
}

// ResolveDomain 使用DoH解析域名
func (doh *DoHHelper) ResolveDomain(domain string) ([]string, error) {
	logger.Debug("Resolving domain", zap.String("func", "ResolveDomain"), zap.String("domain", domain))
	
	if domain == "" {
		logger.Error("Domain cannot be empty", zap.String("func", "ResolveDomain"))
		return nil, errors.NewAppError(http.StatusBadRequest, "domain cannot be empty", "")
	}

	if !doh.shouldUseDoH(domain) {
		logger.Debug("Using standard DNS for domain", zap.String("func", "ResolveDomain"), zap.String("domain", domain))
		// 使用标准DNS解析
		ips, err := net.LookupHost(domain)
		if err != nil {
			logger.Error("Standard DNS lookup failed", 
				zap.String("func", "ResolveDomain"), 
				zap.String("domain", domain), 
				zap.Error(err))
			return nil, errors.WrapError(err, fmt.Sprintf("failed to lookup domain %s", domain))
		}
		return ips, nil
	}

	// 检查缓存
	if cachedIP := doh.getCachedIP(domain); cachedIP != "" {
		logger.Debug("Using cached IP for domain", 
			zap.String("func", "ResolveDomain"), 
			zap.String("domain", domain), 
			zap.String("ip", cachedIP))
		return []string{cachedIP}, nil
	}

	// 使用DoH解析
	ips, err := doh.resolveWithDoH(domain)
	if err != nil {
		logger.Warn("DoH resolution failed, falling back to standard DNS", 
			zap.String("func", "ResolveDomain"), 
			zap.String("domain", domain), 
			zap.Error(err))
		// DoH失败，回退到标准DNS
		ips, fallbackErr := net.LookupHost(domain)
		if fallbackErr != nil {
			logger.Error("Fallback DNS lookup also failed", 
				zap.String("func", "ResolveDomain"), 
				zap.String("domain", domain), 
				zap.Error(fallbackErr))
			return nil, errors.WrapError(fallbackErr, fmt.Sprintf("both DoH and standard DNS failed for domain %s", domain))
		}
		return ips, nil
	}

	logger.Info("Domain resolved successfully with DoH", 
		zap.String("func", "ResolveDomain"), 
		zap.String("domain", domain), 
		zap.Strings("ips", ips))
	return ips, nil
}

// resolveWithDoH 使用DoH解析域名
func (doh *DoHHelper) resolveWithDoH(domain string) ([]string, error) {
	logger.Debug("Resolving domain with DoH", zap.String("func", "resolveWithDoH"), zap.String("domain", domain))
	
	var lastError error

	// 尝试所有DoH服务器
	for _, server := range doh.dohServers {
		ips, err := doh.queryDoHServer(server, domain)
		if err != nil {
			logger.Warn("DoH server query failed", 
				zap.String("func", "resolveWithDoH"), 
				zap.String("server", server), 
				zap.String("domain", domain), 
				zap.Error(err))
			lastError = err
			continue
		}

		if len(ips) > 0 {
			logger.Info("DoH server query succeeded", 
				zap.String("func", "resolveWithDoH"), 
				zap.String("server", server), 
				zap.String("domain", domain), 
				zap.Strings("ips", ips))
			return ips, nil
		}
	}

	logger.Error("All DoH servers failed", 
		zap.String("func", "resolveWithDoH"), 
		zap.String("domain", domain), 
		zap.Error(lastError))
	return nil, errors.WrapError(lastError, fmt.Sprintf("all DoH servers failed for domain %s", domain))
}

// queryDoHServer 查询DoH服务器
func (doh *DoHHelper) queryDoHServer(server, domain string) ([]string, error) {
	logger.Debug("Querying DoH server", 
		zap.String("func", "queryDoHServer"), 
		zap.String("server", server), 
		zap.String("domain", domain))
	
	if server == "" {
		logger.Error("Server cannot be empty", zap.String("func", "queryDoHServer"))
		return nil, errors.NewAppError(http.StatusBadRequest, "server cannot be empty", "")
	}
	
	if domain == "" {
		logger.Error("Domain cannot be empty", zap.String("func", "queryDoHServer"))
		return nil, errors.NewAppError(http.StatusBadRequest, "domain cannot be empty", "")
	}

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
		logger.Error("Failed to create HTTP request", 
			zap.String("func", "queryDoHServer"), 
			zap.String("url", url), 
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to create request")
	}

	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("User-Agent", "MoviePilot-DoH/1.0")

	logger.Debug("Sending DoH request", 
		zap.String("func", "queryDoHServer"), 
		zap.String("url", url))

	resp, err := doh.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send HTTP request", 
			zap.String("func", "queryDoHServer"), 
			zap.String("url", url), 
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("DoH server returned non-OK status", 
			zap.String("func", "queryDoHServer"), 
			zap.String("server", server), 
			zap.Int("status_code", resp.StatusCode))
		return nil, errors.NewAppError(resp.StatusCode, 
			fmt.Sprintf("server returned status: %d", resp.StatusCode), "")
	}

	// 解析响应
	var dohResponse DoHResponse
	if err := json.NewDecoder(resp.Body).Decode(&dohResponse); err != nil {
		logger.Error("Failed to decode DoH response", 
			zap.String("func", "queryDoHServer"), 
			zap.String("server", server), 
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to decode response")
	}

	// 提取IP地址
	var ips []string
	for _, answer := range dohResponse.Answer {
		if answer.Type == 1 { // A记录
			ips = append(ips, answer.Data)
			// 缓存结果
			doh.setCachedIP(domain, answer.Data, answer.TTL)
			logger.Debug("Found A record", 
				zap.String("func", "queryDoHServer"), 
				zap.String("domain", domain), 
				zap.String("ip", answer.Data), 
				zap.Int("ttl", answer.TTL))
		}
	}

	if len(ips) == 0 {
		logger.Warn("No A records found in DoH response", 
			zap.String("func", "queryDoHServer"), 
			zap.String("domain", domain))
	}

	logger.Debug("DoH query completed", 
		zap.String("func", "queryDoHServer"), 
		zap.String("domain", domain), 
		zap.Strings("ips", ips))
	
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
	logger.Debug("Clearing DNS cache", zap.String("func", "ClearCache"))
	
	doh.cacheMutex.Lock()
	defer doh.cacheMutex.Unlock()

	cacheSize := len(doh.cache)
	doh.cache = make(map[string]*DNSCacheEntry)
	
	logger.Info("DNS cache cleared", zap.String("func", "ClearCache"), zap.Int("cleared_entries", cacheSize))
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
	logger.Debug("Testing DoH server", zap.String("func", "TestDoHServer"), zap.String("server", server))
	
	if server == "" {
		logger.Error("Server cannot be empty", zap.String("func", "TestDoHServer"))
		return errors.NewAppError(http.StatusBadRequest, "server cannot be empty", "")
	}
	
	// 使用一个常见的域名进行测试
	testDomain := "google.com"
	
	_, err := doh.queryDoHServer(server, testDomain)
	if err != nil {
		logger.Error("DoH server test failed", 
			zap.String("func", "TestDoHServer"), 
			zap.String("server", server), 
			zap.Error(err))
		return errors.WrapError(err, fmt.Sprintf("DoH server %s test failed", server))
	}

	logger.Info("DoH server test passed", 
		zap.String("func", "TestDoHServer"), 
		zap.String("server", server))
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
	logger.Debug("Importing DoH configuration", zap.String("func", "ImportConfig"))
	
	if config == nil {
		logger.Error("Config cannot be nil", zap.String("func", "ImportConfig"))
		return errors.NewAppError(http.StatusBadRequest, "config cannot be nil", "")
	}

	if enabled, ok := config["enabled"].(bool); ok {
		if enabled {
			if domains, ok := config["domains"].([]string); ok {
				if err := doh.Enable(domains); err != nil {
					logger.Error("Failed to enable DoH during config import", 
						zap.String("func", "ImportConfig"), 
						zap.Error(err))
					return errors.WrapError(err, "failed to enable DoH")
				}
			} else {
				logger.Warn("Enabled is true but no domains provided", zap.String("func", "ImportConfig"))
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
		} else {
			logger.Warn("Invalid timeout format in config", 
				zap.String("func", "ImportConfig"), 
				zap.String("timeout", timeoutStr), 
				zap.Error(err))
		}
	}

	logger.Info("DoH configuration imported successfully", zap.String("func", "ImportConfig"))
	return nil
}