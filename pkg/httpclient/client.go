// Package httpclient 提供HTTP客户端工具
// 封装常用的HTTP请求方法，支持超时、重试、日志记录等功能
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client HTTP客户端结构
type Client struct {
	httpClient *http.Client
	logger     *zap.Logger
	baseURL    string
	headers    map[string]string
}

// Options 客户端配置选项
type Options struct {
	BaseURL string            // 基础URL
	Timeout time.Duration     // 超时时间
	Headers map[string]string // 默认请求头
	Logger  *zap.Logger       // 日志记录器
}

// NewClient 创建新的HTTP客户端
// opts: 客户端配置选项
// 返回: 配置好的HTTP客户端实例
func NewClient(opts Options) *Client {
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	client := &Client{
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		logger:  opts.Logger,
		baseURL: opts.BaseURL,
		headers: opts.Headers,
	}

	if client.headers == nil {
		client.headers = make(map[string]string)
	}

	return client
}

// Get 发送GET请求
// ctx: 上下文
// path: 请求路径
// result: 响应结果指针
// 返回: 错误信息
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodGet, path, nil, result)
}

// Post 发送POST请求
// ctx: 上下文
// path: 请求路径
// body: 请求体
// result: 响应结果指针
// 返回: 错误信息
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.request(ctx, http.MethodPost, path, body, result)
}

// Put 发送PUT请求
// ctx: 上下文
// path: 请求路径
// body: 请求体
// result: 响应结果指针
// 返回: 错误信息
func (c *Client) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.request(ctx, http.MethodPut, path, body, result)
}

// Delete 发送DELETE请求
// ctx: 上下文
// path: 请求路径
// result: 响应结果指针
// 返回: 错误信息
func (c *Client) Delete(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodDelete, path, nil, result)
}

// request 统一请求方法
// ctx: 上下文
// method: HTTP方法
// path: 请求路径
// body: 请求体
// result: 响应结果指针
// 返回: 错误信息
func (c *Client) request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// 设置默认请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 如果有body，设置Content-Type
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 记录请求日志
	if c.logger != nil {
		c.logger.Debug("HTTP Request",
			zap.String("method", method),
			zap.String("url", url),
		)
	}

	// 发送请求
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("HTTP Request Failed",
				zap.String("method", method),
				zap.String("url", url),
				zap.Error(err),
			)
		}
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 记录响应时间
	duration := time.Since(startTime)
	if c.logger != nil {
		c.logger.Debug("HTTP Response",
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status", resp.StatusCode),
			zap.Duration("duration", duration),
		)
	}

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// SetHeader 设置请求头
// key: 请求头键
// value: 请求头值
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// RemoveHeader 移除请求头
// key: 请求头键
func (c *Client) RemoveHeader(key string) {
	delete(c.headers, key)
}
