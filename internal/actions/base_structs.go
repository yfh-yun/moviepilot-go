package actions
import (
	"fmt"

	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/repository"
	"go.uber.org/zap"
)

// ActionContext 动作上下文
type ActionContext struct {
	Fileitems       []*model.FileItem
	Downloads      []*model.DownloadHistory
	Medias         []*model.MediaInfo
	UserConfig      *model.UserConfig
	WorkflowID      int
	ShouldStop      bool
}

// ActionParams 动作参数接口
type ActionParams interface {
	Validate() error
}

// BaseAction 动作基础接口
type BaseAction interface {
	Name() string
	Description() string
	Data() interface{}
	Done() bool
	Success() bool
	Message() string
	SetDone(message string)
	CheckCache(workflowID int, key string) bool
	SaveCache(workflowID int, data interface{})
	Execute(workflowID int, params interface{}, context *ActionContext) (*ActionContext, error)
}

// BaseActionImpl 基础动作实现
type BaseActionImpl struct {
	actionID       string
	doneFlag       bool
	message        string
	systemConfigRepo repository.SystemConfigRepository
}

// NewBaseActionImpl 创建基础动作实现
func NewBaseActionImpl(actionID string, systemConfigRepo repository.SystemConfigRepository) *BaseActionImpl {
	return &BaseActionImpl{
		actionID:       actionID,
		doneFlag:       false,
		message:        "",
		systemConfigRepo: systemConfigRepo,
	}
}

// Name 获取动作名称
func (a *BaseActionImpl) Name() string {
	return a.actionID
}

// Description 获取动作描述
func (a *BaseActionImpl) Description() string {
	return a.actionID + "动作"
}

// Data 获取动作数据
func (a *BaseActionImpl) Data() interface{} {
	return map[string]interface{}{
		"action_id": a.actionID,
	}
}

// Done 判断动作是否完成
func (a *BaseActionImpl) Done() bool {
	return a.doneFlag
}

// Success 判断动作是否成功
func (a *BaseActionImpl) Success() bool {
	return !a.doneFlag || a.message == ""
}

// Message 获取执行信息
func (a *BaseActionImpl) Message() string {
	return a.message
}

// SetDone 标记动作完成
func (a *BaseActionImpl) SetDone(message string) {
	zap.Debug("Marking action as done", 
		zap.String("action_id", a.actionID),
		zap.String("message", message))
	
	a.message = message
	a.doneFlag = true
}

// checkCache 检查是否处理过
func (a *BaseActionImpl) CheckCache(workflowID int, key string) bool {
	zap.Debug("Checking action cache", 
		zap.Int("workflow_id", workflowID),
		zap.String("key", key))
	
	workflowKey := fmt.Sprintf("WorkflowCache-%d", workflowID)
	workflowCache, err := a.getWorkflowCache(workflowKey)
	if err != nil {
		zap.Error("Failed to get workflow cache", zap.Error(err))
		return false
	}
	
	actionCache, ok := workflowCache[a.actionID].([]interface{})
	if !ok {
		return false
	}
	
	// 检查key是否在缓存中
	for _, item := range actionCache {
		if itemStr, ok := item.(string); ok && itemStr == key {
			return true
		}
	}
	
	return false
}

// saveCache 保存缓存
func (a *BaseActionImpl) SaveCache(workflowID int, data interface{}) {
	zap.Debug("Saving action cache", 
		zap.Int("workflow_id", workflowID),
		zap.Any("data", data))
	
	workflowKey := fmt.Sprintf("WorkflowCache-%d", workflowID)
	workflowCache, err := a.getWorkflowCache(workflowKey)
	if err != nil {
		zap.Error("Failed to get workflow cache", zap.Error(err))
		return
	}
	
	actionCache, ok := workflowCache[a.actionID].([]interface{})
	if !ok {
		actionCache = []interface{}{}
	}
	
	// 添加数据到缓存
	switch v := data.(type) {
	case string:
		actionCache = append(actionCache, v)
	case []string:
		actionCache = append(actionCache, v...)
	default:
		actionCache = append(actionCache, fmt.Sprintf("%v", v))
	}
	
	workflowCache[a.actionID] = actionCache
	a.setWorkflowCache(workflowKey, workflowCache)
}

// getWorkflowCache 获取工作流缓存
func (a *BaseActionImpl) getWorkflowCache(key string) (map[string]interface{}, error) {
	value, err := a.systemConfigRepo.GetValue(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow cache: %w", err)
	}
	
	if value == "" {
		return make(map[string]interface{}), nil
	}
	
	// 这里应该解析JSON，简化处理
	return make(map[string]interface{}), nil
}

// setWorkflowCache 设置工作流缓存
func (a *BaseActionImpl) setWorkflowCache(key string, cache map[string]interface{}) {
	// 这里应该序列化为JSON，简化处理
	_ = a.systemConfigRepo.SetValue(key, fmt.Sprintf("%v", cache), "string", "工作流缓存")
}

// ActionChain 动作链
type ActionChain struct {
	actions []BaseAction
}

// NewActionChain 创建动作链
func NewActionChain() *ActionChain {
	return &ActionChain{
		actions: []BaseAction{},
	}
}

// AddAction 添加动作
func (ac *ActionChain) AddAction(action BaseAction) {
	ac.actions = append(ac.actions, action)
}

// Execute 执行动作链
func (ac *ActionChain) Execute(workflowID int, params map[string]interface{}, context *ActionContext) (*ActionContext, error) {
	for _, action := range ac.actions {
		if context.ShouldStop {
			break
		}
		
		// 检查缓存
		cacheKey := fmt.Sprintf("%s_%s_%d", action.Name(), "cache", workflowID)
		if action.CheckCache(workflowID, cacheKey) {
			zap.Info("Action already processed, skipping", 
				zap.String("action", action.Name()),
				zap.String("cache_key", cacheKey))
			continue
		}
		
		// 执行动作
		actionContext, err := action.Execute(workflowID, params, context)
		if err != nil {
			zap.Error("Action execution failed", 
				zap.String("action", action.Name()),
				zap.Error(err))
			return context, fmt.Errorf("action %s failed: %w", action.Name(), err)
		}
		
		// 保存缓存
		action.SaveCache(workflowID, cacheKey)
		
		// 检查是否完成
		if action.Done() {
			zap.Info("Action completed", 
				zap.String("action", action.Name()),
				zap.String("message", action.Message()))
		}
		
		// 更新上下文
		context = actionContext
	}
	
	return context, nil
}
