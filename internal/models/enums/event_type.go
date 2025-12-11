package enums

// EventType 异步广播事件
type EventType string

const (
	// 插件需要重载
	EventTypePluginReload EventType = "plugin.reload"
	// 触发插件动作
	EventTypePluginAction EventType = "plugin.action"
	// 插件触发事件
	EventTypePluginTriggered EventType = "plugin.triggered"
	// 执行命令
	EventTypeCommandExecute EventType = "command.excute"
	// 站点已删除
	EventTypeSiteDeleted EventType = "site.deleted"
	// 站点已更新
	EventTypeSiteUpdated EventType = "site.updated"
	// 站点已刷新
	EventTypeSiteRefreshed EventType = "site.refreshed"
	// 转移完成
	EventTypeTransferComplete EventType = "transfer.complete"
	// 下载已添加
	EventTypeDownloadAdded EventType = "download.added"
	// 删除历史记录
	EventTypeHistoryDeleted EventType = "history.deleted"
	// 删除下载源文件
	EventTypeDownloadFileDeleted EventType = "downloadfile.deleted"
	// 删除下载任务
	EventTypeDownloadDeleted EventType = "download.deleted"
	// 收到用户外来消息
	EventTypeUserMessage EventType = "user.message"
	// 收到Webhook消息
	EventTypeWebhookMessage EventType = "webhook.message"
	// 发送消息通知
	EventTypeNoticeMessage EventType = "notice.message"
	// 订阅已添加
	EventTypeSubscribeAdded EventType = "subscribe.added"
	// 订阅已调整
	EventTypeSubscribeModified EventType = "subscribe.modified"
	// 订阅已删除
	EventTypeSubscribeDeleted EventType = "subscribe.deleted"
	// 订阅已完成
	EventTypeSubscribeComplete EventType = "subscribe.complete"
	// 系统错误
	EventTypeSystemError EventType = "system.error"
	// 刮削元数据
	EventTypeMetadataScrape EventType = "metadata.scrape"
	// 模块需要重载
	EventTypeModuleReload EventType = "module.reload"
	// 配置项更新
	EventTypeConfigChanged EventType = "config.updated"
	// 消息交互动作
	EventTypeMessageAction EventType = "message.action"
	// 执行工作流
	EventTypeWorkflowExecute EventType = "workflow.execute"
)

// EventTypeNames EventType中文名称翻译字典
var EventTypeNames = map[EventType]string{
	EventTypePluginReload:        "插件重载",
	EventTypePluginAction:        "触发插件动作",
	EventTypePluginTriggered:     "触发插件事件",
	EventTypeCommandExecute:      "执行命令",
	EventTypeSiteDeleted:         "站点已删除",
	EventTypeSiteUpdated:         "站点已更新",
	EventTypeSiteRefreshed:       "站点已刷新",
	EventTypeTransferComplete:    "整理完成",
	EventTypeDownloadAdded:       "添加下载",
	EventTypeHistoryDeleted:      "删除历史记录",
	EventTypeDownloadFileDeleted: "删除下载源文件",
	EventTypeDownloadDeleted:     "删除下载任务",
	EventTypeUserMessage:         "收到用户消息",
	EventTypeWebhookMessage:      "收到Webhook消息",
	EventTypeNoticeMessage:       "发送消息通知",
	EventTypeSubscribeAdded:      "添加订阅",
	EventTypeSubscribeModified:   "订阅已调整",
	EventTypeSubscribeDeleted:    "订阅已删除",
	EventTypeSubscribeComplete:   "订阅已完成",
	EventTypeSystemError:         "系统错误",
	EventTypeMetadataScrape:      "刮削元数据",
	EventTypeModuleReload:        "模块重载",
	EventTypeConfigChanged:       "配置项更新",
	EventTypeMessageAction:       "消息交互动作",
	EventTypeWorkflowExecute:     "执行工作流",
}

// ChainEventType 同步链式事件
type ChainEventType string

const (
	// 名称识别
	ChainEventTypeNameRecognize ChainEventType = "name.recognize"
	// 认证验证
	ChainEventTypeAuthVerification ChainEventType = "auth.verification"
	// 认证拦截
	ChainEventTypeAuthIntercept ChainEventType = "auth.intercept"
	// 命令注册
	ChainEventTypeCommandRegister ChainEventType = "command.register"
	// 整理重命名
	ChainEventTypeTransferRename ChainEventType = "transfer.rename"
	// 整理拦截
	ChainEventTypeTransferIntercept ChainEventType = "transfer.intercept"
	// 资源选择
	ChainEventTypeResourceSelection ChainEventType = "resource.selection"
	// 资源下载
	ChainEventTypeResourceDownload ChainEventType = "resource.download"
	// 探索数据源
	ChainEventTypeDiscoverSource ChainEventType = "discover.source"
	// 媒体识别转换
	ChainEventTypeMediaRecognizeConvert ChainEventType = "media.recognize.convert"
	// 推荐数据源
	ChainEventTypeRecommendSource ChainEventType = "recommend.source"
	// 工作流执行
	ChainEventTypeWorkflowExecution ChainEventType = "workflow.execution"
	// 存储操作选择
	ChainEventTypeStorageOperSelection ChainEventType = "storage.operation"
)
