package subscribe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Add 添加订阅
// 对应Python: add() 和 async_add()
func (s *Service) Add(ctx context.Context, req *AddSubscribeRequest) (*AddSubscribeResponse, error) {
	s.logger.Info("开始添加订阅",
		zap.String("title", req.Title),
		zap.String("year", req.Year),
		zap.String("type", req.MediaType),
	)

	// 1. 解析标题生成元数据
	meta := s.parseTitle(req.Title, req.Year, req.MediaType, req.Season)

	// 2. 识别媒体信息
	mediaInfo, err := s.recognizeMediaForAdd(ctx, req, meta)
	if err != nil {
		s.logger.Error("识别媒体信息失败", zap.Error(err))
		return &AddSubscribeResponse{
			Message: fmt.Sprintf("识别媒体信息失败: %v", err),
		}, err
	}

	if mediaInfo == nil {
		s.logger.Warn("未识别到媒体信息",
			zap.String("title", req.Title),
			zap.Any("tmdb_id", req.TMDBID),
			zap.String("douban_id", req.DoubanID),
		)
		return &AddSubscribeResponse{
			Message: "未识别到媒体信息",
		}, fmt.Errorf("未识别到媒体信息")
	}

	// 3. 处理电视剧总集数
	if mediaInfo.Type == "tv" {
		if handleErr := s.handleTVEpisodes(ctx, req, mediaInfo, meta); handleErr != nil {
			s.logger.Error("处理电视剧集数失败", zap.Error(handleErr))
			return &AddSubscribeResponse{
				Message: fmt.Sprintf("处理电视剧集数失败: %v", handleErr),
			}, handleErr
		}
	} else {
		// 电影避免season为0的问题
		req.Season = nil
	}

	// 4. 更新媒体图片
	if obtainErr := s.mediaService.ObtainImages(ctx, mediaInfo); obtainErr != nil {
		s.logger.Warn("更新媒体图片失败", zap.Error(err))
	}

	// 5. 合并媒体信息
	s.mergeMediaInfo(mediaInfo, req)

	// 6. 合并默认配置
	s.mergeDefaultConfig(ctx, req, mediaInfo.Type)

	// 7. 创建订阅对象
	subscribe := s.buildSubscribe(req, mediaInfo, meta)

	// 8. 保存到数据库
	subscribeID, err := s.subscribeRepo.Add(ctx, subscribe)
	if err != nil {
		s.logger.Error("保存订阅失败",
			zap.String("title", mediaInfo.Title),
			zap.Error(err),
		)

		// 发送失败消息
		if req.Message {
			s.sendFailureMessage(ctx, req, mediaInfo, meta, err.Error())
		}

		return &AddSubscribeResponse{
			Message: fmt.Sprintf("保存订阅失败: %v", err),
		}, err
	}

	subscribe.ID = subscribeID

	// 9. 发送成功消息
	if req.Message {
		s.sendSuccessMessage(ctx, req, mediaInfo, meta)
	}

	// 10. 发送订阅添加事件
	s.eventService.SendEventAsync(ctx, "SubscribeAdded", map[string]any{
		"subscribe_id": subscribeID,
		"username":     req.Username,
		"mediainfo":    mediaInfo,
	})

	// 11. 统计订阅（异步）
	go s.recordSubscribeStats(context.Background(), req, mediaInfo, meta)

	s.logger.Info("订阅添加成功",
		zap.Int("subscribe_id", subscribeID),
		zap.String("title", subscribe.Name),
	)

	return &AddSubscribeResponse{
		SubscribeID: subscribeID,
		Message:     "订阅添加成功",
	}, nil
}

// parseTitle 解析标题生成元数据
func (s *Service) parseTitle(title, year, mediaType string, season *int) *MetaInfo {
	meta := &MetaInfo{
		Name:  title,
		Title: title,
		Year:  year,
		Type:  mediaType,
	}

	if season != nil {
		meta.BeginSeason = *season
		meta.Season = *season
		meta.Type = "tv"
	}

	return meta
}

