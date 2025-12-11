package entity

import (
	"time"
)

// SiteRSSFeed RSS订阅表
type SiteRSSFeed struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SiteID      uint      `gorm:"index;not null" json:"site_id"`
	SiteName    string    `gorm:"type:varchar(200)" json:"site_name"`
	URL         string    `gorm:"type:varchar(500);not null" json:"url"`
	Title       string    `gorm:"type:varchar(200)" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Enabled     bool      `gorm:"default:true;index" json:"enabled"`
	Interval    int       `gorm:"default:300" json:"interval"` // 更新间隔（秒）
	LastUpdate  time.Time `json:"last_update"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SiteRSSFeed) TableName() string {
	return "site_rss_feeds"
}
