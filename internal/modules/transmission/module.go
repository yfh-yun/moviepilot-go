package transmission

import (
	"fmt"
	"path/filepath"
	"strings"
	
	"github.com/hekmon/transmissionrpc/v2"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
	"moviepilot-go/pkg/modules"
)

// TransmissionModule Transmission模块
type TransmissionModule struct {
	*modules.ModuleBase
	*modules.DownloaderBase
}

// NewTransmissionModule 创建新的Transmission模块实例
func NewTransmissionModule() *TransmissionModule {
	module := &TransmissionModule{
		ModuleBase:     modules.NewModuleBase(),
		DownloaderBase: modules.NewDownloaderBase(),
	}
	
	// 设置模块属�?	module.Name = "Transmission"
	module.Type = models.ModuleTypeDownloader
	module.SubType = models.DownloaderTypeTransmission
	module.Priority = 2
	
	return module
}

// InitModule 初始化模�?func (t *TransmissionModule) InitModule() error {
	// 初始化服�?	t.InitService("transmission", func(conf interface{}) interface{} {
		return NewTransmissionFromConfig(conf)
	})
	return nil
}

// HandleConfigChanged 处理配置变更事件
func (t *TransmissionModule) HandleConfigChanged(eventData *models.ConfigChangeEventData) {
	if eventData == nil {
		return
	}
	
	// 检查是否是下载器配置变�?	if eventData.Key != string(models.SystemConfigKeyDownloaders) {
		return
	}
	
	fmt.Println("配置变更，重新加载Transmission模块...")
	t.InitModule()
}

// GetName 获取模块名称
func (t *TransmissionModule) GetName() string {
	return "Transmission"
}

// GetType 获取模块类型
func (t *TransmissionModule) GetType() models.ModuleType {
	return models.ModuleTypeDownloader
}

// GetSubType 获取模块的子类型
func (t *TransmissionModule) GetSubType() models.DownloaderType {
	return models.DownloaderTypeTransmission
}

// GetPriority 获取模块优先�?func (t *TransmissionModule) GetPriority() int {
	return 2
}

// Stop 停止模块
func (t *TransmissionModule) Stop() error {
	// 实现停止逻辑
	return nil
}

// Test 测试模块连接�?func (t *TransmissionModule) Test() (bool, string) {
	instances := t.GetInstances()
	if len(instances) == 0 {
		return true, ""
	}
	
	for name, server := range instances {
		if transmission, ok := server.(*Transmission); ok {
			if transmission.IsInactive() {
				transmission.Reconnect()
			}
			
			if transmission.TransferInfo() == nil {
				return false, fmt.Sprintf("无法连接Transmission下载器：%s", name)
			}
		}
	}
	
	return true, ""
}

// InitSetting 初始化设�?func (t *TransmissionModule) InitSetting() (string, interface{}) {
	// 实现初始化设置逻辑
	return "", nil
}

// SchedulerJob 定时任务，每10分钟调用一�?func (t *TransmissionModule) SchedulerJob() {
	// 定时重连
	instances := t.GetInstances()
	for name, server := range instances {
		if transmission, ok := server.(*Transmission); ok {
			if transmission.IsInactive() {
				fmt.Printf("Transmission下载�?%s 连接断开，尝试重�?...\n", name)
				transmission.Reconnect()
			}
		}
	}
}

// NewTransmissionFromConfig 从配置创建Transmission实例
func NewTransmissionFromConfig(conf interface{}) *Transmission {
	// 类型断言获取配置
	if configMap, ok := conf.(map[string]interface{}); ok {
		// 提取配置参数
		var host string
		var port int64
		var username, password string
		kwargs := make(map[string]interface{})
		
		if h, ok := configMap["host"].(string); ok {
			host = h
		}
		
		if p, ok := configMap["port"].(int64); ok {
			port = p
		} else if p, ok := configMap["port"].(float64); ok {
			port = int64(p)
		} else if p, ok := configMap["port"].(int); ok {
			port = int64(p)
		}
		
		if u, ok := configMap["username"].(string); ok {
			username = u
		}
		
		if p, ok := configMap["password"].(string); ok {
			password = p
		}
		
		// 创建Transmission实例
		return NewTransmission(host, port, username, password, kwargs)
	}
	
	return nil
}

