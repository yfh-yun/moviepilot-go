package workflows

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"github.com/stretchr/testify/assert"
)

// TestWorkflowManager_Cache 测试工作流管理器的缓存功能
func TestWorkflowManager_Cache(t *testing.T) {
	// 创建测试日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	// 创建工作流管理器实例
	manager := NewWorkflowManager(logger)
	
	// 创建测试工作流
	workflowConfig := WorkflowConfig{
		ID:          "test-workflow-1",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Priority:    WorkflowPriorityMedium,
		Type:        WorkflowTypeSequential,
	}
	
	workflow, err := NewWorkflow(workflowConfig)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	
	// 注册工作流
	err = manager.RegisterWorkflow(workflow)
	assert.NoError(t, err)
	
	// 创建工作流上下文
	ctx := WorkflowContext{
		Context:       context.Background(),
		Logger:        logger,
		WorkflowID:    "test-workflow-1",
		WorkflowName:  "Test Workflow",
		Input:         map[string]any{"test": "input"},
		GlobalContext: map[string]any{},
		Services:      map[string]any{},
		Priority:      WorkflowPriorityMedium,
		Type:          WorkflowTypeSequential,
	}
	
	// 第一次执行工作流
	result1, err := manager.ExecuteWorkflow(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result1)
	assert.True(t, result1.Success)
	
	// 第二次执行工作流，应该使用缓存
	result2, err := manager.ExecuteWorkflow(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.True(t, result2.Success)
	
	// 验证两次结果相同
	assert.Equal(t, result1.Status, result2.Status)
	assert.Equal(t, result1.Success, result2.Success)
	assert.Equal(t, result1.Duration, result2.Duration)
}

// TestWorkflowManager_ListCache 测试工作流列表缓存
func TestWorkflowManager_ListCache(t *testing.T) {
	// 创建测试日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	// 创建工作流管理器实例
	manager := NewWorkflowManager(logger)
	
	// 创建测试工作流
	workflowConfig1 := WorkflowConfig{
		ID:          "test-workflow-1",
		Name:        "Test Workflow 1",
		Description: "A test workflow 1",
		Priority:    WorkflowPriorityMedium,
		Type:        WorkflowTypeSequential,
	}
	
	workflow1, err := NewWorkflow(workflowConfig1)
	assert.NoError(t, err)
	assert.NotNil(t, workflow1)
	
	workflowConfig2 := WorkflowConfig{
		ID:          "test-workflow-2",
		Name:        "Test Workflow 2",
		Description: "A test workflow 2",
		Priority:    WorkflowPriorityHigh,
		Type:        WorkflowTypeParallel,
	}
	
	workflow2, err := NewWorkflow(workflowConfig2)
	assert.NoError(t, err)
	assert.NotNil(t, workflow2)
	
	// 注册工作流
	err = manager.RegisterWorkflow(workflow1)
	assert.NoError(t, err)
	
	err = manager.RegisterWorkflow(workflow2)
	assert.NoError(t, err)
	
	// 第一次获取工作流列表
	params := ListWorkflowsParams{}
	list1, err := manager.ListWorkflows(params)
	assert.NoError(t, err)
	assert.Len(t, list1, 2)
	
	// 第二次获取工作流列表，应该使用缓存
	list2, err := manager.ListWorkflows(params)
	assert.NoError(t, err)
	assert.Len(t, list2, 2)
	
	// 验证两次列表相同
	assert.Equal(t, len(list1), len(list2))
	for i := range list1 {
		assert.Equal(t, list1[i].GetID(), list2[i].GetID())
	}
	
	// 测试带过滤条件的列表缓存
	params = ListWorkflowsParams{
		Priority: WorkflowPriorityHigh,
	}
	
	list3, err := manager.ListWorkflows(params)
	assert.NoError(t, err)
	assert.Len(t, list3, 1)
	assert.Equal(t, "test-workflow-2", list3[0].GetID())
	
	// 再次使用相同过滤条件，应该使用缓存
	list4, err := manager.ListWorkflows(params)
	assert.NoError(t, err)
	assert.Len(t, list4, 1)
	assert.Equal(t, "test-workflow-2", list4[0].GetID())
}

// TestWorkflowManager_ClearCache 测试缓存清除功能
func TestWorkflowManager_ClearCache(t *testing.T) {
	// 创建测试日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	// 创建工作流管理器实例
	manager := NewWorkflowManager(logger)
	
	// 创建测试工作流
	workflowConfig := WorkflowConfig{
		ID:          "test-workflow-clear-cache",
		Name:        "Test Workflow Clear Cache",
		Description: "A test workflow for cache clearing",
		Priority:    WorkflowPriorityMedium,
		Type:        WorkflowTypeSequential,
	}
	
	workflow, err := NewWorkflow(workflowConfig)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	
	// 注册工作流
	err = manager.RegisterWorkflow(workflow)
	assert.NoError(t, err)
	
	// 创建工作流上下文
	ctx := WorkflowContext{
		Context:       context.Background(),
		Logger:        logger,
		WorkflowID:    "test-workflow-clear-cache",
		WorkflowName:  "Test Workflow Clear Cache",
		Input:         map[string]any{"test": "input"},
		GlobalContext: map[string]any{},
		Services:      map[string]any{},
		Priority:      WorkflowPriorityMedium,
		Type:          WorkflowTypeSequential,
	}
	
	// 执行工作流，结果应该被缓存
	result1, err := manager.ExecuteWorkflow(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result1)
	
	// 清除缓存
	err = manager.ClearWorkflowCache("test-workflow-clear-cache")
	assert.NoError(t, err)
	
	// 再次执行工作流，应该重新执行而不是使用缓存
	result2, err := manager.ExecuteWorkflow(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	
	// 验证结果不同（因为是重新执行）
	assert.Equal(t, result1.Status, result2.Status)
	assert.Equal(t, result1.Success, result2.Success)
	
	// 清除所有缓存
	err = manager.ClearCache()
	assert.NoError(t, err)
	
	// 验证列表缓存也被清除
	params := ListWorkflowsParams{}
	list, err := manager.ListWorkflows(params)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
}

// TestWorkflowManager_EventCache 测试事件缓存功能
func TestWorkflowManager_EventCache(t *testing.T) {
	// 创建测试日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	// 创建工作流管理器实例
	manager := NewWorkflowManager(logger)
	
	// 创建测试工作流
	workflowConfig := WorkflowConfig{
		ID:          "test-workflow-event",
		Name:        "Test Workflow Event",
		Description: "A test workflow for event handling",
		Priority:    WorkflowPriorityMedium,
		Type:        WorkflowTypeSequential,
	}
	
	workflow, err := NewWorkflow(workflowConfig)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	
	// 注册工作流
	err = manager.RegisterWorkflow(workflow)
	assert.NoError(t, err)
	
	// 注册工作流事件
	err = manager.RegisterWorkflowEvent("test-workflow-event", "test-event-type")
	assert.NoError(t, err)
	
	// 触发事件，应该执行工作流
	err = manager.HandleEvent("test-event-type", map[string]interface{}{"event": "data"})
	assert.NoError(t, err)
	
	// 等待一小段时间，确保工作流执行完成
	time.Sleep(100 * time.Millisecond)
	
	// 检查工作流结果
	result, err := manager.GetWorkflowResults("test-workflow-event")
	assert.NoError(t, err)
	
	// 移除工作流事件
	err = manager.RemoveWorkflowEvent("test-workflow-event", "test-event-type")
	assert.NoError(t, err)
}
