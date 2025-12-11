package indexer

import (
	"context"
	"time"
)

// TorrentCategory 种子分类
type TorrentCategory string

const (
	// CategoryMovie 电影
	CategoryMovie TorrentCategory = "movie"
	// CategoryTV 剧集
	CategoryTV TorrentCategory = "tv"
	// CategoryAnime 动漫
	CategoryAnime TorrentCategory = "anime"
	// CategoryMusic 音乐
	CategoryMusic TorrentCategory = "music"
	// CategoryOther 其他
	CategoryOther TorrentCategory = "other"
)

// Torrent 种子信息
type Torrent struct {
	// Title 标题
	Title string `json:"title"`
	// Description 描述
	Description string `json:"description,omitempty"`
	// Link 种子链接或磁力链接
	Link string `json:"link"`
	// MagnetURL 磁力链接
	MagnetURL string `json:"magnet_url,omitempty"`
	// Size 文件大小（字节）
	Size int64 `json:"size"`
	// Seeders 做种数
	Seeders int `json:"seeders"`
	// Leechers 下载数
	Leechers int `json:"leechers"`
	// PublishDate 发布时间
	PublishDate time.Time `json:"publish_date"`
	// Category 分类
	Category TorrentCategory `json:"category"`
	// IndexerName 索引器名称
	IndexerName string `json:"indexer_name"`
	// IMDBID IMDB ID（可选）
	IMDBID string `json:"imdb_id,omitempty"`
	// TMDBID TMDB ID（可选）
	TMDBID int `json:"tmdb_id,omitempty"`
	// Extra 额外信息
	Extra map[string]any `json:"extra,omitempty"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	// Query 搜索关键词
	Query string
	// Category 分类过滤
	Category TorrentCategory
	// IMDBID IMDB ID 过滤
	IMDBID string
	// TMDBID TMDB ID 过滤
	TMDBID int
	// Limit 结果数量限制
	Limit int
	// Offset 结果偏移量
	Offset int
	// MinSeeders 最小做种数
	MinSeeders int
	// Extra 额外参数
	Extra map[string]any
}

// Client 索引器客户端接口
// Jackett、Prowlarr 等索引器都需要实现此接口
type Client interface {
	// Name 返回索引器名称
	Name() string

	// Search 搜索种子
	Search(ctx context.Context, opts SearchOptions) ([]*Torrent, error)

	// TestConnection 测试连接
	TestConnection(ctx context.Context) error

	// GetCapabilities 获取索引器能力（支持的分类、搜索参数等）
	GetCapabilities(ctx context.Context) (*Capabilities, error)
}

// Capabilities 索引器能力
type Capabilities struct {
	// SupportedCategories 支持的分类
	SupportedCategories []TorrentCategory `json:"supported_categories"`
	// SupportIMDBSearch 是否支持 IMDB 搜索
	SupportIMDBSearch bool `json:"support_imdb_search"`
	// SupportTMDBSearch 是否支持 TMDB 搜索
	SupportTMDBSearch bool `json:"support_tmdb_search"`
	// MaxResults 最大结果数
	MaxResults int `json:"max_results"`
}

// Factory 索引器客户端工厂
type Factory struct {
	clients map[string]Client
}

// NewFactory 创建工厂实例
func NewFactory() *Factory {
	return &Factory{
		clients: make(map[string]Client),
	}
}

// Register 注册索引器客户端
func (f *Factory) Register(client Client) {
	f.clients[client.Name()] = client
}

// Get 获取指定名称的索引器客户端
func (f *Factory) Get(name string) (Client, bool) {
	client, ok := f.clients[name]
	return client, ok
}

// List 列出所有已注册的索引器客户端
func (f *Factory) List() []string {
	names := make([]string, 0, len(f.clients))
	for name := range f.clients {
		names = append(names, name)
	}
	return names
}

// GetAll 获取所有索引器客户端
func (f *Factory) GetAll() []Client {
	clients := make([]Client, 0, len(f.clients))
	for _, client := range f.clients {
		clients = append(clients, client)
	}
	return clients
}
