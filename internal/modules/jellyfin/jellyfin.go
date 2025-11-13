package jellyfin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"moviepilot-go/internal/common"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/entity"
	"moviepilot-go/internal/utils"
)

type Jellyfin struct {
	Host         string
	ApiKey       string
	PlayHost     string
	SyncLibraries []string
	User         string
	ServerId     string
	
	client *resty.Client
}

func NewJellyfin(host string, apiKey string, playHost string, syncLibraries []string) *Jellyfin {
	if host == "" || apiKey == "" {
		common.LOG.Error("Jellyfin服务器配置不完整！！")
		return nil
	}
	
	j := &Jellyfin{
		Host:          utils.StandardizeBaseURL(host),
		PlayHost:      utils.StandardizeBaseURL(playHost),
		ApiKey:        apiKey,
		SyncLibraries: syncLibraries,
		client:        resty.New(),
	}
	
	j.User = j.GetUser(config.CONFIG.SUPERUSER)
	j.ServerId = j.GetServerId()
	
	return j
}

func (j *Jellyfin) IsInactive() bool {
	if j.Host == "" || j.ApiKey == "" {
		return false
	}
	return j.User == ""
}

func (j *Jellyfin) Reconnect() {
	j.User = j.GetUser("")
	j.ServerId = j.GetServerId()
}

func (j *Jellyfin) GetJellyfinFolders() []map[string]interface{} {
	if j.Host == "" || j.ApiKey == "" {
		return []map[string]interface{}{}
	}
	
	url := fmt.Sprintf("%sLibrary/SelectableMediaFolders", j.Host)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Library/SelectableMediaFolders 出错�?v", err)
		return []map[string]interface{}{}
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Library/SelectableMediaFolders 未获取到返回数据")
		return []map[string]interface{}{}
	}
	
	var result []map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Library/SelectableMediaFolders返回数据出错�?v", err)
		return []map[string]interface{}{}
	}
	
	return result
}

func (j *Jellyfin) GetJellyfinVirtualFolders() []map[string]interface{} {
	if j.Host == "" || j.ApiKey == "" {
		return []map[string]interface{}{}
	}
	
	url := fmt.Sprintf("%sLibrary/VirtualFolders", j.Host)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Library/VirtualFolders 出错�?v", err)
		return []map[string]interface{}{}
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Library/VirtualFolders 未获取到返回数据")
		return []map[string]interface{}{}
	}
	
	var libraryItems []map[string]interface{}
	err = json.Unmarshal(resp.Body(), &libraryItems)
	if err != nil {
		common.LOG.Errorf("解析Library/VirtualFolders返回数据出错�?v", err)
		return []map[string]interface{}{}
	}
	
	var libraries []map[string]interface{}
	for _, libraryItem := range libraryItems {
		libraryId := libraryItem["ItemId"]
		libraryName := libraryItem["Name"]
		pathInfos := libraryItem["LibraryOptions"].(map[string]interface{})["PathInfos"].([]interface{})
		
		var libraryPaths []string
		for _, pathInfo := range pathInfos {
			pathMap := pathInfo.(map[string]interface{})
			if pathMap["NetworkPath"] != nil && pathMap["NetworkPath"] != "" {
				libraryPaths = append(libraryPaths, pathMap["NetworkPath"].(string))
			} else {
				libraryPaths = append(libraryPaths, pathMap["Path"].(string))
			}
		}
		
		if libraryName != nil && len(libraryPaths) > 0 {
			libraries = append(libraries, map[string]interface{}{
				"Id":   libraryId,
				"Name": libraryName,
				"Path": libraryPaths,
			})
		}
	}
	
	return libraries
}

func (j *Jellyfin) getJellyfinLibrarys(username string) []map[string]interface{} {
	if j.Host == "" || j.ApiKey == "" {
		return []map[string]interface{}{}
	}
	
	user := j.User
	if username != "" {
		user = j.GetUser(username)
	}
	
	url := fmt.Sprintf("%sUsers/%s/Views", j.Host, user)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users/Views 出错�?v", err)
		return []map[string]interface{}{}
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Users/Views 未获取到返回数据")
		return []map[string]interface{}{}
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Users/Views返回数据出错�?v", err)
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
}