// recognizeMediaForAdd 为添加订阅识别媒体信息
func (s *Service) recognizeMediaForAdd(ctx context.Context, req *AddSubscribeRequest, meta *MetaInfo) (*MediaInfo, error) {
	var mediaInfo *MediaInfo
	var err error

	// 获取识别源配置
	recognizeSource := "themoviedb" // 默认使用TMDB
	if val, configErr := s.systemConfigRepo.Get(ctx, "RECOGNIZE_SOURCE"); configErr == nil && val != nil {
		if source, ok := val.(string); ok {
			recognizeSource = source
		}
	}

	if recognizeSource == "themoviedb" {
		// TMDB识别模式
		if req.TMDBID != nil {
			// 使用TMDBID识别
			mediaInfo, err = s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
				Meta:         meta,
				MediaType:    req.MediaType,
				TMDBID:       req.TMDBID,
				EpisodeGroup: req.EpisodeGroup,
				Cache:        false,
			})
		} else if req.DoubanID != "" {
			// 豆瓣ID转TMDB
			mediaInfo, err = s.mediaService.GetTMDBInfoByDoubanID(ctx, req.DoubanID, req.MediaType)
		} else if req.MediaID != "" {
			// 广播事件解析媒体信息
			mediaInfo, err = s.getEventMedia(ctx, req.MediaID, meta)
		}
	} else {
		// 豆瓣识别模式
		if req.DoubanID != "" {
			mediaInfo, err = s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
				Meta:      meta,
				MediaType: req.MediaType,
				DoubanID:  req.DoubanID,
				Cache:     false,
			})

			// 豆瓣标题处理
			if mediaInfo != nil {
				parsedMeta := s.parseTitle(mediaInfo.Title, "", "", nil)
				mediaInfo.Title = parsedMeta.Name
				if req.Season == nil && parsedMeta.BeginSeason > 0 {
					season := parsedMeta.BeginSeason
					req.Season = &season
				}
			}
		} else if req.MediaID != "" {
			mediaInfo, err = s.getEventMedia(ctx, req.MediaID, meta)
		}
	}

	// 使用名称识别兜底
	if mediaInfo == nil && err == nil {
		mediaInfo, err = s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
			Meta:         meta,
			EpisodeGroup: req.EpisodeGroup,
			Cache:        false,
		})
	}

	return mediaInfo, err
}

// getEventMedia 广播事件解析媒体信息
func (s *Service) getEventMedia(ctx context.Context, mediaID string, meta *MetaInfo) (*MediaInfo, error) {
	// 获取识别源配置
	recognizeSource := "themoviedb"
	if val, err := s.systemConfigRepo.Get(ctx, "RECOGNIZE_SOURCE"); err == nil && val != nil {
		if source, ok := val.(string); ok {
			recognizeSource = source
		}
	}

	// 发送媒体识别转换事件
	eventData := &MediaRecognizeConvertEventData{
		MediaID:     mediaID,
		ConvertType: recognizeSource,
	}

	result, err := s.eventService.SendMediaRecognizeConvertEvent(ctx, eventData)
	if err != nil {
		return nil, err
	}

	if result == nil || result.MediaDict == nil {
		return nil, fmt.Errorf("事件返回数据为空")
	}

	// 从事件返回的数据中获取ID
	newID := result.MediaDict["id"]
	if newID == nil {
		return nil, fmt.Errorf("事件返回数据中没有ID")
	}

	// 根据转换类型识别媒体
	if result.ConvertType == "themoviedb" {
		if tmdbID, ok := newID.(int); ok {
			return s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
				Meta:   meta,
				TMDBID: &tmdbID,
			})
		}
	} else if result.ConvertType == "douban" {
		if doubanID, ok := newID.(string); ok {
			return s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
				Meta:     meta,
				DoubanID: doubanID,
			})
		}
	}

	return nil, fmt.Errorf("无法识别媒体ID类型")
}

// handleTVEpisodes 处理电视剧集数
func (s *Service) handleTVEpisodes(ctx context.Context, req *AddSubscribeRequest, mediaInfo *MediaInfo, meta *MetaInfo) error {
	// 设置默认季度
	if req.Season == nil {
		season := 1
		req.Season = &season
	}

	// 如果用户已指定总集数，直接使用
	if req.TotalEpisode > 0 {
		if req.LackEpisode == 0 {
			req.LackEpisode = req.TotalEpisode
		}
		meta.Season = *req.Season
		return nil
	}

	// 获取总集数
	if mediaInfo.Seasons == nil || req.EpisodeGroup != "" {
		// 补充媒体信息
		fullMediaInfo, err := s.mediaService.RecognizeMedia(ctx, &RecognizeMediaRequest{
			MediaType:    mediaInfo.Type,
			TMDBID:       &mediaInfo.TMDBID,
			DoubanID:     mediaInfo.DoubanID,
			BangumiID:    &mediaInfo.BangumiID,
			EpisodeGroup: req.EpisodeGroup,
			Cache:        false,
		})
		if err != nil {
			return fmt.Errorf("补充媒体信息失败: %w", err)
		}
		if fullMediaInfo == nil {
			return fmt.Errorf("媒体信息识别失败")
		}
		mediaInfo.Seasons = fullMediaInfo.Seasons
	}

	if mediaInfo.Seasons == nil {
		return fmt.Errorf("媒体信息中没有季集信息")
	}

	episodes, ok := mediaInfo.Seasons[*req.Season]
	if !ok || len(episodes) == 0 {
		return fmt.Errorf("未获取到第 %d 季的总集数", *req.Season)
	}

	// 设置总集数和缺失集数
	req.TotalEpisode = len(episodes)
	if req.LackEpisode == 0 {
		req.LackEpisode = req.TotalEpisode
	}
	meta.Season = *req.Season

	return nil
}

