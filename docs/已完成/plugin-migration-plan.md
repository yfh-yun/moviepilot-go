# plugin.py 插件系统迁移计划

> Python: `app/core/plugin.py`  
> Go: `pkg/plugin/` + `python-plugins/` + `internal/infrastructure/events/` + `internal/business/services/*`

---

## 1. Python 插件系统概览

### 1.1 文件职责

`app/core/plugin.py` 是 MoviePilot 的 **主插件管理器**，核心职责：

- 插件发现与加载（从 `app/plugins` 目录扫描）
- 插件配置管理（读取/保存到 `systemconfig`）
- 插件数据管理（通过 `plugindata_oper`）
- 插件生命周期管理（start/stop/reload）
- 热更新（DEV 或 `PLUGIN_AUTO_RELOAD` 模式 + watchdog 监控 .py 文件改动）
- 与事件系统集成（根据插件状态启用/禁用事件处理器、响应 ConfigChanged / PluginReload 等事件）
- 插件能力暴露：命令、API、定时服务、动作、模块、前端联邦组件（remoteEntry.js）、仪表盘等
- 插件安装与依赖管理（通过 `PluginHelper` 与远程仓库 / pip 交互）

> 注意：当前这套是 **同进程内 Python 插件系统**，而你的 Go 版本规划中还有一个单独的 `python-plugins` gRPC 服务，需要做架构层拆分。

---

## 2. 关键组件分析（Python）

### 2.1 PluginMonitorHandler（热更新监控）

- 继承 `watchdog.events.FileSystemEventHandler`
- 监听 `app/plugins` 目录下 `.py` 文件的修改：
  - 排除 `__pycache__` 等
  - 通过读取对应目录下的 `__init__.py`，找到 `class XXXX(_PluginBase)` 的类名 → 作为 `pid`
  - 调用 `PluginManager().reload_plugin(pid)`，并用 `rate_limit_window` 做 2s 窗口限流

### 2.2 PluginManager 核心属性

```python
class PluginManager(metaclass=Singleton):
    def __init__(self):
        self._plugins: dict = {}            # 插件类
        self._running_plugins: dict = {}    # 插件实例
        self._config_key: str = "plugin.%s" # 配置存储 key 前缀
        self._observer: Observer = None     # 文件监控
        if settings.DEV or settings.PLUGIN_AUTO_RELOAD:
            self.__start_monitor()
```

### 2.3 插件启动流程 start(pid=None)

1. 获取已安装插件列表：`SystemConfigOper().get(UserInstalledPlugins)`
2. `_load_selective_plugins(pid, installed_plugins, check_module)`：
   - 扫描 `app/plugins` 目录
   - 只处理在已安装列表/指定 pid 中的目录
   - 只加载包含 `__init__.py` 的包
   - import `app.plugins.<dir>` 并 `reload`
   - 在模块中找到 **第一个满足 `hasattr(obj, 'init_plugin') and hasattr(obj, 'plugin_name')` 的类** 作为插件类
3. 对每个插件类：
   - 调用 `__set_and_check_auth_level(plugin)` 检查认证等级（根据插件元信息和系统认证配置）
   - 记录插件 class 到 `_plugins`
   - 实例化为 `plugin_obj`
   - `plugin_obj.init_plugin(self.get_plugin_config(plugin_id))` 应用配置
   - 将实例放入 `_running_plugins`
   - 检查插件状态 `plugin_obj.get_state()`：
     - 若启用 → `eventmanager.enable_event_handler(plugin)`
     - 否则 → `eventmanager.disable_event_handler(plugin)`

### 2.4 插件停止 & 重载

- `stop(pid=None)`：
  - 可停单个或全部插件
  - 对每个插件：
    - `eventmanager.disable_event_handler(type(plugin))`
    - `__stop_plugin(plugin)`：调用 `close()` 和 `stop_service()`（如有）
  - 清理 `_plugins` 和 `_running_plugins` 中的对象
  - 调用 `_clear_plugin_modules(pid)`：从 `sys.modules` 中删除对应的模块缓存

- `reload_plugin(plugin_id)`：
  - `stop(plugin_id)` → 清理
  - `start(plugin_id)` → 再加载
  - 发送 `EventType.PluginReload` 广播事件

### 2.5 配置、数据与状态

- 配置：
  - 存储在 `systemconfig` 表中，key 为 `plugin.<pid>`，value 为 dict
  - 读：`get_plugin_config(pid)`
  - 写：`save_plugin_config(pid, conf, force=False)`
  - 删：`delete_plugin_config(pid)`

