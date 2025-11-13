package emby

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils"
)

// Emby Emby服务器类
type Emby struct {
	host          string
	playhost      string
	apikey        string
	syncLibraries []string
	user          interface{}
	username      string
	folders       []map[string]interface{}
	serverid      string
}

// NewEmby 创建Emby实例
func NewEmby(host, apikey, playHost, username string, syncLibraries []string, kwargs map[string]interface{}) *Emby {
	if host == "" || apikey == "" {
		logger.Error("Emby服务器配置不完整�?)
		return nil
	}

	emby := &Emby{
		host:          utils.UrlUtils.StandardizeBaseURL(host),
		playhost:      utils.UrlUtils.StandardizeBaseURL(playHost),
		apikey:        apikey,
		username:      username,
		syncLibraries: syncLibraries,
	}

	emby.user = emby.GetUser(username)
	emby.folders = emby.GetEmbyFolders()
	emby.serverid = emby.GetServerID()

	return emby
}

// IsInactive 判断是否需要重�?func (e *Emby) IsInactive() bool {
	if e.host == "" || e.apikey == "" {
		return false
	}
	return e.user == nil
}

// Reconnect 重连
func (e *Emby) Reconnect() {
	e.user = e.GetUser("")
	e.folders = e.GetEmbyFolders()
}

// GetEmbyFolders 获取Emby媒体库路径列�?func (e *Emby) GetEmbyFolders() []map[string]interface{} {
	if e.host == "" || e.apikey == "" {
		return []map[string]interface{}{}
	}
	
	url := fmt.Sprintf("%s/emby/Library/SelectableMediaFolders", e.host)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Library/SelectableMediaFolders 出错�?s", err.Error())
		return []map[string]interface{}{}
	}
	
	if res != nil {
		defer res.Body.Close()
		var result []map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("Library/SelectableMediaFolders 未获取到返回数据")
			return []map[string]interface{}{}
		}
		return result
	} else {
		logger.Error("Library/SelectableMediaFolders 未获取到返回数据")
		return []map[string]interface{}{}
	}
}

// GetEmbyVirtualFolders 获取Emby媒体库所有路径列表（包含共享路径�?func (e *Emby) GetEmbyVirtualFolders() []map[string]interface{} {
	if e.host == "" || e.apikey == "" {
		return []map[string]interface{}{}
	}
	
	url := fmt.Sprintf("%s/emby/Library/VirtualFolders/Query", e.host)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Library/VirtualFolders/Query 出错�?s", err.Error())
		return []map[string]interface{}{}
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("Library/VirtualFolders/Query 未获取到返回数据")
			return []map[string]interface{}{}
		}
		
		libraryItems, ok := result["Items"].([]interface{})
		if !ok {
			logger.Error("Library/VirtualFolders/Query 返回数据格式错误")
			return []map[string]interface{}{}
		}
		
		librarys := []map[string]interface{}{}
		for _, item := range libraryItems {
			libraryItem, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			
			libraryID := libraryItem["ItemId"]
			libraryName := libraryItem["Name"]
			
			libraryOptions, ok := libraryItem["LibraryOptions"].(map[string]interface{})
			if !ok {
				continue
			}
			
			pathInfos, ok := libraryOptions["PathInfos"].([]interface{})
			if !ok {
				continue
			}
			
			libraryPaths := []string{}
			for _, pathInfo := range pathInfos {
				path, ok := pathInfo.(map[string]interface{})
				if !ok {
					continue
				}
				
				if networkPath, exists := path["NetworkPath"]; exists && networkPath != nil {
					libraryPaths = append(libraryPaths, networkPath.(string))
				} else if pathValue, exists := path["Path"]; exists && pathValue != nil {
					libraryPaths = append(libraryPaths, pathValue.(string))
				}
			}
			
			if libraryName != nil && len(libraryPaths) > 0 {
				librarys = append(librarys, map[string]interface{}{
					"Id":   libraryID,
					"Name": libraryName,
					"Path": libraryPaths,
				})
			}
		}
		return librarys
	} else {
		logger.Error("Library/VirtualFolders/Query 未获取到返回数据")
		return []map[string]interface{}{}
	}
}

// getEmbyLibrarys 获取Emby媒体库列�?func (e *Emby) getEmbyLibrarys(username string) []map[string]interface{} {
	if e.host == "" || e.apikey == "" {
		return []map[string]interface{}{}
	}
	
	var user interface{}
	if username != "" {
		user = e.GetUser(username)
	} else {
		user = e.user
	}
	
	url := fmt.Sprintf("%s/emby/Users/%v/Views", e.host, user)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接User/Views 出错�?s", err.Error())
		return []map[string]interface{}{}
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("User/Views 未获取到返回数据")
			return []map[string]interface{}{}
		}
		
		items, ok := result["Items"].([]interface{})
		if !ok {
			return []map[string]interface{}{}
		}
		
		libraries := make([]map[string]interface{}, len(items))
		for i, item := range items {
			libraries[i] = item.(map[string]interface{})
		}
		return libraries
	} else {
		logger.Error("User/Views 未获取到返回数据")
		return []map[string]interface{}{}
	}
}

