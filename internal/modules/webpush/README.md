# WebPush 模块

## 概述

WebPush 模块是 MoviePilot 系统中用于发送 Web 推送通知的组件，支持向浏览器发送实时通知。

## 核心组件

### WebPushModule (WebPush模块)
作为系统模块，集成到 MoviePilot 的模块系统中：
- 模块生命周期管理
- 消息发送
- 配置管理

## 功能特性

### 消息处理
1. **消息发送** - 支持向订阅用户发送Web推送通知
2. **用户过滤** - 支持指定接收用户
3. **内容处理** - 支持标题和内容的灵活组合

### 配置管理
1. **配置变更处理** - 自动响应系统配置变更
2. **用户过滤** - 根据配置过滤消息接收用户

## 配置说明

需要以下配置项：
- `WEBPUSH_USERNAME` - 接收推送通知的用户名列表（逗号分隔）

## 使用方法

### 初始化
```go
webpushModule := NewWebPushModule()
webpushModule.InitModule()
```

### 发送消息
```go
webpushModule.PostMessage(&Notification{
    Title:    "测试消息",
    Text:     "这是一条测试消息",
    Username: "user123",
})
```

## 注意事项

1. 需要正确配置VAPID密钥
2. 浏览器需要先订阅推送服务
3. 网络连接正常，能够访问推送服务
4. 注意用户权限验证