package trimemedia

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/utils"
)

// User 用户信息
type User struct {
	GUID     string `json:"guid"`
	Username string `json:"username"`
	IsAdmin  int    `json:"is_admin"`
}

// Category 媒体库分�?type Category string

const (
	CategoryMovie  Category = "Movie"
	CategoryTV     Category = "TV"
	CategoryMix    Category = "Mix"
	CategoryOthers Category = "Others"
)

// Type 媒体类型
type Type string

const (
	TypeMovie     Type = "Movie"
	TypeTV        Type = "TV"
	TypeSeason    Type = "Season"
	TypeEpisode   Type = "Episode"
	TypeVideo     Type = "Video"
	TypeDirectory Type = "Directory"
)

// MediaDb 媒体库信�?type MediaDb struct {
	GUID     string   `json:"guid"`
	Category Category `json:"category"`
	Name     string   `json:"name,omitempty"`
	Posters  []string `json:"posters,omitempty"`
	DirList  []string `json:"dir_list,omitempty"`
}

// MediaDbSummary 媒体库统计信�?type MediaDbSummary struct {
	Favorite int `json:"favorite"`
	Movie    int `json:"movie"`
	TV       int `json:"tv"`
	Video    int `json:"video"`
	Total    int `json:"total"`
}

// Version 版本信息
type Version struct {
	Frontend string `json:"version,omitempty"`        // 飞牛影视版本
	Backend  string `json:"mediasrvVersion,omitempty"` // 服务版本
}

// Item 媒体�?type Item struct {
	GUID          string `json:"guid"`
	AncestorGUID  string `json:"ancestor_guid"`
	Type          Type   `json:"type,omitempty"`
	TVTitle       string `json:"tv_title,omitempty"`       // 当type为Episode时是剧名
	ParentTitle   string `json:"parent_title,omitempty"`   // 季名
	Title         string `json:"title,omitempty"`          // 分集名称或标�?	OriginalTitle string `json:"original_title,omitempty"` // 原始标题
	Overview      string `json:"overview,omitempty"`
	Poster        string `json:"poster,omitempty"`
	Backdrops     string `json:"backdrops,omitempty"`
	Posters       string `json:"posters,omitempty"`
	DoubanID      int    `json:"douban_id,omitempty"`
	IMDbID        string `json:"imdb_id,omitempty"`
	TrimID        string `json:"trim_id,omitempty"`
	ReleaseDate   string `json:"release_date,omitempty"`
	AirDate       string `json:"air_date,omitempty"`
	VoteAverage   string `json:"vote_average,omitempty"`
	SeasonNumber  int    `json:"season_number,omitempty"`
	EpisodeNumber int    `json:"episode_number,omitempty"`
	Duration      int    `json:"duration,omitempty"` // 片长(�?
	TS            int    `json:"ts,omitempty"`       // 已播�?�?
	Watched       int    `json:"watched,omitempty"`  // 1:已看�?}

// TMDBID 获取TMDB ID
func (i *Item) TMDBID() int {
	if i.TrimID == "" {
		return 0
	}
	
	if strings.HasPrefix(i.TrimID, "tt") || strings.HasPrefix(i.TrimID, "tm") {
		// 飞牛给tmdbid加了前缀用以区分tv或movie
		if id, err := strconv.Atoi(i.TrimID[2:]); err == nil {
			return id
		}
	}
	return 0
}

// Api 飞牛影视API客户�?type Api struct {
	host        string
	token       string
	apikey      string
	apiPath     string
	httpClient  *http.Client
	version     *Version
}