// GetLibrarys 获取媒体服务器所有媒体库列表
func (e *Emby) GetLibrarys(username string, hidden bool) []*schemas.MediaServerLibrary {
	if e.host == "" || e.apikey == "" {
		return []*schemas.MediaServerLibrary{}
	}
	
	libraries := []*schemas.MediaServerLibrary{}
	embyLibraries := e.getEmbyLibrarys(username)
	
	for _, library := range embyLibraries {
		// 检查是否需要隐�?		if hidden && len(e.syncLibraries) > 0 && !containsString(e.syncLibraries, "all") {
			if libraryID, ok := library["Id"].(string); ok {
				if !containsString(e.syncLibraries, libraryID) {
					continue
				}
			}
		}
		
		var libraryType string
		if collectionType, ok := library["CollectionType"].(string); ok {
			switch collectionType {
			case "movies":
				libraryType = string(types.MediaTypeMovie)
			case "tvshows":
				libraryType = string(types.MediaTypeTV)
			default:
				libraryType = string(types.MediaTypeUnknown)
			}
		} else {
			libraryType = string(types.MediaTypeUnknown)
		}
		
		image := e.getLocalImageByID(library["Id"].(string))
		
		libraryPath := ""
		if path, ok := library["Path"].(string); ok {
			libraryPath = path
		}
		
		link := fmt.Sprintf("%s/web/index.html#!/videos?serverId=%s&parentId=%s", 
			e.playhost, e.serverid, library["Id"].(string))
		
		libraries = append(libraries, &schemas.MediaServerLibrary{
			Server:     "emby",
			ID:         library["Id"].(string),
			Name:       library["Name"].(string),
			Path:       libraryPath,
			Type:       libraryType,
			Image:      image,
			Link:       link,
			ServerType: "emby",
		})
	}
	
	return libraries
}

// GetUser 获得管理员用�?func (e *Emby) GetUser(userName string) interface{} {
	if e.host == "" || e.apikey == "" {
		return nil
	}
	
	url := fmt.Sprintf("%s/Users", e.host)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Users出错�?s", err.Error())
		return nil
	}
	
	if res != nil {
		defer res.Body.Close()
		var users []map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
			logger.Error("Users 未获取到返回数据")
			return nil
		}
		
		// 先查询是否有与当前用户名称匹配的
		if userName != "" {
			for _, user := range users {
				if name, ok := user["Name"].(string); ok && name == userName {
					return user["Id"]
				}
			}
		}
		
		// 查询管理�?		for _, user := range users {
			if policy, ok := user["Policy"].(map[string]interface{}); ok {
				if isAdministrator, ok := policy["IsAdministrator"].(bool); ok && isAdministrator {
					return user["Id"]
				}
			}
		}
	} else {
		logger.Error("Users 未获取到返回数据")
	}
	
	return nil
}

// Authenticate 用户认证
func (e *Emby) Authenticate(username, password string) string {
	if e.host == "" || e.apikey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%s/emby/Users/AuthenticateByName", e.host)
	
	headers := map[string]string{
		"X-Emby-Authorization": fmt.Sprintf(`MediaBrowser Client="MoviePilot", Device="requests", DeviceId="1", Version="1.0.0", Token="%s"`, e.apikey),
		"Content-Type":         "application/json",
		"Accept":               "application/json",
	}
	
	data := map[string]string{
		"Username": username,
		"Pw":       password,
	}
	
	jsonData, _ := json.Marshal(data)
	
	res, err := utils.RequestUtils.PostRes(url, headers, string(jsonData), 0)
	if err != nil {
		logger.Errorf("连接Users/AuthenticateByName出错�?s", err.Error())
		return ""
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("Users/AuthenticateByName 未获取到返回数据")
			return ""
		}
		
		if authToken, ok := result["AccessToken"].(string); ok && authToken != "" {
			logger.Infof("用户 %s Emby认证成功", username)
			return authToken
		}
	} else {
		logger.Error("Users/AuthenticateByName 未获取到返回数据")
	}
	
	return ""
}

// GetServerID 获得服务器信�?func (e *Emby) GetServerID() string {
	if e.host == "" || e.apikey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%s/System/Info", e.host)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接System/Info出错�?s", err.Error())
		return ""
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("System/Info 未获取到返回数据")
			return ""
		}
		
		if id, ok := result["Id"].(string); ok {
			return id
		}
	} else {
		logger.Error("System/Info 未获取到返回数据")
	}
	
	return ""
}

// GetUserCount 获得用户数量
func (e *Emby) GetUserCount() int {
	if e.host == "" || e.apikey == "" {
		return 0
	}
	
	url := fmt.Sprintf("%s/emby/Users/Query", e.host)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Users/Query出错�?s", err.Error())
		return 0
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("Users/Query 未获取到返回数据")
			return 0
		}
		
		if count, ok := result["TotalRecordCount"].(float64); ok {
			return int(count)
		}
	} else {
		logger.Error("Users/Query 未获取到返回数据")
	}
	
	return 0
}

// GetMediasCount 获得电影、电视剧、动漫媒体数�?func (e *Emby) GetMediasCount() *schemas.Statistic {
	if e.host == "" || e.apikey == "" {
		return &schemas.Statistic{}
	}
	
	url := fmt.Sprintf("%s/emby/Items/Counts", e.host)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Items/Counts出错�?s", err.Error())
		return &schemas.Statistic{}
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("Items/Counts 未获取到返回数据")
			return &schemas.Statistic{}
		}
		
		statistic := &schemas.Statistic{}
		if movieCount, ok := result["MovieCount"].(float64); ok {
			statistic.MovieCount = int(movieCount)
		}
		if seriesCount, ok := result["SeriesCount"].(float64); ok {
			statistic.TVCount = int(seriesCount)
		}
		if episodeCount, ok := result["EpisodeCount"].(float64); ok {
			statistic.EpisodeCount = int(episodeCount)
		}
		
		return statistic
	} else {
		logger.Error("Items/Counts 未获取到返回数据")
		return &schemas.Statistic{}
	}
}

