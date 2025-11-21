package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/pkg/utils"
)

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
	logger          logger.Logger
	workflowRepo    repository.WorkflowRepository
	variableStore   *VariableStore
	conditionEvaluator *ConditionEvaluator
	actionExecutor  *ActionExecutor
	loopController  *LoopController
	errorHandler    *ErrorHandler
}

// NewWorkflowEngine 创建工作流引擎
func NewWorkflowEngine(
	logger logger.Logger,
	workflowRepo repository.WorkflowRepository,
) *WorkflowEngine {
	return &WorkflowEngine{
		logger:              logger,
		workflowRepo:        workflowRepo,
		variableStore:       NewVariableStore(),
		conditionEvaluator:  NewConditionEvaluator(),
		actionExecutor:      NewActionExecutor(),
		loopController:      NewLoopController(),
		errorHandler:        NewErrorHandler(),
	}
}

// ExecuteWorkflow 执行工作流
func (e *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID int64, input map[string]interface{}, fromBeginning bool) (*model.WorkflowExecutionResult, error) {
	// 获取工作流定义
	workflow, err := e.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("获取工作流失败: %w", err)
	}

	if workflow == nil {
		return nil, fmt.Errorf("工作流不存在: %d", workflowID)
	}

	// 创建执行上下文
	executionContext := &WorkflowExecutionContext{
		WorkflowID:  workflowID,
		Variables:    make(map[string]interface{}),
		Input:        input,
		StartedAt:    time.Now(),
		FromBeginning: fromBeginning,
		Status:       model.WorkflowStatusRunning,
	}

	// 初始化变量
	if input != nil {
		for key, value := range input {
			executionContext.Variables[key] = value
		}
	}

	// 执行工作流
	result, err := e.executeWorkflowInternal(ctx, workflow, executionContext)
	if err != nil {
		e.errorHandler.HandleError(ctx, executionContext, err)
		return &model.WorkflowExecutionResult{
			Success:  false,
			Message:  err.Error(),
			Variables: executionContext.Variables,
		}, nil
	}

	result.Variables = executionContext.Variables
	return result, nil
}

// executeWorkflowInternal 内部执行工作流
func (e *WorkflowEngine) executeWorkflowInternal(ctx context.Context, workflow *model.Workflow, context *WorkflowExecutionContext) (*model.WorkflowExecutionResult, error) {
	e.logger.Info("开始执行工作流", 
		logger.Int64("workflow_id", workflow.ID),
		logger.String("name", workflow.Name))

	// 执行前置条件
	if workflow.PreConditions != nil && len(workflow.PreConditions) > 0 {
		preConditionMet, err := e.evaluateConditions(ctx, workflow.PreConditions, context)
		if err != nil {
			return nil, fmt.Errorf("评估前置条件失败: %w", err)
		}

		if !preConditionMet {
			return &model.WorkflowExecutionResult{
				Success: true,
				Message: "前置条件未满足，跳过执行",
				Status:  model.WorkflowStatusSkipped,
			}, nil
		}
	}

	// 执行工作流步骤
	for _, flow := range workflow.Flows {
		result, err := e.executeFlow(ctx, flow, context)
		if err != nil {
			return nil, fmt.Errorf("执行流程失败: %w", err)
		}

		if !result.Success {
			// 根据错误处理策略决定是否继续
			if flow.ErrorHandling == model.ErrorHandlingStop {
				return result, nil
			} else if flow.ErrorHandling == model.ErrorHandlingRetry {
				// 实现重试逻辑
				return e.retryFlow(ctx, flow, context)
			}
			// ErrorHandlingContinue: 继续执行
		}
	}

	// 执行后置条件
	if workflow.PostConditions != nil && len(workflow.PostConditions) > 0 {
		postConditionMet, err := e.evaluateConditions(ctx, workflow.PostConditions, context)
		if err != nil {
			return nil, fmt.Errorf("评估后置条件失败: %w", err)
		}

		if !postConditionMet {
			return &model.WorkflowExecutionResult{
				Success: false,
				Message: "后置条件未满足",
				Status:  model.WorkflowStatusFailed,
			}, nil
		}
	}

	context.Status = model.WorkflowStatusCompleted
	context.FinishedAt = time.Now()

	return &model.WorkflowExecutionResult{
		Success:   true,
		Message:   "工作流执行成功",
		Status:    context.Status,
		StartTime: context.StartedAt,
		EndTime:   context.FinishedAt,
	}, nil
}

