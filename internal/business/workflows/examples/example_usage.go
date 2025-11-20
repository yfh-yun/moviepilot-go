// Package examples 提供动作系统使用示例
package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/business/workflows/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/business/workflows/registry"
	"github.com/yfh-yun/moviepilot-go/internal/business/workflows/types"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// ExampleActionUsage 展示动作系统使用示例
func ExampleActionUsage() {
	// 获取默认注册表
	reg := registry.GetDefaultRegistry()

	// 创建动作上下文
	actionContext := &types.ActionContext{
		WorkflowID: 12345,
		Variables:  make(map[string]interface{}),
		Metadata:   make(map[string]string),
		CreatedAt:  time.Now(),
	}

	// 示例1: 使用文件扫描器
	exampleFileScanner(reg, actionContext)

	// 示例2: 使用媒体获取器
	exampleMediaFetcher(reg, actionContext)

	// 示例3: 使用消息发送器
	exampleMessageSender(reg, actionContext)

	// 示例4: 使用插件调用器
	examplePluginInvoker(reg, actionContext)
}

// exampleFileScanner 文件扫描器示例
func exampleFileScanner(reg *registry.ActionRegistry, context *types.ActionContext) {
	logger.Info("=== File Scanner Example ===")

	// 创建文件扫描器动作
	action, err := reg.CreateAction("file_scanner")
	if err != nil {
		logger.Error("Failed to create file scanner", "error", err)
		return
	}

	// 初始化动作
	if err := action.Initialize(); err != nil {
		logger.Error("Failed to initialize file scanner", "error", err)
		return
	}
	defer action.Cleanup()

	// 准备参数
	params := map[string]interface{}{
		"scan_path":        []string{"/tmp/downloads", "/tmp/media"},
		"include_patterns": []string{"*.mp4", "*.mkv"},
		"max_file_size":    1024 * 1024 * 1024, // 1GB
		"enable_hash_check": true,
		"parallel_scans":   2,
	}

	// 执行动作
	ctx := context.Background()
	updatedContext, err := action.Execute(ctx, 12345, params, context)
	if err != nil {
		logger.Error("File scanner execution failed", "error", err)
		return
	}

	// 获取结果
	data := action.GetData()
	if scanResult, ok := data["scan_result"]; ok {
		logger.Info("File scan completed", "result", scanResult)
	}

	logger.Info("Updated context", "variables", updatedContext.Variables)
}

// exampleMediaFetcher 媒体获取器示例
func exampleMediaFetcher(reg *registry.ActionRegistry, context *types.ActionContext) {
	logger.Info("=== Media Fetcher Example ===")

	// 创建媒体获取器动作
	action, err := reg.CreateAction("media_fetcher")
	if err != nil {
		logger.Error("Failed to create media fetcher", "error", err)
		return
	}

	// 初始化动作
	if err := action.Initialize(); err != nil {
		logger.Error("Failed to initialize media fetcher", "error", err)
		return
	}
	defer action.Cleanup()

	// 准备参数
	params := map[string]interface{}{
		"source_type": "movie",
		"sources":     []string{"tmdb", "douban"},
		"keywords":    "复仇者联盟",
		"limit":       10,
		"year":        2019,
		"rating":      7.0,
		"sort_by":     "rating",
	}

	// 执行动作
	ctx := context.Background()
	updatedContext, err := action.Execute(ctx, 12345, params, context)
	if err != nil {
		logger.Error("Media fetcher execution failed", "error", err)
		return
	}

	// 获取结果
	data := action.GetData()
	if medias, ok := data["medias"]; ok {
		logger.Info("Media fetch completed", "count", len(medias.([]*types.MediaInfo)))
	}

	logger.Info("Updated context", "variables", updatedContext.Variables)
}

