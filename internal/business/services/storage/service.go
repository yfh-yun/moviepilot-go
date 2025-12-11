package storage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service 定义与存储相关的核心接口。
type Service interface {
	// Scan 根据参数扫描文件系统，返回文件项
	Scan(opts ScanOptions) ([]FileItem, error)
	// Transfer 根据计划执行文件转移，返回每个任务的结果
	Transfer(tasks []TransferTask) ([]TransferResult, error)
}

// ScanOptions 描述扫描行为。
type ScanOptions struct {
	RootPath       string
	Include        []string
	Exclude        []string
	MaxDepth       int
	FollowSymlinks bool
}

// FileItem 表示扫描到的文件。
type FileItem struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// TransferMode 文件转移模式。
type TransferMode string

const (
	TransferModeMove     TransferMode = "move"
	TransferModeCopy     TransferMode = "copy"
	TransferModeHardlink TransferMode = "hardlink"
)

// TransferTask 描述一次转移任务。
type TransferTask struct {
	SourcePath string
	TargetPath string
	Mode       TransferMode
	Overwrite  bool
}

// TransferResult 记录转移结果。
type TransferResult struct {
	Task   TransferTask
	Status string
	Error  error
}

// LocalService 是基于本地文件系统的实现。
type LocalService struct {
	logger *zap.Logger
}

// NewLocalService 创建实例。
func NewLocalService(logger *zap.Logger) *LocalService {
	return &LocalService{logger: logger}
}

// NewStorageService 创建存储服务实例（路由使用）
func NewStorageService(db *gorm.DB, logger *zap.Logger) *LocalService {
	return &LocalService{logger: logger}
}

// Scan 执行目录扫描。
func (s *LocalService) Scan(opts ScanOptions) ([]FileItem, error) {
	if opts.RootPath == "" {
		return nil, errors.New("root path is required")
	}

	info, err := os.Stat(opts.RootPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("root path must be directory")
	}

	root := filepath.Clean(opts.RootPath)
	rootDepth := depth(root)
	files := make([]FileItem, 0)

	walkFn := func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if s.logger != nil {
				s.logger.Warn("scan error", zap.String("path", path), zap.Error(walkErr))
			}
			return nil
		}

		currentDepth := depth(path) - rootDepth
		if opts.MaxDepth > 0 && currentDepth > opts.MaxDepth {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !opts.FollowSymlinks && entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		if len(opts.Include) > 0 && !matchAny(path, opts.Include) {
			return nil
		}
		if len(opts.Exclude) > 0 && matchAny(path, opts.Exclude) {
			return nil
		}

		stat, err := entry.Info()
		if err != nil {
			return nil
		}

		files = append(files, FileItem{
			Path:    path,
			Size:    stat.Size(),
			ModTime: stat.ModTime(),
			IsDir:   entry.IsDir(),
		})
		return nil
	}

	if err := filepath.WalkDir(root, walkFn); err != nil {
		return nil, err
	}

	return files, nil
}

// Transfer 目前为占位实现，仅返回任务列表，后续可对接实际文件操作或插件。
func (s *LocalService) Transfer(tasks []TransferTask) ([]TransferResult, error) {
	results := make([]TransferResult, 0, len(tasks))
	for _, task := range tasks {
		results = append(results, TransferResult{
			Task:   task,
			Status: "planned",
			Error:  nil,
		})
	}
	return results, nil
}

func depth(path string) int {
	cleaned := filepath.Clean(path)
	if cleaned == string(os.PathSeparator) {
		return 0
	}
	parts := strings.Split(cleaned, string(os.PathSeparator))
	return len(parts)
}

func matchAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		if ok, _ := filepath.Match(pattern, filepath.Base(path)); ok {
			return true
		}
	}
	return false
}