func (j *Jellyfin) GetLibrarys(username string, hidden bool) []entity.MediaServerLibrary {
	if j.Host == "" || j.ApiKey == "" {
		return []entity.MediaServerLibrary{}
	}
	
	libraries := []entity.MediaServerLibrary{}
	rawLibraries := j.getJellyfinLibrarys(username)
	
	for _, library := range rawLibraries {
		// Check if library should be hidden
		if hidden && len(j.SyncLibraries) > 0 && !utils.ContainsString(j.SyncLibraries, "all") {
			if library["Id"] != nil && !utils.ContainsString(j.SyncLibraries, library["Id"].(string)) {
				continue
			}
		}
		
		var libraryType string
		var link string
		
		collectionType, _ := library["CollectionType"].(string)
		id, _ := library["Id"].(string)
		name, _ := library["Name"].(string)
		path, _ := library["Path"].(string)
		
		switch collectionType {
		case "movies":
			libraryType = "movie"
			link = fmt.Sprintf("%sweb/index.html#!/movies.html?topParentId=%s", j.PlayHostOrHost(), id)
		case "tvshows":
			libraryType = "tv"
			link = fmt.Sprintf("%sweb/index.html#!/tv.html?topParentId=%s", j.PlayHostOrHost(), id)
		default:
			libraryType = "unknown"
			link = fmt.Sprintf("%sweb/index.html#!/library.html?topParentId=%s", j.PlayHostOrHost(), id)
		}
		
		image := j.getLocalImageById(id)
		
		lib := entity.MediaServerLibrary{
			Server:     "jellyfin",
			Id:         id,
			Name:       name,
			Path:       path,
			Type:       libraryType,
			Image:      image,
			Link:       link,
			ServerType: "jellyfin",
		}
		
		libraries = append(libraries, lib)
	}
	
	return libraries
}

func (j *Jellyfin) GetItems(parent string, startIndex int, limit int) []*entity.MediaServerItem {
	if parent == "" || j.Host == "" || j.ApiKey == "" {
		return nil
	}
	
	url := fmt.Sprintf("%sUsers/%s/Items", j.Host, j.User)
	params := map[string]string{
		"ParentId": parent,
		"api_key":  j.ApiKey,
		"Fields":   "ProviderIds,OriginalTitle,ProductionYear,Path,UserDataPlayCount,UserDataLastPlayedDate,ParentId",
	}
	
	if limit != -1 {
		params["StartIndex"] = strconv.Itoa(startIndex)
		params["Limit"] = strconv.Itoa(limit)
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users/Items出错�?v", err)
		return nil
	}
	
	if resp.StatusCode() != http.StatusOK {
		return nil
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Users/Items返回数据出错�?v", err)
		return nil
	}
	
	items, ok := result["Items"].([]interface{})
	if !ok {
		return []*entity.MediaServerItem{}
	}
	
	var mediaItems []*entity.MediaServerItem
	for _, item := range items {
		if item == nil {
			continue
		}
		
		itemMap := item.(map[string]interface{})
		itemType, _ := itemMap["Type"].(string)
		
		if strings.Contains(itemType, "Folder") {
			// Recursive call for folders
			if id, ok := itemMap["Id"].(string); ok {
				folderItems := j.GetItems(id, 0, -1)
				mediaItems = append(mediaItems, folderItems...)
			}
		} else if itemType == "Movie" || itemType == "Series" {
			if mediaItem := j.formatItemInfo(itemMap); mediaItem != nil {
				mediaItems = append(mediaItems, mediaItem)
			}
		}
	}
	
	return mediaItems
}

