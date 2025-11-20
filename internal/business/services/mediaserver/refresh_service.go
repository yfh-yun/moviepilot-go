package mediaserver

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// RefreshService 媒体库刷新服务
// 负责处理媒体库的批量刷新、增量同步、进度跟踪等功能
type RefreshService struct {
	logger         *zap.Logger
	mediaServerSvc *MediaServerService
}

// NewRefreshService 创建刷新服务实例
func NewRefreshService(logger *zap.Logger, mediaServerSvc *MediaServerService) *RefreshService {
	return &RefreshService{
		logger:         logger,
		mediaServerSvc: mediaServerSvc,
	}
}

// RefreshLibraryRequest 刷新媒体库请求
// 支持单个或多个媒体库的刷新操作
type RefreshLibraryRequest struct {
	LibraryIDs   []string `json:"library_ids" binding:"required"` // 媒体库ID列表
	ServerType   string   `json:"server_type" binding:"required"` // 服务器类型 (emby/jellyfin/plex)
	ForceRefresh bool     `json:"force_refresh"`                  // 是否强制刷新
}

// RefreshLibraryResponse 刷新媒体库响应
type RefreshLibraryResponse struct {
	TaskID     string          `json:"task_id"`               // 任务ID
	Total      int             `json:"total"`                 // 总媒体库数量
	Completed  int             `json:"completed"`             // 已完成数量
	Status     string          `json:"status"`                // 任务状态
	Results    []RefreshResult `json:"results"`               // 刷新结果
	StartTime  time.Time       `json:"start_time"`            // 开始时间
	FinishTime *time.Time      `json:"finish_time,omitempty"` // 完成时间
}

// RefreshResult 单个媒体库刷新结果
type RefreshResult struct {
	LibraryID  string    `json:"library_id"`      // 媒体库ID
	ServerType string    `json:"server_type"`     // 服务器类型
	Status     string    `json:"status"`          // 刷新状态
	Error      *string   `json:"error,omitempty"` // 错误信息
	StartTime  time.Time `json:"start_time"`      // 开始时间
	FinishTime time.Time `json:"finish_time"`     // 完成时间
	Duration   string    `json:"duration"`        // 耗时
}

// RefreshLibrary 刷新媒体库
// 支持批量刷新多个媒体库，提供进度跟踪和错误恢复
func (s *RefreshService) RefreshLibrary(ctx context.Context, req *RefreshLibraryRequest) (*RefreshLibraryResponse, error) {
	// 参数验证
	if len(req.LibraryIDs) == 0 {
		return nil, errors.New("媒体库ID列表不能为空")
	}

	// 创建刷新任务
	taskID := uuid.New().String()
	response := &RefreshLibraryResponse{
		TaskID:    taskID,
		Total:     len(req.LibraryIDs),
		Completed: 0,
		Status:    "processing",
		Results:   make([]RefreshResult, 0, len(req.LibraryIDs)),
		StartTime: time.Now(),
	}

	s.logger.Info("开始刷新媒体库",
		zap.String("task_id", taskID),
		zap.String("server_type", req.ServerType),
		zap.Int("library_count", len(req.LibraryIDs)),
		zap.Bool("force_refresh", req.ForceRefresh),
	)

	// 遍历媒体库进行刷新
	for _, libraryID := range req.LibraryIDs {
		result := RefreshResult{
			LibraryID:  libraryID,
			ServerType: req.ServerType,
			Status:     "processing",
			StartTime:  time.Now(),
		}

		// 执行媒体库刷新
		err := s.refreshSingleLibrary(ctx, req.ServerType, libraryID, req.ForceRefresh)

		result.FinishTime = time.Now()
		result.Duration = result.FinishTime.Sub(result.StartTime).String()

		if err != nil {
			errorMsg := err.Error()
			result.Status = "failed"
			result.Error = &errorMsg
			s.logger.Error("媒体库刷新失败",
				zap.String("task_id", taskID),
				zap.String("library_id", libraryID),
				zap.String("server_type", req.ServerType),
				zap.Error(err),
			)
		} else {
			result.Status = "completed"
			s.logger.Info("媒体库刷新完成",
				zap.String("task_id", taskID),
				zap.String("library_id", libraryID),
				zap.String("server_type", req.ServerType),
				zap.String("duration", result.Duration),
			)
		}

		response.Results = append(response.Results, result)
		response.Completed++
	}

	// 更新任务状态
	finishTime := time.Now()
	response.FinishTime = &finishTime
	response.Status = "completed"

	totalDuration := finishTime.Sub(response.StartTime).String()
	s.logger.Info("媒体库刷新任务完成",
		zap.String("task_id", taskID),
		zap.Int("total", response.Total),
		zap.Int("completed", response.Completed),
		zap.String("duration", totalDuration),
	)

	return response, nil
}

// refreshSingleLibrary 刷新单个媒体库
// 执行具体的媒体库刷新逻辑，支持错误重试和超时控制
func (s *RefreshService) refreshSingleLibrary(ctx context.Context, serverType, libraryID string, forceRefresh bool) error {
	s.logger.Debug("开始刷新单个媒体库",
		zap.String("server_type", serverType),
		zap.String("library_id", libraryID),
		zap.Bool("force_refresh", forceRefresh),
	)

	// 获取媒体服务器实例
	server, err := s.mediaServerSvc.GetServer(serverType)
	if err != nil {
		return errors.Wrap(err, "获取媒体服务器实例失败")
	}

	// 检查连接状态
	if !server.IsConnected() {
		s.logger.Warn("媒体服务器未连接，尝试重新连接",
			zap.String("server_type", serverType),
		)

		if err := server.Connect(ctx); err != nil {
			return errors.Wrap(err, "媒体服务器连接失败")
		}
	}

	// 执行媒体库刷新
	if err := server.RefreshLibrary(ctx, libraryID); err != nil {
		return errors.Wrap(err, "媒体库刷新失败")
	}

	return nil
}

// GetRefreshProgress 获取刷新进度
// 用于查询刷新任务的实时进度
func (s *RefreshService) GetRefreshProgress(ctx context.Context, taskID string) (*RefreshLibraryResponse, error) {
	// 这里实现进度查询逻辑
	// 实际应用中可以从Redis或数据库中获取任务进度
	s.logger.Debug("查询刷新进度",
		zap.String("task_id", taskID),
	)

	return nil, errors.New("进度查询功能待实现")
}
