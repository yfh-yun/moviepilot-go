package helper

import (
	"encoding/json"
	"fmt"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PluginHelper 插件市场管理，下载安装插件到本地
type PluginHelper struct {
	BaseURL          string
	InstallRegURL    string
	InstallReportURL string
	InstallStatURL   string
}

// NewPluginHelper 创建PluginHelper实例
func NewPluginHelper() *PluginHelper {
	cfg := config.GetConfig()
	serverHost := "https://movie-pilot.org" // 默认值，实际应该从配置获�?	
	return &PluginHelper{
		BaseURL:          "https://raw.githubusercontent.com/{user}/{repo}/main/",
		InstallRegURL:    fmt.Sprintf("%s/plugin/install/{pid}", serverHost),
		InstallReportURL: fmt.Sprintf("%s/plugin/install", serverHost),
		InstallStatURL:   fmt.Sprintf("%s/plugin/statistic", serverHost),
	}
}

// GetPlugins 获取Github所有最新插件列�?func (p *PluginHelper) GetPlugins(repoURL string, packageVersion *string, force bool) (map[string]interface{}, error) {
	/*
	 * 获取Github所有最新插件列�?	 * :param repoURL: Github仓库地址
	 * :param packageVersion: 首选插件版�?(�?"v2", "v3")，如果不指定则获�?v1 版本
	 * :param force: 是否强制刷新，忽略缓�?	 */
	// 如果强制刷新，直接调用不带缓存的版本
	if force {
		return p.requestPlugins(repoURL, packageVersion)
	} else {
		return p.requestPluginsCached(repoURL, packageVersion)
	}
}

// requestPluginsCached 获取Github所有最新插件列表（使用缓存�?func (p *PluginHelper) requestPluginsCached(repoURL string, packageVersion *string) (map[string]interface{}, error) {
	/*
	 * 获取Github所有最新插件列表（使用缓存�?	 * :param repoURL: Github仓库地址
	 * :param packageVersion: 首选插件版�?(�?"v2", "v3")，如果不指定则获�?v1 版本
	 */
	return p.requestPlugins(repoURL, packageVersion)
}

// requestPlugins 获取Github所有最新插件列表（不使用缓存）
func (p *PluginHelper) requestPlugins(repoURL string, packageVersion *string) (map[string]interface{}, error) {
	/*
	 * 获取Github所有最新插件列表（不使用缓存）
	 * :param repoURL: Github仓库地址
	 * :param packageVersion: 首选插件版�?(�?"v2", "v3")，如果不指定则获�?v1 版本
	 */
	if repoURL == "" {
		return nil, nil
	}

	user, repo := p.getRepoInfo(repoURL)
	if user == "" || repo == "" {
		return nil, nil
	}

	rawURL := strings.Replace(p.BaseURL, "{user}", user, -1)
	rawURL = strings.Replace(rawURL, "{repo}", repo, -1)
	
	var packageURL string
	if packageVersion != nil && *packageVersion != "" {
		packageURL = fmt.Sprintf("%spackage.%s.json", rawURL, *packageVersion)
	} else {
		packageURL = rawURL + "package.json"
	}

	res, err := p.requestWithFallback(packageURL, nil, 60, false)
	if err != nil || res == nil {
		return nil, err
	}
	
	if res.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(res.Content), &result); err != nil {
			if !strings.Contains(res.Content, "404: Not Found") {
				// 日志记录：插件包数据解析失败
				fmt.Printf("插件包数据解析失败：%s\n", res.Content)
			}
			return nil, err
		}
		return result, nil
	}
	
	return make(map[string]interface{}), nil
}

