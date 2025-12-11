package modules

import (
	"moviepilot-go/internal/business/domains/module"
)

// ModuleAdapter 现有Module接口到新Module接口的适配器
type ModuleAdapter struct {
	Module
}

// ID 获取模块ID
func (ma *ModuleAdapter) ID() string {
	return ma.Module.GetID()
}

// Type 获取模块类型
func (ma *ModuleAdapter) Type() module.ModuleType {
	// 默认返回other类型，后续可以根据模块实际类型进行映射
	return module.ModuleTypeOther
}

// SubType 获取模块子类型
func (ma *ModuleAdapter) SubType() string {
	// 默认返回空字符串，后续可以根据模块实际子类型进行映射
	return ""
}

// Init 初始化模块
func (ma *ModuleAdapter) Init(cfg any) error {
	return ma.Module.Initialize()
}

// Stop 停止模块
func (ma *ModuleAdapter) Stop() error {
	return ma.Module.Stop()
}

// Test 测试模块连接
func (ma *ModuleAdapter) Test() (bool, string) {
	// 现有模块可能没有Test方法，默认返回true
	return true, "测试通过"
}

// SettingInfo 获取模块配置开关信息
func (ma *ModuleAdapter) SettingInfo() (*module.Setting, bool) {
	// 现有模块没有SettingInfo方法，默认返回nil和false，表示不使用配置开关
	return nil, false
}

// NewModuleAdapter 创建现有Module到新Module的适配器
func NewModuleAdapter(m Module) module.Module {
	return &ModuleAdapter{
		Module: m,
	}
}