// NewApi 创建新的API客户�?func NewApi(host, apikey string) *Api {
	return &Api{
		apiPath:    "/api/v1",
		host:       strings.TrimRight(host, "/"),
		apikey:     apikey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Token 获取token
func (a *Api) Token() string {
	return a.token
}

// Host 获取主机地址
func (a *Api) Host() string {
	return a.host
}

// ApiKey 获取API密钥
func (a *Api) ApiKey() string {
	return a.apikey
}

// Version 获取版本信息
func (a *Api) Version() *Version {
	return a.version
}

// SysVersion 获取飞牛影视版本�?func (a *Api) SysVersion() *Version {
	res := a.request("/sys/version", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.(map[string]interface{})
			a.version = &Version{
				Frontend: utils.GetStringValue(data, "version"),
				Backend:  utils.GetStringValue(data, "mediasrvVersion"),
			}
			return a.version
		}
	}
	return nil
}

// Login 登录飞牛影视
func (a *Api) Login(username, password string) string {
	data := map[string]interface{}{
		"username":  username,
		"password":  password,
		"app_name":  "trimemedia-web",
	}
	
	res := a.request("/login", "POST", nil, data, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.(map[string]interface{})
			a.token = utils.GetStringValue(data, "token")
			return a.token
		}
	}
	return ""
}

// Logout 退出账�?func (a *Api) Logout() bool {
	if a.token == "" {
		return true
	}
	
	res := a.request("/user/logout", "POST", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			a.token = ""
			return true
		}
	}
	return false
}

// UserList 获取用户列表(仅管理员有权访问)
func (a *Api) UserList() []User {
	res := a.request("/manager/user/list", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.([]interface{})
			users := make([]User, 0, len(data))
			for _, item := range data {
				info := item.(map[string]interface{})
				user := User{
					GUID:     utils.GetStringValue(info, "guid"),
					Username: utils.GetStringValue(info, "username"),
					IsAdmin:  int(utils.GetFloatValue(info, "is_admin", 0)),
				}
				users = append(users, user)
			}
			return users
		}
		return []User{}
	}
	return nil
}

// UserInfo 获取当前用户信息
func (a *Api) UserInfo() *User {
	res := a.request("/user/info", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.(map[string]interface{})
			user := &User{
				GUID:     utils.GetStringValue(data, "guid"),
				Username: utils.GetStringValue(data, "username"),
				IsAdmin:  int(utils.GetFloatValue(data, "is_admin", 0)),
			}
			return user
		}
	}
	return nil
}

// MediaDbSum 获取媒体数量统计
func (a *Api) MediaDbSum() *MediaDbSummary {
	res := a.request("/mediadb/sum", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.(map[string]interface{})
			summary := &MediaDbSummary{
				Favorite: int(utils.GetFloatValue(data, "favorite", 0)),
				Movie:    int(utils.GetFloatValue(data, "movie", 0)),
				TV:       int(utils.GetFloatValue(data, "tv", 0)),
				Video:    int(utils.GetFloatValue(data, "video", 0)),
				Total:    int(utils.GetFloatValue(data, "total", 0)),
			}
			return summary
		}
	}
	return nil
}

// MediaDbList 获取媒体库列�?普通用�?
func (a *Api) MediaDbList() []MediaDb {
	res := a.request("/mediadb/list", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.([]interface{})
			items := make([]MediaDb, 0, len(data))
			for _, item := range data {
				info := item.(map[string]interface{})
				mdb := MediaDb{
					GUID:     utils.GetStringValue(info, "guid"),
					Category: Category(utils.GetStringValue(info, "category")),
					Name:     utils.GetStringValue(info, "title"),
				}
				
				// 处理海报
				if posters, ok := info["posters"].([]interface{}); ok {
					mdb.Posters = make([]string, len(posters))
					for i, poster := range posters {
						if posterStr, ok := poster.(string); ok {
							mdb.Posters[i] = a.buildImgApiURL(posterStr)
						}
					}
				}
				
				items = append(items, mdb)
			}
			return items
		}
		return []MediaDb{}
	}
	return nil
}

// buildImgApiURL 构建图片API URL
func (a *Api) buildImgApiURL(imgPath string) string {
	if imgPath == "" {
		return ""
	}
	if !strings.HasPrefix(imgPath, "/") {
		imgPath = "/" + imgPath
	}
	return a.apiPath + "/sys/img" + imgPath
}

