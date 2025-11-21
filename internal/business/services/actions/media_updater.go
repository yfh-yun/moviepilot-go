package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/response"
)

// MediaUpdater 媒体更新器接口
type MediaUpdater interface {
	// UpdateMedia 更新单个媒体
	UpdateMedia(ctx context.Context, req *MediaUpdateRequest) (*MediaUpdateResult, error)
	// BatchUpdateMedias 批量更新媒体
	BatchUpdateMedias(ctx context.Context, req *MediaUpdateRequest) (*MediaUpdateResponse, error)
	// GetUpdateTask 获取更新任务
	GetUpdateTask(ctx context.Context, taskID string) (*MediaUpdateTask, error)
	// ListUpdateTasks 列出更新任务
	ListUpdateTasks(ctx context.Context, query *MediaUpdateTaskQuery) (*MediaUpdateTaskListResponse, error)
	// CancelUpdateTask 取消更新任务
	CancelUpdateTask(ctx context.Context, taskID string) error
	// GetUpdateStats 获取更新统计信息
	GetUpdateStats(ctx context.Context) (*MediaUpdateStats, error)
}

// MediaRepository 媒体仓库接口
type MediaRepository interface {
	// GetMediaByID 根据ID获取媒体
	GetMediaByID(ctx context.Context, mediaID string) (interface{}, error)
	// UpdateMedia 更新媒体信息
	UpdateMedia(ctx context.Context, mediaID string, updates map[string]interface{}) error
	// ListMediasByFilter 根据条件列出媒体
	ListMediasByFilter(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]interface{}, int64, error)
	// GetMediaUpdateTime 获取媒体最后更新时间
	GetMediaUpdateTime(ctx context.Context, mediaID string) (*time.Time, error)
}

// MediaMetadataProvider 媒体元数据提供者接口
type MediaMetadataProvider interface {
	// GetMediaMetadata 获取媒体元数据
	GetMediaMetadata(ctx context.Context, mediaID string, mediaType string) (map[string]interface{}, error)
	// IsMetadataUpToDate 检查元数据是否最新
	IsMetadataUpToDate(ctx context.Context, mediaID string, lastUpdateTime time.Time) (bool, error)
}

// mediaUpdater 媒体更新器实现
type mediaUpdater struct {
	mediaRepo          MediaRepository
	metadataProvider   MediaMetadataProvider
	taskManager        *UpdateTaskManager
	logger             logger.Logger
	config             *MediaUpdateConfig
}

// NewMediaUpdater 创建媒体更新器实例
func NewMediaUpdater(
	mediaRepo MediaRepository,
	metadataProvider MediaMetadataProvider,
	logger logger.Logger,
	config *MediaUpdateConfig,
) MediaUpdater {
	return &mediaUpdater{
		mediaRepo:        mediaRepo,
		metadataProvider: metadataProvider,
		taskManager:      NewUpdateTaskManager(logger),
		logger:           logger,
		config:           config,
	}
}

