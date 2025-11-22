package workflow

import (
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"moviepilot-go/internal/actions"
)

const actionParamsNamespace = "action_params"

// ActionTask 将 actions.BaseAction 适配为 workflow.Task。
type ActionTask struct {
	*BasicTask
	action        actions.BaseAction
	paramsKey     string
	sharedContext *actions.ActionContext
}

// NewActionTask 创建一个新的 ActionTask。
func NewActionTask(id string, action actions.BaseAction, logger *zap.Logger) (*ActionTask, error) {
	if action == nil {
		return nil, fmt.Errorf("action is required")
	}

	task := &ActionTask{
		BasicTask: NewBasicTask(id, action.Name(), nil, logger),
		action:    action,
		paramsKey: action.Name(),
	}
	task.taskFunc = task.run

	return task, nil
}

// WithParamsKey 设置 Action 参数读取的 key。
func (t *ActionTask) WithParamsKey(key string) *ActionTask {
	if key != "" {
		t.paramsKey = key
	}
	return t
}

// WithSharedContext 设置 ActionContext（可让多个 Action 共用）。
func (t *ActionTask) WithSharedContext(ctx *actions.ActionContext) *ActionTask {
	if ctx != nil {
		t.sharedContext = ctx
	}
	return t
}

func (t *ActionTask) run(ctx *WorkflowContext) (*TaskResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("workflow context is nil")
	}

	actionCtx := t.ensureActionContext()
	params := t.extractParams(ctx)

	workflowID := parseWorkflowID(ctx.Workflow)
	resultCtx, err := t.action.Execute(workflowID, params, actionCtx)
	if err != nil {
		return nil, err
	}

	t.sharedContext = resultCtx
	t.writeBackToWorkflow(ctx, resultCtx)

	output := map[string]interface{}{
		"action":          t.action.Name(),
		"files_count":     lenSafeFileItems(resultCtx),
		"medias_count":    lenSafeMedias(resultCtx),
		"downloads_count": lenSafeDownloads(resultCtx),
		"transfers_count": lenSafeTransfers(resultCtx),
	}

	return &TaskResult{
		Status: TaskCompleted,
		Output: output,
	}, nil
}

func (t *ActionTask) ensureActionContext() *actions.ActionContext {
	if t.sharedContext == nil {
		t.sharedContext = &actions.ActionContext{}
	}
	t.sharedContext.Ensure()
	return t.sharedContext
}

func (t *ActionTask) extractParams(ctx *WorkflowContext) any {
	if ctx == nil || ctx.Variables == nil {
		return nil
	}

	if t.paramsKey != "" {
		if val, ok := ctx.Variables[t.paramsKey]; ok {
			return val
		}
	}

	if raw, ok := ctx.Variables[actionParamsNamespace]; ok {
		if paramMap, ok := raw.(map[string]any); ok {
			if v, exists := paramMap[t.paramsKey]; exists {
				return v
			}
		}
	}

	return nil
}

func (t *ActionTask) writeBackToWorkflow(ctx *WorkflowContext, actionCtx *actions.ActionContext) {
	if ctx == nil || actionCtx == nil {
		return
	}

	if ctx.Variables == nil {
		ctx.Variables = make(map[string]interface{})
	}

	ctx.Variables["action_context"] = actionCtx
	ctx.Variables[fmt.Sprintf("%s_files", t.ID())] = actionCtx.Files
	ctx.Variables[fmt.Sprintf("%s_medias", t.ID())] = actionCtx.Medias
	ctx.Variables[fmt.Sprintf("%s_downloads", t.ID())] = actionCtx.Downloads
	ctx.Variables[fmt.Sprintf("%s_transfers", t.ID())] = actionCtx.Transfers

	if actionCtx.WorkflowMetadata != nil {
		for k, v := range actionCtx.WorkflowMetadata {
			ctx.Variables[k] = v
		}
	}
}

func parseWorkflowID(workflow *Workflow) int {
	if workflow == nil {
		return 0
	}
	if id, err := strconv.Atoi(workflow.ID); err == nil {
		return id
	}
	return 0
}

func lenSafeFileItems(ctx *actions.ActionContext) int {
	if ctx == nil {
		return 0
	}
	return len(ctx.Files)
}

func lenSafeMedias(ctx *actions.ActionContext) int {
	if ctx == nil {
		return 0
	}
	return len(ctx.Medias)
}

func lenSafeDownloads(ctx *actions.ActionContext) int {
	if ctx == nil {
		return 0
	}
	return len(ctx.Downloads)
}

func lenSafeTransfers(ctx *actions.ActionContext) int {
	if ctx == nil {
		return 0
	}
	return len(ctx.Transfers)
}
