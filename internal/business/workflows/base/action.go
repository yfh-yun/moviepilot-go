// Package base 提供动作系统的基础实现
package base

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/types"

	"go.uber.org/zap"
)

// BaseAction 基础动作实现
// 提供所有动作的通用功能实现
type BaseAction struct {
	// 动作标识
	actionID string

	// 状态管理
	done    bool
	success bool
	message string

	// 缓存管理
	cache    interfaces.Cache
	cacheKey string

	// 配置
	config map[string]interface{}

	// 统计信息
	stats *interfaces.ActionStats

	// 观察者
	observers []interfaces.ActionObserver

	// 验证器
	validator interfaces.ActionValidator

	// 日志
	logger *zap.Logger

	// 创建时间
	createdAt time.Time
}

// NewBaseAction 创建基础动作实例
func NewBaseAction(actionID string, cache interfaces.Cache) *BaseAction {
	return &BaseAction{
		actionID:  actionID,
		cache:     cache,
		cacheKey:  fmt.Sprintf("WorkflowCache-%d", 0), // 工作流ID将在执行时设置
		config:    make(map[string]interface{}),
		stats:     &interfaces.ActionStats{},
		observers: make([]interfaces.ActionObserver, 0),
		logger:    logger.Logger,
		createdAt: time.Now(),
	}
}

// Name 获取动作名称
// 子类需要实现此方法
func (ba *BaseAction) Name() string {
	return "BaseAction"
}

// Description 获取动作描述
// 子类需要实现此方法
func (ba *BaseAction) Description() string {
	return "基础动作类"
}

// Data 获取动作数据
// 子类需要实现此方法
func (ba *BaseAction) Data() map[string]interface{} {
	return map[string]interface{}{
		"action_id": ba.actionID,
		"name":      ba.Name(),
		"desc":      ba.Description(),
		"config":    ba.config,
		"stats":     ba.stats,
		"created_at": ba.createdAt,
	}
}

// IsDone 检查动作是否完成
func (ba *BaseAction) IsDone() bool {
	return ba.done
}

// IsSuccess 检查动作是否成功
func (ba *BaseAction) IsSuccess() bool {
	return ba.success
}

// GetMessage 获取执行信息
func (ba *BaseAction) GetMessage() string {
	return ba.message
}

// SetDone 标记动作完成
func (ba *BaseAction) SetDone(message string) {
	ba.done = true
	ba.success = true
	ba.message = message
	ba.stats.SuccessCount++
	ba.stats.LastSuccess = time.Now()
	
	ba.logger.Info("动作完成",
		zap.String("action_id", ba.actionID),
		zap.String("message", message))
}

// SetError 标记动作错误
func (ba *BaseAction) SetError(message string) {
	ba.done = true
	ba.success = false
	ba.message = message
	ba.stats.ErrorCount++
	ba.stats.LastError = time.Now()
	
	ba.logger.Error("动作失败",
		zap.String("action_id", ba.actionID),
		zap.String("message", message))
}

// Execute 执行动作
// 子类需要实现此方法
func (ba *BaseAction) Execute(ctx context.Context, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error) {
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

	// 通知观察者：动作开始
	ba.notifyObservers(func(observer interfaces.ActionObserver) {
		observer.OnActionStart(ctx, ba, workflowID)
	})

	// 验证参数
	if ba.validator != nil {
		if err := ba.validator.ValidateParams(params); err != nil {
			ba.SetError(fmt.Sprintf("参数验证失败: %s", err.Error()))
			ba.notifyObservers(func(observer interfaces.ActionObserver) {
				observer.OnActionError(ctx, ba, workflowID, err)
			})
			return context, err
		}

		if err := ba.validator.ValidateContext(context); err != nil {
			ba.SetError(fmt.Sprintf("上下文验证失败: %s", err.Error()))
			ba.notifyObservers(func(observer interfaces.ActionObserver) {
				observer.OnActionError(ctx, ba, workflowID, err)
			})
			return context, err
		}
	}

	// 执行具体逻辑
	result, err := ba.doExecute(ctx, workflowID, params, context)
	if err != nil {
		ba.SetError(err.Error())
		ba.notifyObservers(func(observer interfaces.ActionObserver) {
			observer.OnActionError(ctx, ba, workflowID, err)
		})
		return result, err
	}

	// 更新统计信息
	duration := time.Since(startTime)
	ba.stats.TotalDuration += duration
	if ba.stats.ExecuteCount > 0 {
		ba.stats.AverageDuration = time.Duration(int64(ba.stats.TotalDuration) / ba.stats.ExecuteCount)
	}

	// 通知观察者：动作完成
	ba.notifyObservers(func(observer interfaces.ActionObserver) {
		observer.OnActionComplete(ctx, ba, workflowID, result, nil)
	})

	ba.logger.Info("动作执行完成",
		zap.String("action_id", ba.actionID),
		zap.Int64("workflow_id", workflowID),
		zap.Bool("success", ba.success),
		zap.Duration("duration", duration),
	)

	return result, nil
}

