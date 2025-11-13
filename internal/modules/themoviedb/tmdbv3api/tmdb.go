package tmdbv3api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"moviepilot-go/pkg/config"
	"moviepilot-go/internal/utils"
)

// TMDb TMDb API结构�?type TMDb struct {
	apiKey             string
	language           string
	sessionID          string
	proxies            map[string]string
	domain             string
	page               *int
	totalResults       *int
	totalPages         *int
	remaining          int
	reset              *int
	timeout            int
	waitOnRateLimit    bool
	debugEnabled       bool
	cacheEnabled       bool
	objCached          bool
	mutex              sync.Mutex
}

// HTTPResponse HTTP响应结构�?type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// TMDbException TMDb异常结构�?type TMDbException struct {
	Message string
}

// NewTMDb 创建TMDb实例
func NewTMDb(objCached bool, language *string) *TMDb {
	tmdb := &TMDb{
		apiKey:          config.Config.TMDB_API_KEY,
		language:        config.Config.TMDB_LOCALE,
		proxies:         make(map[string]string),
		domain:          config.Config.TMDB_API_DOMAIN,
		remaining:       40,
		timeout:         15,
		waitOnRateLimit: true,
		debugEnabled:    false,
		cacheEnabled:    true,
		objCached:       objCached,
	}
	
	// 设置默认语言
	if tmdb.language == "" {
		tmdb.language = "en-US"
	}
	
	// 设置代理
	if config.Config.PROXY != "" {
		tmdb.proxies["http"] = config.Config.PROXY
		tmdb.proxies["https"] = config.Config.PROXY
	}
	
	// 如果传入了language参数，则使用该参�?	if language != nil && *language != "" {
		tmdb.language = *language
	}
	
	return tmdb
}

// Page 获取当前页码
func (t *TMDb) Page() *int {
	return t.page
}

// TotalResults 获取总结果数
func (t *TMDb) TotalResults() *int {
	return t.totalResults
}

// TotalPages 获取总页�?func (t *TMDb) TotalPages() *int {
	return t.totalPages
}

// APIKey 获取API密钥
func (t *TMDb) APIKey() string {
	return t.apiKey
}

// Domain 获取域名
func (t *TMDb) Domain() string {
	return t.domain
}

// Proxies 获取代理设置
func (t *TMDb) Proxies() map[string]string {
	return t.proxies
}

// SetProxies 设置代理
func (t *TMDb) SetProxies(proxies map[string]string) {
	t.proxies = proxies
}

// SetAPIKey 设置API密钥
func (t *TMDb) SetAPIKey(apiKey string) {
	t.apiKey = apiKey
}

// SetDomain 设置域名
func (t *TMDb) SetDomain(domain string) {
	t.domain = domain
}

// Language 获取语言
func (t *TMDb) Language() string {
	return t.language
}

// SetLanguage 设置语言
func (t *TMDb) SetLanguage(language string) {
	t.language = language
}

// HasSession 检查是否有会话
func (t *TMDb) HasSession() bool {
	return t.sessionID != ""
}

// SessionID 获取会话ID
func (t *TMDb) SessionID() (string, error) {
	if !t.HasSession() {
		return "", &TMDbException{"Must Authenticate to create a session run Authentication(username, password)"}
	}
	return t.sessionID, nil
}

// SetSessionID 设置会话ID
func (t *TMDb) SetSessionID(sessionID string) {
	t.sessionID = sessionID
}

// WaitOnRateLimit 获取是否等待频率限制
func (t *TMDb) WaitOnRateLimit() bool {
	return t.waitOnRateLimit
}

// SetWaitOnRateLimit 设置是否等待频率限制
func (t *TMDb) SetWaitOnRateLimit(waitOnRateLimit bool) {
	t.waitOnRateLimit = waitOnRateLimit
}

// Debug 获取调试模式
func (t *TMDb) Debug() bool {
	return t.debugEnabled
}

// SetDebug 设置调试模式
func (t *TMDb) SetDebug(debug bool) {
	t.debugEnabled = debug
}

// Cache 获取缓存模式
func (t *TMDb) Cache() bool {
	return t.cacheEnabled
}

// SetCache 设置缓存模式
func (t *TMDb) SetCache(cache bool) {
	t.cacheEnabled = cache
}

// CachedRequest 缓存请求
func (t *TMDb) CachedRequest(method, url string, data map[string]string, jsonData map[string]interface{}) (*HTTPResponse, error) {
	// 简化实现，实际应该实现缓存逻辑
	return t.Request(method, url, data, jsonData)
}

// Request 发起HTTP请求
func (t *TMDb) Request(method, url string, data map[string]string, jsonData map[string]interface{}) (*HTTPResponse, error) {
	var response *utils.HTTPResponse
	var err error
	
	if method == "GET" {
		response, err = utils.RequestUtils.GetRes(url, data, t.proxies, t.timeout)
	} else {
		// 将jsonData转换为字�?		var jsonBytes []byte
		if jsonData != nil {
			jsonBytes, err = json.Marshal(jsonData)
			if err != nil {
				return nil, &TMDbException{fmt.Sprintf("JSON序列化失�? %v", err)}
			}
		}
		headers := map[string]string{"Content-Type": "application/json"}
		response, err = utils.RequestUtils.PostRes(url, jsonBytes, headers, t.proxies, t.timeout)
	}
	
	if err != nil {
		return nil, &TMDbException{fmt.Sprintf("无法连接TheMovieDb，请检查网络连接！错误: %v", err)}
	}
	
	if response == nil {
		return nil, &TMDbException{"无法连接TheMovieDb，请检查网络连接！"}
	}
	
	return &HTTPResponse{
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
		Body:       response.Body,
	}, nil
}

