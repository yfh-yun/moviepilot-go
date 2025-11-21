package actions

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"moviepilot-go/pkg/logger"
)

// MediaFilterHandler 媒体过滤器处理器接口
type MediaFilterHandler interface {
	// FilterMedias 过滤媒体列表
	FilterMedias(c *gin.Context)
	// ValidateFilter 验证过滤条件
	ValidateFilter(c *gin.Context)
	// GetFilterStats 获取过滤统计信息
	GetFilterStats(c *gin.Context)
	// GetFilterSuggestions 获取过滤建议
	GetFilterSuggestions(c *gin.Context)
	// GetAvailableFilterFields 获取可用的过滤字段
	GetAvailableFilterFields(c *gin.Context)
	// GetFilterPreview 获取过滤预览
	GetFilterPreview(c *gin.Context)
}

// mediaFilterHandler 媒体过滤器处理器实现
type mediaFilterHandler struct {
	mediaFilter MediaFilter
	mediaFetcher MediaFetcher
	logger       logger.Logger
}

// NewMediaFilterHandler 创建媒体过滤器处理器实例
func NewMediaFilterHandler(
	mediaFilter MediaFilter,
	mediaFetcher MediaFetcher,
	logger logger.Logger,
) MediaFilterHandler {
	return &mediaFilterHandler{
		mediaFilter: mediaFilter,
		mediaFetcher: mediaFetcher,
		logger:       logger,
	}
}

// FilterMedias 过滤媒体列表
// @Summary 过滤媒体列表
// @Description 根据指定条件过滤媒体列表，支持多条件组合、排序和分页
// @Tags media-filters
// @Accept json
// @Produce json
// @Param params body MediaFilterParams true "过滤参数"
// @Success 200 {object} MediaFilterResponse "过滤结果"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/media-filters/filter [post]
func (h *mediaFilterHandler) FilterMedias(c *gin.Context) {
	ctx := c.Request.Context()
	log := h.logger.WithContext(ctx)
	log.Debug("Received filter medias request")

	// 解析请求参数
	var params MediaFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Error("Invalid filter parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的过滤参数",
			Details: err.Error(),
		})
		return
	}

	// 获取媒体数据
	mediaResponse, err := h.mediaFetcher.GetMediaList(ctx, &GetMediaListParams{
		MediaType: params.MediaTypes,
		Status:    params.Status,
		Source:    params.Source,
	})
	if err != nil {
		log.Error("Failed to get media list", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "获取媒体列表失败",
			Details: err.Error(),
		})
		return
	}

	// 应用过滤
	result, err := h.mediaFilter.FilterMedias(ctx, mediaResponse.Medias, &params)
	if err != nil {
		log.Error("Failed to filter medias", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "过滤媒体失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Successfully filtered medias", "returned_count", len(result.Medias))
	c.JSON(http.StatusOK, result)
}

// ValidateFilter 验证过滤条件
// @Summary 验证过滤条件
// @Description 验证过滤条件的有效性，返回验证结果和错误信息（如果有）
// @Tags media-filters
// @Accept json
// @Produce json
// @Param params body MediaFilterParams true "过滤参数"
// @Success 200 {object} ValidateFilterResponse "验证结果"
// @Failure 400 {object} ErrorResponse "请求格式错误"
// @Router /api/v1/media-filters/validate [post]
func (h *mediaFilterHandler) ValidateFilter(c *gin.Context) {
	ctx := c.Request.Context()
	log := h.logger.WithContext(ctx)
	log.Debug("Received validate filter request")

	// 解析请求参数
	var params MediaFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Error("Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "请求格式错误",
			Details: err.Error(),
		})
		return
	}

	// 验证过滤条件
	result, err := h.mediaFilter.ValidateFilter(&params)
	if err != nil {
		log.Error("Failed to validate filter", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "验证过滤条件失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Successfully validated filter", "valid", result.Valid)
	c.JSON(http.StatusOK, result)
}

