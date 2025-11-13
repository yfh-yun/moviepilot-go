package models

// ExistMediaInfo 媒体服务器存在媒体信�?type ExistMediaInfo struct {
	// 类型 电影、电视剧
	Type MediaType `json:"type,omitempty"`
	// �?	Seasons map[int][]interface{} `json:"seasons,omitempty"`
	// 媒体服务器类型：plex、jellyfin、emby、trimemedia
	ServerType string `json:"server_type,omitempty"`
	// 媒体服务器名�?	Server string `json:"server,omitempty"`
	// 媒体ID
	ItemID interface{} `json:"itemid,omitempty"`
}

// NotExistMediaInfo 媒体服务器不存在媒体信息
type NotExistMediaInfo struct {
	// �?	Season int `json:"season,omitempty"`
	// 剧集列表
	Episodes []interface{} `json:"episodes,omitempty"`
	// 总集�?	TotalEpisode int `json:"total_episode,omitempty"`
	// 开始集
	StartEpisode int `json:"start_episode,omitempty"`
}

// RefreshMediaItem 媒体库刷新信�?type RefreshMediaItem struct {
	// 标题
	Title string `json:"title,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// 类型
	Type MediaType `json:"type,omitempty"`
	// 类别
	Category string `json:"category,omitempty"`
	// 目录
	TargetPath string `json:"target_path,omitempty"`
}

// MediaServerLibrary 媒体服务器媒体库信息
type MediaServerLibrary struct {
	// 服务�?	Server string `json:"server,omitempty"`
	// ID
	ID interface{} `json:"id,omitempty"`
	// 名称
	Name string `json:"name,omitempty"`
	// 路径
	Path interface{} `json:"path,omitempty"`
	// 类型
	Type string `json:"type,omitempty"`
	// 封面�?	Image string `json:"image,omitempty"`
	// 封面图列�?	ImageList []string `json:"image_list,omitempty"`
	// 跳转链接
	Link string `json:"link,omitempty"`
	// 服务器类�?	ServerType string `json:"server_type,omitempty"`
}

// MediaServerItemUserState 媒体服务器媒体用户状�?type MediaServerItemUserState struct {
	// 已播�?	Played bool `json:"played,omitempty"`
	// 继续播放
	Resume bool `json:"resume,omitempty"`
	// 上次播放时间 10位时间戳
	LastPlayedDate string `json:"last_played_date,omitempty"`
	// 播放次数(不等于完播次数，理解为浏览次�?
	PlayCount int `json:"play_count,omitempty"`
	// 播放进度
	Percentage float64 `json:"percentage,omitempty"`
}

// MediaServerItem 媒体服务器媒体信�?type MediaServerItem struct {
	// ID
	ID interface{} `json:"id,omitempty"`
	// 服务�?	Server string `json:"server,omitempty"`
	// 媒体库ID
	Library interface{} `json:"library,omitempty"`
	// ID
	ItemID string `json:"item_id,omitempty"`
	// 类型
	ItemType string `json:"item_type,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 原标�?	OriginalTitle string `json:"original_title,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// TMDBID
	TmdbID int `json:"tmdbid,omitempty"`
	// IMDBID
	ImdbID string `json:"imdbid,omitempty"`
	// TVDBID
	TvdbID string `json:"tvdbid,omitempty"`
	// 路径
	Path string `json:"path,omitempty"`
	// 季集
	SeasonInfo map[int][]interface{} `json:"seasoninfo,omitempty"`
	// 备注
	Note interface{} `json:"note,omitempty"`
	// 同步时间
	LstModDate string `json:"lst_mod_date,omitempty"`
	// 用户状�?	UserState *MediaServerItemUserState `json:"user_state,omitempty"`
}

// MediaServerSeasonInfo 媒体服务器媒体剧集信�?type MediaServerSeasonInfo struct {
	// �?	Season int `json:"season,omitempty"`
	// 剧集列表
	Episodes []int `json:"episodes,omitempty"`
}

