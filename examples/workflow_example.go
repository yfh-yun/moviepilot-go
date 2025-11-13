package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
)

func main() {
	// 获取工作流帮助类实例
	workflowHelper := helper.GetWorkflowHelper()

	// 示例1: 检查工作流分享功能是否开�?	fmt.Println("=== 检查工作流分享功能 ===")
	enabled, message := workflowHelper.CheckWorkflowShareEnabled()
	fmt.Printf("功能开�? %v, 消息: %s\n", enabled, message)

	// 示例2: 获取用户UUID
	fmt.Println("\n=== 获取用户UUID ===")
	uuid := workflowHelper.GetUserUUID()
	fmt.Printf("用户UUID: %s\n", uuid)

	// 示例3: 验证工作�?	fmt.Println("\n=== 验证工作�?===")
	valid, message := workflowHelper.ValidateWorkflow(nil)
	fmt.Printf("验证结果: %v, 消息: %s\n", valid, message)

	// 示例4: 获取分享列表
	fmt.Println("\n=== 获取分享列表 ===")
	shares := workflowHelper.GetShares("", 1, 30)
	fmt.Printf("获取�?%d 个分享\n", len(shares))

	// 示例5: 工作流分享（需要实际的工作流ID�?	fmt.Println("\n=== 工作流分�?===")
	success, message := workflowHelper.WorkflowShare(1, "测试分享", "这是一个测试分�?, "测试用户")
	fmt.Printf("分享结果: %v, 消息: %s\n", success, message)

	// 示例6: 删除分享（需要实际的分享ID�?	fmt.Println("\n=== 删除分享 ===")
	success, message = workflowHelper.ShareDelete(1)
	fmt.Printf("删除结果: %v, 消息: %s\n", success, message)

	// 示例7: 复用分享的工作流（需要实际的分享ID�?	fmt.Println("\n=== 复用分享的工作流 ===")
	success, message = workflowHelper.WorkflowFork(1)
	fmt.Printf("复用结果: %v, 消息: %s\n", success, message)
}
