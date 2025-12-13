package workflows

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"moviepilot-go/pkg/cache"
)

// WorkflowManager 工作流管理器，对应Python中的WorkFlowManager类
// 实现工作流的注册、执行、事件处理等功能
// 添加缓存支持，提高性能
type WorkflowManager struct {
	// 所有已注册的工作流
	workflows map[string]Workflow
	
	// 事件到工作流的映射
	eventWorkflows map[string][]string
	
	// 工作流执行结果缓存
	resultsCache cache.CacheBackend
	
	// 工作流列表缓存
	listCache cache.CacheBackend
	
	// 互斥锁，保护并发访问
	mutex sync.RWMutex
	
	// 日志记录器
	logger *zap.Logger
}

// NewWorkflowManager 创建工作流管理器实例
func NewWorkflowManager(logger *zap.Logger) *WorkflowManager {
	// 创建缓存实例
	resultsCache := cache.NewMemoryBackend(cache.Config{
		MaxSize:    1000,
		DefaultTTL: 30 * time.Minute,
	})
	
	listCache := cache.NewMemoryBackend(cache.Config{
		MaxSize:    100,
		DefaultTTL: 5 * time.Minute,
	})
	
	return &WorkflowManager{
		workflows:      make(map[string]Workflow),
		eventWorkflows: make(map[string][]string),
		resultsCache:   resultsCache,
		listCache:      listCache,
		logger:         logger,
	}
}

// RegisterWorkflow 注册工作流
func (wm *WorkflowManager) RegisterWorkflow(workflow Workflow) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	workflowID := workflow.GetID()
	wm.workflows[workflowID] = workflow
	
	wm.logger.Info("Workflow registered", zap.String("workflow_id", workflowID), zap.String("workflow_name", workflow.GetName()))
	
	// 注册后清除列表缓存
	wm.listCache.Clear("workflows")
	
	return nil
}

// GetWorkflow 获取工作流
func (wm *WorkflowManager) GetWorkflow(workflowID string) (Workflow, error) {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	workflow, exists := wm.workflows[workflowID]
	if !exists {
		return nil, nil
	}
	
	return workflow, nil
}

// ExecuteWorkflow 执行工作流
func (wm *WorkflowManager) ExecuteWorkflow(ctx WorkflowContext) (*WorkflowResult, error) {
	// 尝试从缓存获取结果
	cacheKey := "result:" + ctx.WorkflowID
	if result, found, err := wm.resultsCache.Get(cacheKey, "workflow_results"); err == nil && found {
		if cachedResult, ok := result.(*WorkflowResult); ok {
			wm.logger.Debug("Using cached workflow result", zap.String("workflow_id", ctx.WorkflowID))
			return cachedResult, nil
		}
	}
	
	// 缓存未命中，执行工作流
	workflow, err := wm.GetWorkflow(ctx.WorkflowID)
	if err != nil {
		return nil, err
	}
	
	if workflow == nil {
		return nil, nil
	}
	
	// 初始化工作流
	if initErr := workflow.Initialize(ctx); initErr != nil {
		return nil, initErr
	}
	
	// 执行工作流
	result, err := workflow.Execute(ctx)
	
	// 缓存执行结果
	if result != nil {
		wm.resultsCache.Set(cacheKey, result, 30*time.Minute, "workflow_results")
	}
	
	return result, err
}

// ExecuteWorkflowAsync 异步执行工作流
func (wm *WorkflowManager) ExecuteWorkflowAsync(ctx WorkflowContext) (string, error) {
	// 异步执行，不缓存结果
	go func() {
		_, err := wm.ExecuteWorkflow(ctx)
		if err != nil {
			wm.logger.Error("Async workflow execution failed", zap.String("workflow_id", ctx.WorkflowID), zap.String("error", err.Error()))
		}
	}()
	
	return ctx.WorkflowID, nil
}

// ListWorkflows 列出所有工作流
func (wm *WorkflowManager) ListWorkflows(params ListWorkflowsParams) ([]Workflow, error) {
	// 尝试从缓存获取列表
	cacheKey := "list:" + params.Status + ":" + params.Priority + ":" + params.Type
	if cachedList, found, err := wm.listCache.Get(cacheKey, "workflow_lists"); err == nil && found {
		if workflows, ok := cachedList.([]Workflow); ok {
			wm.logger.Debug("Using cached workflow list")
			return workflows, nil
		}
	}
	
	// 缓存未命中，构建列表
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	var workflows []Workflow
	
	// 遍历所有工作流
	for _, workflow := range wm.workflows {
		// 应用过滤条件
		if (params.Status == "" || workflow.GetStatus() == params.Status) &&
		   (params.Priority == "" || workflow.GetPriority() == params.Priority) &&
		   (params.Type == "" || workflow.GetType() == params.Type) {
			workflows = append(workflows, workflow)
		}
	}
	
	// 应用分页
	if params.Limit > 0 {
		end := params.Offset + params.Limit
		if end > len(workflows) {
			end = len(workflows)
		}
		
		if params.Offset < len(workflows) {
			workflows = workflows[params.Offset:end]
		} else {
			workflows = []Workflow{}
		}
	}
	
	// 缓存结果
	wm.listCache.Set(cacheKey, workflows, 5*time.Minute, "workflow_lists")
	
	return workflows, nil
}

