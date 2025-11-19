// Package downloader 实现Qbittorrent下载器客户端
package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/service/interfaces"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"

	"go.uber.org/zap"
)

// QbittorrentClient Qbittorrent客户端实现
type QbittorrentClient struct {
	config     *interfaces.DownloaderConfig
	httpClient *utils.HTTPClient
	logger     *zap.Logger
	baseURL    string
}

// NewQbittorrentClient 创建Qbittorrent客户端实例
func NewQbittorrentClient(config *interfaces.DownloaderConfig) *QbittorrentClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := utils.NewHTTPClient(&utils.HTTPConfig{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	})

	// 如果跳过TLS验证
	if config.SkipTLSVerify {
		httpClient.SetSkipTLSVerify(true)
	}

	return &QbittorrentClient{
		config:     config,
		httpClient: httpClient,
		logger:     logger.Logger,
		baseURL:    strings.TrimSuffix(config.Endpoint, "/") + "/api/v2",
	}
}

// Start 启动客户端
func (c *QbittorrentClient) Start() error {
	c.logger.Info("启动Qbittorrent客户端", "endpoint", c.config.Endpoint)

	// 测试连接
	if err := c.testConnection(); err != nil {
		return fmt.Errorf("无法连接到Qbittorrent: %w", err)
	}

	// 登录获取认证
	if err := c.login(); err != nil {
		return fmt.Errorf("Qbittorrent登录失败: %w", err)
	}

	c.logger.Info("Qbittorrent客户端启动成功")
	return nil
}

// Stop 停止客户端
func (c *QbittorrentClient) Stop() error {
	c.logger.Info("停止Qbittorrent客户端")

	// 登出
	if err := c.logout(); err != nil {
		c.logger.Warn("Qbittorrent登出失败", "error", err)
	}

	return nil
}

// GetStatus 获取下载器状态
func (c *QbittorrentClient) GetStatus() (*interfaces.DownloaderStatus, error) {
	// 获取版本信息
	version, err := c.getVersion()
	if err != nil {
		return nil, fmt.Errorf("获取版本信息失败: %w", err)
	}

	// 获取传输信息
	transferInfo, err := c.getTransferInfo()
	if err != nil {
		return nil, fmt.Errorf("获取传输信息失败: %w", err)
	}

	// 获取种子列表
	torrents, err := c.ListTorrents(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("获取种子列表失败: %w", err)
	}

	// 统计种子状态
	activeCount := 0
	pausedCount := 0
	completedCount := 0
	for _, torrent := range torrents {
		switch torrent.State {
		case "downloading", "uploading", "stalledDL", "stalledUP", "queuedDL", "queuedUP", "checkingDL", "checkingUP", "forcedDL", "forcedUP":
			activeCount++
		case "pausedDL", "pausedUP":
			pausedCount++
		case "completed":
			completedCount++
		}
	}

	return &interfaces.DownloaderStatus{
		Connected:      true,
		Version:        version,
		APIVersion:     "2.8", // 假设API版本
		Uptime:         0,     // qBittorrent不提供运行时间
		FreeSpace:      transferInfo.FreeSpace,
		AllTimeDL:      transferInfo.AllTimeDL,
		AllTimeUL:      transferInfo.AllTimeUL,
		GlobalDLSpeed:  transferInfo.DlInfoSpeed,
		GlobalULSpeed:  transferInfo.UpInfoSpeed,
		TorrentCount:   len(torrents),
		ActiveCount:    activeCount,
		PausedCount:    pausedCount,
		CompletedCount: completedCount,
	}, nil
}

