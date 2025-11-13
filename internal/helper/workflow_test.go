package helper

import (
	"testing"
)

func TestWorkflowHelper_GetWorkflowHelper(t *testing.T) {
	// 测试获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	if workflowHelper == nil {
		t.Error("Expected WorkflowHelper instance, got nil")
	}
	
	// 测试单例模式
	workflowHelper2 := GetWorkflowHelper()
	if workflowHelper != workflowHelper2 {
		t.Error("Expected same instance for singleton pattern")
	}
}

func TestWorkflowHelper_CheckWorkflowShareEnabled(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试检查工作流分享功能是否开�?	enabled, message := workflowHelper.CheckWorkflowShareEnabled()
	t.Logf("Enabled: %v, Message: %s", enabled, message)
}

func TestWorkflowHelper_ValidateWorkflow(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试验证空工作流
	valid, message := workflowHelper.ValidateWorkflow(nil)
	if valid {
		t.Error("Expected validation to fail for nil workflow")
	}
	if message != "工作流不存在" {
		t.Errorf("Expected '工作流不存在', got '%s'", message)
	}
}

func TestWorkflowHelper_GetUserUUID(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试获取用户UUID
	uuid := workflowHelper.GetUserUUID()
	if uuid == "" {
		t.Error("Expected non-empty UUID")
	}
	t.Logf("User UUID: %s", uuid)
}

func TestWorkflowHelper_WorkflowShare(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试工作流分享（使用不存在的工作流ID�?	success, message := workflowHelper.WorkflowShare(999999, "测试分享", "测试描述", "测试用户")
	t.Logf("WorkflowShare result - Success: %v, Message: %s", success, message)
}

func TestWorkflowHelper_ShareDelete(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试删除分享（使用不存在的分享ID�?	success, message := workflowHelper.ShareDelete(999999)
	t.Logf("ShareDelete result - Success: %v, Message: %s", success, message)
}

func TestWorkflowHelper_WorkflowFork(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试复用分享的工作流（使用不存在的分享ID�?	success, message := workflowHelper.WorkflowFork(999999)
	t.Logf("WorkflowFork result - Success: %v, Message: %s", success, message)
}

func TestWorkflowHelper_GetShares(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试获取分享列表
	shares := workflowHelper.GetShares("", 1, 30)
	t.Logf("GetShares result - Count: %d", len(shares))
}

func TestWorkflowHelper_AsyncMethods(t *testing.T) {
	// 获取工作流帮助类实例
	workflowHelper := GetWorkflowHelper()
	
	// 测试异步工作流分�?	success, message := workflowHelper.AsyncWorkflowShare(999999, "测试分享", "测试描述", "测试用户")
	t.Logf("AsyncWorkflowShare result - Success: %v, Message: %s", success, message)
	
	// 测试异步删除分享
	success, message = workflowHelper.AsyncShareDelete(999999)
	t.Logf("AsyncShareDelete result - Success: %v, Message: %s", success, message)
	
	// 测试异步复用分享的工作流
	success, message = workflowHelper.AsyncWorkflowFork(999999)
	t.Logf("AsyncWorkflowFork result - Success: %v, Message: %s", success, message)
	
	// 测试异步获取分享列表
	shares := workflowHelper.AsyncGetShares("", 1, 30)
	t.Logf("AsyncGetShares result - Count: %d", len(shares))
}
