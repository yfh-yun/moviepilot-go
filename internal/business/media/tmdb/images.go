package tmdb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// GetMovieImages 获取电影图片
func (s *TMDBService) GetMovieImages(ctx context.Context, id int, language string) (*MovieImages, error) {
	endpoint := fmt.Sprintf("/movie/%d/images?language=%s", id, language)
	var result MovieImages
	err := s.makeCachedRequest(ctx, endpoint, &result, 6*time.Hour)
	if err != nil {
		logger.Error("Failed to get movie images", zap.Error(err), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetTVImages 获取电视剧图片
func (s *TMDBService) GetTVImages(ctx context.Context, id int, language string) (*TVImages, error) {
	endpoint := fmt.Sprintf("/tv/%d/images?language=%s", id, language)
	var result TVImages
	err := s.makeCachedRequest(ctx, endpoint, &result, 6*time.Hour)
	if err != nil {
		logger.Error("Failed to get TV images", zap.Error(err), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// GetPersonImages 获取人物图片
func (s *TMDBService) GetPersonImages(ctx context.Context, id int, language string) (*PersonImages, error) {
	endpoint := fmt.Sprintf("/person/%d/images?language=%s", id, language)
	var result PersonImages
	err := s.makeCachedRequest(ctx, endpoint, &result, 6*time.Hour)
	if err != nil {
		logger.Error("Failed to get person images", zap.Error(err), zap.Int("id", id), zap.String("language", language))
		return nil, err
	}
	return &result, nil
}

// DownloadImage 下载单个图片
func (s *TMDBService) DownloadImage(ctx context.Context, filePath string, config *ImageConfig) ([]byte, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path is required")
	}

	// 使用默认配置
	if config == nil {
		config = s.imageConfig
	}

	// 构建完整URL
	imageURL := s.buildImageURL(filePath, config)

	// TODO: 修复缓存类型问题后重新启用
	// cacheKey := fmt.Sprintf("tmdb:image:%s", filePath)
	// if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
	// 	logger.Debug("Image cache hit", zap.String("key", cacheKey))
	// 	return cached.([]byte), nil
	// }

	logger.Debug("Downloading image", zap.String("url", imageURL), zap.String("file_path", filePath))

	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		logger.Error("Failed to create image request", zap.Error(err), zap.String("url", imageURL))
		return nil, err
	}

	// 设置User-Agent
	if config.UserAgent != "" {
		req.Header.Set("User-Agent", config.UserAgent)
	} else {
		req.Header.Set("User-Agent", "MoviePilot-TMDB/1.0")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error("Failed to download image", zap.Error(err), zap.String("url", imageURL))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Image download error", zap.Int("status", resp.StatusCode), zap.String("url", imageURL))
		return nil, fmt.Errorf("image download error: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read image data", zap.Error(err), zap.String("url", imageURL))
		return nil, err
	}

	// 缓存图片数据
	// cacheKey := fmt.Sprintf("tmdb:image:%s", filePath)
	// if err := s.cache.Set(ctx, cacheKey, data, 24*time.Hour); err != nil {
	// 	logger.Warn("Failed to cache image", zap.Error(err), zap.String("key", cacheKey))
	// }

	logger.Debug("Image downloaded successfully", zap.String("url", imageURL), zap.Int("size", len(data)))
	return data, nil
}

// DownloadImages 并发下载多个图片
func (s *TMDBService) DownloadImages(ctx context.Context, filePaths []string, config *ImageConfig) (map[string][]byte, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("file paths are required")
	}

	// 使用默认配置
	if config == nil {
		config = s.imageConfig
	}

	// 设置默认并发数
	maxConcurrency := config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	logger.Info("Starting batch image download", zap.Int("count", len(filePaths)), zap.Int("max_concurrency", maxConcurrency))

	results := make(map[string][]byte)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建工作池
	semaphore := make(chan struct{}, maxConcurrency)
	errChan := make(chan error, len(filePaths))

	for _, filePath := range filePaths {
		if filePath == "" {
			continue
		}

		wg.Add(1)
		go func(fp string) {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()

				data, err := s.DownloadImage(ctx, fp, config)
				if err != nil {
					errChan <- err
					return
				}

				mu.Lock()
				results[fp] = data
				mu.Unlock()

			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}(filePath)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
			logger.Error("Image download failed", zap.Error(err))
		}
	}

	if len(errors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all image downloads failed, first error: %v", errors[0])
	}

	logger.Info("Batch image download completed", zap.Int("successful", len(results)), zap.Int("failed", len(errors)), zap.Int("total", len(filePaths)))

	if len(errors) > 0 {
		logger.Warn("Some images failed to download", zap.Int("failed_count", len(errors)))
	}

	return results, nil
}

// buildImageURL 构建图片URL
func (s *TMDBService) buildImageURL(filePath string, config *ImageConfig) string {
	// 清理文件路径
	filePath = strings.TrimPrefix(filePath, "/")

	// 构建URL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://image.tmdb.org/t/p"
	}

	size := config.Size
	if size == "" {
		size = "w500"
	}

	return fmt.Sprintf("%s/%s/%s", baseURL, size, filePath)
}

// GetBestImage 获取最佳质量的图片（委托给工具函数）
func (s *TMDBService) GetBestImage(images []Image, preferLanguage string) *Image {
	return GetBestImage(images, preferLanguage)
}

// calculateImageScore 已迁移到 image_utils.go，此方法保留用于兼容
func (s *TMDBService) calculateImageScore(img Image, preferLanguage string) float64 {
	return calculateImageScore(img, preferLanguage)
}

// GetImageExtension 获取图片文件扩展名（委托给工具函数）
func (s *TMDBService) GetImageExtension(filePath string) string {
	return GetImageExtension(filePath)
}

// FilterImagesBySize 按尺寸过滤图片（委托给工具函数）
func (s *TMDBService) FilterImagesBySize(images []Image, minWidth, minHeight int) []Image {
	return FilterImagesBySize(images, minWidth, minHeight)
}

// FilterImagesByAspectRatio 按宽高比过滤图片（委托给工具函数）
func (s *TMDBService) FilterImagesByAspectRatio(images []Image, minRatio, maxRatio float64) []Image {
	return FilterImagesByAspectRatio(images, minRatio, maxRatio)
}

// FilterImagesByLanguage 按语言过滤图片（委托给工具函数）
func (s *TMDBService) FilterImagesByLanguage(images []Image, languages []string) []Image {
	return FilterImagesByLanguage(images, languages)
}
