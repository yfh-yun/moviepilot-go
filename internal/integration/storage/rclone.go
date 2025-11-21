package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"moviepilot-go/internal/infrastructure/config"
	"moviepilot-go/pkg/logger"
)

// RCloneStorage RClone存储实现
type RCloneStorage struct {
	name       string
	config     *config.RCloneConfig
	logger     *logger.Logger
	rclonePath string
	remoteName string
}

// NewRCloneStorage 创建RClone存储实例
func NewRCloneStorage(cfg *config.Config) (*RCloneStorage, error) {
	// 检查rclone是否可用
	rclonePath, err := exec.LookPath("rclone")
	if err != nil {
		return nil, fmt.Errorf("rclone未安装，请先安装rclone")
	}

	storage := &RCloneStorage{
		name:       "RClone",
		config:     &cfg.RClone,
		logger:     logger.NewLogger("storage-rclone"),
		rclonePath: rclonePath,
		remoteName: cfg.RClone.RemoteName,
	}

	// 验证远程配置
	if err := storage.validateRemote(); err != nil {
		return nil, fmt.Errorf("RClone远程配置验证失败: %w", err)
	}

	return storage, nil
}

// validateRemote 验证远程配置
func (s *RCloneStorage) validateRemote() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 执行rclone lsd命令验证远程
	cmd := exec.CommandContext(ctx, s.rclonePath, "lsd", s.remoteName+":")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("RClone远程验证失败: %s, 输出: %s", err, string(output))
	}

	s.logger.Infof("RClone远程验证成功: %s", s.remoteName)
	return nil
}

// Name 获取存储名称
func (s *RCloneStorage) Name() string {
	return s.name
}

// Type 获取存储类型
func (s *RCloneStorage) Type() string {
	return "rclone"
}

// IsAvailable 检查存储是否可用
func (s *RCloneStorage) IsAvailable() bool {
	// 检查rclone是否可执行
	if _, err := os.Stat(s.rclonePath); err != nil {
		return false
	}

	// 简单验证远程
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.rclonePath, "lsd", s.remoteName+":")
	return cmd.Run() == nil
}

// ListFiles 列出文件
func (s *RCloneStorage) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	remotePath := s.buildRemotePath(path)

	cmd := exec.CommandContext(ctx, s.rclonePath, "lsjson", remotePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("RClone列出文件失败: %w", err)
	}

	// 解析JSON输出
	var files []struct {
		Path     string    `json:"Path"`
		Name     string    `json:"Name"`
		Size     int64     `json:"Size"`
		ModTime  time.Time `json:"ModTime"`
		IsDir    bool      `json:"IsDir"`
		MimeType string    `json:"MimeType"`
	}

	if err := s.parseJSONOutput(output, &files); err != nil {
		return nil, fmt.Errorf("解析RClone输出失败: %w", err)
	}

	var result []FileInfo
	for _, file := range files {
		fileType := FileTypeFile
		if file.IsDir {
			fileType = FileTypeDirectory
		}

		result = append(result, FileInfo{
			Name:    file.Name,
			Path:    filepath.Join(path, file.Name),
			Size:    file.Size,
			Type:    fileType,
			ModTime: file.ModTime,
		})
	}

	return result, nil
}

// UploadFile 上传文件
func (s *RCloneStorage) UploadFile(ctx context.Context, localPath, remotePath string) error {
	sourcePath := localPath
	destPath := s.buildRemotePath(remotePath)

	s.logger.Infof("开始上传文件: %s -> %s", sourcePath, destPath)

	// 使用copy命令上传文件
	cmd := exec.CommandContext(ctx, s.rclonePath, "copy",
		sourcePath, destPath,
		"--progress",       // 显示进度
		"--transfers", "4", // 并发传输数
		"--checkers", "8", // 并发检查数
		"--buffer-size", "64M", // 缓冲区大小
		"--retries", "3", // 重试次数
		"--low-level-retries", "10", // 低级重试
	)

	// 实时输出进度
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("RClone上传文件失败: %w", err)
	}

	s.logger.Infof("文件上传成功: %s", remotePath)
	return nil
}

// DownloadFile 下载文件
func (s *RCloneStorage) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	sourcePath := s.buildRemotePath(remotePath)
	destPath := localPath

	s.logger.Infof("开始下载文件: %s -> %s", sourcePath, destPath)

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 使用copy命令下载文件
	cmd := exec.CommandContext(ctx, s.rclonePath, "copy",
		sourcePath, destPath,
		"--progress",
		"--transfers", "4",
		"--checkers", "8",
		"--buffer-size", "64M",
		"--retries", "3",
		"--low-level-retries", "10",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("RClone下载文件失败: %w", err)
	}

	s.logger.Infof("文件下载成功: %s", localPath)
	return nil
}

