package jellyfin

import (
	"moviepilot-go/internal/entity"
	"moviepilot-go/internal/modules"
)

type JellyfinModule struct {
	modules.BaseMediaServerModule[*Jellyfin]
}

func NewJellyfinModule() *JellyfinModule {
	m := &JellyfinModule{}
	m.InitService("jellyfin", func(conf *entity.MediaServerConfig) *Jellyfin {
		return NewJellyfin(conf.Host, conf.ApiKey, conf.PlayHost, conf.SyncLibraries)
	})
	return m
}

func (m *JellyfinModule) GetInstances() map[string]*Jellyfin {
	instances := make(map[string]*Jellyfin)
	for name, instance := range m.BaseMediaServerModule.GetInstances() {
		instances[name] = instance.(*Jellyfin)
	}
	return instances
}

func (m *JellyfinModule) GetInstance(name string) *Jellyfin {
	instance := m.BaseMediaServerModule.GetInstance(name)
	if instance != nil {
		if j, ok := instance.(*Jellyfin); ok {
			return j
		}
	}
	return nil
}

// 实现Jellyfin模块的所有接口方�?func (m *JellyfinModule) GetName() string {
	return "Jellyfin"
}

func (m *JellyfinModule) GetType() entity.ModuleType {
	return entity.ModuleTypeMediaServer
}

func (m *JellyfinModule) GetSubType() entity.MediaServerType {
	return entity.MediaServerTypeJellyfin
}

func (m *JellyfinModule) GetPriority() int {
	return 2
}

func (m *JellyfinModule) Test() (bool, string) {
	if len(m.GetInstances()) == 0 {
		return false, "没有配置Jellyfin服务�?
	}
	
	for name, server := range m.GetInstances() {
		if server.IsInactive() {
			server.Reconnect()
		}
		if server.GetUser("") == "" {
			return false, "无法连接Jellyfin服务器：" + name
		}
	}
	
	return true, "Jellyfin服务器连接正�?
}

func (m *JellyfinModule) MediaExists(mediaInfo *entity.MediaInfo, itemId, serverName string) *entity.ExistMediaInfo {
	// 确定要查询的服务�?	var servers map[string]*Jellyfin
	if serverName != "" {
		if server := m.GetInstance(serverName); server != nil {
			servers = map[string]*Jellyfin{serverName: server}
		} else {
			return nil
		}
	} else {
		servers = m.GetInstances()
	}
	
	// 遍历服务器查找媒�?	for name, server := range servers {
		if server == nil {
			continue
		}
		
		if mediaInfo.Type == "movie" {
			// 检查电�?			if itemId != "" {
				// 根据itemId查找
				movie := server.GetItemInfo(itemId)
				if movie != nil {
					common.LOG.Infof("媒体�?%s 中找到了 %v", name, movie)
					return &entity.ExistMediaInfo{
						Type:       "movie",
						ServerType: "jellyfin",
						Server:     name,
						ItemId:     movie.ItemId,
					}
				}
			}
			
			// 根据标题、年份、tmdbId查找
			movies := server.GetMovies(mediaInfo.Title, mediaInfo.Year, mediaInfo.TmdbID)
			if len(movies) == 0 {
				common.LOG.Infof("%s 没有在媒体库 %s �?, mediaInfo.TitleWithYear(), name)
				continue
			} else {
				common.LOG.Infof("媒体�?%s 中找到了 %v", name, movies)
				return &entity.ExistMediaInfo{
					Type:       "movie",
					ServerType: "jellyfin",
					Server:     name,
					ItemId:     movies[0].ItemId,
				}
			}
		} else {
			// 检查电视剧
			foundItemId, tvs := server.GetTvEpisodes(itemId, mediaInfo.Title, mediaInfo.Year, mediaInfo.TmdbID, 0)
			if len(tvs) == 0 {
				common.LOG.Infof("%s 没有在媒体库 %s �?, mediaInfo.TitleWithYear(), name)
				continue
			} else {
				common.LOG.Infof("%s 在媒体库 %s 中找到了这些季集�?v", mediaInfo.TitleWithYear(), name, tvs)
				return &entity.ExistMediaInfo{
					Type:       "tv",
					Seasons:    tvs,
					ServerType: "jellyfin",
					Server:     name,
					ItemId:     foundItemId,
				}
			}
		}
	}
	
	return nil
}

