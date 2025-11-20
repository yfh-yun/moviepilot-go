package repositories

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"gorm.io/gorm"
)

// systemConfigRepository 系统配置仓储实现
type systemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置仓储
func NewSystemConfigRepository(db *gorm.DB) interfaces.SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

// Create 创建系统配置
func (r *systemConfigRepository) Create(config *model.SystemConfig) error {
	if config == nil {
		return errors.New("system config cannot be nil")
	}
	return r.db.Create(config).Error
}

// GetByID 根据ID获取系统配置
func (r *systemConfigRepository) GetByID(id uint) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.First(&config, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetByKey 根据Key获取配置
func (r *systemConfigRepository) GetByKey(key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetByKeys 根据Keys批量获取配置
func (r *systemConfigRepository) GetByKeys(keys []string) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.Where("key IN ?", keys).Find(&configs).Error
	return configs, err
}

// GetAll 获取所有配置
func (r *systemConfigRepository) GetAll() ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.Find(&configs).Error
	return configs, err
}

// Update 更新系统配置
func (r *systemConfigRepository) Update(config *model.SystemConfig) error {
	if config == nil {
		return errors.New("system config cannot be nil")
	}
	return r.db.Save(config).Error
}

// UpdateByKey 根据Key更新配置值
func (r *systemConfigRepository) UpdateByKey(key, value string) error {
	result := r.db.Model(&model.SystemConfig{}).Where("key = ?", key).Update("value", value)
	if result.Error != nil {
		return fmt.Errorf("failed to update system config: %w", result.Error)
	}
	
	// 如果没有更新任何记录，创建新配置
	if result.RowsAffected == 0 {
		config := &model.SystemConfig{
			Key:   key,
			Value: value,
		}
		return r.Create(config)
	}
	
	return nil
}

// UpdateByKeys 批量更新配置
func (r *systemConfigRepository) UpdateByKeys(configs map[string]string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			result := tx.Model(&model.SystemConfig{}).Where("key = ?", key).Update("value", value)
			if result.Error != nil {
				return fmt.Errorf("failed to update config %s: %w", key, result.Error)
			}
			
			// 如果没有更新任何记录，创建新配置
			if result.RowsAffected == 0 {
				config := &model.SystemConfig{
					Key:   key,
					Value: value,
				}
				if err := tx.Create(config).Error; err != nil {
					return fmt.Errorf("failed to create config %s: %w", key, err)
				}
			}
		}
		return nil
	})
}

// Delete 删除系统配置
func (r *systemConfigRepository) Delete(id uint) error {
	return r.db.Delete(&model.SystemConfig{}, id).Error
}

// DeleteByKey 根据Key删除配置
func (r *systemConfigRepository) DeleteByKey(key string) error {
	return r.db.Where("key = ?", key).Delete(&model.SystemConfig{}).Error
}

// List 分页获取系统配置列表
func (r *systemConfigRepository) List(offset, limit int) ([]*model.SystemConfig, int64, error) {
	var configs []*model.SystemConfig
	var total int64

	err := r.db.Model(&model.SystemConfig{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&configs).Error
	return configs, total, err
}

// Count 统计系统配置数量
func (r *systemConfigRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.SystemConfig{}).Count(&count).Error
	return count, err
}

// GetConfigsByType 根据类型获取配置
func (r *systemConfigRepository) GetConfigsByType(configType string) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.Where("type = ?", configType).Find(&configs).Error
	return configs, err
}

// SetDefaultValue 设置默认值
func (r *systemConfigRepository) SetDefaultValue(key, value, configType, remark string) error {
	// 检查配置是否已存在
	config, err := r.GetByKey(key)
	if err != nil {
		return fmt.Errorf("failed to check existing config: %w", err)
	}

	if config != nil {
		// 配置已存在，不更新
		return nil
	}

	// 创建新配置
	newConfig := &model.SystemConfig{
		Key:     key,
		Value:   value,
		Type:    configType,
		Remark:  remark,
	}
	
	return r.Create(newConfig)
}