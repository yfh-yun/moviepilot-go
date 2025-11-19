package utils

import (
	"fmt"
	"sync"
	"time"
)

// ProgressHelper 进度处理助手
type ProgressHelper struct {
	key      string
	progress map[string]*ProgressEntry
	mutex    sync.RWMutex
}

// ProgressEntry 进度条目
type ProgressEntry struct {
	Enable bool                   `json:"enable"`
	Value  int                    `json:"value"`
	Text   string                 `json:"text"`
	Data   map[string]interface{} `json:"data"`
	Start  time.Time              `json:"start"`
	Update time.Time              `json:"update"`
}

// ProgressKey 进度键类型
type ProgressKey string

const (
	ProgressKeySearch    ProgressKey = "search"
	ProgressKeyDownload  ProgressKey = "download"
	ProgressKeyTransfer  ProgressKey = "transfer"
	ProgressKeyScan      ProgressKey = "scan"
	ProgressKeySync      ProgressKey = "sync"
	ProgressKeyBackup    ProgressKey = "backup"
	ProgressKeyRestore   ProgressKey = "restore"
	ProgressKeyInstall   ProgressKey = "install"
	ProgressKeyUninstall ProgressKey = "uninstall"
)

// NewProgressHelper 创建进度助手实例
func NewProgressHelper(key interface{}) *ProgressHelper {
	var keyStr string
	if k, ok := key.(ProgressKey); ok {
		keyStr = string(k)
	} else if k, ok := key.(string); ok {
		keyStr = k
	} else {
		keyStr = fmt.Sprintf("%v", key)
	}

	return &ProgressHelper{
		key:      keyStr,
		progress: make(map[string]*ProgressEntry),
	}
}

// Start 开始进度
func (ph *ProgressHelper) Start() {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	ph.reset()
	current := ph.progress[ph.key]
	if current == nil {
		current = &ProgressEntry{}
		ph.progress[ph.key] = current
	}

	current.Enable = true
	current.Start = time.Now()
	current.Update = time.Now()
}

// End 结束进度
func (ph *ProgressHelper) End() {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	current := ph.progress[ph.key]
	if current == nil {
		return
	}

	current.Enable = false
	current.Value = 100
	current.Text = "完成"
	current.Update = time.Now()
}

// Update 更新进度
func (ph *ProgressHelper) Update(value int, text string) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	current := ph.progress[ph.key]
	if current == nil {
		current = &ProgressEntry{}
		ph.progress[ph.key] = current
	}

	if value < 0 {
		value = 0
	} else if value > 100 {
		value = 100
	}

	current.Value = value
	current.Text = text
	current.Update = time.Now()
}

// SetValue 设置进度值
func (ph *ProgressHelper) SetValue(value int) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	current := ph.progress[ph.key]
	if current == nil {
		current = &ProgressEntry{}
		ph.progress[ph.key] = current
	}

	if value < 0 {
		value = 0
	} else if value > 100 {
		value = 100
	}

	current.Value = value
	current.Update = time.Now()
}

// SetText 设置进度文本
func (ph *ProgressHelper) SetText(text string) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	current := ph.progress[ph.key]
	if current == nil {
		current = &ProgressEntry{}
		ph.progress[ph.key] = current
	}

	current.Text = text
	current.Update = time.Now()
}

// SetData 设置进度数据
func (ph *ProgressHelper) SetData(data map[string]interface{}) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	current := ph.progress[ph.key]
	if current == nil {
		current = &ProgressEntry{}
		ph.progress[ph.key] = current
	}

	current.Data = data
	current.Update = time.Now()
}

// AddData 添加进度数据
func (ph *ProgressHelper) AddData(key string, value interface{}) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	current := ph.progress[ph.key]
	if current == nil {
		current = &ProgressEntry{}
		ph.progress[ph.key] = current
	}

	if current.Data == nil {
		current.Data = make(map[string]interface{})
	}

	current.Data[key] = value
	current.Update = time.Now()
}

// GetProgress 获取进度
func (ph *ProgressHelper) GetProgress() *ProgressEntry {
	ph.mutex.RLock()
	defer ph.mutex.RUnlock()

	if current, exists := ph.progress[ph.key]; exists {
		// 返回副本
		return &ProgressEntry{
			Enable: current.Enable,
			Value:  current.Value,
			Text:   current.Text,
			Data:   current.Data,
			Start:  current.Start,
			Update: current.Update,
		}
	}

	return &ProgressEntry{
		Enable: false,
		Value:  0,
		Text:   "请稍候...",
		Data:   make(map[string]interface{}),
	}
}

// GetValue 获取进度值
func (ph *ProgressHelper) GetValue() int {
	progress := ph.GetProgress()
	return progress.Value
}

// GetText 获取进度文本
func (ph *ProgressHelper) GetText() string {
	progress := ph.GetProgress()
	return progress.Text
}

// GetData 获取进度数据
func (ph *ProgressHelper) GetData() map[string]interface{} {
	progress := ph.GetProgress()
	return progress.Data
}

// IsEnabled 检查进度是否启用
func (ph *ProgressHelper) IsEnabled() bool {
	progress := ph.GetProgress()
	return progress.Enable
}

// IsFinished 检查进度是否完成
func (ph *ProgressHelper) IsFinished() bool {
	progress := ph.GetProgress()
	return !progress.Enable && progress.Value == 100
}

// GetElapsed 获取已用时间
func (ph *ProgressHelper) GetElapsed() time.Duration {
	progress := ph.GetProgress()
	if progress.Start.IsZero() {
		return 0
	}
	return time.Since(progress.Start)
}

