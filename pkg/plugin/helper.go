package plugin

import (
	"archive/zip"
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

// PluginHelper 插件助手，用于管理插件的安装、卸载、更新等操作
type PluginHelper struct {
	logger           *zap.Logger
	baseURL          string
	installReg       string
	installReport    string
	installStatistic string
	pluginDir        string
	tempDir          string
	mutex            sync.RWMutex
}

// PluginPackage 插件包信息
type PluginPackage struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Author      string `json:"author"`
	AuthorURL   string `json:"author_url"`
	Label       string `json:"label"`
	Release     bool   `json:"release"`
	V2          bool   `json:"v2,omitempty"`
	V3          bool   `json:"v3,omitempty"`
}

// NewPluginHelper 创建插件助手实例
func NewPluginHelper(pluginDir, tempDir string) *PluginHelper {
	return &PluginHelper{
		logger:           logger.GetLogger(),
		baseURL:          "https://raw.githubusercontent.com/{user}/{repo}/main/",
		installReg:       "{mp_server_host}/plugin/install/{pid}",
		installReport:    "{mp_server_host}/plugin/install",
		installStatistic: "{mp_server_host}/plugin/statistic",
		pluginDir:        pluginDir,
		tempDir:          tempDir,
	}
}

// GetPlugins 获取插件列表
func (h *PluginHelper) GetPlugins(repoURL, packageVersion string, force bool) (map[string]*PluginPackage, error) {
	// 解析仓库信息
	user, repo, err := h.parseRepoURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("解析仓库地址失败: %w", err)
	}

	// 构建请求URL
	rawURL := strings.Replace(h.baseURL, "{user}", user, 1)
	rawURL = strings.Replace(rawURL, "{repo}", repo, 1)

	var packageURL string
	if packageVersion != "" {
		packageURL = fmt.Sprintf("%spackage.%s.json", rawURL, packageVersion)
	} else {
		packageURL = fmt.Sprintf("%spackage.json", rawURL)
	}

	// 发送请求
	resp, err := http.Get(packageURL)
	if err != nil {
		return nil, fmt.Errorf("请求插件列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取插件列表失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var plugins map[string]*PluginPackage
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("解析插件列表失败: %w", err)
	}

	return plugins, nil
}

// GetPluginPackageVersion 获取插件包版本
func (h *PluginHelper) GetPluginPackageVersion(pid, repoURL, packageVersion string) (string, error) {
	// 解析仓库信息
	user, repo, err := h.parseRepoURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("解析仓库地址失败: %w", err)
	}

	// 构建请求URL
	rawURL := strings.Replace(h.baseURL, "{user}", user, 1)
	rawURL = strings.Replace(rawURL, "{repo}", repo, 1)

	// 1. 检查指定版本的插件
	if packageVersion != "" {
		packageURL := fmt.Sprintf("%spackage.%s.json", rawURL, packageVersion)
		resp, err := http.Get(packageURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			var plugins map[string]*PluginPackage
			if err := json.NewDecoder(resp.Body).Decode(&plugins); err == nil {
				if _, exists := plugins[pid]; exists {
					resp.Body.Close()
					return packageVersion, nil
				}
			}
			resp.Body.Close()
		}
	}

	// 2. 检查默认版本的插件
	packageURL := fmt.Sprintf("%spackage.json", rawURL)
	resp, err := http.Get(packageURL)
	if err != nil {
		return "", fmt.Errorf("请求插件列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取插件列表失败，状态码: %d", resp.StatusCode)
	}

	var plugins map[string]*PluginPackage
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return "", fmt.Errorf("解析插件列表失败: %w", err)
	}

	plugin := plugins[pid]
	if plugin == nil {
		return "", fmt.Errorf("插件不存在: %s", pid)
	}

	// 检查插件是否兼容指定版本
	if packageVersion == "v2" && plugin.V2 {
		return "", nil // 兼容v2版本
	} else if packageVersion == "v3" && plugin.V3 {
		return "", nil // 兼容v3版本
	}

	return "", fmt.Errorf("插件不兼容指定版本: %s", packageVersion)
}

