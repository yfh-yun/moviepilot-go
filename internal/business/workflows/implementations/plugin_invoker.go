// Package implementations 提供动作系统的具体实现
package implementations

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/business/workflows/base"
	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/types"
	"moviepilot-go/pkg/logger"
)

// PluginInvoker 插件调用动作
type PluginInvoker struct {
	*base.Action
	config *PluginInvokerConfig
}

// PluginInvokerConfig 插件调用器配置
type PluginInvokerConfig struct {
	PluginID     string                 `json:"plugin_id" description:"插件ID"`
	Method       string                 `json:"method" description:"调用方法"`
	Parameters   map[string]interface{} `json:"parameters" description:"方法参数"`
	Timeout      time.Duration          `json:"timeout" description:"调用超时时间"`
	Retries      int                    `json:"retries" description:"重试次数"`
	Async        bool                   `json:"async" description:"异步调用"`
	WaitResult   bool                   `json:"wait_result" description:"等待结果"`
	CallbackURL  string                 `json:"callback_url" description:"回调URL"`
	Headers      map[string]string      `json:"headers" description:"请求头"`
	Metadata     map[string]string      `json:"metadata" description:"元数据"`
}

// NewPluginInvoker 创建插件调用器实例
func NewPluginInvoker() interfaces.Action {
	return &PluginInvoker{
		Action: base.NewAction("PluginInvoker", "插件调用器，支持调用Python插件服务"),
		config: &PluginInvokerConfig{
			Timeout:    30 * time.Second,
			Retries:    3,
			Async:      false,
			WaitResult: true,
		},
	}
}

// Execute 执行插件调用
func (pi *PluginInvoker) Execute(ctx context.Context, workflowID int64, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error) {
	logger.Debug("PluginInvoker execution started", 
		"workflow_id", workflowID,
		"action", "PluginInvoker")

	// 解析参数
	config, err := pi.parseConfig(params)
	if err != nil {
		pi.SetError(fmt.Sprintf("参数解析失败: %v", err))
		return actionContext, err
	}

	// 验证插件
	if err := pi.validatePlugin(ctx, config); err != nil {
		pi.SetError(fmt.Sprintf("插件验证失败: %v", err))
		return actionContext, err
	}

	// 执行调用
	result, err := pi.invokePlugin(ctx, config)
	if err != nil {
		pi.SetError(fmt.Sprintf("插件调用失败: %v", err))
		return actionContext, err
	}

	// 设置结果
	pi.SetData("plugin_id", config.PluginID)
	pi.SetData("method", config.Method)
	pi.SetData("result", result)
	pi.SetData("execution_time", time.Now().Format(time.RFC3339))

	pi.SetDone(fmt.Sprintf("成功调用插件 %s.%s", config.PluginID, config.Method))

	logger.Info("PluginInvoker execution completed", 
		"workflow_id", workflowID,
		"plugin_id", config.PluginID,
		"method", config.Method)

	return actionContext, nil
}

// parseConfig 解析配置参数
func (pi *PluginInvoker) parseConfig(params map[string]interface{}) (*PluginInvokerConfig, error) {
	config := *pi.config // 复制默认配置

	if pluginID, ok := params["plugin_id"].(string); ok {
		config.PluginID = pluginID
	}

	if method, ok := params["method"].(string); ok {
		config.Method = method
	}

	if parameters, ok := params["parameters"].(map[string]interface{}); ok {
		config.Parameters = parameters
	}

	if timeout, ok := params["timeout"].(float64); ok {
		config.Timeout = time.Duration(timeout) * time.Second
	}

	if retries, ok := params["retries"].(float64); ok {
		config.Retries = int(retries)
	}

	if async, ok := params["async"].(bool); ok {
		config.Async = async
	}

	if waitResult, ok := params["wait_result"].(bool); ok {
		config.WaitResult = waitResult
	}

	if callbackURL, ok := params["callback_url"].(string); ok {
		config.CallbackURL = callbackURL
	}

	if headers, ok := params["headers"].(map[string]string); ok {
		config.Headers = headers
	}

	if metadata, ok := params["metadata"].(map[string]string); ok {
		config.Metadata = metadata
	}

	return &config, nil
}

