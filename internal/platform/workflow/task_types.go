package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
)

// BasicTask 基础任务实现
type BasicTask struct {
	id            string
	name          string
	description   string
	taskFunc      TaskFunc
	status        TaskStatus
	dependencies  []string
	retryStrategy RetryStrategy
	logger        logger.Logger
	metadata      map[string]interface{}
}

// TaskFunc 任务函数类型
type TaskFunc func(ctx *WorkflowContext) (*TaskResult, error)

// NewBasicTask 创建基础任务
func NewBasicTask(id, name string, taskFunc TaskFunc, logger logger.Logger) *BasicTask {
	return &BasicTask{
		id:           id,
		name:         name,
		taskFunc:     taskFunc,
		status:       TaskPending,
		dependencies: []string{},
		metadata:     make(map[string]interface{}),
		logger:       logger,
	}
}

// ID 获取任务ID
func (t *BasicTask) ID() string {
	return t.id
}

// Name 获取任务名称
func (t *BasicTask) Name() string {
	return t.name
}

// Description 获取任务描述
func (t *BasicTask) Description() string {
	return t.description
}

// Status 获取任务状态
func (t *BasicTask) Status() TaskStatus {
	return t.status
}

// SetStatus 设置任务状态
func (t *BasicTask) SetStatus(status TaskStatus) {
	t.status = status
}

// Dependencies 获取任务依赖
func (t *BasicTask) Dependencies() []string {
	return t.dependencies
}

// AddDependency 添加任务依赖
func (t *BasicTask) AddDependency(taskID string) {
	t.dependencies = append(t.dependencies, taskID)
}

// SetDescription 设置任务描述
func (t *BasicTask) SetDescription(description string) {
	t.description = description
}

// SetRetryStrategy 设置重试策略
func (t *BasicTask) SetRetryStrategy(strategy RetryStrategy) {
	t.retryStrategy = strategy
}

// GetRetryStrategy 获取重试策略
func (t *BasicTask) GetRetryStrategy() RetryStrategy {
	return t.retryStrategy
}

// SetMetadata 设置元数据
func (t *BasicTask) SetMetadata(key string, value interface{}) {
	t.metadata[key] = value
}

// GetMetadata 获取元数据
func (t *BasicTask) GetMetadata(key string) (interface{}, bool) {
	value, exists := t.metadata[key]
	return value, exists
}

// CanRun 检查任务是否可以运行
func (t *BasicTask) CanRun(ctx *WorkflowContext) bool {
	// 检查依赖任务是否都已完成
	for _, depID := range t.dependencies {
		result, exists := ctx.Results[depID]
		if !exists || result.Status != TaskCompleted {
			return false
		}
	}
	return true
}

// Run 执行任务
func (t *BasicTask) Run(ctx *WorkflowContext) (*TaskResult, error) {
	if t.taskFunc == nil {
		return nil, fmt.Errorf("task function not set for task %s", t.id)
	}

	return t.taskFunc(ctx)
}

// ConditionTask 条件任务
type ConditionTask struct {
	BasicTask
	condition     ConditionFunc
	thenTask      Task
	elseTask      Task
	conditionMet  bool
}

// ConditionFunc 条件函数类型
type ConditionFunc func(ctx *WorkflowContext) bool

// NewConditionTask 创建条件任务
func NewConditionTask(id, name string, condition ConditionFunc, logger logger.Logger) *ConditionTask {
	task := &ConditionTask{
		BasicTask: *NewBasicTask(id, name, nil, logger),
		condition: condition,
	}

	// 覆盖Run方法
	task.taskFunc = task.conditionRun

	return task
}

// SetThenTask 设置条件为真时执行的任务
func (t *ConditionTask) SetThenTask(task Task) {
	t.thenTask = task
	t.AddDependency(task.ID()) // 添加依赖
}

// SetElseTask 设置条件为假时执行的任务
func (t *ConditionTask) SetElseTask(task Task) {
	t.elseTask = task
	t.AddDependency(task.ID()) // 添加依赖
}

