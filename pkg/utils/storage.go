package utils

import (
	"fmt"
	"sync"
)

// StorageHelper 存储帮助类
type StorageHelper struct {
	storages map[string]*StorageService
	mutex    sync.RWMutex
}

// StorageService 存储服务信息
type StorageService struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string                 `json:"type"`
	Config     map[string]interface{} `json:"config"`
}

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal    StorageType = "local"
	StorageTypeFTP      StorageType = "ftp"
	StorageTypeSFTP     StorageType = "sftp"
	StorageTypeS3       StorageType = "s3"
	StorageTypeOSS      StorageType = "oss"
	StorageTypeCOS      StorageType = "cos"
	StorageTypeWebDAV   StorageType = "webdav"
	StorageTypeOneDrive StorageType = "onedrive"
)

// NewStorageHelper 创建存储助手实例
func NewStorageHelper() *StorageHelper {
	return &StorageHelper{
		storages: make(map[string]*StorageService),
	}
}

// GetStorages 获取所有存储设置
func (sh *StorageHelper) GetStorages() []*StorageService {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	storages := make([]*StorageService, 0, len(sh.storages))
	for _, storage := range sh.storages {
		storages = append(storages, storage)
	}

	return storages
}

// GetStorage 获取指定存储配置
func (sh *StorageHelper) GetStorage(storageType string) (*StorageService, error) {
	if storageType == "" {
		return nil, fmt.Errorf("storage type cannot be empty")
	}

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	storage, exists := sh.storages[storageType]
	if !exists {
		return nil, fmt.Errorf("storage not found: %s", storageType)
	}

	return storage, nil
}

// SetStorage 设置存储配置
func (sh *StorageHelper) SetStorage(storageType string, config map[string]interface{}) error {
	if storageType == "" {
		return fmt.Errorf("storage type cannot be empty")
	}

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证存储类型
	if !sh.isValidStorageType(storageType) {
		return fmt.Errorf("unsupported storage type: %s", storageType)
	}

	// 验证配置
	if err := sh.validateStorageConfig(storageType, config); err != nil {
		return fmt.Errorf("invalid config for %s: %v", storageType, err)
	}

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	sh.storages[storageType] = &StorageService{
		Type:   storageType,
		Config: config,
	}

	return nil
}

// RemoveStorage 移除存储配置
func (sh *StorageHelper) RemoveStorage(storageType string) error {
	if storageType == "" {
		return fmt.Errorf("storage type cannot be empty")
	}

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	if _, exists := sh.storages[storageType]; !exists {
		return fmt.Errorf("storage not found: %s", storageType)
	}

	delete(sh.storages, storageType)
	return nil
}

// UpdateStorage 更新存储配置
func (sh *StorageHelper) UpdateStorage(storageType string, config map[string]interface{}) error {
	return sh.SetStorage(storageType, config)
}

// IsStorageEnabled 检查存储是否启用
func (sh *StorageHelper) IsStorageEnabled(storageType string) bool {
	storage, err := sh.GetStorage(storageType)
	if err != nil {
		return false
	}

	// 检查配置中是否有enabled字段
	if enabled, exists := storage.Config["enabled"]; exists {
		if enabledBool, ok := enabled.(bool); ok {
			return enabledBool
		}
	}

	// 默认认为存在配置就是启用的
	return true
}

// GetStorageConfig 获取存储配置
func (sh *StorageHelper) GetStorageConfig(storageType string) (map[string]interface{}, error) {
	storage, err := sh.GetStorage(storageType)
	if err != nil {
		return nil, err
	}

	return storage.Config, nil
}

// isValidStorageType 检查是否为有效的存储类型
func (sh *StorageHelper) isValidStorageType(storageType string) bool {
	validTypes := []string{
		string(StorageTypeLocal),
		string(StorageTypeFTP),
		string(StorageTypeSFTP),
		string(StorageTypeS3),
		string(StorageTypeOSS),
		string(StorageTypeCOS),
		string(StorageTypeWebDAV),
		string(StorageTypeOneDrive),
	}

	for _, validType := range validTypes {
		if storageType == validType {
			return true
		}
	}

	return false
}

// validateStorageConfig 验证存储配置
func (sh *StorageHelper) validateStorageConfig(storageType string, config map[string]interface{}) error {
	switch storageType {
	case string(StorageTypeLocal):
		return sh.validateLocalConfig(config)
	case string(StorageTypeFTP):
		return sh.validateFTPConfig(config)
	case string(StorageTypeSFTP):
		return sh.validateSFTPConfig(config)
	case string(StorageTypeS3):
		return sh.validateS3Config(config)
	case string(StorageTypeOSS):
		return sh.validateOSSConfig(config)
	case string(StorageTypeCOS):
		return sh.validateCOSConfig(config)
	case string(StorageTypeWebDAV):
		return sh.validateWebDAVConfig(config)
	case string(StorageTypeOneDrive):
		return sh.validateOneDriveConfig(config)
	default:
		return fmt.Errorf("unknown storage type: %s", storageType)
	}
}

