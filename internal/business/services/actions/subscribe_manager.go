// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// SubscribeManager 订阅管理器
// 提供订阅创建、管理、缓存和重复检查功能
type SubscribeManager struct {
	subscribeService SubscribeService
	mediaService     MediaService
	cache            *WorkflowCache
	logger           *zap.Logger
	mutex            sync.RWMutex
}

// SubscribeService 订阅服务接口
type SubscribeService interface {
	CreateSubscribe(ctx context.Context, subscribe *model.Subscribe) error
	GetSubscribeByMedia(ctx context.Context, mediaType, title string, year int) (*model.Subscribe, error)
	GetSubscribe(ctx context.Context, subscribeID uint) (*model.Subscribe, error)
	ListSubscribes(ctx context.Context, userID uint) ([]*model.Subscribe, error)
	UpdateSubscribeStatus(ctx context.Context, subscribeID uint, status string) error
	DeleteSubscribe(ctx context.Context, subscribeID uint) error
}

// MediaService 媒体服务接口
type MediaService interface {
	RecognizeMedia(ctx context.Context, metaInfo *MetaInfo) (*model.Media, error)
	MediaExists(ctx context.Context, media *model.Media) (*MediaExistResult, error)
}

// NewSubscribeManager 创建订阅管理器实例
func NewSubscribeManager(
	subscribeService SubscribeService,
	mediaService MediaService,
	cache *WorkflowCache,
) *SubscribeManager {
	return &SubscribeManager{
		subscribeService: subscribeService,
		mediaService:     mediaService,
		cache:            cache,
		logger:           logger.NewLogger("subscribe_manager"),
	}
}

// AddSubscribeAction 添加订阅动作
// 实现Python项目中的add_subscribe.py功能
type AddSubscribeAction struct {
	manager *SubscribeManager
	logger  *zap.Logger
}

// NewAddSubscribeAction 创建添加订阅动作实例
func NewAddSubscribeAction(manager *SubscribeManager) *AddSubscribeAction {
	return &AddSubscribeAction{
		manager: manager,
		logger:  logger.NewLogger("add_subscribe_action"),
	}
}

