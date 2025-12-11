# module.py 模块系统迁移计划

> Python: `app/core/module.py`  
> Go: `internal/platform/` + `internal/integration/` + `internal/business/services/`

---

## 1. Python 模块系统分析

### 1.1 模块管理器概览

```python
class ModuleManager(metaclass=Singleton):
    SubType = Union[DownloaderType, MediaServerType, MessageChannel, StorageSchema, OtherModulesType]

    def __init__(self):
        self._modules: dict = {}          # 所有可用模块类
        self._running_modules: dict = {}  # 已通过配置开关加载并运行的模块实例
        self.load_modules()
```

模块系统的定位：
- **统一管理 app.modules 下的“外部服务模块”**，例如：
  - 下载器（qBittorrent、Transmission 等）
  - 媒体服务器（Emby/Plex/Jellyfin）
  - 消息通道（Telegram/Slack/WeChat）
  - 存储后端（rclone、云盘等）
  - 其他第三方集成模块
- 通过配置开关(`settings.*`) 决定哪些模块启用
- 提供统一的查询接口：
  - 按 module_id 获取
  - 按类型 / 子类型 获取
  - 按是否实现某方法获取（如所有实现 `test` 的模块）

### 1.2 模块加载逻辑

```python
def load_modules(self):
    modules = ModuleHelper.load(
        "app.modules",
        filter_func=lambda _, obj: hasattr(obj, 'init_module') and hasattr(obj, 'init_setting')
    )
    self._running_modules = {}
    self._modules = {}
    for module in modules:
        module_id = module.__name__
        self._modules[module_id] = module
        try:
            _module = module()                     # 实例化
            if self.check_setting(_module.init_setting()):
                _module.init_module()              # 根据配置开关决定是否初始化
                self._running_modules[module_id] = _module
                logger.debug(f"Moudle Loaded：{module_id}")
        except Exception as err:
            logger.error("Load Module Error...", exc_info=True)
```

要点：
- `ModuleHelper.load("app.modules", filter_func=...)` 会扫描 `app/modules` 包下所有符合条件的类：
  - 必须具有 `init_module` 和 `init_setting` 方法
- 每个模块类需要实现统一接口约定：
  - `init_setting() -> tuple[switch_name, value] | None`：返回配置开关及开启值
  - `init_module()`：模块初始化（例如创建 HTTP client、检查配置、注册事件等）
  - 可选：`stop()`, `test()`, `get_type()`, `get_subtype()` 等

### 1.3 配置开关 check_setting

```python
@staticmethod
def check_setting(setting: Optional[tuple]) -> bool:
    if not setting:
        return True
    switch, value = setting
    option = getattr(settings, switch)
    if not option:
        return False
    if option and value is True:
        return True
    if value in option:
        return True
    return False
```

含义：
- 如果模块 `init_setting()` 返回 None → 默认启用
- 返回 `(switch, value)`：
  - 从全局配置 `settings.<switch>` 中取值：
    - 若为 False/空 → 不启用
    - 若 `value is True` 且 option 为真 → 启用（简单开关）
    - 若 `value` 在 `option` 中 → 启用（option 为字符串/列表时，支持多值匹配）

### 1.4 生命周期管理

```python
def stop(self):
    logger.info("正在停止所有模块...")
    for module_id, module in self._running_modules.items():
        if hasattr(module, "stop"):
            module.stop()
    logger.info("所有模块停止完成")


def reload(self):
    self.stop()
    self.load_modules()
    eventmanager.send_event(etype=EventType.ModuleReload, data={})
```

- `stop()`：遍历所有运行中的模块，如果有 `stop()` 方法则调用
- `reload()`：先 stop，再重新 load，并发送 `ModuleReload` 事件通知系统其它部分

### 1.5 模块查询接口

```python
def get_running_module(self, module_id: str) -> Any

def get_running_modules(self, method: str) -> Generator
    # 返回所有实现了指定方法名的运行模块

def get_running_type_modules(self, module_type: ModuleType) -> Generator
    # 调用 module.get_type() == module_type

def get_running_subtype_module(self, module_subtype: SubType) -> Generator
    # 调用 module.get_subtype() == module_subtype

# 静态模块信息

def get_module(self, module_id: str) -> Any

def get_modules(self) -> dict

def get_module_ids(self) -> List[str]
```

用途：
- 为业务链路提供统一入口，如在下载/通知/媒体服务器 service 中，按类型/子类型选择模块下发请求。

---

## 2. Go 侧总体设计

### 2.1 模块的定位

在 Go 架构中，`app/modules` 里的内容本质上是“**外部服务适配器**”：
- 下载器适配：qBittorrent/Transmission → `internal/integration/qbittorrent`, `.../transmission`
- 媒体服务器适配：Emby/Plex/Jellyfin → `internal/platform/emby`, `.../plex`, `.../jellyfin`
- 消息通道适配：Telegram/Slack/WeChat → `internal/integration/telegram`, `...`
- 存储适配：rclone/云盘 → `internal/platform/storage/...`

