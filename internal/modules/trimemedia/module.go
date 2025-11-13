package trimemedia

import (
	"fmt"
	
	"moviepilot-go/pkg/models"
	"moviepilot-go/pkg/modules"
)

// TrimeMediaModule 飞牛影视模块
type TrimeMediaModule struct {
	*modules.ModuleBase
	*modules.MediaServerBase
}

// NewTrimeMediaModule 创建新的飞牛影视模块实例
func NewTrimeMediaModule() *TrimeMediaModule {
	module := &TrimeMediaModule{
		ModuleBase:       modules.NewModuleBase(),
		MediaServerBase:  modules.NewMediaServerBase(),
	}
	
	// 设置模块属�?	module.Name = "飞牛影视"
	module.Type = models.ModuleTypeMediaServer
	module.SubType = models.MediaServerTypeTrimeMedia
	module.Priority = 4
	
	return module
}

// InitModule 初始化模�?func (t *TrimeMediaModule) InitModule() error {
	// 初始化服�?	t.InitService("trimemedia", func(conf interface{}) interface{} {
		// 类型断言获取配置
		if configMap, ok := conf.(map[string]interface{}); ok {
			// 提取配置参数
			var host, username, password, playHost string
			var syncLibraries []string
			
			if h, ok := configMap["host"].(string); ok {
				host = h
			}
			
			if u, ok := configMap["username"].(string); ok {
				username = u
			}
			
			if p, ok := configMap["password"].(string); ok {
				password = p
			}
			
			if ph, ok := configMap["play_host"].(string); ok {
				playHost = ph
			}
			
			if sl, ok := configMap["sync_libraries"].([]string); ok {
				syncLibraries = sl
			}
			
			return NewTrimeMedia(host, username, password, playHost, syncLibraries)
		}
		return nil
	})
	
	return nil
}

// HandleConfigChanged 处理配置变更事件
func (t *TrimeMediaModule) HandleConfigChanged(eventData *models.ConfigChangeEventData) {
	if eventData == nil {
		return
	}
	
	// 检查是否是媒体服务器配置变�?	if eventData.Key != string(models.SystemConfigKeyMediaServers) {
		return
	}
	
	fmt.Println("配置变更，重新加载飞牛影视模�?..")
	t.InitModule()
}

// GetName 获取模块名称
func (t *TrimeMediaModule) GetName() string {
	return "飞牛影视"
}

// GetType 获取模块类型
func (t *TrimeMediaModule) GetType() models.ModuleType {
	return models.ModuleTypeMediaServer
}

// GetSubType 获取模块的子类型
func (t *TrimeMediaModule) GetSubType() models.MediaServerType {
	return models.MediaServerTypeTrimeMedia
}

// GetPriority 获取模块优先�?func (t *TrimeMediaModule) GetPriority() int {
	return 4
}

// InitSetting 初始化设�?func (t *TrimeMediaModule) InitSetting() (string, interface{}) {
	// 实现初始化设置逻辑
	return "", nil
}

// SchedulerJob 定时任务，每10分钟调用一�?func (t *TrimeMediaModule) SchedulerJob() {
	// 定时重连
	instances := t.GetInstances()
	for name, server := range instances {
		if trimeMedia, ok := server.(*TrimeMedia); ok {
			if trimeMedia.IsConfigured() && trimeMedia.IsInactive() {
				fmt.Printf("飞牛影视 %s 连接断开，尝试重�?...\n", name)
				trimeMedia.Reconnect()
			}
		}
	}
}

// Stop 停止模块
func (t *TrimeMediaModule) Stop() error {
	instances := t.GetInstances()
	for _, server := range instances {
		if trimeMedia, ok := server.(*TrimeMedia); ok {
			if trimeMedia.IsAuthenticated() {
				trimeMedia.Disconnect()
			}
		}
	}
	return nil
}

// Test 测试模块连接�?func (t *TrimeMediaModule) Test() (bool, string) {
	instances := t.GetInstances()
	if len(instances) == 0 {
		return true, ""
	}
	
	for name, server := range instances {
		if trimeMedia, ok := server.(*TrimeMedia); ok {
			if !trimeMedia.IsConfigured() {
				return false, fmt.Sprintf("飞牛影视配置不完整：%s", name)
			}
			
			if trimeMedia.IsInactive() && !trimeMedia.Reconnect() {
				return false, fmt.Sprintf("无法连接飞牛影视�?s", name)
			}
		}
	}
	
	return true, ""
}

