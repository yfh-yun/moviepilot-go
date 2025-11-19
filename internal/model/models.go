package models

import (
	"gorm.io/gorm"
	"time"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User 用户表
type User struct {
	BaseModel
	Name         string `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Email        string `gorm:"size:255" json:"email"`
	PasswordHash string `gorm:"column:hashed_password;size:255" json:"-"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`
	IsSuperuser  bool   `gorm:"default:false" json:"is_superuser"`
	Avatar       string `gorm:"size:500" json:"avatar"`
	IsOTP        bool   `gorm:"column:is_otp;default:false" json:"is_otp"`
	OTPSecret    string `gorm:"column:otp_secret;size:255" json:"-"`
	Permissions  string `gorm:"type:json" json:"permissions"`
	Settings     string `gorm:"type:json" json:"settings"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Media 媒体表
type Media struct {
	BaseModel
	TMDBID        *int     `gorm:"index" json:"tmdb_id"`
	IMDBID        *string  `gorm:"size:20" json:"imdb_id"`
	TVDBID        *int     `gorm:"index" json:"tvdb_id"`
	DoubanID      *string  `gorm:"index;size:50" json:"douban_id"`
	BangumiID     *int     `gorm:"index" json:"bangumi_id"`
	Title         string   `gorm:"size:500;not null" json:"title"`
	OriginalTitle string   `gorm:"size:500" json:"original_title"`
	Year          *string  `gorm:"size:10" json:"year"`
	Type          string   `gorm:"size:20;not null" json:"type"` // movie, tv
	Season        *int     `json:"season"`
	Episode       *int     `json:"episode"`
	Poster        string   `gorm:"size:500" json:"poster"`
	Backdrop      string   `gorm:"size:500" json:"backdrop"`
	Vote          *float64 `json:"vote"`
	Description   string   `gorm:"type:text" json:"description"`
	Genres        string   `gorm:"type:json" json:"genres"`
	Countries     string   `gorm:"type:json" json:"countries"`
	Language      string   `gorm:"size:10" json:"language"`
	Runtime       *int     `json:"runtime"`
	Status        string   `gorm:"size:20;default:active" json:"state"`
}

// TableName 指定表名
func (Media) TableName() string {
	return "medias"
}

// Subscribe 订阅表
type Subscribe struct {
	BaseModel
	Name            string     `gorm:"size:500;not null;index" json:"name"`
	Year            *string    `gorm:"size:10" json:"year"`
	Type            string     `gorm:"size:20" json:"type"`
	Keyword         string     `gorm:"size:500" json:"keyword"`
	TMDBID          *int       `gorm:"index" json:"tmdb_id"`
	IMDBID          *string    `gorm:"size:20" json:"imdb_id"`
	TVDBID          *int       `gorm:"index" json:"tvdb_id"`
	DoubanID        *string    `gorm:"index;size:50" json:"douban_id"`
	BangumiID       *int       `gorm:"index;size:50" json:"bangumi_id"`
	MediaID         *string    `gorm:"index;size:100" json:"media_id"`
	Season          *int       `json:"season"`
	Poster          string     `gorm:"size:500" json:"poster"`
	Backdrop        string     `gorm:"size:500" json:"backdrop"`
	Vote            *float64   `json:"vote"`
	Description     string     `gorm:"type:text" json:"description"`
	Filter          string     `gorm:"type:text" json:"filter"`
	Include         string     `gorm:"type:text" json:"include"`
	Exclude         string     `gorm:"type:text" json:"exclude"`
	Quality         string     `gorm:"size:100" json:"quality"`
	Resolution      string     `gorm:"size:100" json:"resolution"`
	Effect          string     `gorm:"size:100" json:"effect"`
	TotalEpisode    *int       `gorm:"column:total_episode" json:"total_episode"`
	StartEpisode    *int       `gorm:"column:start_episode" json:"start_episode"`
	LackEpisode     *int       `gorm:"column:lack_episode" json:"lack_episode"`
	Note            string     `gorm:"type:json" json:"note"`
	State           string     `gorm:"size:10;not null;index;default:N" json:"state"` // N-新建 R-订阅中 P-待定 S-暂停
	LastUpdate      *time.Time `gorm:"column:last_update" json:"last_update"`
	Username        string     `gorm:"size:100" json:"username"`
	Sites           string     `gorm:"type:json" json:"sites"`
	Downloader      string     `gorm:"size:50" json:"downloader"`
	BestVersion     int        `gorm:"column:best_version;default:0" json:"best_version"`
	CurrentPriority *int       `gorm:"column:current_priority" json:"current_priority"`
	MediaCategory   string     `gorm:"column:media_category;size:100" json:"media_category"`
	EpisodeGroup    string     `gorm:"column:episode_group;size:100" json:"episode_group"`
}

// TableName 指定表名
func (Subscribe) TableName() string {
	return "subscribes"
}

// DownloadHistory 下载历史表
type DownloadHistory struct {
	BaseModel
	Path           string     `gorm:"size:500;not null;index" json:"path"`                        // 保存路径
	Type           string     `gorm:"size:20;not null" json:"type"`                              // 类型 电影/电视剧
	Title          string     `gorm:"size:500;not null" json:"title"`                            // 标题
	Year           *string    `gorm:"size:10" json:"year"`                                      // 年份
	TMDBID         *int       `gorm:"index" json:"tmdb_id"`
	IMDBID         *string    `gorm:"size:20" json:"imdb_id"`
	TVDBID         *int       `gorm:"index" json:"tvdb_id"`
	DoubanID       *string    `gorm:"index;size:50" json:"douban_id"`
	Seasons        *string    `gorm:"column:seasons;size:20" json:"seasons"`                    // Sxx
	Episodes       *string    `gorm:"column:episodes;size:20" json:"episodes"`                  // Exx
	Image          string     `gorm:"size:500" json:"image"`                                    // 海报
	Downloader     string     `gorm:"size:50" json:"downloader"`                               // 下载器
	DownloadHash   string     `gorm:"column:download_hash;size:100;index" json:"download_hash"` // 下载任务Hash
	TorrentName    string     `gorm:"column:torrent_name;size:500" json:"torrent_name"`        // 种子名称
	TorrentDesc    string     `gorm:"column:torrent_description;size:1000" json:"torrent_description"` // 种子描述
	TorrentSite    string     `gorm:"column:torrent_site;size:100" json:"torrent_site"`        // 种子站点
	UserID         string     `gorm:"column:userid;size:100" json:"user_id"`                   // 下载用户
	Username       string     `gorm:"size:100" json:"username"`                                // 下载用户名/插件名
	Channel        string     `gorm:"size:50" json:"channel"`                                 // 下载渠道
	Date           string     `gorm:"size:50" json:"date"`                                     // 创建时间
	Note           string     `gorm:"type:json" json:"note"`                                   // 附加信息
	MediaCategory  string     `gorm:"column:media_category;size:100" json:"media_category"`   // 自定义媒体类别
	EpisodeGroup   string     `gorm:"column:episode_group;size:100" json:"episode_group"`     // 剧集组
	
	// 新增字段以兼容现有功能
	SavePath       string     `gorm:"column:save_path;size:500" json:"save_path"`               // 保存路径
	TotalSize      int64      `gorm:"column:total_size" json:"total_size"`                      // 总大小
	LeftSize       int64      `gorm:"column:left_size" json:"left_size"`                        // 剩余大小
	State          string     `gorm:"size:20;index" json:"state"`                              // downloading, completed, failed, paused
	Progress       float64    `json:"progress"`                                               // 进度
	Speed          int64      `json:"speed"`                                                  // 速度
	SiteName       string     `gorm:"column:site_name;size:100" json:"site_name"`              // 站点名称
	SiteUserID     *int       `gorm:"column:site_user_id" json:"site_user_id"`                  // 站点用户ID
	CreateTime     *time.Time `gorm:"column:create_time" json:"create_time"`                    // 创建时间
	UpdateTime     *time.Time `gorm:"column:update_time" json:"update_time"`                    // 更新时间
	CompletedAt    *time.Time `gorm:"column:completed_at" json:"completed_at"`                 // 完成时间
}

// TableName 指定表名
func (DownloadHistory) TableName() string {
	return "downloadhistories"
}

// TransferHistory 转移历史表
type TransferHistory struct {
	BaseModel
	Type          string     `gorm:"size:20;not null" json:"type"`
	Title         string     `gorm:"size:500;not null" json:"title"`
	Year          *string    `gorm:"size:10" json:"year"`
	Season        *int       `json:"season"`
	Episode       *int       `json:"episode"`
	TMDBID        *int       `gorm:"index" json:"tmdb_id"`
	IMDBID        *string    `gorm:"size:20" json:"imdb_id"`
	TVDBID        *int       `gorm:"index" json:"tvdb_id"`
	DoubanID      *string    `gorm:"size:50" json:"douban_id"`
	Source        string     `gorm:"size:500;not null" json:"source"`
	SourcePath    string     `gorm:"column:source_path;size:1000;not null" json:"source_path"`
	Target        string     `gorm:"size:500;not null" json:"target"`
	TargetPath    string     `gorm:"column:target_path;size:1000;not null" json:"target_path"`
	Status        string     `gorm:"size:20;default:pending" json:"status"` // pending, success, failed
	FailReason    string     `gorm:"column:fail_reason;size:500" json:"fail_reason"`
	Date          *time.Time `json:"date"`
	DownloadHash  string     `gorm:"column:download_hash;size:100" json:"download_hash"`
	UserID        string     `gorm:"column:userid;size:100" json:"user_id"`
	Username      string     `gorm:"size:100" json:"username"`
	Note          string     `gorm:"type:json" json:"note"`
	MediaCategory string     `gorm:"column:media_category;size:100" json:"media_category"`
	EpisodeGroup  string     `gorm:"column:episode_group;size:100" json:"episode_group"`
}

// TableName 指定表名
func (TransferHistory) TableName() string {
	return "transfer_histories"
}

// MediaServer 媒体服务器配置表
type MediaServer struct {
	BaseModel
	Name     string `gorm:"size:100;not null" json:"name"`
	Type     string `gorm:"size:20;not null" json:"type"` // emby, jellyfin, plex
	Host     string `gorm:"size:200;not null" json:"host"`
	Port     int    `gorm:"not null" json:"port"`
	SSL      bool   `gorm:"default:false" json:"ssl"`
	APIKey   string `gorm:"column:api_key;size:100;not null" json:"api_key"`
	Username string `gorm:"size:100" json:"username"`
	Password string `gorm:"size:255" json:"-"`
	IsActive bool   `gorm:"column:is_active;default:true" json:"is_active"`
	SyncLibs string `gorm:"column:sync_libs;type:json" json:"sync_libs"`
	Settings string `gorm:"type:json" json:"settings"`
}

// MediaServerItem 媒体服务器项目表 (对应Python中的MediaServerItem)
type MediaServerItem struct {
	BaseModel
	Server      string     `gorm:"size:50;not null;index" json:"server"`      // 服务器类型 emby/jellyfin/plex
	Library     string     `gorm:"size:100" json:"library"`              // 媒体库
	ItemID      string     `gorm:"column:item_id;size:100;not null;index" json:"item_id"` // 媒体项ID
	TMDBID      *int       `gorm:"index" json:"tmdb_id"`          // TMDBID
	Type        string     `gorm:"size:20;not null" json:"type"`         // 类型
	Title       string     `gorm:"size:500;not null;index" json:"title"`   // 标题
	OriginalTitle string     `gorm:"column:original_title;size:500" json:"original_title"` // 原标题
	Year        *int       `gorm:"index" json:"year"`              // 年份
	Season      *int       `gorm:"index" json:"season"`             // 季
	Episode     *int       `gorm:"index" json:"episode"`            // 集
	IMDBID      *string    `gorm:"size:20" json:"imdb_id"`          // IMDBID
	TVDBID      *int       `gorm:"index" json:"tvdb_id"`          // TVDBID
	Poster      string     `gorm:"size:500" json:"poster"`             // 海报
	Backdrop    string     `gorm:"size:500" json:"backdrop"`           // 背景图
	Overview    string     `gorm:"type:text" json:"overview"`          // 简介
	Genre       string     `gorm:"size:500" json:"genre"`              // 类型
	Rating      *float64   `json:"rating"`                       // 评分
	ParentItemID string     `gorm:"column:parent_item_id;size:100;index" json:"parent_item_id"` // 父级项目ID
	Path        string     `gorm:"size:1000" json:"path"`             // 路径
	Size        int64      `gorm:"default:0" json:"size"`               // 文件大小
	Duration    int        `gorm:"default:0" json:"duration"`           // 时长(秒)
	VideoCodec  string     `gorm:"column:video_codec;size:50" json:"video_codec"`   // 视频编码
	AudioCodec  string     `gorm:"column:audio_codec;size:50" json:"audio_codec"`   // 音频编码
	Resolution  string     `gorm:"size:20" json:"resolution"`            // 分辨率
	Container   string     `gorm:"size:20" json:"container"`             // 容器格式
	Bitrate     int        `gorm:"default:0" json:"bitrate"`             // 比特率
	Width       int        `gorm:"default:0" json:"width"`               // 视频宽度
	Height      int        `gorm:"default:0" json:"height"`              // 视频高度
	Framerate   float64    `gorm:"default:0" json:"framerate"`           // 帧率
	SeasonInfo  string     `gorm:"column:season_info;type:json" json:"season_info"` // 季信息
	PlayCount   int        `gorm:"column:play_count;default:0" json:"play_count"`   // 播放次数
	LastPlayed  *time.Time `gorm:"column:last_played" json:"last_played"`           // 最后播放时间
	Watched     bool       `gorm:"default:false" json:"watched"`                     // 是否已观看
	UserData    string     `gorm:"column:user_data;type:json" json:"user_data"`       // 用户数据
	ExtraData   string     `gorm:"column:extra_data;type:json" json:"extra_data"`     // 额外数据
	Note        string     `gorm:"type:json" json:"note"`                        // 备注
	LstModDate  string     `gorm:"column:lst_mod_date;size:50" json:"lst_mod_date"` // 最后修改时间
}

// TableName 指定表名
func (MediaServer) TableName() string {
	return "mediaservers"
}

// TableName 指定表名
func (MediaServerItem) TableName() string {
	return "mediaserveritems"
}

// Message 消息表 (对应Python中的Message)
type Message struct {
	BaseModel
	Channel   string     `gorm:"size:50" json:"channel"`                       // 消息渠道
	Source    string     `gorm:"size:100" json:"source"`                        // 来源
	Type      string     `gorm:"column:mtype;size:50" json:"type"`              // 消息类型
	Title     string     `gorm:"size:500" json:"title"`                         // 标题
	Text      string     `gorm:"type:text" json:"text"`                          // 文本内容
	Image     string     `gorm:"size:500" json:"image"`                         // 图片
	Link      string     `gorm:"size:500" json:"link"`                          // 链接
	UserID    string     `gorm:"column:userid;size:100" json:"user_id"`         // 用户ID
	Username  string     `gorm:"size:100" json:"username"`                      // 用户名
	Action    int        `gorm:"default:1" json:"action"`                       // 消息方向：0-接收息，1-发送消息
	RegTime   string     `gorm:"column:reg_time;size:50" json:"reg_time"`       // 注册时间
	Note      string     `gorm:"type:json" json:"note"`                          // 附件json
	
	// 兼容字段
	Content   string     `gorm:"type:text" json:"content"`                       // 内容
	Level     string     `gorm:"size:20;default:info" json:"level"`             // info, warning, error
	IsRead    bool       `gorm:"column:is_read;default:false" json:"is_read"`   // 是否已读
	ReadAt    *time.Time `gorm:"column:read_at" json:"read_at"`                // 读取时间
	Extra     string     `gorm:"type:json" json:"extra"`                       // 额外信息
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}

// TableName 指定表名
func (PluginData) TableName() string {
	return "plugindatas"
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "systemconfigs"
}

// TableName 指定表名
func (UserConfig) TableName() string {
	return "userconfigs"
}

// TableName 指定表名
func (Site) TableName() string {
	return "sites"
}

// TableName 指定表名
func (SiteUserData) TableName() string {
	return "siteuserdatas"
}

// TableName 指定表名
func (SiteStatistic) TableName() string {
	return "sitestatistics"
}

// PluginData 插件数据表 (对应Python中的PluginData)
type PluginData struct {
	BaseModel
	PluginKey string `gorm:"column:plugin_key;size:100;not null;index" json:"plugin_key"`
	DataKey   string `gorm:"column:data_key;size:100;not null;index" json:"data_key"`
	DataValue string `gorm:"column:data_value;type:text" json:"data_value"`
	UserID    string `gorm:"column:userid;size:100" json:"user_id"`
}

// SiteIcon 站点图标表 (对应Python中的SiteIcon)
type SiteIcon struct {
	BaseModel
	SiteName  string `gorm:"column:site_name;size:100;not null;uniqueIndex" json:"site_name"`
	Icon      string `gorm:"type:blob" json:"icon"`                     // 图标Base64数据
	URL       string `gorm:"size:500" json:"url"`                        // 图标URL
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"` // 更新时间
}

// SiteUserData 站点用户数据表 (对应Python中的SiteUserData)
type SiteUserData struct {
	BaseModel
	SiteName        string     `gorm:"column:site_name;size:100;not null;index" json:"site_name"`   // 站点域名
	Username        string     `gorm:"size:100" json:"username"`                             // 用户名
	UserID          string     `gorm:"column:userid;size:100" json:"userid"`             // 用户ID
	UserLevel       string     `gorm:"column:user_level;size:100" json:"user_level"`       // 用户等级
	JoinAt          *time.Time `gorm:"column:join_at" json:"join_at"`                     // 加入时间
	Uploaded        int64      `gorm:"default:0" json:"uploaded"`                       // 上传量
	Downloaded      int64      `gorm:"default:0" json:"downloaded"`                     // 下载量
	Ratio           float64    `gorm:"default:0" json:"ratio"`                         // 分享率
	Seeding         int        `gorm:"default:0" json:"seeding"`                        // 做种数
	Leeching        int        `gorm:"default:0" json:"leeching"`                       // 下载数
	Bonus           float64    `gorm:"default:0" json:"bonus"`                         // 积分
	Invites         int        `gorm:"default:0" json:"invites"`                       // 邀请数
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at"`           // 更新时间
	Domain          string     `gorm:"size:200" json:"domain"`                            // 站点域名
	UpdatedDay      string     `gorm:"column:updated_day;size:20" json:"updated_day"`    // 更新日期
	UpdatedTime     string     `gorm:"column:updated_time;size:20" json:"updated_time"`  // 更新时间
	ErrMsg          string     `gorm:"column:err_msg;size:500" json:"err_msg"`         // 错误信息
	Upload          float64    `gorm:"default:0" json:"upload"`                         // 上传量 (别名)
	Download        float64    `gorm:"default:0" json:"download"`                       // 下载量 (别名)
	MessageUnread   int        `gorm:"column:message_unread;default:0" json:"message_unread"` // 未读消息
	SeedingInfo     string     `gorm:"column:seeding_info;type:json" json:"seeding_info"` // 做种信息JSON
	SeedingSize     float64    `gorm:"column:seeding_size;default:0" json:"seeding_size"` // 做种体积
	LeechingSize    float64    `gorm:"column:leeching_size;default:0" json:"leeching_size"` // 下载体积
	SeedingInfoSize string     `gorm:"column:seeding_info_size;type:json" json:"seeding_info_size"` // 做种信息体积JSON
	MessageUnreadContents string     `gorm:"column:message_unread_contents;type:json" json:"message_unread_contents"` // 未读消息内容JSON
}

// SiteStatistic 站点统计表 (对应Python中的SiteStatistic)
type SiteStatistic struct {
	BaseModel
	SiteName   string     `gorm:"column:site_name;size:100;not null;index" json:"site_name"` // 站点名称
	Success    int        `gorm:"default:0" json:"success"`                            // 成功次数
	Fail       int        `gorm:"default:0" json:"fail"`                               // 失败次数
	Seconds    int        `gorm:"default:0" json:"seconds"`                            // 平均响应时间(秒)
	LstState   int        `gorm:"column:lst_state;default:0" json:"lst_state"`    // 最后状态 0-成功 1-失败
	LstModDate string     `gorm:"column:lst_mod_date;size:50" json:"lst_mod_date"` // 最后修改日期
	Note       string     `gorm:"type:json" json:"note"`                             // 访问时间记录
	
	// 兼容字段
	Uploaded   int64      `gorm:"default:0" json:"uploaded"`                           // 上传量
	Downloaded int64      `gorm:"default:0" json:"downloaded"`                         // 下载量
	Ratio      float64    `gorm:"default:0" json:"ratio"`                             // 分享率
	Seeding    int        `gorm:"default:0" json:"seeding"`                            // 做种数
	Leeching   int        `gorm:"default:0" json:"leeching"`                           // 下载数
	Bonus      float64    `gorm:"default:0" json:"bonus"`                             // 积分
}

// TableName 指定表名
func (PluginData) TableName() string {
	return "plugin_datas"
}

// SystemConfig 系统配置表
type SystemConfig struct {
	BaseModel
	Key    string `gorm:"size:100;not null;uniqueIndex" json:"key"`
	Value  string `gorm:"type:text" json:"value"`
	Type   string `gorm:"size:20;default:string" json:"type"` // string, int, bool, json
	Remark string `gorm:"size:500" json:"remark"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// UserConfig 用户配置表
type UserConfig struct {
	BaseModel
	UserID string `gorm:"column:userid;not null;index" json:"user_id"`
	Key    string `gorm:"size:100;not null" json:"key"`
	Value  string `gorm:"type:text" json:"value"`
	Type   string `gorm:"size:20;default:string" json:"type"` // string, int, bool, json
}

// TableName 指定表名
func (UserConfig) TableName() string {
	return "user_configs"
}

// Site 站点表
type Site struct {
	BaseModel
	Name           string `gorm:"size:100;not null" json:"name"`                      // 站点名
	Domain         string `gorm:"size:200;not null;uniqueIndex" json:"domain"`        // 域名Key
	URL            string `gorm:"size:500;not null" json:"url"`                       // 站点地址
	Pri            int    `gorm:"default:1" json:"pri"`                               // 站点优先级
	RSS            string `gorm:"size:500" json:"rss"`                                // RSS地址，未启用
	Cookie         string `gorm:"type:text" json:"cookie"`                             // Cookie
	UserAgent      string `gorm:"column:user_agent;size:500" json:"user_agent"`       // User-Agent
	APIKey         string `gorm:"column:apikey;size:500" json:"apikey"`              // ApiKey
	Token          string `gorm:"size:500" json:"token"`                              // Token
	Proxy          int    `gorm:"default:0" json:"proxy"`                             // 是否使用代理 0-否，1-是
	Filter         string `gorm:"size:500" json:"filter"`                             // 过滤规则
	Render         int    `gorm:"default:0" json:"render"`                            // 是否渲染
	Public         int    `gorm:"default:0" json:"public"`                            // 是否公开站点
	Note           string `gorm:"type:json" json:"note"`                              // 附加信息
	LimitInterval  int    `gorm:"column:limit_interval;default:0" json:"limit_interval"` // 流控单位周期
	LimitCount     int    `gorm:"column:limit_count;default:0" json:"limit_count"`   // 流控次数
	LimitSeconds   int    `gorm:"column:limit_seconds;default:0" json:"limit_seconds"` // 流控间隔
	Timeout        int    `gorm:"default:15" json:"timeout"`                          // 超时时间
	IsActive       bool   `gorm:"column:is_active;default:true" json:"is_active"`      // 是否启用
	LstModDate     string `gorm:"column:lst_mod_date;size:50" json:"lst_mod_date"`    // 创建时间
	Downloader     string `gorm:"size:50" json:"downloader"`                          // 下载器
	
	// 兼容字段
	SignURL        string `gorm:"column:sign_url;size:500" json:"sign_url"`           // 签到URL
	LoginPage      string `gorm:"column:login_page;size:500" json:"login_page"`       // 登录页面
	ProxyURL       string `gorm:"column:proxy;size:500" json:"proxy_url"`             // 代理URL
	RenderBool     bool   `gorm:"column:render_bool;default:false" json:"render_bool"`  // 是否渲染(布尔值)
	PublicBool     bool   `gorm:"column:public_bool;default:false" json:"public_bool"`  // 是否公开站点(布尔值)
	IsRSS          bool   `gorm:"column:is_rss;default:false" json:"is_rss"`          // 是否启用RSS
	Subscribed     bool   `gorm:"default:false" json:"subscribed"`                     // 是否订阅
	FailLimit      int    `gorm:"column:fail_limit;default:3" json:"fail_limit"`      // 失败限制
	FailCount      int    `gorm:"column:fail_count;default:0" json:"fail_count"`      // 失败次数
	Priority       int    `gorm:"default:0" json:"priority"`                          // 优先级
	Settings       string `gorm:"type:json" json:"settings"`                          // 设置
}

// TableName 指定表名
func (Site) TableName() string {
	return "sites"
}

// SiteUserData 站点用户数据表 (对应Python中的SiteUserData)
type SiteUserData struct {
	BaseModel
	SiteName    string     `gorm:"column:site_name;size:100;not null;index" json:"site_name"`
	Username    string     `gorm:"size:100" json:"username"`
	UserLevel   string     `gorm:"column:user_level;size:100" json:"user_level"`
	JoinAt      *time.Time `gorm:"column:join_at" json:"join_at"`
	Uploaded    int64      `gorm:"default:0" json:"uploaded"`
	Downloaded  int64      `gorm:"default:0" json:"downloaded"`
	Ratio       float64    `gorm:"default:0" json:"ratio"`
	Seeding     int        `gorm:"default:0" json:"seeding"`
	Leeching    int        `gorm:"default:0" json:"leeching"`
	Bonus       float64    `gorm:"default:0" json:"bonus"`
	Invites     int        `gorm:"default:0" json:"invites"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
	Domain      string     `gorm:"size:200" json:"domain"`                           // 站点域名
	UpdatedDay  string     `gorm:"column:updated_day;size:20" json:"updated_day"`    // 更新日期
	UpdatedTime string     `gorm:"column:updated_time;size:20" json:"updated_time"`  // 更新时间
	ErrMsg      string     `gorm:"column:err_msg;size:500" json:"err_msg"`           // 错误信息
}

// TableName 指定表名
func (SiteUserData) TableName() string {
	return "site_user_datas"
}

// SiteStatistic 站点统计表 (对应Python中的SiteStatistic)
type SiteStatistic struct {
	BaseModel
	SiteName   string     `gorm:"column:site_name;size:100;not null;index" json:"site_name"`
	Success    int        `gorm:"default:0" json:"success"`                     // 成功次数
	Fail       int        `gorm:"default:0" json:"fail"`                        // 失败次数
	Seconds    int        `gorm:"default:0" json:"seconds"`                     // 平均响应时间(秒)
	LstState   int        `gorm:"column:lst_state;default:0" json:"lst_state"`    // 最后状态 0-成功 1-失败
	LstModDate string     `gorm:"column:lst_mod_date;size:50" json:"lst_mod_date"` // 最后修改日期
	Note       string     `gorm:"type:json" json:"note"`                         // 访问时间记录
	
	// 兼容字段
	Date       *time.Time `json:"date"`                                        // 日期
	Uploaded   int64      `gorm:"default:0" json:"uploaded"`                   // 上传量
	Downloaded int64      `gorm:"default:0" json:"downloaded"`                 // 下载量
	Ratio      float64    `gorm:"default:0" json:"ratio"`                       // 分享率
	Seeding    int        `gorm:"default:0" json:"seeding"`                    // 做种数
	Leeching   int        `gorm:"default:0" json:"leeching"`                   // 下载数
	Bonus      float64    `gorm:"default:0" json:"bonus"`                      // 积分
}

// TableName 指定表名
func (SiteStatistic) TableName() string {
	return "site_statistics"
}

// SubscribeHistory 订阅历史表
type SubscribeHistory struct {
	BaseModel
	SubscribeID uint       `gorm:"column:subscribe_id;not null;index" json:"subscribe_id"`
	Title       string     `gorm:"size:500;not null" json:"title"`
	Year        *string    `gorm:"size:10" json:"year"`
	Type        string     `gorm:"size:20;not null" json:"type"`
	Season      *int       `json:"season"`
	Episode     *int       `json:"episode"`
	TMDBID      *int       `gorm:"index" json:"tmdb_id"`
	IMDBID      *string    `gorm:"size:20" json:"imdb_id"`
	TVDBID      *int       `gorm:"index" json:"tvdb_id"`
	DoubanID    *string    `gorm:"size:50" json:"douban_id"`
	Poster      string     `gorm:"size:500" json:"poster"`
	Description string     `gorm:"type:text" json:"description"`
	Status      string     `gorm:"size:20;default:new" json:"status"` // new, downloading, completed, failed
	Torrents    string     `gorm:"type:json" json:"torrents"`
	Date        *time.Time `json:"date"`
}

// TableName 指定表名
func (SubscribeHistory) TableName() string {
	return "subscribe_histories"
}

// DownloadFiles 下载文件表
type DownloadFiles struct {
	BaseModel
	DownloadHash string `gorm:"column:download_hash;size:100;not null;index" json:"download_hash"`
	FullPath     string `gorm:"column:full_path;size:1000;not null;index" json:"full_path"`
	SavePath     string `gorm:"column:save_path;size:1000;not null" json:"save_path"`
	FileName     string `gorm:"column:file_name;size:500;not null" json:"file_name"`
	FileSize     int64  `gorm:"column:file_size;default:0" json:"file_size"`
	State        int    `gorm:"default:0" json:"state"` // 0-正常 1-已删除
	Date         *time.Time `json:"date"`
}

// TableName 指定表名
func (DownloadFiles) TableName() string {
	return "download_files"
}

// SiteIcon 站点图标表
type SiteIcon struct {
	BaseModel
	SiteName  string     `gorm:"column:site_name;size:100;not null;uniqueIndex" json:"site_name"`
	Icon      string     `gorm:"type:blob" json:"icon"`
	URL       string     `gorm:"size:500" json:"url"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (SiteIcon) TableName() string {
	return "site_icons"
}

// Workflow 工作流表
type Workflow struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Type        string `gorm:"size:20;not null" json:"type"` // download, transfer, subscribe
	Description string `gorm:"type:text" json:"description"`
	Triggers    string `gorm:"type:json" json:"triggers"`
	Actions     string `gorm:"type:json" json:"actions"`
	Conditions  string `gorm:"type:json" json:"conditions"`
	IsEnabled   bool   `gorm:"column:is_enabled;default:true" json:"is_enabled"`
	Priority    int    `gorm:"default:0" json:"priority"`
}

// TableName 指定表名
func (Workflow) TableName() string {
	return "workflows"
}

// File 文件表
type File struct {
	BaseModel
	Path       string `gorm:"size:1000;not null;index" json:"path"`
	Name       string `gorm:"size:500;not null" json:"name"`
	Size       int64  `gorm:"default:0" json:"size"`
	Type       string `gorm:"size:50" json:"type"` // file, directory
	Extension  string `gorm:"size:50" json:"extension"`
	MimeType   string `gorm:"column:mime_type;size:100" json:"mime_type"`
	MD5        string `gorm:"size:32;index" json:"md5"`
	ParentID   *uint  `gorm:"column:parent_id" json:"parent_id"`
	IsHidden   bool   `gorm:"column:is_hidden;default:false" json:"is_hidden"`
	AccessTime time.Time `gorm:"column:access_time" json:"access_time"`
	ModifyTime time.Time `gorm:"column:modify_time" json:"modify_time"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}

// Search 搜索历史表
type Search struct {
	BaseModel
	Keyword   string `gorm:"size:500;not null;index" json:"keyword"`
	Type      string `gorm:"size:20;not null" json:"type"` // movie, tv
	TMDBID    *int   `gorm:"index" json:"tmdb_id"`
	IMDBID    *string `gorm:"size:20" json:"imdb_id"`
	Title     string `gorm:"size:500;not null" json:"title"`
	Year      *string `gorm:"size:10" json:"year"`
	Season    *int   `json:"season"`
	Episode   *int   `json:"episode"`
	Poster    string `gorm:"size:500" json:"poster"`
	Backdrop  string `gorm:"size:500" json:"backdrop"`
	Vote      *float64 `json:"vote"`
	Description string `gorm:"type:text" json:"description"`
	Username  string `gorm:"size:100" json:"username"`
}

// TableName 指定表名
func (Search) TableName() string {
	return "searches"
}

// Plugin 插件表
type Plugin struct {
	BaseModel
	Key         string `gorm:"size:100;not null;uniqueIndex" json:"key"`
	Name        string `gorm:"size:500;not null" json:"name"`
	State       bool   `gorm:"default:false" json:"state"` // false-禁用 true-启用
	Version     string `gorm:"size:50" json:"version"`
	Icon        string `gorm:"size:500" json:"icon"`
	Author      string `gorm:"size:100" json:"author"`
	Description string `gorm:"type:text" json:"description"`
	DescriptionV2 string `gorm:"column:description_v2;type:text" json:"description_v2"`
	Note        string `gorm:"type:text" json:"note"`
	HasPage     bool   `gorm:"column:has_page;default:false" json:"has_page"`
	PageURL     string `gorm:"column:page_url;size:500" json:"page_url"`
}

// TableName 指定表名
func (Plugin) TableName() string {
	return "plugins"
}

// ScrapeEvent 刮削事件表
type ScrapeEvent struct {
	BaseModel
	ID             string `gorm:"primaryKey;size:100" json:"id"`
	FilePath       string `gorm:"size:1000;not null" json:"file_path"`
	MediaType      string `gorm:"size:50" json:"media_type"` // movie, tv, episode
	ForceScrape    bool   `gorm:"column:force_scrape;default:false" json:"force_scrape"`
	GenerateNFO    bool   `gorm:"column:generate_nfo;default:true" json:"generate_nfo"`
	DownloadImages bool   `gorm:"column:download_images;default:true" json:"download_images"`
	RefreshMediaServer bool `gorm:"column:refresh_media_server;default:false" json:"refresh_media_server"`
	TriggeredBy     string `gorm:"size:100" json:"triggered_by"` // user, system, workflow
	Status          string `gorm:"size:20;default:pending" json:"status"` // pending, processing, completed, failed
	ProcessedAt     *time.Time `gorm:"column:processed_at" json:"processed_at"`
	CompletedAt     *time.Time `gorm:"column:completed_at" json:"completed_at"`
	Error           string `gorm:"type:text" json:"error"`
	RetryCount      int    `gorm:"column:retry_count;default:0" json:"retry_count"`
	MaxRetries      int    `gorm:"column:max_retries;default:3" json:"max_retries"`
}

// TableName 指定表名
func (ScrapeEvent) TableName() string {
	return "scrape_events"
}

// ScrapeResult 刮削结果
type ScrapeResult struct {
	ID                  string    `gorm:"primaryKey;size:100" json:"id"`
	EventID             string    `gorm:"column:event_id;size:100;index" json:"event_id"`
	FilePath            string    `gorm:"size:1000" json:"file_path"`
	MediaID             string    `gorm:"size:100" json:"media_id"`
	Title               string    `gorm:"size:500" json:"title"`
	Year                int       `json:"year"`
	MediaType           string    `gorm:"size:50" json:"media_type"`
	Poster              string    `gorm:"size:500" json:"poster"`
	Backdrop            string    `gorm:"size:500" json:"backdrop"`
	Overview            string    `gorm:"type:text" json:"overview"`
	Rating              float64   `json:"rating"`
	Genres              []string  `gorm:"type:json" json:"genres"`
	Status              string    `gorm:"size:20" json:"status"` // processing, completed, failed
	StartTime           time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime             *time.Time `gorm:"column:end_time" json:"end_time"`
	Duration            time.Duration `json:"duration"`
	NFOGenerated        bool      `gorm:"column:nfo_generated" json:"nfo_generated"`
	NFOError            string    `gorm:"column:nfo_error;type:text" json:"nfo_error"`
	ImagesDownloaded     bool      `gorm:"column:images_downloaded" json:"images_downloaded"`
	ImagesError         string    `gorm:"column:images_error;type:text" json:"images_error"`
	MediaServerRefreshed bool      `gorm:"column:media_server_refreshed" json:"media_server_refreshed"`
	MediaServerError     string    `gorm:"column:media_server_error;type:text" json:"media_server_error"`
	Error               string    `gorm:"type:text" json:"error"`
}

// TableName 指定表名
func (ScrapeResult) TableName() string {
	return "scrape_results"
}

// MediaInfo 媒体信息表
type MediaInfo struct {
	BaseModel
	MediaID       string    `gorm:"column:media_id;size:100;index" json:"media_id"`
	Source        string    `gorm:"size:50" json:"source"` // tmdb, imdb, tvdb
	Title         string    `gorm:"size:500;not null" json:"title"`
	OriginalTitle string    `gorm:"size:500" json:"original_title"`
	SortTitle     string    `gorm:"size:500" json:"sort_title"`
	Year          int       `json:"year"`
	MediaType     string    `gorm:"column:media_type;size:50" json:"media_type"` // movie, tv, episode
	Season        *int      `json:"season"`
	Episode       *int      `json:"episode"`
	Overview      string    `gorm:"type:text" json:"overview"`
	Poster        string    `gorm:"size:500" json:"poster"`
	Backdrop      string    `gorm:"size:500" json:"backdrop"`
	Rating        float64   `json:"rating"`
	VoteCount     int       `gorm:"column:vote_count" json:"vote_count"`
	ReleaseDate   string    `gorm:"column:release_date;size:50" json:"release_date"`
	Genres        []string  `gorm:"type:json" json:"genres"`
	Keywords      []string  `gorm:"type:json" json:"keywords"`
	FilePath      string    `gorm:"column:file_path;size:1000;index" json:"file_path"`
	FileSize      int64     `gorm:"column:file_size" json:"file_size"`
	FileHash      string    `gorm:"column:file_hash;size:64;index" json:"file_hash"`
	RecognizedAt   time.Time `gorm:"column:recognized_at" json:"recognized_at"`
	SyncedAt      time.Time `gorm:"column:synced_at" json:"synced_at"`
	Extra         map[string]interface{} `gorm:"type:json" json:"extra"`
}

// TableName 指定表名
func (MediaInfo) TableName() string {
	return "media_infos"
}

// MediaDetail 媒体详细信息
type MediaDetail struct {
	MediaID       string                 `json:"media_id"`
	Source        string                 `json:"source"`
	Title         string                 `json:"title"`
	OriginalTitle string                 `json:"original_title"`
	SortTitle     string                 `json:"sort_title"`
	Year          int                    `json:"year"`
	MediaType     string                 `json:"media_type"`
	Season        *int                   `json:"season"`
	Episode       *int                   `json:"episode"`
	Overview      string                 `json:"overview"`
	Poster        string                 `json:"poster"`
	Backdrop      string                 `json:"backdrop"`
	Logo          string                 `json:"logo"`
	Banner        string                 `json:"banner"`
	Thumb         string                 `json:"thumb"`
	Rating        float64                `json:"rating"`
	VoteCount     int                    `json:"vote_count"`
	Popularity    float64                `json:"popularity"`
	ReleaseDate   string                 `json:"release_date"`
	Runtime       string                 `json:"runtime"`
	Status        string                 `json:"status"`
	Genres        []string               `json:"genres"`
	Keywords      []string               `json:"keywords"`
	ProductionCompanies []string          `json:"production_companies"`
	ProductionCountries []string          `json:"production_countries"`
	SpokenLanguages []string             `json:"spoken_languages"`
	Networks      []string               `json:"networks"`
	Credits       map[string]interface{}  `json:"credits"`
	Videos        []Video                `json:"videos"`
	Images        map[string][]string      `json:"images"`
	ExternalIDs   map[string]string      `json:"external_ids"`
	Extra         map[string]interface{}  `json:"extra"`
}

// Video 视频信息
type Video struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Site    string `json:"site"`
	Type    string `json:"type"`
	Size    int    `json:"size"`
	Official bool   `json:"official"`
}

// MediaStatistics 媒体统计信息
type MediaStatistics struct {
	TotalCount     int `json:"total_count"`
	MovieCount     int `json:"movie_count"`
	TVShowCount    int `json:"tvshow_count"`
	EpisodeCount   int `json:"episode_count"`
	TotalSize      int64 `json:"total_size"`
	AverageSize    int64 `json:"average_size"`
	LastScanned    time.Time `json:"last_scanned"`
	RecentlyAdded   int `json:"recently_added"`
	GenreDistribution map[string]int `json:"genre_distribution"`
	YearDistribution   map[string]int `json:"year_distribution"`
}

// Actor 演员信息（用于NFO）
type Actor struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Order   int    `json:"order"`
	Thumb   string `json:"thumb"`
	Profile string `json:"profile"`
}