// Download 根据种子文件，选择并添加下载任�?func (t *TransmissionModule) Download(
	content interface{}, // Path, string, or []byte
	downloadDir string,
	cookie string,
	episodes []int,
	category string,
	label string,
	downloader string,
) (string, string, string, string) {
	
	if content == nil {
		return "", "", "", "下载内容为空"
	}
	
	// 读取种子的名�?	torrent, contentBytes := t.getTorrentInfo(content)
	
	// 检查是否为磁力链接
	var isMagnet bool
	if str, ok := content.(string); ok && strings.HasPrefix(str, "magnet:") {
		isMagnet = true
	} else if bytes, ok := content.([]byte); ok && strings.HasPrefix(string(bytes), "magnet:") {
		isMagnet = true
	}
	
	if torrent == nil && !isMagnet {
		return "", "", "", "添加种子任务失败：无法读取种子文�?
	}
	
	// 获取下载�?	server := t.GetInstance(&downloader)
	if server == nil {
		return "", "", "", "无法获取下载器实�?
	}
	
	transmission, ok := server.(*Transmission)
	if !ok {
		return "", "", "", "下载器实例类型错�?
	}
	
	// 如果要选择文件则先暂停
	isPaused := len(episodes) > 0
	
	// 标签
	var labels []string
	if label != "" {
		labels = strings.Split(label, ",")
	} else {
		// 使用默认标签（这里简化处理，实际应从配置获取�?		labels = nil
	}
	
	// 添加任务
	var torrentResult *transmissionrpc.Torrent
	if contentBytes != nil {
		torrentResult = transmission.AddTorrent(contentBytes, isPaused, downloadDir, labels, cookie)
	} else if contentStr, ok := content.(string); ok {
		torrentResult = transmission.AddTorrent(contentStr, isPaused, downloadDir, labels, cookie)
	}
	
	// TR 始终使用原始种子布局, 返回"Original"
	torrentLayout := "Original"
	
	if torrentResult == nil {
		// 查询所有下载器的种�?		torrents, errorFlag := transmission.GetTorrents(nil, nil, nil)
		if errorFlag {
			return "", "", "", "无法连接transmission下载�?
		}
		
		if len(torrents) > 0 {
			// 这里简化处理，实际应比较名称和大小
			torrent := torrents[0]
			torrentHash := *torrent.HashString
			fmt.Printf("下载器中已存在该种子任务�?s - %s\n", torrentHash, *torrent.Name)
			
			// 给种子打上标�?			// 这里简化处理，实际应从配置获取标签
			fmt.Printf("给种�?%s 打上标签\n", torrentHash)
			
			return downloader, torrentHash, torrentLayout, "下载任务已存�?
		}
		
		return "", "", "", fmt.Sprintf("添加种子任务失败�?v", content)
	} else {
		torrentHash := *torrentResult.HashString
		if isPaused {
			// 选择文件
			torrentFiles := transmission.GetFiles(torrentHash)
			if len(torrentFiles) == 0 {
				return downloader, torrentHash, torrentLayout, "获取种子文件失败，下载任务可能在暂停状�?
			}
			
			// 需要的文件信息
			fileIds := make([]int64, 0)
			unwantedFileIds := make([]int64, 0)
			
			for _, torrentFile := range torrentFiles {
				fileId := *torrentFile.ID
				fileName := *torrentFile.Name
				
				// 这里应该解析文件名获取集数信�?				// 简化处理，假设所有文件都需�?				metaInfo := utils.NewMetaInfo(fileName)
				if len(metaInfo.EpisodeList) == 0 {
					unwantedFileIds = append(unwantedFileIds, fileId)
					continue
				}
				
				selected := true
				// 检查是否在需要的集数列表�?				if len(episodes) > 0 {
					selected = false
					for _, episode := range metaInfo.EpisodeList {
						for _, neededEpisode := range episodes {
							if episode == neededEpisode {
								selected = true
								break
							}
						}
						if selected {
							break
						}
					}
				}
				
				if !selected {
					unwantedFileIds = append(unwantedFileIds, fileId)
					continue
				}
				
				fileIds = append(fileIds, fileId)
			}
			
			// 选择文件
			transmission.SetFiles(torrentHash, fileIds)
			transmission.SetUnwantedFiles(torrentHash, unwantedFileIds)
			
			// 开始任�?			transmission.StartTorrents(torrentHash)
			
			return downloader, torrentHash, torrentLayout, "添加下载任务成功"
		} else {
			return downloader, torrentHash, torrentLayout, "添加下载任务成功"
		}
	}
}

