package actions

import (
	"time"
)

// 媒体过滤相关枚举类型

// MediaFilterField 媒体过滤字段
type MediaFilterField string

const (
	MediaFilterFieldID              MediaFilterField = "id"
	MediaFilterFieldTitle           MediaFilterField = "title"
	MediaFilterFieldOriginalTitle   MediaFilterField = "original_title"
	MediaFilterFieldType            MediaFilterField = "type"
	MediaFilterFieldStatus          MediaFilterField = "status"
	MediaFilterFieldYear            MediaFilterField = "year"
	MediaFilterFieldRating          MediaFilterField = "rating"
	MediaFilterFieldVotes           MediaFilterField = "votes"
	MediaFilterFieldRuntime         MediaFilterField = "runtime"
	MediaFilterFieldSeasonCount     MediaFilterField = "season_count"
	MediaFilterFieldEpisodeCount    MediaFilterField = "episode_count"
	MediaFilterFieldAirDate         MediaFilterField = "air_date"
	MediaFilterFieldFirstAirDate    MediaFilterField = "first_air_date"
	MediaFilterFieldLastAirDate     MediaFilterField = "last_air_date"
	MediaFilterFieldReleaseDate     MediaFilterField = "release_date"
	MediaFilterFieldOverview        MediaFilterField = "overview"
	MediaFilterFieldGenres          MediaFilterField = "genres"
	MediaFilterFieldTags            MediaFilterField = "tags"
	MediaFilterFieldStudio          MediaFilterField = "studio"
	MediaFilterFieldDirector        MediaFilterField = "director"
	MediaFilterFieldCast            MediaFilterField = "cast"
	MediaFilterFieldWriter          MediaFilterField = "writer"
	MediaFilterFieldIMDBID          MediaFilterField = "imdb_id"
	MediaFilterFieldTMDBID          MediaFilterField = "tmdb_id"
	MediaFilterFieldTVDBID          MediaFilterField = "tvdb_id"
	MediaFilterFieldSource          MediaFilterField = "source"
	MediaFilterFieldCover           MediaFilterField = "cover"
	MediaFilterFieldBackdrop        MediaFilterField = "backdrop"
	MediaFilterFieldTrailer         MediaFilterField = "trailer"
	MediaFilterFieldLogo            MediaFilterField = "logo"
	MediaFilterFieldLocalStatus     MediaFilterField = "local_status"
	MediaFilterFieldSubscribeStatus MediaFilterField = "subscribe_status"
	MediaFilterFieldDownloadStatus  MediaFilterField = "download_status"
	MediaFilterFieldCreateTime      MediaFilterField = "create_time"
	MediaFilterFieldUpdateTime      MediaFilterField = "update_time"
	MediaFilterFieldSortTitle       MediaFilterField = "sort_title"
	MediaFilterFieldLanguage        MediaFilterField = "language"
	MediaFilterFieldCountry         MediaFilterField = "country"
	MediaFilterFieldNetwork         MediaFilterField = "network"
	MediaFilterFieldCollection      MediaFilterField = "collection"
	MediaFilterFieldQuality         MediaFilterField = "quality"
	MediaFilterFieldCodec           MediaFilterField = "codec"
	MediaFilterFieldResolution      MediaFilterField = "resolution"
	MediaFilterFieldAudio           MediaFilterField = "audio"
	MediaFilterFieldVideoFormat     MediaFilterField = "video_format"
	MediaFilterFieldFolderSize      MediaFilterField = "folder_size"
	MediaFilterFieldFilePath        MediaFilterField = "file_path"
	MediaFilterFieldFolderPath      MediaFilterField = "folder_path"
	MediaFilterFieldMediaServer     MediaFilterField = "media_server"
	MediaFilterFieldSubtitleStatus  MediaFilterField = "subtitle_status"
	MediaFilterFieldCustom1         MediaFilterField = "custom1"
	MediaFilterFieldCustom2         MediaFilterField = "custom2"
	MediaFilterFieldCustom3         MediaFilterField = "custom3"
)

