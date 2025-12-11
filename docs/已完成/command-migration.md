# command.py 命令系统迁移计划

> Python: `app/command.py`  
> Go: `internal/business/services/command/` + 事件系统

---

## 1. Python 命令系统概览

- **职责**：
  - 解析消息中的命令（如 `/subscribe`、`/status`）
  - 命令注册与路由
  - 调用对应业务逻辑（订阅、搜索、系统控制等）
- **入口**：
  - 从通知渠道（Telegram/WeChat 等）收到消息事件 → 解析命令 → 执行

---

## 2. Go 目标设计

- **位置**：`internal/business/services/command/`
- **核心接口示例**：

```go
type Service interface {
    Register(cmd Handler) error
    Execute(ctx context.Context, input string) error
    List() []Info
}

type Handler interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args []string) error
}
```

- **集成方式**：
  - 通过事件系统订阅“收到消息”事件（见 `event-migration-plan.md`）。
  - 将命令文本交给 `CommandService` 统一处理。

---

## 3. 命令映射表（初版）

| Python 命令 | Go Handler 目标 | 状态 | 说明 |
|-------------|-----------------|------|------|
| `/subscribe` | `SubscribeCommandHandler` | ⏳ 待迁移 | 新建订阅 |
| `/unsubscribe` | `UnsubscribeCommandHandler` | ⏳ 待迁移 | 取消订阅 |
| `/status` | `StatusCommandHandler` | ⏳ 待迁移 | 系统状态 |
| `/help` | `HelpCommandHandler` | ⏳ 待迁移 | 帮助信息 |

> 后续按 `command.py` 中实际命令完整补全表格，并在 Go 端实现后更新状态。

---

## 4. 迁移步骤

1. 在 Python 端列出所有命令及其参数、权限要求。
2. 在 Go 的 `internal/business/services/command/` 中实现 `Service` 与若干 `Handler`：
   - 每个命令一个 handler，内部调用对应业务 Service。
3. 在事件系统中注册命令入口：
   - 当收到消息事件时，调用 `CommandService.Execute`。
4. 为关键命令编写单元测试：
   - 正常路径 + 错误参数 + 权限不足等场景。

---

## 5. 检查清单

- [ ] 所有命令都在映射表中出现。
- [ ] 命令执行逻辑不直接访问 Repository，而是通过业务 Service。
- [ ] 日志使用 `pkg/logger` 记录命令执行情况。
- [ ] 命令系统与通知渠道解耦（通过事件总线/接口抽象）。
