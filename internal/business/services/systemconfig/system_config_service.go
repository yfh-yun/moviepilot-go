package systemconfig

import (
	"context"
	"sync"

	"moviepilot-go/internal/repositories/interfaces"
)

// Service 提供类似 Python SystemConfigOper 的系统配置管理能力：
// - 启动时将配置加载到内存
// - 所有读写以进程内缓存为主，再按需落库
// - 提供 Set/Get/All/Delete 接口
type Service struct {
	repo interfaces.SystemConfigRepository
	mu   sync.RWMutex
	data map[string]string
}

// NewService 创建系统配置服务
func NewService(repo interfaces.SystemConfigRepository) *Service {
	return &Service{
		repo: repo,
		data: make(map[string]string),
	}
}

// Initialize 从仓储加载所有配置到内存
func (s *Service) Initialize(ctx context.Context) error {
	if s.repo == nil {
		return nil
	}
	configs, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]string, len(configs))
	for _, c := range configs {
		if c != nil {
			s.data[c.Key] = c.Value
		}
	}
	return nil
}

// Get 获取单个配置值（从内存缓存读取）
func (s *Service) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// All 返回所有配置的副本
func (s *Service) All() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string]string, len(s.data))
	for k, v := range s.data {
		copy[k] = v
	}
	return copy
}

// Set 设置配置值；当新旧值不同时才更新仓储
// 返回值：
//
//	true  - 已更新（插入或修改或删除）
//	false - 无需更新（值未变化）
func (s *Service) Set(ctx context.Context, key, value string) (bool, error) {
	if s.repo == nil {
		return false, nil
	}

	s.mu.Lock()
	old, exists := s.data[key]
	if value == "" {
		delete(s.data, key)
	} else {
		s.data[key] = value
	}
	s.mu.Unlock()

	if exists && old == value {
		return false, nil
	}

	if value == "" {
		if err := s.repo.Delete(ctx, key); err != nil {
			return false, err
		}
		return true, nil
	}

	if err := s.repo.Set(ctx, key, value); err != nil {
		return false, err
	}
	return true, nil
}

// Delete 删除配置：更新缓存并从仓储删除
func (s *Service) Delete(ctx context.Context, key string) error {
	if s.repo == nil {
		return nil
	}

	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()

	return s.repo.Delete(ctx, key)
}
