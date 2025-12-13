// Package services MoviePilot业务服务层
package services

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/logger"
)

// SubscriptionScheduler 订阅任务调度服务接口
type SubscriptionScheduler interface {
	// Start 启动调度器
	Start(ctx context.Context) error
	// Stop 停止调度器
	Stop() error
	// ScheduleCheck 调度订阅检查
	ScheduleCheck(ctx context.Context, subscription *models.Subscription) error
	// CheckSubscription 检查单个订阅
	CheckSubscription(ctx context.Context, subscription *models.Subscription) error
	// CheckAllSubscriptions 检查所有订阅
	CheckAllSubscriptions(ctx context.Context) error
}

// SubscriptionSchedulerImpl 订阅任务调度服务实现
type SubscriptionSchedulerImpl struct {
	db          *gorm.DB
	logger      *zap.Logger
	mediaScanner MediaScanner
	running     bool
	mutex       sync.Mutex
	workers     int
	queue       chan *models.Subscription
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewSubscriptionScheduler 创建订阅任务调度服务实例
func NewSubscriptionScheduler(db *gorm.DB, mediaScanner MediaScanner) SubscriptionScheduler {
	return &SubscriptionSchedulerImpl{
		db:          db,
		logger:      logger.GetLogger(),
		mediaScanner: mediaScanner,
		running:     false,
		workers:     5,
		queue:       make(chan *models.Subscription, 100),
	}
}

// Start 启动调度器
func (s *SubscriptionSchedulerImpl) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return nil
	}

	// 创建上下文
	s.ctx, s.cancel = context.WithCancel(ctx)

	// 启动工作协程
	for i := 0; i < s.workers; i++ {
		go s.worker(i)
	}

	// 启动定时检查协程
	go s.scheduleRegularChecks()

	s.running = true
	s.logger.Info("订阅任务调度器已启动",
		zap.Int("workers", s.workers),
		zap.Int("queue_capacity", cap(s.queue)),
	)

	return nil
}

// Stop 停止调度器
func (s *SubscriptionSchedulerImpl) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	// 关闭队列
	close(s.queue)

	s.running = false
	s.logger.Info("订阅任务调度器已停止")

	return nil
}

// ScheduleCheck 调度订阅检查
func (s *SubscriptionSchedulerImpl) ScheduleCheck(ctx context.Context, subscription *models.Subscription) error {
	if !s.running {
		return nil
	}

	select {
	case s.queue <- subscription:
		s.logger.Debug("订阅检查已加入队列",
			zap.Uint("subscription_id", subscription.ID),
			zap.String("title", subscription.Title),
		)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Warn("订阅检查队列已满，丢弃检查任务",
			zap.Uint("subscription_id", subscription.ID),
			zap.String("title", subscription.Title),
		)
		return nil
	}
}

// CheckSubscription 检查单个订阅
func (s *SubscriptionSchedulerImpl) CheckSubscription(ctx context.Context, subscription *models.Subscription) error {
	if subscription.Status != models.SubscriptionStatusActive || !subscription.Monitor {
		s.logger.Debug("订阅未处于活跃状态或未启用监控，跳过检查",
			zap.Uint("subscription_id", subscription.ID),
			zap.String("status", string(subscription.Status)),
			zap.Bool("monitor", subscription.Monitor),
		)
		return nil
	}

	// 记录检查开始时间
	start := time.Now()
	s.logger.Info("开始检查订阅",
		zap.Uint("subscription_id", subscription.ID),
		zap.String("title", subscription.Title),
		zap.Int("tmdb_id", subscription.TMDBID),
	)

	// TODO: 实现订阅检查逻辑
	// 1. 查询TMDB获取最新信息
	// 2. 检查是否有新的季/集
	// 3. 如果有更新，创建下载任务
	// 4. 更新订阅状态

	// 更新上次检查时间
	now := time.Now()
	subscription.LastCheckAt = &now

	// 计算下次检查时间（默认30分钟后）
	nextCheck := now.Add(30 * time.Minute)
	subscription.NextCheckAt = &nextCheck

	// 保存更新
	if err := s.db.Save(subscription).Error; err != nil {
		s.logger.Error("更新订阅失败",
			zap.Uint("subscription_id", subscription.ID),
			zap.Error(err),
		)
		return err
	}

	// 记录检查完成时间
	duration := time.Since(start)
	s.logger.Info("订阅检查完成",
		zap.Uint("subscription_id", subscription.ID),
		zap.String("title", subscription.Title),
		zap.Duration("duration", duration),
	)

	return nil
}

// CheckAllSubscriptions 检查所有订阅
func (s *SubscriptionSchedulerImpl) CheckAllSubscriptions(ctx context.Context) error {
	// 获取所有需要检查的订阅
	var subscriptions []*models.Subscription
	if err := s.db.Where("status = ? AND monitor = ?", models.SubscriptionStatusActive, true).Find(&subscriptions).Error; err != nil {
		s.logger.Error("获取订阅列表失败",
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("开始检查所有订阅",
		zap.Int("count", len(subscriptions)),
	)

	// 并行检查所有订阅
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.workers)

	for _, subscription := range subscriptions {
		wg.Add(1)
		sem <- struct{}{}

		go func(sub *models.Subscription) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := s.CheckSubscription(ctx, sub); err != nil {
				s.logger.Error("检查订阅失败",
					zap.Uint("subscription_id", sub.ID),
					zap.String("title", sub.Title),
					zap.Error(err),
				)
			}
		}(subscription)
	}

	wg.Wait()
	s.logger.Info("所有订阅检查完成")

	return nil
}

// worker 工作协程
func (s *SubscriptionSchedulerImpl) worker(id int) {
	s.logger.Debug("订阅检查工作协程已启动",
		zap.Int("worker_id", id),
	)

	for {
		select {
		case subscription, ok := <-s.queue:
			if !ok {
				// 队列已关闭
				s.logger.Debug("订阅检查工作协程已退出",
					zap.Int("worker_id", id),
				)
				return
			}

			// 检查订阅
			if err := s.CheckSubscription(s.ctx, subscription); err != nil {
				s.logger.Error("检查订阅失败",
					zap.Int("worker_id", id),
					zap.Uint("subscription_id", subscription.ID),
					zap.Error(err),
				)
			}

		case <-s.ctx.Done():
			// 上下文已取消
			s.logger.Debug("订阅检查工作协程已退出",
				zap.Int("worker_id", id),
			)
			return
		}
	}
}

// scheduleRegularChecks 定期检查所有订阅
func (s *SubscriptionSchedulerImpl) scheduleRegularChecks() {
	// 创建定时触发器
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("订阅定期检查已启动",
		zap.Duration("interval", 30*time.Minute),
	)

	for {
		select {
		case <-ticker.C:
			// 定期检查所有订阅
			if err := s.CheckAllSubscriptions(s.ctx); err != nil {
				s.logger.Error("定期检查订阅失败",
					zap.Error(err),
				)
			}

		case <-s.ctx.Done():
			// 上下文已取消
			s.logger.Info("订阅定期检查已停止")
			return
		}
	}
}