package mediaserver

import (
	"context"
	"time"
)

// MediaType 媒体类型
// 统一抽象电影、剧集、季、剧集单集等
// 方便后续在业务层做聚合

type MediaType string

const (
	MediaTypeMovie   MediaType = "movie"
	MediaTypeEpisode MediaType = "episode"
	MediaTypeSeason  MediaType = "season"
	MediaTypeSeries  MediaType = "series"
	MediaTypeUnknown MediaType = "unknown"
)

// ExternalID 外部ID（TMDB/TVDB/IMDB等）
type ExternalID struct {
	TMDBID   *int    `json:"tmdb_id,omitempty"`
	TVDBID   *int    `json:"tvdb_id,omitempty"`
	IMDBID   *string `json:"imdb_id,omitempty"`
	DoubanID *string `json:"douban_id,omitempty"`
}

// MediaLibrary 媒体库信息
type MediaLibrary struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      MediaType  `json:"type"`
	ItemCount int64      `json:"item_count"`
	LastScan  *time.Time `json:"last_scan,omitempty"`
}

// MediaItem 媒体条目（电影/剧集/季/单集）
type MediaItem struct {
	ID           string     `json:"id"`
	ServerID     string     `json:"server_id"`
	LibraryID    string     `json:"library_id"`
	Name         string     `json:"name"`
	OriginalName string     `json:"original_name,omitempty"`
	Type         MediaType  `json:"type"`
	Year         *int       `json:"year,omitempty"`
	Season       *int       `json:"season,omitempty"`
	Episode      *int       `json:"episode,omitempty"`
	ExternalIDs  ExternalID `json:"external_ids"`
	IsMissing    bool       `json:"is_missing"`
}

// ServerInfo 服务器基础信息
type ServerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"` // emby/plex/jellyfin
	URL     string `json:"url"`
}

// SearchQuery 媒体搜索条件
type SearchQuery struct {
	Keyword string    `json:"keyword"`
	Type    MediaType `json:"type"`
	Year    *int      `json:"year,omitempty"`
	Limit   int       `json:"limit,omitempty"`
}

// MediaServerClient 媒体服务器统一接口
// Emby/Plex/Jellyfin 均需实现此接口

type MediaServerClient interface {
	// TestConnection 测试连接是否正常
	TestConnection(ctx context.Context) error

	// GetServerInfo 获取服务器基础信息
	GetServerInfo(ctx context.Context) (*ServerInfo, error)

	// ListLibraries 列出媒体库
	ListLibraries(ctx context.Context) ([]*MediaLibrary, error)

	// ScanLibrary 触发媒体库扫描
	ScanLibrary(ctx context.Context, libraryID string) error

	// GetItem 根据ID获取单个媒体条目
	GetItem(ctx context.Context, id string) (*MediaItem, error)

	// SearchItems 按关键字/年份等搜索媒体条目
	SearchItems(ctx context.Context, query SearchQuery) ([]*MediaItem, error)
}

// Factory 媒体服务器客户端工厂
// 用于根据名称获取具体实现（emby/plex/jellyfin）

type Factory struct {
	clients map[string]MediaServerClient
}

// NewFactory 创建工厂实例
func NewFactory() *Factory {
	return &Factory{
		clients: make(map[string]MediaServerClient),
	}
}

// Register 注册媒体服务器客户端
func (f *Factory) Register(name string, client MediaServerClient) {
	f.clients[name] = client
}

// GetClient 获取媒体服务器客户端
func (f *Factory) GetClient(name string) (MediaServerClient, bool) {
	c, ok := f.clients[name]
	return c, ok
}

// ListClients 列出已注册的客户端名称
func (f *Factory) ListClients() []string {
	names := make([]string, 0, len(f.clients))
	for name := range f.clients {
		names = append(names, name)
	}
	return names
}
