package plex

import (
	"fmt"
	"net/url"
	"strconv"

	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// PlexModule Plex模块
type PlexModule struct {
	modules.ModuleBase
}

// NewPlexModule 创建Plex模块实例
func NewPlexModule() *PlexModule {
	return &PlexModule{}
}

// InitModule 初始化模�?func (pm *PlexModule) InitModule() error {
	// 初始化模�?	// 这里应该初始化Plex服务
	utils.Log.Info("初始化Plex模块")
	return nil
}

// GetName 获取模块名称
func (pm *PlexModule) GetName() string {
	return "Plex"
}

// GetType 获取模块类型
func (pm *PlexModule) GetType() models.ModuleType {
	return models.ModuleTypeMediaServer
}

// GetSubtype 获取模块子类�?func (pm *PlexModule) GetSubtype() string {
	return "Plex"
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (pm *PlexModule) GetPriority() int {
	return 3
}

// InitSetting 初始化设�?func (pm *PlexModule) InitSetting() (string, interface{}) {
	return "", nil
}

// Stop 停止模块
func (pm *PlexModule) Stop() {
	// 停止模块服务
	instances := pm.GetInstances()
	for _, server := range instances {
		if plex, ok := server.(*Plex); ok && plex != nil {
			plex.Close()
		}
	}
}

// Test 测试模块连接�?func (pm *PlexModule) Test() (bool, string) {
	instances := pm.GetInstances()
	if len(instances) == 0 {
		return true, ""
	}
	
	for name, server := range instances {
		if plex, ok := server.(*Plex); ok {
			if plex.IsInactive() {
				plex.Reconnect()
			}
			
			if len(plex.GetLibrarys(false)) == 0 {
				return false, fmt.Sprintf("无法连接Plex服务器：%s", name)
			}
		}
	}
	
	return true, ""
}

// SchedulerJob 定时任务，每10分钟调用一�?func (pm *PlexModule) SchedulerJob() {
	// 定时重连
	instances := pm.GetInstances()
	for name, server := range instances {
		if plex, ok := server.(*Plex); ok && plex.IsInactive() {
			utils.Log.Infof("Plex %s 服务器连接断开，尝试重�?...", name)
			plex.Reconnect()
		}
	}
}

// UserAuthenticate 使用Plex用户辅助完成用户认证
func (pm *PlexModule) UserAuthenticate(credentials *models.AuthCredentials, serviceName *string) *models.AuthCredentials {
	// Plex认证
	if credentials == nil || credentials.GrantType != "password" {
		return nil
	}
	
	// 确定要认证的服务器列�?	var servers map[string]interface{}
	if serviceName != nil {
		// 如果指定了服务名，获取该服务实例
		server := pm.GetInstance(serviceName)
		if server != nil {
			servers = map[string]interface{}{*serviceName: server}
		} else {
			servers = map[string]interface{}{}
		}
	} else {
		// 如果没有指定服务名，遍历所有服�?		servers = pm.GetInstances()
	}
	
	// 遍历要认证的服务�?	for name, server := range servers {
		if plex, ok := server.(*Plex); ok {
			// 触发认证拦截事件
			// 注意：Go版本中事件系统可能需要不同的实现方式
			/*
			interceptEvent := eventmanager.SendEvent(
				ChainEventTypeAuthIntercept,
				&AuthInterceptCredentials{
					Username: credentials.Username,
					Channel:  pm.GetName(),
					Service:  name,
					Status:   "triggered",
				},
			)
			
			if interceptEvent != nil && interceptEvent.EventData != nil {
				interceptData := interceptEvent.EventData.(*AuthInterceptCredentials)
				if interceptData.Cancel {
					continue
				}
			}
			*/
			
			authResultToken, authResultUsername := plex.Authenticate(credentials.Username, credentials.Password)
			if authResultToken != nil && authResultUsername != nil {
				credentials.Channel = pm.GetName()
				credentials.Service = name
				credentials.Token = *authResultToken
				// Plex 传入可能为邮箱，这里调整为用户名返回
				credentials.Username = *authResultUsername
				return credentials
			}
		}
	}
	
	return nil
}

