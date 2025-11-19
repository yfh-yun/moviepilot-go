// Package transmission Transmission下载器客户端
package transmission

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"net/http"
	"time"
)

// Client Transmission客户端
type Client struct {
	protocol  string
	host      string
	port      int
	username  string
	password  string
	client    *http.Client
	sessionID string
	ctx       context.Context
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	HashString         string    `json:"hashString"`
	Status             int       `json:"status"`
	TotalSize          int64     `json:"totalSize"`
	PercentDone        float64   `json:"percentDone"`
	DownloadedEver     int64     `json:"downloadedEver"`
	UploadedEver       int64     `json:"uploadedEver"`
	UploadRatio        float64   `json:"uploadRatio"`
	RateDownload       int64     `json:"rateDownload"`
	RateUpload         int64     `json:"rateUpload"`
	Eta                int64     `json:"eta"`
	LeftUntilDone      int64     `json:"leftUntilDone"`
	PeersGettingFromUs int       `json:"peersGettingFromUs"`
	PeersSendingToUs   int       `json:"peersSendingToUs"`
	AddedDate          int64     `json:"addedDate"`
	DoneDate           int64     `json:"doneDate"`
	ActivityDate       int64     `json:"activityDate"`
	DownloadDir        string    `json:"downloadDir"`
	Error              int       `json:"error"`
	ErrorString        string    `json:"errorString"`
	Labels             []string  `json:"labels"`
	TrackerStats       []Tracker `json:"trackerStats"`
	QueuePosition      int       `json:"queuePosition"`
}

// Tracker Tracker信息
type Tracker struct {
	Announce string `json:"announce"`
	ID       int    `json:"id"`
	Tier     int    `json:"tier"`
}

// TorrentFile 种子文件信息
type TorrentFile struct {
	BytesCompleted int64  `json:"bytesCompleted"`
	Length         int64  `json:"length"`
	Name           string `json:"name"`
}

// SessionStats 会话统计
type SessionStats struct {
	ActiveTorrentCount int64 `json:"activeTorrentCount"`
	DownloadSpeed      int64 `json:"downloadSpeed"`
	UploadSpeed        int64 `json:"uploadSpeed"`
	PausedTorrentCount int64 `json:"pausedTorrentCount"`
	TorrentCount       int64 `json:"torrentCount"`
	DownloadedBytes    int64 `json:"current-stats.downloadedBytes"`
	UploadedBytes      int64 `json:"current-stats.uploadedBytes"`
}

// Session 会话信息
type Session struct {
	SpeedLimitDown        int64  `json:"speed-limit-down"`
	SpeedLimitUp          int64  `json:"speed-limit-up"`
	SpeedLimitDownEnabled bool   `json:"speed-limit-down-enabled"`
	SpeedLimitUpEnabled   bool   `json:"speed-limit-up-enabled"`
	DownloadDir           string `json:"download-dir"`
}

// rpcRequest RPC请求结构
type rpcRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Tag       int                    `json:"tag,omitempty"`
}

// rpcResponse RPC响应结构
type rpcResponse struct {
	Arguments interface{} `json:"arguments"`
	Result    string      `json:"result"`
	Tag       int         `json:"tag"`
}

// NewClient 创建Transmission客户端
func NewClient(protocol, host string, port int, username, password string) (*Client, error) {
	if protocol == "" {
		protocol = "http"
	}

	client := &Client{
		protocol: protocol,
		host:     host,
		port:     port,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		ctx: context.Background(),
	}

	// 测试连接
	if _, err := client.GetSession(); err != nil {
		return nil, fmt.Errorf("transmission connection failed: %w", err)
	}

	logger.Info("Transmission client connected successfully", "host", host, "port", port)
	return client, nil
}

