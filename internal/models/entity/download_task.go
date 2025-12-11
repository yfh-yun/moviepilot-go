package entity

import (
	"time"
)

// DownloadTask 下载任务实体
type DownloadTask struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	SubscriptionID *uint      `gorm:"index" json:"subscription_id,omitempty"`
	DownloaderType string     `gorm:"size:20;not null;index" json:"downloader_type"` // qbittorrent, transmission
	Hash           string     `gorm:"size:64;uniqueIndex;not null" json:"hash"`
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
	UpdatedAt      time.Time  `gorm:"not null" json:"updated_at"`
}

// TableName 指定表名
func (DownloadTask) TableName() string {
	return "download_tasks"
}

// DownloadHistory 下载历史实体
type DownloadHistory struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	TaskID         *uint      `gorm:"index" json:"task_id,omitempty"`
	Hash           string     `gorm:"size:64;not null;index" json:"hash"`
	Name           string     `gorm:"size:200;not null" json:"name"`
	Size           int64      `json:"size,omitempty"`
	DownloaderType string     `gorm:"size:20;not null" json:"downloader_type"`
	Status         string     `gorm:"size:20;not null;index" json:"status"`           // completed, failed, deleted
	DownloadTime   int        `json:"download_time,omitempty"`                        // 下载耗时（秒）
	AverageSpeed   int64      `json:"average_speed,omitempty"`                        // 平均速度（字节/秒）
	FinalRatio     float64    `gorm:"type:decimal(5,2)" json:"final_ratio,omitempty"` // 最终分享率
	SavePath       string     `gorm:"size:500" json:"save_path,omitempty"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;index" json:"created_at"`
}

// TableName 指定表名
func (DownloadHistory) TableName() string {
	return "download_history"
}
