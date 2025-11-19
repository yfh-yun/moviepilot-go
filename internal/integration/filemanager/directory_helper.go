// Package filemanager 目录助手
package filemanager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// DirectoryHelper 目录助手
type DirectoryHelper struct {
	logger *zap.Logger
}

// NewDirectoryHelper 创建目录助手
func NewDirectoryHelper() *DirectoryHelper {
	return &DirectoryHelper{
		logger: logger.Logger,
	}
}

// TransferDirectoryConf 转移目录配置
type TransferDirectoryConf struct {
	Name           string `json:"name"`
	DownloadPath   string `json:"download_path"`
	LibraryPath    string `json:"library_path"`
	Storage        string `json:"storage"`
	LibraryStorage string `json:"library_storage"`
	TransferType   string `json:"transfer_type"`
	Enabled        bool   `json:"enabled"`
}

// GetDirs 获取目录配置
func (dh *DirectoryHelper) GetDirs(ctx context.Context) ([]*TransferDirectoryConf, error) {
	// 这里应该从数据库或配置文件中读取目录配置
	// 为了演示，返回一些示例配置
	dirs := []*TransferDirectoryConf{
		{
			Name:           "Movies",
			DownloadPath:   "/downloads/movies",
			LibraryPath:    "/library/movies",
			Storage:        "local",
			LibraryStorage: "local",
			TransferType:   "move",
			Enabled:        true,
		},
		{
			Name:           "TV Shows",
			DownloadPath:   "/downloads/tv",
			LibraryPath:    "/library/tv",
			Storage:        "local",
			LibraryStorage: "local",
			TransferType:   "move",
			Enabled:        true,
		},
	}

	return dirs, nil
}

// Exists 检查路径是否存在
func (dh *DirectoryHelper) Exists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return err == nil
}

// IsSameDisk 检查两个路径是否在同一磁盘
func (dh *DirectoryHelper) IsSameDisk(path1, path2 string) bool {
	if path1 == "" || path2 == "" {
		return false
	}

	// 获取路径的设备信息
	stat1, err := os.Stat(path1)
	if err != nil {
		return false
	}

	stat2, err := os.Stat(path2)
	if err != nil {
		return false
	}

	// 比较设备ID
	return os.SameFile(stat1, stat2)
}

// CreateDirectory 创建目录
func (dh *DirectoryHelper) CreateDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	return os.MkdirAll(path, 0755)
}

// GetDirectorySize 获取目录大小
func (dh *DirectoryHelper) GetDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// ListDirectories 列出目录
func (dh *DirectoryHelper) ListDirectories(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(path, entry.Name()))
		}
	}

	return dirs, nil
}

// ListFiles 列出文件
func (dh *DirectoryHelper) ListFiles(path string, recursive bool) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	var files []string

	if recursive {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, filePath)
			}
			return nil
		})
		return files, err
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(path, entry.Name()))
			}
		}
	}

	return files, nil
}

// GetFileInfo 获取文件信息
func (dh *DirectoryHelper) GetFileInfo(path string) (*models.FileInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &models.FileInfo{
		Path:    path,
		Name:    filepath.Base(path),
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime().Unix(),
	}, nil
}

// IsMediaFile 判断是否为媒体文件
func (dh *DirectoryHelper) IsMediaFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	mediaExtensions := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".m4v":  true,
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".aac":  true,
		".ogg":  true,
		".m4a":  true,
		".wma":  true,
	}

	return mediaExtensions[ext]
}

// IsSubtitleFile 判断是否为字幕文件
func (dh *DirectoryHelper) IsSubtitleFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	subtitleExtensions := map[string]bool{
		".srt":  true,
		".ass":  true,
		".ssa":  true,
		".sub":  true,
		".vtt":  true,
		".sup":  true,
	}

	return subtitleExtensions[ext]
}

// IsVideoFile 判断是否为视频文件
func (dh *DirectoryHelper) IsVideoFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	videoExtensions := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".m4v":  true,
		".3gp":  true,
		".asf":  true,
		".rm":   true,
		".rmvb": true,
		".divx": true,
		".xvid": true,
	}

	return videoExtensions[ext]
}

// IsAudioFile 判断是否为音频文件
func (dh *DirectoryHelper) IsAudioFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	audioExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".aac":  true,
		".ogg":  true,
		".m4a":  true,
		".wma":  true,
		".ape":  true,
		".opus": true,
	}

	return audioExtensions[ext]
}

// GetRelativePath 获取相对路径
func (dh *DirectoryHelper) GetRelativePath(basePath, targetPath string) (string, error) {
	if basePath == "" || targetPath == "" {
		return "", fmt.Errorf("base path and target path cannot be empty")
	}

	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// JoinPath 连接路径
func (dh *DirectoryHelper) JoinPath(paths ...string) string {
	return filepath.Join(paths...)
}

// SplitPath 分割路径
func (dh *DirectoryHelper) SplitPath(path string) (dir, file string) {
	return filepath.Split(path)
}

// CleanPath 清理路径
func (dh *DirectoryHelper) CleanPath(path string) string {
	return filepath.Clean(path)
}

// GetParentDir 获取父目录
func (dh *DirectoryHelper) GetParentDir(path string) string {
	return filepath.Dir(path)
}

// GetFileName 获取文件名
func (dh *DirectoryHelper) GetFileName(path string) string {
	return filepath.Base(path)
}

// GetFileExt 获取文件扩展名
func (dh *DirectoryHelper) GetFileExt(path string) string {
	return filepath.Ext(path)
}

// GetFileNameWithoutExt 获取不带扩展名的文件名
func (dh *DirectoryHelper) GetFileNameWithoutExt(path string) string {
	name := dh.GetFileName(path)
	return strings.TrimSuffix(name, dh.GetFileExt(path))
}

// IsHidden 判断文件或目录是否为隐藏
func (dh *DirectoryHelper) IsHidden(path string) bool {
	name := dh.GetFileName(path)
	return strings.HasPrefix(name, ".")
}

// GetDiskUsage 获取磁盘使用情况
func (dh *DirectoryHelper) GetDiskUsage(path string) (*DiskUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return &DiskUsage{
		Total: total,
		Used:  used,
		Free:  free,
	}, nil
}

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// FindFiles 查找文件
func (dh *DirectoryHelper) FindFiles(rootPath, pattern string, recursive bool) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !recursive && filepath.Dir(path) != rootPath {
			return nil
		}

		if !info.IsDir() {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

// GetFileHash 获取文件哈希
func (dh *DirectoryHelper) GetFileHash(path string, algorithm string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var hash hash.Hash
	switch strings.ToLower(algorithm) {
	case "md5":
		hash = md5.New()
	case "sha1":
		hash = sha1.New()
	case "sha256":
		hash = sha256.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}