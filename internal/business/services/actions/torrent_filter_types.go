package actions

import (
	"time"
)

// TorrentFilterField 种子过滤字段枚举
type TorrentFilterField string

const (
	// 基本信息
	TorrentFilterFieldID          TorrentFilterField = "id"
	TorrentFilterFieldName        TorrentFilterField = "name"
	TorrentFilterFieldHash        TorrentFilterField = "hash"
	TorrentFilterFieldSize        TorrentFilterField = "size"
	TorrentFilterFieldProgress    TorrentFilterField = "progress"
	TorrentFilterFieldStatus      TorrentFilterField = "status"
	TorrentFilterFieldType        TorrentFilterField = "type"
	TorrentFilterFieldCategory    TorrentFilterField = "category"
	TorrentFilterFieldTags        TorrentFilterField = "tags"
	TorrentFilterFieldTracker     TorrentFilterField = "tracker"
	TorrentFilterFieldTrackerStatus TorrentFilterField = "tracker_status"
	TorrentFilterFieldDownloader  TorrentFilterField = "downloader"

	// 速度信息
	TorrentFilterFieldDownloadSpeed TorrentFilterField = "download_speed"
	TorrentFilterFieldUploadSpeed   TorrentFilterField = "upload_speed"
	TorrentFilterFieldRatio         TorrentFilterField = "ratio"
	TorrentFilterFieldSeedingTime   TorrentFilterField = "seeding_time"
	TorrentFilterFieldActiveTime    TorrentFilterField = "active_time"

	// 时间信息
	TorrentFilterFieldCreateTime    TorrentFilterField = "create_time"
	TorrentFilterFieldAddTime       TorrentFilterField = "add_time"
	TorrentFilterFieldCompletedTime TorrentFilterField = "completed_time"
	TorrentFilterFieldLastActiveTime TorrentFilterField = "last_active_time"

	// 文件信息
	TorrentFilterFieldFileCount     TorrentFilterField = "file_count"
	TorrentFilterFieldPath          TorrentFilterField = "path"
	TorrentFilterFieldSavePath      TorrentFilterField = "save_path"
	TorrentFilterFieldMediaInfo     TorrentFilterField = "media_info"

	// 媒体关联
	TorrentFilterFieldMediaType     TorrentFilterField = "media_type"
	TorrentFilterFieldMediaID       TorrentFilterField = "media_id"
	TorrentFilterFieldSeason        TorrentFilterField = "season"
	TorrentFilterFieldEpisode       TorrentFilterField = "episode"
	TorrentFilterFieldQuality       TorrentFilterField = "quality"

	// 其他信息
	TorrentFilterFieldPriority      TorrentFilterField = "priority"
	TorrentFilterFieldAutoTMM       TorrentFilterField = "auto_tmm"
	TorrentFilterFieldComment       TorrentFilterField = "comment"
	TorrentFilterFieldCreator       TorrentFilterField = "creator"
	TorrentFilterFieldNumComplete   TorrentFilterField = "num_complete"
	TorrentFilterFieldNumLeechs     TorrentFilterField = "num_leechs"
	TorrentFilterFieldNumSeeds      TorrentFilterField = "num_seeds"

	// 自定义字段
	TorrentFilterFieldCustom1       TorrentFilterField = "custom1"
	TorrentFilterFieldCustom2       TorrentFilterField = "custom2"
	TorrentFilterFieldCustom3       TorrentFilterField = "custom3"
)

// TorrentSortField 种子排序字段枚举
type TorrentSortField string

