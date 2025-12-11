package subscribe

import (
	"context"

	"moviepilot-go/internal/models/dto"
)

// SubscribeRepository 订阅仓储接口
type SubscribeRepository interface {
	// Add 添加订阅
	Add(ctx context.Context, subscribe *Subscribe) (int, error)

	// Get 获取订阅
	Get(ctx context.Context, id int) (*Subscribe, error)

	// List 获取订阅列表
	List(ctx context.Context, states []string) ([]*Subscribe, error)

	// Update 更新订阅
	Update(ctx context.Context, id int, updates map[string]any) error

	// Delete 删除订阅
	Delete(ctx context.Context, id int) error

	// Exists 判断订阅是否存在
	Exists(ctx context.Context, tmdbID *int, doubanID string, season int) (bool, error)

	// ExistHistory 判断订阅历史是否存在
	ExistHistory(ctx context.Context, tmdbID *int, doubanID string, season int) (bool, error)

	// AddHistory 添加订阅历史
	AddHistory(ctx context.Context, subscribe *Subscribe) error
}

// SystemConfigRepository 系统配置仓储接口
type SystemConfigRepository interface {
	// Get 获取配置
	Get(ctx context.Context, key string) (any, error)

	// Set 设置配置
	Set(ctx context.Context, key string, value any) error
}

// SiteRepository 站点仓储接口
type SiteRepository interface {
	// GetDomainsByIDs 根据站点ID获取域名列表
	GetDomainsByIDs(ctx context.Context, ids []int) ([]string, error)
}

// DownloadHistoryRepository 下载历史仓储接口
type DownloadHistoryRepository interface {
	// GetByMediaID 根据媒体ID获取下载历史
	GetByMediaID(ctx context.Context, tmdbID *int, doubanID string) ([]*DownloadHistory, error)

	// GetFilesByHash 根据Hash获取下载文件
	GetFilesByHash(ctx context.Context, hash string) ([]*DownloadFile, error)
}

// MediaService 媒体服务接口
type MediaService interface {
	// RecognizeMedia 识别媒体信息
	RecognizeMedia(ctx context.Context, req *RecognizeMediaRequest) (*MediaInfo, error)

	// GetTMDBInfoByDoubanID 根据豆瓣ID获取TMDB信息
	GetTMDBInfoByDoubanID(ctx context.Context, doubanID string, mediaType string) (*MediaInfo, error)

	// ObtainImages 获取媒体图片
	ObtainImages(ctx context.Context, mediaInfo *MediaInfo) error

	// MediaFiles 获取媒体库文件
	MediaFiles(ctx context.Context, mediaInfo *MediaInfo) ([]*MediaFileItem, error)
}

// SearchService 搜索服务接口
type SearchService interface {
	// Process 处理搜索请求
	Process(ctx context.Context, req *SearchRequest) ([]*Context, error)
}

// DownloadService 下载服务接口
type DownloadService interface {
	// BatchDownload 批量下载
	BatchDownload(ctx context.Context, req *BatchDownloadRequest) (*BatchDownloadResult, error)

	// GetNoExistsInfo 获取缺失媒体信息
	GetNoExistsInfo(ctx context.Context, req *GetNoExistsInfoRequest) (*GetNoExistsInfoResult, error)
}

// TorrentsService 种子服务接口
type TorrentsService interface {
	// Refresh 刷新站点资源
	Refresh(ctx context.Context, sites []int) (map[string][]*Context, error)
}

// FilterService 过滤服务接口
type FilterService interface {
	// FilterTorrents 过滤种子
	FilterTorrents(ctx context.Context, req *FilterTorrentsRequest) ([]*TorrentInfo, error)
}

// EventService 事件服务接口
type EventService interface {
	// SendEvent 发送事件
	SendEvent(ctx context.Context, eventType string, data any) error

	// SendEventAsync 异步发送事件
	SendEventAsync(ctx context.Context, eventType string, data any)

	// SendMediaRecognizeConvertEvent 发送媒体识别转换事件
	SendMediaRecognizeConvertEvent(ctx context.Context, data *MediaRecognizeConvertEventData) (*MediaRecognizeConvertEventData, error)
}

