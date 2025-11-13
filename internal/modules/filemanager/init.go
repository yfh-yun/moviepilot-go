package filemanager

// init.go 文件管理器模块初始化文件

import (
	"moviepilot-go/internal/modules"
)

// 模块实例变量
var instance modules.Module

// GetInstance 获取模块实例
func GetInstance() modules.Module {
	if instance == nil {
		instance = NewFileManagerModule()
	}
	return instance
}

// RegisterModule 注册模块
func RegisterModule() {
	// 注册文件管理器模�?	modules.RegisterModule("filemanager", GetInstance)
}