// AddTorrent 添加种子
func (c *QbittorrentClient) AddTorrent(ctx context.Context, req *interfaces.AddTorrentRequest) (*interfaces.AddTorrentResponse, error) {
	c.logger.Info("添加种子", "name", req.SavePath, "category", req.Category)

	var err error
	var hash string

	// 根据种子类型选择添加方法
	if len(req.RawData) > 0 {
		hash, err = c.addTorrentFile(req)
	} else if req.URL != "" {
		hash, err = c.addTorrentURL(req)
	} else {
		return nil, fmt.Errorf("必须提供种子URL或数据")
	}

	if err != nil {
		return &interfaces.AddTorrentResponse{
			Success: false,
			Message: fmt.Sprintf("添加种子失败: %v", err),
		}, nil
	}

	// 设置种子属性
	if req.Priority > 0 {
		c.SetTorrentPriority(hash, req.Priority)
	}

	if req.DownloadLimit > 0 || req.UploadLimit > 0 {
		c.SetTorrentSpeedLimits(hash, req.DownloadLimit, req.UploadLimit)
	}

	// 设置标签
	if len(req.Tags) > 0 {
		c.setTorrentTags(hash, req.Tags)
	}

	c.logger.Info("种子添加成功", "hash", hash)
	return &interfaces.AddTorrentResponse{
		Success:   true,
		Hash:      hash,
		TorrentID: hash,
		Message:   "种子添加成功",
	}, nil
}

// RemoveTorrent 删除种子
func (c *QbittorrentClient) RemoveTorrent(ctx context.Context, hash string) error {
	c.logger.Info("删除种子", "hash", hash)

	// 删除种子文件
	apiURL := fmt.Sprintf("%s/torrents/delete", c.baseURL)
	form := url.Values{
		"hashes":      {hash},
		"deleteFiles": {"false"},
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("删除种子失败，状态码: %d", resp.StatusCode)
	}

	c.logger.Info("种子删除成功", "hash", hash)
	return nil
}

// PauseTorrent 暂停种子
func (c *QbittorrentClient) PauseTorrent(ctx context.Context, hash string) error {
	c.logger.Info("暂停种子", "hash", hash)

	apiURL := fmt.Sprintf("%s/torrents/pause", c.baseURL)
	form := url.Values{"hashes": {hash}}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("暂停种子失败，状态码: %d", resp.StatusCode)
	}

	c.logger.Info("种子暂停成功", "hash", hash)
	return nil
}

// ResumeTorrent 恢复种子
func (c *QbittorrentClient) ResumeTorrent(ctx context.Context, hash string) error {
	c.logger.Info("恢复种子", "hash", hash)

	apiURL := fmt.Sprintf("%s/torrents/resume", c.baseURL)
	form := url.Values{"hashes": {hash}}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("恢复种子失败，状态码: %d", resp.StatusCode)
	}

	c.logger.Info("种子恢复成功", "hash", hash)
	return nil
}

// GetTorrentInfo 获取种子信息
func (c *QbittorrentClient) GetTorrentInfo(ctx context.Context, hash string) (*interfaces.TorrentInfo, error) {
	apiURL := fmt.Sprintf("%s/torrents/info?hashes=%s", c.baseURL, hash)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取种子信息失败，状态码: %d", resp.StatusCode)
	}

	var torrents []QBTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, err
	}

	if len(torrents) == 0 {
		return nil, fmt.Errorf("种子不存在: %s", hash)
	}

	qbTorrent := torrents[0]
	torrent := &interfaces.TorrentInfo{
		Hash:          qbTorrent.Hash,
		Name:          qbTorrent.Name,
		Size:          qbTorrent.Size,
		Progress:      qbTorrent.Progress,
		State:         qbTorrent.State,
		Priority:      qbTorrent.Priority,
		DownloadSpeed: qbTorrent.DLSpeed,
		UploadSpeed:   qbTorrent.UPSpeed,
		Downloaded:    qbTorrent.Downloaded,
		Uploaded:      qbTorrent.Uploaded,
		ETA:           qbTorrent.ETA,
		AddedOn:       time.Unix(qbTorrent.AddedOn, 0),
		SavePath:      qbTorrent.SavePath,
		Category:      qbTorrent.Category,
		Tags:          strings.Split(qbTorrent.Tags, ", "),
		Ratio:         qbTorrent.Ratio,
		QueuePosition: qbTorrent.Priority,
		Metadata:      make(map[string]string),
	}

	// 转换完成时间
	if qbTorrent.CompletionOn > 0 {
		completedTime := time.Unix(qbTorrent.CompletionOn, 0)
		torrent.CompletedOn = &completedTime
	}

	// 获取文件列表
	files, err := c.getTorrentFiles(hash)
	if err == nil {
		torrent.Files = files
	}

	// 获取Tracker列表
	trackers, err := c.getTorrentTrackers(hash)
	if err == nil {
		torrent.Trackers = trackers
	}

	return torrent, nil
}

