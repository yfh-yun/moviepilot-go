package chain

import (
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repository"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/internal/utils"
)

// TorrentsChain 种子处理链
type TorrentsChain struct {
	logger         *utils.Logger
	torrentsRepo   *repository.TorrentsRepository
	siteService    *service.SiteService
	torrentService *service.TorrentService
	downloadChain  *DownloadChain
	mutex          sync.RWMutex
}

// NewTorrentsChain 创建种子业务链实例
func NewTorrentsChain(
	logger *utils.Logger,
	torrentsRepo *repository.TorrentsRepository,
	siteService *service.SiteService,
	torrentService *service.TorrentService,
	downloadChain *DownloadChain,
) *TorrentsChain {
	return &TorrentsChain{
		logger:         logger,
		torrentsRepo:   torrentsRepo,
		siteService:    siteService,
		torrentService: torrentService,
		downloadChain:  downloadChain,
	}
}

// RefreshTorrents 刷新种子列表
func (t *TorrentsChain) RefreshTorrents(siteNames []string) (*models.TorrentsRefreshResult, error) {
	t.logger.Info("开始刷新种子列表", "sites", siteNames)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 获取要刷新的站点
	sites, err := t.getRefreshSites(siteNames)
	if err != nil {
		return nil, fmt.Errorf("获取刷新站点失败: %v", err)
	}

	if len(sites) == 0 {
		t.logger.Info("没有可刷新的站点")
		return &models.TorrentsRefreshResult{
			TotalSites:   0,
			SuccessSites: 0,
			ErrorSites:   0,
			NewTorrents:  0,
		}, nil
	}

	var successSites int
	var errorSites int
	var newTorrents int
	var errors []error

	// 并行刷新所有站点
	var wg sync.WaitGroup
	resultChan := make(chan *models.SiteRefreshResult, len(sites))
	errorChan := make(chan error, len(sites))

	for _, site := range sites {
		wg.Add(1)
		go func(s *models.Site) {
			defer wg.Done()
			t.refreshSingleSite(s, resultChan, errorChan)
		}(site)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	// 收集结果
	var siteResults []*models.SiteRefreshResult

	for result := range resultChan {
		siteResults = append(siteResults, result)
		successSites++
		newTorrents += result.NewTorrents
	}

	for err := range errorChan {
		errors = append(errors, err)
		errorSites++
	}

	result := &models.TorrentsRefreshResult{
		TotalSites:   len(sites),
		SuccessSites: successSites,
		ErrorSites:   errorSites,
		NewTorrents:  newTorrents,
		SiteResults:  siteResults,
		Errors:       errors,
		RefreshTime:  time.Now(),
	}

	t.logger.Info("种子列表刷新完成",
		"total", len(sites),
		"success", successSites,
		"errors", errorSites,
		"new", newTorrents)

	return result, nil
}

// GetTorrents 获取种子列表
func (t *TorrentsChain) GetTorrents(siteName string, page int, pageSize int, filters map[string]interface{}) (*models.TorrentsListResult, error) {
	t.logger.Debug("获取种子列表", "site", siteName, "page", page, "pageSize", pageSize)

	torrents, total, err := t.torrentsRepo.GetTorrents(siteName, page, pageSize, filters)
	if err != nil {
		t.logger.Error("获取种子列表失败", "error", err)
		return nil, err
	}

	result := &models.TorrentsListResult{
		Torrents: torrents,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	t.logger.Debug("种子列表获取完成", "site", siteName, "count", len(torrents), "total", total)
	return result, nil
}

// DownloadTorrent 下载种子
func (t *TorrentsChain) DownloadTorrent(torrentID string, downloadPath string, downloaderName string) (*models.DownloadResult, error) {
	t.logger.Info("下载种子", "torrentID", torrentID, "downloader", downloaderName)

	// 获取种子信息
	torrentInfo, err := t.torrentsRepo.GetTorrentByID(torrentID)
	if err != nil {
		t.logger.Error("获取种子信息失败", "error", err)
		return nil, err
	}

	if torrentInfo == nil {
		return nil, fmt.Errorf("种子不存在")
	}

	// 调用下载链下载种子
	result, err := t.downloadChain.DownloadAndAdd(torrentInfo, torrentInfo.SiteName, downloadPath, downloaderName)
	if err != nil {
		t.logger.Error("下载种子失败", "error", err)
		return nil, err
	}

	// 更新种子状态为已下载
	torrentInfo.Downloaded = true
	torrentInfo.DownloadedAt = time.Now()
	err = t.torrentsRepo.UpdateTorrent(torrentInfo)
	if err != nil {
		t.logger.Warn("更新种子状态失败", "error", err)
	}

	t.logger.Info("种子下载成功", "title", torrentInfo.Title, "torrentID", torrentID)
	return result.DownloadResult, nil
}

// GetTorrentDetail 获取种子详情
func (t *TorrentsChain) GetTorrentDetail(torrentID string) (*models.TorrentDetail, error) {
	t.logger.Debug("获取种子详情", "torrentID", torrentID)

	torrentInfo, err := t.torrentsRepo.GetTorrentByID(torrentID)
	if err != nil {
		t.logger.Error("获取种子详情失败", "error", err)
		return nil, err
	}

	if torrentInfo == nil {
		return nil, fmt.Errorf("种子不存在")
	}

	// 获取种子文件列表（如果有）
	files, err := t.torrentService.GetTorrentFiles(torrentInfo)
	if err != nil {
		t.logger.Warn("获取种子文件列表失败", "error", err)
	}

	detail := &models.TorrentDetail{
		TorrentInfo: torrentInfo,
		Files:       files,
	}

	t.logger.Debug("种子详情获取完成", "title", torrentInfo.Title)
	return detail, nil
}

// SearchTorrents 搜索种子
func (t *TorrentsChain) SearchTorrents(query string, siteName string, filters map[string]interface{}) ([]*models.TorrentInfo, error) {
	t.logger.Info("搜索种子", "query", query, "site", siteName)

	torrents, err := t.torrentsRepo.SearchTorrents(query, siteName, filters)
	if err != nil {
		t.logger.Error("搜索种子失败", "error", err)
		return nil, err
	}

	t.logger.Info("种子搜索完成", "query", query, "count", len(torrents))
	return torrents, nil
}

// GetStatistics 获取种子统计信息
func (t *TorrentsChain) GetStatistics() (*models.TorrentsStatistics, error) {
	t.logger.Debug("获取种子统计信息")

	stats, err := t.torrentsRepo.GetStatistics()
	if err != nil {
		t.logger.Error("获取种子统计失败", "error", err)
		return nil, err
	}

	t.logger.Debug("种子统计获取完成", "total", stats.TotalTorrents)
	return stats, nil
}

// CleanupOldTorrents 清理旧种子
func (t *TorrentsChain) CleanupOldTorrents(days int) (*models.CleanupResult, error) {
	t.logger.Info("清理旧种子", "days", days)

	deletedCount, err := t.torrentsRepo.DeleteOldTorrents(days)
	if err != nil {
		t.logger.Error("清理旧种子失败", "error", err)
		return nil, err
	}

	result := &models.CleanupResult{
		DeletedCount: deletedCount,
		CleanedAt:    time.Now(),
	}

	t.logger.Info("旧种子清理完成", "deleted", deletedCount)
	return result, nil
}

// 内部辅助方法

// getRefreshSites 获取刷新站点
func (t *TorrentsChain) getRefreshSites(siteNames []string) ([]*models.Site, error) {
	if len(siteNames) == 0 {
		// 获取所有启用的站点
		return t.siteService.GetEnabledSites()
	}

	var sites []*models.Site
	for _, siteName := range siteNames {
		site, err := t.siteService.GetSiteByName(siteName)
		if err != nil {
			t.logger.Warn("获取站点失败", "site", siteName, "error", err)
			continue
		}
		if site.Enabled {
			sites = append(sites, site)
		}
	}

	if len(sites) == 0 {
		return nil, fmt.Errorf("没有可用的刷新站点")
	}

	return sites, nil
}

// refreshSingleSite 刷新单个站点
func (t *TorrentsChain) refreshSingleSite(site *models.Site, resultChan chan<- *models.SiteRefreshResult, errorChan chan<- error) {
	t.logger.Debug("刷新站点种子", "site", site.Name)

	siteResult := &models.SiteRefreshResult{
		SiteName: site.Name,
	}

	// 从站点获取最新的种子列表
	newTorrents, err := t.torrentService.GetSiteTorrents(site)
	if err != nil {
		t.logger.Warn("获取站点种子失败", "site", site.Name, "error", err)
		errorChan <- fmt.Errorf("%s: %v", site.Name, err)
		return
	}

	// 保存新种子到数据库
	for _, torrent := range newTorrents {
		torrent.SiteName = site.Name
		torrent.CreatedAt = time.Now()

		existing, err := t.torrentsRepo.GetTorrentByHash(torrent.Hash)
		if err != nil {
			t.logger.Warn("检查种子是否存在失败", "hash", torrent.Hash, "error", err)
			continue
		}

		if existing == nil {
			// 新种子
			err = t.torrentsRepo.CreateTorrent(torrent)
			if err != nil {
				t.logger.Warn("保存新种子失败", "title", torrent.Title, "error", err)
			} else {
				siteResult.NewTorrents++
			}
		} else {
			// 更新现有种子信息
			existing.Seeders = torrent.Seeders
			existing.Leechers = torrent.Leechers
			existing.UpdatedAt = time.Now()
			err = t.torrentsRepo.UpdateTorrent(existing)
			if err != nil {
				t.logger.Warn("更新种子信息失败", "title", torrent.Title, "error", err)
			}
		}
	}

	siteResult.TotalTorrents = len(newTorrents)
	resultChan <- siteResult

	t.logger.Debug("站点种子刷新完成", "site", site.Name, "new", siteResult.NewTorrents, "total", siteResult.TotalTorrents)
}