// mergeMediaInfo 合并媒体信息
func (s *Service) mergeMediaInfo(mediaInfo *MediaInfo, req *AddSubscribeRequest) {
	if req.DoubanID != "" {
		mediaInfo.DoubanID = req.DoubanID
	}
	if req.BangumiID != nil {
		mediaInfo.BangumiID = *req.BangumiID
	}
}

// mergeDefaultConfig 合并默认配置
func (s *Service) mergeDefaultConfig(ctx context.Context, req *AddSubscribeRequest, mediaType string) {
	if req.Quality == "" {
		req.Quality = s.getDefaultSubscribeConfig(ctx, mediaType, "quality")
	}
	if req.Resolution == "" {
		req.Resolution = s.getDefaultSubscribeConfig(ctx, mediaType, "resolution")
	}
	if req.Effect == "" {
		req.Effect = s.getDefaultSubscribeConfig(ctx, mediaType, "effect")
	}
	if req.Include == "" {
		req.Include = s.getDefaultSubscribeConfig(ctx, mediaType, "include")
	}
	if req.Exclude == "" {
		req.Exclude = s.getDefaultSubscribeConfig(ctx, mediaType, "exclude")
	}
	if len(req.Sites) == 0 {
		if sites := s.getDefaultSubscribeConfig(ctx, mediaType, "sites"); sites != "" {
			// TODO: 解析站点配置
		}
	}
	if req.Downloader == "" {
		req.Downloader = s.getDefaultSubscribeConfig(ctx, mediaType, "downloader")
	}
	if req.SavePath == "" {
		req.SavePath = s.getDefaultSubscribeConfig(ctx, mediaType, "save_path")
	}
	if len(req.FilterGroups) == 0 {
		if filterGroups := s.getDefaultSubscribeConfig(ctx, mediaType, "filter_groups"); filterGroups != "" {
			req.FilterGroups = strings.Split(filterGroups, ",")
		}
	}
}

