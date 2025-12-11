package transmission

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/integration/downloader"
	"moviepilot-go/pkg/logger"
)

// Client Transmission 客户端
type Client struct {
	baseURL   string
	username  string
	password  string
	client    *http.Client
	logger    *zap.Logger
	sessionID string
}

// Config Transmission 配置
type Config struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

// NewClient 创建 Transmission 客户端
func NewClient(config Config) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	client := &Client{
		baseURL:  strings.TrimRight(config.BaseURL, "/"),
		username: config.Username,
		password: config.Password,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger.GetLogger(),
	}

	return client, nil
}

// RPC 执行 RPC 调用
func (c *Client) RPC(ctx context.Context, method string, arguments any) (json.RawMessage, error) {
	request := rpcRequest{
		Method:    method,
		Arguments: arguments,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/transmission/rpc", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 设置认证
	if c.username != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	// 设置 Session ID
	if c.sessionID != "" {
		req.Header.Set("X-Transmission-Session-Id", c.sessionID)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 处理 409 错误（需要更新 Session ID）
	if resp.StatusCode == http.StatusConflict {
		c.sessionID = resp.Header.Get("X-Transmission-Session-Id")
		c.logger.Debug("更新 Session ID", zap.String("session_id", c.sessionID))
		// 重试请求
		return c.RPC(ctx, method, arguments)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC 请求失败: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var response rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("RPC 调用失败: %s", response.Result)
	}

	return response.Arguments, nil
}

// AddTorrent 添加种子
func (c *Client) AddTorrent(ctx context.Context, req *downloader.AddTorrentRequest) (*downloader.Torrent, error) {
	args := map[string]any{}

	// 添加种子URL或文件
	if req.URL != "" {
		args["filename"] = req.URL
	} else if len(req.TorrentData) > 0 {
		args["metainfo"] = base64.StdEncoding.EncodeToString(req.TorrentData)
	} else {
		return nil, fmt.Errorf("必须提供URL或种子文件")
	}

	// 添加其他参数
	if req.SavePath != "" {
		args["download-dir"] = req.SavePath
	}
	if req.Paused {
		args["paused"] = true
	}

	result, err := c.RPC(ctx, "torrent-add", args)
	if err != nil {
		return nil, err
	}

	var addResult struct {
		TorrentAdded     *trTorrent `json:"torrent-added"`
		TorrentDuplicate *trTorrent `json:"torrent-duplicate"`
	}

	if err := json.Unmarshal(result, &addResult); err != nil {
		return nil, err
	}

	var torrent *trTorrent
	if addResult.TorrentAdded != nil {
		torrent = addResult.TorrentAdded
		c.logger.Info("种子添加成功", zap.String("name", torrent.Name))
	} else if addResult.TorrentDuplicate != nil {
		torrent = addResult.TorrentDuplicate
		c.logger.Warn("种子已存在", zap.String("name", torrent.Name))
	} else {
		return nil, fmt.Errorf("添加种子失败：未返回种子信息")
	}

	// 设置分类（通过标签模拟）
	if req.Category != "" && torrent.ID > 0 {
		_ = c.SetTorrentTags(ctx, fmt.Sprintf("%d", torrent.ID), []string{req.Category})
	}

	// 设置标签
	if len(req.Tags) > 0 && torrent.ID > 0 {
		_ = c.SetTorrentTags(ctx, fmt.Sprintf("%d", torrent.ID), req.Tags)
	}

	return torrent.toTorrent(), nil
}

// ListTorrents 列出种子
func (c *Client) ListTorrents(ctx context.Context, filter *downloader.TorrentFilter) ([]*downloader.Torrent, error) {
	args := map[string]any{
		"fields": []string{
			"id", "name", "status", "percentDone", "totalSize",
			"downloadedEver", "uploadedEver", "rateDownload", "rateUpload",
			"eta", "uploadRatio", "labels", "downloadDir",
			"addedDate", "doneDate", "peersConnected", "peersGettingFromUs",
			"peersSendingToUs", "error", "errorString",
		},
	}

	// 添加过滤条件
	if filter != nil && len(filter.Hashes) > 0 {
		// Transmission 使用 ID 而不是 hash
		// 这里需要转换，暂时不支持
		c.logger.Warn("Transmission 不支持按 hash 过滤")
	}

	result, err := c.RPC(ctx, "torrent-get", args)
	if err != nil {
		return nil, err
	}

	var listResult struct {
		Torrents []trTorrent `json:"torrents"`
	}

	if err := json.Unmarshal(result, &listResult); err != nil {
		return nil, err
	}

	torrents := make([]*downloader.Torrent, 0, len(listResult.Torrents))
	for _, tr := range listResult.Torrents {
		torrent := tr.toTorrent()

		// 应用过滤器
		if filter != nil {
			if filter.Category != "" && !contains(torrent.Tags, filter.Category) {
				continue
			}
			if filter.Tag != "" && !contains(torrent.Tags, filter.Tag) {
				continue
			}
			if filter.State != "" && torrent.State != filter.State {
				continue
			}
		}

		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// GetTorrentInfo 获取种子详情
func (c *Client) GetTorrentInfo(ctx context.Context, hash string) (*downloader.TorrentInfo, error) {
	// Transmission 使用 ID，这里假设 hash 就是 ID
	args := map[string]any{
		"ids": []any{hash},
		"fields": []string{
			"id", "name", "status", "percentDone", "totalSize",
			"downloadedEver", "uploadedEver", "rateDownload", "rateUpload",
			"eta", "uploadRatio", "labels", "downloadDir",
			"addedDate", "doneDate", "peersConnected", "peersGettingFromUs",
			"peersSendingToUs", "error", "errorString",
			"files", "fileStats", "trackers", "trackerStats",
			"pieceCount", "pieceSize", "comment", "creator", "dateCreated",
		},
	}

	result, err := c.RPC(ctx, "torrent-get", args)
	if err != nil {
		return nil, err
	}

	var listResult struct {
		Torrents []trTorrent `json:"torrents"`
	}

	if err := json.Unmarshal(result, &listResult); err != nil {
		return nil, err
	}

	if len(listResult.Torrents) == 0 {
		return nil, fmt.Errorf("种子不存在: %s", hash)
	}

	tr := listResult.Torrents[0]
	info := &downloader.TorrentInfo{
		Torrent:      tr.toTorrent(),
		TotalSize:    tr.TotalSize,
		PieceSize:    tr.PieceSize,
		NumPieces:    tr.PieceCount,
		Comment:      tr.Comment,
		Creator:      tr.Creator,
		CreationDate: time.Unix(tr.DateCreated, 0),
	}

	// 转换文件列表
	if len(tr.Files) > 0 {
		info.Files = make([]*downloader.TorrentFile, len(tr.Files))
		for i, file := range tr.Files {
			info.Files[i] = &downloader.TorrentFile{
				Name:     file.Name,
				Size:     file.Length,
				Progress: 0,
				Priority: 0,
				IsSeed:   false,
			}

			// 从 fileStats 获取进度
			if i < len(tr.FileStats) {
				info.Files[i].Progress = float64(tr.FileStats[i].BytesCompleted) / float64(file.Length)
				info.Files[i].Priority = tr.FileStats[i].Priority
			}
		}
	}

	// 统计 seeders 和 leechers
	for _, tracker := range tr.TrackerStats {
		if tracker.SeederCount > 0 {
			info.Seeders += tracker.SeederCount
		}
		if tracker.LeecherCount > 0 {
			info.Leechers += tracker.LeecherCount
		}
	}

	return info, nil
}

// PauseTorrent 暂停种子
func (c *Client) PauseTorrent(ctx context.Context, hash string) error {
	args := map[string]any{
		"ids": []any{hash},
	}

	_, err := c.RPC(ctx, "torrent-stop", args)
	if err != nil {
		return err
	}

	c.logger.Info("种子已暂停", zap.String("id", hash))
	return nil
}

// ResumeTorrent 恢复种子
func (c *Client) ResumeTorrent(ctx context.Context, hash string) error {
	args := map[string]any{
		"ids": []any{hash},
	}

	_, err := c.RPC(ctx, "torrent-start", args)
	if err != nil {
		return err
	}

	c.logger.Info("种子已恢复", zap.String("id", hash))
	return nil
}

// RemoveTorrent 删除种子
func (c *Client) RemoveTorrent(ctx context.Context, hash string, deleteFiles bool) error {
	method := "torrent-remove"
	args := map[string]any{
		"ids":               []any{hash},
		"delete-local-data": deleteFiles,
	}

	_, err := c.RPC(ctx, method, args)
	if err != nil {
		return err
	}

	c.logger.Info("种子已删除", zap.String("id", hash), zap.Bool("delete_files", deleteFiles))
	return nil
}

// SetTorrentCategory 设置种子分类（通过标签模拟）
func (c *Client) SetTorrentCategory(ctx context.Context, hash string, category string) error {
	return c.SetTorrentTags(ctx, hash, []string{category})
}

// SetTorrentTags 设置种子标签
func (c *Client) SetTorrentTags(ctx context.Context, hash string, tags []string) error {
	args := map[string]any{
		"ids":    []any{hash},
		"labels": tags,
	}

	_, err := c.RPC(ctx, "torrent-set", args)
	if err != nil {
		return err
	}

	c.logger.Info("标签设置成功", zap.String("id", hash), zap.Strings("tags", tags))
	return nil
}

// GetVersion 获取版本
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	result, err := c.RPC(ctx, "session-get", map[string]any{})
	if err != nil {
		return "", err
	}

	var session struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(result, &session); err != nil {
		return "", err
	}

	return session.Version, nil
}

// TestConnection 测试连接
func (c *Client) TestConnection(ctx context.Context) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}

	c.logger.Info("Transmission 连接测试成功", zap.String("version", version))
	return nil
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
