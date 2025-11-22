package media

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/media/tmdb"
	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/cache"
)

// TMDBService 调用 TMDB API 获取更准确的媒体信息。
type TMDBService struct {
	client   *tmdb.Client
	cache    cache.Cache
	logger   *zap.Logger
	fallback Service
	
	// 缓存键映射
	cacheKeys sync.Map
}

// NewTMDBService 创建 TMDB 识别服务；若 apiKey 为空则返回 nil。
func NewTMDBService(apiKey string, logger *zap.Logger, fallback Service, cache cache.Cache) *TMDBService {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil
	}

	// 创建TMDB客户端
	tmdbClient := tmdb.NewClient(tmdb.Config{
		APIKey: apiKey,
		Cache:  cache,
		Logger: logger,
	})

	return &TMDBService{
		client:   tmdbClient,
		cache:    cache,
		logger:   logger,
		fallback: fallback,
	}
}

// Identify 根据 TMDB 搜索结果生成 Media 信息，失败时回退到 fallback。
func (s *TMDBService) Identify(files []FileItem, opts IdentifyOptions) ([]models.Media, error) {
	results := make([]models.Media, 0, len(files))
	for _, file := range files {
		media, err := s.identifySingle(file)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("tmdb identify failed", zap.String("file", file.Path), zap.Error(err))
			}
			fallbackMedia := s.fallbackMedia(file, opts)
			results = append(results, fallbackMedia)
			continue
		}
		results = append(results, media)
	}
	return results, nil
}

func (s *TMDBService) identifySingle(file FileItem) (models.Media, error) {
	query := sanitize(file.Path)
	if query == "" {
		query = filepath.Base(file.Path)
	}
	typeHint := guessType(file.Path)
	cacheKey := fmt.Sprintf("identify:%s|%s", strings.ToLower(query), typeHint)

	// 检查缓存
	if s.cache != nil {
		var media models.Media
		err := s.cache.GetJSON(context.Background(), cacheKey, &media)
		if err == nil {
			if s.logger != nil {
				s.logger.Debug("TMDB service cache hit", zap.String("file", file.Path), zap.String("cache_key", cacheKey))
			}
			return media, nil
		}
	}

	media, err := s.lookupTMDB(query, typeHint)
	if err != nil {
		return models.Media{}, err
	}

	// 存入缓存
	if s.cache != nil {
		if err := s.cache.SetJSON(context.Background(), cacheKey, media, time.Hour); err != nil && s.logger != nil {
			s.logger.Warn("failed to cache TMDB result", zap.Error(err), zap.String("cache_key", cacheKey))
		}
	}

	return media, nil
}

func (s *TMDBService) lookupTMDB(query, typeHint string) (models.Media, error) {
	ctx := context.Background()
	
	// 根据类型选择搜索方式
	var err error
	var media models.Media
	
	switch strings.ToLower(typeHint) {
	case "movie":
		media, err = s.searchMovie(ctx, query)
	case "tv", "series":
		media, err = s.searchTV(ctx, query)
	default:
		// 使用多搜索
		media, err = s.searchMulti(ctx, query, typeHint)
	}
	
	if err != nil {
		return models.Media{}, fmt.Errorf("tmdb lookup failed: %w", err)
	}

	return media, nil
}

func (s *TMDBService) searchMovie(ctx context.Context, query string) (models.Media, error) {
	// 尝试从文件名提取年份
	year := extractYearFromQuery(query)
	
	resp, err := s.client.SearchMovie(ctx, query, year, 1)
	if err != nil {
		return models.Media{}, err
	}
	
	if len(resp.Results) == 0 {
		return models.Media{}, fmt.Errorf("no movie found for query: %s", query)
	}
	
	// 获取最佳匹配
	result := resp.Results[0]
	
	// 获取详细信息
	details, err := s.client.GetMovieDetails(ctx, result.ID)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("failed to get movie details", zap.Int("id", result.ID), zap.Error(err))
		}
		// 使用搜索结果
		return s.movieResultToMedia(&result), nil
	}
	
	return s.movieDetailsToMedia(details), nil
}

func (s *TMDBService) searchTV(ctx context.Context, query string) (models.Media, error) {
	year := extractYearFromQuery(query)
	
	resp, err := s.client.SearchTV(ctx, query, year, 1)
	if err != nil {
		return models.Media{}, err
	}
	
	if len(resp.Results) == 0 {
		return models.Media{}, fmt.Errorf("no TV show found for query: %s", query)
	}
	
	result := resp.Results[0]
	
	// 获取详细信息
	details, err := s.client.GetTVDetails(ctx, result.ID)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("failed to get TV details", zap.Int("id", result.ID), zap.Error(err))
		}
		return s.tvResultToMedia(&result), nil
	}
	
	return s.tvDetailsToMedia(details), nil
}

