package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// DoHClient DoH客户端
type DoHClient struct {
	httpClient *http.Client
	resolvers  []string
	timeout    time.Duration
	cache      map[string]string
	cacheMutex sync.RWMutex
}

// NewDoHClient 创建DoH客户端
func NewDoHClient(resolvers []string, timeout time.Duration) *DoHClient {
	return &DoHClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		resolvers: resolvers,
		timeout:   timeout,
		cache:     make(map[string]string),
	}
}

// DNSQueryResult DNS查询结果
type DNSQueryResult struct {
	Status     int             `json:"Status"`
	TC         bool            `json:"TC"`
	RD         bool            `json:"RD"`
	RA         bool            `json:"RA"`
	AD         bool            `json:"AD"`
	CD         bool            `json:"CD"`
	Question   []DNSQuestion   `json:"Question"`
	Answer     []DNSAnswer     `json:"Answer"`
	Authority  []DNSAuthority  `json:"Authority,omitempty"`
	Additional []DNSAdditional `json:"Additional,omitempty"`
}

// DNSQuestion DNS查询问题
type DNSQuestion struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

// DNSAnswer DNS查询答案
type DNSAnswer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

// DNSAuthority DNS权威记录
type DNSAuthority struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

// DNSAdditional DNS附加记录
type DNSAdditional struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

// QueryJSON 使用JSON格式查询DNS
func (c *DoHClient) QueryJSON(ctx context.Context, host string, qtype string) (string, error) {
	c.logger().Debug("使用JSON格式查询DNS", zap.String("host", host), zap.String("type", qtype))

	// 检查缓存
	cacheKey := fmt.Sprintf("%s:%s", host, qtype)
	if ip, ok := c.getFromCache(cacheKey); ok {
		c.logger().Debug("从缓存获取DNS结果", zap.String("host", host), zap.String("ip", ip))
		return ip, nil
	}

	// 遍历解析器
	for _, resolver := range c.resolvers {
		// 构建请求URL
		reqURL := fmt.Sprintf("https://%s/dns-query?name=%s&type=%s", resolver, url.QueryEscape(host), qtype)
		c.logger().Debug("发送DoH JSON请求", zap.String("url", reqURL))

		// 创建请求
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			c.logger().Error("创建DoH请求失败", zap.Error(err))
			continue
		}
		req.Header.Set("Accept", "application/dns-json")

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger().Error("发送DoH请求失败", zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			c.logger().Error("DoH响应状态错误", zap.Int("status", resp.StatusCode))
			continue
		}

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger().Error("读取DoH响应失败", zap.Error(err))
			continue
		}

		// 解析JSON响应
		var result DNSQueryResult
		if err := json.Unmarshal(body, &result); err != nil {
			c.logger().Error("解析DoH响应失败", zap.Error(err))
			continue
		}

		// 检查查询状态
		if result.Status != 0 {
			c.logger().Error("DoH查询失败", zap.Int("status", result.Status))
			continue
		}

		// 提取第一个A记录
		for _, answer := range result.Answer {
			if answer.Type == 1 { // A记录
				// 缓存结果
				c.addToCache(cacheKey, answer.Data)
				return answer.Data, nil
			}
		}
	}

	return "", fmt.Errorf("所有DoH解析器查询失败")
}

// QueryMessage 使用DNS消息格式查询DNS
func (c *DoHClient) QueryMessage(ctx context.Context, host string) (string, error) {
	c.logger().Debug("使用DNS消息格式查询DNS", zap.String("host", host))

	// 检查缓存
	cacheKey := fmt.Sprintf("%s:A", host)
	if ip, ok := c.getFromCache(cacheKey); ok {
		c.logger().Debug("从缓存获取DNS结果", zap.String("host", host), zap.String("ip", ip))
		return ip, nil
	}

	// 构建DNS查询消息
	message, err := c.buildDNSQuery(host, 1) // 1 = A记录
	if err != nil {
		return "", fmt.Errorf("构建DNS查询消息失败: %w", err)
	}

	// 遍历解析器
	for _, resolver := range c.resolvers {
		// 发送请求
		ip, err := c.sendDNSQuery(ctx, resolver, message)
		if err != nil {
			c.logger().Error("发送DNS查询失败", zap.Error(err))
			continue
		}

		// 缓存结果
		c.addToCache(cacheKey, ip)
		return ip, nil
	}

	return "", fmt.Errorf("所有DoH解析器查询失败")
}

