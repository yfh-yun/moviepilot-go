package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	urlutil "moviepilot-go/pkg/utils/url"
)

// Client HTTP客户端封装
type Client struct {
	client  *http.Client
	timeout time.Duration
	headers map[string]string
	proxy   string
}

// ClientConfig HTTP客户端配置
type ClientConfig struct {
	Timeout    time.Duration
	Headers    map[string]string
	Proxy      string
	UserAgent  string
	MaxRetries int
}

// NewClient 创建新的HTTP客户端
func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	// 设置默认User-Agent
	if config.UserAgent == "" {
		config.UserAgent = "MoviePilot/1.0.0"
	}
	config.Headers["User-Agent"] = config.UserAgent

	client := &http.Client{
		Timeout: config.Timeout,
	}

	// 设置代理
	if config.Proxy != "" {
		if proxyURL, err := neturl.Parse(config.Proxy); err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	return &Client{
		client:  client,
		timeout: config.Timeout,
		headers: config.Headers,
		proxy:   config.Proxy,
	}
}

// Get 发送GET请求
func (c *Client) Get(ctx context.Context, url string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	return c.doRequest(req)
}

// Head 发送HEAD请求
func (c *Client) Head(ctx context.Context, url string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	return c.doRequest(req)
}

// GetImage 获取图片内容
func (c *Client) GetImage(ctx context.Context, imageURL string, ifNoneMatch string) (*ImageResponse, error) {
	// 替换URL中的占位符
	imageURL = urlutil.ReplaceImageURLPlaceholders(imageURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建图片请求失败: %w", err)
	}

	// 设置If-None-Match头
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}

	// 设置Accept头
	req.Header.Set("Accept", "image/*")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求图片失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取图片内容失败: %w", err)
	}

	// 构建响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[strings.ToLower(key)] = values[0]
		}
	}

	// 获取Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = urlutil.GetImageMimeType(imageURL)
	}

	return &ImageResponse{
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Headers:     headers,
		Data:        body,
	}, nil
}