// UserAuthenticate 使用飞牛影视用户辅助完成用户认证
func (t *TrimeMediaModule) UserAuthenticate(
	credentials *models.AuthCredentials,
	serviceName string,
) *models.AuthCredentials {
	// 飞牛影视认证
	if credentials == nil || credentials.GrantType != "password" {
		return nil
	}
	
	// 确定要认证的服务器列�?	var servers map[string]interface{}
	if serviceName != "" {
		// 如果指定了服务名，获取该服务实例
		if server := t.GetInstance(&serviceName); server != nil {
			servers = map[string]interface{}{serviceName: server}
		} else {
			servers = make(map[string]interface{})
		}
	} else {
		// 如果没有指定服务名，遍历所有服�?		servers = t.GetInstances()
	}
	
	// 遍历要认证的服务�?	for name, server := range servers {
		if trimeMedia, ok := server.(*TrimeMedia); ok {
			// TODO: 触发认证拦截事件
			/*
			interceptEvent := eventmanager.SendEvent(
				ChainEventTypeAuthIntercept,
				&models.AuthInterceptCredentials{
					Username: credentials.Username,
					Channel:  t.GetName(),
					Service:  name,
					Status:   "triggered",
				},
			)
			
			if interceptEvent != nil && interceptEvent.EventData != nil {
				interceptData := interceptEvent.EventData.(*models.AuthInterceptCredentials)
				if interceptData.Cancel {
					continue
				}
			}
			*/
			
			token := trimeMedia.Authenticate(credentials.Username, credentials.Password)
			if token != "" {
				credentials.Channel = t.GetName()
				credentials.Service = name
				credentials.Token = token
				return credentials
			}
		}
	}
	
	return nil
}

// WebhookParser 解析Webhook报文�?func (t *TrimeMediaModule) WebhookParser(
	body interface{},
	form map[string]interface{},
	args map[string]interface{},
) *models.WebhookEventInfo {
	
	source, _ := args["source"].(string)
	if source != "" {
		server := t.GetInstance(&source)
		if trimeMedia, ok := server.(*TrimeMedia); ok {
			result := trimeMedia.GetWebhookMessage(body)
			if result != nil {
				result.ServerName = source
			}
			return result
		}
		return nil
	}
	
	instances := t.GetInstances()
	for _, server := range instances {
		if trimeMedia, ok := server.(*TrimeMedia); ok {
			result := trimeMedia.GetWebhookMessage(body)
			if result != nil {
				return result
			}
		}
	}
	
	return nil
}

// MediaExists 判断媒体文件是否存在
func (t *TrimeMediaModule) MediaExists(
	mediaInfo *models.MediaInfo,
	itemID string,
	serverName string,
) *models.ExistMediaInfo {
	
	var servers map[string]interface{}
	if serverName != "" {
		servers = map[string]interface{}{serverName: t.GetInstance(&serverName)}
	} else {
		servers = t.GetInstances()
	}
	
	for name, s := range servers {
		if s == nil {
			continue
		}
		
		if trimeMedia, ok := s.(*TrimeMedia); ok {
			if mediaInfo.Type == models.Movie {
				if itemID != "" {
					movie := trimeMedia.GetItemInfo(itemID)
					if movie != nil {
						fmt.Printf("媒体�?%s 中找到了 %v\n", name, movie)
						return &models.ExistMediaInfo{
							Type:        models.Movie,
							ServerType:  "trimemedia",
							Server:      name,
							ItemID:      movie.ItemID,
						}
					}
				}
				
				movies := trimeMedia.GetMovies(mediaInfo.Title, mediaInfo.Year, mediaInfo.TmdbID)
				if len(movies) == 0 {
					fmt.Printf("%s 没有在媒体库 %s 中\n", mediaInfo.TitleYear, name)
					continue
				} else {
					fmt.Printf("媒体�?%s 中找到了 %v\n", name, movies)
					return &models.ExistMediaInfo{
						Type:        models.Movie,
						ServerType:  "trimemedia",
						Server:      name,
						ItemID:      movies[0].ItemID,
					}
				}
			} else {
				foundItemID, tvs := trimeMedia.GetTVEpisodes(itemID, mediaInfo.Title, mediaInfo.Year, mediaInfo.TmdbID, -1)
				if len(tvs) == 0 {
					fmt.Printf("%s 没有在媒体库 %s 中\n", mediaInfo.TitleYear, name)
					continue
				} else {
					fmt.Printf("%s 在媒体库 %s 中找到了这些季集�?v\n", mediaInfo.TitleYear, name, tvs)
					return &models.ExistMediaInfo{
						Type:        models.TV,
						Seasons:     tvs,
						ServerType:  "trimemedia",
						Server:      name,
						ItemID:      foundItemID,
					}
				}
			}
		}
	}
	
	return nil
}

