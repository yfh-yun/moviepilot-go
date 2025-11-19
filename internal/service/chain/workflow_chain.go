package chain

import (
	"context"
	"errors"
	"fmt"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/repository"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
)

// WorkflowChain 工作流处理链
type WorkflowChain struct {
	cache               *cache.Cache
	logger              *logger.Logger
	workflowRepo        *repository.WorkflowRepository
	workflowService     *service.WorkflowService
	workflowManager     *WorkflowManager
	workflowShareHelper *WorkflowShareHelper
}

// NewWorkflowChain 创建工作流处理链实例
func NewWorkflowChain(cache *cache.Cache, logger *logger.Logger,
	workflowRepo *repository.WorkflowRepository, serverHost string, shareEnabled bool, proxyURL string) *WorkflowChain {

	// 创建工作流服务
	workflowService := service.NewWorkflowService(workflowRepo, logger)

	// 创建工作流管理器
	workflowManager := NewWorkflowManager(cache, logger, workflowRepo)

	// 创建工作流分享助手
	workflowShareHelper := NewWorkflowShareHelper(cache, logger, workflowRepo, serverHost, shareEnabled, proxyURL)

	return &WorkflowChain{
		cache:               cache,
		logger:              logger,
		workflowRepo:        workflowRepo,
		workflowService:     workflowService,
		workflowManager:     workflowManager,
		workflowShareHelper: workflowShareHelper,
	}
}

// ExecuteWorkflow 执行工作流
func (c *WorkflowChain) ExecuteWorkflow(ctx context.Context, workflowID int64, fromBeginning bool) (*model.WorkflowExecutionResult, error) {
	c.logger.Info("执行工作流", "workflowID", workflowID, "fromBeginning", fromBeginning)

	// 使用工作流管理器执行工作流
	result, err := c.workflowManager.ExecuteWorkflow(ctx, workflowID, fromBeginning)
	if err != nil {
		c.logger.Error("执行工作流失败", "error", err)
		return nil, err
	}

	c.logger.Info("工作流执行完成", "workflowID", workflowID, "success", result.Success, "message", result.Message)
	return result, nil
}

// validateWorkflow 验证工作流
func (c *WorkflowChain) validateWorkflow(workflow *model.Workflow) error {
	if workflow == nil {
		return errors.New("工作流不能为空")
	}

	if len(workflow.Actions) == 0 {
		return errors.New("工作流无动作")
	}

	if len(workflow.Flows) == 0 {
		return errors.New("工作流无流程")
	}

	return nil
}

// CreateWorkflow 创建工作流
func (c *WorkflowChain) CreateWorkflow(ctx context.Context, workflowData model.WorkflowCreateData) (*model.Workflow, error) {
	c.logger.Info("创建工作流", "name", workflowData.Name)

	workflow, err := c.workflowService.CreateWorkflow(ctx, workflowData)
	if err != nil {
		c.logger.Error("创建工作流失败", "error", err)
		return nil, err
	}

	c.logger.Info("创建工作流成功", "workflowID", workflow.ID)
	return workflow, nil
}

// UpdateWorkflow 更新工作流
func (c *WorkflowChain) UpdateWorkflow(ctx context.Context, workflowID int64, updateData model.WorkflowUpdateData) (*model.Workflow, error) {
	c.logger.Info("更新工作流", "workflowID", workflowID)

	workflow, err := c.workflowService.UpdateWorkflow(ctx, workflowID, updateData)
	if err != nil {
		c.logger.Error("更新工作流失败", "error", err)
		return nil, err
	}

	c.logger.Info("更新工作流成功", "workflowID", workflowID)
	return workflow, nil
}

// DeleteWorkflow 删除工作流
func (c *WorkflowChain) DeleteWorkflow(ctx context.Context, workflowID int64) error {
	c.logger.Info("删除工作流", "workflowID", workflowID)

	err := c.workflowService.DeleteWorkflow(ctx, workflowID)
	if err != nil {
		c.logger.Error("删除工作流失败", "error", err)
		return err
	}

	c.logger.Info("删除工作流成功", "workflowID", workflowID)
	return nil
}

