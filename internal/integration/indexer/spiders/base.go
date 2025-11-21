// Package spiders 站点爬虫基础类
package spiders

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"moviepilot-go/internal/integration/indexer"
	"moviepilot-go/pkg/httpclient"
)

// BaseSpider 基础爬虫
type BaseSpider struct {
	name       string
	domain     string
	baseURL    string
	httpClient *httpclient.Client
	cookies    map[string]string
	headers    map[string]string
}

// NewBaseSpider 创建基础爬虫
func NewBaseSpider(name, domain string) *BaseSpider {
	return &BaseSpider{
		name:    name,
		domain:  domain,
		baseURL: fmt.Sprintf("https://%s", domain),
		httpClient: httpclient.NewClient(httpclient.Options{
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
			},
		}),
		cookies: make(map[string]string),
		headers: make(map[string]string),
	}
}

// GetName 获取爬虫名称
func (bs *BaseSpider) GetName() string {
	return bs.name
}

// GetDomain 获取域名
func (bs *BaseSpider) GetDomain() string {
	return bs.domain
}

// SetCookies 设置Cookies
func (bs *BaseSpider) SetCookies(cookies map[string]string) {
	bs.cookies = cookies
	bs.updateHttpClient()
}

// SetHeaders 设置Headers
func (bs *BaseSpider) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		bs.headers[k] = v
	}
	bs.updateHttpClient()
}

// updateHttpClient 更新HTTP客户端
func (bs *BaseSpider) updateHttpClient() {
	// 构建Cookie字符串
	var cookieStrings []string
	for k, v := range bs.cookies {
		cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", k, v))
	}

	// 更新Headers
	if len(cookieStrings) > 0 {
		bs.headers["Cookie"] = strings.Join(cookieStrings, "; ")
	}

	bs.httpClient.SetHeaders(bs.headers)
}

// Get 获取页面
func (bs *BaseSpider) Get(ctx context.Context, path string, params map[string]string) (*http.Response, error) {
	url := bs.baseURL + path
	if len(params) > 0 {
		url += "?" + bs.buildQuery(params)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	return bs.httpClient.Do(req)
}

// Post 提交数据
func (bs *BaseSpider) Post(ctx context.Context, path string, data map[string]string) (*http.Response, error) {
	url := bs.baseURL + path

	formData := url.Values{}
	for k, v := range data {
		formData.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return bs.httpClient.Do(req)
}

// buildQuery 构建查询字符串
func (bs *BaseSpider) buildQuery(params map[string]string) string {
	var pairs []string
	for k, v := range params {
		pairs = append(pairs, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
	}
	return strings.Join(pairs, "&")
}

// Login 登录
func (bs *BaseSpider) Login(ctx context.Context, username, password string) error {
	// 基础登录实现，子类可以重写
	loginData := map[string]string{
		"username": username,
		"password": password,
		"submit":   "登录",
	}

	resp, err := bs.Post(ctx, "/login.php", loginData)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// 提取cookies
	for _, cookie := range resp.Cookies() {
		bs.cookies[cookie.Name] = cookie.Value
	}
	bs.updateHttpClient()

	return nil
}

// IsLoggedIn 检查是否已登录
func (bs *BaseSpider) IsLoggedIn(ctx context.Context) bool {
	resp, err := bs.Get(ctx, "/index.php", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查响应中是否包含登录页面特征
	return resp.StatusCode == http.StatusOK && !strings.Contains(resp.Request.URL.String(), "login")
}

// Search 搜索种子 - 基础实现
func (bs *BaseSpider) Search(ctx context.Context, keyword, mediaType string) ([]*indexer.TorrentInfo, error) {
	searchData := map[string]string{
		"search":    keyword,
		"cat":       mediaType,
		"search_in": "title",
		"dead":      "0",
	}

	resp, err := bs.Post(ctx, "/torrents.php", searchData)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	// 这里需要使用对应的解析器来解析结果
	// 基础实现返回空结果
	return []*indexer.TorrentInfo{}, nil
}

// GetTorrentDetail 获取种子详情 - 基础实现
func (bs *BaseSpider) GetTorrentDetail(ctx context.Context, id string) (*indexer.TorrentDetail, error) {
	path := fmt.Sprintf("/details.php?id=%s", id)
	resp, err := bs.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get torrent detail failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get torrent detail failed with status: %d", resp.StatusCode)
	}

	// 这里需要使用对应的解析器来解析结果
	// 基础实现返回空结果
	return &indexer.TorrentDetail{}, nil
}

// GetUserInfo 获取用户信息 - 基础实现
func (bs *BaseSpider) GetUserInfo(ctx context.Context) (*indexer.UserInfo, error) {
	resp, err := bs.Get(ctx, "/user.php", nil)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user info failed with status: %d", resp.StatusCode)
	}

	// 这里需要使用对应的解析器来解析结果
	// 基础实现返回空结果
	return &indexer.UserInfo{}, nil
}

// Download 下载种子 - 基础实现
func (bs *BaseSpider) Download(ctx context.Context, id string) ([]byte, error) {
	path := fmt.Sprintf("/download.php?id=%s", id)
	resp, err := bs.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("download torrent failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download torrent failed with status: %d", resp.StatusCode)
	}

	// 读取响应体
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read torrent content failed: %w", err)
	}

	return content, nil
}