// MdbList 获取媒体库列�?管理�?
func (a *Api) MdbList() []MediaDb {
	res := a.request("/mdb/list", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.([]interface{})
			items := make([]MediaDb, 0, len(data))
			for _, item := range data {
				info := item.(map[string]interface{})
				mdb := MediaDb{
					GUID:     utils.GetStringValue(info, "guid"),
					Category: Category(utils.GetStringValue(info, "category")),
					Name:     utils.GetStringValue(info, "name"),
				}
				
				// 处理海报
				if posters, ok := info["posters"].([]interface{}); ok {
					mdb.Posters = make([]string, len(posters))
					for i, poster := range posters {
						if posterStr, ok := poster.(string); ok {
							mdb.Posters[i] = a.buildImgApiURL(posterStr)
						}
					}
				}
				
				// 处理目录列表
				if dirList, ok := info["dir_list"].([]interface{}); ok {
					mdb.DirList = make([]string, len(dirList))
					for i, dir := range dirList {
						if dirStr, ok := dir.(string); ok {
							mdb.DirList[i] = dirStr
						}
					}
				}
				
				items = append(items, mdb)
			}
			return items
		}
		return []MediaDb{}
	}
	return nil
}

// MdbScanAll 扫描所有媒体库
func (a *Api) MdbScanAll() bool {
	res := a.request("/mdb/scanall", "POST", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			return true
		}
	}
	return false
}

// MdbScan 扫描指定媒体�?func (a *Api) MdbScan(mdb *MediaDb) bool {
	res := a.request("/mdb/scan/"+mdb.GUID, "POST", nil, map[string]interface{}{}, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			return true
		}
	}
	return false
}

// TaskRunning 获取当前正在运行的任�?func (a *Api) TaskRunning() bool {
	res := a.request("/task/running", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			// TODO 具体正在运行的任�?			return true
		}
	}
	return false
}

// buildItem 构造媒体Item
func (a *Api) buildItem(info map[string]interface{}) *Item {
	item := &Item{
		GUID:          utils.GetStringValue(info, "guid"),
		AncestorGUID:  utils.GetStringValue(info, "ancestor_guid"),
		Type:          Type(utils.GetStringValue(info, "type")),
		TVTitle:       utils.GetStringValue(info, "tv_title"),
		ParentTitle:   utils.GetStringValue(info, "parent_title"),
		Title:         utils.GetStringValue(info, "title"),
		OriginalTitle: utils.GetStringValue(info, "original_title"),
		Overview:      utils.GetStringValue(info, "overview"),
		Poster:        utils.GetStringValue(info, "poster"),
		Backdrops:     utils.GetStringValue(info, "backdrops"),
		Posters:       utils.GetStringValue(info, "posters"),
		IMDbID:        utils.GetStringValue(info, "imdb_id"),
		TrimID:        utils.GetStringValue(info, "trim_id"),
		ReleaseDate:   utils.GetStringValue(info, "release_date"),
		AirDate:       utils.GetStringValue(info, "air_date"),
		VoteAverage:   utils.GetStringValue(info, "vote_average"),
		Duration:      int(utils.GetFloatValue(info, "duration", 0)),
		TS:            int(utils.GetFloatValue(info, "ts", 0)),
		Watched:       int(utils.GetFloatValue(info, "watched", 0)),
	}
	
	// 处理季号和集�?	if seasonNumber, ok := info["season_number"].(float64); ok {
		item.SeasonNumber = int(seasonNumber)
	}
	
	if episodeNumber, ok := info["episode_number"].(float64); ok {
		item.EpisodeNumber = int(episodeNumber)
	}
	
	if doubanID, ok := info["douban_id"].(float64); ok {
		item.DoubanID = int(doubanID)
	}
	
	// Item详情接口才有posters和backdrops
	item.Posters = a.buildImgApiURL(item.Posters)
	item.Backdrops = a.buildImgApiURL(item.Backdrops)
	if item.Poster == "" {
		item.Poster = item.Posters
	}
	
	return item
}

