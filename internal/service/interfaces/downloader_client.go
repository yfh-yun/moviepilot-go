// Package interfaces 定义下载器客户端接口
package interfaces

import (
	"context"
	"time"
)

// DownloaderClient 下载器客户端接口
type DownloaderClient interface {
	// 基础操作
	Start() error
	Stop() error
	GetStatus() (*DownloaderStatus, error)
	
	// 下载管理
	AddTorrent(ctx context.Context, req *AddTorrentRequest) (*AddTorrentResponse, error)
	RemoveTorrent(ctx context.Context, hash string) error
	PauseTorrent(ctx context.Context, hash string) error
	ResumeTorrent(ctx context.Context, hash string) error
	
	// 信息查询
	GetTorrentInfo(ctx context.Context, hash string) (*TorrentInfo, error)
	ListTorrents(ctx context.Context, filter *TorrentFilter) ([]*TorrentInfo, error)
	GetGlobalTransferInfo(ctx context.Context) (*TransferInfo, error)
	
	// 配置管理
	GetPreferences() (*Preferences, error)
	SetPreferences(prefs *Preferences) error
	
	// 下载策略
	SetTorrentPriority(hash string, priority int) error
	SetTorrentSpeedLimits(hash string, downloadLimit, uploadLimit int) error
	
	// 版本和兼容性
	GetVersion() (string, error)
	GetAPIVersion() (string, error)
}

// DownloaderType 下载器类型
type DownloaderType string

const (
	DownloaderTypeQBittorrent DownloaderType = "qbittorrent"
	DownloaderTypeTransmission DownloaderType = "transmission"
	DownloaderTypeAria2        DownloaderType = "aria2"
)

// DownloaderConfig 下载器配置
type DownloaderConfig struct {
	Type     DownloaderType `json:"type" yaml:"type"`
	Endpoint string         `json:"endpoint" yaml:"endpoint"`
	Username string         `json:"username" yaml:"username"`
	Password string         `json:"password" yaml:"password"`
	Timeout  time.Duration  `json:"timeout" yaml:"timeout"`
	
	// 连接配置
	MaxConnections int           `json:"max_connections" yaml:"max_connections"`
	RetryInterval  time.Duration `json:"retry_interval" yaml:"retry_interval"`
	MaxRetries     int           `json:"max_retries" yaml:"max_retries"`
	
	// 安全配置
	SkipTLSVerify bool `json:"skip_tls_verify" yaml:"skip_tls_verify"`
	CustomCA      string `json:"custom_ca" yaml:"custom_ca"`
}

// AddTorrentRequest 添加种子请求
type AddTorrentRequest struct {
	// 种子数据（二选一）
	URL        string `json:"url,omitempty"`        // 种子URL或磁力链接
	RawData    []byte `json:"raw_data,omitempty"`   // 种子文件数据
	
	// 下载配置
	SavePath       string            `json:"save_path,omitempty"`        // 保存路径
	DownloadPath   string            `json:"download_path,omitempty"`    // 下载目录
	Category       string            `json:"category,omitempty"`         // 分类
	Tags           []string          `json:"tags,omitempty"`             // 标签
	Priority       int               `json:"priority,omitempty"`         // 优先级
	Sequential     bool              `json:"sequential,omitempty"`        // 按顺序下载
	FirstLast      bool              `json:"first_last_piece,omitempty"`  // 首尾优先
	
	// 文件选择（可选）
	FileIDs        []int             `json:"file_ids,omitempty"`         // 下载文件ID列表
	RenameFiles    map[string]string `json:"rename_files,omitempty"`     // 文件重命名映射
	
	// 速度限制
	DownloadLimit  int64 `json:"download_limit,omitempty"`  // 下载速度限制 (bytes/s)
	UploadLimit    int64 `json:"upload_limit,omitempty"`      // 上传速度限制 (bytes/s)
	
	// 其他选项
	Paused         bool     `json:"paused,omitempty"`          // 添加后暂停
	SkipHashing    bool     `json:"skip_hashing,omitempty"`    // 跳过哈希检查
	SkipChecking   bool     `json:"skip_checking,omitempty"`   // 跳过完整性检查
	
	// 元数据
	AutoTMM        bool              `json:"auto_tmm,omitempty"`     // 自动 Torrent Management Mode
	RatioLimit     float64           `json:"ratio_limit,omitempty"`   // 分享率限制
	SeedingTime    time.Duration     `json:"seeding_time,omitempty"`  // 做种时间限制
	
	// 扩展数据
	Metadata       map[string]string `json:"metadata,omitempty"`       // 自定义元数据
}

// AddTorrentResponse 添加种子响应
type AddTorrentResponse struct {
	Success   bool   `json:"success"`
	Hash      string `json:"hash,omitempty"`
	Message   string `json:"message,omitempty"`
	TorrentID string `json:"torrent_id,omitempty"`
}