// FilterOperator 过滤操作符
type FilterOperator string

const (
	FilterOperatorEq        FilterOperator = "eq"        // 等于
	FilterOperatorNe        FilterOperator = "ne"        // 不等于
	FilterOperatorGt        FilterOperator = "gt"        // 大于
	FilterOperatorGte       FilterOperator = "gte"       // 大于等于
	FilterOperatorLt        FilterOperator = "lt"        // 小于
	FilterOperatorLte       FilterOperator = "lte"       // 小于等于
	FilterOperatorLike      FilterOperator = "like"      // 包含
	FilterOperatorNotLike   FilterOperator = "not_like"  // 不包含
	FilterOperatorIn        FilterOperator = "in"        // 在范围内
	FilterOperatorNotIn     FilterOperator = "not_in"    // 不在范围内
	FilterOperatorRegex     FilterOperator = "regex"     // 正则匹配
	FilterOperatorNotRegex  FilterOperator = "not_regex" // 正则不匹配
	FilterOperatorBetween   FilterOperator = "between"   // 在两者之间
	FilterOperatorIsNull    FilterOperator = "is_null"   // 为空
	FilterOperatorIsNotNull FilterOperator = "is_not_null" // 不为空
	FilterOperatorStartsWith FilterOperator = "starts_with" // 以...开始
	FilterOperatorEndsWith   FilterOperator = "ends_with" // 以...结束
)

// MediaSortField 媒体排序字段
type MediaSortField string

const (
	MediaSortFieldID              MediaSortField = "id"
	MediaSortFieldTitle           MediaSortField = "title"
	MediaSortFieldOriginalTitle   MediaSortField = "original_title"
	MediaSortFieldType            MediaSortField = "type"
	MediaSortFieldYear            MediaSortField = "year"
	MediaSortFieldRating          MediaSortField = "rating"
	MediaSortFieldVotes           MediaSortField = "votes"
	MediaSortFieldRuntime         MediaSortField = "runtime"
	MediaSortFieldSeasonCount     MediaSortField = "season_count"
	MediaSortFieldEpisodeCount    MediaSortField = "episode_count"
	MediaSortFieldAirDate         MediaSortField = "air_date"
	MediaSortFieldFirstAirDate    MediaSortField = "first_air_date"
	MediaSortFieldLastAirDate     MediaSortField = "last_air_date"
	MediaSortFieldReleaseDate     MediaSortField = "release_date"
	MediaSortFieldCreateTime      MediaSortField = "create_time"
	MediaSortFieldUpdateTime      MediaSortField = "update_time"
	MediaSortFieldSortTitle       MediaSortField = "sort_title"
	MediaSortFieldLocalStatus     MediaSortField = "local_status"
	MediaSortFieldSubscribeStatus MediaSortField = "subscribe_status"
	MediaSortFieldDownloadStatus  MediaSortField = "download_status"
	MediaSortFieldFolderSize      MediaSortField = "folder_size"
	MediaSortFieldQuality         MediaSortField = "quality"
	MediaSortFieldResolution      MediaSortField = "resolution"
)

// SortOrder 排序顺序
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"  // 升序
	SortOrderDesc SortOrder = "desc" // 降序
)

// 过滤条件相关结构体

// FilterCondition 单个过滤条件
type FilterCondition struct {
	Field    MediaFilterField `json:"field" binding:"required"`     // 字段名
	Operator FilterOperator   `json:"operator" binding:"required"` // 操作符
	Value    interface{}      `json:"value" binding:"required"`    // 值
}

// FilterGroup 过滤条件组
type FilterGroup struct {
	Logic      string          `json:"logic" binding:"required,oneof=and or"` // 逻辑关系：and或or
	Conditions []interface{}   `json:"conditions" binding:"required"`        // 条件列表，可以是FilterCondition或FilterGroup
}

