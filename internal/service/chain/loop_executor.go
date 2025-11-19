package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/model"
	"go.uber.org/zap"
)

// LoopExecutor 循环执行器
type LoopExecutor struct {
	logger       *zap.Logger
	actionChain  *ActionChain
	maxIterations int
	maxLoopTime   time.Duration
}

// NewLoopExecutor 创建循环执行器
func NewLoopExecutor(logger *zap.Logger, actionChain *ActionChain) *LoopExecutor {
	return &LoopExecutor{
		logger:        logger,
		actionChain:   actionChain,
		maxIterations: 1000, // 最大循环次数限制
		maxLoopTime:   time.Hour * 1, // 最大循环时间限制
	}
}

// ExecuteLoop 执行循环
func (e *LoopExecutor) ExecuteLoop(ctx context.Context, loop *model.Loop, workflowCtx *model.WorkflowContext) error {
	// 验证循环参数
	if loop.MaxIterations > 0 && loop.MaxIterations > e.maxIterations {
		return fmt.Errorf("循环次数超过最大限制: %d > %d", loop.MaxIterations, e.maxIterations)
	}

	e.logger.Info("开始执行循环",
		zap.String("loop_type", string(loop.Type)),
		zap.String("iterator_var", loop.IteratorVar),
		zap.Int("max_iterations", loop.MaxIterations))

	startTime := time.Now()
	iteration := 0

	// 根据循环类型执行
	switch loop.Type {
	case model.LoopTypeForEach:
		return e.executeForEach(ctx, loop, workflowCtx, &iteration, startTime)
	case model.LoopTypeWhile:
		return e.executeWhile(ctx, loop, workflowCtx, &iteration, startTime)
	case model.LoopTypeFor:
		return e.executeFor(ctx, loop, workflowCtx, &iteration, startTime)
	default:
		return fmt.Errorf("不支持的循环类型: %s", loop.Type)
	}
}

// executeForEach 执行foreach循环
func (e *LoopExecutor) executeForEach(ctx context.Context, loop *model.Loop, workflowCtx *model.WorkflowContext, iteration *int, startTime time.Time) error {
	// 获取遍历的数据源
	dataSource, err := e.resolveDataSource(loop.DataSource, workflowCtx)
	if err != nil {
		return fmt.Errorf("解析数据源失败: %w", err)
	}

	items, ok := dataSource.([]interface{})
	if !ok {
		return fmt.Errorf("数据源不是数组类型")
	}

	// 遍历每个元素
	for index, item := range items {
		// 检查循环限制
		if err := e.checkLoopLimits(iteration, startTime); err != nil {
			return err
		}

		// 设置迭代变量
		if loop.IteratorVar != "" {
			workflowCtx.SetVariable(loop.IteratorVar, item)
		}

		// 设置索引变量
		if loop.IndexVar != "" {
			workflowCtx.SetVariable(loop.IndexVar, index)
		}

		e.logger.Debug("执行foreach迭代",
			zap.Int("iteration", *iteration),
			zap.Int("index", index),
			zap.Any("item", item))

		// 执行循环体中的动作
		for _, actionID := range loop.Actions {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := e.actionChain.ExecuteAction(ctx, actionID, workflowCtx); err != nil {
					e.logger.Error("循环中的动作执行失败",
						zap.String("action_id", actionID),
						zap.Int("iteration", *iteration),
						zap.Error(err))
					return fmt.Errorf("循环动作执行失败: %w", err)
				}
			}
		}

		*iteration++
	}

	e.logger.Info("foreach循环完成", zap.Int("total_iterations", *iteration))
	return nil
}

// executeWhile 执行while循环
func (e *LoopExecutor) executeWhile(ctx context.Context, loop *model.Loop, workflowCtx *model.WorkflowContext, iteration *int, startTime time.Time) error {
	conditionEvaluator := NewConditionEvaluator(e.logger)

	for {
		// 检查循环限制
		if err := e.checkLoopLimits(iteration, startTime); err != nil {
			return err
		}

		// 评估循环条件
		shouldContinue, err := conditionEvaluator.EvaluateConditions(ctx, loop.Conditions, loop.ConditionOperator)
		if err != nil {
			return fmt.Errorf("评估循环条件失败: %w", err)
		}

		if !shouldContinue {
			e.logger.Info("while循环条件不满足，退出循环", zap.Int("iteration", *iteration))
			break
		}

		e.logger.Debug("执行while迭代", zap.Int("iteration", *iteration))

		// 执行循环体中的动作
		for _, actionID := range loop.Actions {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := e.actionChain.ExecuteAction(ctx, actionID, workflowCtx); err != nil {
					e.logger.Error("while循环中的动作执行失败",
						zap.String("action_id", actionID),
						zap.Int("iteration", *iteration),
						zap.Error(err))
					return fmt.Errorf("while循环动作执行失败: %w", err)
				}
			}
		}

		*iteration++
	}

	e.logger.Info("while循环完成", zap.Int("total_iterations", *iteration))
	return nil
}

