package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/cache"
)

// Client TMDB API客户端
type Client struct {
	apiKey       string
	baseURL      string
	imageBaseURL string
	httpClient   *http.Client
	cache        cache.Cache
	limiter      *rateLimiter
	logger       *zap.Logger

	// 配置
	timeout    time.Duration
	maxRetries int
}

// Config TMDB客户端配置
type Config struct {
	APIKey       string
	BaseURL      string
	ImageBaseURL string
	Timeout      time.Duration
	MaxRetries   int
	Cache        cache.Cache
	Logger       *zap.Logger
}

// NewClient 创建TMDB客户端
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.themoviedb.org/3"
	}
	if config.ImageBaseURL == "" {
		config.ImageBaseURL = "https://image.tmdb.org/t/p/w500"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &Client{
		apiKey:       config.APIKey,
		baseURL:      config.BaseURL,
		imageBaseURL: config.ImageBaseURL,
		httpClient:   &http.Client{Timeout: config.Timeout},
		cache:        config.Cache,
		limiter:      newRateLimiter(40, 10*time.Second), // 40 req/10s
		logger:       config.Logger,
		timeout:      config.Timeout,
		maxRetries:   config.MaxRetries,
	}
}

// rateLimiter 限流器
type rateLimiter struct {
	tokens     float64
	capacity   float64
	lastRefill time.Time
	mutex      sync.Mutex
	rate       float64
	interval   time.Duration
}

func newRateLimiter(capacity int, interval time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:     float64(capacity),
		capacity:   float64(capacity),
		lastRefill: time.Now(),
		rate:       float64(capacity) / interval.Seconds(),
		interval:   interval,
	}
}

func (r *rateLimiter) allow() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.tokens += elapsed * r.rate
	if r.tokens > r.capacity {
		r.tokens = r.capacity
	}
	r.lastRefill = now

	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

