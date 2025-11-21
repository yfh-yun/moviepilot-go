// Package actions 提供订阅相关的数据类型定义
package actions

import (
	"time"
)

// SubscribeType 订阅类型枚举
type SubscribeType string

const (
	SubscribeTypeRSS    SubscribeType = "rss"     // RSS订阅
	SubscribeTypeTorrent SubscribeType = "torrent" // 种子订阅
	SubscribeTypeMedia   SubscribeType = "media"   // 媒体订阅
	SubscribeTypeKeyword SubscribeType = "keyword" // 关键词订阅
	SubscribeTypeUser    SubscribeType = "user"    // 用户订阅
)

// SubscribeStatus 订阅状态枚举
type SubscribeStatus string

const (
	SubscribeStatusActive   SubscribeStatus = "active"   // 活跃状态
	SubscribeStatusPaused   SubscribeStatus = "paused"   // 暂停状态
	SubscribeStatusStopped  SubscribeStatus = "stopped"  // 停止状态
	SubscribeStatusError    SubscribeStatus = "error"    // 错误状态
	SubscribeStatusDisabled SubscribeStatus = "disabled" // 禁用状态
)

// SubscribeConfig 订阅配置结构
type SubscribeConfig struct {
	// 基础配置
	Type         SubscribeType   `json:"type"`         // 订阅类型
	URL          string          `json:"url,omitempty"`          // 订阅源URL (RSS订阅)
	Keywords     []string        `json:"keywords,omitempty"`     // 关键词列表 (关键词订阅)
	UserID       string          `json:"user_id,omitempty"`      // 用户ID (用户订阅)
	MediaID      string          `json:"media_id,omitempty"`     // 媒体ID (媒体订阅)
	TorrentHash  string          `json:"torrent_hash,omitempty"` // 种子哈希 (种子订阅)

	// 过滤配置
	Categories   []string        `json:"categories,omitempty"`   // 分类过滤
	ExcludeWords []string        `json:"exclude_words,omitempty"` // 排除关键词
	MinSize      int64           `json:"min_size,omitempty"`      // 最小大小 (字节)
	MaxSize      int64           `json:"max_size,omitempty"`      // 最大大小 (字节)
	Quality      string          `json:"quality,omitempty"`       // 质量要求
	Resolution   string          `json:"resolution,omitempty"`    // 分辨率要求

	// 时间配置
	UpdateInterval int           `json:"update_interval"`       // 更新间隔 (分钟)
	LastUpdate     *time.Time    `json:"last_update,omitempty"` // 最后更新时间
	NextUpdate     *time.Time    `json:"next_update,omitempty"` // 下次更新时间
	CreatedAt      time.Time     `json:"created_at"`            // 创建时间
	ExpireAt       *time.Time    `json:"expire_at,omitempty"`   // 过期时间

	// 行为配置
	AutoDownload  bool           `json:"auto_download"`      // 是否自动下载
	SavePath      string         `json:"save_path,omitempty"` // 保存路径
	Downloader    string         `json:"downloader,omitempty"` // 下载器
	Labels        []string       `json:"labels,omitempty"`    // 标签
	Notify        bool           `json:"notify"`              // 是否通知

	// 状态配置
	Status        SubscribeStatus `json:"status"`             // 订阅状态
	ErrorCount    int             `json:"error_count"`        // 错误次数
	LastError     string          `json:"last_error,omitempty"` // 最后错误信息

	// 自定义配置
	Extra         map[string]interface{} `json:"extra,omitempty"` // 额外配置
}

// SubscribeResult 订阅操作结果
type SubscribeResult struct {
	SubscribeID   string          `json:"subscribe_id"`    // 订阅ID
	Success       bool            `json:"success"`         // 是否成功
	Message       string          `json:"message"`         // 操作消息
	Error         string          `json:"error,omitempty"` // 错误信息
	SubscribeConfig *SubscribeConfig `json:"config,omitempty"` // 订阅配置
	CreatedAt     time.Time       `json:"created_at"`      // 创建时间
}

// SubscribeInfo 订阅信息
type SubscribeInfo struct {
	ID           string          `json:"id"`            // 订阅ID
	Name         string          `json:"name"`          // 订阅名称
	Type         SubscribeType   `json:"type"`          // 订阅类型
	Status       SubscribeStatus `json:"status"`        // 订阅状态
	Description  string          `json:"description"`   // 订阅描述
	Config       *SubscribeConfig `json:"config,omitempty"` // 订阅配置
	Stats        *SubscribeStats `json:"stats,omitempty"` // 订阅统计
	CreatedAt    time.Time       `json:"created_at"`     // 创建时间
	UpdatedAt    time.Time       `json:"updated_at"`     // 更新时间
}

