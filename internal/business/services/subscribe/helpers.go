package subscribe

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// GetSubSites 获取订阅站点
// 对应Python: get_sub_sites()
func (s *Service) GetSubSites(ctx context.Context, subscribe *Subscribe) ([]int, error) {
	// 从系统配置获取默认订阅站点
	defaultSites := []int{}
	if val, err := s.systemConfigRepo.Get(ctx, "RssSites"); err == nil && val != nil {
		if sites, ok := val.([]int); ok {
			defaultSites = sites
		} else if sites, ok := val.([]any); ok {
			for _, site := range sites {
				if siteID, ok := site.(int); ok {
					defaultSites = append(defaultSites, siteID)
				}
			}
		}
	}

	// 如果订阅未指定站点，直接返回默认站点
	if len(subscribe.Sites) == 0 {
		return defaultSites, nil
	}

	// 如果默认订阅站点未设置，直接返回订阅指定站点
	if len(defaultSites) == 0 {
		return subscribe.Sites, nil
	}

	// 计算交集
	intersectionSites := []int{}
	siteMap := make(map[int]bool)
	for _, site := range defaultSites {
		siteMap[site] = true
	}

	for _, site := range subscribe.Sites {
		if siteMap[site] {
			intersectionSites = append(intersectionSites, site)
		}
	}

	// 如果交集为空，返回默认站点
	if len(intersectionSites) == 0 {
		return defaultSites, nil
	}

	return intersectionSites, nil
}

// GetSubscribedSites 获取所有订阅站点
// 对应Python: get_subscribed_sites()
func (s *Service) GetSubscribedSites(ctx context.Context) ([]int, error) {
	retSites := []int{}
	siteMap := make(map[int]bool)

	// 获取所有订阅
	subscribes, err := s.subscribeRepo.List(ctx, s.GetStatesForSearch("R"))
	if err != nil {
		return nil, err
	}

	if len(subscribes) == 0 {
		// 没有订阅
		return nil, nil
	}

	// 收集所有订阅的站点
	for _, subscribe := range subscribes {
		sites, err := s.GetSubSites(ctx, subscribe)
		if err != nil {
			continue
		}
		for _, site := range sites {
			if !siteMap[site] {
				siteMap[site] = true
				retSites = append(retSites, site)
			}
		}
	}

	return retSites, nil
}

// CheckAndHandleExistingMedia 检查媒体是否存在
// 对应Python: check_and_handle_existing_media()
func (s *Service) CheckAndHandleExistingMedia(
	ctx context.Context,
	subscribe *Subscribe,
	meta *MetaInfo,
	mediaInfo *MediaInfo,
	mediaKey any,
) (bool, map[any]map[int]*NotExistMediaInfo, error) {
	var existFlag bool
	var noExists map[any]map[int]*NotExistMediaInfo

	// 非洗版
	if !subscribe.BestVersion {
		// 每季总集数
		totals := make(map[int]int)
		if subscribe.Season > 0 && subscribe.TotalEpisode > 0 {
			totals[subscribe.Season] = subscribe.TotalEpisode
		}

		// 查询媒体库缺失的媒体信息
		result, err := s.downloadService.GetNoExistsInfo(ctx, &GetNoExistsInfoRequest{
			Meta:      meta,
			MediaInfo: mediaInfo,
			Totals:    totals,
		})
		if err != nil {
			return false, nil, err
		}

		existFlag = result.ExistFlag
		noExists = result.NoExists
	} else {
		// 洗版，如果已经满足了优先级，则认为已经洗版完成
		if subscribe.CurrentPriority != nil && *subscribe.CurrentPriority == 100 {
			existFlag = true
			noExists = make(map[any]map[int]*NotExistMediaInfo)
		} else {
			existFlag = false
			if meta.Type == "tv" {
				// 对于电视剧，构造缺失的媒体信息
				noExists = map[any]map[int]*NotExistMediaInfo{
					mediaKey: {
						subscribe.Season: {
							Season:       subscribe.Season,
							Episodes:     []int{},
							TotalEpisode: subscribe.TotalEpisode,
							StartEpisode: subscribe.StartEpisode,
						},
					},
				}
			} else {
				noExists = make(map[any]map[int]*NotExistMediaInfo)
			}
		}
	}

	// 如果媒体已存在，执行订阅完成操作
	if existFlag {
		if !subscribe.BestVersion {
			s.logger.Info("媒体库中已存在", zap.String("title", mediaInfo.TitleYear))
		}
		return true, noExists, s.FinishSubscribeOrNot(ctx, subscribe, meta, mediaInfo, nil, noExists, true)
	}

	// 获取已下载的集数或电影
	downloaded := s.getDownloaded(subscribe)
	if meta.Type == "tv" {
		// 对于电视剧类型，整合缺失集数并剔除已下载的集数
		existFlag, noExists = s.getSubscribeNoExits(
			fmt.Sprintf("%s %d", subscribe.Name, meta.Season),
			noExists,
			mediaKey,
			meta.BeginSeason,
			subscribe.TotalEpisode,
			subscribe.StartEpisode,
			downloaded,
		)
	} else if meta.Type == "movie" {
		// 对于电影类型，直接根据是否已下载判断
		existFlag = len(downloaded) > 0
	}

	// 如果已下载完毕，执行订阅完成操作
	if existFlag {
		s.logger.Info("已全部下载", zap.String("title", mediaInfo.TitleYear))
		return true, noExists, s.FinishSubscribeOrNot(ctx, subscribe, meta, mediaInfo, nil, noExists, true)
	}

	// 返回结果，表示媒体未完全下载或存在
	return false, noExists, nil
}