// MessageService 消息服务接口
type MessageService interface {
	// PostMessage 发送消息
	PostMessage(ctx context.Context, notification *Notification, meta *MetaInfo, mediaInfo *MediaInfo, extra map[string]any) error

	// Put 发送系统消息
	Put(ctx context.Context, message string, title string, role string) error
}

// TorrentHelper 种子助手接口
type TorrentHelper interface {
	// MatchTorrent 匹配种子
	MatchTorrent(mediaInfo *MediaInfo, torrentMeta *MetaInfo, torrent *TorrentInfo) bool

	// FilterTorrent 过滤种子
	FilterTorrent(torrentInfo *TorrentInfo, filterParams map[string]any) bool
}

// WordsMatcher 识别词匹配器接口
type WordsMatcher interface {
	// Prepare 准备识别词
	Prepare(text string, customWords []string) (string, []string)
}

// SubscribeHelper 订阅助手接口
type SubscribeHelper interface {
	// SubRegAsync 异步注册订阅统计
	SubRegAsync(ctx context.Context, data map[string]any) error

	// SubDoneAsync 异步完成订阅统计
	SubDoneAsync(ctx context.Context, data map[string]any) error

	// GetShares 获取分享列表
	GetShares(ctx context.Context) ([]map[string]any, error)
}

// TmdbService TMDB服务接口
type TmdbService interface {
	// GetEpisodes 获取剧集信息
	GetEpisodes(ctx context.Context, tmdbID int, season int, episodeGroup string) ([]*Episode, error)
}

// DownloadHistory 下载历史
type DownloadHistory struct {
	TorrentName  string `json:"torrent_name"`
	TorrentSite  string `json:"torrent_site"`
	DownloadHash string `json:"download_hash"`
}

// DownloadFile 下载文件
type DownloadFile struct {
	Downloader string `json:"downloader"`
	FilePath   string `json:"filepath"`
	FullPath   string `json:"fullpath"`
}

// MediaFileItem 媒体库文件项
type MediaFileItem struct {
	Storage string `json:"storage"`
	Path    string `json:"path"`
}

// Episode 剧集信息
type Episode struct {
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	EpisodeNumber int    `json:"episode_number"`
	StillPath     string `json:"still_path"`
}

// AnalyticsService 订阅分析服务接口
type AnalyticsService interface {
	// GetOverallStats 获取总体统计
	GetOverallStats(ctx context.Context) (*OverallStats, error)

	// GetSubscribeStats 获取订阅统计
	GetSubscribeStats(ctx context.Context, id uint) (*SubscribeStats, error)

	// GetTrendData 获取趋势数据
	GetTrendData(ctx context.Context, days int) ([]*TrendPoint, error)

	// GetTopSubscribes 获取热门订阅
	GetTopSubscribes(ctx context.Context, limit int) ([]*TopSubscribe, error)
}

// HistoryService 订阅历史服务接口
type HistoryService interface {
	// GetMatchHistory 获取匹配历史
	GetMatchHistory(ctx context.Context, id uint, limit int) ([]*MatchRecord, error)

	// GetMatchStats 获取匹配统计
	GetMatchStats(ctx context.Context, id uint) (*MatchStats, error)
}

// ShareService 订阅分享服务接口
type ShareService interface {
	// ShareSubscribe 分享订阅
	ShareSubscribe(ctx context.Context, req *dto.ShareSubscribeRequest) (*dto.SubscribeShare, error)

	// DeleteShare 删除分享
	DeleteShare(ctx context.Context, shareID int) error

	// ForkSubscribe 复用订阅
	ForkSubscribe(ctx context.Context, shareID int) error

	// GetShares 获取分享列表
	GetShares(ctx context.Context, name string, page, count int) ([]*dto.SubscribeShare, error)

	// GetShareStatistics 获取分享统计
	GetShareStatistics(ctx context.Context, userID string) ([]*dto.ShareStatistics, error)

	// FollowUser 关注用户
	FollowUser(ctx context.Context, userID, shareUID string) error

	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, userID, shareUID string) error

	// GetFollowedUsers 获取关注列表
	GetFollowedUsers(ctx context.Context, userID string) ([]string, error)

	// GetPopularShares 获取热门分享
	GetPopularShares(ctx context.Context, limit int) ([]*dto.SubscribeShare, error)
}
