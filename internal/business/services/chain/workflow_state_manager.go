package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/models"
	"go.uber.org/zap"
)

// WorkflowStateManager 工作流状态管理器
type WorkflowStateManager struct {
	logger     *zap.Logger
	states     map[string]*model.WorkflowState
	mutex      sync.RWMutex
	history    map[string][]*model.WorkflowStateSnapshot
	maxHistory int
}

// NewWorkflowStateManager 创建工作流状态管理器
func NewWorkflowStateManager(logger *zap.Logger) *WorkflowStateManager {
	return &WorkflowStateManager{
		logger:     logger,
		states:     make(map[string]*model.WorkflowState),
		history:    make(map[string][]*model.WorkflowStateSnapshot),
		maxHistory: 100, // 最大历史记录数
	}
}

// CreateState 创建工作流状态
func (m *WorkflowStateManager) CreateState(workflowID, instanceID string) *model.WorkflowState {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	state := &model.WorkflowState{
		WorkflowID:    workflowID,
		InstanceID:    instanceID,
		Status:        model.WorkflowStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CurrentStep:   "",
		Variables:     make(map[string]interface{}),
		Metrics:       make(map[string]interface{}),
		Error:         nil,
	}

	m.states[instanceID] = state
	m.history[instanceID] = append(m.history[instanceID], m.createSnapshot(state))

	m.logger.Info("创建工作流状态",
		zap.String("workflow_id", workflowID),
		zap.String("instance_id", instanceID),
		zap.String("status", string(state.Status)))

	return state
}

// GetState 获取工作流状态
func (m *WorkflowStateManager) GetState(instanceID string) (*model.WorkflowState, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.states[instanceID]
	return state, exists
}

// UpdateState 更新工作流状态
func (m *WorkflowStateManager) UpdateState(instanceID string, updateFn func(*model.WorkflowState)) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	state, exists := m.states[instanceID]
	if !exists {
		return fmt.Errorf("工作流状态不存在: %s", instanceID)
	}

	// 保存更新前的状态快照
	oldStatus := state.Status
	oldStep := state.CurrentStep

	// 执行状态更新
	updateFn(state)
	state.UpdatedAt = time.Now()

	// 记录状态变更
	m.logger.Debug("更新工作流状态",
		zap.String("instance_id", instanceID),
		zap.String("old_status", string(oldStatus)),
		zap.String("new_status", string(state.Status)),
		zap.String("old_step", oldStep),
		zap.String("new_step", state.CurrentStep))

	// 创建状态快照
	snapshot := m.createSnapshot(state)
	m.history[instanceID] = append(m.history[instanceID], snapshot)

	// 限制历史记录数量
	if len(m.history[instanceID]) > m.maxHistory {
		m.history[instanceID] = m.history[instanceID][1:]
	}

	return nil
}

// SetStatus 设置工作流状态
func (m *WorkflowStateManager) SetStatus(instanceID string, status model.WorkflowStatus, reason string) error {
	return m.UpdateState(instanceID, func(state *model.WorkflowState) {
		state.Status = status
		if reason != "" {
			state.StatusReason = reason
		}
	})
}

// SetCurrentStep 设置当前步骤
func (m *WorkflowStateManager) SetCurrentStep(instanceID, stepID string) error {
	return m.UpdateState(instanceID, func(state *model.WorkflowState) {
		state.CurrentStep = stepID
	})
}

// SetVariable 设置变量
func (m *WorkflowStateManager) SetVariable(instanceID, key string, value interface{}) error {
	return m.UpdateState(instanceID, func(state *model.WorkflowState) {
		if state.Variables == nil {
			state.Variables = make(map[string]interface{})
		}
		state.Variables[key] = value
	})
}

// GetVariable 获取变量
func (m *WorkflowStateManager) GetVariable(instanceID, key string) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.states[instanceID]
	if !exists || state.Variables == nil {
		return nil, false
	}

	value, exists := state.Variables[key]
	return value, exists
}

// SetError 设置错误信息
func (m *WorkflowStateManager) SetError(instanceID string, err error) error {
	return m.UpdateState(instanceID, func(state *model.WorkflowState) {
		state.Status = model.WorkflowStatusFailed
		state.Error = err
		state.FailedAt = time.Now()
	})
}

// SetProgress 设置进度
func (m *WorkflowStateManager) SetProgress(instanceID string, completed, total int) error {
	return m.UpdateState(instanceID, func(state *model.WorkflowState) {
		state.Progress.Completed = completed
		state.Progress.Total = total
		if total > 0 {
			state.Progress.Percentage = float64(completed) / float64(total) * 100
		}
	})
}

// SetMetric 设置指标
func (m *WorkflowStateManager) SetMetric(instanceID, key string, value interface{}) error {
	return m.UpdateState(instanceID, func(state *model.WorkflowState) {
		if state.Metrics == nil {
			state.Metrics = make(map[string]interface{})
		}
		state.Metrics[key] = value
	})
}

