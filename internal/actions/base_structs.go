package actions

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
)

// ActionContext 表示 Action 链路中需要共享的数据载体。
type ActionContext struct {
	WorkflowID       int
	Files            []FileItem
	Medias           []models.Media
	Downloads        []models.DownloadHistory
	Transfers        []models.TransferHistory
	UserConfig       map[string]any
	WorkflowMetadata map[string]any
	StartedAt        time.Time
	UpdatedAt        time.Time
}

// Ensure initializes optional maps to避免 nil map 操作。
func (ctx *ActionContext) Ensure() {
	if ctx.UserConfig == nil {
		ctx.UserConfig = make(map[string]any)
	}
	if ctx.WorkflowMetadata == nil {
		ctx.WorkflowMetadata = make(map[string]any)
	}
	if ctx.StartedAt.IsZero() {
		ctx.StartedAt = time.Now()
	}
	ctx.UpdatedAt = time.Now()
}

// FileItem 表示扫描得到的文件或目录。
type FileItem struct {
	Path     string
	Size     int64
	ModTime  time.Time
	IsDir    bool
	Metadata map[string]string
}

// BaseAction 约束所有 Action 的统一接口。
type BaseAction interface {
	Name() string
	Description() string
	Data() any
	Done() bool
	Success() bool
	Message() string
	CheckCache(workflowID int, key string) bool
	SaveCache(workflowID int, data any)
	Execute(workflowID int, params any, ctx *ActionContext) (*ActionContext, error)
}

// BaseActionImpl 为具体 Action 提供通用字段与方法，具体 Action 通过内嵌来复用。
type BaseActionImpl struct {
	name        string
	description string
	data        any
	done        bool
	success     bool
	message     string
	logger      *zap.Logger
}

// NewBaseActionImpl 创建一个基础实现实例。
func NewBaseActionImpl(name, description string, logger *zap.Logger) BaseActionImpl {
	return BaseActionImpl{
		name:        name,
		description: description,
		logger:      logger,
	}
}

// Name 返回 Action 名称。
func (b *BaseActionImpl) Name() string { return b.name }

// Description 返回 Action 描述。
func (b *BaseActionImpl) Description() string { return b.description }

// Data 返回 Action 业务数据。
func (b *BaseActionImpl) Data() any { return b.data }

// Done 指示是否已执行完成。
func (b *BaseActionImpl) Done() bool { return b.done }

// Success 指示是否执行成功。
func (b *BaseActionImpl) Success() bool { return b.success }

// Message 返回人类可读的执行信息。
func (b *BaseActionImpl) Message() string { return b.message }

// Logger 返回注入的日志实例（可能为 nil）。
func (b *BaseActionImpl) Logger() *zap.Logger { return b.logger }

// SetResult 统一更新执行结果。
func (b *BaseActionImpl) SetResult(success bool, message string, data any) {
	b.done = true
	b.success = success
	b.message = message
	b.data = data
}

var actionCache = struct {
	sync.RWMutex
	data map[int]map[string]any
}{
	data: make(map[int]map[string]any),
}

// CheckCache 判断 workflow 对应的缓存中是否存在 key。
func (b *BaseActionImpl) CheckCache(workflowID int, key string) bool {
	actionCache.RLock()
	defer actionCache.RUnlock()
	if cache, ok := actionCache.data[workflowID]; ok {
		_, exists := cache[key]
		return exists
	}
	return false
}

// SaveCache 按 Action 名称保存数据，便于简单复用。
func (b *BaseActionImpl) SaveCache(workflowID int, data any) {
	b.saveCacheWithKey(workflowID, b.name, data)
}

// SaveCacheWithKey 允许子类指定自定义 key。
func (b *BaseActionImpl) SaveCacheWithKey(workflowID int, key string, data any) {
	b.saveCacheWithKey(workflowID, key, data)
}

func (b *BaseActionImpl) saveCacheWithKey(workflowID int, key string, data any) {
	if key == "" {
		if b.logger != nil {
			b.logger.Warn("cache key is empty", zap.String("action", b.name))
		}
		return
	}
	actionCache.Lock()
	defer actionCache.Unlock()
	if _, ok := actionCache.data[workflowID]; !ok {
		actionCache.data[workflowID] = make(map[string]any)
	}
	actionCache.data[workflowID][key] = data
}

// GetCacheValue 尝试读取缓存值。
func (b *BaseActionImpl) GetCacheValue(workflowID int, key string) (any, bool) {
	actionCache.RLock()
	defer actionCache.RUnlock()
	if cache, ok := actionCache.data[workflowID]; ok {
		val, exists := cache[key]
		return val, exists
	}
	return nil, false
}

// Execute 由具体 Action 实现。
func (b *BaseActionImpl) Execute(_ int, _ any, ctx *ActionContext) (*ActionContext, error) {
	return ctx, fmt.Errorf("action %s has not implemented Execute", b.name)
}