// WebhookEventInfo Webhook事件信息
type WebhookEventInfo struct {
	Event       string                 `json:"event,omitempty"`
	Channel     string                 `json:"channel,omitempty"`
	ServerName  string                 `json:"server_name,omitempty"`
	ItemType    string                 `json:"item_type,omitempty"`
	ItemName    string                 `json:"item_name,omitempty"`
	ItemID      string                 `json:"item_id,omitempty"`
	ItemPath    string                 `json:"item_path,omitempty"`
	SeasonID    string                 `json:"season_id,omitempty"`
	EpisodeID   string                 `json:"episode_id,omitempty"`
	TmdbID      string                 `json:"tmdb_id,omitempty"`
	Overview    string                 `json:"overview,omitempty"`
	Percentage  float64                `json:"percentage,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	DeviceName  string                 `json:"device_name,omitempty"`
	Client      string                 `json:"client,omitempty"`
	UserName    string                 `json:"user_name,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"`
	ItemFavorite bool                   `json:"item_favorite,omitempty"`
	SaveReason  string                 `json:"save_reason,omitempty"`
	ItemIsvirtual bool                  `json:"item_isvirtual,omitempty"`
	MediaType   string                 `json:"media_type,omitempty"`
	JSONObject  map[string]interface{} `json:"json_object,omitempty"`
}

// MediaServerPlayItem 媒体服务器可播放项目信息
type MediaServerPlayItem struct {
	ID                interface{} `json:"id,omitempty"`
	Title             string      `json:"title,omitempty"`
	Subtitle          string      `json:"subtitle,omitempty"`
	Type              string      `json:"type,omitempty"`
	Image             string      `json:"image,omitempty"`
	Link              string      `json:"link,omitempty"`
	Percent           float64     `json:"percent,omitempty"`
	BackdropImageTags []string    `json:"BackdropImageTags,omitempty"`
	ServerType        string      `json:"server_type,omitempty"`
}

// NewExistMediaInfo 创建一个新�?ExistMediaInfo 实例
func NewExistMediaInfo() *ExistMediaInfo {
	return &ExistMediaInfo{
		Seasons: make(map[int][]interface{}),
	}
}

// NewNotExistMediaInfo 创建一个新�?NotExistMediaInfo 实例
func NewNotExistMediaInfo() *NotExistMediaInfo {
	return &NotExistMediaInfo{
		Episodes: make([]interface{}, 0),
	}
}

// NewRefreshMediaItem 创建一个新�?RefreshMediaItem 实例
func NewRefreshMediaItem() *RefreshMediaItem {
	return &RefreshMediaItem{}
}

// NewMediaServerLibrary 创建一个新�?MediaServerLibrary 实例
func NewMediaServerLibrary() *MediaServerLibrary {
	return &MediaServerLibrary{
		ImageList: make([]string, 0),
	}
}

// NewMediaServerItemUserState 创建一个新�?MediaServerItemUserState 实例
func NewMediaServerItemUserState() *MediaServerItemUserState {
	return &MediaServerItemUserState{}
}

// NewMediaServerItem 创建一个新�?MediaServerItem 实例
func NewMediaServerItem() *MediaServerItem {
	return &MediaServerItem{
		SeasonInfo: make(map[int][]interface{}),
	}
}

// NewMediaServerSeasonInfo 创建一个新�?MediaServerSeasonInfo 实例
func NewMediaServerSeasonInfo() *MediaServerSeasonInfo {
	return &MediaServerSeasonInfo{
		Episodes: make([]int, 0),
	}
}

// NewWebhookEventInfo 创建一个新�?WebhookEventInfo 实例
func NewWebhookEventInfo() *WebhookEventInfo {
	return &WebhookEventInfo{
		JSONObject: make(map[string]interface{}),
	}
}

// NewMediaServerPlayItem 创建一个新�?MediaServerPlayItem 实例
func NewMediaServerPlayItem() *MediaServerPlayItem {
	return &MediaServerPlayItem{
		BackdropImageTags: make([]string, 0),
	}
}
