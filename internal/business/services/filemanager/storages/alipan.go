package storages

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/integration/filemanager"
)

// AliPanProvider 阿里云盘存储提供商
type AliPanProvider struct {
	config  *AliPanConfig
	client  *http.Client
	logger  *zap.Logger
}

// AliPanConfig 阿里云盘配置
type AliPanConfig struct {
	AppID       string `json:"app_id"`
	AppSecret   string `json:"app_secret"`
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserAgent   string `json:"user_agent"`
	Timeout     int    `json:"timeout"`
	RootPath    string `json:"root_path"`
}

// AliPanTokenInfo 阿里云盘令牌信息
type AliPanTokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// AliPanFileInfo 阿里云盘文件信息
type AliPanFileInfo struct {
	FileID      string    `json:"file_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	ParentID    string    `json:"parent_file_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ContentHash string    `json:"content_hash"`
	Url         string    `json:"url,omitempty"`
}

// AliPanUploadResult 上传结果
type AliPanUploadResult struct {
	FileID     string `json:"file_id"`
	UploadID   string `json:"upload_id"`
	UploadURL  string `json:"upload_url"`
	StatusCode int    `json:"status_code"`
}

// NewAliPanProvider 创建阿里云盘提供商
func NewAliPanProvider(logger *zap.Logger) filemanager.StorageProvider {
	return &AliPanProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetName 获取存储名称
func (p *AliPanProvider) GetName() string {
	return "AliPan"
}

// GetType 获取存储类型
func (p *AliPanProvider) GetType() string {
	return "alipan"
}

// ValidateConfig 验证配置
func (p *AliPanProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg AliPanConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.AppID == "" {
		return fmt.Errorf("AppID不能为空")
	}
	
	if cfg.AppSecret == "" {
		return fmt.Errorf("AppSecret不能为空")
	}
	
	if cfg.AccessToken == "" && cfg.RefreshToken == "" {
		return fmt.Errorf("必须提供AccessToken或RefreshToken")
	}
	
	return nil
}

// Initialize 初始化存储
func (p *AliPanProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &AliPanConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析阿里云盘配置失败: %w", err)
	}
	
	// 验证配置
	if err := p.ValidateConfig(config); err != nil {
		return err
	}
	
	// 设置默认值
	if p.config.UserAgent == "" {
		p.config.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	
	if p.config.Timeout > 0 {
		p.client.Timeout = time.Duration(p.config.Timeout) * time.Second
	}
	
	// 设置默认根路径
	if p.config.RootPath == "" {
		p.config.RootPath = "/"
	}
	
	// 测试连接和令牌
	if err := p.testConnection(); err != nil {
		return fmt.Errorf("阿里云盘连接测试失败: %w", err)
	}
	
	p.logger.Info("阿里云盘存储提供商初始化成功")
	return nil
}

// ListFiles 列出文件
func (p *AliPanProvider) ListFiles(ctx context.Context, path string) ([]*filemanager.FileInfo, error) {
	parentID, err := p.pathToParentID(path)
	if err != nil {
		return nil, err
	}
	
	files, err := p.listFiles(ctx, parentID)
	if err != nil {
		return nil, err
	}
	
	return p.convertFiles(files)
}

// UploadFile 上传文件
func (p *AliPanProvider) UploadFile(ctx context.Context, localPath, remotePath string, progress *filemanager.UploadProgress) error {
	// 检查本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer file.Close()
	
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	
	return p.UploadStream(ctx, file, remotePath, fileInfo.Size(), progress)
}

// UploadStream 上传流
func (p *AliPanProvider) UploadStream(ctx context.Context, stream io.Reader, remotePath string, size int64, progress *filemanager.UploadProgress) error {
	// 获取父目录ID
	dirPath := filepath.Dir(remotePath)
	parentID, err := p.pathToParentID(dirPath)
	if err != nil {
		return err
	}
	
	fileName := filepath.Base(remotePath)
	
	// 创建上传
	uploadResult, err := p.createUpload(ctx, parentID, fileName, size)
	if err != nil {
		return err
	}
	
	// 分片上传
	if size > 10*1024*1024 { // 大于10MB使用分片上传
		return p.uploadLargeFile(ctx, stream, uploadResult, size, progress)
	} else {
		return p.uploadSmallFile(ctx, stream, uploadResult, size, progress)
	}
}

// DownloadFile 下载文件
func (p *AliPanProvider) DownloadFile(ctx context.Context, remotePath, localPath string, progress *filemanager.DownloadProgress) error {
	// 创建本地目录
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}
	
	// 创建本地文件
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer localFile.Close()
	
	// 获取下载流
	stream, err := p.DownloadStream(ctx, remotePath)
	if err != nil {
		return err
	}
	defer stream.Close()
	
	// 复制数据并更新进度
	if progress != nil {
		progress.TotalBytes = progress.TotalBytes
	}
	
	written, err := io.Copy(localFile, stream)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	
	// 更新进度
	if progress != nil {
		progress.DownloadedBytes = written
		progress.Percentage = 100.0
	}
	
	return nil
}

