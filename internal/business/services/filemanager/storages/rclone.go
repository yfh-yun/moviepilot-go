package storages

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/integration/filemanager"
)

// RcloneProvider Rclone云存储提供商
type RcloneProvider struct {
	config  *RcloneConfig
	logger  *zap.Logger
	binPath string
}

// RcloneConfig Rclone配置
type RcloneConfig struct {
	RemoteName     string            `json:"remote_name"`
	RemoteType     string            `json:"remote_type"`
	Config         map[string]string `json:"config"`
	BinPath        string            `json:"bin_path"`
	Timeout        int               `json:"timeout"`
	RootPath       string            `json:"root_path"`
	MaxRetries     int               `json:"max_retries"`
	EnableLogging  bool              `json:"enable_logging"`
	LogFile        string            `json:"log_file"`
	Transfers      int               `json:"transfers"`
	Checkers       int               `json:"checkers"`
	BufferSize     string            `json:"buffer_size"`
	DisableChecksum bool             `json:"disable_checksum"`
}

// RcloneRemote Rclone远程存储配置
type RcloneRemote struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
}

// RcloneItem Rclone文件项目
type RcloneItem struct {
	Name      string    `json:"Name"`
	Size      int64     `json:"Size"`
	ModTime   time.Time `json:"ModTime"`
	IsDir     bool      `json:"IsDir"`
	Path      string    `json:"Path"`
	MimeType  string    `json:"MimeType"`
	Hash      string    `json:"Hash,omitempty"`
}

// RcloneStats Rclone统计信息
type RcloneStats struct {
	Transfers   int64   `json:"transfers"`
	Bytes       int64   `json:"bytes"`
	Speed       float64 `json:"speed"`
	AverageSpeed float64 `json:"averageSpeed"`
	BytesTotal  int64   `json:"bytesTotal"`
	Errors      int64   `json:"errors"`
}

// NewRcloneProvider 创建Rclone提供商
func NewRcloneProvider(logger *zap.Logger) filemanager.StorageProvider {
	return &RcloneProvider{
		logger: logger,
	}
}

// GetName 获取存储名称
func (p *RcloneProvider) GetName() string {
	return "Rclone"
}

// GetType 获取存储类型
func (p *RcloneProvider) GetType() string {
	return "rclone"
}

// ValidateConfig 验证配置
func (p *RcloneProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg RcloneConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.RemoteName == "" {
		return fmt.Errorf("RemoteName不能为空")
	}
	
	if cfg.RemoteType == "" {
		return fmt.Errorf("RemoteType不能为空")
	}
	
	if cfg.Config == nil {
		return fmt.Errorf("Config不能为空")
	}
	
	return nil
}

// Initialize 初始化存储
func (p *RcloneProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &RcloneConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析Rclone配置失败: %w", err)
	}
	
	// 验证配置
	if err := p.ValidateConfig(config); err != nil {
		return err
	}
	
	// 查找rclone二进制文件
	if err := p.findRcloneBinary(); err != nil {
		return fmt.Errorf("查找rclone二进制文件失败: %w", err)
	}
	
	// 配置远程存储
	if err := p.configureRemote(); err != nil {
		return fmt.Errorf("配置rclone远程存储失败: %w", err)
	}
	
	// 测试连接
	if err := p.testConnection(); err != nil {
		return fmt.Errorf("Rclone连接测试失败: %w", err)
	}
	
	// 设置默认值
	if p.config.MaxRetries == 0 {
		p.config.MaxRetries = 3
	}
	
	if p.config.Transfers == 0 {
		p.config.Transfers = 4
	}
	
	if p.config.Checkers == 0 {
		p.config.Checkers = 8
	}
	
	if p.config.BufferSize == "" {
		p.config.BufferSize = "16M"
	}
	
	if p.config.RootPath == "" {
		p.config.RootPath = "/"
	}
	
	p.logger.Info("Rclone存储提供商初始化成功")
	return nil
}

// ListFiles 列出文件
func (p *RcloneProvider) ListFiles(ctx context.Context, path string) ([]*filemanager.FileInfo, error) {
	remotePath := p.buildRemotePath(path)
	
	cmd := exec.CommandContext(ctx, p.binPath, "lsjson", remotePath)
	cmd.Env = p.buildEnv()
	
	output, err := p.runCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("列出文件失败: %w", err)
	}
	
	var items []RcloneItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("解析文件列表失败: %w", err)
	}
	
	return p.convertItems(items)
}

