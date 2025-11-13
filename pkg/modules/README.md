# 模块系统

## 概述

模块系统是 MoviePilot 的核心组件之一，负责管理系统中的各种功能模块，包括下载器、媒体服务器、消息通道等。

## 核心组件

### ModuleBase (模块基类)
定义了所有模块必须实现的基本接口和方法：

- `InitModule()` - 模块初始化
- `InitSetting()` - 模块开关设置
- `GetName()` - 获取模块名称
- `GetType()` - 获取模块类型
- `GetSubType()` - 获取模块子类型
- `GetPriority()` - 获取模块优先级
- `Stop()` - 停止模块服务
- `Test()` - 模块测试

### ServiceBase (服务基类)
抽象服务基类，负责服务的初始化、获取实例和配置管理：

- `InitService()` - 初始化服务
- `GetInstances()` - 获取服务实例列表
- `GetInstance()` - 获取指定名称的服务实例
- `GetConfigs()` - 获取已启用的服务配置字典
- `GetConfig()` - 获取指定名称的服务配置
- `GetDefaultConfigName()` - 获取默认服务配置的名称

### MessageBase (消息基类)
消息模块的基类，继承自 ServiceBase：

- `CheckMessage()` - 检查消息渠道及消息类型，判断是否处理消息

### DownloaderBase (下载器基类)
下载器模块的基类，继承自 ServiceBase：

- `GetDefaultConfigName()` - 获取默认下载器配置名称
- 特殊处理默认下载器的查找逻辑

### MediaServerBase (媒体服务器基类)
媒体服务器模块的基类，继承自 ServiceBase

## 设计原则

1. **模块化**: 每个功能组件都被设计为独立的模块，可以单独启用或禁用
2. **可扩展**: 通过继承基类可以轻松创建新的模块类型
3. **配置驱动**: 模块的行为通过配置文件进行控制
4. **优先级控制**: 同类型模块可以通过优先级控制执行顺序
5. **测试支持**: 每个模块都支持内置测试功能

## 使用方法

### 创建新模块

要创建一个新的模块，需要：

1. 继承 `ModuleBase` 或相应的子类
2. 实现必要的抽象方法
3. 在系统配置中注册模块

### 示例

```go
type MyModule struct {
    *modules.ModuleBase
}

func NewMyModule() *MyModule {
    module := &MyModule{
        ModuleBase: modules.NewModuleBase(),
    }
    module.Name = "MyModule"
    module.Type = models.ModuleTypeOther
    return module
}

func (m *MyModule) InitModule() error {
    // 初始化逻辑
    return nil
}

func (m *MyModule) Stop() error {
    // 停止逻辑
    return nil
}

// 实现其他必需的方法...
```

## 配置管理

模块系统通过配置提供者获取服务配置，支持动态配置更新和多实例管理。