package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"` // 不在 JSON 中返回密码
	Nickname     string         `gorm:"size:100" json:"nickname"`
	Avatar       string         `gorm:"size:500" json:"avatar"`
	Status       string         `gorm:"size:20;default:active;not null" json:"status"` // active, disabled, locked
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"size:50" json:"last_login_ip"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // 软删除

	// 关联
	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Sites []Site `gorm:"foreignKey:UserID" json:"sites,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM 钩子：创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 可以在这里添加创建前的逻辑
	return nil
}

// BeforeUpdate GORM 钩子：更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 可以在这里添加更新前的逻辑
	return nil
}

// IsActive 检查用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// IsDisabled 检查用户是否被禁用
func (u *User) IsDisabled() bool {
	return u.Status == "disabled"
}

// IsLocked 检查用户是否被锁定
func (u *User) IsLocked() bool {
	return u.Status == "locked"
}
