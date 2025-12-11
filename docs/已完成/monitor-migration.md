# monitor.py 文件监控迁移计划

> Python: `app/monitor.py`  \
> Go: `internal/monitor/`

---

## 1. Python 监控系统概览

- **职责**：
  - 监控指定目录（下载目录、媒体目录等）的文件变化。
  - 根据变更触发相应工作流（刮削、整理、入库等）。
  - 与缓存/数据库协作，避免重复处理。
- **技术栈**：watchdog / 文件系统事件。

---

## 2. Go 目标设计

- **位置**：`internal/monitor/`
- **可能子模块**：
  - `filewatch/`：封装 fsnotify 事件。
  - `pipeline/`：事件到工作流的分发。

示例接口：

```go
type FileWatcher interface {
    Watch(path string, handler EventHandler) error
    Stop() error
}

type EventHandler func(event Event)

type Event struct {
    Path string
    Op   Operation // Create/Write/Remove/Rename
}
```

---

## 3. 事件流迁移

1. 目录变化事件 → `FileWatcher` 产生 `Event`。
2. 将 `Event` 转换为领域事件（如“下载完成”、“新媒体文件出现”）。
3. 通过事件系统或工作流引擎触发后续处理：
   - 刮削 → 整理 → 入库 → 通知。

---

## 4. 迁移步骤

1. 明确 Python `monitor.py` 中监控的目录和触发规则。
2. 在 Go 中实现 `internal/monitor/filewatch`，封装 fsnotify。
3. 将监控事件通过业务 Service 或 workflow 系统分发：
   - 避免监控层直接调用数据库或复杂业务逻辑。
4. 为关键路径增加日志与调试开关。

---
