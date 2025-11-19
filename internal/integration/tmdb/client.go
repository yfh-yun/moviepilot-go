// Package tmdb TMDB API集成客户端
// 提供电影数据库(The Movie Database)的API访问功能
package tmdb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	// DefaultBaseURL TMDB API默认基础URL
	DefaultBaseURL = "https://api.themoviedb.org/3"
	// DefaultImageBaseURL TMDB图片默认基础URL
	DefaultImageBaseURL = "https://image.tmdb.org/t/p"
)

// Client TMDB API客户端
type Client struct {
	httpClient *httpclient.Client
	apiKey     string
	language   string
	logger     *zap.Logger
}

// NewClient 创建TMDB客户端
// apiKey: TMDB API密钥
// 返回: TMDB客户端实例
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = viper.GetString("tmdb.api_key")
	}

	language := viper.GetString("tmdb.language")
	if language == "" {
		language = "zh-CN"
	}

	client := &Client{
		httpClient: httpclient.NewClient(httpclient.Options{
			BaseURL: DefaultBaseURL,
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Logger: logger.Logger,
		}),
		apiKey:   apiKey,
		language: language,
		logger:   logger.Logger,
	}

	return client
}

// MovieDetails 电影详情响应
type MovieDetails struct {
	ID                  int64               `json:"id"`
	IMDBID              string              `json:"imdb_id"`
	Title               string              `json:"title"`
	OriginalTitle       string              `json:"original_title"`
	Overview            string              `json:"overview"`
	ReleaseDate         string              `json:"release_date"`
	PosterPath          string              `json:"poster_path"`
	BackdropPath        string              `json:"backdrop_path"`
	VoteAverage         float64             `json:"vote_average"`
	VoteCount           int64               `json:"vote_count"`
	Popularity          float64             `json:"popularity"`
	Adult               bool                `json:"adult"`
	Budget              int64               `json:"budget"`
	Revenue             int64               `json:"revenue"`
	Runtime             int                 `json:"runtime"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	Genres              []Genre             `json:"genres"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
}

// TVDetails 电视剧详情响应
type TVDetails struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	OriginalName     string    `json:"original_name"`
	Overview         string    `json:"overview"`
	FirstAirDate     string    `json:"first_air_date"`
	LastAirDate      string    `json:"last_air_date"`
	PosterPath       string    `json:"poster_path"`
	BackdropPath     string    `json:"backdrop_path"`
	VoteAverage      float64   `json:"vote_average"`
	VoteCount        int64     `json:"vote_count"`
	Popularity       float64   `json:"popularity"`
	NumberOfSeasons  int       `json:"number_of_seasons"`
	NumberOfEpisodes int       `json:"number_of_episodes"`
	Status           string    `json:"status"`
	Type             string    `json:"type"`
	Genres           []Genre   `json:"genres"`
	Networks         []Network `json:"networks"`
}

// Genre 类型
type Genre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ProductionCompany 制作公司
type ProductionCompany struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

// Network 电视网络
type Network struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

// Account 账户信息
type Account struct {
	ID              int64      `json:"id"`
	Username        string     `json:"username"`
	Name            string     `json:"name"`
	IncludeAdult    bool       `json:"include_adult"`
	Language        string     `json:"language"`
	Country         string     `json:"country"`
	Avatar          Avatar     `json:"avatar"`
}

// Avatar 头像
type Avatar struct {
	Gravatar     AvatarInfo `json:"gravatar"`
	Tmdb         AvatarInfo `json:"tmdb"`
}

// AvatarInfo 头像信息
type AvatarInfo struct {
	Hash string `json:"hash"`
}

// Credits 演职员信息
type Credits struct {
	ID      int64    `json:"id"`
	Cast    []Cast   `json:"cast"`
	Crew    []Crew   `json:"crew"`
}

// Cast 演员信息
type Cast struct {
	CastID        int64  `json:"cast_id"`
	Character     string `json:"character"`
	CreditID      string `json:"credit_id"`
	Gender        int    `json:"gender"`
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Order         int    `json:"order"`
	ProfilePath   string `json:"profile_path"`
}

// Crew 工作人员信息
type Crew struct {
	CreditID    string `json:"credit_id"`
	Department  string `json:"department"`
	Gender      int    `json:"gender"`
	ID          int64  `json:"id"`
	Job         string `json:"job"`
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
}