// Execute 执行添加订阅动作
func (a *AddSubscribeAction) Execute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionCtx *ActionContext,
) (*ActionContext, error) {
	a.logger.Info("开始添加订阅",
		zap.Int64("workflow_id", workflowID),
		zap.Int("media_count", len(actionCtx.Medias)))

	addedSubscribes := []*model.Subscribe{}
	hasError := false

	for _, media := range actionCtx.Medias {
		// 检查缓存
		cacheKey := fmt.Sprintf("%s-%s-%d-%d", media.Type, media.Title, media.Year, media.Season)
		if a.manager.cache.CheckCache(workflowID, cacheKey) {
			a.logger.Info("订阅已添加过，跳过",
				zap.String("title", media.Title),
				zap.String("type", media.Type),
				zap.Int("year", media.Year))
			continue
		}

		// 检查订阅是否已存在
		exists, err := a.manager.subscribeService.GetSubscribeByMedia(ctx, media.Type, media.Title, media.Year)
		if err != nil {
			a.logger.Error("检查订阅存在性失败", zap.Error(err))
			hasError = true
			continue
		}

		if exists != nil {
			a.logger.Info("订阅已存在", zap.String("title", media.Title))
			// 保存缓存
			a.manager.cache.SaveCache(workflowID, cacheKey, true, 24*time.Hour)
			continue
		}

		// 识别媒体信息
		mediaInfo, err := a.manager.mediaService.RecognizeMedia(ctx, &MetaInfo{
			Title:  media.Title,
			Year:   media.Year,
			Type:   media.Type,
			Season: media.Season,
		})
		if err != nil {
			a.logger.Error("媒体识别失败", zap.Error(err), zap.String("title", media.Title))
			hasError = true
			continue
		}

		// 创建订阅
		subscribe := &model.Subscribe{
			Title:     media.Title,
			Year:      media.Year,
			Type:      media.Type,
			Season:    media.Season,
			TMDBID:    mediaInfo.TMDBID,
			DoubanID:  mediaInfo.DoubanID,
			BangumiID: mediaInfo.BangumiID,
			Status:    "active",
			UserID:    1, // 超级用户ID，可以从配置获取
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		subscribeID, err := a.manager.subscribeService.CreateSubscribe(ctx, subscribe)
		if err != nil {
			a.logger.Error("创建订阅失败", zap.Error(err), zap.String("title", media.Title))
			hasError = true
			continue
		}

		subscribe.ID = subscribeID
		addedSubscribes = append(addedSubscribes, subscribe)

		// 保存缓存
		a.manager.cache.SaveCache(workflowID, cacheKey, true, 24*time.Hour)

		a.logger.Info("订阅添加成功",
			zap.Uint("subscribe_id", subscribeID),
			zap.String("title", media.Title),
			zap.String("type", media.Type))
	}

	if len(addedSubscribes) > 0 {
		a.logger.Info("订阅添加完成", zap.Int("added_count", len(addedSubscribes)))

		// 更新上下文
		for _, subscribe := range addedSubscribes {
			actionCtx.Subscribes = append(actionCtx.Subscribes, &Subscribe{
				ID:       int(subscribe.ID),
				Title:    subscribe.Title,
				Year:     subscribe.Year,
				Type:     subscribe.Type,
				Season:   subscribe.Season,
				Status:   subscribe.Status,
				TMDBID:   subscribe.TMDBID,
				DoubanID: subscribe.DoubanID,
			})
		}
	}

	if hasError {
		a.logger.Warn("部分订阅添加失败")
	}

	// 更新工作流变量
	actionCtx.Variables["subscribes_added_count"] = len(addedSubscribes)
	actionCtx.Variables["subscribes_has_error"] = hasError

	return actionCtx, nil
}

// CheckSubscriptionExists 检查订阅是否存在
func (m *SubscribeManager) CheckSubscriptionExists(ctx context.Context, media *MediaInfo) (bool, error) {
	// 检查数据库
	existing, err := m.subscribeService.GetSubscribeByMedia(ctx, media.Type, media.Title, media.Year)
	if err != nil {
		return false, err
	}

	return existing != nil, nil
}

// ValidateSubscription 验证订阅参数
func (m *SubscribeManager) ValidateSubscription(media *MediaInfo) error {
	if media.Title == "" {
		return fmt.Errorf("媒体标题不能为空")
	}

	if media.Type == "" {
		return fmt.Errorf("媒体类型不能为空")
	}

	if media.Type == "movie" && media.Year == 0 {
		return fmt.Errorf("电影年份不能为空")
	}

	if media.Type == "tv" && media.Season <= 0 {
		return fmt.Errorf("电视剧季号必须大于0")
	}

	return nil
}

// GetSubscriptionStatistics 获取订阅统计信息
func (m *SubscribeManager) GetSubscriptionStatistics(ctx context.Context) (*SubscriptionStatistics, error) {
	// 这里可以添加更详细的统计逻辑
	return &SubscriptionStatistics{
		TotalSubscriptions:     0, // 从数据库获取
		ActiveSubscriptions:    0,
		CompletedSubscriptions: 0,
		LastUpdateTime:         time.Now(),
	}, nil
}

// BatchCreateSubscriptions 批量创建订阅
func (m *SubscribeManager) BatchCreateSubscriptions(
	ctx context.Context,
	medias []*MediaInfo,
	workflowID int64,
) ([]*model.Subscribe, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	subscribes := []*model.Subscribe{}
	errors := []error{}

	for _, media := range medias {
		wg.Add(1)
		go func(m *MediaInfo) {
			defer wg.Done()

			// 验证参数
			if err := m.ValidateSubscription(m); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("媒体 %s 验证失败: %w", m.Title, err))
				mu.Unlock()
				return
			}

			// 检查是否已存在
			exists, err := m.CheckSubscriptionExists(ctx, m)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("检查媒体 %s 存在性失败: %w", m.Title, err))
				mu.Unlock()
				return
			}

			if exists {
				return // 已存在，跳过
			}

			// 识别媒体信息
			mediaInfo, err := m.mediaService.RecognizeMedia(ctx, &MetaInfo{
				Title:  m.Title,
				Year:   m.Year,
				Type:   m.Type,
				Season: m.Season,
			})
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("识别媒体 %s 失败: %w", m.Title, err))
				mu.Unlock()
				return
			}

			// 创建订阅
			subscribe := &model.Subscribe{
				Title:     m.Title,
				Year:      m.Year,
				Type:      m.Type,
				Season:    m.Season,
				TMDBID:    mediaInfo.TMDBID,
				DoubanID:  mediaInfo.DoubanID,
				BangumiID: mediaInfo.BangumiID,
				Status:    "active",
				UserID:    1, // 从配置获取
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			subscribeID, err := m.subscribeService.CreateSubscribe(ctx, subscribe)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("创建订阅 %s 失败: %w", m.Title, err))
				mu.Unlock()
				return
			}

			subscribe.ID = subscribeID
			mu.Lock()
			subscribes = append(subscribes, subscribe)
			mu.Unlock()

			// 保存缓存
			cacheKey := fmt.Sprintf("%s-%s-%d-%d", m.Type, m.Title, m.Year, m.Season)
			m.cache.SaveCache(workflowID, cacheKey, true, 24*time.Hour)

		}(media)
	}

	wg.Wait()

	if len(errors) > 0 {
		return subscribes, fmt.Errorf("批量创建订阅时发生错误: %v", errors)
	}

	return subscribes, nil
}

// 数据结构定义

// MetaInfo 元信息
type MetaInfo struct {
	Title   string `json:"title"`
	Year    int    `json:"year"`
	Type    string `json:"type"` // "movie", "tv"
	Season  int    `json:"season"`
	Episode int    `json:"episode"`
}

// MediaExistResult 媒体存在结果
type MediaExistResult struct {
	Exists  bool          `json:"exists"`
	Seasons map[int][]int `json:"seasons"` // 季 -> 集列表
	Paths   []string      `json:"paths"`   // 文件路径列表
	Quality string        `json:"quality"` // 质量信息
}

// SubscriptionStatistics 订阅统计信息
type SubscriptionStatistics struct {
	TotalSubscriptions     int       `json:"total_subscriptions"`
	ActiveSubscriptions    int       `json:"active_subscriptions"`
	CompletedSubscriptions int       `json:"completed_subscriptions"`
	FailedSubscriptions    int       `json:"failed_subscriptions"`
	LastUpdateTime         time.Time `json:"last_update_time"`
}

// Subscribe 订阅信息（用于工作流上下文）
type Subscribe struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Type     string `json:"type"`
	Season   int    `json:"season"`
	Status   string `json:"status"`
	TMDBID   int    `json:"tmdb_id"`
	DoubanID int    `json:"douban_id"`
}
