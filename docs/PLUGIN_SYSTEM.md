# MoviePilot 插件系统

MoviePilot 插件系统是一个灵活的混合插件架构，支持多种类型的插件：

- **Go原生插件**：编译为 `.so` 文件的Go插件
- **Python脚本插件**：通过HTTP/gRPC通信的Python插件
- **WebAssembly插件**：编译为WASM的Web插件

## 架构概述

### 核心组件

1. **HybridPluginManager**：混合插件管理器，统一管理所有类型的插件
2. **NativePluginManager**：Go原生插件管理器
3. **PythonPluginManager**：Python插件管理器
4. **EventBus**：事件总线，用于插件间通信
5. **ConfigManager**：配置管理器，管理插件配置

### 插件生命周期

```
Unloaded → Loaded → Initialized → Running → Stopped → Unloaded
    ↑                                                    ↓
    ←←←←←←←←←←←←←←← Error ←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←
```

## 快速开始

### 1. 创建插件

#### Go原生插件示例

```go
package main

import (
    "fmt"
    "github.com/yfh-yun/moviepilot-go/pkg/plugin"
)

type MyPlugin struct {
    id    string
    name  string
    state plugin.PluginState
}

func NewPlugin() plugin.Plugin {
    return &MyPlugin{
        id:    "my-plugin",
        name:  "My Plugin",
        state: plugin.StateUnloaded,
    }
}

func (p *MyPlugin) ID() string { return p.id }
func (p *MyPlugin) Name() string { return p.name }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Type() plugin.PluginType { return plugin.PluginTypeNative }
func (p *MyPlugin) Description() string { return "My custom plugin" }

func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    p.state = plugin.StateInitialized
    fmt.Println("Plugin initialized")
    return nil
}

func (p *MyPlugin) Start() error {
    p.state = plugin.StateRunning
    fmt.Println("Plugin started")
    return nil
}

func (p *MyPlugin) Stop() error {
    p.state = plugin.StateStopped
    fmt.Println("Plugin stopped")
    return nil
}

func (p *MyPlugin) Destroy() error {
    p.state = plugin.StateUnloaded
    fmt.Println("Plugin destroyed")
    return nil
}

func (p *MyPlugin) GetState() plugin.PluginState {
    return p.state
}

func (p *MyPlugin) HandleEvent(event plugin.Event) error {
    fmt.Printf("Received event: %s\n", event.Type)
    return nil
}

func (p *MyPlugin) GetConfigForm() *plugin.ConfigForm {
    return &plugin.ConfigForm{
        Title:       "My Plugin Configuration",
        Description: "Configure my plugin",
        Fields: []plugin.FormField{
            {
                Name:     "enabled",
                Label:    "Enable",
                Type:     "boolean",
                Required: false,
                Default:  true,
            },
        },
    }
}

func (p *MyPlugin) GetAPIRoutes() []plugin.APIRoute {
    return []plugin.APIRoute{
        {
            Method:  "GET",
            Path:    "/hello",
            Handler: "handleHello",
        },
    }
}

func (p *MyPlugin) GetCommands() []plugin.Command {
    return []plugin.Command{
        {
            Name:        "hello",
            Description: "Say hello",
            Usage:       "hello",
            Handler:     "handleHelloCommand",
        },
    }
}

func (p *MyPlugin) GetServices() []plugin.Service {
    return []plugin.Service{
        {
            Name:        "hello-service",
            Description: "Hello service",
            Handler:     "handleHelloService",
        },
    }
}

func main() {} // 必须有空的main函数
```

### 2. 编译插件

```bash
# 进入插件目录
cd plugins/example

# 编译为.so文件
go build -buildmode=plugin -o my_plugin.so my_plugin.go

# 或使用Makefile
make build
```

### 3. 配置插件系统

创建配置文件 `configs/plugins.json`：

```json
{
  "plugins": {
    "native": {
      "path": "./plugins",
      "enabled": true
    },
    "python": {
      "host": "localhost",
      "port": 5000,
      "timeout": 30,
      "enabled": true
    },
    "web": {
      "enabled": false,
      "path": "./web-plugins"
    }
  }
}
```

### 4. 使用插件管理器

```go
package main

import (
    "log"
    "github.com/yfh-yun/moviepilot-go/pkg/plugin"
)

func main() {
    // 创建配置
    config := &plugin.Config{
        Plugins: plugin.PluginConfig{
            Native: plugin.NativePluginConfig{
                Path:    "./plugins",
                Enabled: true,
            },
        },
    }

    // 创建插件管理器
    manager, err := plugin.NewHybridPluginManager(config)
    if err != nil {
        log.Fatal(err)
    }

    // 加载插件
    err = manager.LoadPlugin("./plugins/my_plugin.so", plugin.PluginTypeNative)
    if err != nil {
        log.Fatal(err)
    }

    // 初始化插件
    err = manager.InitializePlugin("my-plugin")
    if err != nil {
        log.Fatal(err)
    }

    // 启动插件
    err = manager.StartPlugin("my-plugin")
    if err != nil {
        log.Fatal(err)
    }

    // 调用插件方法
    result, err := manager.CallPluginMethod("my-plugin", "hello", "World")
    if err != nil {
        log.Fatal(err)
    }
    println(result)

    // 发布事件
    event := plugin.CreateEvent("test.event", "main", map[string]interface{}{
        "message": "Hello World",
    })
    manager.PublishEvent(event)

    // 停止插件
    err = manager.StopPlugin("my-plugin")
    if err != nil {
        log.Fatal(err)
    }
}
```

