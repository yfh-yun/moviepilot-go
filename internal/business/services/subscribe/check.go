package subscribe

import (
	"context"

	"go.uber.org/zap"
)

// Check 检查订阅，更新订阅信息
// 对应Python: check()
func (s *Service) Check(ctx context.Context) error {
	s.logger.Info("开始检查订阅")

	// 查询所有订阅
	subscribes, err := s.subscribeRepo.List(ctx, []string{})
	if err != nil {
		return err
	}

	// 遍历订阅
	for _, subscribe := range subscribes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		s.logger.Info("开始更新订阅元数据", zap.String("name", subscribe.Name))

		// 生成元数据
		meta := &MetaInfo{
			Name:        subscribe.Name,
			Year:        subscribe.Year,
			BeginSeason: subscribe.Season,
			Type:        subscribe.Type,
		}

		// 识别媒体信息
		mediaInfo, err := s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
			Meta:         meta,
			MediaType:    subscribe.Type,
			TMDBID:       subscribe.TMDBID,
			DoubanID:     subscribe.DoubanID,
			EpisodeGroup: subscribe.EpisodeGroup,
			Cache:        false,
		})

		if err != nil || mediaInfo == nil {
			s.logger.Warn("未识别到媒体信息",
				zap.String("name", subscribe.Name),
				zap.Any("tmdb_id", subscribe.TMDBID),
				zap.String("douban_id", subscribe.DoubanID),
			)
			continue
		}

		// 对于电视剧，获取当前季的总集数
		totalEpisode := subscribe.TotalEpisode
		lackEpisode := subscribe.LackEpisode

		if subscribe.Type == "tv" && mediaInfo.Seasons != nil {
			episodes, ok := mediaInfo.Seasons[subscribe.Season]
			if ok && len(episodes) > 0 && !subscribe.ManualTotalEpisode {
				newTotalEpisode := len(episodes)
				if newTotalEpisode != subscribe.TotalEpisode {
					// 总集数变化
					lackEpisode = subscribe.LackEpisode + (newTotalEpisode - subscribe.TotalEpisode)
					totalEpisode = newTotalEpisode

					s.logger.Info("总集数变化",
						zap.String("name", subscribe.Name),
						zap.Int("old_total", subscribe.TotalEpisode),
						zap.Int("new_total", totalEpisode),
						zap.Int("lack_episode", lackEpisode),
					)
				}
			}
		}

		// 更新TMDB信息
		updates := map[string]any{
			"name":          mediaInfo.Title,
			"year":          mediaInfo.Year,
			"vote":          mediaInfo.VoteAverage,
			"poster":        mediaInfo.PosterPath,
			"backdrop":      mediaInfo.BackdropPath,
			"description":   mediaInfo.Overview,
			"imdbid":        mediaInfo.IMDBID,
			"tvdbid":        mediaInfo.TVDBID,
			"total_episode": totalEpisode,
			"lack_episode":  lackEpisode,
		}

		if err := s.subscribeRepo.Update(ctx, subscribe.ID, updates); err != nil {
			s.logger.Error("更新订阅元数据失败",
				zap.String("name", subscribe.Name),
				zap.Error(err),
			)
			continue
		}

		s.logger.Info("订阅元数据更新完成", zap.String("name", subscribe.Name))
	}

	s.logger.Info("订阅检查完成")
	return nil
}
