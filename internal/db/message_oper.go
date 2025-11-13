package db

import (
	"time"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/db/models"
	"moviepilot-go/pkg/models"
)

// MessageOper 消息数据管理
type MessageOper struct {
	DB *gorm.DB
}

// NewMessageOper 创建消息数据管理实例
func NewMessageOper(db *gorm.DB) *MessageOper {
	return &MessageOper{
		DB: db,
	}
}

// Add 新增消息
func (m *MessageOper) Add(
	channel string,
	source string,
	mtype string,
	title string,
	text string,
	image string,
	link string,
	userid string,
	action int,
	note map[string]interface{},
	extra map[string]interface{},
) error {
	// 构造消息对象
	message := &models.Message{
		Channel:  channel,
		Source:   source,
		MType:    mtype,
		Title:    title,
		Text:     text,
		Image:    image,
		Link:     link,
		UserID:   userid,
		Action:   action,
		RegTime:  time.Now().Format("2006-01-02 15:04:05"),
		Note:     note,
	}
	
	// 从extra中提取Message中存在的字段
	if extra != nil {
		// 这里我们只处理已知的字段，忽略Message中不存在的字段
		// 在Go中，结构体字段是静态的，所以不需要像Python那样动态处理
	}
	
	// 创建消息记录
	return m.DB.Create(message).Error
}

// ListByPage 分页获取消息列表
func (m *MessageOper) ListByPage(page int, count int) ([]models.Message, error) {
	message := &models.Message{}
	return message.ListByPage(m.DB, page, count)
}