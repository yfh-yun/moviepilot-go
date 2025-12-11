package actions

import (
	"sync"

	"go.uber.org/zap"
)

// ActionManager 定义动作管理器接口
type ActionManager interface {
	// RegisterAction 注册动作
	RegisterAction(action Action) error

	// GetAction 获取动作
	GetAction(actionID string) (Action, error)

	// ExecuteAction 执行动作
	ExecuteAction(ctx ActionContext) (*ActionResult, error)

	// CancelAction 取消动作执行
	CancelAction(actionID string) error

	// GetActionStatus 获取动作状态
	GetActionStatus(actionID string) (string, error)

	// GetActions 获取所有动作
	GetActions() map[string]Action
}

// DefaultActionManager 实现默认的动作管理器
type DefaultActionManager struct {
	// actions 已注册的动作列表
	actions map[string]Action

	// mutex 互斥锁，用于保护actions的并发访问
	mutex sync.RWMutex

	// logger 日志记录器
	logger *zap.Logger
}

// NewDefaultActionManager 创建新的动作管理器实例
func NewDefaultActionManager(logger *zap.Logger) *DefaultActionManager {
	return &DefaultActionManager{
		actions: make(map[string]Action),
		mutex:   sync.RWMutex{},
		logger:  logger,
	}
}

// RegisterAction 注册动作
func (m *DefaultActionManager) RegisterAction(action Action) error {
	// 检查动作是否已注册
	m.mutex.Lock()
	defer m.mutex.Unlock()

	actionID := action.GetActionID()
	if _, exists := m.actions[actionID]; exists {
		return nil // 动作已注册，无需重复注册
	}

	// 注册动作
	m.actions[actionID] = action
	m.logger.Info("Action registered", zap.String("action_id", actionID), zap.String("action_name", action.GetName()), zap.String("action_type", action.GetType()))

	return nil
}

// GetAction 获取动作
func (m *DefaultActionManager) GetAction(actionID string) (Action, error) {
	// 获取动作
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	action, exists := m.actions[actionID]
	if !exists {
		return nil, nil // 动作不存在
	}

	return action, nil
}

// ExecuteAction 执行动作
func (m *DefaultActionManager) ExecuteAction(ctx ActionContext) (*ActionResult, error) {
	// 获取动作
	m.mutex.RLock()
	action, exists := m.actions[ctx.ActionID]
	m.mutex.RUnlock()

	if !exists {
		// 如果动作不存在，创建新动作
		// 注意：这里使用了空的Services实例，实际应用中应传入完整的服务实例
		services := NewServices()
		factory := NewActionFactory(m.logger, services)
		var err error
		action, err = factory.Create(ctx.ActionName)
		if err != nil {
			m.logger.Error("Failed to create action", zap.String("action_name", ctx.ActionName), zap.Error(err))
			return nil, err
		}

		// 注册动作
		m.RegisterAction(action)
	}

	// 初始化动作
	if !action.IsInitialized() {
		if err := action.Initialize(ctx); err != nil {
			m.logger.Error("Failed to initialize action", zap.String("action_id", action.GetActionID()), zap.Error(err))
			return nil, err
		}
	}

	// 执行动作
	result, err := action.Execute(ctx)
	if err != nil {
		m.logger.Error("Failed to execute action", zap.String("action_id", action.GetActionID()), zap.Error(err))
		return result, err
	}

	m.logger.Info("Action executed successfully", zap.String("action_id", action.GetActionID()), zap.String("action_name", action.GetName()), zap.String("status", result.Status), zap.Duration("duration", result.Duration))
	return result, nil
}

// CancelAction 取消动作执行
func (m *DefaultActionManager) CancelAction(actionID string) error {
	// 获取动作
	m.mutex.RLock()
	action, exists := m.actions[actionID]
	m.mutex.RUnlock()

	if !exists {
		return nil // 动作不存在，无需取消
	}

	// 取消动作执行
	if err := action.Cancel(); err != nil {
		m.logger.Error("Failed to cancel action", zap.String("action_id", actionID), zap.Error(err))
		return err
	}

	m.logger.Info("Action cancelled", zap.String("action_id", actionID))
	return nil
}

// GetActionStatus 获取动作状态
func (m *DefaultActionManager) GetActionStatus(actionID string) (string, error) {
	// 获取动作
	m.mutex.RLock()
	action, exists := m.actions[actionID]
	m.mutex.RUnlock()

	if !exists {
		return ActionStatusPending, nil // 动作不存在，返回待执行状态
	}

	// 获取动作状态
	return action.GetStatus(), nil
}

// GetActions 获取所有动作
func (m *DefaultActionManager) GetActions() map[string]Action {
	// 获取所有动作
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 返回动作列表的副本
	actions := make(map[string]Action)
	for k, v := range m.actions {
		actions[k] = v
	}

	return actions
}
