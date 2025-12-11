package progress

import (
	"context"
	"sync"
	"time"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// ProgressState 进度状态
type ProgressState struct {
	Enable bool      `json:"enable"`
	Value  int       `json:"value"`
	Text   string    `json:"text"`
	Data   string    `json:"data"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

// ProgressService 进度服务
type ProgressService struct {
	*base.ServiceBase
	mu    sync.RWMutex
	state map[string]*ProgressState
}

// NewProgressService 创建ProgressService实例
func NewProgressService() *ProgressService {
	logger.Debug("Creating new ProgressService instance")
	return &ProgressService{
		ServiceBase: base.NewServiceBase(),
		state:       make(map[string]*ProgressState),
	}
}

// Initialize 初始化服务
func (s *ProgressService) Initialize() error {
	logger.Info("Initializing ProgressService")

	// 启动定期清理goroutine，清理超过24小时的进度记录
	go s.cleanupExpiredProgress()

	return nil
}

// Name 获取服务名称
func (s *ProgressService) Name() string {
	return "ProgressService"
}

// Close 关闭服务
func (s *ProgressService) Close() error {
	logger.Info("Closing ProgressService")
	return nil
}

// Start 开始进度跟踪
func (s *ProgressService) Start(ctx context.Context, key string) {
	logger.Debug("Starting progress tracking", zap.String("key", key))

	s.mu.Lock()
	defer s.mu.Unlock()

	s.state[key] = &ProgressState{
		Enable: true,
		Value:  0,
		Text:   "",
		Data:   "",
		Start:  time.Now(),
		End:    time.Time{},
	}
}

// End 结束进度跟踪
func (s *ProgressService) End(ctx context.Context, key string) {
	logger.Debug("Ending progress tracking", zap.String("key", key))

	s.mu.Lock()
	defer s.mu.Unlock()

	if state, exists := s.state[key]; exists {
		state.Enable = false
		state.Value = 100
		state.End = time.Now()
	}
}

// Update 更新进度
func (s *ProgressService) Update(ctx context.Context, key string, value int, text string, data string) {
	logger.Debug("Updating progress",
		zap.String("key", key),
		zap.Int("value", value),
		zap.String("text", text))

	s.mu.Lock()
	defer s.mu.Unlock()

	if state, exists := s.state[key]; exists {
		if value >= 0 && value <= 100 {
			state.Value = value
		}
		if text != "" {
			state.Text = text
		}
		if data != "" {
			state.Data = data
		}
	}
}

// Get 获取进度
func (s *ProgressService) Get(ctx context.Context, key string) (*ProgressState, bool) {
	logger.Debug("Getting progress", zap.String("key", key))

	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.state[key]
	return state, exists
}

// GetAll 获取所有进度
func (s *ProgressService) GetAll(ctx context.Context) map[string]*ProgressState {
	logger.Debug("Getting all progress states")

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本，避免并发修改问题
	result := make(map[string]*ProgressState, len(s.state))
	for key, state := range s.state {
		result[key] = &ProgressState{
			Enable: state.Enable,
			Value:  state.Value,
			Text:   state.Text,
			Data:   state.Data,
			Start:  state.Start,
			End:    state.End,
		}
	}

	return result
}

// Delete 删除进度
func (s *ProgressService) Delete(ctx context.Context, key string) {
	logger.Debug("Deleting progress", zap.String("key", key))

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.state, key)
}

// cleanupExpiredProgress 清理过期的进度记录
func (s *ProgressService) cleanupExpiredProgress() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for key, state := range s.state {
				// 清理超过24小时的进度记录
				if now.Sub(state.Start) > 24*time.Hour {
					delete(s.state, key)
					logger.Debug("Cleaned up expired progress", zap.String("key", key))
				}
			}
			s.mu.Unlock()
		}
	}
}