## API接口

插件系统提供以下HTTP API接口：

### 插件生命周期管理

- `POST /api/v1/plugin-manager/load` - 加载插件
- `POST /api/v1/plugin-manager/{pluginId}/initialize` - 初始化插件
- `POST /api/v1/plugin-manager/{pluginId}/start` - 启动插件
- `POST /api/v1/plugin-manager/{pluginId}/stop` - 停止插件
- `POST /api/v1/plugin-manager/{pluginId}/unload` - 卸载插件

### 插件信息和方法调用

- `GET /api/v1/plugin-manager/{pluginId}/info` - 获取插件信息
- `POST /api/v1/plugin-manager/{pluginId}/call` - 调用插件方法
- `GET /api/v1/plugin-manager/plugins` - 列出所有插件

### 事件管理

- `POST /api/v1/plugin-manager/events` - 发布事件

## 事件系统

插件可以订阅和处理事件：

```go
// 实现EventHandler接口
type EventHandler interface {
    HandleEvent(event Event) error
    GetEventSubscriptions() []string
}

// 在插件中实现事件处理
func (p *MyPlugin) GetEventSubscriptions() []string {
    return []string{"user.created", "media.downloaded"}
}

func (p *MyPlugin) HandleEvent(event plugin.Event) error {
    switch event.Type {
    case "user.created":
        // 处理用户创建事件
        userID := event.Data["user_id"]
        fmt.Printf("User created: %v\n", userID)
    case "media.downloaded":
        // 处理媒体下载事件
        mediaPath := event.Data["path"]
        fmt.Printf("Media downloaded: %v\n", mediaPath)
    }
    return nil
}
```

## 配置管理

每个插件都有自己的配置文件，存储在 `configs/plugins/{pluginId}.json`：

```json
{
  "greeting": "Hello, World!",
  "enabled": true,
  "api_key": "your-api-key"
}
```

插件可以在初始化时获取配置：

```go
func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    if greeting, ok := config["greeting"].(string); ok {
        p.greeting = greeting
    }
    return nil
}
```

## Python插件支持

Python插件通过HTTP/gRPC与主应用通信。创建Python插件：

1. 创建插件目录结构：
   ```
   python-plugins/
   ├── plugins/
   │   └── my_plugin/
   │       ├── __init__.py
   │       ├── plugin.py
   │       └── plugin.json
   └── cmd/
       └── server/
           └── main.py
   ```

2. 实现插件接口：
   ```python
   # plugins/my_plugin/plugin.py
   class MyPlugin:
       def __init__(self):
           self.id = "my-python-plugin"
           self.name = "My Python Plugin"
           self.version = "1.0.0"
       
       def initialize(self, config):
           print("Python plugin initialized")
           return True
       
       def start(self):
           print("Python plugin started")
           return True
       
       def stop(self):
           print("Python plugin stopped")
           return True
   ```

3. 启动Python插件服务：
   ```bash
   cd python-plugins
   python cmd/server/main.py
   ```

## 最佳实践

1. **错误处理**：始终返回有意义的错误信息
2. **日志记录**：使用结构化日志记录插件行为
3. **配置验证**：验证配置参数的有效性
4. **资源清理**：在Destroy方法中清理资源
5. **并发安全**：确保插件是线程安全的
6. **版本兼容**：使用语义化版本控制

## 故障排除

### 常见问题

1. **插件加载失败**
   - 检查.so文件是否存在
   - 确认插件导出了NewPlugin函数
   - 检查Go版本兼容性

2. **插件初始化失败**
   - 检查配置文件格式
   - 验证必需的配置参数
   - 查看日志了解详细错误

3. **Python插件通信失败**
   - 确认Python插件服务正在运行
   - 检查网络连接和端口配置
   - 验证API端点是否正确

### 调试技巧

1. 启用调试日志：
   ```bash
   export PLUGIN_DEBUG=true
   ```

2. 使用插件管理器示例：
   ```bash
   go run examples/plugin_manager_example.go
   ```

3. 检查插件状态：
   ```bash
   curl http://localhost:3001/api/v1/plugin-manager/plugins
   ```

## 扩展开发

如需扩展插件系统，可以：

1. 添加新的插件类型
2. 实现自定义事件处理器
3. 创建插件开发工具
4. 添加插件市场功能

更多信息请参考示例代码和API文档。