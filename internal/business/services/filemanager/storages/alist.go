package storages

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/integration/filemanager"
)

// AListProvider AList文件列表服务提供商
type AListProvider struct {
	config  *AListConfig
	client  *http.Client
	logger  *zap.Logger
}

// AListConfig AList配置
type AListConfig struct {
	ServerURL    string `json:"server_url"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Token        string `json:"token"`
	RootPath     string `json:"root_path"`
	Timeout      int    `json:"timeout"`
	Insecure     bool   `json:"insecure"`
	AutoRefresh  bool   `json:"auto_refresh"`
}

// AListResponse AList API响应
type AListResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// AListLoginResponse 登录响应
type AListLoginResponse struct {
	Token string `json:"token"`
}

// AListDirResponse 目录响应
type AListDirResponse struct {
	Content  []AListFile `json:"content"`
	Total    int         `json:"total"`
	Readme   string      `json:"readme"`
	Write    bool        `json:"write"`
	Provider string      `json:"provider"`
}

// AListFile 文件信息
type AListFile struct {
	Name         string            `json:"name"`
	Size         int64             `json:"size"`
	IsDir        bool              `json:"is_dir"`
	ModifiedTime string            `json:"modified_time"`
	CreatedTime  string            `json:"created_time"`
	Sign         string            `json:"sign"`
	Thumb        string            `json:"thumb"`
	Type         int               `json:"type"`
	HashInfo     string            `json:"hashinfo"`
	Hash         string            `json:"hash"`
	FsID         string            `json:"fs_id"`
	Driver       string            `json:"driver"`
	FilePath     string            `json:"raw_url"`
	Repository   string            `json:"repository"`
	InternalType string            `json:"internal_type"`
	Additional   map[string]string `json:"additional"`
}

// AListUploadResponse 上传响应
type AListUploadResponse struct {
	TaskID string `json:"task_id"`
	FileID string `json:"file_id"`
}

// AListDownloadURLResponse 下载URL响应
type AListDownloadURLResponse struct {
	DownURL      string `json:"raw_url"`
	ContentURL   string `json:"content_url"`
	Expire       int    `json:"expire"`
	AuthURL      string `json:"auth_url"`
	DownloadURL  string `json:"download_url"`
}

// NewAListProvider 创建AList提供商
func NewAListProvider(logger *zap.Logger) filemanager.StorageProvider {
	return &AListProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetName 获取存储名称
func (p *AListProvider) GetName() string {
	return "AList"
}

// GetType 获取存储类型
func (p *AListProvider) GetType() string {
	return "alist"
}

// ValidateConfig 验证配置
func (p *AListProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg AListConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.ServerURL == "" {
		return fmt.Errorf("ServerURL不能为空")
	}
	
	// 验证URL格式
	if !strings.HasPrefix(cfg.ServerURL, "http://") && !strings.HasPrefix(cfg.ServerURL, "https://") {
		return fmt.Errorf("ServerURL必须是有效的HTTP/HTTPS URL")
	}
	
	return nil
}

// Initialize 初始化存储
func (p *AListProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &AListConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析AList配置失败: %w", err)
	}
	
	// 验证配置
	if err := p.ValidateConfig(config); err != nil {
		return err
	}
	
	// 确保ServerURL以/结尾
	if !strings.HasSuffix(p.config.ServerURL, "/") {
		p.config.ServerURL += "/"
	}
	
	// 设置超时
	if p.config.Timeout > 0 {
		p.client.Timeout = time.Duration(p.config.Timeout) * time.Second
	}
	
	// 跳过HTTPS证书验证
	if p.config.Insecure {
		// 实现跳过TLS验证的Transport
	}
	
	// 登录获取token
	if p.config.Token == "" {
		if err := p.login(); err != nil {
			return fmt.Errorf("AList登录失败: %w", err)
		}
	}
	
	// 测试连接
	if err := p.testConnection(); err != nil {
		return fmt.Errorf("AList连接测试失败: %w", err)
	}
	
	p.logger.Info("AList存储提供商初始化成功")
	return nil
}

// ListFiles 列出文件
func (p *AListProvider) ListFiles(ctx context.Context, path string) ([]*filemanager.FileInfo, error) {
	apiURL := p.config.ServerURL + "api/fs/list"
	
	data := map[string]interface{}{
		"path":     path,
		"password": "",
		"page":     1,
		"per_page": 0, // 0表示获取所有
		"refresh":  false,
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return nil, err
	}
	
	var listResp AListDirResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, fmt.Errorf("解析AList响应失败: %w", err)
	}
	
	return p.convertFiles(listResp.Content)
}

// UploadFile 上传文件
func (p *AListProvider) UploadFile(ctx context.Context, localPath, remotePath string, progress *filemanager.UploadProgress) error {
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
func (p *AListProvider) UploadStream(ctx context.Context, stream io.Reader, remotePath string, size int64, progress *filemanager.UploadProgress) error {
	// 创建上传任务
	apiURL := p.config.ServerURL + "api/fs/put"
	
	data := map[string]interface{}{
		"path":         remotePath,
		"password":     "",
		"as_task":      false,
	}
	
	// 这里需要实现multipart/form-data上传
	// 由于Go标准库的multipart/form-data比较复杂，这里简化处理
	return fmt.Errorf("流上传需要实现multipart/form-data支持")
}

// DownloadFile 下载文件
func (p *AListProvider) DownloadFile(ctx context.Context, remotePath, localPath string, progress *filemanager.DownloadProgress) error {
	// 创建本地目录
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}
	
	// 获取下载URL
	downloadURL, err := p.GetURL(ctx, remotePath, 0)
	if err != nil {
		return err
	}
	
	// 创建本地文件
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer localFile.Close()
	
	// 下载文件
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("创建下载请求失败: %w", err)
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}
	
	// 复制数据并更新进度
	if progress != nil {
		progress.TotalBytes = resp.ContentLength
	}
	
	written, err := io.Copy(localFile, resp.Body)
	if err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	
	// 更新进度
	if progress != nil {
		progress.DownloadedBytes = written
		if progress.TotalBytes > 0 {
			progress.Percentage = float64(written) / float64(progress.TotalBytes) * 100
		}
	}
	
	return nil
}

// DownloadStream 下载流
func (p *AListProvider) DownloadStream(ctx context.Context, remotePath string) (io.ReadCloser, error) {
	// 获取下载URL
	downloadURL, err := p.GetURL(ctx, remotePath, 0)
	if err != nil {
		return nil, err
	}
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建下载请求失败: %w", err)
	}
	
	// 发送请求
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
func (p *AListProvider) DeleteFile(ctx context.Context, path string) error {
	apiURL := p.config.ServerURL + "api/fs/remove"
	
	data := map[string]interface{}{
		"dir":     filepath.Dir(path),
		"names":   []string{filepath.Base(path)},
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return err
	}
	
	var result struct {
		Failed []string `json:"failed"`
		Success int     `json:"success"`
	}
	
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("解析删除响应失败: %w", err)
	}
	
	if result.Success == 0 && len(result.Failed) > 0 {
		return fmt.Errorf("删除失败: %v", result.Failed)
	}
	
	return nil
}

// CreateDirectory 创建目录
func (p *AListProvider) CreateDirectory(ctx context.Context, path string) error {
	apiURL := p.config.ServerURL + "api/fs/mkdir"
	
	data := map[string]interface{}{
		"path": path,
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return err
	}
	
	// 检查响应
	var result struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("解析创建目录响应失败: %w", err)
	}
	
	if result.Code != 200 {
		return fmt.Errorf("创建目录失败")
	}
	
	return nil
}

// DeleteDirectory 删除目录
func (p *AListProvider) DeleteDirectory(ctx context.Context, path string, recursive bool) error {
	apiURL := p.config.ServerURL + "api/fs/remove"
	
	data := map[string]interface{}{
		"dir":  path,
		"names": []string{filepath.Base(path)},
	}
	
	if recursive {
		data["force"] = true
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return err
	}
	
	var result struct {
		Failed []string `json:"failed"`
		Success int     `json:"success"`
	}
	
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("解析删除目录响应失败: %w", err)
	}
	
	if result.Success == 0 && len(result.Failed) > 0 {
		return fmt.Errorf("删除目录失败: %v", result.Failed)
	}
	
	return nil
}

// RenameFile 重命名文件
func (p *AListProvider) RenameFile(ctx context.Context, oldPath, newPath string) error {
	apiURL := p.config.ServerURL + "api/fs/rename"
	
	data := map[string]interface{}{
		"path":     oldPath,
		"new_name": filepath.Base(newPath),
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return err
	}
	
	var result struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("解析重命名响应失败: %w", err)
	}
	
	if result.Code != 200 {
		return fmt.Errorf("重命名失败")
	}
	
	return nil
}

// CopyFile 复制文件
func (p *AListProvider) CopyFile(ctx context.Context, sourcePath, targetPath string) error {
	apiURL := p.config.ServerURL + "api/fs/copy"
	
	data := map[string]interface{}{
		"src_dir":    filepath.Dir(sourcePath),
		"dst_dir":    filepath.Dir(targetPath),
		"src_file_names": []string{filepath.Base(sourcePath)},
		"dst_file_names": []string{filepath.Base(targetPath)},
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return err
	}
	
	var result struct {
		Failed []string `json:"failed"`
		Success int     `json:"success"`
	}
	
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("解析复制响应失败: %w", err)
	}
	
	if result.Success == 0 && len(result.Failed) > 0 {
		return fmt.Errorf("复制失败: %v", result.Failed)
	}
	
	return nil
}

// MoveFile 移动文件
func (p *AListProvider) MoveFile(ctx context.Context, sourcePath, targetPath string) error {
	apiURL := p.config.ServerURL + "api/fs/move"
	
	data := map[string]interface{}{
		"src_dir":    filepath.Dir(sourcePath),
		"dst_dir":    filepath.Dir(targetPath),
		"src_file_names": []string{filepath.Base(sourcePath)},
		"dst_file_names": []string{filepath.Base(targetPath)},
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return err
	}
	
	var result struct {
		Failed []string `json:"failed"`
		Success int     `json:"success"`
	}
	
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("解析移动响应失败: %w", err)
	}
	
	if result.Success == 0 && len(result.Failed) > 0 {
		return fmt.Errorf("移动失败: %v", result.Failed)
	}
	
	return nil
}

// GetFileInfo 获取文件信息
func (p *AListProvider) GetFileInfo(ctx context.Context, path string) (*filemanager.FileInfo, error) {
	apiURL := p.config.ServerURL + "api/fs/get"
	
	data := map[string]interface{}{
		"path": path,
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return nil, err
	}
	
	var alistFile AListFile
	if err := json.Unmarshal(resp, &alistFile); err != nil {
		return nil, fmt.Errorf("解析文件信息响应失败: %w", err)
	}
	
	return p.convertFileInfo(alistFile), nil
}

// GetURL 获取文件访问URL
func (p *AListProvider) GetURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	apiURL := p.config.ServerURL + "api/fs/get"
	
	data := map[string]interface{}{
		"path": path,
	}
	
	resp, err := p.postAPI(ctx, apiURL, data)
	if err != nil {
		return "", err
	}
	
	var urlResp AListDownloadURLResponse
	if err := json.Unmarshal(resp, &urlResp); err != nil {
		return "", fmt.Errorf("解析下载URL响应失败: %w", err)
	}
	
	return urlResp.DownURL, nil
}

// IsHealthy 健康检查
func (p *AListProvider) IsHealthy(ctx context.Context) bool {
	// 测试获取根目录文件列表
	_, err := p.ListFiles(ctx, "/")
	if err != nil {
		p.logger.Debug("AList健康检查失败", zap.Error(err))
		return false
	}
	
	return true
}

// GetConfig 获取配置
func (p *AListProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if password, exists := result["password"]; exists {
		if str, ok := password.(string); ok && len(str) > 5 {
			result["password"] = str[:5] + "***"
		}
	}
	
	if token, exists := result["token"]; exists {
		if str, ok := token.(string); ok && len(str) > 20 {
			result["token"] = str[:20] + "***"
		}
	}
	
	return result
}

// Close 关闭存储
func (p *AListProvider) Close() error {
	return nil
}

// login 登录AList
func (p *AListProvider) login() error {
	apiURL := p.config.ServerURL + "api/auth/login"
	
	data := map[string]interface{}{
		"username": p.config.Username,
		"password": p.config.Password,
	}
	
	resp, err := p.postAPI(context.Background(), apiURL, data)
	if err != nil {
		return err
	}
	
	var loginResp AListLoginResponse
	if err := json.Unmarshal(resp, &loginResp); err != nil {
		return fmt.Errorf("解析登录响应失败: %w", err)
	}
	
	p.config.Token = loginResp.Token
	return nil
}

// testConnection 测试连接
func (p *AListProvider) testConnection() error {
	// 测试获取根目录
	_, err := p.ListFiles(context.Background(), "/")
	return err
}

// postAPI 发送API请求
func (p *AListProvider) postAPI(ctx context.Context, apiURL string, data map[string]interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("编码请求数据失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if p.config.Token != "" {
		req.Header.Set("Authorization", p.config.Token)
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	
	var alistResp AListResponse
	if err := json.Unmarshal(body, &alistResp); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}
	
	if alistResp.Code != 200 {
		return nil, fmt.Errorf("AList API错误: %s", alistResp.Message)
	}
	
	if alistResp.Data != nil {
		return json.Marshal(alistResp.Data)
	}
	
	return body, nil
}

// convertFiles 转换文件列表
func (p *AListProvider) convertFiles(files []AListFile) ([]*filemanager.FileInfo, error) {
	result := make([]*filemanager.FileInfo, len(files))
	
	for i, file := range files {
		result[i] = p.convertFileInfo(file)
	}
	
	return result, nil
}

// convertFileInfo 转换文件信息
func (p *AListProvider) convertFileInfo(file AListFile) *filemanager.FileInfo {
	modifiedTime := time.Time{}
	createdTime := time.Time{}
	
	if file.ModifiedTime != "" {
		if mt, err := time.Parse(time.RFC3339, file.ModifiedTime); err == nil {
			modifiedTime = mt
		}
	}
	
	if file.CreatedTime != "" {
		if ct, err := time.Parse(time.RFC3339, file.CreatedTime); err == nil {
			createdTime = ct
		}
	}
	
	return &filemanager.FileInfo{
		Name:         file.Name,
		Path:         file.FilePath,
		Size:         file.Size,
		IsDir:        file.IsDir,
		ModifiedTime: modifiedTime,
		CreatedTime:  createdTime,
		ContentType:  file.InternalType,
		Metadata: map[string]string{
			"driver":  file.Driver,
			"sign":    file.Sign,
			"thumb":   file.Thumb,
			"fs_id":   file.FsID,
		},
	}
}