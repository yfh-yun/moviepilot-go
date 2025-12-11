# Notification 通知模块

> **版本**: v1.0.0  
> **更新时间**: 2025-12-02

---

## 📋 目录

1. [概述](#概述)
2. [架构设计](#架构设计)
3. [支持的通知渠道](#支持的通知渠道)
4. [快速开始](#快速开始)
5. [API 参考](#api-参考)
6. [最佳实践](#最佳实践)

---

## 概述

通知模块提供了统一的通知接口，支持多种通知渠道（Telegram、WeChat 企业微信等），并提供广播功能，可以同时向多个渠道发送消息。

### 核心特性

- ✅ **统一接口**：所有通知渠道实现相同的接口
- ✅ **多渠道支持**：Telegram、WeChat 企业微信
- ✅ **广播功能**：一次发送到多个渠道
- ✅ **并发发送**：使用 goroutine 并发发送，提高效率
- ✅ **错误处理**：完善的错误处理和日志记录
- ✅ **工厂模式**：动态注册和管理通知客户端

---

## 架构设计

### 核心组件

```
notification/
├── interface.go       # 统一接口定义
├── router.go          # 通知路由器
├── telegram/          # Telegram 客户端
│   └── client.go
└── wechat/            # WeChat 企业微信客户端
    └── client.go
```

### 接口定义

```go
type Client interface {
    Name() string
    SendText(ctx context.Context, message string) error
    SendImage(ctx context.Context, imageURL string, caption string) error
    SendFile(ctx context.Context, fileURL string, filename string) error
    SendMarkdown(ctx context.Context, markdown string) error
    Send(ctx context.Context, msg *Message) error
    TestConnection(ctx context.Context) error
}
```

### 消息类型

```go
type NotificationType string

const (
    NotificationTypeText     NotificationType = "text"
    NotificationTypeImage    NotificationType = "image"
    NotificationTypeFile     NotificationType = "file"
    NotificationTypeMarkdown NotificationType = "markdown"
)
```

---

## 支持的通知渠道

### 1. Telegram

**特性**：
- ✅ 支持文本消息
- ✅ 支持图片消息
- ✅ 支持文件消息
- ✅ 支持 Markdown 格式
- ✅ 支持 HTML 格式

**配置示例**：

```go
cfg := telegram.Config{
    BotToken:  "your_bot_token",
    ChatID:    "your_chat_id",
    ParseMode: "Markdown", // 可选：Markdown, HTML
    Timeout:   10 * time.Second,
}

client, err := telegram.NewClient(cfg)
```

**获取 Bot Token**：
1. 在 Telegram 中搜索 @BotFather
2. 发送 `/newbot` 创建新 Bot
3. 按提示设置 Bot 名称
4. 获取 Bot Token

**获取 Chat ID**：
1. 将 Bot 添加到群组或与 Bot 私聊
2. 访问：`https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
3. 在返回的 JSON 中找到 `chat.id`

### 2. WeChat 企业微信

**特性**：
- ✅ 支持文本消息
- ✅ 支持 Markdown 消息
- ✅ 支持图片消息（简化实现）
- ✅ 支持文件消息（简化实现）
- ✅ 自动管理 access_token

**配置示例**：

```go
cfg := wechat.Config{
    CorpID:  "your_corp_id",
    AgentID: "your_agent_id",
    Secret:  "your_secret",
    Timeout: 10 * time.Second,
}

client, err := wechat.NewClient(cfg)
```

**获取配置信息**：
1. 登录企业微信管理后台
2. 进入"应用管理" → "自建应用"
3. 创建应用并获取 AgentID 和 Secret
4. 在"我的企业"中获取 CorpID

---

## 快速开始

### 1. 创建通知客户端

```go
package main

import (
    "context"
    "time"

    "moviepilot-go/internal/integration/notification"
    "moviepilot-go/internal/integration/notification/telegram"
    "moviepilot-go/internal/integration/notification/wechat"
)

func main() {
    // 创建 Telegram 客户端
    tgClient, err := telegram.NewClient(telegram.Config{
        BotToken: "your_bot_token",
        ChatID:   "your_chat_id",
    })
    if err != nil {
        panic(err)
    }

    // 创建 WeChat 客户端
    wxClient, err := wechat.NewClient(wechat.Config{
        CorpID:  "your_corp_id",
        AgentID: "your_agent_id",
        Secret:  "your_secret",
    })
    if err != nil {
        panic(err)
    }

    // 创建工厂并注册客户端
    factory := notification.NewFactory()
    factory.Register(tgClient)
    factory.Register(wxClient)

    // 创建路由器
    router := notification.NewRouter(factory)

    // 发送消息
    ctx := context.Background()
    err = router.BroadcastText(ctx, "Hello from MoviePilot!")
    if err != nil {
        panic(err)
    }
}
```

### 2. 发送不同类型的消息

```go
// 发送文本消息
err := router.BroadcastText(ctx, "订阅更新：《The Last of Us》S01E05 已下载完成")

// 发送图片消息
err := router.BroadcastImage(ctx, 
    "https://image.tmdb.org/t/p/w500/poster.jpg",
    "《The Last of Us》海报")

// 发送 Markdown 消息
msg := &notification.Message{
    Type: notification.NotificationTypeMarkdown,
    Content: `
**订阅更新**

- 标题：The Last of Us
- 季集：S01E05
- 状态：✅ 下载完成
`,
}
err := router.Broadcast(ctx, msg)
```

### 3. 向指定渠道发送消息

```go
// 只发送到 Telegram
msg := &notification.Message{
    Type:    notification.NotificationTypeText,
    Content: "这条消息只发送到 Telegram",
}
err := router.SendToChannel(ctx, "telegram", msg)

// 只发送到 WeChat
err := router.SendToChannel(ctx, "wechat", msg)
```

### 4. 测试连接

```go
// 测试所有渠道
results := router.TestAllChannels(ctx)
for channel, err := range results {
    if err != nil {
        fmt.Printf("❌ %s: %v\n", channel, err)
    } else {
        fmt.Printf("✅ %s: 连接成功\n", channel)
    }
}
```

---

## API 参考

### Router 方法

#### BroadcastText

向所有渠道广播文本消息。

```go
func (r *Router) BroadcastText(ctx context.Context, message string) error
```

**参数**：
- `ctx`: 上下文
- `message`: 消息内容

**返回**：
- `error`: 如果所有渠道都失败，返回错误

#### BroadcastImage

向所有渠道广播图片消息。

```go
func (r *Router) BroadcastImage(ctx context.Context, imageURL string, caption string) error
```

#### Broadcast

向所有渠道广播通用消息。

```go
func (r *Router) Broadcast(ctx context.Context, msg *Message) error
```

#### SendToChannel

向指定渠道发送消息。

```go
func (r *Router) SendToChannel(ctx context.Context, channelName string, msg *Message) error
```

#### TestAllChannels

测试所有渠道的连接。

```go
func (r *Router) TestAllChannels(ctx context.Context) map[string]error
```

---

## 最佳实践

### 1. 错误处理

广播消息时，即使部分渠道失败，也会继续发送到其他渠道：

```go
err := router.BroadcastText(ctx, "测试消息")
if err != nil {
    // 部分渠道发送失败
    log.Warn("部分通知发送失败", zap.Error(err))
} else {
    // 所有渠道发送成功
    log.Info("通知发送成功")
}
```

### 2. 超时控制

使用 context 控制超时：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := router.BroadcastText(ctx, "测试消息")
```

### 3. 消息格式化

Telegram 支持 Markdown 和 HTML 格式：

```go
// Markdown 格式
message := `
*订阅更新*

标题：_The Last of Us_
状态：✅ 下载完成
`

// HTML 格式（需要在 Telegram Config 中设置 ParseMode: "HTML"）
message := `
<b>订阅更新</b>

标题：<i>The Last of Us</i>
状态：✅ 下载完成
`
```

### 4. 批量通知

对于批量通知，建议使用 goroutine 并发发送：

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        msg := fmt.Sprintf("下载完成：%s", i.Title)
        router.BroadcastText(ctx, msg)
    }(item)
}
wg.Wait()
```

### 5. 日志记录

所有通知操作都会自动记录日志：

```
INFO  通知发送成功  channel=telegram
INFO  通知发送成功  channel=wechat
ERROR 发送通知失败  channel=telegram error="request timeout"
```

---

## 扩展新的通知渠道

### 1. 实现 Client 接口

```go
package email

import (
    "context"
    "moviepilot-go/internal/integration/notification"
)

type Client struct {
    // ...
}

func (c *Client) Name() string {
    return "email"
}

func (c *Client) SendText(ctx context.Context, message string) error {
    // 实现发送逻辑
    return nil
}

// 实现其他接口方法...

var _ notification.Client = (*Client)(nil)
```

### 2. 注册到工厂

```go
emailClient := email.NewClient(emailConfig)
factory.Register(emailClient)
```

---

## 常见问题

### 1. Telegram 发送失败

**问题**：`Telegram API 返回错误: status=400`

**解决方案**：
- 检查 Bot Token 是否正确
- 检查 Chat ID 是否正确
- 确保 Bot 已添加到群组（如果发送到群组）

### 2. WeChat 获取 access_token 失败

**问题**：`获取 access_token 失败: code=40013`

**解决方案**：
- 检查 CorpID 是否正确
- 检查 Secret 是否正确
- 确保应用已启用

### 3. 消息发送超时

**问题**：`context deadline exceeded`

**解决方案**：
- 增加超时时间：`Timeout: 30 * time.Second`
- 检查网络连接
- 检查 API 服务是否正常

---

## 更新日志

### v1.0.0 (2025-12-02)

- ✅ 初始版本
- ✅ 支持 Telegram
- ✅ 支持 WeChat 企业微信
- ✅ 实现通知路由器
- ✅ 支持广播功能
- ✅ 完善的错误处理和日志记录

---

## 参考资源

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [企业微信 API](https://developer.work.weixin.qq.com/document/path/90664)
- [MoviePilot Go 项目文档](../../docs/README.md)
