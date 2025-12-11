package storage

import (
	"context"
	"sort"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// DirectoryService 目录管理服务
type DirectoryService interface {
	// GetDirectories 获取所有目录配置
	GetDirectories(ctx context.Context) ([]*TransferDirectoryConf, error)

	// GetDownloadDirectories 获取所有下载目录
	GetDownloadDirectories(ctx context.Context) ([]*TransferDirectoryConf, error)

	// GetLocalDownloadDirectories 获取所有本地下载目录
	GetLocalDownloadDirectories(ctx context.Context) ([]*TransferDirectoryConf, error)

	// GetLibraryDirectories 获取所有媒体库目录
	GetLibraryDirectories(ctx context.Context) ([]*TransferDirectoryConf, error)

	// GetLocalLibraryDirectories 获取所有本地媒体库目录
	GetLocalLibraryDirectories(ctx context.Context) ([]*TransferDirectoryConf, error)
}

// TransferDirectoryConf 传输目录配置
type TransferDirectoryConf struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	DownloadPath string `json:"download_path"`
	LibraryPath  string `json:"library_path"`
	Storage      string `json:"storage"`
	Priority     int    `json:"priority"`
	Enabled      bool   `json:"enabled"`
	Category     string `json:"category"`
	Comment      string `json:"comment"`
}

// directoryService 目录管理服务实现
type directoryService struct {
	logger *zap.Logger
}

// NewDirectoryService 创建目录管理服务
func NewDirectoryService() DirectoryService {
	return &directoryService{
		logger: logger.GetLogger(),
	}
}

// GetDirectories 获取所有目录配置
func (s *directoryService) GetDirectories(ctx context.Context) ([]*TransferDirectoryConf, error) {
	s.logger.Info("获取所有目录配置")

	// TODO: 从数据库或配置中获取目录配置
	// 这里简化处理，返回模拟数据
	dirs := []*TransferDirectoryConf{
		{
			ID:           1,
			Name:         "默认目录",
			DownloadPath: "/downloads",
			LibraryPath:  "/library",
			Storage:      "local",
			Priority:     100,
			Enabled:      true,
			Category:     "movie",
			Comment:      "默认下载和媒体库目录",
		},
		{
			ID:           2,
			Name:         "电视剧目录",
			DownloadPath: "/downloads/tv",
			LibraryPath:  "/library/tv",
			Storage:      "local",
			Priority:     90,
			Enabled:      true,
			Category:     "tv",
			Comment:      "电视剧下载和媒体库目录",
		},
	}

	return dirs, nil
}

// GetDownloadDirectories 获取所有下载目录
func (s *directoryService) GetDownloadDirectories(ctx context.Context) ([]*TransferDirectoryConf, error) {
	s.logger.Info("获取所有下载目录")

	// 1. 获取所有目录
	dirs, err := s.GetDirectories(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 过滤出有下载路径的目录
	downloadDirs := make([]*TransferDirectoryConf, 0, len(dirs))
	for _, dir := range dirs {
		if dir.DownloadPath != "" {
			downloadDirs = append(downloadDirs, dir)
		}
	}

	// 3. 按优先级排序
	sort.Slice(downloadDirs, func(i, j int) bool {
		return downloadDirs[i].Priority > downloadDirs[j].Priority
	})

	return downloadDirs, nil
}

// GetLocalDownloadDirectories 获取所有本地下载目录
func (s *directoryService) GetLocalDownloadDirectories(ctx context.Context) ([]*TransferDirectoryConf, error) {
	s.logger.Info("获取所有本地下载目录")

	// 1. 获取所有下载目录
	downloadDirs, err := s.GetDownloadDirectories(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 过滤出本地存储的目录
	localDirs := make([]*TransferDirectoryConf, 0, len(downloadDirs))
	for _, dir := range downloadDirs {
		if dir.Storage == "local" {
			localDirs = append(localDirs, dir)
		}
	}

	return localDirs, nil
}

// GetLibraryDirectories 获取所有媒体库目录
func (s *directoryService) GetLibraryDirectories(ctx context.Context) ([]*TransferDirectoryConf, error) {
	s.logger.Info("获取所有媒体库目录")

	// 1. 获取所有目录
	dirs, err := s.GetDirectories(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 过滤出有媒体库路径的目录
	libraryDirs := make([]*TransferDirectoryConf, 0, len(dirs))
	for _, dir := range dirs {
		if dir.LibraryPath != "" {
			libraryDirs = append(libraryDirs, dir)
		}
	}

	// 3. 按优先级排序
	sort.Slice(libraryDirs, func(i, j int) bool {
		return libraryDirs[i].Priority > libraryDirs[j].Priority
	})

	return libraryDirs, nil
}

// GetLocalLibraryDirectories 获取所有本地媒体库目录
func (s *directoryService) GetLocalLibraryDirectories(ctx context.Context) ([]*TransferDirectoryConf, error) {
	s.logger.Info("获取所有本地媒体库目录")

	// 1. 获取所有媒体库目录
	libraryDirs, err := s.GetLibraryDirectories(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 过滤出本地存储的目录
	localDirs := make([]*TransferDirectoryConf, 0, len(libraryDirs))
	for _, dir := range libraryDirs {
		if dir.Storage == "local" {
			localDirs = append(localDirs, dir)
		}
	}

	return localDirs, nil
}