// MediaStatistic 媒体数量统计
func (t *TrimeMediaModule) MediaStatistic(serverName string) []models.Statistic {
	var servers []interface{}
	
	if serverName != "" {
		serverObj := t.GetInstance(&serverName)
		if serverObj == nil {
			return nil
		}
		servers = []interface{}{serverObj}
	} else {
		instances := t.GetInstances()
		servers = make([]interface{}, 0, len(instances))
		for _, s := range instances {
			servers = append(servers, s)
		}
	}
	
	mediaStatistics := make([]models.Statistic, 0)
	for _, s := range servers {
		if trimeMedia, ok := s.(*TrimeMedia); ok {
			mediaStatistic := trimeMedia.GetMediasCount()
			if mediaStatistic == nil {
				continue
			}
			
			mediaStatistic.UserCount = trimeMedia.GetUserCount()
			mediaStatistics = append(mediaStatistics, *mediaStatistic)
		}
	}
	
	return mediaStatistics
}

// MediaServerLibrarys 媒体库列�?func (t *TrimeMediaModule) MediaServerLibrarys(
	serverName string,
	hidden bool,
	kwargs map[string]interface{},
) []models.MediaServerLibrary {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj != nil {
		if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
			return trimeMedia.GetLibrarys(hidden)
		}
	}
	
	return nil
}

// MediaServerItems 获取媒体服务器项目列�?func (t *TrimeMediaModule) MediaServerItems(
	serverName string,
	libraryID string,
	startIndex int,
	limit int,
) []models.MediaServerItem {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj != nil {
		if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
			return trimeMedia.GetItems(libraryID, startIndex, limit)
		}
	}
	
	return nil
}

// MediaServerItemInfo 媒体库项目详�?func (t *TrimeMediaModule) MediaServerItemInfo(
	serverName string,
	itemID string,
) *models.MediaServerItem {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj != nil {
		if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
			return trimeMedia.GetItemInfo(itemID)
		}
	}
	
	return nil
}

// MediaServerTVEpisodes 获取剧集信息
func (t *TrimeMediaModule) MediaServerTVEpisodes(
	serverName string,
	itemID string,
) []models.MediaServerSeasonInfo {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj == nil {
		return nil
	}
	
	if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
		_, seasonInfo := trimeMedia.GetTVEpisodes(itemID, "", "", 0, -1)
		if seasonInfo == nil {
			return []models.MediaServerSeasonInfo{}
		}
		
		result := make([]models.MediaServerSeasonInfo, 0, len(seasonInfo))
		for season, episodes := range seasonInfo {
			result = append(result, models.MediaServerSeasonInfo{
				Season:   season,
				Episodes: episodes,
			})
		}
		
		return result
	}
	
	return nil
}

// MediaServerPlaying 获取媒体服务器正在播放信�?func (t *TrimeMediaModule) MediaServerPlaying(
	serverName string,
	count int,
	kwargs map[string]interface{},
) []models.MediaServerPlayItem {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj == nil {
		return []models.MediaServerPlayItem{}
	}
	
	if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
		result := trimeMedia.GetResume(count)
		if result == nil {
			return []models.MediaServerPlayItem{}
		}
		return result
	}
	
	return []models.MediaServerPlayItem{}
}

// MediaServerPlayURL 获取媒体库播放地址
func (t *TrimeMediaModule) MediaServerPlayURL(
	serverName string,
	itemID string,
) string {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj == nil {
		return ""
	}
	
	if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
		return trimeMedia.GetPlayURL(itemID)
	}
	
	return ""
}

// MediaServerLatest 获取媒体服务器最新入库条�?func (t *TrimeMediaModule) MediaServerLatest(
	serverName string,
	count int,
	kwargs map[string]interface{},
) []models.MediaServerPlayItem {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj == nil {
		return []models.MediaServerPlayItem{}
	}
	
	if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
		result := trimeMedia.GetLatest(count)
		if result == nil {
			return []models.MediaServerPlayItem{}
		}
		return result
	}
	
	return []models.MediaServerPlayItem{}
}

// MediaServerLatestImages 获取媒体服务器最新入库条目的图片
func (t *TrimeMediaModule) MediaServerLatestImages(
	serverName string,
	count int,
	remote bool,
	kwargs map[string]interface{},
) []string {
	
	serverObj := t.GetInstance(&serverName)
	if serverObj == nil {
		return []string{}
	}
	
	if trimeMedia, ok := serverObj.(*TrimeMedia); ok {
		result := trimeMedia.GetLatestBackdrops(count, remote)
		if result == nil {
			return []string{}
		}
		return result
	}
	
	return []string{}
}