// UploadFile 上传文件
func (p *RcloneProvider) UploadFile(ctx context.Context, localPath, remotePath string, progress *filemanager.UploadProgress) error {
	remotePath = p.buildRemotePath(remotePath)
	
	// 构建rclone copy命令
	args := []string{"copy", localPath, remotePath}
	
	// 添加参数
	args = p.addCommonArgs(args)
	
	cmd := exec.CommandContext(ctx, p.binPath, args...)
	cmd.Env = p.buildEnv()
	
	// 如果需要监控进度，使用进度回调
	if progress != nil {
		return p.runUploadCommand(cmd, progress)
	}
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}
	
	return nil
}

// UploadStream 上传流
func (p *RcloneProvider) UploadStream(ctx context.Context, stream io.Reader, remotePath string, size int64, progress *filemanager.UploadProgress) error {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "rclone_upload_*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 复制流到临时文件
	written, err := io.Copy(tempFile, stream)
	if err != nil {
		return fmt.Errorf("保存临时文件失败: %w", err)
	}
	
	// 上传临时文件
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}
	
	return p.UploadFile(ctx, tempFile.Name(), remotePath, progress)
}

// DownloadFile 下载文件
func (p *RcloneProvider) DownloadFile(ctx context.Context, remotePath, localPath string, progress *filemanager.DownloadProgress) error {
	remotePath = p.buildRemotePath(remotePath)
	
	// 创建本地目录
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}
	
	// 构建rclone copy命令
	args := []string{"copy", remotePath, localPath}
	
	// 添加参数
	args = p.addCommonArgs(args)
	
	cmd := exec.CommandContext(ctx, p.binPath, args...)
	cmd.Env = p.buildEnv()
	
	// 如果需要监控进度，使用进度回调
	if progress != nil {
		return p.runDownloadCommand(cmd, progress)
	}
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	
	return nil
}

// DownloadStream 下载流
func (p *RcloneProvider) DownloadStream(ctx context.Context, remotePath string) (io.ReadCloser, error) {
	remotePath := p.buildRemotePath(remotePath)
	
	cmd := exec.CommandContext(ctx, p.binPath, "cat", remotePath)
	cmd.Env = p.buildEnv()
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建输出管道失败: %w", err)
	}
	
	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动rclone命令失败: %w", err)
	}
	
	// 返回可关闭的读取器
	return &rcloneReadCloser{
		pipe: stdout,
		cmd:   cmd,
	}, nil
}

// DeleteFile 删除文件
func (p *RcloneProvider) DeleteFile(ctx context.Context, path string) error {
	remotePath := p.buildRemotePath(path)
	
	cmd := exec.CommandContext(ctx, p.binPath, "delete", remotePath)
	cmd.Env = p.buildEnv()
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	
	return nil
}

// CreateDirectory 创建目录
func (p *RcloneProvider) CreateDirectory(ctx context.Context, path string) error {
	remotePath := p.buildRemotePath(path)
	
	cmd := exec.CommandContext(ctx, p.binPath, "mkdir", remotePath)
	cmd.Env = p.buildEnv()
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	return nil
}

// DeleteDirectory 删除目录
func (p *RcloneProvider) DeleteDirectory(ctx context.Context, path string, recursive bool) error {
	remotePath := p.buildRemotePath(path)
	
	args := []string{"rmdir", remotePath}
	if recursive {
		args = []string{"rmdirs", remotePath}
	}
	
	cmd := exec.CommandContext(ctx, p.binPath, args...)
	cmd.Env = p.buildEnv()
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}
	
	return nil
}

// RenameFile 重命名文件
func (p *RcloneProvider) RenameFile(ctx context.Context, oldPath, newPath string) error {
	oldRemotePath := p.buildRemotePath(oldPath)
	newRemotePath := p.buildRemotePath(newPath)
	
	cmd := exec.CommandContext(ctx, p.binPath, "moveto", oldRemotePath, newRemotePath)
	cmd.Env = p.buildEnv()
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}
	
	return nil
}

// CopyFile 复制文件
func (p *RcloneProvider) CopyFile(ctx context.Context, sourcePath, targetPath string) error {
	sourceRemotePath := p.buildRemotePath(sourcePath)
	targetRemotePath := p.buildRemotePath(targetPath)
	
	cmd := exec.CommandContext(ctx, p.binPath, "copy", sourceRemotePath, targetRemotePath)
	cmd.Env = p.buildEnv()
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}
	
	return nil
}

// MoveFile 移动文件
func (p *RcloneProvider) MoveFile(ctx context.Context, sourcePath, targetPath string) error {
	sourceRemotePath := p.buildRemotePath(sourcePath)
	targetRemotePath := p.buildRemotePath(targetPath)
	
	cmd := exec.CommandContext(ctx, p.binPath, "move", sourceRemotePath, targetRemotePath)
	cmd.Env = p.buildEnv()
	
	_, err := p.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}
	
	return nil
}