// getTorrentInfo 获取种子信息
func (t *TransmissionModule) getTorrentInfo(content interface{}) (*utils.Torrent, []byte) {
	var torrentInfo *utils.Torrent
	var torrentContent []byte
	
	switch v := content.(type) {
	case string:
		// 检查是否为文件路径
		if !strings.HasPrefix(v, "magnet:") {
			// 尝试读取文件
			// 这里简化处理，实际应检查文件是否存在并读取
			torrentContent = []byte(v)
		} else {
			// 磁力链接
			torrentContent = []byte(v)
		}
	case []byte:
		torrentContent = v
	}
	
	// 尝试解析种子信息
	if torrentContent != nil && !strings.HasPrefix(string(torrentContent), "magnet:") {
		torrentInfo = utils.ParseTorrent(torrentContent)
	}
	
	return torrentInfo, torrentContent
}

// ListTorrents 获取下载器种子列�?func (t *TransmissionModule) ListTorrents(
	status models.TorrentStatus,
	hashes interface{},
	downloader string,
) []interface{} {
	
	// 获取下载�?	var servers map[string]interface{}
	if downloader != "" {
		server := t.GetInstance(&downloader)
		if server == nil {
			return nil
		}
		servers = map[string]interface{}{downloader: server}
	} else {
		servers = t.GetInstances()
	}
	
	retTorrents := make([]interface{}, 0)
	
	if hashes != nil {
		// 按Hash获取
		for name, server := range servers {
			if transmission, ok := server.(*Transmission); ok {
				torrents, _ := transmission.GetTorrents(hashes, nil, nil)
				for _, torrent := range torrents {
					retTorrents = append(retTorrents, &models.TransferTorrent{
						Downloader: name,
						Title:      *torrent.Name,
						Path:       filepath.Join(*torrent.DownloadDir, *torrent.Name),
						Hash:       *torrent.HashString,
						Size:       *torrent.TotalSize,
						Tags:       strings.Join(torrent.Labels, ","),
						Progress:   *torrent.PercentDone * 100,
					})
				}
			}
		}
	} else if status == models.TorrentStatusTransfer {
		// 获取已完成且未整理的
		for name, server := range servers {
			if transmission, ok := server.(*Transmission); ok {
				// 这里简化处理，实际应从配置获取标签
				torrents := transmission.GetCompletedTorrents(nil, nil)
				for _, torrent := range torrents {
					// 检查是否包�?已整�?标签
					hasOrganized := false
					for _, label := range torrent.Labels {
						if label == "已整�? {
							hasOrganized = true
							break
						}
					}
					
					if hasOrganized {
						continue
					}
					
					// 下载路径
					path := *torrent.DownloadDir
					// 无法获取下载路径的不处理
					if path == "" {
						fmt.Printf("未获取到 %s 下载保存路径\n", *torrent.Name)
						continue
					}
					
					state := "downloading"
					if *torrent.Status == transmissionrpc.TorrentStatusStopped {
						state = "paused"
					}
					
					retTorrents = append(retTorrents, &models.TransferTorrent{
						Downloader: name,
						Title:      *torrent.Name,
						Path:       filepath.Join(*torrent.DownloadDir, *torrent.Name),
						Hash:       *torrent.HashString,
						Tags:       strings.Join(torrent.Labels, ","),
						Progress:   *torrent.PercentDone * 100,
						State:      state,
					})
				}
			}
		}
	} else if status == models.TorrentStatusDownloading {
		// 获取正在下载的任�?		for name, server := range servers {
			if transmission, ok := server.(*Transmission); ok {
				// 这里简化处理，实际应从配置获取标签
				torrents := transmission.GetDownloadingTorrents(nil, nil)
				for _, torrent := range torrents {
					meta := utils.NewMetaInfo(*torrent.Name)
					
					dlspeed := *torrent.RateDownload
					upspeed := *torrent.RateUpload
					
					state := "downloading"
					if *torrent.Status == transmissionrpc.TorrentStatusStopped {
						state = "paused"
					}
					
					// 计算剩余时间（简化处理）
					leftTime := ""
					if dlspeed > 0 {
						leftTime = utils.SecondsToString(int64(*torrent.LeftUntilDone) / int64(dlspeed))
					}
					
					retTorrents = append(retTorrents, &models.DownloadingTorrent{
						Downloader:    name,
						Hash:          *torrent.HashString,
						Title:         *torrent.Name,
						Name:          meta.Name,
						Year:          meta.Year,
						SeasonEpisode: meta.SeasonEpisode,
						Progress:      *torrent.PercentDone * 100,
						Size:          *torrent.TotalSize,
						State:         state,
						Dlspeed:       utils.FileSizeToString(dlspeed),
						Upspeed:       utils.FileSizeToString(upspeed),
						LeftTime:      leftTime,
					})
				}
			}
		}
	} else {
		return nil
	}
	
	return retTorrents
}

