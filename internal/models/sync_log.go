package models

import (
	"time"
)

// SyncLog 同步日志模型
type SyncLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SiteID       uint      `gorm:"not null;index" json:"site_id"`
	Success      bool      `gorm:"not null" json:"success"`
	DurationMs   int       `json:"duration_ms,omitempty"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	SyncedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP;not null" json:"synced_at"`

	// 关联
	Site Site `gorm:"foreignKey:SiteID" json:"site,omitempty"`
}

// TableName 指定表名
func (SyncLog) TableName() string {
	return "sync_logs"
}
