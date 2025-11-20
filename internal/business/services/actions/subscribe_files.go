package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/config"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/context"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/events"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/meta"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/meta"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/chain"
	"go.uber.org/zap"
)

// SubscribeFilesHandler 订阅文件处理器
type SubscribeFilesHandler struct {
	rwLock             sync.RWMutex
	lockTimeout        time.Duration
	downloadChain      *chain.DownloadChain
	mediaChain         *chain.MediaChain
	searchChain        *chain.SearchChain
	tmdbChain          *chain.TmdbChain
	torrentsChain      *chain.TorrentsChain
	subscribeOper      *operations.SubscribeOper
	downloadHistoryOper *operations.DownloadHistoryOper
	siteOper           *operations.SiteOper
	systemConfigOper   *operations.SystemConfigOper
	subscribeHelper    *subscribe.SubscribeHelper
	torrentHelper      *torrent.TorrentHelper
	wordsMatcher       *meta.WordsMatcher
	eventManager       *event.EventManager
	logger             *zap.Logger
	config             *config.Config
}

// NewSubscribeFilesHandler 创建订阅文件处理器
func NewSubscribeFilesHandler(
	downloadChain *chain.DownloadChain,
	mediaChain *chain.MediaChain,
	searchChain *chain.SearchChain,
	tmdbChain *chain.TmdbChain,
	torrentsChain *chain.TorrentsChain,
	subscribeOper *operations.SubscribeOper,
	downloadHistoryOper *operations.DownloadHistoryOper,
	siteOper *operations.SiteOper,
	systemConfigOper *operations.SystemConfigOper,
	subscribeHelper *subscribe.SubscribeHelper,
	torrentHelper *torrent.TorrentHelper,
	wordsMatcher *meta.WordsMatcher,
	eventManager *event.EventManager,
	config *config.Config,
) *SubscribeFilesHandler {
	return &SubscribeFilesHandler{
		lockTimeout:        2 * time.Hour,
		downloadChain:      downloadChain,
		mediaChain:         mediaChain,
		searchChain:        searchChain,
		tmdbChain:          tmdbChain,
		torrentsChain:      torrentsChain,
		subscribeOper:      subscribeOper,
		downloadHistoryOper: downloadHistoryOper,
		siteOper:           siteOper,
		systemConfigOper:   systemConfigOper,
		subscribeHelper:    subscribeHelper,
		torrentHelper:      torrentHelper,
		wordsMatcher:       wordsMatcher,
		eventManager:       eventManager,
		logger:             logger.Global,
		config:             config,
	}
}

// SubscribeProcessOptions 订阅处理选项
type SubscribeProcessOptions struct {
	SubscribeID    string                    `json:"subscribe_id"`
	MediaID        string                    `json:"media_id"`
	UserName       string                    `json:"user_name"`
	Force          bool                      `json:"force"`
	CustomOptions  map[string]interface{}     `json:"custom_options,omitempty"`
}

// SubscribeProcessResult 订阅处理结果
type SubscribeProcessResult struct {
	Success        bool                      `json:"success"`
	Message        string                    `json:"message"`
	MediaInfo      *context.MediaInfo        `json:"media_info,omitempty"`
	TorrentInfo    *context.TorrentInfo      `json:"torrent_info,omitempty"`
	Downloaded     bool                      `json:"downloaded"`
	Subscribe      *models.Subscribe         `json:"subscribe,omitempty"`
	ProcessedItems []map[string]interface{}   `json:"processed_items,omitempty"`
	Errors         []string                  `json:"errors,omitempty"`
}