// GetFilterStats 获取过滤统计信息
// @Summary 获取过滤统计信息
// @Description 根据过滤条件获取媒体的统计信息，包括各种分类统计数据
// @Tags media-filters
// @Accept json
// @Produce json
// @Param params body MediaFilterParams true "过滤参数"
// @Success 200 {object} MediaFilterStats "统计信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/media-filters/stats [post]
func (h *mediaFilterHandler) GetFilterStats(c *gin.Context) {
	ctx := c.Request.Context()
	log := h.logger.WithContext(ctx)
	log.Debug("Received get filter stats request")

	// 解析请求参数
	var params MediaFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Error("Invalid filter parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的过滤参数",
			Details: err.Error(),
		})
		return
	}

	// 获取媒体数据
	mediaResponse, err := h.mediaFetcher.GetMediaList(ctx, &GetMediaListParams{
		MediaType: params.MediaTypes,
		Status:    params.Status,
		Source:    params.Source,
	})
	if err != nil {
		log.Error("Failed to get media list", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "获取媒体列表失败",
			Details: err.Error(),
		})
		return
	}

	// 获取统计信息
	stats, err := h.mediaFilter.GetFilterStats(ctx, mediaResponse.Medias, &params)
	if err != nil {
		log.Error("Failed to get filter stats", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "获取统计信息失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Successfully got filter stats", "total", stats.Total)
	c.JSON(http.StatusOK, stats)
}

// GetFilterSuggestions 获取过滤建议
// @Summary 获取过滤建议
// @Description 根据指定字段获取过滤建议值
// @Tags media-filters
// @Produce json
// @Param field query string true "过滤字段名"
// @Param limit query int false "结果限制数量" default(50)
// @Success 200 {object} MediaFilterSuggestionsResponse "过滤建议"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/media-filters/suggestions [get]
func (h *mediaFilterHandler) GetFilterSuggestions(c *gin.Context) {
	ctx := c.Request.Context()
	log := h.logger.WithContext(ctx)
	log.Debug("Received get filter suggestions request")

	// 解析请求参数
	field := MediaFilterField(c.Query("field"))
	if field == "" {
		log.Error("Field parameter is required")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "缺少字段参数",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}

	// 获取媒体数据
	mediaResponse, err := h.mediaFetcher.GetMediaList(ctx, &GetMediaListParams{Limit: 1000}) // 限制获取数量以提高性能
	if err != nil {
		log.Error("Failed to get media list", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "获取媒体列表失败",
			Details: err.Error(),
		})
		return
	}

	// 获取过滤建议
	suggestions, err := h.mediaFilter.GetFilterSuggestions(ctx, mediaResponse.Medias, field, limit)
	if err != nil {
		log.Error("Failed to get filter suggestions", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "获取过滤建议失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Successfully got filter suggestions", "field", field, "suggestion_count", len(suggestions.Suggestions))
	c.JSON(http.StatusOK, suggestions)
}

// GetAvailableFilterFields 获取可用的过滤字段
// @Summary 获取可用的过滤字段
// @Description 返回系统支持的所有媒体过滤字段信息
// @Tags media-filters
// @Produce json
// @Success 200 {object} AvailableFilterFieldsResponse "可用字段列表"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/media-filters/fields [get]
func (h *mediaFilterHandler) GetAvailableFilterFields(c *gin.Context) {
	ctx := c.Request.Context()
	log := h.logger.WithContext(ctx)
	log.Debug("Received get available filter fields request")

	// 获取媒体过滤器的实例
	filter, ok := h.mediaFilter.(*mediaFilter)
	if !ok {
		log.Error("Invalid media filter implementation")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "内部服务错误",
		})
		return
	}

	// 获取可用字段
	fields := filter.getAvailableFilterFields()

	// 构建字段信息
	var fieldInfos []FilterFieldInfo
	for _, field := range fields {
		fieldInfo := FilterFieldInfo{
			Field:     field,
			Name:      h.getFieldDisplayName(field),
			Type:      h.getFieldType(field),
			Operators: h.getFieldOperators(field),
		}
		fieldInfos = append(fieldInfos, fieldInfo)
	}

	response := AvailableFilterFieldsResponse{
		Success: true,
		Fields:  fieldInfos,
		Total:   len(fieldInfos),
	}

	log.Info("Successfully got available filter fields", "total", len(fieldInfos))
	c.JSON(http.StatusOK, response)
}