// buildDNSQuery 构建DNS查询消息
func (c *DoHClient) buildDNSQuery(host string, qtype uint16) ([]byte, error) {
	var buf bytes.Buffer

	// DNS Header
	// ID: 0
	binary.Write(&buf, binary.BigEndian, uint16(0))
	// Flags: 标准递归查询
	binary.Write(&buf, binary.BigEndian, uint16(0x0100))
	// QDCOUNT: 1
	binary.Write(&buf, binary.BigEndian, uint16(1))
	// ANCOUNT: 0
	binary.Write(&buf, binary.BigEndian, uint16(0))
	// NSCOUNT: 0
	binary.Write(&buf, binary.BigEndian, uint16(0))
	// ARCOUNT: 0
	binary.Write(&buf, binary.BigEndian, uint16(0))

	// DNS Question
	// QNAME: 域名序列
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if len(label) > 63 {
			return nil, fmt.Errorf("域名标签过长")
		}
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	buf.WriteByte(0) // 结束标签
	// QTYPE: A
	binary.Write(&buf, binary.BigEndian, qtype)
	// QCLASS: IN
	binary.Write(&buf, binary.BigEndian, uint16(1))

	return buf.Bytes(), nil
}

// sendDNSQuery 发送DNS查询消息
func (c *DoHClient) sendDNSQuery(ctx context.Context, resolver string, message []byte) (string, error) {
	// 编码DNS消息
	b64message := base64.RawURLEncoding.EncodeToString(message)

	// 构建请求URL
	reqURL := fmt.Sprintf("https://%s/dns-query?dns=%s", resolver, b64message)
	c.logger().Debug("发送DoH消息请求", zap.String("url", reqURL))

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建DoH请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/dns-message")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送DoH请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DoH响应状态错误: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取DoH响应失败: %w", err)
	}

	// 解析DNS响应
	// 跳过头部和问题部分，直接获取第一个A记录
	// 头部: 12字节
	// 问题部分: 域名长度 + 4字节(QTYPE + QCLASS)
	// 答案部分: 12字节(NAME + TYPE + CLASS + TTL) + 2字节(RDLENGTH) + 4字节(RDATA)
	// 这里简化处理，直接解析第一个A记录
	// 实际实现中应该完整解析DNS响应

	// 查找第一个A记录
	pos := 12 // 跳过头部

	// 跳过问题部分
	for {
		labelLen := int(body[pos])
		if labelLen == 0 {
			pos++
			break
		}
		pos += labelLen + 1
	}
	pos += 4 // 跳过QTYPE和QCLASS

	// 读取答案部分
	if pos+16 > len(body) {
		return "", fmt.Errorf("DNS响应格式错误")
	}

	// 跳过NAME(压缩)、TYPE、CLASS、TTL
	pos += 12

	// 读取RDLENGTH
	rdLength := binary.BigEndian.Uint16(body[pos : pos+2])
	pos += 2

	// 读取RDATA
	if pos+int(rdLength) > len(body) {
		return "", fmt.Errorf("DNS响应格式错误")
	}

	// 解析A记录
	if rdLength != 4 {
		return "", fmt.Errorf("DNS响应不是A记录")
	}

	// 转换为IP地址
	ip := fmt.Sprintf("%d.%d.%d.%d", body[pos], body[pos+1], body[pos+2], body[pos+3])
	return ip, nil
}

// getFromCache 从缓存获取
func (c *DoHClient) getFromCache(key string) (string, bool) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	ip, ok := c.cache[key]
	return ip, ok
}

// addToCache 添加到缓存
func (c *DoHClient) addToCache(key, value string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache[key] = value
}

// ClearCache 清空缓存
func (c *DoHClient) ClearCache() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache = make(map[string]string)
}

// logger 获取日志记录器
func (c *DoHClient) logger() *zap.Logger {
	return logger.GetLogger()
}
