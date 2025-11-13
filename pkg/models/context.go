package models

// MetaInfo 识别元数�?type MetaInfo struct {
	// 是否处理的文�?	IsFile bool `json:"isfile,omitempty"`
	
	// 原字符串
	OrgString string `json:"org_string,omitempty"`
	
	// 原标�?	Title string `json:"title,omitempty"`
	
	// 副标�?	Subtitle string `json:"subtitle,omitempty"`
	
	// 类型 电影、电视剧
	Type string `json:"type,omitempty"`
	
	// 名称
	Name string `json:"name,omitempty"`
	
	// 识别的中文名
	CnName string `json:"cn_name,omitempty"`
	
	// 识别的英文名
	EnName string `json:"en_name,omitempty"`
	
	// 年份
	Year string `json:"year,omitempty"`
	
	// 总季�?	TotalSeason int `json:"total_season,omitempty"`
	
	// 识别的开始季 数字
	BeginSeason int `json:"begin_season,omitempty"`
	
	// 识别的结束季 数字
	EndSeason int `json:"end_season,omitempty"`
	
	// 总集�?	TotalEpisode int `json:"total_episode,omitempty"`
	
	// 识别的开始集
	BeginEpisode int `json:"begin_episode,omitempty"`
	
	// 识别的结束集
	EndEpisode int `json:"end_episode,omitempty"`
	
	// SxxExx
	SeasonEpisode string `json:"season_episode,omitempty"`
	
	// 集列�?	EpisodeList []int `json:"episode_list,omitempty"`
	
	// Partx Cd Dvd Disk Disc
	Part string `json:"part,omitempty"`
	
	// 识别的资源类�?	ResourceType string `json:"resource_type,omitempty"`
	
	// 识别的效�?	ResourceEffect string `json:"resource_effect,omitempty"`
	
	// 识别的分辨率
	ResourcePix string `json:"resource_pix,omitempty"`
	
	// 识别的制作组/字幕�?	ResourceTeam string `json:"resource_team,omitempty"`
	
	// 视频编码
	VideoEncode string `json:"video_encode,omitempty"`
	
	// 音频编码
	AudioEncode string `json:"audio_encode,omitempty"`
	
	// 资源类型
	Edition string `json:"edition,omitempty"`
	
	// 流媒体平�?	WebSource string `json:"web_source,omitempty"`
	
	// 应用的识别词信息
	ApplyWords []string `json:"apply_words,omitempty"`
}

