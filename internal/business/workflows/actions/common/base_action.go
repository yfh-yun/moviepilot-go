package common

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/systemconfig"
	"moviepilot-go/internal/business/workflows/actions/base"

	"github.com/google/uuid"
)

// BaseAction 实现动作的基础类，使用模板方法模式
type BaseAction struct {
	name                string                // 动作名称
	actionType          string                // 动作类型
	status              string                // 动作状态
	actionID            string                // 动作ID
	initialized         bool                  // 是否已初始化
	logger              *zap.Logger           // 日志记录器
	input               map[string]any        // 输入参数
	globalContext       map[string]any        // 全局上下文
	services            map[string]any        // 服务实例
	systemConfigService *systemconfig.Service // 系统配置服务，用于缓存管理
	doneFlag            bool                  // 完成标志
	message             string                // 执行信息
	cacheKeyFormat      string                // 缓存键格式
}

// NewBaseAction 创建新的基础动作实例
func NewBaseAction(name, actionType string) *BaseAction {
	return &BaseAction{
		name:           name,
		actionType:     actionType,
		status:         base.ActionStatusPending,
		actionID:       uuid.New().String(),
		doneFlag:       false,
		message:        "",
		cacheKeyFormat: "WorkflowCache-%s",
	}
}

// GetName 获取动作名称
func (a *BaseAction) GetName() string {
	return a.name
}

// GetType 获取动作类型
func (a *BaseAction) GetType() string {
	return a.actionType
}

// Initialize 初始化动作
func (a *BaseAction) Initialize(ctx base.ActionContext) error {
	if a.initialized {
		return nil
	}

	a.logger = ctx.Logger
	a.input = ctx.Input
	a.globalContext = ctx.GlobalContext
	a.services = ctx.Services
	a.initialized = true

	// 获取系统配置服务实例
	if service, ok := ctx.Services["SystemConfigService"].(*systemconfig.Service); ok {
		a.systemConfigService = service
	}

	a.logger.Debug("Action initialized", zap.String("action_name", a.name), zap.String("action_id", a.actionID))
	return nil
}

// IsInitialized 检查动作是否已初始化
func (a *BaseAction) IsInitialized() bool {
	return a.initialized
}

// Execute 执行动作，实现模板方法模式
func (a *BaseAction) Execute(ctx base.ActionContext) (*base.ActionResult, error) {
	// 记录开始时间
	startTime := time.Now()
	result := &base.ActionResult{
		Success:  false,
		Output:   make(map[string]any),
		Status:   base.ActionStatusRunning,
		Duration: 0,
	}

	// 设置状态为运行中
	a.status = base.ActionStatusRunning
	a.logger.Debug("Action started", zap.String("action_name", a.name), zap.String("action_id", a.actionID))

	// 1. 执行前处理
	if err := a.preExecute(ctx); err != nil {
		a.status = base.ActionStatusFailed
		result.Status = base.ActionStatusFailed
		result.ErrorMessage = err.Error()
		result.Duration = time.Since(startTime)
		a.logger.Error("Action preExecute failed", zap.String("action_name", a.name), zap.String("action_id", a.actionID), zap.Error(err))
		return result, err
	}

	// 2. 执行核心逻辑
	output, err := a.execute(ctx)
	if err != nil {
		a.status = base.ActionStatusFailed
		result.Status = base.ActionStatusFailed
		result.ErrorMessage = err.Error()
		result.Duration = time.Since(startTime)
		a.logger.Error("Action execute failed", zap.String("action_name", a.name), zap.String("action_id", a.actionID), zap.Error(err))

		// 执行失败处理
		if rollbackErr := a.onFailure(ctx, err); rollbackErr != nil {
			a.logger.Error("Action rollback failed", zap.String("action_name", a.name), zap.String("action_id", a.actionID), zap.Error(rollbackErr))
		}

		return result, err
	}

	// 3. 执行后处理
	if err := a.postExecute(ctx); err != nil {
		a.status = base.ActionStatusFailed
		result.Status = base.ActionStatusFailed
		result.ErrorMessage = err.Error()
		result.Duration = time.Since(startTime)
		a.logger.Error("Action postExecute failed", zap.String("action_name", a.name), zap.String("action_id", a.actionID), zap.Error(err))
		return result, err
	}

	// 4. 执行成功处理
	if err := a.onSuccess(ctx, output); err != nil {
		a.logger.Error("Action onSuccess failed", zap.String("action_name", a.name), zap.String("action_id", a.actionID), zap.Error(err))
	}

	// 设置最终状态
	a.status = base.ActionStatusCompleted
	result.Success = true
	result.Status = base.ActionStatusCompleted
	result.Output = output
	result.Duration = time.Since(startTime)

	a.logger.Info("Action completed successfully", zap.String("action_name", a.name), zap.String("action_id", a.actionID), zap.Duration("duration", result.Duration))
	return result, nil
}

// GetStatus 获取动作状态
func (a *BaseAction) GetStatus() string {
	return a.status
}

