package workflow

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// TestEngine_ExecuteParallel 测试并行执行工作流功能
func TestEngine_ExecuteParallel(t *testing.T) {
	// 创建工作流引擎
	engine := NewEngine()

	// 创建测试工作流1
	workflow1 := &Workflow{
		ID:          "test-workflow-1",
		Name:        "测试工作流1",
		Description: "测试工作流1",
		Steps: []Step{
			{
				ID:         "step-1-1",
				Name:       "步骤1",
				Type:       StepTypeAction,
				Action:     "note",
				Parameters: map[string]any{"message": "工作流1步骤1执行"},
			},
			{
				ID:         "step-1-2",
				Name:       "步骤2",
				Type:       StepTypeAction,
				Action:     "note",
				Parameters: map[string]any{"message": "工作流1步骤2执行"},
			},
		},
		Variables: map[string]any{"workflow": "1"},
	}

	// 创建测试工作流2
	workflow2 := &Workflow{
		ID:          "test-workflow-2",
		Name:        "测试工作流2",
		Description: "测试工作流2",
		Steps: []Step{
			{
				ID:         "step-2-1",
				Name:       "步骤1",
				Type:       StepTypeAction,
				Action:     "note",
				Parameters: map[string]any{"message": "工作流2步骤1执行"},
			},
			{
				ID:         "step-2-2",
				Name:       "步骤2",
				Type:       StepTypeAction,
				Action:     "note",
				Parameters: map[string]any{"message": "工作流2步骤2执行"},
			},
		},
		Variables: map[string]any{"workflow": "2"},
	}

	// 并行执行两个工作流
	ctx := context.Background()
	workflows := []*Workflow{workflow1, workflow2}

	results, err := engine.ExecuteParallel(ctx, workflows)
	if err != nil {
		t.Fatalf("并行执行工作流失败: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("并行执行结果数量错误，期望2个，实际%v个", len(results))
	}

	logger.GetLogger().Info("并行执行工作流测试完成", zap.Int("workflow_count", len(workflows)))
}

// TestEngine_PauseResumeExecution 测试暂停和恢复工作流执行功能
func TestEngine_PauseResumeExecution(t *testing.T) {
	// 该测试需要更复杂的异步流程设计，暂时注释
	// 实际使用中，暂停恢复功能需要与具体的动作实现配合
	// 由于当前动作执行是模拟的，执行速度太快，无法测试暂停恢复
	t.Skip("暂停恢复功能需要与具体动作实现配合，当前模拟执行速度太快")
}

// TestEngine_RollbackExecution 测试回滚工作流执行功能
func TestEngine_RollbackExecution(t *testing.T) {
	// 该测试需要更复杂的设计，暂时注释
	// 回滚功能需要与具体的动作实现配合
	t.Skip("回滚功能需要与具体动作实现配合")
}