// MediaInfo 识别媒体信息
type MediaInfo struct {
	// 来源：themoviedb、douban、bangumi
	Source string `json:"source,omitempty"`
	
	// 类型 电影、电视剧、合�?	Type string `json:"type,omitempty"`
	
	// 媒体标题
	Title string `json:"title,omitempty"`
	
	// 英文标题
	EnTitle string `json:"en_title,omitempty"`
	
	// 年份
	Year string `json:"year,omitempty"`
	
	// 标题（年份）
	TitleYear string `json:"title_year,omitempty"`
	
	// 当前指定季，如有
	Season int `json:"season,omitempty"`
	
	// TMDB ID
	TmdbID int `json:"tmdb_id,omitempty"`
	
	// IMDB ID
	ImdbID string `json:"imdb_id,omitempty"`
	
	// TVDB ID
	TvdbID string `json:"tvdb_id,omitempty"`
	
	// 豆瓣ID
	DoubanID string `json:"douban_id,omitempty"`
	
	// Bangumi ID
	BangumiID int `json:"bangumi_id,omitempty"`
	
	// 合集ID
	CollectionID int `json:"collection_id,omitempty"`
	
	// 其它媒体ID前缀
	MediaidPrefix string `json:"mediaid_prefix,omitempty"`
	
	// 其它媒体ID�?	MediaID string `json:"media_id,omitempty"`
	
	// 媒体原语�?	OriginalLanguage string `json:"original_language,omitempty"`
	
	// 媒体原发行标�?	OriginalTitle string `json:"original_title,omitempty"`
	
	// 媒体发行日期
	ReleaseDate string `json:"release_date,omitempty"`
	
	// 背景图片
	BackdropPath string `json:"backdrop_path,omitempty"`
	
	// 海报图片
	PosterPath string `json:"poster_path,omitempty"`
	
	// 评分
	VoteAverage float64 `json:"vote_average,omitempty"`
	
	// 描述
	Overview string `json:"overview,omitempty"`
	
	// 二级分类
	Category string `json:"category,omitempty"`
	
	// 季季集清�?	Seasons map[int][]int `json:"seasons,omitempty"`
	
	// 季详�?	SeasonInfo []map[string]interface{} `json:"season_info,omitempty"`
	
	// 别名和译�?	Names []string `json:"names,omitempty"`
	
	// 演员
	Actors []map[string]interface{} `json:"actors,omitempty"`
	
	// 导演
	Directors []map[string]interface{} `json:"directors,omitempty"`
	
	// 详情链接
	DetailLink string `json:"detail_link,omitempty"`
	
	// 其它TMDB属�?	// 是否成人内容
	Adult bool `json:"adult,omitempty"`
	
	// 创建�?	CreatedBy []map[string]interface{} `json:"created_by,omitempty"`
	
	// 集时�?	EpisodeRunTime []int `json:"episode_run_time,omitempty"`
	
	// 风格
	Genres []map[string]interface{} `json:"genres,omitempty"`
	
	// 首播日期
	FirstAirDate string `json:"first_air_date,omitempty"`
	
	// 首页
	Homepage string `json:"homepage,omitempty"`
	
	// 语种
	Languages []string `json:"languages,omitempty"`
	
	// 最后上映日�?	LastAirDate string `json:"last_air_date,omitempty"`
	
	// 流媒体平�?	Networks []map[string]interface{} `json:"networks,omitempty"`
	
	// 集数
	NumberOfEpisodes int `json:"number_of_episodes,omitempty"`
	
	// 季数
	NumberOfSeasons int `json:"number_of_seasons,omitempty"`
	
	// 原产�?	OriginCountry []string `json:"origin_country,omitempty"`
	
	// 原名
	OriginalName string `json:"original_name,omitempty"`
	
	// 出品公司
	ProductionCompanies []map[string]interface{} `json:"production_companies,omitempty"`
	
	// 出品�?	ProductionCountries []map[string]interface{} `json:"production_countries,omitempty"`
	
	// 语种
	SpokenLanguages []map[string]interface{} `json:"spoken_languages,omitempty"`
	
	// 状�?	Status string `json:"status,omitempty"`
	
	// 标签
	Tagline string `json:"tagline,omitempty"`
	
	// 风格ID
	GenreIds []int `json:"genre_ids,omitempty"`
	
	// 评价数量
	VoteCount int `json:"vote_count,omitempty"`
	
	// 流行�?	Popularity int `json:"popularity,omitempty"`
	
	// 时长
	Runtime int `json:"runtime,omitempty"`
	
	// 下一�?	NextEpisodeToAir map[string]interface{} `json:"next_episode_to_air,omitempty"`
	
	// 全部剧集�?	EpisodeGroups []map[string]interface{} `json:"episode_groups,omitempty"`
	
	// 剧集�?	EpisodeGroup string `json:"episode_group,omitempty"`
}

