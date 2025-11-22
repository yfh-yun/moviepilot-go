package subscribe

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repository/interfaces"
)

// Service 订阅服务接口
type Service interface {
	CreateSubscribe(req CreateSubscribeRequest) (*models.Subscribe, error)
	UpdateSubscribe(id uint, req UpdateSubscribeRequest) error
	DeleteSubscribe(id uint) error
	GetSubscribe(id uint) (*models.Subscribe, error)
	ListSubscribes(opts ListOptions) ([]models.Subscribe, int64, error)
	PauseSubscribe(id uint) error
	ResumeSubscribe(id uint) error
	GetActiveSubscribes() ([]models.Subscribe, error)
}

// DefaultService 默认订阅服务实现
type DefaultService struct {
	repo   interfaces.SubscribeRepository
	logger *zap.Logger
}

// NewService 创建订阅服务
func NewService(repo interfaces.SubscribeRepository, logger *zap.Logger) Service {
	return &DefaultService{
		repo:   repo,
		logger: logger,
	}
}

// CreateSubscribeRequest 创建订阅请求
type CreateSubscribeRequest struct {
	Name       string  `json:"name" binding:"required"`
	Year       *string `json:"year"`
	Type       string  `json:"type" binding:"required,oneof=movie tv"`
	Season     *int    `json:"season"`
	TMDBID     *int    `json:"tmdb_id"`
	IMDBID     *string `json:"imdb_id"`
	Quality    string  `json:"quality"`
	Resolution string  `json:"resolution"`
	Include    string  `json:"include"`
	Exclude    string  `json:"exclude"`
}

// UpdateSubscribeRequest 更新订阅请求
type UpdateSubscribeRequest struct {
	Name       *string `json:"name"`
	Quality    *string `json:"quality"`
	Resolution *string `json:"resolution"`
	Include    *string `json:"include"`
	Exclude    *string `json:"exclude"`
}

// ListOptions 列表选项
type ListOptions struct {
	Page     int
	PageSize int
	State    string
	Type     string
	OrderBy  string
}

// CreateSubscribe 创建订阅
func (s *DefaultService) CreateSubscribe(req CreateSubscribeRequest) (*models.Subscribe, error) {
	ctx := context.Background()

	// TODO: 检查是否已存在 - 需要实现 FindByTMDBID 方法
	// if req.TMDBID != nil {
	//	existing, _ := s.repo.FindByTMDBID(ctx, *req.TMDBID, req.Type, req.Season)
	//	if existing != nil {
	//		return nil, fmt.Errorf("subscription already exists for this media")
	//	}
	// }

	subscribe := &models.Subscribe{
		Name:       req.Name,
		Year:       req.Year,
		Type:       req.Type,
		Season:     req.Season,
		TMDBID:     req.TMDBID,
		IMDBID:     req.IMDBID,
		Quality:    req.Quality,
		Resolution: req.Resolution,
		Include:    req.Include,
		Exclude:    req.Exclude,
		State:      "N", // 新建状态
	}

	if err := s.repo.Create(ctx, subscribe); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("subscription created",
			zap.Uint("id", subscribe.ID),
			zap.String("name", subscribe.Name),
			zap.String("type", subscribe.Type))
	}

	return subscribe, nil
}

// UpdateSubscribe 更新订阅
func (s *DefaultService) UpdateSubscribe(id uint, req UpdateSubscribeRequest) error {
	ctx := context.Background()

	subscribe, err := s.repo.GetByID(ctx, fmt.Sprintf("%d", id))
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		subscribe.Name = *req.Name
	}
	if req.Quality != nil {
		subscribe.Quality = *req.Quality
	}
	if req.Resolution != nil {
		subscribe.Resolution = *req.Resolution
	}
	if req.Include != nil {
		subscribe.Include = *req.Include
	}
	if req.Exclude != nil {
		subscribe.Exclude = *req.Exclude
	}

	now := time.Now()
	subscribe.LastUpdate = &now

	if err := s.repo.Update(ctx, subscribe); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("subscription updated", zap.Uint("id", id))
	}

	return nil
}

// DeleteSubscribe 删除订阅
func (s *DefaultService) DeleteSubscribe(id uint) error {
	ctx := context.Background()

	if err := s.repo.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("subscription deleted", zap.Uint("id", id))
	}

	return nil
}

// GetSubscribe 获取订阅详情
func (s *DefaultService) GetSubscribe(id uint) (*models.Subscribe, error) {
	ctx := context.Background()

	subscribe, err := s.repo.GetByID(ctx, fmt.Sprintf("%d", id))
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	return subscribe, nil
}

// ListSubscribes 列表订阅
func (s *DefaultService) ListSubscribes(opts ListOptions) ([]models.Subscribe, int64, error) {
	ctx := context.Background()

	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.Page < 1 {
		opts.Page = 1
	}

	params := interfaces.ListSubscribeParams{
		Page:     opts.Page,
		PageSize: opts.PageSize,
		Status:   opts.State,
		Type:     opts.Type,
	}

	subscribes, total, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// 转换为非指针切片
	result := make([]models.Subscribe, len(subscribes))
	for i, sub := range subscribes {
		result[i] = *sub
	}

	return result, total, nil
}

// PauseSubscribe 暂停订阅
func (s *DefaultService) PauseSubscribe(id uint) error {
	ctx := context.Background()

	subscribe, err := s.repo.GetByID(ctx, fmt.Sprintf("%d", id))
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	subscribe.State = "S"
	if err := s.repo.Update(ctx, subscribe); err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("subscription paused", zap.Uint("id", id))
	}

	return nil
}

// ResumeSubscribe 恢复订阅
func (s *DefaultService) ResumeSubscribe(id uint) error {
	ctx := context.Background()

	subscribe, err := s.repo.GetByID(ctx, fmt.Sprintf("%d", id))
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	subscribe.State = "R"
	if err := s.repo.Update(ctx, subscribe); err != nil {
		return fmt.Errorf("failed to resume subscription: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("subscription resumed", zap.Uint("id", id))
	}

	return nil
}

// GetActiveSubscribes 获取所有活跃订阅
func (s *DefaultService) GetActiveSubscribes() ([]models.Subscribe, error) {
	ctx := context.Background()

	subscribes, err := s.repo.GetActiveSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为非指针切片
	result := make([]models.Subscribe, len(subscribes))
	for i, sub := range subscribes {
		result[i] = *sub
	}

	return result, nil
}
