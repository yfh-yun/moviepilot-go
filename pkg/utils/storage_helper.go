package utils

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// StorageConf 存储配置
type StorageConf struct {
	Type    string         `json:"type"`
	Name    string         `json:"name"`
	Enabled bool           `json:"enabled"`
	Config  map[string]any `json:"config"`
}

// StorageConfigStore 配置存储接口
type StorageConfigStore interface {
	// Get 获取配置
	Get(key string, defaultValue any) any
	// Set 设置配置
	Set(key string, value any)
}

// StorageSystemConfigKey 系统配置键
type StorageSystemConfigKey string

const (
	StorageSystemConfigKeyStorages StorageSystemConfigKey = "Storages"
)

// StorageHelper 存储帮助类
type StorageHelper struct {
	logger      *zap.Logger
	configStore StorageConfigStore
	mutex       sync.RWMutex
}

// NewStorageHelper 创建存储帮助类实例
func NewStorageHelper(configStore StorageConfigStore) *StorageHelper {
	return &StorageHelper{
		logger:      logger.GetLogger(),
		configStore: configStore,
	}
}

// GetStorages 获取所有存储配置
func (h *StorageHelper) GetStorages() []*StorageConf {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 从配置存储中获取存储配置
	storages := h.configStore.Get(string(StorageSystemConfigKeyStorages), []*StorageConf{}).([]*StorageConf)
	if storages == nil {
		return []*StorageConf{}
	}

	return storages
}

// GetStorage 获取指定类型的存储配置
func (h *StorageHelper) GetStorage(storageType string) *StorageConf {
	storages := h.GetStorages()
	for _, s := range storages {
		if s.Type == storageType {
			return s
		}
	}
	return nil
}

// SetStorage 设置存储配置
func (h *StorageHelper) SetStorage(storageType string, config map[string]any) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storages := h.GetStorages()
	found := false

	for _, s := range storages {
		if s.Type == storageType {
			s.Config = config
			found = true
			break
		}
	}

	if !found {
		// 如果存储不存在，创建新的存储配置
		newStorage := &StorageConf{
			Type:    storageType,
			Name:    storageType,
			Enabled: true,
			Config:  config,
		}
		storages = append(storages, newStorage)
	}

	// 保存配置
	h.configStore.Set(string(StorageSystemConfigKeyStorages), storages)
	h.logger.Info("设置存储配置成功", zap.String("type", storageType))
	return nil
}

// AddStorage 添加存储配置
func (h *StorageHelper) AddStorage(storageType, name string, config map[string]any) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storages := h.GetStorages()

	// 检查存储类型是否已存在
	for _, s := range storages {
		if s.Type == storageType {
			return fmt.Errorf("存储类型已存在: %s", storageType)
		}
	}

	// 创建新的存储配置
	newStorage := &StorageConf{
		Type:    storageType,
		Name:    name,
		Enabled: true,
		Config:  config,
	}

	// 添加到存储列表
	storages = append(storages, newStorage)

	// 保存配置
	h.configStore.Set(string(StorageSystemConfigKeyStorages), storages)
	h.logger.Info("添加存储配置成功", zap.String("type", storageType), zap.String("name", name))
	return nil
}

// ResetStorage 重置存储配置
func (h *StorageHelper) ResetStorage(storageType string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storages := h.GetStorages()
	found := false

	for _, s := range storages {
		if s.Type == storageType {
			s.Config = make(map[string]any)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("存储类型不存在: %s", storageType)
	}

	// 保存配置
	h.configStore.Set(string(StorageSystemConfigKeyStorages), storages)
	h.logger.Info("重置存储配置成功", zap.String("type", storageType))
	return nil
}

// DeleteStorage 删除存储配置
func (h *StorageHelper) DeleteStorage(storageType string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storages := h.GetStorages()
	for i, s := range storages {
		if s.Type == storageType {
			// 从切片中删除元素
			storages = append(storages[:i], storages[i+1:]...)
			// 保存配置
			h.configStore.Set(string(StorageSystemConfigKeyStorages), storages)
			h.logger.Info("删除存储配置成功", zap.String("type", storageType))
			return nil
		}
	}

	return fmt.Errorf("存储类型不存在: %s", storageType)
}

// UpdateStorage 更新存储配置
func (h *StorageHelper) UpdateStorage(storageType string, name string, enabled bool, config map[string]any) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storages := h.GetStorages()
	found := false

	for _, s := range storages {
		if s.Type == storageType {
			s.Name = name
			s.Enabled = enabled
			s.Config = config
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("存储类型不存在: %s", storageType)
	}

	// 保存配置
	h.configStore.Set(string(StorageSystemConfigKeyStorages), storages)
	h.logger.Info("更新存储配置成功", zap.String("type", storageType))
	return nil
}

// EnableStorage 启用存储
func (h *StorageHelper) EnableStorage(storageType string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storages := h.GetStorages()
	found := false

	for _, s := range storages {
		if s.Type == storageType {
			s.Enabled = true
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("存储类型不存在: %s", storageType)
	}

	// 保存配置
	h.configStore.Set(string(StorageSystemConfigKeyStorages), storages)
	h.logger.Info("启用存储成功", zap.String("type", storageType))
	return nil
}

// DisableStorage 禁用存储
func (h *StorageHelper) DisableStorage(storageType string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storages := h.GetStorages()
	found := false

	for _, s := range storages {
		if s.Type == storageType {
			s.Enabled = false
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("存储类型不存在: %s", storageType)
	}

	// 保存配置
	h.configStore.Set(string(StorageSystemConfigKeyStorages), storages)
	h.logger.Info("禁用存储成功", zap.String("type", storageType))
	return nil
}

// GetEnabledStorages 获取所有启用的存储配置
func (h *StorageHelper) GetEnabledStorages() []*StorageConf {
	storages := h.GetStorages()
	enabledStorages := make([]*StorageConf, 0)

	for _, s := range storages {
		if s.Enabled {
			enabledStorages = append(enabledStorages, s)
		}
	}

	return enabledStorages
}
