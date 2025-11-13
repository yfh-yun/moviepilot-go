package indexer

import (
	"moviepilot-go/pkg/modules"
	"moviepilot-go/pkg/models"
)

// Ensure IndexerModule implements the Module interface
var _ modules.Module = (*IndexerModule)(nil)

// NewIndexerModule 创建索引模块实例
func NewIndexerModule() *IndexerModule {
	return &IndexerModule{}
}

// GetInstance 获取模块实例
func GetInstance() modules.Module {
	return NewIndexerModule()
}

// RegisterModule 注册模块
func RegisterModule() {
	// 注册索引模块到模块管理器
	// moduleManager.Register(&IndexerModule{})
}