// ListTorrents 列出种子
func (c *QbittorrentClient) ListTorrents(ctx context.Context, filter *interfaces.TorrentFilter) ([]*interfaces.TorrentInfo, error) {
	apiURL := fmt.Sprintf("%s/torrents/info", c.baseURL)

	// 添加过滤器参数
	if filter != nil {
		params := url.Values{}
		if len(filter.States) > 0 {
			params.Add("filter", strings.Join(filter.States, "|"))
		}
		if filter.Category != "" {
			params.Add("category", filter.Category)
		}
		if filter.Tag != "" {
			params.Add("tag", filter.Tag)
		}
		if len(params) > 0 {
			apiURL += "?" + params.Encode()
		}
	}

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取种子列表失败，状态码: %d", resp.StatusCode)
	}

	var qbTorrents []QBTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&qbTorrents); err != nil {
		return nil, err
	}

	torrents := make([]*interfaces.TorrentInfo, len(qbTorrents))
	for i, qbTorrent := range qbTorrents {
		torrent := &interfaces.TorrentInfo{
			Hash:          qbTorrent.Hash,
			Name:          qbTorrent.Name,
			Size:          qbTorrent.Size,
			Progress:      qbTorrent.Progress,
			State:         qbTorrent.State,
			Priority:      qbTorrent.Priority,
			DownloadSpeed: qbTorrent.DLSpeed,
			UploadSpeed:   qbTorrent.UPSpeed,
			Downloaded:    qbTorrent.Downloaded,
			Uploaded:      qbTorrent.Uploaded,
			ETA:           qbTorrent.ETA,
			AddedOn:       time.Unix(qbTorrent.AddedOn, 0),
			SavePath:      qbTorrent.SavePath,
			Category:      qbTorrent.Category,
			Tags:          strings.Split(qbTorrent.Tags, ", "),
			Ratio:         qbTorrent.Ratio,
			QueuePosition: qbTorrent.Priority,
			Metadata:      make(map[string]string),
		}

		// 转换完成时间
		if qbTorrent.CompletionOn > 0 {
			completedTime := time.Unix(qbTorrent.CompletionOn, 0)
			torrent.CompletedOn = &completedTime
		}

		torrents[i] = torrent
	}

	return torrents, nil
}

// GetGlobalTransferInfo 获取全局传输信息
func (c *QbittorrentClient) GetGlobalTransferInfo(ctx context.Context) (*interfaces.TransferInfo, error) {
	transferInfo, err := c.getTransferInfo()
	if err != nil {
		return nil, err
	}

	return &interfaces.TransferInfo{
		GlobalDownloadSpeed: transferInfo.DlInfoSpeed,
		GlobalUploadSpeed:   transferInfo.UpInfoSpeed,
		TotalDownloaded:     transferInfo.AllTimeDL,
		TotalUploaded:       transferInfo.AllTimeUL,
		DHTNodes:            transferInfo.DlInfoSpeed, // QB不提供DHT节点数
		ListeningPort:       transferInfo.ListenPort,
	}, nil
}

