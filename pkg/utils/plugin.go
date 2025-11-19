package utils

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

// PluginHelper 插件市场管理，下载安装插件到本地
type PluginHelper struct {
	baseURL             string
	installRegURL       string
	installReportURL    string
	installStatisticURL string
	httpClient          *http.Client
	redisClient         *redis.Client
	semaphore           *semaphore.Weighted
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Homepage    string            `json:"homepage"`
	License     string            `json:"license"`
	Release     bool              `json:"release"`
	Files       []PluginFile      `json:"files"`
	Dependencies []string         `json:"dependencies"`
	Config      map[string]interface{} `json:"config"`
	Tags        []string          `json:"tags"`
	V2          bool              `json:"v2"`
	V3          bool              `json:"v3"`
}

// PluginFile 插件文件信息
type PluginFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url"`
	SHA256      string `json:"sha256"`
	Type        string `json:"type"` // file, dir
}

// PluginPackage 插件包信息
type PluginPackage struct {
	Version   string                    `json:"version"`
	Plugins   map[string]*PluginMetadata `json:"plugins"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

// PluginInstallResult 插件安装结果
type PluginInstallResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Version string `json:"version,omitempty"`
}

// NewPluginHelper 创建插件助手实例
func NewPluginHelper(redisClient *redis.Client) *PluginHelper {
	return &PluginHelper{
		baseURL:             "https://raw.githubusercontent.com/{user}/{repo}/main/",
		installRegURL:        "http://localhost:3000/plugin/install/{pid}",
		installReportURL:     "http://localhost:3000/plugin/install",
		installStatisticURL:  "http://localhost:3000/plugin/statistic",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		redisClient: redisClient,
		semaphore:   semaphore.NewWeighted(5), // 限制并发安装数量
	}
}

// GetPlugins 获取Github所有最新插件列表
func (p *PluginHelper) GetPlugins(ctx context.Context, repoURL string, packageVersion string, force bool) (map[string]*PluginMetadata, error) {
	if force {
		return p.requestPlugins(ctx, repoURL, packageVersion)
	}
	
	// 使用缓存
	cacheKey := fmt.Sprintf("plugins:%s:%s", repoURL, packageVersion)
	cached, err := p.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var result map[string]*PluginMetadata
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result, nil
		}
	}
	
	// 缓存未命中，请求远程
	plugins, err := p.requestPlugins(ctx, repoURL, packageVersion)
	if err != nil {
		return nil, err
	}
	
	// 缓存结果
	if data, err := json.Marshal(plugins); err == nil {
		p.redisClient.Set(ctx, cacheKey, data, 30*time.Minute)
	}
	
	return plugins, nil
}

// requestPlugins 请求插件列表
func (p *PluginHelper) requestPlugins(ctx context.Context, repoURL string, packageVersion string) (map[string]*PluginMetadata, error) {
	if repoURL == "" {
		return nil, errors.New("repo URL is empty")
	}
	
	user, repo, err := p.parseRepoInfo(repoURL)
	if err != nil {
		return nil, err
	}
	
	rawURL := strings.ReplaceAll(p.baseURL, "{user}", user)
	rawURL = strings.ReplaceAll(rawURL, "{repo}", repo)
	
	var packageURL string
	if packageVersion != "" {
		packageURL = fmt.Sprintf("%spackage.%s.json", rawURL, packageVersion)
	} else {
		packageURL = rawURL + "package.json"
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", packageURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "MoviePilot-Go/1.0")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	var pluginPackage PluginPackage
	if err := json.NewDecoder(resp.Body).Decode(&pluginPackage); err != nil {
		return nil, errors.Wrap(err, "failed to decode plugin package")
	}
	
	return pluginPackage.Plugins, nil
}

// GetPluginPackageVersion 检查并获取指定插件的可用版本
func (p *PluginHelper) GetPluginPackageVersion(ctx context.Context, pid string, repoURL string, packageVersion string) (string, error) {
	// 如果没有指定版本，使用默认版本
	if packageVersion == "" {
		packageVersion = "v2" // 默认版本
	}
	
	// 优先检查指定版本的插件
	plugins, err := p.GetPlugins(ctx, repoURL, packageVersion, false)
	if err != nil {
		return "", err
	}
	
	if _, exists := plugins[pid]; exists {
		return packageVersion, nil
	}
	
	// 检查全局package.json文件
	plugins, err = p.GetPlugins(ctx, repoURL, "", false)
	if err != nil {
		return "", err
	}
	
	if plugin, exists := plugins[pid]; exists {
		// 检查插件是否支持指定版本
		if packageVersion == "v2" && plugin.V2 {
			return "", nil // 使用package.json
		}
		if packageVersion == "v3" && plugin.V3 {
			return "", nil // 使用package.json
		}
	}
	
	return "", fmt.Errorf("plugin %s not found for version %s", pid, packageVersion)
}

// parseRepoInfo 解析GitHub仓库信息
func (p *PluginHelper) parseRepoInfo(repoURL string) (string, string, error) {
	if repoURL == "" {
		return "", "", errors.New("repo URL is empty")
	}
	
	if !strings.HasSuffix(repoURL, "/") {
		repoURL += "/"
	}
	
	if strings.Count(repoURL, "/") < 6 {
		repoURL += "main/"
	}
	
	parts := strings.Split(repoURL, "/")
	if len(parts) < 5 {
		return "", "", errors.New("invalid repo URL format")
	}
	
	user := parts[len(parts)-4]
	repo := parts[len(parts)-3]
	
	if user == "" || repo == "" {
		return "", "", errors.New("failed to extract user and repo from URL")
	}
	
	return user, repo, nil
}

// Install 安装插件
func (p *PluginHelper) Install(ctx context.Context, pid string, repoURL string, packageVersion string, forceInstall bool) (*PluginInstallResult, error) {
	// 限制并发安装
	if err := p.semaphore.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer p.semaphore.Release(1)
	
	// 验证参数
	if pid == "" || repoURL == "" {
		return &PluginInstallResult{
			Success: false,
			Message: "参数错误",
		}, errors.New("pid and repoURL are required")
	}
	
	user, repo, err := p.parseRepoInfo(repoURL)
	if err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "不支持的插件仓库地址格式",
		}, err
	}
	
	userRepo := fmt.Sprintf("%s/%s", user, repo)
	
	if packageVersion == "" {
		packageVersion = "v2" // 默认版本
	}
	
	// 检查插件版本兼容性
	actualVersion, err := p.GetPluginPackageVersion(ctx, pid, repoURL, packageVersion)
	if err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: fmt.Sprintf("%s 没有找到适用于当前版本的插件", pid),
		}, err
	}
	
	// 获取插件元数据
	metadata, err := p.getPluginMetadata(ctx, pid, repoURL, actualVersion)
	if err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "获取插件元数据失败",
		}, err
	}
	
	// 执行安装
	if metadata.Release {
		return p.installFromRelease(ctx, pid, userRepo, metadata.Version, forceInstall)
	} else {
		return p.installFromFileList(ctx, pid, userRepo, actualVersion, forceInstall)
	}
}

// getPluginMetadata 获取插件元数据
func (p *PluginHelper) getPluginMetadata(ctx context.Context, pid string, repoURL string, packageVersion string) (*PluginMetadata, error) {
	plugins, err := p.GetPlugins(ctx, repoURL, packageVersion, false)
	if err != nil {
		return nil, err
	}
	
	metadata, exists := plugins[pid]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pid)
	}
	
	return metadata, nil
}

// installFromRelease 从Release安装插件
func (p *PluginHelper) installFromRelease(ctx context.Context, pid string, userRepo string, version string, forceInstall bool) (*PluginInstallResult, error) {
	releaseTag := fmt.Sprintf("%s_v%s", pid, version)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s.zip", userRepo, releaseTag, pid)
	
	// 下载zip文件
	resp, err := p.httpClient.Get(downloadURL)
	if err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "下载插件失败",
		}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return &PluginInstallResult{
			Success: false,
			Message: fmt.Sprintf("下载失败: HTTP %d", resp.StatusCode),
		}, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	
	// 创建临时文件
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.zip", pid))
	if err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "创建临时文件失败",
		}, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 保存zip文件
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "保存插件文件失败",
		}, err
	}
	
	// 解压并安装插件
	pluginDir := filepath.Join("app", "plugins", strings.ToLower(pid))
	
	if !forceInstall {
		// 备份现有插件
		if _, err := os.Stat(pluginDir); err == nil {
			backupDir := pluginDir + ".bak." + strconv.FormatInt(time.Now().Unix(), 10)
			if err := os.Rename(pluginDir, backupDir); err != nil {
				return &PluginInstallResult{
					Success: false,
					Message: "备份现有插件失败",
				}, err
			}
		}
	}
	
	// 确保插件目录存在
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "创建插件目录失败",
		}, err
	}
	
	// 解压zip文件
	if err := p.extractZip(tempFile.Name(), pluginDir); err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "解压插件失败",
		}, err
	}
	
	return &PluginInstallResult{
		Success: true,
		Message: "插件安装成功",
		Version: version,
	}, nil
}

// installFromFileList 从文件列表安装插件
func (p *PluginHelper) installFromFileList(ctx context.Context, pid string, userRepo string, packageVersion string, forceInstall bool) (*PluginInstallResult, error) {
	// 获取文件列表
	fileList, err := p.getFileList(ctx, pid, userRepo, packageVersion)
	if err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "获取文件列表失败",
		}, err
	}
	
	// 安装文件
	if err := p.downloadFiles(ctx, pid, fileList, userRepo, packageVersion, forceInstall); err != nil {
		return &PluginInstallResult{
			Success: false,
			Message: "下载文件失败",
		}, err
	}
	
	return &PluginInstallResult{
		Success: true,
		Message: "插件安装成功",
	}, nil
}

// getFileList 获取插件的文件列表
func (p *PluginHelper) getFileList(ctx context.Context, pid string, userRepo string, packageVersion string) ([]PluginFile, error) {
	var apiURL string
	if packageVersion != "" {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/contents/plugins.%s/%s", userRepo, packageVersion, strings.ToLower(pid))
	} else {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/contents/plugins/%s", userRepo, strings.ToLower(pid))
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "MoviePilot-Go/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}
	
	var contents []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}
	
	var files []PluginFile
	for _, content := range contents {
		var item struct {
			Name        string `json:"name"`
			Path        string `json:"path"`
			Size        int64  `json:"size"`
			DownloadURL string `json:"download_url"`
			Type        string `json:"type"`
		}
		
		if err := json.Unmarshal(content, &item); err != nil {
			continue
		}
		
		files = append(files, PluginFile{
			Name:        item.Name,
			Path:        item.Path,
			Size:        item.Size,
			DownloadURL: item.DownloadURL,
			Type:        item.Type,
		})
	}
	
	return files, nil
}

// downloadFiles 下载插件文件
func (p *PluginHelper) downloadFiles(ctx context.Context, pid string, fileList []PluginFile, userRepo string, packageVersion string, forceInstall bool) error {
	pluginDir := filepath.Join("app", "plugins", strings.ToLower(pid))
	
	if !forceInstall {
		// 备份现有插件
		if _, err := os.Stat(pluginDir); err == nil {
			backupDir := pluginDir + ".bak." + strconv.FormatInt(time.Now().Unix(), 10)
			if err := os.Rename(pluginDir, backupDir); err != nil {
				return err
			}
		}
	}
	
	// 确保插件目录存在
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}
	
	for _, file := range fileList {
		if file.Type == "dir" {
			// 创建目录
			dirPath := filepath.Join(pluginDir, file.Name)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return err
			}
			continue
		}
		
		if file.DownloadURL == "" {
			continue
		}
		
		// 下载文件
		req, err := http.NewRequestWithContext(ctx, "GET", file.DownloadURL, nil)
		if err != nil {
			return err
		}
		
		req.Header.Set("User-Agent", "MoviePilot-Go/1.0")
		
		resp, err := p.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download %s failed: %d", file.Name, resp.StatusCode)
		}
		
		// 保存文件
		filePath := filepath.Join(pluginDir, file.Name)
		fileDir := filepath.Dir(filePath)
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return err
		}
		
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		
		if _, err := io.Copy(f, resp.Body); err != nil {
			return err
		}
	}
	
	return nil
}

// extractZip 解压zip文件
func (p *PluginHelper) extractZip(zipPath string, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	
	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)
		
		// 确保路径在目标目录内
		if !strings.HasPrefix(path, destDir+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}
		
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}
		
		// 创建文件
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()
		
		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()
		
		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}
	
	return nil
}

// InstallReg 安装插件统计
func (p *PluginHelper) InstallReg(ctx context.Context, pid string, repoURL string) error {
	if pid == "" {
		return errors.New("pid is required")
	}
	
	url := strings.ReplaceAll(p.installRegURL, "{pid}", pid)
	
	payload := map[string]interface{}{
		"plugin_id": pid,
		"repo_url":  repoURL,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoviePilot-Go/1.0")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed: %d", resp.StatusCode)
	}
	
	return nil
}

// GetStatistic 获取插件安装统计
func (p *PluginHelper) GetStatistic(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.installStatisticURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "MoviePilot-Go/1.0")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statistic request failed: %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// Uninstall 卸载插件
func (p *PluginHelper) Uninstall(pid string, backup bool) error {
	pluginDir := filepath.Join("app", "plugins", strings.ToLower(pid))
	
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found", pid)
	}
	
	if backup {
		// 创建备份
		backupDir := pluginDir + ".uninstall." + strconv.FormatInt(time.Now().Unix(), 10)
		if err := os.Rename(pluginDir, backupDir); err != nil {
			return err
		}
	} else {
		// 直接删除
		if err := os.RemoveAll(pluginDir); err != nil {
			return err
		}
	}
	
	return nil
}

// ListInstalledPlugins 列出已安装的插件
func (p *PluginHelper) ListInstalledPlugins() ([]string, error) {
	pluginDir := "app/plugins"
	
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	
	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			plugins = append(plugins, entry.Name())
		}
	}
	
	return plugins, nil
}

// ValidatePlugin 验证插件完整性
func (p *PluginHelper) ValidatePlugin(pid string) error {
	pluginDir := filepath.Join("app", "plugins", strings.ToLower(pid))
	
	// 检查插件目录是否存在
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin directory not found")
	}
	
	// 检查必要文件
	requiredFiles := []string{"__init__.py", "plugin.py"}
	for _, file := range requiredFiles {
		filePath := filepath.Join(pluginDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file %s not found", file)
		}
	}
	
	return nil
}