// request 发送RPC请求
func (c *Client) request(method string, args map[string]interface{}) ([]byte, error) {
	reqURL := fmt.Sprintf("%s://%s:%d/transmission/rpc", c.protocol, c.host, c.port)

	reqData := rpcRequest{
		Method:    method,
		Arguments: args,
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加Basic认证
	if c.username != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	// 添加Session ID
	if c.sessionID != "" {
		req.Header.Set("X-Transmission-Session-Id", c.sessionID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 如果返回409,需要更新Session ID
	if resp.StatusCode == 409 {
		c.sessionID = resp.Header.Get("X-Transmission-Session-Id")
		// 重试请求
		return c.request(method, args)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Result != "success" {
		return nil, fmt.Errorf("rpc error: %s", rpcResp.Result)
	}

	result, err := json.Marshal(rpcResp.Arguments)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetTorrents 获取种子列表
func (c *Client) GetTorrents(ids []int64, fields []string) ([]*TorrentInfo, error) {
	args := make(map[string]interface{})

	if len(fields) == 0 {
		fields = []string{
			"id", "name", "hashString", "status", "totalSize", "percentDone",
			"downloadedEver", "uploadedEver", "uploadRatio", "rateDownload",
			"rateUpload", "eta", "leftUntilDone", "peersGettingFromUs",
			"peersSendingToUs", "addedDate", "doneDate", "activityDate",
			"downloadDir", "error", "errorString", "labels", "trackerStats",
			"queuePosition",
		}
	}
	args["fields"] = fields

	if len(ids) > 0 {
		args["ids"] = ids
	}

	body, err := c.request("torrent-get", args)
	if err != nil {
		return nil, err
	}

	var result struct {
		Torrents []*TorrentInfo `json:"torrents"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Torrents, nil
}

// AddTorrent 添加种子
func (c *Client) AddTorrent(torrentURL string, torrentFile []byte, downloadDir string, isPaused bool, labels []string) (*TorrentInfo, error) {
	args := make(map[string]interface{})

	if torrentURL != "" {
		args["filename"] = torrentURL
	} else if torrentFile != nil {
		args["metainfo"] = base64.StdEncoding.EncodeToString(torrentFile)
	} else {
		return nil, fmt.Errorf("either torrentURL or torrentFile must be provided")
	}

	if downloadDir != "" {
		args["download-dir"] = downloadDir
	}
	if isPaused {
		args["paused"] = true
	}
	if len(labels) > 0 {
		args["labels"] = labels
	}

	body, err := c.request("torrent-add", args)
	if err != nil {
		return nil, err
	}

	var result struct {
		TorrentAdded     *TorrentInfo `json:"torrent-added"`
		TorrentDuplicate *TorrentInfo `json:"torrent-duplicate"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.TorrentAdded != nil {
		return result.TorrentAdded, nil
	}
	if result.TorrentDuplicate != nil {
		return result.TorrentDuplicate, nil
	}

	return nil, fmt.Errorf("failed to add torrent")
}

// DeleteTorrents 删除种子
func (c *Client) DeleteTorrents(ids []int64, deleteFiles bool) error {
	args := map[string]interface{}{
		"ids":               ids,
		"delete-local-data": deleteFiles,
	}

	_, err := c.request("torrent-remove", args)
	return err
}

// StartTorrents 启动种子
func (c *Client) StartTorrents(ids []int64) error {
	args := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.request("torrent-start", args)
	return err
}

// StopTorrents 停止种子
func (c *Client) StopTorrents(ids []int64) error {
	args := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.request("torrent-stop", args)
	return err
}

// VerifyTorrents 校验种子
func (c *Client) VerifyTorrents(ids []int64) error {
	args := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.request("torrent-verify", args)
	return err
}

// SetTorrent 设置种子属性
func (c *Client) SetTorrent(id int64, labels []string, downloadLimit, uploadLimit int64) error {
	args := map[string]interface{}{
		"ids": []int64{id},
	}

	if len(labels) > 0 {
		args["labels"] = labels
	}
	if downloadLimit >= 0 {
		args["downloadLimited"] = true
		args["downloadLimit"] = downloadLimit
	}
	if uploadLimit >= 0 {
		args["uploadLimited"] = true
		args["uploadLimit"] = uploadLimit
	}

	_, err := c.request("torrent-set", args)
	return err
}

// GetTorrentFiles 获取种子文件列表
func (c *Client) GetTorrentFiles(id int64) ([]*TorrentFile, error) {
	args := map[string]interface{}{
		"ids":    []int64{id},
		"fields": []string{"files"},
	}

	body, err := c.request("torrent-get", args)
	if err != nil {
		return nil, err
	}

	var result struct {
		Torrents []struct {
			Files []*TorrentFile `json:"files"`
		} `json:"torrents"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Torrents) == 0 {
		return nil, fmt.Errorf("torrent not found")
	}

	return result.Torrents[0].Files, nil
}

// SetFilePriority 设置文件优先级
func (c *Client) SetFilePriority(id int64, fileIDs []int, wanted bool) error {
	args := map[string]interface{}{
		"ids": []int64{id},
	}

	if wanted {
		args["files-wanted"] = fileIDs
	} else {
		args["files-unwanted"] = fileIDs
	}

	_, err := c.request("torrent-set", args)
	return err
}

// GetSessionStats 获取会话统计
func (c *Client) GetSessionStats() (*SessionStats, error) {
	body, err := c.request("session-stats", nil)
	if err != nil {
		return nil, err
	}

	var stats SessionStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetSession 获取会话信息
func (c *Client) GetSession() (*Session, error) {
	body, err := c.request("session-get", nil)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// SetSession 设置会话
func (c *Client) SetSession(downloadLimit, uploadLimit int64, downloadLimitEnabled, uploadLimitEnabled bool) error {
	args := map[string]interface{}{
		"speed-limit-down":         downloadLimit,
		"speed-limit-up":           uploadLimit,
		"speed-limit-down-enabled": downloadLimitEnabled,
		"speed-limit-up-enabled":   uploadLimitEnabled,
	}

	_, err := c.request("session-set", args)
	return err
}

// AddTrackers 添加Tracker
func (c *Client) AddTrackers(id int64, trackers []string) error {
	// Transmission使用tracker-add方法
	for _, tracker := range trackers {
		args := map[string]interface{}{
			"ids":        []int64{id},
			"trackerAdd": []string{tracker},
		}
		if _, err := c.request("torrent-set", args); err != nil {
			return err
		}
	}
	return nil
}

// GetStatusString 获取状态字符串
func (t *TorrentInfo) GetStatusString() string {
	status := t.Status
	statusMap := map[int]string{
		0: "stopped",
		1: "check_pending",
		2: "checking",
		3: "download_pending",
		4: "downloading",
		5: "seed_pending",
		6: "seeding",
	}

	if str, ok := statusMap[status]; ok {
		return str
	}
	return "unknown"
}

// IsCompleted 是否已完成
func (t *TorrentInfo) IsCompleted() bool {
	return t.PercentDone >= 1.0
}

// IsSeeding 是否正在做种
func (t *TorrentInfo) IsSeeding() bool {
	return t.Status == 6 // seeding
}

// IsDownloading 是否正在下载
func (t *TorrentInfo) IsDownloading() bool {
	return t.Status == 4 // downloading
}
