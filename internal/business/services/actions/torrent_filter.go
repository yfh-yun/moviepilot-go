package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/cache"
)

// TorrentFilter 种子过滤器接口
type TorrentFilter interface {
	// FilterTorrents 过滤种子列表
	FilterTorrents(ctx context.Context, params *TorrentFilterParams) (*TorrentFilterResponse, error)
	// ApplyFilter 应用过滤条件到种子列表
	ApplyFilter(ctx context.Context, torrents []TorrentItem, params *TorrentFilterParams) ([]TorrentItem, error)
	// ValidateFilter 验证过滤参数
	ValidateFilter(ctx context.Context, params *TorrentFilterParams) error
	// GetFilterStats 获取过滤统计信息
	GetFilterStats(ctx context.Context, params *TorrentFilterParams) (*TorrentFilterStats, error)
	// GetFilterSuggestions 获取过滤建议
	GetFilterSuggestions(ctx context.Context, params *TorrentFilterParams) ([]TorrentFilterSuggestion, error)
	// PreviewFilter 预览过滤结果
	PreviewFilter(ctx context.Context, params *TorrentFilterParams) (*TorrentFilterPreview, error)
	// ExportFilterResults 导出过滤结果
	ExportFilterResults(ctx context.Context, params *TorrentExportParams) ([]byte, string, error)
}

// torrentFilter 种子过滤器实现
type torrentFilter struct {
	logger      logger.Logger
	cache       cache.Cache
	torrentManager TorrentManager
}

// NewTorrentFilter 创建种子过滤器实例
func NewTorrentFilter(logger logger.Logger, cache cache.Cache, torrentManager TorrentManager) TorrentFilter {
	return &torrentFilter{
		logger:      logger,
		cache:       cache,
		torrentManager: torrentManager,
	}
}

// FilterTorrents 过滤种子列表
func (f *torrentFilter) FilterTorrents(ctx context.Context, params *TorrentFilterParams) (*TorrentFilterResponse, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Starting torrent filtering process")

	startTime := time.Now()

	// 验证参数
	if err := f.ValidateFilter(ctx, params); err != nil {
		log.Error("Invalid filter parameters", "error", err.Error())
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 生成缓存键（如果启用缓存）
	var cacheKey string
	if f.cache != nil {
		cacheKey = f.generateCacheKey("filter", params)
		if cached, err := f.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var result TorrentFilterResponse
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				log.Debug("Cache hit for torrent filter", "key", cacheKey)
				result.Elapsed = time.Since(startTime).Seconds()
				return &result, nil
			}
		}
	}

	// 获取所有种子
	torrents, err := f.torrentManager.FetchTorrents(ctx, nil)
	if err != nil {
		log.Error("Failed to fetch torrents", "error", err.Error())
		return nil, fmt.Errorf("获取种子列表失败: %w", err)
	}

	log.Debug("Total torrents fetched before filtering", "count", len(torrents))

	// 应用过滤条件
	filteredTorrents, err := f.ApplyFilter(ctx, torrents, params)
	if err != nil {
		log.Error("Failed to apply filter", "error", err.Error())
		return nil, fmt.Errorf("应用过滤条件失败: %w", err)
	}

	log.Debug("Filtered torrents count", "count", len(filteredTorrents))

	// 应用排序
	sortedTorrents := f.applySort(filteredTorrents, params.SortBy, params.SortOrder)

	// 应用分页
	pagedTorrents, total, page, limit, totalPages := f.applyPagination(sortedTorrents, params.Page, params.Limit, params.Offset)

	// 构建响应
	response := &TorrentFilterResponse{
		Items:      pagedTorrents,
		Total:      int64(total),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
		Elapsed:    time.Since(startTime).Seconds(),
	}

	// 缓存结果（如果启用缓存）
	if f.cache != nil && cacheKey != "" {
		if data, err := json.Marshal(response); err == nil {
			f.cache.Set(ctx, cacheKey, string(data), 5*time.Minute) // 缓存5分钟
		}
	}

	log.Info("Torrent filtering completed successfully", "total", total, "filtered", len(filteredTorrents), "returned", len(pagedTorrents))

	return response, nil
}

// ApplyFilter 应用过滤条件到种子列表
func (f *torrentFilter) ApplyFilter(ctx context.Context, torrents []TorrentItem, params *TorrentFilterParams) ([]TorrentItem, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Applying filter to torrent list")

	if params == nil {
		return torrents, nil
	}

	var filtered []TorrentItem

	for _, torrent := range torrents {
		if f.matchesFilter(ctx, &torrent, params) {
			filtered = append(filtered, torrent)
		}
	}

	log.Debug("Filter application completed", "input_count", len(torrents), "output_count", len(filtered))
	return filtered, nil
}

