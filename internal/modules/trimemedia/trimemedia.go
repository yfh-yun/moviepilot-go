package trimemedia

import (
	"fmt"
	"path/filepath"
	"strings"
	
	"moviepilot-go/pkg/models"
)

// TrimeMedia 飞牛影视客户�?type TrimeMedia struct {
	username       string
	password       string
	userinfo       *User
	host           string
	playhost       string
	libraries      map[string]*MediaDb
	syncLibraries  []string
	api            *Api
}

// NewTrimeMedia 创建新的TrimeMedia实例
func NewTrimeMedia(
	host string,
	username string,
	password string,
	playHost string,
	syncLibraries []string,
) *TrimeMedia {
	
	if host == "" || username == "" || password == "" {
		fmt.Println("飞牛影视配置不完整！�?)
		return nil
	}
	
	t := &TrimeMedia{
		username:      username,
		password:      password,
		host:          host,
		syncLibraries: syncLibraries,
		libraries:     make(map[string]*MediaDb),
	}
	
	if !t.Reconnect() {
		fmt.Printf("请检查服务端地址 %s\n", host)
		return nil
	}
	
	if result := t.createAPI(playHost); result != nil {
		t.playhost = result.api.Host()
		result.api.Close()
	} else if playHost != "" {
		fmt.Printf("请检查外网播放地址 %s\n", playHost)
		// 标准化播放地址
		t.playhost = strings.TrimRight(playHost, "/")
	}
	
	return t
}

// ApiCreateResult API创建结果
type ApiCreateResult struct {
	api     *Api
	version *Version
}

// createAPI 创建一个飞牛API
func (t *TrimeMedia) createAPI(host string) *ApiCreateResult {
	if host == "" {
		return nil
	}
	
	apiKey := "16CCEB3D-AB42-077D-36A1-F355324E4237"
	host = strings.TrimRight(host, "/")
	
	if !strings.HasSuffix(host, "/v") {
		// 尝试补上结尾�?v 测试能否正常访问
		api := NewApi(host+"/v", apiKey)
		if fnver := api.SysVersion(); fnver != nil {
			return &ApiCreateResult{
				api:     api,
				version: fnver,
			}
		}
	}
	
	// 测试用户配置的地址
	api := NewApi(host, apiKey)
	if fnver := api.SysVersion(); fnver != nil {
		return &ApiCreateResult{
			api:     api,
			version: fnver,
		}
	}
	
	return nil
}

// Close 关闭连接
func (t *TrimeMedia) Close() {
	t.Disconnect()
}

// IsConfigured 是否已配�?func (t *TrimeMedia) IsConfigured() bool {
	return t.host != "" && t.username != "" && t.password != ""
}

// IsAuthenticated 是否已登�?func (t *TrimeMedia) IsAuthenticated() bool {
	return t.IsConfigured() && 
		t.api != nil && 
		t.api.Token() != "" && 
		t.userinfo != nil
}

// IsInactive 判断是否需要重�?func (t *TrimeMedia) IsInactive() bool {
	if !t.IsAuthenticated() {
		return true
	}
	t.userinfo = t.api.UserInfo()
	return t.userinfo == nil
}

// Reconnect 重连
func (t *TrimeMedia) Reconnect() bool {
	if !t.IsConfigured() {
		return false
	}
	
	t.Disconnect()
	
	if result := t.createAPI(t.host); result != nil {
		t.api = result.api
		// 版本�?0.8.53, 服务版本:0.8.23
		if result.version != nil {
			fmt.Printf("版本�?%s, 服务版本:%s\n", result.version.Frontend, result.version.Backend)
		}
	} else {
		return false
	}
	
	if t.api.Login(t.username, t.password) == "" {
		return false
	}
	
	t.userinfo = t.api.UserInfo()
	if t.userinfo == nil {
		return false
	}
	
	fmt.Printf("%s 成功登录飞牛影视\n", t.username)
	
	// 刷新媒体库列�?	t.GetLibrarys(false)
	
	return true
}

// Disconnect 断开与飞牛的连接
func (t *TrimeMedia) Disconnect() {
	if t.api != nil {
		t.api.Logout()
		t.api.Close()
		t.api = nil
		t.userinfo = nil
		fmt.Printf("%s 已断开飞牛影视\n", t.username)
	}
}

