package models

import (
	"time"
)

// TransferTorrent 待转移任务信�?type TransferTorrent struct {
	Downloader string     `json:"downloader,omitempty"`
	Title      string     `json:"title,omitempty"`
	Path       string     `json:"path,omitempty"`
	Hash       string     `json:"hash,omitempty"`
	Tags       string     `json:"tags,omitempty"`
	Size       int        `json:"size,omitempty"`
	UserID     string     `json:"userid,omitempty"`
	Progress   float64    `json:"progress,omitempty"`
	State      string     `json:"state,omitempty"`
}

// DownloadingTorrent 下载中任务信�?type DownloadingTorrent struct {
	Downloader      string                 `json:"downloader,omitempty"`
	Hash            string                 `json:"hash,omitempty"`
	Title           string                 `json:"title,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Year            string                 `json:"year,omitempty"`
	SeasonEpisode   string                 `json:"season_episode,omitempty"`
	Size            float64                `json:"size,omitempty"`
	Progress        float64                `json:"progress,omitempty"`
	State           string                 `json:"state,omitempty"`
	Upspeed         string                 `json:"upspeed,omitempty"`
	Dlspeed         string                 `json:"dlspeed,omitempty"`
	Media           map[string]interface{} `json:"media,omitempty"`
	UserID          string                 `json:"userid,omitempty"`
	Username        string                 `json:"username,omitempty"`
	LeftTime        string                 `json:"left_time,omitempty"`
}

// TransferTask 文件整理任务
type TransferTask struct {
	FileItem              FileItem             `json:"fileitem"`
	Meta                  interface{}          `json:"meta,omitempty"`
	MediaInfo             interface{}          `json:"mediainfo,omitempty"`
	TargetDirectory       *TransferDirectoryConf `json:"target_directory,omitempty"`
	TargetStorage         string               `json:"target_storage,omitempty"`
	TargetPath            string               `json:"target_path,omitempty"`
	TransferType          string               `json:"transfer_type,omitempty"`
	Scrape                bool                 `json:"scrape,omitempty"`
	LibraryTypeFolder     bool                 `json:"library_type_folder,omitempty"`
	LibraryCategoryFolder bool                 `json:"library_category_folder,omitempty"`
	EpisodesInfo          []TmdbEpisode        `json:"episodes_info,omitempty"`
	Username              string               `json:"username,omitempty"`
	Downloader            string               `json:"downloader,omitempty"`
	DownloadHash          string               `json:"download_hash,omitempty"`
	DownloadHistory       *DownloadHistory     `json:"download_history,omitempty"`
	Manual                bool                 `json:"manual,omitempty"`
	Background            bool                 `json:"background,omitempty"`
}

// TransferJobTask 文件整理作业任务
type TransferJobTask struct {
	FileItem     *FileItem  `json:"fileitem,omitempty"`
	Meta         *MetaInfo  `json:"meta,omitempty"`
	State        string     `json:"state,omitempty"`
	Downloader   string     `json:"downloader,omitempty"`
	DownloadHash string     `json:"download_hash,omitempty"`
}

// TransferJob 文件整理作业
type TransferJob struct {
	Media  *MediaInfo         `json:"media,omitempty"`
	Season int                `json:"season,omitempty"`
	Tasks  []TransferJobTask  `json:"tasks,omitempty"`
}

// TransferInfo 文件整理结果
type TransferInfo struct {
	// 是否成功标志
	Success bool `json:"success"`
	
	// 整理文件路径
	FileItem *FileItem `json:"fileitem,omitempty"`
	
	// 转移后的目录项，媒体的根目录
	TargetDirItem *FileItem `json:"target_diritem,omitempty"`
	
	// 转移后路�?	TargetItem *FileItem `json:"target_item,omitempty"`
	
	// 整理方式
	TransferType string `json:"transfer_type,omitempty"`
	
	// 处理文件�?	FileCount int `json:"file_count,omitempty"`
	
	// 处理文件清单
	FileList []interface{} `json:"file_list,omitempty"`
	
	// 目标文件清单
	FileListNew []interface{} `json:"file_list_new,omitempty"`
	
	// 总文件大�?	TotalSize int `json:"total_size,omitempty"`
	
	// 失败清单
	FailList []interface{} `json:"fail_list,omitempty"`
	
	// 处理字幕文件清单
	SubtitleList []interface{} `json:"subtitle_list,omitempty"`
	
	// 目标字幕文件清单
	SubtitleListNew []interface{} `json:"subtitle_list_new,omitempty"`
	
	// 处理音频文件清单
	AudioList []interface{} `json:"audio_list,omitempty"`
	
	// 目标音频文件清单
	AudioListNew []interface{} `json:"audio_list_new,omitempty"`
	
	// 错误信息
	Message string `json:"message,omitempty"`
	
	// 是否需要刮�?	NeedScrape bool `json:"need_scrape,omitempty"`
	
	// 是否需要通知
	NeedNotify bool `json:"need_notify,omitempty"`
}

// TransferQueue 异步整理队列信息
type TransferQueue struct {
	// 任务信息
	Task     *TransferTask  `json:"task,omitempty"`
	// 回调函数 (在Go中使用函数类�?
	Callback interface{}    `json:"-"`
	// 整理结果
	Result   *TransferInfo  `json:"result,omitempty"`
}

// EpisodeFormat 剧集自定义识别格�?type EpisodeFormat struct {
	Format string `json:"format,omitempty"`
	Detail string `json:"detail,omitempty"`
	Part   string `json:"part,omitempty"`
	Offset string `json:"offset,omitempty"`
}

// ManualTransferItem 手动整理�?type ManualTransferItem struct {
	// 文件�?	FileItem FileItem `json:"fileitem"`
	
	// 日志ID
	LogID int `json:"logid,omitempty"`
	
	// 目标存储
	TargetStorage string `json:"target_storage,omitempty"`
	
	// 目标路径
	TargetPath string `json:"target_path,omitempty"`
	
	// TMDB ID
	TMDBID int `json:"tmdbid,omitempty"`
	
	// 豆瓣ID
	DoubanID string `json:"doubanid,omitempty"`
	
	// 类型
	TypeName string `json:"type_name,omitempty"`
	
	// 季号
	Season int `json:"season,omitempty"`
	
	// 整理方式
	TransferType string `json:"transfer_type,omitempty"`
	
	// 自定义格�?	EpisodeFormat string `json:"episode_format,omitempty"`
	
	// 指定集数
	EpisodeDetail string `json:"episode_detail,omitempty"`
	
	// 指定PART
	EpisodePart string `json:"episode_part,omitempty"`
	
	// 集数偏移
	EpisodeOffset string `json:"episode_offset,omitempty"`
	
	// 最小文件大�?	MinFilesize int `json:"min_filesize,omitempty"`
	
	// 刮削
	Scrape bool `json:"scrape"`
	
	// 媒体库类型子目录
	LibraryTypeFolder bool `json:"library_type_folder,omitempty"`
	
	// 媒体库类别子目录
	LibraryCategoryFolder bool `json:"library_category_folder,omitempty"`
	
	// 复用历史识别信息
	FromHistory bool `json:"from_history,omitempty"`
	
	// 剧集�?	EpisodeGroup string `json:"episode_group,omitempty"`
}

// MetaInfo 媒体元信�?type MetaInfo struct {
	// 原始文件�?	OriginalName string `json:"original_name,omitempty"`
	
	// 识别后的名称
	Name string `json:"name,omitempty"`
	
	// 年份
	Year string `json:"year,omitempty"`
	
	// SXXEXX
	SeasonEpisode string `json:"season_episode,omitempty"`
	
	// SXX
	Season string `json:"season,omitempty"`
	
	// EEXX
	Episode string `json:"episode,omitempty"`
	
	// 资源类型
	ResourceType string `json:"resource_type,omitempty"`
	
	// 效果
	Effect string `json:"effect,omitempty"`
	
	// 视频编码
	VideoEncode string `json:"video_encode,omitempty"`
	
	// 音频编码
	AudioEncode string `json:"audio_encode,omitempty"`
	
	// 名称（兼容旧版本�?	Title string `json:"title,omitempty"`
	
	// 种子名称（兼容旧版本�?	TorrentName string `json:"torrent_name,omitempty"`
	
	// 描述（兼容旧版本�?	Description string `json:"description,omitempty"`
	
	// 分辨�?	Resolution string `json:"resolution,omitempty"`
	
	// 发布�?	ReleaseGroup string `json:"release_group,omitempty"`
	
	// 处理日期
	Date time.Time `json:"date,omitempty"`
	
	// 其他信息
	OtherInfo map[string]interface{} `json:"other_info,omitempty"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	// 类型 电影、电视剧
	Type string `json:"type,omitempty"`
	
	// 媒体标题
	Title string `json:"title,omitempty"`
	
	// 年份
	Year string `json:"year,omitempty"`
	
	// 标题（原标题�?	OriginalTitle string `json:"original_title,omitempty"`
	
	// 季号
	Season int `json:"season,omitempty"`
	
	// TMDB ID
	TMDBID int `json:"tmdbid,omitempty"`
	
	// IMDB ID
	IMDBID string `json:"imdbid,omitempty"`
	
	// TVDB ID
	TVDBID int `json:"tvdbid,omitempty"`
	
	// 豆瓣ID
	DoubanID string `json:"doubanid,omitempty"`
	
	// 媒体原语�?	OriginalLanguage string `json:"original_language,omitempty"`
	
	// 媒体原语种名�?	OriginalLanguageName string `json:"original_language_name,omitempty"`
	
	// 媒体分类
	Category string `json:"category,omitempty"`
	
	// TMDB INFO
	TMDBInfo map[string]interface{} `json:"tmdb_info,omitempty"`
	
	// 豆瓣INFO
	DoubanInfo map[string]interface{} `json:"douban_info,omitempty"`
	
	// 季信�?	SeasonInfo map[string]interface{} `json:"season_info,omitempty"`
	
	// 剧集信息
	EpisodeInfo map[string]interface{} `json:"episode_info,omitempty"`
	
	// 演员
	Actors []interface{} `json:"actors,omitempty"`
	
	// 导演
	Directors []interface{} `json:"directors,omitempty"`
	
	// 其他信息
	OtherInfo map[string]interface{} `json:"other_info,omitempty"`
}