// UpdateMedia 更新单个媒体
func (u *mediaUpdater) UpdateMedia(ctx context.Context, req *MediaUpdateRequest) (*MediaUpdateResult, error) {
	if req.MediaID == "" {
		return nil, errors.New("媒体ID不能为空")
	}

	u.logger.Debug("开始更新单个媒体", "media_id", req.MediaID)
	startTime := time.Now()

	// 获取媒体信息
	media, err := u.mediaRepo.GetMediaByID(ctx, req.MediaID)
	if err != nil {
		u.logger.Error("获取媒体信息失败", "media_id", req.MediaID, "error", err.Error())
		return nil, fmt.Errorf("获取媒体信息失败: %w", err)
	}

	// 检查是否需要更新
	if !u.shouldUpdateMedia(ctx, req, req.MediaID) {
		executionTime := time.Since(startTime).Milliseconds()
		u.logger.Info("媒体不需要更新", "media_id", req.MediaID)
		return &MediaUpdateResult{
			MediaID:       req.MediaID,
			Success:       true,
			UpdatedFields: []MediaUpdateField{},
			ChangedFields: []MediaUpdateField{},
			ExecutionTime: executionTime,
			UpdatedAt:     time.Now(),
		}, nil
	}

	// 获取元数据
	metadata, err := u.metadataProvider.GetMediaMetadata(ctx, req.MediaID, req.MediaType)
	if err != nil {
		u.logger.Error("获取媒体元数据失败", "media_id", req.MediaID, "error", err.Error())
		return nil, fmt.Errorf("获取媒体元数据失败: %w", err)
	}

	// 准备更新字段
	updates, updatedFields, changedFields := u.prepareMediaUpdates(media, metadata, req)

	// 执行更新
	if len(updates) > 0 {
		err = u.mediaRepo.UpdateMedia(ctx, req.MediaID, updates)
		if err != nil {
			u.logger.Error("更新媒体信息失败", "media_id", req.MediaID, "error", err.Error())
			return nil, fmt.Errorf("更新媒体信息失败: %w", err)
		}
	}

	executionTime := time.Since(startTime).Milliseconds()
	result := &MediaUpdateResult{
		MediaID:       req.MediaID,
		Success:       true,
		UpdatedFields: updatedFields,
		ChangedFields: changedFields,
		ExecutionTime: executionTime,
		UpdatedAt:     time.Now(),
	}

	u.logger.Info("媒体更新成功", "media_id", req.MediaID, "changed_fields", changedFields)
	return result, nil
}

// BatchUpdateMedias 批量更新媒体
func (u *mediaUpdater) BatchUpdateMedias(ctx context.Context, req *MediaUpdateRequest) (*MediaUpdateResponse, error) {
	// 验证请求
	if len(req.MediaIDs) == 0 && req.MediaID == "" && req.MediaType == "" {
		return nil, errors.New("必须指定媒体ID、媒体ID列表或媒体类型")
	}

	// 创建更新任务
	task := u.taskManager.CreateTask(req)
	taskID := task.ID

	u.logger.Info("创建批量更新任务", "task_id", taskID, "media_count", len(req.MediaIDs))

	// 异步执行批量更新
	go u.executeBatchUpdate(ctx, task)

	// 返回任务响应
	return &MediaUpdateResponse{
		TaskID:    taskID,
		Status:    task.Status,
		Total:     u.calculateTotalMedias(req),
		Completed: 0,
		Failed:    0,
		Skipped:   0,
		StartTime: task.CreatedAt,
		Message:   "批量更新任务已创建并开始执行",
	}, nil
}

// GetUpdateTask 获取更新任务
func (u *mediaUpdater) GetUpdateTask(ctx context.Context, taskID string) (*MediaUpdateTask, error) {
	task := u.taskManager.GetTask(taskID)
	if task == nil {
		return nil, errors.New("任务不存在")
	}
	return task, nil
}