func (s *TMDBService) searchMulti(ctx context.Context, query, typeHint string) (models.Media, error) {
	resp, err := s.client.SearchMulti(ctx, query, 1)
	if err != nil {
		return models.Media{}, err
	}
	
	if len(resp.Results) == 0 {
		return models.Media{}, fmt.Errorf("no media found for query: %s", query)
	}
	
	// 优先选择匹配类型的第一个结果
	preferredType := strings.ToLower(typeHint)
	if preferredType == "" {
		preferredType = "movie"
	}
	
	for _, result := range resp.Results {
		if strings.ToLower(result.MediaType) == preferredType {
			return s.multiResultToMedia(&result), nil
		}
	}
	
	// 如果没有匹配类型，返回第一个结果
	return s.multiResultToMedia(&resp.Results[0]), nil
}

func (s *TMDBService) movieResultToMedia(result *tmdb.MovieResult) models.Media {
	year := extractYear(result.ReleaseDate)
	var yearPtr *string
	if year != "" {
		yearPtr = &year
	}
	
	var tmdbID = result.ID
	
	return models.Media{
		BaseModel:     models.BaseModel{},
		TMDBID:        &tmdbID,
		Title:         result.Title,
		OriginalTitle: result.OriginalTitle,
		Year:          yearPtr,
		Type:          "movie",
		Description:   result.Overview,
		Poster:        s.client.BuildImageURL(result.PosterPath),
		Backdrop:      s.client.BuildImageURL(result.BackdropPath),
		Vote:          &result.VoteAverage,
	}
}

func (s *TMDBService) movieDetailsToMedia(details *tmdb.MovieDetails) models.Media {
	year := extractYear(details.ReleaseDate)
	var yearPtr *string
	if year != "" {
		yearPtr = &year
	}
	
	// 处理类型
	var genres []string
	for _, genre := range details.Genres {
		genres = append(genres, genre.Name)
	}
	
	// 处理国家
	var countries []string
	for _, country := range details.ProductionCountries {
		countries = append(countries, country.Name)
	}
	
	// 处理语言
	language := details.OriginalLanguage
	if language == "" && len(details.SpokenLanguages) > 0 {
		language = details.SpokenLanguages[0].ISO6391
	}
	
	return models.Media{
		BaseModel:     models.BaseModel{},
		TMDBID:        &details.ID,
		IMDBID:        &details.IMDBID,
		Title:         details.Title,
		OriginalTitle: details.OriginalTitle,
		Year:          yearPtr,
		Type:          "movie",
		Description:   details.Overview,
		Poster:        s.client.BuildImageURL(details.PosterPath),
		Backdrop:      s.client.BuildImageURL(details.BackdropPath),
		Vote:          &details.VoteAverage,
		Genres:        genresJSON(genres),
		Countries:     countriesJSON(countries),
		Language:      language,
		Runtime:       &details.Runtime,
	}
}

func (s *TMDBService) tvResultToMedia(result *tmdb.TVResult) models.Media {
	year := extractYear(result.FirstAirDate)
	var yearPtr *string
	if year != "" {
		yearPtr = &year
	}
	
	var tmdbID = result.ID
	
	return models.Media{
		BaseModel:     models.BaseModel{},
		TMDBID:        &tmdbID,
		Title:         result.Name,
		OriginalTitle: result.OriginalName,
		Year:          yearPtr,
		Type:          "tv",
		Description:   result.Overview,
		Poster:        s.client.BuildImageURL(result.PosterPath),
		Backdrop:      s.client.BuildImageURL(result.BackdropPath),
		Vote:          &result.VoteAverage,
	}
}

func (s *TMDBService) tvDetailsToMedia(details *tmdb.TVDetails) models.Media {
	year := extractYear(details.FirstAirDate)
	var yearPtr *string
	if year != "" {
		yearPtr = &year
	}
	
	// 处理类型
	var genres []string
	for _, genre := range details.Genres {
		genres = append(genres, genre.Name)
	}
	
	// 处理国家
	var countries []string
	for _, country := range details.OriginCountry {
		countries = append(countries, country)
	}
	
	// 处理语言
	language := details.OriginalLanguage
	if language == "" && len(details.Languages) > 0 {
		language = details.Languages[0]
	}
	
	var runtime *int
	if len(details.EpisodeRunTime) > 0 {
		runtime = &details.EpisodeRunTime[0]
	}
	
	return models.Media{
		BaseModel:     models.BaseModel{},
		TMDBID:        &details.ID,
		Title:         details.Name,
		OriginalTitle: details.OriginalName,
		Year:          yearPtr,
		Type:          "tv",
		Description:   details.Overview,
		Poster:        s.client.BuildImageURL(details.PosterPath),
		Backdrop:      s.client.BuildImageURL(details.BackdropPath),
		Vote:          &details.VoteAverage,
		Genres:        genresJSON(genres),
		Countries:     countriesJSON(countries),
		Language:      language,
		Runtime:       runtime,
	}
}

