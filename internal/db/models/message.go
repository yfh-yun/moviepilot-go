package models

import (
	"gorm.io/gorm"
	
	"moviepilot-go/pkg/models"
)

// Message 消息表模型
type Message struct {
	models.Message
}

// ListByPage 分页获取消息列表
func (m *Message) ListByPage(db *gorm.DB, page int, count int) ([]models.Message, error) {
	var messages []models.Message
	offset := (page - 1) * count
	err := db.Order("reg_time DESC").Offset(offset).Limit(count).Find(&messages).Error
	return messages, err
}