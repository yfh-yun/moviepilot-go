// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/business/services/actions/types"

	"go.uber.org/zap"
)

// ActionManager 动作管理器
// 负责管理所有动作的注册、执行和生命周期
type ActionManager struct {
	// 仓储接口
	downloadRepo  interfaces.DownloadRepository
	mediaRepo     interfaces.MediaRepository
	messageRepo   interfaces.MessageRepository
	userRepo      interfaces.UserRepository
	subscribeRepo interfaces.SubscribeRepository

	// 缓存和工具
	cache *WorkflowCache

	// 动作注册表
	actions map[string]BaseAction
	mutex   sync.RWMutex

	// 日志
	logger *zap.Logger

	// 统计信息
	stats *ManagerStats
}

// ManagerStats 管理器统计信息
type ManagerStats struct {
	RegisteredActions int                     `json:"registered_actions"`
	TotalExecutions   int64                   `json:"total_executions"`
	SuccessExecutions int64                   `json:"success_executions"`
	ErrorExecutions   int64                   `json:"error_executions"`
	AverageDuration   time.Duration           `json:"average_duration"`
	LastExecute       time.Time               `json:"last_execute"`
	ActionStats       map[string]*ActionStats `json:"action_stats"`
	mutex             sync.RWMutex
}

// NewActionManager 创建动作管理器实例
func NewActionManager(
	downloadRepo interfaces.DownloadRepository,
	mediaRepo interfaces.MediaRepository,
	messageRepo interfaces.MessageRepository,
	userRepo interfaces.UserRepository,
	subscribeRepo interfaces.SubscribeRepository,
	cache *WorkflowCache,
) *ActionManager {
	am := &ActionManager{
		downloadRepo:  downloadRepo,
		mediaRepo:     mediaRepo,
		messageRepo:   messageRepo,
		userRepo:      userRepo,
		subscribeRepo: subscribeRepo,
		cache:         cache,
		actions:       make(map[string]BaseAction),
		logger:        logger.Logger,
		stats: &ManagerStats{
			ActionStats: make(map[string]*ActionStats),
		},
	}

	// 注册默认动作
	am.registerDefaultActions()

	return am
}

// registerDefaultActions 注册默认动作
func (am *ActionManager) registerDefaultActions() {
	// 注册添加下载动作
	am.RegisterAction(NewAddDownloadAction(am.downloadRepo, am.mediaRepo, am.cache))

	// 注册获取媒体数据动作
	am.RegisterAction(NewFetchMediasAction(am.cache))

	// 注册发送消息动作
	am.RegisterAction(NewSendMessageAction(am.messageRepo, am.userRepo, am.cache))

	// 注册添加订阅动作
	am.RegisterAction(NewAddSubscribeAction(am.subscribeRepo, am.mediaRepo, am.cache))

	am.logger.Info("默认动作注册完成", zap.Int("count", 4))
}

// RegisterAction 注册动作
func (am *ActionManager) RegisterAction(action BaseAction) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	actionID := action.GetActionID()
	if _, exists := am.actions[actionID]; exists {
		return fmt.Errorf("动作 %s 已存在", actionID)
	}

	// 初始化动作
	if err := action.Initialize(); err != nil {
		return fmt.Errorf("初始化动作 %s 失败: %w", actionID, err)
	}

	am.actions[actionID] = action
	am.stats.RegisteredActions++
	am.stats.ActionStats[actionID] = &ActionStats{}

	am.logger.Info("动作注册成功",
		zap.String("action_id", actionID),
		zap.String("name", action.Name()),
	)

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
		am.logger.Warn("清理动作失败", zap.String("action_id", actionID), zap.Error(err))
	}

	delete(am.actions, actionID)
	am.stats.RegisteredActions--
	delete(am.stats.ActionStats, actionID)

	am.logger.Info("动作注销成功", zap.String("action_id", actionID))

	return nil
}

// GetAction 获取动作
func (am *ActionManager) GetAction(actionID string) (BaseAction, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	action, exists := am.actions[actionID]
	if !exists {
		return nil, fmt.Errorf("动作 %s 不存在", actionID)
	}

	return action, nil
}

// ListActions 列出所有动作
func (am *ActionManager) ListActions() map[string]BaseAction {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// 创建副本
	result := make(map[string]BaseAction)
	for id, action := range am.actions {
		result[id] = action
	}

	return result
}

