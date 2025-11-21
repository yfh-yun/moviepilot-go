package actions

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
)

// MediaFilter 媒体过滤器接口
type MediaFilter interface {
	// FilterMedias 过滤媒体列表
	FilterMedias(ctx context.Context, medias []*MediaItem, params *MediaFilterParams) (*MediaFilterResponse, error)
	// ApplyFilter 应用过滤条件
	ApplyFilter(medias []*MediaItem, params *MediaFilterParams) ([]*MediaItem, error)
	// ValidateFilter 验证过滤条件
	ValidateFilter(params *MediaFilterParams) (*ValidateFilterResponse, error)
	// GetFilterStats 获取过滤统计信息
	GetFilterStats(ctx context.Context, medias []*MediaItem, params *MediaFilterParams) (*MediaFilterStats, error)
	// GetFilterSuggestions 获取过滤建议
	GetFilterSuggestions(ctx context.Context, medias []*MediaItem, field MediaFilterField, limit int) (*MediaFilterSuggestionsResponse, error)
}

// mediaFilter 媒体过滤器实现
type mediaFilter struct {
	logger logger.Logger
}

// NewMediaFilter 创建媒体过滤器实例
func NewMediaFilter(logger logger.Logger) MediaFilter {
	return &mediaFilter{
		logger: logger,
	}
}

