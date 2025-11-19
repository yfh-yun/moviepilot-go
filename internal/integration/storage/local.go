package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LocalStorage 本地存储提供商
type LocalStorage struct {
	name      string
	basePath  string
	connected bool
	mu        sync.RWMutex
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(name, basePath string) *LocalStorage {
	return &LocalStorage{
		name:     name,
		basePath: basePath,
	}
}

// Name 返回存储名称
func (ls *LocalStorage) Name() string {
	return ls.name
}

// Type 返回存储类型
func (ls *LocalStorage) Type() string {
	return ProviderLocal
}

// IsConnected 检查是否连接
func (ls *LocalStorage) IsConnected() bool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.connected
}

// Connect 连接存储
func (ls *LocalStorage) Connect(ctx context.Context) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// 检查基础路径是否存在
	if err := os.MkdirAll(ls.basePath, 0755); err != nil {
		return fmt.Errorf("创建基础路径失败: %w", err)
	}

	// 检查路径是否可写
	testFile := filepath.Join(ls.basePath, ".test_write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("路径不可写: %w", err)
	}
	os.Remove(testFile)

	ls.connected = true
	return nil
}

// Disconnect 断开连接
func (ls *LocalStorage) Disconnect() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.connected = false
	return nil
}

// Upload 上传文件
func (ls *LocalStorage) Upload(ctx context.Context, filePath string, reader io.Reader, size int64) error {
	if !ls.IsConnected() {
		return ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, filePath)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建临时文件
	tempFile := fullPath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer file.Close()
	defer os.Remove(tempFile) // 清理临时文件

	// 写入文件
	written, err := io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	if size > 0 && written != size {
		return fmt.Errorf("文件大小不匹配: 期望 %d, 实际 %d", size, written)
	}

	// 重命名为目标文件
	if err := os.Rename(tempFile, fullPath); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// Download 下载文件
func (ls *LocalStorage) Download(ctx context.Context, filePath string) (io.ReadCloser, error) {
	if !ls.IsConnected() {
		return nil, ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, filePath)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	return file, nil
}

// Delete 删除文件
func (ls *LocalStorage) Delete(ctx context.Context, filePath string) error {
	if !ls.IsConnected() {
		return ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, filePath)

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (ls *LocalStorage) Exists(ctx context.Context, filePath string) (bool, error) {
	if !ls.IsConnected() {
		return false, ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, filePath)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("检查文件存在失败: %w", err)
	}

	return true, nil
}

// Move 移动文件
func (ls *LocalStorage) Move(ctx context.Context, srcPath, dstPath string) error {
	if !ls.IsConnected() {
		return ErrNotConnected
	}

	srcFullPath := filepath.Join(ls.basePath, srcPath)
	dstFullPath := filepath.Join(ls.basePath, dstPath)

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dstFullPath), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	err := os.Rename(srcFullPath, dstFullPath)
	if err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	return nil
}

// Copy 复制文件
func (ls *LocalStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	if !ls.IsConnected() {
		return ErrNotConnected
	}

	srcFullPath := filepath.Join(ls.basePath, srcPath)
	dstFullPath := filepath.Join(ls.basePath, dstPath)

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dstFullPath), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	srcFile, err := os.Open(srcFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstFullPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	return nil
}

// List 列出目录内容
func (ls *LocalStorage) List(ctx context.Context, path string) ([]FileInfo, error) {
	if !ls.IsConnected() {
		return nil, ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var fileInfos []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:         entry.Name(),
			Path:         filepath.Join(path, entry.Name()),
			IsDir:        entry.IsDir(),
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
			MimeType:     ls.getMimeType(entry.Name(), entry.IsDir()),
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

// Mkdir 创建目录
func (ls *LocalStorage) Mkdir(ctx context.Context, path string) error {
	if !ls.IsConnected() {
		return ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, path)

	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	return nil
}

// Rmdir 删除目录
func (ls *LocalStorage) Rmdir(ctx context.Context, path string) error {
	if !ls.IsConnected() {
		return ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, path)

	err := os.RemoveAll(fullPath)
	if err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}

	return nil
}

// GetQuota 获取配额信息
func (ls *LocalStorage) GetQuota(ctx context.Context) (*QuotaInfo, error) {
	if !ls.IsConnected() {
		return nil, ErrNotConnected
	}

	var stat syscall.Statfs_t
	err := syscall.Statfs(ls.basePath, &stat)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}

	// 计算总容量和可用空间
	total := stat.Blocks * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	used := total - available

	return &QuotaInfo{
		Total:     int64(total),
		Used:      int64(used),
		Available: int64(available),
	}, nil
}

// GetFileInfo 获取文件信息
func (ls *LocalStorage) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	if !ls.IsConnected() {
		return nil, ErrNotConnected
	}

	fullPath := filepath.Join(ls.basePath, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	isDir := info.IsDir()
	checksum, _ := ls.calculateChecksum(fullPath, isDir)

	fileInfo := &FileInfo{
		Name:         filepath.Base(path),
		Path:         path,
		IsDir:        isDir,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		MimeType:     ls.getMimeType(info.Name(), isDir),
		Checksum:     checksum,
	}

	return fileInfo, nil
}

// getMimeType 获取MIME类型
func (ls *LocalStorage) getMimeType(name string, isDir bool) string {
	if isDir {
		return "inode/directory"
	}

	ext := filepath.Ext(name)
	switch strings.ToLower(ext) {
	case ".txt", ".md":
		return "text/plain"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".mp4":
		return "video/mp4"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

// calculateChecksum 计算文件校验和
func (ls *LocalStorage) calculateChecksum(path string, isDir bool) (string, error) {
	if isDir {
		return "", nil // 目录不计算校验和
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
