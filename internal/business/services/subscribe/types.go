package subscribe

import (
	"time"
)

// Subscribe 订阅模型
type Subscribe struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Year               string    `json:"year"`
	Type               string    `json:"type"` // movie/tv
	TMDBID             *int      `json:"tmdb_id"`
	DoubanID           string    `json:"douban_id"`
	BangumiID          *int      `json:"bangumi_id"`
	Season             int       `json:"season"`
	Poster             string    `json:"poster"`
	Backdrop           string    `json:"backdrop"`
	Vote               float64   `json:"vote"`
	Description        string    `json:"description"`
	IMDBID             string    `json:"imdb_id"`
	TVDBID             string    `json:"tvdb_id"`
	Keyword            string    `json:"keyword"`
	Quality            string    `json:"quality"`
	Resolution         string    `json:"resolution"`
	Effect             string    `json:"effect"`
	Include            string    `json:"include"`
	Exclude            string    `json:"exclude"`
	TotalEpisode       int       `json:"total_episode"`
	StartEpisode       int       `json:"start_episode"`
	LackEpisode        int       `json:"lack_episode"`
	State              string    `json:"state"` // N:新建 R:订阅中 P:待定 S:暂停
	BestVersion        bool      `json:"best_version"`
	CurrentPriority    *int      `json:"current_priority"`
	SearchIMDBID       bool      `json:"search_imdbid"`
	Sites              []int     `json:"sites"`
	Downloader         string    `json:"downloader"`
	SavePath           string    `json:"save_path"`
	FilterGroups       []string  `json:"filter_groups"`
	CustomWords        string    `json:"custom_words"`
	MediaCategory      string    `json:"media_category"`
	EpisodeGroup       string    `json:"episode_group"`
	Note               []int     `json:"note"` // 已下载集数
	ManualTotalEpisode bool      `json:"manual_total_episode"`
	Date               time.Time `json:"date"`
	LastUpdate         time.Time `json:"last_update"`
	Username           string    `json:"username"`
}

