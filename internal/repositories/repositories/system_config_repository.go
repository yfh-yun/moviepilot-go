package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// SystemConfigRepositoryImpl 系统配置仓储实现
type SystemConfigRepositoryImpl struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置仓储实例
func NewSystemConfigRepository(db *gorm.DB) interfaces.SystemConfigRepository {
	return &SystemConfigRepositoryImpl{db: db}
}

// Get 根据key获取配置
func (r *SystemConfigRepositoryImpl) Get(ctx context.Context, key string) (*database.SystemConfig, error) {
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// Set 设置配置
func (r *SystemConfigRepositoryImpl) Set(ctx context.Context, key string, value string) error {
	// 先查询是否存在
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新记录
			config = database.SystemConfig{
				Key:   key,
				Value: value,
			}
			return r.db.WithContext(ctx).Create(&config).Error
		}
		return err
	}

	// 存在，更新记录
	config.Value = value
	return r.db.WithContext(ctx).Save(&config).Error
}

// Delete 删除配置
func (r *SystemConfigRepositoryImpl) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&database.SystemConfig{}).Error
}

// List 获取所有配置
func (r *SystemConfigRepositoryImpl) List(ctx context.Context) ([]*database.SystemConfig, error) {
	var configs []*database.SystemConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}

// BatchSet 批量设置配置
func (r *SystemConfigRepositoryImpl) BatchSet(ctx context.Context, configs map[string]string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			var config database.SystemConfig
			err := tx.Where("key = ?", key).First(&config).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 不存在，创建新记录
					config = database.SystemConfig{
						Key:   key,
						Value: value,
					}
					if err := tx.Create(&config).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}

			// 存在，更新记录
			config.Value = value
			if err := tx.Save(&config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchGet 批量获取配置
func (r *SystemConfigRepositoryImpl) BatchGet(ctx context.Context, keys []string) (map[string]string, error) {
	var configs []*database.SystemConfig
	err := r.db.WithContext(ctx).Where("key IN ?", keys).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(configs))
	for _, config := range configs {
		result[config.Key] = config.Value
	}
	return result, nil
}

// Exists 检查配置是否存在
func (r *SystemConfigRepositoryImpl) Exists(ctx context.Context, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.SystemConfig{}).Where("key = ?", key).Count(&count).Error
	return count > 0, err
}

// GetByPrefix 根据前缀获取配置
func (r *SystemConfigRepositoryImpl) GetByPrefix(ctx context.Context, prefix string) ([]*database.SystemConfig, error) {
	var configs []*database.SystemConfig
	err := r.db.WithContext(ctx).Where("key LIKE ?", prefix+"%").Find(&configs).Error
	return configs, err
}

// DeleteByPrefix 根据前缀删除配置
func (r *SystemConfigRepositoryImpl) DeleteByPrefix(ctx context.Context, prefix string) error {
	return r.db.WithContext(ctx).Where("key LIKE ?", prefix+"%").Delete(&database.SystemConfig{}).Error
}

// GetType 获取配置类型
func (r *SystemConfigRepositoryImpl) GetType(ctx context.Context, key string) (string, error) {
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.Type, nil
}

// SetWithType 设置配置和类型
func (r *SystemConfigRepositoryImpl) SetWithType(ctx context.Context, key string, value string, configType string) error {
	// 先查询是否存在
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新记录
			config = database.SystemConfig{
				Key:   key,
				Value: value,
				Type:  configType,
			}
			return r.db.WithContext(ctx).Create(&config).Error
		}
		return err
	}

	// 存在，更新记录
	config.Value = value
	config.Type = configType
	return r.db.WithContext(ctx).Save(&config).Error
}

// BatchSetWithTypes 批量设置配置和类型
func (r *SystemConfigRepositoryImpl) BatchSetWithTypes(ctx context.Context, configs map[string]interfaces.ConfigItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, item := range configs {
			var config database.SystemConfig
			err := tx.Where("key = ?", key).First(&config).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 不存在，创建新记录
					config = database.SystemConfig{
						Key:   key,
						Value: item.Value,
						Type:  item.Type,
					}
					if err := tx.Create(&config).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}

			// 存在，更新记录
			config.Value = item.Value
			config.Type = item.Type
			if err := tx.Save(&config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// 批量删除配置
func (r *SystemConfigRepositoryImpl) BatchDelete(ctx context.Context, keys []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, key := range keys {
			if err := tx.Where("key = ?", key).Delete(&database.SystemConfig{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