- 数据：
  - `PluginDataOper().del_data(pid)` 删除插件相关数据

- 状态：
  - `get_plugin_state(pid)`：返回 `plugin.get_state()`

### 2.6 插件暴露的能力（对外接口）

- 命令：`get_plugin_commands(pid=None)`
  - 插件可实现 `get_command()` 返回列表：
    - `{"cmd": "/xx", "event": EventType.xx, "desc": "..", "data": {}}`

- API：`get_plugin_apis(pid=None)`
  - 插件可实现 `get_api()`，返回路由描述：
    - `path`, `endpoint`, `methods`, `summary`, `description`, `allow_anonymous`, `auth`

- 定时服务：`get_plugin_services(pid=None)`
  - 插件可实现 `get_service()` 定义 APScheduler 任务：
    - `id`, `name`, `trigger`, `func`, `kwargs`, `func_kwargs`

- 模块：`get_plugin_modules(pid=None)`
  - 插件可实现 `get_module()` 提供模块函数集

- 动作：`get_plugin_actions(pid=None)`
  - 插件可实现 `get_actions()`，用于工作流、按钮触发等

- 前端联邦组件：
  - `get_plugin_remotes(pid=None)` → 返回 `id`, `url`, `name`
  - `get_plugin_remote_entry(plugin_id, dist_path)` → `/plugin/file/<id>/<dist_path>/remoteEntry.js`

- 仪表盘：
  - `get_plugin_dashboard_meta()`：所有插件 dashboard 的元信息
  - `get_plugin_dashboard(pid, key, user_agent)`：根据插件实现渲染仪表盘（不同 render_mode，如 vue）

- 反射式调用：
  - `get_plugin_attr(pid, attr)`
  - `run_plugin_method(pid, method, *args, **kwargs)`
  - `async_run_plugin_method(pid, method, *args, **kwargs)`

### 2.7 安装/同步/依赖

- `sync()`：
  - 获取已安装插件列表（systemconfig）
  - 获取在线插件列表（`get_online_plugins`，在后半部分定义）
  - 对比版本，找出需要安装/更新的插件
  - 用 `PluginHelper().install(...)` 并发安装（ThreadPoolExecutor），记录成功/失败

- `install_plugin_missing_dependencies()`：
  - 用 `PluginHelper` 检测缺失依赖并尝试安装

> 后半部分（未展示）还包括：插件市场访问、签名/认证、license 校验、远程仓库管理等。

---

## 3. Go 侧架构设计（总体）

你的 Go 版规划中已经有：

- `pkg/plugin/`：插件系统核心
- `python-plugins/`：独立 Python 插件服务，通过 gRPC 与 Go 主应用通信

因此插件系统迁移的目标是：

1. 将 **主应用内的 Python 插件管理逻辑** 重构为 Go 侧的插件抽象层：
   - 插件注册、启停、配置、状态、能力描述（命令、服务、API、动作、仪表盘等）
2. 将具体 Python 插件 **迁移/保留在 `python-plugins` 服务** 中，Go 通过 gRPC 调用
3. 保持与现有前端/API 的合同基本兼容（命令、仪表盘、前端联邦组件等）

### 3.1 目标结构建议

```text
pkg/plugin/
  ├── interface.go        # 插件接口定义（Go 侧抽象）
  ├── manager.go          # PluginManager（Go 版）
  ├── registry.go         # 插件注册表（Go 原生插件）
  ├── bridge.go           # Python 插件桥接（gRPC client）
  └── model.go            # 配置/状态/能力描述结构

python-plugins/
  ├── ...                 # 已有，作为插件运行时环境

internal/apis/handlers/plugin/
  ├── handler.go          # 插件管理 API（列表、安装、配置、仪表盘等）

internal/schedulers/
  ├── plugin_jobs.go      # 从插件获取定时任务并注册到调度器

internal/business/services/plugin/
  ├── service.go          # 业务层封装（命令派发、动作执行等）
```

### 3.2 Go 插件接口（pkg/plugin/interface.go）