func (j *Jellyfin) GetResume(num int, username string) []entity.MediaServerPlayItem {
	if j.Host == "" || j.ApiKey == "" {
		return nil
	}
	
	user := j.User
	if username != "" {
		user = j.GetUser(username)
	}
	
	url := fmt.Sprintf("%sUsers/%s/Items/Resume", j.Host, user)
	params := map[string]string{
		"Limit":      "100",
		"MediaTypes": "Video",
		"Fields":     "ProductionYear,Path",
		"api_key":    j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users/Items/Resume出错�?v", err)
		return []entity.MediaServerPlayItem{}
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Users/Items/Resume 未获取到返回数据")
		return []entity.MediaServerPlayItem{}
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Users/Items/Resume返回数据出错�?v", err)
		return []entity.MediaServerPlayItem{}
	}
	
	items, ok := result["Items"].([]interface{})
	if !ok {
		return []entity.MediaServerPlayItem{}
	}
	
	retResume := []entity.MediaServerPlayItem{}
	// 用户媒体库文件夹列表（排除黑名单�?	libraryFolders := j.GetUserLibraryFolders()
	
	for _, item := range items {
		if len(retResume) >= num {
			break
		}
		
		itemMap := item.(map[string]interface{})
		itemType, _ := itemMap["Type"].(string)
		
		if itemType != "Movie" && itemType != "Episode" {
			continue
		}
		
		itemPath, _ := itemMap["Path"].(string)
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
			mediaType = "movie"
		} else {
			mediaType = "tv"
		}
		
		itemId, _ := itemMap["Id"].(string)
		link := j.GetPlayUrl(itemId)
		
		var image string
		if backdropTags, ok := itemMap["BackdropImageTags"].([]interface{}); ok && len(backdropTags) > 0 {
			if tag, ok := backdropTags[0].(string); ok {
				image = j.GetBackdropUrl(itemId, tag, false)
			}
		} else {
			image = j.getLocalImageById(itemId)
		}
		
		// 小部分剧集无[xxx-S01E01-thumb.jpg]图片
		imageResp, _ := j.client.R().Get(image)
		if imageResp != nil && imageResp.StatusCode() == 404 {
			image = j.GenerateImageLink(itemId, "Backdrop", false)
		}
		
		var title, subtitle string
		itemName, _ := itemMap["Name"].(string)
		if mediaType == "movie" {
			title = itemName
			if prodYear, ok := itemMap["ProductionYear"].(float64); ok {
				subtitle = strconv.Itoa(int(prodYear))
			}
		} else {
			if seriesName, ok := itemMap["SeriesName"].(string); ok {
				title = seriesName
			}
			if seasonNum, ok := itemMap["ParentIndexNumber"].(float64); ok {
				if epNum, ok := itemMap["IndexNumber"].(float64); ok {
					if epName, ok := itemMap["Name"].(string); ok {
						subtitle = fmt.Sprintf("S%d:%d - %s", int(seasonNum), int(epNum), epName)
					}
				}
			}
		}
		
		var percent *float64
		if userData, ok := itemMap["UserData"].(map[string]interface{}); ok {
			if playedPct, ok := userData["PlayedPercentage"].(float64); ok {
				percent = &playedPct
			}
		}
		
		retResume = append(retResume, entity.MediaServerPlayItem{
			Id:         itemId,
			Title:      title,
			Subtitle:   subtitle,
			Type:       mediaType,
			Image:      image,
			Link:       link,
			Percent:    percent,
			ServerType: "jellyfin",
		})
	}
	
	return retResume
}

func (j *Jellyfin) GetUserCount() int {
	if j.Host == "" || j.ApiKey == "" {
		return 0
	}
	
	url := fmt.Sprintf("%sUsers", j.Host)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users出错�?v", err)
		return 0
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Users 未获取到返回数据")
		return 0
	}
	
	var users []interface{}
	err = json.Unmarshal(resp.Body(), &users)
	if err != nil {
		common.LOG.Errorf("解析Users返回数据出错�?v", err)
		return 0
	}
	
	return len(users)
}

func (j *Jellyfin) GetUser(userName string) string {
	if j.Host == "" || j.ApiKey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%sUsers", j.Host)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users出错�?v", err)
		return ""
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Users 未获取到返回数据")
		return ""
	}
	
	var users []map[string]interface{}
	err = json.Unmarshal(resp.Body(), &users)
	if err != nil {
		common.LOG.Errorf("解析Users返回数据出错�?v", err)
		return ""
	}
	
	// First check if there's a match with userName
	if userName != "" {
		for _, user := range users {
			if name, ok := user["Name"].(string); ok && name == userName {
				if id, ok := user["Id"].(string); ok {
					return id
				}
			}
		}
	}
	
	// Check for admin user
	for _, user := range users {
		if policy, ok := user["Policy"].(map[string]interface{}); ok {
			if isAdmin, ok := policy["IsAdministrator"].(bool); ok && isAdmin {
				if id, ok := user["Id"].(string); ok {
					return id
				}
			}
		}
	}
	
	return ""
}

func (j *Jellyfin) Authenticate(username, password string) string {
	if j.Host == "" || j.ApiKey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%sUsers/authenticatebyname", j.Host)
	
	authHeader := fmt.Sprintf(`MediaBrowser Client="MoviePilot", Device="requests", DeviceId="1", Version="1.0.0", Token="%s"`, j.ApiKey)
	
	requestData := map[string]string{
		"Username": username,
		"Pw":       password,
	}
	
	requestBody, _ := json.Marshal(requestData)
	
	resp, err := j.client.R().
		SetHeader("X-Emby-Authorization", authHeader).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(requestBody).
		Post(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users/AuthenticateByName出错�?v", err)
		return ""
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Users/AuthenticateByName 未获取到返回数据")
		return ""
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Users/AuthenticateByName返回数据出错�?v", err)
		return ""
	}
	
	authToken, ok := result["AccessToken"].(string)
	if !ok || authToken == "" {
		return ""
	}
	
	common.LOG.Infof("用户 %s Jellyfin认证成功", username)
	return authToken
}

func (j *Jellyfin) GetServerId() string {
	if j.Host == "" || j.ApiKey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%sSystem/Info", j.Host)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接System/Info出错�?v", err)
		return ""
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("System/Info 未获取到返回数据")
		return ""
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析System/Info返回数据出错�?v", err)
		return ""
	}
	
	serverId, ok := result["Id"].(string)
	if !ok {
		return ""
	}
	
	return serverId
}