func (s *TMDBService) multiResultToMedia(result *tmdb.MultiResult) models.Media {
	year := extractYear(result.ReleaseDate)
	if year == "" {
		year = extractYear(result.FirstAirDate)
	}
	var yearPtr *string
	if year != "" {
		yearPtr = &year
	}
	
	title := result.Title
	if title == "" {
		title = result.Name
	}
	original := result.OriginalTitle
	if original == "" {
		original = result.OriginalName
	}
	
	var tmdbID = result.ID
	
	return models.Media{
		BaseModel:     models.BaseModel{},
		TMDBID:        &tmdbID,
		Title:         title,
		OriginalTitle: original,
		Year:          yearPtr,
		Type:          result.MediaType,
		Description:   result.Overview,
		Poster:        s.client.BuildImageURL(result.PosterPath),
		Backdrop:      s.client.BuildImageURL(result.BackdropPath),
		Vote:          &result.VoteAverage,
	}
}

// 辅助函数
func extractYearFromQuery(query string) int {
	// 简单的年份提取逻辑
	// 可以在这里添加更复杂的解析
	return 0
}

func genresJSON(genres []string) string {
	data, _ := json.Marshal(genres)
	return string(data)
}

func countriesJSON(countries []string) string {
	data, _ := json.Marshal(countries)
	return string(data)
}

func (s *TMDBService) fallbackMedia(file FileItem, opts IdentifyOptions) models.Media {
	if s.fallback != nil {
		medias, err := s.fallback.Identify([]FileItem{file}, opts)
		if err == nil && len(medias) > 0 {
			return medias[0]
		}
	}
	title := sanitize(file.Path)
	if title == "" {
		title = filepath.Base(file.Path)
	}
	return models.Media{
		Title: title,
		Type:  guessType(file.Path),
	}
}

// DownloadPoster 下载电影/电视剧海报
func (s *TMDBService) DownloadPoster(ctx context.Context, media models.Media, outputDir string, opts tmdb.DownloadOptions) (string, error) {
	if media.Poster == "" || media.TMDBID == nil {
		return "", fmt.Errorf("no poster path or TMDB ID available")
	}
	
	// 从海报URL中提取路径
	posterPath := ""
	if strings.HasPrefix(media.Poster, "http") {
		// 已经是完整URL，需要提取相对路径
		parts := strings.Split(media.Poster, "/")
		if len(parts) > 1 {
			posterPath = strings.Join(parts[1:], "/")
		}
	} else {
		posterPath = media.Poster
	}
	
	// 创建下载器
	downloader := tmdb.NewImageDownloader(s.client, s.cache, s.logger, opts)
	
	// 下载海报
	return downloader.DownloadPoster(ctx, media.Type, posterPath)
}

// DownloadBackdrop 下载电影/电视剧背景图
func (s *TMDBService) DownloadBackdrop(ctx context.Context, media models.Media, outputDir string, opts tmdb.DownloadOptions) (string, error) {
	if media.Backdrop == "" || media.TMDBID == nil {
		return "", fmt.Errorf("no backdrop path or TMDB ID available")
	}
	
	// 从背景URL中提取路径
	backdropPath := ""
	if strings.HasPrefix(media.Backdrop, "http") {
		// 已经是完整URL，需要提取相对路径
		parts := strings.Split(media.Backdrop, "/")
		if len(parts) > 1 {
			backdropPath = strings.Join(parts[1:], "/")
		}
	} else {
		backdropPath = media.Backdrop
	}
	
	// 创建下载器
	downloader := tmdb.NewImageDownloader(s.client, s.cache, s.logger, opts)
	
	// 下载背景图
	return downloader.DownloadBackdrop(ctx, media.Type, backdropPath)
}

// DownloadAllImages 下载所有相关图片
func (s *TMDBService) DownloadAllImages(ctx context.Context, media models.Media, outputDir string, opts tmdb.DownloadOptions) ([]string, error) {
	var results []string
	
	// 下载海报
	if posterPath, err := s.DownloadPoster(ctx, media, outputDir, opts); err == nil {
		results = append(results, posterPath)
	} else if s.logger != nil {
		s.logger.Warn("failed to download poster", zap.Error(err))
	}
	
	// 下载背景图
	if backdropPath, err := s.DownloadBackdrop(ctx, media, outputDir, opts); err == nil {
		results = append(results, backdropPath)
	} else if s.logger != nil {
		s.logger.Warn("failed to download backdrop", zap.Error(err))
	}
	
	return results, nil
}

// extractYear 从日期字符串中提取年份
func extractYear(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return ""
}
