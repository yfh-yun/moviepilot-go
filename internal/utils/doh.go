package utils

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"go.uber.org/zap"
)

// DoHHelper DoH帮助类，用于处理DNS over HTTPS解析
type DoHHelper struct {
	cache     map[string]string
	cacheLock sync.RWMutex
	origLookupHost func(ctx context.Context, network, host string) (addrs []string, err error)
}

// dohHelperInstance DoHHelper单例实例
var dohHelperInstance *DoHHelper
var dohHelperOnce sync.Once

// Answer DoH JSON响应中的Answer字段
type Answer struct {
	Data string `json:"data"`
}

// DoHResponse DoH JSON响应结构
type DoHResponse struct {
	Answer []Answer `json:"Answer"`
}

// NewDoHHelper 创建DoHHelper单例实例
func NewDoHHelper() *DoHHelper {
	dohHelperOnce.Do(func() {
		dohHelperInstance = &DoHHelper{
			cache: make(map[string]string),
			origLookupHost: net.DefaultResolver.LookupHost,
		}
		dohHelperInstance.init()
	})
	return dohHelperInstance
}

// init 初始化DoHHelper
func (dh *DoHHelper) init() {
	/*
		初始化DoHHelper
	*/
	settings := config.GetConfig()
	dh.EnableDoH(settings.DohEnable)
}

// EnableDoH 启用或禁用DoH解析
func (dh *DoHHelper) EnableDoH(enable bool) {
	/*
		启用或禁用DoH解析
	*/
	if enable {
		// 替换net.LookupHost方法
		net.DefaultResolver.LookupHost = dh.patchedLookupHost
	} else {
		// 恢复原始的net.LookupHost方法
		net.DefaultResolver.LookupHost = dh.origLookupHost
	}
}

// patchedLookupHost net.LookupHost的补丁版�?func (dh *DoHHelper) patchedLookupHost(ctx context.Context, network, host string) (addrs []string, err error) {
	/*
		net.LookupHost的补丁版�?	*/
	settings := config.GetConfig()
	
	// 检查主机是否在DoH域名列表�?	dohDomains := strings.Split(settings.DohDomains, ",")
	shouldUseDoH := false
	for _, domain := range dohDomains {
		if strings.TrimSpace(domain) == host {
			shouldUseDoH = true
			break
		}
	}
	
	if !shouldUseDoH {
		return dh.origLookupHost(ctx, network, host)
	}
	
	// 检查主机是否已解析
	dh.cacheLock.RLock()
	ip := dh.cache[host]
	dh.cacheLock.RUnlock()
	
	if ip != "" {
		logger.GetLoggerManager().Info("已解�?[%s] �?[%s] (缓存)", host, ip)
		return []string{ip}, nil
	}
	
	// 使用DoH解析主机
	dohResolvers := strings.Split(settings.DohResolvers, ",")
	resultChan := make(chan string, len(dohResolvers))
	var wg sync.WaitGroup
	
	for _, resolver := range dohResolvers {
		resolver = strings.TrimSpace(resolver)
		if resolver == "" {
			continue
		}
		
		wg.Add(1)
		go func(resolver, host string) {
			defer wg.Done()
			ip := dh.dohQuery(resolver, host)
			if ip != "" {
				resultChan <- ip
			}
		}(resolver, host)
	}
	
	// 等待所有goroutine完成或有结果返回
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 获取第一个成功的解析结果
	ip = ""
	select {
	case ip = <-resultChan:
		if ip != "" {
			logger.GetLoggerManager().Info("已解�?[%s] �?[%s]", host, ip)
			dh.cacheLock.Lock()
			dh.cache[host] = ip
			dh.cacheLock.Unlock()
		}
	case <-time.After(5 * time.Second):
		// 超时处理
		logger.GetLoggerManager().Warn("DoH解析超时: %s", host)
	}
	
	if ip != "" {
		return []string{ip}, nil
	}
	
	// 如果DoH解析失败，使用原始解析方�?	return dh.origLookupHost(ctx, network, host)
}

// ClearCache 清空DoH解析缓存
func (dh *DoHHelper) ClearCache() {
	/*
		清空DoH解析缓存
	*/
	dh.cacheLock.Lock()
	defer dh.cacheLock.Unlock()
	dh.cache = make(map[string]string)
}