// GetWorkflowResults 获取工作流执行结果
func (wm *WorkflowManager) GetWorkflowResults(workflowID string) (*WorkflowResult, error) {
	// 尝试从缓存获取结果
	cacheKey := "result:" + workflowID
	if result, found, err := wm.resultsCache.Get(cacheKey, "workflow_results"); err == nil && found {
		if cachedResult, ok := result.(*WorkflowResult); ok {
			wm.logger.Debug("Using cached workflow result", zap.String("workflow_id", workflowID))
			return cachedResult, nil
		}
	}
	
	return nil, nil
}

// RegisterWorkflowEvent 注册工作流事件触发器
func (wm *WorkflowManager) RegisterWorkflowEvent(workflowID, eventType string) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	// 添加到事件到工作流的映射
	if _, exists := wm.eventWorkflows[eventType]; !exists {
		wm.eventWorkflows[eventType] = []string{}
	}
	
	// 检查是否已经注册
	for _, id := range wm.eventWorkflows[eventType] {
		if id == workflowID {
			return nil
		}
	}
	
	// 添加到映射
	wm.eventWorkflows[eventType] = append(wm.eventWorkflows[eventType], workflowID)
	wm.logger.Info("Workflow event registered", zap.String("workflow_id", workflowID), zap.String("event_type", eventType))
	
	return nil
}

// RemoveWorkflowEvent 移除工作流事件触发器
func (wm *WorkflowManager) RemoveWorkflowEvent(workflowID, eventType string) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	// 检查事件是否存在
	if _, exists := wm.eventWorkflows[eventType]; !exists {
		return nil
	}
	
	// 查找并移除工作流ID
	workflows := wm.eventWorkflows[eventType]
	for i, id := range workflows {
		if id == workflowID {
			wm.eventWorkflows[eventType] = append(workflows[:i], workflows[i+1:]...)
			break
		}
	}
	
	// 如果事件下没有工作流了，删除事件
	if len(wm.eventWorkflows[eventType]) == 0 {
		delete(wm.eventWorkflows, eventType)
	}
	
	wm.logger.Info("Workflow event removed", zap.String("workflow_id", workflowID), zap.String("event_type", eventType))
	
	return nil
}

// HandleEvent 处理事件，触发相应的工作流
func (wm *WorkflowManager) HandleEvent(eventType string, eventData map[string]interface{}) error {
	wm.mutex.RLock()
	// 复制工作流ID列表，避免锁竞争
	var workflowIDs []string
	if ids, exists := wm.eventWorkflows[eventType]; exists {
		workflowIDs = make([]string, len(ids))
		copy(workflowIDs, ids)
	}
	wm.mutex.RUnlock()
	
	if len(workflowIDs) == 0 {
		return nil
	}
	
	// 触发所有相关工作流
	for _, workflowID := range workflowIDs {
		go func(id string) {
			// 检查工作流是否存在且可用
			workflow, err := wm.GetWorkflow(id)
			if err != nil || workflow == nil {
				return
			}
			
			// 创建工作流上下文
			ctx := WorkflowContext{
				Context:       context.Background(),
				Logger:        wm.logger,
				WorkflowID:    id,
				WorkflowName:  workflow.GetName(),
				Input:         eventData,
				GlobalContext: map[string]any{},
				Services:      map[string]any{},
				Priority:      workflow.GetPriority(),
				Type:          workflow.GetType(),
			}
			
			// 异步执行工作流
			if _, err := wm.ExecuteWorkflowAsync(ctx); err != nil {
				wm.logger.Error("Failed to trigger workflow on event", 
					zap.String("workflow_id", id),
					zap.String("event_type", eventType),
					zap.String("error", err.Error()))
			}
		}(workflowID)
	}
	
	return nil
}

// ClearCache 清除缓存
func (wm *WorkflowManager) ClearCache() error {
	// 清除所有缓存
	if err := wm.resultsCache.Clear("workflow_results"); err != nil {
		return err
	}
	
	if err := wm.listCache.Clear("workflow_lists"); err != nil {
		return err
	}
	
	wm.logger.Info("Workflow cache cleared")
	return nil
}

// ClearWorkflowCache 清除特定工作流的缓存
func (wm *WorkflowManager) ClearWorkflowCache(workflowID string) error {
	// 清除特定工作流的结果缓存
	cacheKey := "result:" + workflowID
	if err := wm.resultsCache.Delete(cacheKey, "workflow_results"); err != nil {
		return err
	}
	
	// 清除列表缓存，因为工作流状态可能已改变
	if err := wm.listCache.Clear("workflows"); err != nil {
		return err
	}
	
	wm.logger.Info("Workflow cache cleared for specific workflow", zap.String("workflow_id", workflowID))
	return nil
}

// Stop 停止工作流管理器
func (wm *WorkflowManager) Stop() error {
	// 关闭缓存
	if err := wm.resultsCache.Close(); err != nil {
		return err
	}
	
	if err := wm.listCache.Close(); err != nil {
		return err
	}
	
	// 清空映射
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	wm.workflows = make(map[string]Workflow)
	wm.eventWorkflows = make(map[string][]string)
	
	wm.logger.Info("Workflow manager stopped")
	return nil
}
