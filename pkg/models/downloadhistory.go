package models

// DownloadHistory 下载历史记录
type DownloadHistory struct {
	// ID
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
	
	// 保存路径
	Path string `json:"path,omitempty" gorm:"index"`
	
	// 类型：电影、电视剧
	Type string `json:"type,omitempty"`
	
	// 标题
	Title string `json:"title,omitempty"`
	
	// 年份
	Year string `json:"year,omitempty"`
	
	// TMDBID
	TmdbID int `json:"tmdbid,omitempty" gorm:"index"`
	
	// IMDBID
	ImdbID string `json:"imdbid,omitempty"`
	
	// TVDBID
	TvdbID int `json:"tvdbid,omitempty"`
	
	// 豆瓣ID
	DoubanID string `json:"doubanid,omitempty"`
	
	// 季Sxx
	Seasons string `json:"seasons,omitempty"`
	
	// 集Exx
	Episodes string `json:"episodes,omitempty"`
	
	// 海报
	Image string `json:"image,omitempty"`
	
	// 下载器
	Downloader string `json:"downloader,omitempty"`
	
	// 下载器Hash
	DownloadHash string `json:"download_hash,omitempty" gorm:"index"`
	
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
	Note interface{} `json:"note,omitempty"`
	
	// 自定义媒体类别
	MediaCategory string `json:"media_category,omitempty"`
	
	// 自定义剧集组
	EpisodeGroup string `json:"episode_group,omitempty"`
}

// TableName 设置表名
func (DownloadHistory) TableName() string {
	return "downloadhistory"
}

// DownloadFiles 下载文件记录
type DownloadFiles struct {
	// ID
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
	
	// 下载器
	Downloader string `json:"downloader,omitempty"`
	
	// 下载任务Hash
	DownloadHash string `json:"download_hash,omitempty" gorm:"index"`
	
	// 完整路径
	Fullpath string `json:"fullpath,omitempty" gorm:"index"`
	
	// 保存路径
	Savepath string `json:"savepath,omitempty" gorm:"index"`
	
	// 文件相对路径/名称
	Filepath string `json:"filepath,omitempty"`
	
	// 种子名称
	Torrentname string `json:"torrentname,omitempty"`
	
	// 状态 0-已删除 1-正常
	State int `json:"state" gorm:"default:1;not null"`
}

// TableName 设置表名
func (DownloadFiles) TableName() string {
	return "downloadfiles"
}