// ProcessSubscribeFile 处理订阅文件
func (h *SubscribeFilesHandler) ProcessSubscribeFile(ctx context.Context, opts *SubscribeProcessOptions) (*SubscribeProcessResult, error) {
	h.rwLock.Lock()
	defer h.rwLock.Unlock()

	// 设置锁超时定时器
	timeoutTimer := time.NewTimer(h.lockTimeout)
	defer timeoutTimer.Stop()

	result := &SubscribeProcessResult{
		Success:        true,
		ProcessedItems: make([]map[string]interface{}, 0),
		Errors:         make([]string, 0),
	}

	// 获取订阅信息
	subscribe, err := h.getSubscribeInfo(ctx, opts.SubscribeID)
	if err != nil {
		return nil, fmt.Errorf("获取订阅信息失败: %w", err)
	}
	result.Subscribe = subscribe

	// 解析媒体信息
	mediaInfo, err := h.parseMediaInfo(ctx, opts.MediaID, subscribe.MediaType)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}
	result.MediaInfo = mediaInfo

	// 生成搜索关键词
	searchKeywords := h.generateSearchKeywords(mediaInfo, subscribe)

	// 执行搜索
	searchResults, err := h.searchChain.Search(ctx, &chain.SearchRequest{
		Keywords:       searchKeywords,
		MediaType:      subscribe.MediaType,
		Quality:        subscribe.Quality,
		Resolution:     subscribe.Resolution,
		Effect:         subscribe.Effect,
		PrioritySites:  subscribe.PrioritySites,
		FilterRules:    subscribe.FilterRules,
		UserAgent:      opts.UserName,
	})
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("搜索失败: %v", err))
		return result, nil
	}

	// 处理搜索结果
	for _, torrentInfo := range searchResults.Torrents {
		torrentResult, err := h.processTorrent(ctx, torrentInfo, mediaInfo, subscribe, opts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("处理种子 %s 失败: %v", torrentInfo.Name, err))
			continue
		}

		if torrentResult.Downloaded {
			result.Downloaded = true
			result.TorrentInfo = torrentResult
			break
		}

		result.ProcessedItems = append(result.ProcessedItems, map[string]interface{}{
			"torrent_name": torrentInfo.Name,
			"site":         torrentInfo.SiteName,
			"size":         torrentInfo.Size,
			"processed":    false,
			"reason":       torrentResult.Message,
		})
	}

	return result, nil
}

// getSubscribeInfo 获取订阅信息
func (h *SubscribeFilesHandler) getSubscribeInfo(ctx context.Context, subscribeID string) (*models.Subscribe, error) {
	subscribe, err := h.subscribeOper.GetByID(ctx, subscribeID)
	if err != nil {
		return nil, err
	}
	if subscribe == nil {
		return nil, fmt.Errorf("订阅不存在: %s", subscribeID)
	}
	return subscribe, nil
}

// parseMediaInfo 解析媒体信息
func (h *SubscribeFilesHandler) parseMediaInfo(ctx context.Context, mediaID string, mediaType string) (*context.MediaInfo, error) {
	// 通过事件广播解析媒体信息
	eventData := &event.MediaRecognizeConvertEventData{
		MediaID:     mediaID,
		ConvertType: h.config.Recognize.Source,
	}

	response := h.eventManager.SendEvent(event.EventTypeMediaRecognizeConvert, eventData)
	if response != nil && response.EventData != nil {
		if data, ok := response.EventData.(*event.MediaRecognizeConvertEventData); ok && data.MediaDict != nil {
			newID, ok := data.MediaDict["id"].(string)
			if ok {
				if data.ConvertType == "themoviedb" {
					return h.mediaChain.RecognizeMedia(&context.MediaRecognizeRequest{
						Meta:    &meta.MetaBase{},
						TMDBID:  newID,
					})
				} else if data.ConvertType == "douban" {
					return h.mediaChain.RecognizeMedia(&context.MediaRecognizeRequest{
						Meta:     &meta.MetaBase{},
						DoubanID: newID,
					})
				}
			}
		}
	}

	return nil, fmt.Errorf("无法解析媒体信息: %s", mediaID)
}