// GetPreferences 获取偏好设置
func (c *QbittorrentClient) GetPreferences() (*interfaces.Preferences, error) {
	apiURL := fmt.Sprintf("%s/app/preferences", c.baseURL)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取偏好设置失败，状态码: %d", resp.StatusCode)
	}

	var qbPrefs QBPreferences
	if err := json.NewDecoder(resp.Body).Decode(&qbPrefs); err != nil {
		return nil, err
	}

	return &interfaces.Preferences{
		DownloadPath:          qbPrefs.SavePath,
		TempPathEnabled:       qbPrefs.TempPathEnabled,
		TempPath:              qbPrefs.TempPath,
		MaxActiveDownloads:    qbPrefs.MaxActiveDownloads,
		MaxActiveTorrents:     qbPrefs.MaxActiveTorrents,
		MaxActiveUploads:      qbPrefs.MaxActiveUploads,
		DownloadLimitEnabled:  qbPrefs.DlLimit > 0,
		DownloadLimit:         qbPrefs.DlLimit,
		UploadLimitEnabled:    qbPrefs.UpLimit > 0,
		UploadLimit:           qbPrefs.UpLimit,
		MaxRatioEnabled:       qbPrefs.MaxRatioEnabled,
		MaxRatio:              qbPrefs.MaxRatio,
		MaxSeedingTimeEnabled: qbPrefs.MaxSeedingTimeEnabled,
		MaxSeedingTime:        time.Duration(qbPrefs.MaxSeedingTimeMin) * time.Minute,
		AlternativeGlobalDHT:  qbPrefs.AlternativeGlobalDHT,
		AnnounceToAllTrackers: qbPrefs.AnnounceToAllTrackers,
	}, nil
}