func (j *Jellyfin) GetMediasCount() entity.Statistic {
	if j.Host == "" || j.ApiKey == "" {
		return entity.Statistic{}
	}
	
	url := fmt.Sprintf("%sItems/Counts", j.Host)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Items/Counts出错�?v", err)
		return entity.Statistic{}
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Items/Counts 未获取到返回数据")
		return entity.Statistic{}
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Items/Counts返回数据出错�?v", err)
		return entity.Statistic{}
	}
	
	stat := entity.Statistic{}
	
	if movieCount, ok := result["MovieCount"].(float64); ok {
		stat.MovieCount = int(movieCount)
	}
	
	if seriesCount, ok := result["SeriesCount"].(float64); ok {
		stat.TvCount = int(seriesCount)
	}
	
	if episodeCount, ok := result["EpisodeCount"].(float64); ok {
		stat.EpisodeCount = int(episodeCount)
	}
	
	return stat
}

func (j *Jellyfin) getJellyfinSeriesIdByName(name, year string) string {
	if j.Host == "" || j.ApiKey == "" || j.User == "" {
		return ""
	}
	
	url := fmt.Sprintf("%sUsers/%s/Items", j.Host, j.User)
	params := map[string]string{
		"IncludeItemTypes": "Series",
		"Recursive":        "true",
		"searchTerm":       name,
		"Limit":            "10",
		"api_key":          j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Items出错�?v", err)
		return ""
	}
	
	if resp.StatusCode() != http.StatusOK {
		return ""
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Items返回数据出错�?v", err)
		return ""
	}
	
	items, ok := result["Items"].([]interface{})
	if !ok {
		return ""
	}
	
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		itemName, _ := itemMap["Name"].(string)
		
		if itemName == name {
			itemYear := ""
			if prodYear, ok := itemMap["ProductionYear"].(float64); ok {
				itemYear = strconv.FormatFloat(prodYear, 'f', -1, 64)
			}
			
			if year == "" || itemYear == year {
				if id, ok := itemMap["Id"].(string); ok {
					return id
				}
			}
		}
	}
	
	return ""
}

func (j *Jellyfin) GetMovies(title, year string, tmdbId int) []entity.MediaServerItem {
	if j.Host == "" || j.ApiKey == "" || j.User == "" {
		return nil
	}
	
	url := fmt.Sprintf("%sUsers/%s/Items", j.Host, j.User)
	params := map[string]string{
		"IncludeItemTypes": "Movie",
		"Fields":           "ProviderIds,OriginalTitle,ProductionYear,Path,UserDataPlayCount,UserDataLastPlayedDate,ParentId",
		"StartIndex":       "0",
		"Recursive":        "true",
		"searchTerm":       title,
		"Limit":            "10",
		"api_key":          j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Items出错�?v", err)
		return nil
	}
	
	if resp.StatusCode() != http.StatusOK {
		return nil
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Items返回数据出错�?v", err)
		return nil
	}
	
	items, ok := result["Items"].([]interface{})
	if !ok {
		return []entity.MediaServerItem{}
	}
	
	retMovies := []entity.MediaServerItem{}
	for _, item := range items {
		if item == nil {
			continue
		}
		
		itemMap := item.(map[string]interface{})
		mediaServerItem := j.formatItemInfo(itemMap)
		
		if mediaServerItem != nil {
			tmdbMatch := tmdbId == 0 || mediaServerItem.Tmdbid == tmdbId
			titleMatch := mediaServerItem.Title == title
			yearMatch := year == "" || strconv.Itoa(mediaServerItem.Year) == year
			
			if tmdbMatch && titleMatch && yearMatch {
				retMovies = append(retMovies, *mediaServerItem)
			}
		}
	}
	
	return retMovies
}

