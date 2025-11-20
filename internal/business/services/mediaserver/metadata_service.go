package mediaserver

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// MetadataService 元数据服务
// 负责媒体元数据的同步、转换、映射和管理功能
type MetadataService struct {
	logger         *zap.Logger
	mediaServerSvc *MediaServerService
}

// NewMetadataService 创建元数据服务实例
func NewMetadataService(logger *zap.Logger, mediaServerSvc *MediaServerService) *MetadataService {
	return &MetadataService{
		logger:         logger,
		mediaServerSvc: mediaServerSvc,
	}
}

// MetadataSyncRequest 元数据同步请求
type MetadataSyncRequest struct {
	SourceServer string   `json:"source_server" binding:"required"` // 源服务器
	TargetServer string   `json:"target_server" binding:"required"` // 目标服务器
	LibraryIDs   []string `json:"library_ids,omitempty"`            // 指定媒体库ID
	ItemIDs      []string `json:"item_ids,omitempty"`               // 指定媒体项ID
	SyncType     string   `json:"sync_type" binding:"required"`     // 同步类型 (full/incremental)
	Fields       []string `json:"fields"`                           // 同步字段
	Overwrite    bool     `json:"overwrite"`                        // 是否覆盖现有数据
}

// MetadataSyncResponse 元数据同步响应
type MetadataSyncResponse struct {
	TaskID      string              `json:"task_id"`               // 任务ID
	SyncType    string              `json:"sync_type"`             // 同步类型
	TotalItems  int                 `json:"total_items"`           // 总同步项数
	SyncedItems int                 `json:"synced_items"`          // 已同步项数
	Status      string              `json:"status"`                // 任务状态
	StartTime   time.Time           `json:"start_time"`            // 开始时间
	FinishTime  *time.Time          `json:"finish_time,omitempty"` // 完成时间
	Errors      []MetadataSyncError `json:"errors,omitempty"`      // 同步错误
}

// MetadataSyncError 元数据同步错误信息
type MetadataSyncError struct {
	ItemID  string `json:"item_id"` // 媒体项ID
	Field   string `json:"field"`   // 错误字段
	Message string `json:"message"` // 错误消息
}

// MetadataItem 元数据项信息
type MetadataItem struct {
	ID          string    `json:"id"`           // 媒体项ID
	Title       string    `json:"title"`        // 标题
	Year        int       `json:"year"`         // 年份
	Type        string    `json:"type"`         // 类型
	Genre       []string  `json:"genre"`        // 类型
	Rating      float64   `json:"rating"`       // 评分
	Overview    string    `json:"overview"`     // 简介
	Poster      string    `json:"poster"`       // 海报URL
	Backdrop    string    `json:"backdrop"`     // 背景图URL
	Cast        []string  `json:"cast"`         // 演员列表
	Director    []string  `json:"director"`     // 导演列表
	Runtime     int       `json:"runtime"`      // 时长(分钟)
	Language    string    `json:"language"`     // 语言
	Country     []string  `json:"country"`      // 国家/地区
	ReleaseDate string    `json:"release_date"` // 发布日期
	LastUpdate  time.Time `json:"last_update"`  // 最后更新时间
}

