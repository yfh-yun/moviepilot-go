// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/services/actions/types"

	"go.uber.org/zap"
)

// BaseAction 基础动作接口
// 对应Python版本的BaseAction抽象基类
type BaseAction interface {
	// 基本属性
	Name() string
	Description() string
	Data() map[string]interface{}

	// 状态管理
	IsDone() bool
	IsSuccess() bool
	GetMessage() string
	SetDone(message string)
	SetError(message string)

	// 执行方法
	Execute(ctx context.Context, workflowID int64, params map[string]interface{}, context *types.ActionContext) (*types.ActionContext, error)

	// 缓存管理
	CheckCache(ctx context.Context, workflowID int64, key string) bool
	SaveCache(ctx context.Context, workflowID int64, data interface{}) error
	ClearCache(ctx context.Context, workflowID int64) error

	// 生命周期
	Initialize() error
	Cleanup() error
}

// BaseActionImpl 基础动作实现
// 提供所有动作的通用功能实现
type BaseActionImpl struct {
	// 动作标识
	actionID string

	// 状态管理
	done    bool
	success bool
	message string

	// 缓存管理
	cache    *WorkflowCache
	cacheKey string

	// 日志
	logger *zap.Logger

	// 统计信息
	stats *ActionStats
}

// ActionStats 动作统计信息
type ActionStats struct {
	ExecuteCount    int64         `json:"execute_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	LastExecute     time.Time     `json:"last_execute"`
	LastSuccess     time.Time     `json:"last_success"`
	LastError       time.Time     `json:"last_error"`
}

// NewBaseAction 创建基础动作实例
func NewBaseAction(actionID string, cache *WorkflowCache) *BaseActionImpl {
	return &BaseActionImpl{
		actionID: actionID,
		cache:    cache,
		cacheKey: fmt.Sprintf("WorkflowCache-%d", 0), // 工作流ID将在执行时设置
		logger:   logger.Logger,
		stats:    &ActionStats{},
	}
}

// Name 获取动作名称
// 子类需要实现此方法
func (ba *BaseActionImpl) Name() string {
	return "BaseAction"
}

// Description 获取动作描述
// 子类需要实现此方法
func (ba *BaseActionImpl) Description() string {
	return "基础动作类"
}

// Data 获取动作数据
// 子类需要实现此方法
func (ba *BaseActionImpl) Data() map[string]interface{} {
	return map[string]interface{}{
		"action_id": ba.actionID,
		"name":      ba.Name(),
		"desc":      ba.Description(),
	}
}

// IsDone 检查动作是否完成
func (ba *BaseActionImpl) IsDone() bool {
	return ba.done
}

// IsSuccess 检查动作是否成功
func (ba *BaseActionImpl) IsSuccess() bool {
	return ba.success
}

// GetMessage 获取执行信息
func (ba *BaseActionImpl) GetMessage() string {
	return ba.message
}

// SetDone 标记动作完成
func (ba *BaseActionImpl) SetDone(message string) {
	ba.done = true
	ba.success = true
	ba.message = message
	ba.stats.SuccessCount++
	ba.stats.LastSuccess = time.Now()
}

// SetError 标记动作错误
func (ba *BaseActionImpl) SetError(message string) {
	ba.done = true
	ba.success = false
	ba.message = message
	ba.stats.ErrorCount++
	ba.stats.LastError = time.Now()
}

// CheckCache 检查缓存
func (ba *BaseActionImpl) CheckCache(ctx context.Context, workflowID int64, key string) bool {
	if ba.cache == nil {
		return false
	}

	workflowKey := fmt.Sprintf("WorkflowCache-%d", workflowID)
	workflowCache, err := ba.cache.Get(ctx, workflowKey)
	if err != nil {
		ba.logger.Warn("获取工作流缓存失败", zap.Error(err))
		return false
	}

	if workflowCache == nil {
		return false
	}

	// 将缓存转换为map类型
	cacheMap, ok := workflowCache.(map[string]interface{})
	if !ok {
		return false
	}

	// 获取动作缓存
	actionCache, exists := cacheMap[ba.actionID]
	if !exists {
		return false
	}

	// 检查key是否在动作缓存中
	actionCacheList, ok := actionCache.([]interface{})
	if !ok {
		return false
	}

	for _, item := range actionCacheList {
		if item == key {
			return true
		}
	}

	return false
}

// SaveCache 保存缓存
func (ba *BaseActionImpl) SaveCache(ctx context.Context, workflowID int64, data interface{}) error {
	if ba.cache == nil {
		return nil
	}

	workflowKey := fmt.Sprintf("WorkflowCache-%d", workflowID)

	// 获取现有缓存
	workflowCache, err := ba.cache.Get(ctx, workflowKey)
	var cacheMap map[string]interface{}

	if err != nil || workflowCache == nil {
		cacheMap = make(map[string]interface{})
	} else {
		var ok bool
		cacheMap, ok = workflowCache.(map[string]interface{})
		if !ok {
			cacheMap = make(map[string]interface{})
		}
	}

	// 获取动作缓存
	actionCache, exists := cacheMap[ba.actionID]
	var actionCacheList []interface{}

	if !exists {
		actionCacheList = make([]interface{}, 0)
	} else {
		var ok bool
		actionCacheList, ok = actionCache.([]interface{})
		if !ok {
			actionCacheList = make([]interface{}, 0)
		}
	}

	// 添加新数据
	switch v := data.(type) {
	case []interface{}:
		actionCacheList = append(actionCacheList, v...)
	case []string:
		for _, item := range v {
			actionCacheList = append(actionCacheList, item)
		}
	case string:
		actionCacheList = append(actionCacheList, v)
	default:
		actionCacheList = append(actionCacheList, v)
	}

	// 更新缓存
	cacheMap[ba.actionID] = actionCacheList
	return ba.cache.Set(ctx, workflowKey, cacheMap, 24*time.Hour)
}

// ClearCache 清理缓存
func (ba *BaseActionImpl) ClearCache(ctx context.Context, workflowID int64) error {
	if ba.cache == nil {
		return nil
	}

	workflowKey := fmt.Sprintf("WorkflowCache-%d", workflowID)
	return ba.cache.Delete(ctx, workflowKey)
}

// Initialize 初始化动作
func (ba *BaseActionImpl) Initialize() error {
	ba.logger.Info("初始化动作", zap.String("action_id", ba.actionID))
	return nil
}

// Cleanup 清理动作
func (ba *BaseActionImpl) Cleanup() error {
	ba.logger.Info("清理动作", zap.String("action_id", ba.actionID))
	return nil
}

// Execute 执行动作
// 子类需要实现此方法
func (ba *BaseActionImpl) Execute(ctx context.Context, workflowID int64, params map[string]interface{}, context *types.ActionContext) (*types.ActionContext, error) {
	startTime := time.Now()

	// 更新缓存键
	ba.cacheKey = fmt.Sprintf("WorkflowCache-%d", workflowID)

	// 更新统计信息
	ba.stats.ExecuteCount++
	ba.stats.LastExecute = startTime

	// 重置状态
	ba.done = false
	ba.success = false
	ba.message = ""

	// 子类实现具体的执行逻辑
	result, err := ba.doExecute(ctx, workflowID, params, context)

	// 更新统计信息
	duration := time.Since(startTime)
	ba.stats.TotalDuration += duration
	ba.stats.AverageDuration = time.Duration(int64(ba.stats.TotalDuration) / ba.stats.ExecuteCount)

	ba.logger.Info("动作执行完成",
		zap.String("action_id", ba.actionID),
		zap.Int64("workflow_id", workflowID),
		zap.Bool("success", ba.success),
		zap.Duration("duration", duration),
	)

	return result, err
}

// doExecute 执行动作的具体逻辑
// 子类需要实现此方法
func (ba *BaseActionImpl) doExecute(ctx context.Context, workflowID int64, params map[string]interface{}, context *types.ActionContext) (*types.ActionContext, error) {
	// 默认实现：直接返回上下文
	ba.SetDone("动作执行完成")
	return context, nil
}

// GetStats 获取统计信息
func (ba *BaseActionImpl) GetStats() *ActionStats {
	return ba.stats
}

// ResetStats 重置统计信息
func (ba *BaseActionImpl) ResetStats() {
	ba.stats = &ActionStats{}
}

// ValidateParams 验证参数
func (ba *BaseActionImpl) ValidateParams(params map[string]interface{}) error {
	// 默认实现：不验证
	return nil
}

// GetActionID 获取动作ID
func (ba *BaseActionImpl) GetActionID() string {
	return ba.actionID
}

// SetActionID 设置动作ID
func (ba *BaseActionImpl) SetActionID(actionID string) {
	ba.actionID = actionID
}

// Clone 克隆动作
func (ba *BaseActionImpl) Clone() BaseAction {
	// 创建新的实例
	newAction := NewBaseAction(ba.actionID, ba.cache)

	// 复制状态
	newAction.done = ba.done
	newAction.success = ba.success
	newAction.message = ba.message

	return newAction
}

// ToJSON 转换为JSON
func (ba *BaseActionImpl) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"action_id":   ba.actionID,
		"name":        ba.Name(),
		"description": ba.Description(),
		"done":        ba.done,
		"success":     ba.success,
		"message":     ba.message,
		"stats":       ba.stats,
	}
}

// String 字符串表示
func (ba *BaseActionImpl) String() string {
	return fmt.Sprintf("Action[%s]: %s - %s", ba.actionID, ba.Name(), ba.message)
}
