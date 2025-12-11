package subscribe

import (
	"context"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	// LockTimeout 锁超时时间（2小时）
	LockTimeout = 2 * time.Hour
)

// Service 订阅服务
type Service struct {
	// Repositories
	subscribeRepo    SubscribeRepository
	systemConfigRepo SystemConfigRepository
	siteRepo         SiteRepository
	downloadHistRepo DownloadHistoryRepository

	// Services
	mediaService    MediaService
	searchService   SearchService
	downloadService DownloadService
	torrentsService TorrentsService
	filterService   FilterService
	eventService    EventService
	messageService  MessageService
	subscribeHelper SubscribeHelper
	torrentHelper   TorrentHelper
	wordsMatcher    WordsMatcher
	tmdbService     TmdbService

	// Concurrency
	searchLock sync.RWMutex
	matchLock  sync.RWMutex

	// Logger
	logger *zap.Logger
}

// NewService 创建订阅服务
func NewService(
	subscribeRepo SubscribeRepository,
	systemConfigRepo SystemConfigRepository,
	siteRepo SiteRepository,
	downloadHistRepo DownloadHistoryRepository,
	mediaService MediaService,
	searchService SearchService,
	downloadService DownloadService,
	torrentsService TorrentsService,
	filterService FilterService,
	eventService EventService,
	messageService MessageService,
	subscribeHelper SubscribeHelper,
	torrentHelper TorrentHelper,
	wordsMatcher WordsMatcher,
	tmdbService TmdbService,
	logger *zap.Logger,
) *Service {
	return &Service{
		subscribeRepo:    subscribeRepo,
		systemConfigRepo: systemConfigRepo,
		siteRepo:         siteRepo,
		downloadHistRepo: downloadHistRepo,
		mediaService:     mediaService,
		searchService:    searchService,
		downloadService:  downloadService,
		torrentsService:  torrentsService,
		filterService:    filterService,
		eventService:     eventService,
		messageService:   messageService,
		subscribeHelper:  subscribeHelper,
		torrentHelper:    torrentHelper,
		wordsMatcher:     wordsMatcher,
		tmdbService:      tmdbService,
		logger:           logger,
	}
}

// NewBasicService 提供一个只依赖 SubscribeRepository 和 Logger 的简化构造函数，
// 便于在尚未完成全部依赖注入时先使用基础订阅能力。
func NewBasicService(subscribeRepo SubscribeRepository, logger *zap.Logger) *Service {
	return &Service{
		subscribeRepo: subscribeRepo,
		logger:        logger,
	}
}

// Exists 判断订阅是否存在
// 对应Python: exists()
func (s *Service) Exists(ctx context.Context, mediaInfo *MediaInfo, meta *MetaInfo) (bool, error) {
	season := 0
	if meta != nil {
		season = meta.BeginSeason
	}

	return s.subscribeRepo.Exists(ctx, &mediaInfo.TMDBID, mediaInfo.DoubanID, season)
}

// GetStatesForSearch 获取搜索状态列表
// 对应Python: get_states_for_search()
// 状态说明:
//
//	N: New（新建，未处理）
//	R: Resolved（订阅中）
//	P: Pending（待定，信息待进一步更新，允许搜索，不允许完成）
//	S: Suspended（暂停，订阅不参与任何动作，暂时停止处理）
func (s *Service) GetStatesForSearch(state string) []string {
	// 如果状态是 R 或 P，则视为一起搜索
	if state == "R" || state == "P" {
		return []string{"R", "P"}
	}
	return []string{state}
}

// GetSubscribeSourceKeyword 获取订阅来源关键字
// 对应Python: get_subscribe_source_keyword()
func (s *Service) GetSubscribeSourceKeyword(subscribe *Subscribe) string {
	if subscribe.Keyword != "" {
		return subscribe.Keyword
	}
	return subscribe.Name
}

