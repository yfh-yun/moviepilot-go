package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/repository/models"
	"strings"
	"time"
)

// WorkflowService 工作流服务
type WorkflowService struct {
	workflowRepo interfaces.WorkflowRepository
	logger       Logger
}

// Logger 日志接口
type Logger interface {
	Info(ctx context.Context, msg string, fields ...interface{})
	Error(ctx context.Context, msg string, fields ...interface{})
	Debug(ctx context.Context, msg string, fields ...interface{})
}

// NewWorkflowService 创建工作流服务实例
func NewWorkflowService(workflowRepo interfaces.WorkflowRepository, logger Logger) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		logger:       logger,
	}
}

// CreateWorkflowRequest 创建工作流请求
// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" validate:"required,oneof=media download subscribe notification system"`
	Trigger     string                 `json:"trigger" validate:"required"`
	Actions     []WorkflowAction       `json:"actions" validate:"required,min=1"`
	Conditions  []WorkflowCondition    `json:"conditions"`
	Priority    int                    `json:"priority" validate:"min=0,max=1000"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
}

// WorkflowAction 工作流动作
type WorkflowAction struct {
	Name     string                 `json:"name" validate:"required"`
	Type     string                 `json:"type" validate:"required"`
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority" validate:"min=0,max=1000"`
}

// WorkflowCondition 工作流条件
type WorkflowCondition struct {
	Type     string                 `json:"type" validate:"required"`
	Operator string                 `json:"operator" validate:"required"`
	Value    interface{}            `json:"value"`
	Config   map[string]interface{} `json:"config"`
}

// UpdateWorkflowRequest 更新工作流请求
type UpdateWorkflowRequest struct {
	ID          uint                   `json:"id" validate:"required"`
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" validate:"required,oneof=media download subscribe notification system"`
	Trigger     string                 `json:"trigger" validate:"required"`
	Actions     []WorkflowAction       `json:"actions" validate:"required,min=1"`
	Conditions  []WorkflowCondition    `json:"conditions"`
	Priority    int                    `json:"priority" validate:"min=0,max=1000"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
}

// ListWorkflowsRequest 工作流列表请求
type ListWorkflowsRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Keyword  string `json:"keyword"`
	Type     string `json:"type"`
	Enabled  *bool  `json:"enabled"`
}

// WorkflowResponse 工作流响应
type WorkflowResponse struct {
	ID            uint                   `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`
	Trigger       string                 `json:"trigger"`
	Actions       []WorkflowAction       `json:"actions"`
	Conditions    []WorkflowCondition    `json:"conditions"`
	Priority      int                    `json:"priority"`
	Enabled       bool                   `json:"enabled"`
	Status        string                 `json:"status"`
	LastExecution *time.Time             `json:"last_execution"`
	Config        map[string]interface{} `json:"config"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CreateWorkflow 创建工作流
func (s *WorkflowService) CreateWorkflow(ctx context.Context, req *CreateWorkflowRequest) (*WorkflowResponse, error) {
	s.logger.Info(ctx, "Creating workflow", "name", req.Name, "type", req.Type)

	// 转换为模型
	workflow := &models.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Trigger:     req.Trigger,
		Priority:    req.Priority,
		Enabled:     req.Enabled,
		Status:      "idle",
	}

	// 序列化动作和条件
	actionsJSON, err := json.Marshal(req.Actions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal actions: %w", err)
	}
	workflow.Actions = string(actionsJSON)

	if len(req.Conditions) > 0 {
		conditionsJSON, err := json.Marshal(req.Conditions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal conditions: %w", err)
		}
		workflow.Conditions = string(conditionsJSON)
	}

	// 序列化配置
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		workflow.Config = string(configJSON)
	}

	// 创建记录
	if err := s.workflowRepo.Create(workflow); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return s.toWorkflowResponse(workflow), nil
}

// GetWorkflow 获取工作流详情
func (s *WorkflowService) GetWorkflow(ctx context.Context, id uint) (*WorkflowResponse, error) {
	s.logger.Debug(ctx, "Getting workflow", "id", id)

	workflow, err := s.workflowRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow %d: %w", id, err)
	}
	if workflow == nil {
		return nil, errors.New("workflow not found")
	}

	return s.toWorkflowResponse(workflow), nil
}

