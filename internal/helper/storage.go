package helper

import (
	"moviepilot-go/internal/db"
	"moviepilot-go/pkg/models"
)

// StorageHelper 存储帮助�?type StorageHelper struct{}

// NewStorageHelper 创建StorageHelper实例
func NewStorageHelper() *StorageHelper {
	return &StorageHelper{}
}

// GetStoragies 获取所有存储设�?func (s *StorageHelper) GetStoragies() []models.StorageConf {
	/*
	 * 获取所有存储设�?	 */
	storageConfs := db.NewSystemConfigOper().Get(models.SystemConfigKeyStorages)
	if storageConfs == nil {
		return []models.StorageConf{}
	}
	
	// 类型断言获取存储配置数据
	storageList, ok := storageConfs.([]interface{})
	if !ok {
		return []models.StorageConf{}
	}
	
	// 转换为StorageConf数组
	result := make([]models.StorageConf, 0, len(storageList))
	for _, storage := range storageList {
		// 将map转换为StorageConf结构�?		if storageMap, ok := storage.(map[string]interface{}); ok {
			storageConf := models.StorageConf{}
			
			if storageType, exists := storageMap["type"]; exists {
				if typeStr, ok := storageType.(string); ok {
					storageConf.Type = typeStr
				}
			}
			
			if name, exists := storageMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					storageConf.Name = nameStr
				}
			}
			
			if config, exists := storageMap["config"]; exists {
				if configMap, ok := config.(map[string]interface{}); ok {
					storageConf.Config = configMap
				}
			}
			
			result = append(result, storageConf)
		}
	}
	
	return result
}

// GetStorage 获取指定存储配置
func (s *StorageHelper) GetStorage(storage string) *models.StorageConf {
	/*
	 * 获取指定存储配置
	 */
	storagies := s.GetStoragies()
	for _, storageConf := range storagies {
		// 注意：这里需要使用局部变量避免闭包问�?		s := storageConf
		if s.Type == storage {
			return &s
		}
	}
	return nil
}

// SetStorage 设置存储配置
func (s *StorageHelper) SetStorage(storage string, conf map[string]interface{}) {
	/*
	 * 设置存储配置
	 */
	storagies := s.GetStoragies()
	
	if len(storagies) == 0 {
		// 如果没有存储配置，则创建新的
		storagies = append(storagies, models.StorageConf{
			Type:   storage,
			Config: conf,
		})
	} else {
		// 查找并更新现有配�?		found := false
		for i := range storagies {
			if storagies[i].Type == storage {
				storagies[i].Config = conf
				found = true
				break
			}
		}
		
		// 如果没有找到，则添加新的配置
		if !found {
			storagies = append(storagies, models.StorageConf{
				Type:   storage,
				Config: conf,
			})
		}
	}
	
	// 转换为可存储的格�?	storageData := make([]map[string]interface{}, 0, len(storagies))
	for _, storageConf := range storagies {
		storageMap := map[string]interface{}{
			"type":   storageConf.Type,
			"name":   storageConf.Name,
			"config": storageConf.Config,
		}
		storageData = append(storageData, storageMap)
	}
	
	// 保存到系统配�?	db.NewSystemConfigOper().Set(models.SystemConfigKeyStorages, storageData)
}

// AddStorage 添加存储配置
func (s *StorageHelper) AddStorage(storage, name string, conf map[string]interface{}) {
	/*
	 * 添加存储配置
	 */
	storagies := s.GetStoragies()
	
	if len(storagies) == 0 {
		// 如果没有存储配置，则创建新的
		storagies = append(storagies, models.StorageConf{
			Type:   storage,
			Name:   name,
			Config: conf,
		})
	} else {
		// 添加新的配置
		storagies = append(storagies, models.StorageConf{
			Type:   storage,
			Name:   name,
			Config: conf,
		})
	}
	
	// 转换为可存储的格�?	storageData := make([]map[string]interface{}, 0, len(storagies))
	for _, storageConf := range storagies {
		storageMap := map[string]interface{}{
			"type":   storageConf.Type,
			"name":   storageConf.Name,
			"config": storageConf.Config,
		}
		storageData = append(storageData, storageMap)
	}
	
	// 保存到系统配�?	db.NewSystemConfigOper().Set(models.SystemConfigKeyStorages, storageData)
}

// ResetStorage 重置存储配置
func (s *StorageHelper) ResetStorage(storage string) {
	/*
	 * 重置存储配置
	 */
	storagies := s.GetStoragies()
	
	for i := range storagies {
		if storagies[i].Type == storage {
			storagies[i].Config = make(map[string]interface{})
			break
		}
	}
	
	// 转换为可存储的格�?	storageData := make([]map[string]interface{}, 0, len(storagies))
	for _, storageConf := range storagies {
		storageMap := map[string]interface{}{
			"type":   storageConf.Type,
			"name":   storageConf.Name,
			"config": storageConf.Config,
		}
		storageData = append(storageData, storageMap)
	}
	
	// 保存到系统配�?	db.NewSystemConfigOper().Set(models.SystemConfigKeyStorages, storageData)
}