func (j *Jellyfin) GetTvEpisodes(itemId, title, year string, tmdbId, season int) (string, map[int][]int) {
	if j.Host == "" || j.ApiKey == "" || j.User == "" {
		return "", nil
	}
	
	// 查TVID
	if itemId == "" {
		itemId = j.getJellyfinSeriesIdByName(title, year)
		if itemId == "" {
			return "", map[int][]int{}
		}
	}
	
	// 验证tmdbid是否相同
	if itemId != "" {
		itemInfo := j.GetItemInfo(itemId)
		if itemInfo != nil && tmdbId > 0 && itemInfo.Tmdbid > 0 {
			if strconv.Itoa(tmdbId) != strconv.Itoa(itemInfo.Tmdbid) {
				return "", map[int][]int{}
			}
		}
	}
	
	var seasonParam *int
	if season > 0 {
		seasonParam = &season
	}
	
	url := fmt.Sprintf("%sShows/%s/Episodes", j.Host, itemId)
	params := map[string]string{
		"userId":   j.User,
		"isMissing": "false",
		"api_key":   j.ApiKey,
	}
	
	if seasonParam != nil {
		params["season"] = strconv.Itoa(*seasonParam)
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Shows/Id/Episodes出错�?v", err)
		return "", nil
	}
	
	if resp.StatusCode() != http.StatusOK {
		return "", map[int][]int{}
	}
	
	var tvInfo map[string]interface{}
	err = json.Unmarshal(resp.Body(), &tvInfo)
	if err != nil {
		common.LOG.Errorf("解析Shows/Id/Episodes返回数据出错�?v", err)
		return "", map[int][]int{}
	}
	
	items, ok := tvInfo["Items"].([]interface{})
	if !ok {
		return "", map[int][]int{}
	}
	
	// 返回的季集信�?	seasonEpisodes := make(map[int][]int)
	for _, item := range items {
		resItem := item.(map[string]interface{})
		
		seasonIndexFloat, ok := resItem["ParentIndexNumber"].(float64)
		if !ok {
			continue
		}
		seasonIndex := int(seasonIndexFloat)
		
		if seasonParam != nil && *seasonParam != seasonIndex {
			continue
		}
		
		episodeIndexFloat, ok := resItem["IndexNumber"].(float64)
		if !ok {
			continue
		}
		episodeIndex := int(episodeIndexFloat)
		
		if seasonEpisodes[seasonIndex] == nil {
			seasonEpisodes[seasonIndex] = []int{}
		}
		seasonEpisodes[seasonIndex] = append(seasonEpisodes[seasonIndex], episodeIndex)
	}
	
	return itemId, seasonEpisodes
}

func (j *Jellyfin) GetRemoteImageById(itemId, imageType string) string {
	if j.Host == "" || j.ApiKey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%sItems/%s/RemoteImages", j.Host, itemId)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Items/Id/RemoteImages出错�?v", err)
		return j.GenerateImageLink(itemId, imageType, true)
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Info("Items/RemoteImages 未获取到返回数据，采用本地图�?)
		return j.GenerateImageLink(itemId, imageType, true)
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Items/Id/RemoteImages返回数据出错�?v", err)
		return ""
	}
	
	images, ok := result["Images"].([]interface{})
	if !ok {
		return ""
	}
	
	for _, image := range images {
		imageMap := image.(map[string]interface{})
		providerName, _ := imageMap["ProviderName"].(string)
		imgType, _ := imageMap["Type"].(string)
		
		if providerName == "TheMovieDb" && imgType == imageType {
			if imgUrl, ok := imageMap["Url"].(string); ok {
				return imgUrl
			}
		}
	}
	
	return ""
}

func (j *Jellyfin) GetItemPathById(itemId string) string {
	if j.Host == "" || j.ApiKey == "" {
		return ""
	}
	
	url := fmt.Sprintf("%sItems/%s/PlaybackInfo", j.Host, itemId)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Items/Id/PlaybackInfo出错�?v", err)
		return ""
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Items/Id/PlaybackInfo 未获取到返回数据，不设置 Path")
		return ""
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Items/Id/PlaybackInfo返回数据出错�?v", err)
		return ""
	}
	
	mediaSources, ok := result["MediaSources"].([]interface{})
	if !ok || len(mediaSources) == 0 {
		return ""
	}
	
	firstSource := mediaSources[0].(map[string]interface{})
	if path, ok := firstSource["Path"].(string); ok {
		return path
	}
	
	return ""
}

func (j *Jellyfin) GenerateImageLink(itemId, imageType string, hostType bool) string {
	if j.PlayHost == "" {
		common.LOG.Error("Jellyfin外网播放地址未能获取或为�?)
		return ""
	}
	
	// 检测是否为TV
	parentId := j.getItemIdAncestors(itemId, 0, "ParentBackdropItemId")
	if parentId != "" {
		itemId = parentId
	}
	
	host := j.Host
	if hostType {
		host = j.PlayHost
	}
	
	url := fmt.Sprintf("%sItems/%s/Images/%s", host, itemId, imageType)
	
	resp, err := j.client.R().Get(url)
	if err != nil {
		common.LOG.Errorf("连接Items/Id/Images出错�?v", err)
		return ""
	}
	
	if resp.StatusCode() == 404 {
		common.LOG.Errorf("Items/Id/Images 未获取到返回数据或无该影�?s图片", imageType)
		return ""
	}
	
	common.LOG.Infof("影片图片链接:%s", resp.RawResponse.Request.URL.String())
	return resp.RawResponse.Request.URL.String()
}

func (j *Jellyfin) getItemIdAncestors(itemId string, index int, key string) string {
	url := fmt.Sprintf("%sItems/%s/Ancestors", j.Host, itemId)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Items/Id/Ancestors出错�?v", err)
		return ""
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Items/Id/Ancestors 未获取到返回数据")
		return ""
	}
	
	var result []map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Items/Id/Ancestors返回数据出错�?v", err)
		return ""
	}
	
	if len(result) <= index {
		return ""
	}
	
	if val, ok := result[index][key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	
	return ""
}

