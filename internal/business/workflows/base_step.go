package workflows

import (
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/workflows/actions"
)

// BaseStep 定义步骤基础实现
type BaseStep struct {
	ID           string         // 步骤ID
	Name         string         // 步骤名称
	Description  string         // 步骤描述
	ActionName   string         // 动作名称
	ActionType   string         // 动作类型
	Input        map[string]any // 输入参数
	Output       map[string]any // 输出参数
	Status       string         // 步骤状态
	Dependencies []string       // 依赖的步骤ID
	Action       actions.Action // 动作实例
	Logger       *zap.Logger    // 日志记录器
	CreatedAt    time.Time      // 创建时间
	UpdatedAt    time.Time      // 更新时间
}

// GetID 获取步骤ID
func (s *BaseStep) GetID() string {
	return s.ID
}

// GetName 获取步骤名称
func (s *BaseStep) GetName() string {
	return s.Name
}

// GetDescription 获取步骤描述
func (s *BaseStep) GetDescription() string {
	return s.Description
}

// GetStatus 获取步骤状态
func (s *BaseStep) GetStatus() string {
	return s.Status
}

// GetActionName 获取步骤对应的动作名称
func (s *BaseStep) GetActionName() string {
	return s.ActionName
}

// GetActionType 获取步骤对应的动作类型
func (s *BaseStep) GetActionType() string {
	return s.ActionType
}

// GetInput 获取步骤输入参数
func (s *BaseStep) GetInput() map[string]any {
	return s.Input
}

// GetOutput 获取步骤输出参数
func (s *BaseStep) GetOutput() map[string]any {
	return s.Output
}

// GetDependencies 获取步骤依赖的步骤ID列表
func (s *BaseStep) GetDependencies() []string {
	return s.Dependencies
}

// Execute 执行步骤
func (s *BaseStep) Execute(ctx StepContext) (*StepResult, error) {
	startTime := time.Now()

	// 更新步骤状态为执行中
	s.Status = "running"
	s.UpdatedAt = time.Now()

	s.Logger.Info("Step execution started", zap.String("step_id", s.ID), zap.String("step_name", s.Name), zap.String("action_name", s.ActionName))

	// 创建动作上下文
	actionCtx := &actions.ActionContext{
		Context:       ctx,
		Logger:        s.Logger,
		WorkflowID:    ctx.WorkflowID,
		ActionName:    s.ActionName,
		ActionType:    s.ActionType,
		Input:         s.Input,
		GlobalContext: ctx.GlobalContext,
		Services:      ctx.Services,
	}

	// 执行动作
	result, err := s.Action.Execute(*actionCtx)

	// 处理执行结果
	if err != nil || !result.Success {
		// 更新步骤状态为执行失败
		s.Status = "failed"
		s.UpdatedAt = time.Now()

		duration := time.Since(startTime)
		errorMsg := result.ErrorMessage
		if err != nil {
			errorMsg = err.Error()
		}

		stepResult := &StepResult{
			StepID:       s.ID,
			StepName:     s.Name,
			ActionName:   s.ActionName,
			ActionType:   s.ActionType,
			Success:      false,
			ErrorMessage: errorMsg,
			Output:       result.Output,
			Duration:     duration,
			Status:       "failed",
			Dependencies: s.Dependencies,
		}

		s.Logger.Error("Step execution failed", zap.String("step_id", s.ID), zap.String("step_name", s.Name), zap.String("error", errorMsg), zap.Duration("duration", duration))

		return stepResult, err
	}

	// 更新步骤状态为执行完成
	s.Status = "completed"
	s.Output = result.Output
	s.UpdatedAt = time.Now()

	// 创建成功的执行结果
	duration := time.Since(startTime)
	stepResult := &StepResult{
		StepID:       s.ID,
		StepName:     s.Name,
		ActionName:   s.ActionName,
		ActionType:   s.ActionType,
		Success:      true,
		ErrorMessage: "",
		Output:       result.Output,
		Duration:     duration,
		Status:       "completed",
		Dependencies: s.Dependencies,
	}

	s.Logger.Info("Step execution completed successfully", zap.String("step_id", s.ID), zap.String("step_name", s.Name), zap.Duration("duration", duration))

	return stepResult, nil
}

// Rollback 回滚步骤
func (s *BaseStep) Rollback() error {
	// 更新步骤状态为回滚中
	s.Status = "rolling_back"
	s.UpdatedAt = time.Now()

	s.Logger.Info("Step rollback started", zap.String("step_id", s.ID), zap.String("step_name", s.Name), zap.String("action_name", s.ActionName))

	// TODO: 实现步骤回滚逻辑
	// 这里需要根据动作类型执行相应的回滚操作

	// 更新步骤状态为已回滚
	s.Status = "rolled_back"
	s.UpdatedAt = time.Now()

	s.Logger.Info("Step rollback completed", zap.String("step_id", s.ID), zap.String("step_name", s.Name))

	return nil
}

// NewStep 创建新的步骤实例
func NewStep(config StepConfig) (Step, error) {
	return &BaseStep{
		ID:           config.ID,
		Name:         config.Name,
		Description:  config.Description,
		ActionName:   config.ActionName,
		ActionType:   config.ActionType,
		Input:        config.Input,
		Output:       make(map[string]any),
		Status:       "pending",
		Dependencies: config.Dependencies,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}
