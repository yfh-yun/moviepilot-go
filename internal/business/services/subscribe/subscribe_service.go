package subscribe

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// SubscribeService 订阅服务
// 原SubscribeChain，负责订阅管理
type SubscribeService struct {
	*base.ServiceBase
	repo interfaces.SubscribeRepository
}

// NewSubscribeService 创建SubscribeService实例
func NewSubscribeService(repo interfaces.SubscribeRepository) *SubscribeService {
	return &SubscribeService{
		ServiceBase: base.NewServiceBase(),
		repo:        repo,
	}
}

// Initialize 初始化服务
func (s *SubscribeService) Initialize() error {
	logger.Info("Initializing SubscribeService")
	return nil
}

// Name 获取服务名称
func (s *SubscribeService) Name() string {
	return "SubscribeService"
}

// Close 关闭服务
func (s *SubscribeService) Close() error {
	logger.Info("Closing SubscribeService")
	return nil
}

// CreateSubscribe 创建订阅
func (s *SubscribeService) CreateSubscribe(ctx context.Context, req *dto.AddSubscribeRequest) (*database.Subscribe, error) {
	logger.Info("Creating subscribe",
		zap.String("title", req.Title),
		zap.String("type", req.MediaType))

	// 检查是否已存在相同的订阅
	exists, err := s.repo.Exists(ctx, req.TMDBID, &req.DoubanID, req.Season)
	if err != nil {
		logger.Error("Failed to check subscribe existence", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("订阅已存在")
	}

	// 创建订阅
	now := time.Now()
	subscribe := &database.Subscribe{
		Name:       req.Title,
		Type:       req.MediaType,
		TMDBID:     req.TMDBID,
		DoubanID:   &req.DoubanID,
		Season:     req.Season,
		State:      "R", // R表示运行中
		Username:   req.Username,
		LastUpdate: &now,
		Date:       now.Format("2006-01-02 15:04:05"),
	}

	if err := s.repo.Create(ctx, subscribe); err != nil {
		logger.Error("Failed to create subscribe", zap.Error(err))
		return nil, err
	}

	logger.Info("Subscribe created successfully", zap.Uint("id", subscribe.ID))
	return subscribe, nil
}

// GetSubscribeByID 根据ID获取订阅
func (s *SubscribeService) GetSubscribeByID(ctx context.Context, id string) (*database.Subscribe, error) {
	logger.Debug("Getting subscribe by ID", zap.String("id", id))

	subscribe, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get subscribe", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	if subscribe == nil {
		return nil, fmt.Errorf("订阅不存在")
	}

	return subscribe, nil
}

// UpdateSubscribe 更新订阅
func (s *SubscribeService) UpdateSubscribe(ctx context.Context, id string, req *dto.AddSubscribeRequest) (*database.Subscribe, error) {
	logger.Info("Updating subscribe", zap.String("id", id))

	// 获取现有订阅
	subscribe, err := s.GetSubscribeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	subscribe.Name = req.Title
	subscribe.Type = req.MediaType
	subscribe.TMDBID = req.TMDBID
	subscribe.DoubanID = &req.DoubanID
	subscribe.Season = req.Season
	now := time.Now()
	subscribe.LastUpdate = &now

	if err := s.repo.Update(ctx, subscribe); err != nil {
		logger.Error("Failed to update subscribe", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	logger.Info("Subscribe updated successfully", zap.String("id", id))
	return subscribe, nil
}

// DeleteSubscribe 删除订阅
func (s *SubscribeService) DeleteSubscribe(ctx context.Context, id string) error {
	logger.Info("Deleting subscribe", zap.String("id", id))

	// 检查订阅是否存在
	_, err := s.GetSubscribeByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete subscribe", zap.String("id", id), zap.Error(err))
		return err
	}

	logger.Info("Subscribe deleted successfully", zap.String("id", id))
	return nil
}

// ListSubscribes 获取订阅列表
func (s *SubscribeService) ListSubscribes(ctx context.Context, page, pageSize int, status, mediaType, userID string) ([]*database.Subscribe, int64, error) {
	logger.Debug("Listing subscribes",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("status", status),
		zap.String("type", mediaType))

	repoParams := interfaces.ListSubscribeParams{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Type:     mediaType,
		UserID:   userID,
	}

	subscribes, total, err := s.repo.List(ctx, repoParams)
	if err != nil {
		logger.Error("Failed to list subscribes", zap.Error(err))
		return nil, 0, err
	}

	return subscribes, total, nil
}

// GetActiveSubscribes 获取活跃订阅
func (s *SubscribeService) GetActiveSubscribes(ctx context.Context) ([]*database.Subscribe, error) {
	logger.Debug("Getting active subscribes")

	subscribes, err := s.repo.ListActive(ctx)
	if err != nil {
		logger.Error("Failed to get active subscribes", zap.Error(err))
		return nil, err
	}

	return subscribes, nil
}

// UpdateSubscribeState 更新订阅状态
func (s *SubscribeService) UpdateSubscribeState(ctx context.Context, id uint, state string) error {
	logger.Info("Updating subscribe state", zap.Uint("id", id), zap.String("state", state))

	if err := s.repo.UpdateState(ctx, id, state); err != nil {
		logger.Error("Failed to update subscribe state", zap.Uint("id", id), zap.Error(err))
		return err
	}

	return nil
}

// GetSubscribeStatistics 获取订阅统计信息
func (s *SubscribeService) GetSubscribeStatistics(ctx context.Context) (map[string]int64, error) {
	logger.Debug("Getting subscribe statistics")

	stats, err := s.repo.Statistics(ctx)
	if err != nil {
		logger.Error("Failed to get subscribe statistics", zap.Error(err))
		return nil, err
	}

	return stats, nil
}

// RefreshSubscribe 刷新订阅
func (s *SubscribeService) RefreshSubscribe(ctx context.Context, id string) error {
	logger.Info("Refreshing subscribe", zap.String("id", id))

	// TODO: 实现订阅刷新逻辑
	// 1. 获取订阅信息
	// 2. 调用搜索服务查找最新资源
	// 3. 根据规则过滤资源
	// 4. 触发下载

	return fmt.Errorf("not implemented yet")
}

// RefreshAllSubscribes 刷新所有订阅
func (s *SubscribeService) RefreshAllSubscribes(ctx context.Context) error {
	logger.Info("Refreshing all subscribes")

	// 获取所有活跃订阅
	subscribes, err := s.GetActiveSubscribes(ctx)
	if err != nil {
		return err
	}

	logger.Info("Found active subscribes", zap.Int("count", len(subscribes)))

	// TODO: 实现批量刷新逻辑
	// 可以使用goroutine并发刷新

	return fmt.Errorf("not implemented yet")
}