// GetParams 获取订阅过滤参数
// 对应Python: get_params()
func (s *Service) GetParams(ctx context.Context, subscribe *Subscribe) map[string]any {
	// 获取默认过滤规则
	defaultRule := make(map[string]any)
	if val, err := s.systemConfigRepo.Get(ctx, "SubscribeDefaultParams"); err == nil && val != nil {
		if rule, ok := val.(map[string]any); ok {
			defaultRule = rule
		}
	}

	params := make(map[string]any)

	// 应用订阅配置或默认配置
	if subscribe.Include != "" {
		params["include"] = subscribe.Include
	} else if v, ok := defaultRule["include"]; ok && v != nil {
		params["include"] = v
	}

	if subscribe.Exclude != "" {
		params["exclude"] = subscribe.Exclude
	} else if v, ok := defaultRule["exclude"]; ok && v != nil {
		params["exclude"] = v
	}

	if subscribe.Quality != "" {
		params["quality"] = subscribe.Quality
	} else if v, ok := defaultRule["quality"]; ok && v != nil {
		params["quality"] = v
	}

	if subscribe.Resolution != "" {
		params["resolution"] = subscribe.Resolution
	} else if v, ok := defaultRule["resolution"]; ok && v != nil {
		params["resolution"] = v
	}

	if subscribe.Effect != "" {
		params["effect"] = subscribe.Effect
	} else if v, ok := defaultRule["effect"]; ok && v != nil {
		params["effect"] = v
	}

	// 添加其他默认参数
	if v, ok := defaultRule["tv_size"]; ok && v != nil {
		params["tv_size"] = v
	}
	if v, ok := defaultRule["movie_size"]; ok && v != nil {
		params["movie_size"] = v
	}
	if v, ok := defaultRule["min_seeders"]; ok && v != nil {
		params["min_seeders"] = v
	}
	if v, ok := defaultRule["min_seeders_time"]; ok && v != nil {
		params["min_seeders_time"] = v
	}

	return params
}

// GetSubscribe 获取订阅详情
func (s *Service) GetSubscribe(ctx context.Context, id uint) (*Subscribe, error) {
	return s.subscribeRepo.Get(ctx, int(id))
}

// ListSubscribes 获取订阅列表
func (s *Service) ListSubscribes(ctx context.Context, opts ListOptions) ([]*Subscribe, int64, error) {
	// 构建状态过滤条件
	var states []string
	if opts.State != "" {
		states = s.GetStatesForSearch(opts.State)
	}

	// 获取订阅列表
	subscribes, err := s.subscribeRepo.List(ctx, states)
	if err != nil {
		return nil, 0, err
	}

	// 类型过滤
	if opts.Type != "" {
		filtered := make([]*Subscribe, 0)
		for _, sub := range subscribes {
			if sub.Type == opts.Type {
				filtered = append(filtered, sub)
			}
		}
		subscribes = filtered
	}

	// 分页处理
	total := int64(len(subscribes))
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 20
	}

	start := (opts.Page - 1) * opts.PageSize
	end := start + opts.PageSize

	if start >= len(subscribes) {
		return []*Subscribe{}, total, nil
	}
	if end > len(subscribes) {
		end = len(subscribes)
	}

	return subscribes[start:end], total, nil
}

// GetSubscribeByMediaID 根据媒体ID获取订阅
func (s *Service) GetSubscribeByMediaID(ctx context.Context, mediaID string, season *int, title string) (*Subscribe, error) {
	// 根据媒体ID类型查询订阅
	if len(mediaID) > 0 {
		if mediaID[:5] == "tmdb:" {
			// TMDB ID
			tmdbID, err := strconv.Atoi(mediaID[5:])
			if err != nil {
				return nil, err
			}
			return s.subscribeRepo.Get(ctx, tmdbID) // This needs to be updated to GetByTMDBID
		} else if mediaID[:7] == "douban:" {
			// Douban ID
			// This needs to be updated to GetByDoubanID
			return nil, nil
		} else if mediaID[:8] == "bangumi:" {
			// Bangumi ID
			_, err := strconv.Atoi(mediaID[8:])
			if err != nil {
				return nil, err
			}
			// This needs to be updated to GetByBangumiID
			return nil, nil
		}
	}
	return nil, nil
}

// RefreshSubscribes 刷新所有订阅
func (s *Service) RefreshSubscribes(ctx context.Context) error {
	// This would typically trigger a scheduled job
	// For now, we'll just return nil
	return nil
}

// CheckSubscribes 刷新订阅TMDB信息
func (s *Service) CheckSubscribes(ctx context.Context) error {
	// This would typically trigger a scheduled job to refresh TMDB info
	// For now, we'll just return nil
	return nil
}

// SearchAllSubscribes 搜索所有订阅
func (s *Service) SearchAllSubscribes(ctx context.Context) error {
	// This would typically trigger a background search job
	// For now, we'll just return nil
	return nil
}

// SearchSubscribe 搜索指定订阅
func (s *Service) SearchSubscribe(ctx context.Context, subscribeID uint) error {
	// This would typically trigger a background search job for a specific subscription
	// For now, we'll just return nil
	return nil
}