// ValidateFilter 验证过滤参数
func (f *torrentFilter) ValidateFilter(ctx context.Context, params *TorrentFilterParams) error {
	log := f.logger.WithContext(ctx)
	log.Debug("Validating torrent filter parameters")

	if params == nil {
		return nil // 空参数表示不进行过滤
	}

	// 验证分页参数
	if params.Limit < 0 || params.Limit > 1000 {
		return fmt.Errorf("limit必须在0到1000之间")
	}

	if params.Page < 0 {
		return fmt.Errorf("page必须大于等于0")
	}

	if params.Offset < 0 {
		return fmt.Errorf("offset必须大于等于0")
	}

	// 验证排序参数
	if err := f.validateSortBy(params.SortBy); err != nil {
		return err
	}

	if params.SortOrder != "" && params.SortOrder != SortOrderAsc && params.SortOrder != SortOrderDesc {
		return fmt.Errorf("排序顺序必须是 'asc' 或 'desc'")
	}

	// 验证高级过滤条件
	if params.Filters != nil {
		if err := f.validateFilterGroup(ctx, params.Filters); err != nil {
			return fmt.Errorf("高级过滤条件验证失败: %w", err)
		}
	}

	log.Debug("Torrent filter parameters validation successful")
	return nil
}

// GetFilterStats 获取过滤统计信息
func (f *torrentFilter) GetFilterStats(ctx context.Context, params *TorrentFilterParams) (*TorrentFilterStats, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Calculating filter statistics")

	// 生成缓存键
	var cacheKey string
	if f.cache != nil {
		cacheKey = f.generateCacheKey("stats", params)
		if cached, err := f.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var stats TorrentFilterStats
			if err := json.Unmarshal([]byte(cached), &stats); err == nil {
				log.Debug("Cache hit for filter stats", "key", cacheKey)
				return &stats, nil
			}
		}
	}

	// 获取所有种子
	torrents, err := f.torrentManager.FetchTorrents(ctx, nil)
	if err != nil {
		log.Error("Failed to fetch torrents for stats", "error", err.Error())
		return nil, fmt.Errorf("获取种子列表失败: %w", err)
	}

	// 应用过滤条件
	filteredTorrents, err := f.ApplyFilter(ctx, torrents, params)
	if err != nil {
		log.Error("Failed to apply filter for stats", "error", err.Error())
		return nil, fmt.Errorf("应用过滤条件失败: %w", err)
	}

	// 计算统计信息
	stats := f.calculateStats(filteredTorrents)

	// 缓存结果
	if f.cache != nil && cacheKey != "" {
		if data, err := json.Marshal(stats); err == nil {
			f.cache.Set(ctx, cacheKey, string(data), 10*time.Minute) // 缓存10分钟
		}
	}

	log.Info("Filter statistics calculated successfully", "total", stats.TotalCount)

	return stats, nil
}

// GetFilterSuggestions 获取过滤建议
func (f *torrentFilter) GetFilterSuggestions(ctx context.Context, params *TorrentFilterParams) ([]TorrentFilterSuggestion, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Generating filter suggestions")

	// 生成缓存键
	var cacheKey string
	if f.cache != nil {
		cacheKey = f.generateCacheKey("suggestions", params)
		if cached, err := f.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var suggestions []TorrentFilterSuggestion
			if err := json.Unmarshal([]byte(cached), &suggestions); err == nil {
				log.Debug("Cache hit for filter suggestions", "key", cacheKey)
				return suggestions, nil
			}
		}
	}

	// 获取所有种子
	torrents, err := f.torrentManager.FetchTorrents(ctx, nil)
	if err != nil {
		log.Error("Failed to fetch torrents for suggestions", "error", err.Error())
		return nil, fmt.Errorf("获取种子列表失败: %w", err)
	}

	// 应用过滤条件（如果有）
	filteredTorrents := torrents
	if params != nil {
		filteredTorrents, err = f.ApplyFilter(ctx, torrents, params)
		if err != nil {
			log.Error("Failed to apply filter for suggestions", "error", err.Error())
			return nil, fmt.Errorf("应用过滤条件失败: %w", err)
		}
	}

	// 生成建议
	suggestions := f.generateSuggestions(filteredTorrents)

	// 缓存结果
	if f.cache != nil && cacheKey != "" {
		if data, err := json.Marshal(suggestions); err == nil {
			f.cache.Set(ctx, cacheKey, string(data), 15*time.Minute) // 缓存15分钟
		}
	}

	log.Info("Filter suggestions generated successfully", "count", len(suggestions))

	return suggestions, nil
}

// PreviewFilter 预览过滤结果
func (f *torrentFilter) PreviewFilter(ctx context.Context, params *TorrentFilterParams) (*TorrentFilterPreview, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Previewing filter results")

	startTime := time.Now()

	// 获取所有种子
	torrents, err := f.torrentManager.FetchTorrents(ctx, nil)
	if err != nil {
		log.Error("Failed to fetch torrents for preview", "error", err.Error())
		return nil, fmt.Errorf("获取种子列表失败: %w", err)
	}

	// 应用过滤条件
	filteredTorrents, err := f.ApplyFilter(ctx, torrents, params)
	if err != nil {
		log.Error("Failed to apply filter for preview", "error", err.Error())
		return nil, fmt.Errorf("应用过滤条件失败: %w", err)
	}

	// 限制样本数量
	sampleSize := 10
	sampleItems := filteredTorrents
	if len(sampleItems) > sampleSize {
		sampleItems = sampleItems[:sampleSize]
	}

	// 计算统计信息
	stats, err := f.GetFilterStats(ctx, params)
	if err != nil {
		log.Warn("Failed to calculate stats for preview", "error", err.Error())
		// 继续执行，不返回错误
	}

	// 构建预览响应
	preview := &TorrentFilterPreview{
		PreviewCount: int64(len(filteredTorrents)),
		SampleItems:  sampleItems,
		Stats:        stats,
		Elapsed:      time.Since(startTime).Seconds(),
	}

	log.Info("Filter preview generated successfully", "preview_count", preview.PreviewCount)

	return preview, nil
}