// MediaFilterParams 媒体过滤参数
type MediaFilterParams struct {
	// 基础过滤条件
	Filters *FilterGroup `json:"filters,omitempty"` // 高级过滤条件组

	// 快捷过滤条件
	MediaTypes  []MediaType      `json:"media_types,omitempty"`  // 媒体类型过滤
	Status      []MediaStatus    `json:"status,omitempty"`       // 状态过滤
	Years       []int            `json:"years,omitempty"`        // 年份过滤
	Genres      []string         `json:"genres,omitempty"`       // 类型过滤
	Tags        []string         `json:"tags,omitempty"`         // 标签过滤
	RatingMin   *float64         `json:"rating_min,omitempty"`   // 最低评分
	RatingMax   *float64         `json:"rating_max,omitempty"`   // 最高评分
	VotesMin    *int             `json:"votes_min,omitempty"`    // 最低投票数
	RuntimeMin  *int             `json:"runtime_min,omitempty"`  // 最低时长
	RuntimeMax  *int             `json:"runtime_max,omitempty"`  // 最高时长
	Language    []string         `json:"language,omitempty"`     // 语言过滤
	Country     []string         `json:"country,omitempty"`      // 国家过滤
	Networks    []string         `json:"networks,omitempty"`     // 网络/平台过滤
	Collections []string         `json:"collections,omitempty"`  // 合集过滤
	Studios     []string         `json:"studios,omitempty"`      // 工作室过滤

	// 本地状态过滤
	LocalStatus         []LocalMediaStatus `json:"local_status,omitempty"`         // 本地状态
	SubscribeStatus     []SubscribeStatus  `json:"subscribe_status,omitempty"`     // 订阅状态
	DownloadStatus      []DownloadStatus   `json:"download_status,omitempty"`      // 下载状态
	SubtitleStatus      []SubtitleStatus   `json:"subtitle_status,omitempty"`      // 字幕状态
	HasLocal            *bool              `json:"has_local,omitempty"`            // 是否有本地文件
	HasSubscription     *bool              `json:"has_subscription,omitempty"`     // 是否有订阅
	HasBackdrop         *bool              `json:"has_backdrop,omitempty"`         // 是否有背景图
	HasCover            *bool              `json:"has_cover,omitempty"`            // 是否有封面
	HasTrailer          *bool              `json:"has_trailer,omitempty"`          // 是否有预告片
	HasWatchHistory     *bool              `json:"has_watch_history,omitempty"`     // 是否有观看历史

	// 时间范围过滤
	AirDateStart      *time.Time `json:"air_date_start,omitempty"`      // 上映日期开始
	AirDateEnd        *time.Time `json:"air_date_end,omitempty"`        // 上映日期结束
	CreateTimeStart   *time.Time `json:"create_time_start,omitempty"`   // 创建时间开始
	CreateTimeEnd     *time.Time `json:"create_time_end,omitempty"`     // 创建时间结束
	UpdateTimeStart   *time.Time `json:"update_time_start,omitempty"`   // 更新时间开始
	UpdateTimeEnd     *time.Time `json:"update_time_end,omitempty"`     // 更新时间结束
	LastWatchTimeStart *time.Time `json:"last_watch_time_start,omitempty"` // 最后观看时间开始
	LastWatchTimeEnd   *time.Time `json:"last_watch_time_end,omitempty"`   // 最后观看时间结束

	// 高级过滤
	TextSearch       string  `json:"text_search,omitempty"`       // 文本搜索关键词
	ExcludeIDs       []int64 `json:"exclude_ids,omitempty"`       // 排除ID列表
	IncludeIDs       []int64 `json:"include_ids,omitempty"`       // 包含ID列表
	Quality          []string `json:"quality,omitempty"`          // 质量过滤
	Resolution       []string `json:"resolution,omitempty"`       // 分辨率过滤
	Codec            []string `json:"codec,omitempty"`            // 编码过滤
	Audio            []string `json:"audio,omitempty"`            // 音频格式过滤
	VideoFormat      []string `json:"video_format,omitempty"`     // 视频格式过滤
	FolderSizeMin    *int64   `json:"folder_size_min,omitempty"`  // 最小文件夹大小
	FolderSizeMax    *int64   `json:"folder_size_max,omitempty"`  // 最大文件夹大小
	MediaServerID    string   `json:"media_server_id,omitempty"`  // 媒体服务器ID
	DownloaderID     string   `json:"downloader_id,omitempty"`    // 下载器ID

	// 分页和排序
	SortBy    MediaSortField `json:"sort_by" binding:"omitempty"`    // 排序字段
	SortOrder SortOrder      `json:"sort_order" binding:"omitempty,oneof=asc desc"` // 排序顺序
	Limit     int            `json:"limit" binding:"omitempty,min=1,max=1000"` // 每页数量
	Offset    int            `json:"offset" binding:"omitempty,min=0"` // 偏移量
	Page      int            `json:"page" binding:"omitempty,min=1"`  // 页码（优先使用offset）
}