// MovieNFO 电影NFO信息
type MovieNFO struct {
	Title         string    `xml:"title"`
	OriginalTitle string    `xml:"originaltitle"`
	SortTitle     string    `xml:"sorttitle"`
	Tagline       string    `xml:"tagline"`
	Plot          string    `xml:"plot"`
	Outline       string    `xml:"outline"`
	Runtime       string    `xml:"runtime"`
	Thumb         string    `xml:"thumb"`
	Fanart        string    `xml:"fanart"`
	Poster        string    `xml:"poster"`
	Logo          string    `xml:"logo"`
	ClearArt      string    `xml:"clearart"`
	Banner        string    `xml:"banner"`
	IMDBID        string    `xml:"imdbid"`
	TMDBID        string    `xml:"tmdbid"`
	Rating        float64   `xml:"rating"`
	Votes         int       `xml:"votes"`
	Year          int       `xml:"year"`
	Top250        int       `xml:"top250"`
	Trailer       string    `xml:"trailer"`
	Watched       bool      `xml:"watched"`
	PlayCount     int       `xml:"playcount"`
	DateAdded     string    `xml:"dateadded"`
	LastPlayed    string    `xml:"lastplayed"`
	Genres        []string  `xml:"genre"`
	Tags          []string  `xml:"tag"`
	Credits       []string  `xml:"credits"`
	Directors     []string  `xml:"director"`
	Actors        []Actor   `xml:"actor"`
	Studios       []string  `xml:"studio"`
	Countries     []string  `xml:"country"`
	FileInfo      *FileInfo `xml:"fileinfo"`
}

