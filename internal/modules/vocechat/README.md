# VoceChat 模块

VoceChat 模块是 MoviePilot 的一个通知模块，用于通过 VoceChat 服务发送通知消息。

## 功能特性

- 发送文本消息
- 发送媒体信息列表
- 发送种子信息列表
- 接收并解析来自 VoceChat 的消息
- 支持向频道或个人用户发送消息

## 配置参数

- `VOCECHAT_HOST`: VoceChat 服务器地址
- `VOCECHAT_API_KEY`: API 密钥
- `VOCECHAT_CHANNEL_ID`: 频道 ID

## 实现细节

该模块实现了以下接口方法：

- `InitModule()`: 初始化模块
- `HandleConfigChanged()`: 处理配置变更事件
- `GetName()`: 获取模块名称
- `GetType()`: 获取模块类型
- `GetSubType()`: 获取模块子类型
- `GetPriority()`: 获取模块优先级
- `Stop()`: 停止模块
- `Test()`: 测试模块连接性
- `MessageParser()`: 解析消息内容
- `PostMessage()`: 发送消息
- `PostMediasMessage()`: 发送媒体信息列表
- `PostTorrentsMessage()`: 发送种子信息列表
- `RegisterCommands()`: 注册命令

## VoceChat 客户端

模块包含一个 VoceChat 客户端实现，提供以下功能：

- `GetState()`: 获取服务状态
- `SendMsg()`: 发送文本消息
- `SendMediasMsg()`: 发送媒体信息列表
- `SendTorrentsMsg()`: 发送种子信息列表
- `sendRequest()`: 发送请求到 VoceChat 服务器