// UpdateWorkflow 更新工作流
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, req *UpdateWorkflowRequest) (*WorkflowResponse, error) {
	s.logger.Info(ctx, "Updating workflow", "id", req.ID)

	// 获取现有工作流
	existingWorkflow, err := s.workflowRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow %d: %w", req.ID, err)
	}
	if existingWorkflow == nil {
		return nil, errors.New("workflow not found")
	}

	// 更新字段
	existingWorkflow.Name = req.Name
	existingWorkflow.Description = req.Description
	existingWorkflow.Type = req.Type
	existingWorkflow.Trigger = req.Trigger
	existingWorkflow.Priority = req.Priority
	existingWorkflow.Enabled = req.Enabled

	// 序列化动作和条件
	actionsJSON, err := json.Marshal(req.Actions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal actions: %w", err)
	}
	existingWorkflow.Actions = string(actionsJSON)

	if len(req.Conditions) > 0 {
		conditionsJSON, err := json.Marshal(req.Conditions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal conditions: %w", err)
		}
		existingWorkflow.Conditions = string(conditionsJSON)
	} else {
		existingWorkflow.Conditions = ""
	}

	// 序列化配置
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		existingWorkflow.Config = string(configJSON)
	} else {
		existingWorkflow.Config = ""
	}

	// 更新记录
	if err := s.workflowRepo.Update(existingWorkflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return s.toWorkflowResponse(existingWorkflow), nil
}

// DeleteWorkflow 删除工作流
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id uint) error {
	s.logger.Info(ctx, "Deleting workflow", "id", id)

	if err := s.workflowRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete workflow %d: %w", id, err)
	}

	return nil
}

// ListWorkflows 获取工作流列表
func (s *WorkflowService) ListWorkflows(ctx context.Context, req *ListWorkflowsRequest) (*WorkflowListResponse, error) {
	s.logger.Debug(ctx, "Listing workflows", "page", req.Page, "page_size", req.PageSize)

	offset := (req.Page - 1) * req.PageSize

	var workflows []*models.Workflow
	var total int64
	var err error

	if req.Keyword != "" {
		workflows, total, err = s.workflowRepo.Search(req.Keyword, offset, req.PageSize)
	} else {
		workflows, total, err = s.workflowRepo.List(offset, req.PageSize)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	// 转换为响应
	responses := make([]*WorkflowResponse, 0, len(workflows))
	for _, workflow := range workflows {
		responses = append(responses, s.toWorkflowResponse(workflow))
	}

	return &WorkflowListResponse{
		Workflows:  responses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (int(total) + req.PageSize - 1) / req.PageSize,
	}, nil
}

// ExecuteWorkflow 执行工作流
func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, id uint, triggerData map[string]interface{}) (*WorkflowExecutionResult, error) {
	s.logger.Info(ctx, "Executing workflow", "id", id)

	// 获取工作流
	workflow, err := s.workflowRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow %d: %w", id, err)
	}
	if workflow == nil {
		return nil, errors.New("workflow not found")
	}

	// 检查工作流是否启用
	if !workflow.Enabled {
		return &WorkflowExecutionResult{
			WorkflowID: id,
			Status:     "skipped",
			Reason:     "workflow is disabled",
		}, nil
	}

	// 解析动作
	var actions []WorkflowAction
	if err := json.Unmarshal([]byte(workflow.Actions), &actions); err != nil {
		return nil, fmt.Errorf("failed to parse workflow actions: %w", err)
	}

	// 解析条件（如果有）
	var conditions []WorkflowCondition
	if workflow.Conditions != "" {
		if err := json.Unmarshal([]byte(workflow.Conditions), &conditions); err != nil {
			return nil, fmt.Errorf("failed to parse workflow conditions: %w", err)
		}
	}

	// 解析配置
	var config map[string]interface{}
	if workflow.Config != "" {
		if err := json.Unmarshal([]byte(workflow.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to parse workflow config: %w", err)
		}
	}

	// 验证条件
	if len(conditions) > 0 {
		if !s.evaluateConditions(conditions, triggerData) {
			return &WorkflowExecutionResult{
				WorkflowID: id,
				Status:     "skipped",
				Reason:     "conditions not met",
			}, nil
		}
	}

	// 执行动作
	result := &WorkflowExecutionResult{
		WorkflowID:   id,
		WorkflowName: workflow.Name,
		Status:       "running",
		StartTime:    time.Now(),
		Actions:      make([]ActionExecutionResult, 0, len(actions)),
	}

	for i, action := range actions {
		actionResult := s.executeAction(ctx, action, triggerData, config)
		actionResult.Order = i + 1
		result.Actions = append(result.Actions, actionResult)

		// 如果动作失败且配置了停止执行，则中断
		if actionResult.Status == "failed" && config != nil {
			if stopOnFailure, ok := config["stop_on_failure"].(bool); ok && stopOnFailure {
				result.Status = "failed"
				result.Reason = fmt.Sprintf("action '%s' failed", action.Name)
				break
			}
		}
	}

	// 更新结果状态
	if result.Status == "running" {
		result.Status = "completed"
	}
	result.EndTime = time.Now()

	// 更新最后执行时间
	if err := s.workflowRepo.UpdateLastExecution(id, result.StartTime); err != nil {
		s.logger.Error(ctx, "Failed to update workflow last execution time", "id", id, "error", err)
	}

	return result, nil
}