// ExecuteAction 执行动作
func (am *ActionManager) ExecuteAction(
	ctx context.Context,
	workflowID int64,
	actionID string,
	params map[string]interface{},
	actionContext *types.ActionContext,
) (*types.ActionContext, error) {
	startTime := time.Now()

	// 获取动作
	action, err := am.GetAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("获取动作失败: %w", err)
	}

	am.logger.Info("开始执行动作",
		zap.Int64("workflow_id", workflowID),
		zap.String("action_id", actionID),
		zap.String("action_name", action.Name()),
	)

	// 更新统计信息
	am.updateStats(true, false, time.Since(startTime))

	// 执行动作
	result, err := action.Execute(ctx, workflowID, params, actionContext)
	if err != nil {
		am.updateStats(false, true, time.Since(startTime))
		am.logger.Error("动作执行失败",
			zap.Int64("workflow_id", workflowID),
			zap.String("action_id", actionID),
			zap.Error(err),
		)
		return result, err
	}

	// 更新统计信息
	am.updateStats(false, false, time.Since(startTime))

	am.logger.Info("动作执行完成",
		zap.Int64("workflow_id", workflowID),
		zap.String("action_id", actionID),
		zap.Bool("success", action.IsSuccess()),
		zap.String("message", action.GetMessage()),
		zap.Duration("duration", time.Since(startTime)),
	)

	return result, nil
}

// ExecuteActions 执行多个动作
func (am *ActionManager) ExecuteActions(
	ctx context.Context,
	workflowID int64,
	actionConfigs []ActionConfig,
	initialContext *types.ActionContext,
) (*types.ActionContext, error) {
	currentContext := initialContext

	for _, config := range actionConfigs {
		// 检查是否启用
		if !config.Enabled {
			am.logger.Debug("跳过已禁用的动作", zap.String("action_id", config.ActionID))
			continue
		}

		// 检查工作流是否已停止
		if am.isWorkflowStopped(ctx, workflowID) {
			am.logger.Info("工作流已停止，终止动作执行", zap.Int64("workflow_id", workflowID))
			break
		}

		// 执行动作
		var err error
		currentContext, err = am.ExecuteAction(ctx, workflowID, config.ActionID, config.Params, currentContext)
		if err != nil {
			if config.StopOnError {
				return currentContext, fmt.Errorf("动作 %s 执行失败，停止执行: %w", config.ActionID, err)
			}
			am.logger.Warn("动作执行失败，继续执行下一个动作",
				zap.String("action_id", config.ActionID),
				zap.Error(err),
			)
		}

		// 检查动作是否成功
		action, _ := am.GetAction(config.ActionID)
		if action != nil && !action.IsSuccess() {
			if config.StopOnError {
				return currentContext, fmt.Errorf("动作 %s 执行失败，停止执行: %s", config.ActionID, action.GetMessage())
			}
		}
	}

	return currentContext, nil
}

// ActionConfig 动作配置
type ActionConfig struct {
	ActionID    string                 `json:"action_id"`
	Name        string                 `json:"name"`
	Params      map[string]interface{} `json:"params"`
	Order       int                    `json:"order"`
	Enabled     bool                   `json:"enabled"`
	StopOnError bool                   `json:"stop_on_error"`
	Timeout     int                    `json:"timeout"`
	Retry       int                    `json:"retry"`
}

// GetStats 获取统计信息
func (am *ActionManager) GetStats() *ManagerStats {
	am.stats.mutex.RLock()
	defer am.stats.mutex.RUnlock()

	// 创建副本
	stats := &ManagerStats{
		RegisteredActions: am.stats.RegisteredActions,
		TotalExecutions:   am.stats.TotalExecutions,
		SuccessExecutions: am.stats.SuccessExecutions,
		ErrorExecutions:   am.stats.ErrorExecutions,
		AverageDuration:   am.stats.AverageDuration,
		LastExecute:       am.stats.LastExecute,
		ActionStats:       make(map[string]*ActionStats),
	}

	// 复制动作统计
	for id, actionStats := range am.stats.ActionStats {
		stats.ActionStats[id] = &ActionStats{
			ExecuteCount:    actionStats.ExecuteCount,
			SuccessCount:    actionStats.SuccessCount,
			ErrorCount:      actionStats.ErrorCount,
			TotalDuration:   actionStats.TotalDuration,
			AverageDuration: actionStats.AverageDuration,
			LastExecute:     actionStats.LastExecute,
			LastSuccess:     actionStats.LastSuccess,
			LastError:       actionStats.LastError,
		}
	}

	return stats
}

