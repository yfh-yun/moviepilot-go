package models

import (
	"gorm.io/gorm"
	
	"moviepilot-go/pkg/models"
)

// PluginData 插件数据表模型
type PluginData struct {
	models.PluginData
}

// GetPluginData 获取插件所有数据
func (p *PluginData) GetPluginData(db *gorm.DB, pluginID string) ([]models.PluginData, error) {
	var pluginDataList []models.PluginData
	err := db.Where("plugin_id = ?", pluginID).Find(&pluginDataList).Error
	return pluginDataList, err
}

// GetPluginDataByKey 根据key获取插件数据
func (p *PluginData) GetPluginDataByKey(db *gorm.DB, pluginID string, key string) (*models.PluginData, error) {
	var pluginData models.PluginData
	err := db.Where("plugin_id = ? AND key = ?", pluginID, key).First(&pluginData).Error
	if err != nil {
		return nil, err
	}
	return &pluginData, nil
}

// DelPluginDataByKey 根据key删除插件数据
func (p *PluginData) DelPluginDataByKey(db *gorm.DB, pluginID string, key string) error {
	return db.Where("plugin_id = ? AND key = ?", pluginID, key).Delete(&models.PluginData{}).Error
}

// DelPluginData 删除插件所有数据
func (p *PluginData) DelPluginData(db *gorm.DB, pluginID string) error {
	return db.Where("plugin_id = ?", pluginID).Delete(&models.PluginData{}).Error
}

// GetPluginDataByPluginID 根据插件ID获取插件所有数据
func (p *PluginData) GetPluginDataByPluginID(db *gorm.DB, pluginID string) ([]models.PluginData, error) {
	var pluginDataList []models.PluginData
	err := db.Where("plugin_id = ?", pluginID).Find(&pluginDataList).Error
	return pluginDataList, err
}

// Truncate 清空插件数据表
func (p *PluginData) Truncate(db *gorm.DB) error {
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.PluginData{}).Error
}