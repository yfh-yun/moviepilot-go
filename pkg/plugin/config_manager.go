package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// configManager 配置管理器实现
type configManager struct {
	configPath string
	configs    map[string]map[string]interface{}
	mutex      sync.RWMutex
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *configManager {
	return &configManager{
		configPath: "./configs/plugins",
		configs:    make(map[string]map[string]interface{}),
	}
}

// GetPluginConfig 获取插件配置
func (cm *configManager) GetPluginConfig(pluginID string) (map[string]interface{}, error) {
	cm.mutex.RLock()
	config, exists := cm.configs[pluginID]
	cm.mutex.RUnlock()

	if exists {
		return config, nil
	}

	// 从文件加载配置
	config, err := cm.loadConfigFromFile(pluginID)
	if err != nil {
		return nil, err
	}

	cm.mutex.Lock()
	cm.configs[pluginID] = config
	cm.mutex.Unlock()

	return config, nil
}

// SetPluginConfig 设置插件配置
func (cm *configManager) SetPluginConfig(pluginID string, config map[string]interface{}) error {
	// 保存到内存
	cm.mutex.Lock()
	cm.configs[pluginID] = config
	cm.mutex.Unlock()

	// 保存到文件
	return cm.saveConfigToFile(pluginID, config)
}

// loadConfigFromFile 从文件加载配置
func (cm *configManager) loadConfigFromFile(pluginID string) (map[string]interface{}, error) {
	configFile := filepath.Join(cm.configPath, pluginID+".json")

	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 返回默认配置
		return make(map[string]interface{}), nil
	}

	// 读取文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析JSON
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// saveConfigToFile 保存配置到文件
func (cm *configManager) saveConfigToFile(pluginID string, config map[string]interface{}) error {
	// 确保目录存在
	if err := os.MkdirAll(cm.configPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	configFile := filepath.Join(cm.configPath, pluginID+".json")
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}