// executeFlow 执行单个流程
func (e *WorkflowEngine) executeFlow(ctx context.Context, flow *model.WorkflowFlow, context *WorkflowExecutionContext) (*model.FlowExecutionResult, error) {
	e.logger.Info("执行流程", logger.String("flow_name", flow.Name))

	// 处理循环逻辑
	if flow.Loop != nil {
		return e.executeLoopFlow(ctx, flow, context)
	}

	// 处理条件逻辑
	if flow.Condition != nil && len(flow.Condition.Conditions) > 0 {
		conditionMet, err := e.evaluateConditions(ctx, flow.Condition.Conditions, context)
		if err != nil {
			return nil, fmt.Errorf("评估条件失败: %w", err)
		}

		if !conditionMet {
			e.logger.Info("条件不满足，跳过流程", logger.String("flow_name", flow.Name))
			return &model.FlowExecutionResult{
				Success: true,
				Message: "条件不满足，跳过执行",
				Skipped: true,
			}, nil
		}
	}

	// 执行动作
	results := make([]*model.ActionExecutionResult, 0)
	for _, action := range flow.Actions {
		result, err := e.actionExecutor.ExecuteAction(ctx, action, context)
		if err != nil {
			e.logger.Error("执行动作失败", 
				logger.String("action_name", action.Name),
				logger.Error(err))
			
			return &model.FlowExecutionResult{
				Success: false,
				Message: fmt.Sprintf("执行动作失败: %s", err.Error()),
			}, nil
		}

		results = append(results, result)

		// 如果动作失败且流程配置为失败时停止
		if !result.Success && flow.StopOnFailure {
			return &model.FlowExecutionResult{
				Success: false,
				Message: fmt.Sprintf("动作执行失败，停止流程: %s", action.Name),
				Results: results,
			}, nil
		}
	}

	return &model.FlowExecutionResult{
		Success: true,
		Message: "流程执行成功",
		Results: results,
	}, nil
}

// executeLoopFlow 执行循环流程
func (e *WorkflowEngine) executeLoopFlow(ctx context.Context, flow *model.WorkflowFlow, context *WorkflowExecutionContext) (*model.FlowExecutionResult, error) {
	loop := flow.Loop
	results := make([]*model.ActionExecutionResult, 0)

	// 评估循环变量
	loopVariable, err := e.evaluateLoopVariable(ctx, loop.Variable, context)
	if err != nil {
		return nil, fmt.Errorf("评估循环变量失败: %w", err)
	}

	maxIterations := loop.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 100 // 默认最大循环次数
	}

	iterationCount := 0
	var loopResult *model.FlowExecutionResult

	// 执行循环
	switch loop.Type {
	case model.LoopTypeForEach:
		loopResult, err = e.executeForEachLoop(ctx, flow, loopVariable, loop, context, &iterationCount, maxIterations, &results)
	case model.LoopTypeWhile:
		loopResult, err = e.executeWhileLoop(ctx, flow, loop, context, &iterationCount, maxIterations, &results)
	case model.LoopTypeUntil:
		loopResult, err = e.executeUntilLoop(ctx, flow, loop, context, &iterationCount, maxIterations, &results)
	default:
		return nil, fmt.Errorf("不支持的循环类型: %s", loop.Type)
	}

	if err != nil {
		return nil, err
	}

	loopResult.Results = results
	return loopResult, nil
}