// getEmbySeriesIDByName 根据名称查询Emby中剧集的SeriesId
func (e *Emby) getEmbySeriesIDByName(name, year string) string {
	if e.host == "" || e.apikey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%s/emby/Items", e.host)
	params := map[string]string{
		"IncludeItemTypes":      "Series",
		"Fields":                "ProductionYear",
		"StartIndex":            "0",
		"Recursive":             "true",
		"SearchTerm":            name,
		"Limit":                 "10",
		"IncludeSearchTypes":    "false",
		"api_key":               e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Items出错�?s", err.Error())
		return ""
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return ""
		}
		
		if resItems, ok := result["Items"].([]interface{}); ok {
			for _, item := range resItems {
				resItem, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				
				itemName, nameOk := resItem["Name"].(string)
				itemYear, yearOk := resItem["ProductionYear"].(float64)
				
				if nameOk && itemName == name {
					if year == "" || (yearOk && strconv.FormatFloat(itemYear, 'f', -1, 64) == year) {
						if id, idOk := resItem["Id"].(string); idOk {
							return id
						}
					}
				}
			}
		}
	}
	
	return ""
}

// GetMovies 根据标题和年份，检查电影是否在Emby中存在，存在则返回列�?func (e *Emby) GetMovies(title, year string, tmdbID *int) []*schemas.MediaServerItem {
	if e.host == "" || e.apikey == "" {
		return nil
	}
	
	url := fmt.Sprintf("%s/emby/Items", e.host)
	params := map[string]string{
		"IncludeItemTypes":   "Movie",
		"Fields":             "ProviderIds,OriginalTitle,ProductionYear,Path,UserDataPlayCount,UserDataLastPlayedDate,ParentId",
		"StartIndex":         "0",
		"Recursive":          "true",
		"SearchTerm":         title,
		"Limit":              "10",
		"IncludeSearchTypes": "false",
		"api_key":            e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Items出错�?s", err.Error())
		return nil
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return nil
		}
		
		if resItems, ok := result["Items"].([]interface{}); ok {
			retMovies := []*schemas.MediaServerItem{}
			for _, item := range resItems {
				resItem, ok := item.(map[string]interface{})
				if !ok || resItem == nil {
					continue
				}
				
				mediaserverItem := e.formatItemInfo(resItem)
				if mediaserverItem != nil {
					match := true
					
					// 检查tmdb_id
					if tmdbID != nil && mediaserverItem.TmdbID != *tmdbID {
						match = false
					}
					
					// 检查标�?					if mediaserverItem.Title != title {
						match = false
					}
					
					// 检查年�?					if year != "" && strconv.Itoa(mediaserverItem.Year) != year {
						match = false
					}
					
					if match {
						retMovies = append(retMovies, mediaserverItem)
					}
				}
			}
			return retMovies
		}
	}
	
	return []*schemas.MediaServerItem{}
}

// GetTVEpisodes 根据标题和年份和季，返回Emby中的剧集列表
func (e *Emby) GetTVEpisodes(itemID, title, year string, tmdbID *int, season *int) (string, map[int][]int) {
	if e.host == "" || e.apikey == "" {
		return "", map[int][]int{}
	}
	
	// 电视�?	if itemID == "" {
		itemID = e.getEmbySeriesIDByName(title, year)
		if itemID == "" {
			return "", map[int][]int{}
		}
	}
	
	// 验证tmdbid是否相同
	itemInfo := e.GetItemInfo(itemID)
	if itemInfo != nil && tmdbID != nil && itemInfo.TmdbID != 0 {
		if strconv.Itoa(*tmdbID) != strconv.Itoa(itemInfo.TmdbID) {
			return "", make(map[int][]int)
		}
	}
	
	// 查集的信�?	var seasonParam *int
	if season != nil {
		seasonParam = season
	}
	
	url := fmt.Sprintf("%s/emby/Shows/%s/Episodes", e.host, itemID)
	params := map[string]string{
		"IsMissing": "false",
		"api_key":   e.apikey,
	}
	
	if seasonParam != nil {
		params["Season"] = strconv.Itoa(*seasonParam)
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Shows/Id/Episodes出错�?s", err.Error())
		return "", map[int][]int{}
	}
	
	if res != nil {
		defer res.Body.Close()
		var tvItem map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&tvItem); err != nil {
			return "", map[int][]int{}
		}
		
		if resItems, ok := tvItem["Items"].([]interface{}); ok {
			seasonEpisodes := make(map[int][]int)
			for _, item := range resItems {
				resItem, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				
				seasonIndexFloat, seasonOk := resItem["ParentIndexNumber"].(float64)
				if !seasonOk || seasonIndexFloat == 0 {
					continue
				}
				
				seasonIndex := int(seasonIndexFloat)
				if seasonParam != nil && *seasonParam != seasonIndex {
					continue
				}
				
				episodeIndexFloat, episodeOk := resItem["IndexNumber"].(float64)
				if !episodeOk || episodeIndexFloat == 0 {
					continue
				}
				
				episodeIndex := int(episodeIndexFloat)
				if _, exists := seasonEpisodes[seasonIndex]; !exists {
					seasonEpisodes[seasonIndex] = []int{}
				}
				
				seasonEpisodes[seasonIndex] = append(seasonEpisodes[seasonIndex], episodeIndex)
			}
			
			// 返回
			return itemID, seasonEpisodes
		}
	}
	
	return "", make(map[int][]int)
}