// GetLibrarys 获取媒体服务器所有媒体库列表
func (t *TrimeMedia) GetLibrarys(hidden bool) []models.MediaServerLibrary {
	if !t.IsAuthenticated() {
		return []models.MediaServerLibrary{}
	}
	
	var mdbList []MediaDb
	if t.userinfo.IsAdmin == 1 {
		mdbList = t.api.MdbList()
	} else {
		tempList := t.api.MediaDbList()
		if tempList != nil {
			mdbList = make([]MediaDb, len(tempList))
			for i, item := range tempList {
				mdbList[i] = item
			}
		}
	}
	
	// 更新libraries缓存
	t.libraries = make(map[string]*MediaDb)
	for i := range mdbList {
		lib := &mdbList[i]
		t.libraries[lib.GUID] = lib
	}
	
	libraries := make([]models.MediaServerLibrary, 0)
	for _, library := range t.libraries {
		if hidden && t.isLibraryBlocked(library.GUID) {
			continue
		}
		
		var libraryType string
		switch library.Category {
		case CategoryMovie:
			libraryType = string(models.Movie)
		case CategoryTV:
			libraryType = string(models.TV)
		case CategoryOthers:
			// 忽略这个�?			continue
		default:
			libraryType = string(models.Unknown)
		}
		
		// 构建图片列表
		imageList := make([]string, len(library.Posters))
		for i, poster := range library.Posters {
			imageList[i] = fmt.Sprintf("%s%s?w=256", t.api.Host(), poster)
		}
		
		// 构建链接
		linkHost := t.playhost
		if linkHost == "" {
			linkHost = t.api.Host()
		}
		link := fmt.Sprintf("%s/library/%s", linkHost, library.GUID)
		
		libraries = append(libraries, models.MediaServerLibrary{
			Server:     "trimemedia",
			ID:         library.GUID,
			Name:       library.Name,
			Type:       libraryType,
			Path:       library.DirList,
			ImageList:  imageList,
			Link:       link,
			ServerType: "trimemedia",
		})
	}
	
	return libraries
}

// GetUserCount 获取用户数量(非管理员不能调用)
func (t *TrimeMedia) GetUserCount() int {
	if !t.IsAuthenticated() {
		return 0
	}
	
	if t.userinfo == nil || t.userinfo.IsAdmin != 1 {
		return 0
	}
	
	users := t.api.UserList()
	if users != nil {
		return len(users)
	}
	
	return 0
}

// GetMediasCount 获取媒体数量
func (t *TrimeMedia) GetMediasCount() *models.Statistic {
	if !t.IsAuthenticated() {
		return &models.Statistic{}
	}
	
	info := t.api.MediaDbSum()
	if info == nil {
		return &models.Statistic{}
	}
	
	return &models.Statistic{
		MovieCount: info.Movie,
		TvCount:    info.TV,
	}
}

// Authenticate 用户认证
func (t *TrimeMedia) Authenticate(username, password string) string {
	if username == "" || password == "" {
		return ""
	}
	
	if !t.IsConfigured() {
		return ""
	}
	
	if result := t.createAPI(t.host); result != nil {
		defer func() {
			result.api.Logout()
			result.api.Close()
		}()
		
		return result.api.Login(username, password)
	}
	
	return ""
}

// GetMovies 根据标题和年份，检查电影是否在飞牛中存在，存在则返回列�?func (t *TrimeMedia) GetMovies(title string, year string, tmdbID int) []models.MediaServerItem {
	if !t.IsAuthenticated() {
		return nil
	}
	
	movies := make([]models.MediaServerItem, 0)
	items := t.api.SearchList(title)
	if items == nil {
		return nil
	}
	
	for _, item := range items {
		if item.Type != TypeMovie {
			continue
		}
		
		// 检查TMDB ID
		matchTmdbID := (tmdbID == 0 || tmdbID == item.TMDBID())
		
		// 检查标�?		matchTitle := (title == item.Title || title == item.OriginalTitle)
		
		// 检查年�?		matchYear := true
		if year != "" && item.ReleaseDate != "" {
			matchYear = (item.ReleaseDate[:4] == year)
		}
		
		if matchTmdbID && matchTitle && matchYear {
			movies = append(movies, t.buildMediaServerItem(&item))
		}
	}
	
	return movies
}

// getSeriesIDByName 根据名称和年份获取剧集ID
func (t *TrimeMedia) getSeriesIDByName(name, year string) string {
	items := t.api.SearchList(name)
	if items == nil {
		return ""
	}
	
	for _, item := range items {
		if item.Type != TypeTV {
			continue
		}
		
		// 可惜搜索接口不下发original_title 也不能指定分类、年�?		if name == item.Title || name == item.OriginalTitle {
			if year == "" || (item.AirDate != "" && item.AirDate[:4] == year) {
				return item.GUID
			}
		}
	}
	
	return ""
}