// SubscribeStats 订阅统计信息
type SubscribeStats struct {
	TotalItems    int       `json:"total_items"`     // 总订阅项数
	DownloadedItems int     `json:"downloaded_items"` // 已下载项数
	FailedItems   int       `json:"failed_items"`    // 失败项数
	LastSuccess   time.Time `json:"last_success"`    // 最后成功时间
	LastUpdate    time.Time `json:"last_update"`     // 最后更新时间
	AverageSize   int64     `json:"average_size"`    // 平均大小
}

// AddSubscribeParams 添加订阅参数
type AddSubscribeParams struct {
	Name          string          `json:"name" binding:"required"` // 订阅名称
	Description   string          `json:"description"`               // 订阅描述
	Config        SubscribeConfig `json:"config" binding:"required"` // 订阅配置
	Tags          []string        `json:"tags,omitempty"`            // 订阅标签
}

// UpdateSubscribeParams 更新订阅参数
type UpdateSubscribeParams struct {
	Name          string          `json:"name,omitempty"`           // 订阅名称
	Description   string          `json:"description,omitempty"`    // 订阅描述
	Config        SubscribeConfig `json:"config,omitempty"`         // 订阅配置
	Tags          []string        `json:"tags,omitempty"`           // 订阅标签
	Status        *SubscribeStatus `json:"status,omitempty"`        // 订阅状态
}

// SubscribeFilter 订阅过滤条件
type SubscribeFilter struct {
	Types    []SubscribeType   `json:"types,omitempty"`    // 按类型过滤
	Statuses []SubscribeStatus `json:"statuses,omitempty"` // 按状态过滤
	Tags     []string          `json:"tags,omitempty"`     // 按标签过滤
	Keywords []string          `json:"keywords,omitempty"` // 名称关键词
	Limit    int               `json:"limit,omitempty"`    // 限制数量
	Offset   int               `json:"offset,omitempty"`   // 偏移量
}

// SubscribeItem 订阅项
type SubscribeItem struct {
	ID            string                 `json:"id"`             // 项目ID
	SubscribeID   string                 `json:"subscribe_id"`   // 所属订阅ID
	Title         string                 `json:"title"`          // 标题
	Description   string                 `json:"description"`    // 描述
	URL           string                 `json:"url"`            // URL
	TorrentURL    string                 `json:"torrent_url,omitempty"` // 种子URL
	Magnet        string                 `json:"magnet,omitempty"` // 磁力链接
	Hash          string                 `json:"hash,omitempty"` // 哈希值
	Size          int64                  `json:"size,omitempty"` // 大小
	Categories    []string               `json:"categories,omitempty"` // 分类
	PublishDate   time.Time              `json:"publish_date"`   // 发布日期
	Downloaded    bool                   `json:"downloaded"`     // 是否已下载
	DownloadID    string                 `json:"download_id,omitempty"` // 下载ID
	Metadata      map[string]interface{} `json:"metadata,omitempty"` // 元数据
	CreatedAt     time.Time              `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time              `json:"updated_at"`     // 更新时间
}

// SubscribeItemFilter 订阅项过滤条件
type SubscribeItemFilter struct {
	SubscribeID  string    `json:"subscribe_id,omitempty"` // 按订阅ID过滤
	Downloaded   *bool     `json:"downloaded,omitempty"`   // 是否已下载
	Categories   []string  `json:"categories,omitempty"`   // 按分类过滤
	Keywords     []string  `json:"keywords,omitempty"`     // 标题关键词
	StartDate    time.Time `json:"start_date,omitempty"`   // 开始日期
	EndDate      time.Time `json:"end_date,omitempty"`     // 结束日期
	MinSize      int64     `json:"min_size,omitempty"`     // 最小大小
	MaxSize      int64     `json:"max_size,omitempty"`     // 最大大小
	Limit        int       `json:"limit,omitempty"`        // 限制数量
	Offset       int       `json:"offset,omitempty"`       // 偏移量
	OrderBy      string    `json:"order_by,omitempty"`     // 排序字段
	OrderDir     string    `json:"order_dir,omitempty"`    // 排序方向
}
