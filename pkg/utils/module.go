package utils

import (
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"sync"

	"go.uber.org/zap"
	
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/errors"
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
	logger.Debug("Creating new ModuleHelper instance", zap.String("func", "NewModuleHelper"))
	return &ModuleHelper{
		loadedModules: make(map[string]interface{}),
	}
}

// Load 加载模块
func (mh *ModuleHelper) Load(packagePath string, filterFunc FilterFuncType) ([]interface{}, error) {
	if packagePath == "" {
		err := errors.NewAppError(400, "Package path cannot be empty", "")
		logger.Error("Invalid package path for module loading", 
			zap.String("error", err.Error()),
			zap.String("func", "Load"))
		return nil, err
	}

	logger.Debug("Loading modules from package", 
		zap.String("package_path", packagePath),
		zap.String("func", "Load"))

	if filterFunc == nil {
		filterFunc = mh.defaultFilter
		logger.Debug("Using default filter for module loading", zap.String("func", "Load"))
	}

	var modules []interface{}
	loadedNames := make(map[string]bool)

	// 在Go中，动态加载模块比Python复杂
	// 这里提供一个简化的实现，主要适用于插件系统
	
	// 扫描目录中的.so文件（Go插件）
	pluginFiles, err := mh.findPluginFiles(packagePath)
	if err != nil {
		logger.Error("Failed to find plugin files", 
			zap.String("package_path", packagePath),
			zap.String("error", err.Error()),
			zap.String("func", "Load"))
		return nil, errors.NewAppError(500, "Failed to find plugin files", err.Error())
	}

	logger.Debug("Found plugin files", 
		zap.String("package_path", packagePath),
		zap.Strings("plugin_files", pluginFiles),
		zap.String("func", "Load"))

	for _, pluginFile := range pluginFiles {
		module, err := mh.loadPlugin(pluginFile, filterFunc, loadedNames)
		if err != nil {
			// 记录错误但继续处理其他模块
			logger.Warn("Failed to load plugin, continuing with others", 
				zap.String("plugin_file", pluginFile),
				zap.String("error", err.Error()),
				zap.String("func", "Load"))
			continue
		}

		if module != nil {
			modules = append(modules, module)
			logger.Debug("Successfully loaded module", 
				zap.String("plugin_file", pluginFile),
				zap.String("func", "Load"))
		}
	}

	logger.Info("Module loading completed", 
		zap.String("package_path", packagePath),
		zap.Int("loaded_count", len(modules)),
		zap.String("func", "Load"))

	return modules, nil
}

// LoadPlugin 加载单个插件
func (mh *ModuleHelper) LoadPlugin(pluginPath string) (interface{}, error) {
	if pluginPath == "" {
		err := errors.NewAppError(400, "Plugin path cannot be empty", "")
		logger.Error("Invalid plugin path", 
			zap.String("error", err.Error()),
			zap.String("func", "LoadPlugin"))
		return nil, err
	}

	logger.Debug("Loading single plugin", 
		zap.String("plugin_path", pluginPath),
		zap.String("func", "LoadPlugin"))

	// 加载Go插件
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		logger.Error("Failed to open plugin", 
			zap.String("plugin_path", pluginPath),
			zap.String("error", err.Error()),
			zap.String("func", "LoadPlugin"))
		return nil, errors.NewAppError(500, "Failed to open plugin", err.Error())
	}

	// 查找导出的符号
	symbols, err := mh.findPluginSymbols(plug)
	if err != nil {
		logger.Error("Failed to find plugin symbols", 
			zap.String("plugin_path", pluginPath),
			zap.String("error", err.Error()),
			zap.String("func", "LoadPlugin"))
		return nil, errors.NewAppError(500, "Failed to find symbols", err.Error())
	}

	if len(symbols) == 0 {
		err := errors.NewAppError(404, "No symbols found in plugin", pluginPath)
		logger.Error("No symbols found in plugin", 
			zap.String("plugin_path", pluginPath),
			zap.String("error", err.Error()),
			zap.String("func", "LoadPlugin"))
		return nil, err
	}

	logger.Info("Successfully loaded plugin", 
		zap.String("plugin_path", pluginPath),
		zap.Int("symbol_count", len(symbols)),
		zap.String("func", "LoadPlugin"))

	// 返回第一个找到的符号
	for _, symbol := range symbols {
		return symbol, nil
	}
	return nil, nil
}

