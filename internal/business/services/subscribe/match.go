package subscribe

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Match 从缓存中匹配订阅
// 对应Python: match()
func (s *Service) Match(ctx context.Context, req *MatchSubscribeRequest) error {
	if len(req.Torrents) == 0 {
		s.logger.Warn("没有缓存资源，无法匹配订阅")
		return nil
	}

	// 获取锁
	lockAcquired := false
	lockChan := make(chan bool, 1)

	go func() {
		s.matchLock.Lock()
		lockChan <- true
	}()

	select {
	case <-lockChan:
		lockAcquired = true
		defer s.matchLock.Unlock()
		s.logger.Debug("match lock acquired", zap.Time("time", time.Now()))
	case <-time.After(LockTimeout):
		s.logger.Warn("match上锁超时")
		return fmt.Errorf("match上锁超时")
	}

	if !lockAcquired {
		return fmt.Errorf("获取锁失败")
	}

	// 预识别所有未识别的种子
	processedTorrents := s.preRecognizeTorrents(ctx, req.Torrents)

	// 所有订阅
	subscribes, err := s.subscribeRepo.List(ctx, s.GetStatesForSearch("R"))
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

		s.logger.Info("开始匹配订阅", zap.String("name", subscribe.Name))

		var mediaKey any
		if subscribe.TMDBID != nil && *subscribe.TMDBID != 0 {
			mediaKey = *subscribe.TMDBID
		} else if subscribe.DoubanID != "" {
			mediaKey = subscribe.DoubanID
		}

		// 生成元数据
		meta := &MetaInfo{
			Name:        subscribe.Name,
			Year:        subscribe.Year,
			BeginSeason: subscribe.Season,
			Type:        subscribe.Type,
		}

		// 订阅的站点域名列表
		var domains []string
		if len(subscribe.Sites) > 0 {
			domains, err = s.siteRepo.GetDomainsByIDs(ctx, subscribe.Sites)
			if err != nil {
				s.logger.Error("获取站点域名失败", zap.Error(err))
			}
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
			)
			continue
		}

		// 检查媒体是否存在
		existFlag, noExists, err := s.CheckAndHandleExistingMedia(ctx, subscribe, meta, mediaInfo, mediaKey)
		if err != nil {
			s.logger.Error("检查媒体存在失败", zap.Error(err))
			continue
		}
		if existFlag {
			continue
		}

		// 订阅识别词
		var customWordsList []string
		if subscribe.CustomWords != "" {
			customWordsList = []string{subscribe.CustomWords}
		}

		// 遍历预识别后的种子
		matchContexts := s.matchTorrentsWithSubscribe(
			ctx,
			subscribe,
			meta,
			mediaInfo,
			noExists,
			processedTorrents,
			domains,
			customWordsList,
		)

		if len(matchContexts) == 0 {
			// 未匹配到资源
			s.logger.Info("未匹配到符合条件的资源",
				zap.String("title", mediaInfo.TitleYear),
			)
			s.FinishSubscribeOrNot(ctx, subscribe, meta, mediaInfo, nil, noExists, false)
			continue
		}

		// 开始批量择优下载
		s.logger.Info("匹配完成",
			zap.String("title", mediaInfo.TitleYear),
			zap.Int("count", len(matchContexts)),
		)

		downloads, lefts, err := s.batchDownload(ctx, subscribe, matchContexts, noExists)
		if err != nil {
			s.logger.Error("批量下载失败", zap.Error(err))
		}

		// 同步外部修改，更新订阅信息
		subscribe, err = s.subscribeRepo.Get(ctx, subscribe.ID)
		if err != nil {
			s.logger.Error("获取订阅失败", zap.Error(err))
			continue
		}

		// 判断是否要完成订阅
		if subscribe != nil {
			s.FinishSubscribeOrNot(ctx, subscribe, meta, mediaInfo, downloads, lefts, false)
		}
	}

	return nil
}