// GetWorkflowList 获取工作流列表
func (c *WorkflowChain) GetWorkflowList(ctx context.Context, page, pageSize int) ([]*model.Workflow, int64, error) {
	c.logger.Info("获取工作流列表", "page", page, "pageSize", pageSize)

	workflows, total, err := c.workflowRepo.GetWorkflowList(ctx, page, pageSize)
	if err != nil {
		c.logger.Error("获取工作流列表失败", "error", err)
		return nil, 0, err
	}

	c.logger.Info("获取工作流列表成功", "count", len(workflows))
	return workflows, total, nil
}

// GetEnabledWorkflows 获取启用的工作流列表
func (c *WorkflowChain) GetEnabledWorkflows(ctx context.Context) ([]*model.Workflow, error) {
	c.logger.Info("获取启用的工作流列表")

	workflows, err := c.workflowRepo.GetEnabledWorkflows(ctx)
	if err != nil {
		c.logger.Error("获取启用的工作流列表失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取启用的工作流列表成功", "count", len(workflows))
	return workflows, nil
}

// GetTimerTriggeredWorkflows 获取定时触发的工作流列表
func (c *WorkflowChain) GetTimerTriggeredWorkflows(ctx context.Context) ([]*model.Workflow, error) {
	c.logger.Info("获取定时触发的工作流列表")

	workflows, err := c.workflowRepo.GetTimerTriggeredWorkflows(ctx)
	if err != nil {
		c.logger.Error("获取定时触发的工作流列表失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取定时触发的工作流列表成功", "count", len(workflows))
	return workflows, nil
}

// GetEventTriggeredWorkflows 获取事件触发的工作流列表
func (c *WorkflowChain) GetEventTriggeredWorkflows(ctx context.Context) ([]*model.Workflow, error) {
	c.logger.Info("获取事件触发的工作流列表")

	workflows, err := c.workflowRepo.GetEventTriggeredWorkflows(ctx)
	if err != nil {
		c.logger.Error("获取事件触发的工作流列表失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取事件触发的工作流列表成功", "count", len(workflows))
	return workflows, nil
}

// EnableWorkflow 启用工作流
func (c *WorkflowChain) EnableWorkflow(ctx context.Context, workflowID int64) error {
	c.logger.Info("启用工作流", "workflowID", workflowID)

	err := c.workflowService.EnableWorkflow(ctx, workflowID)
	if err != nil {
		c.logger.Error("启用工作流失败", "error", err)
		return err
	}

	c.logger.Info("启用工作流成功", "workflowID", workflowID)
	return nil
}

// DisableWorkflow 禁用工作流
func (c *WorkflowChain) DisableWorkflow(ctx context.Context, workflowID int64) error {
	c.logger.Info("禁用工作流", "workflowID", workflowID)

	err := c.workflowService.DisableWorkflow(ctx, workflowID)
	if err != nil {
		c.logger.Error("禁用工作流失败", "error", err)
		return err
	}

	c.logger.Info("禁用工作流成功", "workflowID", workflowID)
	return nil
}

// GetWorkflowExecutionHistory 获取工作流执行历史
func (c *WorkflowChain) GetWorkflowExecutionHistory(ctx context.Context, workflowID int64, page, pageSize int) ([]*model.WorkflowExecution, int64, error) {
	c.logger.Info("获取工作流执行历史", "workflowID", workflowID, "page", page, "pageSize", pageSize)

	executions, total, err := c.workflowRepo.GetWorkflowExecutionHistory(ctx, workflowID, page, pageSize)
	if err != nil {
		c.logger.Error("获取工作流执行历史失败", "error", err)
		return nil, 0, err
	}

	c.logger.Info("获取工作流执行历史成功", "count", len(executions))
	return executions, total, nil
}

// GetWorkflowStats 获取工作流统计信息
func (c *WorkflowChain) GetWorkflowStats(ctx context.Context) (*model.WorkflowStats, error) {
	c.logger.Info("获取工作流统计信息")

	stats, err := c.workflowRepo.GetWorkflowStats(ctx)
	if err != nil {
		c.logger.Error("获取工作流统计信息失败", "error", err)
		return nil, err
	}

	return stats, nil
}

// GetListActions 获取所有可用动作列表
func (c *WorkflowChain) GetListActions() []map[string]interface{} {
	c.logger.Info("获取工作流动作列表")
	return c.workflowManager.GetListActions()
}

// HandleEvent 处理工作流事件
func (c *WorkflowChain) HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error {
	c.logger.Info("处理工作流事件", "eventType", eventType)

	err := c.workflowManager.HandleEvent(ctx, eventType, eventData)
	if err != nil {
		c.logger.Error("处理工作流事件失败", "eventType", eventType, "error", err)
		return err
	}

	return nil
}

// StopWorkflow 停止工作流执行
func (c *WorkflowChain) StopWorkflow(ctx context.Context, workflowID int64) error {
	c.logger.Info("停止工作流", "workflowID", workflowID)

	c.workflowManager.StopWorkflow(workflowID)
	return nil
}

// GetRunningWorkflows 获取运行中的工作流
func (c *WorkflowChain) GetRunningWorkflows(ctx context.Context) map[int64]*WorkflowExecutionContext {
	c.logger.Info("获取运行中的工作流")
	return c.workflowManager.GetRunningWorkflows()
}

// LoadWorkflowEvents 加载工作流事件触发器
func (c *WorkflowChain) LoadWorkflowEvents(ctx context.Context, workflowID *int64) error {
	c.logger.Info("加载工作流事件触发器")

	err := c.workflowManager.LoadWorkflowEvents(workflowID)
	if err != nil {
		c.logger.Error("加载工作流事件触发器失败", "error", err)
		return err
	}

	return nil
}

// UpdateWorkflowEvent 更新工作流事件触发器
func (c *WorkflowChain) UpdateWorkflowEvent(ctx context.Context, workflow *model.Workflow) error {
	c.logger.Info("更新工作流事件触发器", "workflowID", workflow.ID)

	err := c.workflowManager.UpdateWorkflowEvent(workflow)
	if err != nil {
		c.logger.Error("更新工作流事件触发器失败", "workflowID", workflow.ID, "error", err)
		return err
	}

	return nil
}

// GetEventWorkflows 获取所有事件触发的工作流
func (c *WorkflowChain) GetEventWorkflows(ctx context.Context) map[string][]int64 {
	c.logger.Info("获取事件触发的工作流")
	return c.workflowManager.GetEventWorkflows()
}

// WorkflowShare 分享工作流
func (c *WorkflowChain) WorkflowShare(ctx context.Context, workflowID int64,
	shareTitle, shareComment, shareUser string) (bool, string) {

	c.logger.Info("分享工作流", "workflowID", workflowID, "title", shareTitle)

	// 验证参数
	if err := c.workflowShareHelper.ValidateShareRequest(shareTitle, shareComment, shareUser); err != nil {
		return false, err.Error()
	}

	success, message := c.workflowShareHelper.WorkflowShare(ctx, workflowID, shareTitle, shareComment, shareUser)
	if success {
		c.logger.Info("工作流分享成功", "workflowID", workflowID)
	} else {
		c.logger.Error("工作流分享失败", "workflowID", workflowID, "error", message)
	}

	return success, message
}

// ShareDelete 删除工作流分享
func (c *WorkflowChain) ShareDelete(ctx context.Context, shareID int64) (bool, string) {
	c.logger.Info("删除工作流分享", "shareID", shareID)

	success, message := c.workflowShareHelper.ShareDelete(ctx, shareID)
	if success {
		c.logger.Info("工作流分享删除成功", "shareID", shareID)
	} else {
		c.logger.Error("工作流分享删除失败", "shareID", shareID, "error", message)
	}

	return success, message
}

// WorkflowFork 复用分享的工作流
func (c *WorkflowChain) WorkflowFork(ctx context.Context, shareID int64) (bool, string) {
	c.logger.Info("复用分享的工作流", "shareID", shareID)

	success, message := c.workflowShareHelper.WorkflowFork(ctx, shareID)
	if success {
		c.logger.Info("工作流复用成功", "shareID", shareID)
	} else {
		c.logger.Error("工作流复用失败", "shareID", shareID, "error", message)
	}

	return success, message
}

// GetShares 获取工作流分享列表
func (c *WorkflowChain) GetShares(ctx context.Context, name *string, page, count int) ([]WorkflowShareItem, error) {
	c.logger.Info("获取工作流分享列表", "page", page, "count", count)

	items, err := c.workflowShareHelper.GetShares(ctx, name, page, count)
	if err != nil {
		c.logger.Error("获取工作流分享列表失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取工作流分享列表成功", "count", len(items))
	return items, nil
}

// SearchShares 搜索工作流分享
func (c *WorkflowChain) SearchShares(ctx context.Context, keyword string, page, count int) ([]WorkflowShareItem, error) {
	c.logger.Info("搜索工作流分享", "keyword", keyword, "page", page, "count", count)

	items, err := c.workflowShareHelper.SearchShares(ctx, keyword, page, count)
	if err != nil {
		c.logger.Error("搜索工作流分享失败", "keyword", keyword, "error", err)
		return nil, err
	}

	c.logger.Info("搜索工作流分享成功", "keyword", keyword, "count", len(items))
	return items, nil
}

// GetPopularShares 获取热门工作流分享
func (c *WorkflowChain) GetPopularShares(ctx context.Context, limit int) ([]WorkflowShareItem, error) {
	c.logger.Info("获取热门工作流分享", "limit", limit)

	items, err := c.workflowShareHelper.GetPopularShares(ctx, limit)
	if err != nil {
		c.logger.Error("获取热门工作流分享失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取热门工作流分享成功", "count", len(items))
	return items, nil
}

// GetMyShares 获取我的工作流分享
func (c *WorkflowChain) GetMyShares(ctx context.Context, page, count int) ([]WorkflowShareItem, error) {
	c.logger.Info("获取我的工作流分享", "page", page, "count", count)

	items, err := c.workflowShareHelper.GetMyShares(ctx, page, count)
	if err != nil {
		c.logger.Error("获取我的工作流分享失败", "error", err)
		return nil, err
	}

	c.logger.Info("获取我的工作流分享成功", "count", len(items))
	return items, nil
}

// GetShareStats 获取工作流分享统计信息
func (c *WorkflowChain) GetShareStats(ctx context.Context) (map[string]interface{}, error) {
	c.logger.Info("获取工作流分享统计信息")

	stats, err := c.workflowShareHelper.GetShareStats(ctx)
	if err != nil {
		c.logger.Error("获取工作流分享统计信息失败", "error", err)
		return nil, err
	}

	return stats, nil
}

// CreateWorkflowFromShare 从分享创建工作流
func (c *WorkflowChain) CreateWorkflowFromShare(ctx context.Context, shareID int64, userID int64) (*model.Workflow, error) {
	c.logger.Info("从分享创建工作流", "shareID", shareID)

	// 首先获取分享数据
	shares, err := c.workflowShareHelper.GetShares(ctx, nil, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("获取分享数据失败: %w", err)
	}

	var targetShare *WorkflowShareItem
	for _, share := range shares {
		if share.ID == shareID {
			targetShare = &share
			break
		}
	}

	if targetShare == nil {
		return nil, fmt.Errorf("未找到指定的分享")
	}

	// 从分享创建工作流
	workflow, err := c.workflowShareHelper.CreateWorkflowFromShare(ctx, targetShare, userID)
	if err != nil {
		c.logger.Error("从分享创建工作流失败", "shareID", shareID, "error", err)
		return nil, err
	}

	c.logger.Info("从分享创建工作流成功", "shareID", shareID, "workflowID", workflow.ID)
	return workflow, nil
}

// GetUserUUID 获取用户UUID
func (c *WorkflowChain) GetUserUUID() string {
	uuid := c.workflowShareHelper.GetUserUUID()
	c.logger.Debug("获取用户UUID", "uuid", uuid)
	return uuid
}

// ParseWorkflowActions 解析工作流动作JSON
func (c *WorkflowChain) ParseWorkflowActions(actionsJSON string) ([]model.WorkflowAction, error) {
	return c.workflowShareHelper.ParseWorkflowActions(actionsJSON)
}

// ParseWorkflowFlows 解析工作流流程JSON
func (c *WorkflowChain) ParseWorkflowFlows(flowsJSON string) ([]model.WorkflowFlow, error) {
	return c.workflowShareHelper.ParseWorkflowFlows(flowsJSON)
}

// Stop 停止工作流链
func (c *WorkflowChain) Stop() {
	c.logger.Info("停止工作流链")

	if c.workflowManager != nil {
		c.workflowManager.Stop()
	}
}
