package models

import (
	"time"
)

// AuthLog 认证日志模型
type AuthLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       *uint     `gorm:"index" json:"user_id,omitempty"`
	Action       string    `gorm:"size:50;not null" json:"action"` // login, logout, register, password_reset
	IPAddress    string    `gorm:"size:50" json:"ip_address"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	Status       string    `gorm:"size:20;not null" json:"status"` // success, failed
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (AuthLog) TableName() string {
	return "auth_logs"
}

// IsSuccess 检查操作是否成功
func (a *AuthLog) IsSuccess() bool {
	return a.Status == "success"
}

// IsFailed 检查操作是否失败
func (a *AuthLog) IsFailed() bool {
	return a.Status == "failed"
}
