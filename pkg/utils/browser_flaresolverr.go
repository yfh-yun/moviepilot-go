package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// FlareSolverrRequest FlareSolverr请求结构
type FlareSolverrRequest struct {
	Cmd       string            `json:"cmd"`
	URL       string            `json:"url"`
	Cookies   []Cookie          `json:"cookies,omitempty"`
	Proxy     map[string]string `json:"proxy,omitempty"`
	Timeout   int               `json:"timeout,omitempty"`
	UserAgent string            `json:"userAgent,omitempty"`
}

// Cookie Cookie结构
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires,omitempty"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

// FlareSolverrResponse FlareSolverr响应结构
type FlareSolverrResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Solution struct {
		URL            string   `json:"url"`
		Status         int      `json:"status"`
		Cookies        []Cookie `json:"cookies"`
		UserAgent      string   `json:"userAgent"`
		Response       string   `json:"response"`
		SessionID      string   `json:"sessionId"`
		StartTimestamp int64    `json:"startTimestamp"`
		EndTimestamp   int64    `json:"endTimestamp"`
		WallTime       float64  `json:"wallTime"`
	} `json:"solution"`
}

// FlareSolverrHelper FlareSolverr助手
type FlareSolverrHelper struct {
	url    string
	logger *zap.Logger
}

// NewFlareSolverrHelper 创建FlareSolverr助手
func NewFlareSolverrHelper(url string) *FlareSolverrHelper {
	return &FlareSolverrHelper{
		url:    url,
		logger: logger.GetLogger(),
	}
}

// SolveCloudflare 使用FlareSolverr解决Cloudflare验证
func (f *FlareSolverrHelper) SolveCloudflare(url string, cookies []Cookie, proxyConfig map[string]string, timeout int) (*FlareSolverrResponse, error) {
	f.logger.Info("使用FlareSolverr解决Cloudflare验证", zap.String("url", url), zap.String("flaresolverr_url", f.url))

	if f.url == "" {
		return nil, fmt.Errorf("未配置FlareSolverr URL")
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/v1", f.url)

	// 构建请求
	req := FlareSolverrRequest{
		Cmd:     "request.get",
		URL:     url,
		Cookies: cookies,
		Proxy:   proxyConfig,
		Timeout: timeout,
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// 发送请求
	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 执行请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var fsResp FlareSolverrResponse
	if err := json.NewDecoder(resp.Body).Decode(&fsResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if fsResp.Status != "ok" {
		return nil, fmt.Errorf("FlareSolverr请求失败: %s", fsResp.Message)
	}

	f.logger.Info("FlareSolverr解决Cloudflare验证成功", zap.String("url", url))
	return &fsResp, nil
}

// GetCookieString 从FlareSolverr响应中获取Cookie字符串
func (f *FlareSolverrHelper) GetCookieString(cookies []Cookie) string {
	if len(cookies) == 0 {
		return ""
	}

	cookieStr := ""
	for _, cookie := range cookies {
		if cookie.Name != "" && cookie.Value != "" {
			if cookieStr != "" {
				cookieStr += "; "
			}
			cookieStr += fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
		}
	}

	return cookieStr
}