// MetaInfo 元数据信息
type MetaInfo struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	Year         string `json:"year"`
	Type         string `json:"type"`
	BeginSeason  int    `json:"begin_season"`
	Season       int    `json:"season"`
	BeginEpisode int    `json:"begin_episode"`
	EpisodeList  []int  `json:"episode_list"`
	SeasonList   []int  `json:"season_list"`
	EpisodeGroup string `json:"episode_group"`
	OrgString    string `json:"org_string"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	TMDBID        int           `json:"tmdb_id"`
	DoubanID      string        `json:"douban_id"`
	BangumiID     int           `json:"bangumi_id"`
	IMDBID        string        `json:"imdb_id"`
	TVDBID        string        `json:"tvdb_id"`
	Title         string        `json:"title"`
	TitleYear     string        `json:"title_year"`
	OriginalTitle string        `json:"original_title"`
	Year          string        `json:"year"`
	Type          string        `json:"type"`
	Overview      string        `json:"overview"`
	PosterPath    string        `json:"poster_path"`
	BackdropPath  string        `json:"backdrop_path"`
	VoteAverage   float64       `json:"vote_average"`
	Seasons       map[int][]int `json:"seasons"` // season -> episodes
	Category      string        `json:"category"`
	EpisodeGroup  string        `json:"episode_group"`
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Site        int       `json:"site"`
	SiteName    string    `json:"site_name"`
	Size        int64     `json:"size"`
	Seeders     int       `json:"seeders"`
	Leechers    int       `json:"leechers"`
	DownloadURL string    `json:"download_url"`
	PageURL     string    `json:"page_url"`
	PriOrder    int       `json:"pri_order"` // 优先级
	PubDate     time.Time `json:"pub_date"`
	Hash        string    `json:"hash"`
}

// Context 种子上下文
type Context struct {
	MetaInfo                *MetaInfo    `json:"meta_info"`
	TorrentInfo             *TorrentInfo `json:"torrent_info"`
	MediaInfo               *MediaInfo   `json:"media_info"`
	MediaRecognizeFailCount int          `json:"media_recognize_fail_count"`
}

// NotExistMediaInfo 缺失媒体信息
type NotExistMediaInfo struct {
	Season       int   `json:"season"`
	Episodes     []int `json:"episodes"`
	TotalEpisode int   `json:"total_episode"`
	StartEpisode int   `json:"start_episode"`
}

// AddSubscribeRequest 添加订阅请求
type AddSubscribeRequest struct {
	Title         string   `json:"title" binding:"required"`
	Year          string   `json:"year"`
	MediaType     string   `json:"media_type"`
	TMDBID        *int     `json:"tmdb_id"`
	DoubanID      string   `json:"douban_id"`
	BangumiID     *int     `json:"bangumi_id"`
	MediaID       string   `json:"media_id"`
	EpisodeGroup  string   `json:"episode_group"`
	Season        *int     `json:"season"`
	Channel       string   `json:"channel"`
	Source        string   `json:"source"`
	UserID        string   `json:"user_id"`
	Username      string   `json:"username"`
	Message       bool     `json:"message"`
	ExistOK       bool     `json:"exist_ok"`
	Quality       string   `json:"quality"`
	Resolution    string   `json:"resolution"`
	Effect        string   `json:"effect"`
	Include       string   `json:"include"`
	Exclude       string   `json:"exclude"`
	BestVersion   bool     `json:"best_version"`
	SearchIMDBID  bool     `json:"search_imdbid"`
	Sites         []int    `json:"sites"`
	Downloader    string   `json:"downloader"`
	SavePath      string   `json:"save_path"`
	FilterGroups  []string `json:"filter_groups"`
	CustomWords   string   `json:"custom_words"`
	MediaCategory string   `json:"media_category"`
	TotalEpisode  int      `json:"total_episode"`
	LackEpisode   int      `json:"lack_episode"`
}

// AddSubscribeResponse 添加订阅响应
type AddSubscribeResponse struct {
	SubscribeID int    `json:"subscribe_id"`
	Message     string `json:"message"`
}

// UpdateSubscribeRequest 更新订阅请求
type UpdateSubscribeRequest struct {
	Name          *string  `json:"name"`
	Quality       *string  `json:"quality"`
	Resolution    *string  `json:"resolution"`
	Effect        *string  `json:"effect"`
	Include       *string  `json:"include"`
	Exclude       *string  `json:"exclude"`
	TotalEpisode  *int     `json:"total_episode"`
	StartEpisode  *int     `json:"start_episode"`
	BestVersion   *bool    `json:"best_version"`
	SearchIMDBID  *bool    `json:"search_imdbid"`
	Sites         []int    `json:"sites"`
	Downloader    *string  `json:"downloader"`
	SavePath      *string  `json:"save_path"`
	FilterGroups  []string `json:"filter_groups"`
	CustomWords   *string  `json:"custom_words"`
	MediaCategory *string  `json:"media_category"`
	State         *string  `json:"state"`
}

// ListOptions 列表查询选项
type ListOptions struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	State    string `json:"state"`
	Type     string `json:"type"`
}

// SearchSubscribeRequest 搜索订阅请求
type SearchSubscribeRequest struct {
	SubscribeID *int   `json:"subscribe_id"`
	State       string `json:"state"` // N:新建 R:订阅中 P:待定 S:暂停
	Manual      bool   `json:"manual"`
}

// MatchSubscribeRequest 匹配订阅请求
type MatchSubscribeRequest struct {
	Torrents map[string][]*Context `json:"torrents"`
}

// CheckSubscribeRequest 检查订阅请求
type CheckSubscribeRequest struct {
	// 无参数
}

// RefreshSubscribeRequest 刷新订阅请求
type RefreshSubscribeRequest struct {
	// 无参数
}

// Notification 通知
type Notification struct {
	Channel  string `json:"channel"`
	Source   string `json:"source"`
	MType    string `json:"mtype"`
	CType    string `json:"ctype"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Image    string `json:"image"`
	Link     string `json:"link"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// RecognizeMediaRequest 识别媒体请求
type RecognizeMediaRequest struct {
	Meta         *MetaInfo `json:"meta"`
	MediaType    string    `json:"media_type"`
	TMDBID       *int      `json:"tmdb_id"`
	DoubanID     string    `json:"douban_id"`
	BangumiID    *int      `json:"bangumi_id"`
	EpisodeGroup string    `json:"episode_group"`
	Cache        bool      `json:"cache"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	MediaInfo    *MediaInfo                         `json:"media_info"`
	Keyword      string                             `json:"keyword"`
	NoExists     map[any]map[int]*NotExistMediaInfo `json:"no_exists"`
	Sites        []int                              `json:"sites"`
	RuleGroups   []string                           `json:"rule_groups"`
	Area         string                             `json:"area"` // title/imdbid
	CustomWords  []string                           `json:"custom_words"`
	FilterParams map[string]any                     `json:"filter_params"`
}

// BatchDownloadRequest 批量下载请求
type BatchDownloadRequest struct {
	Contexts   []*Context                         `json:"contexts"`
	NoExists   map[any]map[int]*NotExistMediaInfo `json:"no_exists"`
	Username   string                             `json:"username"`
	SavePath   string                             `json:"save_path"`
	Downloader string                             `json:"downloader"`
	Source     string                             `json:"source"`
}

// BatchDownloadResult 批量下载结果
type BatchDownloadResult struct {
	Downloads []*Context                         `json:"downloads"`
	Lefts     map[any]map[int]*NotExistMediaInfo `json:"lefts"`
}

// FilterTorrentsRequest 过滤种子请求
type FilterTorrentsRequest struct {
	RuleGroups  []string       `json:"rule_groups"`
	TorrentList []*TorrentInfo `json:"torrent_list"`
	MediaInfo   *MediaInfo     `json:"media_info"`
}

// GetNoExistsInfoRequest 获取缺失信息请求
type GetNoExistsInfoRequest struct {
	Meta      *MetaInfo   `json:"meta"`
	MediaInfo *MediaInfo  `json:"media_info"`
	Totals    map[int]int `json:"totals"` // season -> total_episode
}

