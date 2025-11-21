package chain

import (
	"fmt"
	"sync"
	"time"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/utils"
)

// SubscribeChain 订阅处理链
type SubscribeChain struct {
	logger        *utils.Logger
	subscribeRepo *repository.SubscribeRepository
	mediaService  *service.MediaService
	downloadChain *DownloadChain
	searchChain   *SearchChain
	notifyService *service.NotifyService
	mutex         sync.RWMutex
}

// NewSubscribeChain 创建订阅业务链实例
func NewSubscribeChain(
	logger *utils.Logger,
	subscribeRepo *repository.SubscribeRepository,
	mediaService *service.MediaService,
	downloadChain *DownloadChain,
	searchChain *SearchChain,
	notifyService *service.NotifyService,
) *SubscribeChain {
	return &SubscribeChain{
		logger:        logger,
		subscribeRepo: subscribeRepo,
		mediaService:  mediaService,
		downloadChain: downloadChain,
		searchChain:   searchChain,
		notifyService: notifyService,
	}
}

// AddSubscribe 添加订阅
func (s *SubscribeChain) AddSubscribe(mediaInfo *models.MediaInfo, options *models.SubscribeOptions) (*models.Subscribe, error) {
	s.logger.Info("添加订阅", "title", mediaInfo.Title, "type", mediaInfo.MediaType)

	// 检查是否已存在相同订阅
	existing, err := s.subscribeRepo.GetByMediaID(mediaInfo.MediaID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("该媒体已订阅")
	}

	// 创建订阅记录
	subscribe := &models.Subscribe{
		MediaID:     mediaInfo.MediaID,
		Title:       mediaInfo.Title,
		MediaType:   mediaInfo.MediaType,
		Year:        mediaInfo.Year,
		TMDBID:      mediaInfo.TMDBID,
		Poster:      mediaInfo.Poster,
		Overview:    mediaInfo.Overview,
		Options:     options,
		Status:      "active",
		CreatedAt:   time.Now(),
		LastCheckAt: time.Now(),
	}

	// 保存订阅
	err = s.subscribeRepo.Create(subscribe)
	if err != nil {
		s.logger.Error("保存订阅失败", "error", err)
		return nil, err
	}

	s.logger.Info("订阅添加成功", "title", mediaInfo.Title, "subscribeID", subscribe.ID)
	return subscribe, nil
}

// RemoveSubscribe 移除订阅
func (s *SubscribeChain) RemoveSubscribe(subscribeID string) error {
	s.logger.Info("移除订阅", "subscribeID", subscribeID)

	// 获取订阅信息
	subscribe, err := s.subscribeRepo.GetByID(subscribeID)
	if err != nil {
		return fmt.Errorf("获取订阅信息失败: %v", err)
	}

	if subscribe == nil {
		return fmt.Errorf("订阅不存在")
	}

	// 删除订阅
	err = s.subscribeRepo.Delete(subscribeID)
	if err != nil {
		s.logger.Error("删除订阅失败", "error", err)
		return err
	}

	s.logger.Info("订阅已移除", "title", subscribe.Title)
	return nil
}

// UpdateSubscribe 更新订阅
func (s *SubscribeChain) UpdateSubscribe(subscribeID string, options *models.SubscribeOptions) error {
	s.logger.Info("更新订阅", "subscribeID", subscribeID)

	// 获取订阅信息
	subscribe, err := s.subscribeRepo.GetByID(subscribeID)
	if err != nil {
		return fmt.Errorf("获取订阅信息失败: %v", err)
	}

	if subscribe == nil {
		return fmt.Errorf("订阅不存在")
	}

	// 更新订阅选项
	subscribe.Options = options
	subscribe.UpdatedAt = time.Now()

	err = s.subscribeRepo.Update(subscribe)
	if err != nil {
		s.logger.Error("更新订阅失败", "error", err)
		return err
	}

	s.logger.Info("订阅已更新", "title", subscribe.Title)
	return nil
}