func (m *JellyfinModule) MediaStatistic(serverName string) []*entity.Statistic {
	var servers []*Jellyfin
	
	if serverName != "" {
		if server := m.GetInstance(serverName); server != nil {
			servers = append(servers, server)
		} else {
			return nil
		}
	} else {
		for _, server := range m.GetInstances() {
			servers = append(servers, server)
		}
	}
	
	mediaStatistics := []*entity.Statistic{}
	for _, server := range servers {
		mediaStatistic := server.GetMediasCount()
		if mediaStatistic.MovieCount == 0 && mediaStatistic.TvCount == 0 && mediaStatistic.EpisodeCount == 0 {
			continue
		}
		mediaStatistic.UserCount = server.GetUserCount()
		mediaStatistics = append(mediaStatistics, &mediaStatistic)
	}
	
	return mediaStatistics
}

func (m *JellyfinModule) MediaServerLibrarys(serverName, username string, hidden bool) []entity.MediaServerLibrary {
	server := m.GetInstance(serverName)
	if server != nil {
		return server.GetLibrarys(username, hidden)
	}
	return nil
}

func (m *JellyfinModule) MediaServerItems(serverName string, libraryId string, startIndex, limit int) []*entity.MediaServerItem {
	server := m.GetInstance(serverName)
	if server != nil {
		return server.GetItems(libraryId, startIndex, limit)
	}
	return nil
}

func (m *JellyfinModule) MediaServerItemInfo(serverName, itemId string) *entity.MediaServerItem {
	server := m.GetInstance(serverName)
	if server != nil {
		return server.GetItemInfo(itemId)
	}
	return nil
}

func (m *JellyfinModule) MediaServerTvEpisodes(serverName, itemId string) []entity.MediaServerSeasonInfo {
	server := m.GetInstance(serverName)
	if server == nil {
		return nil
	}
	
	_, seasonInfo := server.GetTvEpisodes(itemId, "", "", 0, 0)
	if seasonInfo == nil {
		return []entity.MediaServerSeasonInfo{}
	}
	
	seasons := []entity.MediaServerSeasonInfo{}
	for season, episodes := range seasonInfo {
		seasons = append(seasons, entity.MediaServerSeasonInfo{
			Season:   season,
			Episodes: episodes,
		})
	}
	
	return seasons
}

func (m *JellyfinModule) MediaServerPlaying(serverName string, count int, username string) []entity.MediaServerPlayItem {
	server := m.GetInstance(serverName)
	if server == nil {
		return []entity.MediaServerPlayItem{}
	}
	
	return server.GetResume(count, username)
}

func (m *JellyfinModule) MediaServerPlayUrl(serverName, itemId string) string {
	server := m.GetInstance(serverName)
	if server == nil {
		return ""
	}
	
	return server.GetPlayUrl(itemId)
}

func (m *JellyfinModule) MediaServerLatest(serverName string, count int, username string) []entity.MediaServerPlayItem {
	server := m.GetInstance(serverName)
	if server == nil {
		return []entity.MediaServerPlayItem{}
	}
	
	return server.GetLatest(count, username)
}

func (m *JellyfinModule) MediaServerLatestImages(serverName string, count int, username string, remote bool) []string {
	server := m.GetInstance(serverName)
	if server == nil {
		return []string{}
	}
	
	links := []string{}
	items := m.MediaServerLatest(serverName, count, username)
	
	for _, item := range items {
		if len(item.BackdropImageTags) > 0 {
			if tag, ok := item.BackdropImageTags[0].(string); ok {
				imageUrl := server.GetBackdropUrl(item.Id, tag, remote)
				if imageUrl != "" {
					links = append(links, imageUrl)
				}
			}
		}
	}
	
	return links
}