// NewTransferTorrent 创建一个新�?TransferTorrent 实例
func NewTransferTorrent() *TransferTorrent {
	return &TransferTorrent{
		Size:     0,
		Progress: 0.0,
	}
}

// NewDownloadingTorrent 创建一个新�?DownloadingTorrent 实例
func NewDownloadingTorrent() *DownloadingTorrent {
	return &DownloadingTorrent{
		Size:      0.0,
		Progress:  0.0,
		State:     "downloading",
		Media:     make(map[string]interface{}),
	}
}

// NewTransferTask 创建一个新�?TransferTask 实例
func NewTransferTask() *TransferTask {
	return &TransferTask{
		Scrape:                false,
		LibraryTypeFolder:     false,
		LibraryCategoryFolder: false,
		EpisodesInfo:          make([]TmdbEpisode, 0),
		Manual:                false,
		Background:            true,
	}
}

// NewTransferJobTask 创建一个新�?TransferJobTask 实例
func NewTransferJobTask() *TransferJobTask {
	return &TransferJobTask{}
}

// NewTransferJob 创建一个新�?TransferJob 实例
func NewTransferJob() *TransferJob {
	return &TransferJob{
		Tasks: make([]TransferJobTask, 0),
	}
}

// NewTransferInfo 创建一个新�?TransferInfo 实例
func NewTransferInfo() *TransferInfo {
	return &TransferInfo{
		Success:           true,
		FileCount:         0,
		FileList:          make([]interface{}, 0),
		FileListNew:       make([]interface{}, 0),
		TotalSize:         0,
		FailList:          make([]interface{}, 0),
		SubtitleList:      make([]interface{}, 0),
		SubtitleListNew:   make([]interface{}, 0),
		AudioList:         make([]interface{}, 0),
		AudioListNew:      make([]interface{}, 0),
		NeedScrape:        false,
		NeedNotify:        false,
	}
}

