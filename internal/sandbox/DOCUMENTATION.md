# 插件系统设计文档

## 概述

MoviePilot的插件系统支持两种类型的插件：

1. **Go插件** - 编译为.so文件的原生Go插件
2. **Python插件** - 作为无状态子服务在独立Docker容器中运行的Python插件

## 系统架构

### 核心组件

1. **PluginManager** - 插件管理器，负责插件的加载、启用、禁用和协调
2. **PluginLoader** - 插件加载器，负责从文件系统加载Go插件
3. **PluginService** - 插件服务，负责管理Python插件容器
4. **PluginBase** - 插件基类，定义了插件的标准接口
5. **PluginChain** - 插件处理链，提供插件间通信机制

### Python插件沙箱

Python插件在独立的Docker容器中运行，通过HTTP/JSON与主服务通信。

#### 工作流程

1. 用户将Python插件放置在`plugins/`目录下
2. 主服务通过`PluginService`加载插件
3. `PluginService`启动一个新的Docker容器运行该插件
4. 插件在容器中作为HTTP服务运行
5. 主服务通过HTTP请求与插件通信

#### 容器要求

Python插件容器需要：
- 监听本地端口（默认8080）
- 提供`/execute`端点处理命令
- 通过标准输出返回结果

## Go插件开发

### 基本结构

Go插件需要实现`plugins.Plugin`接口：

```go
type Plugin interface {
    // 基本方法
    InitPlugin(config map[string]interface{}) error
    GetName() string
    GetDesc() string
    GetOrder() int
    GetState() bool
    StopService()
    
    // 配置和数据管理
    UpdateConfig(config map[string]interface{}, pluginID *string) bool
    GetConfig(pluginID *string) interface{}
    GetDataPath(pluginID *string) string
    SaveData(key string, value interface{}, pluginID *string)
    GetData(key *string, pluginID *string) interface{}
    DelData(key string, pluginID *string) interface{}
    
    // UI和API
    GetCommand() []map[string]interface{}
    GetRenderMode() (string, *string)
    GetAPI() []map[string]interface{}
    GetForm() ([]map[string]interface{}, map[string]interface{})
    GetPage() []map[string]interface{}
    GetService() []map[string]interface{}
    GetDashboard(key string, kwargs map[string]interface{}) (*map[string]interface{}, *map[string]interface{}, *[]map[string]interface{})
    GetDashboardMeta() []map[string]string
    GetModule() map[string]interface{}
    GetActions() []map[string]interface{}
    
    // 消息和通信
    PostMessage(channel *MessageChannel, mtype *NotificationType,
        title, text, image, link, userid, username *string, kwargs map[string]interface{})
    
    // 资源管理
    Close()
}
```

### 示例插件

```go
package main

import (
    "github.com/moviepilot/moviepilot-go/pkg/plugins"
)

type ExamplePlugin struct {
    *plugins.PluginBase
}

func NewPlugin() plugins.Plugin {
    plugin := &ExamplePlugin{
        PluginBase: plugins.NewPluginBase(),
    }
    plugin.PluginName = "Example Plugin"
    plugin.PluginDesc = "这是一个示例插件"
    plugin.PluginOrder = 100
    return plugin
}

func (p *ExamplePlugin) InitPlugin(config map[string]interface{}) error {
    // 初始化插件逻辑
    return nil
}

func (p *ExamplePlugin) GetState() bool {
    // 返回插件状态
    return true
}

func (p *ExamplePlugin) StopService() {
    // 停止插件服务
}

// 实现其他必需的方法...

func main() {}
```

## Python插件开发

### 基本结构

Python插件需要包含以下文件：
- `plugin.py` - 插件主文件
- `manifest.yaml` - 插件元数据
- `requirements.txt` - 依赖列表

### plugin.py 示例

```python
from app.plugins import _PluginBase
from app.schemas import EventType

class ExamplePlugin(_PluginBase):
    # 插件名称
    plugin_name = "示例插件"
    # 插件描述
    plugin_desc = "这是一个示例插件"
    # 插件顺序
    plugin_order = 100

    def init_plugin(self, config: dict = None):
        # 初始化插件
        pass

    def get_state(self) -> bool:
        # 返回插件状态
        return True

    def get_command(self) -> List[Dict[str, Any]]:
        # 注册插件命令
        return [{
            "cmd": "/example",
            "event": EventType.Example,
            "desc": "示例命令",
            "category": "示例"
        }]

    def get_api(self) -> List[Dict[str, Any]]:
        # 注册插件API
        return [{
            "path": "/example",
            "endpoint": self.example_api,
            "methods": ["GET"],
            "summary": "示例API"
        }]

    def stop_service(self):
        # 停止插件服务
        pass

    def example_api(self):
        # API处理函数
        return {"message": "Hello from example plugin"}
```

## API端点

### 插件管理端点

- `POST /api/v1/plugin/load` - 加载插件
- `POST /api/v1/plugin/execute` - 执行插件
- `POST /api/v1/plugin/unload` - 卸载插件
- `GET /api/v1/plugin/list` - 列出插件

### 请求格式

```json
{
  "plugin_path": "example_plugin",
  "action": "execute",
  "params": {
    "key": "value"
  }
}
```

### 响应格式

```json
{
  "success": true,
  "data": {},
  "message": "Success message"
}
```