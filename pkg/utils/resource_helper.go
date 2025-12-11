package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// ResourceHelper 资源包管理助手
type ResourceHelper struct {
	logger        *zap.Logger
	repoURL       string
	filesAPI      string
	baseDir       string
	autoUpdate    bool
	githubProxy   string
	proxy         string
	githubHeaders map[string]string
	mutex         sync.RWMutex
}

// ResourceInfo 资源包信息
type ResourceInfo struct {
	Version   string               `json:"version"`
	Resources map[string]*Resource `json:"resources"`
}

// Resource 资源明细
type Resource struct {
	Type     string `json:"type"`
	Platform string `json:"platform"`
	Target   string `json:"target"`
	Version  string `json:"version"`
}

// ResourceFileInfo 文件信息
type ResourceFileInfo struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
	Path        string `json:"path"`
}

// NewResourceHelper 创建资源助手实例
func NewResourceHelper(repoURL, filesAPI, baseDir string, autoUpdate bool, githubProxy, proxy string, githubHeaders map[string]string) *ResourceHelper {
	return &ResourceHelper{
		logger:        logger.GetLogger(),
		repoURL:       repoURL,
		filesAPI:      filesAPI,
		baseDir:       baseDir,
		autoUpdate:    autoUpdate,
		githubProxy:   githubProxy,
		proxy:         proxy,
		githubHeaders: githubHeaders,
	}
}

// Check 检测资源包更新
func (h *ResourceHelper) Check() error {
	if !h.autoUpdate {
		h.logger.Debug("资源包自动更新已关闭")
		return nil
	}

	// 检测是否为可执行文件模式
	if isFrozen() {
		h.logger.Debug("可执行文件模式下，跳过资源包更新")
		return nil
	}

	h.logger.Info("开始检测资源包版本...")

	// 获取远程资源包信息
	resourceInfo, err := h.getRemoteResourceInfo()
	if err != nil {
		h.logger.Error("获取资源包信息失败", zap.Error(err))
		return fmt.Errorf("获取资源包信息失败: %w", err)
	}

	if resourceInfo == nil {
		h.logger.Warn("无法连接资源包仓库")
		return nil
	}

	h.logger.Info("最新资源包版本", zap.String("version", fmt.Sprintf("v%s", resourceInfo.Version)))

	// 检查需要更新的资源
	needUpdates, err := h.checkNeedUpdates(resourceInfo)
	if err != nil {
		h.logger.Error("检查资源更新失败", zap.Error(err))
		return fmt.Errorf("检查资源更新失败: %w", err)
	}

	if len(needUpdates) == 0 {
		h.logger.Info("所有资源已最新，无需更新")
		return nil
	}

	// 下载资源文件
	if err := h.downloadResources(needUpdates); err != nil {
		h.logger.Error("下载资源文件失败", zap.Error(err))
		return fmt.Errorf("下载资源文件失败: %w", err)
	}

	h.logger.Info("资源包更新完成")
	return nil
}

// getRemoteResourceInfo 获取远程资源包信息
func (h *ResourceHelper) getRemoteResourceInfo() (*ResourceInfo, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", h.repoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range h.githubHeaders {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var resourceInfo ResourceInfo
	if err := json.NewDecoder(resp.Body).Decode(&resourceInfo); err != nil {
		return nil, fmt.Errorf("解析资源包信息失败: %w", err)
	}

	return &resourceInfo, nil
}

// checkNeedUpdates 检查需要更新的资源
func (h *ResourceHelper) checkNeedUpdates(resourceInfo *ResourceInfo) (map[string]string, error) {
	needUpdates := make(map[string]string)

	// 获取本地平台
	localPlatform := platform()

	// 遍历资源
	for rname, resource := range resourceInfo.Resources {
		// 检查平台是否匹配
		if resource.Platform != "" && resource.Platform != localPlatform {
			h.logger.Debug("资源平台不匹配，跳过", zap.String("resource", rname), zap.String("platform", resource.Platform))
			continue
		}

		// 检查版本号
		localVersion, err := h.getLocalResourceVersion(resource.Type)
		if err != nil {
			h.logger.Error("获取本地资源版本失败", zap.String("type", resource.Type), zap.Error(err))
			continue
		}

		// 比较版本
		if compareVersion(resource.Version, localVersion) > 0 {
			h.logger.Info("资源包有更新", zap.String("resource", rname), zap.String("local_version", localVersion), zap.String("remote_version", resource.Version))
			needUpdates[rname] = resource.Target
		}
	}

	return needUpdates, nil
}