// GetRemoteImageByID 根据ItemId从Emby查询TMDB的图片地址
func (e *Emby) GetRemoteImageByID(itemID, imageType string) string {
	if e.host == "" || e.apikey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%s/emby/Items/%s/RemoteImages", e.host, itemID)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 10)
	if err != nil {
		logger.Errorf("连接Items/Id/RemoteImages出错�?s", err.Error())
		return ""
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Info("Items/RemoteImages 未获取到返回数据，采用本地图�?)
			return e.GenerateExternalImageLink(itemID, imageType)
		}
		
		if images, ok := result["Images"].([]interface{}); ok {
			for _, image := range images {
				img, ok := image.(map[string]interface{})
				if !ok {
					continue
				}
				
				providerName, providerOk := img["ProviderName"].(string)
				imgType, typeOk := img["Type"].(string)
				
				if providerOk && providerName == "TheMovieDb" && typeOk && imgType == imageType {
					if url, urlOk := img["Url"].(string); urlOk {
						return url
					}
				}
			}
		}
		
		// 数据为空
		logger.Info("Items/RemoteImages 未获取到返回数据，采用本地图�?)
		return e.GenerateExternalImageLink(itemID, imageType)
	}
	
	return ""
}

// GenerateExternalImageLink 根据ItemId和imageType查询本地对应图片
func (e *Emby) GenerateExternalImageLink(itemID, imageType string) string {
	if e.playhost == "" {
		logger.Error("Emby外网播放地址未能获取或为�?)
		return ""
	}
	
	url := fmt.Sprintf("%s/Items/%s/Images/%s", e.playhost, itemID, imageType)
	
	res, err := utils.RequestUtils.GetRes(url, nil, nil, 0)
	if err != nil {
		logger.Errorf("连接Items/Id/Images出错�?s", err.Error())
		return ""
	}
	
	if res != nil {
		defer res.Body.Close()
		if res.StatusCode != 404 {
			logger.Infof("影片图片链接:%s", res.Request.URL.String())
			return res.Request.URL.String()
		} else {
			logger.Infof("Items/Id/Images 未获取到返回数据或无该影�?s图片", imageType)
			return ""
		}
	} else {
		logger.Info("Items/Id/Images 未获取到返回数据")
	}
	
	return ""
}

// refreshEmbyLibraryByID 通知Emby刷新一个项目的媒体�?func (e *Emby) refreshEmbyLibraryByID(itemID string) bool {
	if e.host == "" || e.apikey == "" {
		return false
	}
	
	url := fmt.Sprintf("%s/emby/Items/%s/Refresh", e.host, itemID)
	params := map[string]string{
		"Recursive": "true",
		"api_key":   e.apikey,
	}
	
	res, err := utils.RequestUtils.PostRes(url, nil, "", 0, params)
	if err != nil {
		logger.Errorf("连接Items/Id/Refresh出错�?s", err.Error())
		return false
	}
	
	if res != nil {
		defer res.Body.Close()
		return true
	} else {
		logger.Infof("刷新媒体库对�?%s 失败，无法连接Emby�?, itemID)
	}
	
	return false
}

// RefreshRootLibrary 通知Emby刷新整个媒体�?func (e *Emby) RefreshRootLibrary() bool {
	if e.host == "" || e.apikey == "" {
		return false
	}
	
	url := fmt.Sprintf("%s/emby/Library/Refresh", e.host)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.PostRes(url, nil, "", 0, params)
	if err != nil {
		logger.Errorf("连接Library/Refresh出错�?s", err.Error())
		return false
	}
	
	if res != nil {
		defer res.Body.Close()
		return true
	} else {
		logger.Info("刷新媒体库失败，无法连接Emby�?)
	}
	
	return false
}

// RefreshLibraryByItems 按类型、名称、年份来刷新媒体�?func (e *Emby) RefreshLibraryByItems(items []*schemas.RefreshMediaItem) bool {
	if len(items) == 0 {
		return false
	}
	
	// 收集要刷新的媒体库信�?	logger.Info("开始刷新Emby媒体�?..")
	libraryIDs := []string{}
	
	for _, item := range items {
		if item.Title == "" || item.Year == "" || item.Type == "" {
			continue
		}
		
		libraryID := e.getEmbyLibraryIDByItem(item)
		if libraryID != "" && !containsString(libraryIDs, libraryID) {
			libraryIDs = append(libraryIDs, libraryID)
		}
	}
	
	// 开始刷新媒体库
	if containsString(libraryIDs, "/") {
		return e.RefreshRootLibrary()
	}
	
	for _, libraryID := range libraryIDs {
		if libraryID != "/" {
			return e.refreshEmbyLibraryByID(libraryID)
		}
	}
	
	logger.Info("Emby媒体库刷新完�?)
	return true
}

// getEmbyLibraryIDByItem 根据媒体信息查询在哪个媒体库，返回要刷新的位置的ID
func (e *Emby) getEmbyLibraryIDByItem(item *schemas.RefreshMediaItem) string {
	if item.Title == "" || item.Year == "" || item.Type == "" {
		return ""
	}
	
	if item.Type != string(types.MediaTypeMovie) {
		itemID := e.getEmbySeriesIDByName(item.Title, item.Year)
		if itemID != "" {
			// 存在电视剧，则直接刷新这个电视剧就行
			return itemID
		}
	} else {
		if len(e.GetMovies(item.Title, item.Year, nil)) > 0 {
			// 已存在，不用刷新
			return ""
		}
	}
	
	// 查找需要刷新的媒体库ID
	itemPath := item.TargetPath
	// 匹配子目�?	for _, folder := range e.folders {
		if subFolders, ok := folder["SubFolders"].([]interface{}); ok {
			for _, subfolder := range subFolders {
				if subfolderMap, ok := subfolder.(map[string]interface{}); ok {
					if path, pathOk := subfolderMap["Path"].(string); pathOk {
						// 匹配子目�?						subfolderPath := path
						// 检查路径是否匹�?						rel, err := filepath.Rel(subfolderPath, itemPath)
						if err == nil && !strings.Contains(rel, "..") {
							if id, idOk := folder["Id"].(string); idOk {
								return id
							}
						}
					}
				}
			}
		}
	}
	
	// 如果找不到，只要路径中有分类目录名就命中
	for _, folder := range e.folders {
		if subFolders, ok := folder["SubFolders"].([]interface{}); ok {
			for _, subfolder := range subFolders {
				if subfolderMap, ok := subfolder.(map[string]interface{}); ok {
					if path, pathOk := subfolderMap["Path"].(string); pathOk {
						re := regexp.MustCompile(fmt.Sprintf(`[/\\]%s`, item.Category))
						if re.MatchString(path) {
							if id, idOk := folder["Id"].(string); idOk {
								return id
							}
						}
					}
				}
			}
		}
	}
	
	// 刷新根目�?	return "/"
}

// formatItemInfo 格式化item
func (e *Emby) formatItemInfo(item map[string]interface{}) *schemas.MediaServerItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("格式化Emby项目信息时出�? %v", r)
		}
	}()
	
	var userState *schemas.MediaServerItemUserState
	if userData, ok := item["UserData"].(map[string]interface{}); ok && userData != nil {
		var resume bool
		if playbackPos, ok := userData["PlaybackPositionTicks"].(float64); ok {
			resume = playbackPos > 0
		}
		
		var lastPlayedDate *time.Time
		if lastPlayedStr, ok := userData["LastPlayedDate"].(string); ok {
			if strings.Contains(lastPlayedStr, ".") {
				lastPlayedStr = strings.Split(lastPlayedStr, ".")[0]
			}
			
			if parsedTime, err := time.Parse("2006-01-02T15:04:05", lastPlayedStr); err == nil {
				lastPlayedDate = &parsedTime
			}
		}
		
		userState = &schemas.MediaServerItemUserState{
			Played:         getBoolValue(userData, "Played"),
			Resume:         resume,
			LastPlayedDate: lastPlayedDate,
			PlayCount:      getIntValue(userData, "PlayCount"),
			Percentage:     getFloatValue(userData, "PlayedPercentage"),
		}
	} else {
		userState = nil
	}
	
	var tmdbID int
	if providerIds, ok := item["ProviderIds"].(map[string]interface{}); ok {
		if tmdbStr, tmdbOk := providerIds["Tmdb"].(string); tmdbOk {
			if tmdbInt, err := strconv.Atoi(tmdbStr); err == nil {
				tmdbID = tmdbInt
			}
		}
	}
	
	var imdbID *string
	if providerIds, ok := item["ProviderIds"].(map[string]interface{}); ok {
		if imdbStr, imdbOk := providerIds["Imdb"].(string); imdbOk {
			imdbID = &imdbStr
		}
	}
	
	var tvdbID *int
	if providerIds, ok := item["ProviderIds"].(map[string]interface{}); ok {
		if tvdbStr, tvdbOk := providerIds["Tvdb"].(string); tvdbOk {
			if tvdbInt, err := strconv.Atoi(tvdbStr); err == nil {
				tvdbID = &tvdbInt
			}
		}
	}
	
	mediaItem := &schemas.MediaServerItem{
		Server:        "emby",
		Library:       getStringValue(item, "ParentId"),
		ItemID:        getStringValue(item, "Id"),
		ItemType:      getStringValue(item, "Type"),
		Title:         getStringValue(item, "Name"),
		OriginalTitle: getStringValue(item, "OriginalTitle"),
		Year:          getIntValue(item, "ProductionYear"),
		TmdbID:        tmdbID,
		ImdbID:        imdbID,
		TvdbID:        tvdbID,
		Path:          getStringValue(item, "Path"),
		UserState:     userState,
	}
	
	return mediaItem
}

