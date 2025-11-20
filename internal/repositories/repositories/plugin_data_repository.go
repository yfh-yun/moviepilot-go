package repositories

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"gorm.io/gorm"
)

// pluginDataRepository 插件数据仓储实现
type pluginDataRepository struct {
	db *gorm.DB
}

// NewPluginDataRepository 创建插件数据仓储
func NewPluginDataRepository(db *gorm.DB) interfaces.PluginDataRepository {
	return &pluginDataRepository{db: db}
}

// Create 创建插件数据
func (r *pluginDataRepository) Create(pluginData *model.PluginData) error {
	if pluginData == nil {
		return errors.New("plugin data cannot be nil")
	}
	return r.db.Create(pluginData).Error
}

// GetByID 根据ID获取插件数据
func (r *pluginDataRepository) GetByID(id uint) (*model.PluginData, error) {
	var pluginData model.PluginData
	err := r.db.First(&pluginData, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pluginData, nil
}

// GetByPluginKey 根据插件Key获取数据
func (r *pluginDataRepository) GetByPluginKey(pluginKey string) ([]*model.PluginData, error) {
	var pluginDataList []*model.PluginData
	err := r.db.Where("plugin_key = ?", pluginKey).Find(&pluginDataList).Error
	return pluginDataList, err
}

// GetByKey 根据数据Key获取数据
func (r *pluginDataRepository) GetByKey(key string) ([]*model.PluginData, error) {
	var pluginDataList []*model.PluginData
	err := r.db.Where("key = ?", key).Find(&pluginDataList).Error
	return pluginDataList, err
}

// GetByPluginKeyAndDataKey 根据插件Key和数据Key获取数据
func (r *pluginDataRepository) GetByPluginKeyAndDataKey(pluginKey, dataKey string) (*model.PluginData, error) {
	var pluginData model.PluginData
	err := r.db.Where("plugin_key = ? AND key = ?", pluginKey, dataKey).First(&pluginData).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pluginData, nil
}

// GetByUserID 根据用户ID获取插件数据
func (r *pluginDataRepository) GetByUserID(userID string) ([]*model.PluginData, error) {
	var pluginDataList []*model.PluginData
	err := r.db.Where("user_id = ?", userID).Find(&pluginDataList).Error
	return pluginDataList, err
}

// Update 更新插件数据
func (r *pluginDataRepository) Update(pluginData *model.PluginData) error {
	if pluginData == nil {
		return errors.New("plugin data cannot be nil")
	}
	return r.db.Save(pluginData).Error
}

// UpdateByPluginKeyAndDataKey 根据插件Key和数据Key更新数据
func (r *pluginDataRepository) UpdateByPluginKeyAndDataKey(pluginKey, dataKey, dataValue string) error {
	result := r.db.Model(&model.PluginData{}).
		Where("plugin_key = ? AND key = ?", pluginKey, dataKey).
		Update("value", dataValue)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update plugin data: %w", result.Error)
	}
	
	// 如果没有更新任何记录，创建新记录
	if result.RowsAffected == 0 {
		pluginData := &model.PluginData{
			PluginKey: pluginKey,
			Key:       dataKey,
			Value:     dataValue,
		}
		return r.Create(pluginData)
	}
	
	return nil
}

// Delete 删除插件数据
func (r *pluginDataRepository) Delete(id uint) error {
	return r.db.Delete(&model.PluginData{}, id).Error
}

// DeleteByPluginKey 根据插件Key删除数据
func (r *pluginDataRepository) DeleteByPluginKey(pluginKey string) error {
	return r.db.Where("plugin_key = ?", pluginKey).Delete(&model.PluginData{}).Error
}

// DeleteByUserID 根据用户ID删除数据
func (r *pluginDataRepository) DeleteByUserID(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.PluginData{}).Error
}

// ClearData 清空插件数据
func (r *pluginDataRepository) ClearData(pluginKey string) error {
	return r.db.Where("plugin_key = ?", pluginKey).Delete(&model.PluginData{}).Error
}

// List 分页获取插件数据列表
func (r *pluginDataRepository) List(offset, limit int) ([]*model.PluginData, int64, error) {
	var pluginDataList []*model.PluginData
	var total int64

	err := r.db.Model(&model.PluginData{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&pluginDataList).Error
	return pluginDataList, total, err
}

// Count 统计插件数据数量
func (r *pluginDataRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.PluginData{}).Count(&count).Error
	return count, err
}

// CountByPluginKey 根据插件Key统计数量
func (r *pluginDataRepository) CountByPluginKey(pluginKey string) (int64, error) {
	var count int64
	err := r.db.Model(&model.PluginData{}).Where("plugin_key = ?", pluginKey).Count(&count).Error
	return count, err
}