// TestNetwork 测试网络连通性
func (c *Client) TestNetwork(ctx context.Context, testURL, expectedContent string) (*NetworkTestResult, error) {
	startTime := time.Now()

	resp, err := c.Get(ctx, testURL)
	if err != nil {
		return &NetworkTestResult{
			Success:      false,
			Error:        err.Error(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	responseTime := time.Since(startTime).Milliseconds()
	result := &NetworkTestResult{
		Success:       true,
		StatusCode:    resp.StatusCode,
		ResponseTime:  responseTime,
		ContentType:   resp.Headers["content-type"],
		ContentLength: len(resp.Data),
	}

	// 检查期望内容
	if expectedContent != "" {
		content := string(resp.Data)
		if strings.Contains(content, expectedContent) {
			result.ContentMatch = true
		} else {
			result.ContentMatch = false
			result.Error = fmt.Sprintf("响应内容不包含期望的字符串: %s", expectedContent)
		}
	}

	return result, nil
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(req *http.Request) (*Response, error) {
	// 设置请求头
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 构建响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[strings.ToLower(key)] = values[0]
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Data:       body,
	}, nil
}

// Response HTTP响应
type Response struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Data       []byte            `json:"data"`
}

// ImageResponse 图片响应
type ImageResponse struct {
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	Data        []byte            `json:"data"`
}

// NetworkTestResult 网络测试结果
type NetworkTestResult struct {
	Success       bool   `json:"success"`
	StatusCode    int    `json:"status_code,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	ContentLength int    `json:"content_length,omitempty"`
	ContentMatch  bool   `json:"content_match,omitempty"`
	ResponseTime  int64  `json:"response_time"`
	Error         string `json:"error,omitempty"`
}

// GetGithubAPI 获取Github API响应
func GetGithubAPI(ctx context.Context, apiURL string) (*Response, error) {
	config := ClientConfig{
		Timeout:   10 * time.Second,
		UserAgent: "MoviePilot-GitHub-Client/1.0.0",
	}

	client := NewClient(config)

	// 替换代理占位符
	apiURL = urlutil.ReplaceProxyPlaceholders(apiURL)

	return client.Get(ctx, apiURL)
}

// DefaultClient 默认HTTP客户端
func DefaultClient() *Client {
	return NewClient(ClientConfig{})
}

// Cookie 解析后的单个 cookie 键值
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CookieParse 解析 cookie 字符串为字典或数组
// 对应 Python cookie_parse 函数
func CookieParse(cookiesStr string, array bool) interface{} {
	if cookiesStr == "" {
		if array {
			return []Cookie{}
		}
		return map[string]string{}
	}

	cookieMap := make(map[string]string)
	parts := strings.Split(cookiesStr, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if name != "" {
			cookieMap[name] = value
		}
	}

	if array {
		res := make([]Cookie, 0, len(cookieMap))
		for k, v := range cookieMap {
			res = append(res, Cookie{Name: k, Value: v})
		}
		return res
	}

	return cookieMap
}

// ParseCookieString 解析 "k1=v1; k2=v2" 形式的 cookie 字符串为 map
// 对应 Python cookie_parse(cookies_str, array=False)
// 保留此函数以保持向后兼容性
func ParseCookieString(cookiesStr string) map[string]string {
	result := make(map[string]string)
	if cookiesStr == "" {
		return result
	}

	parts := strings.Split(cookiesStr, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if name != "" {
			result[name] = value
		}
	}
	return result
}

// CookiesToArray 将 cookie map 转换为 Cookie 数组
// 对应 Python cookie_parse(..., array=True)
// 保留此函数以保持向后兼容性
func CookiesToArray(cookieMap map[string]string) []Cookie {
	if len(cookieMap) == 0 {
		return nil
	}
	res := make([]Cookie, 0, len(cookieMap))
	for k, v := range cookieMap {
		res = append(res, Cookie{Name: k, Value: v})
	}
	return res
}

// ParseCacheControl 解析 Cache-Control 头，返回缓存指令和 max-age（如果存在）
func ParseCacheControl(header string) (cacheDirective string, maxAge *int) {
	if header == "" {
		return "", nil
	}

	directives := strings.Split(header, ",")
	for _, d := range directives {
		directive := strings.TrimSpace(d)
		if strings.HasPrefix(strings.ToLower(directive), "max-age") {
			parts := strings.SplitN(directive, "=", 2)
			if len(parts) == 2 {
				if v, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					maxAge = &v
				}
			}
		} else if directive == "no-cache" || directive == "private" || directive == "public" ||
			directive == "no-store" || directive == "must-revalidate" {
			cacheDirective = directive
		}
	}

	return cacheDirective, maxAge
}

// GenerateCacheHeaders 生成 ETag 和 Cache-Control 头
func GenerateCacheHeaders(etag string, cacheControl string, maxAge *int) map[string]string {
	headers := make(map[string]string)
	if etag != "" {
		headers["ETag"] = etag
	}

	if cacheControl != "" && maxAge != nil {
		headers["Cache-Control"] = fmt.Sprintf("%s, max-age=%d", cacheControl, *maxAge)
	} else if cacheControl != "" {
		headers["Cache-Control"] = cacheControl
	} else if maxAge != nil {
		headers["Cache-Control"] = fmt.Sprintf("max-age=%d", *maxAge)
	}

	return headers
}

// DetectEncodingFromHTML 根据 HTML 响应内容尝试探测编码，目前实现简化为只识别 utf-8 场景
// 与 Python RequestUtils.detect_encoding_from_html_response 语义类似，但未引入 chardet
func DetectEncodingFromHTML(resp *http.Response) string {
	if resp == nil {
		return "utf-8"
	}

	// 1. 从 Content-Type 中检查 charset
	contentType := resp.Header.Get("Content-Type")
	if regexp.MustCompile(`(?i)charset=["']?utf-8["']?`).MatchString(contentType) {
		return "utf-8"
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return "utf-8"
	}

	// 2. 检查 BOM
	if len(body) >= 3 && body[0] == 0xef && body[1] == 0xbb && body[2] == 0xbf {
		return "utf-8"
	}

	// 3. 检查 meta charset
	if regexp.MustCompile(`(?i)charset=["']?utf-8["']?`).Match(body) {
		return "utf-8"
	}

	// 兜底使用 utf-8
	return "utf-8"
}

// GetDecodedHTMLContent 获取解码后的 HTML 文本内容
func GetDecodedHTMLContent(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	if len(body) == 0 {
		return ""
	}

	encoding := DetectEncodingFromHTML(resp)
	// 当前仅处理 utf-8，其他编码直接按默认方式转换为字符串
	_ = encoding
	return string(body)
}

// RequestOptions 对应 Python RequestUtils 的构造参数子集
type RequestOptions struct {
	Headers   map[string]string
	Cookies   map[string]string
	Proxy     string
	Timeout   time.Duration
	UserAgent string
}

// RequestUtils 提供与 Python RequestUtils 语义接近的同步 HTTP 工具
type RequestUtils struct {
	client  *Client
	headers map[string]string
	cookies map[string]string
}

// NewRequestUtils 创建 RequestUtils 实例
func NewRequestUtils(opts RequestOptions) *RequestUtils {
	headers := make(map[string]string)
	for k, v := range opts.Headers {
		if v != "" {
			headers[k] = v
		}
	}

	config := ClientConfig{
		Timeout:   opts.Timeout,
		Headers:   headers,
		Proxy:     opts.Proxy,
		UserAgent: opts.UserAgent,
	}

	client := NewClient(config)

	return &RequestUtils{
		client:  client,
		headers: headers,
		cookies: opts.Cookies,
	}
}

// buildCookieHeader 构建 Cookie 头字符串
func (r *RequestUtils) buildCookieHeader() string {
	if len(r.cookies) == 0 {
		return ""
	}
	parts := make([]string, 0, len(r.cookies))
	for k, v := range r.cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "; ")
}

// Request 发起 HTTP 请求，返回通用 Response
func (r *RequestUtils) Request(ctx context.Context, method, rawURL string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 复用底层 Client 的 headers
	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	// 设置 Cookie 头
	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	return r.client.doRequest(req)
}

// Get 发送 GET 请求并返回字符串内容
func (r *RequestUtils) Get(ctx context.Context, rawURL string) (string, error) {
	resp, err := r.Request(ctx, http.MethodGet, rawURL)
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Data) == 0 {
		return "", nil
	}
	return string(resp.Data), nil
}