因此 Go 侧不再需要通过 `importlib` 动态扫描 Python 包，而是：
- 每个模块对应一个 **强类型的 client/service** 包
- 通过 **接口 + 依赖注入** 将这些模块纳入业务服务
- 用一个统一的 **ModuleRegistry / ModuleManager** 负责：
  - 注册所有模块实现
  - 根据配置判断启用哪个模块
  - 提供按类型/子类型检索模块的能力

### 2.2 目录与结构建议

```text
internal/platform/
  ├── downloader/          # 下载器模块
  │   ├── qbittorrent/
  │   ├── transmission/
  │   └── ...
  ├── mediaserver/         # 媒体服务器模块
  │   ├── emby/
  │   ├── plex/
  │   └── jellyfin/
  ├── message/             # 消息通道模块
  │   ├── telegram/
  │   ├── slack/
  │   └── wechat/
  └── storage/             # 存储模块
      ├── rclone/
      ├── alipan/
      └── ...

internal/integration/
  └── ...                  # 若有通用 HTTP/SDK 封装可放这里

internal/business/services/
  ├── download/            # 下载服务，依赖 downloader 接口
  ├── mediaserver/
  ├── notification/
  └── storage/

internal/business/domains/module/
  ├── types.go             # ModuleType, SubType 等
  └── registry.go          # ModuleRegistry / ModuleManager（Go 版）
```

### 2.3 Go 版 ModuleRegistry 设计

```go
// internal/business/domains/module/types.go

package module

// ModuleType / SubType 与 Python 枚举对应

type ModuleType string

const (
    ModuleTypeDownloader   ModuleType = "downloader"
    ModuleTypeMediaServer  ModuleType = "mediaserver"
    ModuleTypeMessage      ModuleType = "message"
    ModuleTypeStorage      ModuleType = "storage"
    ModuleTypeOther        ModuleType = "other"
)

// 子类型，如 DownloaderType、MediaServerType 等，可各自定义枚举

type DownloaderType string
// ... MediaServerType, MessageChannel, StorageSchema, OtherModulesType

// 通用模块接口（最小约束）

type Module interface {
    ID() string
    Type() ModuleType
    SubType() string      // 或具体枚举接口
    Init(cfg *config.Config) error
    Stop() error
    Test() (bool, string)
}
```

```go
// internal/business/domains/module/registry.go

type Registry struct {
    logger *zap.Logger

    modules       map[string]Module // 所有模块（已构造）
    running       map[string]Module // 已启用模块
    cfg           *config.Config
}

func NewRegistry(cfg *config.Config) *Registry {
    return &Registry{
        logger:  logger.GetLogger(),
        modules: make(map[string]Module),
        running: make(map[string]Module),
        cfg:     cfg,
    }
}

// Register 在应用启动时由各模块调用
func (r *Registry) Register(m Module) {
    id := m.ID()
    r.modules[id] = m
}

// LoadModules 根据配置加载模块
func (r *Registry) LoadModules() {
    r.running = make(map[string]Module)
    for id, m := range r.modules {
        if !r.checkSetting(m) {
            continue
        }
        if err := m.Init(r.cfg); err != nil {
            r.logger.Error("failed to init module", zap.String("id", id), zap.Error(err))
            continue
        }
        r.running[id] = m
        r.logger.Debug("module loaded", zap.String("id", id))
    }
}

func (r *Registry) Stop() {
    r.logger.Info("stopping all modules...")
    for id, m := range r.running {
        if err := m.Stop(); err != nil {
            r.logger.Error("failed to stop module", zap.String("id", id), zap.Error(err))
        } else {
            r.logger.Debug("module stopped", zap.String("id", id))
        }
    }
}

func (r *Registry) Reload(bus events.Bus) {
    r.Stop()
    r.LoadModules()
    // 发送 ModuleReload 事件（对应 Python 逻辑）
    _ = bus.PublishBroadcast(context.Background(), events.EventTypeModuleReload, nil, 10)
}

// 查询接口

func (r *Registry) GetRunningModule(id string) Module {
    return r.running[id]
}

func (r *Registry) GetRunningModulesByType(t ModuleType) []Module {
    var result []Module
    for _, m := range r.running {
        if m.Type() == t {
            result = append(result, m)
        }
    }
    return result
}

func (r *Registry) GetRunningModulesBySubType(subType string) []Module {
    var result []Module
    for _, m := range r.running {
        if m.SubType() == subType {
            result = append(result, m)
        }
    }
    return result
}
```

### 2.4 配置开关 checkSetting 映射

可以模仿 Python 的设计，在模块实现中提供类似方法，或者统一用 struct tag / 配置字段控制。

简单方案：每个模块在 `Init` 中自行检查配置；为对齐 Python 行为，可以在 Registry 中集中做：