// TVShowNFO 电视剧NFO信息
type TVShowNFO struct {
	Title         string    `xml:"title"`
	OriginalTitle string    `xml:"originaltitle"`
	SortTitle     string    `xml:"sorttitle"`
	Plot          string    `xml:"plot"`
	Outline       string    `xml:"outline"`
	Runtime       string    `xml:"runtime"`
	Thumb         string    `xml:"thumb"`
	Fanart        string    `xml:"fanart"`
	Poster        string    `xml:"poster"`
	Season        int       `xml:"season"`
	Episode       int       `xml:"episode"`
	IMDBID        string    `xml:"imdbid"`
	TMDBID        string    `xml:"tmdbid"`
	TVDBID        string    `xml:"tvdbid"`
	Rating        float64   `xml:"rating"`
	Votes         int       `xml:"votes"`
	Year          int       `xml:"year"`
	Premiered     string    `xml:"premiered"`
	Ended         string    `xml:"ended"`
	Status        string    `xml:"status"`
	Watched       bool      `xml:"watched"`
	PlayCount     int       `xml:"playcount"`
	DateAdded     string    `xml:"dateadded"`
	LastPlayed    string    `xml:"lastplayed"`
	Genres        []string  `xml:"genre"`
	Tags          []string  `xml:"tag"`
	Credits       []string  `xml:"credits"`
	Directors     []string  `xml:"director"`
	Actors        []Actor   `xml:"actor"`
	Studios       []string  `xml:"studio"`
	Countries     []string  `xml:"country"`
	FileInfo      *FileInfo `xml:"fileinfo"`
}