// MediaFilterResponse 媒体过滤响应
type MediaFilterResponse struct {
	Success         bool        `json:"success"`          // 是否成功
	Medias          []*MediaItem `json:"medias"`           // 媒体列表
	Total           int         `json:"total"`            // 总数（过滤前）
	Filtered        int         `json:"filtered"`         // 过滤后总数
	Page            int         `json:"page"`             // 当前页码
	PageSize        int         `json:"page_size"`        // 每页数量
	TotalPages      int         `json:"total_pages"`      // 总页数
	ProcessingTime  time.Duration `json:"processing_time"` // 处理时间
	AppliedFilters  *FilterGroup `json:"applied_filters,omitempty"` // 应用的过滤条件
	SortBy          MediaSortField `json:"sort_by,omitempty"` // 排序字段
	SortOrder       SortOrder    `json:"sort_order,omitempty"` // 排序顺序
}

// MediaFilterStats 媒体过滤统计
type MediaFilterStats struct {
	Total           int                    `json:"total"`             // 总数
	ByMediaType     map[MediaType]int      `json:"by_media_type"`     // 按媒体类型统计
	ByStatus        map[MediaStatus]int    `json:"by_status"`         // 按状态统计
	ByLocalStatus   map[LocalMediaStatus]int `json:"by_local_status"`   // 按本地状态统计
	ByYear          map[int]int           `json:"by_year"`           // 按年份统计
	ByGenre         map[string]int        `json:"by_genre"`          // 按类型统计
	ByTag           map[string]int        `json:"by_tag"`            // 按标签统计
	ByQuality       map[string]int        `json:"by_quality"`        // 按质量统计
	ByResolution    map[string]int        `json:"by_resolution"`     // 按分辨率统计
	RatingStats     *RatingStats          `json:"rating_stats"`      // 评分统计
	RuntimeStats    *RuntimeStats         `json:"runtime_stats"`     // 时长统计
	FileSizeStats   *FileSizeStats        `json:"file_size_stats"`   // 文件大小统计
	LastUpdated     time.Time             `json:"last_updated"`      // 最后更新时间
}

// RatingStats 评分统计
type RatingStats struct {
	Min     float64 `json:"min"`     // 最低评分
	Max     float64 `json:"max"`     // 最高评分
	Average float64 `json:"average"` // 平均评分
	Count   int     `json:"count"`   // 统计数量
}

// RuntimeStats 时长统计
type RuntimeStats struct {
	Min     int     `json:"min"`     // 最小时长
	Max     int     `json:"max"`     // 最大时长
	Average float64 `json:"average"` // 平均时长
	Count   int     `json:"count"`   // 统计数量
}

