package dto

import ()

// Workflow 工作流信息
type Workflow struct {
	// 工作流ID
	ID int `json:"id,omitempty"`
	// 工作流名称
	Name string `json:"name,omitempty"`
	// 工作流描述
	Description string `json:"description,omitempty"`
	// 定时器
	Timer string `json:"timer,omitempty"`
	// 触发类型：timer-定时触发 event-事件触发 manual-手动触发
	TriggerType string `json:"trigger_type,omitempty"`
	// 事件类型（当trigger_type为event时使用）
	EventType string `json:"event_type,omitempty"`
	// 事件条件（JSON格式，用于过滤事件）
	EventConditions map[string]any `json:"event_conditions,omitempty"`
	// 状态
	State string `json:"state,omitempty"`
	// 已执行动作
	CurrentAction string `json:"current_action,omitempty"`
	// 任务执行结果
	Result string `json:"result,omitempty"`
	// 已执行次数
	RunCount int `json:"run_count,omitempty"`
	// 任务列表
	Actions []any `json:"actions,omitempty"`
	// 任务流
	Flows []any `json:"flows,omitempty"`
	// 创建时间
	AddTime string `json:"add_time,omitempty"`
	// 最后执行时间
	LastTime string `json:"last_time,omitempty"`
}

// ActionParams 动作基础参数
type ActionParams struct {
	// 是否需要循环
	Loop bool `json:"loop,omitempty"`
	// 循环间隔 (秒)
	LoopInterval int `json:"loop_interval,omitempty"`
}

// Action 动作信息
type Action struct {
	// 动作ID
	ID string `json:"id,omitempty"`
	// 动作类型 (类名)
	Type string `json:"type,omitempty"`
	// 动作名称
	Name string `json:"name,omitempty"`
	// 动作描述
	Description string `json:"description,omitempty"`
	// 位置
	Position map[string]any `json:"position,omitempty"`
	// 参数
	Data map[string]any `json:"data,omitempty"`
}

// ActionExecution 动作执行情况
type ActionExecution struct {
	// 当前动作（名称）
	Action string `json:"action,omitempty"`
	// 执行结果
	Result bool `json:"result,omitempty"`
	// 执行消息
	Message string `json:"message,omitempty"`
}

// ActionContext 动作基础上下文，各动作通用数据
type ActionContext struct {
	// 文本类内容
	Content string `json:"content,omitempty"`
	// 资源列表
	Torrents []*Context `json:"torrents,omitempty"`
	// 媒体列表
	Medias []*MediaInfo `json:"medias,omitempty"`
	// 文件列表
	FileItems []*FileItem `json:"fileitems,omitempty"`
	// 下载任务列表
	Downloads []*DownloadTask `json:"downloads,omitempty"`
	// 站点列表
	Sites []*Site `json:"sites,omitempty"`
	// 订阅列表
	Subscribes []*Subscribe `json:"subscribes,omitempty"`
	// 执行历史
	ExecuteHistory []*ActionExecution `json:"execute_history,omitempty"`
	// 执行进度（%）
	Progress int `json:"progress,omitempty"`
}

// ActionFlow 工作流流程
type ActionFlow struct {
	// 流程ID
	ID string `json:"id,omitempty"`
	// 源动作
	Source string `json:"source,omitempty"`
	// 目标动作
	Target string `json:"target,omitempty"`
	// 是否动画流程
	Animated bool `json:"animated,omitempty"`
}

// WorkflowShare 工作流分享信息
type WorkflowShare struct {
	// 分享ID
	ID int `json:"id,omitempty"`
	// 分享标题
	ShareTitle string `json:"share_title,omitempty"`
	// 分享说明
	ShareComment string `json:"share_comment,omitempty"`
	// 分享人
	ShareUser string `json:"share_user,omitempty"`
	// 分享人唯一ID
	ShareUID string `json:"share_uid,omitempty"`
	// 工作流名称
	Name string `json:"name,omitempty"`
	// 工作流描述
	Description string `json:"description,omitempty"`
	// 定时器
	Timer string `json:"timer,omitempty"`
	// 触发类型
	TriggerType string `json:"trigger_type,omitempty"`
	// 事件类型
	EventType string `json:"event_type,omitempty"`
	// 事件条件
	EventConditions string `json:"event_conditions,omitempty"`
	// 任务列表(JSON字符串)
	Actions string `json:"actions,omitempty"`
	// 任务流(JSON字符串)
	Flows string `json:"flows,omitempty"`
	// 执行上下文(JSON字符串)
	Context string `json:"context,omitempty"`
	// 分享时间
	Date string `json:"date,omitempty"`
	// 复用人次
	Count int `json:"count,omitempty"`
}