```go
package plugin

import (
    "context"
)

type State string

const (
    StateDisabled State = "disabled"
    StateEnabled  State = "enabled"
)

// Plugin 定义 Go 侧插件抽象（包括 Python 侧插件）
type Plugin interface {
    ID() string
    Name() string
    Version() string

    // 配置生命周期
    Init(ctx context.Context, cfg map[string]interface{}) error
    Stop(ctx context.Context) error

    // 状态查询
    State(ctx context.Context) State

    // 能力暴露（可选）
    Commands(ctx context.Context) ([]Command, error)
    APIs(ctx context.Context) ([]API, error)
    Services(ctx context.Context) ([]Service, error)
    Modules(ctx context.Context) (map[string]interface{}, error)
    Actions(ctx context.Context) ([]ActionGroup, error)
    DashboardMeta(ctx context.Context) ([]DashboardMeta, error)
    Dashboard(ctx context.Context, key, userAgent string) (*Dashboard, error)
    RenderMode(ctx context.Context) (string, string) // mode, distPath
}

// 以下结构体对应 Python 侧返回的数据结构

type Command struct {
    Cmd   string                 `json:"cmd"`
    Event string                 `json:"event"`
    Desc  string                 `json:"desc"`
    Data  map[string]interface{} `json:"data"`
}

type API struct {
    Path          string   `json:"path"`
    Methods       []string `json:"methods"`
    Summary       string   `json:"summary"`
    Description   string   `json:"description"`
    AllowAnonymous bool    `json:"allow_anonymous"`
    Auth          string   `json:"auth"`
    // Endpoint 由 Go 侧包装，不直接暴露
}

// Service, Module, ActionGroup, DashboardMeta, Dashboard ...
```

> 对于 Go 原生插件，`Plugin` 接口直接由 Go 实现；对于 Python 插件，通过 gRPC 客户端实现一个代理类，实现上述接口，内部转发请求到 `python-plugins` 服务。

### 3.3 PluginManager（Go 版）

```go
// pkg/plugin/manager.go

type Manager struct {
    logger *zap.Logger

    plugins       map[string]Plugin          // 所有已注册插件（Go + Python 代理）
    running       map[string]Plugin          // 当前启用的插件
    configStore   ConfigStore                // 插件配置持久化（封装 systemconfig_oper 对应的 repository）
    dataStore     DataStore                  // 插件数据持久化（封装 PluginDataOper）
    eventBus      events.Bus                 // 事件总线（用于 PluginReload 等）
}

func NewManager(cfgStore ConfigStore, dataStore DataStore, bus events.Bus) *Manager

func (m *Manager) Register(p Plugin) {
    m.plugins[p.ID()] = p
}

func (m *Manager) Start(ctx context.Context, pid string) error
func (m *Manager) Stop(ctx context.Context, pid string) error
func (m *Manager) Reload(ctx context.Context, pid string) error
func (m *Manager) StartAll(ctx context.Context) error
func (m *Manager) StopAll(ctx context.Context) error
```

启动/停止逻辑：
- 读取配置 `plugin.<pid>`（通过 `ConfigStore`）
- 调用 `p.Init(ctx, cfg)`
- 根据 `p.State(ctx)` 决定是否在 `running` map 中启用
- 向事件总线发送：
  - 插件 Reload：`EventTypePluginReload`

### 3.4 Python 插件桥接（bridge.go）

- 在 `pkg/plugin/bridge.go` 中实现一个 `PythonPlugin` 类型：
  - 实现 `Plugin` 接口
  - 内部持有 gRPC client：调用 `python-plugins` 服务暴露的接口：
    - `InitPlugin`, `StopPlugin`
    - `GetCommands`, `GetAPIs`, `GetServices`, `GetDashboard`, `GetActions`, `GetRemotes`, etc.

这样可以：
- Go 主应用只关心 `Plugin` 抽象
- Python 插件运行在独立服务中，负责具体逻辑

---

## 4. 功能映射表（Python → Go）

| Python 功能 | 位置 | Go 对应 | 位置 |
|-------------|------|---------|------|
| 插件扫描 & 加载 (`_load_selective_plugins`, `start`) | 同进程内，从 `app/plugins` 目录动态扫描 | 插件注册 + 通过配置启用 | `pkg/plugin.Manager.Register/Start`，Python 端通过 gRPC 报告自身注册信息或由 Go 侧从插件市场读取元数据 |
| 插件配置读取/保存 | `SystemConfigOper` + `_config_key` | 抽象 `ConfigStore` 接口（底层 Postgres） | `internal/repositories` + `pkg/plugin` |
| 插件数据删除 (`PluginDataOper`) | DB 操作 | `DataStore` 接口 | 同上 |
| 插件启停/重载 | `start/stop/reload_plugin` | `Manager.Start/Stop/Reload` | `pkg/plugin` |
| 事件集成（启用/禁用 handler） | `eventmanager.enable/disable_event_handler` | 在 `Plugin` 代理中通过 `events.Bus` 注册 handler 或在插件内部注册 | `internal/infrastructure/events` + 业务 service |
| 热更新（watchdog） | `PluginMonitorHandler` + `Observer` | 可以用 fsnotify 监控本地插件包（若仍保留本地 Python 插件）或依赖 Python-plugins 自身热更新 | `internal/monitor/filewatch` 或 `python-plugins` 内部实现 |
| 插件命令/API/服务/动作/仪表盘 | 一系列 `get_*` 方法反射调用 | 定义 `Plugin` 接口的能力方法，并由 Go/Python 具体实现填充结构体 | `pkg/plugin`, `python-plugins` |
| 插件安装/更新/依赖处理 | `PluginHelper` + pip + GitHub 插件市场 | 在 Go 侧通过 HTTP 调用插件市场 API + 调度 Python-plugins/脚本执行 pip，或保持在 Python 端实现 | `internal/integration` + `python-plugins` |

