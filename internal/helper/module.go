package helper

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"moviepilot-go/internal/logger"
)

// FilterFuncType 过滤函数类型定义
type FilterFuncType func(name string, obj interface{}) bool

// defaultFilter 默认过滤�?func defaultFilter(name string, obj interface{}) bool {
	/*
	 * 默认过滤�?	 */
	return name != "" && obj != nil
}

// ModuleHelper 模块动态加载帮助类
type ModuleHelper struct{}

// NewModuleHelper 创建模块帮助类实�?func NewModuleHelper() *ModuleHelper {
	return &ModuleHelper{}
}

// Load 导入模块
func (mh *ModuleHelper) Load(packagePath string, filterFunc FilterFuncType) []interface{} {
	/*
	 * 导入模块
	 * :param packagePath: 父包�?	 * :param filterFunc: 子模块过滤函数，入参为模块名和模块对象，返回True则导入，否则不导�?	 * :return: 导入的模块对象列�?	 */

	if filterFunc == nil {
		filterFunc = defaultFilter
	}

	submodules := make([]interface{}, 0)
	loadedModules := make(map[string]bool)

	// 在Go中，我们通过解析源代码文件来模拟模块加载
	// 这里需要根据packagePath找到对应的目�?	// 由于Go的插件系统限制，我们采用解析源码的方式实�?	
	// TODO: 实际实现需要根据项目结构确定目录路�?	// 这里简化处理，假设packagePath对应的是文件系统中的目录路径
	dirPath := filepath.Join("internal", strings.ReplaceAll(packagePath, ".", string(filepath.Separator)))
	
	// 解析目录中的所有Go文件
	files, err := filepath.Glob(filepath.Join(dirPath, "*.go"))
	if err != nil {
		logger.GetLoggerManager().Debugf("加载模块 %s 失败�?v", packagePath, err)
		return submodules
	}

	for _, file := range files {
		// 跳过以_开头的文件
		if strings.HasPrefix(filepath.Base(file), "_") {
			continue
		}
		
		// 解析Go源文�?		objs, err := mh.parseGoFile(file)
		if err != nil {
			logger.GetLoggerManager().Debugf("解析文件 %s 失败�?v", file, err)
			continue
		}

		// 应用过滤�?		for name, obj := range objs {
			// 跳过以_开头的名称
			if strings.HasPrefix(name, "_") {
				continue
			}
			
			// 检查对象是否为类型
			if reflect.TypeOf(obj) != nil {
				objType := reflect.TypeOf(obj)
				if objType.Kind() == reflect.Ptr || objType.Kind() == reflect.Struct || objType.Kind() == reflect.Interface {
					if filterFunc(name, obj) {
						if _, exists := loadedModules[name]; !exists {
							loadedModules[name] = true
							submodules = append(submodules, obj)
						}
					}
				}
			}
		}
	}

	return submodules
}

// LoadWithPreFilter 导入子模块（带预过滤�?func (mh *ModuleHelper) LoadWithPreFilter(packagePath string, filterFunc FilterFuncType) []interface{} {
	/*
	 * 导入子模�?	 * :param packagePath: 父包�?	 * :param filterFunc: 子模块过滤函数，入参为模块名和模块对象，返回True则导入，否则不导�?	 * :return: 导入的模块对象列�?	 */

	if filterFunc == nil {
		filterFunc = defaultFilter
	}

	submodules := make([]interface{}, 0)
	
	// TODO: 实际实现需要根据项目结构确定目录路�?	dirPath := filepath.Join("internal", strings.ReplaceAll(packagePath, ".", string(filepath.Separator)))
	
	// 解析目录中的所有Go文件
	files, err := filepath.Glob(filepath.Join(dirPath, "*.go"))
	if err != nil {
		logger.GetLoggerManager().Debugf("加载模块 %s 失败�?v", packagePath, err)
		return submodules
	}

	for _, file := range files {
		// 跳过以_开头的文件
		if strings.HasPrefix(filepath.Base(file), "_") {
			continue
		}
		
		// 预检查模块中的对�?		candidates, err := mh.parseGoFile(file)
		if err != nil {
			logger.GetLoggerManager().Debugf("解析文件 %s 失败�?v", file, err)
			continue
		}
		
		// 确定是否需要处理此文件
		shouldProcess := false
		for name, obj := range candidates {
			if !strings.HasPrefix(name, "_") {
				objType := reflect.TypeOf(obj)
				if objType != nil && (objType.Kind() == reflect.Ptr || objType.Kind() == reflect.Struct || objType.Kind() == reflect.Interface) {
					if filterFunc(name, obj) {
						shouldProcess = true
						break
					}
				}
			}
		}
		
		if shouldProcess {
			// 重新加载模块对象
			objs, err := mh.reloadModuleObjects(file)
			if err != nil {
				logger.GetLoggerManager().Debugf("重新加载模块对象 %s 失败�?v", file, err)
				continue
			}
			
			// 应用过滤�?			for name, obj := range objs {
				if !strings.HasPrefix(name, "_") {
					objType := reflect.TypeOf(obj)
					if objType != nil && (objType.Kind() == reflect.Ptr || objType.Kind() == reflect.Struct || objType.Kind() == reflect.Interface) {
						if filterFunc(name, obj) {
							submodules = append(submodules, obj)
						}
					}
				}
			}
		}
	}

	return submodules
}