// CheckSubscribes 检查所有订阅
func (s *SubscribeChain) CheckSubscribes() (*models.SubscribeCheckResult, error) {
	s.logger.Info("开始检查所有订阅")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取所有活跃订阅
	subscribes, err := s.subscribeRepo.GetActiveSubscribes()
	if err != nil {
		s.logger.Error("获取活跃订阅失败", "error", err)
		return nil, err
	}

	if len(subscribes) == 0 {
		s.logger.Info("没有活跃订阅需要检查")
		return &models.SubscribeCheckResult{
			TotalCount:   0,
			FoundCount:   0,
			ErrorCount:   0,
			NewDownloads: 0,
		}, nil
	}

	var foundCount int
	var errorCount int
	var newDownloads int
	var errors []error

	// 并行检查每个订阅
	var wg sync.WaitGroup
	resultChan := make(chan *models.SubscribeCheckItem, len(subscribes))
	errorChan := make(chan error, len(subscribes))

	for _, subscribe := range subscribes {
		wg.Add(1)
		go func(sub *models.Subscribe) {
			defer wg.Done()
			s.checkSingleSubscribe(sub, resultChan, errorChan)
		}(subscribe)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	// 收集结果
	var checkItems []*models.SubscribeCheckItem

	for result := range resultChan {
		checkItems = append(checkItems, result)
		if result.Found {
			foundCount++
		}
		if result.Downloaded {
			newDownloads++
		}
	}

	for err := range errorChan {
		errors = append(errors, err)
		errorCount++
	}

	result := &models.SubscribeCheckResult{
		TotalCount:   len(subscribes),
		FoundCount:   foundCount,
		ErrorCount:   errorCount,
		NewDownloads: newDownloads,
		CheckItems:   checkItems,
		Errors:       errors,
		CheckTime:    time.Now(),
	}

	s.logger.Info("订阅检查完成",
		"total", len(subscribes),
		"found", foundCount,
		"downloads", newDownloads,
		"errors", errorCount)

	return result, nil
}

// GetSubscribes 获取订阅列表
func (s *SubscribeChain) GetSubscribes(status string, page int, pageSize int) (*models.SubscribeListResult, error) {
	s.logger.Debug("获取订阅列表", "status", status, "page", page, "pageSize", pageSize)

	subscribes, total, err := s.subscribeRepo.GetSubscribes(status, page, pageSize)
	if err != nil {
		s.logger.Error("获取订阅列表失败", "error", err)
		return nil, err
	}

	result := &models.SubscribeListResult{
		Subscribes: subscribes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}

	s.logger.Debug("订阅列表获取完成", "count", len(subscribes), "total", total)
	return result, nil
}

// GetSubscribeDetail 获取订阅详情
func (s *SubscribeChain) GetSubscribeDetail(subscribeID string) (*models.SubscribeDetail, error) {
	s.logger.Debug("获取订阅详情", "subscribeID", subscribeID)

	subscribe, err := s.subscribeRepo.GetByID(subscribeID)
	if err != nil {
		s.logger.Error("获取订阅详情失败", "error", err)
		return nil, err
	}

	if subscribe == nil {
		return nil, fmt.Errorf("订阅不存在")
	}

	// 获取订阅历史记录
	history, err := s.subscribeRepo.GetHistoryBySubscribeID(subscribeID)
	if err != nil {
		s.logger.Warn("获取订阅历史失败", "error", err)
	}

	detail := &models.SubscribeDetail{
		Subscribe: subscribe,
		History:   history,
	}

	s.logger.Debug("订阅详情获取完成", "title", subscribe.Title)
	return detail, nil
}

// SearchAndSubscribe 搜索并订阅媒体
func (s *SubscribeChain) SearchAndSubscribe(query string, mediaType string, options *models.SubscribeOptions) (*models.SearchAndSubscribeResult, error) {
	s.logger.Info("搜索并订阅", "query", query, "type", mediaType)

	// 搜索媒体
	searchResult, err := s.searchChain.SearchByMedia(&models.MediaInfo{
		Title:     query,
		MediaType: mediaType,
	}, nil, true)
	if err != nil {
		return nil, fmt.Errorf("搜索媒体失败: %v", err)
	}

	if len(searchResult.Torrents) == 0 {
		return nil, fmt.Errorf("未找到匹配的媒体")
	}

	// 获取第一个匹配的媒体信息
	// 在实际应用中，这里应该让用户选择具体要订阅哪个媒体
	firstTorrent := searchResult.Torrents[0]

	// 创建媒体信息（简化版）
	mediaInfo := &models.MediaInfo{
		Title:     firstTorrent.Title,
		MediaType: mediaType,
		// 需要更完整的媒体信息，这里简化处理
	}

	// 添加订阅
	subscribe, err := s.AddSubscribe(mediaInfo, options)
	if err != nil {
		return nil, fmt.Errorf("添加订阅失败: %v", err)
	}

	result := &models.SearchAndSubscribeResult{
		SearchResult: searchResult,
		Subscribe:    subscribe,
	}

	s.logger.Info("搜索并订阅完成", "title", mediaInfo.Title)
	return result, nil
}

// 内部辅助方法

// checkSingleSubscribe 检查单个订阅
func (s *SubscribeChain) checkSingleSubscribe(subscribe *models.Subscribe, resultChan chan<- *models.SubscribeCheckItem, errorChan chan<- error) {
	s.logger.Debug("检查订阅", "title", subscribe.Title)

	checkItem := &models.SubscribeCheckItem{
		SubscribeID: subscribe.ID,
		Title:       subscribe.Title,
		MediaType:   subscribe.MediaType,
	}

	// 检查是否需要更新订阅信息
	if s.needUpdateMediaInfo(subscribe) {
		err := s.updateMediaInfo(subscribe)
		if err != nil {
			checkItem.Error = err.Error()
			errorChan <- err
		}
	}

	// 搜索匹配的种子
	searchResult, err := s.searchChain.SearchByMedia(&models.MediaInfo{
		Title:     subscribe.Title,
		MediaType: subscribe.MediaType,
		Year:      subscribe.Year,
		TMDBID:    subscribe.TMDBID,
	}, nil, false)

	if err != nil {
		checkItem.Error = err.Error()
		errorChan <- err
		resultChan <- checkItem
		return
	}

	if len(searchResult.Torrents) > 0 {
		checkItem.Found = true
		checkItem.TorrentCount = len(searchResult.Torrents)

		// 选择最佳的种子进行下载
		bestTorrent := s.selectBestTorrent(searchResult.Torrents, subscribe.Options)
		if bestTorrent != nil {
			// 下载种子
			downloadResult, err := s.downloadChain.DownloadAndAdd(bestTorrent, "", subscribe.Options.DownloadPath, subscribe.Options.Downloader)
			if err != nil {
				checkItem.Error = err.Error()
				errorChan <- err
			} else {
				checkItem.Downloaded = true
				checkItem.TorrentTitle = bestTorrent.Title

				// 记录下载历史
				s.recordDownloadHistory(subscribe, bestTorrent, downloadResult)

				// 发送通知
				s.sendDownloadNotification(subscribe, bestTorrent)
			}
		}
	}

	// 更新订阅的最后检查时间
	subscribe.LastCheckAt = time.Now()
	err = s.subscribeRepo.Update(subscribe)
	if err != nil {
		s.logger.Warn("更新订阅检查时间失败", "error", err)
	}

	resultChan <- checkItem
	s.logger.Debug("订阅检查完成", "title", subscribe.Title, "found", checkItem.Found, "downloaded", checkItem.Downloaded)
}

// needUpdateMediaInfo 检查是否需要更新媒体信息
func (s *SubscribeChain) needUpdateMediaInfo(subscribe *models.Subscribe) bool {
	// 如果媒体信息超过30天未更新，则需要更新
	return time.Since(subscribe.LastMediaUpdateAt) > 30*24*time.Hour
}

// updateMediaInfo 更新媒体信息
func (s *SubscribeChain) updateMediaInfo(subscribe *models.Subscribe) error {
	// 从TMDB或其他API获取最新的媒体信息
	// 这里简化处理，实际需要调用媒体服务

	subscribe.LastMediaUpdateAt = time.Now()
	return s.subscribeRepo.Update(subscribe)
}

// selectBestTorrent 选择最佳的种子
func (s *SubscribeChain) selectBestTorrent(torrents []*models.TorrentInfo, options *models.SubscribeOptions) *models.TorrentInfo {
	if len(torrents) == 0 {
		return nil
	}

	// 简单的选择逻辑：选择种子数量最多的
	var bestTorrent *models.TorrentInfo
	maxSeeders := 0

	for _, torrent := range torrents {
		if torrent.Seeders > maxSeeders {
			maxSeeders = torrent.Seeders
			bestTorrent = torrent
		}
	}

	return bestTorrent
}

// recordDownloadHistory 记录下载历史
func (s *SubscribeChain) recordDownloadHistory(subscribe *models.Subscribe, torrent *models.TorrentInfo, downloadResult *models.CombinedDownloadResult) {
	history := &models.SubscribeHistory{
		SubscribeID:  subscribe.ID,
		TorrentTitle: torrent.Title,
		TorrentHash:  torrent.Hash,
		Size:         torrent.Size,
		DownloadedAt: time.Now(),
		Status:       "downloaded",
	}

	err := s.subscribeRepo.SaveHistory(history)
	if err != nil {
		s.logger.Warn("保存订阅历史失败", "error", err)
	}
}

// sendDownloadNotification 发送下载通知
func (s *SubscribeChain) sendDownloadNotification(subscribe *models.Subscribe, torrent *models.TorrentInfo) {
	message := fmt.Sprintf("订阅的媒体【%s】已找到匹配资源并开始下载", subscribe.Title)

	err := s.notifyService.SendNotification("订阅下载通知", message, "info")
	if err != nil {
		s.logger.Warn("发送下载通知失败", "error", err)
	}
}