// dohQuery 使用给定的DoH解析器查询给定主机的IP地址
func (dh *DoHHelper) dohQuery(resolver string, host string) string {
	/*
		使用给定的DoH解析器查询给定主机的IP地址
	*/
	
	// 构造DNS查询消息（RFC 1035�?	header := make([]byte, 12)
	// ID: 0
	binary.BigEndian.PutUint16(header[0:2], 0)
	// FLAGS: 标准递归查询
	binary.BigEndian.PutUint16(header[2:4], 0x0100)
	// QDCOUNT: 1
	binary.BigEndian.PutUint16(header[4:6], 1)
	// ANCOUNT: 0
	binary.BigEndian.PutUint16(header[6:8], 0)
	// NSCOUNT: 0
	binary.BigEndian.PutUint16(header[8:10], 0)
	// ARCOUNT: 0
	binary.BigEndian.PutUint16(header[10:12], 0)
	
	// 构造问题部�?	var question []byte
	labels := strings.Split(host, ".")
	for _, label := range labels {
		question = append(question, byte(len(label)))
		question = append(question, []byte(label)...)
	}
	question = append(question, 0x00) // 结束�?	
	// QTYPE: A (1)
	question = append(question, 0x00, 0x01)
	// QCLASS: IN (1)
	question = append(question, 0x00, 0x01)
	
	message := append(header, question...)
	
	// 发送GET请求到DoH解析器（RFC 8484�?	b64message := base64.URLEncoding.EncodeToString(message)
	// 去掉填充字符
	b64message = strings.TrimRight(b64message, "=")
	
	url := fmt.Sprintf("https://%s/dns-query?dns=%s", resolver, b64message)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.GetLoggerManager().Error("创建DoH请求失败", zap.Error(err))
		return ""
	}
	
	req.Header.Set("Content-Type", "application/dns-message")
	
	logger.GetLoggerManager().Debug("DoH请求: %s", url)
	
	resp, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Error("解析�?%s)请求错误: %v", resolver, err)
		return ""
	}
	defer resp.Body.Close()
	
	logger.GetLoggerManager().Debug("解析�?%s)响应: %d", resolver, resp.StatusCode)
	
	if resp.StatusCode != 200 {
		return ""
	}
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLoggerManager().Error("读取解析�?%s)响应失败: %v", resolver, err)
		return ""
	}
	
	// 解析DNS响应消息（RFC 1035�?	// name（压缩）:2 + type:2 + class:2 + ttl:4 + rdlength:2 = 12字节
	firstRdataStart := len(header) + len(question) + 12
	// rdata（A记录�? 4字节
	firstRdataEnd := firstRdataStart + 4
	
	// 检查响应体长度
	if len(respBody) < firstRdataEnd {
		logger.GetLoggerManager().Error("解析�?%s)响应体长度不�?, resolver)
		return ""
	}
	
	// 将rdata转换为IP地址
	ip := net.IP(respBody[firstRdataStart:firstRdataEnd]).String()
	return ip
}

// DoHQueryJSON 使用给定的DoH解析器查询给定主机的IP地址（JSON格式�?func (dh *DoHHelper) DoHQueryJSON(resolver string, host string) string {
	/*
		使用给定的DoH解析器查询给定主机的IP地址（JSON格式�?	*/
	url := fmt.Sprintf("https://%s/dns-query?name=%s&type=A", resolver, host)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.GetLoggerManager().Error("创建DoH JSON请求失败", zap.Error(err))
		return ""
	}
	
	req.Header.Set("Accept", "application/dns-json")
	
	logger.GetLoggerManager().Debug("DoH请求: %s", url)
	
	resp, err := client.Do(req)
	if err != nil {
		logger.GetLoggerManager().Error("解析�?%s)请求错误: %v", resolver, err)
		return ""
	}
	defer resp.Body.Close()
	
	logger.GetLoggerManager().Debug("解析�?%s)响应: %d", resolver, resp.StatusCode)
	
	if resp.StatusCode != 200 {
		return ""
	}
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLoggerManager().Error("读取解析�?%s)响应失败: %v", resolver, err)
		return ""
	}
	
	logger.GetLoggerManager().Debug("<== body: %s", string(respBody))
	
	var dohResp DoHResponse
	err = json.Unmarshal(respBody, &dohResp)
	if err != nil {
		logger.GetLoggerManager().Error("解析解析�?%s)JSON响应失败: %v", resolver, err)
		return ""
	}
	
	if len(dohResp.Answer) > 0 {
		return dohResp.Answer[0].Data
	}
	
	return ""
}