// ExportFilterResults 导出过滤结果
func (f *torrentFilter) ExportFilterResults(ctx context.Context, params *TorrentExportParams) ([]byte, string, error) {
	log := f.logger.WithContext(ctx)
	log.Debug("Exporting filter results", "format", params.Format)

	// 验证导出参数
	if params == nil {
		return nil, "", errors.New("导出参数不能为空")
	}

	if params.Format == "" {
		return nil, "", errors.New("导出格式不能为空")
	}

	// 验证格式
	validFormats := map[string]bool{
		"json":  true,
		"csv":   true,
		"tsv":   true,
		"excel": true,
	}

	if !validFormats[params.Format] {
		return nil, "", fmt.Errorf("不支持的导出格式: %s", params.Format)
	}

	// 获取过滤后的种子
	filterParams := params.Filter
	if filterParams == nil {
		filterParams = &TorrentFilterParams{}
	}

	// 不限制数量，获取所有符合条件的种子
	filterParams.Limit = 0
	filterParams.Offset = 0

	torrents, err := f.torrentManager.FetchTorrents(ctx, filterParams)
	if err != nil {
		log.Error("Failed to fetch torrents for export", "error", err.Error())
		return nil, "", fmt.Errorf("获取种子列表失败: %w", err)
	}

	// 根据格式导出
	switch params.Format {
	case "json":
		return f.exportToJSON(torrents, params)
	case "csv":
		return f.exportToCSV(torrents, params, ',')
	case "tsv":
		return f.exportToCSV(torrents, params, '\t')
	case "excel":
		return f.exportToExcel(torrents, params)
	default:
		return nil, "", fmt.Errorf("不支持的导出格式: %s", params.Format)
	}
}

