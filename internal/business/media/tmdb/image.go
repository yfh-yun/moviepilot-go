package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/cache"
)

// DownloadOptions 图片下载选项
type DownloadOptions struct {
	// 目标目录
	OutputDir string
	// 图片类型：poster, backdrop, logo等
	ImageType string
	// 图片尺寸
	Size string
	// 是否覆盖已存在的文件
	Overwrite bool
	// 最大并发下载数
	MaxConcurrent int
}

// ImageDownloader TMDB图片下载器
type ImageDownloader struct {
	client *Client
	cache  cache.Cache
	logger *zap.Logger
	opts   DownloadOptions
}

// NewImageDownloader 创建图片下载器
func NewImageDownloader(client *Client, cache cache.Cache, logger *zap.Logger, opts DownloadOptions) *ImageDownloader {
	// 设置默认值
	if opts.Size == "" {
		opts.Size = "w500"
	}
	if opts.MaxConcurrent == 0 {
		opts.MaxConcurrent = 5
	}

	return &ImageDownloader{
		client: client,
		cache:  cache,
		logger: logger,
		opts:   opts,
	}
}

// DownloadPoster 下载海报
func (d *ImageDownloader) DownloadPoster(ctx context.Context, mediaType string, posterPath string) (string, error) {
	if posterPath == "" {
		return "", fmt.Errorf("poster path is empty")
	}

	// 构建完整URL
	imageURL := d.client.BuildImageURL(posterPath, d.opts.Size)

	// 生成本地文件名
	ext := filepath.Ext(posterPath)
	if ext == "" {
		ext = ".jpg"
	}

	// 根据媒体类型确定文件名
	var filename string
	switch strings.ToLower(mediaType) {
	case "movie":
		filename = fmt.Sprintf("poster%s", ext)
	case "tv", "series":
		filename = fmt.Sprintf("poster%s", ext)
	default:
		filename = fmt.Sprintf("image%s", ext)
	}

	// 下载图片
	return d.downloadImage(ctx, imageURL, filename)
}

// DownloadBackdrop 下载背景图
func (d *ImageDownloader) DownloadBackdrop(ctx context.Context, mediaType string, backdropPath string) (string, error) {
	if backdropPath == "" {
		return "", fmt.Errorf("backdrop path is empty")
	}

	// 构建完整URL
	imageURL := d.client.BuildImageURL(backdropPath, "w1280")

	// 生成本地文件名
	ext := filepath.Ext(backdropPath)
	if ext == "" {
		ext = ".jpg"
	}

	// 根据媒体类型确定文件名
	var filename string
	switch strings.ToLower(mediaType) {
	case "movie":
		filename = fmt.Sprintf("backdrop%s", ext)
	case "tv", "series":
		filename = fmt.Sprintf("backdrop%s", ext)
	default:
		filename = fmt.Sprintf("fanart%s", ext)
	}

	// 下载图片
	return d.downloadImage(ctx, imageURL, filename)
}

