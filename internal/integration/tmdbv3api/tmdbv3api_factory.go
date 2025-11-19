// Package tmdbv3api TMDBv3 API工厂模块
// 提供TMDBv3 API客户端的创建和配置功能
package tmdbv3api

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/tmdb"
)

// TMDBFactory TMDB API工厂
type TMDBFactory struct {
	defaultLanguage string
	defaultCache   bool
	defaultTTL     time.Duration
}

// NewTMDBFactory 创建TMDB工厂实例
func NewTMDBFactory() *TMDBFactory {
	return &TMDBFactory{
		defaultLanguage: "zh-CN",
		defaultCache:   true,
		defaultTTL:     15 * time.Minute,
	}
}

// SetDefaultLanguage 设置默认语言
func (f *TMDBFactory) SetDefaultLanguage(language string) {
	f.defaultLanguage = language
}

// SetDefaultCache 设置默认缓存状态
func (f *TMDBFactory) SetDefaultCache(cache bool) {
	f.defaultCache = cache
}

// SetDefaultTTL 设置默认缓存TTL
func (f *TMDBFactory) SetDefaultTTL(ttl time.Duration) {
	f.defaultTTL = ttl
}

// CreateClient 创建TMDB客户端
// apiKey: TMDB API密钥
func (f *TMDBFactory) CreateClient(apiKey string) *TMDb {
	return NewTMDb(apiKey, f.defaultLanguage, f.defaultCache)
}

// CreateClientWithOptions 创建带选项的TMDB客户端
type ClientOptions struct {
	Language   string        `json:"language"`
	Cache      bool          `json:"cache"`
	CacheTTL   time.Duration `json:"cache_ttl"`
	Timeout    time.Duration `json:"timeout"`
	Debug      bool          `json:"debug"`
	ProxyURL   string        `json:"proxy_url"`
}

// CreateClientWithOptions 创建带选项的TMDB客户端
func (f *TMDBFactory) CreateClientWithOptions(apiKey string, options ClientOptions) *TMDb {
	// 使用默认值填充未设置的选项
	if options.Language == "" {
		options.Language = f.defaultLanguage
	}
	if options.CacheTTL == 0 {
		options.CacheTTL = f.defaultTTL
	}

	client := NewTMDb(apiKey, options.Language, options.Cache)
	client.SetCacheTTL(options.CacheTTL)

	return client
}

// TMDBClientPool TMDB客户端池
type TMDBClientPool struct {
	clients map[string]*TMDb
	factory *TMDBFactory
}

// NewTMDBClientPool 创建TMDB客户端池
func NewTMDBClientPool() *TMDBClientPool {
	return &TMDBClientPool{
		clients: make(map[string]*TMDb),
		factory: NewTMDBFactory(),
	}
}

// GetClient 获取或创建TMDB客户端
// name: 客户端名称
// apiKey: TMDB API密钥
func (p *TMDBClientPool) GetClient(name, apiKey string) *TMDb {
	if client, exists := p.clients[name]; exists {
		return client
	}

	client := p.factory.CreateClient(apiKey)
	p.clients[name] = client
	return client
}

// GetClientWithOptions 获取或创建带选项的TMDB客户端
func (p *TMDBClientPool) GetClientWithOptions(name, apiKey string, options ClientOptions) *TMDb {
	if client, exists := p.clients[name]; exists {
		return client
	}

	client := p.factory.CreateClientWithOptions(apiKey, options)
	p.clients[name] = client
	return client
}

// RemoveClient 移除客户端
func (p *TMDBClientPool) RemoveClient(name string) {
	delete(p.clients, name)
}

// ClearAllClients 清除所有客户端
func (p *TMDBClientPool) ClearAllClients() {
	p.clients = make(map[string]*TMDb)
}

// GetClientNames 获取所有客户端名称
func (p *TMDBClientPool) GetClientNames() []string {
	names := make([]string, 0, len(p.clients))
	for name := range p.clients {
		names = append(names, name)
	}
	return names
}

// 全局TMDB实例
var (
	globalFactory *TMDBFactory
	globalPool    *TMDBClientPool
)

// init 初始化全局实例
func init() {
	globalFactory = NewTMDBFactory()
	globalPool = NewTMDBClientPool()
}

// GetGlobalFactory 获取全局工厂实例
func GetGlobalFactory() *TMDBFactory {
	return globalFactory
}

// GetGlobalPool 获取全局客户端池实例
func GetGlobalPool() *TMDBClientPool {
	return globalPool
}

// CreateGlobalClient 创建全局客户端
func CreateGlobalClient(name, apiKey string) *TMDb {
	return globalPool.GetClient(name, apiKey)
}

// GetGlobalClient 获取全局客户端
func GetGlobalClient(name string) *TMDb {
	if client, exists := globalPool.clients[name]; exists {
		return client
	}
	return nil
}

// QuickSearch 快速搜索（使用默认客户端）
func QuickSearch(ctx context.Context, query string, page int) (*tmdb.SearchResult, error) {
	client := GetGlobalClient("default")
	if client == nil {
		return nil, fmt.Errorf("default client not found")
	}

	search := client.NewSearch()
	return search.Multi(ctx, query, page)
}

// QuickMovieDetails 快速获取电影详情（使用默认客户端）
func QuickMovieDetails(ctx context.Context, movieID int64) (*tmdb.MovieDetails, error) {
	client := GetGlobalClient("default")
	if client == nil {
		return nil, fmt.Errorf("default client not found")
	}

	movie := client.NewMovie()
	return movie.Details(ctx, movieID, "")
}

// QuickTrendingMovies 快速获取趋势电影（使用默认客户端）
func QuickTrendingMovies(ctx context.Context, timeWindow string) (*tmdb.Trending, error) {
	client := GetGlobalClient("default")
	if client == nil {
		return nil, fmt.Errorf("default client not found")
	}

	trending := client.NewTrending()
	return trending.Movie(ctx, timeWindow)
}