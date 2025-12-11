// Package file 文件配置提供者
package file

import (
	"context"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileProvider 文件配置提供者
type FileProvider struct {
	basePath string
}

// NewFileProvider 创建文件配置提供者
func NewFileProvider(basePath string) *FileProvider {
	return &FileProvider{
		basePath: basePath,
	}
}

// Load 加载配置文件
func (p *FileProvider) Load(ctx context.Context, path string) (map[string]any, error) {
	fullPath := filepath.Join(p.basePath, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// Watch 监听文件变化
func (p *FileProvider) Watch(ctx context.Context, path string, callback func(map[string]any)) error {
	// TODO: 实现文件监听
	return nil
}

// Save 保存配置文件
func (p *FileProvider) Save(ctx context.Context, path string, config map[string]any) error {
	fullPath := filepath.Join(p.basePath, path)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(fullPath, data, 0644)
}

// Validate 验证配置
func (p *FileProvider) Validate(ctx context.Context, config map[string]any, rules map[string]any) error {
	// TODO: 实现配置验证
	return nil
}
