package media

// MediaScrapeRequest 媒体识别请求
type MediaScrapeRequest struct {
	FilePath string `json:"file_path" binding:"required"` // 媒体文件路径
	MediaID  string `json:"media_id,omitempty"`           // 媒体ID（可选）
	Season   int    `json:"season,omitempty"`             // 季号（剧集）
	Episode  int    `json:"episode,omitempty"`            // 集号（剧集）
}

// MediaScrapeResponse 媒体识别响应
type MediaScrapeResponse struct {
	MediaID  string `json:"media_id"`          // 媒体ID
	FilePath string `json:"file_path"`         // 文件路径
	Status   string `json:"status"`            // 状态: pending, processing, completed, failed
	Message  string `json:"message"`           // 状态消息
	TMDBID   int    `json:"tmdb_id,omitempty"` // TMDB ID
	Title    string `json:"title,omitempty"`   // 标题
	Year     int    `json:"year,omitempty"`    // 年份
	Type     string `json:"type,omitempty"`    // 类型: movie, tv
}

// MediaSearchRequest 媒体搜索请求
type MediaSearchRequest struct {
	Keyword string `json:"keyword" binding:"required"` // 搜索关键词
	Type    string `json:"type,omitempty"`             // 媒体类型: movie, tv, all
	Year    int    `json:"year,omitempty"`             // 年份
	Limit   int    `json:"limit,omitempty"`            // 限制数量
}

// MediaSearchResult 媒体搜索结果
type MediaSearchResult struct {
	TMDBID        int     `json:"tmdb_id"`        // TMDB ID
	Title         string  `json:"title"`          // 标题
	OriginalTitle string  `json:"original_title"` // 原标题
	Year          int     `json:"year"`           // 年份
	Type          string  `json:"type"`           // 类型: movie, tv
	Overview      string  `json:"overview"`       // 简介
	PosterPath    string  `json:"poster_path"`    // 海报路径
	BackdropPath  string  `json:"backdrop_path"`  // 背景图路径
	Rating        float64 `json:"rating"`         // 评分
	ReleaseDate   string  `json:"release_date"`   // 发布日期
}

// MediaSearchResponse 媒体搜索响应
type MediaSearchResponse struct {
	Keyword string              `json:"keyword"` // 搜索关键词
	Type    string              `json:"type"`    // 媒体类型
	Page    int                 `json:"page"`    // 当前页
	Size    int                 `json:"size"`    // 每页数量
	Total   int                 `json:"total"`   // 总数量
	Results []MediaSearchResult `json:"results"` // 搜索结果
	Message string              `json:"message"` // 状态消息
}

// MediaInfoRequest 媒体信息请求
type MediaInfoRequest struct {
	MediaID string `json:"media_id" binding:"required"` // 媒体ID
	Source  string `json:"source,omitempty"`            // 数据源: tmdb, douban, bangumi
}

// MediaInfoResponse 媒体信息响应
type MediaInfoResponse struct {
	MediaID       string         `json:"media_id"`           // 媒体ID
	TMDBID        int            `json:"tmdb_id"`            // TMDB ID
	Title         string         `json:"title"`              // 标题
	OriginalTitle string         `json:"original_title"`     // 原标题
	Year          int            `json:"year"`               // 年份
	Type          string         `json:"type"`               // 类型: movie, tv
	Overview      string         `json:"overview"`           // 简介
	Status        string         `json:"status"`             // 状态
	Message       string         `json:"message"`            // 状态消息
	Metadata      map[string]any `json:"metadata,omitempty"` // 额外元数据
	CreatedAt     string         `json:"created_at"`         // 创建时间
	UpdatedAt     string         `json:"updated_at"`         // 更新时间
}

// MediaBatchScrapeRequest 批量媒体识别请求
type MediaBatchScrapeRequest struct {
	FilePaths []string `json:"file_paths" binding:"required"` // 文件路径列表
	AutoMatch bool     `json:"auto_match,omitempty"`          // 自动匹配
	Force     bool     `json:"force,omitempty"`               // 强制重新识别
}

// MediaBatchScrapeResponse 批量媒体识别响应
type MediaBatchScrapeResponse struct {
	TaskID    string                `json:"task_id"`           // 任务ID
	Status    string                `json:"status"`            // 状态: created, processing, completed, failed
	Total     int                   `json:"total"`             // 总数量
	Processed int                   `json:"processed"`         // 已处理数量
	Failed    int                   `json:"failed"`            // 失败数量
	Message   string                `json:"message"`           // 状态消息
	Results   []MediaScrapeResponse `json:"results,omitempty"` // 详细结果（可选）
}

// MediaRefreshRequest 媒体元数据刷新请求
type MediaRefreshRequest struct {
	MediaID string `json:"media_id" binding:"required"` // 媒体ID
	Force   bool   `json:"force,omitempty"`             // 强制刷新
}

// MediaRefreshResponse 媒体元数据刷新响应
type MediaRefreshResponse struct {
	MediaID string `json:"media_id"` // 媒体ID
	Status  string `json:"status"`   // 状态: refreshing, completed, failed
	Message string `json:"message"`  // 状态消息
}

// MediaListItem 媒体列表项
type MediaListItem struct {
	MediaID   string `json:"media_id"`   // 媒体ID
	Title     string `json:"title"`      // 标题
	Year      int    `json:"year"`       // 年份
	Type      string `json:"type"`       // 类型: movie, tv
	Status    string `json:"status"`     // 状态: identified, pending, processing
	FilePath  string `json:"file_path"`  // 文件路径
	FileSize  int64  `json:"file_size"`  // 文件大小
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// MediaListRequest 媒体列表请求
type MediaListRequest struct {
	Type     string `json:"type,omitempty"`      // 类型过滤: movie, tv, all
	Status   string `json:"status,omitempty"`    // 状态过滤
	Page     int    `json:"page,omitempty"`      // 页码
	Size     int    `json:"size,omitempty"`      // 每页数量
	SortBy   string `json:"sort_by,omitempty"`   // 排序字段
	SortDesc bool   `json:"sort_desc,omitempty"` // 降序排列
}

// MediaListResponse 媒体列表响应
type MediaListResponse struct {
	Page    int             `json:"page"`    // 当前页
	Size    int             `json:"size"`    // 每页数量
	Total   int             `json:"total"`   // 总数量
	Items   []MediaListItem `json:"items"`   // 媒体列表
	Message string          `json:"message"` // 状态消息
}