// EpisodeNFO 剧集NFO信息
type EpisodeNFO struct {
	Title         string    `xml:"title"`
	ShowTitle     string    `xml:"showtitle"`
	Season        int       `xml:"season"`
	Episode       int       `xml:"episode"`
	Plot          string    `xml:"plot"`
	Outline       string    `xml:"outline"`
	Runtime       string    `xml:"runtime"`
	Thumb         string    `xml:"thumb"`
	Fanart        string    `xml:"fanart"`
	IMDBID        string    `xml:"imdbid"`
	TMDBID        string    `xml:"tmdbid"`
	TVDBID        string    `xml:"tvdbid"`
	Rating        float64   `xml:"rating"`
	Votes         int       `xml:"votes"`
	Year          int       `xml:"year"`
	Aired         string    `xml:"aired"`
	Watched       bool      `xml:"watched"`
	PlayCount     int       `xml:"playcount"`
	DateAdded     string    `xml:"dateadded"`
	LastPlayed    string    `xml:"lastplayed"`
	Directors     []string  `xml:"director"`
	Actors        []Actor   `xml:"actor"`
	Credits       []string  `xml:"credits"`
	FileInfo      *FileInfo `xml:"fileinfo"`
}

// FileInfo NFO文件信息
type FileInfo struct {
	StreamDetails []StreamDetail `xml:"streamdetails"`
}