// GetPageContent 获取页面内容
func (bs *BaseSpider) GetPageContent(ctx context.Context, path string, params map[string]string) (string, error) {
	resp, err := bs.Get(ctx, path, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// PostPageContent 提交并获取页面内容
func (bs *BaseSpider) PostPageContent(ctx context.Context, path string, data map[string]string) (string, error) {
	resp, err := bs.Post(ctx, path, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// ExtractFormToken 提取表单token
func (bs *BaseSpider) ExtractFormToken(html string) string {
	// 提取CSRF token或其他表单token
	tokenRegex := regexp.MustCompile(`<input[^>]*name=["']token["'][^>]*value=["']([^"']+)["']`)
	matches := tokenRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// ExtractCaptcha 提取验证码
func (bs *BaseSpider) ExtractCaptcha(html string) string {
	// 提取验证码图片URL
	captchaRegex := regexp.MustCompile(`<img[^>]*src=["']([^"']*captcha[^"']*)["']`)
	matches := captchaRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// SolveCaptcha 解决验证码 - 基础实现
func (bs *BaseSpider) SolveCaptcha(ctx context.Context, captchaURL string) (string, error) {
	// 基础实现不支持验证码识别
	return "", fmt.Errorf("captcha solving not implemented")
}

// RateLimit 速率限制
func (bs *BaseSpider) RateLimit(ctx context.Context) {
	// 基础速率限制，子类可以重写
	select {
	case <-ctx.Done():
		return
	case <-time.After(1 * time.Second):
		return
	}
}

// Retry 重试机制
func (bs *BaseSpider) Retry(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			
			// 如果是上下文取消，直接返回
			if ctx.Err() != nil {
				return ctx.Err()
			}
			
			// 等待重试
			backoff := time.Duration(i+1) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				continue
			}
		} else {
			return nil
		}
	}
	
	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ValidateResponse 验证响应
func (bs *BaseSpider) ValidateResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("access forbidden")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found")
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}
	return nil
}

// GetBaseURL 获取基础URL
func (bs *BaseSpider) GetBaseURL() string {
	return bs.baseURL
}

// SetBaseURL 设置基础URL
func (bs *BaseSpider) SetBaseURL(baseURL string) {
	bs.baseURL = baseURL
}

// GetCookies 获取Cookies
func (bs *BaseSpider) GetCookies() map[string]string {
	result := make(map[string]string)
	for k, v := range bs.cookies {
		result[k] = v
	}
	return result
}

// GetHeaders 获取Headers
func (bs *BaseSpider) GetHeaders() map[string]string {
	result := make(map[string]string)
	for k, v := range bs.headers {
		result[k] = v
	}
	return result
}