// Package qbittorrent qBittorrent下载器客户端
package qbittorrent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"moviepilot-go/pkg/logger"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client qBittorrent客户端
type Client struct {
	host     string
	port     int
	username string
	password string
	client   *http.Client
	cookie   string
	ctx      context.Context
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	Hash              string  `json:"hash"`
	Name              string  `json:"name"`
	Category          string  `json:"category"`
	Tags              string  `json:"tags"`
	State             string  `json:"state"`
	Size              int64   `json:"size"`
	Progress          float64 `json:"progress"`
	Downloaded        int64   `json:"downloaded"`
	Uploaded          int64   `json:"uploaded"`
	Ratio             float64 `json:"ratio"`
	DownloadSpeed     int64   `json:"dlspeed"`
	UploadSpeed       int64   `json:"upspeed"`
	Eta               int64   `json:"eta"`
	NumSeeds          int     `json:"num_seeds"`
	NumLeechs         int     `json:"num_leechs"`
	SavePath          string  `json:"save_path"`
	ContentPath       string  `json:"content_path"`
	AddedDate         int64   `json:"added_on"`
	CompletionDate    int64   `json:"completion_on"`
	Tracker           string  `json:"tracker"`
	TotalSize         int64   `json:"total_size"`
	CompletedSize     int64   `json:"completed"`
	SequentialDL      bool    `json:"seq_dl"`
	FirstLastPiecePri bool    `json:"f_l_piece_prio"`
}

