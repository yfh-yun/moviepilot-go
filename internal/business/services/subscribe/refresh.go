package subscribe

import (
	"context"

	"go.uber.org/zap"
)

// Refresh 刷新订阅
// 对应Python: refresh()
func (s *Service) Refresh(ctx context.Context) error {
	s.logger.Info("开始刷新订阅")

	// 获取订阅中涉及的所有站点
	sites, err := s.GetSubscribedSites(ctx)
	if err != nil {
		return err
	}

	if sites == nil {
		s.logger.Info("没有订阅")
		return nil
	}

	s.logger.Info("刷新站点资源", zap.Ints("sites", sites))

	// 触发刷新站点资源
	torrents, err := s.torrentsService.Refresh(ctx, sites)
	if err != nil {
		return err
	}

	// 从缓存中匹配订阅
	return s.Match(ctx, &MatchSubscribeRequest{
		Torrents: torrents,
	})
}
