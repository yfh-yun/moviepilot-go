package subscribe

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// Follow 刷新follow的用户分享，并自动添加订阅
// 对应Python: follow()
func (s *Service) Follow(ctx context.Context) error {
	s.logger.Info("开始刷新follow用户分享订阅")

	// 获取follow用户列表
	followUsers := []string{}
	if val, err := s.systemConfigRepo.Get(ctx, "FollowSubscribers"); err == nil && val != nil {
		if users, ok := val.([]string); ok {
			followUsers = users
		} else if users, ok := val.([]any); ok {
			for _, user := range users {
				if userStr, ok := user.(string); ok {
					followUsers = append(followUsers, userStr)
				}
			}
		}
	}

	if len(followUsers) == 0 {
		s.logger.Info("没有配置follow用户")
		return nil
	}

	successCount := 0

	// 获取分享列表
	shares, err := s.subscribeHelper.GetShares(ctx)
	if err != nil {
		return err
	}

	for _, shareSub := range shares {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		uid, ok := shareSub["share_uid"].(string)
		if !ok || uid == "" {
			continue
		}

		// 检查是否在follow列表中
		if !contains(followUsers, uid) {
			continue
		}

		// 提取订阅信息
		var tmdbID *int
		var doubanID string
		var season int

		if id, ok := shareSub["tmdbid"].(int); ok {
			tmdbID = &id
		}
		if id, ok := shareSub["doubanid"].(string); ok {
			doubanID = id
		}
		if s, ok := shareSub["season"].(int); ok {
			season = s
		}

		// 订阅已存在则跳过
		exists, err := s.subscribeRepo.Exists(ctx, tmdbID, doubanID, season)
		if err != nil {
			s.logger.Error("检查订阅存在失败", zap.Error(err))
			continue
		}
		if exists {
			continue
		}

		// 已经订阅过跳过
		existHistory, err := s.subscribeRepo.ExistHistory(ctx, tmdbID, doubanID, season)
		if err != nil {
			s.logger.Error("检查订阅历史失败", zap.Error(err))
			continue
		}
		if existHistory {
			continue
		}

		// 构建添加订阅请求
		req := &AddSubscribeRequest{
			Title:    getStringValue(shareSub, "name"),
			Year:     getStringValue(shareSub, "year"),
			Username: "订阅分享",
			ExistOK:  true,
		}

		if tmdbID != nil {
			req.TMDBID = tmdbID
		}
		if doubanID != "" {
			req.DoubanID = doubanID
		}
		if season > 0 {
			req.Season = &season
		}

		// 其他配置
		if mediaType, ok := shareSub["type"].(string); ok {
			req.MediaType = mediaType
		}
		if bestVersion, ok := shareSub["best_version"].(bool); ok {
			req.BestVersion = bestVersion
		}
		if savePath, ok := shareSub["save_path"].(string); ok {
			req.SavePath = savePath
		}
		if searchIMDBID, ok := shareSub["search_imdbid"].(bool); ok {
			req.SearchIMDBID = searchIMDBID
		}
		if customWords, ok := shareSub["custom_words"].(string); ok {
			req.CustomWords = customWords
		}
		if mediaCategory, ok := shareSub["media_category"].(string); ok {
			req.MediaCategory = mediaCategory
		}
		if filterGroups, ok := shareSub["filter_groups"].([]string); ok {
			req.FilterGroups = filterGroups
		}

		// 添加订阅
		resp, err := s.Add(ctx, req)
		if err != nil || resp.SubscribeID == 0 {
			s.logger.Error("follow用户分享订阅添加失败",
				zap.String("title", req.Title),
				zap.String("message", resp.Message),
			)
			continue
		}

		successCount++
		s.logger.Info("follow用户分享订阅添加成功",
			zap.String("title", req.Title),
		)
	}

	s.logger.Info("follow用户分享订阅刷新完成",
		zap.Int("success_count", successCount),
	)

	return nil
}

