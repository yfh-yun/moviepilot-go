package workflows

import (
	"context"
	"testing"
	"time"

	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/registry"
	"moviepilot-go/internal/business/workflows/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestActionRegistry 测试动作注册表
func TestActionRegistry(t *testing.T) {
	reg := registry.NewActionRegistry()

	// 测试注册动作
	err := reg.Register("test_action", func() interfaces.Action {
		return &TestAction{}
	})
	require.NoError(t, err)

	// 测试重复注册
	err = reg.Register("test_action", func() interfaces.Action {
		return &TestAction{}
	})
	assert.Error(t, err)

	// 测试获取动作工厂
	factory, err := reg.GetFactory("test_action")
	require.NoError(t, err)
	assert.NotNil(t, factory)

	// 测试获取不存在的动作
	_, err = reg.GetFactory("nonexistent")
	assert.Error(t, err)

	// 测试创建动作
	action, err := reg.CreateAction("test_action")
	require.NoError(t, err)
	assert.NotNil(t, action)

	// 测试列出动作
	actions := reg.ListActions()
	assert.Contains(t, actions, "test_action")

	// 测试注销动作
	err = reg.Unregister("test_action")
	require.NoError(t, err)

	// 测试注销不存在的动作
	err = reg.Unregister("test_action")
	assert.Error(t, err)
}

// TestDefaultRegistry 测试默认注册表
func TestDefaultRegistry(t *testing.T) {
	reg := registry.GetDefaultRegistry()
	assert.NotNil(t, reg)

	// 检查内置动作是否已注册
	actions := reg.ListActions()
	expectedActions := []string{
		"download", "scan", "file_scanner", "media_fetcher",
		"message_sender", "plugin_invoker", "rss_parser",
		"subscribe_manager", "transfer_manager", "workflow_cache",
	}

	for _, expected := range expectedActions {
		assert.Contains(t, actions, expected, "Expected action %s to be registered", expected)
	}
}

// TestActionExecution 测试动作执行
func TestActionExecution(t *testing.T) {
	reg := registry.GetDefaultRegistry()

	// 测试文件扫描器
	t.Run("FileScanner", func(t *testing.T) {
		action, err := reg.CreateAction("file_scanner")
		require.NoError(t, err)

		err = action.Initialize()
		require.NoError(t, err)

		// 准备测试参数
		params := map[string]interface{}{
			"scan_path":        []string{"/tmp"},
			"include_patterns": []string{"*.log"},
			"max_file_size":    1024,
			"enable_hash_check": false,
		}

		// 创建动作上下文
		actionContext := &types.ActionContext{
			WorkflowID: 12345,
			Variables:  make(map[string]interface{}),
			Metadata:   make(map[string]string),
			CreatedAt:  time.Now(),
		}

		// 执行动作
		ctx := context.Background()
		updatedContext, err := action.Execute(ctx, 12345, params, actionContext)
		require.NoError(t, err)
		assert.NotNil(t, updatedContext)

		// 检查执行结果
		assert.True(t, action.IsDone())
		assert.True(t, action.IsSuccess())

		// 清理
		err = action.Cleanup()
		require.NoError(t, err)
	})

	// 测试媒体获取器
	t.Run("MediaFetcher", func(t *testing.T) {
		action, err := reg.CreateAction("media_fetcher")
		require.NoError(t, err)

		err = action.Initialize()
		require.NoError(t, err)

		// 准备测试参数
		params := map[string]interface{}{
			"keywords": "test",
			"limit":    5,
			"sources":  []string{"tmdb"},
		}

		// 创建动作上下文
		actionContext := &types.ActionContext{
			WorkflowID: 12346,
			Variables:  make(map[string]interface{}),
			Metadata:   make(map[string]string),
			CreatedAt:  time.Now(),
		}

		// 执行动作
		ctx := context.Background()
		updatedContext, err := action.Execute(ctx, 12346, params, actionContext)
		require.NoError(t, err)
		assert.NotNil(t, updatedContext)

		// 检查执行结果
		assert.True(t, action.IsDone())
		assert.True(t, action.IsSuccess())

		// 清理
		err = action.Cleanup()
		require.NoError(t, err)
	})

	// 测试消息发送器
	t.Run("MessageSender", func(t *testing.T) {
		action, err := reg.CreateAction("message_sender")
		require.NoError(t, err)

		err = action.Initialize()
		require.NoError(t, err)

		// 准备测试参数
		params := map[string]interface{}{
			"channels": []string{"webhook"},
			"title":    "Test Message",
			"content":  "This is a test message",
			"priority": "normal",
		}

		// 创建动作上下文
		actionContext := &types.ActionContext{
			WorkflowID: 12347,
			Variables:  make(map[string]interface{}),
			Metadata:   make(map[string]string),
			CreatedAt:  time.Now(),
		}

		// 执行动作
		ctx := context.Background()
		updatedContext, err := action.Execute(ctx, 12347, params, actionContext)
		require.NoError(t, err)
		assert.NotNil(t, updatedContext)

		// 检查执行结果
		assert.True(t, action.IsDone())
		assert.True(t, action.IsSuccess())

		// 清理
		err = action.Cleanup()
		require.NoError(t, err)
	})
}

// TestActionCaching 测试动作缓存
func TestActionCaching(t *testing.T) {
	reg := registry.GetDefaultRegistry()

	action, err := reg.CreateAction("file_scanner")
	require.NoError(t, err)

	err = action.Initialize()
	require.NoError(t, err)
	defer action.Cleanup()

	// 创建动作上下文
	actionContext := &types.ActionContext{
		WorkflowID: 12348,
		Variables:  make(map[string]interface{}),
		Metadata:   make(map[string]string),
		CreatedAt:  time.Now(),
	}

	// 准备测试参数
	params := map[string]interface{}{
		"scan_path": []string{"/tmp"},
		"limit":     10,
	}

	ctx := context.Background()

	// 第一次执行
	_, err = action.Execute(ctx, 12348, params, actionContext)
	require.NoError(t, err)

	// 检查缓存
	cached := action.CheckCache(ctx, 12348, "scan_result")
	assert.True(t, cached)

	// 清除缓存
	err = action.ClearCache(ctx, 12348)
	require.NoError(t, err)

	// 检查缓存已清除
	cached = action.CheckCache(ctx, 12348, "scan_result")
	assert.False(t, cached)
}

// TestActionErrorHandling 测试动作错误处理
func TestActionErrorHandling(t *testing.T) {
	reg := registry.GetDefaultRegistry()

	action, err := reg.CreateAction("plugin_invoker")
	require.NoError(t, err)

	err = action.Initialize()
	require.NoError(t, err)
	defer action.Cleanup()

	// 创建动作上下文
	actionContext := &types.ActionContext{
		WorkflowID: 12349,
		Variables:  make(map[string]interface{}),
		Metadata:   make(map[string]string),
		CreatedAt:  time.Now(),
	}

	// 准备无效参数（缺少必需的plugin_id）
	params := map[string]interface{}{
		"method": "test_method",
	}

	// 执行动作（应该失败）
	ctx := context.Background()
	_, err = action.Execute(ctx, 12349, params, actionContext)
	assert.Error(t, err)

	// 检查错误状态
	assert.True(t, action.IsDone())
	assert.False(t, action.IsSuccess())
	assert.NotEmpty(t, action.GetMessage())
}

// TestActionMetadata 测试动作元数据
func TestActionMetadata(t *testing.T) {
	reg := registry.GetDefaultRegistry()

	actions := reg.ListActions()
	for _, actionName := range actions {
		t.Run(actionName, func(t *testing.T) {
			info, err := reg.GetActionInfo(actionName)
			require.NoError(t, err)

			assert.NotEmpty(t, info.Name)
			assert.NotEmpty(t, info.Description)
			assert.NotEmpty(t, info.Version)
			assert.NotEmpty(t, info.Author)
			assert.NotEmpty(t, info.Category)

			// 创建动作实例测试
			action, err := reg.CreateAction(actionName)
			require.NoError(t, err)

			assert.Equal(t, info.Name, action.Name())
			assert.Equal(t, info.Description, action.Description())
			assert.Equal(t, info.Version, action.Version())
			assert.Equal(t, info.Author, action.Author())
			assert.Equal(t, info.Category, action.Category())
		})
	}
}

// TestAction 测试动作
type TestAction struct {
	done    bool
	success bool
	message string
	data    map[string]interface{}
}

// Name 返回动作名称
func (t *TestAction) Name() string {
	return "TestAction"
}

// Description 返回动作描述
func (t *TestAction) Description() string {
	return "Test action for unit testing"
}

// Version 返回动作版本
func (t *TestAction) Version() string {
	return "1.0.0"
}

// Author 返回动作作者
func (t *TestAction) Author() string {
	return "Test Author"
}

// Category 返回动作类别
func (t *TestAction) Category() string {
	return "test"
}

// Tags 返回动作标签
func (t *TestAction) Tags() []string {
	return []string{"test", "unit"}
}

// Execute 执行动作
func (t *TestAction) Execute(ctx context.Context, workflowID int64, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error) {
	t.data = params
	t.done = true
	t.success = true
	t.message = "Test action executed successfully"
	return actionContext, nil
}

// IsDone 检查是否完成
func (t *TestAction) IsDone() bool {
	return t.done
}

// IsSuccess 检查是否成功
func (t *TestAction) IsSuccess() bool {
	return t.success
}

// GetMessage 获取消息
func (t *TestAction) GetMessage() string {
	return t.message
}

// SetDone 设置完成状态
func (t *TestAction) SetDone(message string) {
	t.done = true
	t.success = true
	t.message = message
}

// SetError 设置错误状态
func (t *TestAction) SetError(message string) {
	t.done = true
	t.success = false
	t.message = message
}

// GetData 获取数据
func (t *TestAction) GetData() map[string]interface{} {
	return t.data
}

// SetData 设置数据
func (t *TestAction) SetData(key string, value interface{}) {
	if t.data == nil {
		t.data = make(map[string]interface{})
	}
	t.data[key] = value
}

// CheckCache 检查缓存
func (t *TestAction) CheckCache(ctx context.Context, workflowID int64, key string) bool {
	return false
}

// GetCache 获取缓存
func (t *TestAction) GetCache(ctx context.Context, workflowID int64, key string) (interface{}, error) {
	return nil, nil
}

// SaveCache 保存缓存
func (t *TestAction) SaveCache(ctx context.Context, workflowID int64, key string, data interface{}, ttl time.Duration) error {
	return nil
}

// ClearCache 清除缓存
func (t *TestAction) ClearCache(ctx context.Context, workflowID int64) error {
	return nil
}

// Initialize 初始化
func (t *TestAction) Initialize() error {
	return nil
}

// Cleanup 清理
func (t *TestAction) Cleanup() error {
	return nil
}