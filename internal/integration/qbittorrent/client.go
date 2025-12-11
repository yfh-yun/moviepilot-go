package qbittorrent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/integration/downloader"
	"moviepilot-go/pkg/logger"
)

// Client qBittorrent 客户端
type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
	logger   *zap.Logger
	cookie   string
}

// Config qBittorrent 配置
type Config struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

// NewClient 创建 qBittorrent 客户端
func NewClient(config Config) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建cookie jar失败: %w", err)
	}

	client := &Client{
		baseURL:  strings.TrimRight(config.BaseURL, "/"),
		username: config.Username,
		password: config.Password,
		client: &http.Client{
			Timeout: config.Timeout,
			Jar:     jar,
		},
		logger: logger.GetLogger(),
	}

	return client, nil
}

// Login 登录
func (c *Client) Login(ctx context.Context) error {
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v2/auth/login", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("登录失败: status=%d, body=%s", resp.StatusCode, string(body))
	}

	if string(body) != "Ok." {
		return fmt.Errorf("登录失败: %s", string(body))
	}

	c.logger.Info("qBittorrent 登录成功")
	return nil
}

// Logout 登出
func (c *Client) Logout(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v2/auth/logout", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// AddTorrent 添加种子
func (c *Client) AddTorrent(ctx context.Context, req *downloader.AddTorrentRequest) (*downloader.Torrent, error) {
	// 确保已登录
	if err := c.ensureLoggedIn(ctx); err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加种子URL或文件
	if req.URL != "" {
		_ = writer.WriteField("urls", req.URL)
	} else if len(req.TorrentData) > 0 {
		part, err := writer.CreateFormFile("torrents", "torrent.torrent")
		if err != nil {
			return nil, err
		}
		_, _ = part.Write(req.TorrentData)
	} else {
		return nil, fmt.Errorf("必须提供URL或种子文件")
	}

	// 添加其他参数
	if req.SavePath != "" {
		_ = writer.WriteField("savepath", req.SavePath)
	}
	if req.Category != "" {
		_ = writer.WriteField("category", req.Category)
	}
	if len(req.Tags) > 0 {
		_ = writer.WriteField("tags", strings.Join(req.Tags, ","))
	}
	if req.Paused {
		_ = writer.WriteField("paused", "true")
	}
	if req.SkipChecking {
		_ = writer.WriteField("skip_checking", "true")
	}
	if req.SequentialDownload {
		_ = writer.WriteField("sequentialDownload", "true")
	}
	if req.FirstLastPiecePrio {
		_ = writer.WriteField("firstLastPiecePrio", "true")
	}

	writer.Close()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v2/torrents/add", body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("添加种子请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("添加种子失败: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	result := string(respBody)
	if result != "Ok." {
		return nil, fmt.Errorf("添加种子失败: %s", result)
	}

	c.logger.Info("种子添加成功", zap.String("url", req.URL))

	// 等待种子出现在列表中
	time.Sleep(time.Second)

	// 尝试获取刚添加的种子信息
	torrents, err := c.ListTorrents(ctx, nil)
	if err == nil && len(torrents) > 0 {
		// 返回最新添加的种子
		return torrents[0], nil
	}

	return &downloader.Torrent{}, nil
}

// ListTorrents 列出种子
func (c *Client) ListTorrents(ctx context.Context, filter *downloader.TorrentFilter) ([]*downloader.Torrent, error) {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return nil, err
	}

	params := url.Values{}
	if filter != nil {
		if filter.Category != "" {
			params.Set("category", filter.Category)
		}
		if filter.Tag != "" {
			params.Set("tag", filter.Tag)
		}
		if filter.State != "" {
			params.Set("filter", string(filter.State))
		}
		if len(filter.Hashes) > 0 {
			params.Set("hashes", strings.Join(filter.Hashes, "|"))
		}
	}

	reqURL := c.baseURL + "/api/v2/torrents/info"
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取种子列表失败: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var qbTorrents []qbTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&qbTorrents); err != nil {
		return nil, fmt.Errorf("解析种子列表失败: %w", err)
	}

	torrents := make([]*downloader.Torrent, 0, len(qbTorrents))
	for _, qt := range qbTorrents {
		torrents = append(torrents, qt.toTorrent())
	}

	return torrents, nil
}

