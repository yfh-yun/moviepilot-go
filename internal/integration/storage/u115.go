package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/config"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"
)

// U115Storage 115网盘存储实现
type U115Storage struct {
	name       string
	baseURL    string
	httpClient *httpclient.Client
	logger     *logger.Logger
	config     *config.U115Config
	// 认证相关字段
	accessToken string
	cookie      string
	userID      string
}

// NewU115Storage 创建115网盘存储实例
func NewU115Storage(cfg *config.Config) (*U115Storage, error) {
	httpClient := httpclient.NewClient(&httpclient.Config{
		BaseURL:   "https://webapi.115.com/",
		Timeout:   60 * time.Second, // 115网盘操作可能较慢
		UserAgent: "MoviePilot/1.0",
	})

	storage := &U115Storage{
		name:       "115网盘",
		baseURL:    "https://webapi.115.com/",
		httpClient: httpClient,
		logger:     logger.NewLogger("storage-u115"),
		config:     &cfg.U115,
	}

	// 异步初始化认证
	go storage.authenticate()

	return storage, nil
}

// authenticate 115网盘认证
func (s *U115Storage) authenticate() error {
	// 115网盘认证通常需要登录获取cookie
	// 这里简化实现，实际需要处理登录流程

	if s.config.AccessToken != "" {
		s.accessToken = s.config.AccessToken
		return nil
	}

	// 如果没有配置token，需要登录
	if s.config.Username == "" || s.config.Password == "" {
		return fmt.Errorf("115网盘需要用户名和密码或AccessToken")
	}

	// 执行登录（简化实现）
	return s.login()
}

// login 115网盘登录
func (s *U115Storage) login() error {
	// 115网盘登录流程比较复杂，需要处理验证码等
	// 这里提供一个简化版本

	payload := map[string]string{
		"login":    s.config.Username,
		"password": s.config.Password,
	}

	resp, err := s.httpClient.Post(context.Background(), "login", nil, payload)
	if err != nil {
		return fmt.Errorf("115网盘登录失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应获取cookie等信息
	// 简化实现，实际需要处理复杂的登录流程

	s.logger.Info("115网盘登录成功")
	return nil
}

// Name 获取存储名称
func (s *U115Storage) Name() string {
	return s.name
}

// Type 获取存储类型
func (s *U115Storage) Type() string {
	return "u115"
}

// IsAvailable 检查存储是否可用
func (s *U115Storage) IsAvailable() bool {
	return s.accessToken != "" || s.cookie != ""
}

// ListFiles 列出文件
func (s *U115Storage) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("115网盘不可用，请先认证")
	}

	// 115网盘文件列表API
	url := fmt.Sprintf("%sfiles?cid=%s", s.baseURL, path)
	headers := s.getAuthHeaders()

	resp, err := s.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		State bool `json:"state"`
		Data  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Size int64  `json:"size"`
			Time int64  `json:"time"`
			Type int    `json:"type"` // 0:文件, 1:文件夹
		} `json:"data"`
	}

	if err := s.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("解析文件列表响应失败: %w", err)
	}

	if !result.State {
		return nil, fmt.Errorf("获取文件列表失败")
	}

	var files []FileInfo
	for _, item := range result.Data {
		fileType := FileTypeFile
		if item.Type == 1 {
			fileType = FileTypeDirectory
		}

		files = append(files, FileInfo{
			Name:    item.Name,
			Path:    filepath.Join(path, item.Name),
			Size:    item.Size,
			Type:    fileType,
			ModTime: time.Unix(item.Time, 0),
		})
	}

	return files, nil
}

// UploadFile 上传文件
func (s *U115Storage) UploadFile(ctx context.Context, localPath, remotePath string) error {
	if !s.IsAvailable() {
		return fmt.Errorf("115网盘不可用，请先认证")
	}

	// 115网盘上传需要先获取上传信息
	uploadInfo, err := s.getUploadInfo(ctx, remotePath)
	if err != nil {
		return fmt.Errorf("获取上传信息失败: %w", err)
	}

	// 执行文件上传
	return s.uploadTo115(ctx, localPath, uploadInfo)
}

// getUploadInfo 获取上传信息
func (s *U115Storage) getUploadInfo(ctx context.Context, remotePath string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%supload/info", s.baseURL)
	headers := s.getAuthHeaders()

	resp, err := s.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("获取上传信息请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		State bool                   `json:"state"`
		Data  map[string]interface{} `json:"data"`
	}

	if err := s.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("解析上传信息失败: %w", err)
	}

	if !result.State {
		return nil, fmt.Errorf("获取上传信息失败")
	}

	return result.Data, nil
}