// NewTransferQueue 创建一个新�?TransferQueue 实例
func NewTransferQueue() *TransferQueue {
	return &TransferQueue{}
}

// NewEpisodeFormat 创建一个新�?EpisodeFormat 实例
func NewEpisodeFormat() *EpisodeFormat {
	return &EpisodeFormat{}
}

// NewManualTransferItem 创建一个新�?ManualTransferItem 实例
func NewManualTransferItem() *ManualTransferItem {
	return &ManualTransferItem{
		MinFilesize:           0,
		Scrape:                false,
		LibraryTypeFolder:     false,
		LibraryCategoryFolder: false,
		FromHistory:           false,
	}
}

// NewMetaInfo 创建一个新�?MetaInfo 实例
func NewMetaInfo() *MetaInfo {
	return &MetaInfo{
		OtherInfo: make(map[string]interface{}),
	}
}

// NewMediaInfo 创建一个新�?MediaInfo 实例
func NewMediaInfo() *MediaInfo {
	return &MediaInfo{
		TMDBInfo:    make(map[string]interface{}),
		DoubanInfo:  make(map[string]interface{}),
		SeasonInfo:  make(map[string]interface{}),
		EpisodeInfo: make(map[string]interface{}),
		Actors:      make([]interface{}, 0),
		Directors:   make([]interface{}, 0),
		OtherInfo:   make(map[string]interface{}),
	}
}