// ResetStats 重置统计信息
func (am *ActionManager) ResetStats() {
	am.stats.mutex.Lock()
	defer am.stats.mutex.Unlock()

	am.stats.TotalExecutions = 0
	am.stats.SuccessExecutions = 0
	am.stats.ErrorExecutions = 0
	am.stats.AverageDuration = 0
	am.stats.LastExecute = time.Time{}
	am.stats.ActionStats = make(map[string]*ActionStats)

	// 重置所有动作的统计信息
	am.mutex.RLock()
	for _, action := range am.actions {
		if baseAction, ok := action.(*BaseActionImpl); ok {
			baseAction.ResetStats()
		}
	}
	am.mutex.RUnlock()

	am.logger.Info("统计信息已重置")
}

// updateStats 更新统计信息
func (am *ActionManager) updateStats(isStart, isError bool, duration time.Duration) {
	am.stats.mutex.Lock()
	defer am.stats.mutex.Unlock()

	if isStart {
		am.stats.TotalExecutions++
		am.stats.LastExecute = time.Now()
	} else {
		if isError {
			am.stats.ErrorExecutions++
		} else {
			am.stats.SuccessExecutions++
		}

		// 更新平均执行时间
		if am.stats.TotalExecutions > 0 {
			am.stats.AverageDuration = time.Duration(int64(am.stats.AverageDuration)*int64(am.stats.TotalExecutions-1)+int64(duration)) / int64(am.stats.TotalExecutions)
		}
	}
}

// isWorkflowStopped 检查工作流是否已停止
func (am *ActionManager) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// ValidateActionConfig 验证动作配置
func (am *ActionManager) ValidateActionConfig(config ActionConfig) error {
	// 检查动作是否存在
	_, err := am.GetAction(config.ActionID)
	if err != nil {
		return fmt.Errorf("动作 %s 不存在: %w", config.ActionID, err)
	}

	// 检查参数
	action, _ := am.GetAction(config.ActionID)
	if baseAction, ok := action.(*BaseActionImpl); ok {
		if err := baseAction.ValidateParams(config.Params); err != nil {
			return fmt.Errorf("参数验证失败: %w", err)
		}
	}

	// 检查配置
	if config.Order < 0 {
		return fmt.Errorf("执行顺序不能为负数")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("超时时间不能为负数")
	}

	if config.Retry < 0 {
		return fmt.Errorf("重试次数不能为负数")
	}

	return nil
}

// CloneAction 克隆动作
func (am *ActionManager) CloneAction(actionID string, newActionID string) error {
	action, err := am.GetAction(actionID)
	if err != nil {
		return fmt.Errorf("获取源动作失败: %w", err)
	}

	// 克隆动作
	clonedAction := action.Clone()
	clonedAction.SetActionID(newActionID)

	// 注册克隆的动作
	return am.RegisterAction(clonedAction)
}

// ExportActions 导出动作配置
func (am *ActionManager) ExportActions() map[string]interface{} {
	actions := am.ListActions()
	result := make(map[string]interface{})

	for id, action := range actions {
		result[id] = map[string]interface{}{
			"name":        action.Name(),
			"description": action.Description(),
			"data":        action.Data(),
			"stats":       action.(*BaseActionImpl).GetStats(),
		}
	}

	return result
}

// ImportActions 导入动作配置
func (am *ActionManager) ImportActions(configs map[string]interface{}) error {
	for id, config := range configs {
		configMap, ok := config.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := configMap["name"].(string)
		description, _ := configMap["description"].(string)

		am.logger.Info("导入动作配置",
			zap.String("action_id", id),
			zap.String("name", name),
			zap.String("description", description),
		)

		// 这里可以根据配置创建动作
		// 暂时只记录日志
	}

	return nil
}

// Cleanup 清理管理器
func (am *ActionManager) Cleanup() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// 清理所有动作
	for id, action := range am.actions {
		if err := action.Cleanup(); err != nil {
			am.logger.Warn("清理动作失败", zap.String("action_id", id), zap.Error(err))
		}
	}

	// 清空动作注册表
	am.actions = make(map[string]BaseAction)
	am.stats.RegisteredActions = 0

	am.logger.Info("动作管理器清理完成")

	return nil
}
