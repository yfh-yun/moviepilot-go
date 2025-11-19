package plex

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Client 表示Plex API客户端
type Client struct {
	config     *ClientConfig
	httpClient *http.Client
	logger     *zap.Logger
	baseURL    string
}

// NewClient 创建新的Plex客户端实例
func NewClient(config *ClientConfig, logger *zap.Logger) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:  logger,
		baseURL: config.URL,
	}
}

// HealthCheck 检查Plex服务器健康状态
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/identity", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "创建健康检查请求失败")
	}

	c.addAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "发送健康检查请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Plex服务器健康检查失败，状态码: %d", resp.StatusCode)
	}

	c.logger.Debug("Plex服务器健康检查成功")
	return nil
}

// GetServerInfo 获取服务器信息
func (c *Client) GetServerInfo(ctx context.Context) (*MediaServerInfo, error) {
	url := fmt.Sprintf("%s/identity", c.baseURL)

	var response struct {
		XMLName xml.Name        `xml:"MediaContainer"`
		Info    MediaServerInfo `xml:"attributes"`
	}

	if err := c.doXMLRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, errors.Wrap(err, "获取服务器信息失败")
	}

	c.logger.Debug("获取Plex服务器信息成功", zap.String("server", response.Info.Name))
	return &response.Info, nil
}

// GetUsers 获取用户列表
func (c *Client) GetUsers(ctx context.Context) ([]UserInfo, error) {
	url := fmt.Sprintf("%s/users", c.baseURL)

	var response struct {
		XMLName xml.Name   `xml:"MediaContainer"`
		Users   []UserInfo `xml:"User"`
	}

	if err := c.doXMLRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, errors.Wrap(err, "获取用户列表失败")
	}

	c.logger.Debug("获取Plex用户列表成功", zap.Int("count", len(response.Users)))
	return response.Users, nil
}

// GetLibraries 获取媒体库列表
func (c *Client) GetLibraries(ctx context.Context) ([]LibraryInfo, error) {
	url := fmt.Sprintf("%s/library/sections", c.baseURL)

	var response struct {
		XMLName   xml.Name      `xml:"MediaContainer"`
		Libraries []LibraryInfo `xml:"Directory"`
	}

	if err := c.doXMLRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, errors.Wrap(err, "获取媒体库列表失败")
	}

	c.logger.Debug("获取Plex媒体库列表成功", zap.Int("count", len(response.Libraries)))
	return response.Libraries, nil
}

// GetLibraryItems 获取指定媒体库中的项目
func (c *Client) GetLibraryItems(ctx context.Context, libraryKey string, params map[string]string) ([]MediaItem, error) {
	url := fmt.Sprintf("%s/library/sections/%s/all", c.baseURL, libraryKey)

	// 添加查询参数
	query := c.buildQuery(params)
	if query != "" {
		url = fmt.Sprintf("%s?%s", url, query)
	}

	var response struct {
		XMLName xml.Name    `xml:"MediaContainer"`
		Items   []MediaItem `xml:"Video"`
	}

	if err := c.doXMLRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, errors.Wrap(err, "获取媒体库项目失败")
	}

	c.logger.Debug("获取Plex媒体库项目成功",
		zap.String("library", libraryKey),
		zap.Int("count", len(response.Items)))

	return response.Items, nil
}

// RefreshLibrary 刷新指定媒体库
func (c *Client) RefreshLibrary(ctx context.Context, libraryKey string) error {
	url := fmt.Sprintf("%s/library/sections/%s/refresh", c.baseURL, libraryKey)

	if err := c.doRequest(ctx, "GET", url, nil, nil); err != nil {
		return errors.Wrap(err, "刷新媒体库失败")
	}

	c.logger.Info("刷新Plex媒体库成功", zap.String("library", libraryKey))
	return nil
}

// GetPlaybackSessions 获取播放会话信息
func (c *Client) GetPlaybackSessions(ctx context.Context) ([]PlaybackInfo, error) {
	url := fmt.Sprintf("%s/status/sessions", c.baseURL)

	var response struct {
		XMLName  xml.Name       `xml:"MediaContainer"`
		Sessions []PlaybackInfo `xml:"Video"`
	}

	if err := c.doXMLRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, errors.Wrap(err, "获取播放会话失败")
	}

	c.logger.Debug("获取Plex播放会话成功", zap.Int("count", len(response.Sessions)))
	return response.Sessions, nil
}

// doRequest 执行HTTP请求并进行重试
func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var lastErr error

	for i := 0; i <= c.config.RetryCount; i++ {
		if i > 0 {
			time.Sleep(c.config.RetryDelay)
			c.logger.Warn("重试请求",
				zap.String("method", method),
				zap.String("url", url),
				zap.Int("attempt", i))
		}

		if err := c.executeRequest(ctx, method, url, body, result); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return errors.Wrap(lastErr, fmt.Sprintf("请求失败，重试%d次", c.config.RetryCount))
}

// doXMLRequest 执行XML格式的HTTP请求
func (c *Client) doXMLRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var lastErr error

	for i := 0; i <= c.config.RetryCount; i++ {
		if i > 0 {
			time.Sleep(c.config.RetryDelay)
			c.logger.Warn("重试XML请求",
				zap.String("method", method),
				zap.String("url", url),
				zap.Int("attempt", i))
		}

		if err := c.executeXMLRequest(ctx, method, url, body, result); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return errors.Wrap(lastErr, fmt.Sprintf("XML请求失败，重试%d次", c.config.RetryCount))
}

// executeRequest 执行单个HTTP请求
func (c *Client) executeRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return errors.Wrap(err, "序列化请求体失败")
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return errors.Wrap(err, "创建请求失败")
	}

	c.addAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "发送请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Wrap(err, "解析响应失败")
		}
	}

	return nil
}

// executeXMLRequest 执行XML格式的HTTP请求
func (c *Client) executeXMLRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return errors.Wrap(err, "创建XML请求失败")
	}

	c.addAuthHeaders(req)
	req.Header.Set("Accept", "application/xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "发送XML请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("XML请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := xml.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Wrap(err, "解析XML响应失败")
		}
	}

	return nil
}

// addAuthHeaders 添加认证头信息
func (c *Client) addAuthHeaders(req *http.Request) {
	req.Header.Set("X-Plex-Token", c.config.Token)
	req.Header.Set("Accept", "application/json")
}

// buildQuery 构建查询参数字符串
func (c *Client) buildQuery(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var query string
	for key, value := range params {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("%s=%s", key, value)
	}

	return query
}

// IsConnected 检查客户端是否已连接
func (c *Client) IsConnected() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.HealthCheck(ctx) == nil
}
