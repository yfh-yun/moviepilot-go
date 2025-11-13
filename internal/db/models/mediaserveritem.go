package models

import (
	"gorm.io/gorm"
	
	"moviepilot-go/pkg/models"
)

// MediaServerItem 媒体服务器媒体条目表模型
type MediaServerItem struct {
	models.MediaServerItem
}

// GetByItemID 根据item_id获取媒体服务器条目
func (m *MediaServerItem) GetByItemID(db *gorm.DB, itemID string) (*MediaServerItem, error) {
	var item MediaServerItem
	err := db.Where("item_id = ?", itemID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Empty 清空媒体服务器数据
func (m *MediaServerItem) Empty(db *gorm.DB, server *string) error {
	if server == nil {
		return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&MediaServerItem{}).Error
	} else {
		return db.Where("server = ?", *server).Delete(&MediaServerItem{}).Error
	}
}

// ExistByTmdbID 根据tmdbid和类型判断媒体服务器数据是否存在
func (m *MediaServerItem) ExistByTmdbID(db *gorm.DB, tmdbid int, mtype string) (*MediaServerItem, error) {
	var item MediaServerItem
	err := db.Where("tmdbid = ? AND item_type = ?", tmdbid, mtype).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ExistsByTitle 根据标题、类型、年份判断媒体服务器数据是否存在
func (m *MediaServerItem) ExistsByTitle(db *gorm.DB, title string, mtype string, year string) (*MediaServerItem, error) {
	var item MediaServerItem
	err := db.Where("title = ? AND item_type = ? AND year = ?", title, mtype, year).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}