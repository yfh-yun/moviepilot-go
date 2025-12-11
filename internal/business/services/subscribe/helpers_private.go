package subscribe

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// updateSubscribeNote 更新已下载信息到note字段
// 对应Python: __update_subscribe_note()
func (s *Service) updateSubscribeNote(ctx context.Context, subscribe *Subscribe, downloads []*Context) error {
	// 查询现有Note
	if len(downloads) == 0 {
		return nil
	}

	note := []int{}
	if len(subscribe.Note) > 0 {
		note = subscribe.Note
	}

	for _, context := range downloads {
		meta := context.MetaInfo
		mediaInfo := context.MediaInfo

		if subscribe.TMDBID != nil && mediaInfo.TMDBID != 0 &&
			mediaInfo.TMDBID != *subscribe.TMDBID {
			continue
		}
		if subscribe.DoubanID != "" && mediaInfo.DoubanID != "" &&
			mediaInfo.DoubanID != subscribe.DoubanID {
			continue
		}

		items := []int{}
		if mediaInfo.Type == "tv" {
			// 电视剧有集数，使用 episode_list
			items = meta.EpisodeList
		} else if mediaInfo.Type == "movie" {
			// 电影只有一个条目，设置为 [1]
			items = []int{1}
		}

		if len(items) == 0 {
			continue
		}

		// 合并已下载的集数或电影项（去重）
		itemMap := make(map[int]bool)
		for _, item := range note {
			itemMap[item] = true
		}
		for _, item := range items {
			if !itemMap[item] {
				note = append(note, item)
				itemMap[item] = true
			}
		}
	}

	// 更新订阅
	if len(note) > 0 {
		return s.subscribeRepo.Update(ctx, subscribe.ID, map[string]any{
			"note": note,
		})
	}

	return nil
}

// updateLackEpisodes 更新订阅剩余集数及时间
// 对应Python: __update_lack_episodes()
func (s *Service) updateLackEpisodes(
	ctx context.Context,
	lefts map[any]map[int]*NotExistMediaInfo,
	subscribe *Subscribe,
	mediaInfo *MediaInfo,
	updateDate bool,
) error {
	updateData := make(map[string]any)

	if updateDate {
		updateData["last_update"] = time.Now().Format("2006-01-02 15:04:05")
	}

	if subscribe.Type == "tv" {
		lackEpisode := 0

		if len(lefts) == 0 {
			// 如果 lefts 为空，表示没有缺失集数，直接设置 lack_episode 为 0
			s.logger.Info("没有缺失集数，直接更新为 0",
				zap.String("title", mediaInfo.TitleYear),
			)
		} else {
			var mediaKey any
			if subscribe.TMDBID != nil && *subscribe.TMDBID != 0 {
				mediaKey = *subscribe.TMDBID
			} else if subscribe.DoubanID != "" {
				mediaKey = subscribe.DoubanID
			}

			if mediaKey != nil {
				leftSeasons := lefts[mediaKey]
				for _, seasonInfo := range leftSeasons {
					season := seasonInfo.Season
					if season == subscribe.Season {
						leftEpisodes := seasonInfo.Episodes
						if len(leftEpisodes) == 0 {
							lackEpisode = seasonInfo.TotalEpisode
						} else {
							lackEpisode = len(leftEpisodes)
						}
						s.logger.Info("更新缺失集数",
							zap.String("title", mediaInfo.TitleYear),
							zap.Int("season", season),
							zap.Int("lack_episode", lackEpisode),
						)
						break
					}
				}
			}
		}

		updateData["lack_episode"] = lackEpisode
	}

	// 更新数据库
	if len(updateData) > 0 {
		return s.subscribeRepo.Update(ctx, subscribe.ID, updateData)
	}

	return nil
}

// finishSubscribe 完成订阅
// 对应Python: __finish_subscribe()
func (s *Service) finishSubscribe(
	ctx context.Context,
	subscribe *Subscribe,
	meta *MetaInfo,
	mediaInfo *MediaInfo,
) error {
	// 如果订阅状态为待定（P），说明订阅信息尚未完全更新，无法完成订阅
	if subscribe.State == "P" {
		return nil
	}

	// 完成订阅
	msgstr := "订阅"
	if subscribe.BestVersion {
		msgstr = "洗版"
	}

	s.logger.Info(fmt.Sprintf("完成%s", msgstr),
		zap.String("title", mediaInfo.TitleYear),
	)

	// 新增订阅历史
	if err := s.subscribeRepo.AddHistory(ctx, subscribe); err != nil {
		s.logger.Error("添加订阅历史失败", zap.Error(err))
	}

	// 删除订阅
	if err := s.subscribeRepo.Delete(ctx, subscribe.ID); err != nil {
		return err
	}

	// 发送通知
	link := "#/subscribe/tv?tab=mysub"
	if mediaInfo.Type == "movie" {
		link = "#/subscribe/movie?tab=mysub"
	}

	// 完成订阅按规则发送消息
	notification := &Notification{
		MType:    "Subscribe",
		CType:    "SubscribeComplete",
		Image:    mediaInfo.PosterPath,
		Link:     link,
		Username: subscribe.Username,
	}

	extra := map[string]any{
		"msgstr":   msgstr,
		"username": subscribe.Username,
	}

	if err := s.messageService.PostMessage(ctx, notification, meta, mediaInfo, extra); err != nil {
		s.logger.Error("发送完成消息失败", zap.Error(err))
	}

	// 发送事件
	s.eventService.SendEventAsync(ctx, "SubscribeComplete", map[string]any{
		"subscribe_id":   subscribe.ID,
		"subscribe_info": subscribe,
		"mediainfo":      mediaInfo,
	})

	// 统计订阅
	data := map[string]any{
		"tmdbid":   mediaInfo.TMDBID,
		"doubanid": mediaInfo.DoubanID,
	}

	if err := s.subscribeHelper.SubDoneAsync(ctx, data); err != nil {
		s.logger.Error("统计订阅失败", zap.Error(err))
	}

	return nil
}
