package chain

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"
)

// DownloadChain 下载处理链
type DownloadChain struct {
	logger            *utils.Logger
	downloadRepo      *repository.DownloadRepository
	downloaderService *service.DownloaderService
	torrentService    *service.TorrentService
	siteService       *service.SiteService
	torrentClient     *service.TorrentClient
}

// NewDownloadChain 创建下载业务链实例
func NewDownloadChain(
	logger *utils.Logger,
	downloadRepo *repository.DownloadRepository,
	downloaderService *service.DownloaderService,
	torrentService *service.TorrentService,
	siteService *service.SiteService,
	torrentClient *service.TorrentClient,
) *DownloadChain {
	return &DownloadChain{
		logger:            logger,
		downloadRepo:      downloadRepo,
		downloaderService: downloaderService,
		torrentService:    torrentService,
		siteService:       siteService,
		torrentClient:     torrentClient,
	}
}

// DownloadTorrent 下载种子文件
func (d *DownloadChain) DownloadTorrent(torrentInfo *models.TorrentInfo, siteName string) (*models.DownloadResult, error) {
	d.logger.Info("开始下载种子", "title", torrentInfo.Title, "site", siteName)

	// 获取站点信息
	site, err := d.siteService.GetSiteByName(siteName)
	if err != nil {
		return nil, fmt.Errorf("获取站点信息失败: %v", err)
	}

	// 下载种子文件
	torrentData, err := d.downloadTorrentFile(torrentInfo.URL, site)
	if err != nil {
		d.logger.Error("种子文件下载失败", "error", err)
		return nil, err
	}

	// 解析种子文件
	torrent, err := d.torrentService.ParseTorrent(torrentData)
	if err != nil {
		d.logger.Error("种子文件解析失败", "error", err)
		return nil, err
	}

	// 保存下载记录
	downloadRecord := &models.DownloadRecord{
		TorrentID:    torrentInfo.ID,
		Title:        torrentInfo.Title,
		SiteName:     siteName,
		Size:         torrentInfo.Size,
		DownloadURL:  torrentInfo.URL,
		DownloadedAt: time.Now(),
		Status:       "completed",
	}

	err = d.downloadRepo.Create(downloadRecord)
	if err != nil {
		d.logger.Warn("保存下载记录失败", "error", err)
	}

	result := &models.DownloadResult{
		Success:     true,
		TorrentData: torrentData,
		TorrentInfo: torrent,
		Message:     "下载成功",
	}

	d.logger.Info("种子下载完成", "title", torrentInfo.Title, "size", len(torrentData))
	return result, nil
}

// AddToDownloader 添加种子到下载器
func (d *DownloadChain) AddToDownloader(torrentData []byte, downloadPath string, downloaderName string) (*models.DownloadTask, error) {
	d.logger.Info("添加种子到下载器", "downloader", downloaderName, "path", downloadPath)

	// 获取下载器
	downloader, err := d.downloaderService.GetDownloader(downloaderName)
	if err != nil {
		return nil, fmt.Errorf("获取下载器失败: %v", err)
	}

	// 添加下载任务
	task, err := downloader.AddTorrent(torrentData, downloadPath)
	if err != nil {
		d.logger.Error("添加下载任务失败", "error", err)
		return nil, err
	}

	// 保存任务记录
	downloadTask := &models.DownloadTask{
		TaskID:       task.ID,
		Title:        task.Name,
		Downloader:   downloaderName,
		DownloadPath: downloadPath,
		Status:       "added",
		AddedAt:      time.Now(),
		TorrentHash:  task.Hash,
	}

	err = d.downloadRepo.CreateTask(downloadTask)
	if err != nil {
		d.logger.Warn("保存下载任务记录失败", "error", err)
	}

	d.logger.Info("下载任务添加成功", "taskID", task.ID, "title", task.Name)
	return downloadTask, nil
}

// DownloadAndAdd 下载并添加种子到下载器
func (d *DownloadChain) DownloadAndAdd(torrentInfo *models.TorrentInfo, siteName string, downloadPath string, downloaderName string) (*models.CombinedDownloadResult, error) {
	d.logger.Info("下载并添加种子", "title", torrentInfo.Title, "site", siteName)

	// 下载种子
	downloadResult, err := d.DownloadTorrent(torrentInfo, siteName)
	if err != nil {
		return nil, err
	}

	// 添加到下载器
	task, err := d.AddToDownloader(downloadResult.TorrentData, downloadPath, downloaderName)
	if err != nil {
		return nil, err
	}

	result := &models.CombinedDownloadResult{
		Success:        true,
		DownloadResult: downloadResult,
		DownloadTask:   task,
		Message:        "下载和添加任务成功",
	}

	d.logger.Info("下载和添加任务完成", "title", torrentInfo.Title)
	return result, nil
}

