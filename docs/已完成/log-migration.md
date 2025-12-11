# log.py 日志系统迁移计划

> Python: `app/log.py`  
> Go: `pkg/logger/`

---

## 1. Python 日志系统概览

- 配置日志级别、格式（含颜色）。
- 管理日志文件目录与轮转策略。
- 为核心模块提供统一的 logger 实例。

---

## 2. Go 目标实现

- **位置**：`pkg/logger/logger.go`
- **技术栈**：zap + lumberjack
- **特性**：
  - 结构化 JSON 日志（可选 console 模式）。
  - 按大小/数量/天数轮转日志文件。
  - 支持多路输出（stdout + 文件）。
  - 上下文字段：`request_id`、`user_id`、`trace_id` 等。

---

## 3. 配置与环境变量

- 通过环境变量控制：
  - `LOGGER_LEVEL`（debug/info/warn/error/fatal/panic）
  - `LOGGER_FORMAT`（json/console）
  - `LOGGER_OUTPUT`（stdout/file/both）
  - `LOGGER_FILE`、`LOGGER_MAX_SIZE` 等轮转参数。

---

## 4. 使用规范回顾

- 所有日志通过 `pkg/logger` 间接使用：
  - 入口：`logger.Init()` / `logger.GetLogger()`。
  - 业务层：`logger.WithContext(ctx)` 获取带上下文的 logger。
- 禁止直接使用 `fmt.Println` / 标准库 `log` 输出业务日志。

---

## 5. 检查清单

- [ ] 所有层级代码都不再直接依赖 Python 旧日志方案。
- [ ] Go 代码中的日志调用统一使用 `pkg/logger`。
- [ ] 生产环境下默认使用 JSON + 文件轮转。
- [ ] 不在日志中记录敏感信息（密码、Token 等）。
