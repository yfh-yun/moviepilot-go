package utils

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// ProgressKey 进度键类型
type ProgressKey string

// ProgressData 进度数据
type ProgressData struct {
	Enable bool           `json:"enable"`
	Value  float64        `json:"value"`
	Text   string         `json:"text"`
	Data   map[string]any `json:"data"`
	Expire time.Time      `json:"expire"`
}

// TTLCache TTL缓存接口
type TTLCache interface {
	// Get 获取缓存值
	Get(key string) (any, bool)
	// Set 设置缓存值
	Set(key string, value any, ttl time.Duration)
	// Delete 删除缓存值
	Delete(key string)
}

// ProgressHelper 进度帮助类
type ProgressHelper struct {
	logger   *zap.Logger
	key      ProgressKey
	progress TTLCache
	mutex    sync.RWMutex
}

// NewProgressHelper 创建进度帮助类实例
func NewProgressHelper(key ProgressKey, cache TTLCache) *ProgressHelper {
	return &ProgressHelper{
		logger:   logger.GetLogger(),
		key:      key,
		progress: cache,
	}
}

// reset 重置进度
func (h *ProgressHelper) reset() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 创建默认进度数据
	progressData := ProgressData{
		Enable: false,
		Value:  0,
		Text:   "请稍候...",
		Data:   make(map[string]any),
		Expire: time.Now().Add(24 * time.Hour),
	}

	// 设置缓存，TTL为24小时
	h.progress.Set(string(h.key), progressData, 24*time.Hour)
}

// Start 开始进度
func (h *ProgressHelper) Start() {
	h.reset()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 获取当前进度
	progress, found := h.progress.Get(string(h.key))
	if !found {
		h.logger.Error("进度数据不存在")
		return
	}

	// 更新进度状态
	progressData, ok := progress.(ProgressData)
	if !ok {
		h.logger.Error("无效的进度数据类型")
		return
	}

	progressData.Enable = true
	h.progress.Set(string(h.key), progressData, 24*time.Hour)
}

// End 结束进度
func (h *ProgressHelper) End() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 获取当前进度
	progress, found := h.progress.Get(string(h.key))
	if !found {
		h.logger.Error("进度数据不存在")
		return
	}

	// 更新进度状态
	progressData, ok := progress.(ProgressData)
	if !ok {
		h.logger.Error("无效的进度数据类型")
		return
	}

	progressData.Enable = false
	progressData.Value = 100
	progressData.Text = ""
	h.progress.Set(string(h.key), progressData, 24*time.Hour)
}

// Update 更新进度
func (h *ProgressHelper) Update(value float64, text string, data map[string]any) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 获取当前进度
	progress, found := h.progress.Get(string(h.key))
	if !found {
		h.logger.Error("进度数据不存在")
		return
	}

	// 更新进度状态
	progressData, ok := progress.(ProgressData)
	if !ok {
		h.logger.Error("无效的进度数据类型")
		return
	}

	// 检查进度是否启用
	if !progressData.Enable {
		return
	}

	// 更新进度值
	if value > 0 {
		progressData.Value = value
	}

	// 更新进度文本
	if text != "" {
		progressData.Text = text
	}

	// 更新进度数据
	if data != nil {
		if progressData.Data == nil {
			progressData.Data = make(map[string]any)
		}
		for k, v := range data {
			progressData.Data[k] = v
		}
	}

	// 保存更新后的进度
	h.progress.Set(string(h.key), progressData, 24*time.Hour)
}

// Get 获取进度
func (h *ProgressHelper) Get() *ProgressData {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 获取当前进度
	progress, found := h.progress.Get(string(h.key))
	if !found {
		return nil
	}

	// 转换为ProgressData类型
	progressData, ok := progress.(ProgressData)
	if !ok {
		h.logger.Error("无效的进度数据类型")
		return nil
	}

	// 检查是否过期
	if time.Now().After(progressData.Expire) {
		h.progress.Delete(string(h.key))
		return nil
	}

	return &progressData
}

// Delete 删除进度
func (h *ProgressHelper) Delete() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 删除进度数据
	h.progress.Delete(string(h.key))
}

// IsEnabled 检查进度是否启用
func (h *ProgressHelper) IsEnabled() bool {
	progress := h.Get()
	return progress != nil && progress.Enable
}

// GetValue 获取进度值
func (h *ProgressHelper) GetValue() float64 {
	progress := h.Get()
	if progress == nil {
		return 0
	}
	return progress.Value
}

// GetText 获取进度文本
func (h *ProgressHelper) GetText() string {
	progress := h.Get()
	if progress == nil {
		return ""
	}
	return progress.Text
}

// GetData 获取进度数据
func (h *ProgressHelper) GetData() map[string]any {
	progress := h.Get()
	if progress == nil {
		return nil
	}
	return progress.Data
}
