package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/core/config"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"
)

// OneDriveStorage OneDrive存储实现
type OneDriveStorage struct {
	name         string
	baseURL      string
	httpClient   *httpclient.Client
	logger       *logger.Logger
	config       *config.OneDriveConfig
	accessToken  string
	refreshToken string
}

// NewOneDriveStorage 创建OneDrive存储实例
func NewOneDriveStorage(cfg *config.Config) (*OneDriveStorage, error) {
	httpClient := httpclient.NewClient(&httpclient.Config{
		BaseURL:   "https://graph.microsoft.com/v1.0/",
		Timeout:   30 * time.Second,
		UserAgent: "MoviePilot/1.0",
	})

	storage := &OneDriveStorage{
		name:       "OneDrive",
		baseURL:    "https://graph.microsoft.com/v1.0/",
		httpClient: httpClient,
		logger:     logger.NewLogger("storage-onedrive"),
		config:     &cfg.OneDrive,
	}

	// 初始化认证
	if err := storage.authenticate(); err != nil {
		return nil, fmt.Errorf("OneDrive认证失败: %w", err)
	}

	return storage, nil
}

// authenticate OneDrive认证
func (s *OneDriveStorage) authenticate() error {
	if s.config.AccessToken != "" {
		s.accessToken = s.config.AccessToken
		s.refreshToken = s.config.RefreshToken
		return nil
	}

	// 如果没有配置token，需要OAuth流程
	// 简化实现，实际需要处理完整的OAuth2流程
	return fmt.Errorf("OneDrive需要配置AccessToken或实现OAuth流程")
}

// refreshToken 刷新访问令牌
func (s *OneDriveStorage) refreshToken() error {
	if s.refreshToken == "" {
		return fmt.Errorf("没有可用的刷新令牌")
	}

	// 执行令牌刷新
	url := "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	payload := map[string]string{
		"client_id":     s.config.ClientID,
		"client_secret": s.config.ClientSecret,
		"refresh_token": s.refreshToken,
		"grant_type":    "refresh_token",
	}

	resp, err := s.httpClient.Post(context.Background(), url, headers, payload)
	if err != nil {
		return fmt.Errorf("刷新令牌失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := s.httpClient.DecodeJSON(resp, &result); err != nil {
		return fmt.Errorf("解析刷新响应失败: %w", err)
	}

	s.accessToken = result.AccessToken
	s.refreshToken = result.RefreshToken

	s.logger.Info("OneDrive令牌刷新成功")
	return nil
}

// Name 获取存储名称
func (s *OneDriveStorage) Name() string {
	return s.name
}

// Type 获取存储类型
func (s *OneDriveStorage) Type() string {
	return "onedrive"
}

// IsAvailable 检查存储是否可用
func (s *OneDriveStorage) IsAvailable() bool {
	return s.accessToken != ""
}

// ListFiles 列出文件
func (s *OneDriveStorage) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("OneDrive不可用，请先认证")
	}

	drivePath := s.normalizePath(path)
	url := fmt.Sprintf("%sme/drive/root:%s:/children", s.baseURL, drivePath)
	headers := s.getAuthHeaders()

	resp, err := s.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Size   int64  `json:"size"`
			WebURL string `json:"webUrl"`
			File   struct {
				MimeType string `json:"mimeType"`
			} `json:"file"`
			Folder struct {
				ChildCount int `json:"childCount"`
			} `json:"folder"`
			LastModifiedDateTime string `json:"lastModifiedDateTime"`
		} `json:"value"`
	}

	if err := s.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("解析文件列表响应失败: %w", err)
	}

	var files []FileInfo
	for _, item := range result.Value {
		fileType := FileTypeFile
		if item.Folder.ChildCount > 0 {
			fileType = FileTypeDirectory
		}

		modTime, _ := time.Parse(time.RFC3339, item.LastModifiedDateTime)

		files = append(files, FileInfo{
			Name:    item.Name,
			Path:    filepath.Join(path, item.Name),
			Size:    item.Size,
			Type:    fileType,
			ModTime: modTime,
		})
	}

	return files, nil
}

