// Package trimemedia Trimedia服务实现
package trimemedia

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/integration/trimemedia"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// TrimediaService Trimedia服务
type TrimediaService struct {
	client *trimemedia.APIClient
	logger *zap.Logger
}

// TrimediaConfig Trimedia配置
type TrimediaConfig struct {
	Host   string `json:"host"`    // Trimedia服务地址
	APIKey string `json:"api_key"` // API密钥
}

// NewTrimediaService 创建Trimedia服务
func NewTrimediaService(config *TrimediaConfig) *TrimediaService {
	if config == nil || config.Host == "" || config.APIKey == "" {
		logger.Logger.Error("Trimedia config is incomplete")
		return nil
	}

	client := trimemedia.NewAPIClient(config.Host, config.APIKey)
	if client == nil {
		logger.Logger.Error("Failed to create Trimedia API client")
		return nil
	}

	service := &TrimediaService{
		client: client,
		logger: logger.Logger,
	}

	// 初始化连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	version, err := client.SysVersion(ctx)
	if err != nil {
		service.logger.Error("Failed to connect to Trimedia server",
			zap.String("host", config.Host),
			zap.Error(err))
		return nil
	}

	service.logger.Info("Trimedia service initialized successfully",
		zap.String("host", config.Host),
		zap.String("frontend_version", version.Frontend),
		zap.String("backend_version", version.Backend))

	return service
}

// APIKey 获取API密钥
// 兼容Python版本的apikey()方法
func (ts *TrimediaService) APIKey() string {
	if ts.client == nil {
		return ""
	}
	return ts.client.APIKey()
}