// Videos 视频信息
type Videos struct {
	ID      int64   `json:"id"`
	Results []Video `json:"results"`
}

// Video 视频信息
type Video struct {
	ID        string `json:"id"`
	Iso639_1  string `json:"iso_639_1"`
	Iso3166_1 string `json:"iso_3166_1"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	Official  bool   `json:"official"`
	PublishedAt string `json:"published_at"`
	Site      string `json:"site"`
	Size      int    `json:"size"`
	Type      string `json:"type"`
}

// ExternalIDs 外部ID
type ExternalIDs struct {
	IMDBID     string `json:"imdb_id"`
	FreebaseMid string `json:"freebase_mid"`
	FreebaseID  string `json:"freebase_id"`
	TVDBID     int64  `json:"tvdb_id"`
	TVRageID   int64  `json:"tvrage_id"`
	FacebookID string `json:"facebook_id"`
	TwitterID  string `json:"twitter_id"`
	InstagramID string `json:"instagram_id"`
}

// Keywords 关键词
type Keywords struct {
	ID       int64     `json:"id"`
	Keywords []Keyword `json:"keywords"`
}

// Keyword 关键词
type Keyword struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Reviews 评论
type Reviews struct {
	ID      int64    `json:"id"`
	Page    int       `json:"page"`
	Results []Review  `json:"results"`
	TotalPages int    `json:"total_pages"`
	TotalResults int  `json:"total_results"`
}

// Review 评论
type Review struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
	AuthorDetails AuthorDetails `json:"author_details"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	URL      string `json:"url"`
}

// AuthorDetails 作者详情
type AuthorDetails struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarPath string `json:"avatar_path"`
	Rating    float64 `json:"rating"`
}

// Similar 相似媒体
type Similar struct {
	Page          int          `json:"page"`
	Results      []SearchItem `json:"results"`
	TotalPages   int          `json:"total_pages"`
	TotalResults int          `json:"total_results"`
}

// Recommendations 推荐
type Recommendations struct {
	Page          int          `json:"page"`
	Results      []SearchItem `json:"results"`
	TotalPages   int          `json:"total_pages"`
	TotalResults int          `json:"total_results"`
}