// executeForEachLoop 执行for each循环
func (e *WorkflowEngine) executeForEachLoop(ctx context.Context, flow *model.WorkflowFlow, loopVariable interface{}, loop *model.WorkflowLoop, context *WorkflowExecutionContext, iterationCount *int, maxIterations int, results *[]*model.ActionExecutionResult) (*model.FlowExecutionResult, error) {
	items, err := e.convertToSlice(loopVariable)
	if err != nil {
		return nil, fmt.Errorf("转换循环变量失败: %w", err)
	}

	e.logger.Info("开始forEach循环", 
		logger.String("item_name", loop.ItemName),
		logger.Int("total_items", len(items)),
		logger.Int("max_iterations", maxIterations))

	for index, item := range items {
		if *iterationCount >= maxIterations {
			e.logger.Warn("达到最大循环次数，停止执行", logger.Int("max_iterations", maxIterations))
			break
		}

		// 设置循环变量
		context.Variables[loop.ItemName] = item
		context.Variables[fmt.Sprintf("%s_index", loop.ItemName)] = index

		// 执行循环体
		result, err := e.executeLoopBody(ctx, flow, context)
		if err != nil {
			return nil, fmt.Errorf("执行循环体失败: %w", err)
		}

		*results = append(*results, result...)
		*iterationCount++

		// 检查是否应该跳出循环
		if loop.BreakCondition != nil {
			shouldBreak, err := e.evaluateConditions(ctx, []model.Condition{*loop.BreakCondition}, context)
			if err != nil {
				return nil, fmt.Errorf("评估跳出条件失败: %w", err)
			}
			if shouldBreak {
				e.logger.Info("满足跳出条件，结束循环")
				break
			}
		}
	}

	// 清理循环变量
	delete(context.Variables, loop.ItemName)
	delete(context.Variables, fmt.Sprintf("%s_index", loop.ItemName))

	return &model.FlowExecutionResult{
		Success: true,
		Message: fmt.Sprintf("forEach循环执行完成，共执行%d次", *iterationCount),
	}, nil
}

// executeWhileLoop 执行while循环
func (e *WorkflowEngine) executeWhileLoop(ctx context.Context, flow *model.WorkflowFlow, loop *model.WorkflowLoop, context *WorkflowExecutionContext, iterationCount *int, maxIterations int, results *[]*model.ActionExecutionResult) (*model.FlowExecutionResult, error) {
	e.logger.Info("开始while循环", logger.Int("max_iterations", maxIterations))

	for *iterationCount < maxIterations {
		// 评估循环条件
		conditionMet, err := e.evaluateConditions(ctx, []model.Condition{*loop.Condition}, context)
		if err != nil {
			return nil, fmt.Errorf("评估循环条件失败: %w", err)
		}

		if !conditionMet {
			e.logger.Info("循环条件不满足，退出循环")
			break
		}

		// 执行循环体
		result, err := e.executeLoopBody(ctx, flow, context)
		if err != nil {
			return nil, fmt.Errorf("执行循环体失败: %w", err)
		}

		*results = append(*results, result...)
		*iterationCount++
		context.Variables["loop_iteration"] = *iterationCount

		// 检查跳出条件
		if loop.BreakCondition != nil {
			shouldBreak, err := e.evaluateConditions(ctx, []model.Condition{*loop.BreakCondition}, context)
			if err != nil {
				return nil, fmt.Errorf("评估跳出条件失败: %w", err)
			}
			if shouldBreak {
				e.logger.Info("满足跳出条件，结束循环")
				break
			}
		}
	}

	delete(context.Variables, "loop_iteration")

	return &model.FlowExecutionResult{
		Success: true,
		Message: fmt.Sprintf("while循环执行完成，共执行%d次", *iterationCount),
	}, nil
}

// executeUntilLoop 执行until循环
func (e *WorkflowEngine) executeUntilLoop(ctx context.Context, flow *model.WorkflowFlow, loop *model.WorkflowLoop, context *WorkflowExecutionContext, iterationCount *int, maxIterations int, results *[]*model.ActionExecutionResult) (*model.FlowExecutionResult, error) {
	e.logger.Info("开始until循环", logger.Int("max_iterations", maxIterations))

	if loop.BreakCondition == nil {
		return nil, fmt.Errorf("until循环必须有跳出条件")
	}

	for *iterationCount < maxIterations {
		// 执行循环体
		result, err := e.executeLoopBody(ctx, flow, context)
		if err != nil {
			return nil, fmt.Errorf("执行循环体失败: %w", err)
		}

		*results = append(*results, result...)
		*iterationCount++
		context.Variables["loop_iteration"] = *iterationCount

		// 评估跳出条件
		shouldBreak, err := e.evaluateConditions(ctx, []model.Condition{*loop.BreakCondition}, context)
		if err != nil {
			return nil, fmt.Errorf("评估跳出条件失败: %w", err)
		}

		if shouldBreak {
			e.logger.Info("满足跳出条件，结束循环")
			break
		}
	}

	delete(context.Variables, "loop_iteration")

	return &model.FlowExecutionResult{
		Success: true,
		Message: fmt.Sprintf("until循环执行完成，共执行%d次", *iterationCount),
	}, nil
}

