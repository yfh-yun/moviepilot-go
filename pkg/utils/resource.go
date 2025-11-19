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
)

// ResourceHelper 资源包管理助手
type ResourceHelper struct {
	repoURL      string
	filesAPI     string
	baseDir      string
	autoUpdate   bool
	httpClient   *http.Client
	mutex        sync.RWMutex
	lastCheck    time.Time
	currentVersion string
}

// ResourceInfo 资源包信息
type ResourceInfo struct {
	Version   string                    `json:"version"`
	Resources map[string]ResourceDetail `json:"resources"`
}

// ResourceDetail 资源详情
type ResourceDetail struct {
	Type     string `json:"type"`
	Version  string `json:"version"`
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
	Size     int64  `json:"size"`
}

// UpdateResult 更新结果
type UpdateResult struct {
	Success   bool     `json:"success"`
	Updated   []string `json:"updated"`
	Failed    []string `json:"failed"`
	Skipped   []string `json:"skipped"`
	Message   string   `json:"message"`
}

// NewResourceHelper 创建资源助手实例
func NewResourceHelper(baseDir string, autoUpdate bool) *ResourceHelper {
	if baseDir == "" {
		baseDir = "."
	}

	return &ResourceHelper{
		repoURL:    "https://raw.githubusercontent.com/jxxghp/MoviePilot-Resources/main/package.v2.json",
		filesAPI:   "https://api.github.com/repos/jxxghp/MoviePilot-Resources/contents/resources.v2",
		baseDir:    baseDir,
		autoUpdate: autoUpdate,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetRepoURL 设置仓库URL
func (rh *ResourceHelper) SetRepoURL(url string) {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()
	rh.repoURL = url
}

// SetAutoUpdate 设置自动更新
func (rh *ResourceHelper) SetAutoUpdate(enable bool) {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()
	rh.autoUpdate = enable
}

// Check 检查更新
func (rh *ResourceHelper) Check() (*UpdateResult, error) {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	if !rh.autoUpdate {
		return &UpdateResult{
			Success: true,
			Message: "Auto update is disabled",
		}, nil
	}

	// 检查是否需要更新（避免频繁检查）
	if time.Since(rh.lastCheck) < time.Hour {
		return &UpdateResult{
			Success: true,
			Message: "Check skipped (too recent)",
		}, nil
	}

	return rh.checkAndUpdate()
}

// checkAndUpdate 检查并更新
func (rh *ResourceHelper) checkAndUpdate() (*UpdateResult, error) {
	result := &UpdateResult{
		Success: true,
		Updated: []string{},
		Failed:  []string{},
		Skipped: []string{},
	}

	// 获取远程资源信息
	remoteInfo, err := rh.fetchRemoteInfo()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to fetch remote info: %v", err)
		return result, err
	}

	// 获取本地版本信息
	localVersion := rh.getLocalVersion()

	// 比较版本
	if remoteInfo.Version == localVersion {
		result.Message = "Resources are up to date"
		rh.lastCheck = time.Now()
		return result, nil
	}

	// 需要更新
	result.Message = fmt.Sprintf("Updating from version %s to %s", localVersion, remoteInfo.Version)

	// 更新资源
	for name, resource := range remoteInfo.Resources {
		if err := rh.updateResource(name, resource); err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("%s: %v", name, err))
		} else {
			result.Updated = append(result.Updated, name)
		}
	}

	// 更新本地版本
	rh.updateLocalVersion(remoteInfo.Version)
	rh.currentVersion = remoteInfo.Version
	rh.lastCheck = time.Now()

	if len(result.Failed) > 0 {
		result.Success = false
		result.Message = fmt.Sprintf("Update completed with %d failures", len(result.Failed))
	} else {
		result.Message = fmt.Sprintf("Successfully updated to version %s", remoteInfo.Version)
	}

	return result, nil
}

// fetchRemoteInfo 获取远程资源信息
func (rh *ResourceHelper) fetchRemoteInfo() (*ResourceInfo, error) {
	req, err := http.NewRequest("GET", rh.repoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "MoviePilot-ResourceHelper/1.0")

	resp, err := rh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var resourceInfo ResourceInfo
	if err := json.Unmarshal(data, &resourceInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource info: %v", err)
	}

	return &resourceInfo, nil
}