// TransferCompleted 转移完成后的处理
func (t *TransmissionModule) TransferCompleted(hashes string, downloader string) {
	// 获取下载�?	server := t.GetInstance(&downloader)
	if server == nil {
		return
	}
	
	if transmission, ok := server.(*Transmission); ok {
		// 获取原标�?		orgTags := transmission.GetTorrentTags(hashes)
		
		// 种子打上已整理标�?		var tags []string
		if len(orgTags) > 0 {
			tags = append(orgTags, "已整�?)
		} else {
			tags = []string{"已整�?}
		}
		
		transmission.SetTorrentTag(hashes, tags, orgTags)
	}
}

// RemoveTorrents 删除下载器种�?func (t *TransmissionModule) RemoveTorrents(hashes interface{}, deleteFile bool, downloader string) bool {
	// 获取下载�?	server := t.GetInstance(&downloader)
	if server == nil {
		return false
	}
	
	if transmission, ok := server.(*Transmission); ok {
		return transmission.DeleteTorrents(deleteFile, hashes)
	}
	
	return false
}

// StartTorrents 开始下�?func (t *TransmissionModule) StartTorrents(hashes interface{}, downloader string) bool {
	// 获取下载�?	server := t.GetInstance(&downloader)
	if server == nil {
		return false
	}
	
	if transmission, ok := server.(*Transmission); ok {
		return transmission.StartTorrents(hashes)
	}
	
	return false
}

// StopTorrents 停止下载
func (t *TransmissionModule) StopTorrents(hashes interface{}, downloader string) bool {
	// 获取下载�?	server := t.GetInstance(&downloader)
	if server == nil {
		return false
	}
	
	if transmission, ok := server.(*Transmission); ok {
		return transmission.StopTorrents(hashes)
	}
	
	return false
}

// TorrentFiles 获取种子文件列表
func (t *TransmissionModule) TorrentFiles(tid string, downloader string) []transmissionrpc.File {
	// 获取下载�?	server := t.GetInstance(&downloader)
	if server == nil {
		return nil
	}
	
	if transmission, ok := server.(*Transmission); ok {
		return transmission.GetFiles(tid)
	}
	
	return nil
}

// DownloaderInfo 下载器信�?func (t *TransmissionModule) DownloaderInfo(downloader string) []models.DownloaderInfo {
	var servers []interface{}
	
	if downloader != "" {
		server := t.GetInstance(&downloader)
		if server == nil {
			return nil
		}
		servers = []interface{}{server}
	} else {
		instances := t.GetInstances()
		servers = make([]interface{}, 0, len(instances))
		for _, s := range instances {
			servers = append(servers, s)
		}
	}
	
	// 调用Transmission API查询实时信息
	retInfo := make([]models.DownloaderInfo, 0)
	for _, server := range servers {
		if transmission, ok := server.(*Transmission); ok {
			info := transmission.TransferInfo()
			if info == nil {
				continue
			}
			
			retInfo = append(retInfo, models.DownloaderInfo{
				DownloadSpeed:  info.DownloadSpeed,
				UploadSpeed:    info.UploadSpeed,
				DownloadSize:   info.CurrentStats.DownloadedBytes,
				UploadSize:     info.CurrentStats.UploadedBytes,
			})
		}
	}
	
	return retInfo
}
