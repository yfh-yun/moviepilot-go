package qbittorrent

import (
	"fmt"
	"path/filepath"
	
	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// QbittorrentModule Qbittorrent模块
type QbittorrentModule struct {
	modules.ModuleBase
	modules.DownloaderBase[*Qbittorrent]
}

// NewQbittorrentModule 创建Qbittorrent模块实例
func NewQbittorrentModule() *QbittorrentModule {
	qm := &QbittorrentModule{}
	
	// 初始化模块服�?	// TODO: 实现完整的初始化逻辑
	// qm.InitService("qbittorrent", &Qbittorrent{})
	
	return qm
}

// InitModule 初始化模�?func (qm *QbittorrentModule) InitModule() error {
	// 初始化模�?	// super().init_service(service_name=Qbittorrent.__name__.lower(), service_type=Qbittorrent)
	return nil
}

// GetName 获取模块名称
func (qm *QbittorrentModule) GetName() string {
	return "Qbittorrent"
}

// GetType 获取模块类型
func (qm *QbittorrentModule) GetType() models.ModuleType {
	return models.ModuleTypeDownloader
}

// GetSubtype 获取模块子类�?func (qm *QbittorrentModule) GetSubtype() string {
	return "Qbittorrent"
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (qm *QbittorrentModule) GetPriority() int {
	return 1
}

// Stop 停止模块
func (qm *QbittorrentModule) Stop() {
	// 停止模块逻辑
}

// Test 测试模块连接�?func (qm *QbittorrentModule) Test() (*bool, string) {
	instances := qm.GetInstances()
	if len(instances) == 0 {
		return nil, ""
	}
	
	for name, server := range instances {
		if server.IsInactive() {
			server.Reconnect()
		}
		
		if server.TransferInfo() == nil {
			result := false
			return &result, fmt.Sprintf("无法连接Qbittorrent下载器：%s", name)
		}
	}
	
	result := true
	return &result, ""
}

// InitSetting 初始化设�?func (qm *QbittorrentModule) InitSetting() (string, interface{}) {
	return "", nil
}

// SchedulerJob 定时任务，每10分钟调用一�?func (qm *QbittorrentModule) SchedulerJob() {
	instances := qm.GetInstances()
	for name, server := range instances {
		if server.IsInactive() {
			utils.Log.Infof("Qbittorrent下载�?%s 连接断开，尝试重�?...", name)
			server.Reconnect()
		}
	}
}

// Download 根据种子文件，选择并添加下载任�?func (qm *QbittorrentModule) Download(
	content interface{},
	downloadDir string,
	cookie string,
	episodes []int,
	category *string,
	label *string,
	downloader *string,
) (*models.DownloadResult, error) {
	
	if content == nil {
		return nil, fmt.Errorf("下载内容为空")
	}
	
	// 获取下载�?	server := qm.GetInstance(downloader)
	if server == nil {
		return nil, fmt.Errorf("无法获取下载器实�?)
	}
	
	// 生成随机Tag
	tag := utils.StringUtils.GenerateRandomStr(10)
	var tags []string
	
	if label != nil && *label != "" {
		tags = append([]string{*label}, tag)
	} else if utils.Config.TORRENT_TAG != "" {
		tags = []string{tag, utils.Config.TORRENT_TAG}
	} else {
		tags = []string{tag}
	}
	
	// 如果要选择文件则先暂停
	isPaused := len(episodes) > 0
	
	// 添加任务
	kwargs := map[string]interface{}{
		"ignore_category_check": false,
	}
	
	state := server.AddTorrent(content, isPaused, &downloadDir, tags, category, &cookie, kwargs)
	
	// 获取种子内容布局: `Original: 原始, Subfolder: 创建子文件夹, NoSubfolder: 不创建子文件夹`
	torrentLayout := server.GetContentLayout()
	
	if !state {
		// 查询所有下载器的种�?		torrents, error := server.GetTorrents(nil, "", utils.Config.TORRENT_TAG)
		if error {
			return nil, fmt.Errorf("无法连接qbittorrent下载�?)
		}
		
		if len(torrents) > 0 {
			// TODO: 实现种子检查逻辑
			utils.Log.Warn("下载器中已存在该种子任务")
			return nil, fmt.Errorf("下载任务已存�?)
		}
		
		return nil, fmt.Errorf("添加种子任务失败�?v", content)
	} else {
		// 获取种子Hash
		torrentHash := server.GetTorrentIdByTag(tags, "")
		if torrentHash == nil {
			return nil, fmt.Errorf("下载任务添加成功，但获取Qbittorrent任务信息失败�?v", content)
		} else {
			if isPaused {
				// 种子文件
				torrentFiles := server.GetFiles(*torrentHash)
				if torrentFiles == nil {
					return &models.DownloadResult{
						Downloader:   downloader,
						Hash:         torrentHash,
						Layout:       torrentLayout,
						ErrorMessage: "获取种子文件失败，下载任务可能在暂停状�?,
					}, nil
				}
				
				// 不需要的文件ID
				var fileIds []interface{}
				// 需要的集清�?				var successEpisodes []int
				
				// TODO: 实现文件选择逻辑
				
				if len(successEpisodes) > 0 && len(fileIds) > 0 {
					// 选择文件
					server.SetFiles(map[string]interface{}{
						"torrent_hash": *torrentHash,
						"file_ids":     fileIds,
						"priority":     0,
					})
				}
				
				// 开始任�?				if server.IsForceResume() {
					// 强制继续
					server.TorrentsSetForceStart(*torrentHash)
				} else {
					server.StartTorrents(*torrentHash)
				}
				
				return &models.DownloadResult{
					Downloader:   downloader,
					Hash:         torrentHash,
					Layout:       torrentLayout,
					ErrorMessage: fmt.Sprintf("添加下载成功，已选择集数�?v", successEpisodes),
				}, nil
			} else {
				if server.IsForceResume() {
					server.TorrentsSetForceStart(*torrentHash)
				}
				return &models.DownloadResult{
					Downloader:   downloader,
					Hash:         torrentHash,
					Layout:       torrentLayout,
					ErrorMessage: "添加下载成功",
				}, nil
			}
		}
	}
}

// ListTorrents 获取下载器种子列�?func (qm *QbittorrentModule) ListTorrents(
	status models.TorrentStatus,
	hashes interface{},
	downloader *string,
) []interface{} {
	
	// 获取下载�?	var servers map[string]*Qbittorrent
	if downloader != nil {
		server := qm.GetInstance(downloader)
		if server == nil {
			return nil
		}
		servers = map[string]*Qbittorrent{*downloader: server}
	} else {
		servers = qm.GetInstances()
	}
	
	retTorrents := make([]interface{}, 0)
	
	if hashes != nil {
		// 按Hash获取
		for name, server := range servers {
			torrents, _ := server.GetTorrents(hashes, "", utils.Config.TORRENT_TAG)
			for _, torrent := range torrents {
				// TODO: 实现转换逻辑
				_ = name
				_ = torrent
			}
		}
	} else if status == models.TorrentStatusTransfer {
		// 获取已完成且未整理的
		for name, server := range servers {
			torrents := server.GetCompletedTorrents(nil, utils.Config.TORRENT_TAG)
			for _, torrent := range torrents {
				// TODO: 实现转换逻辑
				_ = name
				_ = torrent
			}
		}
	} else if status == models.TorrentStatusDownloading {
		// 获取正在下载的任�?		for name, server := range servers {
			torrents := server.GetDownloadingTorrents(nil, utils.Config.TORRENT_TAG)
			for _, torrent := range torrents {
				// TODO: 实现转换逻辑
				_ = name
				_ = torrent
			}
		}
	} else {
		return nil
	}
	
	return retTorrents
}

