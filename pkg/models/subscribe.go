package models

// Subscribe 订阅信息模型
type Subscribe struct {
	ID              int         `json:"id,omitempty"`
	Name            string      `json:"name,omitempty"`             // 订阅名称
	Year            string      `json:"year,omitempty"`             // 订阅年份
	Type            string      `json:"type,omitempty"`             // 订阅类型 电影/电视�?	Keyword         string      `json:"keyword,omitempty"`          // 搜索关键�?	TMDBID          int         `json:"tmdbid,omitempty"`
	DoubanID        string      `json:"doubanid,omitempty"`
	BangumiID       int         `json:"bangumiid,omitempty"`
	MediaID         string      `json:"mediaid,omitempty"`
	Season          int         `json:"season,omitempty"`            // 季号
	Poster          string      `json:"poster,omitempty"`            // 海报
	Backdrop        string      `json:"backdrop,omitempty"`          // 背景�?	Vote            int         `json:"vote,omitempty"`              // 评分
	Description     string      `json:"description,omitempty"`       // 描述
	Filter          string      `json:"filter,omitempty"`            // 过滤规则
	Include         string      `json:"include,omitempty"`           // 包含
	Exclude         string      `json:"exclude,omitempty"`           // 排除
	Quality         string      `json:"quality,omitempty"`           // 质量
	Resolution      string      `json:"resolution,omitempty"`        // 分辨�?	Effect          string      `json:"effect,omitempty"`            // 特效
	TotalEpisode    int         `json:"total_episode,omitempty"`     // 总集�?	StartEpisode    int         `json:"start_episode,omitempty"`     // 开始集�?	LackEpisode     int         `json:"lack_episode,omitempty"`      // 缺失集数
	Note            interface{} `json:"note,omitempty"`              // 附加信息
	State           string      `json:"state,omitempty"`             // 状态：N-新建�?R-订阅�?	LastUpdate      string      `json:"last_update,omitempty"`       // 最后更新时�?	Username        string      `json:"username,omitempty"`          // 订阅用户
	Sites           []int       `json:"sites,omitempty"`             // 订阅站点
	Downloader      string      `json:"downloader,omitempty"`        // 下载�?	BestVersion     int         `json:"best_version,omitempty"`      // 是否洗版
	CurrentPriority int         `json:"current_priority,omitempty"`  // 当前优先�?	SavePath        string      `json:"save_path,omitempty"`         // 保存路径
	SearchIMDbID    int         `json:"search_imdbid,omitempty"`     // 是否使用 imdbid 搜索
	Date            string      `json:"date,omitempty"`              // 时间
	CustomWords     string      `json:"custom_words,omitempty"`      // 自定义识别词
	MediaCategory   string      `json:"media_category,omitempty"`    // 自定义媒体类�?	FilterGroups    []string    `json:"filter_groups,omitempty"`     // 过滤规则�?	EpisodeGroup    string      `json:"episode_group,omitempty"`     // 剧集�?}

// SubscribeShare 订阅分享模型
type SubscribeShare struct {
	ID           int    `json:"id,omitempty"`            // 分享ID
	SubscribeID  int    `json:"subscribe_id,omitempty"`  // 订阅ID
	ShareTitle   string `json:"share_title,omitempty"`   // 分享标题
	ShareComment string `json:"share_comment,omitempty"` // 分享说明
	ShareUser    string `json:"share_user,omitempty"`    // 分享�?	ShareUID     string `json:"share_uid,omitempty"`     // 分享人唯一ID
	Name         string `json:"name,omitempty"`          // 订阅名称
	Year         string `json:"year,omitempty"`          // 订阅年份
	Type         string `json:"type,omitempty"`          // 订阅类型 电影/电视�?	Keyword      string `json:"keyword,omitempty"`       // 搜索关键�?	TMDBID       int    `json:"tmdbid,omitempty"`
	DoubanID     string `json:"doubanid,omitempty"`
	BangumiID    int    `json:"bangumiid,omitempty"`
	Season       int    `json:"season,omitempty"`        // 季号
	Poster       string `json:"poster,omitempty"`        // 海报
	Backdrop     string `json:"backdrop,omitempty"`      // 背景�?	Vote         int    `json:"vote,omitempty"`          // 评分
	Description  string `json:"description,omitempty"`   // 描述
	Include      string `json:"include,omitempty"`       // 包含
	Exclude      string `json:"exclude,omitempty"`       // 排除
	Quality      string `json:"quality,omitempty"`       // 质量
	Resolution   string `json:"resolution,omitempty"`    // 分辨�?	Effect       string `json:"effect,omitempty"`        // 特效
	TotalEpisode int    `json:"total_episode,omitempty"` // 总集�?	Date         string `json:"date,omitempty"`          // 时间
	CustomWords  string `json:"custom_words,omitempty"`  // 自定义识别词
	MediaCategory string `json:"media_category,omitempty"` // 自定义媒体类�?	EpisodeGroup string `json:"episode_group,omitempty"`  // 自定义剧集组
	Count        int    `json:"count,omitempty"`          // 复用人次
}

// SubscribeShareStatistics 订阅分享统计模型
type SubscribeShareStatistics struct {
	ShareUser        string `json:"share_user,omitempty"`         // 分享�?	ShareCount       int    `json:"share_count,omitempty"`        // 分享数量
	TotalReuseCount  int    `json:"total_reuse_count,omitempty"`  // 总复用人�?}

// SubscribeDownloadFileInfo 订阅下载文件信息模型
type SubscribeDownloadFileInfo struct {
	TorrentTitle string `json:"torrent_title,omitempty"` // 种子名称
	SiteName     string `json:"site_name,omitempty"`     // 站点名称
	Downloader   string `json:"downloader,omitempty"`    // 下载�?	Hash         string `json:"hash,omitempty"`          // hash
	FilePath     string `json:"file_path,omitempty"`     // 文件路径
}

// SubscribeLibraryFileInfo 订阅媒体库文件信息模�?type SubscribeLibraryFileInfo struct {
	Storage  string `json:"storage,omitempty"`   // 存储
	FilePath string `json:"file_path,omitempty"` // 文件路径
}

// SubscribeEpisodeInfo 订阅剧集信息模型
type SubscribeEpisodeInfo struct {
	Title       string                      `json:"title,omitempty"`        // 标题
	Description string                      `json:"description,omitempty"`  // 描述
	Backdrop    string                      `json:"backdrop,omitempty"`     // 背景�?	Download    []SubscribeDownloadFileInfo `json:"download,omitempty"`     // 下载文件信息
	Library     []SubscribeLibraryFileInfo  `json:"library,omitempty"`      // 媒体库文件信�?}

// SubscribeInfo 订阅完整信息模型
type SubscribeInfo struct {
	Subscribe *Subscribe                    `json:"subscribe,omitempty"` // 订阅信息
	Episodes  map[int]*SubscribeEpisodeInfo `json:"episodes,omitempty"`  // 集信�?{集号: {download: 文件路径，library: 文件路径, backdrop: url, title: 标题, description: 描述}}
}

// NewSubscribe 创建一个新�?Subscribe 实例
func NewSubscribe() *Subscribe {
	return &Subscribe{
		Vote:         0,
		TotalEpisode: 0,
		StartEpisode: 0,
		LackEpisode:  0,
		Sites:        make([]int, 0),
		BestVersion:  0,
		SearchIMDbID: 0,
		FilterGroups: make([]string, 0),
	}
}

// NewSubscribeShare 创建一个新�?SubscribeShare 实例
func NewSubscribeShare() *SubscribeShare {
	return &SubscribeShare{
		Vote:         0,
		TotalEpisode: 0,
		Count:        0,
	}
}

// NewSubscribeShareStatistics 创建一个新�?SubscribeShareStatistics 实例
func NewSubscribeShareStatistics() *SubscribeShareStatistics {
	return &SubscribeShareStatistics{
		ShareCount:      0,
		TotalReuseCount: 0,
	}
}

// NewSubscribeDownloadFileInfo 创建一个新�?SubscribeDownloadFileInfo 实例
func NewSubscribeDownloadFileInfo() *SubscribeDownloadFileInfo {
	return &SubscribeDownloadFileInfo{}
}

// NewSubscribeLibraryFileInfo 创建一个新�?SubscribeLibraryFileInfo 实例
func NewSubscribeLibraryFileInfo() *SubscribeLibraryFileInfo {
	return &SubscribeLibraryFileInfo{
		Storage: "local",
	}
}

// NewSubscribeEpisodeInfo 创建一个新�?SubscribeEpisodeInfo 实例
func NewSubscribeEpisodeInfo() *SubscribeEpisodeInfo {
	return &SubscribeEpisodeInfo{
		Download: make([]SubscribeDownloadFileInfo, 0),
		Library:  make([]SubscribeLibraryFileInfo, 0),
	}
}

// NewSubscribeInfo 创建一个新�?SubscribeInfo 实例
func NewSubscribeInfo() *SubscribeInfo {
	return &SubscribeInfo{
		Episodes: make(map[int]*SubscribeEpisodeInfo),
	}
}