// WorkflowListResponse 工作流列表响应
type WorkflowListResponse struct {
	Workflows  []*WorkflowResponse `json:"workflows"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// WorkflowExecutionResult 工作流执行结果
type WorkflowExecutionResult struct {
	WorkflowID   uint                    `json:"workflow_id"`
	WorkflowName string                  `json:"workflow_name"`
	Status       string                  `json:"status"` // running, completed, failed, skipped
	Reason       string                  `json:"reason"`
	StartTime    time.Time               `json:"start_time"`
	EndTime      time.Time               `json:"end_time"`
	Actions      []ActionExecutionResult `json:"actions"`
}

// ActionExecutionResult 动作执行结果
type ActionExecutionResult struct {
	Order    int                    `json:"order"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"` // success, failed, skipped
	Error    string                 `json:"error"`
	Duration time.Duration          `json:"duration"`
	Result   map[string]interface{} `json:"result"`
}

// evaluateConditions 评估条件
func (s *WorkflowService) evaluateConditions(conditions []WorkflowCondition, data map[string]interface{}) bool {
	for _, condition := range conditions {
		if !s.evaluateCondition(condition, data) {
			return false
		}
	}
	return true
}

// evaluateCondition 评估单个条件
func (s *WorkflowService) evaluateCondition(condition WorkflowCondition, data map[string]interface{}) bool {
	// 这里实现简单的条件评估逻辑
	// 实际实现应该根据条件类型和操作符进行更复杂的评估

	value, exists := data[condition.Type]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "equals":
		return value == condition.Value
	case "not_equals":
		return value != condition.Value
	case "contains":
		if str, ok := value.(string); ok {
			if target, ok := condition.Value.(string); ok {
				return strings.Contains(str, target)
			}
		}
		return false
	case "greater_than":
		return s.compareNumbers(value, condition.Value, func(a, b float64) bool { return a > b })
	case "less_than":
		return s.compareNumbers(value, condition.Value, func(a, b float64) bool { return a < b })
	default:
		return false
	}
}

// compareNumbers 比较数字
func (s *WorkflowService) compareNumbers(a, b interface{}, compareFunc func(float64, float64) bool) bool {
	aNum, ok := toFloat64(a)
	if !ok {
		return false
	}

	bNum, ok := toFloat64(b)
	if !ok {
		return false
	}

	return compareFunc(aNum, bNum)
}