// GetJSON 发送 GET 请求并解析 JSON 响应为泛型 map
func (r *RequestUtils) GetJSON(ctx context.Context, rawURL string) (map[string]any, error) {
	resp, err := r.Request(ctx, http.MethodGet, rawURL)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Data) == 0 {
		return nil, nil
	}
	var out map[string]any
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}
	return out, nil
}

// PostJSON 发送 POST 请求并解析 JSON 响应
func (r *RequestUtils) PostJSON(ctx context.Context, rawURL string, body any) (map[string]any, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	resp, err := r.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Data) == 0 {
		return nil, nil
	}
	var out map[string]any
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}
	return out, nil
}

// Post 发送 POST 请求
func (r *RequestUtils) Post(ctx context.Context, rawURL string, data any, jsonData map[string]any) (*Response, error) {
	var body io.Reader
	var contentType string

	if jsonData != nil {
		bodyBytes, err := json.Marshal(jsonData)
		if err != nil {
			return nil, fmt.Errorf("marshal JSON failed: %w", err)
		}
		body = strings.NewReader(string(bodyBytes))
		contentType = "application/json"
	} else if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
			contentType = "application/x-www-form-urlencoded"
		case []byte:
			body = bytes.NewReader(d)
			contentType = "application/octet-stream"
		case io.Reader:
			body = d
			contentType = "application/octet-stream"
		default:
			return nil, fmt.Errorf("unsupported data type: %T", data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	return r.client.doRequest(req)
}

// Put 发送 PUT 请求
func (r *RequestUtils) Put(ctx context.Context, rawURL string, data any) (*Response, error) {
	var body io.Reader
	var contentType string

	if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
			contentType = "application/x-www-form-urlencoded"
		case []byte:
			body = bytes.NewReader(d)
			contentType = "application/octet-stream"
		case io.Reader:
			body = d
			contentType = "application/octet-stream"
		default:
			return nil, fmt.Errorf("unsupported data type: %T", data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	return r.client.doRequest(req)
}

// Delete 发送 DELETE 请求
func (r *RequestUtils) Delete(ctx context.Context, rawURL string, data any) (*Response, error) {
	var body io.Reader
	var contentType string

	if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
			contentType = "application/x-www-form-urlencoded"
		case []byte:
			body = bytes.NewReader(d)
			contentType = "application/octet-stream"
		case io.Reader:
			body = d
			contentType = "application/octet-stream"
		default:
			return nil, fmt.Errorf("unsupported data type: %T", data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	return r.client.doRequest(req)
}

// GetRes 发送 GET 请求并返回响应对象，支持更多选项
func (r *RequestUtils) GetRes(ctx context.Context, rawURL string, params map[string]string, data any, jsonData map[string]any, allowRedirects bool, raiseException bool) (*Response, error) {
	// 构建请求URL，添加查询参数
	url := rawURL
	if params != nil {
		parsedURL, err := neturl.Parse(rawURL)
		if err == nil {
			q := parsedURL.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			parsedURL.RawQuery = q.Encode()
			url = parsedURL.String()
		}
	}

	// 构建请求体
	var body io.Reader
	var contentType string

	if jsonData != nil {
		bodyBytes, err := json.Marshal(jsonData)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("marshal JSON failed: %w", err)
			}
			return nil, nil
		}
		body = strings.NewReader(string(bodyBytes))
		contentType = "application/json"
	} else if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
			contentType = "application/x-www-form-urlencoded"
		case []byte:
			body = bytes.NewReader(d)
			contentType = "application/octet-stream"
		case io.Reader:
			body = d
			contentType = "application/octet-stream"
		default:
			if raiseException {
				return nil, fmt.Errorf("unsupported data type: %T", data)
			}
			return nil, nil
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, body)
	if err != nil {
		if raiseException {
			return nil, fmt.Errorf("create request failed: %w", err)
		}
		return nil, nil
	}

	// 设置请求头
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	// 设置Cookie头
	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	// 设置重定向策略
	client := r.client.client
	if !allowRedirects {
		// 创建临时客户端，禁用重定向
		tmpClient := *client
		tmpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		resp, err := tmpClient.Do(req)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			return nil, nil
		}
		defer resp.Body.Close()

		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("read response failed: %w", err)
			}
			return nil, nil
		}

		// 构建响应头
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[strings.ToLower(key)] = values[0]
			}
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    headers,
			Data:       respBody,
		}, nil
	}

	// 执行请求
	resp, err := r.client.doRequest(req)
	if err != nil {
		if raiseException {
			return nil, err
		}
		return nil, nil
	}

	return resp, nil
}