// GetFileInfo 获取文件信息
func (p *RcloneProvider) GetFileInfo(ctx context.Context, path string) (*filemanager.FileInfo, error) {
	remotePath := p.buildRemotePath(path)
	
	cmd := exec.CommandContext(ctx, p.binPath, "lsjson", remotePath)
	cmd.Env = p.buildEnv()
	
	output, err := p.runCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}
	
	var items []RcloneItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("解析文件信息失败: %w", err)
	}
	
	if len(items) == 0 {
		return nil, fmt.Errorf("文件不存在: %s", path)
	}
	
	return p.convertItem(items[0]), nil
}

// GetURL 获取文件访问URL
func (p *RcloneProvider) GetURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	// Rclone本身不提供直接访问URL，这个方法取决于具体的远程类型
	// 对于支持预签名的存储（如S3），可以生成预签名URL
	return "", fmt.Errorf("Rclone不支持直接生成访问URL")
}

// IsHealthy 健康检查
func (p *RcloneProvider) IsHealthy(ctx context.Context) bool {
	// 测试列出根目录
	_, err := p.ListFiles(ctx, "/")
	if err != nil {
		p.logger.Debug("Rclone健康检查失败", zap.Error(err))
		return false
	}
	
	return true
}

// GetConfig 获取配置
func (p *RcloneProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if cfg, exists := result["config"]; exists {
		if configMap, ok := cfg.(map[string]interface{}); ok {
			for key, value := range configMap {
				if str, ok := value.(string); ok && len(str) > 10 {
					// 隐藏可能敏感的配置项
					if strings.Contains(strings.ToLower(key), "key") ||
						strings.Contains(strings.ToLower(key), "secret") ||
						strings.Contains(strings.ToLower(key), "token") ||
						strings.Contains(strings.ToLower(key), "password") {
						configMap[key] = str[:10] + "***"
					}
				}
			}
		}
	}
	
	return result
}

// Close 关闭存储
func (p *RcloneProvider) Close() error {
	return nil
}

// findRcloneBinary 查找rclone二进制文件
func (p *RcloneProvider) findRcloneBinary() error {
	// 如果指定了路径，先检查指定路径
	if p.config.BinPath != "" {
		if _, err := os.Stat(p.config.BinPath); err == nil {
			p.binPath = p.config.BinPath
			return nil
		}
	}
	
	// 查找PATH中的rclone
	path, err := exec.LookPath("rclone")
	if err != nil {
		return fmt.Errorf("系统中未找到rclone二进制文件: %w", err)
	}
	
	p.binPath = path
	return nil
}

// configureRemote 配置远程存储
func (p *RcloneProvider) configureRemote() error {
	// 创建远程存储配置
	remote := RcloneRemote{
		Name:   p.config.RemoteName,
		Type:   p.config.RemoteType,
		Config: p.config.Config,
	}
	
	configJSON, err := json.Marshal(remote)
	if err != nil {
		return fmt.Errorf("编码远程配置失败: %w", err)
	}
	
	cmd := exec.Command(p.binPath, "config", "create", p.config.RemoteName, p.config.RemoteType)
	cmd.Stdin = strings.NewReader(string(configJSON))
	
	_, err = p.runCommand(cmd)
	if err != nil {
		// 配置可能已存在，这是正常的
		p.logger.Debug("配置远程存储（可能已存在）", zap.Error(err))
	}
	
	return nil
}

// testConnection 测试连接
func (p *RcloneProvider) testConnection() error {
	cmd := exec.Command(p.binPath, "lsd", p.config.RemoteName+":")
	
	_, err := p.runCommand(cmd)
	return err
}

// buildRemotePath 构建远程路径
func (p *RcloneProvider) buildRemotePath(path string) string {
	if p.config.RootPath != "/" {
		path = strings.TrimPrefix(path, "/")
		return fmt.Sprintf("%s:%s/%s", p.config.RemoteName, strings.TrimSuffix(p.config.RootPath, "/"), path)
	}
	
	return fmt.Sprintf("%s:%s", p.config.RemoteName, path)
}

// buildEnv 构建环境变量
func (p *RcloneProvider) buildEnv() []string {
	env := os.Environ()
	
	if p.config.EnableLogging && p.config.LogFile != "" {
		env = append(env, fmt.Sprintf("RCLONE_LOG_FILE=%s", p.config.LogFile))
	}
	
	return env
}