// toFloat64 转换为float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// executeAction 执行动作
func (s *WorkflowService) executeAction(ctx context.Context, action WorkflowAction, triggerData map[string]interface{}, config map[string]interface{}) ActionExecutionResult {
	startTime := time.Now()
	result := ActionExecutionResult{
		Name:   action.Name,
		Type:   action.Type,
		Status: "success",
		Result: make(map[string]interface{}),
	}

	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// 这里实现具体的动作执行逻辑
	// 根据action.Type调用不同的执行器

	switch action.Type {
	case "download":
		result = s.executeDownloadAction(ctx, action, triggerData)
	case "notification":
		result = s.executeNotificationAction(ctx, action, triggerData)
	case "media_processing":
		result = s.executeMediaProcessingAction(ctx, action, triggerData)
	case "file_operation":
		result = s.executeFileOperationAction(ctx, action, triggerData)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown action type: %s", action.Type)
	}

	return result
}

// executeDownloadAction 执行下载动作
func (s *WorkflowService) executeDownloadAction(ctx context.Context, action WorkflowAction, triggerData map[string]interface{}) ActionExecutionResult {
	// 模拟下载动作执行
	s.logger.Info(ctx, "Executing download action", "name", action.Name)

	return ActionExecutionResult{
		Name:   action.Name,
		Type:   action.Type,
		Status: "success",
		Result: map[string]interface{}{
			"message": "Download action executed successfully",
		},
	}
}

// executeNotificationAction 执行通知动作
func (s *WorkflowService) executeNotificationAction(ctx context.Context, action WorkflowAction, triggerData map[string]interface{}) ActionExecutionResult {
	// 模拟通知动作执行
	s.logger.Info(ctx, "Executing notification action", "name", action.Name)

	return ActionExecutionResult{
		Name:   action.Name,
		Type:   action.Type,
		Status: "success",
		Result: map[string]interface{}{
			"message": "Notification action executed successfully",
		},
	}
}

// executeMediaProcessingAction 执行媒体处理动作
func (s *WorkflowService) executeMediaProcessingAction(ctx context.Context, action WorkflowAction, triggerData map[string]interface{}) ActionExecutionResult {
	// 模拟媒体处理动作执行
	s.logger.Info(ctx, "Executing media processing action", "name", action.Name)

	return ActionExecutionResult{
		Name:   action.Name,
		Type:   action.Type,
		Status: "success",
		Result: map[string]interface{}{
			"message": "Media processing action executed successfully",
		},
	}
}

// executeFileOperationAction 执行文件操作动作
func (s *WorkflowService) executeFileOperationAction(ctx context.Context, action WorkflowAction, triggerData map[string]interface{}) ActionExecutionResult {
	// 模拟文件操作动作执行
	s.logger.Info(ctx, "Executing file operation action", "name", action.Name)

	return ActionExecutionResult{
		Name:   action.Name,
		Type:   action.Type,
		Status: "success",
		Result: map[string]interface{}{
			"message": "File operation action executed successfully",
		},
	}
}

// toWorkflowResponse 转换为工作流响应
func (s *WorkflowService) toWorkflowResponse(workflow *models.Workflow) *WorkflowResponse {
	response := &WorkflowResponse{
		ID:            workflow.ID,
		Name:          workflow.Name,
		Description:   workflow.Description,
		Type:          workflow.Type,
		Trigger:       workflow.Trigger,
		Priority:      workflow.Priority,
		Enabled:       workflow.Enabled,
		Status:        workflow.Status,
		LastExecution: workflow.LastExecution,
		CreatedAt:     workflow.CreatedAt,
		UpdatedAt:     workflow.UpdatedAt,
	}

	// 解析动作
	if workflow.Actions != "" {
		var actions []WorkflowAction
		if err := json.Unmarshal([]byte(workflow.Actions), &actions); err == nil {
			response.Actions = actions
		}
	}

	// 解析条件
	if workflow.Conditions != "" {
		var conditions []WorkflowCondition
		if err := json.Unmarshal([]byte(workflow.Conditions), &conditions); err == nil {
			response.Conditions = conditions
		}
	}

	// 解析配置
	if workflow.Config != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(workflow.Config), &config); err == nil {
			response.Config = config
		}
	}

	return response
}
