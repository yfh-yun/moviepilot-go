package helper

import (
	"moviepilot-go/internal/utils/cache"
	"moviepilot-go/pkg/models"
	"sync"
)

// ProgressHelper 处理进度辅助�?type ProgressHelper struct {
	key      string
	progress *cache.TTLCache
}

// ProgressData 进度数据结构
type ProgressData struct {
	Enable bool                   `json:"enable"`
	Value  float64                `json:"value"`
	Text   string                 `json:"text"`
	Data   map[string]interface{} `json:"data"`
}

var (
	// 单例实例
	instances = make(map[string]*ProgressHelper)
	// 保护instances的互斥锁
	instanceMutex sync.RWMutex
)

// NewProgressHelper 创建ProgressHelper实例（单例模式）
func NewProgressHelper(key interface{}) *ProgressHelper {
	// 转换key为字符串
	var keyStr string
	switch k := key.(type) {
	case models.ProgressKey:
		keyStr = string(k)
	case string:
		keyStr = k
	default:
		keyStr = ""
	}

	// 使用读锁检查实例是否已存在
	instanceMutex.RLock()
	if instance, exists := instances[keyStr]; exists {
		instanceMutex.RUnlock()
		return instance
	}
	instanceMutex.RUnlock()

	// 使用写锁创建新实�?	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	// 双重检�?	if instance, exists := instances[keyStr]; exists {
		return instance
	}

	// 创建新的ProgressHelper实例
	helper := &ProgressHelper{
		key:      keyStr,
		progress: cache.NewTTLCache("progress", 1024, 24*60*60), // 24小时TTL
	}

	instances[keyStr] = helper
	return helper
}

// reset 重置进度
func (p *ProgressHelper) reset() {
	/*
	 * 重置进度
	 */
	data := &ProgressData{
		Enable: false,
		Value:  0,
		Text:   "请稍�?..",
		Data:   make(map[string]interface{}),
	}
	p.progress.Set(p.key, data)
}

// Start 开始进�?func (p *ProgressHelper) Start() {
	/*
	 * 开始进�?	 */
	p.reset()
	
	current, exists := p.progress.Get(p.key)
	if !exists {
		return
	}
	
	if data, ok := current.(*ProgressData); ok {
		data.Enable = true
		p.progress.Set(p.key, data)
	}
}

// End 结束进度
func (p *ProgressHelper) End() {
	/*
	 * 结束进度
	 */
	current, exists := p.progress.Get(p.key)
	if !exists {
		return
	}
	
	if data, ok := current.(*ProgressData); ok {
		data.Enable = false
		data.Value = 100
		data.Text = ""
		p.progress.Set(p.key, data)
	}
}

// Update 更新进度
func (p *ProgressHelper) Update(value *float64, text *string, data map[string]interface{}) {
	/*
	 * 更新进度
	 */
	current, exists := p.progress.Get(p.key)
	if !exists {
		return
	}
	
	if progressData, ok := current.(*ProgressData); ok {
		if !progressData.Enable {
			return
		}
		
		if value != nil {
			progressData.Value = *value
		}
		
		if text != nil {
			progressData.Text = *text
		}
		
		if data != nil {
			if progressData.Data == nil {
				progressData.Data = make(map[string]interface{})
			}
			for k, v := range data {
				progressData.Data[k] = v
			}
		}
		
		p.progress.Set(p.key, progressData)
	}
}

// Get 获取进度
func (p *ProgressHelper) Get() *ProgressData {
	current, exists := p.progress.Get(p.key)
	if !exists {
		return nil
	}
	
	if data, ok := current.(*ProgressData); ok {
		return data
	}
	
	return nil
}