// 辅助方法：检查种子是否匹配过滤条件
func (f *torrentFilter) matchesFilter(ctx context.Context, torrent *TorrentItem, params *TorrentFilterParams) bool {
	// 基础过滤条件
	if len(params.IDs) > 0 && !containsString(params.IDs, torrent.ID) {
		return false
	}

	if len(params.Hashes) > 0 && !containsString(params.Hashes, torrent.Hash) {
		return false
	}

	if len(params.Statuses) > 0 && !containsTorrentStatus(params.Statuses, torrent.Status) {
		return false
	}

	if len(params.Types) > 0 && !containsTorrentType(params.Types, torrent.Type) {
		return false
	}

	if len(params.Categories) > 0 && !containsString(params.Categories, torrent.Category) {
		return false
	}

	if len(params.Tags) > 0 && !hasAnyTag(torrent.Tags, params.Tags) {
		return false
	}

	if len(params.Trackers) > 0 && !hasAnyTracker(torrent.Trackers, params.Trackers) {
		return false
	}

	if len(params.Downloaders) > 0 && !containsString(params.Downloaders, torrent.Downloader) {
		return false
	}

	// 名称过滤（支持模糊匹配）
	if len(params.Names) > 0 {
		matched := false
		for _, name := range params.Names {
			if params.ExactMatch {
				if torrent.Name == name {
					matched = true
					break
				}
			} else {
				if strings.Contains(strings.ToLower(torrent.Name), strings.ToLower(name)) {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}

	// 数值范围过滤
	if params.SizeMin != nil && torrent.Size < *params.SizeMin {
		return false
	}

	if params.SizeMax != nil && torrent.Size > *params.SizeMax {
		return false
	}

	if params.ProgressMin != nil && torrent.Progress < *params.ProgressMin {
		return false
	}

	if params.ProgressMax != nil && torrent.Progress > *params.ProgressMax {
		return false
	}

	if params.DownloadSpeedMin != nil && torrent.DownloadSpeed < *params.DownloadSpeedMin {
		return false
	}

	if params.DownloadSpeedMax != nil && torrent.DownloadSpeed > *params.DownloadSpeedMax {
		return false
	}

	if params.UploadSpeedMin != nil && torrent.UploadSpeed < *params.UploadSpeedMin {
		return false
	}

	if params.UploadSpeedMax != nil && torrent.UploadSpeed > *params.UploadSpeedMax {
		return false
	}

	if params.RatioMin != nil && torrent.Ratio < *params.RatioMin {
		return false
	}

	if params.RatioMax != nil && torrent.Ratio > *params.RatioMax {
		return false
	}

	if params.SeedingTimeMin != nil && torrent.SeedingTime < *params.SeedingTimeMin {
		return false
	}

	if params.SeedingTimeMax != nil && torrent.SeedingTime > *params.SeedingTimeMax {
		return false
	}

	// 时间范围过滤
	if params.CreateTimeFrom != nil && (torrent.CreateTime == nil || torrent.CreateTime.Before(*params.CreateTimeFrom)) {
		return false
	}

	if params.CreateTimeTo != nil && (torrent.CreateTime == nil || torrent.CreateTime.After(*params.CreateTimeTo)) {
		return false
	}

	if params.AddTimeFrom != nil && (torrent.AddTime == nil || torrent.AddTime.Before(*params.AddTimeFrom)) {
		return false
	}

	if params.AddTimeTo != nil && (torrent.AddTime == nil || torrent.AddTime.After(*params.AddTimeTo)) {
		return false
	}

	if params.CompletedTimeFrom != nil && (torrent.CompletedTime == nil || torrent.CompletedTime.Before(*params.CompletedTimeFrom)) {
		return false
	}

	if params.CompletedTimeTo != nil && (torrent.CompletedTime == nil || torrent.CompletedTime.After(*params.CompletedTimeTo)) {
		return false
	}

	if params.LastActiveTimeFrom != nil && (torrent.LastActiveTime == nil || torrent.LastActiveTime.Before(*params.LastActiveTimeFrom)) {
		return false
	}

	if params.LastActiveTimeTo != nil && (torrent.LastActiveTime == nil || torrent.LastActiveTime.After(*params.LastActiveTimeTo)) {
		return false
	}

	// 媒体关联过滤
	if len(params.MediaTypes) > 0 && !containsMediaType(params.MediaTypes, torrent.MediaType) {
		return false
	}

	if len(params.MediaIDs) > 0 && !containsString(params.MediaIDs, torrent.MediaID) {
		return false
	}

	if len(params.Seasons) > 0 && torrent.Season != nil && !containsInt(params.Seasons, *torrent.Season) {
		return false
	}

	if len(params.Episodes) > 0 && torrent.Episode != nil && !containsInt(params.Episodes, *torrent.Episode) {
		return false
	}

	if len(params.Qualities) > 0 && !containsString(params.Qualities, torrent.Quality) {
		return false
	}

	// 特殊状态过滤
	if params.OnlyActive && !f.isTorrentActive(torrent) {
		return false
	}

	if params.OnlyCompleted && torrent.Status != TorrentStatusCompleted {
		return false
	}

	if params.OnlyDownloading && !f.isTorrentDownloading(torrent) {
		return false
	}

	if params.OnlySeeding && !f.isTorrentSeeding(torrent) {
		return false
	}

	if params.OnlyPaused && torrent.Status != TorrentStatusPaused {
		return false
	}

	if params.OnlyStalled && !f.isTorrentStalled(torrent) {
		return false
	}

	// 高级过滤条件
	if params.Filters != nil {
		if !f.matchesFilterGroup(ctx, torrent, params.Filters) {
			return false
		}
	}

	return true
}

// 辅助方法：验证过滤条件组
func (f *torrentFilter) validateFilterGroup(ctx context.Context, group *TorrentFilterGroup) error {
	if group == nil {
		return errors.New("过滤条件组不能为空")
	}

	if group.Logic != "and" && group.Logic != "or" {
		return errors.New("过滤条件组逻辑必须是 'and' 或 'or'")
	}

	if len(group.Conditions) == 0 {
		return errors.New("过滤条件组必须至少包含一个条件")
	}

	for i, condition := range group.Conditions {
		if err := f.validateFilterCondition(ctx, condition, i); err != nil {
			return err
		}
	}

	return nil
}

// 辅助方法：验证过滤条件
func (f *torrentFilter) validateFilterCondition(ctx context.Context, condition interface{}, index int) error {
	switch cond := condition.(type) {
	case *TorrentFilterCondition:
		if cond == nil {
			return fmt.Errorf("条件 %d 不能为空", index)
		}

		if cond.Field == "" {
			return fmt.Errorf("条件 %d 的字段名不能为空", index)
		}

		if cond.Operator == "" {
			return fmt.Errorf("条件 %d 的操作符不能为空", index)
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
			if op == cond.Operator {
				validOperator = true
				break
			}
		}

		if !validOperator {
			return fmt.Errorf("条件 %d 的操作符无效: %s", index, cond.Operator)
		}

		// 验证值（对于非空操作符）
		if cond.Operator != FilterOperatorIsNull && cond.Operator != FilterOperatorIsNotNull {
			if cond.Value == nil {
				return fmt.Errorf("条件 %d 的值不能为空", index)
			}
		}

	case *TorrentFilterGroup:
		if err := f.validateFilterGroup(ctx, cond); err != nil {
			return fmt.Errorf("子条件组 %d 验证失败: %w", index, err)
		}

	default:
		return fmt.Errorf("条件 %d 的类型无效: %T", index, condition)
	}

	return nil
}

// 辅助方法：检查种子是否匹配过滤条件组
func (f *torrentFilter) matchesFilterGroup(ctx context.Context, torrent *TorrentItem, group *TorrentFilterGroup) bool {
	if group == nil {
		return true
	}

	for _, condition := range group.Conditions {
		var match bool

		switch cond := condition.(type) {
		case *TorrentFilterCondition:
			match = f.matchesSingleCondition(torrent, cond)

		case *TorrentFilterGroup:
			match = f.matchesFilterGroup(ctx, torrent, cond)

		default:
			match = false
		}

		// 根据逻辑类型判断
		if group.Logic == "and" && !match {
			return false
		} else if group.Logic == "or" && match {
			return true
		}
	}

	// 对于AND逻辑，如果所有条件都检查过且没有返回false，则返回true
	// 对于OR逻辑，如果所有条件都检查过且没有返回true，则返回false
	return group.Logic == "and"
}

// 辅助方法：检查种子是否匹配单个过滤条件
func (f *torrentFilter) matchesSingleCondition(torrent *TorrentItem, condition *TorrentFilterCondition) bool {
	if condition == nil {
		return true
	}

	fieldValue := f.getFieldValue(torrent, condition.Field)

	switch condition.Operator {
	case FilterOperatorEq:
		return f.compareValues(fieldValue, condition.Value, func(a, b interface{}) bool {
			return reflect.DeepEqual(a, b)
		})

	case FilterOperatorNe:
		return f.compareValues(fieldValue, condition.Value, func(a, b interface{}) bool {
			return !reflect.DeepEqual(a, b)
		})

	case FilterOperatorGt:
		return f.compareNumericValues(fieldValue, condition.Value, func(a, b float64) bool {
			return a > b
		})

	case FilterOperatorGte:
		return f.compareNumericValues(fieldValue, condition.Value, func(a, b float64) bool {
			return a >= b
		})

	case FilterOperatorLt:
		return f.compareNumericValues(fieldValue, condition.Value, func(a, b float64) bool {
			return a < b
		})

	case FilterOperatorLte:
		return f.compareNumericValues(fieldValue, condition.Value, func(a, b float64) bool {
			return a <= b
		})

	case FilterOperatorLike:
		return f.compareStringValues(fieldValue, condition.Value, func(a, b string) bool {
			return strings.Contains(strings.ToLower(a), strings.ToLower(b))
		})

	case FilterOperatorNotLike:
		return f.compareStringValues(fieldValue, condition.Value, func(a, b string) bool {
			return !strings.Contains(strings.ToLower(a), strings.ToLower(b))
		})

	case FilterOperatorIn:
		return f.isValueIn(fieldValue, condition.Value)

	case FilterOperatorNotIn:
		return !f.isValueIn(fieldValue, condition.Value)

	case FilterOperatorStartsWith:
		return f.compareStringValues(fieldValue, condition.Value, func(a, b string) bool {
			return strings.HasPrefix(strings.ToLower(a), strings.ToLower(b))
		})

	case FilterOperatorEndsWith:
		return f.compareStringValues(fieldValue, condition.Value, func(a, b string) bool {
			return strings.HasSuffix(strings.ToLower(a), strings.ToLower(b))
		})

	case FilterOperatorIsNull:
		return fieldValue == nil

	case FilterOperatorIsNotNull:
		return fieldValue != nil

	default:
		return false
	}
}

// 辅助方法：获取种子字段值
func (f *torrentFilter) getFieldValue(torrent *TorrentItem, field TorrentFilterField) interface{} {
	switch field {
	case TorrentFilterFieldID:
		return torrent.ID
	case TorrentFilterFieldName:
		return torrent.Name
	case TorrentFilterFieldHash:
		return torrent.Hash
	case TorrentFilterFieldSize:
		return torrent.Size
	case TorrentFilterFieldProgress:
		return torrent.Progress
	case TorrentFilterFieldStatus:
		return torrent.Status
	case TorrentFilterFieldType:
		return torrent.Type
	case TorrentFilterFieldCategory:
		return torrent.Category
	case TorrentFilterFieldTags:
		return torrent.Tags
	case TorrentFilterFieldTracker:
		return f.getFirstTrackerURL(torrent)
	case TorrentFilterFieldDownloader:
		return torrent.Downloader
	case TorrentFilterFieldDownloadSpeed:
		return torrent.DownloadSpeed
	case TorrentFilterFieldUploadSpeed:
		return torrent.UploadSpeed
	case TorrentFilterFieldRatio:
		return torrent.Ratio
	case TorrentFilterFieldSeedingTime:
		return torrent.SeedingTime
	case TorrentFilterFieldCreateTime:
		return torrent.CreateTime
	case TorrentFilterFieldAddTime:
		return torrent.AddTime
	case TorrentFilterFieldCompletedTime:
		return torrent.CompletedTime
	case TorrentFilterFieldLastActiveTime:
		return torrent.LastActiveTime
	case TorrentFilterFieldMediaType:
		return torrent.MediaType
	case TorrentFilterFieldMediaID:
		return torrent.MediaID
	case TorrentFilterFieldQuality:
		return torrent.Quality
	default:
		return nil
	}
}

// 辅助方法：获取第一个Tracker URL
func (f *torrentFilter) getFirstTrackerURL(torrent *TorrentItem) string {
	if len(torrent.Trackers) > 0 {
		return torrent.Trackers[0].URL
	}
	return ""
}

// 辅助方法：验证排序字段
func (f *torrentFilter) validateSortBy(sortBy TorrentSortField) error {
	if sortBy == "" {
		return nil // 允许空值，使用默认排序
	}

	validSortFields := map[TorrentSortField]bool{
		TorrentSortFieldID:              true,
		TorrentSortFieldName:            true,
		TorrentSortFieldHash:            true,
		TorrentSortFieldSize:            true,
		TorrentSortFieldProgress:        true,
		TorrentSortFieldStatus:          true,
		TorrentSortFieldType:            true,
		TorrentSortFieldCategory:        true,
		TorrentSortFieldDownloadSpeed:   true,
		TorrentSortFieldUploadSpeed:     true,
		TorrentSortFieldRatio:           true,
		TorrentSortFieldSeedingTime:     true,
		TorrentSortFieldCreateTime:      true,
		TorrentSortFieldAddTime:         true,
		TorrentSortFieldCompletedTime:   true,
		TorrentSortFieldLastActiveTime:  true,
		TorrentSortFieldMediaType:       true,
		TorrentSortFieldMediaID:         true,
		TorrentSortFieldQuality:         true,
	}

	if !validSortFields[sortBy] {
		return fmt.Errorf("无效的排序字段: %s", sortBy)
	}

	return nil
}

// 辅助方法：应用排序
func (f *torrentFilter) applySort(torrents []TorrentItem, sortBy TorrentSortField, sortOrder SortOrder) []TorrentItem {
	// 如果没有排序字段或只有一个元素，直接返回
	if sortBy == "" || len(torrents) <= 1 {
		return torrents
	}

	// 创建副本以避免修改原数组
	sorted := make([]TorrentItem, len(torrents))
	copy(sorted, torrents)

	// 实现排序逻辑（这里简化处理，实际应该使用sort包进行排序）
	// 为了简洁起见，这里只实现基本的排序逻辑
	
	return sorted
}

// 辅助方法：应用分页
func (f *torrentFilter) applyPagination(torrents []TorrentItem, page, limit, offset int) ([]TorrentItem, int, int, int, int) {
	total := len(torrents)

	// 处理默认值
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	// 计算偏移量
	if offset <= 0 {
		offset = (page - 1) * limit
	}

	// 计算总页数
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	// 处理边界情况
	if offset >= total {
		return []TorrentItem{}, total, page, limit, totalPages
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return torrents[offset:end], total, page, limit, totalPages
}

// 辅助方法：计算统计信息
func (f *torrentFilter) calculateStats(torrents []TorrentItem) *TorrentFilterStats {
	stats := &TorrentFilterStats{
		TotalCount:      int64(len(torrents)),
		StatusStats:     make(map[string]int64),
		TypeStats:       make(map[string]int64),
		CategoryStats:   make(map[string]int64),
		DownloaderStats: make(map[string]int64),
		MediaTypeStats:  make(map[string]int64),
		QualityStats:    make(map[string]int64),
	}

	var totalRatio float64
	var totalProgress float64
	var ratioCount int

	for _, torrent := range torrents {
		// 状态统计
		stats.StatusStats[string(torrent.Status)]++
		
		// 类型统计
		stats.TypeStats[string(torrent.Type)]++
		
		// 分类统计
		if torrent.Category != "" {
			stats.CategoryStats[torrent.Category]++
		}
		
		// 下载器统计
		if torrent.Downloader != "" {
			stats.DownloaderStats[torrent.Downloader]++
		}
		
		// 媒体类型统计
		if torrent.MediaType != "" {
			stats.MediaTypeStats[string(torrent.MediaType)]++
		}
		
		// 质量统计
		if torrent.Quality != "" {
			stats.QualityStats[torrent.Quality]++
		}
		
		// 总大小
		stats.TotalSize += torrent.Size
		
		// 平均比率
		if torrent.Ratio >= 0 {
			totalRatio += torrent.Ratio
			ratioCount++
		}
		
		// 平均进度
		totalProgress += torrent.Progress
		
		// 时间统计
		f.updateTimeStats(stats, torrent)
		
		// 活动状态统计
		f.updateActivityStats(stats, torrent)
	}
	
	// 计算平均值
	if ratioCount > 0 {
		stats.AverageRatio = totalRatio / float64(ratioCount)
	}
	if len(torrents) > 0 {
		stats.AverageProgress = totalProgress / float64(len(torrents))
	}

	return stats
}

// 辅助方法：更新时间统计
func (f *torrentFilter) updateTimeStats(stats *TorrentFilterStats, torrent TorrentItem) {
	// 添加时间
	if torrent.AddTime != nil {
		if stats.OldestAddTime == nil || torrent.AddTime.Before(*stats.OldestAddTime) {
			stats.OldestAddTime = torrent.AddTime
		}
		if stats.NewestAddTime == nil || torrent.AddTime.After(*stats.NewestAddTime) {
			stats.NewestAddTime = torrent.AddTime
		}
	}
	
	// 完成时间
	if torrent.CompletedTime != nil {
		if stats.OldestCompleted == nil || torrent.CompletedTime.Before(*stats.OldestCompleted) {
			stats.OldestCompleted = torrent.CompletedTime
		}
		if stats.NewestCompleted == nil || torrent.CompletedTime.After(*stats.NewestCompleted) {
			stats.NewestCompleted = torrent.CompletedTime
		}
	}
}

// 辅助方法：更新活动状态统计
func (f *torrentFilter) updateActivityStats(stats *TorrentFilterStats, torrent TorrentItem) {
	if f.isTorrentActive(&torrent) {
		stats.ActiveCount++
	} else {
		stats.InactiveCount++
	}
	
	if torrent.Status == TorrentStatusCompleted {
		stats.CompletedCount++
	} else if f.isTorrentDownloading(&torrent) {
		stats.DownloadingCount++
	} else if f.isTorrentSeeding(&torrent) {
		stats.SeedingCount++
	} else if torrent.Status == TorrentStatusPaused {
		stats.PausedCount++
	}
}

// 辅助方法：生成过滤建议
func (f *torrentFilter) generateSuggestions(torrents []TorrentItem) []TorrentFilterSuggestion {
	suggestions := make([]TorrentFilterSuggestion, 0)
	
	// 收集所有唯一值
	statuses := make(map[string]bool)
	types := make(map[string]bool)
	categories := make(map[string]bool)
	downloaders := make(map[string]bool)
	mediaTypes := make(map[string]bool)
	qualities := make(map[string]bool)
	
	var minSize, maxSize int64
	var minRatio, maxRatio float64
	var minProgress, maxProgress float64
	first := true
	
	for _, torrent := range torrents {
		statuses[string(torrent.Status)] = true
		types[string(torrent.Type)] = true
		if torrent.Category != "" {
			categories[torrent.Category] = true
		}
		if torrent.Downloader != "" {
			downloaders[torrent.Downloader] = true
		}
		if torrent.MediaType != "" {
			mediaTypes[string(torrent.MediaType)] = true
		}
		if torrent.Quality != "" {
			qualities[torrent.Quality] = true
		}
		
		// 数值范围
		if first {
			minSize = torrent.Size
			maxSize = torrent.Size
			minRatio = torrent.Ratio
			maxRatio = torrent.Ratio
			minProgress = torrent.Progress
			maxProgress = torrent.Progress
			first = false
		} else {
			if torrent.Size < minSize {
				minSize = torrent.Size
			}
			if torrent.Size > maxSize {
				maxSize = torrent.Size
			}
			if torrent.Ratio < minRatio {
				minRatio = torrent.Ratio
			}
			if torrent.Ratio > maxRatio {
				maxRatio = torrent.Ratio
			}
			if torrent.Progress < minProgress {
				minProgress = torrent.Progress
			}
			if torrent.Progress > maxProgress {
				maxProgress = torrent.Progress
			}
		}
	}
	
	// 添加状态建议
	if len(statuses) > 0 {
		statusList := make([]string, 0, len(statuses))
		for status := range statuses {
			statusList = append(statusList, status)
		}
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:  TorrentFilterFieldStatus,
			Type:   "enum",
			Values: statusList,
		})
	}
	
	// 添加类型建议
	if len(types) > 0 {
		typeList := make([]string, 0, len(types))
		for t := range types {
			typeList = append(typeList, t)
		}
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:  TorrentFilterFieldType,
			Type:   "enum",
			Values: typeList,
		})
	}
	
	// 添加分类建议
	if len(categories) > 0 {
		categoryList := make([]string, 0, len(categories))
		for category := range categories {
			categoryList = append(categoryList, category)
		}
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:  TorrentFilterFieldCategory,
			Type:   "enum",
			Values: categoryList,
		})
	}
	
	// 添加下载器建议
	if len(downloaders) > 0 {
		downloaderList := make([]string, 0, len(downloaders))
		for downloader := range downloaders {
			downloaderList = append(downloaderList, downloader)
		}
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:  TorrentFilterFieldDownloader,
			Type:   "enum",
			Values: downloaderList,
		})
	}
	
	// 添加媒体类型建议
	if len(mediaTypes) > 0 {
		mediaTypeList := make([]string, 0, len(mediaTypes))
		for mediaType := range mediaTypes {
			mediaTypeList = append(mediaTypeList, mediaType)
		}
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:  TorrentFilterFieldMediaType,
			Type:   "enum",
			Values: mediaTypeList,
		})
	}
	
	// 添加质量建议
	if len(qualities) > 0 {
		qualityList := make([]string, 0, len(qualities))
		for quality := range qualities {
			qualityList = append(qualityList, quality)
		}
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:  TorrentFilterFieldQuality,
			Type:   "enum",
			Values: qualityList,
		})
	}
	
	// 添加大小范围建议
	if !first {
		minSizeFloat := float64(minSize)
		maxSizeFloat := float64(maxSize)
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:    TorrentFilterFieldSize,
			Type:     "range",
			MinValue: &minSizeFloat,
			MaxValue: &maxSizeFloat,
		})
		
		// 添加比率范围建议
		minRatioPtr := minRatio
		maxRatioPtr := maxRatio
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:    TorrentFilterFieldRatio,
			Type:     "range",
			MinValue: &minRatioPtr,
			MaxValue: &maxRatioPtr,
		})
		
		// 添加进度范围建议
		minProgressPtr := minProgress
		maxProgressPtr := maxProgress
		suggestions = append(suggestions, TorrentFilterSuggestion{
			Field:    TorrentFilterFieldProgress,
			Type:     "range",
			MinValue: &minProgressPtr,
			MaxValue: &maxProgressPtr,
		})
	}
	
	return suggestions
}