// conditionRun 执行条件任务
func (t *ConditionTask) conditionRun(ctx *WorkflowContext) (*TaskResult, error) {
	// 计算条件
	t.conditionMet = t.condition(ctx)
	t.logger.Info("Condition evaluated", "task_id", t.id, "condition_met", t.conditionMet)

	// 根据条件执行对应的任务
	var result *TaskResult
	var err error

	if t.conditionMet && t.thenTask != nil {
		t.logger.Info("Executing then branch", "task_id", t.id)
		result, err = t.thenTask.Run(ctx)
	} else if !t.conditionMet && t.elseTask != nil {
		t.logger.Info("Executing else branch", "task_id", t.id)
		result, err = t.elseTask.Run(ctx)
	} else {
		// 如果条件为真但没有thenTask，或者条件为假但没有elseTask
		t.logger.Info("No branch to execute", "task_id", t.id)
		result = &TaskResult{
			Status: TaskCompleted,
			Output: map[string]interface{}{"condition_met": t.conditionMet},
		}
	}

	if err != nil {
		return nil, err
	}

	// 添加条件结果到输出
	if result.Output == nil {
		result.Output = make(map[string]interface{})
	}
	result.Output["condition_met"] = t.conditionMet

	return result, nil
}

// LoopTask 循环任务
type LoopTask struct {
	BasicTask
	loopFunc      LoopFunc
	taskToRepeat  Task
	maxIterations int
	currentIter   int
}

// LoopFunc 循环条件函数类型
type LoopFunc func(ctx *WorkflowContext, iteration int) bool

// NewLoopTask 创建循环任务
func NewLoopTask(id, name string, loopFunc LoopFunc, maxIterations int, logger logger.Logger) *LoopTask {
	task := &LoopTask{
		BasicTask:     *NewBasicTask(id, name, nil, logger),
		loopFunc:      loopFunc,
		maxIterations: maxIterations,
		currentIter:   0,
	}

	// 覆盖Run方法
	task.taskFunc = task.loopRun

	return task
}

// SetTaskToRepeat 设置要重复执行的任务
func (t *LoopTask) SetTaskToRepeat(task Task) {
	t.taskToRepeat = task
}

// loopRun 执行循环任务
func (t *LoopTask) loopRun(ctx *WorkflowContext) (*TaskResult, error) {
	if t.taskToRepeat == nil {
		return nil, fmt.Errorf("task to repeat not set for loop task %s", t.id)
	}

	results := make([]*TaskResult, 0)
	t.currentIter = 0

	// 执行循环
	for {
		// 检查循环条件
		if !t.loopFunc(ctx, t.currentIter) || t.currentIter >= t.maxIterations {
			t.logger.Info("Loop condition exited", "task_id", t.id, "iterations", t.currentIter)
			break
		}

		// 重置任务状态
		t.taskToRepeat.SetStatus(TaskPending)

		// 执行任务
		result, err := t.taskToRepeat.Run(ctx)
		if err != nil {
			return nil, fmt.Errorf("loop iteration %d failed: %w", t.currentIter, err)
		}

		// 记录结果
		results = append(results, result)
		t.currentIter++

		// 如果任务失败，停止循环
		if result.Status != TaskCompleted {
			t.logger.Warn("Loop task iteration failed", "task_id", t.id, "iteration", t.currentIter)
			break
		}
	}

	// 构建最终结果
	return &TaskResult{
		Status:   TaskCompleted,
		Output:   map[string]interface{}{"iterations": t.currentIter, "results": results},
		Metadata: map[string]interface{}{"loop_results": results},
	}, nil
}

// ParallelTask 并行任务
type ParallelTask struct {
	BasicTask
	subTasks      []Task
	concurrent    int
	waitForAll    bool
}

// NewParallelTask 创建并行任务
func NewParallelTask(id, name string, concurrent int, waitForAll bool, logger logger.Logger) *ParallelTask {
	task := &ParallelTask{
		BasicTask: *NewBasicTask(id, name, nil, logger),
		subTasks:  make([]Task, 0),
		concurrent: concurrent,
		waitForAll: waitForAll,
	}

	if task.concurrent <= 0 {
		task.concurrent = 5 // 默认并发数
	}

	// 覆盖Run方法
	task.taskFunc = task.parallelRun

	return task
}

// AddTask 添加并行任务
func (t *ParallelTask) AddTask(task Task) {
	t.subTasks = append(t.subTasks, task)
}