// GetRemaining 估算剩余时间（简单实现）
func (ph *ProgressHelper) GetRemaining() time.Duration {
	progress := ph.GetProgress()
	if progress.Start.IsZero() || progress.Value <= 0 {
		return 0
	}

	elapsed := time.Since(progress.Start)
	if progress.Value >= 100 {
		return 0
	}

	// 简单线性估算
	estimatedTotal := elapsed * time.Duration(100) / time.Duration(progress.Value)
	remaining := estimatedTotal - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// reset 重置进度
func (ph *ProgressHelper) reset() {
	ph.progress[ph.key] = &ProgressEntry{
		Enable: false,
		Value:  0,
		Text:   "请稍候...",
		Data:   make(map[string]interface{}),
		Start:  time.Time{},
		Update: time.Time{},
	}
}

// Reset 重置进度
func (ph *ProgressHelper) Reset() {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	ph.reset()
}

// Destroy 销毁进度
func (ph *ProgressHelper) Destroy() {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	delete(ph.progress, ph.key)
}

// ProgressManager 进度管理器
type ProgressManager struct {
	progresses map[string]*ProgressHelper
	mutex      sync.RWMutex
}

// NewProgressManager 创建进度管理器
func NewProgressManager() *ProgressManager {
	return &ProgressManager{
		progresses: make(map[string]*ProgressHelper),
	}
}

// GetProgress 获取或创建进度助手
func (pm *ProgressManager) GetProgress(key interface{}) *ProgressHelper {
	var keyStr string
	if k, ok := key.(ProgressKey); ok {
		keyStr = string(k)
	} else if k, ok := key.(string); ok {
		keyStr = k
	} else {
		keyStr = fmt.Sprintf("%v", key)
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if progress, exists := pm.progresses[keyStr]; exists {
		return progress
	}

	progress := NewProgressHelper(keyStr)
	pm.progresses[keyStr] = progress
	return progress
}

// RemoveProgress 移除进度
func (pm *ProgressManager) RemoveProgress(key interface{}) {
	var keyStr string
	if k, ok := key.(ProgressKey); ok {
		keyStr = string(k)
	} else if k, ok := key.(string); ok {
		keyStr = k
	} else {
		keyStr = fmt.Sprintf("%v", key)
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if progress, exists := pm.progresses[keyStr]; exists {
		progress.Destroy()
		delete(pm.progresses, keyStr)
	}
}

// GetAllProgresses 获取所有进度
func (pm *ProgressManager) GetAllProgresses() map[string]*ProgressEntry {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*ProgressEntry)
	for key, progress := range pm.progresses {
		result[key] = progress.GetProgress()
	}

	return result
}

// ClearAll 清空所有进度
func (pm *ProgressManager) ClearAll() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for _, progress := range pm.progresses {
		progress.Destroy()
	}

	pm.progresses = make(map[string]*ProgressHelper)
}

// GetActiveProgresses 获取活跃的进度
func (pm *ProgressManager) GetActiveProgresses() map[string]*ProgressEntry {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*ProgressEntry)
	for key, progress := range pm.progresses {
		if progress.IsEnabled() {
			result[key] = progress.GetProgress()
		}
	}

	return result
}

// GetProgressCount 获取进度数量
func (pm *ProgressManager) GetProgressCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return len(pm.progresses)
}

// GetActiveCount 获取活跃进度数量
func (pm *ProgressManager) GetActiveCount() int {
	active := pm.GetActiveProgresses()
	return len(active)
}

// CleanupExpired 清理过期的进度（超过1小时未更新）
func (pm *ProgressManager) CleanupExpired() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	expiredKeys := make([]string, 0)
	expiry := time.Now().Add(-time.Hour)

	for key, progress := range pm.progresses {
		progressEntry := progress.GetProgress()
		if !progressEntry.Enable && !progressEntry.Update.IsZero() && progressEntry.Update.Before(expiry) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		if progress, exists := pm.progresses[key]; exists {
			progress.Destroy()
			delete(pm.progresses, key)
		}
	}
}

// BatchUpdate 批量更新进度
func (pm *ProgressManager) BatchUpdate(updates map[string]struct {
	Value int
	Text  string
}) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for key, update := range updates {
		if progress, exists := pm.progresses[key]; exists {
			progress.Update(update.Value, update.Text)
		}
	}
}

// ExportProgresses 导出进度信息
func (pm *ProgressManager) ExportProgresses() map[string]interface{} {
	progresses := pm.GetAllProgresses()
	result := make(map[string]interface{})

	for key, progress := range progresses {
		result[key] = map[string]interface{}{
			"enable": progress.Enable,
			"value":  progress.Value,
			"text":   progress.Text,
			"data":   progress.Data,
			"start":  progress.Start,
			"update": progress.Update,
		}
	}

	return result
}

// 全局进度管理器实例
var globalProgressManager = NewProgressManager()

// GetGlobalProgressManager 获取全局进度管理器
func GetGlobalProgressManager() *ProgressManager {
	return globalProgressManager
}

// GetProgress 获取全局进度助手
func GetProgress(key interface{}) *ProgressHelper {
	return globalProgressManager.GetProgress(key)
}

// RemoveProgress 移除全局进度
func RemoveProgress(key interface{}) {
	globalProgressManager.RemoveProgress(key)
}

// GetAllProgresses 获取所有全局进度
func GetAllProgresses() map[string]*ProgressEntry {
	return globalProgressManager.GetAllProgresses()
}