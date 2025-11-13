package emby

import (
	"moviepilot-go/internal/core/event"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
)

// EmbyModule Emby模块
type EmbyModule struct {
	instances map[string]*Emby
}

// NewEmbyModule 创建EmbyModule实例
func NewEmbyModule() *EmbyModule {
	return &EmbyModule{
		instances: make(map[string]*Emby),
	}
}

// InitModule 初始化模�?func (e *EmbyModule) InitModule() error {
	// 初始化模块逻辑
	logger.Info("初始化Emby模块...")
	// 这里应该从配置中读取Emby服务器配置并初始化实�?	return nil
}

// GetName 获取模块名称
func (e *EmbyModule) GetName() string {
	return "Emby"
}

// GetType 获取模块类型
func (e *EmbyModule) GetType() types.ModuleType {
	return types.ModuleTypeMediaServer
}

// GetSubtype 获取模块子类�?func (e *EmbyModule) GetSubtype() types.MediaServerType {
	return types.MediaServerTypeEmby
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (e *EmbyModule) GetPriority() int {
	return 1
}

// Stop 停止模块
func (e *EmbyModule) Stop() {
	// 停止模块逻辑
}

// Test 测试模块连接�?func (e *EmbyModule) Test() (bool, string) {
	if len(e.instances) == 0 {
		return false, "没有配置Emby服务�?
	}
	
	for name, server := range e.instances {
		if server.IsInactive() {
			server.Reconnect()
		}
		
		if server.GetUser("") == nil {
			return false, "无法连接Emby服务器：" + name
		}
	}
	
	return true, ""
}

// InitSetting 初始化设�?func (e *EmbyModule) InitSetting() (string, interface{}) {
	return "", nil
}

// SchedulerJob 定时任务，每10分钟调用一�?func (e *EmbyModule) SchedulerJob() {
	// 定时重连
	for name, server := range e.instances {
		if server.IsInactive() {
			logger.Infof("Emby服务�?%s 连接断开，尝试重�?...", name)
			server.Reconnect()
		}
	}
}

// UserAuthenticate 使用Emby用户辅助完成用户认证
func (e *EmbyModule) UserAuthenticate(credentials *schemas.AuthCredentials, serviceName string) *schemas.AuthCredentials {
	// Emby认证
	if credentials == nil || credentials.GrantType != "password" {
		return nil
	}
	
	// 确定要认证的服务器列�?	var servers map[string]*Emby
	if serviceName != "" {
		// 如果指定了服务名，获取该服务实例
		if server, exists := e.instances[serviceName]; exists {
			servers = map[string]*Emby{serviceName: server}
		} else {
			servers = make(map[string]*Emby)
		}
	} else {
		// 如果没有指定服务名，使用所有服�?		servers = e.instances
	}
	
	// 遍历要认证的服务�?	for name, server := range servers {
		// 触发认证拦截事件
		interceptEvent := event.SendEvent(
			types.ChainEventTypeAuthIntercept,
			&schemas.AuthInterceptCredentials{
				Username: credentials.Username,
				Channel:  e.GetName(),
				Service:  name,
				Status:   "triggered",
			},
		)
		
		if interceptEvent != nil && interceptEvent.Data != nil {
			interceptData := interceptEvent.Data.(*schemas.AuthInterceptCredentials)
			if interceptData.Cancel {
				continue
			}
		}
		
		token := server.Authenticate(credentials.Username, credentials.Password)
		if token != "" {
			credentials.Channel = e.GetName()
			credentials.Service = name
			credentials.Token = token
			return credentials
		}
	}
	
	return nil
}

// WebhookParser 解析Webhook报文�?func (e *EmbyModule) WebhookParser(body, form, args interface{}) *schemas.WebhookEventInfo {
	// 类型断言
	formMap, _ := form.(map[string]interface{})
	argsMap, _ := args.(map[string]interface{})
	
	source, _ := argsMap["source"].(string)
	if source != "" {
		if server, exists := e.instances[source]; exists {
			result := server.GetWebhookMessage(formMap, argsMap)
			if result != nil {
				result.ServerName = source
			}
			return result
		}
		return nil
	}
	
	for _, server := range e.instances {
		if server != nil {
			result := server.GetWebhookMessage(formMap, argsMap)
			if result != nil {
				return result
			}
		}
	}
	
	return nil
}

