package adapters

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// MediaItem 定义媒体项
type MediaItem struct {
	ID            string         `json:"id"`             // 媒体ID
	Title         string         `json:"title"`          // 标题
	OriginalTitle string         `json:"original_title"` // 原始标题
	Type          string         `json:"type"`           // 媒体类型，movie或tv
	Year          int            `json:"year"`           // 年份
	ReleaseDate   time.Time      `json:"release_date"`   // 发布日期
	Rating        float64        `json:"rating"`         // 评分
	Genres        []string       `json:"genres"`         // 类型
	Countries     []string       `json:"countries"`      // 国家/地区
	Languages     []string       `json:"languages"`      // 语言
	Overview      string         `json:"overview"`       // 简介
	Poster        string         `json:"poster"`         // 海报URL
	Backdrop      string         `json:"backdrop"`       // 背景图URL
	Runtime       int            `json:"runtime"`        // 时长（分钟）
	Status        string         `json:"status"`         // 状态
	Metadata      map[string]any `json:"metadata"`       // 元数据
	ServerName    string         `json:"server_name"`    // 媒体服务器名称
	LibraryName   string         `json:"library_name"`   // 媒体库名称
	CreatedAt     time.Time      `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time      `json:"updated_at"`     // 更新时间
}

// MediaEpisode 定义剧集信息
type MediaEpisode struct {
	ID            string         `json:"id"`             // 剧集ID
	SeasonNumber  int            `json:"season_number"`  // 季数
	EpisodeNumber int            `json:"episode_number"` // 集数
	Title         string         `json:"title"`          // 标题
	Overview      string         `json:"overview"`       // 简介
	AirDate       time.Time      `json:"air_date"`       // 播出日期
	Rating        float64        `json:"rating"`         // 评分
	Runtime       int            `json:"runtime"`        // 时长（分钟）
	StillPath     string         `json:"still_path"`     // 剧照URL
	Metadata      map[string]any `json:"metadata"`       // 元数据
}

// MediaSeason 定义季信息
type MediaSeason struct {
	ID           string         `json:"id"`            // 季ID
	SeasonNumber int            `json:"season_number"` // 季数
	Title        string         `json:"title"`         // 标题
	Overview     string         `json:"overview"`      // 简介
	AirDate      time.Time      `json:"air_date"`      // 播出日期
	PosterPath   string         `json:"poster_path"`   // 海报URL
	EpisodeCount int            `json:"episode_count"` // 集数
	Episodes     []MediaEpisode `json:"episodes"`      // 剧集列表
}

// MediaDetails 定义媒体详情
type MediaDetails struct {
	MediaItem
	Seasons []MediaSeason `json:"seasons"` // 季列表
	Cast    []MediaCast   `json:"cast"`    // 演员列表
	Crew    []MediaCrew   `json:"crew"`    // 制作人员列表
}

// MediaCast 定义演员信息
type MediaCast struct {
	ID          string `json:"id"`           // 演员ID
	Name        string `json:"name"`         // 演员名称
	Character   string `json:"character"`    // 角色名称
	ProfilePath string `json:"profile_path"` // 头像URL
	Order       int    `json:"order"`        // 排序
}

// MediaCrew 定义制作人员信息
type MediaCrew struct {
	ID          string `json:"id"`           // 人员ID
	Name        string `json:"name"`         // 人员名称
	Department  string `json:"department"`   // 部门
	Job         string `json:"job"`          // 职位
	ProfilePath string `json:"profile_path"` // 头像URL
}

// MediaService 定义媒体服务接口
type MediaService interface {
	// GetMedias 获取媒体列表
	GetMedias(ctx context.Context, params GetMediasParams) ([]MediaItem, error)

	// GetMediaDetails 获取媒体详情
	GetMediaDetails(ctx context.Context, mediaID string, params GetMediaDetailsParams) (*MediaDetails, error)

	// SearchMedias 搜索媒体
	SearchMedias(ctx context.Context, params SearchMediasParams) ([]MediaItem, error)

	// GetMediaEpisodes 获取媒体剧集
	GetMediaEpisodes(ctx context.Context, mediaID string, seasonNumber int) ([]MediaEpisode, error)

	// RefreshMedia 刷新媒体信息
	RefreshMedia(ctx context.Context, mediaID string) (*MediaItem, error)

	// GetMediaLibraries 获取媒体库列表
	GetMediaLibraries(ctx context.Context, serverName string) ([]MediaLibrary, error)
}

// MediaLibrary 定义媒体库
type MediaLibrary struct {
	ID         string         `json:"id"`          // 媒体库ID
	Name       string         `json:"name"`        // 媒体库名称
	Type       string         `json:"type"`        // 媒体库类型
	Path       string         `json:"path"`        // 媒体库路径
	ServerName string         `json:"server_name"` // 媒体服务器名称
	ItemCount  int            `json:"item_count"`  // 媒体数量
	Metadata   map[string]any `json:"metadata"`    // 元数据
	CreatedAt  time.Time      `json:"created_at"`  // 创建时间
	UpdatedAt  time.Time      `json:"updated_at"`  // 更新时间
}

// GetMediasParams 获取媒体列表参数
type GetMediasParams struct {
	ServerName    string    `json:"server_name"`    // 媒体服务器名称
	LibraryName   string    `json:"library_name"`   // 媒体库名称
	MediaType     string    `json:"media_type"`     // 媒体类型
	RecentlyAdded bool      `json:"recently_added"` // 是否只获取最近添加的媒体
	Limit         int       `json:"limit"`          // 返回结果数量限制
	Offset        int       `json:"offset"`         // 偏移量
	SortBy        string    `json:"sort_by"`        // 排序字段
	SortOrder     string    `json:"sort_order"`     // 排序顺序
	StartDate     time.Time `json:"start_date"`     // 开始日期
	EndDate       time.Time `json:"end_date"`       // 结束日期
}

// GetMediaDetailsParams 获取媒体详情参数
type GetMediaDetailsParams struct {
	ServerName     string `json:"server_name"`     // 媒体服务器名称
	IncludeSeasons bool   `json:"include_seasons"` // 是否包含季信息
	IncludeCast    bool   `json:"include_cast"`    // 是否包含演员信息
	IncludeCrew    bool   `json:"include_crew"`    // 是否包含制作人员信息
}

// SearchMediasParams 搜索媒体参数
type SearchMediasParams struct {
	ServerName string   `json:"server_name"` // 媒体服务器名称
	Query      string   `json:"query"`       // 搜索关键词
	MediaType  string   `json:"media_type"`  // 媒体类型
	Year       int      `json:"year"`        // 年份
	Genres     []string `json:"genres"`      // 类型
	Limit      int      `json:"limit"`       // 返回结果数量限制
	Offset     int      `json:"offset"`      // 偏移量
}

// MediaServiceAdapter 媒体服务适配器实现
type MediaServiceAdapter struct {
	logger *zap.Logger
	// 实际的媒体服务客户端可以在这里注入
}

// NewMediaServiceAdapter 创建新的媒体服务适配器实例
func NewMediaServiceAdapter(logger *zap.Logger) *MediaServiceAdapter {
	return &MediaServiceAdapter{
		logger: logger,
	}
}

// GetMedias 获取媒体列表
func (a *MediaServiceAdapter) GetMedias(ctx context.Context, params GetMediasParams) ([]MediaItem, error) {
	// 实际实现中，这里应该调用核心业务服务的媒体API
	// 这里使用模拟实现，返回空列表
	a.logger.Info("Getting medias", zap.String("server_name", params.ServerName), zap.String("library_name", params.LibraryName), zap.String("media_type", params.MediaType))
	return []MediaItem{}, nil
}

// GetMediaDetails 获取媒体详情
func (a *MediaServiceAdapter) GetMediaDetails(ctx context.Context, mediaID string, params GetMediaDetailsParams) (*MediaDetails, error) {
	// 实际实现中，这里应该调用核心业务服务的媒体API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Getting media details", zap.String("media_id", mediaID), zap.String("server_name", params.ServerName))
	return nil, nil
}

// SearchMedias 搜索媒体
func (a *MediaServiceAdapter) SearchMedias(ctx context.Context, params SearchMediasParams) ([]MediaItem, error) {
	// 实际实现中，这里应该调用核心业务服务的媒体API
	// 这里使用模拟实现，返回空列表
	a.logger.Info("Searching medias", zap.String("server_name", params.ServerName), zap.String("query", params.Query), zap.String("media_type", params.MediaType))
	return []MediaItem{}, nil
}

// GetMediaEpisodes 获取媒体剧集
func (a *MediaServiceAdapter) GetMediaEpisodes(ctx context.Context, mediaID string, seasonNumber int) ([]MediaEpisode, error) {
	// 实际实现中，这里应该调用核心业务服务的媒体API
	// 这里使用模拟实现，返回空列表
	a.logger.Info("Getting media episodes", zap.String("media_id", mediaID), zap.Int("season_number", seasonNumber))
	return []MediaEpisode{}, nil
}

// RefreshMedia 刷新媒体信息
func (a *MediaServiceAdapter) RefreshMedia(ctx context.Context, mediaID string) (*MediaItem, error) {
	// 实际实现中，这里应该调用核心业务服务的媒体API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Refreshing media", zap.String("media_id", mediaID))
	return nil, nil
}

// GetMediaLibraries 获取媒体库列表
func (a *MediaServiceAdapter) GetMediaLibraries(ctx context.Context, serverName string) ([]MediaLibrary, error) {
	// 实际实现中，这里应该调用核心业务服务的媒体API
	// 这里使用模拟实现，返回空列表
	a.logger.Info("Getting media libraries", zap.String("server_name", serverName))
	return []MediaLibrary{}, nil
}

// MockMediaService 模拟媒体服务实现，用于测试
type MockMediaService struct {
	logger    *zap.Logger
	medias    map[string]MediaItem
	libraries map[string][]MediaLibrary
}

// NewMockMediaService 创建新的模拟媒体服务实例
func NewMockMediaService(logger *zap.Logger) *MockMediaService {
	return &MockMediaService{
		logger:    logger,
		medias:    make(map[string]MediaItem),
		libraries: make(map[string][]MediaLibrary),
	}
}

// GetMedias 获取媒体列表（模拟实现）
func (m *MockMediaService) GetMedias(ctx context.Context, params GetMediasParams) ([]MediaItem, error) {
	m.logger.Info("Mock getting medias", zap.String("server_name", params.ServerName), zap.String("library_name", params.LibraryName), zap.String("media_type", params.MediaType))

	var medias []MediaItem
	for _, media := range m.medias {
		if (params.ServerName == "" || media.ServerName == params.ServerName) &&
			(params.LibraryName == "" || media.LibraryName == params.LibraryName) &&
			(params.MediaType == "" || media.Type == params.MediaType) {
			medias = append(medias, media)
		}
	}

	return medias, nil
}

// GetMediaDetails 获取媒体详情（模拟实现）
func (m *MockMediaService) GetMediaDetails(ctx context.Context, mediaID string, params GetMediaDetailsParams) (*MediaDetails, error) {
	m.logger.Info("Mock getting media details", zap.String("media_id", mediaID), zap.String("server_name", params.ServerName))

	media, exists := m.medias[mediaID]
	if !exists {
		return nil, nil
	}

	// 创建模拟媒体详情
	details := &MediaDetails{
		MediaItem: media,
		Seasons:   []MediaSeason{},
		Cast:      []MediaCast{},
		Crew:      []MediaCrew{},
	}

	return details, nil
}

// SearchMedias 搜索媒体（模拟实现）
func (m *MockMediaService) SearchMedias(ctx context.Context, params SearchMediasParams) ([]MediaItem, error) {
	m.logger.Info("Mock searching medias", zap.String("server_name", params.ServerName), zap.String("query", params.Query), zap.String("media_type", params.MediaType))

	var medias []MediaItem
	for _, media := range m.medias {
		if (params.ServerName == "" || media.ServerName == params.ServerName) &&
			(params.MediaType == "" || media.Type == params.MediaType) &&
			(params.Year == 0 || media.Year == params.Year) {
			medias = append(medias, media)
		}
	}

	return medias, nil
}

// GetMediaEpisodes 获取媒体剧集（模拟实现）
func (m *MockMediaService) GetMediaEpisodes(ctx context.Context, mediaID string, seasonNumber int) ([]MediaEpisode, error) {
	m.logger.Info("Mock getting media episodes", zap.String("media_id", mediaID), zap.Int("season_number", seasonNumber))

	return []MediaEpisode{}, nil
}

// RefreshMedia 刷新媒体信息（模拟实现）
func (m *MockMediaService) RefreshMedia(ctx context.Context, mediaID string) (*MediaItem, error) {
	m.logger.Info("Mock refreshing media", zap.String("media_id", mediaID))

	media, exists := m.medias[mediaID]
	if !exists {
		return nil, nil
	}

	// 更新媒体的更新时间
	media.UpdatedAt = time.Now()
	m.medias[mediaID] = media

	return &media, nil
}

// GetMediaLibraries 获取媒体库列表（模拟实现）
func (m *MockMediaService) GetMediaLibraries(ctx context.Context, serverName string) ([]MediaLibrary, error) {
	m.logger.Info("Mock getting media libraries", zap.String("server_name", serverName))

	libraries, exists := m.libraries[serverName]
	if !exists {
		return []MediaLibrary{}, nil
	}

	return libraries, nil
}