// GetPluginPackageVersion 检查并获取指定插件的可用版本，支持多版本优先级加载和版本兼容性检�?func (p *PluginHelper) GetPluginPackageVersion(pid string, repoURL string, packageVersion *string) *string {
	/*
	 * 检查并获取指定插件的可用版本，支持多版本优先级加载和版本兼容性检�?	 * 1. 如果未指定版本，则使用系统配置的默认版本（通过 settings.VERSION_FLAG 设置�?	 * 2. 优先检查指定版本的插件（如 `package.v2.json`�?	 * 3. 如果插件不存在于指定版本，检�?`package.json` 文件，查看该插件是否兼容指定版本
	 * 4. 如果插件不存在或不兼容指定版本，返回 `None`
	 * :param pid: 插件 ID，用于在插件列表中查�?	 * :param repoURL: 插件仓库�?URL，指定用于获取插件信息的 GitHub 仓库地址
	 * :param packageVersion: 首选插件版�?(�?"v2", "v3")，如不指定则默认使用系统配置的版�?	 * :return: 返回可用的插件版本号 (�?"v2"，如果指定版本不可用则返回空字符串表�?v1)，如果插件不可用则返�?None
	 */
	 
	// 如果没有指定版本，则使用当前系统配置的版本（�?"v2"�?	versionFlag := "v1" // 默认值，实际应该从配置获�?	if packageVersion == nil || *packageVersion == "" {
		packageVersion = &versionFlag
	}

	// 优先检查指定版本的插件，即 package.v(x).json 文件中是否存在该插件，如果存在，返回该版本号
	plugins, _ := p.GetPlugins(repoURL, packageVersion, false)
	if plugins != nil {
		if _, exists := plugins[pid]; exists {
			return packageVersion
		}
	}

	// 如果指定版本的插件不存在，检查全局 package.json 文件，查看插件是否兼容指定的版本
	globalPlugins, _ := p.GetPlugins(repoURL, nil, false)
	if globalPlugins != nil {
		if plugin, exists := globalPlugins[pid]; exists {
			// 检查插件是否明确支持当前指定的版本（如 v2 �?v3），如果支持，返回空字符串表示使�?package.json（v1�?			if pluginMap, ok := plugin.(map[string]interface{}); ok {
				if val, exists := pluginMap[*packageVersion]; exists {
					if val == true {
						emptyStr := ""
						return &emptyStr
					}
				}
			}
		}
	}

	// 如果所有版本都不存在或插件不兼容，返回 nil，表示插件不可用
	return nil
}

// getRepoInfo 获取GitHub仓库信息
func (p *PluginHelper) getRepoInfo(repoURL string) (string, string) {
	/*
	 * 获取GitHub仓库信息
	 */
	if repoURL == "" {
		return "", ""
	}
	
	if !strings.HasSuffix(repoURL, "/") {
		repoURL += "/"
	}
	
	// 确保URL格式正确
	if strings.Count(repoURL, "/") < 6 {
		repoURL = fmt.Sprintf("%smain/", repoURL)
	}
	
	parts := strings.Split(repoURL, "/")
	if len(parts) >= 6 {
		user := parts[len(parts)-4]
		repo := parts[len(parts)-3]
		return user, repo
	}
	
	// 日志记录：解析GitHub仓库地址失败
	fmt.Printf("解析GitHub仓库地址失败�?s\n", repoURL)
	return "", ""
}