// GetTorrentInfo 获取种子详情
func (c *Client) GetTorrentInfo(ctx context.Context, hash string) (*downloader.TorrentInfo, error) {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return nil, err
	}

	// 获取基本信息
	torrents, err := c.ListTorrents(ctx, &downloader.TorrentFilter{
		Hashes: []string{hash},
	})
	if err != nil {
		return nil, err
	}
	if len(torrents) == 0 {
		return nil, fmt.Errorf("种子不存在: %s", hash)
	}

	torrent := torrents[0]

	// 获取文件列表
	files, err := c.getTorrentFiles(ctx, hash)
	if err != nil {
		c.logger.Warn("获取种子文件列表失败", zap.Error(err))
	}

	// 获取tracker信息
	trackers, err := c.getTorrentTrackers(ctx, hash)
	if err != nil {
		c.logger.Warn("获取tracker信息失败", zap.Error(err))
	}

	info := &downloader.TorrentInfo{
		Torrent: torrent,
		Files:   files,
	}

	// 从tracker获取做种者和下载者数量
	if len(trackers) > 0 {
		for _, tracker := range trackers {
			if tracker.NumSeeds > 0 {
				info.Seeders = tracker.NumSeeds
			}
			if tracker.NumLeeches > 0 {
				info.Leechers = tracker.NumLeeches
			}
		}
	}

	return info, nil
}

// PauseTorrent 暂停种子
func (c *Client) PauseTorrent(ctx context.Context, hash string) error {
	return c.controlTorrent(ctx, "pause", hash)
}

// ResumeTorrent 恢复种子
func (c *Client) ResumeTorrent(ctx context.Context, hash string) error {
	return c.controlTorrent(ctx, "resume", hash)
}

// RemoveTorrent 删除种子
func (c *Client) RemoveTorrent(ctx context.Context, hash string, deleteFiles bool) error {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return err
	}

	data := url.Values{}
	data.Set("hashes", hash)
	data.Set("deleteFiles", strconv.FormatBool(deleteFiles))

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v2/torrents/delete", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除种子失败: status=%d, body=%s", resp.StatusCode, string(body))
	}

	c.logger.Info("种子删除成功", zap.String("hash", hash), zap.Bool("delete_files", deleteFiles))
	return nil
}

// SetTorrentCategory 设置种子分类
func (c *Client) SetTorrentCategory(ctx context.Context, hash string, category string) error {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return err
	}

	data := url.Values{}
	data.Set("hashes", hash)
	data.Set("category", category)

	return c.postForm(ctx, "/api/v2/torrents/setCategory", data)
}

// SetTorrentTags 设置种子标签
func (c *Client) SetTorrentTags(ctx context.Context, hash string, tags []string) error {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return err
	}

	data := url.Values{}
	data.Set("hashes", hash)
	data.Set("tags", strings.Join(tags, ","))

	return c.postForm(ctx, "/api/v2/torrents/addTags", data)
}

// GetVersion 获取版本
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/app/version", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(body)), nil
}

// TestConnection 测试连接
func (c *Client) TestConnection(ctx context.Context) error {
	if err := c.Login(ctx); err != nil {
		return err
	}

	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}

	c.logger.Info("qBittorrent 连接测试成功", zap.String("version", version))
	return nil
}

// ensureLoggedIn 确保已登录
func (c *Client) ensureLoggedIn(ctx context.Context) error {
	// 简单检查：尝试获取版本，如果失败则重新登录
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/app/version", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil || resp.StatusCode == http.StatusForbidden {
		// 需要重新登录
		return c.Login(ctx)
	}
	resp.Body.Close()

	return nil
}

// controlTorrent 控制种子（暂停/恢复等）
func (c *Client) controlTorrent(ctx context.Context, action string, hash string) error {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return err
	}

	data := url.Values{}
	data.Set("hashes", hash)

	endpoint := fmt.Sprintf("/api/v2/torrents/%s", action)
	return c.postForm(ctx, endpoint, data)
}

// postForm 发送表单请求
func (c *Client) postForm(ctx context.Context, endpoint string, data url.Values) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// getTorrentFiles 获取种子文件列表
func (c *Client) getTorrentFiles(ctx context.Context, hash string) ([]*downloader.TorrentFile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/torrents/files?hash="+hash, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var qbFiles []qbTorrentFile
	if err := json.NewDecoder(resp.Body).Decode(&qbFiles); err != nil {
		return nil, err
	}

	files := make([]*downloader.TorrentFile, 0, len(qbFiles))
	for _, qf := range qbFiles {
		files = append(files, qf.toTorrentFile())
	}

	return files, nil
}

// getTorrentTrackers 获取tracker信息
func (c *Client) getTorrentTrackers(ctx context.Context, hash string) ([]qbTracker, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/torrents/trackers?hash="+hash, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var trackers []qbTracker
	if err := json.NewDecoder(resp.Body).Decode(&trackers); err != nil {
		return nil, err
	}

	return trackers, nil
}