// findPluginFiles 查找插件文件
func (mh *ModuleHelper) findPluginFiles(packagePath string) ([]string, error) {
	// 在Go中，插件文件通常是.so文件
	pattern := filepath.Join(packagePath, "*.so")
	
	logger.Debug("Searching for plugin files", 
		zap.String("package_path", packagePath),
		zap.String("pattern", pattern),
		zap.String("func", "findPluginFiles"))
	
	matches, err := filepath.Glob(pattern)
	if err != nil {
		logger.Error("Failed to glob plugin files", 
			zap.String("package_path", packagePath),
			zap.String("pattern", pattern),
			zap.String("error", err.Error()),
			zap.String("func", "findPluginFiles"))
		return nil, errors.NewAppError(500, "Failed to glob plugin files", err.Error())
	}

	logger.Debug("Found plugin files", 
		zap.String("package_path", packagePath),
		zap.Strings("files", matches),
		zap.Int("count", len(matches)),
		zap.String("func", "findPluginFiles"))

	return matches, nil
}

// loadPlugin 加载插件
func (mh *ModuleHelper) loadPlugin(pluginPath string, filterFunc FilterFuncType, loadedNames map[string]bool) (interface{}, error) {
	logger.Debug("Loading plugin", 
		zap.String("plugin_path", pluginPath),
		zap.String("func", "loadPlugin"))
	
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		logger.Error("Failed to open plugin", 
			zap.String("plugin_path", pluginPath),
			zap.String("error", err.Error()),
			zap.String("func", "loadPlugin"))
		return nil, errors.NewAppError(500, "Failed to open plugin", err.Error())
	}

	symbols, err := mh.findPluginSymbols(plug)
	if err != nil {
		logger.Error("Failed to find symbols in plugin", 
			zap.String("plugin_path", pluginPath),
			zap.String("error", err.Error()),
			zap.String("func", "loadPlugin"))
		return nil, errors.NewAppError(500, "Failed to find symbols", err.Error())
	}

	logger.Debug("Found symbols in plugin", 
		zap.String("plugin_path", pluginPath),
		zap.Int("symbol_count", len(symbols)),
		zap.String("func", "loadPlugin"))

	for name, symbol := range symbols {
		// 跳过私有符号
		if strings.HasPrefix(name, "_") {
			logger.Debug("Skipping private symbol", 
				zap.String("symbol_name", name),
				zap.String("plugin_path", pluginPath),
				zap.String("func", "loadPlugin"))
			continue
		}

		// 检查是否已加载
		if loadedNames[name] {
			logger.Debug("Symbol already loaded, skipping", 
				zap.String("symbol_name", name),
				zap.String("plugin_path", pluginPath),
				zap.String("func", "loadPlugin"))
			continue
		}

		// 应用过滤器
		if filterFunc(name, symbol) {
			loadedNames[name] = true
			
			// 缓存模块
			mh.moduleMutex.Lock()
			mh.loadedModules[name] = symbol
			mh.moduleMutex.Unlock()

			logger.Info("Successfully loaded and cached module", 
				zap.String("symbol_name", name),
				zap.String("plugin_path", pluginPath),
				zap.String("func", "loadPlugin"))

			return symbol, nil
		}
	}

	logger.Debug("No matching symbols found in plugin", 
		zap.String("plugin_path", pluginPath),
		zap.String("func", "loadPlugin"))

	return nil, nil
}

// findPluginSymbols 查找插件符号
func (mh *ModuleHelper) findPluginSymbols(plug *plugin.Plugin) (map[string]interface{}, error) {
	logger.Debug("Finding plugin symbols", zap.String("func", "findPluginSymbols"))
	
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
			logger.Debug("Found plugin symbol", 
				zap.String("symbol_name", symbolName),
				zap.String("func", "findPluginSymbols"))
		} else {
			logger.Debug("Symbol not found in plugin", 
				zap.String("symbol_name", symbolName),
				zap.String("error", err.Error()),
				zap.String("func", "findPluginSymbols"))
		}
	}

	logger.Debug("Plugin symbol search completed", 
		zap.Int("found_count", len(symbols)),
		zap.Strings("symbols", mh.getSymbolNames(symbols)),
		zap.String("func", "findPluginSymbols"))

	return symbols, nil
}