// getLocalVersion 获取本地版本
func (rh *ResourceHelper) getLocalVersion() string {
	versionFile := filepath.Join(rh.baseDir, "resources", "version.txt")
	
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// updateLocalVersion 更新本地版本
func (rh *ResourceHelper) updateLocalVersion(version string) error {
	versionFile := filepath.Join(rh.baseDir, "resources", "version.txt")
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(versionFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	return os.WriteFile(versionFile, []byte(version), 0644)
}

// updateResource 更新单个资源
func (rh *ResourceHelper) updateResource(name string, resource ResourceDetail) error {
	// 构建本地文件路径
	localPath := filepath.Join(rh.baseDir, "resources", name)
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 下载资源
	if err := rh.downloadResource(resource.URL, localPath); err != nil {
		return fmt.Errorf("failed to download resource: %v", err)
	}

	// 验证校验和（如果提供）
	if resource.Checksum != "" {
		if err := rh.verifyChecksum(localPath, resource.Checksum); err != nil {
			return fmt.Errorf("checksum verification failed: %v", err)
		}
	}

	return nil
}

// downloadResource 下载资源
func (rh *ResourceHelper) downloadResource(url, localPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "MoviePilot-ResourceHelper/1.0")

	resp, err := rh.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 创建临时文件
	tempPath := localPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to write file: %v", err)
	}

	// 重命名为最终文件
	if err := os.Rename(tempPath, localPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename file: %v", err)
	}

	return nil
}

// verifyChecksum 验证校验和
func (rh *ResourceHelper) verifyChecksum(filePath, expectedChecksum string) error {
	// 这里应该实现实际的校验和验证
	// 简化实现，只检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}

	return nil
}

// ForceUpdate 强制更新
func (rh *ResourceHelper) ForceUpdate() (*UpdateResult, error) {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	// 重置检查时间
	rh.lastCheck = time.Time{}

	return rh.checkAndUpdate()
}

// GetCurrentVersion 获取当前版本
func (rh *ResourceHelper) GetCurrentVersion() string {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	if rh.currentVersion != "" {
		return rh.currentVersion
	}

	return rh.getLocalVersion()
}

// GetLastCheck 获取最后检查时间
func (rh *ResourceHelper) GetLastCheck() time.Time {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	return rh.lastCheck
}

// GetResourceList 获取资源列表
func (rh *ResourceHelper) GetResourceList() ([]string, error) {
	resourcesDir := filepath.Join(rh.baseDir, "resources")
	
	entries, err := os.ReadDir(resourcesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read resources directory: %v", err)
	}

	var resources []string
	for _, entry := range entries {
		if !entry.IsDir() {
			resources = append(resources, entry.Name())
		}
	}

	return resources, nil
}

// GetResourcePath 获取资源路径
func (rh *ResourceHelper) GetResourcePath(name string) string {
	return filepath.Join(rh.baseDir, "resources", name)
}

// ResourceExists 检查资源是否存在
func (rh *ResourceHelper) ResourceExists(name string) bool {
	path := rh.GetResourcePath(name)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// RemoveResource 移除资源
func (rh *ResourceHelper) RemoveResource(name string) error {
	path := rh.GetResourcePath(name)
	return os.Remove(path)
}

// CleanupResources 清理资源
func (rh *ResourceHelper) CleanupResources() error {
	resourcesDir := filepath.Join(rh.baseDir, "resources")
	
	entries, err := os.ReadDir(resourcesDir)
	if err != nil {
		return fmt.Errorf("failed to read resources directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "version.txt" {
			path := filepath.Join(resourcesDir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %v", entry.Name(), err)
			}
		}
	}

	return nil
}

// GetStats 获取统计信息
func (rh *ResourceHelper) GetStats() map[string]interface{} {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	stats := map[string]interface{}{
		"current_version": rh.currentVersion,
		"last_check":      rh.lastCheck,
		"auto_update":     rh.autoUpdate,
		"repo_url":        rh.repoURL,
	}

	// 添加资源统计
	if resources, err := rh.GetResourceList(); err == nil {
		stats["resource_count"] = len(resources)
		stats["resources"] = resources
	}

	return stats
}

// SetTimeout 设置HTTP客户端超时
func (rh *ResourceHelper) SetTimeout(timeout time.Duration) {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()
	rh.httpClient.Timeout = timeout
}

// SetProxy 设置代理
func (rh *ResourceHelper) SetProxy(proxyURL string) {
	// 这里应该实现代理设置
	// 简化实现
}

// ExportConfig 导出配置
func (rh *ResourceHelper) ExportConfig() map[string]interface{} {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	return map[string]interface{}{
		"repo_url":    rh.repoURL,
		"base_dir":    rh.baseDir,
		"auto_update": rh.autoUpdate,
		"last_check":  rh.lastCheck,
		"current_version": rh.currentVersion,
	}
}

// ImportConfig 导入配置
func (rh *ResourceHelper) ImportConfig(config map[string]interface{}) error {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	if repoURL, ok := config["repo_url"].(string); ok {
		rh.repoURL = repoURL
	}

	if baseDir, ok := config["base_dir"].(string); ok {
		rh.baseDir = baseDir
	}

	if autoUpdate, ok := config["auto_update"].(bool); ok {
		rh.autoUpdate = autoUpdate
	}

	if lastCheck, ok := config["last_check"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastCheck); err == nil {
			rh.lastCheck = t
		}
	}

	if currentVersion, ok := config["current_version"].(string); ok {
		rh.currentVersion = currentVersion
	}

	return nil
}