// GetItemInfo 获取单个项目详情
func (e *Emby) GetItemInfo(itemid string) *schemas.MediaServerItem {
	if itemid == "" {
		return nil
	}
	
	if e.host == "" || e.apikey == "" {
		return nil
	}
	
	url := fmt.Sprintf("%s/emby/Users/%v/Items/%s", e.host, e.user, itemid)
	params := map[string]string{
		"api_key": e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接/Users/%v/Items/%s出错�?s", e.user, itemid, err.Error())
		return nil
	}
	
	if res != nil && res.StatusCode == 200 {
		defer res.Body.Close()
		var item map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&item); err != nil {
			return nil
		}
		
		iteminfo := e.formatItemInfo(item)
		return iteminfo
	}
	
	return nil
}

// GetItems 获取媒体服务器项目列表，支持分页和不分页逻辑，默认不分页获取所有数�?func (e *Emby) GetItems(parent string, startIndex, limit int) []*schemas.MediaServerItem {
	if parent == "" || e.host == "" || e.apikey == "" {
		return nil
	}
	
	url := fmt.Sprintf("%s/emby/Users/%v/Items", e.host, e.user)
	params := map[string]string{
		"ParentId": parent,
		"api_key":  e.apikey,
		"Fields":   "ProviderIds,OriginalTitle,ProductionYear,Path,UserDataPlayCount,UserDataLastPlayedDate,ParentId",
	}
	
	if limit != -1 {
		params["StartIndex"] = strconv.Itoa(startIndex)
		params["Limit"] = strconv.Itoa(limit)
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Users/Items出错�?s", err.Error())
		return nil
	}
	
	if res != nil && res.StatusCode == 200 {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return nil
		}
		
		if items, ok := result["Items"].([]interface{}); ok {
			mediaItems := []*schemas.MediaServerItem{}
			for _, item := range items {
				if item == nil {
					continue
				}
				
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				
				if itemType, typeOk := itemMap["Type"].(string); typeOk {
					if strings.Contains(itemType, "Folder") {
						// 递归处理文件�?						if id, idOk := itemMap["Id"].(string); idOk {
							subItems := e.GetItems(id, 0, -1)
							mediaItems = append(mediaItems, subItems...)
						}
					} else if itemType == "Movie" || itemType == "Series" {
						formattedItem := e.formatItemInfo(itemMap)
						if formattedItem != nil {
							mediaItems = append(mediaItems, formattedItem)
						}
					}
				}
			}
			return mediaItems
		}
	}
	
	return nil
}