// addCommonArgs 添加通用参数
func (p *RcloneProvider) addCommonArgs(args []string) []string {
	if p.config.Transfers > 0 {
		args = append(args, "--transfers", strconv.Itoa(p.config.Transfers))
	}
	
	if p.config.Checkers > 0 {
		args = append(args, "--checkers", strconv.Itoa(p.config.Checkers))
	}
	
	if p.config.BufferSize != "" {
		args = append(args, "--buffer-size", p.config.BufferSize)
	}
	
	if p.config.DisableChecksum {
		args = append(args, "--checksum=false")
	}
	
	if p.config.Timeout > 0 {
		args = append(args, "--timeout", strconv.Itoa(p.config.Timeout))
	}
	
	return args
}

// runCommand 运行命令
func (p *RcloneProvider) runCommand(cmd *exec.Cmd) ([]byte, error) {
	p.logger.Debug("执行rclone命令", zap.String("cmd", cmd.String()))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Error("rclone命令执行失败",
			zap.String("cmd", cmd.String()),
			zap.Error(err),
			zap.String("output", string(output)))
		return nil, fmt.Errorf("执行命令失败: %w\n输出: %s", err, string(output))
	}
	
	return output, nil
}

// convertItems 转换文件列表
func (p *RcloneProvider) convertItems(items []RcloneItem) ([]*filemanager.FileInfo, error) {
	result := make([]*filemanager.FileInfo, len(items))
	
	for i, item := range items {
		result[i] = p.convertItem(item)
	}
	
	return result, nil
}

// convertItem 转换文件信息
func (p *RcloneProvider) convertItem(item RcloneItem) *filemanager.FileInfo {
	return &filemanager.FileInfo{
		Name:         item.Name,
		Path:         item.Path,
		Size:         item.Size,
		IsDir:        item.IsDir,
		ModifiedTime: item.ModTime,
		CreatedTime:  item.ModTime, // rclone只提供修改时间
		ContentType:  item.MimeType,
		Metadata: map[string]string{
			"hash": item.Hash,
		},
	}
}

// runUploadCommand 运行上传命令（带进度监控）
func (p *RcloneProvider) runUploadCommand(cmd *exec.Cmd, progress *filemanager.UploadProgress) error {
	return p.runProgressCommand(cmd, "upload", progress)
}

// runDownloadCommand 运行下载命令（带进度监控）
func (p *RcloneProvider) runDownloadCommand(cmd *exec.Cmd, progress *filemanager.DownloadProgress) error {
	return p.runProgressCommand(cmd, "download", progress)
}

// runProgressCommand 运行带进度监控的命令
func (p *RcloneProvider) runProgressCommand(cmd *exec.Cmd, transferType string, progress interface{}) error {
	// 添加进度参数
	cmd.Args = append(cmd.Args, "--progress")
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建输出管道失败: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}
	
	// 解析进度输出
	go p.parseProgressOutput(stdout, transferType, progress)
	
	return cmd.Wait()
}

// parseProgressOutput 解析进度输出
func (p *RcloneProvider) parseProgressOutput(stdout io.Reader, transferType string, progress interface{}) {
	scanner := bufio.NewScanner(stdout)
	
	for scanner.Scan() {
		line := scanner.Text()
		p.logger.Debug("rclone进度输出", zap.String("line", line))
		
		// 解析进度信息
		// rclone的进度输出格式比较复杂，这里简化处理
		if transferType == "upload" {
			if uploadProgress, ok := progress.(*filemanager.UploadProgress); ok {
				// 简单的进度解析逻辑
				if strings.Contains(line, "Transferred:") {
					// 这里可以解析更详细的进度信息
					uploadProgress.Percentage = float64(uploadProgress.UploadedBytes) / float64(uploadProgress.TotalBytes) * 100
				}
			}
		} else if transferType == "download" {
			if downloadProgress, ok := progress.(*filemanager.DownloadProgress); ok {
				// 简单的进度解析逻辑
				if strings.Contains(line, "Transferred:") {
					downloadProgress.Percentage = float64(downloadProgress.DownloadedBytes) / float64(downloadProgress.TotalBytes) * 100
				}
			}
		}
	}
}

// rcloneReadCloser rclone读取器封装
type rcloneReadCloser struct {
	pipe io.ReadCloser
	cmd   *exec.Cmd
}

func (r *rcloneReadCloser) Read(p []byte) (n int, err error) {
	return r.pipe.Read(p)
}

func (r *rcloneReadCloser) Close() error {
	r.pipe.Close()
	
	// 等待命令完成
	if r.cmd.Process != nil {
		r.cmd.Process.Kill()
	}
	
	return r.cmd.Wait()
}