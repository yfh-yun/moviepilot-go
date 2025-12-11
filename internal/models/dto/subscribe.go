package dto

import "time"

// Subscribe 订阅信息
type Subscribe struct {
	ID int `json:"id,omitempty"`
	// 订阅名称
	Name string `json:"name,omitempty"`
	// 订阅年份
	Year string `json:"year,omitempty"`
	// 订阅类型 电影/电视剧
	Type string `json:"type,omitempty"`
	// 搜索关键字
	Keyword   string `json:"keyword,omitempty"`
	TmdbID    *int   `json:"tmdb_id,omitempty"`
	DoubanID  string `json:"douban_id,omitempty"`
	BangumiID *int   `json:"bangumi_id,omitempty"`
	MediaID   string `json:"media_id,omitempty"`
	// 季号
	Season *int `json:"season,omitempty"`
	// 海报
	Poster string `json:"poster,omitempty"`
	// 背景图
	Backdrop string `json:"backdrop,omitempty"`
	// 评分
	Vote int `json:"vote,omitempty"`
	// 描述
	Description string `json:"description,omitempty"`
	// 过滤规则
	Filter string `json:"filter,omitempty"`
	// 包含
	Include string `json:"include,omitempty"`
	// 排除
	Exclude string `json:"exclude,omitempty"`
	// 质量
	Quality string `json:"quality,omitempty"`
	// 分辨率
	Resolution string `json:"resolution,omitempty"`
	// 特效
	Effect string `json:"effect,omitempty"`
	// 总集数
	TotalEpisode int `json:"total_episode,omitempty"`
	// 开始集数
	StartEpisode int `json:"start_episode,omitempty"`
	// 缺失集数
	LackEpisode int `json:"lack_episode,omitempty"`
	// 附加信息
	Note any `json:"note,omitempty"`
	// 状态：N-新建， R-订阅中
	State string `json:"state,omitempty"`
	// 最后更新时间
	LastUpdate string `json:"last_update,omitempty"`
	// 订阅用户
	Username string `json:"username,omitempty"`
	// 订阅站点
	Sites []int `json:"sites,omitempty"`
	// 下载器
	Downloader string `json:"downloader,omitempty"`
	// 是否洗版
	BestVersion int `json:"best_version,omitempty"`
	// 当前优先级
	CurrentPriority *int `json:"current_priority,omitempty"`
	// 保存路径
	SavePath string `json:"save_path,omitempty"`
	// 是否使用 imdbid 搜索
	SearchImdbID int `json:"search_imdbid,omitempty"`
	// 时间
	Date string `json:"date,omitempty"`
	// 自定义识别词
	CustomWords string `json:"custom_words,omitempty"`
	// 自定义媒体类别
	MediaCategory string `json:"media_category,omitempty"`
	// 过滤规则组
	FilterGroups []string `json:"filter_groups,omitempty"`
	// 剧集组
	EpisodeGroup string `json:"episode_group,omitempty"`
}

// SubscribeShare 订阅分享
type SubscribeShare struct {
	// 分享ID
	ID int `json:"id,omitempty"`
	// 订阅ID
	SubscribeID int `json:"subscribe_id,omitempty"`
	// 分享标题
	ShareTitle string `json:"share_title,omitempty"`
	// 分享说明
	ShareComment string `json:"share_comment,omitempty"`
	// 分享人
	ShareUser string `json:"share_user,omitempty"`
	// 分享人唯一ID
	ShareUID string `json:"share_uid,omitempty"`
	// 订阅名称
	Name string `json:"name,omitempty"`
	// 订阅年份
	Year string `json:"year,omitempty"`
	// 订阅类型 电影/电视剧
	Type string `json:"type,omitempty"`
	// 搜索关键字
	Keyword   string `json:"keyword,omitempty"`
	TmdbID    *int   `json:"tmdb_id,omitempty"`
	DoubanID  string `json:"douban_id,omitempty"`
	BangumiID *int   `json:"bangumi_id,omitempty"`
	// 季号
	Season *int `json:"season,omitempty"`
	// 海报
	Poster string `json:"poster,omitempty"`
	// 背景图
	Backdrop string `json:"backdrop,omitempty"`
	// 评分
	Vote int `json:"vote,omitempty"`
	// 描述
	Description string `json:"description,omitempty"`
	// 包含
	Include string `json:"include,omitempty"`
	// 排除
	Exclude string `json:"exclude,omitempty"`
	// 质量
	Quality string `json:"quality,omitempty"`
	// 分辨率
	Resolution string `json:"resolution,omitempty"`
	// 特效
	Effect string `json:"effect,omitempty"`
	// 总集数
	TotalEpisode int `json:"total_episode,omitempty"`
	// 时间
	Date string `json:"date,omitempty"`
	// 自定义识别词
	CustomWords string `json:"custom_words,omitempty"`
	// 自定义媒体类别
	MediaCategory string `json:"media_category,omitempty"`
	// 自定义剧集组
	EpisodeGroup string `json:"episode_group,omitempty"`
	// 复用人次
	Count int `json:"count,omitempty"`
	// 分享时间
	ShareTime time.Time `json:"share_time,omitempty"`
	// 复用次数
	ForkCount int `json:"fork_count,omitempty"`
	// 点赞次数
	LikeCount int `json:"like_count,omitempty"`
}

// SubscribeShareStatistics 订阅分享统计
type SubscribeShareStatistics struct {
	// 分享人
	ShareUser string `json:"share_user,omitempty"`
	// 分享数量
	ShareCount int `json:"share_count,omitempty"`
	// 总复用人次
	TotalReuseCount int `json:"total_reuse_count,omitempty"`
}