// InstallPlugin 安装插件
func (h *PluginHelper) InstallPlugin(pid, repoURL, packageVersion string, forceInstall bool) error {
	// 解析仓库信息
	user, repo, err := h.parseRepoURL(repoURL)
	if err != nil {
		return fmt.Errorf("解析仓库地址失败: %w", err)
	}

	userRepo := fmt.Sprintf("%s/%s", user, repo)

	// 获取插件包版本
	pkgVersion, err := h.GetPluginPackageVersion(pid, repoURL, packageVersion)
	if err != nil {
		return fmt.Errorf("获取插件包版本失败: %w", err)
	}
	if pkgVersion == "" {
		pkgVersion = ""
	}

	// 获取插件元信息
	meta, err := h.getPluginMeta(pid, repoURL, &pkgVersion)
	if err != nil {
		return fmt.Errorf("获取插件元信息失败: %w", err)
	}

	// 决定安装方式
	if meta.Release {
		// 使用Release安装
		return h.installFromRelease(pid, userRepo, fmt.Sprintf("%s_v%s", pid, meta.Version), forceInstall)
	} else {
		// 使用文件列表安装
		return h.installFromFileList(pid, userRepo, &pkgVersion, forceInstall)
	}
}

// UninstallPlugin 卸载插件
func (h *PluginHelper) UninstallPlugin(pid string) error {
	pluginDir := filepath.Join(h.pluginDir, strings.ToLower(pid))

	// 检查插件是否存在
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("插件不存在: %s", pid)
	}

	// 删除插件目录
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("删除插件目录失败: %w", err)
	}

	h.logger.Info("插件已卸载", zap.String("pid", pid))
	return nil
}

// UpdatePlugin 更新插件
func (h *PluginHelper) UpdatePlugin(pid, repoURL, packageVersion string) error {
	// 先卸载旧插件
	if err := h.UninstallPlugin(pid); err != nil {
		return fmt.Errorf("卸载旧插件失败: %w", err)
	}

	// 安装新插件
	return h.InstallPlugin(pid, repoURL, packageVersion, true)
}

// GetInstalledPlugins 获取已安装的插件列表
func (h *PluginHelper) GetInstalledPlugins() ([]string, error) {
	// 读取插件目录
	entries, err := os.ReadDir(h.pluginDir)
	if err != nil {
		return nil, fmt.Errorf("读取插件目录失败: %w", err)
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			plugins = append(plugins, entry.Name())
		}
	}

	return plugins, nil
}

// parseRepoURL 解析仓库URL，提取user和repo
func (h *PluginHelper) parseRepoURL(repoURL string) (user, repo string, err error) {
	// 标准化URL
	if !strings.HasSuffix(repoURL, "/") {
		repoURL += "/"
	}

	// 解析user和repo
	parts := strings.Split(repoURL, "/")
	if len(parts) < 6 {
		return "", "", fmt.Errorf("不支持的仓库地址格式")
	}

	user = parts[len(parts)-4]
	repo = parts[len(parts)-3]
	return user, repo, nil
}

// getPluginMeta 获取插件元信息
func (h *PluginHelper) getPluginMeta(pid, repoURL string, packageVersion *string) (*PluginPackage, error) {
	plugins, err := h.GetPlugins(repoURL, func() string {
		if packageVersion != nil {
			return *packageVersion
		}
		return ""
	}(), true)
	if err != nil {
		return nil, err
	}

	plugin := plugins[pid]
	if plugin == nil {
		return nil, fmt.Errorf("插件不存在: %s", pid)
	}

	return plugin, nil
}