// StreamDetail 流详情
type StreamDetail struct {
	Type         string `xml:"type"`
	Language     string `xml:"language"`
	Codec        string `xml:"codec"`
	CodecID      string `xml:"codecid"`
	Width        int    `xml:"width"`
	Height       int    `xml:"height"`
	Bitrate      int    `xml:"bitrate"`
	Duration     string `xml:"duration"`
	Channels     int    `xml:"channels"`
	SamplingRate string `xml:"samplingrate"`
	Default      bool   `xml:"default"`
	Forced       bool   `xml:"forced"`
}

// MediaInfo 媒体信息（用于缓存和识别）
type MediaInfo struct {
	ID           string    `json:"id"`
	MediaID      string    `json:"media_id"`
	Title        string    `json:"title"`
	OriginalTitle string   `json:"original_title"`
	Year         int       `json:"year"`
	MediaType    string    `json:"media_type"` // movie, tv, episode
	Season       *int      `json:"season"`
	Episode      *int      `json:"episode"`
	Overview     string    `json:"overview"`
	Poster       string    `json:"poster"`
	Backdrop     string    `json:"backdrop"`
	Rating       float64   `json:"rating"`
	Genres       []string  `json:"genres"`
	ReleaseDate  string    `json:"release_date"`
	IMDBID       string    `json:"imdb_id"`
	TMDBID       int       `json:"tmdb_id"`
	TVDBID       int       `json:"tvdb_id"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	RecognizedAt time.Time `json:"recognized_at"`
	SyncedAt     time.Time `json:"synced_at"`
	Extra        map[string]interface{} `json:"extra"`
}

// MediaDetail 媒体详细信息
type MediaDetail struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Overview    string                   `json:"overview"`
	Poster      string                   `json:"poster"`
	Backdrop    string                   `json:"backdrop"`
	Rating      float64                  `json:"rating"`
	Genres      []string                 `json:"genres"`
	Runtime     int                      `json:"runtime"`
	ReleaseDate string                   `json:"release_date"`
	IMDBID      string                   `json:"imdb_id"`
	TMDBID      int                      `json:"tmdb_id"`
	TVDBID      int                      `json:"tvdb_id"`
	Extra       map[string]interface{}   `json:"extra"`
}

// ScrapeEvent 刮削事件
type ScrapeEvent struct {
	ID                string `json:"id"`
	FilePath          string `json:"file_path"`
	MediaType         string `json:"media_type"`
	ForceScrape       bool   `json:"force_scrape"`
	GenerateNFO       bool   `json:"generate_nfo"`
	DownloadImages    bool   `json:"download_images"`
	RefreshMediaServer bool   `json:"refresh_media_server"`
	UserID            string `json:"user_id"`
	CreatedAt         time.Time `json:"created_at"`
	ProcessedAt       *time.Time `json:"processed_at"`
}

// ScrapeResult 刮削结果
type ScrapeResult struct {
	EventID             string     `json:"event_id"`
	MediaID             string     `json:"media_id"`
	FilePath            string     `json:"file_path"`
	Title               string     `json:"title"`
	Year                int        `json:"year"`
	MediaType           string     `json:"media_type"`
	Status              string     `json:"status"` // processing, completed, failed
	StartTime           time.Time  `json:"start_time"`
	EndTime             time.Time  `json:"end_time"`
	Duration            time.Duration `json:"duration"`
	Poster              string     `json:"poster"`
	Backdrop            string     `json:"backdrop"`
	Overview            string     `json:"overview"`
	Rating              float64    `json:"rating"`
	Genres              []string   `json:"genres"`
	NFOGenerated        bool       `json:"nfo_generated"`
	NFOError            string     `json:"nfo_error,omitempty"`
	ImagesDownloaded    bool       `json:"images_downloaded"`
	ImagesError         string     `json:"images_error,omitempty"`
	MediaServerRefreshed bool       `json:"media_server_refreshed"`
	MediaServerError     string     `json:"media_server_error,omitempty"`
	Error               string     `json:"error,omitempty"`
}

// 下载历史记录
type DownloadHistory struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	MediaID     string     `json:"media_id" gorm:"index"`
	MediaTitle  string     `json:"media_title"`
	MediaType   string     `json:"media_type" gorm:"index"`
	Source      string     `json:"source" gorm:"index"`
	Status      string     `json:"status" gorm:"index"` // success, failed, pending
	Size        int64      `json:"size"`
	Downloaded  int64      `json:"downloaded"`
	Speed       int64      `json:"speed"` // bytes per second
	Progress    int        `json:"progress"` // percentage
	Error       string     `json:"error,omitempty"`
	CreateTime  time.Time  `json:"create_time" gorm:"index"`
	UpdateTime  time.Time  `json:"update_time"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	
	// 关联字段
	Media      *MediaInfo `json:"media,omitempty" gorm:"foreignKey:MediaID"`
}

