package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StorageManager 存储管理器实现
type StorageManager struct {
	providers map[string]StorageProvider
	mu        sync.RWMutex
	logger    *logger.Logger
}

// NewStorageManager 创建存储管理器
func NewStorageManager() *StorageManager {
	return &StorageManager{
		providers: make(map[string]StorageProvider),
		logger:    logger.NewLogger("storage"),
	}
}

// RegisterProvider 注册存储提供商
func (sm *StorageManager) RegisterProvider(name string, provider StorageProvider) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.providers[name]; exists {
		return NewError("PROVIDER_EXISTS", "存储提供商已存在", name)
	}

	sm.providers[name] = provider
	sm.logger.Infof("存储提供商注册成功: %s (%s)", name, provider.Type())

	return nil
}

// GetProvider 获取存储提供商
func (sm *StorageManager) GetProvider(name string) (StorageProvider, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	provider, exists := sm.providers[name]
	if !exists {
		return nil, NewError("PROVIDER_NOT_FOUND", "存储提供商未找到", name)
	}

	return provider, nil
}

// ListProviders 列出所有存储提供商
func (sm *StorageManager) ListProviders() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var providers []string
	for name := range sm.providers {
		providers = append(providers, name)
	}

	return providers
}

// RemoveProvider 移除存储提供商
func (sm *StorageManager) RemoveProvider(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	provider, exists := sm.providers[name]
	if !exists {
		return NewError("PROVIDER_NOT_FOUND", "存储提供商未找到", name)
	}

	// 断开连接
	if err := provider.Disconnect(); err != nil {
		sm.logger.Warnf("断开存储提供商连接失败: %s - %v", name, err)
	}

	delete(sm.providers, name)
	sm.logger.Infof("存储提供商移除成功: %s", name)

	return nil
}

// ConnectAll 连接所有存储提供商
func (sm *StorageManager) ConnectAll(ctx context.Context) map[string]error {
	sm.mu.RLock()
	providers := make(map[string]StorageProvider)
	for name, provider := range sm.providers {
		providers[name] = provider
	}
	sm.mu.RUnlock()

	errors := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, provider := range providers {
		wg.Add(1)
		go func(name string, provider StorageProvider) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := provider.Connect(ctx); err != nil {
				mu.Lock()
				errors[name] = err
				mu.Unlock()
				sm.logger.Errorf("连接存储提供商失败: %s - %v", name, err)
			} else {
				sm.logger.Infof("存储提供商连接成功: %s", name)
			}
		}(name, provider)
	}

	wg.Wait()
	return errors
}

// DisconnectAll 断开所有存储提供商
func (sm *StorageManager) DisconnectAll() map[string]error {
	sm.mu.RLock()
	providers := make(map[string]StorageProvider)
	for name, provider := range sm.providers {
		providers[name] = provider
	}
	sm.mu.RUnlock()

	errors := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, provider := range providers {
		wg.Add(1)
		go func(name string, provider StorageProvider) {
			defer wg.Done()

			if err := provider.Disconnect(); err != nil {
				mu.Lock()
				errors[name] = err
				mu.Unlock()
				sm.logger.Errorf("断开存储提供商连接失败: %s - %v", name, err)
			} else {
				sm.logger.Infof("存储提供商断开连接成功: %s", name)
			}
		}(name, provider)
	}

	wg.Wait()
	return errors
}

// GetConnectedProviders 获取已连接的存储提供商
func (sm *StorageManager) GetConnectedProviders() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var connected []string
	for name, provider := range sm.providers {
		if provider.IsConnected() {
			connected = append(connected, name)
		}
	}

	return connected
}

// HealthCheck 健康检查
func (sm *StorageManager) HealthCheck(ctx context.Context) map[string]bool {
	sm.mu.RLock()
	providers := make(map[string]StorageProvider)
	for name, provider := range sm.providers {
		providers[name] = provider
	}
	sm.mu.RUnlock()

	health := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, provider := range providers {
		wg.Add(1)
		go func(name string, provider StorageProvider) {
			defer wg.Done()

			if !provider.IsConnected() {
				mu.Lock()
				health[name] = false
				mu.Unlock()
				return
			}

			// 简单健康检查：尝试列出根目录
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if _, err := provider.List(ctx, "/"); err != nil {
				mu.Lock()
				health[name] = false
				mu.Unlock()
				sm.logger.Warnf("存储提供商健康检查失败: %s - %v", name, err)
			} else {
				mu.Lock()
				health[name] = true
				mu.Unlock()
			}
		}(name, provider)
	}

	wg.Wait()
	return health
}

// GetQuotaSummary 获取所有存储的配额汇总
func (sm *StorageManager) GetQuotaSummary(ctx context.Context) (*QuotaSummary, error) {
	sm.mu.RLock()
	providers := make(map[string]StorageProvider)
	for name, provider := range sm.providers {
		providers[name] = provider
	}
	sm.mu.RUnlock()

	summary := &QuotaSummary{
		Providers: make(map[string]*QuotaInfo),
	}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, provider := range providers {
		if !provider.IsConnected() {
			continue
		}

		wg.Add(1)
		go func(name string, provider StorageProvider) {
			defer wg.Done()

			quota, err := provider.GetQuota(ctx)
			if err != nil {
				sm.logger.Warnf("获取存储配额失败: %s - %v", name, err)
				return
			}

			mu.Lock()
			summary.Providers[name] = quota
			summary.Total += quota.Total
			summary.Used += quota.Used
			summary.Available += quota.Available
			mu.Unlock()
		}(name, provider)
	}

	wg.Wait()
	return summary, nil
}

// QuotaSummary 配额汇总信息
type QuotaSummary struct {
	Total     int64                 `json:"total"`
	Used      int64                 `json:"used"`
	Available int64                 `json:"available"`
	Providers map[string]*QuotaInfo `json:"providers"`
}