// FinishSubscribeOrNot 判断是否完成订阅
// 对应Python: finish_subscribe_or_not()
func (s *Service) FinishSubscribeOrNot(
	ctx context.Context,
	subscribe *Subscribe,
	meta *MetaInfo,
	mediaInfo *MediaInfo,
	downloads []*Context,
	lefts map[any]map[int]*NotExistMediaInfo,
	force bool,
) error {
	var mediaKey any
	if subscribe.TMDBID != nil && *subscribe.TMDBID != 0 {
		mediaKey = *subscribe.TMDBID
	} else if subscribe.DoubanID != "" {
		mediaKey = subscribe.DoubanID
	}

	// 是否有剩余集
	noLefts := len(lefts) == 0
	if !noLefts && mediaKey != nil {
		if _, ok := lefts[mediaKey]; !ok {
			noLefts = true
		}
	}

	// 是否完成订阅
	if !subscribe.BestVersion {
		// 订阅存在待定策略，不管是否已完成，均需更新订阅信息
		// 更新订阅已下载信息
		if err := s.updateSubscribeNote(ctx, subscribe, downloads); err != nil {
			s.logger.Error("更新订阅note失败", zap.Error(err))
		}

		// 更新订阅剩余集数和时间
		if err := s.updateLackEpisodes(ctx, lefts, subscribe, mediaInfo, len(downloads) > 0); err != nil {
			s.logger.Error("更新缺失集数失败", zap.Error(err))
		}

		// 判断是否需要完成订阅
		if (noLefts && meta.Type == "tv") ||
			(len(downloads) > 0 && meta.Type == "movie") ||
			force {
			return s.finishSubscribe(ctx, subscribe, meta, mediaInfo)
		} else {
			// 未下载到内容且不完整
			s.logger.Info("未下载完整，继续订阅",
				zap.String("title", mediaInfo.TitleYear),
			)
		}
	} else if len(downloads) > 0 {
		// 洗版下载到了内容，更新资源优先级
		return s.UpdateSubscribePriority(ctx, subscribe, meta, mediaInfo, downloads)
	} else if subscribe.CurrentPriority != nil && *subscribe.CurrentPriority == 100 {
		// 洗版完成
		return s.finishSubscribe(ctx, subscribe, meta, mediaInfo)
	} else {
		// 洗版，未下载到内容
		s.logger.Info("继续洗版",
			zap.String("title", mediaInfo.TitleYear),
		)
	}

	return nil
}

// UpdateSubscribePriority 更新订阅优先级
// 对应Python: update_subscribe_priority()
func (s *Service) UpdateSubscribePriority(
	ctx context.Context,
	subscribe *Subscribe,
	meta *MetaInfo,
	mediaInfo *MediaInfo,
	downloads []*Context,
) error {
	if len(downloads) == 0 {
		return nil
	}

	if !subscribe.BestVersion {
		return nil
	}

	// 当前下载资源的优先级
	priority := 0
	for _, item := range downloads {
		if item.TorrentInfo.PriOrder > priority {
			priority = item.TorrentInfo.PriOrder
		}
	}

	// 订阅存在待定策略，不管是否已完成，均需更新订阅信息
	updates := map[string]any{
		"current_priority": priority,
		"last_update":      fmt.Sprintf("%v", ctx.Value("now")),
	}

	if err := s.subscribeRepo.Update(ctx, subscribe.ID, updates); err != nil {
		return err
	}

	if priority == 100 {
		// 洗版完成
		return s.finishSubscribe(ctx, subscribe, meta, mediaInfo)
	} else {
		// 正在洗版，更新资源优先级
		s.logger.Info("正在洗版，更新资源优先级",
			zap.String("title", mediaInfo.TitleYear),
			zap.Int("priority", priority),
		)
	}

	return nil
}

// getDownloaded 获取已下载集数
// 对应Python: __get_downloaded()
func (s *Service) getDownloaded(subscribe *Subscribe) []int {
	if subscribe.BestVersion {
		return []int{}
	}

	note := subscribe.Note
	if len(note) == 0 {
		return []int{}
	}

	// 针对 TV 类型，返回已下载的集数
	if subscribe.Type == "tv" {
		s.logger.Info("已下载集数",
			zap.String("name", subscribe.Name),
			zap.Int("season", subscribe.Season),
			zap.Ints("episodes", note),
		)
		return note
	}

	// 针对 Movie 类型，直接返回已下载的电影
	if subscribe.Type == "movie" {
		s.logger.Info("已下载内容",
			zap.String("name", subscribe.Name),
			zap.Ints("note", note),
		)
		return note
	}

	return []int{}
}