// GetNoExistsInfoResult 获取缺失信息结果
type GetNoExistsInfoResult struct {
	ExistFlag bool                               `json:"exist_flag"`
	NoExists  map[any]map[int]*NotExistMediaInfo `json:"no_exists"`
}

// MediaRecognizeConvertEventData 媒体识别转换事件数据
type MediaRecognizeConvertEventData struct {
	MediaID     string         `json:"mediaid"`
	ConvertType string         `json:"convert_type"` // themoviedb/douban
	MediaDict   map[string]any `json:"media_dict"`
}

// SubscribeSourceKeyword 订阅来源关键字
type SubscribeSourceKeyword struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Year      string `json:"year"`
	Type      string `json:"type"`
	Season    int    `json:"season"`
	TMDBID    *int   `json:"tmdbid"`
	IMDBID    string `json:"imdbid"`
	TVDBID    string `json:"tvdbid"`
	DoubanID  string `json:"doubanid"`
	BangumiID *int   `json:"bangumiid"`
}

// OverallStats 总体统计
type OverallStats struct {
	TotalSubscribes  int     `json:"total_subscribes"`   // 总订阅数
	ActiveSubscribes int     `json:"active_subscribes"`  // 活跃订阅数
	TotalDownloads   int     `json:"total_downloads"`    // 总下载数
	TodayDownloads   int     `json:"today_downloads"`    // 今日下载数
	WeekDownloads    int     `json:"week_downloads"`     // 本周下载数
	MonthDownloads   int     `json:"month_downloads"`    // 本月下载数
	SuccessRate      float64 `json:"success_rate"`       // 成功率
	AverageMatchTime float64 `json:"average_match_time"` // 平均匹配时间(秒)
	LastUpdateTime   string  `json:"last_update_time"`   // 最后更新时间
}

// SubscribeStats 订阅统计
type SubscribeStats struct {
	SubscribeID      uint      `json:"subscribe_id"`      // 订阅ID
	SubscribeName    string    `json:"subscribe_name"`    // 订阅名称
	TotalDownloads   int       `json:"total_downloads"`   // 总下载数
	SuccessDownloads int       `json:"success_downloads"` // 成功下载数
	FailedDownloads  int       `json:"failed_downloads"`  // 失败下载数
	LastMatchTime    time.Time `json:"last_match_time"`   // 最后匹配时间
	AverageSize      int64     `json:"average_size"`      // 平均大小(字节)
	TotalSize        int64     `json:"total_size"`        // 总大小(字节)
	SuccessRate      float64   `json:"success_rate"`      // 成功率
	State            string    `json:"state"`             // 订阅状态
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Date      string `json:"date"`      // 日期 (YYYY-MM-DD)
	Downloads int    `json:"downloads"` // 下载数
	Matches   int    `json:"matches"`   // 匹配数
	Success   int    `json:"success"`   // 成功数
}

// TopSubscribe 热门订阅
type TopSubscribe struct {
	SubscribeID   uint    `json:"subscribe_id"`   // 订阅ID
	SubscribeName string  `json:"subscribe_name"` // 订阅名称
	Poster        string  `json:"poster"`         // 海报
	Type          string  `json:"type"`           // 类型 (movie/tv)
	Downloads     int     `json:"downloads"`      // 下载数
	SuccessRate   float64 `json:"success_rate"`   // 成功率
	LastUpdate    string  `json:"last_update"`    // 最后更新时间
}

// MatchRecord 匹配记录
type MatchRecord struct {
	ID           uint      `json:"id"`            // 记录ID
	SubscribeID  uint      `json:"subscribe_id"`  // 订阅ID
	TorrentTitle string    `json:"torrent_title"` // 种子标题
	TorrentSite  string    `json:"torrent_site"`  // 种子站点
	TorrentSize  int64     `json:"torrent_size"`  // 种子大小
	MatchTime    time.Time `json:"match_time"`    // 匹配时间
	DownloadTime time.Time `json:"download_time"` // 下载时间
	Status       string    `json:"status"`        // 状态 (matched/downloaded/failed)
	Season       int       `json:"season"`        // 季
	Episodes     []int     `json:"episodes"`      // 集数列表
	Quality      string    `json:"quality"`       // 质量
	Resolution   string    `json:"resolution"`    // 分辨率
}

// MatchStats 匹配统计
type MatchStats struct {
	SubscribeID         uint    `json:"subscribe_id"`         // 订阅ID
	TotalMatches        int     `json:"total_matches"`        // 总匹配数
	SuccessMatches      int     `json:"success_matches"`      // 成功匹配数
	FailedMatches       int     `json:"failed_matches"`       // 失败匹配数
	AverageMatchTime    float64 `json:"average_match_time"`   // 平均匹配时间(秒)
	LastMatchTime       string  `json:"last_match_time"`      // 最后匹配时间
	MostMatchedSite     string  `json:"most_matched_site"`    // 最常匹配站点
	PreferredQuality    string  `json:"preferred_quality"`    // 偏好质量
	PreferredResolution string  `json:"preferred_resolution"` // 偏好分辨率
}