// 转移历史记录
type TransferHistory struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	MediaID      string     `json:"media_id" gorm:"index"`
	MediaTitle   string     `json:"media_title"`
	MediaType    string     `json:"media_type" gorm:"index"`
	SourcePath   string     `json:"source_path"`
	DestPath     string     `json:"dest_path"`
	Status       string     `json:"status" gorm:"index"` // success, failed, pending
	Size         int64      `json:"size"`
	Transferred  int64      `json:"transferred"`
	Speed        int64      `json:"speed"` // bytes per second
	Progress     int        `json:"progress"` // percentage
	TransferMode string     `json:"transfer_mode" gorm:"index"` // move, copy, sync
	Error        string     `json:"error,omitempty"`
	CreateTime   time.Time  `json:"create_time" gorm:"index"`
	UpdateTime   time.Time  `json:"update_time"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	
	// 关联字段
	Media       *MediaInfo `json:"media,omitempty" gorm:"foreignKey:MediaID"`
}

// 订阅历史记录
type SubscribeHistory struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	MediaID    string     `json:"media_id" gorm:"index"`
	MediaTitle string     `json:"media_title"`
	MediaType  string     `json:"media_type" gorm:"index"`
	Season     int        `json:"season" gorm:"index"`
	Episode    int        `json:"episode" gorm:"index"`
	Status     string     `json:"status" gorm:"index"` // success, failed, pending
	Error      string     `json:"error,omitempty"`
	CreateTime time.Time  `json:"create_time" gorm:"index"`
	UpdateTime time.Time  `json:"update_time"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	
	// 关联字段
	Media      *MediaInfo `json:"media,omitempty" gorm:"foreignKey:MediaID"`
}