// exampleMessageSender 消息发送器示例
func exampleMessageSender(reg *registry.ActionRegistry, context *types.ActionContext) {
	logger.Info("=== Message Sender Example ===")

	// 创建消息发送器动作
	action, err := reg.CreateAction("message_sender")
	if err != nil {
		logger.Error("Failed to create message sender", "error", err)
		return
	}

	// 初始化动作
	if err := action.Initialize(); err != nil {
		logger.Error("Failed to initialize message sender", "error", err)
		return
	}
	defer action.Cleanup()

	// 准备参数
	params := map[string]interface{}{
		"channels": []string{"webhook", "email"},
		"title":    "MoviePilot 通知",
		"content":  "您订阅的电影《复仇者联盟》已下载完成",
		"priority": "high",
		"tags":     []string{"download", "movie", "notification"},
		"variables": map[string]string{
			"movie_name": "复仇者联盟",
			"download_time": time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	// 执行动作
	ctx := context.Background()
	updatedContext, err := action.Execute(ctx, 12345, params, context)
	if err != nil {
		logger.Error("Message sender execution failed", "error", err)
		return
	}

	// 获取结果
	data := action.GetData()
	if results, ok := data["results"]; ok {
		logger.Info("Message send completed", "results", results)
	}

	logger.Info("Updated context", "variables", updatedContext.Variables)
}

// examplePluginInvoker 插件调用器示例
func examplePluginInvoker(reg *registry.ActionRegistry, context *types.ActionContext) {
	logger.Info("=== Plugin Invoker Example ===")

	// 创建插件调用器动作
	action, err := reg.CreateAction("plugin_invoker")
	if err != nil {
		logger.Error("Failed to create plugin invoker", "error", err)
		return
	}

	// 初始化动作
	if err := action.Initialize(); err != nil {
		logger.Error("Failed to initialize plugin invoker", "error", err)
		return
	}
	defer action.Cleanup()

	// 准备参数
	params := map[string]interface{}{
		"plugin_id": "site_pter",
		"method":    "search_torrents",
		"parameters": map[string]interface{}{
			"keyword": "复仇者联盟",
			"category": "movie",
			"limit":   20,
		},
		"timeout":     30,
		"retries":     3,
		"async":       false,
		"wait_result": true,
	}

	// 执行动作
	ctx := context.Background()
	updatedContext, err := action.Execute(ctx, 12345, params, context)
	if err != nil {
		logger.Error("Plugin invoker execution failed", "error", err)
		return
	}

	// 获取结果
	data := action.GetData()
	if result, ok := data["result"]; ok {
		logger.Info("Plugin invocation completed", "result", result)
	}

	logger.Info("Updated context", "variables", updatedContext.Variables)
}

// ExampleWorkflow 示例工作流
func ExampleWorkflow() {
	logger.Info("=== Example Workflow ===")

	// 获取注册表
	reg := registry.GetDefaultRegistry()

	// 创建工作流上下文
	context := &types.ActionContext{
		WorkflowID: 67890,
		Variables: map[string]interface{}{
			"user_id":    "user123",
			"movie_name": "复仇者联盟",
			"year":       2019,
		},
		Metadata: map[string]string{
			"workflow_name": "movie_download_workflow",
			"trigger":       "user_request",
		},
		CreatedAt: time.Now(),
	}

	// 步骤1: 媒体获取
	if err := executeWorkflowStep(reg, context, "media_fetcher", map[string]interface{}{
		"keywords": context.Variables["movie_name"],
		"year":     context.Variables["year"],
		"limit":    5,
	}); err != nil {
		logger.Error("Workflow step 1 failed", "error", err)
		return
	}

	// 步骤2: 文件扫描
	if err := executeWorkflowStep(reg, context, "file_scanner", map[string]interface{}{
		"scan_path":        []string{"/downloads"},
		"include_patterns": []string{"*.mp4", "*.mkv"},
	}); err != nil {
		logger.Error("Workflow step 2 failed", "error", err)
		return
	}

	// 步骤3: 消息通知
	if err := executeWorkflowStep(reg, context, "message_sender", map[string]interface{}{
		"channels": []string{"webhook"},
		"title":    "工作流完成",
		"content":  fmt.Sprintf("电影 %s 的处理流程已完成", context.Variables["movie_name"]),
	}); err != nil {
		logger.Error("Workflow step 3 failed", "error", err)
		return
	}

	logger.Info("Workflow completed successfully", "workflow_id", context.WorkflowID)
}

// executeWorkflowStep 执行工作流步骤
func executeWorkflowStep(reg *registry.ActionRegistry, context *types.ActionContext, actionName string, params map[string]interface{}) error {
	// 创建动作
	action, err := reg.CreateAction(actionName)
	if err != nil {
		return fmt.Errorf("failed to create action %s: %v", actionName, err)
	}

	// 初始化动作
	if err := action.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize action %s: %v", actionName, err)
	}
	defer action.Cleanup()

	// 执行动作
	ctx := context.Background()
	updatedContext, err := action.Execute(ctx, context.WorkflowID, params, context)
	if err != nil {
		return fmt.Errorf("failed to execute action %s: %v", actionName, err)
	}

	// 更新上下文
	*context = *updatedContext

	logger.Info("Workflow step completed", 
		"action", actionName, 
		"workflow_id", context.WorkflowID,
		"success", action.IsSuccess())

	return nil
}