// getDefaultSubscribeConfig 获取订阅默认配置
// 对应Python: __get_default_subscribe_config()
func (s *Service) getDefaultSubscribeConfig(ctx context.Context, mediaType, key string) string {
	var configKey string
	if mediaType == "tv" {
		configKey = "DefaultTvSubscribeConfig"
	} else if mediaType == "movie" {
		configKey = "DefaultMovieSubscribeConfig"
	} else {
		return ""
	}

	value, err := s.systemConfigRepo.Get(ctx, configKey)
	if err != nil || value == nil {
		return ""
	}

	if config, ok := value.(map[string]any); ok {
		if val, exists := config[key]; exists && val != nil {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}

	return ""
}

// buildSubscribe 构建订阅对象
func (s *Service) buildSubscribe(req *AddSubscribeRequest, mediaInfo *MediaInfo, _ *MetaInfo) *Subscribe {
	subscribe := &Subscribe{
		Name:          mediaInfo.Title,
		Year:          mediaInfo.Year,
		Type:          mediaInfo.Type,
		TMDBID:        &mediaInfo.TMDBID,
		DoubanID:      mediaInfo.DoubanID,
		BangumiID:     &mediaInfo.BangumiID,
		IMDBID:        mediaInfo.IMDBID,
		TVDBID:        mediaInfo.TVDBID,
		Poster:        mediaInfo.PosterPath,
		Backdrop:      mediaInfo.BackdropPath,
		Vote:          mediaInfo.VoteAverage,
		Description:   mediaInfo.Overview,
		State:         "N", // 新建状态
		Username:      req.Username,
		Quality:       req.Quality,
		Resolution:    req.Resolution,
		Effect:        req.Effect,
		Include:       req.Include,
		Exclude:       req.Exclude,
		BestVersion:   req.BestVersion,
		SearchIMDBID:  req.SearchIMDBID,
		Sites:         req.Sites,
		Downloader:    req.Downloader,
		SavePath:      req.SavePath,
		FilterGroups:  req.FilterGroups,
		CustomWords:   req.CustomWords,
		MediaCategory: req.MediaCategory,
		EpisodeGroup:  req.EpisodeGroup,
		Date:          time.Now(),
		LastUpdate:    time.Now(),
	}

	if req.Season != nil {
		subscribe.Season = *req.Season
	}

	// 电视剧设置集数信息
	if mediaInfo.Type == "tv" {
		subscribe.TotalEpisode = req.TotalEpisode
		subscribe.LackEpisode = req.LackEpisode
	}

	return subscribe
}

// sendSuccessMessage 发送成功消息
func (s *Service) sendSuccessMessage(ctx context.Context, req *AddSubscribeRequest, mediaInfo *MediaInfo, meta *MetaInfo) {
	link := "#/subscribe/tv?tab=mysub"
	if mediaInfo.Type == "movie" {
		link = "#/subscribe/movie?tab=mysub"
	}

	notification := &Notification{
		Channel:  req.Channel,
		Source:   req.Source,
		MType:    "Subscribe",
		CType:    "SubscribeAdded",
		Image:    mediaInfo.PosterPath,
		Link:     link,
		UserID:   req.UserID,
		Username: req.Username,
	}

	extra := map[string]any{
		"username": req.Username,
	}

	if err := s.messageService.PostMessage(ctx, notification, meta, mediaInfo, extra); err != nil {
		s.logger.Error("发送成功消息失败", zap.Error(err))
	}
}

// sendFailureMessage 发送失败消息
func (s *Service) sendFailureMessage(ctx context.Context, req *AddSubscribeRequest, mediaInfo *MediaInfo, meta *MetaInfo, errMsg string) {
	seasonText := ""
	if meta.Season > 0 {
		seasonText = fmt.Sprintf(" 第%d季", meta.Season)
	}

	notification := &Notification{
		Channel:  req.Channel,
		Source:   req.Source,
		MType:    "Subscribe",
		Title:    fmt.Sprintf("%s %s%s 添加订阅失败！", mediaInfo.Title, mediaInfo.Year, seasonText),
		Text:     errMsg,
		Image:    mediaInfo.PosterPath,
		UserID:   req.UserID,
		Username: req.Username,
	}

	if err := s.messageService.PostMessage(ctx, notification, meta, mediaInfo, nil); err != nil {
		s.logger.Error("发送失败消息失败", zap.Error(err))
	}
}

// recordSubscribeStats 记录订阅统计
func (s *Service) recordSubscribeStats(ctx context.Context, req *AddSubscribeRequest, mediaInfo *MediaInfo, meta *MetaInfo) {
	data := map[string]any{
		"name":        req.Title,
		"year":        req.Year,
		"type":        meta.Type,
		"tmdbid":      mediaInfo.TMDBID,
		"imdbid":      mediaInfo.IMDBID,
		"tvdbid":      mediaInfo.TVDBID,
		"doubanid":    mediaInfo.DoubanID,
		"bangumiid":   mediaInfo.BangumiID,
		"season":      meta.BeginSeason,
		"poster":      mediaInfo.PosterPath,
		"backdrop":    mediaInfo.BackdropPath,
		"vote":        mediaInfo.VoteAverage,
		"description": mediaInfo.Overview,
	}

	if err := s.subscribeHelper.SubRegAsync(ctx, data); err != nil {
		s.logger.Error("记录订阅统计失败", zap.Error(err))
	}
}

// GetSubscribeSourceKeywordJSON 获取订阅来源关键字JSON
// 对应Python: get_subscribe_source_keyword()
func (s *Service) GetSubscribeSourceKeywordJSON(subscribe *Subscribe) string {
	keyword := &SubscribeSourceKeyword{
		ID:        subscribe.ID,
		Name:      subscribe.Name,
		Year:      subscribe.Year,
		Type:      subscribe.Type,
		Season:    subscribe.Season,
		TMDBID:    subscribe.TMDBID,
		IMDBID:    subscribe.IMDBID,
		TVDBID:    subscribe.TVDBID,
		DoubanID:  subscribe.DoubanID,
		BangumiID: subscribe.BangumiID,
	}

	data, _ := json.Marshal(keyword)
	return fmt.Sprintf("Subscribe|%s", string(data))
}
