package models
import (
	"time"
)

import "gorm.io/gorm"

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// UserConfig 用户配置表 (对应Python中的UserConfig)
type UserConfig struct {
	BaseModel
	Username string `gorm:"size:100;not null;index" json:"username"` // 用户名
	Key      string `gorm:"size:100;not null" json:"key"`         // 配置键
	Value    string `gorm:"type:text" json:"value"`               // 值
}

// SiteIcon 站点图标表 (对应Python中的SiteIcon)
type SiteIcon struct {
	BaseModel
	SiteName string `gorm:"column:site_name;size:100;not null;uniqueIndex" json:"site_name"` // 站点名称
	Icon     string `gorm:"type:blob" json:"icon"`                              // 图标Base64数据
	URL      string `gorm:"size:500" json:"url"`                                // 图标URL
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`     // 更新时间
}

// SiteUserData 站点用户数据表 (对应Python中的SiteUserData)
type SiteUserData struct {
	BaseModel
	SiteName         string     `gorm:"column:site_name;size:100;not null;index" json:"site_name"` // 站点域名
	Username           string     `gorm:"size:100" json:"username"`                             // 用户名
	UserID            string     `gorm:"column:userid;size:100" json:"userid"`               // 用户ID
	UserLevel           string     `gorm:"column:user_level;size:100" json:"user_level"`         // 用户等级
	JoinAt             *time.Time `gorm:"column:join_at" json:"join_at"`                     // 加入时间
	Uploaded            int64       `gorm:"default:0" json:"uploaded"`                        // 上传量
	Downloaded          int64       `gorm:"default:0" json:"downloaded"`                     // 下载量
	Ratio               float64     `gorm:"default:0" json:"ratio"`                           // 分享率
	Seeding              int         `gorm:"default:0" json:"seeding"`                         // 做种数
	Leeching             int         `gorm:"default:0" json:"leeching"`                        // 下载数
	Bonus                float64     `gorm:"default:0" json:"bonus"`                           // 积分
	Invites              int         `gorm:"default:0" json:"invites"`                         // 邀请数
	UpdatedAt            *time.Time `gorm:"column:updated_at" json:"updated_at"`           // 更新时间
	Domain              string     `gorm:"size:200" json:"domain"`                             // 站点域名
	UpdatedDay          string     `gorm:"column:updated_day;size:20" json:"updated_day"`   // 更新日期
	UpdatedTime          string     `gorm:"column:updated_time;size:20" json:"updated_time"` // 更新时间
	ErrMsg               string     `gorm:"column:err_msg;size:500" json:"err_msg"`         // 错误信息
	Upload              float64    `gorm:"default:0" json:"upload"`                         // 上传量 (别名)
	Download             float64    `gorm:"default:0" json:"download"`                       // 下载量 (别名)
	MessageUnread        int         `gorm:"column:message_unread;default:0" json:"message_unread"` // 未读消息
	SeedingInfo          string     `gorm:"column:seeding_info;type:json" json:"seeding_info"` // 做种信息JSON
	SeedingSize         float64    `gorm:"column:seeding_size;default:0" json:"seeding_size"` // 做种体积
	LeechingSize         float64    `gorm:"column:leeching_size;default:0" json:"leeching_size"` // 下载体积
	SeedingInfoSize     string     `gorm:"column:seeding_info_size;type:json" json:"seeding_info_size"` // 做种信息体积JSON
	MessageUnreadContents string     `gorm:"column:message_unread_contents;type:json" json:"message_unread_contents"` // 未读消息内容JSON
}

// SiteStatistic 站点统计表 (对应Python中的SiteStatistic)
type SiteStatistic struct {
	BaseModel
	SiteName   string `gorm:"column:site_name;size:100;not null;index" json:"site_name"` // 站点名称
	Success    int    `gorm:"default:0" json:"success"`                            // 成功次数
	Fail       int    `gorm:"default:0" json:"fail"`                               // 失败次数
	Seconds    int    `gorm:"default:0" json:"seconds"`                            // 平均响应时间(秒)
	LstState   int    `gorm:"column:lst_state;default:0" json:"lst_state"`    // 最后状态 0-成功 1-失败
	LstModDate string `gorm:"column:lst_mod_date;size:50" json:"lst_mod_date"` // 最后修改日期
	Note       string `gorm:"type:json" json:"note"`                             // 访问时间记录
	
	// 兼容字段
	Uploaded   int64  `gorm:"default:0" json:"uploaded"`                           // 上传量
	Downloaded int64  `gorm:"default:0" json:"downloaded"`                         // 下载量
	Ratio      float64 `gorm:"default:0" json:"ratio"`                             // 分享率
	Seeding    int    `gorm:"default:0" json:"seeding"`                            // 做种数
	Leeching   int    `gorm:"default:0" json:"leeching"`                           // 下载数
	Bonus      float64 `gorm:"default:0" json:"bonus"`                             // 积分
}

// SubscribeHistory 订阅历史表 (对应Python中的SubscribeHistory)
type SubscribeHistory struct {
	BaseModel
	Name     string `gorm:"size:500;not null;index" json:"name"` // 标题
	Year     string `gorm:"size:4" json:"year"`           // 年份
	Type     string `gorm:"size:20" json:"type"`         // 类型
	Keyword  string `gorm:"size:500" json:"keyword"`   // 搜索关键字
	TMDBID   int    `gorm:"column:tmdbid;index" json:"tmdbid"`   // TMDBID
	IMDBID   string `gorm:"size:20" json:"imdbid"`   // IMDBID
	TVDBID   int    `gorm:"column:tvdbid" json:"tvdbid"`   // TVDBID
	DoubanID string `gorm:"column:doubanid;index" json:"doubanid"` // 豆瓣ID
	BangumiID int  `gorm:"column:bangumiid;index" json:"bangumiid"` // BangumiID
	MediaID  string `gorm:"column:mediaid;index" json:"mediaid"` // 媒体ID
	Season   int    `gorm:"column:season" json:"season"`         // 季号
	Poster   string `gorm:"size:500" json:"poster"`         // 海报
	Backdrop string `gorm:"size:500" json:"backdrop"`     // 背景图
	Vote     float64 `gorm:"column:vote" json:"vote"`       // 评分，float
	Description string `gorm:"type:text" json:"description"` // 简介
	Filter   string `gorm:"size:500" json:"filter"`     // 过滤规则
	Include  string `gorm:"size:500" json:"include"`     // 包含
	Exclude  string `gorm:"size:500" json:"exclude"`     // 排除
	Quality  string `gorm:"size:100" json:"quality"`     // 质量
	Resolution string `gorm:"size:100" json:"resolution"` // 分辨率
	Effect   string `gorm:"size:100" json:"effect"`     // 特效
	Status   string `gorm:"size:20;not null;index;default:new" json:"state"` // 状态：N-新建 R-订阅中 P-待定 S-暂停
	Error    string `gorm:"size:500" json:"error,omitempty"` // 错误信息
	Files    string `gorm:"type:json" json:"files"`     // 文件列表
}

// Workflow 工作流表 (对应Python中的Workflow)
type Workflow struct {
	BaseModel
	Name        string    `gorm:"size:100;not null;index" json:"name"`         // 名称
	Description string    `gorm:"type:text" json:"description"`              // 描述
	Timer       string    `gorm:"size:100" json:"timer"`                         // 定时器
	TriggerType string    `gorm:"column:trigger_type;size:50;default:'timer'" json:"trigger_type"` // 触发类型：timer-定时触发 event-事件触发 manual-手动触发
	EventType   string    `gorm:"column:event_type;size:50" json:"event_type"`   // 事件类型（当trigger_type为event时使用）
	EventConditions string    `gorm:"column:event_conditions;type:json;default:dict" json:"event_conditions"` // 事件条件（JSON格式，用于过滤事件）
	State       string    `gorm:"size:20;not null;index;default:'W'" json:"state"` // 状态：W-等待 R-运行中 P-暂停 S-成功 F-失败
	CurrentAction string   `gorm:"column:current_action" json:"current_action"` // 已执行动作
	Result       string    `gorm:"size:500" json:"result"`         // 任务执行结果
	RunCount     int       `gorm:"column:run_count;default:0" json:"run_count"`   // 已执行次数
	Actions      string    `gorm:"type:json;default:list" json:"actions"`     // 任务列表
	Flows        string    `gorm:"type:json;default:list" json:"flows"`        // 任务流
	Context       string    `gorm:"type:json;default:dict" json:"context"`     // 执行上下文
	AddTime       string    `gorm:"column:add_time;default: CURRENT_TIMESTAMP" json:"add_time"` // 创建时间
	LastTime      string    `gorm:"column:last_time" json:"last_time"`     // 最后执行时间
	IsEnabled     bool      `gorm:"column:is_enabled;default:true" json:"is_enabled"` // 是否启用
	Priority      int       `gorm:"default:0" json:"priority"` // 优先级
}

// TransferHistoryBase 转移历史基础结构 (对应Python中的TransferHistory)
type TransferHistoryBase struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	Src          string     `gorm:"index" json:"src"`                              // 源路径
	SrcStorage  string     `gorm:"column:src_storage" json:"src_storage"`              // 源存储
	SrcFileitem string     `gorm:"column:src_fileitem;type:json;default:dict" json:"src_fileitem"` // 源文件项
	Dest         string     `gorm:"column:dest" json:"dest"`                            // 目标路径
	DestStorage  string     `gorm:"column:dest_storage" json:"dest_storage"`              // 目标存储
	DestFileitem string     `gorm:"column:dest_fileitem;type:json;default:dict" json:"dest_fileitem"` // 目标文件项
	Mode         string     `gorm:"index" json:"mode"`                             // 转移模式 move/copy/link...
	Type         string     `gorm:"index" json:"type"`                             // 类型 电影/电视剧
	Category    string     `gorm:"index" json:"category"`                         // 二级分类
	Title        string     `gorm:"index" json:"title"`                            // 标题
	Year         string     `gorm:"index" json:"year"`                             // 年份
	Seasons      string     `gorm:"column:seasons" json:"seasons"`                // Sxx
	Episodes     string     `gorm:"column:episodes" json:"episodes"`              // Exx
	Image        string     `gorm:"index" json:"image"`                            // 海报
	Downloader   string     `gorm:"index" json:"downloader"`                      // 下载器
	DownloadHash string     `gorm:"column:download_hash;size:100" json:"download_hash"` // 下载器hash
	Status       string     `gorm:"size:20;index" json:"status"`                   // 状态
	Error        string     `gorm:"size:500" json:"error,omitempty"`              // 错误信息
	CreateTime   time.Time  `gorm:"column:create_time;index" json:"create_time"` // 创建时间
	UpdateTime   time.Time  `gorm:"column:update_time" json:"update_time"`       // 更新时间
	CompletedAt  *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"` // 完成时间
}

// 扩展的TransferHistory，包含所有Python字段
type TransferHistory struct {
	TransferHistoryBase
	TMDBID      int       `gorm:"column:tmdbid;index" json:"tmdbid"`          // TMDBID
	IMDBID      string    `gorm:"column:imdbid" json:"imdbid"`         // IMDBID
	TVDBID      int       `gorm:"column:tvdbid" json:"tvdbid"`          // TVDBID
	DoubanID    string    `gorm:"column:doubanid" json:"doubanid"`       // 豆瓣ID
	FileList     string    `gorm:"column:files;type:json" json:"files"`     // 文件列表
}

// TransferFileItem 转移文件项结构
type TransferFileItem struct {
	Path     string `json:"path"`
	Storage  string `json:"storage"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
}

// TableName 方法
func (UserConfig) TableName() string {
	return "userconfig"
}

func (SiteIcon) TableName() string {
	return "siteicons"
}

func (SiteUserData) TableName() string {
	return "siteuserdatas"
}

func (SiteStatistic) TableName() string {
	return "sitestatistics"
}

func (SubscribeHistory) TableName() string {
	return "subscribehistories"
}

func (Workflow) TableName() string {
	return "workflows"
}

func (TransferHistory) TableName() string {
	return "transferhistories"
}
