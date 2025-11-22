package storage

import (
	"context"
	"io/fs"
	"path/filepath"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/utils/workerpool"
)

// ConcurrentScanner 并发扫描器
type ConcurrentScanner struct {
	concurrency int
	logger      *zap.Logger
}

// NewConcurrentScanner 创建并发扫描器
func NewConcurrentScanner(concurrency int, logger *zap.Logger) *ConcurrentScanner {
	if concurrency <= 0 {
		concurrency = 10 // 默认并发数
	}

	return &ConcurrentScanner{
		concurrency: concurrency,
		logger:      logger,
	}
}

// ScanConcurrent 并发扫描目录
func (s *ConcurrentScanner) ScanConcurrent(ctx context.Context, opts ScanOptions) ([]FileItem, error) {
	var (
		mu    sync.Mutex
		files []FileItem
	)

	// 创建 Goroutine 池
	pool := workerpool.New(s.concurrency)
	defer pool.Wait()

	// 记录开始时间
	if s.logger != nil {
		s.logger.Info("starting concurrent scan",
			zap.String("root_path", opts.RootPath),
			zap.Int("concurrency", s.concurrency))
	}

	// 遍历目录
	err := filepath.WalkDir(opts.RootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("walk error", zap.String("path", path), zap.Error(err))
			}
			return nil // 继续处理其他文件
		}

		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 检查深度限制
		if opts.MaxDepth > 0 {
			depth := calculateDepth(opts.RootPath, path)
			if depth > opts.MaxDepth {
				return nil
			}
		}

		// 检查包含/排除规则
		if !shouldInclude(path, opts.Include, opts.Exclude) {
			return nil
		}

		// 提交任务到池
		pool.Submit(func() {
			fileInfo, err := d.Info()
			if err != nil {
				if s.logger != nil {
					s.logger.Warn("failed to get file info",
						zap.String("path", path),
						zap.Error(err))
				}
				return
			}

			item := FileItem{
				Path:    path,
				Size:    fileInfo.Size(),
				ModTime: fileInfo.ModTime(),
			}

			mu.Lock()
			files = append(files, item)
			mu.Unlock()
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("concurrent scan completed",
			zap.Int("file_count", len(files)))
	}

	return files, nil
}

// calculateDepth 计算路径深度
func calculateDepth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 0
	}

	depth := 0
	for _, c := range rel {
		if c == filepath.Separator {
			depth++
		}
	}
	return depth
}

// shouldInclude 检查文件是否应该包含
func shouldInclude(path string, include, exclude []string) bool {
	// 检查排除规则
	for _, pattern := range exclude {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return false
		}
	}

	// 如果没有包含规则，默认包含
	if len(include) == 0 {
		return true
	}

	// 检查包含规则
	for _, pattern := range include {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}
