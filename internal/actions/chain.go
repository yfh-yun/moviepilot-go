package actions

import "fmt"

// ActionChain 维护顺序执行的 Action 列表。
type ActionChain struct {
	actions     []BaseAction
	stopOnError bool
}

// NewActionChain 创建一个新的 ActionChain。
func NewActionChain(stopOnError bool, actions ...BaseAction) *ActionChain {
	return &ActionChain{
		actions:     actions,
		stopOnError: stopOnError,
	}
}

// Append 追加新的 Action。
func (c *ActionChain) Append(actions ...BaseAction) {
	c.actions = append(c.actions, actions...)
}

// Execute 依次执行链路中的 Actions。
func (c *ActionChain) Execute(workflowID int, params map[string]any, ctx *ActionContext) (*ActionContext, error) {
	if ctx == nil {
		ctx = &ActionContext{}
	}
	ctx.WorkflowID = workflowID
	ctx.Ensure()

	var execErr error

	for _, action := range c.actions {
		if action == nil {
			continue
		}

		var actionParams any
		if params != nil {
			actionParams = params[action.Name()]
		}

		var err error
		ctx, err = action.Execute(workflowID, actionParams, ctx)
		if err != nil {
			execErr = fmt.Errorf("action %s failed: %w", action.Name(), err)
			if c.stopOnError {
				break
			}
		}
	}

	return ctx, execErr
}
