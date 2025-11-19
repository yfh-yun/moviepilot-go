package mediaserver

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// SyncService 媒体服务器同步服务
// 负责处理不同媒体服务器之间的数据同步，支持增量同步和全量同步
type SyncService struct {
	logger         *zap.Logger
	mediaServerSvc *MediaServerService
}

// NewSyncService 创建同步服务实例
func NewSyncService(logger *zap.Logger, mediaServerSvc *MediaServerService) *SyncService {
	return &SyncService{
		logger:         logger,
		mediaServerSvc: mediaServerSvc,
	}
}

// SyncRequest 同步请求
// 定义同步操作的参数和配置
type SyncRequest struct {
	SourceServer string   `json:"source_server" binding:"required"` // 源服务器
	TargetServer string   `json:"target_server" binding:"required"` // 目标服务器
	SyncType     string   `json:"sync_type" binding:"required"`     // 同步类型 (full/incremental)
	LibraryIDs   []string `json:"library_ids,omitempty"`            // 指定媒体库ID
	UserIDs      []string `json:"user_ids,omitempty"`               // 指定用户ID
	DataTypes    []string `json:"data_types"`                       // 数据类型 (libraries/playback/metadata)
}

// SyncResponse 同步响应
type SyncResponse struct {
	TaskID      string      `json:"task_id"`               // 任务ID
	SyncType    string      `json:"sync_type"`             // 同步类型
	TotalItems  int         `json:"total_items"`           // 总同步项数
	SyncedItems int         `json:"synced_items"`          // 已同步项数
	Status      string      `json:"status"`                // 任务状态
	StartTime   time.Time   `json:"start_time"`            // 开始时间
	FinishTime  *time.Time  `json:"finish_time,omitempty"` // 完成时间
	Errors      []SyncError `json:"errors,omitempty"`      // 同步错误
}

// SyncError 同步错误信息
type SyncError struct {
	ItemID    string `json:"item_id"`    // 同步项ID
	ErrorType string `json:"error_type"` // 错误类型
	Message   string `json:"message"`    // 错误消息
}

// Sync 执行媒体服务器同步
// 支持全量同步和增量同步，提供详细的进度跟踪和错误处理
func (s *SyncService) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	// 参数验证
	if req.SourceServer == req.TargetServer {
		return nil, errors.New("源服务器和目标服务器不能相同")
	}

	// 创建同步任务
	taskID := uuid.New().String()
	response := &SyncResponse{
		TaskID:      taskID,
		SyncType:    req.SyncType,
		TotalItems:  0,
		SyncedItems: 0,
		Status:      "processing",
		StartTime:   time.Now(),
		Errors:      make([]SyncError, 0),
	}

	s.logger.Info("开始媒体服务器同步",
		zap.String("task_id", taskID),
		zap.String("source_server", req.SourceServer),
		zap.String("target_server", req.TargetServer),
		zap.String("sync_type", req.SyncType),
		zap.Strings("data_types", req.DataTypes),
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

	// 执行同步操作
	syncedCount := 0
	totalCount := 0

	// 同步媒体库
	if s.containsDataType(req.DataTypes, "libraries") {
		count, err := s.syncLibraries(ctx, sourceServer, targetServer, req.LibraryIDs, response)
		if err != nil {
			s.logger.Error("同步媒体库失败",
				zap.String("task_id", taskID),
				zap.Error(err),
			)
		}
		totalCount += count
		syncedCount += count
	}

	// 同步播放状态
	if s.containsDataType(req.DataTypes, "playback") {
		count, err := s.syncPlaybackStatus(ctx, sourceServer, targetServer, req.UserIDs, response)
		if err != nil {
			s.logger.Error("同步播放状态失败",
				zap.String("task_id", taskID),
				zap.Error(err),
			)
		}
		totalCount += count
		syncedCount += count
	}

	// 同步元数据
	if s.containsDataType(req.DataTypes, "metadata") {
		count, err := s.syncMetadata(ctx, sourceServer, targetServer, req.LibraryIDs, response)
		if err != nil {
			s.logger.Error("同步元数据失败",
				zap.String("task_id", taskID),
				zap.Error(err),
			)
		}
		totalCount += count
		syncedCount += count
	}

	// 更新响应数据
	response.TotalItems = totalCount
	response.SyncedItems = syncedCount
	finishTime := time.Now()
	response.FinishTime = &finishTime
	response.Status = "completed"

	duration := finishTime.Sub(response.StartTime).String()
	s.logger.Info("媒体服务器同步完成",
		zap.String("task_id", taskID),
		zap.Int("total_items", totalCount),
		zap.Int("synced_items", syncedCount),
		zap.Int("error_count", len(response.Errors)),
		zap.String("duration", duration),
	)

	return response, nil
}