const (
	TorrentSortFieldID          TorrentSortField = "id"
	TorrentSortFieldName        TorrentSortField = "name"
	TorrentSortFieldHash        TorrentSortField = "hash"
	TorrentSortFieldSize        TorrentSortField = "size"
	TorrentSortFieldProgress    TorrentSortField = "progress"
	TorrentSortFieldStatus      TorrentSortField = "status"
	TorrentSortFieldType        TorrentSortField = "type"
	TorrentSortFieldCategory    TorrentSortField = "category"
	TorrentSortFieldDownloadSpeed TorrentSortField = "download_speed"
	TorrentSortFieldUploadSpeed   TorrentSortField = "upload_speed"
	TorrentSortFieldRatio         TorrentSortField = "ratio"
	TorrentSortFieldSeedingTime   TorrentSortField = "seeding_time"
	TorrentSortFieldActiveTime    TorrentSortField = "active_time"
	TorrentSortFieldCreateTime    TorrentSortField = "create_time"
	TorrentSortFieldAddTime       TorrentSortField = "add_time"
	TorrentSortFieldCompletedTime TorrentSortField = "completed_time"
	TorrentSortFieldLastActiveTime TorrentSortField = "last_active_time"
	TorrentSortFieldFileCount     TorrentSortField = "file_count"
	TorrentSortFieldMediaType     TorrentSortField = "media_type"
	TorrentSortFieldMediaID       TorrentSortField = "media_id"
	TorrentSortFieldQuality       TorrentSortField = "quality"
	TorrentSortFieldPriority      TorrentSortField = "priority"
	TorrentSortFieldNumComplete   TorrentSortField = "num_complete"
	TorrentSortFieldNumLeechs     TorrentSortField = "num_leechs"
	TorrentSortFieldNumSeeds      TorrentSortField = "num_seeds"
)

// TorrentFilterParams 种子过滤参数结构体
type TorrentFilterParams struct {
	// 基础过滤条件
	IDs              []string          `json:"ids,omitempty"`
	Names            []string          `json:"names,omitempty"`
	Hashes           []string          `json:"hashes,omitempty"`
	Statuses         []TorrentStatus   `json:"statuses,omitempty"`
	Types            []TorrentType     `json:"types,omitempty"`
	Categories       []string          `json:"categories,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Trackers         []string          `json:"trackers,omitempty"`
	Downloaders      []string          `json:"downloaders,omitempty"`

	// 数值范围过滤
	SizeMin          *int64            `json:"size_min,omitempty"`
	SizeMax          *int64            `json:"size_max,omitempty"`
	ProgressMin      *float64          `json:"progress_min,omitempty"`
	ProgressMax      *float64          `json:"progress_max,omitempty"`
	DownloadSpeedMin *int64            `json:"download_speed_min,omitempty"`
	DownloadSpeedMax *int64            `json:"download_speed_max,omitempty"`
	UploadSpeedMin   *int64            `json:"upload_speed_min,omitempty"`
	UploadSpeedMax   *int64            `json:"upload_speed_max,omitempty"`
	RatioMin         *float64          `json:"ratio_min,omitempty"`
	RatioMax         *float64          `json:"ratio_max,omitempty"`
	SeedingTimeMin   *int64            `json:"seeding_time_min,omitempty"`
	SeedingTimeMax   *int64            `json:"seeding_time_max,omitempty"`

	// 时间范围过滤
	CreateTimeFrom   *time.Time        `json:"create_time_from,omitempty"`
	CreateTimeTo     *time.Time        `json:"create_time_to,omitempty"`
	AddTimeFrom      *time.Time        `json:"add_time_from,omitempty"`
	AddTimeTo        *time.Time        `json:"add_time_to,omitempty"`
	CompletedTimeFrom *time.Time       `json:"completed_time_from,omitempty"`
	CompletedTimeTo   *time.Time       `json:"completed_time_to,omitempty"`
	LastActiveTimeFrom *time.Time      `json:"last_active_time_from,omitempty"`
	LastActiveTimeTo   *time.Time      `json:"last_active_time_to,omitempty"`

	// 媒体关联过滤
	MediaTypes       []MediaType       `json:"media_types,omitempty"`
	MediaIDs         []string          `json:"media_ids,omitempty"`
	Seasons          []int             `json:"seasons,omitempty"`
	Episodes         []int             `json:"episodes,omitempty"`
	Qualities        []string          `json:"qualities,omitempty"`

	// 高级过滤条件
	Filters          *TorrentFilterGroup `json:"filters,omitempty"`

	// 排序参数
	SortBy           TorrentSortField  `json:"sort_by,omitempty"`
	SortOrder        SortOrder         `json:"sort_order,omitempty"`

	// 分页参数
	Page             int               `json:"page,omitempty" default:"1"`
	Limit            int               `json:"limit,omitempty" default:"20"`
	Offset           int               `json:"offset,omitempty" default:"0"`

	// 其他选项
	IncludeFiles     bool              `json:"include_files,omitempty"`
	IncludeTrackers  bool              `json:"include_trackers,omitempty"`
	IncludePeers     bool              `json:"include_peers,omitempty"`
	IncludeMediaInfo bool              `json:"include_media_info,omitempty"`
	ExactMatch       bool              `json:"exact_match,omitempty"`
	OnlyActive       bool              `json:"only_active,omitempty"`
	OnlyCompleted    bool              `json:"only_completed,omitempty"`
	OnlyDownloading  bool              `json:"only_downloading,omitempty"`
	OnlySeeding      bool              `json:"only_seeding,omitempty"`
	OnlyPaused       bool              `json:"only_paused,omitempty"`
	OnlyStalled      bool              `json:"only_stalled,omitempty"`
}

// TorrentFilterGroup 种子过滤条件组结构体
type TorrentFilterGroup struct {
	Logic      string                   `json:"logic" binding:"required,oneof=and or"`
	Conditions []interface{}            `json:"conditions" binding:"required,min=1"`
}

// TorrentFilterCondition 种子过滤条件结构体
type TorrentFilterCondition struct {
	Field    TorrentFilterField   `json:"field" binding:"required"`
	Operator FilterOperator       `json:"operator" binding:"required"`
	Value    interface{}          `json:"value"`
}

// TorrentFilterResponse 种子过滤响应结构体
type TorrentFilterResponse struct {
	Items      []TorrentItem      `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
	HasMore    bool               `json:"has_more"`
	Elapsed    float64            `json:"elapsed"`
}