// CacheClear 清除缓存
func (t *TMDb) CacheClear() {
	// 简化实现，实际应该清除缓存
}

// validateAPIKey 验证API密钥
func (t *TMDb) validateAPIKey() error {
	if t.apiKey == "" {
		return &TMDbException{"TheMovieDb API Key 未设置！"}
	}
	return nil
}

// buildURL 构建URL
func (t *TMDb) buildURL(action string, params string) string {
	url := fmt.Sprintf("https://%s/3%s?api_key=%s", t.domain, action, t.apiKey)
	
	if params != "" {
		url += "&" + params
	}
	
	url += "&language=" + t.language
	
	return url
}

// handleHeaders 处理响应�?func (t *TMDb) handleHeaders(headers map[string]string) {
	if remaining, exists := headers["X-RateLimit-Remaining"]; exists {
		if val, err := strconv.Atoi(remaining); err == nil {
			t.remaining = val
		}
	}
	
	if reset, exists := headers["X-RateLimit-Reset"]; exists {
		if val, err := strconv.Atoi(reset); err == nil {
			t.reset = &val
		}
	}
}

// handleRateLimit 处理频率限制
func (t *TMDb) handleRateLimit() int {
	if t.remaining < 1 && t.reset != nil {
		currentTime := int(time.Now().Unix())
		sleepTime := *t.reset - currentTime
		
		if t.waitOnRateLimit {
			// 日志记录应该使用项目中的logger，这里简化处�?			fmt.Printf("达到请求频率限制，休眠：%d �?..\n", sleepTime)
			return sleepTime
		} else {
			// 抛出异常
			return -1 // 用负数表示需要抛出异�?		}
	}
	return 0
}

// processJSONResponse 处理JSON响应
func (t *TMDb) processJSONResponse(jsonData map[string]interface{}, isAsync bool) {
	if page, exists := jsonData["page"]; exists {
		if pageFloat, ok := page.(float64); ok {
			pageInt := int(pageFloat)
			t.page = &pageInt
		}
	}
	
	if totalResults, exists := jsonData["total_results"]; exists {
		if totalResultsFloat, ok := totalResults.(float64); ok {
			totalResultsInt := int(totalResultsFloat)
			t.totalResults = &totalResultsInt
		}
	}
	
	if totalPages, exists := jsonData["total_pages"]; exists {
		if totalPagesFloat, ok := totalPages.(float64); ok {
			totalPagesInt := int(totalPagesFloat)
			t.totalPages = &totalPagesInt
		}
	}
	
	if t.debugEnabled {
		// 日志记录应该使用项目中的logger，这里简化处�?		fmt.Printf("JSON Response: %+v\n", jsonData)
	}
}

// handleErrors 处理错误
func (t *TMDb) handleErrors(jsonData map[string]interface{}) error {
	if errors, exists := jsonData["errors"]; exists {
		return &TMDbException{fmt.Sprintf("%v", errors)}
	}
	
	if success, exists := jsonData["success"]; exists {
		if successBool, ok := success.(bool); ok && !successBool {
			if statusMessage, exists := jsonData["status_message"]; exists {
				return &TMDbException{fmt.Sprintf("%v", statusMessage)}
			}
		}
	}
	
	return nil
}

// RequestObj 发起对象请求
func (t *TMDb) RequestObj(action, params, method string, data map[string]string, jsonData map[string]interface{}, key *string) (interface{}, error) {
	err := t.validateAPIKey()
	if err != nil {
		return nil, err
	}
	
	url := t.buildURL(action, params)
	
	var req *HTTPResponse
	if t.cacheEnabled && t.objCached && method != "POST" {
		req, err = t.CachedRequest(method, url, data, jsonData)
	} else {
		req, err = t.Request(method, url, data, jsonData)
	}
	
	if err != nil {
		return nil, err
	}
	
	if req == nil {
		return nil, nil
	}
	
	t.handleHeaders(req.Headers)
	
	rateLimitResult := t.handleRateLimit()
	if rateLimitResult < 0 {
		return nil, &TMDbException{"达到请求频率限制，请稍后再试�?}
	} else if rateLimitResult > 0 {
		// 日志记录应该使用项目中的logger，这里简化处�?		fmt.Printf("达到请求频率限制，将�?%d 秒后重试...\n", rateLimitResult)
		time.Sleep(time.Duration(rateLimitResult) * time.Second)
		return t.RequestObj(action, params, method, data, jsonData, key)
	}
	
	var jsonDataResp map[string]interface{}
	if err := json.Unmarshal(req.Body, &jsonDataResp); err != nil {
		return nil, &TMDbException{fmt.Sprintf("JSON解析失败: %v", err)}
	}
	
	t.processJSONResponse(jsonDataResp, false)
	
	if err := t.handleErrors(jsonDataResp); err != nil {
		return nil, err
	}
	
	if key != nil {
		return jsonDataResp[*key], nil
	}
	
	return jsonDataResp, nil
}

// Close 关闭连接
func (t *TMDb) Close() {
	// 在Go中HTTP连接会自动管理，无需特殊处理
}

// Error 实现error接口
func (e *TMDbException) Error() string {
	return e.Message
}
