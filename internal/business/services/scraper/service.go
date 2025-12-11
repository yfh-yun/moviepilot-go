package scraper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/cache"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/scraper"
)

// Service 刮削服务接口
type Service interface {
	// ScrapeMovie 刮削电影
	ScrapeMovie(ctx context.Context, title string, year int, opts ScrapeOptions) (*scraper.MovieMetadata, error)

	// ScrapeTV 刮削电视剧
	ScrapeTV(ctx context.Context, title string, year int, opts ScrapeOptions) (*scraper.TVMetadata, error)

	// ScrapeSeason 刮削季
	ScrapeSeason(ctx context.Context, tvID int, seasonNumber int, opts ScrapeOptions) (*scraper.SeasonMetadata, error)

	// ScrapeEpisode 刮削集
	ScrapeEpisode(ctx context.Context, tvID int, seasonNumber int, episodeNumber int, opts ScrapeOptions) (*scraper.EpisodeMetadata, error)

	// BatchScrapeMovies 批量刮削电影
	BatchScrapeMovies(ctx context.Context, requests []ScrapeRequest) ([]*ScrapeResult, error)

	// BatchScrapeTVs 批量刮削电视剧
	BatchScrapeTVs(ctx context.Context, requests []ScrapeRequest) ([]*ScrapeResult, error)
}

// service 刮削服务实现
type service struct {
	scraper   scraper.Scraper
	cache     cache.Cache
	imagePath string
	logger    *zap.Logger
}

// NewService 创建刮削服务
func NewService(apiKey string, cache cache.Cache, imagePath string) Service {
	return &service{
		scraper:   scraper.NewTMDbScraper(apiKey),
		cache:     cache,
		imagePath: imagePath,
		logger:    logger.GetLogger(),
	}
}

// ScrapeOptions 刮削选项
type ScrapeOptions struct {
	// UseCache 是否使用缓存
	UseCache bool

	// DownloadImages 是否下载图片
	DownloadImages bool

	// SaveMetadata 是否保存元数据
	SaveMetadata bool
}

// ScrapeRequest 刮削请求
type ScrapeRequest struct {
	Title string
	Year  int
}

// ScrapeResult 刮削结果
type ScrapeResult struct {
	Title    string
	Year     int
	Metadata any
	Error    error
}

// ScrapeMovie 刮削电影
func (s *service) ScrapeMovie(ctx context.Context, title string, year int, opts ScrapeOptions) (*scraper.MovieMetadata, error) {
	s.logger.Info("刮削电影", zap.String("title", title), zap.Int("year", year))

	// 检查缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:movie:%s:%d", title, year)
		var cached scraper.MovieMetadata
		if err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			s.logger.Debug("从缓存获取电影元数据")
			return &cached, nil
		}
	}

	// 刮削
	metadata, err := s.scraper.ScrapeMovie(ctx, title, year)
	if err != nil {
		return nil, err
	}

	// 下载图片
	if opts.DownloadImages {
		if err := s.downloadMovieImages(ctx, metadata); err != nil {
			s.logger.Warn("下载电影图片失败", zap.Error(err))
		}
	}

	// 保存到缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:movie:%s:%d", title, year)
		if err := s.cache.SetJSON(ctx, cacheKey, metadata, 86400); err != nil {
			s.logger.Warn("保存缓存失败", zap.Error(err))
		}
	}

	s.logger.Info("电影刮削完成", zap.String("title", metadata.Title))
	return metadata, nil
}

// ScrapeTV 刮削电视剧
func (s *service) ScrapeTV(ctx context.Context, title string, year int, opts ScrapeOptions) (*scraper.TVMetadata, error) {
	s.logger.Info("刮削电视剧", zap.String("title", title), zap.Int("year", year))

	// 检查缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:tv:%s:%d", title, year)
		var cached scraper.TVMetadata
		if err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			s.logger.Debug("从缓存获取电视剧元数据")
			return &cached, nil
		}
	}

	// 刮削
	metadata, err := s.scraper.ScrapeTV(ctx, title, year)
	if err != nil {
		return nil, err
	}

	// 下载图片
	if opts.DownloadImages {
		if err := s.downloadTVImages(ctx, metadata); err != nil {
			s.logger.Warn("下载电视剧图片失败", zap.Error(err))
		}
	}

	// 保存到缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:tv:%s:%d", title, year)
		if err := s.cache.SetJSON(ctx, cacheKey, metadata, 86400); err != nil {
			s.logger.Warn("保存缓存失败", zap.Error(err))
		}
	}

	s.logger.Info("电视剧刮削完成", zap.String("name", metadata.Name))
	return metadata, nil
}