// DownloaderStatus 下载器状态
type DownloaderStatus struct {
	Connected     bool      `json:"connected"`
	Version       string    `json:"version"`
	APIVersion    string    `json:"api_version"`
	Uptime        time.Duration `json:"uptime"`
	FreeSpace     int64     `json:"free_space"`
	
	// 传输统计
	AllTimeDL     int64 `json:"alltime_download"`
	AllTimeUL     int64 `json:"alltime_upload"`
	GlobalDLSpeed int64 `json:"global_download_speed"`
	GlobalULSpeed int64 `json:"global_upload_speed"`
	
	// 统计信息
	TorrentCount   int `json:"torrent_count"`
	ActiveCount    int `json:"active_count"`
	PausedCount    int `json:"paused_count"`
	CompletedCount int `json:"completed_count"`
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	// 基本信息
	Hash           string    `json:"hash"`
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	Progress       float64   `json:"progress"`
	State          string    `json:"state"`
	Priority       int       `json:"priority"`
	
	// 状态信息
	DownloadSpeed  int64     `json:"download_speed"`
	UploadSpeed    int64     `json:"upload_speed"`
	Downloaded     int64     `json:"downloaded"`
	Uploaded       int64     `json:"uploaded"`
	ETA            int64     `json:"eta"`            // 预计完成时间 (秒)
	
	// 时间信息
	AddedOn        time.Time `json:"added_on"`
	CompletedOn    *time.Time `json:"completed_on,omitempty"`
	SeedingTime    int64     `json:"seeding_time"`   // 做种时间 (秒)
	
	// 比率限制
	Ratio          float64   `json:"ratio"`
	RatioLimit     float64   `json:"ratio_limit"`
	
	// 路径信息
	SavePath       string    `json:"save_path"`
	DownloadPath   string    `json:"download_path,omitempty"`
	ContentPath    string    `json:"content_path,omitempty"`
	
	// 分类和标签
	Category       string    `json:"category"`
	Tags           []string  `json:"tags"`
	
	// 文件信息
	Files          []TorrentFile `json:"files,omitempty"`
	Trackers       []string       `json:"trackers,omitempty"`
	
	// 队列信息
	QueuePosition  int    `json:"queue_position,omitempty"`
	
	// 元数据
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// TorrentFile 种子文件信息
type TorrentFile struct {
	Index       int    `json:"index"`        // 文件索引
	Name        string `json:"name"`         // 文件名
	Size        int64  `json:"size"`         // 文件大小
	Progress    float64 `json:"progress"`    // 下载进度
	Priority    int    `json:"priority"`     // 下载优先级
	IsSeed      bool   `json:"is_seed"`      // 是否已完成
	Path        string `json:"path"`         // 文件路径
}

// TorrentFilter 种子过滤器
type TorrentFilter struct {
	// 状态过滤
	States      []string `json:"states,omitempty"`       // 状态列表
	Categories  []string `json:"categories,omitempty"`   // 分类列表
	Tags        []string `json:"tags,omitempty"`         // 标签列表
	
	// 范围过滤
	MinSize     int64 `json:"min_size,omitempty"`       // 最小大小
	MaxSize     int64 `json:"max_size,omitempty"`       // 最大大小
	MinProgress float64 `json:"min_progress,omitempty"`  // 最小进度
	MaxProgress float64 `json:"max_progress,omitempty"`  // 最大进度
	
	// 排序
	SortBy      string `json:"sort_by,omitempty"`       // 排序字段
	SortOrder   string `json:"sort_order,omitempty"`     // 排序方向 (asc/desc)
	
	// 分页
	Offset      int    `json:"offset,omitempty"`        // 偏移量
	Limit       int    `json:"limit,omitempty"`         // 限制数量
}

// TransferInfo 传输信息
type TransferInfo struct {
	// 全局速度
	GlobalDownloadSpeed int64 `json:"global_download_speed"`
	GlobalUploadSpeed    int64 `json:"global_upload_speed"`
	
	// 全统计
	TotalDownloaded  int64 `json:"total_downloaded"`
	TotalUploaded    int64 `json:"total_uploaded"`
	
	// 连接信息
	DHTNodes         int    `json:"dht_nodes"`
	ListeningPort    int    `json:"listening_port"`
	
	// 时间统计
	TotalConnectedTime int64  `json:"total_connected_time"`  // 总连接时间 (秒)
	TotalUploadTime    int64  `json:"total_upload_time"`     // 总上传时间 (秒)
	TotalDownloadTime   int64  `json:"total_download_time"`   // 总下载时间 (秒)
}

// Preferences 偏好设置
type Preferences struct {
	// 下载设置
	DownloadPath          string  `json:"download_path"`
	TempPathEnabled       bool    `json:"temp_path_enabled"`
	TempPath              string  `json:"temp_path,omitempty"`
	MaxActiveDownloads    int     `json:"max_active_downloads"`
	MaxActiveTorrents     int     `json:"max_active_torrents"`
	MaxActiveUploads      int     `json:"max_active_uploads"`
	
	// 速度限制
	DownloadLimitEnabled  bool    `json:"download_limit_enabled"`
	DownloadLimit         int64   `json:"download_limit"`
	UploadLimitEnabled    bool    `json:"upload_limit_enabled"`
	UploadLimit           int64   `json:"upload_limit"`
	
	// 种子管理
	MaxRatioEnabled       bool    `json:"max_ratio_enabled"`
	MaxRatio              float64 `json:"max_ratio"`
	MaxSeedingTimeEnabled bool    `json:"max_seeding_time_enabled"`
	MaxSeedingTime        int64   `json:"max_seeding_time"`
	
	// 其他设置
	AlternativeGlobalDHT  bool    `json:"alternative_global_dht"`
	AlternativeWebUI      bool    `json:"alternative_web_ui"`
	AnnounceToAllTrackers bool    `json:"announce_to_all_trackers"`
	AnnounceToAllTiers    bool    `json:"announce_to_all_tiers"`
	
	// UI设置
	Locale                 string  `json:"locale"`
	WebUIAddress           string  `json:"web_ui_address"`
	WebUIPort              int     `json:"web_ui_port"`
	WebUIUsername          string  `json:"web_ui_username"`
	WebUIPassword          string  `json:"web_ui_password"`
}

// DownloaderFactory 下载器工厂接口
type DownloaderFactory interface {
	// 创建下载器客户端
	CreateClient(config *DownloaderConfig) (DownloaderClient, error)
	
	// 获取支持的下载器类型
	GetSupportedTypes() []DownloaderType
	
	// 验证配置
	ValidateConfig(config *DownloaderConfig) error
}