// WebhookParser 解析Webhook报文�?func (pm *PlexModule) WebhookParser(body interface{}, form map[string][]string, args map[string]string) *models.WebhookEventInfo {
	source := args["source"]
	if source != "" {
		server := pm.GetInstance(&source)
		if server == nil {
			return nil
		}
		
		if plex, ok := server.(*Plex); ok {
			// 转换form格式
			urlValues := url.Values{}
			for k, v := range form {
				if len(v) > 0 {
					urlValues.Set(k, v[0])
				}
			}
			
			result := plex.GetWebhookMessage(urlValues)
			if result != nil {
				result.ServerName = &source
			}
			return result
		}
	}
	
	for _, server := range pm.GetInstances() {
		if plex, ok := server.(*Plex); ok {
			// 转换form格式
			urlValues := url.Values{}
			for k, v := range form {
				if len(v) > 0 {
					urlValues.Set(k, v[0])
				}
			}
			
			result := plex.GetWebhookMessage(urlValues)
			if result != nil {
				return result
			}
		}
	}
	
	return nil
}

// MediaExists 判断媒体文件是否存在
func (pm *PlexModule) MediaExists(mediainfo *models.MediaInfo, itemid *string, serverName *string) *models.ExistMediaInfo {
	var servers map[string]interface{}
	
	if serverName != nil {
		servers = map[string]interface{}{*serverName: pm.GetInstance(serverName)}
	} else {
		servers = pm.GetInstances()
	}
	
	for name, s := range servers {
		if s == nil {
			continue
		}
		
		if plex, ok := s.(*Plex); ok {
			if mediainfo.Type == "movie" {
				if itemid != nil {
					movie := plex.GetItemInfo(*itemid)
					if movie != nil {
						utils.Log.Infof("媒体�?%s 中找到了 %v", name, movie)
						return &models.ExistMediaInfo{
							Type:       "movie",
							ServerType: "plex",
							Server:     name,
							ItemID:     movie.ItemID,
						}
					}
				}
				
				var tmdbID *int
				if mediainfo.TmdbID != 0 {
					tmdbID = &mediainfo.TmdbID
				}
				
				movies := plex.GetMovies(mediainfo.Title, mediainfo.OriginalTitle, mediainfo.Year, tmdbID)
				if len(movies) == 0 {
					utils.Log.Infof("%s 没有在媒体库 %s �?, mediainfo.Title, name)
					continue
				} else {
					utils.Log.Infof("媒体�?%s 中找到了 %v", name, movies)
					return &models.ExistMediaInfo{
						Type:       "movie",
						ServerType: "plex",
						Server:     name,
						ItemID:     movies[0].ItemID,
					}
				}
			} else {
				var tmdbID *int
				if mediainfo.TmdbID != 0 {
					tmdbID = &mediainfo.TmdbID
				}
				
				itemID, tvs := plex.GetTVEpisodes("", mediainfo.Title, mediainfo.OriginalTitle, mediainfo.Year, tmdbID)
				if len(tvs) == 0 {
					utils.Log.Infof("%s 没有在媒体库 %s �?, mediainfo.Title, name)
					continue
				} else {
					utils.Log.Infof("%s 在媒体库 %s 中找到了这些季集�?v", mediainfo.Title, name, tvs)
					return &models.ExistMediaInfo{
						Type:       "tv",
						Seasons:    tvs,
						ServerType: "plex",
						Server:     name,
						ItemID:     itemID,
					}
				}
			}
		}
	}
	
	return nil
}

// MediaStatistic 媒体数量统计
func (pm *PlexModule) MediaStatistic(server *string) []models.Statistic {
	var servers []interface{}
	
	if server != nil {
		serverObj := pm.GetInstance(server)
		if serverObj == nil {
			return nil
		}
		servers = []interface{}{serverObj}
	} else {
		for _, s := range pm.GetInstances() {
			servers = append(servers, s)
		}
	}
	
	var mediaStatistics []models.Statistic
	for _, s := range servers {
		if plex, ok := s.(*Plex); ok {
			mediaStatistic := plex.GetMediasCount()
			if mediaStatistic != nil {
				// Go中没有直接修改结构体字段的方式，需要创建新实例
				stat := models.Statistic{
					MovieCount:   mediaStatistic.MovieCount,
					TvCount:      mediaStatistic.TvCount,
					EpisodeCount: mediaStatistic.EpisodeCount,
					UserCount:    1, // Plex默认用户数为1
				}
				mediaStatistics = append(mediaStatistics, stat)
			}
		}
	}
	
	return mediaStatistics
}

// MediaServerLibrarys 媒体库列�?func (pm *PlexModule) MediaServerLibrarys(server *string, hidden bool, kwargs map[string]interface{}) []models.MediaServerLibrary {
	if server != nil {
		serverObj := pm.GetInstance(server)
		if serverObj != nil {
			if plex, ok := serverObj.(*Plex); ok {
				return plex.GetLibrarys(hidden)
			}
		}
	}
	return nil
}

// MediaServerItems 获取媒体服务器项目列�?func (pm *PlexModule) MediaServerItems(server string, libraryID string, startIndex int, limit int) []models.MediaServerItem {
	serverObj := pm.GetInstance(&server)
	if serverObj != nil {
		if plex, ok := serverObj.(*Plex); ok {
			return plex.GetItems(libraryID, startIndex, limit)
		}
	}
	return nil
}