// GetWebhookMessage 解析Emby Webhook报文
func (e *Emby) GetWebhookMessage(form map[string]interface{}, args map[string]interface{}) *schemas.WebhookEventInfo {
	var message map[string]interface{}
	
	if form != nil && form["data"] != nil {
		if dataStr, ok := form["data"].(string); ok {
			if err := json.Unmarshal([]byte(dataStr), &message); err != nil {
				logger.Debugf("解析emby webhook报文出错�?s", err.Error())
				return nil
			}
		}
	} else if len(args) > 0 {
		argsJSON, _ := json.Marshal(args)
		if err := json.Unmarshal(argsJSON, &message); err != nil {
			logger.Debugf("解析emby webhook报文出错�?s", err.Error())
			return nil
		}
	} else {
		return nil
	}
	
	eventType, ok := message["Event"].(string)
	if !ok || eventType == "" {
		return nil
	}
	
	logger.Debugf("接收到emby webhook�?v", message)
	
	eventItem := &schemas.WebhookEventInfo{
		Event:   eventType,
		Channel: "emby",
	}
	
	if item, itemOk := message["Item"].(map[string]interface{}); itemOk {
		eventItem.MediaType = getStringValue(item, "Type")
		
		if itemType, typeOk := item["Type"].(string); typeOk {
			if itemType == "Episode" || itemType == "Series" || itemType == "Season" {
				eventItem.ItemType = "TV"
				
				seriesName := getStringValue(item, "SeriesName")
				parentIndex := getIntValue(item, "ParentIndexNumber")
				indexNumber := getIntValue(item, "IndexNumber")
				itemName := getStringValue(item, "Name")
				
				if seriesName != "" && parentIndex > 0 && indexNumber > 0 {
					eventItem.ItemName = fmt.Sprintf("%s S%dE%d %s", seriesName, parentIndex, indexNumber, itemName)
				} else if seriesName != "" {
					eventItem.ItemName = fmt.Sprintf("%s %s", seriesName, itemName)
				} else {
					eventItem.ItemName = itemName
				}
				
				eventItem.ItemID = getStringValue(item, "SeriesId")
				eventItem.SeasonID = parentIndex
				eventItem.EpisodeID = indexNumber
			} else if itemType == "Audio" {
				eventItem.ItemType = "AUD"
				album := getStringValue(item, "Album")
				fileName := getStringValue(item, "FileName")
				eventItem.ItemName = album
				eventItem.Overview = fileName
				eventItem.ItemID = getStringValue(item, "AlbumId")
			} else {
				eventItem.ItemType = "MOV"
				productionYear := getIntValue(item, "ProductionYear")
				eventItem.ItemName = fmt.Sprintf("%s (%d)", getStringValue(item, "Name"), productionYear)
				eventItem.ItemID = getStringValue(item, "Id")
			}
		}
		
		eventItem.ItemPath = getStringValue(item, "Path")
		
		if providerIds, ok := item["ProviderIds"].(map[string]interface{}); ok {
			if tmdbID, tmdbOk := providerIds["Tmdb"].(string); tmdbOk {
				eventItem.TmdbID = tmdbID
			}
		}
		
		if overview, overviewOk := item["Overview"].(string); overviewOk {
			if len(overview) > 100 {
				eventItem.Overview = overview[:100] + "..."
			} else {
				eventItem.Overview = overview
			}
		}
		
		if transcodingInfo, transcodingOk := message["TranscodingInfo"].(map[string]interface{}); transcodingOk {
			if completionPercentage, completionOk := transcodingInfo["CompletionPercentage"].(float64); completionOk {
				eventItem.Percentage = &completionPercentage
			}
		}
		
		if eventItem.Percentage == nil {
			if playbackInfo, playbackOk := message["PlaybackInfo"].(map[string]interface{}); playbackOk {
				if positionTicks, positionOk := playbackInfo["PositionTicks"].(float64); positionOk {
					if runTimeTicks, runtimeOk := item["RunTimeTicks"].(float64); runtimeOk && runTimeTicks > 0 {
						percentage := positionTicks / runTimeTicks * 100
						eventItem.Percentage = &percentage
					}
				}
			}
		}
	}
	
	if session, sessionOk := message["Session"].(map[string]interface{}); sessionOk {
		eventItem.IP = getStringValue(session, "RemoteEndPoint")
		eventItem.DeviceName = getStringValue(session, "DeviceName")
		eventItem.Client = getStringValue(session, "Client")
	}
	
	if user, userOk := message["User"].(map[string]interface{}); userOk {
		eventItem.UserName = getStringValue(user, "Name")
	}
	
	if itemIsVirtual, isVirtualOk := message["item_isvirtual"].(bool); isVirtualOk {
		eventItem.ItemIsvirtual = &itemIsVirtual
		eventItem.ItemType = getStringValue(message, "item_type")
		eventItem.ItemName = getStringValue(message, "item_name")
		eventItem.ItemPath = getStringValue(message, "item_path")
		eventItem.TmdbID = getStringValue(message, "tmdb_id")
		eventItem.SeasonID = getIntValue(message, "season_id")
		eventItem.EpisodeID = getIntValue(message, "episode_id")
	}
	
	// 获取消息图片
	if eventItem.ItemID != "" {
		// 根据返回的item_id去调用媒体服务器获取
		eventItem.ImageURL = e.GetRemoteImageByID(itemID: eventItem.ItemID, imageType: "Backdrop")
	}
	
	eventItem.JSONObject = message
	
	return eventItem
}

