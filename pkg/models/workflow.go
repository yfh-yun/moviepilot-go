package models

// Workflow 工作流信�?type Workflow struct {
	// 工作流ID
	ID int `json:"id,omitempty"`
	// 工作流名�?	Name string `json:"name,omitempty"`
	// 工作流描�?	Description string `json:"description,omitempty"`
	// 定时�?	Timer string `json:"timer,omitempty"`
	// 触发类型：timer-定时触发 event-事件触发 manual-手动触发
	TriggerType string `json:"trigger_type,omitempty"`
	// 事件类型（当trigger_type为event时使用）
	EventType string `json:"event_type,omitempty"`
	// 事件条件（JSON格式，用于过滤事件）
	EventConditions map[string]interface{} `json:"event_conditions,omitempty"`
	// 状�?	State string `json:"state,omitempty"`
	// 已执行动�?	CurrentAction string `json:"current_action,omitempty"`
	// 任务执行结果
	Result string `json:"result,omitempty"`
	// 已执行次�?	RunCount int `json:"run_count,omitempty"`
	// 任务列表
	Actions []interface{} `json:"actions,omitempty"`
	// 任务�?	Flows []interface{} `json:"flows,omitempty"`
	// 创建时间
	AddTime string `json:"add_time,omitempty"`
	// 最后执行时�?	LastTime string `json:"last_time,omitempty"`
}

// ActionParams 动作基础参数
type ActionParams struct {
	// 是否需要循�?	Loop bool `json:"loop,omitempty"`
	// 循环间隔 (�?
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
	Position map[string]interface{} `json:"position,omitempty"`
	// 参数
	Data map[string]interface{} `json:"data,omitempty"`
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
	// 文本类内�?	Content string `json:"content,omitempty"`
	// 资源列表
	Torrents []Context `json:"torrents,omitempty"`
	// 媒体列表
	Medias []MediaInfo `json:"medias,omitempty"`
	// 文件列表
	FileItems []FileItem `json:"fileitems,omitempty"`
	// 下载任务列表
	Downloads []DownloadTask `json:"downloads,omitempty"`
	// 站点列表
	Sites []Site `json:"sites,omitempty"`
	// 订阅列表
	Subscribes []Subscribe `json:"subscribes,omitempty"`
	// 执行历史
	ExecuteHistory []ActionExecution `json:"execute_history,omitempty"`
	// 执行进度�?�?	Progress int `json:"progress,omitempty"`
}

// ActionFlow 工作流流�?type ActionFlow struct {
	// 流程ID
	ID string `json:"id,omitempty"`
	// 源动�?	Source string `json:"source,omitempty"`
	// 目标动作
	Target string `json:"target,omitempty"`
	// 是否动画流程
	Animated bool `json:"animated,omitempty"`
}

// WorkflowShare 工作流分享信�?type WorkflowShare struct {
	// 分享ID
	ID int `json:"id,omitempty"`
	// 分享标题
	ShareTitle string `json:"share_title,omitempty"`
	// 分享说明
	ShareComment string `json:"share_comment,omitempty"`
	// 分享�?	ShareUser string `json:"share_user,omitempty"`
	// 分享人唯一ID
	ShareUID string `json:"share_uid,omitempty"`
	// 工作流名�?	Name string `json:"name,omitempty"`
	// 工作流描�?	Description string `json:"description,omitempty"`
	// 定时�?	Timer string `json:"timer,omitempty"`
	// 触发类型
	TriggerType string `json:"trigger_type,omitempty"`
	// 事件类型
	EventType string `json:"event_type,omitempty"`
	// 事件条件
	EventConditions string `json:"event_conditions,omitempty"`
	// 任务列表(JSON字符�?
	Actions string `json:"actions,omitempty"`
	// 任务�?JSON字符�?
	Flows string `json:"flows,omitempty"`
	// 执行上下�?JSON字符�?
	Context string `json:"context,omitempty"`
	// 分享时间
	Date string `json:"date,omitempty"`
	// 复用人次
	Count int `json:"count,omitempty"`
}

// NewWorkflow 创建一个新�?Workflow 实例
func NewWorkflow() *Workflow {
	return &Workflow{
		TriggerType:     "timer",
		EventConditions: make(map[string]interface{}),
		RunCount:        0,
		Actions:         make([]interface{}, 0),
		Flows:           make([]interface{}, 0),
	}
}

// NewActionParams 创建一个新�?ActionParams 实例
func NewActionParams() *ActionParams {
	return &ActionParams{
		Loop:        false,
		LoopInterval: 0,
	}
}

// NewAction 创建一个新�?Action 实例
func NewAction() *Action {
	return &Action{
		Position: make(map[string]interface{}),
		Data:     make(map[string]interface{}),
	}
}

// NewActionExecution 创建一个新�?ActionExecution 实例
func NewActionExecution() *ActionExecution {
	return &ActionExecution{}
}

// NewActionContext 创建一个新�?ActionContext 实例
func NewActionContext() *ActionContext {
	return &ActionContext{
		Torrents:       make([]Context, 0),
		Medias:         make([]MediaInfo, 0),
		FileItems:      make([]FileItem, 0),
		Downloads:      make([]DownloadTask, 0),
		Sites:          make([]Site, 0),
		Subscribes:     make([]Subscribe, 0),
		ExecuteHistory: make([]ActionExecution, 0),
		Progress:       0,
	}
}

// NewActionFlow 创建一个新�?ActionFlow 实例
func NewActionFlow() *ActionFlow {
	return &ActionFlow{
		Animated: true,
	}
}

// NewWorkflowShare 创建一个新�?WorkflowShare 实例
func NewWorkflowShare() *WorkflowShare {
	return &WorkflowShare{
		Count: 0,
	}
}