// Install 安装插件，包括依赖安装和文件下载，相关资源支持自动降级策�?func (p *PluginHelper) Install(pid string, repoURL string, packageVersion *string, forceInstall bool) (bool, string) {
	/*
	 * 安装插件，包括依赖安装和文件下载，相关资源支持自动降级策�?	 * 1. 检查并获取插件的指定版本，确认版本兼容�?	 * 2. �?GitHub 获取文件列表（包�?requirements.txt�?	 * 3. 删除旧的插件目录（如非强制安装则进行备份�?	 * 4. 下载并预安装 requirements.txt 中的依赖（如果存在）
	 * 5. 下载并安装插件的其他文件
	 * 6. 再次尝试安装依赖（确保安装完整）
	 * :param pid: 插件 ID
	 * :param repoURL: 插件仓库地址
	 * :param packageVersion: 首选插件版�?(�?"v2", "v3")，如不指定则默认使用系统配置的版�?	 * :param forceInstall: 是否强制安装插件，默认不启用，启用时不进行备份和恢复操作
	 * :return: (是否成功, 错误信息)
	 */
	 
	// 检查是否为可执行文件模式（简化处理）
	if false { // 实际应该检查是否为可执行文件模�?		return false, "可执行文件模式下，只能安装本地插�?
	}

	// 验证参数
	if pid == "" || repoURL == "" {
		return false, "参数错误"
	}

	// �?GitHub �?repo_url 获取用户和项目名
	user, repo := p.getRepoInfo(repoURL)
	if user == "" || repo == "" {
		return false, "不支持的插件仓库地址格式"
	}

	userRepo := fmt.Sprintf("%s/%s", user, repo)

	// 如果没有指定版本，则使用默认版本
	versionFlag := "v1" // 默认值，实际应该从配置获�?	if packageVersion == nil || *packageVersion == "" {
		packageVersion = &versionFlag
	}

	// 1. 优先检查指定版本的插件
	packageVersion = p.GetPluginPackageVersion(pid, repoURL, packageVersion)
	// 如果 package_version 为nil，说明没有找到匹配的插件
	if packageVersion == nil {
		msg := fmt.Sprintf("%s 没有找到适用于当前版本的插件", pid)
		// 日志记录
		fmt.Println(msg)
		return false, msg
	}
	
	// package_version 为空，表示从 package.json 中找到插�?	if *packageVersion == "" {
		msg := fmt.Sprintf("%s �?package.json 中找到适用于当前版本的插件", pid)
		fmt.Println(msg)
	} else {
		msg := fmt.Sprintf("%s �?package.%s.json 中找到适用于当前版本的插件", pid, *packageVersion)
		fmt.Println(msg)
	}

	// 2. 决定安装方式（release �?文件列表）并执行统一安装流程
	meta := p.getPluginMeta(pid, repoURL, packageVersion)
	
	// 是否release打包
	isRelease := false
	if releaseVal, exists := meta["release"]; exists {
		if releaseBool, ok := releaseVal.(bool); ok {
			isRelease = releaseBool
		}
	}
	
	// 插件版本�?	var pluginVersion *string
	if versionVal, exists := meta["version"]; exists {
		if versionStr, ok := versionVal.(string); ok {
			pluginVersion = &versionStr
		}
	}
	
	if isRelease {
		// 使用 插件ID_插件版本�?作为 Release tag
		if pluginVersion == nil || *pluginVersion == "" {
			return false, fmt.Sprintf("未在插件清单中找�?%s 的版本号，无法进�?Release 安装", pid)
		}
		
		// 拼接 release_tag
		releaseTag := fmt.Sprintf("%s_v%s", pid, *pluginVersion)
		
		// 使用 release 进行安装
		return p.installFromRelease(pid, userRepo, releaseTag)
	} else {
		// 如果 release_tag 不存在，说明插件没有发布版本，使用文件列表方式安�?		return p.prepareContentViaFilelistSync(strings.ToLower(pid), userRepo, packageVersion)
	}
}

// getPluginMeta 获取插件元数�?func (p *PluginHelper) getPluginMeta(pid string, repoURL string, packageVersion *string) map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			// 日志记录：获取插件元数据失败
			fmt.Printf("获取插件 %s 元数据失败：%v\n", pid, r)
		}
	}()
	
	var plugins map[string]interface{}
	var err error
	
	if packageVersion == nil || *packageVersion == "" {
		plugins, err = p.GetPlugins(repoURL, nil, false)
	} else {
		plugins, err = p.GetPlugins(repoURL, packageVersion, false)
	}
	
	if err != nil || plugins == nil {
		return make(map[string]interface{})
	}
	
	if meta, exists := plugins[pid]; exists {
		if metaMap, ok := meta.(map[string]interface{}); ok {
			return metaMap
		}
	}
	
	return make(map[string]interface{})
}