// Cancel 取消动作执行
func (a *BaseAction) Cancel() error {
	if a.status != base.ActionStatusRunning {
		return nil
	}

	a.status = base.ActionStatusCancelled
	a.logger.Info("Action cancelled", zap.String("action_name", a.name), zap.String("action_id", a.actionID))
	return nil
}

// GetActionID 获取动作ID
func (a *BaseAction) GetActionID() string {
	return a.actionID
}

// GetLogger 获取日志记录器
func (a *BaseAction) GetLogger() *zap.Logger {
	return a.logger
}

// GetDescription 获取动作描述
func (a *BaseAction) GetDescription() string {
	// 默认实现返回空描述，子类可以覆盖
	return ""
}

// GetData 获取动作参数模板
func (a *BaseAction) GetData() map[string]any {
	// 默认实现返回空模板，子类可以覆盖
	return make(map[string]any)
}

// Done 判断动作是否完成
func (a *BaseAction) Done() bool {
	return a.doneFlag
}

// Message 获取执行信息
func (a *BaseAction) Message() string {
	return a.message
}

// JobDone 标记动作完成
func (a *BaseAction) JobDone(message string) {
	a.message = message
	a.doneFlag = true
	a.status = base.ActionStatusCompleted
}

// Success 判断动作是否成功
func (a *BaseAction) Success() bool {
	return a.status == base.ActionStatusCompleted
}

// preExecute 执行前处理，子类可覆盖
func (a *BaseAction) preExecute(_ base.ActionContext) error {
	// 默认实现为空
	return nil
}

// execute 执行核心逻辑，子类必须覆盖
func (a *BaseAction) execute(_ base.ActionContext) (map[string]any, error) {
	// 默认实现返回空
	return make(map[string]any), nil
}

// postExecute 执行后处理，子类可覆盖
func (a *BaseAction) postExecute(_ base.ActionContext) error {
	// 默认实现为空
	return nil
}

// onSuccess 执行成功处理，子类可覆盖
func (a *BaseAction) onSuccess(_ base.ActionContext, _ map[string]any) error {
	// 默认实现为空
	return nil
}

// onFailure 执行失败处理，子类可覆盖
func (a *BaseAction) onFailure(_ base.ActionContext, _ error) error {
	// 默认实现为空
	return nil
}

// CheckCache 检查指定key是否在当前动作的缓存中
func (a *BaseAction) CheckCache(workflowID string, key string) bool {
	if a.systemConfigService == nil {
		return false
	}

	// 获取工作流缓存
	cacheKey := fmt.Sprintf(a.cacheKeyFormat, workflowID)
	cacheDataStr, exists := a.systemConfigService.Get(cacheKey)
	if !exists {
		return false
	}

	// 解析缓存数据
	var workflowCache map[string][]string
	if err := json.Unmarshal([]byte(cacheDataStr), &workflowCache); err != nil {
		a.logger.Error("Failed to unmarshal workflow cache", zap.Error(err), zap.String("cache_key", cacheKey))
		return false
	}

	// 检查当前动作的缓存中是否包含指定key
	actionCache := workflowCache[a.actionID]
	for _, cachedKey := range actionCache {
		if cachedKey == key {
			return true
		}
	}

	return false
}

// SaveCache 将数据保存到当前动作的缓存中
func (a *BaseAction) SaveCache(workflowID string, data interface{}) error {
	if a.systemConfigService == nil {
		return nil
	}

	// 获取工作流缓存
	cacheKey := fmt.Sprintf(a.cacheKeyFormat, workflowID)
	cacheDataStr, exists := a.systemConfigService.Get(cacheKey)

	var workflowCache map[string][]string
	if exists {
		if err := json.Unmarshal([]byte(cacheDataStr), &workflowCache); err != nil {
			a.logger.Error("Failed to unmarshal workflow cache", zap.Error(err), zap.String("cache_key", cacheKey))
			workflowCache = make(map[string][]string)
		}
	} else {
		workflowCache = make(map[string][]string)
	}

	// 获取当前动作的缓存
	actionCache := workflowCache[a.actionID]

	// 根据数据类型添加到缓存
	switch v := data.(type) {
	case []string:
		actionCache = append(actionCache, v...)
	case string:
		actionCache = append(actionCache, v)
	default:
		a.logger.Error("Invalid cache data type", zap.Any("data", data), zap.String("action_id", a.actionID))
		return nil
	}

	// 更新工作流缓存
	workflowCache[a.actionID] = actionCache

	// 序列化并保存
	updatedCacheData, err := json.Marshal(workflowCache)
	if err != nil {
		a.logger.Error("Failed to marshal workflow cache", zap.Error(err), zap.String("cache_key", cacheKey))
		return err
	}

	// 保存到系统配置服务
	_, err = a.systemConfigService.Set(nil, cacheKey, string(updatedCacheData))
	if err != nil {
		a.logger.Error("Failed to save workflow cache", zap.Error(err), zap.String("cache_key", cacheKey))
		return err
	}

	return nil
}
