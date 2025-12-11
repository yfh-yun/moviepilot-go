package subscribe

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Search 搜索订阅
// 对应Python: search()
func (s *Service) Search(ctx context.Context, req *SearchSubscribeRequest) error {
	// 获取锁
	lockAcquired := false
	lockChan := make(chan bool, 1)

	go func() {
		s.searchLock.Lock()
		lockChan <- true
	}()

	select {
	case <-lockChan:
		lockAcquired = true
		defer s.searchLock.Unlock()
		s.logger.Debug("search lock acquired", zap.Time("time", time.Now()))
	case <-time.After(LockTimeout):
		s.logger.Warn("search上锁超时")
		return fmt.Errorf("search上锁超时")
	}

	if !lockAcquired {
		return fmt.Errorf("获取锁失败")
	}

	// 获取订阅列表
	var subscribes []*Subscribe
	var err error

	if req.SubscribeID != nil {
		subscribe, getErr := s.subscribeRepo.Get(ctx, *req.SubscribeID)
		if getErr != nil {
			return getErr
		}
		if subscribe != nil {
			subscribes = []*Subscribe{subscribe}
		}
	} else {
		states := s.GetStatesForSearch(req.State)
		subscribes, err = s.subscribeRepo.List(ctx, states)
		if err != nil {
			return err
		}
	}

	// 遍历订阅
	for _, subscribe := range subscribes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var mediaKey any
		if subscribe.TMDBID != nil && *subscribe.TMDBID != 0 {
			mediaKey = *subscribe.TMDBID
		} else if subscribe.DoubanID != "" {
			mediaKey = subscribe.DoubanID
		}

		// 自定义识别词列表
		var customWordList []string
		if subscribe.CustomWords != "" {
			customWordList = strings.Split(subscribe.CustomWords, "\n")
		}

		// 校验创建时间，小于1分钟跳过
		if !subscribe.Date.IsZero() {
			now := time.Now()
			if now.Sub(subscribe.Date).Seconds() < 60 {
				s.logger.Debug("订阅新增小于1分钟，暂不搜索",
					zap.String("name", subscribe.Name),
				)
				continue
			}
		}

		// 随机休眠1-5分钟（非手动搜索且状态为R或P）
		if req.SubscribeID == nil && (req.State == "R" || req.State == "P") {
			sleepTime := rand.Intn(240) + 60 // 60-300秒
			s.logger.Info("订阅搜索随机休眠",
				zap.Int("seconds", sleepTime),
			)
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}

		// 搜索单个订阅
		if err := s.searchSingleSubscribe(ctx, subscribe, mediaKey, customWordList); err != nil {
			s.logger.Error("搜索订阅失败",
				zap.String("name", subscribe.Name),
				zap.Error(err),
			)
		}

		// 如果状态为N则更新为R
		if subscribe.State == "N" {
			if err := s.subscribeRepo.Update(ctx, subscribe.ID, map[string]any{
				"state": "R",
			}); err != nil {
				s.logger.Error("更新订阅状态失败", zap.Error(err))
			}
		}
	}

	// 手动触发时发送系统消息
	if req.Manual {
		if len(subscribes) > 0 {
			if req.SubscribeID != nil {
				s.messageService.Put(ctx, fmt.Sprintf("%s 搜索完成！", subscribes[0].Name), "订阅搜索", "system")
			} else {
				s.messageService.Put(ctx, "所有订阅搜索完成！", "订阅搜索", "system")
			}
		} else {
			s.messageService.Put(ctx, "没有找到订阅！", "订阅搜索", "system")
		}
	}

	return nil
}