// doRequest 执行API请求
func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values) (*http.Response, error) {
	// 等待限流
	for !c.limiter.allow() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}

	// 添加API密钥
	if params == nil {
		params = url.Values{}
	}
	params.Set("api_key", c.apiKey)

	// 构建完整URL
	fullURL := fmt.Sprintf("%s/%s?%s", c.baseURL, endpoint, params.Encode())

	var lastErr error
	for i := 0; i <= c.maxRetries; i++ {
		if i > 0 {
			// 指数退避
			backoff := time.Duration(1<<uint(i-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		resp.Body.Close()

		// 对于某些错误码不重试
		if resp.StatusCode == 401 || resp.StatusCode == 404 {
			return nil, fmt.Errorf("TMDB API error: %d", resp.StatusCode)
		}

		lastErr = fmt.Errorf("TMDB API error: %d", resp.StatusCode)
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// getCacheKey 生成缓存键
func (c *Client) getCacheKey(endpoint string, params url.Values) string {
	return fmt.Sprintf("tmdb:%s:%s", endpoint, params.Encode())
}

// getWithCache 带缓存的GET请求
func (c *Client) getWithCache(ctx context.Context, endpoint string, params url.Values, dest any, ttl time.Duration) error {
	if c.cache != nil {
		cacheKey := c.getCacheKey(endpoint, params)
		err := c.cache.GetJSON(ctx, cacheKey, dest)
		if err == nil {
			if c.logger != nil {
				c.logger.Debug("TMDB cache hit", zap.String("endpoint", endpoint), zap.String("cache_key", cacheKey))
			}
			return nil
		}
	}

	// 缓存未命中，请求API
	resp, err := c.doRequest(ctx, endpoint, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// 存入缓存
	if c.cache != nil {
		cacheKey := c.getCacheKey(endpoint, params)
		if err := c.cache.SetJSON(ctx, cacheKey, dest, ttl); err != nil && c.logger != nil {
			c.logger.Warn("failed to cache TMDB response", zap.Error(err), zap.String("cache_key", cacheKey))
		}
	}

	return nil
}

// buildImageURL 构建图片URL
func (c *Client) buildImageURL(path string, size ...string) string {
	if path == "" {
		return ""
	}

	imageSize := "w500"
	if len(size) > 0 && size[0] != "" {
		imageSize = size[0]
	}

	return fmt.Sprintf("%s/t/p/%s%s", c.imageBaseURL, imageSize, path)
}

// BuildImageURL 公开的图片URL构建方法
func (c *Client) BuildImageURL(path string, size ...string) string {
	return c.buildImageURL(path, size...)
}

// SearchMulti 搜索多媒体
func (c *Client) SearchMulti(ctx context.Context, query string, page int) (*MultiSearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("include_adult", "false")
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	var resp MultiSearchResponse
	err := c.getWithCache(ctx, "search/multi", params, &resp, time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SearchMovies 搜索电影（接口兼容方法）
func (c *Client) SearchMovies(ctx context.Context, query string, page int) (*MovieSearchResponse, error) {
	return c.SearchMovie(ctx, query, 0, page)
}

// SearchTVShows 搜索电视剧（接口兼容方法）
func (c *Client) SearchTVShows(ctx context.Context, query string, page int) (*TVSearchResponse, error) {
	return c.SearchTV(ctx, query, 0, page)
}

// SearchPeople 搜索人物
func (c *Client) SearchPeople(ctx context.Context, query string, page int) (*PersonSearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("include_adult", "false")
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	var resp PersonSearchResponse
	err := c.getWithCache(ctx, "search/person", params, &resp, time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SearchMovie 搜索电影
func (c *Client) SearchMovie(ctx context.Context, query string, year int, page int) (*MovieSearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("include_adult", "false")
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if year > 0 {
		params.Set("year", strconv.Itoa(year))
	}

	var resp MovieSearchResponse
	err := c.getWithCache(ctx, "search/movie", params, &resp, time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SearchTV 搜索电视剧
func (c *Client) SearchTV(ctx context.Context, query string, year int, page int) (*TVSearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("include_adult", "false")
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if year > 0 {
		params.Set("first_air_date_year", strconv.Itoa(year))
	}

	var resp TVSearchResponse
	err := c.getWithCache(ctx, "search/tv", params, &resp, time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetMovieDetails 获取电影详情
func (c *Client) GetMovieDetails(ctx context.Context, id int, language ...string) (*MovieDetails, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("movie/%d", id)
	var resp MovieDetails
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetTVDetails 获取电视剧详情
func (c *Client) GetTVDetails(ctx context.Context, id int, language ...string) (*TVDetails, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("tv/%d", id)
	var resp TVDetails
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetSeasonDetails 获取季详情
func (c *Client) GetSeasonDetails(ctx context.Context, tvID, seasonNum int, language ...string) (*SeasonDetails, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("tv/%d/season/%d", tvID, seasonNum)
	var resp SeasonDetails
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetEpisodeDetails 获取集详情
func (c *Client) GetEpisodeDetails(ctx context.Context, tvID, seasonNum, episodeNum int, language ...string) (*EpisodeDetails, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("tv/%d/season/%d/episode/%d", tvID, seasonNum, episodeNum)
	var resp EpisodeDetails
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetMovieCredits 获取电影演职员
func (c *Client) GetMovieCredits(ctx context.Context, id int) (*Credits, error) {
	endpoint := fmt.Sprintf("movie/%d/credits", id)
	var resp Credits
	err := c.getWithCache(ctx, endpoint, nil, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetTVCredits 获取电视剧演职员
func (c *Client) GetTVCredits(ctx context.Context, id int, language ...string) (*Credits, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("tv/%d/credits", id)
	var resp Credits
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetTrending 获取趋势内容
func (c *Client) GetTrending(ctx context.Context, mediaType, timeWindow string, page int) (*TrendingResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	endpoint := fmt.Sprintf("trending/%s/%s", mediaType, timeWindow)
	var resp TrendingResponse
	err := c.getWithCache(ctx, endpoint, params, &resp, 2*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetMovieImages 获取电影图片
func (c *Client) GetMovieImages(ctx context.Context, id int, language ...string) (*MovieImages, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("movie/%d/images", id)
	var resp MovieImages
	err := c.getWithCache(ctx, endpoint, params, &resp, 6*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetTVImages 获取电视剧图片
func (c *Client) GetTVImages(ctx context.Context, id int, language ...string) (*TVImages, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("tv/%d/images", id)
	var resp TVImages
	err := c.getWithCache(ctx, endpoint, params, &resp, 6*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPersonImages 获取人物图片
func (c *Client) GetPersonImages(ctx context.Context, id int) (*PersonImages, error) {
	endpoint := fmt.Sprintf("person/%d/images", id)
	var resp PersonImages
	err := c.getWithCache(ctx, endpoint, nil, &resp, 6*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPersonDetails 获取人物详情
func (c *Client) GetPersonDetails(ctx context.Context, id int, language ...string) (*PersonDetails, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("person/%d", id)
	var resp PersonDetails
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPersonMovieCredits 获取人物电影作品
func (c *Client) GetPersonMovieCredits(ctx context.Context, id int, language ...string) (*PersonMovieCredits, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("person/%d/movie_credits", id)
	var resp PersonMovieCredits
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPersonTVCredits 获取人物电视剧作品
func (c *Client) GetPersonTVCredits(ctx context.Context, id int, language ...string) (*PersonTVCredits, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("person/%d/tv_credits", id)
	var resp PersonTVCredits
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPersonCombinedCredits 获取人物所有作品
func (c *Client) GetPersonCombinedCredits(ctx context.Context, id int, language ...string) (*PersonCombinedCredits, error) {
	params := url.Values{}
	if len(language) > 0 && language[0] != "" {
		params.Set("language", language[0])
	}

	endpoint := fmt.Sprintf("person/%d/combined_credits", id)
	var resp PersonCombinedCredits
	err := c.getWithCache(ctx, endpoint, params, &resp, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetConfiguration 获取TMDB配置
func (c *Client) GetConfiguration(ctx context.Context) (*Configuration, error) {
	endpoint := "configuration"
	var resp Configuration
	err := c.getWithCache(ctx, endpoint, nil, &resp, 7*24*time.Hour) // 缓存7天
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// DiscoverMovies 发现电影
func (c *Client) DiscoverMovies(ctx context.Context, params *DiscoverParams) (*MovieSearchResponse, error) {
	if params == nil {
		params = &DiscoverParams{}
	}

	values := url.Values{}

	// 基础参数
	if params.Language != "" {
		values.Set("language", params.Language)
	}
	if params.Region != "" {
		values.Set("region", params.Region)
	}
	if params.SortBy != "" {
		values.Set("sort_by", params.SortBy)
	}
	if params.Page > 0 {
		values.Set("page", strconv.Itoa(params.Page))
	}

	// 日期范围
	if params.ReleaseDateGTE != "" {
		values.Set("release_date.gte", params.ReleaseDateGTE)
	}
	if params.ReleaseDateLTE != "" {
		values.Set("release_date.lte", params.ReleaseDateLTE)
	}

	// 类型过滤
	if len(params.WithGenres) > 0 {
		values.Set("with_genres", params.WithGenres)
	}
	if len(params.WithoutGenres) > 0 {
		values.Set("without_genres", params.WithoutGenres)
	}
	if len(params.WithCompanies) > 0 {
		values.Set("with_companies", params.WithCompanies)
	}
	if len(params.WithPeople) > 0 {
		values.Set("with_people", params.WithPeople)
	}

	// 注意：DiscoverParams 没有 VoteAverageGTE 和 VoteCountGTE 字段
	// 如果需要评分过滤，可以添加这些字段到结构体中

	// 运行时间过滤
	if params.WithRuntimeGTE > 0 {
		values.Set("with_runtime.gte", strconv.Itoa(params.WithRuntimeGTE))
	}
	if params.WithRuntimeLTE > 0 {
		values.Set("with_runtime.lte", strconv.Itoa(params.WithRuntimeLTE))
	}

	endpoint := "discover/movie"
	var resp MovieSearchResponse
	err := c.getWithCache(ctx, endpoint, values, &resp, 2*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// DiscoverTVShows 发现电视剧
func (c *Client) DiscoverTVShows(ctx context.Context, params *DiscoverParams) (*TVSearchResponse, error) {
	if params == nil {
		params = &DiscoverParams{}
	}

	values := url.Values{}

	// 基础参数
	if params.Language != "" {
		values.Set("language", params.Language)
	}
	if params.Region != "" {
		values.Set("region", params.Region)
	}
	if params.SortBy != "" {
		values.Set("sort_by", params.SortBy)
	}
	if params.Page > 0 {
		values.Set("page", strconv.Itoa(params.Page))
	}

	// 日期范围 - 注意：DiscoverParams 没有 FirstAirDateGTE/LTE 字段
	// 可以使用 PrimaryReleaseDateGTE/LTE 作为替代
	if params.PrimaryReleaseDateGTE != "" {
		values.Set("first_air_date.gte", params.PrimaryReleaseDateGTE)
	}
	if params.PrimaryReleaseDateLTE != "" {
		values.Set("first_air_date.lte", params.PrimaryReleaseDateLTE)
	}

	// 类型过滤
	if len(params.WithGenres) > 0 {
		values.Set("with_genres", params.WithGenres)
	}
	if len(params.WithoutGenres) > 0 {
		values.Set("without_genres", params.WithoutGenres)
	}
	if len(params.WithNetworks) > 0 {
		values.Set("with_networks", params.WithNetworks)
	}
	if len(params.WithCompanies) > 0 {
		values.Set("with_companies", params.WithCompanies)
	}
	if len(params.WithPeople) > 0 {
		values.Set("with_people", params.WithPeople)
	}

	// 注意：DiscoverParams 没有 VoteAverageGTE 和 VoteCountGTE 字段
	// 如果需要评分过滤，可以添加这些字段到结构体中

	// 注意：DiscoverParams 没有 IncludeNullFirstAirDates 和 ScreenedTheatrically 字段
	// 如果需要这些功能，可以添加到结构体中

	endpoint := "discover/tv"
	var resp TVSearchResponse
	err := c.getWithCache(ctx, endpoint, values, &resp, 2*time.Hour)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