// getSymbolNames 获取符号名称列表
func (mh *ModuleHelper) getSymbolNames(symbols map[string]interface{}) []string {
	names := make([]string, 0, len(symbols))
	for name := range symbols {
		names = append(names, name)
	}
	return names
}

// defaultFilter 默认过滤器
func (mh *ModuleHelper) defaultFilter(name string, obj interface{}) bool {
	return name != "" && obj != nil
}

// GetLoadedModules 获取已加载的模块
func (mh *ModuleHelper) GetLoadedModules() map[string]interface{} {
	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	logger.Debug("Getting loaded modules", 
		zap.Int("module_count", len(mh.loadedModules)),
		zap.String("func", "GetLoadedModules"))

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
		err := errors.NewAppError(400, "Module name cannot be empty", "")
		logger.Error("Invalid module name", 
			zap.String("error", err.Error()),
			zap.String("func", "GetModule"))
		return nil, err
	}

	logger.Debug("Getting module", 
		zap.String("module_name", name),
		zap.String("func", "GetModule"))

	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	module, exists := mh.loadedModules[name]
	if !exists {
		err := errors.NewAppError(404, "Module not found", name)
		logger.Error("Module not found", 
			zap.String("module_name", name),
			zap.String("error", err.Error()),
			zap.String("func", "GetModule"))
		return nil, err
	}

	return module, nil
}

// UnloadModule 卸载模块
func (mh *ModuleHelper) UnloadModule(name string) error {
	if name == "" {
		err := errors.NewAppError(400, "Module name cannot be empty", "")
		logger.Error("Invalid module name for unload", 
			zap.String("error", err.Error()),
			zap.String("func", "UnloadModule"))
		return err
	}

	logger.Debug("Unloading module", 
		zap.String("module_name", name),
		zap.String("func", "UnloadModule"))

	mh.moduleMutex.Lock()
	defer mh.moduleMutex.Unlock()

	if _, exists := mh.loadedModules[name]; !exists {
		err := errors.NewAppError(404, "Module not found", name)
		logger.Error("Module not found for unload", 
			zap.String("module_name", name),
			zap.String("error", err.Error()),
			zap.String("func", "UnloadModule"))
		return err
	}

	delete(mh.loadedModules, name)
	logger.Info("Successfully unloaded module", 
		zap.String("module_name", name),
		zap.String("func", "UnloadModule"))
	
	return nil
}

// ClearModules 清空所有模块
func (mh *ModuleHelper) ClearModules() {
	mh.moduleMutex.Lock()
	defer mh.moduleMutex.Unlock()

	moduleCount := len(mh.loadedModules)
	mh.loadedModules = make(map[string]interface{})
	
	logger.Info("Cleared all modules", 
		zap.Int("cleared_count", moduleCount),
		zap.String("func", "ClearModules"))
}

// ReloadModule 重新加载模块
func (mh *ModuleHelper) ReloadModule(name string, packagePath string, filterFunc FilterFuncType) error {
	if name == "" {
		err := errors.NewAppError(400, "Module name cannot be empty", "")
		logger.Error("Invalid module name for reload", 
			zap.String("error", err.Error()),
			zap.String("func", "ReloadModule"))
		return err
	}

	if packagePath == "" {
		err := errors.NewAppError(400, "Package path cannot be empty", "")
		logger.Error("Invalid package path for reload", 
			zap.String("module_name", name),
			zap.String("error", err.Error()),
			zap.String("func", "ReloadModule"))
		return err
	}

	logger.Info("Reloading module", 
		zap.String("module_name", name),
		zap.String("package_path", packagePath),
		zap.String("func", "ReloadModule"))

	var err error
	err = mh.UnloadModule(name)
	if err != nil {
		logger.Error("Failed to unload module for reload", 
			zap.String("module_name", name),
			zap.String("error", err.Error()),
			zap.String("func", "ReloadModule"))
		return err
	}

	var modules []interface{}
	modules, err = mh.Load(packagePath, filterFunc)
	if err != nil {
		logger.Error("Failed to load modules for reload", 
			zap.String("module_name", name),
			zap.String("package_path", packagePath),
			zap.String("error", err.Error()),
			zap.String("func", "ReloadModule"))
		return err
	}

	// 检查是否重新加载成功
	for _, module := range modules {
		if mh.isModuleNamed(module, name) {
			logger.Info("Successfully reloaded module", 
				zap.String("module_name", name),
				zap.String("func", "ReloadModule"))
			return nil
		}
	}

	reloadErr := errors.NewAppError(500, "Failed to reload module", name)
	logger.Error("Failed to reload module - module not found after reload", 
		zap.String("module_name", name),
		zap.String("package_path", packagePath),
		zap.String("error", reloadErr.Error()),
		zap.String("func", "ReloadModule"))
	return reloadErr
}

