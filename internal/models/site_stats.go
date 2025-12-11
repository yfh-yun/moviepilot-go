package models

import (
	"time"
)

// SiteStats 流量统计模型
type SiteStats struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SiteID        uint      `gorm:"not null;uniqueIndex:idx_site_date" json:"site_id"`
	Date          time.Time `gorm:"type:date;not null;uniqueIndex:idx_site_date" json:"date"`
	UploadDelta   int64     `gorm:"default:0" json:"upload_delta"`
	DownloadDelta int64     `gorm:"default:0" json:"download_delta"`
	UploadTotal   int64     `gorm:"default:0" json:"upload_total"`
	DownloadTotal int64     `gorm:"default:0" json:"download_total"`
	Ratio         float64   `gorm:"type:decimal(10,2);default:0" json:"ratio"`
	CreatedAt     time.Time `json:"created_at"`

	// 关联
	Site Site `gorm:"foreignKey:SiteID" json:"site,omitempty"`
}

// TableName 指定表名
func (SiteStats) TableName() string {
	return "site_stats"
}