// executeFor 执行for循环（计数循环）
func (e *LoopExecutor) executeFor(ctx context.Context, loop *model.Loop, workflowCtx *model.WorkflowContext, iteration *int, startTime time.Time) error {
	// 解析循环参数
	start := loop.Start
	end := loop.End
	step := loop.Step
	if step == 0 {
		step = 1
	}

	// 从上下文解析变量
	if startValue, exists := workflowCtx.GetVariable("loop_start"); exists {
		if startNum, ok := startValue.(float64); ok {
			start = int(startNum)
		}
	}
	if endValue, exists := workflowCtx.GetVariable("loop_end"); exists {
		if endNum, ok := endValue.(float64); ok {
			end = int(endNum)
		}
	}

	e.logger.Debug("执行for循环",
		zap.Int("start", start),
		zap.Int("end", end),
		zap.Int("step", step))

	for i := start; i < end; i += step {
		// 检查循环限制
		if err := e.checkLoopLimits(iteration, startTime); err != nil {
			return err
		}

		// 设置迭代变量
		if loop.IteratorVar != "" {
			workflowCtx.SetVariable(loop.IteratorVar, i)
		}

		e.logger.Debug("执行for迭代",
			zap.Int("iteration", *iteration),
			zap.Int("value", i))

		// 执行循环体中的动作
		for _, actionID := range loop.Actions {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := e.actionChain.ExecuteAction(ctx, actionID, workflowCtx); err != nil {
					e.logger.Error("for循环中的动作执行失败",
						zap.String("action_id", actionID),
						zap.Int("iteration", *iteration),
						zap.Error(err))
					return fmt.Errorf("for循环动作执行失败: %w", err)
				}
			}
		}

		*iteration++
	}

	e.logger.Info("for循环完成", zap.Int("total_iterations", *iteration))
	return nil
}

// checkLoopLimits 检查循环限制
func (e *LoopExecutor) checkLoopLimits(iteration *int, startTime time.Time) error {
	// 检查最大循环次数
	if *iteration >= e.maxIterations {
		return fmt.Errorf("循环次数超过最大限制: %d", *iteration)
	}

	// 检查最大执行时间
	if time.Since(startTime) > e.maxLoopTime {
		return fmt.Errorf("循环执行时间超过最大限制: %v", time.Since(startTime))
	}

	return nil
}

// resolveDataSource 解析数据源
func (e *LoopExecutor) resolveDataSource(dataSource interface{}, workflowCtx *model.WorkflowContext) (interface{}, error) {
	switch v := dataSource.(type) {
	case string:
		// 检查是否是变量引用
		if len(v) > 2 && v[0] == '$' && v[1] == '{' && v[len(v)-1] == '}' {
			varName := v[2 : len(v)-1]
			if value, exists := workflowCtx.GetVariable(varName); exists {
				return value, nil
			}
			return nil, fmt.Errorf("变量不存在: %s", varName)
		}
		return v, nil
	case []interface{}:
		return v, nil
	default:
		return dataSource, nil
	}
}

// BreakLoop 循环中断信号
type BreakLoop struct {
	Iteration int
	Message   string
}

// ContinueLoop 循环继续信号
type ContinueLoop struct {
	Iteration int
	Message   string
}

// IsLoopBreak 检查是否是循环中断
func IsLoopBreak(err error) bool {
	_, ok := err.(*BreakLoop)
	return ok
}

// IsLoopContinue 检查是否是循环继续
func IsLoopContinue(err error) bool {
	_, ok := err.(*ContinueLoop)
	return ok
}

// NewBreakLoop 创建循环中断
func NewBreakLoop(iteration int, message string) *BreakLoop {
	return &BreakLoop{
		Iteration: iteration,
		Message:   message,
	}
}

// NewContinueLoop 创建循环继续
func NewContinueLoop(iteration int, message string) *ContinueLoop {
	return &ContinueLoop{
		Iteration: iteration,
		Message:   message,
	}
}

func (b *BreakLoop) Error() string {
	return fmt.Sprintf("循环中断于第%d次迭代: %s", b.Iteration, b.Message)
}

func (c *ContinueLoop) Error() string {
	return fmt.Sprintf("循环继续于第%d次迭代: %s", c.Iteration, c.Message)
}