// GetFilterPreview 获取过滤预览
// @Summary 获取过滤预览
// @Description 快速预览过滤条件的效果，只返回少量数据
// @Tags media-filters
// @Accept json
// @Produce json
// @Param params body MediaFilterParams true "过滤参数"
// @Success 200 {object} MediaFilterPreviewResponse "预览结果"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/media-filters/preview [post]
func (h *mediaFilterHandler) GetFilterPreview(c *gin.Context) {
	ctx := c.Request.Context()
	log := h.logger.WithContext(ctx)
	log.Debug("Received filter preview request")

	// 解析请求参数
	var params MediaFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Error("Invalid filter parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的过滤参数",
			Details: err.Error(),
		})
		return
	}

	// 设置预览限制，最多返回50条数据
	params.Limit = 50
	params.Offset = 0

	// 获取媒体数据
	mediaResponse, err := h.mediaFetcher.GetMediaList(ctx, &GetMediaListParams{
		MediaType: params.MediaTypes,
		Status:    params.Status,
		Source:    params.Source,
	})
	if err != nil {
		log.Error("Failed to get media list", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "获取媒体列表失败",
			Details: err.Error(),
		})
		return
	}

	// 应用过滤
	filteredMedias, err := h.mediaFilter.ApplyFilter(mediaResponse.Medias, &params)
	if err != nil {
		log.Error("Failed to apply filter", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "应用过滤条件失败",
			Details: err.Error(),
		})
		return
	}

	// 排序
	filter, ok := h.mediaFilter.(*mediaFilter)
	if !ok {
		log.Error("Invalid media filter implementation")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "内部服务错误",
		})
		return
	}

	sortedMedias := filter.sortMedias(filteredMedias, params.SortBy, params.SortOrder)

	// 限制预览数量
	previewCount := 50
	if len(sortedMedias) < previewCount {
		previewCount = len(sortedMedias)
	}
	previewMedias := sortedMedias[:previewCount]

	// 构建响应
	response := MediaFilterPreviewResponse{
		Success:       true,
		PreviewMedias: previewMedias,
		TotalFiltered: len(filteredMedias),
		TotalMedias:   len(mediaResponse.Medias),
		PreviewCount:  previewCount,
		ApplyFilters:  params,
	}

	log.Info("Successfully got filter preview", 
		"total_filtered", response.TotalFiltered,
		"preview_count", response.PreviewCount)
	c.JSON(http.StatusOK, response)
}

// 辅助方法：获取字段显示名称
func (h *mediaFilterHandler) getFieldDisplayName(field MediaFilterField) string {
	fieldNames := map[MediaFilterField]string{
		MediaFilterFieldID:              "ID",
		MediaFilterFieldTitle:           "标题",
		MediaFilterFieldOriginalTitle:   "原始标题",
		MediaFilterFieldType:            "媒体类型",
		MediaFilterFieldStatus:          "状态",
		MediaFilterFieldYear:            "年份",
		MediaFilterFieldRating:          "评分",
		MediaFilterFieldVotes:           "投票数",
		MediaFilterFieldRuntime:         "时长",
		MediaFilterFieldSeasonCount:     "季数",
		MediaFilterFieldEpisodeCount:    "集数",
		MediaFilterFieldAirDate:         "播出日期",
		MediaFilterFieldFirstAirDate:    "首播日期",
		MediaFilterFieldLastAirDate:     "完结日期",
		MediaFilterFieldReleaseDate:     "发行日期",
		MediaFilterFieldOverview:        "简介",
		MediaFilterFieldGenres:          "类型",
		MediaFilterFieldTags:            "标签",
		MediaFilterFieldStudio:          "工作室",
		MediaFilterFieldDirector:        "导演",
		MediaFilterFieldCast:            "演员",
		MediaFilterFieldWriter:          "编剧",
		MediaFilterFieldIMDBID:          "IMDB ID",
		MediaFilterFieldTMDBID:          "TMDB ID",
		MediaFilterFieldTVDBID:          "TVDB ID",
		MediaFilterFieldSource:          "来源",
		MediaFilterFieldCover:           "封面",
		MediaFilterFieldBackdrop:        "背景",
		MediaFilterFieldTrailer:         "预告片",
		MediaFilterFieldLogo:            "Logo",
		MediaFilterFieldLocalStatus:     "本地状态",
		MediaFilterFieldSubscribeStatus: "订阅状态",
		MediaFilterFieldDownloadStatus:  "下载状态",
		MediaFilterFieldCreateTime:      "创建时间",
		MediaFilterFieldUpdateTime:      "更新时间",
		MediaFilterFieldSortTitle:       "排序标题",
		MediaFilterFieldLanguage:        "语言",
		MediaFilterFieldCountry:         "国家",
		MediaFilterFieldNetwork:         "网络平台",
		MediaFilterFieldCollection:      "合集",
		MediaFilterFieldQuality:         "质量",
		MediaFilterFieldCodec:           "编码",
		MediaFilterFieldResolution:      "分辨率",
		MediaFilterFieldAudio:           "音频",
		MediaFilterFieldVideoFormat:     "视频格式",
		MediaFilterFieldFolderSize:      "文件夹大小",
		MediaFilterFieldFilePath:        "文件路径",
		MediaFilterFieldFolderPath:      "文件夹路径",
		MediaFilterFieldMediaServer:     "媒体服务器",
		MediaFilterFieldSubtitleStatus:  "字幕状态",
		MediaFilterFieldCustom1:         "自定义字段1",
		MediaFilterFieldCustom2:         "自定义字段2",
		MediaFilterFieldCustom3:         "自定义字段3",
	}

	if name, ok := fieldNames[field]; ok {
		return name
	}
	return string(field)
}