// 系统历史记录
type SystemHistory struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Type       string    `json:"type" gorm:"index"` // download, transfer, subscribe, plugin
	Level      string    `json:"level" gorm:"index"` // info, warning, error
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	CreateTime time.Time `json:"create_time" gorm:"index"`
}

// 历史记录统计信息
type HistoryStats struct {
	// 下载统计
	DownloadTotal     int64 `json:"download_total"`
	DownloadSuccess   int64 `json:"download_success"`
	DownloadFailed    int64 `json:"download_failed"`
	DownloadPending  int64 `json:"download_pending"`
	DownloadSize     int64 `json:"download_size"` // 总下载大小（字节）
	
	// 转移统计
	TransferTotal     int64 `json:"transfer_total"`
	TransferSuccess  int64 `json:"transfer_success"`
	TransferFailed   int64 `json:"transfer_failed"`
	TransferPending  int64 `json:"transfer_pending"`
	TransferSize    int64 `json:"transfer_size"` // 总转移大小（字节）
	
	// 订阅统计
	SubscribeTotal     int64 `json:"subscribe_total"`
	SubscribeSuccess   int64 `json:"subscribe_success"`
	SubscribeFailed    int64 `json:"subscribe_failed"`
	SubscribePending  int64 `json:"subscribe_pending"`
	
	// 系统统计
	SystemTotal       int64 `json:"system_total"`
	SystemInfoCount   int64 `json:"system_info_count"`
	SystemWarningCount int64 `json:"system_warning_count"`
	SystemErrorCount   int64 `json:"system_error_count"`
	
	// 时间统计
	LastUpdateTime    time.Time `json:"last_update_time"`
	FirstRecordTime   time.Time `json:"first_record_time"`
	TodayActivityCount int64   `json:"today_activity_count"`
	WeekActivityCount  int64   `json:"week_activity_count"`
	MonthActivityCount int64  `json:"month_activity_count"`
}
