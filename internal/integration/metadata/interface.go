package metadata

import "context"

// MediaKind 媒体类型（电影/剧集）
type MediaKind string

const (
	MediaKindMovie MediaKind = "movie"
	MediaKindTV    MediaKind = "tv"
)

// Language 语言代码（ISO-639-1）
type Language string

// ProviderName 元数据提供方
type ProviderName string

const (
	ProviderTMDB   ProviderName = "tmdb"
	ProviderTVDB   ProviderName = "tvdb"
	ProviderDouban ProviderName = "douban"
)

// MovieInfo 电影信息（统一结构）
type MovieInfo struct {
	ID          string       `json:"id"`
	Provider    ProviderName `json:"provider"`
	Title       string       `json:"title"`
	Original    string       `json:"original_title,omitempty"`
	Year        int          `json:"year,omitempty"`
	Overview    string       `json:"overview,omitempty"`
	PosterURL   string       `json:"poster_url,omitempty"`
	BackdropURL string       `json:"backdrop_url,omitempty"`
	TMDBID      *int         `json:"tmdb_id,omitempty"`
	IMDBID      *string      `json:"imdb_id,omitempty"`
	DoubanID    *string      `json:"douban_id,omitempty"`
}

// TVSeasonInfo 剧集季信息
type TVSeasonInfo struct {
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	PosterURL    string `json:"poster_url,omitempty"`
}

// TVEpisodeInfo 剧集单集信息
type TVEpisodeInfo struct {
	SeasonNumber  int    `json:"season_number"`
	EpisodeNumber int    `json:"episode_number"`
	Title         string `json:"title"`
	Overview      string `json:"overview,omitempty"`
	AirDate       string `json:"air_date,omitempty"`
}

// TVShowInfo 剧集信息
type TVShowInfo struct {
	ID        string         `json:"id"`
	Provider  ProviderName   `json:"provider"`
	Title     string         `json:"title"`
	Original  string         `json:"original_title,omitempty"`
	Year      int            `json:"year,omitempty"`
	Overview  string         `json:"overview,omitempty"`
	PosterURL string         `json:"poster_url,omitempty"`
	TMDBID    *int           `json:"tmdb_id,omitempty"`
	TVDBID    *int           `json:"tvdb_id,omitempty"`
	DoubanID  *string        `json:"douban_id,omitempty"`
	Seasons   []TVSeasonInfo `json:"seasons,omitempty"`
}

// SearchOptions 搜索可选项
type SearchOptions struct {
	Language Language `json:"language,omitempty"`
	Year     int      `json:"year,omitempty"`
	Page     int      `json:"page,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// MetadataProvider 统一元数据接口
// TMDB/TVDB/豆瓣均需实现

type MetadataProvider interface {
	Name() ProviderName

	// 搜索电影
	SearchMovie(ctx context.Context, keyword string, opts SearchOptions) ([]*MovieInfo, error)

	// 搜索剧集
	SearchTV(ctx context.Context, keyword string, opts SearchOptions) ([]*TVShowInfo, error)

	// 通过 TMDB ID 获取电影	ex: provider 可忽略或用于校验
	GetMovieByTMDB(ctx context.Context, tmdbID int, lang Language) (*MovieInfo, error)

	// 通过 TMDB ID 获取剧集
	GetTVByTMDB(ctx context.Context, tmdbID int, lang Language) (*TVShowInfo, error)

	// 通过本方 ID 获取电影
	GetMovieByID(ctx context.Context, id string, lang Language) (*MovieInfo, error)

	// 通过本方 ID 获取剧集
	GetTVByID(ctx context.Context, id string, lang Language) (*TVShowInfo, error)
}

// Factory 元数据提供方工厂
type Factory struct {
	providers map[ProviderName]MetadataProvider
}

// NewFactory 创建工厂
func NewFactory() *Factory {
	return &Factory{providers: make(map[ProviderName]MetadataProvider)}
}

// Register 注册提供方
func (f *Factory) Register(p MetadataProvider) {
	f.providers[p.Name()] = p
}

// Get 获取指定提供方
func (f *Factory) Get(name ProviderName) (MetadataProvider, bool) {
	p, ok := f.providers[name]
	return p, ok
}

// List 列出所有已注册的提供方
func (f *Factory) List() []ProviderName {
	names := make([]ProviderName, 0, len(f.providers))
	for k := range f.providers {
		names = append(names, k)
	}
	return names
}
