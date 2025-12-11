package actions

import (
	"fmt"
	"sync"
)

// ActionChain 维护顺序执行的 Action 列表

type ActionChain struct {
	actions     []Action
	stopOnError bool
	mutex       sync.RWMutex
}

// NewActionChain 创建一个新的 ActionChain
func NewActionChain(stopOnError bool, actions ...Action) *ActionChain {
	return &ActionChain{
		actions:     actions,
		stopOnError: stopOnError,
	}
}

// Append 追加新的 Action
func (c *ActionChain) Append(actions ...Action) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.actions = append(c.actions, actions...)
}

// Execute 依次执行链路中的 Actions
func (c *ActionChain) Execute(ctx ActionContext) (*ActionResult, error) {
	c.mutex.RLock()
	actions := make([]Action, len(c.actions))
	copy(actions, c.actions)
	c.mutex.RUnlock()

	var finalResult *ActionResult
	var execErr error

	for _, action := range actions {
		if action == nil {
			continue
		}

		// 初始化动作
		if !action.IsInitialized() {
			if err := action.Initialize(ctx); err != nil {
				execErr = fmt.Errorf("action %s initialization failed: %w", action.GetName(), err)
				if c.stopOnError {
					break
				}
				continue
			}
		}

		// 执行动作
		result, err := action.Execute(ctx)
		finalResult = result

		if err != nil {
			execErr = fmt.Errorf("action %s execution failed: %w", action.GetName(), err)
			if c.stopOnError {
				break
			}
		} else if !result.Success {
			execErr = fmt.Errorf("action %s execution failed: %s", action.GetName(), result.ErrorMessage)
			if c.stopOnError {
				break
			}
		}

		// 将当前动作的输出作为下一个动作的输入
		if result.Success && result.Output != nil {
			// 合并输出到上下文输入
			if ctx.Input == nil {
				ctx.Input = make(map[string]any)
			}
			for k, v := range result.Output {
				ctx.Input[k] = v
			}
		}
	}

	return finalResult, execErr
}

// GetActions 获取链中的所有动作
func (c *ActionChain) GetActions() []Action {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	result := make([]Action, len(c.actions))
	copy(result, c.actions)
	return result
}

// Clear 清空链中的所有动作
func (c *ActionChain) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.actions = nil
}