// getLocalResourceVersion 获取本地资源版本
func (h *ResourceHelper) getLocalResourceVersion(resourceType string) (string, error) {
	// 根据资源类型返回本地版本
	switch resourceType {
	case "auth":
		// 站点认证资源
		return h.getAuthVersion()
	case "sites":
		// 站点索引资源
		return h.getIndexerVersion()
	default:
		return "0.0.0", fmt.Errorf("不支持的资源类型: %s", resourceType)
	}
}

// getAuthVersion 获取认证资源版本
func (h *ResourceHelper) getAuthVersion() (string, error) {
	// 读取本地认证资源文件，获取版本号
	authFile := filepath.Join(h.baseDir, "app", "sites", "auth", "version.txt")
	return h.readVersionFile(authFile)
}

// getIndexerVersion 获取索引资源版本
func (h *ResourceHelper) getIndexerVersion() (string, error) {
	// 读取本地索引资源文件，获取版本号
	indexerFile := filepath.Join(h.baseDir, "app", "sites", "indexer", "version.txt")
	return h.readVersionFile(indexerFile)
}

// readVersionFile 读取版本文件
func (h *ResourceHelper) readVersionFile(filePath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "0.0.0", nil
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "0.0.0", fmt.Errorf("读取版本文件失败: %w", err)
	}

	// 去除空格和换行符
	version := strings.TrimSpace(string(data))
	if version == "" {
		return "0.0.0", nil
	}

	return version, nil
}

// downloadResources 下载资源文件
func (h *ResourceHelper) downloadResources(needUpdates map[string]string) error {
	// 获取文件列表
	fileList, err := h.getFileList()
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 下载需要更新的文件
	for _, file := range fileList {
		targetDir, exists := needUpdates[file.Name]
		if !exists {
			continue
		}

		if file.DownloadURL == "" {
			h.logger.Debug("文件没有下载链接，跳过", zap.String("file", file.Name))
			continue
		}

		h.logger.Info("开始下载资源文件", zap.String("file", file.Name))

		// 构建下载URL
		downloadURL := file.DownloadURL
		if h.githubProxy != "" {
			downloadURL = fmt.Sprintf("%s%s", h.githubProxy, downloadURL)
		}

		// 下载文件
		data, err := h.downloadFile(downloadURL)
		if err != nil {
			return fmt.Errorf("下载文件失败 %s: %w", file.Name, err)
		}

		// 构建保存路径
		savePath := filepath.Join(h.baseDir, targetDir, file.Name)

		// 创建目录
		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", filepath.Dir(savePath), err)
		}

		// 写入文件
		if err := os.WriteFile(savePath, data, 0644); err != nil {
			return fmt.Errorf("写入文件失败 %s: %w", savePath, err)
		}

		h.logger.Info("资源文件下载成功", zap.String("file", file.Name), zap.String("path", savePath))
	}

	return nil
}

// getFileList 获取文件列表
func (h *ResourceHelper) getFileList() ([]*ResourceFileInfo, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", h.filesAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range h.githubHeaders {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var fileList []*ResourceFileInfo
	if err := json.NewDecoder(resp.Body).Decode(&fileList); err != nil {
		return nil, fmt.Errorf("解析文件列表失败: %w", err)
	}

	return fileList, nil
}

// downloadFile 下载文件
func (h *ResourceHelper) downloadFile(url string) ([]byte, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 180 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range h.githubHeaders {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return data, nil
}

// compareVersion 比较版本号
// 返回值：1 - v1 > v2, 0 - v1 == v2, -1 - v1 < v2
func compareVersion(v1, v2 string) int {
	// 分割版本号
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// 取最大长度
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	// 比较每个部分
	for i := 0; i < maxLen; i++ {
		// 获取当前部分的版本号
		p1 := 0
		p2 := 0

		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &p1)
		}

		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		// 比较
		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	return 0
}

// SetAutoUpdate 设置自动更新开关
func (h *ResourceHelper) SetAutoUpdate(autoUpdate bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.autoUpdate = autoUpdate
}

// GetAutoUpdate 获取自动更新开关状态
func (h *ResourceHelper) GetAutoUpdate() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.autoUpdate
}

// 本地系统工具函数，避免循环依赖

// isFrozen 检查是否为可执行文件模式
func isFrozen() bool {
	// 简化实现：检查是否存在frozen环境变量
	return os.Getenv("FROZEN") == "1"
}

// platform 获取本地平台
func platform() string {
	// 简化实现：返回当前操作系统
	return "linux"
}