// preRecognizeTorrents 预识别所有未识别的种子
func (s *Service) preRecognizeTorrents(
	ctx context.Context,
	torrents map[string][]*Context,
) map[string][]*Context {
	processedTorrents := make(map[string][]*Context)

	for domain, contexts := range torrents {
		select {
		case <-ctx.Done():
			return processedTorrents
		default:
		}

		processedTorrents[domain] = []*Context{}

		for _, context := range contexts {
			// 如果种子未识别且失败次数未超过3次，尝试识别
			if (context.MediaInfo == nil ||
				(context.MediaInfo.TMDBID == 0 && context.MediaInfo.DoubanID == "")) &&
				context.MediaRecognizeFailCount < 3 {

				s.logger.Debug("尝试重新识别种子",
					zap.String("title", context.TorrentInfo.Title),
					zap.Int("fail_count", context.MediaRecognizeFailCount),
				)

				reMediaInfo, err := s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
					Meta: context.MetaInfo,
				})

				if err == nil && reMediaInfo != nil {
					// 更新种子缓存
					context.MediaInfo = reMediaInfo
					// 重置失败次数
					context.MediaRecognizeFailCount = 0
					s.logger.Debug("种子重新识别成功",
						zap.String("title", context.TorrentInfo.Title),
					)
				} else {
					// 识别失败，增加失败次数
					context.MediaRecognizeFailCount++
					s.logger.Debug("种子媒体识别失败",
						zap.String("title", context.TorrentInfo.Title),
						zap.Int("fail_count", context.MediaRecognizeFailCount),
					)
				}
			} else if context.MediaRecognizeFailCount >= 3 {
				s.logger.Debug("种子已达到最大识别失败次数(3次)，跳过识别",
					zap.String("title", context.TorrentInfo.Title),
				)
			}

			// 添加已预处理
			processedTorrents[domain] = append(processedTorrents[domain], context)
		}
	}

	return processedTorrents
}