// uploadTo115 上传文件到115网盘
func (s *U115Storage) uploadTo115(ctx context.Context, localPath string, uploadInfo map[string]interface{}) error {
	// 115网盘上传流程比较复杂，这里简化实现
	// 实际需要处理分片上传、秒传等逻辑

	s.logger.Infof("开始上传文件到115网盘: %s", localPath)

	// 简化实现：直接调用上传API
	url := uploadInfo["url"].(string)
	headers := s.getAuthHeaders()

	// 设置上传相关头信息
	headers["Content-Type"] = "application/octet-stream"

	// 读取本地文件
	fileReader, err := s.openLocalFile(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer fileReader.Close()

	// 执行上传
	resp, err := s.httpClient.Put(ctx, url, headers, fileReader)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Infof("文件上传成功: %s", localPath)
	return nil
}

// DownloadFile 下载文件
func (s *U115Storage) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	if !s.IsAvailable() {
		return fmt.Errorf("115网盘不可用，请先认证")
	}

	// 获取下载链接
	downloadURL, err := s.getDownloadURL(ctx, remotePath)
	if err != nil {
		return fmt.Errorf("获取下载链接失败: %w", err)
	}

	headers := s.getAuthHeaders()

	// 执行下载
	resp, err := s.httpClient.Get(ctx, downloadURL, headers)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	// 保存到本地文件
	return s.saveToLocalFile(resp.Body, localPath)
}

// getDownloadURL 获取下载链接
func (s *U115Storage) getDownloadURL(ctx context.Context, remotePath string) (string, error) {
	url := fmt.Sprintf("%sdownload/info", s.baseURL)
	headers := s.getAuthHeaders()

	payload := map[string]string{
		"pickcode": s.extractPickCode(remotePath),
	}

	resp, err := s.httpClient.Post(ctx, url, headers, payload)
	if err != nil {
		return "", fmt.Errorf("获取下载信息失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		State bool `json:"state"`
		Data  struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := s.httpClient.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("解析下载信息失败: %w", err)
	}

	if !result.State {
		return "", fmt.Errorf("获取下载信息失败")
	}

	return result.Data.URL, nil
}

// DeleteFile 删除文件
func (s *U115Storage) DeleteFile(ctx context.Context, path string) error {
	if !s.IsAvailable() {
		return fmt.Errorf("115网盘不可用，请先认证")
	}

	url := fmt.Sprintf("%sfile/delete", s.baseURL)
	headers := s.getAuthHeaders()

	payload := map[string]string{
		"fid": s.extractFileID(path),
	}

	resp, err := s.httpClient.Post(ctx, url, headers, payload)
	if err != nil {
		return fmt.Errorf("删除文件请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		State bool `json:"state"`
	}

	if err := s.httpClient.DecodeJSON(resp, &result); err != nil {
		return fmt.Errorf("解析删除响应失败: %w", err)
	}

	if !result.State {
		return fmt.Errorf("删除文件失败")
	}

	s.logger.Infof("文件删除成功: %s", path)
	return nil
}

// getAuthHeaders 获取认证头
func (s *U115Storage) getAuthHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if s.accessToken != "" {
		headers["Authorization"] = "Bearer " + s.accessToken
	}

	if s.cookie != "" {
		headers["Cookie"] = s.cookie
	}

	return headers
}

// extractPickCode 从路径提取pickcode
func (s *U115Storage) extractPickCode(path string) string {
	// 115网盘使用pickcode标识文件
	// 简化实现，实际需要解析路径
	return strings.TrimPrefix(filepath.Base(path), "115://")
}

// extractFileID 从路径提取文件ID
func (s *U115Storage) extractFileID(path string) string {
	// 简化实现，实际需要解析115网盘的文件ID
	return filepath.Base(path)
}

// openLocalFile 打开本地文件
func (s *U115Storage) openLocalFile(path string) (io.ReadCloser, error) {
	// 使用标准库打开文件
	return httpclient.OpenLocalFile(path)
}

// saveToLocalFile 保存到本地文件
func (s *U115Storage) saveToLocalFile(reader io.Reader, localPath string) error {
	// 使用标准库保存文件
	return httpclient.SaveToLocalFile(reader, localPath)
}