// searchSingleSubscribe 搜索单个订阅
func (s *Service) searchSingleSubscribe(
	ctx context.Context,
	subscribe *Subscribe,
	mediaKey any,
	customWordList []string,
) error {
	s.logger.Info("开始搜索订阅", zap.String("name", subscribe.Name))

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
		return fmt.Errorf("未识别到媒体信息")
	}

	// 检查媒体是否存在
	existFlag, noExists, checkErr := s.CheckAndHandleExistingMedia(ctx, subscribe, meta, mediaInfo, mediaKey)
	if checkErr != nil {
		return checkErr
	}
	if existFlag {
		return nil
	}

	// 站点范围
	sites, sitesErr := s.GetSubSites(ctx, subscribe)
	if sitesErr != nil {
		return sitesErr
	}

	// 优先级过滤规则
	var ruleGroups []string
	if subscribe.BestVersion {
		if len(subscribe.FilterGroups) > 0 {
			ruleGroups = subscribe.FilterGroups
		} else {
			if val, configErr := s.systemConfigRepo.Get(ctx, "BestVersionFilterRuleGroups"); configErr == nil && val != nil {
				if groups, ok := val.([]string); ok {
					ruleGroups = groups
				}
			}
		}
	} else {
		if len(subscribe.FilterGroups) > 0 {
			ruleGroups = subscribe.FilterGroups
		} else {
			if val, configErr := s.systemConfigRepo.Get(ctx, "SubscribeFilterRuleGroups"); configErr == nil && val != nil {
				if groups, ok := val.([]string); ok {
					ruleGroups = groups
				}
			}
		}
	}

	// 搜索区域
	area := "title"
	if subscribe.SearchIMDBID {
		area = "imdbid"
	}

	// 搜索资源
	contexts, err := s.searchService.Process(ctx, &SearchRequest{
		MediaInfo:    mediaInfo,
		Keyword:      subscribe.Keyword,
		NoExists:     noExists,
		Sites:        sites,
		RuleGroups:   ruleGroups,
		Area:         area,
		CustomWords:  customWordList,
		FilterParams: s.GetParams(ctx, subscribe),
	})

	if err != nil || len(contexts) == 0 {
		s.logger.Warn("未搜索到资源",
			zap.String("keyword", subscribe.Keyword),
			zap.String("name", subscribe.Name),
		)
		return s.FinishSubscribeOrNot(ctx, subscribe, meta, mediaInfo, nil, noExists, false)
	}

	// 过滤搜索结果
	matchedContexts := s.filterSearchResults(subscribe, contexts)

	if len(matchedContexts) == 0 {
		s.logger.Warn("没有符合过滤条件的资源",
			zap.String("name", subscribe.Name),
		)
		return s.FinishSubscribeOrNot(ctx, subscribe, meta, mediaInfo, nil, noExists, false)
	}

	// 自动下载
	downloads, lefts, err := s.batchDownload(ctx, subscribe, matchedContexts, noExists)
	if err != nil {
		s.logger.Error("批量下载失败", zap.Error(err))
	}

	// 同步外部修改，更新订阅信息
	subscribe, err = s.subscribeRepo.Get(ctx, subscribe.ID)
	if err != nil {
		return err
	}

	// 判断是否应完成订阅
	if subscribe != nil {
		return s.FinishSubscribeOrNot(ctx, subscribe, meta, mediaInfo, downloads, lefts, false)
	}

	return nil
}

// filterSearchResults 过滤搜索结果
func (s *Service) filterSearchResults(subscribe *Subscribe, contexts []*Context) []*Context {
	matchedContexts := []*Context{}

	for _, context := range contexts {
		torrentMeta := context.MetaInfo
		torrentInfo := context.TorrentInfo
		torrentMediaInfo := context.MediaInfo

		// 洗版过滤
		if subscribe.BestVersion {
			// 洗版时，非整季不要
			if torrentMediaInfo.Type == "tv" {
				if len(torrentMeta.EpisodeList) > 0 {
					s.logger.Info("正在洗版，不是整季",
						zap.String("subscribe", subscribe.Name),
						zap.String("torrent", torrentInfo.Title),
					)
					continue
				}
			}

			// 洗版时，优先级小于等于已下载优先级的不要
			if subscribe.CurrentPriority != nil &&
				torrentInfo.PriOrder <= *subscribe.CurrentPriority {
				s.logger.Info("正在洗版，优先级低于或等于已下载优先级",
					zap.String("subscribe", subscribe.Name),
					zap.String("torrent", torrentInfo.Title),
				)
				continue
			}
		}

		// 更新订阅自定义属性
		if subscribe.MediaCategory != "" {
			torrentMediaInfo.Category = subscribe.MediaCategory
		}
		if subscribe.EpisodeGroup != "" {
			torrentMediaInfo.EpisodeGroup = subscribe.EpisodeGroup
		}

		matchedContexts = append(matchedContexts, context)
	}

	return matchedContexts
}

// batchDownload 批量下载
func (s *Service) batchDownload(
	ctx context.Context,
	subscribe *Subscribe,
	contexts []*Context,
	noExists map[any]map[int]*NotExistMediaInfo,
) ([]*Context, map[any]map[int]*NotExistMediaInfo, error) {
	result, err := s.downloadService.BatchDownload(ctx, &BatchDownloadRequest{
		Contexts:   contexts,
		NoExists:   noExists,
		Username:   subscribe.Username,
		SavePath:   subscribe.SavePath,
		Downloader: subscribe.Downloader,
		Source:     s.GetSubscribeSourceKeywordJSON(subscribe),
	})

	if err != nil {
		return nil, noExists, err
	}

	return result.Downloads, result.Lefts, nil
}
