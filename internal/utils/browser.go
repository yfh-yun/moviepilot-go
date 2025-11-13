package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"go.uber.org/zap"
)

// BrowserHelper 浏览器助手结构体
type BrowserHelper struct {
	browserType string
	client      *resty.Client
}

// FlareSolverrResponse FlareSolverr响应结构
type FlareSolverrResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Solution struct {
		Cookies    []FlareSolverrCookie `json:"cookies"`
		UserAgent  string               `json:"userAgent"`
		URL        string               `json:"url"`
		Headers    map[string]string    `json:"headers"`
		Response   string               `json:"response"`
	} `json:"solution"`
}

// FlareSolverrCookie FlareSolverr Cookie结构
type FlareSolverrCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires"`
	Size     int    `json:"size"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
	Session  bool   `json:"session"`
	SameSite string `json:"sameSite"`
}

// FlareSolverrRequest FlareSolverr请求结构
type FlareSolverrRequest struct {
	Cmd       string                 `json:"cmd"`
	URL       string                 `json:"url"`
	Session   string                 `json:"session,omitempty"`
	MaxTimeout int                   `json:"maxTimeout,omitempty"`
	Proxy     *FlareSolverrProxy     `json:"proxy,omitempty"`
	Cookies   []FlareSolverrCookie   `json:"cookies,omitempty"`
}

// FlareSolverrProxy FlareSolverr代理结构
type FlareSolverrProxy struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// FlareSolverrSessionRequest FlareSolverr会话请求结构
type FlareSolverrSessionRequest struct {
	Cmd     string             `json:"cmd"`
	Session string             `json:"session"`
	Proxy   *FlareSolverrProxy `json:"proxy,omitempty"`
}

// NewBrowserHelper 创建新的浏览器助手实�?func NewBrowserHelper(browserType string) *BrowserHelper {
	client := resty.New()
	client.SetTimeout(60 * time.Second)
	
	return &BrowserHelper{
		browserType: browserType,
		client:      client,
	}
}

// NewDefaultBrowserHelper 创建默认浏览器助手实�?chromium)
func NewDefaultBrowserHelper() *BrowserHelper {
	return NewBrowserHelper("chromium")
}

// fsCookieStr 将cookies转换为字符串格式
func (bh *BrowserHelper) fsCookieStr(cookies []FlareSolverrCookie) string {
	if len(cookies) == 0 {
		return ""
	}
	
	var cookieStrings []string
	for _, cookie := range cookies {
		if cookie.Name != "" {
			cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
	}
	
	return strings.Join(cookieStrings, "; ")
}

// flaresolverrRequest 调用FlareSolverr解决Cloudflare并返回结�?func (bh *BrowserHelper) flaresolverrRequest(targetURL, cookies string, proxyConfig map[string]interface{}, timeout int) (*FlareSolverrResponse, error) {
	settings := config.GetConfig()
	
	if settings.FlareSolverrURL == "" {
		logger.GetLoggerManager().Warn("未配置FLARESOLVERR_URL，无法使用FlareSolverr")
		return nil, fmt.Errorf("未配置FLARESOLVERR_URL")
	}
	
	fsAPI := strings.TrimRight(settings.FlareSolverrURL, "/") + "/v1"
	sessionID := ""
	
	defer func() {
		// 清理会话
		if sessionID != "" {
			bh.destroyFlareSolverrSession(fsAPI, sessionID)
		}
	}()
	
	needProxyAuth := false
	if proxyConfig != nil && proxyConfig["server"] != nil {
		server := proxyConfig["server"].(string)
		username, _ := proxyConfig["username"].(string)
		password, _ := proxyConfig["password"].(string)
		needProxyAuth = server != "" && (username != "" || password != "")
	}
	
	var solution *FlareSolverrResponse
	
	if needProxyAuth {
		// 使用session模式支持代理认证
		logger.GetLoggerManager().Debug("检测到flaresolverr代理需要认证，使用session模式")
		
		// 1. 创建会话
		var err error
		sessionID, err = bh.createFlareSolverrSession(fsAPI, proxyConfig)
		if err != nil {
			logger.GetLoggerManager().Error("创建FlareSolverr会话失败", zap.Error(err))
			return nil, err
		}
		
		// 2. 使用会话发送请�?		solution, err = bh.requestWithSession(fsAPI, sessionID, targetURL, cookies, timeout)
		if err != nil {
			logger.GetLoggerManager().Error("使用会话发送请求失�?, zap.Error(err))
			return nil, err
		}
	} else {
		// 使用普通模�?无代理认�?
		var err error
		solution, err = bh.requestWithoutSession(fsAPI, targetURL, cookies, proxyConfig, timeout)
		if err != nil {
			logger.GetLoggerManager().Error("发送普通请求失�?, zap.Error(err))
			return nil, err
		}
	}
	
	return solution, nil
}

// createFlareSolverrSession 创建FlareSolverr会话
func (bh *BrowserHelper) createFlareSolverrSession(apiURL string, proxyConfig map[string]interface{}) (string, error) {
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	
	request := FlareSolverrSessionRequest{
		Cmd:     "sessions.create",
		Session: sessionID,
	}
	
	if proxyConfig != nil && proxyConfig["server"] != nil {
		server := proxyConfig["server"].(string)
		if server != "" {
			proxy := &FlareSolverrProxy{
				URL: server,
			}
			
			if username, ok := proxyConfig["username"].(string); ok {
				proxy.Username = username
			}
			
			if password, ok := proxyConfig["password"].(string); ok {
				proxy.Password = password
			}
			
			request.Proxy = proxy
		}
	}
	
	resp, err := bh.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(apiURL)
	
	if err != nil {
		return "", err
	}
	
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("创建会话失败，状态码�?d", resp.StatusCode())
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", err
	}
	
	if status, ok := result["status"].(string); !ok || status != "ok" {
		message := "未知错误"
		if msg, ok := result["message"].(string); ok {
			message = msg
		}
		return "", fmt.Errorf("创建会话失败�?s", message)
	}
	
	return sessionID, nil
}

// destroyFlareSolverrSession 销毁FlareSolverr会话
func (bh *BrowserHelper) destroyFlareSolverrSession(apiURL, sessionID string) {
	request := FlareSolverrSessionRequest{
		Cmd:     "sessions.destroy",
		Session: sessionID,
	}
	
	resp, err := bh.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(apiURL)
	
	if err != nil {
		logger.Warn().Err(err).Msg("销毁FlareSolverr会话失败")
		return
	}
	
	if resp.StatusCode() != http.StatusOK {
		logger.Warn().Msgf("销毁FlareSolverr会话返回�?00状态码�?d", resp.StatusCode())
		return
	}
	
	logger.Debug().Msgf("已清理FlareSolverr会话�?s", sessionID)
}

// requestWithSession 使用会话发送请�?func (bh *BrowserHelper) requestWithSession(apiURL, sessionID, targetURL, cookies string, timeout int) (*FlareSolverrResponse, error) {
	request := FlareSolverrRequest{
		Cmd:       "request.get",
		URL:       targetURL,
		Session:   sessionID,
		MaxTimeout: timeout * 1000,
	}
	
	// 解析cookies
	if cookies != "" {
		parsedCookies, err := bh.parseCookies(cookies)
		if err == nil {
			request.Cookies = parsedCookies
		} else {
			logger.GetLoggerManager().Debug("解析cookies失败，忽�?, zap.Error(err))
		}
	}
	
	resp, err := bh.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(apiURL)
	
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码�?d", resp.StatusCode())
	}
	
	var result FlareSolverrResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	
	if result.Status != "ok" {
		return nil, fmt.Errorf("FlareSolverr调用失败�?s", result.Message)
	}
	
	return &result, nil
}

// requestWithoutSession 使用普通模式发送请�?func (bh *BrowserHelper) requestWithoutSession(apiURL, targetURL, cookies string, proxyConfig map[string]interface{}, timeout int) (*FlareSolverrResponse, error) {
	request := FlareSolverrRequest{
		Cmd:       "request.get",
		URL:       targetURL,
		MaxTimeout: timeout * 1000,
	}
	
	// 添加代理配置(仅URL，无认证)
	if proxyConfig != nil && proxyConfig["server"] != nil {
		server := proxyConfig["server"].(string)
		if server != "" {
			request.Proxy = &FlareSolverrProxy{
				URL: server,
			}
		}
	}
	
	// 解析cookies
	if cookies != "" {
		parsedCookies, err := bh.parseCookies(cookies)
		if err == nil {
			request.Cookies = parsedCookies
		} else {
			logger.GetLoggerManager().Debug("解析cookies失败，忽�?, zap.Error(err))
		}
	}
	
	resp, err := bh.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(apiURL)
	
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码�?d", resp.StatusCode())
	}
	
	var result FlareSolverrResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	
	if result.Status != "ok" {
		return nil, fmt.Errorf("FlareSolverr调用失败�?s", result.Message)
	}
	
	return &result, nil
}

// parseCookies 解析cookies字符�?func (bh *BrowserHelper) parseCookies(cookieStr string) ([]FlareSolverrCookie, error) {
	// 这里简化处理，实际可能需要更复杂的解析逻辑
	cookies := make([]FlareSolverrCookie, 0)
	pairs := strings.Split(cookieStr, ";")
	
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		equalIndex := strings.Index(pair, "=")
		if equalIndex > 0 && equalIndex < len(pair)-1 {
			name := pair[:equalIndex]
			value := pair[equalIndex+1:]
			cookies = append(cookies, FlareSolverrCookie{
				Name:  name,
				Value: value,
			})
		}
	}
	
	return cookies, nil
}

// GetPageSource 获取网页源码
func (bh *BrowserHelper) GetPageSource(targetURL, cookies, userAgent string, proxies map[string]interface{}, headless bool, timeout int) (string, error) {
	settings := config.GetConfig()
	
	// 如果配置为FlareSolverr，则直接调用获取页面源码
	if settings.BrowserEmulation == "flaresolverr" {
		solution, err := bh.flaresolverrRequest(targetURL, cookies, proxies, timeout)
		if err != nil {
			logger.GetLoggerManager().Error("FlareSolverr获取源码失败", zap.Error(err))
			return "", err
		}
		return solution.Solution.Response, nil
	}
	
	// 使用普通的HTTP客户端获取页面内�?	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}
	
	// 设置User-Agent
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	
	// 设置cookies
	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}
	
	// 设置代理
	if proxies != nil && proxies["server"] != nil {
		proxyURL := proxies["server"].(string)
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}

// Action 访问网页并执行操�?简化版)
func (bh *BrowserHelper) Action(targetURL string, callback func(string) interface{}, cookies, userAgent string, proxies map[string]interface{}, headless bool, timeout int) interface{} {
	source, err := bh.GetPageSource(targetURL, cookies, userAgent, proxies, headless, timeout)
	if err != nil {
		logger.GetLoggerManager().Error("获取网页源码失败", zap.Error(err))
		return nil
	}
	
	return callback(source)
}
