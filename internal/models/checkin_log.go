package models

import (
	"time"
)

// CheckinLog 签到日志
type CheckinLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SiteID       uint      `gorm:"not null;index" json:"site_id"`
	Success      bool      `gorm:"not null" json:"success"`
	Message      string    `gorm:"type:text" json:"message,omitempty"`
	Bonus        int       `gorm:"default:0" json:"bonus"`
	ContinueDays int       `gorm:"default:0" json:"continue_days"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	CheckinTime  time.Time `gorm:"default:CURRENT_TIMESTAMP;not null" json:"checkin_time"`

	// 关联
	Site Site `gorm:"foreignKey:SiteID" json:"site,omitempty"`
}

// TableName 指定表名
func (CheckinLog) TableName() string {
	return "checkin_logs"
}