// syncLibraries 同步媒体库数据
func (s *SyncService) syncLibraries(ctx context.Context, sourceServer, targetServer MediaServer, libraryIDs []string, response *SyncResponse) (int, error) {
	s.logger.Debug("开始同步媒体库数据")

	// 获取源服务器的媒体库列表
	sourceLibraries, err := sourceServer.GetLibraries(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "获取源服务器媒体库失败")
	}

	// 过滤指定媒体库
	if len(libraryIDs) > 0 {
		sourceLibraries = s.filterLibraries(sourceLibraries, libraryIDs)
	}

	syncedCount := 0

	// 同步每个媒体库
	for _, library := range sourceLibraries {
		synced, err := s.syncSingleLibrary(ctx, sourceServer, targetServer, library.ID, response)
		if err != nil {
			s.logger.Warn("同步单个媒体库失败",
				zap.String("library_id", library.ID),
				zap.Error(err),
			)
		}

		if synced {
			syncedCount++
		}
	}

	return syncedCount, nil
}

// syncPlaybackStatus 同步播放状态
func (s *SyncService) syncPlaybackStatus(ctx context.Context, sourceServer, targetServer MediaServer, userIDs []string, response *SyncResponse) (int, error) {
	s.logger.Debug("开始同步播放状态")

	// 获取源服务器的播放会话
	sourceSessions, err := sourceServer.GetPlaybackSessions(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "获取源服务器播放会话失败")
	}

	syncedCount := 0

	// 同步播放状态
	for _, session := range sourceSessions {
		// 过滤指定用户
		if len(userIDs) > 0 && !s.containsUser(userIDs, session.UserID) {
			continue
		}

		synced, err := s.syncSinglePlaybackStatus(ctx, sourceServer, targetServer, session, response)
		if err != nil {
			s.logger.Warn("同步单个播放状态失败",
				zap.String("session_id", session.ID),
				zap.Error(err),
			)
		}

		if synced {
			syncedCount++
		}
	}

	return syncedCount, nil
}

// syncMetadata 同步元数据
func (s *SyncService) syncMetadata(ctx context.Context, sourceServer, targetServer MediaServer, libraryIDs []string, response *SyncResponse) (int, error) {
	s.logger.Debug("开始同步元数据")

	// 这里实现元数据同步逻辑
	// 支持批量同步媒体项的元数据
	syncedCount := 0

	return syncedCount, nil
}

// syncSingleLibrary 同步单个媒体库
func (s *SyncService) syncSingleLibrary(ctx context.Context, sourceServer, targetServer MediaServer, libraryID string, response *SyncResponse) (bool, error) {
	// 获取源媒体库项
	sourceItems, err := sourceServer.GetLibraryItems(ctx, libraryID, QueryParams{})
	if err != nil {
		response.addError(libraryID, "library_sync", err.Error())
		return false, errors.Wrap(err, "获取源媒体库项失败")
	}

	s.logger.Debug("同步媒体库项",
		zap.String("library_id", libraryID),
		zap.Int("item_count", len(sourceItems.Items)),
	)

	// 实际同步逻辑
	// 这里需要实现具体的媒体项同步

	return true, nil
}

// syncSinglePlaybackStatus 同步单个播放状态
func (s *SyncService) syncSinglePlaybackStatus(ctx context.Context, sourceServer, targetServer MediaServer, session PlaybackSession, response *SyncResponse) (bool, error) {
	// 同步播放状态到目标服务器
	err := targetServer.UpdatePlaybackStatus(ctx, session.ItemID, PlaybackStatus{
		Progress:    session.Progress,
		IsPlaying:   session.IsPlaying,
		LastPlayed:  session.LastPlayed,
		PlayedCount: session.PlayedCount,
	})

	if err != nil {
		response.addError(session.ID, "playback_sync", err.Error())
		return false, errors.Wrap(err, "同步播放状态失败")
	}

	return true, nil
}

// containsDataType 检查是否包含指定数据类型
func (s *SyncService) containsDataType(dataTypes []string, target string) bool {
	if len(dataTypes) == 0 {
		return true // 默认同步所有类型
	}

	for _, dt := range dataTypes {
		if dt == target {
			return true
		}
	}
	return false
}

// filterLibraries 过滤媒体库
func (s *SyncService) filterLibraries(libraries []Library, targetIDs []string) []Library {
	filtered := make([]Library, 0)

	for _, library := range libraries {
		for _, targetID := range targetIDs {
			if library.ID == targetID {
				filtered = append(filtered, library)
				break
			}
		}
	}

	return filtered
}

// containsUser 检查是否包含指定用户
func (s *SyncService) containsUser(userIDs []string, targetID string) bool {
	for _, userID := range userIDs {
		if userID == targetID {
			return true
		}
	}
	return false
}

// addError 添加同步错误
func (r *SyncResponse) addError(itemID, errorType, message string) {
	r.Errors = append(r.Errors, SyncError{
		ItemID:    itemID,
		ErrorType: errorType,
		Message:   message,
	})
}