// TransferCompleted 转移完成后的处理
func (qm *QbittorrentModule) TransferCompleted(hashes string, downloader *string) {
	server := qm.GetInstance(downloader)
	if server == nil {
		return
	}
	server.SetTorrentsTag(hashes, []string{"已整�?})
}

// RemoveTorrents 删除下载器种�?func (qm *QbittorrentModule) RemoveTorrents(hashes interface{}, deleteFile *bool, downloader *string) *bool {
	server := qm.GetInstance(downloader)
	if server == nil {
		return nil
	}
	
	deleteFlag := true
	if deleteFile != nil {
		deleteFlag = *deleteFile
	}
	
	result := server.DeleteTorrents(deleteFlag, hashes)
	return &result
}

// StartTorrents 开始下�?func (qm *QbittorrentModule) StartTorrents(hashes interface{}, downloader *string) *bool {
	server := qm.GetInstance(downloader)
	if server == nil {
		return nil
	}
	result := server.StartTorrents(hashes)
	return &result
}

// StopTorrents 停止下载
func (qm *QbittorrentModule) StopTorrents(hashes interface{}, downloader *string) *bool {
	server := qm.GetInstance(downloader)
	if server == nil {
		return nil
	}
	result := server.StopTorrents(hashes)
	return &result
}

// TorrentFiles 获取种子文件列表
func (qm *QbittorrentModule) TorrentFiles(tid string, downloader *string) interface{} {
	server := qm.GetInstance(downloader)
	if server == nil {
		return nil
	}
	return server.GetFiles(tid)
}

// DownloaderInfo 下载器信�?func (qm *QbittorrentModule) DownloaderInfo(downloader *string) []models.DownloaderInfo {
	var servers []*Qbittorrent
	
	if downloader != nil {
		server := qm.GetInstance(downloader)
		if server == nil {
			return nil
		}
		servers = []*Qbittorrent{server}
	} else {
		instanceValues := qm.GetInstances()
		for _, server := range instanceValues {
			servers = append(servers, server)
		}
	}
	
	// 调用Qbittorrent API查询实时信息
	retInfo := make([]models.DownloaderInfo, 0)
	for _, server := range servers {
		info := server.TransferInfo()
		if info == nil {
			continue
		}
		
		// TODO: 实现信息转换逻辑
		_ = info
		
		downloaderInfo := models.DownloaderInfo{
			DownloadSpeed: 0,
			UploadSpeed:   0,
			DownloadSize:  0,
			UploadSize:    0,
		}
		retInfo = append(retInfo, downloaderInfo)
	}
	
	return retInfo
}