// DownloadStream 下载流
func (p *AliPanProvider) DownloadStream(ctx context.Context, remotePath string) (io.ReadCloser, error) {
	fileID, err := p.pathToFileID(remotePath)
	if err != nil {
		return nil, err
	}
	
	downloadURL, err := p.getDownloadURL(ctx, fileID)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建下载请求失败: %w", err)
	}
	
	req.Header.Set("User-Agent", p.config.UserAgent)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}
	
	return resp.Body, nil
}

// DeleteFile 删除文件
func (p *AliPanProvider) DeleteFile(ctx context.Context, path string) error {
	fileID, err := p.pathToFileID(path)
	if err != nil {
		return err
	}
	
	return p.deleteFile(ctx, fileID)
}

// CreateDirectory 创建目录
func (p *AliPanProvider) CreateDirectory(ctx context.Context, path string) error {
	parentID, err := p.pathToParentID(filepath.Dir(path))
	if err != nil {
		return err
	}
	
	dirName := filepath.Base(path)
	
	return p.createDirectory(ctx, parentID, dirName)
}

// DeleteDirectory 删除目录
func (p *AliPanProvider) DeleteDirectory(ctx context.Context, path string, recursive bool) error {
	fileID, err := p.pathToFileID(path)
	if err != nil {
		return err
	}
	
	return p.deleteDirectory(ctx, fileID, recursive)
}

// RenameFile 重命名文件
func (p *AliPanProvider) RenameFile(ctx context.Context, oldPath, newPath string) error {
	fileID, err := p.pathToFileID(oldPath)
	if err != nil {
		return err
	}
	
	newName := filepath.Base(newPath)
	newParentID, err := p.pathToParentID(filepath.Dir(newPath))
	if err != nil {
		return err
	}
	
	return p.renameFile(ctx, fileID, newName, newParentID)
}

// CopyFile 复制文件
func (p *AliPanProvider) CopyFile(ctx context.Context, sourcePath, targetPath string) error {
	sourceFileID, err := p.pathToFileID(sourcePath)
	if err != nil {
		return err
	}
	
	targetParentID, err := p.pathToParentID(filepath.Dir(targetPath))
	if err != nil {
		return err
	}
	
	targetName := filepath.Base(targetPath)
	
	return p.copyFile(ctx, sourceFileID, targetParentID, targetName)
}

// MoveFile 移动文件
func (p *AliPanProvider) MoveFile(ctx context.Context, sourcePath, targetPath string) error {
	// 先复制后删除
	if err := p.CopyFile(ctx, sourcePath, targetPath); err != nil {
		return err
	}
	
	return p.DeleteFile(ctx, sourcePath)
}

// GetFileInfo 获取文件信息
func (p *AliPanProvider) GetFileInfo(ctx context.Context, path string) (*filemanager.FileInfo, error) {
	fileID, err := p.pathToFileID(path)
	if err != nil {
		return nil, err
	}
	
	fileInfo, err := p.getFileInfo(ctx, fileID)
	if err != nil {
		return nil, err
	}
	
	return p.convertFileInfo(fileInfo), nil
}

// GetURL 获取文件访问URL
func (p *AliPanProvider) GetURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	fileID, err := p.pathToFileID(path)
	if err != nil {
		return "", err
	}
	
	return p.getDownloadURL(ctx, fileID)
}

// IsHealthy 健康检查
func (p *AliPanProvider) IsHealthy(ctx context.Context) bool {
	// 测试获取根目录文件列表
	_, err := p.listFiles(ctx, "root")
	if err != nil {
		p.logger.Debug("阿里云盘健康检查失败", zap.Error(err))
		return false
	}
	
	return true
}

// GetConfig 获取配置
func (p *AliPanProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if appSecret, exists := result["app_secret"]; exists {
		if str, ok := appSecret.(string); ok && len(str) > 10 {
			result["app_secret"] = str[:10] + "***"
		}
	}
	
	if accessToken, exists := result["access_token"]; exists {
		if str, ok := accessToken.(string); ok && len(str) > 20 {
			result["access_token"] = str[:20] + "***"
		}
	}
	
	if refreshToken, exists := result["refresh_token"]; exists {
		if str, ok := refreshToken.(string); ok && len(str) > 20 {
			result["refresh_token"] = str[:20] + "***"
		}
	}
	
	return result
}

