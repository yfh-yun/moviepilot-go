package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"moviepilot-go/internal/types"
)

// HTTPUtils HTTP工具�?type HTTPUtils struct {
	client *http.Client
}

// NewHTTPUtils 创建一个新�?HTTPUtils 实例
func NewHTTPUtils() *HTTPUtils {
	// 创建一个自定义的HTTP客户�?	transport := &http.Transport{
		// 跳过证书验证
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// 设置连接超时
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// 设置连接池相关参�?		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // 默认超时时间
	}

	return &HTTPUtils{
		client: client,
	}
}

// RequestUtils 全局HTTP工具实例
var RequestUtils = NewHTTPUtils()

// PostRes 发送POST请求并返回响应体和状态码
func (h *HTTPUtils) PostRes(url string, data []byte, headers map[string]string, proxy map[string]string, timeout int) (*types.HTTPResponse, error) {
	// 创建上下�?	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	// 设置请求�?	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 设置代理
	var client *http.Client
	if proxy != nil && (proxy["http"] != "" || proxy["https"] != "") {
		proxyURL := proxy["http"]
		if proxyURL == "" {
			proxyURL = proxy["https"]
		}
		
		proxyURI, err := url.Parse(proxyURL)
		if err == nil {
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURI),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
			
			client = &http.Client{
				Transport: transport,
				Timeout:   time.Duration(timeout) * time.Second,
			}
		}
	}
	
	if client == nil {
		client = h.client
		if timeout > 0 {
			client.Timeout = time.Duration(timeout) * time.Second
		}
	}

	// 发送请�?	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 构造返回结�?	httpResp := &types.HTTPResponse{
		Body:    body,
		Content: string(body),
		StatusCode: resp.StatusCode,
		Headers: make(map[string]string),
		Cookies: make([]*types.Cookie, 0),
	}

	// 复制响应�?	for key, values := range resp.Header {
		if len(values) > 0 {
			httpResp.Headers[key] = values[0]
		}
	}

	// 复制Cookie
	for _, cookie := range resp.Cookies() {
		httpResp.Cookies = append(httpResp.Cookies, &types.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	return httpResp, nil
}

// GetRes 发送GET请求并返回响应体和状态码
func (h *HTTPUtils) GetRes(url string, headers map[string]string, proxy map[string]string, timeout int) (*types.HTTPResponse, error) {
	// 创建上下�?	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求�?	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 设置代理
	var client *http.Client
	if proxy != nil && (proxy["http"] != "" || proxy["https"] != "") {
		proxyURL := proxy["http"]
		if proxyURL == "" {
			proxyURL = proxy["https"]
		}
		
		proxyURI, err := url.Parse(proxyURL)
		if err == nil {
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURI),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
			
			client = &http.Client{
				Transport: transport,
				Timeout:   time.Duration(timeout) * time.Second,
			}
		}
	}
	
	if client == nil {
		client = h.client
		if timeout > 0 {
			client.Timeout = time.Duration(timeout) * time.Second
		}
	}

	// 发送请�?	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 构造返回结�?	httpResp := &types.HTTPResponse{
		Body:    body,
		Content: string(body),
		StatusCode: resp.StatusCode,
		Headers: make(map[string]string),
		Cookies: make([]*types.Cookie, 0),
	}

	// 复制响应�?	for key, values := range resp.Header {
		if len(values) > 0 {
			httpResp.Headers[key] = values[0]
		}
	}

	// 复制Cookie
	for _, cookie := range resp.Cookies() {
		httpResp.Cookies = append(httpResp.Cookies, &types.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	return httpResp, nil
}

// Request 发送HTTP请求
func (h *HTTPUtils) Request(method, url string, params map[string]string, cookies []*types.Cookie, headers map[string]string, timeout time.Duration) (*types.HTTPResponse, error) {
	// 创建上下�?	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// 处理GET参数
	if method == "GET" && params != nil {
		query := url + "?"
		for key, value := range params {
			query += key + "=" + value + "&"
		}
		// 移除末尾�?
		query = strings.TrimSuffix(query, "&")
		url = query
	}

	// 创建请求
	var req *http.Request
	var err error

	if method == "POST" && params != nil {
		// POST请求参数处理
		data := make(map[string][]string)
		for key, value := range params {
			data[key] = []string{value}
		}
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(h.encodeFormData(data)))
		if req != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		// GET或其他请�?		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// 设置请求�?	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 设置Cookie
	if cookies != nil {
		for _, cookie := range cookies {
			req.AddCookie(&http.Cookie{
				Name:  cookie.Name,
				Value: cookie.Value,
			})
		}
	}

	// 发送请�?	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 构造返回结�?	httpResp := &types.HTTPResponse{
		Content: string(body),
		Status:  resp.StatusCode,
		Headers: make(map[string]string),
		Cookies: make([]*types.Cookie, 0),
	}

	// 复制响应�?	for key, values := range resp.Header {
		if len(values) > 0 {
			httpResp.Headers[key] = values[0]
		}
	}

	// 复制Cookie
	for _, cookie := range resp.Cookies() {
		httpResp.Cookies = append(httpResp.Cookies, &types.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	return httpResp, nil
}

// encodeFormData 编码表单数据
func (h *HTTPUtils) encodeFormData(data map[string][]string) string {
	var result []string
	for key, values := range data {
		for _, value := range values {
			result = append(result, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return strings.Join(result, "&")
}

// Post 发送POST请求
func (h *HTTPUtils) Post(url string, params map[string]string, cookies []*types.Cookie, headers map[string]string, timeout time.Duration) (*types.HTTPResponse, error) {
	return h.Request("POST", url, params, cookies, headers, timeout)
}

// Get 发送GET请求
func (h *HTTPUtils) Get(url string, params map[string]string, cookies []*types.Cookie, headers map[string]string, timeout time.Duration) (*types.HTTPResponse, error) {
	return h.Request("GET", url, params, cookies, headers, timeout)
}

// Session 会话工具�?type Session struct {
	client  *http.Client
	cookies []*http.Cookie
	headers map[string]string
}

// NewSession 创建一个新的会�?func NewSession() *Session {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return &Session{
		client:  client,
		cookies: make([]*http.Cookie, 0),
		headers: make(map[string]string),
	}
}

// Get 发送GET请求
func (s *Session) Get(url string, params map[string]string, headers map[string]string) (*types.HTTPResponse, error) {
	return s.Request("GET", url, params, headers)
}

// Post 发送POST请求
func (s *Session) Post(url string, params map[string]string, headers map[string]string) (*types.HTTPResponse, error) {
	return s.Request("POST", url, params, headers)
}

// Request 发送HTTP请求
func (s *Session) Request(method, url string, params map[string]string, headers map[string]string) (*types.HTTPResponse, error) {
	// 创建上下�?	ctx := context.Background()

	// 处理GET参数
	if method == "GET" && params != nil {
		query := url + "?"
		for key, value := range params {
			query += key + "=" + value + "&"
		}
		// 移除末尾�?
		query = strings.TrimSuffix(query, "&")
		url = query
	}

	// 创建请求
	var req *http.Request
	var err error

	if method == "POST" && params != nil {
		// POST请求参数处理
		data := make(map[string][]string)
		for key, value := range params {
			data[key] = []string{value}
		}
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(s.encodeFormData(data)))
		if req != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		// GET或其他请�?		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// 设置请求�?	for key, value := range s.headers {
		req.Header.Set(key, value)
	}
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 设置Cookie
	for _, cookie := range s.cookies {
		req.AddCookie(cookie)
	}

	// 发送请�?	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 更新Cookie
	s.cookies = resp.Cookies()

	// 构造返回结�?	httpResp := &types.HTTPResponse{
		Content: string(body),
		Status:  resp.StatusCode,
		Headers: make(map[string]string),
		Cookies: make([]*types.Cookie, 0),
	}

	// 复制响应�?	for key, values := range resp.Header {
		if len(values) > 0 {
			httpResp.Headers[key] = values[0]
		}
	}

	// 复制Cookie
	for _, cookie := range resp.Cookies() {
		httpResp.Cookies = append(httpResp.Cookies, &types.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	return httpResp, nil
}

// encodeFormData 编码表单数据
func (s *Session) encodeFormData(data map[string][]string) string {
	var result []string
	for key, values := range data {
		for _, value := range values {
			result = append(result, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return strings.Join(result, "&")
}

// UpdateHeaders 更新会话的请求头
func (s *Session) UpdateHeaders(headers map[string]string) {
	if headers != nil {
		for key, value := range headers {
			s.headers[key] = value
		}
	}
}

// GetCookies 获取会话中的Cookie
func (s *Session) GetCookies() []*types.Cookie {
	cookies := make([]*types.Cookie, len(s.cookies))
	for i, cookie := range s.cookies {
		cookies[i] = &types.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		}
	}
	return cookies
}

// CookieStrToDict 将Cookie字符串转换为字典
func (h *HTTPUtils) CookieStrToDict(cookieStr string) map[string]string {
	cookies := make(map[string]string)
	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			cookies[kv[0]] = kv[1]
		}
	}
	return cookies
}

// AddHeader 添加请求�?func (h *HTTPUtils) AddHeader(key, value string) {
	// 这个方法在Python版本中是为requests库添加默认请求头
	// 在Go中，我们通常在每次请求时设置请求�?	// 这里保留方法以保持接口一致�?}

// GetImage 获取图片内容
func (h *HTTPUtils) GetImage(url string, referer string, timeout time.Duration) ([]byte, error) {
	// 创建上下�?	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置Referer
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	// 发送请�?	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状�?	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
