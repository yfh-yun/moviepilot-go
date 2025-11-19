// Package manager 提供动作管理器的实现
package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/actions/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/actions/types"

	"go.uber.org/zap"
)

// ActionManager 动作管理器实现
type ActionManager struct {
	// 动作注册表
	actions map[string]interfaces.Action
	mutex   sync.RWMutex

	// 动作链注册表
	chains map[string]interfaces.ActionChain
	chainMutex sync.RWMutex

	// 执行状态
	executions map[int64]*ExecutionState
	execMutex  sync.RWMutex

	// 配置
	config *interfaces.ManagerConfig

	// 依赖组件
	factory   interfaces.ActionFactory
	cache     interfaces.Cache
	observers []interfaces.ActionObserver

	// 状态
	running   bool
	startTime time.Time
	stopTime  time.Time

	// 日志
	logger *zap.Logger
}

// ExecutionState 执行状态
type ExecutionState struct {
	Context    *types.ActionContext
	Status     string
	StartTime  time.Time
	UpdateTime time.Time
	CancelFunc context.CancelFunc
}

// NewActionManager 创建动作管理器
func NewActionManager(config *interfaces.ManagerConfig) *ActionManager {
	if config == nil {
		config = &interfaces.ManagerConfig{
			DefaultTimeout:    300,
			MaxConcurrency:    10,
			EnableRetry:       true,
			DefaultRetryCount: 3,
			CacheEnabled:      true,
			CacheTTL:          3600,
			MetricsEnabled:    true,
			WorkerPoolSize:    5,
			QueueSize:         100,
		}
	}

	return &ActionManager{
		actions:   make(map[string]interfaces.Action),
		chains:    make(map[string]interfaces.ActionChain),
		executions: make(map[int64]*ExecutionState),
		config:     config,
		observers:  make([]interfaces.ActionObserver, 0),
		logger:     logger.Logger,
	}
}

// RegisterAction 注册动作
func (am *ActionManager) RegisterAction(action interfaces.Action) error {
	if action == nil {
		return fmt.Errorf("动作不能为空")
	}

	actionID := action.GetActionID()
	if actionID == "" {
		return fmt.Errorf("动作ID不能为空")
	}

	am.mutex.Lock()
	defer am.mutex.Unlock()

	if _, exists := am.actions[actionID]; exists {
		return fmt.Errorf("动作 %s 已存在", actionID)
	}

	// 初始化动作
	if err := action.Initialize(); err != nil {
		return fmt.Errorf("初始化动作失败: %w", err)
	}

	am.actions[actionID] = action

	am.logger.Info("注册动作成功",
		zap.String("action_id", actionID),
		zap.String("action_name", action.Name()))

	return nil
}

// UnregisterAction 注销动作
func (am *ActionManager) UnregisterAction(actionID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	action, exists := am.actions[actionID]
	if !exists {
		return fmt.Errorf("动作 %s 不存在", actionID)
	}

	// 清理动作
	if err := action.Cleanup(); err != nil {
		am.logger.Warn("清理动作失败",
			zap.String("action_id", actionID),
			zap.Error(err))
	}

	delete(am.actions, actionID)

	am.logger.Info("注销动作成功", zap.String("action_id", actionID))
	return nil
}

// GetAction 获取动作
func (am *ActionManager) GetAction(actionID string) (interfaces.Action, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	action, exists := am.actions[actionID]
	return action, exists
}

// ListActions 列出所有动作
func (am *ActionManager) ListActions() []interfaces.Action {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	actions := make([]interfaces.Action, 0, len(am.actions))
	for _, action := range am.actions {
		actions = append(actions, action)
	}

	return actions
}

// GetActionsByType 根据类型获取动作
func (am *ActionManager) GetActionsByType(actionType string) []interfaces.Action {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var actions []interfaces.Action
	for _, action := range am.actions {
		// 这里需要动作提供类型信息，暂时使用名称匹配
		if action.Name() == actionType {
			actions = append(actions, action)
		}
	}

	return actions
}

// CreateChain 创建动作链
func (am *ActionManager) CreateChain(chainID string) interfaces.ActionChain {
	am.chainMutex.Lock()
	defer am.chainMutex.Unlock()

	if chain, exists := am.chains[chainID]; exists {
		return chain
	}

	// 创建新的动作链
	chain := NewActionChain(chainID, am)
	am.chains[chainID] = chain

	am.logger.Info("创建动作链", zap.String("chain_id", chainID))
	return chain
}

// GetChain 获取动作链
func (am *ActionManager) GetChain(chainID string) (interfaces.ActionChain, bool) {
	am.chainMutex.RLock()
	defer am.chainMutex.RUnlock()

	chain, exists := am.chains[chainID]
	return chain, exists
}

