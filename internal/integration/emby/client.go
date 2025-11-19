package emby

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Client 表示Emby API客户端
type Client struct {
	config     *ClientConfig
	httpClient *http.Client
	logger     *zap.Logger
	baseURL    string
}

// NewClient 创建新的Emby客户端实例
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

// HealthCheck 检查Emby服务器健康状态
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/System/Info", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "创建健康检查请求失败")
	}

	req.Header.Set("X-Emby-Token", c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "发送健康检查请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Emby服务器健康检查失败，状态码: %d", resp.StatusCode)
	}

	c.logger.Debug("Emby服务器健康检查成功")
	return nil
}

// GetServerInfo 获取服务器信息
func (c *Client) GetServerInfo(ctx context.Context) (*MediaServerInfo, error) {
	url := fmt.Sprintf("%s/System/Info", c.baseURL)

	var serverInfo MediaServerInfo
	if err := c.doRequest(ctx, "GET", url, nil, &serverInfo); err != nil {
		return nil, errors.Wrap(err, "获取服务器信息失败")
	}

	c.logger.Debug("获取Emby服务器信息成功", zap.String("server", serverInfo.Name))
	return &serverInfo, nil
}

// GetUsers 获取用户列表
func (c *Client) GetUsers(ctx context.Context) ([]UserInfo, error) {
	url := fmt.Sprintf("%s/Users", c.baseURL)

	var users []UserInfo
	if err := c.doRequest(ctx, "GET", url, nil, &users); err != nil {
		return nil, errors.Wrap(err, "获取用户列表失败")
	}

	c.logger.Debug("获取Emby用户列表成功", zap.Int("count", len(users)))
	return users, nil
}

// GetLibraries 获取媒体库列表
func (c *Client) GetLibraries(ctx context.Context) ([]LibraryInfo, error) {
	url := fmt.Sprintf("%s/Library/MediaFolders", c.baseURL)

	var libraries []LibraryInfo
	if err := c.doRequest(ctx, "GET", url, nil, &libraries); err != nil {
		return nil, errors.Wrap(err, "获取媒体库列表失败")
	}

	c.logger.Debug("获取Emby媒体库列表成功", zap.Int("count", len(libraries)))
	return libraries, nil
}

// GetLibraryItems 获取指定媒体库中的项目
func (c *Client) GetLibraryItems(ctx context.Context, libraryID string, params map[string]string) ([]MediaItem, error) {
	url := fmt.Sprintf("%s/Items", c.baseURL)

	// 构建查询参数
	if params == nil {
		params = make(map[string]string)
	}
	params["ParentId"] = libraryID
	params["Recursive"] = "true"

	// 添加查询参数
	query := c.buildQuery(params)
	if query != "" {
		url = fmt.Sprintf("%s?%s", url, query)
	}

	var response struct {
		Items []MediaItem `json:"Items"`
	}

	if err := c.doRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, errors.Wrap(err, "获取媒体库项目失败")
	}

	c.logger.Debug("获取Emby媒体库项目成功",
		zap.String("library", libraryID),
		zap.Int("count", len(response.Items)))

	return response.Items, nil
}

// RefreshLibrary 刷新指定媒体库
func (c *Client) RefreshLibrary(ctx context.Context, libraryID string, request *RefreshRequest) error {
	url := fmt.Sprintf("%s/Items/%s/Refresh", c.baseURL, libraryID)

	if request == nil {
		request = &RefreshRequest{
			MetadataRefreshMode: "Default",
			ImageRefreshMode:    "Default",
			ReplaceAllImages:    false,
			ReplaceAllMetadata:  false,
		}
	}

	if err := c.doRequest(ctx, "POST", url, request, nil); err != nil {
		return errors.Wrap(err, "刷新媒体库失败")
	}

	c.logger.Info("刷新Emby媒体库成功", zap.String("library", libraryID))
	return nil
}

// GetPlaybackSessions 获取播放会话信息
func (c *Client) GetPlaybackSessions(ctx context.Context) ([]PlaybackInfo, error) {
	url := fmt.Sprintf("%s/Sessions", c.baseURL)

	var sessions []PlaybackInfo
	if err := c.doRequest(ctx, "GET", url, nil, &sessions); err != nil {
		return nil, errors.Wrap(err, "获取播放会话失败")
	}

	c.logger.Debug("获取Emby播放会话成功", zap.Int("count", len(sessions)))
	return sessions, nil
}

// UpdatePlaybackStatus 更新播放状态
func (c *Client) UpdatePlaybackStatus(ctx context.Context, sessionID string, status map[string]interface{}) error {
	url := fmt.Sprintf("%s/Sessions/%s/Playing", c.baseURL, sessionID)

	if err := c.doRequest(ctx, "POST", url, status, nil); err != nil {
		return errors.Wrap(err, "更新播放状态失败")
	}

	c.logger.Debug("更新Emby播放状态成功", zap.String("session", sessionID))
	return nil
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

	req.Header.Set("X-Emby-Token", c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

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