// parallelRun 执行并行任务
func (t *ParallelTask) parallelRun(ctx *WorkflowContext) (*TaskResult, error) {
	if len(t.subTasks) == 0 {
		return &TaskResult{Status: TaskCompleted, Output: map[string]interface{}{"tasks_count": 0}},
		nil
	}

	// 创建通道来控制并发
	taskChan := make(chan Task, t.concurrent)
	resultChan := make(chan struct {
		task   Task
		result *TaskResult
		err    error
	}, len(t.subTasks))

	// 启动工作协程
	for i := 0; i < t.concurrent; i++ {
		go func() {
			for task := range taskChan {
				// 重置任务状态
				task.SetStatus(TaskPending)

				// 执行任务
				result, err := task.Run(ctx)

				// 发送结果
				resultChan <- struct {
					task   Task
					result *TaskResult
					err    error
				}{task, result, err}
			}
		}()
	}

	// 提交所有任务
	for _, task := range t.subTasks {
		taskChan <- task
	}
	close(taskChan)

	// 收集结果
	results := make(map[string]*TaskResult)
	var firstErr error
	successfulCount := 0

	for i := 0; i < len(t.subTasks); i++ {
		select {
		case result := <-resultChan:
			if result.err != nil {
				if firstErr == nil {
					firstErr = result.err
				}
				t.logger.Error("Parallel task execution failed", 
					"parent_id", t.id, 
					"task_id", result.task.ID(), 
					"error", result.err.Error())
			} else if result.result != nil {
				results[result.task.ID()] = result.result
				if result.result.Status == TaskCompleted {
					successfulCount++
				}
			}
		}
	}

	// 检查是否所有任务都成功
	totalSuccess := successfulCount == len(t.subTasks)

	if firstErr != nil && t.waitForAll {
		return nil, firstErr
	}

	return &TaskResult{
		Status: TaskCompleted,
		Output: map[string]interface{}{
			"tasks_count":       len(t.subTasks),
			"successful_count":  successfulCount,
			"all_successful":    totalSuccess,
			"individual_results": results,
		},
	}, nil
}

// DelayTask 延迟任务
type DelayTask struct {
	BasicTask
	delayDuration time.Duration
}

// NewDelayTask 创建延迟任务
func NewDelayTask(id, name string, delay time.Duration, logger logger.Logger) *DelayTask {
	task := &DelayTask{
		BasicTask:     *NewBasicTask(id, name, nil, logger),
		delayDuration: delay,
	}

	// 覆盖Run方法
	task.taskFunc = task.delayRun

	return task
}

// delayRun 执行延迟任务
func (t *DelayTask) delayRun(ctx *WorkflowContext) (*TaskResult, error) {
	t.logger.Info("Executing delay task", "task_id", t.id, "delay", t.delayDuration)

	// 创建一个定时器
	timer := time.NewTimer(t.delayDuration)
	defer timer.Stop()

	// 等待延迟完成或上下文取消
	select {
	case <-timer.C:
		// 延迟完成
		t.logger.Info("Delay completed", "task_id", t.id)
		return &TaskResult{
			Status: TaskCompleted,
			Output: map[string]interface{}{"delay_completed": true},
		}, nil
	case <-ctx.Context().Done():
		// 上下文取消
		return &TaskResult{
			Status: TaskFailed,
			Error:  ctx.Context().Err().Error(),
		}, ctx.Context().Err()
	}
}

// TransformTask 转换任务，用于转换数据
type TransformTask struct {
	BasicTask
	transformFunc TransformFunc
	sourceKey     string
	targetKey     string
}

// TransformFunc 转换函数类型
type TransformFunc func(value interface{}) (interface{}, error)

// NewTransformTask 创建转换任务
func NewTransformTask(id, name, sourceKey, targetKey string, transformFunc TransformFunc, logger logger.Logger) *TransformTask {
	task := &TransformTask{
		BasicTask:     *NewBasicTask(id, name, nil, logger),
		transformFunc: transformFunc,
		sourceKey:     sourceKey,
		targetKey:     targetKey,
	}

	// 覆盖Run方法
	task.taskFunc = task.transformRun

	return task
}

