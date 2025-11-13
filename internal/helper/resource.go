package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
)

// ResourceHelper 检测和更新资源�?type ResourceHelper struct {
	repo     string
	filesAPI string
	baseDir  string
}

// ResourceInfo 资源包信�?type ResourceInfo struct {
	Version   string              `json:"version"`
	Resources map[string]Resource `json:"resources"`
}

// Resource 资源信息
type Resource struct {
	Type     string `json:"type"`
	Platform string `json:"platform"`
	Target   string `json:"target"`
	Version  string `json:"version"`
}

// FileInfo 文件信息
type FileInfo struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
}

// NewResourceHelper 创建ResourceHelper实例
func NewResourceHelper() *ResourceHelper {
	githubProxy := config.GetConfig().GITHUB_PROXY
	if githubProxy == "" {
		githubProxy = ""
	} else if githubProxy[len(githubProxy)-1] != '/' {
		githubProxy += "/"
	}

	helper := &ResourceHelper{
		repo:     fmt.Sprintf("%shttps://raw.githubusercontent.com/jxxghp/MoviePilot-Resources/main/package.v2.json", githubProxy),
		filesAPI: "https://api.github.com/repos/jxxghp/MoviePilot-Resources/contents/resources.v2",
		baseDir:  config.GetConfig().ROOT_PATH,
	}

	// 在创建实例时自动检查更�?	helper.Check()
	return helper
}

// getProxies 获取代理设置
func (r *ResourceHelper) getProxies() string {
	cfg := config.GetConfig()
	if cfg.GITHUB_PROXY != "" {
		return ""
	}
	return cfg.PROXY
}

