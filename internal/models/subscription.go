// Package models MoviePilot数据模型定义
package models

import (
	"time"

	"gorm.io/gorm"
)

// SubscriptionStatus 订阅状态枚举
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"    // 活跃
	SubscriptionStatusPaused    SubscriptionStatus = "paused"    // 暂停
	SubscriptionStatusCompleted SubscriptionStatus = "completed" // 已完成
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled" // 已取消
)

// Subscription 订阅模型
type Subscription struct {
	ID            uint               `gorm:"primaryKey" json:"id"`
	UserID        uint               `gorm:"not null;index" json:"user_id"`
	MediaID       *uint              `gorm:"index" json:"media_id,omitempty"`
	MediaType     MediaType          `gorm:"size:20;not null;index" json:"media_type"` // movie, tv, anime
	TMDBID        int                `gorm:"not null;index" json:"tmdb_id"`
	Title         string             `gorm:"size:200;not null" json:"title"`
	Year          int                `gorm:"index" json:"year,omitempty"`
	Season        int                `json:"season,omitempty"`
	Episode       int                `json:"episode,omitempty"`
	Status        SubscriptionStatus `gorm:"size:20;default:'active';not null;index" json:"status"`
	Priority      int                `gorm:"default:5;not null" json:"priority"`
	Monitor       bool               `gorm:"default:true;not null" json:"monitor"` // 是否监控
	AutoDownload  bool               `gorm:"default:true;not null" json:"auto_download"` // 是否自动下载
	AutoDelete    bool               `gorm:"default:false;not null" json:"auto_delete"` // 是否自动删除
	Quality        string            `gorm:"size:50" json:"quality,omitempty"` // 质量要求
	Language       string            `gorm:"size:50" json:"language,omitempty"` // 语言要求
	Subtitle       bool              `gorm:"default:false;not null" json:"subtitle"` // 是否需要字幕
	Tags           []string          `gorm:"type:text[]" json:"tags,omitempty"` // 标签
	Note           string            `gorm:"type:text" json:"note,omitempty"` // 备注
	LastCheckAt    *time.Time        `json:"last_check_at,omitempty"` // 上次检查时间
	NextCheckAt    *time.Time        `json:"next_check_at,omitempty"` // 下次检查时间
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"-"`

	// 关联
	User           User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Media          *MediaItem        `gorm:"foreignKey:MediaID" json:"media,omitempty"`
	Items          []SubscriptionItem `gorm:"foreignKey:SubscriptionID" json:"items,omitempty"`
	DownloadTasks  []DownloadTask     `gorm:"foreignKey:SubscriptionID" json:"download_tasks,omitempty"`
}

// TableName 指定表名
func (Subscription) TableName() string {
	return "subscriptions"
}

// SubscriptionItem 订阅项模型（用于电视剧的季、集）
type SubscriptionItem struct {
	ID              uint               `gorm:"primaryKey" json:"id"`
	SubscriptionID  uint               `gorm:"not null;index" json:"subscription_id"`
	ItemType        string             `gorm:"size:20;not null" json:"item_type"` // season, episode
	Season          int                `json:"season,omitempty"`
	Episode         int                `json:"episode,omitempty"`
	Status          string             `gorm:"size:20;default:'pending';not null;index" json:"status"` // pending, downloading, completed, skipped
	Quality         string             `gorm:"size:50" json:"quality,omitempty"` // 质量
	Language        string             `gorm:"size:50" json:"language,omitempty"` // 语言
	DownloadTaskID  *uint              `gorm:"index" json:"download_task_id,omitempty"` // 关联的下载任务
	LastCheckAt     *time.Time         `json:"last_check_at,omitempty"` // 上次检查时间
	CompletedAt     *time.Time         `json:"completed_at,omitempty"` // 完成时间
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	DeletedAt       gorm.DeletedAt     `gorm:"index" json:"-"`

	// 关联
	Subscription    Subscription       `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
	DownloadTask    *DownloadTask      `gorm:"foreignKey:DownloadTaskID" json:"download_task,omitempty"`
}