// DeleteFile 删除文件
func (s *RCloneStorage) DeleteFile(ctx context.Context, path string) error {
	remotePath := s.buildRemotePath(path)

	s.logger.Infof("删除文件: %s", remotePath)

	cmd := exec.CommandContext(ctx, s.rclonePath, "delete", remotePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("RClone删除文件失败: %s, 输出: %s", err, string(output))
	}

	s.logger.Infof("文件删除成功: %s", path)
	return nil
}

// DeleteDirectory 删除目录
func (s *RCloneStorage) DeleteDirectory(ctx context.Context, path string) error {
	remotePath := s.buildRemotePath(path)

	s.logger.Infof("删除目录: %s", remotePath)

	cmd := exec.CommandContext(ctx, s.rclonePath, "purge", remotePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("RClone删除目录失败: %s, 输出: %s", err, string(output))
	}

	s.logger.Infof("目录删除成功: %s", path)
	return nil
}

// SyncDirectory 同步目录
func (s *RCloneStorage) SyncDirectory(ctx context.Context, sourcePath, destPath string) error {
	sourceRemote := s.buildRemotePath(sourcePath)
	destRemote := s.buildRemotePath(destPath)

	s.logger.Infof("同步目录: %s -> %s", sourceRemote, destRemote)

	cmd := exec.CommandContext(ctx, s.rclonePath, "sync",
		sourceRemote, destRemote,
		"--progress",
		"--transfers", "4",
		"--checkers", "8",
		"--delete-during",   // 同步时删除目标多余文件
		"--delete-excluded", // 删除排除的文件
		"--retries", "3",
		"--low-level-retries", "10",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("RClone同步目录失败: %w", err)
	}

	s.logger.Infof("目录同步成功: %s -> %s", sourcePath, destPath)
	return nil
}

// GetFileInfo 获取文件信息
func (s *RCloneStorage) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	remotePath := s.buildRemotePath(path)

	cmd := exec.CommandContext(ctx, s.rclonePath, "lsjson", remotePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("RClone获取文件信息失败: %w", err)
	}

	var files []struct {
		Path    string    `json:"Path"`
		Name    string    `json:"Name"`
		Size    int64     `json:"Size"`
		ModTime time.Time `json:"ModTime"`
		IsDir   bool      `json:"IsDir"`
	}

	if err := s.parseJSONOutput(output, &files); err != nil {
		return nil, fmt.Errorf("解析RClone输出失败: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("文件不存在: %s", path)
	}

	file := files[0]
	fileType := FileTypeFile
	if file.IsDir {
		fileType = FileTypeDirectory
	}

	return &FileInfo{
		Name:    file.Name,
		Path:    path,
		Size:    file.Size,
		Type:    fileType,
		ModTime: file.ModTime,
	}, nil
}

// buildRemotePath 构建远程路径
func (s *RCloneStorage) buildRemotePath(path string) string {
	if path == "" || path == "/" {
		return s.remoteName + ":"
	}
	return s.remoteName + ":" + strings.TrimPrefix(path, "/")
}

// parseJSONOutput 解析JSON输出
func (s *RCloneStorage) parseJSONOutput(output []byte, target interface{}) error {
	// 移除可能的空行和无效字符
	jsonStr := strings.TrimSpace(string(output))
	if jsonStr == "" {
		return fmt.Errorf("空的RClone输出")
	}

	// 使用标准JSON解析
	return s.decodeJSON([]byte(jsonStr), target)
}

// decodeJSON JSON解码（简化实现）
func (s *RCloneStorage) decodeJSON(data []byte, target interface{}) error {
	// 这里应该使用JSON解码器
	// 简化实现，实际需要处理JSON解析
	return fmt.Errorf("JSON解码未实现")
}

// RunCommand 执行自定义RClone命令
func (s *RCloneStorage) RunCommand(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, s.rclonePath, args...)
	return cmd.CombinedOutput()
}

// GetRemoteStats 获取远程统计信息
func (s *RCloneStorage) GetRemoteStats(ctx context.Context) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, s.rclonePath, "about", s.remoteName+":")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取远程统计失败: %w", err)
	}

	// 解析about命令输出
	stats := make(map[string]interface{})
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				stats[key] = value
			}
		}
	}

	return stats, nil
}