// DynamicImportAllModules 动态导入目录下所有模�?func (mh *ModuleHelper) DynamicImportAllModules(basePath string, packageName string) []string {
	/*
	 * 动态导入目录下所有模�?	 */
	modules := make([]string, 0)
	
	// 查找目录下所有Go文件
	pattern := filepath.Join(basePath, "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		logger.GetLoggerManager().Debugf("查找模块文件失败�?v", err)
		return modules
	}
	
	for _, file := range files {
		fileName := strings.TrimSuffix(filepath.Base(file), ".go")
		if fileName != "__init__" {
			modules = append(modules, fileName)
			// 在Go中，我们不能像Python那样动态导入，但可以记录模块名
			fullModuleName := packageName + "." + fileName
			logger.GetLoggerManager().Debugf("找到模块: %s", fullModuleName)
		}
	}
	
	return modules
}

// parseGoFile 解析Go源文件，提取类型定义
func (mh *ModuleHelper) parseGoFile(filePath string) (map[string]interface{}, error) {
	objs := make(map[string]interface{})
	
	// 创建token文件�?	fset := token.NewFileSet()
	
	// 解析Go源文�?	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return objs, err
	}
	
	// 遍历文件中的声明
	for _, decl := range f.Decls {
		// 处理GenDecl（通用声明，如类型、常量、变量声明）
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				// 处理类型声明
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					name := typeSpec.Name.Name
					// 在Go中，我们无法直接创建类型实例，这里返回类型信�?					objs[name] = typeSpec
				}
			}
		}
		
		// 处理函数声明
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.IsExported() { // 只处理导出的函数
				name := funcDecl.Name.Name
				// 在Go中，函数是一等公民，我们可以返回函数本身的信�?				objs[name] = funcDecl
			}
		}
	}
	
	return objs, nil
}

// reloadModuleObjects 重新加载模块对象
func (mh *ModuleHelper) reloadModuleObjects(filePath string) (map[string]interface{}, error) {
	// 简化实现，直接调用parseGoFile
	return mh.parseGoFile(filePath)
}

// loadPlugin 加载插件（Go插件系统�?func (mh *ModuleHelper) loadPlugin(pluginPath string) (*plugin.Plugin, error) {
	/*
	 * 加载Go插件
	 * :param pluginPath: 插件路径
	 * :return: 插件对象
	 */
	return plugin.Open(pluginPath)
}

// getPluginSymbol 获取插件符号
func (mh *ModuleHelper) getPluginSymbol(p *plugin.Plugin, symbolName string) (plugin.Symbol, error) {
	/*
	 * 获取插件符号
	 * :param p: 插件对象
	 * :param symbolName: 符号名称
	 * :return: 符号对象
	 */
	return p.Lookup(symbolName)
}

// LoadPlugins 加载目录中的所有插�?func (mh *ModuleHelper) LoadPlugins(pluginDir string) []*plugin.Plugin {
	/*
	 * 加载目录中的所有插�?	 * :param pluginDir: 插件目录
	 * :return: 插件对象列表
	 */
	plugins := make([]*plugin.Plugin, 0)
	
	// 查找目录下所有插件文�?	pattern := filepath.Join(pluginDir, "*.so")
	files, err := filepath.Glob(pattern)
	if err != nil {
		logger.GetLoggerManager().Debugf("查找插件文件失败�?v", err)
		return plugins
	}
	
	for _, file := range files {
		// 加载插件
		p, err := mh.loadPlugin(file)
		if err != nil {
			logger.GetLoggerManager().Debugf("加载插件 %s 失败�?v", file, err)
			continue
		}
		
		plugins = append(plugins, p)
	}
	
	return plugins
}

// GetPluginInstance 获取插件实例
func (mh *ModuleHelper) GetPluginInstance(p *plugin.Plugin, instanceName string) (interface{}, error) {
	/*
	 * 获取插件实例
	 * :param p: 插件对象
	 * :param instanceName: 实例名称
	 * :return: 实例对象
	 */
	symbol, err := p.Lookup(instanceName)
	if err != nil {
		return nil, err
	}
	return symbol, nil
}

// GetPluginInfo 获取插件信息
func (mh *ModuleHelper) GetPluginInfo(p *plugin.Plugin) (map[string]interface{}, error) {
	/*
	 * 获取插件信息
	 * :param p: 插件对象
	 * :return: 插件信息
	 */
	info := make(map[string]interface{})
	
	// 获取插件名称
	if nameSymbol, err := p.Lookup("Name"); err == nil {
		if name, ok := nameSymbol.(*string); ok {
			info["name"] = *name
		}
	}
	
	// 获取插件版本
	if versionSymbol, err := p.Lookup("Version"); err == nil {
		if version, ok := versionSymbol.(*string); ok {
			info["version"] = *version
		}
	}
	
	// 获取插件描述
	if descSymbol, err := p.Lookup("Description"); err == nil {
		if desc, ok := descSymbol.(*string); ok {
			info["description"] = *desc
		}
	}
	
	return info, nil
}