// DeleteChain 删除动作链
func (am *ActionManager) DeleteChain(chainID string) error {
	am.chainMutex.Lock()
	defer am.chainMutex.Unlock()

	if _, exists := am.chains[chainID]; !exists {
		return fmt.Errorf("动作链 %s 不存在", chainID)
	}

	delete(am.chains, chainID)

	am.logger.Info("删除动作链", zap.String("chain_id", chainID))
	return nil
}

// ListChains 列出所有动作链
func (am *ActionManager) ListChains() []string {
	am.chainMutex.RLock()
	defer am.chainMutex.RUnlock()

	chainIDs := make([]string, 0, len(am.chains))
	for chainID := range am.chains {
		chainIDs = append(chainIDs, chainID)
	}

	return chainIDs
}

// ExecuteAction 执行动作
func (am *ActionManager) ExecuteAction(ctx context.Context, actionID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error) {
	action, exists := am.GetAction(actionID)
	if !exists {
		return context, fmt.Errorf("动作 %s 不存在", actionID)
	}

	// 创建执行状态
	execState := &ExecutionState{
		Context:    context,
		Status:     "running",
		StartTime:  time.Now(),
		UpdateTime: time.Now(),
	}

	am.execMutex.Lock()
	am.executions[workflowID] = execState
	am.execMutex.Unlock()

	// 设置超时
	timeout := time.Duration(am.config.DefaultTimeout) * time.Second
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		execState.CancelFunc = cancel
	}

	// 执行动作
	result, err := action.Execute(ctx, workflowID, params, context)

	// 更新执行状态
	am.execMutex.Lock()
	if execState, exists := am.executions[workflowID]; exists {
		execState.UpdateTime = time.Now()
		if err != nil {
			execState.Status = "failed"
		} else {
			execState.Status = "completed"
		}
	}
	am.execMutex.Unlock()

	return result, err
}

// ExecuteChain 执行动作链
func (am *ActionManager) ExecuteChain(ctx context.Context, chainID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error) {
	chain, exists := am.GetChain(chainID)
	if !exists {
		return context, fmt.Errorf("动作链 %s 不存在", chainID)
	}

	// 创建执行状态
	execState := &ExecutionState{
		Context:    context,
		Status:     "running",
		StartTime:  time.Now(),
		UpdateTime: time.Now(),
	}

	am.execMutex.Lock()
	am.executions[workflowID] = execState
	am.execMutex.Unlock()

	// 设置超时
	timeout := time.Duration(am.config.DefaultTimeout) * time.Second
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		execState.CancelFunc = cancel
	}

	// 执行动作链
	result, err := chain.Execute(ctx, workflowID, params, context)

	// 更新执行状态
	am.execMutex.Lock()
	if execState, exists := am.executions[workflowID]; exists {
		execState.UpdateTime = time.Now()
		if err != nil {
			execState.Status = "failed"
		} else {
			execState.Status = "completed"
		}
	}
	am.execMutex.Unlock()

	return result, err
}

