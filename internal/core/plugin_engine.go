package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// Plugin 插件接口
type Plugin interface {
	// Name 插件名称
	Name() string
	
	// Version 插件版本
	Version() string
	
	// Events 插件订阅的事件列�?	Events() []string
	
	// HandleEvent 处理事件
	HandleEvent(event *Event) error
	
	// Permissions 插件所需权限
	Permissions() []string
}

// Manifest 插件清单文件结构
type Manifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Events      []string `yaml:"events"`
	Permissions []string `yaml:"permissions"`
	Author      string   `yaml:"author"`
}

// PluginEngine 插件引擎
type PluginEngine struct {
	plugins map[string]Plugin
	logger  *zap.Logger
	mu      sync.RWMutex
}

// NewPluginEngine 创建新的插件引擎
func NewPluginEngine(logger *zap.Logger) *PluginEngine {
	return &PluginEngine{
		plugins: make(map[string]Plugin),
		logger:  logger,
	}
}

// LoadPlugins 从指定目录加载插�?func (pe *PluginEngine) LoadPlugins(pluginDir string) error {
	// 检查目录是否存�?	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		pe.logger.Warn("插件目录不存�?, zap.String("dir", pluginDir))
		return nil
	}
	
	// 读取目录下的所有插件目�?	entries, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return fmt.Errorf("读取插件目录失败: %v", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		pluginPath := filepath.Join(pluginDir, entry.Name())
		if err := pe.loadPlugin(pluginPath); err != nil {
			pe.logger.Error("加载插件失败",
				zap.String("plugin", entry.Name()),
				zap.String("path", pluginPath),
				zap.Error(err))
		}
	}
	
	return nil
}

// loadPlugin 加载单个插件
func (pe *PluginEngine) loadPlugin(pluginPath string) error {
	// 读取插件清单文件
	manifestPath := filepath.Join(pluginPath, "manifest.yaml")
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("读取插件清单文件失败: %v", err)
	}
	
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("解析插件清单文件失败: %v", err)
	}
	
	// 检查必需文件是否存在
	requiredFiles := []string{"plugin.py", "requirements.txt"}
	for _, file := range requiredFiles {
		filePath := filepath.Join(pluginPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			pe.logger.Warn("插件缺少必需文件",
				zap.String("plugin", manifest.Name),
				zap.String("missing_file", file))
		}
	}
	
	// TODO: 实现Python插件的加载逻辑
	// 这里需要调用Python沙箱服务来加载插�?	
	pe.logger.Info("插件加载成功",
		zap.String("plugin", manifest.Name),
		zap.String("version", manifest.Version))
	
	return nil
}

// GetPlugin 获取指定名称的插�?func (pe *PluginEngine) GetPlugin(name string) Plugin {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	
	return pe.plugins[name]
}

// GetPlugins 获取所有插�?func (pe *PluginEngine) GetPlugins() map[string]Plugin {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	
	// 返回副本以避免外部修�?	result := make(map[string]Plugin)
	for name, plugin := range pe.plugins {
		result[name] = plugin
	}
	
	return result
}

// HandleEvent 处理事件，分发给所有订阅该事件的插�?func (pe *PluginEngine) HandleEvent(event *Event) error {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	
	// 遍历所有插件，检查是否订阅了该事�?	for name, plugin := range pe.plugins {
		events := plugin.Events()
		for _, e := range events {
			if strings.EqualFold(e, event.Type) {
				// 插件订阅了该事件，调用处理函�?				if err := plugin.HandleEvent(event); err != nil {
					pe.logger.Error("插件处理事件失败",
						zap.String("plugin", name),
						zap.String("event_type", event.Type),
						zap.Error(err))
				}
				break
			}
		}
	}
	
	return nil
}
