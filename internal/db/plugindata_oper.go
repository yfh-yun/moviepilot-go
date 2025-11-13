package db

import (
	"moviepilot-go/internal/db/models"
	"moviepilot-go/pkg/models"
	
	"gorm.io/gorm"
)

// PluginDataOper 插件数据管理
type PluginDataOper struct {
	DB *gorm.DB
}

// NewPluginDataOper 创建插件数据管理实例
func NewPluginDataOper(db *gorm.DB) *PluginDataOper {
	return &PluginDataOper{
		DB: db,
	}
}

// Save 保存插件数据
func (p *PluginDataOper) Save(pluginID string, key string, value interface{}) error {
	pluginDataModel := &models.PluginData{}
	pluginData, err := pluginDataModel.GetPluginDataByKey(p.DB, pluginID, key)
	
	if err != nil && err != gorm.ErrRecordNotFound {
		// 如果是其他错误，返回错误
		return err
	}
	
	if pluginData != nil {
		// 如果数据存在，更新数据
		return p.DB.Model(&pluginData).Update("value", value).Error
	} else {
		// 如果数据不存在，创建新数据
		newPluginData := &models.PluginData{
			PluginID: pluginID,
			Key:      key,
			Value:    value,
		}
		return p.DB.Create(newPluginData).Error
	}
}

// GetData 获取插件数据
func (p *PluginDataOper) GetData(pluginID string, key string) (interface{}, error) {
	if key != "" {
		// 如果指定了key，获取指定key的数据
		pluginDataModel := &models.PluginData{}
		pluginData, err := pluginDataModel.GetPluginDataByKey(p.DB, pluginID, key)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, nil
			}
			return nil, err
		}
		return pluginData.Value, nil
	} else {
		// 如果没有指定key，获取插件所有数据
		return p.GetDataAll(pluginID)
	}
}

// DelData 删除插件数据
func (p *PluginDataOper) DelData(pluginID string, key string) error {
	if key != "" {
		// 如果指定了key，删除指定key的数据
		pluginDataModel := &models.PluginData{}
		return pluginDataModel.DelPluginDataByKey(p.DB, pluginID, key)
	} else {
		// 如果没有指定key，删除插件所有数据
		pluginDataModel := &models.PluginData{}
		return pluginDataModel.DelPluginData(p.DB, pluginID)
	}
}

// Truncate 清空插件数据
func (p *PluginDataOper) Truncate() error {
	pluginDataModel := &models.PluginData{}
	return pluginDataModel.Truncate(p.DB)
}

// GetDataAll 获取插件所有数据
func (p *PluginDataOper) GetDataAll(pluginID string) ([]models.PluginData, error) {
	pluginDataModel := &models.PluginData{}
	return pluginDataModel.GetPluginDataByPluginID(p.DB, pluginID)
}