// installFromRelease 从Release安装插件
func (h *PluginHelper) installFromRelease(pid, userRepo, releaseTag string, forceInstall bool) error {
	// 构建Release API URL
	releaseAPI := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", userRepo, releaseTag)

	// 请求Release信息
	resp, err := http.Get(releaseAPI)
	if err != nil {
		return fmt.Errorf("请求Release信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取Release信息失败，状态码: %d", resp.StatusCode)
	}

	// 解析Release信息
	var releaseInfo struct {
		Assets []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return fmt.Errorf("解析Release信息失败: %w", err)
	}

	// 查找资产文件
	var assetID int
	assetName := fmt.Sprintf("%s.zip", strings.ToLower(releaseTag))
	found := false

	for _, asset := range releaseInfo.Assets {
		if asset.Name == assetName {
			assetID = asset.ID
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到资产文件: %s", assetName)
	}

	// 下载资产文件
	downloadURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", userRepo, assetID)
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("创建下载请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载资产文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载资产文件失败，状态码: %d", resp.StatusCode)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(h.tempDir, "plugin_*.zip")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tempFilePath := tempFile.Name()
	defer func() {
		tempFile.Close()
		os.Remove(tempFilePath)
	}()

	// 写入临时文件
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	tempFile.Close()

	// 解压文件
	return h.extractReleaseZip(pid, tempFilePath)
}

// installFromFileList 从文件列表安装插件
func (h *PluginHelper) installFromFileList(pid, userRepo string, packageVersion *string, forceInstall bool) error {
	// 获取文件列表
	fileList, err := h.getFileList(pid, userRepo, packageVersion)
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 创建插件目录
	pluginDir := filepath.Join(h.pluginDir, strings.ToLower(pid))
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("创建插件目录失败: %w", err)
	}

	// 下载文件
	return h.downloadFiles(fileList, pluginDir, userRepo, packageVersion)
}

// getFileList 获取插件文件列表
func (h *PluginHelper) getFileList(pid, userRepo string, packageVersion *string) ([]*fileInfo, error) {
	// 构建API URL
	fileAPI := fmt.Sprintf("https://api.github.com/repos/%s/contents/plugins", userRepo)
	if packageVersion != nil {
		fileAPI += fmt.Sprintf(".%s", *packageVersion)
	}
	fileAPI += fmt.Sprintf("/%s", strings.ToLower(pid))

	// 请求文件列表
	resp, err := http.Get(fileAPI)
	if err != nil {
		return nil, fmt.Errorf("请求文件列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取文件列表失败，状态码: %d", resp.StatusCode)
	}

	// 解析文件列表
	var fileInfos []*fileInfo
	if err := json.NewDecoder(resp.Body).Decode(&fileInfos); err != nil {
		return nil, fmt.Errorf("解析文件列表失败: %w", err)
	}

	return fileInfos, nil
}

// fileInfo 文件信息
type fileInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		HTML string `json:"html"`
	} `json:"_links"`
}

// downloadFiles 下载文件
func (h *PluginHelper) downloadFiles(fileList []*fileInfo, pluginDir, userRepo string, packageVersion *string) error {
	// 使用栈来替代递归，避免递归深度过大
	stack := make([]*fileInfo, len(fileList))
	copy(stack, fileList)

	for len(stack) > 0 {
		// 弹出栈顶元素
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if item.Type == "dir" {
			// 递归获取子目录文件
			subFiles, err := h.getFileList(filepath.Join(item.Path, item.Name), userRepo, packageVersion)
			if err != nil {
				return fmt.Errorf("获取子目录文件列表失败: %w", err)
			}
			stack = append(stack, subFiles...)
		} else if item.DownloadURL != "" {
			// 下载文件
			destPath := filepath.Join(pluginDir, strings.TrimPrefix(item.Path, func() string {
				prefix := "plugins"
				if packageVersion != nil {
					prefix += fmt.Sprintf(".%s", *packageVersion)
				}
				return prefix + "/" + strings.ToLower(filepath.Base(pluginDir)) + "/"
			}()))

			// 创建目录
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}

			// 下载文件
			resp, err := http.Get(item.DownloadURL)
			if err != nil {
				return fmt.Errorf("下载文件失败: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("下载文件失败，状态码: %d", resp.StatusCode)
			}

			// 写入文件
			destFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("创建文件失败: %w", err)
			}
			defer destFile.Close()

			if _, err := io.Copy(destFile, resp.Body); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}

			h.logger.Debug("文件下载成功", zap.String("path", destPath))
		}
	}

	return nil
}