// TorrentFilterStats 种子过滤统计结构体
type TorrentFilterStats struct {
	TotalCount       int64          `json:"total_count"`
	StatusStats      map[string]int64 `json:"status_stats"`
	TypeStats        map[string]int64 `json:"type_stats"`
	CategoryStats    map[string]int64 `json:"category_stats"`
	DownloaderStats  map[string]int64 `json:"downloader_stats"`
	MediaTypeStats   map[string]int64 `json:"media_type_stats"`
	QualityStats     map[string]int64 `json:"quality_stats"`

	// 数值统计
	TotalSize        int64          `json:"total_size"`
	AverageRatio     float64        `json:"average_ratio"`
	AverageProgress  float64        `json:"average_progress"`

	// 时间统计
	OldestAddTime    *time.Time     `json:"oldest_add_time"`
	NewestAddTime    *time.Time     `json:"newest_add_time"`
	OldestCompleted  *time.Time     `json:"oldest_completed_time"`
	NewestCompleted  *time.Time     `json:"newest_completed_time"`

	// 活动统计
	ActiveCount      int64          `json:"active_count"`
	InactiveCount    int64          `json:"inactive_count"`
	CompletedCount   int64          `json:"completed_count"`
	DownloadingCount int64          `json:"downloading_count"`
	SeedingCount     int64          `json:"seeding_count"`
	PausedCount      int64          `json:"paused_count"`
}

// TorrentFilterSuggestion 种子过滤建议结构体
type TorrentFilterSuggestion struct {
	Field    TorrentFilterField `json:"field"`
	Type     string             `json:"type"`
	Values   []string           `json:"values,omitempty"`
	MinValue *float64           `json:"min_value,omitempty"`
	MaxValue *float64           `json:"max_value,omitempty"`
}

// TorrentFilterPreview 种子过滤预览结构体
type TorrentFilterPreview struct {
	PreviewCount int64            `json:"preview_count"`
	SampleItems  []TorrentItem    `json:"sample_items,omitempty"`
	Stats        *TorrentFilterStats `json:"stats,omitempty"`
	Elapsed      float64          `json:"elapsed"`
}