// transformRun 执行转换任务
func (t *TransformTask) transformRun(ctx *WorkflowContext) (*TaskResult, error) {
	// 获取源数据
	sourceValue, exists := ctx.Variables[t.sourceKey]
	if !exists {
		t.logger.Warn("Source key not found", "task_id", t.id, "source_key", t.sourceKey)
		return &TaskResult{
			Status: TaskSkipped,
			Output: map[string]interface{}{"skipped": true, "reason": "source_key_not_found"},
		}, nil
	}

	// 转换数据
	transformedValue, err := t.transformFunc(sourceValue)
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	// 存储转换后的数据
	ctx.Variables[t.targetKey] = transformedValue

	t.logger.Info("Data transformation completed", 
		"task_id", t.id, 
		"source_key", t.sourceKey, 
		"target_key", t.targetKey)

	return &TaskResult{
		Status: TaskCompleted,
		Output: map[string]interface{}{
			"source_key":      t.sourceKey,
			"target_key":      t.targetKey,
			"transformed_value": transformedValue,
		},
	}, nil
}

// ErrorHandlingTask 错误处理任务
type ErrorHandlingTask struct {
	BasicTask
	taskToMonitor Task
	errorHandler ErrorHandlerFunc
}

// ErrorHandlerFunc 错误处理函数类型
type ErrorHandlerFunc func(ctx *WorkflowContext, err error) *TaskResult

// NewErrorHandlingTask 创建错误处理任务
func NewErrorHandlingTask(id, name string, taskToMonitor Task, errorHandler ErrorHandlerFunc, logger logger.Logger) *ErrorHandlingTask {
	task := &ErrorHandlingTask{
		BasicTask:     *NewBasicTask(id, name, nil, logger),
		taskToMonitor: taskToMonitor,
		errorHandler: errorHandler,
	}

	// 覆盖Run方法
	task.taskFunc = task.errorHandlingRun

	return task
}

// errorHandlingRun 执行错误处理任务
func (t *ErrorHandlingTask) errorHandlingRun(ctx *WorkflowContext) (*TaskResult, error) {
	// 重置任务状态
	t.taskToMonitor.SetStatus(TaskPending)

	// 执行监控的任务
	result, err := t.taskToMonitor.Run(ctx)

	if err != nil {
		// 错误发生，执行错误处理
		t.logger.Error("Task execution failed, handling error", 
			"task_id", t.id, 
			"monitored_task", t.taskToMonitor.ID(), 
			"error", err.Error())

		// 执行错误处理函数
		handledResult := t.errorHandler(ctx, err)

		// 如果错误处理成功，返回成功结果
		if handledResult.Status == TaskCompleted {
			return &TaskResult{
				Status:    TaskCompleted,
				Output:    handledResult.Output,
				Metadata:  map[string]interface{}{"error_handled": true, "original_error": err.Error()},
			}, nil
		}

		// 错误处理失败，返回原始错误
		return nil, err
	}

	// 任务执行成功
	return result, nil
}

// LogTask 日志任务，用于记录日志
type LogTask struct {
	BasicTask
	logLevel      LogLevel
	logMessage    string
	logVariables  map[string]interface{}
}

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// NewLogTask 创建日志任务
func NewLogTask(id, name, message string, level LogLevel, logger logger.Logger) *LogTask {
	task := &LogTask{
		BasicTask:    *NewBasicTask(id, name, nil, logger),
		logLevel:     level,
		logMessage:   message,
		logVariables: make(map[string]interface{}),
	}

	// 覆盖Run方法
	task.taskFunc = task.logRun

	return task
}

// AddVariable 添加日志变量
func (t *LogTask) AddVariable(key string, value interface{}) {
	t.logVariables[key] = value
}

// logRun 执行日志任务
func (t *LogTask) logRun(ctx *WorkflowContext) (*TaskResult, error) {
	// 合并上下文变量和日志变量
	logFields := make(map[string]interface{})

	// 添加预定义变量
	for k, v := range t.logVariables {
		logFields[k] = v
	}

	// 添加任务信息
	logFields["task_id"] = t.id
	logFields["task_name"] = t.name

	// 解析消息中的变量占位符
	message := t.logMessage
	for key, value := range ctx.Variables {
		placeholder := "${" + key + "}"
		if strings.Contains(message, placeholder) {
			message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// 根据日志级别记录日志
	switch t.logLevel {
	case LogLevelDebug:
		t.logger.Debug(message, logFields)
	case LogLevelInfo:
		t.logger.Info(message, logFields)
	case LogLevelWarn:
		t.logger.Warn(message, logFields)
	case LogLevelError:
		t.logger.Error(message, logFields)
	default:
		t.logger.Info(message, logFields)
	}

	return &TaskResult{
		Status: TaskCompleted,
		Output: map[string]interface{}{"logged": true, "level": t.logLevel, "message": message},
	}, nil
}