// CacheCalendar 预缓存订阅日历
// 对应Python: cache_calendar()
func (s *Service) CacheCalendar(ctx context.Context) error {
	s.logger.Info("开始预缓存订阅日历")

	// 获取所有订阅
	subscribes, err := s.subscribeRepo.List(ctx, []string{})
	if err != nil {
		return err
	}

	for _, subscribe := range subscribes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 识别媒体信息
		if subscribe.Type == "movie" {
			mediaInfo, err := s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
				MediaType:    subscribe.Type,
				TMDBID:       subscribe.TMDBID,
				DoubanID:     subscribe.DoubanID,
				BangumiID:    subscribe.BangumiID,
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
		} else {
			// 电视剧，获取剧集信息
			if subscribe.TMDBID != nil && *subscribe.TMDBID != 0 {
				episodes, err := s.tmdbService.GetEpisodes(ctx, *subscribe.TMDBID, subscribe.Season, subscribe.EpisodeGroup)
				if err != nil || len(episodes) == 0 {
					s.logger.Warn("未识别到季集信息",
						zap.String("name", subscribe.Name),
						zap.Any("tmdb_id", subscribe.TMDBID),
						zap.String("douban_id", subscribe.DoubanID),
						zap.Int("season", subscribe.Season),
					)
					continue
				}
			}
		}
	}

	s.logger.Info("订阅日历预缓存完成")
	return nil
}

// RemoteList 查询订阅并发送消息
// 对应Python: remote_list()
func (s *Service) RemoteList(ctx context.Context, channel, source, userID string) error {
	subscribes, err := s.subscribeRepo.List(ctx, []string{})
	if err != nil {
		return err
	}

	if len(subscribes) == 0 {
		notification := &Notification{
			Channel: channel,
			Source:  source,
			Title:   "没有任何订阅！",
			UserID:  userID,
		}
		return s.messageService.PostMessage(ctx, notification, nil, nil, nil)
	}

	title := fmt.Sprintf("共有 %d 个订阅，回复对应指令操作：\n", len(subscribes))
	title += "- 删除订阅：/subscribe_delete [id]\n"
	title += "- 搜索订阅：/subscribe_search [id]\n"
	title += "- 刷新订阅：/subscribe_refresh"

	messages := []string{}
	for _, subscribe := range subscribes {
		if subscribe.Type == "movie" {
			messages = append(messages, fmt.Sprintf("%d. %s（%s）",
				subscribe.ID, subscribe.Name, subscribe.Year))
		} else {
			downloaded := subscribe.TotalEpisode - subscribe.LackEpisode
			messages = append(messages, fmt.Sprintf("%d. %s（%s）第%d季 [%d/%d]",
				subscribe.ID, subscribe.Name, subscribe.Year, subscribe.Season,
				downloaded, subscribe.TotalEpisode))
		}
	}

	notification := &Notification{
		Channel: channel,
		Source:  source,
		Title:   title,
		Text:    strings.Join(messages, "\n"),
		UserID:  userID,
	}

	return s.messageService.PostMessage(ctx, notification, nil, nil, nil)
}

// RemoteDelete 删除订阅
// 对应Python: remote_delete()
func (s *Service) RemoteDelete(ctx context.Context, argStr, channel, source, userID string) error {
	if argStr == "" {
		notification := &Notification{
			Channel: channel,
			Source:  source,
			Title:   "请输入正确的命令格式：/subscribe_delete [id]，[id]为订阅编号",
			UserID:  userID,
		}
		return s.messageService.PostMessage(ctx, notification, nil, nil, nil)
	}

	argStrs := strings.Fields(argStr)

	for _, arg := range argStrs {
		arg = strings.TrimSpace(arg)
		subscribeID, err := strconv.Atoi(arg)
		if err != nil {
			continue
		}

		subscribe, err := s.subscribeRepo.Get(ctx, subscribeID)
		if err != nil || subscribe == nil {
			notification := &Notification{
				Channel: channel,
				Source:  source,
				Title:   fmt.Sprintf("订阅编号 %d 不存在！", subscribeID),
				UserID:  userID,
			}
			s.messageService.PostMessage(ctx, notification, nil, nil, nil)
			return fmt.Errorf("订阅编号 %d 不存在", subscribeID)
		}

		// 删除订阅
		if err := s.subscribeRepo.Delete(ctx, subscribeID); err != nil {
			return err
		}

		// 统计订阅
		data := map[string]any{
			"tmdbid":   subscribe.TMDBID,
			"doubanid": subscribe.DoubanID,
		}
		s.subscribeHelper.SubDoneAsync(ctx, data)
	}

	// 重新发送消息
	return s.RemoteList(ctx, channel, source, userID)
}