// doExecute 执行动作的具体逻辑
// 子类需要实现此方法
func (ba *BaseAction) doExecute(ctx context.Context, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error) {
	// 默认实现：直接返回上下文
	ba.SetDone("动作执行完成")
	return context, nil
}

// CheckCache 检查缓存
func (ba *BaseAction) CheckCache(ctx context.Context, workflowID int64, key string) bool {
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
func (ba *BaseAction) SaveCache(ctx context.Context, workflowID int64, data interface{}) error {
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
	return ba.cache.Set(ctx, workflowKey, cacheMap, 24*3600) // 24小时TTL
}

// ClearCache 清理缓存
func (ba *BaseAction) ClearCache(ctx context.Context, workflowID int64) error {
	if ba.cache == nil {
		return nil
	}

	workflowKey := fmt.Sprintf("WorkflowCache-%d", workflowID)
	return ba.cache.Delete(ctx, workflowKey)
}

// Initialize 初始化动作
func (ba *BaseAction) Initialize() error {
	ba.logger.Info("初始化动作", zap.String("action_id", ba.actionID))
	
	// 验证配置
	if ba.validator != nil {
		if err := ba.validator.ValidateConfig(ba.config); err != nil {
			return fmt.Errorf("配置验证失败: %w", err)
		}
	}
	
	return nil
}

// Cleanup 清理动作
func (ba *BaseAction) Cleanup() error {
	ba.logger.Info("清理动作", zap.String("action_id", ba.actionID))
	return nil
}

// GetStats 获取统计信息
func (ba *BaseAction) GetStats() *interfaces.ActionStats {
	return ba.stats
}

// ResetStats 重置统计信息
func (ba *BaseAction) ResetStats() {
	ba.stats = &interfaces.ActionStats{}
}

// GetActionID 获取动作ID
func (ba *BaseAction) GetActionID() string {
	return ba.actionID
}

// SetActionID 设置动作ID
func (ba *BaseAction) SetActionID(actionID string) {
	ba.actionID = actionID
}

// Clone 克隆动作
func (ba *BaseAction) Clone() interfaces.Action {
	// 创建新的实例
	newAction := NewBaseAction(ba.actionID, ba.cache)

	// 复制配置
	newAction.config = make(map[string]interface{})
	for k, v := range ba.config {
		newAction.config[k] = v
	}

	// 复制状态
	newAction.done = ba.done
	newAction.success = ba.success
	newAction.message = ba.message

	// 复制观察者
	newAction.observers = make([]interfaces.ActionObserver, len(ba.observers))
	copy(newAction.observers, ba.observers)

	// 复制验证器
	newAction.validator = ba.validator

	return newAction
}

// ToJSON 转换为JSON
func (ba *BaseAction) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"action_id":   ba.actionID,
		"name":        ba.Name(),
		"description": ba.Description(),
		"done":        ba.done,
		"success":     ba.success,
		"message":     ba.message,
		"config":      ba.config,
		"stats":       ba.stats,
		"created_at":  ba.createdAt,
	}
}

// SetConfig 设置配置
func (ba *BaseAction) SetConfig(config map[string]interface{}) {
	ba.config = config
}

// GetConfig 获取配置
func (ba *BaseAction) GetConfig() map[string]interface{} {
	return ba.config
}

// AddObserver 添加观察者
func (ba *BaseAction) AddObserver(observer interfaces.ActionObserver) {
	ba.observers = append(ba.observers, observer)
}

// RemoveObserver 移除观察者
func (ba *BaseAction) RemoveObserver(observer interfaces.ActionObserver) {
	for i, obs := range ba.observers {
		if obs == observer {
			ba.observers = append(ba.observers[:i], ba.observers[i+1:]...)
			break
		}
	}
}

// SetValidator 设置验证器
func (ba *BaseAction) SetValidator(validator interfaces.ActionValidator) {
	ba.validator = validator
}

// notifyObservers 通知所有观察者
func (ba *BaseAction) notifyObservers(notify func(interfaces.ActionObserver)) {
	for _, observer := range ba.observers {
		notify(observer)
	}
}

// String 字符串表示
func (ba *BaseAction) String() string {
	return fmt.Sprintf("Action[%s]: %s - %s", ba.actionID, ba.Name(), ba.message)
}

// MarshalJSON JSON序列化
func (ba *BaseAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(ba.ToJSON())
}

// UnmarshalJSON JSON反序列化
func (ba *BaseAction) UnmarshalJSON(data []byte) error {
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if actionID, ok := temp["action_id"].(string); ok {
		ba.actionID = actionID
	}
	if done, ok := temp["done"].(bool); ok {
		ba.done = done
	}
	if success, ok := temp["success"].(bool); ok {
		ba.success = success
	}
	if message, ok := temp["message"].(string); ok {
		ba.message = message
	}
	if config, ok := temp["config"].(map[string]interface{}); ok {
		ba.config = config
	}

	return nil
}