// FilterMedias 过滤媒体列表
func (f *mediaFilter) FilterMedias(ctx context.Context, medias []*MediaItem, params *MediaFilterParams) (*MediaFilterResponse, error) {
	startTime := time.Now()
	log := f.logger.WithContext(ctx)

	// 记录开始过滤
	log.Debug("Starting to filter medias", "total_media_count", len(medias))

	// 验证参数
	if params == nil {
		params = &MediaFilterParams{}
	}

	// 设置默认值
	if params.Limit <= 0 || params.Limit > 1000 {
		params.Limit = 50
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	// 如果提供了page参数，计算offset
	if params.Page > 0 && params.Offset == 0 {
		params.Offset = (params.Page - 1) * params.Limit
	}
	if params.SortBy == "" {
		params.SortBy = MediaSortFieldCreateTime
	}
	if params.SortOrder == "" {
		params.SortOrder = SortOrderDesc
	}

	// 记录过滤参数
	log.Debug("Filter parameters", 
		"media_types", params.MediaTypes,
		"status", params.Status,
		"years", params.Years,
		"text_search", params.TextSearch,
		"sort_by", params.SortBy,
		"sort_order", params.SortOrder,
		"limit", params.Limit,
		"offset", params.Offset,
	)

	// 应用过滤条件
	filteredMedias, err := f.ApplyFilter(medias, params)
	if err != nil {
		log.Error("Failed to apply filter", "error", err.Error())
		return nil, fmt.Errorf("应用过滤条件失败: %w", err)
	}

	// 记录过滤结果
	log.Debug("Filtering results", 
		"total_before", len(medias), 
		"total_after", len(filteredMedias))

	// 排序
	sortedMedias := f.sortMedias(filteredMedias, params.SortBy, params.SortOrder)

	// 分页
	total := len(sortedMedias)
	pagedMedias := f.paginateMedias(sortedMedias, params.Limit, params.Offset)
	page := (params.Offset / params.Limit) + 1
	totalPages := (total + params.Limit - 1) / params.Limit

	// 构建响应
	response := &MediaFilterResponse{
		Success:        true,
		Medias:         pagedMedias,
		Total:          len(medias),
		Filtered:       total,
		Page:           page,
		PageSize:       params.Limit,
		TotalPages:     totalPages,
		ProcessingTime: time.Since(startTime),
		SortBy:         params.SortBy,
		SortOrder:      params.SortOrder,
	}

	// 如果有应用的过滤条件，添加到响应中
	if params.Filters != nil {
		response.AppliedFilters = params.Filters
	}

	log.Info("Successfully filtered medias", 
		"returned_count", len(response.Medias),
		"filtered_count", response.Filtered,
		"total_count", response.Total,
		"processing_time", response.ProcessingTime)

	return response, nil
}

// ApplyFilter 应用过滤条件
func (f *mediaFilter) ApplyFilter(medias []*MediaItem, params *MediaFilterParams) ([]*MediaItem, error) {
	if params == nil {
		return medias, nil
	}

	var filtered []*MediaItem

	// 遍历媒体列表应用过滤条件
	for _, media := range medias {
		if f.matchMedia(media, params) {
			filtered = append(filtered, media)
		}
	}

	return filtered, nil
}

// ValidateFilter 验证过滤条件
func (f *mediaFilter) ValidateFilter(params *MediaFilterParams) (*ValidateFilterResponse, error) {
	var errors []string

	// 验证基础参数
	if params.Limit < 0 || params.Limit > 1000 {
		errors = append(errors, "limit must be between 0 and 1000")
	}

	if params.Offset < 0 {
		errors = append(errors, "offset must be >= 0")
	}

	if params.Page < 0 {
		errors = append(errors, "page must be >= 0")
	}

	// 验证排序字段
	validSortFields := map[MediaSortField]bool{
		MediaSortFieldID:              true,
		MediaSortFieldTitle:           true,
		MediaSortFieldOriginalTitle:   true,
		MediaSortFieldType:            true,
		MediaSortFieldYear:            true,
		MediaSortFieldRating:          true,
		MediaSortFieldVotes:           true,
		MediaSortFieldRuntime:         true,
		MediaSortFieldSeasonCount:     true,
		MediaSortFieldEpisodeCount:    true,
		MediaSortFieldAirDate:         true,
		MediaSortFieldFirstAirDate:    true,
		MediaSortFieldLastAirDate:     true,
		MediaSortFieldReleaseDate:     true,
		MediaSortFieldCreateTime:      true,
		MediaSortFieldUpdateTime:      true,
		MediaSortFieldSortTitle:       true,
		MediaSortFieldLocalStatus:     true,
		MediaSortFieldSubscribeStatus: true,
		MediaSortFieldDownloadStatus:  true,
		MediaSortFieldFolderSize:      true,
		MediaSortFieldQuality:         true,
		MediaSortFieldResolution:      true,
	}

	if params.SortBy != "" && !validSortFields[params.SortBy] {
		errors = append(errors, fmt.Sprintf("invalid sort field: %s", params.SortBy))
	}

	// 验证排序顺序
	if params.SortOrder != "" && params.SortOrder != SortOrderAsc && params.SortOrder != SortOrderDesc {
		errors = append(errors, "sort_order must be 'asc' or 'desc'")
	}

	// 验证高级过滤条件
	if params.Filters != nil {
		if err := f.validateFilterGroup(params.Filters); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 构建响应
	response := &ValidateFilterResponse{
		Valid: len(errors) == 0,
	}

	if !response.Valid {
		response.Errors = errors
		response.Message = "Filter validation failed"
	} else {
		response.Message = "Filter validation successful"
	}

	return response, nil
}

// GetFilterStats 获取过滤统计信息
func (f *mediaFilter) GetFilterStats(ctx context.Context, medias []*MediaItem, params *MediaFilterParams) (*MediaFilterStats, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Getting filter statistics")

	// 应用过滤条件
	filteredMedias, err := f.ApplyFilter(medias, params)
	if err != nil {
		log.Error("Failed to apply filter for stats", "error", err.Error())
		return nil, fmt.Errorf("应用过滤条件失败: %w", err)
	}

	// 初始化统计数据
	stats := &MediaFilterStats{
		Total:           len(filteredMedias),
		ByMediaType:     make(map[MediaType]int),
		ByStatus:        make(map[MediaStatus]int),
		ByLocalStatus:   make(map[LocalMediaStatus]int),
		ByYear:          make(map[int]int),
		ByGenre:         make(map[string]int),
		ByTag:           make(map[string]int),
		ByQuality:       make(map[string]int),
		ByResolution:    make(map[string]int),
		RatingStats:     &RatingStats{},
		RuntimeStats:    &RuntimeStats{},
		FileSizeStats:   &FileSizeStats{},
		LastUpdated:     time.Now(),
	}

	// 计算统计数据
	totalRating := 0.0
	totalRuntime := 0
	totalFileSize := int64(0)
	ratingCount := 0
	runtimeCount := 0
	fileSizeCount := 0

	for _, media := range filteredMedias {
		// 媒体类型统计
		stats.ByMediaType[media.Type]++

		// 状态统计
		stats.ByStatus[media.Status]++

		// 本地状态统计
		stats.ByLocalStatus[media.LocalStatus]++

		// 年份统计
		if media.Year > 0 {
			stats.ByYear[media.Year]++
		}

		// 类型统计
		for _, genre := range media.Genres {
			stats.ByGenre[genre]++
		}

		// 标签统计
		for _, tag := range media.Tags {
			stats.ByTag[tag]++
		}

		// 质量统计
		if media.Quality != "" {
			stats.ByQuality[media.Quality]++
		}

		// 分辨率统计
		if media.Resolution != "" {
			stats.ByResolution[media.Resolution]++
		}

		// 评分统计
		if media.Rating > 0 {
			totalRating += media.Rating
			ratingCount++
			if media.Rating < stats.RatingStats.Min || stats.RatingStats.Min == 0 {
				stats.RatingStats.Min = media.Rating
			}
			if media.Rating > stats.RatingStats.Max {
				stats.RatingStats.Max = media.Rating
			}
		}

		// 时长统计
		if media.Runtime > 0 {
			totalRuntime += media.Runtime
			runtimeCount++
			if media.Runtime < stats.RuntimeStats.Min || stats.RuntimeStats.Min == 0 {
				stats.RuntimeStats.Min = media.Runtime
			}
			if media.Runtime > stats.RuntimeStats.Max {
				stats.RuntimeStats.Max = media.Runtime
			}
		}

		// 文件大小统计
		if media.FolderSize > 0 {
			totalFileSize += media.FolderSize
			fileSizeCount++
			if media.FolderSize < stats.FileSizeStats.Min || stats.FileSizeStats.Min == 0 {
				stats.FileSizeStats.Min = media.FolderSize
			}
			if media.FolderSize > stats.FileSizeStats.Max {
				stats.FileSizeStats.Max = media.FolderSize
			}
		}
	}

	// 计算平均值
	if ratingCount > 0 {
		stats.RatingStats.Average = totalRating / float64(ratingCount)
		stats.RatingStats.Count = ratingCount
	}

	if runtimeCount > 0 {
		stats.RuntimeStats.Average = float64(totalRuntime) / float64(runtimeCount)
		stats.RuntimeStats.Count = runtimeCount
	}

	if fileSizeCount > 0 {
		stats.FileSizeStats.Average = totalFileSize / int64(fileSizeCount)
		stats.FileSizeStats.Total = totalFileSize
		stats.FileSizeStats.Count = fileSizeCount
	}

	log.Info("Got filter statistics", "total", stats.Total)

	return stats, nil
}

// GetFilterSuggestions 获取过滤建议
func (f *mediaFilter) GetFilterSuggestions(ctx context.Context, medias []*MediaItem, field MediaFilterField, limit int) (*MediaFilterSuggestionsResponse, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Getting filter suggestions", "field", field, "limit", limit)

	if limit <= 0 {
		limit = 50
	}

	// 收集字段值
	valueCount := make(map[string]int)

	for _, media := range medias {
		values := f.getMediaFieldValues(media, field)
		for _, value := range values {
			if value != "" {
				valueCount[value]++
			}
		}
	}

	// 排序并限制结果
	type valueCountPair struct {
		Value string
		Count int
	}

	var pairs []valueCountPair
	for value, count := range valueCount {
		pairs = append(pairs, valueCountPair{Value: value, Count: count})
	}

	// 按计数降序排序
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	// 限制结果数量
	if len(pairs) > limit {
		pairs = pairs[:limit]
	}

	// 提取建议值
	var suggestions []*MediaFilterSuggestion
	suggestion := &MediaFilterSuggestion{
		Field:  field,
		Count:  len(pairs),
		Values: make([]string, 0, len(pairs)),
	}

	for _, pair := range pairs {
		suggestion.Values = append(suggestion.Values, pair.Value)
	}

	suggestions = append(suggestions, suggestion)

	// 构建响应
	response := &MediaFilterSuggestionsResponse{
		Success:    true,
		Suggestions: suggestions,
		Fields:      f.getAvailableFilterFields(),
	}

	log.Info("Got filter suggestions", "field", field, "suggestion_count", len(suggestions))

	return response, nil
}

// 辅助方法：匹配媒体
func (f *mediaFilter) matchMedia(media *MediaItem, params *MediaFilterParams) bool {
	// 1. 应用高级过滤条件
	if params.Filters != nil {
		if !f.matchFilterGroup(media, params.Filters) {
			return false
		}
	}

	// 2. 应用快捷过滤条件
	// 媒体类型过滤
	if len(params.MediaTypes) > 0 && !containsMediaType(params.MediaTypes, media.Type) {
		return false
	}

	// 状态过滤
	if len(params.Status) > 0 && !containsMediaStatus(params.Status, media.Status) {
		return false
	}

	// 年份过滤
	if len(params.Years) > 0 && !containsInt(params.Years, media.Year) {
		return false
	}

	// 类型过滤
	if len(params.Genres) > 0 && !f.intersects(media.Genres, params.Genres) {
		return false
	}

	// 标签过滤
	if len(params.Tags) > 0 && !f.intersects(media.Tags, params.Tags) {
		return false
	}

	// 评分过滤
	if params.RatingMin != nil && media.Rating < *params.RatingMin {
		return false
	}
	if params.RatingMax != nil && media.Rating > *params.RatingMax {
		return false
	}

	// 投票数过滤
	if params.VotesMin != nil && media.Votes < *params.VotesMin {
		return false
	}

	// 时长过滤
	if params.RuntimeMin != nil && media.Runtime < *params.RuntimeMin {
		return false
	}
	if params.RuntimeMax != nil && media.Runtime > *params.RuntimeMax {
		return false
	}

	// 语言过滤
	if len(params.Language) > 0 && !containsString(params.Language, media.Language) {
		return false
	}

	// 国家过滤
	if len(params.Country) > 0 && !containsString(params.Country, media.Country) {
		return false
	}

	// 网络/平台过滤
	if len(params.Networks) > 0 && !containsString(params.Networks, media.Network) {
		return false
	}

	// 合集过滤
	if len(params.Collections) > 0 && !containsString(params.Collections, media.Collection)
	{
		return false
	}

	// 工作室过滤
	if len(params.Studios) > 0 && !containsString(params.Studios, media.Studio) {
		return false
	}

	// 本地状态过滤
	if len(params.LocalStatus) > 0 && !containsLocalMediaStatus(params.LocalStatus, media.LocalStatus) {
		return false
	}

	// 订阅状态过滤
	if len(params.SubscribeStatus) > 0 && !containsSubscribeStatus(params.SubscribeStatus, media.SubscribeStatus) {
		return false
	}

	// 下载状态过滤
	if len(params.DownloadStatus) > 0 && !containsDownloadStatus(params.DownloadStatus, media.DownloadStatus) {
		return false
	}

	// 字幕状态过滤
	if len(params.SubtitleStatus) > 0 && !containsSubtitleStatus(params.SubtitleStatus, media.SubtitleStatus) {
		return false
	}

	// 本地文件存在性过滤
	if params.HasLocal != nil {
		hasLocal := media.LocalStatus != LocalMediaStatusNotFound
		if *params.HasLocal != hasLocal {
			return false
		}
	}

	// 订阅存在性过滤
	if params.HasSubscription != nil {
		hasSubscription := media.SubscribeStatus != SubscribeStatusNone
		if *params.HasSubscription != hasSubscription {
			return false
		}
	}

	// 文本搜索
	if params.TextSearch != "" {
		searchLower := strings.ToLower(params.TextSearch)
		match := strings.Contains(strings.ToLower(media.Title), searchLower) ||
			strings.Contains(strings.ToLower(media.OriginalTitle), searchLower) ||
			strings.Contains(strings.ToLower(media.Overview), searchLower)
		if !match {
			return false
		}
	}

	// 排除ID列表
	if len(params.ExcludeIDs) > 0 && containsInt64(params.ExcludeIDs, media.ID) {
		return false
	}

	// 包含ID列表
	if len(params.IncludeIDs) > 0 && !containsInt64(params.IncludeIDs, media.ID) {
		return false
	}

	// 质量过滤
	if len(params.Quality) > 0 && !containsString(params.Quality, media.Quality) {
		return false
	}

	// 分辨率过滤
	if len(params.Resolution) > 0 && !containsString(params.Resolution, media.Resolution) {
		return false
	}

	// 编码过滤
	if len(params.Codec) > 0 && !containsString(params.Codec, media.Codec) {
		return false
	}

	// 音频格式过滤
	if len(params.Audio) > 0 && !containsString(params.Audio, media.Audio) {
		return false
	}

	// 视频格式过滤
	if len(params.VideoFormat) > 0 && !containsString(params.VideoFormat, media.VideoFormat) {
		return false
	}

	// 文件夹大小过滤
	if params.FolderSizeMin != nil && media.FolderSize < *params.FolderSizeMin {
		return false
	}
	if params.FolderSizeMax != nil && media.FolderSize > *params.FolderSizeMax {
		return false
	}

	// 媒体服务器ID过滤
	if params.MediaServerID != "" && media.MediaServerID != params.MediaServerID {
		return false
	}

	// 下载器ID过滤
	if params.DownloaderID != "" && media.DownloaderID != params.DownloaderID {
		return false
	}

	return true
}

// 辅助方法：匹配过滤条件组
func (f *mediaFilter) matchFilterGroup(media *MediaItem, group *FilterGroup) bool {
	if group == nil || len(group.Conditions) == 0 {
		return true
	}

	if group.Logic == "and" {
		for _, condition := range group.Conditions {
			if !f.matchCondition(media, condition) {
				return false
			}
		}
		return true
	} else if group.Logic == "or" {
		for _, condition := range group.Conditions {
			if f.matchCondition(media, condition) {
				return true
			}
		}
		return false
	}

	return false
}

// 辅助方法：匹配单个条件
func (f *mediaFilter) matchCondition(media *MediaItem, condition interface{}) bool {
	switch cond := condition.(type) {
	case *FilterCondition:
		return f.matchSingleCondition(media, cond)
	case *FilterGroup:
		return f.matchFilterGroup(media, cond)
	default:
		return false
	}
}

// 辅助方法：匹配单个过滤条件
func (f *mediaFilter) matchSingleCondition(media *MediaItem, condition *FilterCondition) bool {
	// 获取字段值
	fieldValue := f.getMediaFieldValue(media, condition.Field)

	// 根据操作符匹配
	switch condition.Operator {
	case FilterOperatorEq:
		return f.compareValues(fieldValue, condition.Value, "eq")
	case FilterOperatorNe:
		return !f.compareValues(fieldValue, condition.Value, "eq")
	case FilterOperatorGt:
		return f.compareValues(fieldValue, condition.Value, "gt")
	case FilterOperatorGte:
		return f.compareValues(fieldValue, condition.Value, "gte")
	case FilterOperatorLt:
		return f.compareValues(fieldValue, condition.Value, "lt")
	case FilterOperatorLte:
		return f.compareValues(fieldValue, condition.Value, "lte")
	case FilterOperatorLike:
		return f.contains(fieldValue, condition.Value)
	case FilterOperatorNotLike:
		return !f.contains(fieldValue, condition.Value)
	case FilterOperatorIn:
		return f.in(fieldValue, condition.Value)
	case FilterOperatorNotIn:
		return !f.in(fieldValue, condition.Value)
	case FilterOperatorRegex:
		return f.regexMatch(fieldValue, condition.Value)
	case FilterOperatorNotRegex:
		return !f.regexMatch(fieldValue, condition.Value)
	case FilterOperatorBetween:
		return f.between(fieldValue, condition.Value)
	case FilterOperatorIsNull:
		return f.isNull(fieldValue)
	case FilterOperatorIsNotNull:
		return !f.isNull(fieldValue)
	case FilterOperatorStartsWith:
		return f.startsWith(fieldValue, condition.Value)
	case FilterOperatorEndsWith:
		return f.endsWith(fieldValue, condition.Value)
	default:
		return false
	}
}

// 辅助方法：获取媒体字段值
func (f *mediaFilter) getMediaFieldValue(media *MediaItem, field MediaFilterField) interface{} {
	switch field {
	case MediaFilterFieldID:
		return media.ID
	case MediaFilterFieldTitle:
		return media.Title
	case MediaFilterFieldOriginalTitle:
		return media.OriginalTitle
	case MediaFilterFieldType:
		return media.Type
	case MediaFilterFieldStatus:
		return media.Status
	case MediaFilterFieldYear:
		return media.Year
	case MediaFilterFieldRating:
		return media.Rating
	case MediaFilterFieldVotes:
		return media.Votes
	case MediaFilterFieldRuntime:
		return media.Runtime
	case MediaFilterFieldSeasonCount:
		return media.SeasonCount
	case MediaFilterFieldEpisodeCount:
		return media.EpisodeCount
	case MediaFilterFieldAirDate:
		return media.AirDate
	case MediaFilterFieldFirstAirDate:
		return media.FirstAirDate
	case MediaFilterFieldLastAirDate:
		return media.LastAirDate
	case MediaFilterFieldReleaseDate:
		return media.ReleaseDate
	case MediaFilterFieldOverview:
		return media.Overview
	case MediaFilterFieldGenres:
		return strings.Join(media.Genres, ", ")
	case MediaFilterFieldTags:
		return strings.Join(media.Tags, ", ")
	case MediaFilterFieldStudio:
		return media.Studio
	case MediaFilterFieldDirector:
		return media.Director
	case MediaFilterFieldCast:
		return strings.Join(media.Cast, ", ")
	case MediaFilterFieldWriter:
		return media.Writer
	case MediaFilterFieldIMDBID:
		return media.IMDBID
	case MediaFilterFieldTMDBID:
		return media.TMDBID
	case MediaFilterFieldTVDBID:
		return media.TVDBID
	case MediaFilterFieldSource:
		return media.Source
	case MediaFilterFieldCover:
		return media.Cover
	case MediaFilterFieldBackdrop:
		return media.Backdrop
	case MediaFilterFieldTrailer:
		return media.Trailer
	case MediaFilterFieldLogo:
		return media.Logo
	case MediaFilterFieldLocalStatus:
		return media.LocalStatus
	case MediaFilterFieldSubscribeStatus:
		return media.SubscribeStatus
	case MediaFilterFieldDownloadStatus:
		return media.DownloadStatus
	case MediaFilterFieldCreateTime:
		return media.CreateTime
	case MediaFilterFieldUpdateTime:
		return media.UpdateTime
	case MediaFilterFieldSortTitle:
		return media.SortTitle
	case MediaFilterFieldLanguage:
		return media.Language
	case MediaFilterFieldCountry:
		return media.Country
	case MediaFilterFieldNetwork:
		return media.Network
	case MediaFilterFieldCollection:
		return media.Collection
	case MediaFilterFieldQuality:
		return media.Quality
	case MediaFilterFieldCodec:
		return media.Codec
	case MediaFilterFieldResolution:
		return media.Resolution
	case MediaFilterFieldAudio:
		return media.Audio
	case MediaFilterFieldVideoFormat:
		return media.VideoFormat
	case MediaFilterFieldFolderSize:
		return media.FolderSize
	case MediaFilterFieldFilePath:
		return media.FilePath
	case MediaFilterFieldFolderPath:
		return media.FolderPath
	case MediaFilterFieldMediaServer:
		return media.MediaServerID
	case MediaFilterFieldSubtitleStatus:
		return media.SubtitleStatus
	case MediaFilterFieldCustom1:
		return media.Custom1
	case MediaFilterFieldCustom2:
		return media.Custom2
	case MediaFilterFieldCustom3:
		return media.Custom3
	default:
		return nil
	}
}

// 辅助方法：获取媒体字段的多个值（用于多值字段如类型、标签等）
func (f *mediaFilter) getMediaFieldValues(media *MediaItem, field MediaFilterField) []string {
	switch field {
	case MediaFilterFieldGenres:
		return media.Genres
	case MediaFilterFieldTags:
		return media.Tags
	case MediaFilterFieldCast:
		return media.Cast
	case MediaFilterFieldQuality:
		if media.Quality != "" {
			return []string{media.Quality}
		}
	case MediaFilterFieldResolution:
		if media.Resolution != "" {
			return []string{media.Resolution}
		}
	case MediaFilterFieldCodec:
		if media.Codec != "" {
			return []string{media.Codec}
		}
	case MediaFilterFieldAudio:
		if media.Audio != "" {
			return []string{media.Audio}
		}
	case MediaFilterFieldVideoFormat:
		if media.VideoFormat != "" {
			return []string{media.VideoFormat}
		}
	case MediaFilterFieldLanguage:
		if media.Language != "" {
			return []string{media.Language}
		}
	case MediaFilterFieldCountry:
		if media.Country != "" {
			return []string{media.Country}
		}
	case MediaFilterFieldNetwork:
		if media.Network != "" {
			return []string{media.Network}
		}
	case MediaFilterFieldStudio:
		if media.Studio != "" {
			return []string{media.Studio}
		}
	case MediaFilterFieldCollection:
		if media.Collection != "" {
			return []string{media.Collection}
		}
	}
	return []string{}
}

// 辅助方法：比较值
func (f *mediaFilter) compareValues(a, b interface{}, op string) bool {
	// 实现不同类型值的比较逻辑
	// 这里是简化实现，实际需要处理更多类型
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	switch op {
	case "eq":
		return aStr == bStr
	case "gt":
		return aStr > bStr
	case "gte":
		return aStr >= bStr
	case "lt":
		return aStr < bStr
	case "lte":
		return aStr <= bStr
	default:
		return false
	}
}

// 辅助方法：包含检查
func (f *mediaFilter) contains(a, b interface{}) bool {
	aStr := strings.ToLower(fmt.Sprintf("%v", a))
	bStr := strings.ToLower(fmt.Sprintf("%v", b))
	return strings.Contains(aStr, bStr)
}

// 辅助方法：在范围内检查
func (f *mediaFilter) in(a interface{}, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)

	// 尝试将b转换为切片
	switch v := b.(type) {
	case [] []
		case []string:
			for _, item := range v {
				if aStr == item {
					return true
				}
			}
		case []interface{}:
			for _, item := range v {
				if aStr == fmt.Sprintf("%v", item) {
					return true
				}
			}
	default:
		// 尝试将b作为字符串处理
		bStr := fmt.Sprintf("%v", b)
		return aStr == bStr
	}

	return false
}

// 辅助方法：正则匹配
func (f *mediaFilter) regexMatch(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	match, err := regexp.MatchString(bStr, aStr)
	return err == nil && match
}

// 辅助方法：范围检查
func (f *mediaFilter) between(a, b interface{}) bool {
	// 简化实现，实际需要处理数值类型
	return false
}

// 辅助方法：空值检查
func (f *mediaFilter) isNull(a interface{}) bool {
	return a == nil || fmt.Sprintf("%v", a) == ""
}

// 辅助方法：开始于检查
func (f *mediaFilter) startsWith(a, b interface{}) bool {
	aStr := strings.ToLower(fmt.Sprintf("%v", a))
	bStr := strings.ToLower(fmt.Sprintf("%v", b))
	return strings.HasPrefix(aStr, bStr)
}

// 辅助方法：结束于检查
func (f *mediaFilter) endsWith(a, b interface{}) bool {
	aStr := strings.ToLower(fmt.Sprintf("%v", a))
	bStr := strings.ToLower(fmt.Sprintf("%v", b))
	return strings.HasSuffix(aStr, bStr)
}

// 辅助方法：验证过滤条件组
func (f *mediaFilter) validateFilterGroup(group *FilterGroup) error {
	if group == nil {
		return fmt.Errorf("filter group cannot be nil")
	}

	if group.Logic != "and" && group.Logic != "or" {
		return fmt.Errorf("filter group logic must be 'and' or 'or'")
	}

	if len(group.Conditions) == 0 {
		return fmt.Errorf("filter group must have at least one condition")
	}

	// 递归验证每个条件
	for i, condition := range group.Conditions {
		switch cond := condition.(type) {
		case *FilterCondition:
			if err := f.validateFilterCondition(cond); err != nil {
				return fmt.Errorf("invalid condition at index %d: %w", i, err)
			}
		case *FilterGroup:
			if err := f.validateFilterGroup(cond); err != nil {
				return fmt.Errorf("invalid subgroup at index %d: %w", i, err)
			}
		default:
			return fmt.Errorf("invalid condition type at index %d", i)
		}
	}

	return nil
}

// 辅助方法：验证单个过滤条件
func (f *mediaFilter) validateFilterCondition(condition *FilterCondition) error {
	if condition == nil {
		return fmt.Errorf("filter condition cannot be nil")
	}

	// 验证字段名
	validFields := f.getAvailableFilterFields()
	validField := false
	for _, field := range validFields {
		if field == condition.Field {
			validField = true
			break
		}
	}
	if !validField {
		return fmt.Errorf("invalid filter field: %s", condition.Field)
	}

	// 验证操作符
	validOperators := []FilterOperator{
		FilterOperatorEq, FilterOperatorNe, FilterOperatorGt, FilterOperatorGte,
		FilterOperatorLt, FilterOperatorLte, FilterOperatorLike, FilterOperatorNotLike,
		FilterOperatorIn, FilterOperatorNotIn, FilterOperatorRegex, FilterOperatorNotRegex,
		FilterOperatorBetween, FilterOperatorIsNull, FilterOperatorIsNotNull,
		FilterOperatorStartsWith, FilterOperatorEndsWith,
	}

	validOperator := false
	for _, op := range validOperators {
		if op == condition.Operator {
			validOperator = true
			break
		}
	}
	if !validOperator {
		return fmt.Errorf("invalid filter operator: %s", condition.Operator)
	}

	// 验证值
	if condition.Operator != FilterOperatorIsNull && condition.Operator != FilterOperatorIsNotNull && condition.Value == nil {
		return fmt.Errorf("filter value cannot be nil for operator: %s", condition.Operator)
	}

	return nil
}

// 辅助方法：获取可用的过滤字段
func (f *mediaFilter) getAvailableFilterFields() []MediaFilterField {
	return []MediaFilterField{
		MediaFilterFieldID, MediaFilterFieldTitle, MediaFilterFieldOriginalTitle,
		MediaFilterFieldType, MediaFilterFieldStatus, MediaFilterFieldYear,
		MediaFilterFieldRating, MediaFilterFieldVotes, MediaFilterFieldRuntime,
		MediaFilterFieldSeasonCount, MediaFilterFieldEpisodeCount,
		MediaFilterFieldAirDate, MediaFilterFieldFirstAirDate,
		MediaFilterFieldLastAirDate, MediaFilterFieldReleaseDate,
		MediaFilterFieldOverview, MediaFilterFieldGenres, MediaFilterFieldTags,
		MediaFilterFieldStudio, MediaFilterFieldDirector, MediaFilterFieldCast,
		MediaFilterFieldWriter, MediaFilterFieldIMDBID, MediaFilterFieldTMDBID,
		MediaFilterFieldTVDBID, MediaFilterFieldSource, MediaFilterFieldCover,
		MediaFilterFieldBackdrop, MediaFilterFieldTrailer, MediaFilterFieldLogo,
		MediaFilterFieldLocalStatus, MediaFilterFieldSubscribeStatus,
		MediaFilterFieldDownloadStatus, MediaFilterFieldCreateTime,
		MediaFilterFieldUpdateTime, MediaFilterFieldSortTitle,
		MediaFilterFieldLanguage, MediaFilterFieldCountry,
		MediaFilterFieldNetwork, MediaFilterFieldCollection,
		MediaFilterFieldQuality, MediaFilterFieldCodec,
		MediaFilterFieldResolution, MediaFilterFieldAudio,
		MediaFilterFieldVideoFormat, MediaFilterFieldFolderSize,
		MediaFilterFieldFilePath, MediaFilterFieldFolderPath,
		MediaFilterFieldMediaServer, MediaFilterFieldSubtitleStatus,
		MediaFilterFieldCustom1, MediaFilterFieldCustom2, MediaFilterFieldCustom3,
	}
}

// 辅助方法：排序媒体
func (f *mediaFilter) sortMedias(medias []*MediaItem, sortBy MediaSortField, sortOrder SortOrder) []*MediaItem {
	// 创建副本以避免修改原始数据
	sorted := make([]*MediaItem, len(medias))
	copy(sorted, medias)

	// 排序
	sort.Slice(sorted, func(i, j int) bool {
		iMedia := sorted[i]
		jMedia := sorted[j]

		var less bool

		switch sortBy {
		case MediaSortFieldID:
			less = iMedia.ID < jMedia.ID
		case MediaSortFieldTitle:
			less = iMedia.Title < jMedia.Title
		case MediaSortFieldOriginalTitle:
			less = iMedia.OriginalTitle < jMedia.OriginalTitle
		case MediaSortFieldType:
			less = iMedia.Type < jMedia.Type
		case MediaSortFieldYear:
			less = iMedia.Year < jMedia.Year
		case MediaSortFieldRating:
			less = iMedia.Rating < jMedia.Rating
		case MediaSortFieldVotes:
			less = iMedia.Votes < jMedia.Votes
		case MediaSortFieldRuntime:
			less = iMedia.Runtime < jMedia.Runtime
		case MediaSortFieldSeasonCount:
			less = iMedia.SeasonCount < jMedia.SeasonCount
		case MediaSortFieldEpisodeCount:
			less = iMedia.EpisodeCount < jMedia.EpisodeCount
		case MediaSortFieldAirDate:
			less = iMedia.AirDate.Before(jMedia.AirDate)
		case MediaSortFieldFirstAirDate:
			less = iMedia.FirstAirDate.Before(jMedia.FirstAirDate)
		case MediaSortFieldLastAirDate:
			less = iMedia.LastAirDate.Before(jMedia.LastAirDate)
		case MediaSortFieldReleaseDate:
			less = iMedia.ReleaseDate.Before(jMedia.ReleaseDate)
		case MediaSortFieldCreateTime:
			less = iMedia.CreateTime.Before(jMedia.CreateTime)
		case MediaSortFieldUpdateTime:
			less = iMedia.UpdateTime.Before(jMedia.UpdateTime)
		case MediaSortFieldSortTitle:
			less = iMedia.SortTitle < jMedia.SortTitle
		case MediaSortFieldLocalStatus:
			less = iMedia.LocalStatus < jMedia.LocalStatus
		case MediaSortFieldSubscribeStatus:
			less = iMedia.SubscribeStatus < jMedia.SubscribeStatus
		case MediaSortFieldDownloadStatus:
			less = iMedia.DownloadStatus < jMedia.DownloadStatus
		case MediaSortFieldFolderSize:
			less = iMedia.FolderSize < jMedia.FolderSize
		case MediaSortFieldQuality:
			less = iMedia.Quality < jMedia.Quality
		case MediaSortFieldResolution:
			less = iMedia.Resolution < jMedia.Resolution
		default:
			// 默认按创建时间排序
			less = iMedia.CreateTime.Before(jMedia.CreateTime)
		}

		// 根据排序顺序反转结果
		if sortOrder == SortOrderDesc {
			less = !less
		}

		return less
	})

	return sorted
}

// 辅助方法：分页媒体
func (f *mediaFilter) paginateMedias(medias []*MediaItem, limit, offset int) []*MediaItem {
	start := offset
	end := offset + limit

	if start >= len(medias) {
		return []*MediaItem{}
	}

	if end > len(medias) {
		end = len(medias)
	}

	return medias[start:end]
}

// 辅助方法：检查两个字符串切片是否有交集
func (f *mediaFilter) intersects(a, b []string) bool {
	bMap := make(map[string]bool)
	for _, v := range b {
		bMap[strings.ToLower(v)] = true
	}

	for _, v := range a {
		if bMap[strings.ToLower(v)] {
			return true
		}
	}

	return false
}

// 辅助函数：包含检查
func containsMediaType(slice []MediaType, item MediaType) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsMediaStatus(slice []MediaStatus, item MediaStatus) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsLocalMediaStatus(slice []LocalMediaStatus, item LocalMediaStatus) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsSubscribeStatus(slice []SubscribeStatus, item SubscribeStatus) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsDownloadStatus(slice []DownloadStatus, item DownloadStatus) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsSubtitleStatus(slice []SubtitleStatus, item SubtitleStatus) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsString(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}

func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInt64(slice []int64, item int64) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