// GetTVEpisodes 根据标题和年份和季，返回飞牛中的剧集列表
func (t *TrimeMedia) GetTVEpisodes(
	itemID string,
	title string,
	year string,
	tmdbID int,
	season int,
) (string, map[int][]int) {
	if !t.IsAuthenticated() {
		return "", nil
	}
	
	// 如果没有提供itemID，则通过名称和年份查�?	if itemID == "" {
		itemID = t.getSeriesIDByName(title, year)
		if itemID == "" {
			return "", nil
		}
	}
	
	itemInfo := t.GetItemInfo(itemID)
	if itemInfo == nil {
		return "", make(map[int][]int)
	}
	
	// 检查TMDB ID
	if tmdbID != 0 && itemInfo.Tmdbid != 0 {
		if tmdbID != itemInfo.Tmdbid {
			return "", make(map[int][]int)
		}
	}
	
	seasons := t.api.SeasonList(itemID)
	if seasons == nil {
		// 季列表获取失�?		return "", make(map[int][]int)
	}
	
	// 如果指定了季，则只保留该�?	if season != -1 {
		filteredSeasons := make([]Item, 0)
		for _, item := range seasons {
			if item.SeasonNumber == season {
				filteredSeasons = append(filteredSeasons, item)
				break
			}
		}
		
		if len(filteredSeasons) == 0 {
			// 没有匹配的季
			return "", make(map[int][]int)
		}
		
		seasons = filteredSeasons
	}
	
	seasonEpisodes := make(map[int][]int)
	for _, item := range seasons {
		episodes := t.api.EpisodeList(item.GUID)
		if episodes == nil {
			continue
		}
		
		for _, episode := range episodes {
			if _, exists := seasonEpisodes[episode.SeasonNumber]; !exists {
				seasonEpisodes[episode.SeasonNumber] = make([]int, 0)
			}
			seasonEpisodes[episode.SeasonNumber] = append(seasonEpisodes[episode.SeasonNumber], episode.EpisodeNumber)
		}
	}
	
	return itemID, seasonEpisodes
}

// RefreshRootLibrary 通知飞牛刷新整个媒体�?非管理员不能调用)
func (t *TrimeMedia) RefreshRootLibrary() bool {
	if !t.IsAuthenticated() {
		return false
	}
	
	if t.userinfo == nil || t.userinfo.IsAdmin != 1 {
		fmt.Println("飞牛仅支持管理员账号刷新媒体�?)
		return false
	}
	
	// 必须调用 否则容易误报 -14 Task duplicate
	t.api.TaskRunning()
	fmt.Println("刷新所有媒体库")
	return t.api.MdbScanAll()
}

// RefreshLibraryByItems 按路径刷新所在的媒体�?非管理员不能调用)
func (t *TrimeMedia) RefreshLibraryByItems(items []models.RefreshMediaItem) bool {
	if !t.IsAuthenticated() {
		return false
	}
	
	if t.userinfo == nil || t.userinfo.IsAdmin != 1 {
		fmt.Println("飞牛仅支持管理员账号刷新媒体�?)
		return false
	}
	
	libraries := make(map[string]bool) // 使用map实现set功能
	for _, item := range items {
		lib := t.matchLibraryByPath(item.TargetPath)
		if lib == nil {
			// 如果有匹配失败的,刷新整个�?			return t.RefreshRootLibrary()
		}
		// 媒体库去�?		libraries[lib.GUID] = true
	}
	
	// 必须调用 否则容易误报 -14 Task duplicate
	t.api.TaskRunning()
	
	for libGUID := range libraries {
		// 逐个刷新
		lib := t.libraries[libGUID]
		fmt.Printf("刷新媒体库：%s\n", lib.Name)
		if !t.api.MdbScan(lib) {
			// 如果失败，刷新整个库
			return t.RefreshRootLibrary()
		}
	}
	
	return true
}

// matchLibraryByPath 根据路径匹配媒体�?func (t *TrimeMedia) matchLibraryByPath(path string) *MediaDb {
	if path == "" {
		return nil
	}
	
	// 判断path是否是parent的子目录
	isSubpath := func(pathStr, parentStr string) bool {
		pathAbs, _ := filepath.Abs(pathStr)
		parentAbs, _ := filepath.Abs(parentStr)
		
		pathParts := strings.Split(filepath.ToSlash(pathAbs), "/")
		parentParts := strings.Split(filepath.ToSlash(parentAbs), "/")
		
		if len(pathParts) < len(parentParts) {
			return false
		}
		
		for i, part := range parentParts {
			if pathParts[i] != part {
				return false
			}
		}
		
		return true
	}
	
	for _, lib := range t.libraries {
		for _, d := range lib.DirList {
			if isSubpath(path, d) {
				return lib
			}
		}
	}
	
	return nil
}

// GetWebhookMessage 获取Webhook消息
func (t *TrimeMedia) GetWebhookMessage(body interface{}) *models.WebhookEventInfo {
	// 暂未实现
	return nil
}

// GetItemInfo 获取单个项目详情
func (t *TrimeMedia) GetItemInfo(itemID string) *models.MediaServerItem {
	if !t.IsAuthenticated() {
		return nil
	}
	
	if item := t.api.Item(itemID); item != nil {
		return t.buildMediaServerItem(item)
	}
	
	return nil
}

// buildMediaServerItem 构建媒体服务器项�?func (t *TrimeMedia) buildMediaServerItem(item *Item) models.MediaServerItem {
	var year string
	if item.AirDate != "" && item.Type == TypeTV {
		year = item.AirDate[:4]
	} else if item.ReleaseDate != "" {
		year = item.ReleaseDate[:4]
	}
	
	userState := models.MediaServerItemUserState{}
	if item.Watched == 1 {
		userState.Played = true
	}
	
	if item.Duration > 0 && item.TS > 0 {
		userState.Percentage = float64(item.TS) / float64(item.Duration) * 100
		userState.Resume = true
	}
	
	var itemType string
	if item.Type == "" {
		itemType = ""
	} else {
		// 将飞牛的媒体类型转为MP能识别的
		if item.Type == TypeTV {
			itemType = "Series"
		} else {
			itemType = string(item.Type)
		}
	}
	
	return models.MediaServerItem{
		Server:        "trimemedia",
		Library:       item.AncestorGUID,
		ItemID:        item.GUID,
		ItemType:      itemType,
		Title:         item.Title,
		OriginalTitle: item.OriginalTitle,
		Year:          year,
		Tmdbid:        item.TMDBID(),
		Imdbid:        item.IMDbID,
		UserState:     userState,
	}
}

