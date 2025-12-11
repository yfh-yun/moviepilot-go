package history

// DownloadHistory 下载历史
type DownloadHistory struct {
	// ID
	ID int `json:"id"`
	// 保存路径
	Path string `json:"path,omitempty"`
	// 类型：电影、电视剧
	Type string `json:"type,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// TMDBID
	TmdbID *int `json:"tmdb_id,omitempty"`
	// IMDBID
	ImdbID string `json:"imdb_id,omitempty"`
	// TVDBID
	TvdbID *int `json:"tvdb_id,omitempty"`
	// 豆瓣ID
	DoubanID string `json:"douban_id,omitempty"`
	// 季Sxx
	Seasons string `json:"seasons,omitempty"`
	// 集Exx
	Episodes string `json:"episodes,omitempty"`
	// 海报
	Image string `json:"image,omitempty"`
	// 下载器Hash
	DownloadHash string `json:"download_hash,omitempty"`
	// 种子名称
	TorrentName string `json:"torrent_name,omitempty"`
	// 种子描述
	TorrentDescription string `json:"torrent_description,omitempty"`
	// 站点
	TorrentSite string `json:"torrent_site,omitempty"`
	// 下载用户
	UserID string `json:"userid,omitempty"`
	// 下载用户名
	Username string `json:"username,omitempty"`
	// 下载渠道
	Channel string `json:"channel,omitempty"`
	// 创建时间
	Date string `json:"date,omitempty"`
	// 备注
	Note any `json:"note,omitempty"`
	// 自定义媒体类别
	MediaCategory string `json:"media_category,omitempty"`
	// 自定义剧集组
	EpisodeGroup string `json:"episode_group,omitempty"`
}

// TransferHistory 转移历史
type TransferHistory struct {
	// ID
	ID int `json:"id"`
	// 源目录
	Src string `json:"src,omitempty"`
	// 目的目录
	Dest string `json:"dest,omitempty"`
	// 转移模式
	Mode string `json:"mode,omitempty"`
	// 类型：电影、电视剧
	Type string `json:"type,omitempty"`
	// 二级分类
	Category string `json:"category,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// TMDBID
	TmdbID *int `json:"tmdb_id,omitempty"`
	// IMDBID
	ImdbID string `json:"imdb_id,omitempty"`
	// TVDBID
	TvdbID *int `json:"tvdb_id,omitempty"`
	// 豆瓣ID
	DoubanID string `json:"douban_id,omitempty"`
	// 季Sxx
	Seasons string `json:"seasons,omitempty"`
	// 集Exx
	Episodes string `json:"episodes,omitempty"`
	// 海报
	Image string `json:"image,omitempty"`
	// 下载器Hash
	DownloadHash string `json:"download_hash,omitempty"`
	// 自定义剧集组
	EpisodeGroup string `json:"episode_group,omitempty"`
	// 状态 1-成功，0-失败
	Status bool `json:"status"`
	// 失败原因
	ErrMsg string `json:"errmsg,omitempty"`
	// 日期
	Date string `json:"date,omitempty"`
}
