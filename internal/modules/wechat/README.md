# 企业微信模块

## 概述

企业微信模块是 MoviePilot 系统中用于与企业微信进行消息交互的组件，支持发送消息、接收消息、创建菜单等功能。

## 核心组件

### WXBizMsgCrypt (消息加解密)
实现企业微信消息的加解密功能，包括：
- 消息签名验证
- 消息加密
- 消息解密

### WeChat (企业微信客户端)
实现企业微信的核心功能：
- Access Token 管理
- 消息发送（文本、图文）
- 媒体列表发送
- 种子列表发送
- 菜单管理（创建、删除）

### WechatModule (微信模块)
作为系统模块，集成到 MoviePilot 的模块系统中：
- 模块生命周期管理
- 消息解析
- 消息发送
- 命令注册

## 功能特性

### 消息处理
1. **消息接收** - 支持接收文本消息和事件消息
2. **消息解密** - 对企业微信加密消息进行解密处理
3. **权限验证** - 验证用户是否有权限执行命令
4. **消息解析** - 解析XML格式的消息内容

### 消息发送
1. **文本消息** - 发送纯文本消息
2. **图文消息** - 发送包含图片的消息
3. **媒体列表** - 发送媒体信息列表
4. **种子列表** - 发送种子信息列表

### 菜单管理
1. **菜单创建** - 根据系统命令自动创建企业微信菜单
2. **菜单删除** - 删除已创建的菜单
3. **菜单分组** - 按类别对命令进行分组

## 配置说明

需要以下配置项：
- `WECHAT_CORPID` - 企业ID
- `WECHAT_APP_SECRET` - 应用密钥
- `WECHAT_APP_ID` - 应用ID
- `WECHAT_TOKEN` - 消息令牌
- `WECHAT_ENCODING_AESKEY` - 消息加解密密钥
- `WECHAT_ADMINS` - 管理员用户ID列表（可选）

## 使用方法

### 初始化
```go
wechatModule := NewWechatModule()
wechatModule.InitModule()
```

### 发送消息
```go
wechatModule.PostMessage(&Notification{
    Title: "测试消息",
    Text:  "这是一条测试消息",
    UserID: "user123",
})
```

### 注册命令
```go
commands := map[string]map[string]interface{}{
    "/test": {
        "description": "测试命令",
        "category":    "测试",
    },
}
wechatModule.RegisterCommands(commands)
```

## 注意事项

1. 确保配置信息正确无误
2. 网络连接正常，能够访问企业微信API
3. 消息加解密密钥需要严格保密
4. 注意Access Token的有效期管理