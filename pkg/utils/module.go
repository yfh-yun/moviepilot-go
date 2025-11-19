package utils

import (
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"sync"
)

// ModuleHelper 模块动态加载助手
type ModuleHelper struct {
	loadedModules map[string]interface{}
	moduleMutex   sync.RWMutex
}

// FilterFuncType 模块过滤器函数类型
type FilterFuncType func(name string, obj interface{}) bool

// ModuleInfo 模块信息
type ModuleInfo struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"`
	Instance interface{} `json:"instance,omitempty"`
	Loaded   bool        `json:"loaded"`
	Error    string      `json:"error,omitempty"`
}

// NewModuleHelper 创建模块助手实例
func NewModuleHelper() *ModuleHelper {
	return &ModuleHelper{
		loadedModules: make(map[string]interface{}),
	}
}

// Load 加载模块
func (mh *ModuleHelper) Load(packagePath string, filterFunc FilterFuncType) ([]interface{}, error) {
	if packagePath == "" {
		return nil, fmt.Errorf("package path cannot be empty")
	}

	if filterFunc == nil {
		filterFunc = mh.defaultFilter
	}

	var modules []interface{}
	loadedNames := make(map[string]bool)

	// 在Go中，动态加载模块比Python复杂
	// 这里提供一个简化的实现，主要适用于插件系统
	
	// 扫描目录中的.so文件（Go插件）
	pluginFiles, err := mh.findPluginFiles(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to find plugin files: %v", err)
	}

	for _, pluginFile := range pluginFiles {
		module, err := mh.loadPlugin(pluginFile, filterFunc, loadedNames)
		if err != nil {
			// 记录错误但继续处理其他模块
			continue
		}

		if module != nil {
			modules = append(modules, module)
		}
	}

	return modules, nil
}

// LoadPlugin 加载单个插件
func (mh *ModuleHelper) LoadPlugin(pluginPath string) (interface{}, error) {
	if pluginPath == "" {
		return nil, fmt.Errorf("plugin path cannot be empty")
	}

	// 加载Go插件
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %v", err)
	}

	// 查找导出的符号
	symbols, err := mh.findPluginSymbols(plug)
	if err != nil {
		return nil, fmt.Errorf("failed to find symbols: %v", err)
	}

	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols found in plugin")
	}

	// 返回第一个找到的符号
	return symbols[0], nil
}

// findPluginFiles 查找插件文件
func (mh *ModuleHelper) findPluginFiles(packagePath string) ([]string, error) {
	// 在Go中，插件文件通常是.so文件
	pattern := filepath.Join(packagePath, "*.so")
	
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob plugin files: %v", err)
	}

	return matches, nil
}

// loadPlugin 加载插件
func (mh *ModuleHelper) loadPlugin(pluginPath string, filterFunc FilterFuncType, loadedNames map[string]bool) (interface{}, error) {
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %v", pluginPath, err)
	}

	symbols, err := mh.findPluginSymbols(plug)
	if err != nil {
		return nil, fmt.Errorf("failed to find symbols in %s: %v", pluginPath, err)
	}

	for name, symbol := range symbols {
		// 跳过私有符号
		if strings.HasPrefix(name, "_") {
			continue
		}

		// 检查是否已加载
		if loadedNames[name] {
			continue
		}

		// 应用过滤器
		if filterFunc(name, symbol) {
			loadedNames[name] = true
			
			// 缓存模块
			mh.moduleMutex.Lock()
			mh.loadedModules[name] = symbol
			mh.moduleMutex.Unlock()

			return symbol, nil
		}
	}

	return nil, nil
}

// findPluginSymbols 查找插件符号
func (mh *ModuleHelper) findPluginSymbols(plug *plugin.Plugin) (map[string]interface{}, error) {
	symbols := make(map[string]interface{})

	// 常见的符号名称
	commonSymbols := []string{
		"Plugin",
		"Handler",
		"Service",
		"Worker",
		"Module",
	}

	for _, symbolName := range commonSymbols {
		if symbol, err := plug.Lookup(symbolName); err == nil {
			symbols[symbolName] = symbol
		}
	}

	return symbols, nil
}

// defaultFilter 默认过滤器
func (mh *ModuleHelper) defaultFilter(name string, obj interface{}) bool {
	return name != "" && obj != nil
}

// GetLoadedModules 获取已加载的模块
func (mh *ModuleHelper) GetLoadedModules() map[string]interface{} {
	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	// 返回副本
	modules := make(map[string]interface{})
	for name, module := range mh.loadedModules {
		modules[name] = module
	}

	return modules
}

// GetModule 获取指定模块
func (mh *ModuleHelper) GetModule(name string) (interface{}, error) {
	if name == "" {
		return nil, fmt.Errorf("module name cannot be empty")
	}

	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	module, exists := mh.loadedModules[name]
	if !exists {
		return nil, fmt.Errorf("module not found: %s", name)
	}

	return module, nil
}

// UnloadModule 卸载模块
func (mh *ModuleHelper) UnloadModule(name string) error {
	if name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	mh.moduleMutex.Lock()
	defer mh.moduleMutex.Unlock()

	if _, exists := mh.loadedModules[name]; !exists {
		return fmt.Errorf("module not found: %s", name)
	}

	delete(mh.loadedModules, name)
	return nil
}