// FileSizeStats 文件大小统计
type FileSizeStats struct {
	Min     int64   `json:"min"`     // 最小大小
	Max     int64   `json:"max"`     // 最大大小
	Average int64   `json:"average"` // 平均大小
	Total   int64   `json:"total"`   // 总大小
	Count   int     `json:"count"`   // 统计数量
}

// MediaFilterSuggestion 媒体过滤建议
type MediaFilterSuggestion struct {
	Field     MediaFilterField `json:"field"`     // 字段名
	Values    []string         `json:"values"`    // 建议值列表
	Count     int              `json:"count"`     // 建议值数量
	Threshold int              `json:"threshold"` // 阈值
}

// MediaFilterSuggestionsResponse 媒体过滤建议响应
type MediaFilterSuggestionsResponse struct {
	Success    bool                    `json:"success"`     // 是否成功
	Suggestions []*MediaFilterSuggestion `json:"suggestions"` // 建议列表
	Fields      []MediaFilterField      `json:"fields"`      // 可过滤字段
}

// 高级过滤相关结构体

// AdvancedMediaFilter 高级媒体过滤
type AdvancedMediaFilter struct {
	Version    string       `json:"version"`     // 过滤器版本
	Name       string       `json:"name"`        // 过滤器名称
	Desc       string       `json:"desc,omitempty"` // 过滤器描述
	Filter     *FilterGroup `json:"filter"`      // 过滤条件
	Sort       *SortConfig  `json:"sort,omitempty"` // 排序配置
	IsPublic   bool         `json:"is_public"`   // 是否公开
	IsDefault  bool         `json:"is_default"`  // 是否默认
	CreateTime time.Time    `json:"create_time"` // 创建时间
	UpdateTime time.Time    `json:"update_time"` // 更新时间
	Creator    string       `json:"creator,omitempty"` // 创建者
}

// SortConfig 排序配置
type SortConfig struct {
	Field MediaSortField `json:"field" binding:"required"` // 排序字段
	Order SortOrder      `json:"order" binding:"required,oneof=asc desc"` // 排序顺序
}

// SaveFilterRequest 保存过滤器请求
type SaveFilterRequest struct {
	Name     string       `json:"name" binding:"required"`     // 过滤器名称
	Desc     string       `json:"desc,omitempty"`              // 过滤器描述
	Filter   *FilterGroup `json:"filter" binding:"required"`  // 过滤条件
	Sort     *SortConfig  `json:"sort,omitempty"`              // 排序配置
	IsPublic bool         `json:"is_public"`                   // 是否公开
}

// UpdateFilterRequest 更新过滤器请求
type UpdateFilterRequest struct {
	Name     string       `json:"name,omitempty"`             // 过滤器名称
	Desc     string       `json:"desc,omitempty"`             // 过滤器描述
	Filter   *FilterGroup `json:"filter,omitempty"`           // 过滤条件
	Sort     *SortConfig  `json:"sort,omitempty"`             // 排序配置
	IsPublic *bool        `json:"is_public,omitempty"`        // 是否公开
}

// FilterListResponse 过滤器列表响应
type FilterListResponse struct {
	Success  bool                  `json:"success"`  // 是否成功
	Filters  []*AdvancedMediaFilter `json:"filters"`  // 过滤器列表
	Total    int                   `json:"total"`    // 总数
	Page     int                   `json:"page"`     // 当前页码
	PageSize int                   `json:"page_size"`// 每页数量
}

// ValidateFilterRequest 验证过滤器请求
type ValidateFilterRequest struct {
	Filter *FilterGroup `json:"filter" binding:"required"` // 过滤条件
}

// ValidateFilterResponse 验证过滤器响应
type ValidateFilterResponse struct {
	Valid   bool     `json:"valid"`    // 是否有效
	Message string   `json:"message,omitempty"` // 消息
	Errors  []string `json:"errors,omitempty"`  // 错误列表
}