// ExecuteActionAsync 异步执行动作
func (am *ActionManager) ExecuteActionAsync(ctx context.Context, actionID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (<-chan *types.ActionContext, <-chan error) {
	resultChan := make(chan *types.ActionContext, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		result, err := am.ExecuteAction(ctx, actionID, workflowID, params, context)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	return resultChan, errorChan
}

// ExecuteChainAsync 异步执行动作链
func (am *ActionManager) ExecuteChainAsync(ctx context.Context, chainID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (<-chan *types.ActionContext, <-chan error) {
	resultChan := make(chan *types.ActionContext, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		result, err := am.ExecuteChain(ctx, chainID, workflowID, params, context)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	return resultChan, errorChan
}

// GetExecutionStatus 获取执行状态
func (am *ActionManager) GetExecutionStatus(workflowID int64) (*interfaces.ExecutionStatus, error) {
	am.execMutex.RLock()
	defer am.execMutex.RUnlock()

	execState, exists := am.executions[workflowID]
	if !exists {
		return nil, fmt.Errorf("工作流执行 %d 不存在", workflowID)
	}

	status := &interfaces.ExecutionStatus{
		WorkflowID:      workflowID,
		Status:          execState.Status,
		StartTime:       timePtr(execState.StartTime),
		UpdateTime:      timePtr(execState.UpdateTime),
		LastUpdateTime:  execState.UpdateTime.Unix(),
	}

	if execState.Context != nil {
		status.CurrentAction = execState.Context.Message
		status.Progress = execState.Context.Progress
		status.Message = execState.Context.Message
		if execState.Context.HasError() {
			status.Error = execState.Context.Error.Message
		}
		status.RetryCount = execState.Context.RetryCount
	}

	return status, nil
}

// StopExecution 停止执行
func (am *ActionManager) StopExecution(workflowID int64) error {
	am.execMutex.Lock()
	defer am.execMutex.Unlock()

	execState, exists := am.executions[workflowID]
	if !exists {
		return fmt.Errorf("工作流执行 %d 不存在", workflowID)
	}

	// 调用取消函数
	if execState.CancelFunc != nil {
		execState.CancelFunc()
	}

	execState.Status = "cancelled"
	execState.UpdateTime = time.Now()

	am.logger.Info("停止工作流执行", zap.Int64("workflow_id", workflowID))
	return nil
}

// PauseExecution 暂停执行
func (am *ActionManager) PauseExecution(workflowID int64) error {
	am.execMutex.Lock()
	defer am.execMutex.Unlock()

	execState, exists := am.executions[workflowID]
	if !exists {
		return fmt.Errorf("工作流执行 %d 不存在", workflowID)
	}

	execState.Status = "paused"
	execState.UpdateTime = time.Now()

	if execState.Context != nil {
		execState.Context.ShouldPause = true
	}

	am.logger.Info("暂停工作流执行", zap.Int64("workflow_id", workflowID))
	return nil
}

// ResumeExecution 恢复执行
func (am *ActionManager) ResumeExecution(workflowID int64) error {
	am.execMutex.Lock()
	defer am.execMutex.Unlock()

	execState, exists := am.executions[workflowID]
	if !exists {
		return fmt.Errorf("工作流执行 %d 不存在", workflowID)
	}

	execState.Status = "running"
	execState.UpdateTime = time.Now()

	if execState.Context != nil {
		execState.Context.ShouldPause = false
	}

	am.logger.Info("恢复工作流执行", zap.Int64("workflow_id", workflowID))
	return nil
}

// ListExecutions 列出所有执行
func (am *ActionManager) ListExecutions() []interfaces.ExecutionInfo {
	am.execMutex.RLock()
	defer am.execMutex.RUnlock()

	executions := make([]interfaces.ExecutionInfo, 0, len(am.executions))
	for workflowID, execState := range am.executions {
		info := interfaces.ExecutionInfo{
			WorkflowID: workflowID,
			Status:     execState.Status,
			StartTime:  execState.StartTime.Unix(),
			EndTime:    execState.UpdateTime.Unix(),
			Duration:   execState.UpdateTime.Sub(execState.StartTime).Milliseconds(),
			Success:    execState.Status == "completed",
		}

		if execState.Context != nil {
			info.Message = execState.Context.Message
			if execState.Context.HasError() {
				info.Error = execState.Context.Error.Message
			}
		}

		executions = append(executions, info)
	}

	return executions
}

// SetGlobalConfig 设置全局配置
func (am *ActionManager) SetGlobalConfig(config *interfaces.ManagerConfig) error {
	am.config = config
	am.logger.Info("更新全局配置")
	return nil
}

// GetGlobalConfig 获取全局配置
func (am *ActionManager) GetGlobalConfig() *interfaces.ManagerConfig {
	return am.config
}

// SetActionConfig 设置动作配置
func (am *ActionManager) SetActionConfig(actionID string, config map[string]interface{}) error {
	action, exists := am.GetAction(actionID)
	if !exists {
		return fmt.Errorf("动作 %s 不存在", actionID)
	}

	// 如果动作实现了配置接口，设置配置
	if baseAction, ok := action.(*BaseAction); ok {
		baseAction.SetConfig(config)
	}

	am.logger.Info("设置动作配置", zap.String("action_id", actionID))
	return nil
}

// GetActionConfig 获取动作配置
func (am *ActionManager) GetActionConfig(actionID string) (map[string]interface{}, error) {
	action, exists := am.GetAction(actionID)
	if !exists {
		return nil, fmt.Errorf("动作 %s 不存在", actionID)
	}

	// 如果动作实现了配置接口，获取配置
	if baseAction, ok := action.(*BaseAction); ok {
		return baseAction.GetConfig(), nil
	}

	return make(map[string]interface{}), nil
}

// AddObserver 添加观察者
func (am *ActionManager) AddObserver(observer interfaces.ActionObserver) error {
	am.observers = append(am.observers, observer)
	am.logger.Info("添加观察者")
	return nil
}

// RemoveObserver 移除观察者
func (am *ActionManager) RemoveObserver(observer interfaces.ActionObserver) error {
	for i, obs := range am.observers {
		if obs == observer {
			am.observers = append(am.observers[:i], am.observers[i+1:]...)
			break
		}
	}
	am.logger.Info("移除观察者")
	return nil
}

// NotifyObservers 通知观察者
func (am *ActionManager) NotifyObservers(event interfaces.ObserverEvent) {
	for _, observer := range am.observers {
		// 这里可以根据事件类型调用不同的观察者方法
		// 简化实现，直接记录日志
		am.logger.Debug("通知观察者",
			zap.String("event_type", event.Type),
			zap.String("action_id", event.ActionID),
			zap.Int64("workflow_id", event.WorkflowID))
	}
}

// SetFactory 设置工厂
func (am *ActionManager) SetFactory(factory interfaces.ActionFactory) error {
	am.factory = factory
	am.logger.Info("设置动作工厂")
	return nil
}

// GetFactory 获取工厂
func (am *ActionManager) GetFactory() interfaces.ActionFactory {
	return am.factory
}

// SetCache 设置缓存
func (am *ActionManager) SetCache(cache interfaces.Cache) error {
	am.cache = cache
	am.logger.Info("设置缓存")
	return nil
}

// GetCache 获取缓存
func (am *ActionManager) GetCache() interfaces.Cache {
	return am.cache
}

// Initialize 初始化
func (am *ActionManager) error {
	am.logger.Info("初始化动作管理器")
	return nil
}

// Start 启动管理器
func (am *ActionManager) Start(ctx context.Context) error {
	am.running = true
	am.startTime = time.Now()
	am.logger.Info("启动动作管理器")
	return nil
}

// Stop 停止管理器
func (am *ActionManager) Stop(ctx context.Context) error {
	am.running = false
	am.stopTime = time.Now()

	// 停止所有执行
	am.execMutex.Lock()
	for workflowID, execState := range am.executions {
		if execState.CancelFunc != nil {
			execState.CancelFunc()
		}
		execState.Status = "cancelled"
	}
	am.execMutex.Unlock()

	// 清理所有动作
	am.mutex.Lock()
	for actionID, action := range am.actions {
		if err := action.Cleanup(); err != nil {
			am.logger.Warn("清理动作失败",
				zap.String("action_id", actionID),
				zap.Error(err))
		}
	}
	am.mutex.Unlock()

	am.logger.Info("停止动作管理器")
	return nil
}

// Shutdown 关闭管理器
func (am *ActionManager) Shutdown(ctx context.Context) error {
	return am.Stop(ctx)
}

// HealthCheck 健康检查
func (am *ActionManager) HealthCheck() *interfaces.HealthStatus {
	checks := make(map[string]interfaces.Check)
	
	// 检查运行状态
	if am.running {
		checks["manager"] = interfaces.Check{Status: "pass", Message: "管理器运行正常"}
	} else {
		checks["manager"] = interfaces.Check{Status: "fail", Message: "管理器未运行"}
	}

	// 检查动作数量
	am.mutex.RLock()
	actionCount := len(am.actions)
	am.mutex.RUnlock()
	checks["actions"] = interfaces.Check{Status: "pass", Message: fmt.Sprintf("已注册 %d 个动作", actionCount)}

	// 检查执行状态
	am.execMutex.RLock()
	executionCount := len(am.executions)
	am.execMutex.RUnlock()
	checks["executions"] = interfaces.Check{Status: "pass", Message: fmt.Sprintf("当前 %d 个执行", executionCount)}

	status := "healthy"
	for _, check := range checks {
		if check.Status == "fail" {
			status = "unhealthy"
			break
		} else if check.Status == "warn" {
			status = "degraded"
		}
	}

	return &interfaces.HealthStatus{
		Status:     status,
		Timestamp:  time.Now().Unix(),
		Checks:     checks,
		Summary:    fmt.Sprintf("动作管理器状态: %s", status),
	}
}

// GetMetrics 获取指标
func (am *ActionManager) GetMetrics() *interfaces.ManagerMetrics {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	am.execMutex.RLock()
	defer am.execMutex.RUnlock()

	metrics := &interfaces.ManagerMetrics{
		TotalActions:     int64(len(am.actions)),
		ActiveExecutions: int64(len(am.executions)),
		LastUpdateTime:   time.Now().Unix(),
	}

	// 统计动作指标
	for _, action := range am.actions {
		stats := action.GetStats()
		metrics.CompletedActions += stats.SuccessCount
		metrics.FailedActions += stats.ErrorCount
	}

	// 计算成功率
	if metrics.TotalActions > 0 {
		metrics.SuccessRate = float64(metrics.CompletedActions) / float64(metrics.TotalActions) * 100
	}

	return metrics
}

// 辅助函数
func timePtr(t time.Time) *int64 {
	unix := t.Unix()
	return &unix
}

// BaseAction 基础动作类型声明（用于类型断言）
type BaseAction interface {
	SetConfig(config map[string]interface{})
	GetConfig() map[string]interface{}
}