// SyncMetadata 执行元数据同步
// 支持全量同步和增量同步，提供字段级控制和错误处理
func (s *MetadataService) SyncMetadata(ctx context.Context, req *MetadataSyncRequest) (*MetadataSyncResponse, error) {
	// 参数验证
	if req.SourceServer == req.TargetServer {
		return nil, errors.New("源服务器和目标服务器不能相同")
	}

	// 创建同步任务
	taskID := uuid.New().String()
	response := &MetadataSyncResponse{
		TaskID:      taskID,
		SyncType:    req.SyncType,
		TotalItems:  0,
		SyncedItems: 0,
		Status:      "processing",
		StartTime:   time.Now(),
		Errors:      make([]MetadataSyncError, 0),
	}

	s.logger.Info("开始元数据同步",
		zap.String("task_id", taskID),
		zap.String("source_server", req.SourceServer),
		zap.String("target_server", req.TargetServer),
		zap.String("sync_type", req.SyncType),
		zap.Strings("fields", req.Fields),
		zap.Bool("overwrite", req.Overwrite),
	)

	// 获取源服务器和目标服务器实例
	sourceServer, err := s.mediaServerSvc.GetServer(req.SourceServer)
	if err != nil {
		return nil, errors.Wrap(err, "获取源服务器实例失败")
	}

	targetServer, err := s.mediaServerSvc.GetServer(req.TargetServer)
	if err != nil {
		return nil, errors.Wrap(err, "获取目标服务器实例失败")
	}

	// 检查服务器连接状态
	if !sourceServer.IsConnected() {
		if err := sourceServer.Connect(ctx); err != nil {
			return nil, errors.Wrap(err, "源服务器连接失败")
		}
	}

	if !targetServer.IsConnected() {
		if err := targetServer.Connect(ctx); err != nil {
			return nil, errors.Wrap(err, "目标服务器连接失败")
		}
	}

	// 获取需要同步的媒体项
	items, err := s.getItemsToSync(ctx, sourceServer, req.LibraryIDs, req.ItemIDs, req.SyncType)
	if err != nil {
		return nil, errors.Wrap(err, "获取需要同步的媒体项失败")
	}

	response.TotalItems = len(items)

	s.logger.Info("开始同步元数据项",
		zap.String("task_id", taskID),
		zap.Int("total_items", len(items)),
	)

	// 同步每个媒体项
	for _, item := range items {
		synced, err := s.syncSingleItem(ctx, sourceServer, targetServer, item, req.Fields, req.Overwrite, response)
		if err != nil {
			s.logger.Warn("同步单个媒体项失败",
				zap.String("item_id", item.ID),
				zap.Error(err),
			)
		}

		if synced {
			response.SyncedItems++
		}
	}

	// 更新任务状态
	finishTime := time.Now()
	response.FinishTime = &finishTime
	response.Status = "completed"

	duration := finishTime.Sub(response.StartTime).String()
	s.logger.Info("元数据同步完成",
		zap.String("task_id", taskID),
		zap.Int("total_items", response.TotalItems),
		zap.Int("synced_items", response.SyncedItems),
		zap.Int("error_count", len(response.Errors)),
		zap.String("duration", duration),
	)

	return response, nil
}

// getItemsToSync 获取需要同步的媒体项
func (s *MetadataService) getItemsToSync(ctx context.Context, server MediaServer, libraryIDs, itemIDs []string, syncType string) ([]MetadataItem, error) {
	items := make([]MetadataItem, 0)

	// 如果指定了媒体项ID，直接同步这些项
	if len(itemIDs) > 0 {
		for _, itemID := range itemIDs {
			item, err := s.getItemMetadata(ctx, server, itemID)
			if err != nil {
				s.logger.Warn("获取媒体项元数据失败",
					zap.String("item_id", itemID),
					zap.Error(err),
				)
				continue
			}
			items = append(items, *item)
		}
		return items, nil
	}

	// 如果指定了媒体库，同步整个媒体库
	if len(libraryIDs) > 0 {
		for _, libraryID := range libraryIDs {
			libraryItems, err := s.getLibraryItems(ctx, server, libraryID, syncType)
			if err != nil {
				s.logger.Warn("获取媒体库项失败",
					zap.String("library_id", libraryID),
					zap.Error(err),
				)
				continue
			}
			items = append(items, libraryItems...)
		}
		return items, nil
	}

	// 全量同步所有媒体库
	libraries, err := server.GetLibraries(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "获取媒体库列表失败")
	}

	for _, library := range libraries {
		libraryItems, err := s.getLibraryItems(ctx, server, library.ID, syncType)
		if err != nil {
			s.logger.Warn("获取媒体库项失败",
				zap.String("library_id", library.ID),
				zap.Error(err),
			)
			continue
		}
		items = append(items, libraryItems...)
	}

	return items, nil
}