// validateLocalConfig 验证本地存储配置
func (sh *StorageHelper) validateLocalConfig(config map[string]interface{}) error {
	if path, exists := config["path"]; exists {
		if pathStr, ok := path.(string); ok && pathStr == "" {
			return fmt.Errorf("local storage path cannot be empty")
		}
	}
	return nil
}

// validateFTPConfig 验证FTP配置
func (sh *StorageHelper) validateFTPConfig(config map[string]interface{}) error {
	if host, exists := config["host"]; !exists || host.(string) == "" {
		return fmt.Errorf("FTP host is required")
	}

	if port, exists := config["port"]; exists {
		if portInt, ok := port.(int); ok && (portInt <= 0 || portInt > 65535) {
			return fmt.Errorf("invalid FTP port: %d", portInt)
		}
	}

	if username, exists := config["username"]; exists {
		if usernameStr, ok := username.(string); ok && usernameStr == "" {
			return fmt.Errorf("FTP username cannot be empty when provided")
		}
	}

	return nil
}

// validateSFTPConfig 验证SFTP配置
func (sh *StorageHelper) validateSFTPConfig(config map[string]interface{}) error {
	if host, exists := config["host"]; !exists || host.(string) == "" {
		return fmt.Errorf("SFTP host is required")
	}

	if port, exists := config["port"]; exists {
		if portInt, ok := port.(int); ok && (portInt <= 0 || portInt > 65535) {
			return fmt.Errorf("invalid SFTP port: %d", portInt)
		}
	}

	if username, exists := config["username"]; !exists || username.(string) == "" {
		return fmt.Errorf("SFTP username is required")
	}

	// 检查认证方式
	password, hasPassword := config["password"]
	privateKey, hasPrivateKey := config["private_key"]

	if !hasPassword && !hasPrivateKey {
		return fmt.Errorf("SFTP requires either password or private key")
	}

	if hasPassword && password.(string) == "" {
		return fmt.Errorf("SFTP password cannot be empty when provided")
	}

	if hasPrivateKey && privateKey.(string) == "" {
		return fmt.Errorf("SFTP private key cannot be empty when provided")
	}

	return nil
}

// validateS3Config 验证S3配置
func (sh *StorageHelper) validateS3Config(config map[string]interface{}) error {
	if endpoint, exists := config["endpoint"]; !exists || endpoint.(string) == "" {
		return fmt.Errorf("S3 endpoint is required")
	}

	if bucket, exists := config["bucket"]; !exists || bucket.(string) == "" {
		return fmt.Errorf("S3 bucket is required")
	}

	if accessKey, exists := config["access_key"]; !exists || accessKey.(string) == "" {
		return fmt.Errorf("S3 access key is required")
	}

	if secretKey, exists := config["secret_key"]; !exists || secretKey.(string) == "" {
		return fmt.Errorf("S3 secret key is required")
	}

	if region, exists := config["region"]; exists && region.(string) == "" {
		return fmt.Errorf("S3 region cannot be empty when provided")
	}

	return nil
}

// validateOSSConfig 验证OSS配置
func (sh *StorageHelper) validateOSSConfig(config map[string]interface{}) error {
	if endpoint, exists := config["endpoint"]; !exists || endpoint.(string) == "" {
		return fmt.Errorf("OSS endpoint is required")
	}

	if bucket, exists := config["bucket"]; !exists || bucket.(string) == "" {
		return fmt.Errorf("OSS bucket is required")
	}

	if accessKey, exists := config["access_key"]; !exists || accessKey.(string) == "" {
		return fmt.Errorf("OSS access key is required")
	}

	if secretKey, exists := config["secret_key"]; !exists || secretKey.(string) == "" {
		return fmt.Errorf("OSS secret key is required")
	}

	return nil
}

// validateCOSConfig 验证COS配置
func (sh *StorageHelper) validateCOSConfig(config map[string]interface{}) error {
	if secretID, exists := config["secret_id"]; !exists || secretID.(string) == "" {
		return fmt.Errorf("COS secret ID is required")
	}

	if secretKey, exists := config["secret_key"]; !exists || secretKey.(string) == "" {
		return fmt.Errorf("COS secret key is required")
	}

	if region, exists := config["region"]; !exists || region.(string) == "" {
		return fmt.Errorf("COS region is required")
	}

	if bucket, exists := config["bucket"]; !exists || bucket.(string) == "" {
		return fmt.Errorf("COS bucket is required")
	}

	return nil
}