// ClearModules 清空所有模块
func (mh *ModuleHelper) ClearModules() {
	mh.moduleMutex.Lock()
	defer mh.moduleMutex.Unlock()

	mh.loadedModules = make(map[string]interface{})
}

// ReloadModule 重新加载模块
func (mh *ModuleHelper) ReloadModule(name string, packagePath string, filterFunc FilterFuncType) error {
	if err := mh.UnloadModule(name); err != nil {
		return err
	}

	modules, err := mh.Load(packagePath, filterFunc)
	if err != nil {
		return err
	}

	// 检查是否重新加载成功
	for _, module := range modules {
		if mh.isModuleNamed(module, name) {
			return nil
		}
	}

	return fmt.Errorf("failed to reload module: %s", name)
}

// isModuleNamed 检查模块是否具有指定名称
func (mh *ModuleHelper) isModuleNamed(module interface{}, name string) bool {
	if module == nil {
		return false
	}

	// 通过反射获取模块信息
	val := reflect.ValueOf(module)
	typ := reflect.TypeOf(module)

	// 检查类型名称
	if typ.Name() == name {
		return true
	}

	// 检查字符串表示
	moduleStr := fmt.Sprintf("%v", module)
	if strings.Contains(moduleStr, name) {
		return true
	}

	return false
}

// GetModuleInfo 获取模块信息
func (mh *ModuleHelper) GetModuleInfo(name string) (*ModuleInfo, error) {
	module, err := mh.GetModule(name)
	if err != nil {
		return &ModuleInfo{
			Name:   name,
			Loaded: false,
			Error:  err.Error(),
		}, nil
	}

	typ := reflect.TypeOf(module)
	
	return &ModuleInfo{
		Name:     name,
		Type:     typ.String(),
		Instance: module,
		Loaded:   true,
	}, nil
}

// GetAllModuleInfo 获取所有模块信息
func (mh *ModuleHelper) GetAllModuleInfo() []*ModuleInfo {
	modules := mh.GetLoadedModules()
	infos := make([]*ModuleInfo, 0, len(modules))

	for name := range modules {
		info, _ := mh.GetModuleInfo(name)
		infos = append(infos, info)
	}

	return infos
}

// ValidateModule 验证模块
func (mh *ModuleHelper) ValidateModule(module interface{}) error {
	if module == nil {
		return fmt.Errorf("module cannot be nil")
	}

	// 检查模块是否为有效的Go对象
	val := reflect.ValueOf(module)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return fmt.Errorf("module pointer is nil")
	}

	// 可以添加更多验证逻辑
	return nil
}

// CallModuleMethod 调用模块方法
func (mh *ModuleHelper) CallModuleMethod(moduleName, methodName string, args ...interface{}) (interface{}, error) {
	module, err := mh.GetModule(moduleName)
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(module)
	method := val.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	// 准备参数
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// 调用方法
	results := method.Call(in)
	if len(results) == 0 {
		return nil, nil
	}

	return results[0].Interface(), nil
}

// GetModuleMethods 获取模块方法列表
func (mh *ModuleHelper) GetModuleMethods(moduleName string) ([]string, error) {
	module, err := mh.GetModule(moduleName)
	if err != nil {
		return nil, err
	}

	typ := reflect.TypeOf(module)
	methods := make([]string, 0)

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if !strings.HasPrefix(method.Name, "_") {
			methods = append(methods, method.Name)
		}
	}

	return methods, nil
}

// GetModuleFields 获取模块字段列表
func (mh *ModuleHelper) GetModuleFields(moduleName string) ([]string, error) {
	module, err := mh.GetModule(moduleName)
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(module)
	typ := val.Type()

	fields := make([]string, 0)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !strings.HasPrefix(field.Name, "_") {
			fields = append(fields, field.Name)
		}
	}

	return fields, nil
}

// GetModuleCount 获取模块数量
func (mh *ModuleHelper) GetModuleCount() int {
	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	return len(mh.loadedModules)
}

// IsModuleLoaded 检查模块是否已加载
func (mh *ModuleHelper) IsModuleLoaded(name string) bool {
	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	_, exists := mh.loadedModules[name]
	return exists
}

// ExportModules 导出模块配置
func (mh *ModuleHelper) ExportModules() map[string]interface{} {
	modules := mh.GetLoadedModules()
	export := make(map[string]interface{})

	for name, module := range modules {
		typ := reflect.TypeOf(module)
		export[name] = map[string]interface{}{
			"type": typ.String(),
			"kind": typ.Kind().String(),
		}
	}

	return export
}

// ScanDirectory 扫描目录中的模块
func (mh *ModuleHelper) ScanDirectory(dirPath string) ([]string, error) {
	if dirPath == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	// 查找所有.so文件
	pattern := filepath.Join(dirPath, "*.so")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %v", err)
	}

	return files, nil
}

// HotReload 热重载模块
func (mh *ModuleHelper) HotReload(packagePath string, filterFunc FilterFuncType) error {
	// 清空当前模块
	mh.ClearModules()

	// 重新加载
	_, err := mh.Load(packagePath, filterFunc)
	return err
}