// buildPlayURL 拼装播放链接
func (t *TrimeMedia) buildPlayURL(host string, item *Item) string {
	switch item.Type {
	case TypeEpisode:
		return fmt.Sprintf("%s/tv/episode/%s", host, item.GUID)
	case TypeSeason:
		return fmt.Sprintf("%s/tv/season/%s", host, item.GUID)
	case TypeMovie:
		return fmt.Sprintf("%s/movie/%s", host, item.GUID)
	case TypeTV:
		return fmt.Sprintf("%s/tv/%s", host, item.GUID)
	default:
		// 其它类型走通用页面，由飞牛来判�?		return fmt.Sprintf("%s/other/%s", host, item.GUID)
	}
}

// buildMediaServerPlayItem 构建媒体服务器播放项�?func (t *TrimeMedia) buildMediaServerPlayItem(item *Item) models.MediaServerPlayItem {
	var title, subtitle, mediaType string
	
	if item.Type == TypeEpisode {
		title = item.TVTitle
		subtitle = fmt.Sprintf("S%d:%d - %s", item.SeasonNumber, item.EpisodeNumber, item.Title)
	} else {
		title = item.Title
		if item.Type == TypeMovie {
			subtitle = "电影"
		} else {
			subtitle = "视频"
		}
	}
	
	if item.Type == TypeMovie || item.Type == TypeVideo {
		mediaType = string(models.Movie)
	} else {
		mediaType = string(models.TV)
	}
	
	host := t.playhost
	if host == "" {
		host = t.api.Host()
	}
	
	percent := 0.0
	if item.Duration > 0 && item.TS > 0 {
		percent = float64(item.TS) / float64(item.Duration) * 100.0
	}
	
	return models.MediaServerPlayItem{
		ID:        item.GUID,
		Title:     title,
		Subtitle:  subtitle,
		Type:      mediaType,
		Image:     fmt.Sprintf("%s%s", t.api.Host(), item.Poster),
		Link:      t.buildPlayURL(host, item),
		Percent:   percent,
		ServerType: "trimemedia",
	}
}