// isModuleNamed 检查模块是否具有指定名称
func (mh *ModuleHelper) isModuleNamed(module interface{}, name string) bool {
	if module == nil {
		return false
	}

	// 通过反射获取模块信息
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
	logger.Debug("Getting module info", 
		zap.String("module_name", name),
		zap.String("func", "GetModuleInfo"))

	module, err := mh.GetModule(name)
	if err != nil {
		logger.Debug("Module not found for info", 
			zap.String("module_name", name),
			zap.String("error", err.Error()),
			zap.String("func", "GetModuleInfo"))
		return &ModuleInfo{
			Name:   name,
			Loaded: false,
			Error:  err.Error(),
		}, nil
	}

	typ := reflect.TypeOf(module)
	
	info := &ModuleInfo{
		Name:     name,
		Type:     typ.String(),
		Instance: module,
		Loaded:   true,
	}

	logger.Debug("Retrieved module info", 
		zap.String("module_name", name),
		zap.String("module_type", info.Type),
		zap.Bool("loaded", info.Loaded),
		zap.String("func", "GetModuleInfo"))

	return info, nil
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
	logger.Debug("Validating module", zap.String("func", "ValidateModule"))

	if module == nil {
		err := errors.NewAppError(400, "Module cannot be nil", "")
		logger.Error("Module validation failed - nil module", 
			zap.String("error", err.Error()),
			zap.String("func", "ValidateModule"))
		return err
	}

	// 检查模块是否为有效的Go对象
	val := reflect.ValueOf(module)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		err := errors.NewAppError(400, "Module pointer is nil", "")
		logger.Error("Module validation failed - nil pointer", 
			zap.String("error", err.Error()),
			zap.String("func", "ValidateModule"))
		return err
	}

	logger.Debug("Module validation passed", 
		zap.String("module_type", val.Type().String()),
		zap.String("func", "ValidateModule"))

	// 可以添加更多验证逻辑
	return nil
}

// CallModuleMethod 调用模块方法
func (mh *ModuleHelper) CallModuleMethod(moduleName, methodName string, args ...interface{}) (interface{}, error) {
	if methodName == "" {
		err := errors.NewAppError(400, "Method name cannot be empty", "")
		logger.Error("Invalid method name", 
			zap.String("module_name", moduleName),
			zap.String("error", err.Error()),
			zap.String("func", "CallModuleMethod"))
		return nil, err
	}

	logger.Debug("Calling module method", 
		zap.String("module_name", moduleName),
		zap.String("method_name", methodName),
		zap.Int("arg_count", len(args)),
		zap.String("func", "CallModuleMethod"))

	module, err := mh.GetModule(moduleName)
	if err != nil {
		logger.Error("Failed to get module for method call", 
			zap.String("module_name", moduleName),
			zap.String("method_name", methodName),
			zap.String("error", err.Error()),
			zap.String("func", "CallModuleMethod"))
		return nil, err
	}

	val := reflect.ValueOf(module)
	method := val.MethodByName(methodName)
	if !method.IsValid() {
		err := errors.NewAppError(404, "Method not found", methodName)
		logger.Error("Method not found in module", 
			zap.String("module_name", moduleName),
			zap.String("method_name", methodName),
			zap.String("error", err.Error()),
			zap.String("func", "CallModuleMethod"))
		return nil, err
	}

	// 准备参数
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// 调用方法
	results := method.Call(in)
	if len(results) == 0 {
		logger.Debug("Method call completed with no return value", 
			zap.String("module_name", moduleName),
			zap.String("method_name", methodName),
			zap.String("func", "CallModuleMethod"))
		return nil, nil
	}

	logger.Debug("Method call completed successfully", 
		zap.String("module_name", moduleName),
		zap.String("method_name", methodName),
		zap.Int("result_count", len(results)),
		zap.String("func", "CallModuleMethod"))

	return results[0].Interface(), nil
}