// ResetSubscribe 重置订阅
func (s *Service) ResetSubscribe(ctx context.Context, subscribeID uint) error {
	// Reset subscription by updating its state and other fields
	updates := map[string]any{
		"note":         "[]",
		"lack_episode": 0, // This should be set to total_episode
		"state":        "R",
	}
	return s.subscribeRepo.Update(ctx, int(subscribeID), updates)
}

// DeleteSubscribeByMediaID 根据媒体ID删除订阅
func (s *Service) DeleteSubscribeByMediaID(ctx context.Context, mediaID string, season *int) error {
	// This would typically delete a subscription by media ID
	// For now, we'll just return nil
	return nil
}

// GetPopularSubscribes 获取热门订阅
func (s *Service) GetPopularSubscribes(ctx context.Context, stype string, page, count int) ([]*Subscribe, error) {
	// This would typically get popular subscriptions
	// For now, we'll just return an empty list
	return []*Subscribe{}, nil
}

// GetUserSubscribes 获取用户订阅
func (s *Service) GetUserSubscribes(ctx context.Context, username string) ([]*Subscribe, error) {
	// This would typically get subscriptions for a specific user
	// For now, we'll just return an empty list
	return []*Subscribe{}, nil
}

// GetSubscribeFiles 获取订阅相关文件
func (s *Service) GetSubscribeFiles(ctx context.Context, subscribeID uint) (*SubscribeInfo, error) {
	// This would typically get files related to a subscription
	// For now, we'll just return nil
	return nil, nil
}

// SeerrSubscribe OverSeerr/JellySeerr通知订阅
func (s *Service) SeerrSubscribe(ctx context.Context, req map[string]any) error {
	// This would typically handle a webhook from OverSeerr/JellySeerr
	// For now, we'll just return nil
	return nil
}

// ShareSubscribe 分享订阅 - Placeholder, implemented in share_service.go
// DeleteShare 删除分享 - Placeholder, implemented in share_service.go
// ForkSubscribe 复用订阅 - Placeholder, implemented in share_service.go
// GetShares 获取分享列表 - Placeholder, implemented in share_service.go
// GetShareStatistics 获取分享统计 - Placeholder, implemented in share_service.go

// UpdateSubscribe 更新订阅
func (s *Service) UpdateSubscribe(ctx context.Context, id uint, req UpdateSubscribeRequest) error {
	// 构建更新字段
	updates := make(map[string]any)

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Quality != nil {
		updates["quality"] = *req.Quality
	}
	if req.Resolution != nil {
		updates["resolution"] = *req.Resolution
	}
	if req.Effect != nil {
		updates["effect"] = *req.Effect
	}
	if req.Include != nil {
		updates["include"] = *req.Include
	}
	if req.Exclude != nil {
		updates["exclude"] = *req.Exclude
	}
	if req.TotalEpisode != nil {
		updates["total_episode"] = *req.TotalEpisode
	}
	if req.StartEpisode != nil {
		updates["start_episode"] = *req.StartEpisode
	}
	if req.BestVersion != nil {
		updates["best_version"] = *req.BestVersion
	}
	if req.SearchIMDBID != nil {
		updates["search_imdbid"] = *req.SearchIMDBID
	}
	if req.Sites != nil {
		updates["sites"] = req.Sites
	}
	if req.Downloader != nil {
		updates["downloader"] = *req.Downloader
	}
	if req.SavePath != nil {
		updates["save_path"] = *req.SavePath
	}
	if req.FilterGroups != nil {
		updates["filter_groups"] = req.FilterGroups
	}
	if req.CustomWords != nil {
		updates["custom_words"] = *req.CustomWords
	}
	if req.MediaCategory != nil {
		updates["media_category"] = *req.MediaCategory
	}
	if req.State != nil {
		updates["state"] = *req.State
	}

	if len(updates) == 0 {
		return nil
	}

	return s.subscribeRepo.Update(ctx, int(id), updates)
}

// DeleteSubscribe 删除订阅
func (s *Service) DeleteSubscribe(ctx context.Context, id uint) error {
	return s.subscribeRepo.Delete(ctx, int(id))
}

// PauseSubscribe 暂停订阅
func (s *Service) PauseSubscribe(ctx context.Context, id uint) error {
	updates := map[string]any{
		"state": "S",
	}
	return s.subscribeRepo.Update(ctx, int(id), updates)
}

// ResumeSubscribe 恢复订阅
func (s *Service) ResumeSubscribe(ctx context.Context, id uint) error {
	updates := map[string]any{
		"state": "R",
	}
	return s.subscribeRepo.Update(ctx, int(id), updates)
}