func (j *Jellyfin) RefreshRootLibrary() bool {
	if j.Host == "" || j.ApiKey == "" {
		return false
	}
	
	url := fmt.Sprintf("%sLibrary/Refresh", j.Host)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Post(url)
		
	if err != nil {
		common.LOG.Errorf("连接Library/Refresh出错�?v", err)
		return false
	}
	
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		common.LOG.Info("刷新媒体库失败，无法连接Jellyfin�?)
		return false
	}
	
	return true
}

func (j *Jellyfin) GetWebhookMessage(body []byte) *entity.WebhookEventInfo {
	if len(body) == 0 {
		return nil
	}
	
	var message map[string]interface{}
	err := json.Unmarshal(body, &message)
	if err != nil {
		common.LOG.Debugf("解析Jellyfin Webhook报文出错�?v", err)
		return nil
	}
	
	if message == nil {
		return nil
	}
	
	common.LOG.Debugf("接收到jellyfin webhook�?v", message)
	
	eventType, ok := message["NotificationType"].(string)
	if !ok {
		return nil
	}
	
	eventItem := &entity.WebhookEventInfo{
		Event:   eventType,
		Channel: "jellyfin",
	}
	
	// Extract fields from message
	if val, ok := message["ItemId"].(string); ok {
		eventItem.ItemId = val
	}
	
	if val, ok := message["Provider_tmdb"].(string); ok {
		eventItem.TmdbId = val
	}
	
	if val, ok := message["Overview"].(string); ok {
		eventItem.Overview = val
	}
	
	if val, ok := message["Favorite"].(bool); ok {
		eventItem.ItemFavorite = &val
	}
	
	if val, ok := message["SaveReason"].(string); ok {
		eventItem.SaveReason = val
	}
	
	if val, ok := message["DeviceName"].(string); ok {
		eventItem.DeviceName = val
	}
	
	if val, ok := message["NotificationUsername"].(string); ok {
		eventItem.UserName = val
	}
	
	if val, ok := message["ClientName"].(string); ok {
		eventItem.Client = val
	}
	
	if itemType, ok := message["ItemType"].(string); ok {
		eventItem.MediaType = itemType
		
		switch itemType {
		case "Episode", "Series", "Season":
			// 剧集
			eventItem.ItemType = "TV"
			
			if val, ok := message["SeasonNumber"].(float64); ok {
				seasonId := int(val)
				eventItem.SeasonId = &seasonId
			}
			
			if val, ok := message["EpisodeNumber"].(float64); ok {
				episodeId := int(val)
				eventItem.EpisodeId = &episodeId
			}
			
			var seriesName, name string
			if val, ok := message["SeriesName"].(string); ok {
				seriesName = val
			}
			if val, ok := message["Name"].(string); ok {
				name = val
			}
			
			eventItem.ItemName = fmt.Sprintf("%s S%dE%d %s", seriesName, *eventItem.SeasonId, *eventItem.EpisodeId, name)
			
		case "Audio":
			// 音乐
			eventItem.ItemType = "AUD"
			
			if val, ok := message["Album"].(string); ok {
				eventItem.ItemName = val
			}
			
			if val, ok := message["Name"].(string); ok {
				eventItem.Overview = val
			}
			
			if val, ok := message["ItemId"].(string); ok {
				eventItem.ItemId = val
			}
			
		default:
			// 电影
			eventItem.ItemType = "MOV"
			
			var itemName string
			var year string
			
			if val, ok := message["Name"].(string); ok {
				itemName = val
			}
			
			if val, ok := message["Year"].(float64); ok {
				year = fmt.Sprintf("(%d)", int(val))
			}
			
			eventItem.ItemName = fmt.Sprintf("%s %s", itemName, year)
		}
	}
	
	if playbackPos, ok1 := message["PlaybackPositionTicks"].(float64); ok1 {
		if runtime, ok2 := message["RunTimeTicks"].(float64); ok2 {
			eventItem.Percentage = &playbackPos
			temp := playbackPos / runtime * 100
			eventItem.Percentage = &temp
		}
	}
	
	// 获取消息图片
	if eventItem.ItemId != "" {
		// 根据返回的item_id去调用媒体服务器获取
		eventItem.ImageUrl = j.GetRemoteImageById(eventItem.ItemId, "Backdrop")
		// jellyfin �?webhook 不含 item_path，需要单独获�?		eventItem.ItemPath = j.GetItemPathById(eventItem.ItemId)
	}
	
	eventItem.JSONObject = message
	
	return eventItem
}