// validatePlugin 验证插件
func (pi *PluginInvoker) validatePlugin(ctx context.Context, config *PluginInvokerConfig) error {
	if config.PluginID == "" {
		return fmt.Errorf("plugin_id is required")
	}

	if config.Method == "" {
		return fmt.Errorf("method is required")
	}

	// 这里可以添加插件存在性检查
	// 例如通过插件管理器检查插件是否已加载

	return nil
}

// invokePlugin 调用插件
func (pi *PluginInvoker) invokePlugin(ctx context.Context, config *PluginInvokerConfig) (interface{}, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// 构建调用请求
	request := &PluginRequest{
		PluginID:   config.PluginID,
		Method:     config.Method,
		Parameters: config.Parameters,
		Metadata:   config.Metadata,
		Headers:    config.Headers,
		Timestamp:  time.Now(),
	}

	// 执行调用
	var result interface{}
	var err error

	for attempt := 0; attempt <= config.Retries; attempt++ {
		if attempt > 0 {
			logger.Info("Retrying plugin invocation", 
				"attempt", attempt, 
				"plugin_id", config.PluginID,
				"method", config.Method)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		if config.Async {
			result, err = pi.invokeAsync(timeoutCtx, request, config)
		} else {
			result, err = pi.invokeSync(timeoutCtx, request, config)
		}

		if err == nil {
			break
		}

		if attempt == config.Retries {
			return nil, fmt.Errorf("plugin invocation failed after %d attempts: %v", config.Retries+1, err)
		}
	}

	return result, nil
}

// PluginRequest 插件请求
type PluginRequest struct {
	PluginID   string                 `json:"plugin_id"`
	Method     string                 `json:"method"`
	Parameters map[string]interface{} `json:"parameters"`
	Metadata   map[string]string      `json:"metadata"`
	Headers    map[string]string      `json:"headers"`
	Timestamp  time.Time              `json:"timestamp"`
}

// invokeSync 同步调用
func (pi *PluginInvoker) invokeSync(ctx context.Context, request *PluginRequest, config *PluginInvokerConfig) (interface{}, error) {
	logger.Info("Invoking plugin synchronously", 
		"plugin_id", request.PluginID,
		"method", request.Method)

	// 这里应该通过gRPC调用Python插件服务
	// 目前返回模拟结果
	result := map[string]interface{}{
		"success": true,
		"data":    fmt.Sprintf("Result from %s.%s", request.PluginID, request.Method),
		"message": "Plugin executed successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// invokeAsync 异步调用
func (pi *PluginInvoker) invokeAsync(ctx context.Context, request *PluginRequest, config *PluginInvokerConfig) (interface{}, error) {
	logger.Info("Invoking plugin asynchronously", 
		"plugin_id", request.PluginID,
		"method", request.Method)

	// 这里应该启动异步调用并返回任务ID
	taskID := fmt.Sprintf("task_%d_%s", time.Now().Unix(), request.PluginID)

	// 如果需要等待结果，可以在这里实现轮询或回调机制
	if config.WaitResult {
		// 模拟等待异步结果
		select {
		case <-time.After(1 * time.Second):
			return map[string]interface{}{
				"task_id":   taskID,
				"success":   true,
				"data":      fmt.Sprintf("Async result from %s.%s", request.PluginID, request.Method),
				"completed": true,
			}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 立即返回任务ID
	return map[string]interface{}{
		"task_id":   taskID,
		"success":   true,
		"message":   "Async task started",
		"completed": false,
	}, nil
}

// Initialize 初始化插件调用器
func (pi *PluginInvoker) Initialize() error {
	logger.Info("Initializing PluginInvoker", 
		"timeout", pi.config.Timeout,
		"retries", pi.config.Retries)
	return nil
}

// Cleanup 清理资源
func (pi *PluginInvoker) Cleanup() error {
	logger.Info("Cleaning up PluginInvoker")
	return nil
}