// downloadImage 下载单个图片
func (d *ImageDownloader) downloadImage(ctx context.Context, imageURL, filename string) (string, error) {
	// 检查输出目录
	if d.opts.OutputDir == "" {
		return "", fmt.Errorf("output directory not specified")
	}

	// 确保目录存在
	if err := os.MkdirAll(d.opts.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// 本地文件路径
	localPath := filepath.Join(d.opts.OutputDir, filename)

	// 检查文件是否已存在
	if !d.opts.Overwrite {
		if _, err := os.Stat(localPath); err == nil {
			if d.logger != nil {
				d.logger.Debug("image already exists, skipping",
					zap.String("url", imageURL),
					zap.String("path", localPath))
			}
			return localPath, nil
		}
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("image:%s", imageURL)
	if d.cache != nil {
		var cached bool
		if _, err := d.cache.Get(ctx, cacheKey); err == nil {
			cached = true
			if d.logger != nil {
				d.logger.Debug("image download cache hit",
					zap.String("url", imageURL),
					zap.String("path", localPath))
			}
		}

		if cached {
			// 检查文件是否确实存在
			if _, err := os.Stat(localPath); err == nil {
				return localPath, nil
			}
		}
	}

	if d.logger != nil {
		d.logger.Info("downloading image",
			zap.String("url", imageURL),
			zap.String("path", localPath))
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 执行请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image download failed with status: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 复制响应体到文件
	_, err = file.ReadFrom(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// 缓存成功状态
	if d.cache != nil {
		if err := d.cache.Set(ctx, cacheKey, "downloaded", 24*time.Hour); err != nil && d.logger != nil {
			d.logger.Warn("failed to cache image download status",
				zap.Error(err),
				zap.String("url", imageURL))
		}
	}

	if d.logger != nil {
		d.logger.Info("image download completed",
			zap.String("url", imageURL),
			zap.String("path", localPath))
	}

	return localPath, nil
}

// DownloadBatch 批量下载多个图片
func (d *ImageDownloader) DownloadBatch(ctx context.Context, mediaType string, imagePaths []string) ([]string, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no image paths provided")
	}

	// 设置默认并发数
	maxConcurrency := d.opts.MaxConcurrent
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	if d.logger != nil {
		d.logger.Info("starting batch image download",
			zap.Int("count", len(imagePaths)),
			zap.Int("max_concurrency", maxConcurrency))
	}

	type result struct {
		path string
		err  error
	}

	resultChan := make(chan result, len(imagePaths))
	semaphore := make(chan struct{}, maxConcurrency)

	// 启动下载任务
	for _, imgPath := range imagePaths {
		if imgPath == "" {
			continue
		}

		go func(path string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 根据图片类型选择下载方法
			var localPath string
			var err error

			// 简单判断：如果路径包含 backdrop，使用 backdrop 下载
			if strings.Contains(strings.ToLower(path), "backdrop") {
				localPath, err = d.DownloadBackdrop(ctx, mediaType, path)
			} else {
				localPath, err = d.DownloadPoster(ctx, mediaType, path)
			}

			resultChan <- result{path: localPath, err: err}
		}(imgPath)
	}

	// 收集结果
	var successPaths []string
	var errors []error

	for i := 0; i < len(imagePaths); i++ {
		select {
		case res := <-resultChan:
			if res.err != nil {
				errors = append(errors, res.err)
				if d.logger != nil {
					d.logger.Warn("image download failed", zap.Error(res.err))
				}
			} else if res.path != "" {
				successPaths = append(successPaths, res.path)
			}
		case <-ctx.Done():
			return successPaths, ctx.Err()
		}
	}

	if d.logger != nil {
		d.logger.Info("batch image download completed",
			zap.Int("successful", len(successPaths)),
			zap.Int("failed", len(errors)),
			zap.Int("total", len(imagePaths)))
	}

	if len(errors) > 0 && len(successPaths) == 0 {
		return nil, fmt.Errorf("all downloads failed, first error: %w", errors[0])
	}

	return successPaths, nil
}

// DownloadImages 批量下载图片
func (d *ImageDownloader) DownloadImages(ctx context.Context, urls []string, prefix string) ([]string, error) {
	// 限制并发数
	semaphore := make(chan struct{}, d.opts.MaxConcurrent)
	resultChan := make(chan downloadResult, len(urls))

	// 启动下载协程
	for i, url := range urls {
		go func(idx int, imgURL string) {
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 生成文件名
			ext := ".jpg"
			if parts := strings.Split(imgURL, "."); len(parts) > 1 {
				ext = "." + parts[len(parts)-1]
			}
			filename := fmt.Sprintf("%s_%d%s", prefix, idx, ext)

			// 下载图片
			localPath, err := d.downloadImage(ctx, imgURL, filename)
			resultChan <- downloadResult{
				Path: localPath,
				Err:  err,
			}
		}(i, url)
	}

	// 收集结果
	results := make([]string, 0, len(urls))
	for i := 0; i < len(urls); i++ {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case result := <-resultChan:
			if result.Err != nil {
				if d.logger != nil {
					d.logger.Warn("image download failed",
						zap.Error(result.Err))
				}
				continue
			}
			results = append(results, result.Path)
		}
	}

	return results, nil
}

// downloadResult 下载结果
type downloadResult struct {
	Path string
	Err  error
}