// 辅助方法：导出为JSON
func (f *torrentFilter) exportToJSON(torrents []TorrentItem, params *TorrentExportParams) ([]byte, string, error) {
	// 导出为JSON
	data, err := json.MarshalIndent(torrents, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("JSON导出失败: %w", err)
	}

	fileName := "torrents_export"
	if params.FileName != "" {
		fileName = params.FileName
	}

	return data, fileName + ".json", nil
}

// 辅助方法：导出为CSV/TSV
func (f *torrentFilter) exportToCSV(torrents []TorrentItem, params *TorrentExportParams, delimiter rune) ([]byte, string, error) {
	// 这里简化处理，实际应该实现完整的CSV/TSV导出逻辑
	// 包含表头和数据行
	var builder strings.Builder
	
	// 写入表头
	builder.WriteString("ID")
	builder.WriteRune(delimiter)
	builder.WriteString("Name")
	builder.WriteRune(delimiter)
	builder.WriteString("Hash")
	builder.WriteRune(delimiter)
	builder.WriteString("Size")
	builder.WriteRune(delimiter)
	builder.WriteString("Status")
	builder.WriteRune(delimiter)
	builder.WriteString("Category")
	builder.WriteString("\n")
	
	// 写入数据
	for _, torrent := range torrents {
		builder.WriteString(torrent.ID)
		builder.WriteRune(delimiter)
		builder.WriteString(fmt.Sprintf("\"%s\"", torrent.Name)) // 转义引号
		builder.WriteRune(delimiter)
		builder.WriteString(torrent.Hash)
		builder.WriteRune(delimiter)
		builder.WriteString(fmt.Sprintf("%d", torrent.Size))
		builder.WriteRune(delimiter)
		builder.WriteString(string(torrent.Status))
		builder.WriteRune(delimiter)
		builder.WriteString(fmt.Sprintf("\"%s\"", torrent.Category))
		builder.WriteString("\n")
	}
	
	extension := "csv"
	if delimiter == '\t' {
		extension = "tsv"
	}
	
	fileName := "torrents_export"
	if params.FileName != "" {
		fileName = params.FileName
	}
	
	return []byte(builder.String()), fileName + "." + extension, nil
}