func (j *Jellyfin) formatItemInfo(item map[string]interface{}) *entity.MediaServerItem {
	defer func() {
		if r := recover(); r != nil {
			common.LOG.Errorf("formatItemInfo panic: %v", r)
		}
	}()
	
	var userState *entity.MediaServerItemUserState
	
	if userData, ok := item["UserData"].(map[string]interface{}); ok && userData != nil {
		resume := false
		if pos, ok := userData["PlaybackPositionTicks"].(float64); ok && pos > 0 {
			resume = true
		}
		
		var lastPlayedDate *time.Time
		if lastPlayedStr, ok := userData["LastPlayedDate"].(string); ok && lastPlayedStr != "" {
			// 处理时间格式，去除毫秒部�?			if strings.Contains(lastPlayedStr, ".") {
				parts := strings.Split(lastPlayedStr, ".")
				lastPlayedStr = parts[0]
			}
			
			if parsedTime, err := time.Parse("2006-01-02T15:04:05", lastPlayedStr); err == nil {
				lastPlayedDate = &parsedTime
			}
		}
		
		var playCount *int
		if pc, ok := userData["PlayCount"].(float64); ok {
			pcInt := int(pc)
			playCount = &pcInt
		}
		
		var percentage *float64
		if pct, ok := userData["PlayedPercentage"].(float64); ok {
			percentage = &pct
		}
		
		userState = &entity.MediaServerItemUserState{
			Played:         userData["Played"].(bool),
			Resume:         resume,
			LastPlayedDate: lastPlayedDate,
			PlayCount:      playCount,
			Percentage:     percentage,
		}
	}
	
	var tmdbId *int
	if providerIds, ok := item["ProviderIds"].(map[string]interface{}); ok {
		if tmdb, ok := providerIds["Tmdb"].(string); ok {
			if id, err := strconv.Atoi(tmdb); err == nil {
				tmdbId = &id
			}
		}
	}
	
	mediaItem := &entity.MediaServerItem{
		Server:        "jellyfin",
		ItemType:      item["Type"].(string),
		Title:         item["Name"].(string),
		Year:          int(item["ProductionYear"].(float64)),
		Tmdbid:        tmdbId,
		Path:          item["Path"].(string),
		UserState:     userState,
	}
	
	if id, ok := item["Id"].(string); ok {
		mediaItem.ItemId = id
	}
	
	if parentId, ok := item["ParentId"].(string); ok {
		mediaItem.Library = parentId
	}
	
	if originalTitle, ok := item["OriginalTitle"].(string); ok {
		mediaItem.OriginalTitle = &originalTitle
	}
	
	if providerIds, ok := item["ProviderIds"].(map[string]interface{}); ok {
		if imdbId, ok := providerIds["Imdb"].(string); ok {
			mediaItem.Imdbid = &imdbId
		}
		
		if tvdbId, ok := providerIds["Tvdb"].(string); ok {
			mediaItem.Tvdbid = &tvdbId
		}
	}
	
	return mediaItem
}

func (j *Jellyfin) GetItemInfo(itemid string) *entity.MediaServerItem {
	if itemid == "" {
		return nil
	}
	
	if j.Host == "" || j.ApiKey == "" {
		return nil
	}
	
	url := fmt.Sprintf("%sUsers/%s/Items/%s", j.Host, j.User, itemid)
	params := map[string]string{
		"api_key": j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users/%s/Items/%s�?v", j.User, itemid, err)
		return nil
	}
	
	if resp.StatusCode() != http.StatusOK {
		return nil
	}
	
	var item map[string]interface{}
	err = json.Unmarshal(resp.Body(), &item)
	if err != nil {
		common.LOG.Errorf("解析Users/%s/Items/%s返回数据出错�?v", j.User, itemid, err)
		return nil
	}
	
	return j.formatItemInfo(item)
}

func (j *Jellyfin) GetData(rawUrl string) (*resty.Response, error) {
	if j.Host == "" || j.ApiKey == "" {
		return nil, fmt.Errorf("Jellyfin服务器配置不完整")
	}
	
	processedUrl := strings.Replace(rawUrl, "[HOST]", j.Host, -1)
	processedUrl = strings.Replace(processedUrl, "[APIKEY]", j.ApiKey, -1)
	processedUrl = strings.Replace(processedUrl, "[USER]", j.User, -1)
	
	resp, err := j.client.R().
		SetHeader("Accept", "application/json").
		Get(processedUrl)
		
	if err != nil {
		common.LOG.Errorf("连接Jellyfin出错�?v", err)
		return nil, err
	}
	
	return resp, nil
}