// GetModuleMethods 获取模块方法列表
func (mh *ModuleHelper) GetModuleMethods(moduleName string) ([]string, error) {
	logger.Debug("Getting module methods", 
		zap.String("module_name", moduleName),
		zap.String("func", "GetModuleMethods"))

	module, err := mh.GetModule(moduleName)
	if err != nil {
		logger.Error("Failed to get module for methods", 
			zap.String("module_name", moduleName),
			zap.String("error", err.Error()),
			zap.String("func", "GetModuleMethods"))
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

	logger.Debug("Retrieved module methods", 
		zap.String("module_name", moduleName),
		zap.Strings("methods", methods),
		zap.Int("method_count", len(methods)),
		zap.String("func", "GetModuleMethods"))

	return methods, nil
}

// GetModuleFields 获取模块字段列表
func (mh *ModuleHelper) GetModuleFields(moduleName string) ([]string, error) {
	logger.Debug("Getting module fields", 
		zap.String("module_name", moduleName),
		zap.String("func", "GetModuleFields"))

	module, err := mh.GetModule(moduleName)
	if err != nil {
		logger.Error("Failed to get module for fields", 
			zap.String("module_name", moduleName),
			zap.String("error", err.Error()),
			zap.String("func", "GetModuleFields"))
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

	logger.Debug("Retrieved module fields", 
		zap.String("module_name", moduleName),
		zap.Strings("fields", fields),
		zap.Int("field_count", len(fields)),
		zap.String("func", "GetModuleFields"))

	return fields, nil
}

// GetModuleCount 获取模块数量
func (mh *ModuleHelper) GetModuleCount() int {
	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	count := len(mh.loadedModules)
	logger.Debug("Getting module count", 
		zap.Int("count", count),
		zap.String("func", "GetModuleCount"))

	return count
}

// IsModuleLoaded 检查模块是否已加载
func (mh *ModuleHelper) IsModuleLoaded(name string) bool {
	mh.moduleMutex.RLock()
	defer mh.moduleMutex.RUnlock()

	_, exists := mh.loadedModules[name]
	
	logger.Debug("Checking if module is loaded", 
		zap.String("module_name", name),
		zap.Bool("loaded", exists),
		zap.String("func", "IsModuleLoaded"))

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
		err := errors.NewAppError(400, "Directory path cannot be empty", "")
		logger.Error("Invalid directory path for scanning", 
			zap.String("error", err.Error()),
			zap.String("func", "ScanDirectory"))
		return nil, err
	}

	logger.Debug("Scanning directory for modules", 
		zap.String("directory_path", dirPath),
		zap.String("func", "ScanDirectory"))

	// 查找所有.so文件
	pattern := filepath.Join(dirPath, "*.so")
	files, err := filepath.Glob(pattern)
	if err != nil {
		logger.Error("Failed to scan directory", 
			zap.String("directory_path", dirPath),
			zap.String("pattern", pattern),
			zap.String("error", err.Error()),
			zap.String("func", "ScanDirectory"))
		return nil, errors.NewAppError(500, "Failed to scan directory", err.Error())
	}

	logger.Debug("Directory scan completed", 
		zap.String("directory_path", dirPath),
		zap.Strings("found_files", files),
		zap.Int("file_count", len(files)),
		zap.String("func", "ScanDirectory"))

	return files, nil
}

// HotReload 热重载模块
func (mh *ModuleHelper) HotReload(packagePath string, filterFunc FilterFuncType) error {
	if packagePath == "" {
		err := errors.NewAppError(400, "Package path cannot be empty", "")
		logger.Error("Invalid package path for hot reload", 
			zap.String("error", err.Error()),
			zap.String("func", "HotReload"))
		return err
	}

	logger.Info("Starting hot reload", 
		zap.String("package_path", packagePath),
		zap.String("func", "HotReload"))

	// 清空当前模块
	mh.ClearModules()

	// 重新加载
	modules, err := mh.Load(packagePath, filterFunc)
	if err != nil {
		logger.Error("Hot reload failed", 
			zap.String("package_path", packagePath),
			zap.String("error", err.Error()),
			zap.String("func", "HotReload"))
		return err
	}

	logger.Info("Hot reload completed successfully", 
		zap.String("package_path", packagePath),
		zap.Int("loaded_count", len(modules)),
		zap.String("func", "HotReload"))

	return nil
}