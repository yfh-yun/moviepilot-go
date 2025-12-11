package mediaserver

import "moviepilot-go/internal/models/enums"

// ExistMediaInfo 媒体服务器存在媒体信息
type ExistMediaInfo struct {
	// 类型 电影、电视剧
	Type enums.MediaType `json:"type,omitempty"`
	// 季
	Seasons map[int][]int `json:"seasons,omitempty"`
	// 媒体服务器类型：plex、jellyfin、emby、trimemedia
	ServerType string `json:"server_type,omitempty"`
	// 媒体服务器名称
	Server string `json:"server,omitempty"`
	// 媒体ID
	ItemID any `json:"itemid,omitempty"` // string or int
}

// NotExistMediaInfo 媒体服务器不存在媒体信息
type NotExistMediaInfo struct {
	// 季
	Season *int `json:"season,omitempty"`
	// 剧集列表
	Episodes []int `json:"episodes,omitempty"`
	// 总集数
	TotalEpisode int `json:"total_episode,omitempty"`
	// 开始集
	StartEpisode int `json:"start_episode,omitempty"`
}

// RefreshMediaItem 媒体库刷新信息
type RefreshMediaItem struct {
	// 标题
	Title string `json:"title,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// 类型
	Type enums.MediaType `json:"type,omitempty"`
	// 类别
	Category string `json:"category,omitempty"`
	// 目录
	TargetPath string `json:"target_path,omitempty"`
}

// MediaServerLibrary 媒体服务器媒体库信息
type MediaServerLibrary struct {
	// 服务器
	Server string `json:"server,omitempty"`
	// ID
	ID any `json:"id,omitempty"` // string or int
	// 名称
	Name string `json:"name,omitempty"`
	// 路径
	Path any `json:"path,omitempty"` // string or list
	// 类型
	Type string `json:"type,omitempty"`
	// 封面图
	Image string `json:"image,omitempty"`
	// 封面图列表
	ImageList []string `json:"image_list,omitempty"`
	// 跳转链接
	Link string `json:"link,omitempty"`
	// 服务器类型
	ServerType string `json:"server_type,omitempty"`
}

// MediaServerItemUserState 媒体服务器项目用户状态
type MediaServerItemUserState struct {
	// 已播放
	Played *bool `json:"played,omitempty"`
	// 继续播放
	Resume *bool `json:"resume,omitempty"`
	// 上次播放时间 10位时间戳
	LastPlayedDate string `json:"last_played_date,omitempty"`
	// 播放次数(不等于完播次数，理解为浏览次数)
	PlayCount *int `json:"play_count,omitempty"`
	// 播放进度
	Percentage *float64 `json:"percentage,omitempty"`
}

// MediaServerItem 媒体服务器媒体信息
type MediaServerItem struct {
	// ID
	ID any `json:"id,omitempty"` // string or int
	// 服务器
	Server string `json:"server,omitempty"`
	// 媒体库ID
	Library any `json:"library,omitempty"` // string or int
	// ID
	ItemID string `json:"item_id,omitempty"`
	// 类型
	ItemType string `json:"item_type,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 原标题
	OriginalTitle string `json:"original_title,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// TMDBID
	TmdbID *int `json:"tmdbid,omitempty"`
	// IMDBID
	ImdbID string `json:"imdbid,omitempty"`
	// TVDBID
	TvdbID string `json:"tvdbid,omitempty"`
	// 路径
	Path string `json:"path,omitempty"`
	// 季集
	SeasonInfo map[int][]int `json:"seasoninfo,omitempty"`
	// 备注
	Note any `json:"note,omitempty"`
	// 同步时间
	LstModDate string `json:"lst_mod_date,omitempty"`
	// 用户状态
	UserState *MediaServerItemUserState `json:"user_state,omitempty"`
}

// MediaServerSeasonInfo 媒体服务器媒体剧集信息
type MediaServerSeasonInfo struct {
	Season   *int  `json:"season,omitempty"`
	Episodes []int `json:"episodes,omitempty"`
}

// WebhookEventInfo Webhook事件信息
type WebhookEventInfo struct {
	Event         string         `json:"event,omitempty"`
	Channel       string         `json:"channel,omitempty"`
	ServerName    string         `json:"server_name,omitempty"`
	ItemType      string         `json:"item_type,omitempty"`
	ItemName      string         `json:"item_name,omitempty"`
	ItemID        string         `json:"item_id,omitempty"`
	ItemPath      string         `json:"item_path,omitempty"`
	SeasonID      string         `json:"season_id,omitempty"`
	EpisodeID     string         `json:"episode_id,omitempty"`
	TmdbID        string         `json:"tmdb_id,omitempty"`
	Overview      string         `json:"overview,omitempty"`
	Percentage    *float64       `json:"percentage,omitempty"`
	IP            string         `json:"ip,omitempty"`
	DeviceName    string         `json:"device_name,omitempty"`
	Client        string         `json:"client,omitempty"`
	UserName      string         `json:"user_name,omitempty"`
	ImageURL      string         `json:"image_url,omitempty"`
	ItemFavorite  *bool          `json:"item_favorite,omitempty"`
	SaveReason    string         `json:"save_reason,omitempty"`
	ItemIsVirtual *bool          `json:"item_isvirtual,omitempty"`
	MediaType     string         `json:"media_type,omitempty"`
	JSONObject    map[string]any `json:"json_object,omitempty"`
}

// MediaServerPlayItem 媒体服务器可播放项目信息
type MediaServerPlayItem struct {
	ID                any      `json:"id,omitempty"` // string or int
	Title             string   `json:"title,omitempty"`
	Subtitle          string   `json:"subtitle,omitempty"`
	Type              string   `json:"type,omitempty"`
	Image             string   `json:"image,omitempty"`
	Link              string   `json:"link,omitempty"`
	Percent           *float64 `json:"percent,omitempty"`
	BackdropImageTags []string `json:"BackdropImageTags,omitempty"`
	ServerType        string   `json:"server_type,omitempty"`
}