// Close 关闭存储
func (p *AliPanProvider) Close() error {
	return nil
}

// testConnection 测试连接
func (p *AliPanProvider) testConnection() error {
	// 尝试刷新令牌
	if err := p.refreshAccessToken(); err != nil {
		return err
	}
	
	// 测试获取根目录
	_, err := p.listFiles(context.Background(), "root")
	return err
}

// refreshAccessToken 刷新访问令牌
func (p *AliPanProvider) refreshAccessToken() error {
	// 实现令牌刷新逻辑
	// 这里需要调用阿里云盘的令牌刷新API
	return nil
}

// pathToParentID 将路径转换为父目录ID
func (p *AliPanProvider) pathToParentID(path string) (string, error) {
	if path == "/" || path == p.config.RootPath {
		return "root", nil
	}
	
	// 实现路径到目录ID的转换逻辑
	// 这里需要递归查找或缓存路径映射
	return "root", nil
}

// pathToFileID 将路径转换为文件ID
func (p *AliPanProvider) pathToFileID(path string) (string, error) {
	// 实现路径到文件ID的转换逻辑
	// 这里需要递归查找或缓存路径映射
	return "", fmt.Errorf("文件不存在: %s", path)
}

// convertFiles 转换文件列表
func (p *AliPanProvider) convertFiles(files []AliPanFileInfo) ([]*filemanager.FileInfo, error) {
	result := make([]*filemanager.FileInfo, len(files))
	
	for i, file := range files {
		result[i] = p.convertFileInfo(&file)
	}
	
	return result, nil
}

// convertFileInfo 转换文件信息
func (p *AliPanProvider) convertFileInfo(file *AliPanFileInfo) *filemanager.FileInfo {
	return &filemanager.FileInfo{
		Name:         file.Name,
		Path:         "", // 需要构建完整路径
		Size:         file.Size,
		IsDir:        file.Type == "folder",
		ModifiedTime: file.UpdatedAt,
		CreatedTime:  file.CreatedAt,
		ContentType:  "", // 根据文件扩展名推断
		Metadata: map[string]string{
			"file_id":      file.FileID,
			"content_hash": file.ContentHash,
		},
	}
}

// 以下是具体API方法的占位符实现
// 在实际实现中，这些方法需要调用阿里云盘的API

func (p *AliPanProvider) listFiles(ctx context.Context, parentID string) ([]AliPanFileInfo, error) {
	// 实现获取文件列表API调用
	return []AliPanFileInfo{}, nil
}

func (p *AliPanProvider) createUpload(ctx context.Context, parentID, fileName string, size int64) (*AliPanUploadResult, error) {
	// 实现创建上传API调用
	return &AliPanUploadResult{}, nil
}

func (p *AliPanProvider) uploadSmallFile(ctx context.Context, stream io.Reader, result *AliPanUploadResult, size int64, progress *filemanager.UploadProgress) error {
	// 实现小文件上传
	return nil
}

func (p *AliPanProvider) uploadLargeFile(ctx context.Context, stream io.Reader, result *AliPanUploadResult, size int64, progress *filemanager.UploadProgress) error {
	// 实现大文件分片上传
	return nil
}

func (p *AliPanProvider) getDownloadURL(ctx context.Context, fileID string) (string, error) {
	// 实现获取下载URL API调用
	return "", nil
}

func (p *AliPanProvider) deleteFile(ctx context.Context, fileID string) error {
	// 实现删除文件API调用
	return nil
}

func (p *AliPanProvider) createDirectory(ctx context.Context, parentID, dirName string) error {
	// 实现创建目录API调用
	return nil
}

func (p *AliPanProvider) deleteDirectory(ctx context.Context, fileID string, recursive bool) error {
	// 实现删除目录API调用
	return nil
}

func (p *AliPanProvider) renameFile(ctx context.Context, fileID, newName, newParentID string) error {
	// 实现重命名文件API调用
	return nil
}

func (p *AliPanProvider) copyFile(ctx context.Context, sourceFileID, targetParentID, targetName string) error {
	// 实现复制文件API调用
	return nil
}

func (p *AliPanProvider) getFileInfo(ctx context.Context, fileID string) (*AliPanFileInfo, error) {
	// 实现获取文件信息API调用
	return &AliPanFileInfo{}, nil
}