// ItemList 获取媒体列表
func (a *Api) ItemList(
	guid string,
	types []Type,
	excludeGroupedVideo bool,
	page int,
	pageSize int,
	sortBy string,
	sort string,
) []Item {
	
	if types == nil {
		types = []Type{TypeMovie, TypeTV, TypeDirectory, TypeVideo}
	}
	
	post := map[string]interface{}{
		"sort_type":    sort,
		"sort_column":  sortBy,
		"page":         page,
		"page_size":    pageSize,
	}
	
	// 处理类型
	typeStrings := make([]string, len(types))
	for i, t := range types {
		typeStrings[i] = string(t)
	}
	post["tags"] = map[string]interface{}{
		"type": typeStrings,
	}
	
	if guid != "" {
		post["ancestor_guid"] = guid
	}
	
	if excludeGroupedVideo {
		post["exclude_grouped_video"] = 1
	}
	
	res := a.request("/item/list", "POST", nil, post, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.(map[string]interface{})
			if list, ok := data["list"].([]interface{}); ok {
				items := make([]Item, 0, len(list))
				for _, item := range list {
					if itemInfo, ok := item.(map[string]interface{}); ok {
						items = append(items, *a.buildItem(itemInfo))
					}
				}
				return items
			}
		}
		return []Item{}
	}
	return nil
}

// SearchList 搜索影片、演�?func (a *Api) SearchList(keywords string) []Item {
	params := map[string]string{
		"q": keywords,
	}
	
	res := a.request("/search/list", "GET", params, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.([]interface{})
			items := make([]Item, 0, len(data))
			for _, item := range data {
				if itemInfo, ok := item.(map[string]interface{}); ok {
					items = append(items, *a.buildItem(itemInfo))
				}
			}
			return items
		}
		return []Item{}
	}
	return nil
}

// Item 查询媒体详情
func (a *Api) Item(guid string) *Item {
	url := fmt.Sprintf("/item/%s", guid)
	res := a.request(url, "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.(map[string]interface{})
			return a.buildItem(data)
		}
	}
	return nil
}

// DelItem 删除媒体
func (a *Api) DelItem(guid string, deleteFile bool) bool {
	url := fmt.Sprintf("/item/%s", guid)
	deleteFlag := 0
	if deleteFile {
		deleteFlag = 1
	}
	
	data := map[string]interface{}{
		"delete_file":  deleteFlag,
		"media_guids":  []interface{}{},
	}
	
	res := a.request(url, "DELETE", nil, data, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			return true
		}
	}
	return false
}

// SeasonList 查询季列�?func (a *Api) SeasonList(tvGUID string) []Item {
	url := fmt.Sprintf("/season/list/%s", tvGUID)
	res := a.request(url, "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.([]interface{})
			items := make([]Item, 0, len(data))
			for _, item := range data {
				if itemInfo, ok := item.(map[string]interface{}); ok {
					items = append(items, *a.buildItem(itemInfo))
				}
			}
			return items
		}
		return []Item{}
	}
	return nil
}

// EpisodeList 查询剧集列表
func (a *Api) EpisodeList(seasonGUID string) []Item {
	url := fmt.Sprintf("/episode/list/%s", seasonGUID)
	res := a.request(url, "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.([]interface{})
			items := make([]Item, 0, len(data))
			for _, item := range data {
				if itemInfo, ok := item.(map[string]interface{}); ok {
					items = append(items, *a.buildItem(itemInfo))
				}
			}
			return items
		}
		return []Item{}
	}
	return nil
}

// PlayList 获取继续观看列表
func (a *Api) PlayList() []Item {
	res := a.request("/play/list", "GET", nil, nil, false)
	if res != nil && res.Success() {
		if res.Data != nil {
			data := res.Data.([]interface{})
			items := make([]Item, 0, len(data))
			for _, item := range data {
				if itemInfo, ok := item.(map[string]interface{}); ok {
					items = append(items, *a.buildItem(itemInfo))
				}
			}
			return items
		}
		return []Item{}
	}
	return nil
}

