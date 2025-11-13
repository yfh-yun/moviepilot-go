package modules

import (
	"moviepilot-go/pkg/models"
)

// ModuleBase 模块基类，实现对应方法，在有需要时会被自动调用，返回nil代表不启用该模块，将继续执行下一模块
// 输入参数与输出参数一致的，或没有输出的，可以被多个模块重复实�?type ModuleBase struct {
	// 模块名称
	Name string
	// 模块类型
	Type models.ModuleType
	// 模块子类�?	SubType interface{}
	// 模块优先�?	Priority int
}

// NewModuleBase 创建一个新的模块基类实�?func NewModuleBase() *ModuleBase {
	return &ModuleBase{}
}

// InitModule 模块初始�?func (m *ModuleBase) InitModule() error {
	// 子类需要实现此方法
	return nil
}

// InitSetting 模块开关设置，返回开关名和开关值，开关值为true时代表有值即打开，不实现该方法或返回nil代表不使用开�?// 部分模块支持同时开启多个，此时设置项以,分隔，开关值使用contains判断
func (m *ModuleBase) InitSetting() (string, interface{}) {
	// 子类可以重写此方�?	return "", nil
}

// GetName 获取模块名称
func (m *ModuleBase) GetName() string {
	return m.Name
}

// GetType 获取模块类型
func (m *ModuleBase) GetType() models.ModuleType {
	return m.Type
}

// GetSubType 获取模块子类型（下载器、媒体服务器、消息通道、存储类型、其他杂项模块类型）
func (m *ModuleBase) GetSubType() interface{} {
	return m.SubType
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (m *ModuleBase) GetPriority() int {
	return m.Priority
}

// Stop 如果关闭时模块有服务需要停止，需要实现此方法
func (m *ModuleBase) Stop() error {
	// 子类需要实现此方法
	return nil
}

// Test 模块测试, 返回测试结果和错误信�?func (m *ModuleBase) Test() (bool, string) {
	// 子类需要实现此方法
	return true, ""
}