// MediaServerItemInfo 媒体库项目详�?func (pm *PlexModule) MediaServerItemInfo(server string, itemID string) *models.MediaServerItem {
	serverObj := pm.GetInstance(&server)
	if serverObj != nil {
		if plex, ok := serverObj.(*Plex); ok {
			return plex.GetItemInfo(itemID)
		}
	}
	return nil
}

// MediaServerTVEpisodes 获取剧集信息
func (pm *PlexModule) MediaServerTVEpisodes(server string, itemID string) []models.MediaServerSeasonInfo {
	serverObj := pm.GetInstance(&server)
	if serverObj == nil {
		return nil
	}
	
	if plex, ok := serverObj.(*Plex); ok {
		_, seasonInfo := plex.GetTVEpisodes(itemID, "", "", "", nil)
		if len(seasonInfo) == 0 {
			return []models.MediaServerSeasonInfo{}
		}
		
		var result []models.MediaServerSeasonInfo
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

// MediaServerPlaying 获取媒体服务器正在播放信�?func (pm *PlexModule) MediaServerPlaying(server string, count int, kwargs map[string]interface{}) []models.MediaServerPlayItem {
	serverObj := pm.GetInstance(&server)
	if serverObj == nil {
		return []models.MediaServerPlayItem{}
	}
	
	if plex, ok := serverObj.(*Plex); ok {
		return plex.GetResume(count)
	}
	
	return []models.MediaServerPlayItem{}
}

// MediaServerLatest 获取媒体服务器最新入库条�?func (pm *PlexModule) MediaServerLatest(server *string, count int, kwargs map[string]interface{}) []models.MediaServerPlayItem {
	var serverObj interface{}
	if server != nil {
		serverObj = pm.GetInstance(server)
	} else {
		// 获取第一个实�?		instances := pm.GetInstances()
		for _, s := range instances {
			serverObj = s
			break
		}
	}
	
	if serverObj == nil {
		return []models.MediaServerPlayItem{}
	}
	
	if plex, ok := serverObj.(*Plex); ok {
		return plex.GetLatest(count)
	}
	
	return []models.MediaServerPlayItem{}
}

// MediaServerLatestImages 获取媒体服务器最新入库条目的图片
func (pm *PlexModule) MediaServerLatestImages(server *string, count int, username *string, kwargs map[string]interface{}) []string {
	serverObj := pm.GetInstance(server)
	if serverObj == nil {
		return []string{}
	}
	
	if plex, ok := serverObj.(*Plex); ok {
		links := []string{}
		items := pm.MediaServerLatest(server, count, kwargs)
		
		for _, item := range items {
			if item.ID != nil {
				itemID := ""
				switch v := item.ID.(type) {
				case string:
					itemID = v
				case int:
					itemID = strconv.Itoa(v)
				}
				
				if itemID != "" {
					link := plex.GetRemoteImageByID(itemID, "Backdrop", 0, false)
					if link != nil {
						links = append(links, *link)
					}
				}
			}
		}
		
		return links
	}
	
	return []string{}
}

// MediaServerPlayURL 获取媒体库播放地址
func (pm *PlexModule) MediaServerPlayURL(server string, itemID string) *string {
	serverObj := pm.GetInstance(&server)
	if serverObj == nil {
		return nil
	}
	
	if plex, ok := serverObj.(*Plex); ok {
		url := plex.GetPlayURL(itemID)
		return &url
	}
	
	return nil
}