// PostRes 发送 POST 请求并返回响应对象，支持更多选项
func (r *RequestUtils) PostRes(ctx context.Context, rawURL string, data any, params map[string]string, allowRedirects bool, files any, jsonData map[string]any, raiseException bool) (*Response, error) {
	// 构建请求URL，添加查询参数
	url := rawURL
	if params != nil {
		parsedURL, err := neturl.Parse(rawURL)
		if err == nil {
			q := parsedURL.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			parsedURL.RawQuery = q.Encode()
			url = parsedURL.String()
		}
	}

	// 构建请求体
	var body io.Reader
	var contentType string

	if jsonData != nil {
		bodyBytes, err := json.Marshal(jsonData)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("marshal JSON failed: %w", err)
			}
			return nil, nil
		}
		body = strings.NewReader(string(bodyBytes))
		contentType = "application/json"
	} else if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
			contentType = "application/x-www-form-urlencoded"
		case []byte:
			body = bytes.NewReader(d)
			contentType = "application/octet-stream"
		case io.Reader:
			body = d
			contentType = "application/octet-stream"
		default:
			if raiseException {
				return nil, fmt.Errorf("unsupported data type: %T", data)
			}
			return nil, nil
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		if raiseException {
			return nil, fmt.Errorf("create request failed: %w", err)
		}
		return nil, nil
	}

	// 设置请求头
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	// 设置Cookie头
	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	// 设置重定向策略
	client := r.client.client
	if !allowRedirects {
		// 创建临时客户端，禁用重定向
		tmpClient := *client
		tmpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		resp, err := tmpClient.Do(req)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			return nil, nil
		}
		defer resp.Body.Close()

		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("read response failed: %w", err)
			}
			return nil, nil
		}

		// 构建响应头
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[strings.ToLower(key)] = values[0]
			}
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    headers,
			Data:       respBody,
		}, nil
	}

	// 执行请求
	resp, err := r.client.doRequest(req)
	if err != nil {
		if raiseException {
			return nil, err
		}
		return nil, nil
	}

	return resp, nil
}