// getSubscribeNoExits 计算订阅缺失集数
// 对应Python: __get_subscribe_no_exits()
func (s *Service) getSubscribeNoExits(
	subscribeName string,
	noExists map[any]map[int]*NotExistMediaInfo,
	mediaKey any,
	beginSeason int,
	totalEpisode int,
	startEpisode int,
	downloadedEpisodes []int,
) (bool, map[any]map[int]*NotExistMediaInfo) {
	// 使用订阅的总集数和开始集数替换no_exists
	if noExists == nil || noExists[mediaKey] == nil {
		return false, noExists
	}

	noExistsItem := noExists[mediaKey]

	if totalEpisode > 0 || startEpisode > 0 {
		s.logger.Info("订阅设定",
			zap.String("name", subscribeName),
			zap.Int("start_episode", startEpisode),
			zap.Int("total_episode", totalEpisode),
		)

		// 该季原缺失信息
		noExistSeason := noExistsItem[beginSeason]
		if noExistSeason != nil {
			// 原集列表
			episodeList := noExistSeason.Episodes
			// 原总集数
			total := noExistSeason.TotalEpisode
			// 原开始集数
			start := noExistSeason.StartEpisode

			// 更新剧集列表、开始集数、总集数
			var episodes []int
			if len(episodeList) == 0 {
				// 整季缺失
				episodes = []int{}
				if startEpisode == 0 {
					startEpisode = start
				}
				if totalEpisode == 0 {
					totalEpisode = total
				}
			} else {
				// 部分缺失
				if startEpisode == 0 && totalEpisode == 0 {
					// 无需调整
					return false, noExists
				}
				if startEpisode == 0 {
					startEpisode = start
				}
				if totalEpisode == 0 {
					totalEpisode = total
				}

				// 新的集列表
				newEpisodes := []int{}
				for i := max(startEpisode, start); i <= totalEpisode; i++ {
					newEpisodes = append(newEpisodes, i)
				}

				// 与原集列表取交集
				episodeMap := make(map[int]bool)
				for _, ep := range episodeList {
					episodeMap[ep] = true
				}

				for _, ep := range newEpisodes {
					if episodeMap[ep] {
						episodes = append(episodes, ep)
					}
				}

				// 交集为空时，说明订阅的剧集均已入库
				if len(episodes) == 0 {
					return true, make(map[any]map[int]*NotExistMediaInfo)
				}
			}

			// 更新集合
			noExists[mediaKey][beginSeason] = &NotExistMediaInfo{
				Season:       beginSeason,
				Episodes:     episodes,
				TotalEpisode: totalEpisode,
				StartEpisode: startEpisode,
			}
		}
	}

	// 根据订阅已下载集数更新缺失集数
	if len(downloadedEpisodes) > 0 {
		s.logger.Info("已下载集数",
			zap.String("name", subscribeName),
			zap.Ints("episodes", downloadedEpisodes),
		)

		// 该季原缺失信息
		noExistSeason := noExistsItem[beginSeason]
		if noExistSeason != nil {
			// 原集列表
			episodeList := noExistSeason.Episodes
			// 原总集数
			total := noExistSeason.TotalEpisode
			// 原开始集数
			start := noExistSeason.StartEpisode

			// 整季缺失
			if len(episodeList) == 0 {
				episodeList = []int{}
				for i := start; i <= total; i++ {
					episodeList = append(episodeList, i)
				}
			}

			// 更新剧集列表
			downloadedMap := make(map[int]bool)
			for _, ep := range downloadedEpisodes {
				downloadedMap[ep] = true
			}

			episodes := []int{}
			for _, ep := range episodeList {
				if !downloadedMap[ep] {
					episodes = append(episodes, ep)
				}
			}

			// 如果存在已下载剧集，则差集为空时，说明所有均已存在
			if len(episodes) == 0 {
				return true, make(map[any]map[int]*NotExistMediaInfo)
			}

			// 更新集合
			noExists[mediaKey][beginSeason] = &NotExistMediaInfo{
				Season:       beginSeason,
				Episodes:     episodes,
				TotalEpisode: total,
				StartEpisode: start,
			}
		} else {
			// 开始集数
			if startEpisode == 0 {
				startEpisode = 1
			}

			// 更新剧集列表
			downloadedMap := make(map[int]bool)
			for _, ep := range downloadedEpisodes {
				downloadedMap[ep] = true
			}

			episodes := []int{}
			for i := startEpisode; i <= totalEpisode; i++ {
				if !downloadedMap[i] {
					episodes = append(episodes, i)
				}
			}

			// 如果存在已下载剧集，则差集为空时，说明所有均已存在
			if len(episodes) == 0 {
				return true, make(map[any]map[int]*NotExistMediaInfo)
			}

			noExists[mediaKey][beginSeason] = &NotExistMediaInfo{
				Season:       beginSeason,
				Episodes:     episodes,
				TotalEpisode: totalEpisode,
				StartEpisode: startEpisode,
			}
		}
	}

	s.logger.Info("缺失剧集数更新",
		zap.String("name", subscribeName),
		zap.Any("no_exists", noExists),
	)

	return false, noExists
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