// ScrapeSeason 刮削季
func (s *service) ScrapeSeason(ctx context.Context, tvID int, seasonNumber int, opts ScrapeOptions) (*scraper.SeasonMetadata, error) {
	s.logger.Info("刮削季", zap.Int("tvID", tvID), zap.Int("seasonNumber", seasonNumber))

	// 检查缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:season:%d:%d", tvID, seasonNumber)
		var cached scraper.SeasonMetadata
		if err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			s.logger.Debug("从缓存获取季元数据")
			return &cached, nil
		}
	}

	// 刮削
	metadata, err := s.scraper.ScrapeSeason(ctx, tvID, seasonNumber)
	if err != nil {
		return nil, err
	}

	// 下载图片
	if opts.DownloadImages && metadata.PosterPath != "" {
		if err := s.downloadSeasonPoster(ctx, metadata); err != nil {
			s.logger.Warn("下载季海报失败", zap.Error(err))
		}
	}

	// 保存到缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:season:%d:%d", tvID, seasonNumber)
		if err := s.cache.SetJSON(ctx, cacheKey, metadata, 86400); err != nil {
			s.logger.Warn("保存缓存失败", zap.Error(err))
		}
	}

	s.logger.Info("季刮削完成", zap.String("name", metadata.Name))
	return metadata, nil
}

// ScrapeEpisode 刮削集
func (s *service) ScrapeEpisode(ctx context.Context, tvID int, seasonNumber int, episodeNumber int, opts ScrapeOptions) (*scraper.EpisodeMetadata, error) {
	s.logger.Info("刮削集",
		zap.Int("tvID", tvID),
		zap.Int("seasonNumber", seasonNumber),
		zap.Int("episodeNumber", episodeNumber))

	// 检查缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:episode:%d:%d:%d", tvID, seasonNumber, episodeNumber)
		var cached scraper.EpisodeMetadata
		if err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			s.logger.Debug("从缓存获取集元数据")
			return &cached, nil
		}
	}

	// 刮削
	metadata, err := s.scraper.ScrapeEpisode(ctx, tvID, seasonNumber, episodeNumber)
	if err != nil {
		return nil, err
	}

	// 下载图片
	if opts.DownloadImages && metadata.StillPath != "" {
		if err := s.downloadEpisodeStill(ctx, metadata); err != nil {
			s.logger.Warn("下载集剧照失败", zap.Error(err))
		}
	}

	// 保存到缓存
	if opts.UseCache {
		cacheKey := fmt.Sprintf("scraper:episode:%d:%d:%d", tvID, seasonNumber, episodeNumber)
		if err := s.cache.SetJSON(ctx, cacheKey, metadata, 86400); err != nil {
			s.logger.Warn("保存缓存失败", zap.Error(err))
		}
	}

	s.logger.Info("集刮削完成", zap.String("name", metadata.Name))
	return metadata, nil
}

// BatchScrapeMovies 批量刮削电影
func (s *service) BatchScrapeMovies(ctx context.Context, requests []ScrapeRequest) ([]*ScrapeResult, error) {
	s.logger.Info("批量刮削电影", zap.Int("count", len(requests)))

	results := make([]*ScrapeResult, len(requests))
	opts := ScrapeOptions{
		UseCache:       true,
		DownloadImages: true,
	}

	for i, req := range requests {
		result := &ScrapeResult{
			Title: req.Title,
			Year:  req.Year,
		}

		metadata, err := s.ScrapeMovie(ctx, req.Title, req.Year, opts)
		if err != nil {
			result.Error = err
		} else {
			result.Metadata = metadata
		}

		results[i] = result

		// 避免请求过快
		time.Sleep(250 * time.Millisecond)
	}

	s.logger.Info("批量刮削电影完成", zap.Int("count", len(requests)))
	return results, nil
}