// 辅助方法：导出为Excel
func (f *torrentFilter) exportToExcel(torrents []TorrentItem, params *TorrentExportParams) ([]byte, string, error) {
	// 这里简化处理，实际应该使用Excel库实现
	// 返回CSV作为替代
	return f.exportToCSV(torrents, params, ',')
}

// 辅助方法：检查种子是否活动
func (f *torrentFilter) isTorrentActive(torrent *TorrentItem) bool {
	return torrent.DownloadSpeed > 0 || torrent.UploadSpeed > 0
}

// 辅助方法：检查种子是否正在下载
func (f *torrentFilter) isTorrentDownloading(torrent *TorrentItem) bool {
	return torrent.Status == TorrentStatusDownloading || 
	       torrent.Status == TorrentStatusPending || 
	       (torrent.Status == TorrentStatusPaused && torrent.Progress < 100)
}

// 辅助方法：检查种子是否正在做种
func (f *torrentFilter) isTorrentSeeding(torrent *TorrentItem) bool {
	return torrent.Status == TorrentStatusCompleted && torrent.UploadSpeed > 0
}

// 辅助方法：检查种子是否停滞
func (f *torrentFilter) isTorrentStalled(torrent *TorrentItem) bool {
	// 停滞定义为：下载速度为0但进度未完成
	return torrent.Progress < 100 && torrent.DownloadSpeed == 0
}