---

## 5. 实施计划

### Phase 1：接口与管理器骨架（Week 1）

- [ ] 在 `pkg/plugin/` 中创建：
  - `interface.go`：定义 `Plugin` 接口和能力结构体（Command/API/Service/Action/Dashboard 等）
  - `manager.go`：定义 `Manager`，完成基础注册/启停/配置读写逻辑骨架
  - `model.go`：封装插件配置/状态/元信息模型
- [ ] 在 `internal/repositories` 中增加插件配置/数据的 Repository 接口（封装现有 `plugindata_oper`、`systemconfig_oper`）

### Phase 2：Python 插件桥接（Week 2-3）

- [ ] 在 `pkg/plugin/bridge.go` 实现 `PythonPlugin`：
  - 使用 gRPC client 调用 `python-plugins` 服务
  - 将 gRPC 返回的数据映射到 `Plugin` 接口结构体
- [ ] 定义 `PluginService` gRPC 协议（若尚未定义）：
  - `InitPlugin`, `StopPlugin`, `GetCommands`, `GetAPIs`, `GetServices`, `GetDashboard`, `GetActions`, `GetRemotes`, etc.
- [ ] 在 `python-plugins` 服务中实现上述 gRPC 接口，对接现有 Python 插件机制

### Phase 3：业务接入与事件集成（Week 4-5）

- [ ] 修改命令系统、调度器、工作流、API 层：
  - 从 `pkg/plugin.Manager` 获取插件命令/API/服务/动作
  - 将插件服务注册到调度器（`internal/schedulers`）
  - 通过 `events.Bus` 实现 PluginReload 等通知
- [ ] 迁移 `get_plugin_dashboard*` 逻辑，统一从 `Plugin.Dashboard`/`DashboardMeta` 获取

### Phase 4：插件安装/市场集成（Week 6+）

- [ ] 在 Go/或 Python-plugins 侧实现插件市场访问逻辑：
  - 查询在线插件列表
  - 下载安装/更新
- [ ] 提供 API：
  - 列出可用插件（本地 + 远程）
  - 安装/卸载/升级插件
  - 安装缺失依赖

---

## 6. 测试策略

### 单元测试

- `Manager`：
  - Register/Start/Stop/Reload 行为
  - 配置读写、状态变化
- `PythonPlugin` 代理：
  - 在测试环境模拟 gRPC server，检查映射正确性

### 集成测试

- 在开发环境启动 `python-plugins` 服务 + Go 主应用：
  - 验证插件列表、命令、API、服务、仪表盘在 Go API 中正确暴露
  - 验证 PluginReload 行为（通过配置变更或 gRPC 指令）

---

## 7. 注意事项

- 避免在 Go 中再现 Python 的反射式动态导入；模块发现应通过：
  - 显式注册（Go 原生插件）
  - Python-plugins 自身的插件列表接口
- 严格通过 `pkg/logger` 打日志，记录插件 ID、版本、操作类型（start/stop/reload/install）
- 处理好 Go ↔ Python 之间的错误/超时；防止 plugin 调用阻塞主应用

---

## 8. 结论

`app/core/plugin.py` 是一个非常重的插件管理中枢。在 Go 迁移方案中，需要将其拆解为：

1. `pkg/plugin`：定义统一的插件接口与管理器，实现 Go 原生插件与 Python 插件的统一抽象；
2. `python-plugins`：作为 Python 插件运行时，通过 gRPC 暴露插件能力；
3. `internal/*`：命令、调度、API、工作流等模块通过 `pkg/plugin.Manager` 获取插件能力，而不直接依赖 Python 代码。

这样既保留了原先插件生态（尤其是现有 Python 插件），又符合你当前 Go + gRPC 双服务架构的方向。