// BatchScrapeTVs 批量刮削电视剧
func (s *service) BatchScrapeTVs(ctx context.Context, requests []ScrapeRequest) ([]*ScrapeResult, error) {
	s.logger.Info("批量刮削电视剧", zap.Int("count", len(requests)))

	results := make([]*ScrapeResult, len(requests))
	opts := ScrapeOptions{
		UseCache:       true,
		DownloadImages: true,
	}

	for i, req := range requests {
		result := &ScrapeResult{
			Title: req.Title,
			Year:  req.Year,
		}

		metadata, err := s.ScrapeTV(ctx, req.Title, req.Year, opts)
		if err != nil {
			result.Error = err
		} else {
			result.Metadata = metadata
		}

		results[i] = result

		// 避免请求过快
		time.Sleep(250 * time.Millisecond)
	}

	s.logger.Info("批量刮削电视剧完成", zap.Int("count", len(requests)))
	return results, nil
}

// downloadMovieImages 下载电影图片
func (s *service) downloadMovieImages(ctx context.Context, metadata *scraper.MovieMetadata) error {
	// 下载海报
	if metadata.PosterPath != "" {
		posterPath := filepath.Join(s.imagePath, "movies", fmt.Sprintf("%d", metadata.TMDBID), "poster.jpg")
		if err := s.downloadImage(ctx, metadata.PosterPath, posterPath); err != nil {
			return fmt.Errorf("下载海报失败: %w", err)
		}
	}

	// 下载背景图
	if metadata.BackdropPath != "" {
		backdropPath := filepath.Join(s.imagePath, "movies", fmt.Sprintf("%d", metadata.TMDBID), "backdrop.jpg")
		if err := s.downloadImage(ctx, metadata.BackdropPath, backdropPath); err != nil {
			return fmt.Errorf("下载背景图失败: %w", err)
		}
	}

	return nil
}

// downloadTVImages 下载电视剧图片
func (s *service) downloadTVImages(ctx context.Context, metadata *scraper.TVMetadata) error {
	// 下载海报
	if metadata.PosterPath != "" {
		posterPath := filepath.Join(s.imagePath, "tv", fmt.Sprintf("%d", metadata.TMDBID), "poster.jpg")
		if err := s.downloadImage(ctx, metadata.PosterPath, posterPath); err != nil {
			return fmt.Errorf("下载海报失败: %w", err)
		}
	}

	// 下载背景图
	if metadata.BackdropPath != "" {
		backdropPath := filepath.Join(s.imagePath, "tv", fmt.Sprintf("%d", metadata.TMDBID), "backdrop.jpg")
		if err := s.downloadImage(ctx, metadata.BackdropPath, backdropPath); err != nil {
			return fmt.Errorf("下载背景图失败: %w", err)
		}
	}

	return nil
}

// downloadSeasonPoster 下载季海报
func (s *service) downloadSeasonPoster(ctx context.Context, metadata *scraper.SeasonMetadata) error {
	posterPath := filepath.Join(s.imagePath, "seasons", fmt.Sprintf("%d", metadata.ID), "poster.jpg")
	return s.downloadImage(ctx, metadata.PosterPath, posterPath)
}

// downloadEpisodeStill 下载集剧照
func (s *service) downloadEpisodeStill(ctx context.Context, metadata *scraper.EpisodeMetadata) error {
	stillPath := filepath.Join(s.imagePath, "episodes", fmt.Sprintf("%d", metadata.ID), "still.jpg")
	return s.downloadImage(ctx, metadata.StillPath, stillPath)
}

// downloadImage 下载图片
func (s *service) downloadImage(ctx context.Context, imagePath string, savePath string) error {
	// 构建完整 URL
	imageURL := fmt.Sprintf("https://image.tmdb.org/t/p/original%s", imagePath)

	// 下载图片
	data, err := s.scraper.DownloadImage(ctx, imageURL)
	if err != nil {
		return err
	}

	// 创建目录
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 保存文件
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}