// TorrentInfo 搜索种子信息
type TorrentInfo struct {
	// 站点ID
	Site int `json:"site,omitempty"`
	
	// 站点名称
	SiteName string `json:"site_name,omitempty"`
	
	// 站点Cookie
	SiteCookie string `json:"site_cookie,omitempty"`
	
	// 站点UA
	SiteUa string `json:"site_ua,omitempty"`
	
	// 站点是否使用代理
	SiteProxy bool `json:"site_proxy,omitempty"`
	
	// 站点优先�?	SiteOrder int `json:"site_order,omitempty"`
	
	// 站点下载�?	SiteDownloader string `json:"site_downloader,omitempty"`
	
	// 种子名称
	Title string `json:"title,omitempty"`
	
	// 种子副标�?	Description string `json:"description,omitempty"`
	
	// IMDB ID
	Imdbid string `json:"imdbid,omitempty"`
	
	// 种子链接
	Enclosure string `json:"enclosure,omitempty"`
	
	// 详情页面
	PageUrl string `json:"page_url,omitempty"`
	
	// 种子大小
	Size float64 `json:"size,omitempty"`
	
	// 做种�?	Seeders int `json:"seeders,omitempty"`
	
	// 下载�?	Peers int `json:"peers,omitempty"`
	
	// 完成�?	Grabs int `json:"grabs,omitempty"`
	
	// 发布时间
	Pubdate string `json:"pubdate,omitempty"`
	
	// 已过时间
	DateElapsed string `json:"date_elapsed,omitempty"`
	
	// 免费截止时间
	Freedate string `json:"freedate,omitempty"`
	
	// 上传因子
	Uploadvolumefactor float64 `json:"uploadvolumefactor,omitempty"`
	
	// 下载因子
	Downloadvolumefactor float64 `json:"downloadvolumefactor,omitempty"`
	
	// HR
	HitAndRun bool `json:"hit_and_run,omitempty"`
	
	// 种子标签
	Labels []string `json:"labels,omitempty"`
	
	// 种子优先�?	PriOrder int `json:"pri_order,omitempty"`
	
	// 促销
	VolumeFactor string `json:"volume_factor,omitempty"`
	
	// 剩余免费时间
	FreedateDiff string `json:"freedate_diff,omitempty"`
}

// Context 上下�?type Context struct {
	// 元数�?	MetaInfo *MetaInfo `json:"meta_info,omitempty"`
	
	// 媒体信息
	MediaInfo *MediaInfo `json:"media_info,omitempty"`
	
	// 种子信息
	TorrentInfo *TorrentInfo `json:"torrent_info,omitempty"`
}

// MediaSeason 季信�?type MediaSeason struct {
	AirDate      string  `json:"air_date,omitempty"`
	EpisodeCount int     `json:"episode_count,omitempty"`
	Name         string  `json:"name,omitempty"`
	Overview     string  `json:"overview,omitempty"`
	PosterPath   string  `json:"poster_path,omitempty"`
	SeasonNumber int     `json:"season_number,omitempty"`
	VoteAverage  float64 `json:"vote_average,omitempty"`
}

// MediaPerson 媒体人物信息
type MediaPerson struct {
	// 来源：themoviedb、douban、bangumi
	Source string `json:"source,omitempty"`
	
	// 公共
	ID   int `json:"id,omitempty"`
	Type interface{} `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Character string `json:"character,omitempty"`
	Images map[string]interface{} `json:"images,omitempty"`
	
	// themoviedb
	ProfilePath string `json:"profile_path,omitempty"`
	Gender interface{} `json:"gender,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
	CreditID string `json:"credit_id,omitempty"`
	AlsoKnownAs []string `json:"also_known_as,omitempty"`
	Birthday string `json:"birthday,omitempty"`
	Deathday string `json:"deathday,omitempty"`
	ImdbID string `json:"imdb_id,omitempty"`
	KnownForDepartment string `json:"known_for_department,omitempty"`
	PlaceOfBirth string `json:"place_of_birth,omitempty"`
	Popularity float64 `json:"popularity,omitempty"`
	Biography string `json:"biography,omitempty"`
	
	// douban
	Roles []map[string]interface{} `json:"roles,omitempty"`
	Title string `json:"title,omitempty"`
	Url string `json:"url,omitempty"`
	Avatar interface{} `json:"avatar,omitempty"`
	LatinName string `json:"latin_name,omitempty"`
	
	// bangumi
	Career []string `json:"career,omitempty"`
	Relation string `json:"relation,omitempty"`
}