// GetDownloadStatus 获取下载状态
func (d *DownloadChain) GetDownloadStatus(taskID string) (*models.DownloadStatus, error) {
	d.logger.Debug("获取下载状态", "taskID", taskID)

	// 从数据库获取任务信息
	task, err := d.downloadRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务信息失败: %v", err)
	}

	// 从下载器获取实时状态
	downloader, err := d.downloaderService.GetDownloader(task.Downloader)
	if err != nil {
		return nil, fmt.Errorf("获取下载器失败: %v", err)
	}

	downloadStatus, err := downloader.GetTaskStatus(taskID)
	if err != nil {
		return nil, fmt.Errorf("获取下载状态失败: %v", err)
	}

	status := &models.DownloadStatus{
		TaskID:        taskID,
		Title:         task.Title,
		Status:        downloadStatus.Status,
		Progress:      downloadStatus.Progress,
		DownloadSpeed: downloadStatus.DownloadSpeed,
		UploadSpeed:   downloadStatus.UploadSpeed,
		Size:          downloadStatus.Size,
		Downloaded:    downloadStatus.Downloaded,
		Uploaded:      downloadStatus.Uploaded,
		Ratio:         downloadStatus.Ratio,
		ETA:           downloadStatus.ETA,
		Peers:         downloadStatus.Peers,
		Seeds:         downloadStatus.Seeds,
	}

	d.logger.Debug("下载状态获取完成", "taskID", taskID, "status", status.Status)
	return status, nil
}

// PauseDownload 暂停下载任务
func (d *DownloadChain) PauseDownload(taskID string) error {
	d.logger.Info("暂停下载任务", "taskID", taskID)

	task, err := d.downloadRepo.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %v", err)
	}

	downloader, err := d.downloaderService.GetDownloader(task.Downloader)
	if err != nil {
		return fmt.Errorf("获取下载器失败: %v", err)
	}

	err = downloader.PauseTask(taskID)
	if err != nil {
		d.logger.Error("暂停下载任务失败", "error", err)
		return err
	}

	// 更新任务状态
	task.Status = "paused"
	err = d.downloadRepo.UpdateTask(task)
	if err != nil {
		d.logger.Warn("更新任务状态失败", "error", err)
	}

	d.logger.Info("下载任务已暂停", "taskID", taskID)
	return nil
}

// ResumeDownload 恢复下载任务
func (d *DownloadChain) ResumeDownload(taskID string) error {
	d.logger.Info("恢复下载任务", "taskID", taskID)

	task, err := d.downloadRepo.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %v", err)
	}

	downloader, err := d.downloaderService.GetDownloader(task.Downloader)
	if err != nil {
		return fmt.Errorf("获取下载器失败: %v", err)
	}

	err = downloader.ResumeTask(taskID)
	if err != nil {
		d.logger.Error("恢复下载任务失败", "error", err)
		return err
	}

	// 更新任务状态
	task.Status = "downloading"
	err = d.downloadRepo.UpdateTask(task)
	if err != nil {
		d.logger.Warn("更新任务状态失败", "error", err)
	}

	d.logger.Info("下载任务已恢复", "taskID", taskID)
	return nil
}

// DeleteDownload 删除下载任务
func (d *DownloadChain) DeleteDownload(taskID string, deleteFiles bool) error {
	d.logger.Info("删除下载任务", "taskID", taskID, "deleteFiles", deleteFiles)

	task, err := d.downloadRepo.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %v", err)
	}

	downloader, err := d.downloaderService.GetDownloader(task.Downloader)
	if err != nil {
		return fmt.Errorf("获取下载器失败: %v", err)
	}

	err = downloader.DeleteTask(taskID, deleteFiles)
	if err != nil {
		d.logger.Error("删除下载任务失败", "error", err)
		return err
	}

	// 删除任务记录
	err = d.downloadRepo.DeleteTask(taskID)
	if err != nil {
		d.logger.Warn("删除任务记录失败", "error", err)
	}

	d.logger.Info("下载任务已删除", "taskID", taskID)
	return nil
}

// GetActiveDownloads 获取活跃下载任务
func (d *DownloadChain) GetActiveDownloads() ([]*models.DownloadStatus, error) {
	d.logger.Debug("获取活跃下载任务")

	tasks, err := d.downloadRepo.GetActiveTasks()
	if err != nil {
		return nil, fmt.Errorf("获取活跃任务失败: %v", err)
	}

	var statuses []*models.DownloadStatus
	for _, task := range tasks {
		status, err := d.GetDownloadStatus(task.TaskID)
		if err != nil {
			d.logger.Warn("获取任务状态失败", "taskID", task.TaskID, "error", err)
			continue
		}
		statuses = append(statuses, status)
	}

	d.logger.Debug("活跃下载任务获取完成", "count", len(statuses))
	return statuses, nil
}

// downloadTorrentFile 下载种子文件
func (d *DownloadChain) downloadTorrentFile(url string, site *models.Site) ([]byte, error) {
	d.logger.Debug("下载种子文件", "url", url)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加站点Cookie
	if site != nil && site.Cookie != "" {
		req.Header.Set("Cookie", site.Cookie)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %s", resp.Status)
	}

	// 读取响应内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 验证是否是有效的种子文件
	if !d.isValidTorrentFile(data) {
		return nil, fmt.Errorf("无效的种子文件")
	}

	d.logger.Debug("种子文件下载成功", "size", len(data))
	return data, nil
}

// isValidTorrentFile 验证是否是有效的种子文件
func (d *DownloadChain) isValidTorrentFile(data []byte) bool {
	// 检查文件头信息
	if len(data) < 16 {
		return false
	}

	// 检查是否为有效的torrent文件格式
	if string(data[:11]) != "d8:announce" && string(data[:13]) != "d13:announce-list" {
		return false
	}

	return true
}