// TorrentBulkActionParams 种子批量操作参数结构体
type TorrentBulkActionParams struct {
	IDs           []string         `json:"ids,omitempty"`
	Hashes        []string         `json:"hashes,omitempty"`
	Filter        *TorrentFilterParams `json:"filter,omitempty"`
	Action        string           `json:"action" binding:"required,oneof=pause resume delete recheck move category tags priority ratio"`

	// 操作相关参数
	Category      string           `json:"category,omitempty"`
	Tags          []string         `json:"tags,omitempty"`
	AddTags       []string         `json:"add_tags,omitempty"`
	RemoveTags    []string         `json:"remove_tags,omitempty"`
	Priority      int              `json:"priority,omitempty"`
	RatioLimit    *float64         `json:"ratio_limit,omitempty"`
	SeedingTimeLimit *int64        `json:"seeding_time_limit,omitempty"`
	SavePath      string           `json:"save_path,omitempty"`
	Downloader    string           `json:"downloader,omitempty"`

	// 高级选项
	Force         bool             `json:"force,omitempty"`
	DeleteFiles   bool             `json:"delete_files,omitempty"`
	Recursive     bool             `json:"recursive,omitempty"`
	ApplyToFiles  bool             `json:"apply_to_files,omitempty"`
}

// TorrentBulkActionResponse 种子批量操作响应结构体
type TorrentBulkActionResponse struct {
	Success       bool             `json:"success"`
	AffectedCount int64            `json:"affected_count"`
	FailedCount   int64            `json:"failed_count"`
	Errors        []string         `json:"errors,omitempty"`
	Elapsed       float64          `json:"elapsed"`
}

// TorrentExportParams 种子导出参数结构体
type TorrentExportParams struct {
	Filter        *TorrentFilterParams `json:"filter,omitempty"`
	Format        string           `json:"format" binding:"required,oneof=json csv tsv excel"`
	IncludeFields []string         `json:"include_fields,omitempty"`
	ExcludeFields []string         `json:"exclude_fields,omitempty"`
	FileName      string           `json:"file_name,omitempty"`
	Compress      bool             `json:"compress,omitempty"`
}

// TorrentFilterValidatorParams 种子过滤验证器参数结构体
type TorrentFilterValidatorParams struct {
	Filter        *TorrentFilterParams `json:"filter,omitempty"`
	ValidateStats bool                `json:"validate_stats,omitempty"`
	ValidatePreview bool              `json:"validate_preview,omitempty"`
}

// TorrentFilterValidatorResponse 种子过滤验证器响应结构体
type TorrentFilterValidatorResponse struct {
	Valid         bool               `json:"valid"`
	Errors        map[string]string  `json:"errors,omitempty"`
	Suggestions   []TorrentFilterSuggestion `json:"suggestions,omitempty"`
}

// TorrentTrackerFilter 种子Tracker过滤结构体
type TorrentTrackerFilter struct {
	URL            string `json:"url,omitempty"`
	Status         string `json:"status,omitempty"`
	Message        string `json:"message,omitempty"`
	NumPeers       int    `json:"num_peers,omitempty"`
	NumSeeds       int    `json:"num_seeds,omitempty"`
	NumLeechs      int    `json:"num_leeches,omitempty"`
	NumDownloaded  int    `json:"num_downloaded,omitempty"`
}

// TorrentFileFilter 种子文件过滤结构体
type TorrentFileFilter struct {
	Name          string   `json:"name,omitempty"`
	Path          string   `json:"path,omitempty"`
	SizeMin       *int64   `json:"size_min,omitempty"`
	SizeMax       *int64   `json:"size_max,omitempty"`
	Extension     []string `json:"extension,omitempty"`
	IsSelected    *bool    `json:"is_selected,omitempty"`
	IsVideo       *bool    `json:"is_video,omitempty"`
	IsAudio       *bool    `json:"is_audio,omitempty"`
	IsSubtitle    *bool    `json:"is_subtitle,omitempty"`
	IsFolder      *bool    `json:"is_folder,omitempty"`
	MediaInfo     *bool    `json:"has_media_info,omitempty"`
}

// TorrentAdvancedFilter 种子高级过滤结构体
type TorrentAdvancedFilter struct {
	CombineMode        string                  `json:"combine_mode" binding:"required,oneof=and or"`
	MainFilters        *TorrentFilterGroup     `json:"main_filters,omitempty"`
	FileFilters        *TorrentFileFilter      `json:"file_filters,omitempty"`
	TrackerFilters     *TorrentTrackerFilter   `json:"tracker_filters,omitempty"`
	IgnoreTorrentFilter bool                   `json:"ignore_torrent_filter,omitempty"`
	IgnoreFileFilter   bool                    `json:"ignore_file_filter,omitempty"`
	IgnoreTrackerFilter bool                   `json:"ignore_tracker_filter,omitempty"`
}