// generateSearchKeywords 生成搜索关键词
func (h *SubscribeFilesHandler) generateSearchKeywords(mediaInfo *context.MediaInfo, subscribe *models.Subscribe) []string {
	var keywords []string

	// 基础关键词
	baseKeyword := mediaInfo.Title
	if mediaInfo.Year > 0 {
		baseKeyword = fmt.Sprintf("%s %d", baseKeyword, mediaInfo.Year)
	}

	// 季集信息
	if mediaInfo.Season > 0 {
		baseKeyword += fmt.Sprintf(" S%02d", mediaInfo.Season)
	}
	if mediaInfo.Episode > 0 {
		baseKeyword += fmt.Sprintf(" E%02d", mediaInfo.Episode)
	}

	keywords = append(keywords, baseKeyword)

	// 添加别名
	if len(mediaInfo.Alias) > 0 {
		for _, alias := range mediaInfo.Alias {
			aliasKeyword := alias
			if mediaInfo.Year > 0 {
				aliasKeyword = fmt.Sprintf("%s %d", aliasKeyword, mediaInfo.Year)
			}
			keywords = append(keywords, aliasKeyword)
		}
	}

	// 添加自定义关键词
	if subscribe.SearchKeywords != "" {
		customKeywords := strings.Split(subscribe.SearchKeywords, ",")
		keywords = append(keywords, customKeywords...)
	}

	return keywords
}

// processTorrent 处理种子
func (h *SubscribeFilesHandler) processTorrent(ctx context.Context, torrentInfo *context.TorrentInfo, mediaInfo *context.MediaInfo, subscribe *models.Subscribe, opts *SubscribeProcessOptions) (*SubscribeProcessResult, error) {
	result := &SubscribeProcessResult{Success: true}

	// 验证种子是否匹配
	if !h.validateTorrentMatch(torrentInfo, mediaInfo, subscribe) {
		result.Message = "种子与订阅不匹配"
		return result, nil
	}

	// 检查是否已下载
	if h.isAlreadyDownloaded(ctx, torrentInfo) {
		result.Message = "种子已下载"
		return result, nil
	}

	// 执行下载
	downloadResult, err := h.downloadChain.Download(ctx, &chain.DownloadRequest{
		TorrentInfo: torrentInfo,
		MediaInfo:   mediaInfo,
		Subscribe:   subscribe,
		UserAgent:   opts.UserName,
	})
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("下载失败: %v", err)
		return result, nil
	}

	if downloadResult.Success {
		result.Downloaded = true
		result.Message = "下载成功"

		// 更新订阅状态
		err = h.updateSubscribeStatus(ctx, subscribe, torrentInfo)
		if err != nil {
			h.logger.Warn("更新订阅状态失败", zap.Error(err))
		}

		// 记录下载历史
		err = h.recordDownloadHistory(ctx, torrentInfo, mediaInfo, subscribe)
		if err != nil {
			h.logger.Warn("记录下载历史失败", zap.Error(err))
		}
	} else {
		result.Message = downloadResult.Message
	}

	return result, nil
}

// validateTorrentMatch 验证种子是否匹配
func (h *SubscribeFilesHandler) validateTorrentMatch(torrentInfo *context.TorrentInfo, mediaInfo *context.MediaInfo, subscribe *models.Subscribe) bool {
	// 使用词汇匹配器验证
	metaInfo := &metainfo.MetaInfo{}
	metaInfo.FromTitle(torrentInfo.Name)

	// 检查标题匹配
	if !h.wordsMatcher.MatchTitle(metaInfo.Title, mediaInfo.Title) {
		return false
	}

	// 检查年份匹配
	if mediaInfo.Year > 0 && metaInfo.Year != mediaInfo.Year {
		return false
	}

	// 检查季集匹配
	if mediaInfo.Season > 0 && metaInfo.Season != mediaInfo.Season {
		return false
	}

	if mediaInfo.Episode > 0 && metaInfo.Episode != mediaInfo.Episode {
		return false
	}

	// 检查质量匹配
	if subscribe.Quality != "" && !h.matchQuality(torrentInfo.Name, subscribe.Quality) {
		return false
	}

	// 检查分辨率匹配
	if subscribe.Resolution != "" && !h.matchResolution(torrentInfo.Name, subscribe.Resolution) {
		return false
	}

	return true
}

// matchQuality 匹配质量
func (h *SubscribeFilesHandler) matchQuality(title, quality string) bool {
	qualityPattern := regexp.MustCompile(fmt.Sprintf(`(?i)%s`, regexp.QuoteMeta(quality)))
	return qualityPattern.MatchString(title)
}