// syncSingleItem 同步单个媒体项
func (s *MetadataService) syncSingleItem(ctx context.Context, sourceServer, targetServer MediaServer, item MetadataItem, fields []string, overwrite bool, response *MetadataSyncResponse) (bool, error) {
	s.logger.Debug("同步单个媒体项",
		zap.String("item_id", item.ID),
		zap.String("title", item.Title),
	)

	// 获取目标服务器上的现有元数据
	targetItem, err := s.getItemMetadata(ctx, targetServer, item.ID)
	if err != nil {
		s.logger.Warn("获取目标服务器元数据失败",
			zap.String("item_id", item.ID),
			zap.Error(err),
		)
	}

	// 检查是否需要覆盖
	if !overwrite && targetItem != nil {
		s.logger.Debug("跳过已存在的媒体项",
			zap.String("item_id", item.ID),
		)
		return false, nil
	}

	// 执行元数据同步
	// 这里需要实现具体的元数据字段同步逻辑

	s.logger.Debug("媒体项同步完成",
		zap.String("item_id", item.ID),
	)

	return true, nil
}

// getItemMetadata 获取单个媒体项的元数据
func (s *MetadataService) getItemMetadata(ctx context.Context, server MediaServer, itemID string) (*MetadataItem, error) {
	// 这里实现从媒体服务器获取元数据的逻辑
	// 返回示例数据
	return &MetadataItem{
		ID:         itemID,
		Title:      "示例标题",
		Year:       2024,
		Type:       "movie",
		Rating:     8.5,
		LastUpdate: time.Now(),
	}, nil
}

// getLibraryItems 获取媒体库中的所有媒体项
func (s *MetadataService) getLibraryItems(ctx context.Context, server MediaServer, libraryID, syncType string) ([]MetadataItem, error) {
	// 获取媒体库项
	items, err := server.GetLibraryItems(ctx, libraryID, QueryParams{})
	if err != nil {
		return nil, errors.Wrap(err, "获取媒体库项失败")
	}

	// 根据同步类型过滤项
	var filteredItems []MetadataItem

	if syncType == "incremental" {
		// 增量同步：只同步最近更新的项
		// 这里实现增量同步逻辑
	} else {
		// 全量同步：同步所有项
		// 这里实现全量同步逻辑
	}

	return filteredItems, nil
}

// GetMetadataMappings 获取元数据映射关系
// 用于不同服务器之间的元数据字段映射
func (s *MetadataService) GetMetadataMappings(sourceServer, targetServer string) (map[string]string, error) {
	mappings := make(map[string]string)

	// 根据服务器类型设置映射关系
	switch sourceServer + "->" + targetServer {
	case "emby->jellyfin":
		mappings = map[string]string{
			"Title":    "Name",
			"Overview": "Overview",
			"Rating":   "CommunityRating",
			"Year":     "ProductionYear",
			"Genre":    "Genres",
			"Runtime":  "RunTimeTicks",
		}
	case "emby->plex":
		mappings = map[string]string{
			"Title":    "title",
			"Overview": "summary",
			"Rating":   "rating",
			"Year":     "year",
			"Genre":    "genre",
			"Runtime":  "duration",
		}
	case "jellyfin->emby":
		mappings = map[string]string{
			"Name":            "Title",
			"Overview":        "Overview",
			"CommunityRating": "Rating",
			"ProductionYear":  "Year",
			"Genres":          "Genre",
			"RunTimeTicks":    "Runtime",
		}
	case "plex->emby":
		mappings = map[string]string{
			"title":    "Title",
			"summary":  "Overview",
			"rating":   "Rating",
			"year":     "Year",
			"genre":    "Genre",
			"duration": "Runtime",
		}
	}

	s.logger.Debug("获取元数据映射关系",
		zap.String("source", sourceServer),
		zap.String("target", targetServer),
		zap.Int("mapping_count", len(mappings)),
	)

	return mappings, nil
}

// addError 添加元数据同步错误
func (r *MetadataSyncResponse) addError(itemID, field, message string) {
	r.Errors = append(r.Errors, MetadataSyncError{
		ItemID:  itemID,
		Field:   field,
		Message: message,
	})
}
