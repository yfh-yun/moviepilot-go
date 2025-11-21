package actions

import (
	"github.com/google/uuid"
	"moviepilot-go/pkg/errors"
)

// CleanTorrentsValidator 种子清理验证器接口
type CleanTorrentsValidator interface {
	// ValidateCleanRequest 验证清理请求参数
	ValidateCleanRequest(req CleanTorrentRequest) error
	
	// ValidateTaskID 验证任务ID
	ValidateTaskID(taskID string) error
	
	// ValidatePagination 验证分页参数
	ValidatePagination(limit, offset int) error
	
	// ValidateCleanStrategy 验证清理策略
	ValidateCleanStrategy(strategy CleanStrategy) error
}

// CleanTorrentsValidatorImpl 种子清理验证器实现
type CleanTorrentsValidatorImpl struct{}

// NewCleanTorrentsValidator 创建种子清理验证器实例
func NewCleanTorrentsValidator() *CleanTorrentsValidatorImpl {
	return &CleanTorrentsValidatorImpl{}
}

// ValidateCleanRequest 验证清理请求参数
func (v *CleanTorrentsValidatorImpl) ValidateCleanRequest(req CleanTorrentRequest) error {
	// 验证清理策略
	if err := v.ValidateCleanStrategy(req.Strategy); err != nil {
		return err
	}

	// 根据策略验证相应的阈值参数
	switch req.Strategy {
	case CleanStrategyByTime:
		if req.TimeThreshold <= 0 {
			return errors.NewValidationError("time_threshold must be greater than 0 when strategy is by_time")
		}
	case CleanStrategyByRatio:
		if req.RatioThreshold < 0 {
			return errors.NewValidationError("ratio_threshold must be greater than or equal to 0 when strategy is by_ratio")
		}
	case CleanStrategyBySeeder:
		if req.SeederThreshold < 0 {
			return errors.NewValidationError("seeder_threshold must be greater than or equal to 0 when strategy is by_seeder")
		}
	case CleanStrategyByStorage:
		if req.StorageThreshold <= 0 {
			return errors.NewValidationError("storage_threshold must be greater than 0 when strategy is by_storage")
		}
	}

	// 至少包含一种状态的种子
	if !req.IncludeCompleted && !req.IncludeDownloading && !req.IncludePaused {
		return errors.NewValidationError("at least one of include_completed, include_downloading, or include_paused must be true")
	}

	// 验证排除标签和Tracker列表（如果提供）
	if len(req.ExcludeTags) > 100 {
		return errors.NewValidationError("exclude_tags cannot contain more than 100 items")
	}
	if len(req.ExcludeTrackers) > 100 {
		return errors.NewValidationError("exclude_trackers cannot contain more than 100 items")
	}

	// 验证下载器列表（如果提供）
	if len(req.Downloaders) > 10 {
		return errors.NewValidationError("downloaders cannot contain more than 10 items")
	}

	return nil
}

// ValidateTaskID 验证任务ID
func (v *CleanTorrentsValidatorImpl) ValidateTaskID(taskID string) error {
	if taskID == "" {
		return errors.NewValidationError("task_id is required")
	}

	// 验证UUID格式
	if _, err := uuid.Parse(taskID); err != nil {
		return errors.NewValidationError("task_id must be a valid UUID format")
	}

	return nil
}

// ValidatePagination 验证分页参数
func (v *CleanTorrentsValidatorImpl) ValidatePagination(limit, offset int) error {
	if limit < 1 {
		return errors.NewValidationError("limit must be greater than 0")
	}
	if limit > 100 {
		return errors.NewValidationError("limit cannot exceed 100")
	}
	if offset < 0 {
		return errors.NewValidationError("offset cannot be negative")
	}

	return nil
}

// ValidateCleanStrategy 验证清理策略
func (v *CleanTorrentsValidatorImpl) ValidateCleanStrategy(strategy CleanStrategy) error {
	validStrategies := []CleanStrategy{
		CleanStrategyByTime,
		CleanStrategyByRatio,
		CleanStrategyBySeeder,
		CleanStrategyByStorage,
	}

	for _, validStrategy := range validStrategies {
		if strategy == validStrategy {
			return nil
		}
	}

	return errors.NewValidationError("invalid clean strategy")
}