// GetHistory 获取状态历史
func (m *WorkflowStateManager) GetHistory(instanceID string) []*model.WorkflowStateSnapshot {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history, exists := m.history[instanceID]
	if !exists {
		return nil
	}

	// 返回副本，避免外部修改
	snapshots := make([]*model.WorkflowStateSnapshot, len(history))
	copy(snapshots, history)
	return snapshots
}

// ListStates 列出所有状态
func (m *WorkflowStateManager) ListStates() []*model.WorkflowState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	states := make([]*model.WorkflowState, 0, len(m.states))
	for _, state := range m.states {
		states = append(states, state)
	}

	return states
}

// ListStatesByStatus 按状态列出工作流
func (m *WorkflowStateManager) ListStatesByStatus(status model.WorkflowStatus) []*model.WorkflowState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	states := make([]*model.WorkflowState, 0)
	for _, state := range m.states {
		if state.Status == status {
			states = append(states, state)
		}
	}

	return states
}

// Cleanup 清理过期状态
func (m *WorkflowStateManager) Cleanup(maxAge time.Duration) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for instanceID, state := range m.states {
		if state.UpdatedAt.Before(cutoff) && 
			(state.Status == model.WorkflowStatusCompleted || 
			 state.Status == model.WorkflowStatusFailed ||
			 state.Status == model.WorkflowStatusCancelled) {
			
			delete(m.states, instanceID)
			delete(m.history, instanceID)
			removed++

			m.logger.Info("清理过期工作流状态",
				zap.String("instance_id", instanceID),
				zap.String("status", string(state.Status)),
				zap.Time("updated_at", state.UpdatedAt))
		}
	}

	return removed
}

// createSnapshot 创建状态快照
func (m *WorkflowStateManager) createSnapshot(state *model.WorkflowState) *model.WorkflowStateSnapshot {
	// 深拷贝变量
	variables := make(map[string]interface{})
	for k, v := range state.Variables {
		variables[k] = v
	}

	// 深拷贝指标
	metrics := make(map[string]interface{})
	for k, v := range state.Metrics {
		metrics[k] = v
	}

	return &model.WorkflowStateSnapshot{
		InstanceID:   state.InstanceID,
		WorkflowID:   state.WorkflowID,
		Status:       state.Status,
		StatusReason: state.StatusReason,
		CurrentStep:  state.CurrentStep,
		Variables:    variables,
		Metrics:      metrics,
		Progress:     state.Progress,
		Error:        state.Error,
		Timestamp:    state.UpdatedAt,
	}
}

// WorkflowErrorHandler 工作流错误处理器
type WorkflowErrorHandler struct {
	logger      *zap.Logger
	stateMgr    *WorkflowStateManager
	actionChain *ActionChain
	retryConfig *RetryConfig
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int
	RetryDelay    time.Duration
	BackoffFactor float64
	MaxDelay      time.Duration
}

// NewWorkflowErrorHandler 创建工作流错误处理器
func NewWorkflowErrorHandler(logger *zap.Logger, stateMgr *WorkflowStateManager, actionChain *ActionChain) *WorkflowErrorHandler {
	return &WorkflowErrorHandler{
		logger:   logger,
		stateMgr: stateMgr,
		actionChain: actionChain,
		retryConfig: &RetryConfig{
			MaxRetries:    3,
			RetryDelay:    time.Second * 5,
			BackoffFactor: 2.0,
			MaxDelay:      time.Minute * 5,
		},
	}
}

// HandleError 处理工作流错误
func (h *WorkflowErrorHandler) HandleError(ctx context.Context, instanceID string, stepID string, err error) error {
	h.logger.Error("处理工作流错误",
		zap.String("instance_id", instanceID),
		zap.String("step_id", stepID),
		zap.Error(err))

	// 记录错误到状态
	if stateErr := h.stateMgr.SetError(instanceID, err); stateErr != nil {
		h.logger.Error("设置工作流错误状态失败", zap.Error(stateErr))
	}

	// 检查是否需要重试
	if h.shouldRetry(err) {
		return h.executeRetry(ctx, instanceID, stepID, err)
	}

	// 执行错误处理策略
	return h.executeErrorStrategy(ctx, instanceID, stepID, err)
}

// shouldRetry 判断是否应该重试
func (h *WorkflowErrorHandler) shouldRetry(err error) bool {
	// 获取重试次数
	state, exists := h.stateMgr.GetState(getCurrentInstanceID())
	if !exists {
		return false
	}

	retryCount, _ := state.GetMetric("retry_count").(int)
	return retryCount < h.retryConfig.MaxRetries
}