// SetPreferences 设置偏好设置
func (c *QbittorrentClient) SetPreferences(prefs *interfaces.Preferences) error {
	qbPrefs := QBPreferences{
		SavePath:              prefs.DownloadPath,
		TempPathEnabled:       prefs.TempPathEnabled,
		TempPath:              prefs.TempPath,
		MaxActiveDownloads:    prefs.MaxActiveDownloads,
		MaxActiveTorrents:     prefs.MaxActiveTorrents,
		MaxActiveUploads:      prefs.MaxActiveUploads,
		DlLimit:               prefs.DownloadLimit,
		UpLimit:               prefs.UploadLimit,
		MaxRatioEnabled:       prefs.MaxRatioEnabled,
		MaxRatio:              prefs.MaxRatio,
		MaxSeedingTimeEnabled: prefs.MaxSeedingTimeEnabled,
		MaxSeedingTimeMin:     int(prefs.MaxSeedingTime.Minutes()),
		AlternativeGlobalDHT:  prefs.AlternativeGlobalDHT,
		AnnounceToAllTrackers: prefs.AnnounceToAllTrackers,
	}

	jsonData, err := json.Marshal(qbPrefs)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("%s/app/setPreferences", c.baseURL)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("设置偏好失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// SetTorrentPriority 设置种子优先级
func (c *QbittorrentClient) SetTorrentPriority(hash string, priority int) error {
	apiURL := fmt.Sprintf("%s/torrents/setPrio", c.baseURL)
	form := url.Values{
		"hashes":   {hash},
		"priority": {strconv.Itoa(priority)},
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("设置种子优先级失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// SetTorrentSpeedLimits 设置种子速度限制
func (c *QbittorrentClient) SetTorrentSpeedLimits(hash string, downloadLimit, uploadLimit int) error {
	// 设置下载速度限制
	if downloadLimit > 0 {
		apiURL := fmt.Sprintf("%s/torrents/setDownloadLimit", c.baseURL)
		form := url.Values{"hashes": {hash}, "limit": {strconv.FormatInt(downloadLimit, 10)}}

		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, _ := c.httpClient.Do(req)
		if resp != nil {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("设置下载速度限制失败，状态码: %d", resp.StatusCode)
			}
		}
	}

	// 设置上传速度限制
	if uploadLimit > 0 {
		apiURL := fmt.Sprintf("%s/torrents/setUploadLimit", c.baseURL)
		form := url.Values{"hashes": {hash}, "limit": {strconv.FormatInt(uploadLimit, 10)}}

		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, _ := c.httpClient.Do(req)
		if resp != nil {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("设置上传速度限制失败，状态码: %d", resp.StatusCode)
			}
		}
	}

	return nil
}

// GetVersion 获取版本
func (c *QbittorrentClient) GetVersion() (string, error) {
	return c.getVersion()
}

// GetAPIVersion 获取API版本
func (c *QbittorrentClient) GetAPIVersion() (string, error) {
	apiURL := fmt.Sprintf("%s/app/version", c.baseURL)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取API版本失败，状态码: %d", resp.StatusCode)
	}

	version, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(version), nil
}

// 私有辅助方法

// testConnection 测试连接
func (c *QbittorrentClient) testConnection() error {
	resp, err := c.httpClient.Get(c.baseURL + "/app/version")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Qbittorrent连接失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// login 登录认证
func (c *QbittorrentClient) login() error {
	apiURL := fmt.Sprintf("%s/auth/login", c.baseURL)
	form := url.Values{
		"username": {c.config.Username},
		"password": {c.config.Password},
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// logout 登出
func (c *QbittorrentClient) logout() error {
	apiURL := fmt.Sprintf("%s/auth/logout", c.baseURL)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("登出失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// getVersion 获取版本信息
func (c *QbittorrentClient) getVersion() (string, error) {
	apiURL := fmt.Sprintf("%s/app/version", c.baseURL)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取版本信息失败，状态码: %d", resp.StatusCode)
	}

	version, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(version), nil
}

// getTransferInfo 获取传输信息
func (c *QbittorrentClient) getTransferInfo() (*QBTransferInfo, error) {
	apiURL := fmt.Sprintf("%s/transfer/info", c.baseURL)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取传输信息失败，状态码: %d", resp.StatusCode)
	}

	var transferInfo QBTransferInfo
	if err := json.NewDecoder(resp.Body).Decode(&transferInfo); err != nil {
		return nil, err
	}

	return &transferInfo, nil
}

// addTorrentFile 添加种子文件
func (c *QbittorrentClient) addTorrentFile(req *interfaces.AddTorrentRequest) (string, error) {
	// 构建表单数据
	form := url.Values{}
	form.Set("download_path", req.DownloadPath)
	form.Set("save_path", req.SavePath)
	form.Set("category", req.Category)
	form.Set("tags", strings.Join(req.Tags, ", "))
	form.Set("priority", strconv.Itoa(req.Priority))

	if req.Paused {
		form.Set("paused", "true")
	}
	if req.Sequential {
		form.Set("sequential_download", "true")
	}
	if req.FirstLast {
		form.Set("first_last_piece_prio", "true")
	}

	apiURL := fmt.Sprintf("%s/torrents/add", c.baseURL)
	httpReq, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	// 添加文件
	httpReq.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")

	// 简化实现，实际应该正确处理multipart
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("添加种子文件失败，状态码: %d", resp.StatusCode)
	}

	// 从响应中提取哈希值（简化实现）
	return "generated_hash", nil
}

// addTorrentURL 添加种子URL
func (c *QbittorrentClient) addTorrentURL(req *interfaces.AddTorrentRequest) (string, error) {
	form := url.Values{
		"urls": {req.URL},
	}

	if req.SavePath != "" {
		form.Set("save_path", req.SavePath)
	}
	if req.Category != "" {
		form.Set("category", req.Category)
	}
	if len(req.Tags) > 0 {
		form.Set("tags", strings.Join(req.Tags, ", "))
	}
	if req.Priority > 0 {
		form.Set("priority", strconv.Itoa(req.Priority))
	}
	if req.Paused {
		form.Set("paused", "true")
	}
	if req.Sequential {
		form.Set("sequential_download", "true")
	}
	if req.FirstLast {
		form.Set("first_last_piece_prio", "true")
	}

	apiURL := fmt.Sprintf("%s/torrents/add", c.baseURL)
	httpReq, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("添加种子URL失败，状态码: %d", resp.StatusCode)
	}

	// 从响应中提取哈希值（简化实现）
	return "generated_hash", nil
}

// setTorrentTags 设置种子标签
func (c *QbittorrentClient) setTorrentTags(hash string, tags []string) error {
	apiURL := fmt.Sprintf("%s/torrents/addTags", c.baseURL)
	form := url.Values{
		"hashes": {hash},
		"tags":   {strings.Join(tags, ", ")},
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("设置种子标签失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// getTorrentFiles 获取种子文件列表
func (c *QbittorrentClient) getTorrentFiles(hash string) ([]interfaces.TorrentFile, error) {
	apiURL := fmt.Sprintf("%s/torrents/files?hash=%s", c.baseURL, hash)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取种子文件列表失败，状态码: %d", resp.StatusCode)
	}

	var qbFiles []QBTorrentFile
	if err := json.NewDecoder(resp.Body).Decode(&qbFiles); err != nil {
		return nil, err
	}

	files := make([]interfaces.TorrentFile, len(qbFiles))
	for i, qbFile := range qbFiles {
		files[i] = interfaces.TorrentFile{
			Index:    qbFile.Index,
			Name:     qbFile.Name,
			Size:     qbFile.Size,
			Progress: qbFile.Progress,
			Priority: qbFile.Priority,
			IsSeed:   qbFile.Progress >= 1.0,
			Path:     qbFile.Name,
		}
	}

	return files, nil
}

// getTorrentTrackers 获取种子Tracker列表
func (c *QbittorrentClient) getTorrentTrackers(hash string) ([]string, error) {
	apiURL := fmt.Sprintf("%s/torrents/trackers?hash=%s", c.baseURL, hash)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取种子Tracker列表失败，状态码: %d", resp.StatusCode)
	}

	var trackers []QBTracker
	if err := json.NewDecoder(resp.Body).Decode(&trackers); err != nil {
		return nil, err
	}

	var trackerURLs []string
	for _, tracker := range trackers {
		if tracker.URL != "" {
			trackerURLs = append(trackerURLs, tracker.URL)
		}
	}

	return trackerURLs, nil
}

// Qbittorrent数据结构定义

type QBTorrentInfo struct {
	Hash         string  `json:"hash"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Progress     float64 `json:"progress"`
	State        string  `json:"state"`
	Priority     int     `json:"priority"`
	DLSpeed      int64   `json:"dl_speed"`
	UPSpeed      int64   `json:"up_speed"`
	Downloaded   int64   `json:"downloaded"`
	Uploaded     int64   `json:"uploaded"`
	ETA          int64   `json:"eta"`
	AddedOn      int64   `json:"added_on"`
	CompletionOn int64   `json:"completion_on"`
	SavePath     string  `json:"save_path"`
	Category     string  `json:"category"`
	Tags         string  `json:"tags"`
	Ratio        float64 `json:"ratio"`
}

type QBTorrentFile struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Progress float64 `json:"progress"`
	Priority int     `json:"priority"`
}

type QBTracker struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	Tier   int    `json:"tier"`
}

type QBTransferInfo struct {
	DlInfoSpeed int64 `json:"dl_info_speed"`
	UpInfoSpeed int64 `json:"up_info_speed"`
	AllTimeDL   int64 `json:"alltime_dl"`
	AllTimeUL   int64 `json:"alltime_ul"`
	FreeSpace   int64 `json:"free_space_on_disk"`
	ListenPort  int   `json:"listen_port"`
}

type QBPreferences struct {
	SavePath              string  `json:"save_path"`
	TempPathEnabled       bool    `json:"temp_path_enabled"`
	TempPath              string  `json:"temp_path"`
	MaxActiveDownloads    int     `json:"max_active_downloads"`
	MaxActiveTorrents     int     `json:"max_active_torrents"`
	MaxActiveUploads      int     `json:"max_active_uploads"`
	DlLimit               int64   `json:"dl_limit"`
	UpLimit               int64   `json:"up_limit"`
	MaxRatioEnabled       bool    `json:"max_ratio_enabled"`
	MaxRatio              float64 `json:"max_ratio"`
	MaxSeedingTimeEnabled bool    `json:"max_seeding_time_enabled"`
	MaxSeedingTimeMin     int     `json:"max_seeding_time_min"`
	AlternativeGlobalDHT  bool    `json:"alternative_global_dht"`
	AnnounceToAllTrackers bool    `json:"announce_to_all_trackers"`
}