// Check 检测是否有更新，如有则下载安装
func (r *ResourceHelper) Check() error {
	cfg := config.GetConfig()

	if !cfg.AUTO_UPDATE_RESOURCE {
		return nil
	}

	if utils.IsFrozen() {
		return nil
	}

	logger.Info("开始检测资源包版本...")

	// 创建带超时的HTTP客户�?	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 设置请求
	req, err := http.NewRequest("GET", r.repo, nil)
	if err != nil {
		logger.Warn("无法连接资源包仓库！")
		return err
	}

	// 添加请求�?	headers := cfg.GITHUB_HEADERS
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请�?	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("无法连接资源包仓库！")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warnf("无法连接资源包仓�? %s", resp.Status)
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("读取资源包信息失败！")
		return err
	}

	// 解析JSON
	var resourceInfo ResourceInfo
	err = json.Unmarshal(body, &resourceInfo)
	if err != nil {
		logger.Error("资源包仓库数据解析失败！")
		return err
	}

	onlineVersion := resourceInfo.Version
	if onlineVersion != "" {
		logger.Infof("最新资源包版本：v%s", onlineVersion)

		// 需要更新的资源�?		needUpdates := make(map[string]string)

		// 资源明细
		resources := resourceInfo.Resources
		if resources == nil {
			resources = make(map[string]Resource)
		}

		for rname, resource := range resources {
			rtype := resource.Type
			platform := resource.Platform
			target := resource.Target
			version := resource.Version

			// 判断平台
			if platform != "" && platform != runtime.GOOS {
				continue
			}

			// 判断版本�?			var localVersion string
			if rtype == "auth" {
				// 站点认证资源
				localVersion = getSitesAuthVersion()
			} else if rtype == "sites" {
				// 站点索引资源
				localVersion = getSitesIndexerVersion()
			} else {
				continue
			}

			if utils.CompareVersion(version, ">", localVersion) {
				logger.Infof("%s 资源包有更新，最新版本：v%s", rname, version)
			} else {
				continue
			}

			// 需要安�?			needUpdates[rname] = target
		}

		if len(needUpdates) > 0 {
			// 下载文件信息列表
			filesClient := &http.Client{
				Timeout: 30 * time.Second,
			}

			filesReq, err := http.NewRequest("GET", r.filesAPI, nil)
			if err != nil {
				return fmt.Errorf("创建文件列表请求失败: %v", err)
			}

			// 添加请求�?			for key, value := range headers {
				filesReq.Header.Set(key, value)
			}

			filesResp, err := filesClient.Do(filesReq)
			if err != nil {
				return fmt.Errorf("连接仓库失败: %v", err)
			}
			defer filesResp.Body.Close()

			if filesResp.StatusCode != http.StatusOK {
				return fmt.Errorf("连接仓库失败�?s - %s", filesResp.Status, filesResp.Status)
			}

			// 解析文件列表
			var filesInfo []FileInfo
			filesBody, err := io.ReadAll(filesResp.Body)
			if err != nil {
				return fmt.Errorf("读取文件列表失败: %v", err)
			}

			err = json.Unmarshal(filesBody, &filesInfo)
			if err != nil {
				return fmt.Errorf("解析文件列表失败: %v", err)
			}

			// 下载资源文件
			success := true
			for _, item := range filesInfo {
				savePath, exists := needUpdates[item.Name]
				if !exists {
					continue
				}

				if item.DownloadURL != "" {
					logger.Infof("开始更新资源文件：%s ...", item.Name)

					downloadURL := fmt.Sprintf("%s%s", config.GetConfig().GITHUB_PROXY, item.DownloadURL)
					if config.GetConfig().GITHUB_PROXY != "" && config.GetConfig().GITHUB_PROXY[len(config.GetConfig().GITHUB_PROXY)-1] != '/' {
						downloadURL = fmt.Sprintf("%s/%s", config.GetConfig().GITHUB_PROXY, item.DownloadURL)
					}

					// 创建带超时的下载客户�?					downloadClient := &http.Client{
						Timeout: 180 * time.Second,
					}

					downloadReq, err := http.NewRequest("GET", downloadURL, nil)
					if err != nil {
						logger.Errorf("创建下载请求失败�?s", err)
						success = false
						break
					}

					// 添加请求�?					for key, value := range headers {
						downloadReq.Header.Set(key, value)
					}

					// 下载资源文件
					downloadResp, err := downloadClient.Do(downloadReq)
					if err != nil {
						logger.Errorf("文件 %s 下载失败�?, item.Name)
						success = false
						break
					} else if downloadResp.StatusCode != http.StatusOK {
						logger.Errorf("下载文件 %s 失败�?s - %s", item.Name, downloadResp.Status, downloadResp.Status)
						success = false
						downloadResp.Body.Close()
						break
					}

					// 读取文件内容
					content, err := io.ReadAll(downloadResp.Body)
					downloadResp.Body.Close()
					if err != nil {
						logger.Errorf("读取文件内容失败�?s", err)
						success = false
						break
					}

					// 创建插件文件�?					filePath := filepath.Join(r.baseDir, savePath, item.Name)
					dir := filepath.Dir(filePath)
					err = os.MkdirAll(dir, 0755)
					if err != nil {
						logger.Errorf("创建目录失败�?s", err)
						success = false
						break
					}

					// 写入文件
					err = os.WriteFile(filePath, content, 0644)
					if err != nil {
						logger.Errorf("写入文件失败�?s", err)
						success = false
						break
					}
				}
			}

			if success {
				logger.Info("资源包更新完成，开始重启服�?..")
				utils.Restart()
			} else {
				logger.Warn("资源包更新失败，跳过升级�?)
			}
		} else {
			logger.Info("所有资源已最新，无需更新")
		}
	}

	return nil
}

// 模拟获取站点认证版本的方�?func getSitesAuthVersion() string {
	// 这里应该从站点帮助类中获取认证版�?	// 暂时返回默认�?	return "0.0.0"
}

// 模拟获取站点索引版本的方�?func getSitesIndexerVersion() string {
	// 这里应该从站点帮助类中获取索引版�?	// 暂时返回默认�?	return "0.0.0"
}