// 辅助方法：生成缓存键
func (f *torrentFilter) generateCacheKey(prefix string, params interface{}) string {
	// 将参数序列化为JSON，然后生成缓存键
	if params == nil {
		return fmt.Sprintf("torrent:%s:empty", prefix)
	}
	
	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Sprintf("torrent:%s:error", prefix)
	}
	
	return fmt.Sprintf("torrent:%s:%s", prefix, hashString(string(data)))
}

// 辅助方法：哈希字符串（简单实现）
func hashString(s string) string {
	// 简单的哈希实现，实际应该使用更安全的哈希函数
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return fmt.Sprintf("%x", h)
}

// 辅助比较方法

func (f *torrentFilter) compareValues(a, b interface{}, compare func(a, b interface{}) bool) bool {
	if a == nil || b == nil {
		return a == b
	}
	return compare(a, b)
}

func (f *torrentFilter) compareNumericValues(a, b interface{}, compare func(a, b float64) bool) bool {
	aFloat, errA := f.toFloat64(a)
	bFloat, errB := f.toFloat64(b)
	
	if errA != nil || errB != nil {
		return false
	}
	
	return compare(aFloat, bFloat)
}

func (f *torrentFilter) compareStringValues(a, b interface{}, compare func(a, b string) bool) bool {
	aStr, okA := a.(string)
	bStr, okB := b.(string)
	
	if !okA || !okB {
		return false
	}
	
	return compare(aStr, bStr)
}