// getAuthx 计算消息签名
func (a *Api) getAuthx(apiPath, body string) string {
	if !strings.HasPrefix(apiPath, "/v") {
		apiPath = "/v" + apiPath
	}
	
	nonce := strconv.Itoa(rand.Intn(900000) + 100000)
	ts := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	
	md5Hash := md5.New()
	md5Hash.Write([]byte(body))
	dataHash := fmt.Sprintf("%x", md5Hash.Sum(nil))
	
	md5Hash = md5.New()
	md5Hash.Write([]byte(fmt.Sprintf(
		"%s_%s_%s_%s_%s_%s",
		"NDzZTVxnRKP8Z0jXg1VAMonaG8akvh",
		apiPath,
		nonce,
		ts,
		dataHash,
		a.apikey,
	)))
	sign := fmt.Sprintf("%x", md5Hash.Sum(nil))
	
	return fmt.Sprintf("nonce=%s&timestamp=%s&sign=%s", nonce, ts, sign)
}

// request 请求飞牛影视API
func (a *Api) request(
	api string,
	method string,
	params map[string]string,
	data map[string]interface{},
	suppressLog bool,
) *Result {
	
	if a.host == "" || api == "" {
		return nil
	}
	
	var apiPath string
	if !strings.HasPrefix(api, "/") {
		apiPath = a.apiPath + "/" + api
	} else {
		apiPath = a.apiPath + api
	}
	
	urlStr := a.host + apiPath
	
	// 处理参数
	var jsonBody string
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			if !suppressLog {
				fmt.Printf("序列化数据失�? %v\n", err)
			}
			return nil
		}
		jsonBody = string(jsonData)
	}
	
	var queriesUnquoted string
	if params != nil {
		queryParams := url.Values{}
		for k, v := range params {
			queryParams.Set(k, v)
		}
		queriesUnquoted = queryParams.Encode()
	}
	
	// 构建请求
	var req *http.Request
	var err error
	
	if method == "GET" {
		if queriesUnquoted != "" {
			urlStr += "?" + queriesUnquoted
		}
		req, err = http.NewRequest(method, urlStr, nil)
	} else {
		req, err = http.NewRequest(method, urlStr, strings.NewReader(jsonBody))
	}
	
	if err != nil {
		if !suppressLog {
			fmt.Printf("创建请求失败: %v\n", err)
		}
		return nil
	}
	
	// 设置请求�?	req.Header.Set("User-Agent", "MoviePilot-Go")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", a.host)
	req.Header.Set("Authorization", a.token)
	req.Header.Set("authx", a.getAuthx(apiPath, jsonBody+queriesUnquoted))
	
	if jsonBody != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// 发送请�?	resp, err := a.httpClient.Do(req)
	if err != nil {
		if !suppressLog {
			fmt.Printf("请求接口 %s 异常: %v\n", urlStr, err)
		}
		return nil
	}
	defer resp.Body.Close()
	
	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		if !suppressLog {
			fmt.Printf("解析响应失败: %v\n", err)
		}
		return nil
	}
	
	msg := utils.GetStringValue(response, "msg")
	code := int(utils.GetFloatValue(response, "code", -1))
	
	if code != 0 {
		if !suppressLog {
			fmt.Printf("请求接口 %s 失败，错误码: %d %s\n", urlStr, code, msg)
		}
		return &Result{
			Code: code,
			Msg:  msg,
			Data: nil,
		}
	}
	
	return &Result{
		Code: 0,
		Msg:  msg,
		Data: response["data"],
	}
}

// Close 关闭API会话
func (a *Api) Close() {
	// 在Go中不需要显式关闭HTTP客户�?}

// Result API响应结果
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 检查请求是否成�?func (r *Result) Success() bool {
	return r.Code == 0
}
