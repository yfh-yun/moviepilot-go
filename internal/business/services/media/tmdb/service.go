package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"moviepilot-go/pkg/cache"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// SimpleTMDBService 简化的TMDB服务实现
type SimpleTMDBService struct {
	client      *http.Client
	baseURL     string
	apiKey      string
	cache       cache.Cache
	imageConfig *ImageConfig
}

// TMDBService 是 SimpleTMDBService 的类型别名，用于兼容旧代码
type TMDBService = SimpleTMDBService

// NewTMDBService 创建TMDB服务
func NewTMDBService(apiKey string, cache cache.Cache) Service {
	return &SimpleTMDBService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.themoviedb.org/3",
		apiKey:  apiKey,
		cache:   cache,
		imageConfig: &ImageConfig{
			Size:           "w500",
			BaseURL:        "https://image.tmdb.org/t/p",
			Quality:        85,
			UserAgent:      "MoviePilot-TMDB/1.0",
			MaxConcurrency: 5,
		},
	}
}

// makeRequest 发起HTTP请求
func (s *SimpleTMDBService) makeRequest(ctx context.Context, endpoint string, result any) error {
	if s.apiKey == "" {
		return fmt.Errorf("TMDB API key is required")
	}

	reqURL := fmt.Sprintf("%s%s?api_key=%s", s.baseURL, endpoint, s.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		logger.Error("Failed to create request", zap.String("error", err.Error()), zap.String("endpoint", endpoint))
		return err
	}

	logger.Debug("Making TMDB API request", zap.String("url", reqURL), zap.String("endpoint", endpoint))

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error("Failed to make request", zap.String("error", err.Error()), zap.String("endpoint", endpoint))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("TMDB API error", zap.Int("status", resp.StatusCode), zap.String("endpoint", endpoint))
		return fmt.Errorf("TMDB API error: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		logger.Error("Failed to decode response", zap.String("error", err.Error()), zap.String("endpoint", endpoint))
		return err
	}

	logger.Debug("TMDB API request successful", zap.String("endpoint", endpoint))
	return nil
}

// makeCachedRequest 发起带缓存的HTTP请求
func (s *SimpleTMDBService) makeCachedRequest(ctx context.Context, endpoint string, result any, cacheTTL time.Duration) error {
	// 检查缓存
	cacheKey := fmt.Sprintf("tmdb:%s", endpoint)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		// cache.Get 返回 string，转为 []byte
		cachedBytes := []byte(cached)
		if err := json.Unmarshal(cachedBytes, result); err == nil {
			logger.Debug("Cache hit", zap.String("key", cacheKey))
			return nil
		}
	}

	// 发起请求
	if err := s.makeRequest(ctx, endpoint, result); err != nil {
		return err
	}

	// 缓存结果
	if data, err := json.Marshal(result); err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, cacheTTL); err != nil {
			logger.Warn("Failed to cache response", zap.String("error", err.Error()), zap.String("key", cacheKey))
		}
	}

	return nil
}

// MultiSearch 多媒体搜索
func (s *SimpleTMDBService) MultiSearch(ctx context.Context, query string, page int) (*MultiSearchResponse, error) {
	endpoint := fmt.Sprintf("/search/multi&query=%s&page=%d", query, page)
	var result MultiSearchResponse
	err := s.makeCachedRequest(ctx, endpoint, &result, 5*time.Minute)
	if err != nil {
		logger.Error("Failed to search multi", zap.String("error", err.Error()), zap.String("query", query), zap.Int("page", page))
		return nil, err
	}
	return &result, nil
}

// SearchMovies 搜索电影
func (s *SimpleTMDBService) SearchMovies(ctx context.Context, query string, page int) (*MovieSearchResponse, error) {
	endpoint := fmt.Sprintf("/search/movie?query=%s&page=%d", query, page)
	var result MovieSearchResponse
	err := s.makeCachedRequest(ctx, endpoint, &result, 5*time.Minute)
	if err != nil {
		logger.Error("Failed to search movies", zap.String("error", err.Error()), zap.String("query", query), zap.Int("page", page))
		return nil, err
	}
	return &result, nil
}

// SearchTVShows 搜索电视剧
func (s *SimpleTMDBService) SearchTVShows(ctx context.Context, query string, page int) (*TVSearchResponse, error) {
	endpoint := fmt.Sprintf("/search/tv?query=%s&page=%d", query, page)
	var result TVSearchResponse
	err := s.makeCachedRequest(ctx, endpoint, &result, 5*time.Minute)
	if err != nil {
		logger.Error("Failed to search TV shows", zap.String("error", err.Error()), zap.String("query", query), zap.Int("page", page))
		return nil, err
	}
	return &result, nil
}

// SearchPeople 搜索人物
func (s *SimpleTMDBService) SearchPeople(ctx context.Context, query string, page int) (*PersonSearchResponse, error) {
	endpoint := fmt.Sprintf("/search/person?query=%s&page=%d", query, page)
	var result PersonSearchResponse
	err := s.makeCachedRequest(ctx, endpoint, &result, 5*time.Minute)
	if err != nil {
		logger.Error("Failed to search people", zap.String("error", err.Error()), zap.String("query", query), zap.Int("page", page))
		return nil, err
	}
	return &result, nil
}