// GetData 自定义URL从媒体服务器获取数据，其中[HOST]、[APIKEY]、[USER]会被替换成实际的�?func (e *Emby) GetData(url string) *http.Response {
	if e.host == "" || e.apikey == "" {
		return nil
	}
	
	url = strings.Replace(url, "[HOST]", e.host, -1)
	url = strings.Replace(url, "[APIKEY]", e.apikey, -1)
	url = strings.Replace(url, "[USER]", fmt.Sprintf("%v", e.user), -1)
	
	res, err := utils.RequestUtils.GetRes(url, nil, nil, 0)
	if err != nil {
		logger.Errorf("连接Emby出错�?s", err.Error())
		return nil
	}
	
	return res
}

// PostData 自定义URL从媒体服务器获取数据，其中[HOST]、[APIKEY]、[USER]会被替换成实际的�?func (e *Emby) PostData(url, data string, headers map[string]string) *http.Response {
	if e.host == "" || e.apikey == "" {
		return nil
	}
	
	url = strings.Replace(url, "[HOST]", e.host, -1)
	url = strings.Replace(url, "[APIKEY]", e.apikey, -1)
	url = strings.Replace(url, "[USER]", fmt.Sprintf("%v", e.user), -1)
	
	res, err := utils.RequestUtils.PostRes(url, headers, data, 0)
	if err != nil {
		logger.Errorf("连接Emby出错�?s", err.Error())
		return nil
	}
	
	return res
}

// GetPlayURL 拼装媒体播放链接
func (e *Emby) GetPlayURL(itemID string) string {
	return fmt.Sprintf("%s/web/index.html#!/item?id=%s&context=home&serverId=%s", 
		coalesceString(e.playhost, e.host), itemID, e.serverid)
}

// GetBackdropURL 获取Emby的Backdrop图片地址
func (e *Emby) GetBackdropURL(itemID, imageTag string, remote bool) string {
	if e.host == "" || e.apikey == "" {
		return ""
	}
	
	if imageTag == "" || itemID == "" {
		return ""
	}
	
	hostURL := e.host
	if remote {
		hostURL = coalesceString(e.playhost, e.host)
	}
	
	return fmt.Sprintf("%s/Items/%s/Images/Backdrop?tag=%s&api_key=%s", hostURL, itemID, imageTag, e.apikey)
}

// getLocalImageByID 根据ItemId从媒体服务器查询本地图片地址
func (e *Emby) getLocalImageByID(itemID string) string {
	if e.host == "" || e.apikey == "" {
		return ""
	}
	
	return fmt.Sprintf("%s/Items/%s/Images/Primary", e.host, itemID)
}

