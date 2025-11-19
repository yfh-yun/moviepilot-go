package repositories

import (
	"fmt"

	"github.com/yfh-yun/moviepilot-go/internal/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userConfigRepository struct {
	db *gorm.DB
}

// NewUserConfigRepository 创建用户配置仓储
func NewUserConfigRepository() interfaces.UserConfigRepository {
	return &userConfigRepository{
		db: database.GetDB(),
	}
}

// Create 创建用户配置
func (r *userConfigRepository) Create(config *model.UserConfig) error {
	logger.Debug("Creating user config", 
		zap.String("username", config.Username),
		zap.String("key", config.Key))
	
	if err := r.db.Create(config).Error; err != nil {
		logger.Error("Failed to create user config", zap.Error(err))
		return fmt.Errorf("failed to create user config: %w", err)
	}
	
	logger.Info("User config created successfully", zap.Uint("id", config.ID))
	return nil
}

// GetByID 根据ID获取用户配置
func (r *userConfigRepository) GetByID(id uint) (*model.UserConfig, error) {
	var config model.UserConfig
	if err := r.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user config with id %d not found", id)
		}
		logger.Error("Failed to get user config by ID", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get user config: %w", err)
	}
	
	return &config, nil
}

// Update 更新用户配置
func (r *userConfigRepository) Update(config *model.UserConfig) error {
	logger.Debug("Updating user config", zap.Uint("id", config.ID))
	
	if err := r.db.Save(config).Error; err != nil {
		logger.Error("Failed to update user config", zap.Uint("id", config.ID), zap.Error(err))
		return fmt.Errorf("failed to update user config: %w", err)
	}
	
	logger.Info("User config updated successfully", zap.Uint("id", config.ID))
	return nil
}

// Delete 删除用户配置
func (r *userConfigRepository) Delete(id uint) error {
	logger.Debug("Deleting user config", zap.Uint("id", id))
	
	if err := r.db.Delete(&model.UserConfig{}, id).Error; err != nil {
		logger.Error("Failed to delete user config", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete user config: %w", err)
	}
	
	logger.Info("User config deleted successfully", zap.Uint("id", id))
	return nil
}

// List 获取用户配置列表
func (r *userConfigRepository) List(offset, limit int) ([]*model.UserConfig, int64, error) {
	var configs []*model.UserConfig
	var total int64
	
	// 获取总数
	if err := r.db.Model(&model.UserConfig{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count user configs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count user configs: %w", err)
	}
	
	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Order("id DESC").Find(&configs).Error; err != nil {
		logger.Error("Failed to list user configs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list user configs: %w", err)
	}
	
	return configs, total, nil
}

// GetByKey 按用户和键获取
func (r *userConfigRepository) GetByKey(username, key string) (*model.UserConfig, error) {
	var config model.UserConfig
	if err := r.db.Where("username = ? AND key = ?", username, key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to get user config by key", 
			zap.String("username", username), 
			zap.String("key", key), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user config by key: %w", err)
	}
	
	return &config, nil
}

// GetByUsername 按用户获取
func (r *userConfigRepository) GetByUsername(username string) ([]*model.UserConfig, error) {
	var configs []*model.UserConfig
	
	if err := r.db.Where("username = ?", username).Order("id DESC").Find(&configs).Error; err != nil {
		logger.Error("Failed to list user configs by username", 
			zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("failed to list user configs by username: %w", err)
	}
	
	return configs, nil
}

// GetByKeyOnly 按键获取
func (r *userConfigRepository) GetByKeyOnly(key string) ([]*model.UserConfig, error) {
	var configs []*model.UserConfig
	
	if err := r.db.Where("key = ?", key).Order("id DESC").Find(&configs).Error; err != nil {
		logger.Error("Failed to list user configs by key", 
			zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("failed to list user configs by key: %w", err)
	}
	
	return configs, nil
}

// SetConfig 设置配置
func (r *userConfigRepository) SetConfig(username, key, value string) error {
	logger.Debug("Setting user config", 
		zap.String("username", username),
		zap.String("key", key),
		zap.String("value", value))
	
	config, err := r.GetByKey(username, key)
	if err != nil {
		return fmt.Errorf("failed to check existing user config: %w", err)
	}
	
	if config != nil {
		// 更新现有配置
		config.Value = value
		if err := r.Update(config); err != nil {
			return fmt.Errorf("failed to update user config: %w", err)
		}
	} else {
		// 创建新配置
		newConfig := &model.UserConfig{
			Username: username,
			Key:      key,
			Value:    value,
		}
		if err := r.Create(newConfig); err != nil {
			return fmt.Errorf("failed to create user config: %w", err)
		}
	}
	
	logger.Info("User config set successfully", 
		zap.String("username", username),
		zap.String("key", key))
	return nil
}

// DeleteByKey 删除配置
func (r *userConfigRepository) DeleteByKey(username, key string) error {
	logger.Debug("Deleting user config", 
		zap.String("username", username),
		zap.String("key", key))
	
	if err := r.db.Where("username = ? AND key = ?", username, key).Delete(&model.UserConfig{}).Error; err != nil {
		logger.Error("Failed to delete user config by key", 
			zap.String("username", username), 
			zap.String("key", key), 
			zap.Error(err))
		return fmt.Errorf("failed to delete user config by key: %w", err)
	}
	
	logger.Info("User config deleted successfully", 
		zap.String("username", username),
		zap.String("key", key))
	return nil
}

// DeleteByUsername 删除用户的所有配置
func (r *userConfigRepository) DeleteByUsername(username string) error {
	logger.Debug("Deleting all user configs", zap.String("username", username))
	
	if err := r.db.Where("username = ?", username).Delete(&model.UserConfig{}).Error; err != nil {
		logger.Error("Failed to delete user configs by username", 
			zap.String("username", username), zap.Error(err))
		return fmt.Errorf("failed to delete user configs by username: %w", err)
	}
	
	logger.Info("All user configs deleted successfully", zap.String("username", username))
	return nil
}

// BatchSet 批量设置配置
func (r *userConfigRepository) BatchSet(username string, configs map[string]string) error {
	logger.Debug("Batch setting user configs", 
		zap.String("username", username),
		zap.Int("count", len(configs)))
	
	for key, value := range configs {
		if err := r.SetConfig(username, key, value); err != nil {
			return fmt.Errorf("failed to set user config %s: %w", key, err)
		}
	}
	
	logger.Info("User configs batch set successfully", 
		zap.String("username", username),
		zap.Int("count", len(configs)))
	return nil
}