// TableName 指定表名
func (SubscriptionItem) TableName() string {
	return "subscription_items"
}

// SubscribeShare 订阅分享模型
type SubscribeShare struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	ShareID     string    `gorm:"size:100;uniqueIndex;not null" json:"share_id"` // 分享ID
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	MediaType   MediaType `gorm:"size:20;not null;index" json:"media_type"` // movie, tv, anime
	Data        string    `gorm:"type:jsonb;not null" json:"data"` // 分享的订阅数据
	Status      string    `gorm:"size:20;default:'active';not null;index" json:"status"` // active, expired, disabled
	ExpireAt    *time.Time `json:"expire_at,omitempty"` // 过期时间
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 关联
	User        User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (SubscribeShare) TableName() string {
	return "subscribe_shares"
}

// RSS RSS订阅模型
type RSS struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	URL           string    `gorm:"size:500;not null;uniqueIndex" json:"url"`
	Enabled       bool      `gorm:"default:true;not null;index" json:"enabled"`
	Priority      int       `gorm:"default:5;not null" json:"priority"`
	LastSyncAt    *time.Time `json:"last_sync_at,omitempty"` // 上次同步时间
	NextSyncAt    *time.Time `json:"next_sync_at,omitempty"` // 下次同步时间
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User          User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (RSS) TableName() string {
	return "rss"
}

// DownloadTask 下载任务模型
type DownloadTask struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	SubscriptionID *uint      `gorm:"index" json:"subscription_id,omitempty"`
	DownloaderType string     `gorm:"size:20;not null;index" json:"downloader_type"` // qbittorrent, transmission, aria2
	Hash           string     `gorm:"size:64;uniqueIndex;not null" json:"hash"` // 种子哈希
	Name           string     `gorm:"size:200;not null" json:"name"`
	SavePath       string     `gorm:"size:500" json:"save_path,omitempty"`
	Size           int64      `gorm:"default:0" json:"size"`
	Status         string     `gorm:"size:20;default:'downloading';not null;index" json:"status"` // downloading, completed, error, paused, seeding
	Progress       float64    `gorm:"type:decimal(5,2);default:0;not null" json:"progress"`
	DownloadSpeed  int64      `gorm:"default:0" json:"download_speed"`
	UploadSpeed    int64      `gorm:"default:0" json:"upload_speed"`
	Downloaded     int64      `gorm:"default:0" json:"downloaded"`
	Uploaded       int64      `gorm:"default:0" json:"uploaded"`
	Ratio          float64    `gorm:"type:decimal(5,2);default:0" json:"ratio"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt      time.Time  `gorm:"not null" json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// 关联
	User           User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Subscription   *Subscription     `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
}

// TableName 指定表名
func (DownloadTask) TableName() string {
	return "download_tasks"
}

// DownloadHistory 下载历史模型
type DownloadHistory struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	TaskID         *uint      `gorm:"index" json:"task_id,omitempty"`
	Hash           string     `gorm:"size:64;not null;index" json:"hash"`
	Name           string     `gorm:"size:200;not null" json:"name"`
	Size           int64      `json:"size,omitempty"`
	DownloaderType string     `gorm:"size:20;not null" json:"downloader_type"`
	Status         string     `gorm:"size:20;not null;index" json:"status"` // completed, failed, deleted
	DownloadTime   int        `json:"download_time,omitempty"` // 下载耗时（秒）
	AverageSpeed   int64      `json:"average_speed,omitempty"` // 平均速度（字节/秒）
	FinalRatio     float64    `gorm:"type:decimal(5,2)" json:"final_ratio,omitempty"` // 最终分享率
	SavePath       string     `gorm:"size:500" json:"save_path,omitempty"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;index" json:"created_at"`

	// 关联
	User           User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (DownloadHistory) TableName() string {
	return "download_history"
}