// validateWebDAVConfig 验证WebDAV配置
func (sh *StorageHelper) validateWebDAVConfig(config map[string]interface{}) error {
	if url, exists := config["url"]; !exists || url.(string) == "" {
		return fmt.Errorf("WebDAV URL is required")
	}

	if username, exists := config["username"]; exists {
		if usernameStr, ok := username.(string); ok && usernameStr == "" {
			return fmt.Errorf("WebDAV username cannot be empty when provided")
		}
	}

	if password, exists := config["password"]; exists {
		if passwordStr, ok := password.(string); ok && passwordStr == "" {
			return fmt.Errorf("WebDAV password cannot be empty when provided")
		}
	}

	return nil
}

// validateOneDriveConfig 验证OneDrive配置
func (sh *StorageHelper) validateOneDriveConfig(config map[string]interface{}) error {
	if clientID, exists := config["client_id"]; !exists || clientID.(string) == "" {
		return fmt.Errorf("OneDrive client ID is required")
	}

	if clientSecret, exists := config["client_secret"]; !exists || clientSecret.(string) == "" {
		return fmt.Errorf("OneDrive client secret is required")
	}

	if refreshToken, exists := config["refresh_token"]; !exists || refreshToken.(string) == "" {
		return fmt.Errorf("OneDrive refresh token is required")
	}

	return nil
}

// GetStorageCount 获取存储配置数量
func (sh *StorageHelper) GetStorageCount() int {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	return len(sh.storages)
}

// GetEnabledStorageCount 获取启用的存储配置数量
func (sh *StorageHelper) GetEnabledStorageCount() int {
	storages := sh.GetStorages()
	count := 0

	for _, storage := range storages {
		if sh.IsStorageEnabled(storage.Type) {
			count++
		}
	}

	return count
}

// ClearStorages 清空所有存储配置
func (sh *StorageHelper) ClearStorages() {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	sh.storages = make(map[string]*StorageService)
}

// ImportStorages 导入存储配置
func (sh *StorageHelper) ImportStorages(configs []*StorageConfig) error {
	if configs == nil {
		return fmt.Errorf("configs cannot be nil")
	}

	for _, config := range configs {
		if err := sh.SetStorage(config.Type, config.Config); err != nil {
			return fmt.Errorf("failed to import storage %s: %v", config.Type, err)
		}
	}

	return nil
}

// ExportStorages 导出存储配置
func (sh *StorageHelper) ExportStorages() []*StorageConfig {
	storages := sh.GetStorages()
	configs := make([]*StorageConfig, 0, len(storages))

	for _, storage := range storages {
		config := &StorageConfig{
			Type:   storage.Type,
			Config: storage.Config,
		}
		configs = append(configs, config)
	}

	return configs
}

// GetStorageByPriority 按优先级获取存储配置
func (sh *StorageHelper) GetStorageByPriority() []*StorageService {
	storages := sh.GetStorages()
	
	// 按优先级排序（如果有priority字段）
	for i := 0; i < len(storages)-1; i++ {
		for j := i + 1; j < len(storages); j++ {
			priority1 := sh.getStoragePriority(storages[i])
			priority2 := sh.getStoragePriority(storages[j])
			
			if priority1 > priority2 {
				storages[i], storages[j] = storages[j], storages[i]
			}
		}
	}

	return storages
}

// getStoragePriority 获取存储优先级
func (sh *StorageHelper) getStoragePriority(storage *StorageService) int {
	if priority, exists := storage.Config["priority"]; exists {
		if priorityInt, ok := priority.(int); ok {
			return priorityInt
		}
	}
	
	// 默认优先级
	switch storage.Type {
	case string(StorageTypeLocal):
		return 1
	case string(StorageTypeSFTP):
		return 2
	case string(StorageTypeFTP):
		return 3
	case string(StorageTypeS3):
		return 4
	case string(StorageTypeOSS):
		return 5
	case string(StorageTypeCOS):
		return 6
	case string(StorageTypeWebDAV):
		return 7
	case string(StorageTypeOneDrive):
		return 8
	default:
		return 99
	}
}

// TestStorageConnection 测试存储连接
func (sh *StorageHelper) TestStorageConnection(storageType string) error {
	storage, err := sh.GetStorage(storageType)
	if err != nil {
		return err
	}

	// 这里应该实现实际的连接测试逻辑
	// 简化实现，只检查配置完整性
	return sh.validateStorageConfig(storageType, storage.Config)
}