// installFromRelease 通过 GitHub Release 资产文件安装插件
func (p *PluginHelper) installFromRelease(pid string, userRepo string, releaseTag string) (bool, string) {
	/*
	 * 通过 GitHub Release 资产文件安装插件�?	 * 规范：release 中存在名�?"{pid}_v{version}.zip" 的资产，zip 根即插件文件�?	 * 将其全部解压�?app/plugins/{pid}
	 */
	// TODO: 实现从Release安装插件的逻辑
	return false, "暂未实现从Release安装插件的功�?
}

// prepareContentViaFilelistSync 同步准备插件内容，通过文件列表获取插件文件和依�?func (p *PluginHelper) prepareContentViaFilelistSync(pid string, userRepo string, packageVersion *string) (bool, string) {
	/*
	 * 同步准备插件内容，通过文件列表获取插件文件和依�?	 */
	// TODO: 实现通过文件列表准备插件内容的逻辑
	return false, "暂未实现通过文件列表准备插件内容的功�?
}

// requestWithFallback 使用自动降级策略，请求资源，优先级依次为镜像站、代理、直�?func (p *PluginHelper) requestWithFallback(url string, headers map[string]string, timeout int, isAPI bool) (*utils.HTTPResponse, error) {
	/*
	 * 使用自动降级策略，请求资源，优先级依次为镜像站、代理、直�?	 * :param url: 目标URL
	 * :param headers: 请求头信�?	 * :param timeout: 请求超时时间
	 * :param isAPI: 是否为GitHub API请求，API请求不走镜像�?	 * :return: 请求成功则返�?Response，失败返�?None
	 */
	 
	// 策略列表
	strategies := make([]struct {
		Name   string
		URL    string
		Params map[string]interface{}
	}, 0)

	// 1. 尝试使用镜像站，镜像站一般不支持API请求，因此API请求直接跳过镜像�?	cfg := config.GetConfig()
	if !isAPI && cfg.GithubProxy != "" {
		proxyURL := fmt.Sprintf("%s%s", strings.TrimRight(cfg.GithubProxy, "/"), url)
		strategies = append(strategies, struct {
			Name   string
			URL    string
			Params map[string]interface{}
		}{
			Name: "镜像�?,
			URL:  proxyURL,
			Params: map[string]interface{}{
				"headers": headers,
				"timeout": timeout,
			},
		})
	}

	// 2. 尝试使用代理
	if cfg.ProxyHost != "" {
		strategies = append(strategies, struct {
			Name   string
			URL    string
			Params map[string]interface{}
		}{
			Name: "代理",
			URL:  url,
			Params: map[string]interface{}{
				"headers": headers,
				"proxies": cfg.Proxy, // 需要实现Proxy配置
				"timeout": timeout,
			},
		})
	}

	// 3. 最后尝试直�?	strategies = append(strategies, struct {
		Name   string
		URL    string
		Params map[string]interface{}
	}{
		Name: "直连",
		URL:  url,
		Params: map[string]interface{}{
			"headers": headers,
			"timeout": timeout,
		},
	})

	// 遍历策略并尝试请�?	for _, strategy := range strategies {
		// 日志记录：尝试使用策略请求URL
		fmt.Printf("[GitHub] 尝试使用策略�?s 请求 URL�?s\n", strategy.Name, strategy.URL)

		// 发起请求
		res, err := utils.RequestUtils.GetRes(strategy.URL, headers, nil, timeout)
		if err == nil && res != nil {
			// 日志记录：请求成�?			fmt.Printf("[GitHub] 请求成功，策略：%s, URL: %s\n", strategy.Name, strategy.URL)
			return res, nil
		}
		
		// 日志记录：请求失�?		fmt.Printf("[GitHub] 请求失败，策略：%s, URL: %s，错误：%v\n", strategy.Name, strategy.URL, err)
	}

	// 日志记录：所有策略均请求失败
	fmt.Printf("[GitHub] 所有策略均请求失败，URL: %s，请检查网络连接或 GitHub 配置\n", url)
	return nil, fmt.Errorf("所有策略均请求失败")
}