// executeRetry 执行重试
func (h *WorkflowErrorHandler) executeRetry(ctx context.Context, instanceID, stepID string, originalErr error) error {
	// 获取当前重试次数
	state, _ := h.stateMgr.GetState(instanceID)
	retryCount, _ := state.GetMetric("retry_count").(int)
	retryCount++

	// 计算延迟时间
	delay := time.Duration(float64(h.retryConfig.RetryDelay) * 
		(1 << (retryCount - 1))) // 指数退避
	if delay > h.retryConfig.MaxDelay {
		delay = h.retryConfig.MaxDelay
	}

	// 更新重试信息
	h.stateMgr.SetMetric(instanceID, "retry_count", retryCount)
	h.stateMgr.SetMetric(instanceID, "retry_delay", delay)
	h.stateMgr.SetStatus(instanceID, model.WorkflowStatusRetrying, 
		fmt.Sprintf("第%d次重试步骤: %s", retryCount, stepID))

	h.logger.Info("执行工作流重试",
		zap.String("instance_id", instanceID),
		zap.String("step_id", stepID),
		zap.Int("retry_count", retryCount),
		zap.Duration("delay", delay))

	// 等待延迟时间
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
	}

	// 重置状态为运行中
	h.stateMgr.SetStatus(instanceID, model.WorkflowStatusRunning, "重试执行中")

	// 重新执行步骤
	workflowCtx := &model.WorkflowContext{
		InstanceID: instanceID,
		Variables:  state.Variables,
	}

	return h.actionChain.ExecuteAction(ctx, stepID, workflowCtx)
}

// executeErrorStrategy 执行错误处理策略
func (h *WorkflowErrorHandler) executeErrorStrategy(ctx context.Context, instanceID, stepID string, err error) error {
	// 根据错误类型执行不同的处理策略
	switch {
	case isTemporaryError(err):
		return h.handleTemporaryError(ctx, instanceID, stepID, err)
	case isBusinessError(err):
		return h.handleBusinessError(ctx, instanceID, stepID, err)
	case isSystemError(err):
		return h.handleSystemError(ctx, instanceID, stepID, err)
	default:
		return h.handleUnknownError(ctx, instanceID, stepID, err)
	}
}

// handleTemporaryError 处理临时错误
func (h *WorkflowErrorHandler) handleTemporaryError(ctx context.Context, instanceID, stepID string, err error) error {
	h.logger.Info("处理临时错误，将稍后重试",
		zap.String("instance_id", instanceID),
		zap.String("step_id", stepID))

	// 设置延迟重试
	h.stateMgr.SetMetric(instanceID, "scheduled_retry_at", time.Now().Add(time.Minute*10))
	h.stateMgr.SetStatus(instanceID, model.WorkflowStatusWaiting, "等待重试临时错误")

	return nil
}

// handleBusinessError 处理业务错误
func (h *WorkflowErrorHandler) handleBusinessError(ctx context.Context, instanceID, stepID string, err error) error {
	h.logger.Error("业务错误，终止工作流",
		zap.String("instance_id", instanceID),
		zap.String("step_id", stepID),
		zap.Error(err))

	// 业务错误直接终止
	h.stateMgr.SetStatus(instanceID, model.WorkflowStatusFailed, 
		fmt.Sprintf("业务错误: %v", err))

	return nil
}

// handleSystemError 处理系统错误
func (h *WorkflowErrorHandler) handleSystemError(ctx context.Context, instanceID, stepID string, err error) error {
	h.logger.Error("系统错误，尝试恢复",
		zap.String("instance_id", instanceID),
		zap.String("step_id", stepID),
		zap.Error(err))

	// 系统错误尝试快速重试
	if h.shouldRetry(err) {
		return h.executeRetry(ctx, instanceID, stepID, err)
	}

	// 重试失败则标记为系统错误
	h.stateMgr.SetStatus(instanceID, model.WorkflowStatusFailed, 
		fmt.Sprintf("系统错误: %v", err))

	return nil
}

// handleUnknownError 处理未知错误
func (h *WorkflowErrorHandler) handleUnknownError(ctx context.Context, instanceID, stepID string, err error) error {
	h.logger.Error("未知错误，终止工作流",
		zap.String("instance_id", instanceID),
		zap.String("step_id", stepID),
		zap.Error(err))

	h.stateMgr.SetStatus(instanceID, model.WorkflowStatusFailed, 
		fmt.Sprintf("未知错误: %v", err))

	return nil
}

// 错误类型判断函数
func isTemporaryError(err error) bool {
	// 检查是否是网络超时、连接错误等临时性错误
	errMsg := err.Error()
	return containsAny(errMsg, []string{
		"timeout", "connection refused", "temporary", 
		"rate limit", "service unavailable",
	})
}

func isBusinessError(err error) bool {
	// 检查是否是业务逻辑错误
	errMsg := err.Error()
	return containsAny(errMsg, []string{
		"validation failed", "permission denied", 
		"business rule", "data not found",
	})
}

func isSystemError(err error) bool {
	// 检查是否是系统级错误
	errMsg := err.Error()
	return containsAny(errMsg, []string{
		"memory", "disk space", "system", 
		"internal error", "panic",
	})
}

func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(str, substr) {
			return true
		}
	}
	return false
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    len(str) > len(substr) && 
		    (str[:len(substr)] == substr || 
		     str[len(str)-len(substr):] == substr || 
		     indexOf(str, substr) >= 0))
}

func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func getCurrentInstanceID() string {
	// 这里应该从上下文中获取当前实例ID
	// 为了简化，返回空字符串
	return ""
}