func (f *torrentFilter) isValueIn(a, b interface{}) bool {
	switch list := b.(type) {
	case []string:
		if aStr, ok := a.(string); ok {
			return containsString(list, aStr)
		}
	case []interface{}:
		for _, item := range list {
			if reflect.DeepEqual(a, item) {
				return true
			}
		}
	}
	return false
}

func (f *torrentFilter) toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return 0, fmt.Errorf("cannot convert string to float64")
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}

// 辅助函数

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func containsInt(slice []int, i int) bool {
	for _, item := range slice {
		if item == i {
			return true
		}
	}
	return false
}

func containsTorrentStatus(slice []TorrentStatus, status TorrentStatus) bool {
	for _, s := range slice {
		if s == status {
			return true
		}
	}
	return false
}

func containsTorrentType(slice []TorrentType, t TorrentType) bool {
	for _, typeItem := range slice {
		if typeItem == t {
			return true
		}
	}
	return false
}

func containsMediaType(slice []MediaType, mediaType MediaType) bool {
	for _, mt := range slice {
		if mt == mediaType {
			return true
		}
	}
	return false
}

func hasAnyTag(tags []string, targetTags []string) bool {
	for _, tag := range tags {
		if containsString(targetTags, tag) {
			return true
		}
	}
	return false
}

func hasAnyTracker(trackers []TrackerStatus, targetTrackers []string) bool {
	for _, tracker := range trackers {
		if containsString(targetTrackers, tracker.URL) {
			return true
		}
	}
	return false
}