// SubscribeDownloadFileInfo 订阅下载文件信息
type SubscribeDownloadFileInfo struct {
	// 种子名称
	TorrentTitle string `json:"torrent_title,omitempty"`
	// 站点名称
	SiteName string `json:"site_name,omitempty"`
	// 下载器
	Downloader string `json:"downloader,omitempty"`
	// hash
	Hash string `json:"hash,omitempty"`
	// 文件路径
	FilePath string `json:"file_path,omitempty"`
}

// SubscribeLibraryFileInfo 订阅媒体库文件信息
type SubscribeLibraryFileInfo struct {
	// 存储
	Storage string `json:"storage,omitempty"`
	// 文件路径
	FilePath string `json:"file_path,omitempty"`
}

// SubscribeEpisodeInfo 订阅剧集信息
type SubscribeEpisodeInfo struct {
	// 标题
	Title string `json:"title,omitempty"`
	// 描述
	Description string `json:"description,omitempty"`
	// 背景图
	Backdrop string `json:"backdrop,omitempty"`
	// 下载文件信息
	Download []*SubscribeDownloadFileInfo `json:"download,omitempty"`
	// 媒体库文件信息
	Library []*SubscribeLibraryFileInfo `json:"library,omitempty"`
}

// SubscribeInfo 订阅详细信息
type SubscribeInfo struct {
	// 订阅信息
	Subscribe *Subscribe `json:"subscribe,omitempty"`
	// 集信息 {集号: {download: 文件路径，library: 文件路径, backdrop: url, title: 标题, description: 描述}}
	Episodes map[int]*SubscribeEpisodeInfo `json:"episodes,omitempty"`
}

// ShareSubscribeRequest 分享订阅请求
type ShareSubscribeRequest struct {
	// 订阅ID
	SubscribeID int `json:"subscribe_id" binding:"required"`
	// 分享标题
	ShareTitle string `json:"share_title" binding:"required"`
	// 分享说明
	ShareComment string `json:"share_comment"`
	// 分享人
	ShareUser string `json:"share_user" binding:"required"`
	// 分享人唯一ID
	ShareUID string `json:"share_uid" binding:"required"`
}

// ShareStatistics 分享统计
type ShareStatistics struct {
	// 分享人
	ShareUser string `json:"share_user"`
	// 分享数量
	ShareCount int `json:"share_count"`
	// 总复用次数
	TotalForks int `json:"total_forks"`
	// 总点赞数
	TotalLikes int `json:"total_likes"`
}

// AddSubscribeRequest 添加订阅请求
type AddSubscribeRequest struct {
	Title         string   `json:"title" binding:"required"`
	Year          string   `json:"year"`
	MediaType     string   `json:"type" binding:"required,oneof=movie tv"`
	TMDBID        *int     `json:"tmdb_id"`
	DoubanID      string   `json:"douban_id"`
	BangumiID     *int     `json:"bangumi_id"`
	Season        *int     `json:"season"`
	TotalEpisode  *int     `json:"total_episode"`
	LackEpisode   *int     `json:"lack_episode"`
	Quality       string   `json:"quality"`
	Resolution    string   `json:"resolution"`
	Effect        string   `json:"effect"`
	Include       string   `json:"include"`
	Exclude       string   `json:"exclude"`
	BestVersion   bool     `json:"best_version"`
	SavePath      string   `json:"save_path"`
	Downloader    string   `json:"downloader"`
	Sites         []int    `json:"sites"`
	FilterGroups  []string `json:"filter_groups"`
	CustomWords   string   `json:"custom_words"`
	MediaCategory string   `json:"media_category"`
	Username      string   `json:"username"`
}

// SubscribeResult 订阅结果
type SubscribeResult struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

// SubscribeSearchRequest 订阅搜索请求
type SubscribeSearchRequest struct {
	SubscribeID uint   `json:"subscribe_id"`
	State       string `json:"state"`
	Manual      bool   `json:"manual"`
}

// RecognizeMediaRequest 识别媒体请求
type RecognizeMediaRequest struct {
	Title    string `json:"title"`
	Year     string `json:"year"`
	Type     string `json:"type"`
	TMDBID   *int   `json:"tmdb_id"`
	DoubanID string `json:"douban_id"`
	Season   int    `json:"season"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	MediaInfo   *MediaInfo     `json:"media_info"`
	Keyword     string         `json:"keyword"`
	NoExists    map[string]any `json:"no_exists"`
	Sites       []int          `json:"sites"`
	FilterRules []string       `json:"filter_rules"`
	CustomWords string         `json:"custom_words"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Site        string `json:"site"`
	SiteID      int    `json:"site_id"`
	Size        int64  `json:"size"`
	Seeders     int    `json:"seeders"`
	Leechers    int    `json:"leechers"`
	DownloadURL string `json:"download_url"`
	PageURL     string `json:"page_url"`
	Priority    int    `json:"priority"`
}

// BatchDownloadRequest 批量下载请求
type BatchDownloadRequest struct {
	Results    []*SearchResult `json:"results"`
	NoExists   map[string]any  `json:"no_exists"`
	Username   string          `json:"username"`
	SavePath   string          `json:"save_path"`
	Downloader string          `json:"downloader"`
}

// BatchDownloadResult 批量下载结果
type BatchDownloadResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	Downloads    []string `json:"downloads"`
}