// GetResume 获得继续观看
func (e *Emby) GetResume(num int, username string) []*schemas.MediaServerPlayItem {
	if e.host == "" || e.apikey == "" {
		return nil
	}
	
	var user interface{}
	if username != "" {
		user = e.GetUser(username)
	} else {
		user = e.user
	}
	
	url := fmt.Sprintf("%s/Users/%v/Items/Resume", e.host, user)
	params := map[string]string{
		"Limit":      "100",
		"MediaTypes": "Video",
		"Fields":     "ProductionYear,Path",
		"api_key":    e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Users/Items/Resume出错�?s", err.Error())
		return []*schemas.MediaServerPlayItem{}
	}
	
	if res != nil {
		defer res.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("Users/Items/Resume 未获取到返回数据")
			return []*schemas.MediaServerPlayItem{}
		}
		
		if items, ok := result["Items"].([]interface{}); ok {
			retResume := []*schemas.MediaServerPlayItem{}
			// 用户媒体库文件夹列表（排除黑名单�?			libraryFolders := e.GetUserLibraryFolders()
			
			for _, item := range items {
				if len(retResume) >= num {
					break
				}
				
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				
				itemType, typeOk := itemMap["Type"].(string)
				if !typeOk || (itemType != "Movie" && itemType != "Episode") {
					continue
				}
				
				itemPath := getStringValue(itemMap, "Path")
				if itemPath != "" && len(libraryFolders) > 0 {
					match := false
					for _, folder := range libraryFolders {
						if strings.HasPrefix(itemPath, folder) {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}
				
				var mediaType string
				if itemType == "Movie" {
					mediaType = string(types.MediaTypeMovie)
				} else {
					mediaType = string(types.MediaTypeTV)
				}
				
				link := e.GetPlayURL(getStringValue(itemMap, "Id"))
				
				var title, subtitle string
				var image string
				
				if mediaType == string(types.MediaTypeMovie) {
					title = getStringValue(itemMap, "Name")
					subtitle = strconv.Itoa(getIntValue(itemMap, "ProductionYear"))
					
					if backdropImageTags, backdropOk := itemMap["BackdropImageTags"].([]interface{}); backdropOk && len(backdropImageTags) > 0 {
						if tag, tagOk := backdropImageTags[0].(string); tagOk {
							image = e.GetBackdropURL(getStringValue(itemMap, "Id"), tag, false)
						}
					}
					
					if image == "" {
						image = e.getLocalImageByID(getStringValue(itemMap, "Id"))
					}
				} else {
					title = getStringValue(itemMap, "SeriesName")
					subtitle = fmt.Sprintf("S%d:%d - %s", 
						getIntValue(itemMap, "ParentIndexNumber"),
						getIntValue(itemMap, "IndexNumber"),
						getStringValue(itemMap, "Name"))
					
					image = e.GetBackdropURL(getStringValue(itemMap, "SeriesId"), 
						getStringValue(itemMap, "SeriesPrimaryImageTag"), false)
					
					if image == "" {
						image = e.getLocalImageByID(getStringValue(itemMap, "SeriesId"))
					}
				}
				
				var percentage *float64
				if userData, userDataOk := itemMap["UserData"].(map[string]interface{}); userDataOk {
					if playedPercentage, playedOk := userData["PlayedPercentage"].(float64); playedOk {
						percentage = &playedPercentage
					}
				}
				
				retResume = append(retResume, &schemas.MediaServerPlayItem{
					ID:          getStringValue(itemMap, "Id"),
					Title:       title,
					Subtitle:    subtitle,
					Type:        mediaType,
					Image:       image,
					Link:        link,
					Percent:     percentage,
					ServerType:  "emby",
				})
			}
			
			return retResume
		}
	} else {
		logger.Error("Users/Items/Resume 未获取到返回数据")
	}
	
	return []*schemas.MediaServerPlayItem{}
}

// GetLatest 获得最近更�?func (e *Emby) GetLatest(num int, username string) []*schemas.MediaServerPlayItem {
	if e.host == "" || e.apikey == "" {
		return nil
	}
	
	var user interface{}
	if username != "" {
		user = e.GetUser(username)
	} else {
		user = e.user
	}
	
	url := fmt.Sprintf("%s/Users/%v/Items/Latest", e.host, user)
	params := map[string]string{
		"Limit":      "100",
		"MediaTypes": "Video",
		"Fields":     "ProductionYear,Path,BackdropImageTags",
		"api_key":    e.apikey,
	}
	
	res, err := utils.RequestUtils.GetRes(url, params, nil, 0)
	if err != nil {
		logger.Errorf("连接Users/Items/Latest出错�?s", err.Error())
		return []*schemas.MediaServerPlayItem{}
	}
	
	if res != nil {
		defer res.Body.Close()
		var result []map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Error("Users/Items/Latest 未获取到返回数据")
			return []*schemas.MediaServerPlayItem{}
		}
		
		retLatest := []*schemas.MediaServerPlayItem{}
		// 用户媒体库文件夹列表（排除黑名单�?		libraryFolders := e.GetUserLibraryFolders()
		
		for _, item := range result {
			if len(retLatest) >= num {
				break
			}
			
			itemType, typeOk := item["Type"].(string)
			if !typeOk || (itemType != "Movie" && itemType != "Series") {
				continue
			}
			
			itemPath := getStringValue(item, "Path")
			if itemPath != "" && len(libraryFolders) > 0 {
				match := false
				for _, folder := range libraryFolders {
					if strings.HasPrefix(itemPath, folder) {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
			
			var mediaType string
			if itemType == "Movie" {
				mediaType = string(types.MediaTypeMovie)
			} else {
				mediaType = string(types.MediaTypeTV)
			}
			
			link := e.GetPlayURL(getStringValue(item, "Id"))
			image := e.getLocalImageByID(getStringValue(item, "Id"))
			
			var backdropImageTags []interface{}
			if tags, ok := item["BackdropImageTags"].([]interface{}); ok {
				backdropImageTags = tags
			}
			
			retLatest = append(retLatest, &schemas.MediaServerPlayItem{
				ID:                getStringValue(item, "Id"),
				Title:             getStringValue(item, "Name"),
				Subtitle:          strconv.Itoa(getIntValue(item, "ProductionYear")),
				Type:              mediaType,
				Image:             image,
				Link:              link,
				BackdropImageTags: backdropImageTags,
				ServerType:        "emby",
			})
		}
		
		return retLatest
	} else {
		logger.Error("Users/Items/Latest 未获取到返回数据")
	}
	
	return []*schemas.MediaServerPlayItem{}
}

// GetUserLibraryFolders 获取Emby媒体库文件夹列表（排除黑名单�?func (e *Emby) GetUserLibraryFolders() []string {
	if e.host == "" || e.apikey == "" {
		return []string{}
	}
	
	libraryFolders := []string{}
	for _, library := range e.GetEmbyVirtualFolders() {
		if len(e.syncLibraries) > 0 {
			if libraryID, ok := library["Id"].(string); ok {
				if !containsString(e.syncLibraries, libraryID) {
					continue
				}
			}
		}
		
		if paths, ok := library["Path"].([]string); ok {
			libraryFolders = append(libraryFolders, paths...)
		} else if paths, ok := library["Path"].([]interface{}); ok {
			for _, path := range paths {
				if pathStr, ok := path.(string); ok {
					libraryFolders = append(libraryFolders, pathStr)
				}
			}
		}
	}
	
	return libraryFolders
}

// Helper functions
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
		if str, ok := val.(string); ok {
			if i, err := strconv.Atoi(str); err == nil {
				return i
			}
		}
	}
	return 0
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return num
		}
	}
	return 0
}

func getBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func coalesceString(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}
