package subtitle

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/utils"
)

// SubtitleModule 字幕下载模块
type SubtitleModule struct {
	// 站点详情页字幕下载链接识别XPATH
	siteSubtitleXpath []string
}

// NewSubtitleModule 创建字幕模块实例
func NewSubtitleModule() *SubtitleModule {
	return &SubtitleModule{
		siteSubtitleXpath: []string{
			`//td[@class="rowhead"][text()="字幕"]/following-sibling::td//a/@href`,
		},
	}
}

// GetName 获取模块名称
func (s *SubtitleModule) GetName() string {
	return "站点字幕"
}

// GetType 获取模块类型
func (s *SubtitleModule) GetType() string {
	return "Other"
}

// GetSubtype 获取模块子类�?func (s *SubtitleModule) GetSubtype() string {
	return "Subtitle"
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (s *SubtitleModule) GetPriority() int {
	return 0
}

// Stop 停止模块
func (s *SubtitleModule) Stop() {
	// 不需要特殊处�?}

// Test 测试模块
func (s *SubtitleModule) Test() (bool, string) {
	// 简单测试，始终返回成功
	return true, ""
}

// DownloadAdded 添加下载任务成功后，从站点下载字幕，保存到下载目�?func (s *SubtitleModule) DownloadAdded(context *utils.Context, downloadDir string, torrentContent interface{}) {
	if !config.Config.DOWNLOAD_SUBTITLE {
		return
	}

	// 没有种子文件不处�?	if torrentContent == nil {
		return
	}

	// 没有详情页不处理
	torrent := context.TorrentInfo
	if torrent.PageUrl == "" {
		return
	}

	// 字幕下载目录
	utils.Log.Infof("开始从站点下载字幕�?s", torrent.PageUrl)

	// 获取种子信息
	folderName, _ := utils.TorrentHelper.GetFileinfoFromTorrentContent(torrentContent)
	
	// 文件保存目录，如果是单文件种子，则folderName是空，此时文件保存目录就是下载目�?	downloadPath := filepath.Join(downloadDir, folderName)
	
	// 等待目录存在
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(downloadPath); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	
	// 目录仍然不存在，且有文件夹名，则创建目录
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) && folderName != "" {
		err := os.MkdirAll(downloadPath, 0755)
		if err != nil {
			utils.Log.Errorf("创建目录失败�?v", err)
			return
		}
	}
	
	// 读取网站代码
	request := utils.RequestUtils.NewRequestUtils(torrent.SiteCookie, torrent.SiteUa, "", nil)
	res, err := request.GetRes(torrent.PageUrl, nil, nil, 30)
	if err != nil {
		utils.Log.Warnf("连接 %s 失败�?v", torrent.PageUrl, err)
		return
	}
	
	if res != nil && res.StatusCode == 200 {
		if len(res.Body) == 0 {
			utils.Log.Warnf("读取页面代码失败�?s", torrent.PageUrl)
			return
		}
		
		// 解析HTML
		doc, err := htmlquery.Parse(strings.NewReader(string(res.Body)))
		if err != nil {
			utils.Log.Errorf("解析HTML失败�?v", err)
			return
		}
		
		// 查找字幕链接
		sublinkList := make([]string, 0)
		for _, xpath := range s.siteSubtitleXpath {
			nodes, err := htmlquery.QueryAll(doc, xpath)
			if err != nil {
				utils.Log.Debugf("XPath查询失败�?s, 错误�?v", xpath, err)
				continue
			}
			
			for _, node := range nodes {
				sublink := node.InnerText()
				if sublink == "" {
					continue
				}
				
				if !strings.HasPrefix(sublink, "http") {
					baseURL := utils.StringUtils.GetBaseURL(torrent.PageUrl)
					if strings.HasPrefix(sublink, "/") {
						sublink = fmt.Sprintf("%s%s", baseURL, sublink)
					} else {
						sublink = fmt.Sprintf("%s/%s", baseURL, sublink)
					}
				}
				sublinkList = append(sublinkList, sublink)
			}
		}
		
		// 下载所有字幕文�?		for _, sublink := range sublinkList {
			utils.Log.Infof("找到字幕下载链接�?s，开始下�?..", sublink)
			
			// 下载
			ret, err := request.GetRes(sublink, nil, nil, 30)
			if err != nil {
				utils.Log.Errorf("下载字幕文件失败�?s, 错误�?v", sublink, err)
				continue
			}
			
			if ret != nil && ret.StatusCode == 200 {
				// 保存文件
				fileName := utils.TorrentHelper.GetUrlFilename(ret, sublink)
				if fileName == "" {
					utils.Log.Warnf("链接不是字幕文件�?s", sublink)
					continue
				}
				
				if strings.ToLower(filepath.Ext(fileName)) == ".zip" {
					// ZIP�?					zipFile := filepath.Join(config.Config.TEMP_PATH, fileName)
					
					// 保存ZIP文件
					err := os.WriteFile(zipFile, ret.Body, 0644)
					if err != nil {
						utils.Log.Errorf("保存ZIP文件失败�?v", err)
						continue
					}
					
					// 解压路径
					zipPath := filepath.Join(config.Config.TEMP_PATH, strings.TrimSuffix(fileName, filepath.Ext(fileName)))
					
					// 解压文件
					err = s.unzipFile(zipFile, zipPath)
					if err != nil {
						utils.Log.Errorf("解压ZIP文件失败�?v", err)
						// 清理临时文件
						os.Remove(zipFile)
						continue
					}
					
					// 遍历转移文件
					err = filepath.Walk(zipPath, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						
						if !info.IsDir() && s.isSubtitleFile(path) {
							targetSubFile := filepath.Join(downloadPath, info.Name())
							if _, err := os.Stat(targetSubFile); err == nil {
								utils.Log.Infof("字幕文件已存在：%s", targetSubFile)
								return nil
							}
							
							utils.Log.Infof("转移字幕 %s �?%s ...", path, targetSubFile)
							err := utils.SystemUtils.Copy(path, targetSubFile)
							if err != nil {
								utils.Log.Errorf("复制文件失败�?v", err)
							}
						}
						return nil
					})
					
					if err != nil {
						utils.Log.Errorf("遍历解压文件失败�?v", err)
					}
					
					// 删除临时文件
					err = os.RemoveAll(zipPath)
					if err != nil {
						utils.Log.Errorf("删除临时解压目录失败�?v", err)
					}
					
					err = os.Remove(zipFile)
					if err != nil {
						utils.Log.Errorf("删除临时ZIP文件失败�?v", err)
					}
				} else {
					subFile := filepath.Join(config.Config.TEMP_PATH, fileName)
					
					// 保存文件
					err := os.WriteFile(subFile, ret.Body, 0644)
					if err != nil {
						utils.Log.Errorf("保存字幕文件失败�?v", err)
						continue
					}
					
					targetSubFile := filepath.Join(downloadPath, fileName)
					utils.Log.Infof("转移字幕 %s �?%s", subFile, targetSubFile)
					err = utils.SystemUtils.Copy(subFile, targetSubFile)
					if err != nil {
						utils.Log.Errorf("复制文件失败�?v", err)
					}
				}
			} else {
				utils.Log.Errorf("下载字幕文件失败�?s", sublink)
				continue
			}
		}
		
		if len(sublinkList) > 0 {
			utils.Log.Infof("%s 页面字幕下载完成", torrent.PageUrl)
		} else {
			utils.Log.Warnf("%s 页面未找到字幕下载链�?, torrent.PageUrl)
		}
	} else if res != nil {
		utils.Log.Warnf("连接 %s 失败，状态码�?d", torrent.PageUrl, res.StatusCode)
	} else {
		utils.Log.Warnf("无法打开链接�?s", torrent.PageUrl)
	}
}

// unzipFile 解压ZIP文件
func (s *SubtitleModule) unzipFile(src, dest string) error {
	// 创建解压目录
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	// 打开ZIP文件
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 遍历ZIP文件中的每个文件
	for _, file := range reader.File {
		// 构建解压后的文件路径
		filePath := filepath.Join(dest, file.Name)

		// 检查是否是目录
		if file.FileInfo().IsDir() {
			// 创建目录
			err := os.MkdirAll(filePath, file.Mode())
			if err != nil {
				return err
			}
			continue
		}

		// 创建文件
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// 打开ZIP中的文件
		srcFile, err := file.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 创建目标文件
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// 复制文件内容
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// isSubtitleFile 判断是否为字幕文�?func (s *SubtitleModule) isSubtitleFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, subExt := range config.Config.RMT_SUBEXT {
		if ext == strings.ToLower(subExt) {
			return true
		}
	}
	return false
}