// Trending 趋势
type Trending struct {
	Page          int          `json:"page"`
	Results      []SearchItem `json:"results"`
	TotalPages   int          `json:"total_pages"`
	TotalResults int          `json:"total_results"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Page         int          `json:"page"`
	Results      []SearchItem `json:"results"`
	TotalPages   int          `json:"total_pages"`
	TotalResults int          `json:"total_results"`
}

// SearchItem 搜索条目
type SearchItem struct {
	ID            int64   `json:"id"`
	MediaType     string  `json:"media_type"`
	Title         string  `json:"title"` // 电影标题
	Name          string  `json:"name"`  // 电视剧名称
	OriginalTitle string  `json:"original_title"`
	OriginalName  string  `json:"original_name"`
	Overview      string  `json:"overview"`
	ReleaseDate   string  `json:"release_date"`   // 电影
	FirstAirDate  string  `json:"first_air_date"` // 电视剧
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	VoteAverage   float64 `json:"vote_average"`
	Popularity    float64 `json:"popularity"`
}

// GetMovieDetails 获取电影详情
// ctx: 上下文
// movieID: 电影ID
// 返回: 电影详情和错误信息
func (c *Client) GetMovieDetails(ctx context.Context, movieID int64) (*MovieDetails, error) {
	path := fmt.Sprintf("/movie/%d?api_key=%s&language=%s", movieID, c.apiKey, c.language)

	var result MovieDetails
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get movie details failed",
			zap.Int64("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get movie details: %w", err)
	}

	c.logger.Info("Get movie details success",
		zap.Int64("movie_id", movieID),
		zap.String("title", result.Title),
	)

	return &result, nil
}

// GetTVDetails 获取电视剧详情
// ctx: 上下文
// tvID: 电视剧ID
// 返回: 电视剧详情和错误信息
func (c *Client) GetTVDetails(ctx context.Context, tvID int64) (*TVDetails, error) {
	path := fmt.Sprintf("/tv/%d?api_key=%s&language=%s", tvID, c.apiKey, c.language)

	var result TVDetails
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get TV details failed",
			zap.Int64("tv_id", tvID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get TV details: %w", err)
	}

	c.logger.Info("Get TV details success",
		zap.Int64("tv_id", tvID),
		zap.String("name", result.Name),
	)

	return &result, nil
}

// SearchMulti 多类型搜索（电影、电视剧、人物）
// ctx: 上下文
// query: 搜索关键词
// page: 页码（从1开始）
// 返回: 搜索结果和错误信息
func (c *Client) SearchMulti(ctx context.Context, query string, page int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}

	path := fmt.Sprintf("/search/multi?api_key=%s&language=%s&query=%s&page=%d",
		c.apiKey, c.language, url.QueryEscape(query), page)

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Search multi failed",
			zap.String("query", query),
			zap.Int("page", page),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search multi: %w", err)
	}

	c.logger.Info("Search multi success",
		zap.String("query", query),
		zap.Int("page", page),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// SearchMovie 搜索电影
// ctx: 上下文
// query: 搜索关键词
// year: 年份（可选，0表示不限）
// page: 页码（从1开始）
// 返回: 搜索结果和错误信息
func (c *Client) SearchMovie(ctx context.Context, query string, year int, page int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}

	path := fmt.Sprintf("/search/movie?api_key=%s&language=%s&query=%s&page=%d",
		c.apiKey, c.language, url.QueryEscape(query), page)

	if year > 0 {
		path += fmt.Sprintf("&year=%d", year)
	}

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Search movie failed",
			zap.String("query", query),
			zap.Int("year", year),
			zap.Int("page", page),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search movie: %w", err)
	}

	c.logger.Info("Search movie success",
		zap.String("query", query),
		zap.Int("year", year),
		zap.Int("page", page),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// SearchTV 搜索电视剧
// ctx: 上下文
// query: 搜索关键词
// firstAirDateYear: 首播年份（可选，0表示不限）
// page: 页码（从1开始）
// 返回: 搜索结果和错误信息
func (c *Client) SearchTV(ctx context.Context, query string, firstAirDateYear int, page int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}

	path := fmt.Sprintf("/search/tv?api_key=%s&language=%s&query=%s&page=%d",
		c.apiKey, c.language, url.QueryEscape(query), page)

	if firstAirDateYear > 0 {
		path += fmt.Sprintf("&first_air_date_year=%d", firstAirDateYear)
	}

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Search TV failed",
			zap.String("query", query),
			zap.Int("year", firstAirDateYear),
			zap.Int("page", page),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search TV: %w", err)
	}

	c.logger.Info("Search TV success",
		zap.String("query", query),
		zap.Int("year", firstAirDateYear),
		zap.Int("page", page),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// GetImageURL 获取图片完整URL
// imagePath: 图片路径（从TMDB API返回）
// size: 图片尺寸（如"w500", "original"等）
// 返回: 完整的图片URL
func (c *Client) GetImageURL(imagePath, size string) string {
	if imagePath == "" {
		return ""
	}

	if size == "" {
		size = "original"
	}

	return fmt.Sprintf("%s/%s%s", DefaultImageBaseURL, size, imagePath)
}

// GetPosterURL 获取海报完整URL
// posterPath: 海报路径
// size: 尺寸（默认w500）
// 返回: 完整的海报URL
func (c *Client) GetPosterURL(posterPath string, size string) string {
	if size == "" {
		size = "w500"
	}
	return c.GetImageURL(posterPath, size)
}

// GetBackdropURL 获取背景图完整URL
// backdropPath: 背景图路径
// size: 尺寸（默认w1280）
// 返回: 完整的背景图URL
func (c *Client) GetBackdropURL(backdropPath string, size string) string {
	if size == "" {
		size = "w1280"
	}
	return c.GetImageURL(backdropPath, size)
}

// ParseTMDBID 从URL解析TMDB ID
// tmdbURL: TMDB网址
// 返回: 媒体类型（movie/tv）、ID和错误信息
func ParseTMDBID(tmdbURL string) (mediaType string, id int64, err error) {
	// 示例: https://www.themoviedb.org/movie/12345
	// 示例: https://www.themoviedb.org/tv/67890

	u, err := url.Parse(tmdbURL)
	if err != nil {
		return "", 0, fmt.Errorf("parse URL: %w", err)
	}

	// 解析路径，格式为 /movie/12345 或 /tv/67890
	parts := splitPath(u.Path)
	if len(parts) < 2 {
		return "", 0, fmt.Errorf("invalid TMDB URL format")
	}

	mediaType = parts[0]
	if mediaType != "movie" && mediaType != "tv" {
		return "", 0, fmt.Errorf("unsupported media type: %s", mediaType)
	}

	id, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("parse ID: %w", err)
	}

	return mediaType, id, nil
}

// GetMovieCredits 获取电影演职员信息
// ctx: 上下文
// movieID: 电影ID
// 返回: 演职员信息和错误信息
func (c *Client) GetMovieCredits(ctx context.Context, movieID int64) (*Credits, error) {
	path := fmt.Sprintf("/movie/%d/credits?api_key=%s&language=%s", movieID, c.apiKey, c.language)

	var result Credits
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get movie credits failed",
			zap.Int64("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get movie credits: %w", err)
	}

	c.logger.Info("Get movie credits success",
		zap.Int64("movie_id", movieID),
		zap.Int("cast_count", len(result.Cast)),
		zap.Int("crew_count", len(result.Crew)),
	)

	return &result, nil
}

// GetMovieVideos 获取电影视频信息
// ctx: 上下文
// movieID: 电影ID
// 返回: 视频信息和错误信息
func (c *Client) GetMovieVideos(ctx context.Context, movieID int64) (*Videos, error) {
	path := fmt.Sprintf("/movie/%d/videos?api_key=%s&language=%s", movieID, c.apiKey, c.language)

	var result Videos
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get movie videos failed",
			zap.Int64("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get movie videos: %w", err)
	}

	c.logger.Info("Get movie videos success",
		zap.Int64("movie_id", movieID),
		zap.Int("video_count", len(result.Results)),
	)

	return &result, nil
}

// GetMovieExternalIDs 获取电影外部ID
// ctx: 上下文
// movieID: 电影ID
// 返回: 外部ID和错误信息
func (c *Client) GetMovieExternalIDs(ctx context.Context, movieID int64) (*ExternalIDs, error) {
	path := fmt.Sprintf("/movie/%d/external_ids?api_key=%s", movieID, c.apiKey)

	var result ExternalIDs
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get movie external IDs failed",
			zap.Int64("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get movie external IDs: %w", err)
	}

	c.logger.Info("Get movie external IDs success",
		zap.Int64("movie_id", movieID),
		zap.String("imdb_id", result.IMDBID),
	)

	return &result, nil
}

// GetMovieKeywords 获取电影关键词
// ctx: 上下文
// movieID: 电影ID
// 返回: 关键词和错误信息
func (c *Client) GetMovieKeywords(ctx context.Context, movieID int64) (*Keywords, error) {
	path := fmt.Sprintf("/movie/%d/keywords?api_key=%s", movieID, c.apiKey)

	var result Keywords
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get movie keywords failed",
			zap.Int64("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get movie keywords: %w", err)
	}

	c.logger.Info("Get movie keywords success",
		zap.Int64("movie_id", movieID),
		zap.Int("keyword_count", len(result.Keywords)),
	)

	return &result, nil
}

// GetMovieSimilar 获取相似电影
// ctx: 上下文
// movieID: 电影ID
// page: 页码（从1开始）
// 返回: 相似电影和错误信息
func (c *Client) GetMovieSimilar(ctx context.Context, movieID int64, page int) (*Similar, error) {
	if page < 1 {
		page = 1
	}

	path := fmt.Sprintf("/movie/%d/similar?api_key=%s&language=%s&page=%d", movieID, c.apiKey, c.language, page)

	var result Similar
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get movie similar failed",
			zap.Int64("movie_id", movieID),
			zap.Int("page", page),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get movie similar: %w", err)
	}

	c.logger.Info("Get movie similar success",
		zap.Int64("movie_id", movieID),
		zap.Int("page", page),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// GetMovieRecommendations 获取电影推荐
// ctx: 上下文
// movieID: 电影ID
// page: 页码（从1开始）
// 返回: 推荐电影和错误信息
func (c *Client) GetMovieRecommendations(ctx context.Context, movieID int64, page int) (*Recommendations, error) {
	if page < 1 {
		page = 1
	}

	path := fmt.Sprintf("/movie/%d/recommendations?api_key=%s&language=%s&page=%d", movieID, c.apiKey, c.language, page)

	var result Recommendations
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get movie recommendations failed",
			zap.Int64("movie_id", movieID),
			zap.Int("page", page),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get movie recommendations: %w", err)
	}

	c.logger.Info("Get movie recommendations success",
		zap.Int64("movie_id", movieID),
		zap.Int("page", page),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// GetTrendingMovies 获取趋势电影
// ctx: 上下文
// timeWindow: 时间窗口 ("day", "week")
// 返回: 趋势电影和错误信息
func (c *Client) GetTrendingMovies(ctx context.Context, timeWindow string) (*Trending, error) {
	if timeWindow == "" {
		timeWindow = "day"
	}

	path := fmt.Sprintf("/trending/movie/%s?api_key=%s&language=%s", timeWindow, c.apiKey, c.language)

	var result Trending
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get trending movies failed",
			zap.String("time_window", timeWindow),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get trending movies: %w", err)
	}

	c.logger.Info("Get trending movies success",
		zap.String("time_window", timeWindow),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// GetTrendingTV 获取趋势电视剧
// ctx: 上下文
// timeWindow: 时间窗口 ("day", "week")
// 返回: 趋势电视剧和错误信息
func (c *Client) GetTrendingTV(ctx context.Context, timeWindow string) (*Trending, error) {
	if timeWindow == "" {
		timeWindow = "day"
	}

	path := fmt.Sprintf("/trending/tv/%s?api_key=%s&language=%s", timeWindow, c.apiKey, c.language)

	var result Trending
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get trending TV failed",
			zap.String("time_window", timeWindow),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get trending TV: %w", err)
	}

	c.logger.Info("Get trending TV success",
		zap.String("time_window", timeWindow),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// GetPopularMovies 获取热门电影
// ctx: 上下文
// page: 页码（从1开始）
// 返回: 热门电影和错误信息
func (c *Client) GetPopularMovies(ctx context.Context, page int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}

	path := fmt.Sprintf("/movie/popular?api_key=%s&language=%s&page=%d", c.apiKey, c.language, page)

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get popular movies failed",
			zap.Int("page", page),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get popular movies: %w", err)
	}

	c.logger.Info("Get popular movies success",
		zap.Int("page", page),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// GetPopularTV 获取热门电视剧
// ctx: 上下文
// page: 页码（从1开始）
// 返回: 热门电视剧和错误信息
func (c *Client) GetPopularTV(ctx context.Context, page int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}

	path := fmt.Sprintf("/tv/popular?api_key=%s&language=%s&page=%d", c.apiKey, c.language, page)

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get popular TV failed",
			zap.Int("page", page),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get popular TV: %w", err)
	}

	c.logger.Info("Get popular TV success",
		zap.Int("page", page),
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// DiscoverMovies 发现电影
// ctx: 上下文
// params: 发现参数
// 返回: 发现结果和错误信息
func (c *Client) DiscoverMovies(ctx context.Context, params map[string]string) (*SearchResult, error) {
	path := fmt.Sprintf("/discover/movie?api_key=%s&language=%s", c.apiKey, c.language)

	for key, value := range params {
		path += fmt.Sprintf("&%s=%s", key, url.QueryEscape(value))
	}

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Discover movies failed",
			zap.Any("params", params),
			zap.Error(err),
		)
		return nil, fmt.Errorf("discover movies: %w", err)
	}

	c.logger.Info("Discover movies success",
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// DiscoverTV 发现电视剧
// ctx: 上下文
// params: 发现参数
// 返回: 发现结果和错误信息
func (c *Client) DiscoverTV(ctx context.Context, params map[string]string) (*SearchResult, error) {
	path := fmt.Sprintf("/discover/tv?api_key=%s&language=%s", c.apiKey, c.language)

	for key, value := range params {
		path += fmt.Sprintf("&%s=%s", key, url.QueryEscape(value))
	}

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Discover TV failed",
			zap.Any("params", params),
			zap.Error(err),
		)
		return nil, fmt.Errorf("discover TV: %w", err)
	}

	c.logger.Info("Discover TV success",
		zap.Int("total_results", result.TotalResults),
	)

	return &result, nil
}

// splitPath 分割路径
func splitPath(path string) []string {
	var parts []string
	current := ""

	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
