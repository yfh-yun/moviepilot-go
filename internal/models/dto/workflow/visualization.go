package workflow

import (
	"time"
)

// Node 工作流节点
// 用于工作流可视化，表示工作流中的一个步骤
type Node struct {
	ID             string          `json:"id"`                        // 节点ID
	Type           string          `json:"type"`                      // 节点类型：action, condition, loop, parallel, delay, notification
	Label          string          `json:"label"`                     // 节点显示名称
	Action         string          `json:"action,omitempty"`          // 动作名称（仅action类型）
	Parameters     map[string]any  `json:"parameters,omitempty"`      // 动作参数（仅action类型）
	Condition      string          `json:"condition,omitempty"`       // 条件表达式（仅condition类型）
	LoopConfig     *LoopConfig     `json:"loop_config,omitempty"`     // 循环配置（仅loop类型）
	ParallelConfig *ParallelConfig `json:"parallel_config,omitempty"` // 并行配置（仅parallel类型）
	DelayConfig    *DelayConfig    `json:"delay_config,omitempty"`    // 延迟配置（仅delay类型）
	Position       *Position       `json:"position"`                  // 节点位置（用于可视化）
	Status         string          `json:"status,omitempty"`          // 执行状态：pending, running, completed, failed, skipped
	StartTime      *time.Time      `json:"start_time,omitempty"`      // 开始执行时间
	EndTime        *time.Time      `json:"end_time,omitempty"`        // 结束执行时间
	Duration       int64           `json:"duration,omitempty"`        // 执行时长（毫秒）
	ErrorMessage   string          `json:"error_message,omitempty"`   // 错误信息（如果执行失败）
}

// LoopConfig 循环配置
type LoopConfig struct {
	LoopType  string `json:"loop_type"`           // 循环类型：for, while
	Count     int    `json:"count,omitempty"`     // 循环次数（for类型）
	Condition string `json:"condition,omitempty"` // 循环条件（while类型）
}

// ParallelConfig 并行配置
type ParallelConfig struct {
	Mode       string `json:"mode"`         // 并行模式：sync, async
	WaitForAll bool   `json:"wait_for_all"` // 是否等待所有分支完成
}

// DelayConfig 延迟配置
type DelayConfig struct {
	DelayType  string `json:"delay_type"`            // 延迟类型：fixed, random
	FixedDelay int    `json:"fixed_delay,omitempty"` // 固定延迟时间（秒，fixed类型）
	MinDelay   int    `json:"min_delay,omitempty"`   // 最小延迟时间（秒，random类型）
	MaxDelay   int    `json:"max_delay,omitempty"`   // 最大延迟时间（秒，random类型）
}

// Position 节点位置
type Position struct {
	X float64 `json:"x"` // X坐标
	Y float64 `json:"y"` // Y坐标
}

// Edge 工作流边
// 用于工作流可视化，表示节点之间的连接关系
type Edge struct {
	ID        string `json:"id"`                  // 边ID
	Source    string `json:"source"`              // 源节点ID
	Target    string `json:"target"`              // 目标节点ID
	Label     string `json:"label,omitempty"`     // 边显示名称
	Condition string `json:"condition,omitempty"` // 条件表达式（仅条件分支）
	Type      string `json:"type,omitempty"`      // 边类型：default, condition, on_success, on_failure
	Status    string `json:"status,omitempty"`    // 执行状态：not_executed, executed, skipped
}

// WorkflowGraph 工作流图
// 用于工作流可视化，表示整个工作流的结构
type WorkflowGraph struct {
	ID          string         `json:"id"`                    // 工作流ID
	Name        string         `json:"name"`                  // 工作流名称
	Description string         `json:"description,omitempty"` // 工作流描述
	Nodes       []Node         `json:"nodes"`                 // 工作流节点列表
	Edges       []Edge         `json:"edges"`                 // 工作流边列表
	Variables   map[string]any `json:"variables,omitempty"`   // 工作流变量
	CreatedAt   time.Time      `json:"created_at"`            // 创建时间
	UpdatedAt   time.Time      `json:"updated_at"`            // 更新时间
	Version     int            `json:"version"`               // 工作流版本
}

// WorkflowExecutionStatus 工作流执行状态
// 用于工作流可视化，表示工作流执行的实时状态
type WorkflowExecutionStatus struct {
	ExecutionID  string                `json:"execution_id"`            // 执行ID
	WorkflowID   string                `json:"workflow_id"`             // 工作流ID
	Status       string                `json:"status"`                  // 执行状态：pending, running, completed, failed, cancelled
	StartTime    time.Time             `json:"start_time"`              // 开始执行时间
	EndTime      *time.Time            `json:"end_time,omitempty"`      // 结束执行时间
	Duration     int64                 `json:"duration,omitempty"`      // 执行时长（毫秒）
	CurrentStep  string                `json:"current_step,omitempty"`  // 当前执行的步骤ID
	NodeStatuses map[string]NodeStatus `json:"node_statuses"`           // 节点执行状态映射
	Variables    map[string]any        `json:"variables,omitempty"`     // 工作流变量
	ErrorMessage string                `json:"error_message,omitempty"` // 错误信息（如果执行失败）
}

// NodeStatus 节点执行状态
type NodeStatus struct {
	Status       string         `json:"status"`                  // 执行状态：pending, running, completed, failed, skipped
	StartTime    *time.Time     `json:"start_time,omitempty"`    // 开始执行时间
	EndTime      *time.Time     `json:"end_time,omitempty"`      // 结束执行时间
	Duration     int64          `json:"duration,omitempty"`      // 执行时长（毫秒）
	ErrorMessage string         `json:"error_message,omitempty"` // 错误信息（如果执行失败）
	Output       map[string]any `json:"output,omitempty"`        // 执行输出
}

// WorkflowExecutionHistory 工作流执行历史
type WorkflowExecutionHistory struct {
	ExecutionID        string         `json:"execution_id"`                   // 执行ID
	WorkflowID         string         `json:"workflow_id"`                    // 工作流ID
	WorkflowName       string         `json:"workflow_name"`                  // 工作流名称
	Status             string         `json:"status"`                         // 执行状态：completed, failed, cancelled
	StartTime          time.Time      `json:"start_time"`                     // 开始执行时间
	EndTime            *time.Time     `json:"end_time"`                       // 结束执行时间
	Duration           int64          `json:"duration"`                       // 执行时长（毫秒）
	ErrorMessage       string         `json:"error_message,omitempty"`        // 错误信息（如果执行失败）
	Variables          map[string]any `json:"variables,omitempty"`            // 工作流变量
	NodeExecutionCount map[string]int `json:"node_execution_count,omitempty"` // 节点执行次数
}

// WorkflowVisualizationResponse 工作流可视化响应
type WorkflowVisualizationResponse struct {
	Graph  *WorkflowGraph           `json:"graph"`            // 工作流图结构
	Status *WorkflowExecutionStatus `json:"status,omitempty"` // 当前执行状态（如果正在执行）
}

// WorkflowExecutionHistoryResponse 工作流执行历史响应
type WorkflowExecutionHistoryResponse struct {
	Executions []WorkflowExecutionHistory `json:"executions"` // 执行历史列表
	Total      int                        `json:"total"`      // 总记录数
	Page       int                        `json:"page"`       // 当前页码
	PageSize   int                        `json:"page_size"`  // 每页记录数
}