// PutRes 发送 PUT 请求并返回响应对象，支持更多选项
func (r *RequestUtils) PutRes(ctx context.Context, rawURL string, data any, params map[string]string, allowRedirects bool, files any, jsonData map[string]any, raiseException bool) (*Response, error) {
	// 构建请求URL，添加查询参数
	url := rawURL
	if params != nil {
		parsedURL, err := neturl.Parse(rawURL)
		if err == nil {
			q := parsedURL.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			parsedURL.RawQuery = q.Encode()
			url = parsedURL.String()
		}
	}

	// 构建请求体
	var body io.Reader
	var contentType string

	if jsonData != nil {
		bodyBytes, err := json.Marshal(jsonData)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("marshal JSON failed: %w", err)
			}
			return nil, nil
		}
		body = strings.NewReader(string(bodyBytes))
		contentType = "application/json"
	} else if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
			contentType = "application/x-www-form-urlencoded"
		case []byte:
			body = bytes.NewReader(d)
			contentType = "application/octet-stream"
		case io.Reader:
			body = d
			contentType = "application/octet-stream"
		default:
			if raiseException {
				return nil, fmt.Errorf("unsupported data type: %T", data)
			}
			return nil, nil
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		if raiseException {
			return nil, fmt.Errorf("create request failed: %w", err)
		}
		return nil, nil
	}

	// 设置请求头
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	// 设置Cookie头
	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	// 设置重定向策略
	client := r.client.client
	if !allowRedirects {
		// 创建临时客户端，禁用重定向
		tmpClient := *client
		tmpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		resp, err := tmpClient.Do(req)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			return nil, nil
		}
		defer resp.Body.Close()

		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("read response failed: %w", err)
			}
			return nil, nil
		}

		// 构建响应头
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[strings.ToLower(key)] = values[0]
			}
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    headers,
			Data:       respBody,
		}, nil
	}

	// 执行请求
	resp, err := r.client.doRequest(req)
	if err != nil {
		if raiseException {
			return nil, err
		}
		return nil, nil
	}

	return resp, nil
}

// DeleteRes 发送 DELETE 请求并返回响应对象，支持更多选项
func (r *RequestUtils) DeleteRes(ctx context.Context, rawURL string, data any, params map[string]string, allowRedirects bool, raiseException bool) (*Response, error) {
	// 构建请求URL，添加查询参数
	url := rawURL
	if params != nil {
		parsedURL, err := neturl.Parse(rawURL)
		if err == nil {
			q := parsedURL.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			parsedURL.RawQuery = q.Encode()
			url = parsedURL.String()
		}
	}

	// 构建请求体
	var body io.Reader
	var contentType string

	if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
			contentType = "application/x-www-form-urlencoded"
		case []byte:
			body = bytes.NewReader(d)
			contentType = "application/octet-stream"
		case io.Reader:
			body = d
			contentType = "application/octet-stream"
		default:
			if raiseException {
				return nil, fmt.Errorf("unsupported data type: %T", data)
			}
			return nil, nil
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, body)
	if err != nil {
		if raiseException {
			return nil, fmt.Errorf("create request failed: %w", err)
		}
		return nil, nil
	}

	// 设置请求头
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range r.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	// 设置Cookie头
	if cookieHeader := r.buildCookieHeader(); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	// 设置重定向策略
	client := r.client.client
	if !allowRedirects {
		// 创建临时客户端，禁用重定向
		tmpClient := *client
		tmpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		resp, err := tmpClient.Do(req)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			return nil, nil
		}
		defer resp.Body.Close()

		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if raiseException {
				return nil, fmt.Errorf("read response failed: %w", err)
			}
			return nil, nil
		}

		// 构建响应头
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[strings.ToLower(key)] = values[0]
			}
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    headers,
			Data:       respBody,
		}, nil
	}

	// 执行请求
	resp, err := r.client.doRequest(req)
	if err != nil {
		if raiseException {
			return nil, err
		}
		return nil, nil
	}

	return resp, nil
}