// SubscribeFilesInfo 获取订阅相关的下载和文件信息
// 对应Python: subscribe_files_info()
func (s *Service) SubscribeFilesInfo(ctx context.Context, subscribeID int) (*SubscribeInfo, error) {
	subscribe, err := s.subscribeRepo.Get(ctx, subscribeID)
	if err != nil || subscribe == nil {
		return nil, fmt.Errorf("订阅不存在")
	}

	subscribeInfo := &SubscribeInfo{
		Subscribe: subscribe,
		Episodes:  make(map[int]*SubscribeEpisodeInfo),
	}

	// 所有集的数据
	if subscribe.TMDBID != nil && *subscribe.TMDBID != 0 && subscribe.Type == "tv" {
		// 查询TMDB中的集信息
		tmdbEpisodes, tmdbErr := s.tmdbService.GetEpisodes(ctx, *subscribe.TMDBID, subscribe.Season, subscribe.EpisodeGroup)
		if tmdbErr == nil && len(tmdbEpisodes) > 0 {
			for _, episode := range tmdbEpisodes {
				info := &SubscribeEpisodeInfo{
					Title:       episode.Name,
					Description: episode.Overview,
					Backdrop:    episode.StillPath,
					Download:    []*SubscribeDownloadFileInfo{},
					Library:     []*SubscribeLibraryFileInfo{},
				}
				subscribeInfo.Episodes[episode.EpisodeNumber] = info
			}
		}
	} else if subscribe.Type == "tv" {
		// 根据开始结束集计算集信息
		startEpisode := subscribe.StartEpisode
		if startEpisode == 0 {
			startEpisode = 1
		}
		for i := startEpisode; i <= subscribe.TotalEpisode; i++ {
			info := &SubscribeEpisodeInfo{
				Title:    fmt.Sprintf("第 %d 集", i),
				Download: []*SubscribeDownloadFileInfo{},
				Library:  []*SubscribeLibraryFileInfo{},
			}
			subscribeInfo.Episodes[i] = info
		}
	} else {
		// 电影
		info := &SubscribeEpisodeInfo{
			Title:    subscribe.Name,
			Download: []*SubscribeDownloadFileInfo{},
			Library:  []*SubscribeLibraryFileInfo{},
		}
		subscribeInfo.Episodes[0] = info
	}

	// 所有下载记录
	downloadHis, downloadErr := s.downloadHistRepo.GetByMediaID(ctx, subscribe.TMDBID, subscribe.DoubanID)
	if downloadErr == nil && len(downloadHis) > 0 {
		for _, his := range downloadHis {
			// 查询下载文件
			files, filesErr := s.downloadHistRepo.GetFilesByHash(ctx, his.DownloadHash)
			if filesErr != nil || len(files) == 0 {
				continue
			}

			for _, file := range files {
				// 识别文件名
				fileMeta := &MetaInfo{
					Name: file.FilePath,
				}

				// 下载文件信息
				fileInfo := &SubscribeDownloadFileInfo{
					TorrentTitle: his.TorrentName,
					SiteName:     his.TorrentSite,
					Downloader:   file.Downloader,
					Hash:         his.DownloadHash,
					FilePath:     file.FullPath,
				}

				if subscribe.Type == "tv" {
					seasonNumber := fileMeta.BeginSeason
					if seasonNumber > 0 && seasonNumber != subscribe.Season {
						continue
					}
					episodeNumber := fileMeta.BeginEpisode
					if episodeNumber > 0 {
						if ep, ok := subscribeInfo.Episodes[episodeNumber]; ok {
							ep.Download = append(ep.Download, fileInfo)
						}
					}
				} else {
					if ep, ok := subscribeInfo.Episodes[0]; ok {
						ep.Download = append(ep.Download, fileInfo)
					}
				}
			}
		}
	}

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
		)
		return subscribeInfo, nil
	}

	// 所有媒体库文件记录
	libraryFileItems, err := s.mediaService.MediaFiles(ctx, mediaInfo)
	if err == nil && len(libraryFileItems) > 0 {
		for _, fileItem := range libraryFileItems {
			// 识别文件名
			fileMeta := &MetaInfo{
				Name: fileItem.Path,
			}

			// 媒体库文件信息
			fileInfo := &SubscribeLibraryFileInfo{
				Storage:  fileItem.Storage,
				FilePath: fileItem.Path,
			}

			if subscribe.Type == "tv" {
				seasonNumber := fileMeta.BeginSeason
				if seasonNumber > 0 && seasonNumber != subscribe.Season {
					continue
				}
				episodeNumber := fileMeta.BeginEpisode
				if episodeNumber > 0 {
					if ep, ok := subscribeInfo.Episodes[episodeNumber]; ok {
						ep.Library = append(ep.Library, fileInfo)
					}
				}
			} else {
				if ep, ok := subscribeInfo.Episodes[0]; ok {
					ep.Library = append(ep.Library, fileInfo)
				}
			}
		}
	}

	return subscribeInfo, nil
}