// matchResolution 匹配分辨率
func (h *SubscribeFilesHandler) matchResolution(title, resolution string) bool {
	resolutionPattern := regexp.MustCompile(fmt.Sprintf(`(?i)%s`, regexp.QuoteMeta(resolution)))
	return resolutionPattern.MatchString(title)
}

// isAlreadyDownloaded 检查是否已下载
func (h *SubscribeFilesHandler) isAlreadyDownloaded(ctx context.Context, torrentInfo *context.TorrentInfo) bool {
	history, err := h.downloadHistoryOper.GetByTorrentHash(ctx, torrentInfo.Hash)
	if err != nil {
		return false
	}
	return history != nil
}

// updateSubscribeStatus 更新订阅状态
func (h *SubscribeFilesHandler) updateSubscribeStatus(ctx context.Context, subscribe *models.Subscribe, torrentInfo *context.TorrentInfo) error {
	now := time.Now()
	
	updates := map[string]interface{}{
		"last_update":   now,
		"last_match":    torrentInfo.Name,
		"download_size": torrentInfo.Size,
	}

	if subscribe.Complete {
		updates["state"] = "completed"
	}

	return h.subscribeOper.Update(ctx, subscribe.ID, updates)
}

// recordDownloadHistory 记录下载历史
func (h *SubscribeFilesHandler) recordDownloadHistory(ctx context.Context, torrentInfo *context.TorrentInfo, mediaInfo *context.MediaInfo, subscribe *models.Subscribe) error {
	history := &models.DownloadHistory{
		TorrentHash:    torrentInfo.Hash,
		TorrentName:    torrentInfo.Name,
		SiteName:       torrentInfo.SiteName,
		MediaID:        mediaInfo.ID,
		MediaTitle:     mediaInfo.Title,
		MediaYear:      mediaInfo.Year,
		MediaSeason:    mediaInfo.Season,
		MediaEpisode:   mediaInfo.Episode,
		SubscribeID:    subscribe.ID,
		DownloadTime:   time.Now(),
		FileSize:       torrentInfo.Size,
	}

	return h.downloadHistoryOper.Create(ctx, history)
}

// BatchProcessSubscribeFiles 批量处理订阅文件
func (h *SubscribeFilesHandler) BatchProcessSubscribeFiles(ctx context.Context, optsList []*SubscribeProcessOptions) ([]*SubscribeProcessResult, error) {
	var results []*SubscribeProcessResult
	var wg sync.WaitGroup
	resultChan := make(chan *SubscribeProcessResult, len(optsList))

	// 并发处理
	for _, opts := range optsList {
		wg.Add(1)
		go func(opt *SubscribeProcessOptions) {
			defer wg.Done()
			result, err := h.ProcessSubscribeFile(ctx, opt)
			if err != nil {
				result = &SubscribeProcessResult{
					Success: false,
					Message: err.Error(),
					Errors:  []string{err.Error()},
				}
			}
			resultChan <- result
		}(opts)
	}

	// 等待完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

// GetSubscribeStatus 获取订阅状态
func (h *SubscribeFilesHandler) GetSubscribeStatus(ctx context.Context, subscribeID string) (*SubscribeStatus, error) {
	subscribe, err := h.subscribeOper.GetByID(ctx, subscribeID)
	if err != nil {
		return nil, err
	}
	if subscribe == nil {
		return nil, fmt.Errorf("订阅不存在: %s", subscribeID)
	}

	// 获取下载历史
	histories, err := h.downloadHistoryOper.GetBySubscribeID(ctx, subscribeID, 10)
	if err != nil {
		return nil, err
	}

	return &SubscribeStatus{
		Subscribe:       subscribe,
		DownloadHistory: histories,
		IsActive:       !subscribe.Complete,
		LastUpdate:     subscribe.LastUpdate,
		TotalDownloads:  len(histories),
	}, nil
}

// SubscribeStatus 订阅状态
type SubscribeStatus struct {
	Subscribe       *models.Subscribe           `json:"subscribe"`
	DownloadHistory []*models.DownloadHistory    `json:"download_history"`
	IsActive       bool                        `json:"is_active"`
	LastUpdate     time.Time                   `json:"last_update"`
	TotalDownloads  int                         `json:"total_downloads"`
}