// GetMovieDetails 获取电影详情
func (s *SimpleTMDBService) GetMovieDetails(ctx context.Context, id int, language string) (*MovieDetails, error) {
	endpoint := fmt.Sprintf("/movie/%d?language=%s&append_to_response=images,credits,keywords", id, language)
	var result MovieDetails
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get movie details", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetTVDetails 获取电视剧详情
func (s *SimpleTMDBService) GetTVDetails(ctx context.Context, id int, language string) (*TVDetails, error) {
	endpoint := fmt.Sprintf("/tv/%d?language=%s&append_to_response=images,credits,keywords", id, language)
	var result TVDetails
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get TV details", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetSeasonDetails 获取季详情
func (s *SimpleTMDBService) GetSeasonDetails(ctx context.Context, tvID, seasonNumber int, language string) (*SeasonDetails, error) {
	endpoint := fmt.Sprintf("/tv/%d/season/%d?language=%s&append_to_response=images,credits", tvID, seasonNumber, language)
	var result SeasonDetails
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get season details", zap.String("error", err.Error()), zap.Int("tv_id", tvID), zap.Int("season", seasonNumber), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetEpisodeDetails 获取集详情
func (s *SimpleTMDBService) GetEpisodeDetails(ctx context.Context, tvID, seasonNumber, episodeNumber int, language string) (*EpisodeDetails, error) {
	endpoint := fmt.Sprintf("/tv/%d/season/%d/episode/%d?language=%s&append_to_response=images,credits", tvID, seasonNumber, episodeNumber, language)
	var result EpisodeDetails
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get episode details", zap.String("error", err.Error()), zap.Int("tv_id", tvID), zap.Int("season", seasonNumber), zap.Int("episode", episodeNumber), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetPersonDetails 获取人物详情
func (s *SimpleTMDBService) GetPersonDetails(ctx context.Context, id int, language string) (*PersonDetails, error) {
	endpoint := fmt.Sprintf("/person/%d?language=%s", id, language)
	var result PersonDetails
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get person details", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetMovieCredits 获取电影演职员
func (s *SimpleTMDBService) GetMovieCredits(ctx context.Context, id int, language string) (*Credits, error) {
	endpoint := fmt.Sprintf("/movie/%d/credits?language=%s", id, language)
	var result Credits
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get movie credits", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetTVCredits 获取电视剧演职员
func (s *SimpleTMDBService) GetTVCredits(ctx context.Context, id int, language string) (*Credits, error) {
	endpoint := fmt.Sprintf("/tv/%d/credits?language=%s", id, language)
	var result Credits
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get TV credits", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetPersonMovieCredits 获取人物电影作品
func (s *SimpleTMDBService) GetPersonMovieCredits(ctx context.Context, id int, language string) (*PersonMovieCredits, error) {
	endpoint := fmt.Sprintf("/person/%d/movie_credits?language=%s", id, language)
	var result PersonMovieCredits
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get person movie credits", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetPersonTVCredits 获取人物电视剧作品
func (s *SimpleTMDBService) GetPersonTVCredits(ctx context.Context, id int, language string) (*PersonTVCredits, error) {
	endpoint := fmt.Sprintf("/person/%d/tv_credits?language=%s", id, language)
	var result PersonTVCredits
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get person TV credits", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetPersonCombinedCredits 获取人物综合作品
func (s *SimpleTMDBService) GetPersonCombinedCredits(ctx context.Context, id int, language string) (*PersonCombinedCredits, error) {
	endpoint := fmt.Sprintf("/person/%d/combined_credits?language=%s", id, language)
	var result PersonCombinedCredits
	err := s.makeCachedRequest(ctx, endpoint, &result, 2*time.Hour)
	if err != nil {
		logger.Error("Failed to get person combined credits", zap.String("error", err.Error()), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetTrending 获取趋势内容
func (s *SimpleTMDBService) GetTrending(ctx context.Context, mediaType, timeWindow string, page int) (*TrendingResponse, error) {
	endpoint := fmt.Sprintf("/trending/%s/%s?page=%d", mediaType, timeWindow, page)
	var result TrendingResponse
	err := s.makeCachedRequest(ctx, endpoint, &result, 30*time.Minute)
	if err != nil {
		logger.Error("Failed to get trending", zap.String("error", err.Error()), zap.String("media_type", mediaType), zap.String("time_window", timeWindow), zap.Int("page", page))
		return nil, err
	}
	return &result, nil
}

// GetConfiguration 获取TMDB配置
func (s *SimpleTMDBService) GetConfiguration(ctx context.Context) (*Configuration, error) {
	endpoint := "/configuration"
	var result Configuration
	err := s.makeCachedRequest(ctx, endpoint, &result, 24*time.Hour)
	if err != nil {
		logger.Error("Failed to get configuration", zap.String("error", err.Error()))
		return nil, err
	}
	return &result, nil
}
