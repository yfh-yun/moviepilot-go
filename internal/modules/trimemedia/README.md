# TrimeMedia 飞牛影视模块

TrimeMedia 模块是 MoviePilot 的一个媒体服务器模块，用于与飞牛影视（TrimeMedia）服务进行交互。

## 功能特性

- 媒体库管理
- 媒体文件搜索和识别
- 用户认证
- 媒体播放控制
- 媒体统计信息
- 定时任务重连
- Webhook 消息处理

## 配置参数

- `host`: 飞牛影视服务端地址
- `username`: 用户名
- `password`: 密码
- `play_host`: 外网播放地址（可选）
- `sync_libraries`: 同步的媒体库列表（可选）

## 实现细节

该模块实现了以下接口方法：

### 模块基础方法
- `InitModule()`: 初始化模块
- `HandleConfigChanged()`: 处理配置变更事件
- `GetName()`: 获取模块名称
- `GetType()`: 获取模块类型
- `GetSubType()`: 获取模块子类型
- `GetPriority()`: 获取模块优先级
- `Stop()`: 停止模块
- `Test()`: 测试模块连接性
- `SchedulerJob()`: 定时任务

### 媒体服务器方法
- `UserAuthenticate()`: 用户认证
- `WebhookParser()`: 解析Webhook报文体
- `MediaExists()`: 判断媒体文件是否存在
- `MediaStatistic()`: 媒体数量统计
- `MediaServerLibrarys()`: 媒体库列表
- `MediaServerItems()`: 获取媒体服务器项目列表
- `MediaServerItemInfo()`: 媒体库项目详情
- `MediaServerTVEpisodes()`: 获取剧集信息
- `MediaServerPlaying()`: 获取媒体服务器正在播放信息
- `MediaServerPlayURL()`: 获取媒体库播放地址
- `MediaServerLatest()`: 获取媒体服务器最新入库条目
- `MediaServerLatestImages()`: 获取媒体服务器最新入库条目的图片

## TrimeMedia 客户端

模块包含一个 TrimeMedia 客户端实现，提供以下功能：

- `IsConfigured()`: 是否已配置
- `IsAuthenticated()`: 是否已登录
- `IsInactive()`: 判断是否需要重连
- `Reconnect()`: 重连
- `Disconnect()`: 断开连接
- `GetLibrarys()`: 获取媒体库列表
- `GetUserCount()`: 获取用户数量
- `GetMediasCount()`: 获取媒体数量统计
- `Authenticate()`: 用户认证
- `GetMovies()`: 获取电影列表
- `GetTVEpisodes()`: 获取剧集列表
- `RefreshRootLibrary()`: 刷新整个媒体库
- `RefreshLibraryByItems()`: 按路径刷新媒体库
- `GetItemInfo()`: 获取项目详情
- `GetItems()`: 获取媒体项目列表
- `GetPlayURL()`: 获取播放链接
- `GetResume()`: 获取继续观看列表
- `GetLatest()`: 获取最近更新列表
- `GetLatestBackdrops()`: 获取最近更新的媒体背景图片