// matchTorrentsWithSubscribe 匹配种子与订阅
func (s *Service) matchTorrentsWithSubscribe(
	ctx context.Context,
	subscribe *Subscribe,
	meta *MetaInfo,
	mediaInfo *MediaInfo,
	noExists map[any]map[int]*NotExistMediaInfo,
	processedTorrents map[string][]*Context,
	domains []string,
	customWordsList []string,
) []*Context {
	matchContexts := []*Context{}

	// 获取订阅站点
	subSites, err := s.GetSubSites(ctx, subscribe)
	if err != nil {
		s.logger.Error("获取订阅站点失败", zap.Error(err))
		return matchContexts
	}

	// 优先级过滤规则
	var ruleGroups []string
	if subscribe.BestVersion {
		if len(subscribe.FilterGroups) > 0 {
			ruleGroups = subscribe.FilterGroups
		} else {
			if val, err := s.systemConfigRepo.Get(ctx, "BestVersionFilterRuleGroups"); err == nil && val != nil {
				if groups, ok := val.([]string); ok {
					ruleGroups = groups
				}
			}
		}
	} else {
		if len(subscribe.FilterGroups) > 0 {
			ruleGroups = subscribe.FilterGroups
		} else {
			if val, err := s.systemConfigRepo.Get(ctx, "SubscribeFilterRuleGroups"); err == nil && val != nil {
				if groups, ok := val.([]string); ok {
					ruleGroups = groups
				}
			}
		}
	}

	for domain, contexts := range processedTorrents {
		select {
		case <-ctx.Done():
			return matchContexts
		default:
		}

		if len(domains) > 0 && !contains(domains, domain) {
			continue
		}

		s.logger.Debug("开始匹配站点",
			zap.String("domain", domain),
			zap.Int("count", len(contexts)),
		)

		for _, context := range contexts {
			torrentMeta := context.MetaInfo
			torrentMediaInfo := context.MediaInfo
			torrentInfo := context.TorrentInfo

			// 不在订阅站点范围的不处理
			if len(subSites) > 0 && !contains(subSites, torrentInfo.Site) {
				s.logger.Debug("不符合订阅站点要求",
					zap.String("site", torrentInfo.SiteName),
					zap.String("title", torrentInfo.Title),
				)
				continue
			}

			// 有自定义识别词时，需要判断是否需要重新识别
			if len(customWordsList) > 0 {
				// TODO: 应用自定义识别词
			}

			// 如果仍然没有识别到媒体信息，尝试标题匹配
			if torrentMediaInfo == nil ||
				(torrentMediaInfo.TMDBID == 0 && torrentMediaInfo.DoubanID == "") {
				s.logger.Debug("重新识别失败，尝试通过标题匹配",
					zap.String("site", torrentInfo.SiteName),
					zap.String("title", torrentInfo.Title),
				)

				if s.torrentHelper.MatchTorrent(mediaInfo, torrentMeta, torrentInfo) {
					// 匹配成功
					s.logger.Info("通过标题匹配到可选资源",
						zap.String("title", mediaInfo.TitleYear),
						zap.String("torrent", torrentInfo.Title),
					)
					torrentMediaInfo = mediaInfo
					context.MediaInfo = mediaInfo
				} else {
					continue
				}
			}

			// 直接比对媒体信息
			if torrentMediaInfo != nil && (torrentMediaInfo.TMDBID != 0 || torrentMediaInfo.DoubanID != "") {
				if torrentMediaInfo.Type != mediaInfo.Type {
					continue
				}
				if torrentMediaInfo.TMDBID != 0 && torrentMediaInfo.TMDBID != mediaInfo.TMDBID {
					continue
				}
				if torrentMediaInfo.DoubanID != "" && torrentMediaInfo.DoubanID != mediaInfo.DoubanID {
					continue
				}
				s.logger.Info("通过媒体ID匹配到可选资源",
					zap.String("title", mediaInfo.TitleYear),
					zap.String("torrent", torrentInfo.Title),
				)
			} else {
				continue
			}

			// 如果是电视剧
			if torrentMediaInfo.Type == "tv" {
				// 有多季的不要
				if len(torrentMeta.SeasonList) > 1 {
					s.logger.Debug("有多季，不处理",
						zap.String("title", torrentInfo.Title),
					)
					continue
				}

				// 比对季
				if torrentMeta.BeginSeason > 0 {
					if meta.BeginSeason != torrentMeta.BeginSeason {
						s.logger.Debug("季不匹配",
							zap.String("title", torrentInfo.Title),
						)
						continue
					}
				} else if meta.BeginSeason != 1 {
					s.logger.Debug("季不匹配",
						zap.String("title", torrentInfo.Title),
					)
					continue
				}

				// 非洗版
				if !subscribe.BestVersion {
					// 不是缺失的剧集不要
					if len(noExists) > 0 {
						var mediaKey any
						if subscribe.TMDBID != nil && *subscribe.TMDBID != 0 {
							mediaKey = *subscribe.TMDBID
						} else if subscribe.DoubanID != "" {
							mediaKey = subscribe.DoubanID
						}

						if mediaKey != nil {
							noExistsInfo, ok := noExists[mediaKey][subscribe.Season]
							if ok && noExistsInfo != nil {
								// 是否有交集
								if len(noExistsInfo.Episodes) > 0 &&
									len(torrentMeta.EpisodeList) > 0 &&
									!hasIntersection(noExistsInfo.Episodes, torrentMeta.EpisodeList) {
									s.logger.Debug("对应剧集未包含缺失的剧集",
										zap.String("title", torrentInfo.Title),
										zap.Ints("torrent_episodes", torrentMeta.EpisodeList),
									)
									continue
								}
							}
						}
					}
				} else {
					// 洗版时，非整季不要
					if meta.Type == "tv" {
						if len(torrentMeta.EpisodeList) > 0 {
							s.logger.Debug("正在洗版，不是整季",
								zap.String("subscribe", subscribe.Name),
								zap.String("torrent", torrentInfo.Title),
							)
							continue
						}
					}
				}
			}

			// 匹配订阅附加参数
			if !s.torrentHelper.FilterTorrent(torrentInfo, s.GetParams(ctx, subscribe)) {
				continue
			}

			// 优先级过滤规则
			if len(ruleGroups) > 0 {
				result, err := s.filterService.FilterTorrents(ctx, &FilterTorrentsRequest{
					RuleGroups:  ruleGroups,
					TorrentList: []*TorrentInfo{torrentInfo},
					MediaInfo:   torrentMediaInfo,
				})
				if err != nil || len(result) == 0 {
					// 不符合过滤规则
					s.logger.Debug("不匹配过滤规则",
						zap.String("title", torrentInfo.Title),
					)
					continue
				}
			}

			// 洗版时，优先级小于已下载优先级的不要
			if subscribe.BestVersion {
				if subscribe.CurrentPriority != nil &&
					torrentInfo.PriOrder <= *subscribe.CurrentPriority {
					s.logger.Info("正在洗版，优先级低于或等于已下载优先级",
						zap.String("subscribe", subscribe.Name),
						zap.String("torrent", torrentInfo.Title),
					)
					continue
				}
			}

			// 匹配成功
			s.logger.Info("匹配成功",
				zap.String("title", mediaInfo.TitleYear),
				zap.String("torrent", torrentInfo.Title),
			)

			// 自定义属性
			if subscribe.MediaCategory != "" {
				torrentMediaInfo.Category = subscribe.MediaCategory
			}
			if subscribe.EpisodeGroup != "" {
				torrentMediaInfo.EpisodeGroup = subscribe.EpisodeGroup
			}

			matchContexts = append(matchContexts, context)
		}
	}

	return matchContexts
}

// contains 检查切片是否包含元素
func contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// hasIntersection 检查两个切片是否有交集
func hasIntersection(a, b []int) bool {
	set := make(map[int]bool)
	for _, v := range a {
		set[v] = true
	}
	for _, v := range b {
		if set[v] {
			return true
		}
	}
	return false
}
