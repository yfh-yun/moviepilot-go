# Transmission 模块

Transmission 模块是 MoviePilot 的一个下载器模块，用于与 Transmission 下载器进行交互。

## 功能特性

- 种子管理（添加、删除、启动、停止）
- 种子状态查询
- 种子文件选择
- 速度限制设置
- 标签管理
- 会话信息获取
- Tracker管理

## 配置参数

- `host`: Transmission服务端地址
- `port`: Transmission服务端端口
- `username`: 用户名
- `password`: 密码

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

### 下载器方法
- `Download()`: 添加下载任务
- `ListTorrents()`: 获取种子列表
- `TransferCompleted()`: 转移完成后的处理
- `RemoveTorrents()`: 删除种子
- `StartTorrents()`: 启动种子
- `StopTorrents()`: 停止种子
- `TorrentFiles()`: 获取种子文件列表
- `DownloaderInfo()`: 获取下载器信息

## Transmission 客户端

模块包含一个 Transmission 客户端实现，提供以下功能：

- `IsInactive()`: 判断是否需要重连
- `Reconnect()`: 重连
- `GetTorrents()`: 获取种子列表
- `GetCompletedTorrents()`: 获取已完成的种子列表
- `GetDownloadingTorrents()`: 获取正在下载的种子列表
- `SetTorrentTag()`: 设置种子标签
- `GetTorrentTags()`: 获取种子标签
- `AddTorrent()`: 添加种子
- `StartTorrents()`: 启动种子
- `StopTorrents()`: 停止种子
- `DeleteTorrents()`: 删除种子
- `GetFiles()`: 获取种子文件列表
- `SetFiles()`: 设置下载文件状态
- `SetUnwantedFiles()`: 设置不想要的文件
- `TransferInfo()`: 获取传输信息
- `SetSpeedLimit()`: 设置速度限制
- `GetSpeedLimit()`: 获取速度限制
- `RecheckTorrents()`: 重新校验种子
- `ChangeTorrent()`: 修改种子参数
- `UpdateTracker()`: 更新Tracker
- `GetSession()`: 获取会话信息