func (j *Jellyfin) PostData(rawUrl, data string, headers map[string]string) (*resty.Response, error) {
	if j.Host == "" || j.ApiKey == "" {
		return nil, fmt.Errorf("Jellyfin服务器配置不完整")
	}
	
	processedUrl := strings.Replace(rawUrl, "[HOST]", j.Host, -1)
	processedUrl = strings.Replace(processedUrl, "[APIKEY]", j.ApiKey, -1)
	processedUrl = strings.Replace(processedUrl, "[USER]", j.User, -1)
	
	req := j.client.R()
	
	// 设置headers
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	
	resp, err := req.SetBody(data).Post(processedUrl)
	if err != nil {
		common.LOG.Errorf("连接Jellyfin出错�?v", err)
		return nil, err
	}
	
	return resp, nil
}

func (j *Jellyfin) GetPlayUrl(itemId string) string {
	return fmt.Sprintf("%sweb/index.html#!/details?id=%s&serverId=%s", j.PlayHostOrHost(), itemId, j.ServerId)
}

func (j *Jellyfin) getLocalImageById(itemId string) string {
	if j.Host == "" || j.ApiKey == "" {
		return ""
	}
	return fmt.Sprintf("%sItems/%s/Images/Primary", j.Host, itemId)
}

func (j *Jellyfin) GetBackdropUrl(itemId, imageTag string, remote bool) string {
	if j.Host == "" || j.ApiKey == "" {
		return ""
	}
	
	if imageTag == "" || itemId == "" {
		return ""
	}
	
	hostUrl := j.Host
	if remote {
		if j.PlayHost != "" {
			hostUrl = j.PlayHost
		}
	}
	
	return fmt.Sprintf("%sItems/%s/Images/Backdrop?tag=%s&api_key=%s", hostUrl, itemId, imageTag, j.ApiKey)
}

func (j *Jellyfin) GetLatest(num int, username string) []entity.MediaServerPlayItem {
	if j.Host == "" || j.ApiKey == "" {
		return nil
	}
	
	user := j.User
	if username != "" {
		user = j.GetUser(username)
	}
	
	url := fmt.Sprintf("%sUsers/%s/Items/Latest", j.Host, user)
	params := map[string]string{
		"Limit":      "100",
		"MediaTypes": "Video",
		"Fields":     "ProductionYear,Path,BackdropImageTags",
		"api_key":    j.ApiKey,
	}
	
	resp, err := j.client.R().
		SetQueryParams(params).
		Get(url)
		
	if err != nil {
		common.LOG.Errorf("连接Users/Items/Latest出错�?v", err)
		return []entity.MediaServerPlayItem{}
	}
	
	if resp.StatusCode() != http.StatusOK {
		common.LOG.Error("Users/Items/Latest 未获取到返回数据")
		return []entity.MediaServerPlayItem{}
	}
	
	var result []map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		common.LOG.Errorf("解析Users/Items/Latest返回数据出错�?v", err)
		return []entity.MediaServerPlayItem{}
	}
	
	retLatest := []entity.MediaServerPlayItem{}
	// 用户媒体库文件夹列表（排除黑名单�?	libraryFolders := j.GetUserLibraryFolders()
	
	for _, item := range result {
		if len(retLatest) >= num {
			break
		}
		
		itemType, _ := item["Type"].(string)
		if itemType != "Movie" && itemType != "Series" {
			continue
		}
		
		itemPath, _ := item["Path"].(string)
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
			mediaType = "movie"
		} else {
			mediaType = "tv"
		}
		
		itemId, _ := item["Id"].(string)
		itemName, _ := item["Name"].(string)
		link := j.GetPlayUrl(itemId)
		image := j.getLocalImageById(itemId)
		
		var backdropTags []interface{}
		if tags, ok := item["BackdropImageTags"].([]interface{}); ok {
			backdropTags = tags
		}
		
		var subtitle string
		if prodYear, ok := item["ProductionYear"].(float64); ok {
			subtitle = strconv.Itoa(int(prodYear))
		}
		
		retLatest = append(retLatest, entity.MediaServerPlayItem{
			Id:                itemId,
			Title:             itemName,
			Subtitle:          subtitle,
			Type:              mediaType,
			Image:             image,
			Link:              link,
			BackdropImageTags: backdropTags,
			ServerType:        "jellyfin",
		})
	}
	
	return retLatest
}

func (j *Jellyfin) GetUserLibraryFolders() []string {
	// 这个方法需要根据实际实现来获取用户媒体库文件夹列表
	// 在Python版本中这个方法似乎没有实现，所以暂时返回空列表
	return []string{}
}

func (j *Jellyfin) PlayHostOrHost() string {
	if j.PlayHost != "" {
		return j.PlayHost
	}
	return j.Host
}