// ListUpdateTasks 列出更新任务
func (u *mediaUpdater) ListUpdateTasks(ctx context.Context, query *MediaUpdateTaskQuery) (*MediaUpdateTaskListResponse, error) {
	tasks, total := u.taskManager.ListTasks(query)

	// 默认分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// 计算分页
	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	return &MediaUpdateTaskListResponse{
		Tasks:      tasks,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// CancelUpdateTask 取消更新任务
func (u *mediaUpdater) CancelUpdateTask(ctx context.Context, taskID string) error {
	if !u.taskManager.CancelTask(taskID) {
		return errors.New("任务不存在或无法取消")
	}
	u.logger.Info("取消更新任务", "task_id", taskID)
	return nil
}

// GetUpdateStats 获取更新统计信息
func (u *mediaUpdater) GetUpdateStats(ctx context.Context) (*MediaUpdateStats, error) {
	return u.taskManager.GetStats(), nil
}

// shouldUpdateMedia 判断媒体是否需要更新
func (u *mediaUpdater) shouldUpdateMedia(ctx context.Context, req *MediaUpdateRequest, mediaID string) bool {
	// 强制更新策略
	if req.UpdateStrategy == StrategyForceUpdate {
		return true
	}

	// 获取媒体最后更新时间
	lastUpdateTime, err := u.mediaRepo.GetMediaUpdateTime(ctx, mediaID)
	if err != nil {
		// 获取失败，假设需要更新
		return true
	}

	// 缺失更新策略
	if req.UpdateStrategy == StrategyOnlyMissing && lastUpdateTime != nil {
		return false
	}

	// 增量或过期更新策略
	if lastUpdateTime != nil && req.UpdateStrategy != StrategyOnlyOutdated {
		// 检查元数据是否过期（超过7天）
		if time.Since(*lastUpdateTime) < 7*24*time.Hour {
			return false
		}
	}

	// 检查元数据是否最新
	if lastUpdateTime != nil {
		isUpToDate, err := u.metadataProvider.IsMetadataUpToDate(ctx, mediaID, *lastUpdateTime)
		if err == nil && isUpToDate {
			return false
		}
	}

	return true
}

// prepareMediaUpdates 准备媒体更新数据
func (u *mediaUpdater) prepareMediaUpdates(media, metadata interface{}, req *MediaUpdateRequest) (
	map[string]interface{},
	[]MediaUpdateField,
	[]MediaUpdateField,
) {
	updates := make(map[string]interface{})
	updatedFields := []MediaUpdateField{}
	changedFields := []MediaUpdateField{}

	// 确定需要更新的字段
	fieldsToUpdate := req.UpdateFields
	if len(fieldsToUpdate) == 0 || containsField(fieldsToUpdate, UpdateFieldAll) {
		fieldsToUpdate = u.getAllUpdateFields()
	}

	// 这里应该根据媒体类型和实际字段结构进行字段映射和更新
	// 暂时返回空结果
	return updates, updatedFields, changedFields
}

// calculateTotalMedias 计算总媒体数量
func (u *mediaUpdater) calculateTotalMedias(req *MediaUpdateRequest) int {
	if len(req.MediaIDs) > 0 {
		return len(req.MediaIDs)
	}
	if req.MediaID != "" {
		return 1
	}
	// 如果指定了媒体类型，需要查询数据库获取总数
	// 暂时返回默认值
	return 100
}

// executeBatchUpdate 执行批量更新
func (u *mediaUpdater) executeBatchUpdate(ctx context.Context, task *MediaUpdateTask) {
	// 更新任务状态
	task.Status = UpdateStatusInProgress
	task.UpdatedAt = time.Now()
	u.taskManager.UpdateTask(task)

	// 模拟批量更新执行
	// 实际实现应该根据task.Request中的条件获取媒体列表并逐一更新

	// 完成任务
	task.Status = UpdateStatusCompleted
	task.Progress.Completed = task.Progress.Total
	task.Progress.Percentage = 100
	task.CompletedAt = &time.Now()
	task.UpdatedAt = time.Now()
	u.taskManager.UpdateTask(task)

	u.logger.Info("批量更新任务完成", "task_id", task.ID)
}

// getAllUpdateFields 获取所有可更新字段
func (u *mediaUpdater) getAllUpdateFields() []MediaUpdateField {
	return []MediaUpdateField{
		UpdateFieldTitle,
		UpdateFieldOverview,
		UpdateFieldPoster,
		UpdateFieldBackdrop,
		UpdateFieldActors,
		UpdateFieldDirectors,
		UpdateFieldGenres,
		UpdateFieldReleaseDate,
		UpdateFieldRating,
		UpdateFieldRuntime,
		UpdateFieldStudio,
		UpdateFieldTags,
	}
}

// containsField 检查字段列表是否包含指定字段
func containsField(fields []MediaUpdateField, field MediaUpdateField) bool {
	for _, f := range fields {
		if f == field {
			return true
		}
	}
	return false
}