// extractReleaseZip 解压Release压缩包
func (h *PluginHelper) extractReleaseZip(pid, zipPath string) error {
	// 打开压缩包
	zipFile, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开压缩包失败: %w", err)
	}
	defer zipFile.Close()

	// 创建插件目录
	pluginDir := filepath.Join(h.pluginDir, strings.ToLower(pid))
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("创建插件目录失败: %w", err)
	}

	// 确定基础前缀
	basePrefix := ""
	if len(zipFile.File) > 0 {
		firstFile := zipFile.File[0]
		if strings.Contains(firstFile.Name, "/") {
			basePrefix = firstFile.Name[:strings.Index(firstFile.Name, "/")+1]
		}
	}

	// 解压文件
	for _, file := range zipFile.File {
		// 跳过目录
		if file.FileInfo().IsDir() {
			continue
		}

		// 计算目标路径
		relPath := strings.TrimPrefix(file.Name, basePrefix)
		destPath := filepath.Join(pluginDir, relPath)

		// 创建目录
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		// 打开源文件
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开源文件失败: %w", err)
		}

		// 创建目标文件
		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("创建目标文件失败: %w", err)
		}

		// 复制文件
		if _, err := io.Copy(destFile, srcFile); err != nil {
			srcFile.Close()
			destFile.Close()
			return fmt.Errorf("复制文件失败: %w", err)
		}

		srcFile.Close()
		destFile.Close()

		h.logger.Debug("文件解压成功", zap.String("path", destPath))
	}

	h.logger.Info("插件安装成功", zap.String("pid", pid))
	return nil
}

// InstallReg 安装统计
func (h *PluginHelper) InstallReg(pid, repoURL, mpServerHost string) error {
	// 替换占位符
	installRegURL := strings.Replace(h.installReg, "{mp_server_host}", mpServerHost, 1)
	installRegURL = strings.Replace(installRegURL, "{pid}", pid, 1)

	// 构建请求体
	reqBody := map[string]any{
		"plugin_id": pid,
		"repo_url":  repoURL,
	}

	// 发送请求
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	resp, err := http.Post(installRegURL, "application/json", strings.NewReader(string(reqJSON)))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("安装统计失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// InstallReport 上报插件安装统计
func (h *PluginHelper) InstallReport(items []struct {
	PID     string
	RepoURL string
}, mpServerHost string) error {
	// 替换占位符
	installReportURL := strings.Replace(h.installReport, "{mp_server_host}", mpServerHost, 1)

	// 构建请求体
	payload := make([]map[string]any, len(items))
	for i, item := range items {
		payload[i] = map[string]any{
			"plugin_id": item.PID,
			"repo_url":  item.RepoURL,
		}
	}

	reqBody := map[string]any{
		"plugins": payload,
	}

	// 发送请求
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	resp, err := http.Post(installReportURL, "application/json", strings.NewReader(string(reqJSON)))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("上报安装统计失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// GetStatistic 获取插件安装统计
func (h *PluginHelper) GetStatistic(mpServerHost string) (map[string]any, error) {
	// 替换占位符
	statisticURL := strings.Replace(h.installStatistic, "{mp_server_host}", mpServerHost, 1)

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(statisticURL)
	if err != nil {
		return nil, fmt.Errorf("请求统计信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取统计信息失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析统计信息失败: %w", err)
	}

	return result, nil
}
