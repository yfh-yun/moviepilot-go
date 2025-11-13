package models

// RadarrMovie Radarr电影信息
type RadarrMovie struct {
	// ID
	ID int `json:"id,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// 是否可用
	IsAvailable bool `json:"isAvailable"`
	// 是否监控
	Monitored bool `json:"monitored"`
	// TMDB ID
	TmdbID int `json:"tmdbId,omitempty"`
	// IMDB ID
	ImdbID string `json:"imdbId,omitempty"`
	// 标题别名
	TitleSlug string `json:"titleSlug,omitempty"`
	// 文件夹名�?	FolderName string `json:"folderName,omitempty"`
	// 路径
	Path string `json:"path,omitempty"`
	// 配置ID
	ProfileID int `json:"profileId,omitempty"`
	// 质量配置ID
	QualityProfileID int `json:"qualityProfileId,omitempty"`
	// 添加时间
	Added string `json:"added,omitempty"`
	// 是否有文�?	HasFile bool `json:"hasFile"`
}

// SonarrSeries Sonarr剧集信息
type SonarrSeries struct {
	// ID
	ID int `json:"id,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 排序标题
	SortTitle string `json:"sortTitle,omitempty"`
	// 季数
	SeasonCount int `json:"seasonCount,omitempty"`
	// 状�?	Status string `json:"status,omitempty"`
	// 概述
	Overview string `json:"overview,omitempty"`
	// 网络
	Network string `json:"network,omitempty"`
	// 播出时间
	AirTime string `json:"airTime,omitempty"`
	// 图片列表
	Images []interface{} `json:"images,omitempty"`
	// 远程海报
	RemotePoster string `json:"remotePoster,omitempty"`
	// 季信�?	Seasons []interface{} `json:"seasons,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// 路径
	Path string `json:"path,omitempty"`
	// 配置ID
	ProfileID int `json:"profileId,omitempty"`
	// 语言配置ID
	LanguageProfileID int `json:"languageProfileId,omitempty"`
	// 是否按季分文件夹
	SeasonFolder bool `json:"seasonFolder"`
	// 是否监控
	Monitored bool `json:"monitored"`
	// 是否使用场面编号
	UseSceneNumbering bool `json:"useSceneNumbering"`
	// 运行时间
	Runtime int `json:"runtime,omitempty"`
	// TMDB ID
	TmdbID int `json:"tmdbId,omitempty"`
	// IMDB ID
	ImdbID string `json:"imdbId,omitempty"`
	// TVDB ID
	TvdbID int `json:"tvdbId,omitempty"`
	// TVRage ID
	TvRageID int `json:"tvRageId,omitempty"`
	// TVMaze ID
	TvMazeID int `json:"tvMazeId,omitempty"`
	// 首播时间
	FirstAired string `json:"firstAired,omitempty"`
	// 剧集类型
	SeriesType string `json:"seriesType,omitempty"`
	// 清理后的标题
	CleanTitle string `json:"cleanTitle,omitempty"`
	// 标题别名
	TitleSlug string `json:"titleSlug,omitempty"`
	// 分级
	Certification string `json:"certification,omitempty"`
	// 类型列表
	Genres []interface{} `json:"genres,omitempty"`
	// 标签列表
	Tags []interface{} `json:"tags,omitempty"`
	// 添加时间
	Added string `json:"added,omitempty"`
	// 评分信息
	Ratings map[string]interface{} `json:"ratings,omitempty"`
	// 质量配置ID
	QualityProfileID int `json:"qualityProfileId,omitempty"`
	// 统计信息
	Statistics map[string]interface{} `json:"statistics,omitempty"`
	// 是否可用
	IsAvailable bool `json:"isAvailable,omitempty"`
	// 是否有文�?	HasFile bool `json:"hasFile,omitempty"`
}

// NewRadarrMovie 创建一个新�?RadarrMovie 实例
func NewRadarrMovie() *RadarrMovie {
	return &RadarrMovie{
		IsAvailable: false,
		Monitored:   false,
		HasFile:     false,
	}
}

// NewSonarrSeries 创建一个新�?SonarrSeries 实例
func NewSonarrSeries() *SonarrSeries {
	return &SonarrSeries{
		Images:            make([]interface{}, 0),
		Seasons:           make([]interface{}, 0),
		SeasonFolder:      false,
		Monitored:         false,
		UseSceneNumbering: false,
		Genres:            make([]interface{}, 0),
		Tags:              make([]interface{}, 0),
		Statistics:        make(map[string]interface{}),
		IsAvailable:       false,
		HasFile:           false,
	}
}
