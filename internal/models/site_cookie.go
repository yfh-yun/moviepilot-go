package models

import (
	"time"
)

// SiteCookie Cookie历史模型
type SiteCookie struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	SiteID    uint       `gorm:"not null;index" json:"site_id"`
	Cookie    string     `gorm:"type:text;not null" json:"cookie"`
	IsValid   bool       `gorm:"default:true;not null" json:"is_valid"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	// 关联
	Site Site `gorm:"foreignKey:SiteID" json:"site,omitempty"`
}

// TableName 指定表名
func (SiteCookie) TableName() string {
	return "site_cookies"
}