// RemoveSite 从订阅中移除与站点相关的设置
// 对应Python: remove_site()
func (s *Service) RemoveSite(ctx context.Context, siteID any) error {
	if siteID == nil {
		return nil
	}

	// 站点被重置
	if siteIDStr, ok := siteID.(string); ok && siteIDStr == "*" {
		// 清空RSS站点配置
		if err := s.systemConfigRepo.Set(ctx, "RssSites", []int{}); err != nil {
			return err
		}

		// 清空所有订阅的站点配置
		subscribes, err := s.subscribeRepo.List(ctx, []string{})
		if err != nil {
			return err
		}

		for _, subscribe := range subscribes {
			if len(subscribe.Sites) == 0 {
				continue
			}
			s.subscribeRepo.Update(ctx, subscribe.ID, map[string]any{
				"sites": []int{},
			})
		}

		return nil
	}

	// 转换站点ID
	var targetSiteID int
	switch v := siteID.(type) {
	case int:
		targetSiteID = v
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		targetSiteID = id
	default:
		return fmt.Errorf("无效的站点ID类型")
	}

	// 从选中的rss站点中移除
	selectedSites := []int{}
	if val, err := s.systemConfigRepo.Get(ctx, "RssSites"); err == nil && val != nil {
		if sites, ok := val.([]int); ok {
			selectedSites = sites
		}
	}

	if contains(selectedSites, targetSiteID) {
		newSites := []int{}
		for _, site := range selectedSites {
			if site != targetSiteID {
				newSites = append(newSites, site)
			}
		}
		s.systemConfigRepo.Set(ctx, "RssSites", newSites)
	}

	// 查询所有订阅
	subscribes, err := s.subscribeRepo.List(ctx, []string{})
	if err != nil {
		return err
	}

	for _, subscribe := range subscribes {
		if len(subscribe.Sites) == 0 {
			continue
		}

		if !contains(subscribe.Sites, targetSiteID) {
			continue
		}

		newSites := []int{}
		for _, site := range subscribe.Sites {
			if site != targetSiteID {
				newSites = append(newSites, site)
			}
		}

		s.subscribeRepo.Update(ctx, subscribe.ID, map[string]any{
			"sites": newSites,
		})
	}

	return nil
}

// getStringValue 从map中获取字符串值
func getStringValue(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// SubscribeInfo 订阅信息
type SubscribeInfo struct {
	Subscribe *Subscribe                    `json:"subscribe"`
	Episodes  map[int]*SubscribeEpisodeInfo `json:"episodes"`
}

// SubscribeEpisodeInfo 订阅剧集信息
type SubscribeEpisodeInfo struct {
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	Backdrop    string                       `json:"backdrop"`
	Download    []*SubscribeDownloadFileInfo `json:"download"`
	Library     []*SubscribeLibraryFileInfo  `json:"library"`
}

// SubscribeDownloadFileInfo 订阅下载文件信息
type SubscribeDownloadFileInfo struct {
	TorrentTitle string `json:"torrent_title"`
	SiteName     string `json:"site_name"`
	Downloader   string `json:"downloader"`
	Hash         string `json:"hash"`
	FilePath     string `json:"file_path"`
}

// SubscribeLibraryFileInfo 订阅媒体库文件信息
type SubscribeLibraryFileInfo struct {
	Storage  string `json:"storage"`
	FilePath string `json:"file_path"`
}