// pipInstallWithFallback 使用自动降级策略安装依赖，并确保新安装的包可被动态导�?func (p *PluginHelper) pipInstallWithFallback(requirementsFile string) (bool, string) {
	/*
	 * 使用自动降级策略安装依赖，并确保新安装的包可被动态导�?	 * :param requirementsFile: 依赖�?requirements.txt 文件路径
	 * :return: (是否成功, 错误信息)
	 */
	 
	// 检查requirements文件是否存在
	if _, err := os.Stat(requirementsFile); os.IsNotExist(err) {
		return false, "requirements.txt 文件不存�?
	}
	
	// 获取wheels目录
	wheelsDir := filepath.Join(filepath.Dir(requirementsFile), "wheels")

	// 构建基础命令
	baseCmd := []string{"/usr/bin/python3", "-m", "pip", "install"} // 使用python3作为示例
	
	// 如果wheels目录存在，增�?--find-links 选项
	if _, err := os.Stat(wheelsDir); err == nil {
		// 日志记录：发现插件内嵌的 wheels 目录
		fmt.Printf("[PIP] 发现插件内嵌�?wheels 目录: %s，将优先从本地安装。\n", wheelsDir)
		baseCmd = append(baseCmd, "--find-links", wheelsDir)
	} else {
		// 日志记录：未发现插件内嵌�?wheels 目录
		fmt.Printf("[PIP] 未发现插件内嵌的 wheels 目录，将仅使用在线源。\n")
	}
	
	baseCmd = append(baseCmd, "-r", requirementsFile)
	
	// 添加策略到列表中
	strategies := make([]struct {
		Name string
		Cmd  []string
	}, 0)
	
	cfg := config.GetConfig()
	
	// 添加镜像站策�?	if cfg.PipProxy != "" {
		cmd := append([]string{}, baseCmd...)
		cmd = append(cmd, "-i", cfg.PipProxy)
		strategies = append(strategies, struct {
			Name string
			Cmd  []string
		}{
			Name: "镜像�?,
			Cmd:  cmd,
		})
	}
	
	// 添加代理策略
	if cfg.ProxyHost != "" {
		cmd := append([]string{}, baseCmd...)
		cmd = append(cmd, "--proxy", cfg.ProxyHost)
		strategies = append(strategies, struct {
			Name string
			Cmd  []string
		}{
			Name: "代理",
			Cmd:  cmd,
		})
	}
	
	// 添加直连策略
	strategies = append(strategies, struct {
		Name string
		Cmd  []string
	}{
		Name: "直连",
		Cmd:  baseCmd,
	})

	// 遍历策略进行安装
	for _, strategy := range strategies {
		// 日志记录：尝试使用策略安装依�?		fmt.Printf("[PIP] 尝试使用策略�?s 安装依赖，命令：%s\n", strategy.Name, strings.Join(strategy.Cmd, " "))
		
		// 执行命令
		cmd := exec.Command(strategy.Cmd[0], strategy.Cmd[1:]...)
		output, err := cmd.CombinedOutput()
		
		if err == nil {
			// 日志记录：安装依赖成�?			fmt.Printf("[PIP] 策略�?s 安装依赖成功，输出：%s\n", strategy.Name, string(output))
			return true, string(output)
		} else {
			// 日志记录：安装依赖失�?			fmt.Printf("[PIP] 策略�?s 安装依赖失败，错误信息：%s\n", strategy.Name, string(output))
		}
	}

	return false, "[PIP] 所有策略均安装依赖失败，请检查网络连接或 PIP 配置"
}
