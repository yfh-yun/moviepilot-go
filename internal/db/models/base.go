package models

import (
	"gorm.io/gorm"
)

// BaseModel 基础模型接口
type BaseModel interface {
	Create(db *gorm.DB) error
	Get(db *gorm.DB, id interface{}) error
	Update(db *gorm.DB, updates interface{}) error
	Delete(db *gorm.DB) error
	List(db *gorm.DB, dest interface{}) error
	ToMap() map[string]interface{}
}

// Base 基础模型结构
type Base struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
}

// Create 创建记录
func (b *Base) Create(db *gorm.DB) error {
	return db.Create(b).Error
}

// Get 根据ID获取记录
func (b *Base) Get(db *gorm.DB, id interface{}) error {
	return db.First(b, id).Error
}

// Update 更新记录
func (b *Base) Update(db *gorm.DB, updates interface{}) error {
	return db.Model(b).Updates(updates).Error
}

// Delete 删除记录
func (b *Base) Delete(db *gorm.DB) error {
	return db.Delete(b).Error
}

// Truncate 清空表
func (b *Base) Truncate(db *gorm.DB) error {
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(b).Error
}

// List 获取所有记录
func (b *Base) List(db *gorm.DB, dest interface{}) error {
	return db.Find(dest).Error
}

// ToMap 转换为map
func (b *Base) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id": b.ID,
	}
}