// TorrentFile 种子文件信息
type TorrentFile struct {
	Index        int     `json:"index"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Progress     float64 `json:"progress"`
	Priority     int     `json:"priority"`
	IsSeed       bool    `json:"is_seed"`
	PieceRange   []int   `json:"piece_range"`
	Availability float64 `json:"availability"`
}

// TransferInfo 传输信息
type TransferInfo struct {
	ConnectionStatus string `json:"connection_status"`
	DhtNodes         int64  `json:"dht_nodes"`
	DownloadSpeed    int64  `json:"dl_info_speed"`
	Downloaded       int64  `json:"dl_info_data"`
	UploadSpeed      int64  `json:"up_info_speed"`
	Uploaded         int64  `json:"up_info_data"`
	FreeSpaceOnDisk  int64  `json:"free_space_on_disk"`
}

// Preferences QB全局偏好设置
type Preferences struct {
	TorrentContentLayout string `json:"torrent_content_layout"`
	SavePath             string `json:"save_path"`
	TempPathEnabled      bool   `json:"temp_path_enabled"`
	TempPath             string `json:"temp_path"`
}

// NewClient 创建qBittorrent客户端
func NewClient(host string, port int, username, password string) (*Client, error) {
	client := &Client{
		host:     host,
		port:     port,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		ctx: context.Background(),
	}

	// 登录认证
	if err := client.login(); err != nil {
		return nil, fmt.Errorf("qBittorrent login failed: %w", err)
	}

	logger.Info("qBittorrent client connected successfully", "host", host, "port", port)
	return client, nil
}

// login 登录qBittorrent
func (c *Client) login() error {
	loginURL := fmt.Sprintf("http://%s:%d/api/v2/auth/login", c.host, c.port)

	data := url.Values{}
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", fmt.Sprintf("http://%s:%d", c.host, c.port))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 保存Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			c.cookie = cookie.Value
			break
		}
	}

	if c.cookie == "" {
		return errors.New("failed to get session cookie")
	}

	return nil
}

// request 发送HTTP请求
func (c *Client) request(method, endpoint string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("http://%s:%d%s", c.host, c.port, endpoint)

	var req *http.Request
	var err error

	if method == "POST" {
		req, err = http.NewRequest(method, reqURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		if params != nil {
			reqURL += "?" + params.Encode()
		}
		req, err = http.NewRequest(method, reqURL, nil)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Referer", fmt.Sprintf("http://%s:%d", c.host, c.port))
	req.AddCookie(&http.Cookie{Name: "SID", Value: c.cookie})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetTorrents 获取种子列表
func (c *Client) GetTorrents(filter, category, tag, sort string, reverse bool, limit, offset int) ([]*TorrentInfo, error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if category != "" {
		params.Set("category", category)
	}
	if tag != "" {
		params.Set("tag", tag)
	}
	if sort != "" {
		params.Set("sort", sort)
	}
	if reverse {
		params.Set("reverse", "true")
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))
	}

	body, err := c.request("GET", "/api/v2/torrents/info", params)
	if err != nil {
		return nil, err
	}

	var torrents []*TorrentInfo
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, err
	}

	return torrents, nil
}

// GetTorrentsByHashes 根据Hash获取种子列表
func (c *Client) GetTorrentsByHashes(hashes []string) ([]*TorrentInfo, error) {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))

	body, err := c.request("GET", "/api/v2/torrents/info", params)
	if err != nil {
		return nil, err
	}

	var torrents []*TorrentInfo
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, err
	}

	return torrents, nil
}

// AddTorrent 添加种子
func (c *Client) AddTorrent(torrentURL string, torrentFile []byte, savePath, category, tags string, isPaused, sequentialDownload, firstLastPiecePrio bool) error {
	params := url.Values{}

	if torrentURL != "" {
		params.Set("urls", torrentURL)
	}
	if savePath != "" {
		params.Set("savepath", savePath)
	}
	if category != "" {
		params.Set("category", category)
	}
	if tags != "" {
		params.Set("tags", tags)
	}
	if isPaused {
		params.Set("paused", "true")
	}
	if sequentialDownload {
		params.Set("sequentialDownload", "true")
	}
	if firstLastPiecePrio {
		params.Set("firstLastPiecePrio", "true")
	}

	_, err := c.request("POST", "/api/v2/torrents/add", params)
	return err
}

// DeleteTorrents 删除种子
func (c *Client) DeleteTorrents(hashes []string, deleteFiles bool) error {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))
	params.Set("deleteFiles", fmt.Sprintf("%t", deleteFiles))

	_, err := c.request("POST", "/api/v2/torrents/delete", params)
	return err
}

// PauseTorrents 暂停种子
func (c *Client) PauseTorrents(hashes []string) error {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))

	_, err := c.request("POST", "/api/v2/torrents/pause", params)
	return err
}

// ResumeTorrents 恢复种子
func (c *Client) ResumeTorrents(hashes []string) error {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))

	_, err := c.request("POST", "/api/v2/torrents/resume", params)
	return err
}

// RecheckTorrents 重新校验种子
func (c *Client) RecheckTorrents(hashes []string) error {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))

	_, err := c.request("POST", "/api/v2/torrents/recheck", params)
	return err
}

// SetForceStart 设置强制启动
func (c *Client) SetForceStart(hashes []string, enable bool) error {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))
	params.Set("value", fmt.Sprintf("%t", enable))

	_, err := c.request("POST", "/api/v2/torrents/setForceStart", params)
	return err
}

// AddTags 添加标签
func (c *Client) AddTags(hashes []string, tags string) error {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))
	params.Set("tags", tags)

	_, err := c.request("POST", "/api/v2/torrents/addTags", params)
	return err
}

// RemoveTags 移除标签
func (c *Client) RemoveTags(hashes []string, tags string) error {
	params := url.Values{}
	params.Set("hashes", strings.Join(hashes, "|"))
	params.Set("tags", tags)

	_, err := c.request("POST", "/api/v2/torrents/removeTags", params)
	return err
}

// GetTorrentFiles 获取种子文件列表
func (c *Client) GetTorrentFiles(hash string) ([]*TorrentFile, error) {
	params := url.Values{}
	params.Set("hash", hash)

	body, err := c.request("GET", "/api/v2/torrents/files", params)
	if err != nil {
		return nil, err
	}

	var files []*TorrentFile
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, err
	}

	return files, nil
}

// SetFilePriority 设置文件优先级
func (c *Client) SetFilePriority(hash string, fileIDs []int, priority int) error {
	params := url.Values{}
	params.Set("hash", hash)

	idStrs := make([]string, len(fileIDs))
	for i, id := range fileIDs {
		idStrs[i] = fmt.Sprintf("%d", id)
	}
	params.Set("id", strings.Join(idStrs, "|"))
	params.Set("priority", fmt.Sprintf("%d", priority))

	_, err := c.request("POST", "/api/v2/torrents/filePrio", params)
	return err
}

// GetTransferInfo 获取传输信息
func (c *Client) GetTransferInfo() (*TransferInfo, error) {
	body, err := c.request("GET", "/api/v2/transfer/info", nil)
	if err != nil {
		return nil, err
	}

	var info TransferInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// SetSpeedLimit 设置速度限制
func (c *Client) SetSpeedLimit(downloadLimit, uploadLimit int64) error {
	// 下载限速
	if downloadLimit >= 0 {
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", downloadLimit*1024))
		if _, err := c.request("POST", "/api/v2/transfer/setDownloadLimit", params); err != nil {
			return err
		}
	}

	// 上传限速
	if uploadLimit >= 0 {
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", uploadLimit*1024))
		if _, err := c.request("POST", "/api/v2/transfer/setUploadLimit", params); err != nil {
			return err
		}
	}

	return nil
}

// GetPreferences 获取全局偏好设置
func (c *Client) GetPreferences() (*Preferences, error) {
	body, err := c.request("GET", "/api/v2/app/preferences", nil)
	if err != nil {
		return nil, err
	}

	var prefs Preferences
	if err := json.Unmarshal(body, &prefs); err != nil {
		return nil, err
	}

	return &prefs, nil
}

// AddTrackers 添加Tracker
func (c *Client) AddTrackers(hash string, trackers []string) error {
	params := url.Values{}
	params.Set("hash", hash)
	params.Set("urls", strings.Join(trackers, "\n"))

	_, err := c.request("POST", "/api/v2/torrents/addTrackers", params)
	return err
}