// MediaExists 判断媒体文件是否存在
func (e *EmbyModule) MediaExists(mediainfo *models.MediaInfo, itemid, serverName string) *schemas.ExistMediaInfo {
	var servers map[string]*Emby
	
	if serverName != "" {
		if server, exists := e.instances[serverName]; exists {
			servers = map[string]*Emby{serverName: server}
		} else {
			servers = make(map[string]*Emby)
		}
	} else {
		servers = e.instances
	}
	
	for name, s := range servers {
		if s == nil {
			continue
		}
		
		if mediainfo.Type == types.MediaTypeMovie {
			if itemid != "" {
				movie := s.GetItemInfo(itemid)
				if movie != nil {
					logger.Infof("媒体�?%s 中找到了 %v", name, movie)
					return &schemas.ExistMediaInfo{
						Type:       types.MediaTypeMovie,
						ServerType: "emby",
						Server:     name,
						ItemID:     movie.ItemID,
					}
				}
			}
			
			movies := s.GetMovies(mediainfo.Title, mediainfo.Year, &mediainfo.TmdbID)
			if len(movies) == 0 {
				logger.Infof("%s 没有在媒体库 %s �?, mediainfo.TitleYear, name)
				continue
			} else {
				logger.Infof("媒体�?%s 中找到了 %v", name, movies)
				return &schemas.ExistMediaInfo{
					Type:       types.MediaTypeMovie,
					ServerType: "emby",
					Server:     name,
					ItemID:     movies[0].ItemID,
				}
			}
		} else {
			seriesID, tvs := s.GetTVEpisodes("", mediainfo.Title, mediainfo.Year, &mediainfo.TmdbID, nil)
			if len(tvs) == 0 {
				logger.Infof("%s 没有在媒体库 %s �?, mediainfo.TitleYear, name)
				continue
			} else {
				logger.Infof("%s 在媒体库 %s 中找到了这些季集�?v", mediainfo.TitleYear, name, tvs)
				return &schemas.ExistMediaInfo{
					Type:       types.MediaTypeTV,
					Seasons:    tvs,
					ServerType: "emby",
					Server:     name,
					ItemID:     seriesID,
				}
			}
		}
	}
	
	return nil
}

// MediaStatistic 媒体数量统计
func (e *EmbyModule) MediaStatistic(serverName string) []*schemas.Statistic {
	var servers []*Emby
	
	if serverName != "" {
		if server, exists := e.instances[serverName]; exists {
			servers = []*Emby{server}
		}
	} else {
		for _, server := range e.instances {
			servers = append(servers, server)
		}
	}
	
	mediaStatistics := []*schemas.Statistic{}
	for _, s := range servers {
		mediaStatistic := s.GetMediasCount()
		if mediaStatistic == nil {
			continue
		}
		mediaStatistic.UserCount = s.GetUserCount()
		mediaStatistics = append(mediaStatistics, mediaStatistic)
	}
	
	return mediaStatistics
}

// MediaserverLibrarys 媒体库列�?func (e *EmbyModule) MediaserverLibrarys(serverName, username string, hidden bool) []*schemas.MediaServerLibrary {
	if server, exists := e.instances[serverName]; exists {
		return server.GetLibrarys(username, hidden)
	}
	return nil
}

// MediaserverItems 获取媒体服务器项目列表，支持分页和不分页逻辑，默认不分页获取所有数�?func (e *EmbyModule) MediaserverItems(serverName, libraryID string, startIndex, limit int) []*schemas.MediaServerItem {
	if server, exists := e.instances[serverName]; exists {
		return server.GetItems(libraryID, startIndex, limit)
	}
	return nil
}

// MediaserverIteminfo 媒体库项目详�?func (e *EmbyModule) MediaserverIteminfo(serverName, itemID string) *schemas.MediaServerItem {
	if server, exists := e.instances[serverName]; exists {
		return server.GetItemInfo(itemID)
	}
	return nil
}

// MediaserverTVEpisodes 获取剧集信息
func (e *EmbyModule) MediaserverTVEpisodes(serverName, itemID string) []*schemas.MediaServerSeasonInfo {
	if server, exists := e.instances[serverName]; exists {
		_, seasonInfo := server.GetTVEpisodes(itemID, "", "", nil, nil)
		if len(seasonInfo) == 0 {
			return []*schemas.MediaServerSeasonInfo{}
		}
		
		result := []*schemas.MediaServerSeasonInfo{}
		for season, episodes := range seasonInfo {
			result = append(result, &schemas.MediaServerSeasonInfo{
				Season:   season,
				Episodes: episodes,
			})
		}
		return result
	}
	
	return nil
}

// MediaserverPlaying 获取媒体服务器正在播放信�?func (e *EmbyModule) MediaserverPlaying(serverName string, count int, username string) []*schemas.MediaServerPlayItem {
	if server, exists := e.instances[serverName]; exists {
		return server.GetResume(count, username)
	}
	return []*schemas.MediaServerPlayItem{}
}

// MediaserverPlayURL 获取媒体库播放地址
func (e *EmbyModule) MediaserverPlayURL(serverName, itemID string) string {
	if server, exists := e.instances[serverName]; exists {
		return server.GetPlayURL(itemID)
	}
	return ""
}

// MediaserverLatest 获取媒体服务器最新入库条�?func (e *EmbyModule) MediaserverLatest(serverName string, count int, username string) []*schemas.MediaServerPlayItem {
	if server, exists := e.instances[serverName]; exists {
		return server.GetLatest(count, username)
	}
	return []*schemas.MediaServerPlayItem{}
}

// MediaserverLatestImages 获取媒体服务器最新入库条目的图片
func (e *EmbyModule) MediaserverLatestImages(serverName string, count int, username string, remote bool) []string {
	if server, exists := e.instances[serverName]; exists {
		links := []string{}
		items := e.MediaserverLatest(serverName, count, username)
		
		for _, item := range items {
			if len(item.BackdropImageTags) > 0 {
				if tag, ok := item.BackdropImageTags[0].(string); ok {
					imageURL := server.GetBackdropURL(item.ID, tag, remote)
					if imageURL != "" {
						links = append(links, imageURL)
					}
				}
			}
		}
		return links
	}
	
	return []string{}
}

// HandleConfigChanged 处理配置变更事件
func (e *EmbyModule) HandleConfigChanged(event *event.Event) {
	if event == nil {
		return
	}
	
	eventData := event.Data.(*schemas.ConfigChangeEventData)
	if eventData.Key != string(types.SystemConfigKeyMediaServers) {
		return
	}
	
	logger.Info("配置变更，重新初始化Emby模块...")
	e.InitModule()
}

// GetInstance 获取指定名称的Emby实例
func (e *EmbyModule) GetInstance(name string) *Emby {
	if server, exists := e.instances[name]; exists {
		return server
	}
	return nil
}

// GetInstances 获取所有Emby实例
func (e *EmbyModule) GetInstances() map[string]*Emby {
	return e.instances
}