// executeLoopBody 执行循环体
func (e *WorkflowEngine) executeLoopBody(ctx context.Context, flow *model.WorkflowFlow, context *WorkflowExecutionContext) ([]*model.ActionExecutionResult, error) {
	results := make([]*model.ActionExecutionResult, 0)

	for _, action := range flow.Actions {
		result, err := e.actionExecutor.ExecuteAction(ctx, action, context)
		if err != nil {
			return nil, err
		}
		results = append(results, result)

		// 如果动作失败且配置为失败时停止
		if !result.Success && flow.StopOnFailure {
			break
		}
	}

	return results, nil
}

// evaluateConditions 评估条件
func (e *WorkflowEngine) evaluateConditions(ctx context.Context, conditions []model.Condition, context *WorkflowExecutionContext) (bool, error) {
	return e.conditionEvaluator.Evaluate(ctx, conditions, context)
}

// evaluateLoopVariable 评估循环变量
func (e *WorkflowEngine) evaluateLoopVariable(ctx context.Context, variable interface{}, context *WorkflowExecutionContext) (interface{}, error) {
	// 如果是字符串，尝试从变量中获取值
	if variableStr, ok := variable.(string); ok {
		if strings.HasPrefix(variableStr, "${") && strings.HasSuffix(variableStr, "}") {
			varName := strings.TrimPrefix(strings.TrimSuffix(variableStr, "}"), "${")
			if val, exists := context.Variables[varName]; exists {
				return val, nil
			}
		}
		// 尝试解析JSON
		var parsed interface{}
		if err := json.Unmarshal([]byte(variableStr), &parsed); err == nil {
			return parsed, nil
		}
		return variableStr, nil
	}

	return variable, nil
}

// convertToSlice 将接口转换为切片
func (e *WorkflowEngine) convertToSlice(value interface{}) ([]interface{}, error) {
	if value == nil {
		return []interface{}{}, nil
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Slice {
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = val.Index(i).Interface()
		}
		return result, nil
	}

	// 如果不是切片，包装成单元素切片
	return []interface{}{value}, nil
}

// retryFlow 重试流程
func (e *WorkflowEngine) retryFlow(ctx context.Context, flow *model.WorkflowFlow, context *WorkflowExecutionContext) (*model.FlowExecutionResult, error) {
	maxRetries := 3
	retryDelay := time.Second * 5

	for i := 0; i < maxRetries; i++ {
		e.logger.Info("重试执行流程", 
			logger.String("flow_name", flow.Name),
			logger.Int("retry_count", i+1))

		time.Sleep(retryDelay)

		result, err := e.executeFlow(ctx, flow, context)
		if err != nil {
			continue
		}

		if result.Success {
			return result, nil
		}
	}

	return &model.FlowExecutionResult{
		Success: false,
		Message: fmt.Sprintf("流程重试%d次后仍然失败", maxRetries),
	}, nil
}

// WorkflowExecutionContext 工作流执行上下文
type WorkflowExecutionContext struct {
	WorkflowID    int64
	Variables      map[string]interface{}
	Input          map[string]interface{}
	StartedAt      time.Time
	FinishedAt     time.Time
	Status         model.WorkflowStatus
	FromBeginning  bool
	CurrentFlowID  string
	ErrorMessage   string
}

// GetVariable 获取变量值
func (c *WorkflowExecutionContext) GetVariable(name string) (interface{}, bool) {
	value, exists := c.Variables[name]
	return value, exists
}

// SetVariable 设置变量值
func (c *WorkflowExecutionContext) SetVariable(name string, value interface{}) {
	if c.Variables == nil {
		c.Variables = make(map[string]interface{})
	}
	c.Variables[name] = value
}

// GetVariableAsString 获取字符串类型的变量值
func (c *WorkflowExecutionContext) GetVariableAsString(name string) string {
	if value, exists := c.GetVariable(name); exists {
		return utils.StringValue(value)
	}
	return ""
}

// GetVariableAsInt 获取整数类型的变量值
func (c *WorkflowExecutionContext) GetVariableAsInt(name string) int {
	if value, exists := c.GetVariable(name); exists {
		return utils.IntValue(value)
	}
	return 0
}

// GetVariableAsBool 获取布尔类型的变量值
func (c *WorkflowExecutionContext) GetVariableAsBool(name string) bool {
	if value, exists := c.GetVariable(name); exists {
		return utils.BoolValue(value)
	}
	return false
}