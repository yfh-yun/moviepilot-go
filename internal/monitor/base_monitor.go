package monitor

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// BaseMonitor 基础监控器，包含通用字段和方法
type BaseMonitor struct {
	Logger  *zap.Logger
	Ctx     context.Context
	Cancel  context.CancelFunc
	Running bool
	Mu      sync.RWMutex
}

// NewBaseMonitor 创建基础监控器
func NewBaseMonitor(logger *zap.Logger) *BaseMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseMonitor{
		Logger: logger,
		Ctx:    ctx,
		Cancel: cancel,
	}
}

// IsRunning 检查监控器是否运行中
func (bm *BaseMonitor) IsRunning() bool {
	bm.Mu.RLock()
	defer bm.Mu.RUnlock()
	return bm.Running
}

// SetRunning 设置运行状态
func (bm *BaseMonitor) SetRunning(running bool) {
	bm.Mu.Lock()
	defer bm.Mu.Unlock()
	bm.Running = running
}

// Stop 停止监控器
func (bm *BaseMonitor) Stop() {
	bm.Mu.Lock()
	defer bm.Mu.Unlock()
	if bm.Running {
		bm.Running = false
		bm.Cancel()
	}
}

// GetContext 获取上下文
func (bm *BaseMonitor) GetContext() context.Context {
	return bm.Ctx
}