// Authenticate 认证
// 兼容Python版本的authenticate()方法
func (ts *TrimediaService) Authenticate(ctx context.Context, username, password string) (string, error) {
	if ts.client == nil {
		return "", fmt.Errorf("trimedia client is nil")
	}

	ts.logger.Info("Attempting Trimedia authentication",
		zap.String("username", username))

	token, err := ts.client.Login(ctx, username, password)
	if err != nil {
		ts.logger.Error("Trimedia authentication failed",
			zap.String("username", username),
			zap.Error(err))
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	if token == "" {
		return "", fmt.Errorf("authentication failed: no token received")
	}

	ts.logger.Info("Trimedia authentication successful",
		zap.String("username", username))

	return token, nil
}

// IsAuthenticated 检查认证状态
func (ts *TrimediaService) IsAuthenticated() bool {
	if ts.client == nil {
		return false
	}
	return ts.client.IsAuthenticated()
}

// GetVersion 获取版本信息
func (ts *TrimediaService) GetVersion(ctx context.Context) (*trimemedia.Version, error) {
	if ts.client == nil {
		return nil, fmt.Errorf("trimedia client is nil")
	}

	return ts.client.SysVersion(ctx)
}

// GetMediaSummary 获取媒体汇总
func (ts *TrimediaService) GetMediaSummary(ctx context.Context) (*trimemedia.MediaSummary, error) {
	if ts.client == nil {
		return nil, fmt.Errorf("trimedia client is nil")
	}

	summary, err := ts.client.GetMediaSummary(ctx)
	if err != nil {
		ts.logger.Error("Failed to get media summary", zap.Error(err))
		return nil, fmt.Errorf("get media summary failed: %w", err)
	}

	ts.logger.Debug("Media summary retrieved",
		zap.Int("total", summary.Total),
		zap.Int("movies", summary.Movie),
		zap.Int("tv", summary.TV),
		zap.Int("videos", summary.Video))

	return summary, nil
}

// SearchMedia 搜索媒体
func (ts *TrimediaService) SearchMedia(ctx context.Context, keyword string, mediaType trimemedia.Category, page int) ([]*trimemedia.MediaItem, error) {
	if ts.client == nil {
		return nil, fmt.Errorf("trimedia client is nil")
	}

	items, err := ts.client.SearchMedia(ctx, keyword, mediaType, page)
	if err != nil {
		ts.logger.Error("Failed to search media",
			zap.String("keyword", keyword),
			zap.String("media_type", string(mediaType)),
			zap.Int("page", page),
			zap.Error(err))
		return nil, fmt.Errorf("search media failed: %w", err)
	}

	ts.logger.Debug("Media search completed",
		zap.String("keyword", keyword),
		zap.String("media_type", string(mediaType)),
		zap.Int("page", page),
		zap.Int("results", len(items)))

	return items, nil
}

// GetMediaDetail 获取媒体详情
func (ts *TrimediaService) GetMediaDetail(ctx context.Context, guid string) (*trimemedia.MediaItem, error) {
	if ts.client == nil {
		return nil, fmt.Errorf("trimedia client is nil")
	}

	detail, err := ts.client.GetMediaDetail(ctx, guid)
	if err != nil {
		ts.logger.Error("Failed to get media detail",
			zap.String("guid", guid),
			zap.Error(err))
		return nil, fmt.Errorf("get media detail failed: %w", err)
	}

	ts.logger.Debug("Media detail retrieved",
		zap.String("guid", guid),
		zap.String("title", detail.Title))

	return detail, nil
}

// GetWatchHistory 获取观看历史
func (ts *TrimediaService) GetWatchHistory(ctx context.Context, mediaType trimemedia.Category, page int) ([]*trimemedia.MediaItem, error) {
	if ts.client == nil {
		return nil, fmt.Errorf("trimedia client is nil")
	}

	items, err := ts.client.GetWatchHistory(ctx, mediaType, page)
	if err != nil {
		ts.logger.Error("Failed to get watch history",
			zap.String("media_type", string(mediaType)),
			zap.Int("page", page),
			zap.Error(err))
		return nil, fmt.Errorf("get watch history failed: %w", err)
	}

	ts.logger.Debug("Watch history retrieved",
		zap.String("media_type", string(mediaType)),
		zap.Int("page", page),
		zap.Int("results", len(items)))

	return items, nil
}

// MarkAsWatched 标记为已观看
func (ts *TrimediaService) MarkAsWatched(ctx context.Context, guid string) error {
	if ts.client == nil {
		return fmt.Errorf("trimedia client is nil")
	}

	err := ts.client.MarkAsWatched(ctx, guid)
	if err != nil {
		ts.logger.Error("Failed to mark as watched",
			zap.String("guid", guid),
			zap.Error(err))
		return fmt.Errorf("mark as watched failed: %w", err)
	}

	ts.logger.Debug("Media marked as watched", zap.String("guid", guid))
	return nil
}

// MarkAsUnwatched 标记为未观看
func (ts *TrimediaService) MarkAsUnwatched(ctx context.Context, guid string) error {
	if ts.client == nil {
		return fmt.Errorf("trimedia client is nil")
	}

	err := ts.client.MarkAsUnwatched(ctx, guid)
	if err != nil {
		ts.logger.Error("Failed to mark as unwatched",
			zap.String("guid", guid),
			zap.Error(err))
		return fmt.Errorf("mark as unwatched failed: %w", err)
	}

	ts.logger.Debug("Media marked as unwatched", zap.String("guid", guid))
	return nil
}

// AddToFavorite 添加到收藏
func (ts *TrimediaService) AddToFavorite(ctx context.Context, guid string) error {
	if ts.client == nil {
		return fmt.Errorf("trimedia client is nil")
	}

	err := ts.client.AddToFavorite(ctx, guid)
	if err != nil {
		ts.logger.Error("Failed to add to favorite",
			zap.String("guid", guid),
			zap.Error(err))
		return fmt.Errorf("add to favorite failed: %w", err)
	}

	ts.logger.Debug("Media added to favorite", zap.String("guid", guid))
	return nil
}

// RemoveFromFavorite 从收藏移除
func (ts *TrimediaService) RemoveFromFavorite(ctx context.Context, guid string) error {
	if ts.client == nil {
		return fmt.Errorf("trimedia client is nil")
	}

	err := ts.client.RemoveFromFavorite(ctx, guid)
	if err != nil {
		ts.logger.Error("Failed to remove from favorite",
			zap.String("guid", guid),
			zap.Error(err))
		return fmt.Errorf("remove from favorite failed: %w", err)
	}

	ts.logger.Debug("Media removed from favorite", zap.String("guid", guid))
	return nil
}

// GetFavorites 获取收藏列表
func (ts *TrimediaService) GetFavorites(ctx context.Context, mediaType trimemedia.Category, page int) ([]*trimemedia.MediaItem, error) {
	if ts.client == nil {
		return nil, fmt.Errorf("trimedia client is nil")
	}

	items, err := ts.client.GetFavorites(ctx, mediaType, page)
	if err != nil {
		ts.logger.Error("Failed to get favorites",
			zap.String("media_type", string(mediaType)),
			zap.Int("page", page),
			zap.Error(err))
		return nil, fmt.Errorf("get favorites failed: %w", err)
	}

	ts.logger.Debug("Favorites retrieved",
		zap.String("media_type", string(mediaType)),
		zap.Int("page", page),
		zap.Int("results", len(items)))

	return items, nil
}

// HealthCheck 健康检查
func (ts *TrimediaService) HealthCheck(ctx context.Context) error {
	if ts.client == nil {
		return fmt.Errorf("trimedia client is nil")
	}

	err := ts.client.HealthCheck(ctx)
	if err != nil {
		ts.logger.Error("Trimedia health check failed", zap.Error(err))
		return fmt.Errorf("health check failed: %w", err)
	}

	ts.logger.Debug("Trimedia health check passed")
	return nil
}

// GetHost 获取服务主机地址
func (ts *TrimediaService) GetHost() string {
	if ts.client == nil {
		return ""
	}
	return ts.client.Host()
}

// GetToken 获取认证令牌
func (ts *TrimediaService) GetToken() string {
	if ts.client == nil {
		return ""
	}
	return ts.client.Token()
}

// SyncWatchHistory 同步观看历史
func (ts *TrimediaService) SyncWatchHistory(ctx context.Context, watchHistory []*WatchRecord) error {
	if ts.client == nil {
		return fmt.Errorf("trimedia client is nil")
	}

	ts.logger.Info("Syncing watch history",
		zap.Int("record_count", len(watchHistory)))

	// 逐条同步观看历史
	for i, record := range watchHistory {
		if err := ts.client.MarkAsWatched(ctx, record.GUID); err != nil {
			ts.logger.Error("Failed to sync watch record",
				zap.Int("index", i),
				zap.String("guid", record.GUID),
				zap.Error(err))
			// 继续同步其他记录
			continue
		}
	}

	ts.logger.Info("Watch history sync completed",
		zap.Int("record_count", len(watchHistory)))

	return nil
}

// WatchRecord 观看记录
type WatchRecord struct {
	GUID      string    `json:"guid"`      // 媒体GUID
	WatchedAt time.Time `json:"watched_at"` // 观看时间
	Duration  int       `json:"duration"`   // 观看时长(秒)
	Progress  float64   `json:"progress"`  // 观看进度(0-1)
}

// BatchMarkAsWatched 批量标记为已观看
func (ts *TrimediaService) BatchMarkAsWatched(ctx context.Context, guids []string) error {
	if ts.client == nil {
		return fmt.Errorf("trimedia client is nil")
	}

	ts.logger.Info("Batch marking as watched",
		zap.Int("guid_count", len(guids)))

	// 逐条标记
	for i, guid := range guids {
		if err := ts.client.MarkAsWatched(ctx, guid); err != nil {
			ts.logger.Error("Failed to batch mark as watched",
				zap.Int("index", i),
				zap.String("guid", guid),
				zap.Error(err))
			// 继续处理其他GUID
			continue
		}
	}

	ts.logger.Info("Batch marking as watched completed",
		zap.Int("guid_count", len(guids)))

	return nil
}

// GetMediaStatistics 获取媒体统计信息
func (ts *TrimediaService) GetMediaStatistics(ctx context.Context) (*MediaStatistics, error) {
	summary, err := ts.GetMediaSummary(ctx)
	if err != nil {
		return nil, err
	}

	stats := &MediaStatistics{
		TotalMedia:  summary.Total,
		MovieCount:  summary.Movie,
		TVCount:     summary.TV,
		VideoCount:  summary.Video,
		FavoriteCount: summary.Favorite,
		LastSync:     time.Now(),
	}

	ts.logger.Info("Media statistics retrieved",
		zap.Int("total_media", stats.TotalMedia),
		zap.Int("movie_count", stats.MovieCount),
		zap.Int("tv_count", stats.TVCount),
		zap.Int("video_count", stats.VideoCount),
		zap.Int("favorite_count", stats.FavoriteCount))

	return stats, nil
}

// MediaStatistics 媒体统计信息
type MediaStatistics struct {
	TotalMedia    int       `json:"total_media"`     // 总媒体数
	MovieCount    int       `json:"movie_count"`     // 电影数量
	TVCount       int       `json:"tv_count"`        // 电视剧数量
	VideoCount    int       `json:"video_count"`     // 视频数量
	FavoriteCount int       `json:"favorite_count"`  // 收藏数量
	LastSync      time.Time `json:"last_sync"`       // 最后同步时间
}

// Close 关闭服务
func (ts *TrimediaService) Close() error {
	// 清理资源
	ts.logger.Info("Trimedia service closed")
	return nil
}