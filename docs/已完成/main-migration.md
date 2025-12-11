# main.py 应用入口迁移计划

> Python: `app/main.py`  \
> Go: `cmd/server/main.go`

---

## 1. Python `main.py` 职责概览

- 设置进程名 / 标识。
- 初始化日志、配置、数据库等核心组件。
- 启动 FastAPI / uvicorn。
- 处理进程信号（SIGINT/SIGTERM），实现优雅关闭。
- （可选）托盘图标及桌面集成。

---

## 2. Go 入口文件现状

- **位置**：`cmd/server/main.go`
- **已实现能力（对照）**：
  - 初始化日志：`pkg/logger.Init`。
  - 加载配置：`internal/infrastructure/config`。
  - 初始化数据库 / 迁移 / 索引 / 基础数据。
  - 初始化调度器、Redis 等基础设施。
  - 创建 Gin Engine + 中间件 + 路由注册。
  - 启动 HTTP Server，支持 Read/Write/Idle Timeout。
  - 捕获 SIGINT/SIGTERM，使用 `ShutdownTimeout` 优雅关闭。

> 结论：Go 侧入口已经覆盖了 Python 入口的大部分核心职责。

---

## 3. 差异与补充点

- **进程名/标识**：
  - 可选：在启动日志中增加更丰富的实例标识（节点 ID、环境等）。
- **系统托盘/桌面集成**：
  - 如仍有需要，应单独作为桌面启动器项目处理，不建议耦合在服务入口。

---

## 4. 迁移确认清单

- [ ] 所有 Python `main.py` 中的初始化步骤在 Go 入口中都有对应实现或明确放弃。
- [ ] 优雅关闭流程已经涵盖所有长连接组件（DB、Redis、调度器等）。
- [ ] 日志中包含启动参数、版本号、环境信息，便于排错。