// GetItems 获取媒体服务器项目列�?func (t *TrimeMedia) GetItems(
	parent string,
	startIndex int,
	limit int,
) []models.MediaServerItem {
	if !t.IsAuthenticated() {
		return nil
	}
	
	pageSize := limit
	if pageSize == 0 {
		pageSize = -1
	}
	
	items := t.api.ItemList(
		parent,
		[]Type{TypeMovie, TypeTV, TypeDirectory},
		true,
		startIndex+1,
		pageSize,
		"create_time",
		"DESC",
	)
	
	if items == nil {
		return []models.MediaServerItem{}
	}
	
	result := make([]models.MediaServerItem, 0)
	for _, item := range items {
		if item.Type == TypeDirectory {
			// 递归获取子目录内�?			subItems := t.GetItems(item.GUID, 0, -1)
			result = append(result, subItems...)
		} else if item.Type == TypeMovie || item.Type == TypeTV {
			result = append(result, t.buildMediaServerItem(&item))
		}
	}
	
	return result
}

// GetPlayURL 获取媒体的外网播放链�?func (t *TrimeMedia) GetPlayURL(itemID string) string {
	if !t.IsAuthenticated() {
		return ""
	}
	
	item := t.api.Item(itemID)
	if item == nil {
		return ""
	}
	
	// 根据查询到的信息拼装出播放链�?	host := t.playhost
	if host == "" {
		host = t.api.Host()
	}
	
	return t.buildPlayURL(host, item)
}

// GetResume 获取继续观看列表
func (t *TrimeMedia) GetResume(num int) []models.MediaServerPlayItem {
	if !t.IsAuthenticated() {
		return nil
	}
	
	items := t.api.PlayList()
	if items == nil {
		return []models.MediaServerPlayItem{}
	}
	
	retResume := make([]models.MediaServerPlayItem, 0)
	for _, item := range items {
		if len(retResume) == num {
			break
		}
		
		if t.isLibraryBlocked(item.AncestorGUID) {
			continue
		}
		
		retResume = append(retResume, t.buildMediaServerPlayItem(&item))
	}
	
	return retResume
}

// GetLatest 获取最近更新列�?func (t *TrimeMedia) GetLatest(num int) []models.MediaServerPlayItem {
	if !t.IsAuthenticated() {
		return nil
	}
	
	pageSize := 100
	if num*5 > pageSize {
		pageSize = num * 5
	}
	
	items := t.api.ItemList(
		"",
		[]Type{TypeMovie, TypeTV},
		true,
		1,
		pageSize,
		"create_time",
		"DESC",
	)
	
	if items == nil {
		return []models.MediaServerPlayItem{}
	}
	
	latest := make([]models.MediaServerPlayItem, 0)
	for _, item := range items {
		if len(latest) == num {
			break
		}
		
		if t.isLibraryBlocked(item.AncestorGUID) {
			continue
		}
		
		latest = append(latest, t.buildMediaServerPlayItem(&item))
	}
	
	return latest
}

// GetLatestBackdrops 获取最近更新的媒体Backdrop图片
func (t *TrimeMedia) GetLatestBackdrops(num int, remote bool) []string {
	if !t.IsAuthenticated() {
		return nil
	}
	
	pageSize := 100
	if num*5 > pageSize {
		pageSize = num * 5
	}
	
	items := t.api.ItemList(
		"",
		[]Type{TypeMovie, TypeTV},
		true,
		1,
		pageSize,
		"create_time",
		"DESC",
	)
	
	if items == nil {
		return []string{}
	}
	
	backdrops := make([]string, 0)
	for _, item := range items {
		if len(backdrops) == num {
			break
		}
		
		if t.isLibraryBlocked(item.AncestorGUID) {
			continue
		}
		
		itemDetails := t.api.Item(item.GUID)
		if itemDetails == nil {
			continue
		}
		
		host := t.api.Host()
		if remote && t.playhost != "" {
			host = t.playhost
		}
		
		var itemImage string
		if itemDetails.Backdrops != "" {
			itemImage = itemDetails.Backdrops
		} else if itemDetails.Posters != "" {
			itemImage = itemDetails.Posters
		} else {
			itemImage = itemDetails.Poster
		}
		
		backdrops = append(backdrops, fmt.Sprintf("%s%s", host, itemImage))
	}
	
	return backdrops
}

// isLibraryBlocked 检查媒体库是否被阻�?func (t *TrimeMedia) isLibraryBlocked(libraryGUID string) bool {
	if library, exists := t.libraries[libraryGUID]; exists {
		if library.Category == CategoryOthers {
			// 忽略这个�?			return true
		}
	}
	
	if t.syncLibraries != nil && len(t.syncLibraries) > 0 {
		// 检查是否在同步库列表中
		allFound := false
		for _, lib := range t.syncLibraries {
			if lib == "all" {
				allFound = true
				break
			}
			if lib == libraryGUID {
				return false // 在列表中，不阻止
			}
		}
		
		// 如果没有找到"all"且没有找到具体的库，则阻�?		return !allFound
	}
	
	return false
}
