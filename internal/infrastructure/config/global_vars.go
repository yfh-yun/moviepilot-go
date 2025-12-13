package config

import (
	"sync"
)

// GlobalVar 全局变量管理结构体
type GlobalVar struct {
	mu sync.RWMutex
	
	// 系统停止事件
	stopEvent chan struct{}
	
	// webpush订阅
	subscriptions []map[string]interface{}
	
	// 需应急停止的工作流
	emergencyStopWorkflows map[int]bool
	
	// 需应急停止文件整理
	emergencyStopTransfer map[string]bool
}

// NewGlobalVar 创建全局变量管理实例
func NewGlobalVar() *GlobalVar {
	return &GlobalVar{
		stopEvent:              make(chan struct{}),
		subscriptions:          make([]map[string]interface{}, 0),
		emergencyStopWorkflows: make(map[int]bool),
		emergencyStopTransfer:  make(map[string]bool),
	}
}

// StopSystem 停止系统
func (g *GlobalVar) StopSystem() {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	select {
	case <-g.stopEvent:
		// 已经停止，不重复发送
	default:
		close(g.stopEvent)
	}
}

// IsSystemStopped 检查系统是否已停止
func (g *GlobalVar) IsSystemStopped() bool {
	select {
	case <-g.stopEvent:
		return true
	default:
		return false
	}
}

// GetSubscriptions 获取webpush订阅列表
func (g *GlobalVar) GetSubscriptions() []map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	// 返回副本，避免外部修改
	result := make([]map[string]interface{}, len(g.subscriptions))
	for i, sub := range g.subscriptions {
		result[i] = make(map[string]interface{})
		for k, v := range sub {
			result[i][k] = v
		}
	}
	
	return result
}

// PushSubscription 添加webpush订阅
func (g *GlobalVar) PushSubscription(subscription map[string]interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.subscriptions = append(g.subscriptions, subscription)
}

// StopWorkflow 停止工作流
func (g *GlobalVar) StopWorkflow(workflowID int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.emergencyStopWorkflows[workflowID] = true
}

// WorkflowResume 恢复工作流
func (g *GlobalVar) WorkflowResume(workflowID int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	delete(g.emergencyStopWorkflows, workflowID)
}

// IsWorkflowStopped 检查工作流是否已停止
func (g *GlobalVar) IsWorkflowStopped(workflowID int) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if g.IsSystemStopped() {
		return true
	}
	
	return g.emergencyStopWorkflows[workflowID]
}

// StopTransfer 停止文件整理
func (g *GlobalVar) StopTransfer(path string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.emergencyStopTransfer[path] = true
}

// IsTransferStopped 检查文件整理是否已停止
func (g *GlobalVar) IsTransferStopped(path string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if g.IsSystemStopped() {
		return true
	}
	
	stopped := g.emergencyStopTransfer[path]
	if stopped {
		// 消费后删除
		g.mu.RUnlock()
		g.mu.Lock()
		delete(g.emergencyStopTransfer, path)
		g.mu.Unlock()
		g.mu.RLock()
	}
	
	return stopped
}

// 全局变量实例
var (
	globalVars *GlobalVar
	once       sync.Once
)

// GetGlobalVars 获取全局变量实例（单例模式）
func GetGlobalVars() *GlobalVar {
	once.Do(func() {
		globalVars = NewGlobalVar()
	})
	
	return globalVars
}