// 辅助方法：获取字段类型
func (h *mediaFilterHandler) getFieldType(field MediaFilterField) FilterFieldType {
	numericFields := map[MediaFilterField]bool{
		MediaFilterFieldID:            true,
		MediaFilterFieldYear:          true,
		MediaFilterFieldRating:        true,
		MediaFilterFieldVotes:         true,
		MediaFilterFieldRuntime:       true,
		MediaFilterFieldSeasonCount:   true,
		MediaFilterFieldEpisodeCount:  true,
		MediaFilterFieldFolderSize:    true,
	}

	dateFields := map[MediaFilterField]bool{
		MediaFilterFieldAirDate:      true,
		MediaFilterFieldFirstAirDate: true,
		MediaFilterFieldLastAirDate:  true,
		MediaFilterFieldReleaseDate:  true,
		MediaFilterFieldCreateTime:   true,
		MediaFilterFieldUpdateTime:   true,
	}

	listFields := map[MediaFilterField]bool{
		MediaFilterFieldGenres: true,
		MediaFilterFieldTags:   true,
		MediaFilterFieldCast:   true,
	}

	if numericFields[field] {
		return FilterFieldTypeNumeric
	} else if dateFields[field] {
		return FilterFieldTypeDate
	} else if listFields[field] {
		return FilterFieldTypeList
	}

	return FilterFieldTypeString
}

// 辅助方法：获取字段支持的操作符
func (h *mediaFilterHandler) getFieldOperators(field MediaFilterField) []FilterOperator {
	// 基础操作符
	baseOperators := []FilterOperator{
		FilterOperatorEq,
		FilterOperatorNe,
		FilterOperatorIsNull,
		FilterOperatorIsNotNull,
	}

	// 根据字段类型添加额外操作符
	fieldType := h.getFieldType(field)
	switch fieldType {
	case FilterFieldTypeNumeric:
		return append(baseOperators, 
			FilterOperatorGt, FilterOperatorGte, 
			FilterOperatorLt, FilterOperatorLte, 
			FilterOperatorBetween,
		)
	case FilterFieldTypeString:
		return append(baseOperators, 
			FilterOperatorLike, FilterOperatorNotLike,
			FilterOperatorIn, FilterOperatorNotIn,
			FilterOperatorRegex, FilterOperatorNotRegex,
			FilterOperatorStartsWith, FilterOperatorEndsWith,
		)
	case FilterFieldTypeDate:
		return append(baseOperators, 
			FilterOperatorGt, FilterOperatorGte, 
			FilterOperatorLt, FilterOperatorLte, 
			FilterOperatorBetween,
		)
	case FilterFieldTypeList:
		return append(baseOperators, 
			FilterOperatorIn, FilterOperatorNotIn,
		)
	default:
		return baseOperators
	}
}

// 错误响应结构
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// 验证过滤响应结构
type ValidateFilterResponse struct {
	Valid   bool     `json:"valid"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

// 媒体过滤预览响应
type MediaFilterPreviewResponse struct {
	Success       bool              `json:"success"`
	PreviewMedias []*MediaItem      `json:"preview_medias"`
	TotalFiltered int               `json:"total_filtered"`
	TotalMedias   int               `json:"total_medias"`
	PreviewCount  int               `json:"preview_count"`
	ApplyFilters  MediaFilterParams `json:"apply_filters"`
}

// 可用过滤字段响应
type AvailableFilterFieldsResponse struct {
	Success bool              `json:"success"`
	Fields  []FilterFieldInfo `json:"fields"`
	Total   int               `json:"total"`
}

// 过滤字段信息
type FilterFieldInfo struct {
	Field     MediaFilterField `json:"field"`
	Name      string           `json:"name"`
	Type      FilterFieldType  `json:"type"`
	Operators []FilterOperator `json:"operators"`
}

// 过滤字段类型
type FilterFieldType string

const (
	FilterFieldTypeString  FilterFieldType = "string"
	FilterFieldTypeNumeric FilterFieldType = "numeric"
	FilterFieldTypeDate    FilterFieldType = "date"
	FilterFieldTypeList    FilterFieldType = "list"
	FilterFieldTypeBoolean FilterFieldType = "boolean"
)