```go
func (r *Registry) checkSetting(m Module) bool {
    // 假设每个模块提供 SettingInfo() (*Setting, bool)
    if cs, ok := m.(interface{ SettingInfo() (*Setting, bool) }); ok {
        setting, exists := cs.SettingInfo()
        if !exists || setting == nil {
            return true
        }
        return r.matchConfig(setting)
    }
    return true
}

type Setting struct {
    Key   string      // 对应 settings 中的字段名
    Value interface{} // True 或某个子类型值
}

func (r *Registry) matchConfig(s *Setting) bool {
    // 从 r.cfg 中查找 Key，并根据 Value 判定是否开启
    // 行为与 Python check_setting 尽量一致
}
```

---

## 3. 与业务和事件系统的集成

### 3.1 业务服务如何使用模块

示例：下载服务 `DownloadService`：

```go
type DownloadService struct {
    modules *module.Registry
}

func (s *DownloadService) getDownloader(t module.DownloaderType) (downloader.Client, error) {
    mods := s.modules.GetRunningModulesBySubType(string(t))
    if len(mods) == 0 {
        return nil, errors.New("no downloader module enabled")
    }
    // 假设模块实现了 downloader.Client 接口
    if d, ok := mods[0].(downloader.Client); ok {
        return d, nil
    }
    return nil, errors.New("invalid downloader module type")
}
```

### 3.2 与事件系统 (`EventManager` → Go events.Bus) 集成

Python `ModuleManager.reload()`：

```python
self.stop()
self.load_modules()
eventmanager.send_event(etype=EventType.ModuleReload, data={})
```

Go 对应：

```go
func (r *Registry) Reload(bus events.Bus) {
    r.Stop()
    r.LoadModules()
    _ = bus.PublishBroadcast(context.Background(), events.EventTypeModuleReload, nil, 10)
}
```

业务模块/插件可以监听这个事件，执行自己的重置逻辑。

---

## 4. 实施计划

### Phase 1：基础结构（Week 1）

- [ ] 创建 `internal/business/domains/module/`：
  - `types.go`：定义 `ModuleType`、各子类型枚举
  - `registry.go`：定义 `Module` 接口与 `Registry` 结构，提供：
    - `Register`, `LoadModules`, `Stop`, `Reload`
    - `GetRunningModule`, `GetRunningModulesByType`, `GetRunningModulesBySubType`
- [ ] 为 configuration 增加模块相关字段映射，准备好开关位

### Phase 2：逐类模块接入（Week 2-3）

- [ ] 下载器模块：
  - 为 qBittorrent / Transmission 等实现 `Module` 接口
  - 在启动时注册到 Registry 中
- [ ] 媒体服务器模块：Emby / Plex / Jellyfin
- [ ] 消息模块：Telegram / Slack / WeChat
- [ ] 存储模块：rclone / 云盘

### Phase 3：替换旧调用（Week 4）

- [ ] 替换链路中对 Python `ModuleManager` 的依赖：
  - 下载链路：统一从 Registry 获取 downloader 实例
  - 通知链路：统一从 Registry 获取 message channel 实例
  - 媒体链路：统一从 Registry 获取 mediaserver 实例
- [ ] 将 `reload` 能力挂到管理 API / 管理命令中

### Phase 4：优化与监控

- [ ] 为模块加载/停止/测试增加结构化日志
- [ ] 暴露一个 `/modules` 调试接口：
  - 列出所有模块及运行状态/类型/子类型
- [ ] 为关键模块增加健康检查与自检逻辑

---

## 5. 测试策略

### 单元测试

- Registry：
  - 注册多个假模块，验证 Load/Stop/Reload 的行为
  - 验证 `GetRunningModulesByType/SubType` 返回正确
  - 验证 checkSetting 的逻辑与 Python 一致

- 模块实现：
  - 为每类模块实现基础测试（Init / Stop / Test）

### 集成测试

- 在集成环境中：
  - 配置不同组合的模块开关
  - 验证启动后哪些模块实际被加载
  - 验证 reload 后，模块能否正确重新初始化

---

## 6. 注意点

- Go 不支持像 Python 那样动态扫描并实例化类，模块注册需要显式进行：
  - 在 `bootstrap.InitModules` 中显式 new 每个模块并调用 `registry.Register()`
- 要避免在 `Registry` 中硬编码具体模块类型，保持接口化
- 日志必须通过 `pkg/logger`，并包含模块 ID / 类型 / 子类型 等关键字段

---

## 7. 结论

`module.py` 提供的是一个“**动态模块开关 + 生命周期管理 + 类型查询**”的统一入口。在 Go 中：

- 使用 `internal/business/domains/module.Registry` 替代 `ModuleManager`
- 把原本 `app.modules` 中的 Python 模块拆成若干强类型 Go 包（downloader/mediaserver/message/storage 等）
- 在启动阶段显式注册模块，并通过配置决定启用哪些
- 与事件系统集成，实现 ModuleReload 通知

这样既保留了模块化扩展能力，又遵守了 Go 的类型系统和你的分层架构规范。