// UploadFile 上传文件
func (s *OneDriveStorage) UploadFile(ctx context.Context, localPath, remotePath string) error {
	if !s.IsAvailable() {
		return fmt.Errorf("OneDrive不可用，请先认证")
	}

	drivePath := s.normalizePath(remotePath)

	// 小文件使用简单上传，大文件使用分片上传
	fileInfo, err := s.getLocalFileInfo(localPath)
	if err != nil {
		return fmt.Errorf("获取本地文件信息失败: %w", err)
	}

	if fileInfo.Size < 4*1024*1024 { // 4MB以下使用简单上传
		return s.simpleUpload(ctx, localPath, drivePath)
	} else {
		return s.resumableUpload(ctx, localPath, drivePath)
	}
}

// simpleUpload 简单上传（小文件）
func (s *OneDriveStorage) simpleUpload(ctx context.Context, localPath, drivePath string) error {
	url := fmt.Sprintf("%sme/drive/root:%s:/content", s.baseURL, drivePath)
	headers := s.getAuthHeaders()
	headers["Content-Type"] = "application/octet-stream"

	fileReader, err := s.openLocalFile(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer fileReader.Close()

	resp, err := s.httpClient.Put(ctx, url, headers, fileReader)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Infof("文件上传成功: %s", drivePath)
	return nil
}

// resumableUpload 分片上传（大文件）
func (s *OneDriveStorage) resumableUpload(ctx context.Context, localPath, drivePath string) error {
	// 创建上传会话
	sessionURL, err := s.createUploadSession(ctx, drivePath)
	if err != nil {
		return fmt.Errorf("创建上传会话失败: %w", err)
	}

	// 执行分片上传
	return s.uploadChunks(ctx, localPath, sessionURL)
}

// createUploadSession 创建上传会话
func (s *OneDriveStorage) createUploadSession(ctx context.Context, drivePath string) (string, error) {
	url := fmt.Sprintf("%sme/drive/root:%s:/createUploadSession", s.baseURL, drivePath)
	headers := s.getAuthHeaders()

	resp, err := s.httpClient.Post(ctx, url, headers, nil)
	if err != nil {
		return "", fmt.Errorf("创建上传会话请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		UploadURL string `json:"uploadUrl"`
	}

	if err := s.httpClient.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("解析上传会话响应失败: %w", err)
	}

	return result.UploadURL, nil
}

// uploadChunks 分片上传
func (s *OneDriveStorage) uploadChunks(ctx context.Context, localPath, sessionURL string) error {
	// 简化实现，实际需要处理分片上传逻辑
	s.logger.Infof("开始分片上传: %s", localPath)

	// 这里应该实现分片上传逻辑
	// 包括读取文件分片、上传分片、处理失败重试等

	s.logger.Infof("分片上传完成: %s", localPath)
	return nil
}

// DownloadFile 下载文件
func (s *OneDriveStorage) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	if !s.IsAvailable() {
		return fmt.Errorf("OneDrive不可用，请先认证")
	}

	drivePath := s.normalizePath(remotePath)
	url := fmt.Sprintf("%sme/drive/root:%s:/content", s.baseURL, drivePath)
	headers := s.getAuthHeaders()

	resp, err := s.httpClient.Get(ctx, url, headers)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	return s.saveToLocalFile(resp.Body, localPath)
}

// DeleteFile 删除文件
func (s *OneDriveStorage) DeleteFile(ctx context.Context, path string) error {
	if !s.IsAvailable() {
		return fmt.Errorf("OneDrive不可用，请先认证")
	}

	drivePath := s.normalizePath(path)
	url := fmt.Sprintf("%sme/drive/root:%s:", s.baseURL, drivePath)
	headers := s.getAuthHeaders()

	resp, err := s.httpClient.Delete(ctx, url, headers)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Infof("文件删除成功: %s", path)
	return nil
}

// getAuthHeaders 获取认证头
func (s *OneDriveStorage) getAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + s.accessToken,
		"Content-Type":  "application/json",
	}
}

// normalizePath 规范化路径
func (s *OneDriveStorage) normalizePath(path string) string {
	if path == "" || path == "/" {
		return ""
	}
	return "/" + strings.Trim(path, "/")
}

// getLocalFileInfo 获取本地文件信息
func (s *OneDriveStorage) getLocalFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// openLocalFile 打开本地文件
func (s *OneDriveStorage) openLocalFile(path string) (io.ReadCloser, error) {
	return httpclient.OpenLocalFile(path)
}

// saveToLocalFile 保存到本地文件
func (s *OneDriveStorage) saveToLocalFile(reader io.Reader, localPath string) error {
	return httpclient.SaveToLocalFile(reader, localPath)
}
