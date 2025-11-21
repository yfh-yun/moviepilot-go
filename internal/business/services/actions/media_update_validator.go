package actions

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/validator"
)

// MediaUpdateValidator 媒体更新验证器接口
type MediaUpdateValidator interface {
	// ValidateUpdateRequest 验证单个更新请求
	ValidateUpdateRequest(c *gin.Context, req *MediaUpdateRequest) error
	// ValidateBatchUpdateRequest 验证批量更新请求
	ValidateBatchUpdateRequest(c *gin.Context, req *MediaUpdateRequest) error
	// ValidateTaskQuery 验证任务查询参数
	ValidateTaskQuery(c *gin.Context, query *MediaUpdateTaskQuery) error
	// ValidateMediaID 验证媒体ID
	ValidateMediaID(mediaID string) error
	// ValidateMediaType 验证媒体类型
	ValidateMediaType(mediaType string) error
	// ValidateUpdateFields 验证更新字段
	ValidateUpdateFields(fields []MediaUpdateField) error
}

// mediaUpdateValidator 媒体更新验证器实现
type mediaUpdateValidator struct {
	logger    logger.Logger
	validator *validator.Validator
}

// NewMediaUpdateValidator 创建媒体更新验证器实例
func NewMediaUpdateValidator(logger logger.Logger, v *validator.Validator) MediaUpdateValidator {
	return &mediaUpdateValidator{
		logger:    logger,
		validator: v,
	}
}

// ValidateUpdateRequest 验证单个更新请求
func (v *mediaUpdateValidator) ValidateUpdateRequest(c *gin.Context, req *MediaUpdateRequest) error {
	// 验证媒体ID
	if err := v.ValidateMediaID(req.MediaID); err != nil {
		return err
	}

	// 验证媒体类型
	if req.MediaType != "" {
		if err := v.ValidateMediaType(req.MediaType); err != nil {
			return err
		}
	}

	// 验证更新策略
	if !v.isValidUpdateStrategy(req.UpdateStrategy) {
		return errors.New("无效的更新策略")
	}

	// 验证更新字段
	if len(req.UpdateFields) > 0 {
		if err := v.ValidateUpdateFields(req.UpdateFields); err != nil {
			return err
		}
	}

	// 验证并发数
	if req.Concurrency < 0 || req.Concurrency > 100 {
		return errors.New("并发数必须在0-100之间")
	}

	// 验证超时时间
	if req.Timeout < 0 || req.Timeout > 86400 {
		return errors.New("超时时间必须在0-86400秒之间")
	}

	// 验证重试次数
	if req.MaxRetries < 0 || req.MaxRetries > 10 {
		return errors.New("最大重试次数必须在0-10之间")
	}

	return nil
}

// ValidateBatchUpdateRequest 验证批量更新请求
func (v *mediaUpdateValidator) ValidateBatchUpdateRequest(c *gin.Context, req *MediaUpdateRequest) error {
	// 验证媒体ID列表或媒体类型或媒体ID至少有一个
	if len(req.MediaIDs) == 0 && req.MediaID == "" && req.MediaType == "" {
		return errors.New("必须指定媒体ID、媒体ID列表或媒体类型")
	}

	// 如果指定了媒体ID列表，验证列表
	if len(req.MediaIDs) > 0 {
		if len(req.MediaIDs) > 1000 {
			return errors.New("媒体ID列表不能超过1000个")
		}

		for _, mediaID := range req.MediaIDs {
			if err := v.ValidateMediaID(mediaID); err != nil {
				return errors.New("无效的媒体ID: " + mediaID)
			}
		}
	}

	// 如果指定了媒体ID，验证单个ID
	if req.MediaID != "" {
		if err := v.ValidateMediaID(req.MediaID); err != nil {
			return err
		}
	}

	// 验证媒体类型
	if req.MediaType != "" {
		if err := v.ValidateMediaType(req.MediaType); err != nil {
			return err
		}
	}

	// 验证更新策略
	if !v.isValidUpdateStrategy(req.UpdateStrategy) {
		return errors.New("无效的更新策略")
	}

	// 验证更新字段
	if len(req.UpdateFields) > 0 {
		if err := v.ValidateUpdateFields(req.UpdateFields); err != nil {
			return err
		}
	}

	// 验证并发数（批量更新对并发数要求更严格）
	if req.Concurrency < 0 || req.Concurrency > 50 {
		return errors.New("批量更新并发数必须在0-50之间")
	}

	// 验证超时时间
	if req.Timeout < 0 || req.Timeout > 86400 {
		return errors.New("超时时间必须在0-86400秒之间")
	}

	// 验证重试次数
	if req.MaxRetries < 0 || req.MaxRetries > 5 {
		return errors.New("批量更新最大重试次数必须在0-5之间")
	}

	return nil
}

// ValidateTaskQuery 验证任务查询参数
func (v *mediaUpdateValidator) ValidateTaskQuery(c *gin.Context, query *MediaUpdateTaskQuery) error {
	// 验证分页参数
	if query.Page < 1 {
		return errors.New("页码必须大于0")
	}

	if query.PageSize < 1 || query.PageSize > 100 {
		return errors.New("每页数量必须在1-100之间")
	}

	// 验证任务状态
	if query.Status != "" {
		if !v.isValidTaskStatus(query.Status) {
			return errors.New("无效的任务状态")
		}
	}

	// 验证媒体类型
	if query.MediaType != "" {
		if err := v.ValidateMediaType(query.MediaType); err != nil {
			return err
		}
	}

	// 验证时间范围
	if query.StartTime != "" {
		if _, err := time.Parse(time.RFC3339, query.StartTime); err != nil {
			return errors.New("开始时间格式错误，应为RFC3339格式")
		}
	}

	if query.EndTime != "" {
		if _, err := time.Parse(time.RFC3339, query.EndTime); err != nil {
			return errors.New("结束时间格式错误，应为RFC3339格式")
		}
	}

	// 验证开始时间不晚于结束时间
	if query.StartTime != "" && query.EndTime != "" {
		startTime, _ := time.Parse(time.RFC3339, query.StartTime)
		endTime, _ := time.Parse(time.RFC3339, query.EndTime)
		if startTime.After(endTime) {
			return errors.New("开始时间不能晚于结束时间")
		}
	}

	return nil
}

// ValidateMediaID 验证媒体ID
func (v *mediaUpdateValidator) ValidateMediaID(mediaID string) error {
	if mediaID == "" {
		return errors.New("媒体ID不能为空")
	}

	if len(mediaID) > 100 {
		return errors.New("媒体ID长度不能超过100个字符")
	}

	// 可以添加更多的媒体ID格式验证
	return nil
}

// ValidateMediaType 验证媒体类型
func (v *mediaUpdateValidator) ValidateMediaType(mediaType string) error {
	if mediaType == "" {
		return errors.New("媒体类型不能为空")
	}

	// 检查媒体类型是否有效
	validMediaTypes := map[string]bool{
		"movie":      true,
		"tv":         true,
		"series":     true,
		"episode":    true,
		"anime":      true,
		"documentary": true,
	}

	if !validMediaTypes[mediaType] {
		return errors.New("无效的媒体类型")
	}

	return nil
}

// ValidateUpdateFields 验证更新字段
func (v *mediaUpdateValidator) ValidateUpdateFields(fields []MediaUpdateField) error {
	if len(fields) == 0 {
		return errors.New("更新字段不能为空")
	}

	// 定义有效的更新字段
	validFields := map[MediaUpdateField]bool{
		UpdateFieldAll:       true,
		UpdateFieldTitle:     true,
		UpdateFieldOverview:  true,
		UpdateFieldPoster:    true,
		UpdateFieldBackdrop:  true,
		UpdateFieldActors:    true,
		UpdateFieldDirectors: true,
		UpdateFieldGenres:    true,
		UpdateFieldReleaseDate: true,
		UpdateFieldRating:    true,
		UpdateFieldRuntime:   true,
		UpdateFieldStudio:    true,
		UpdateFieldTags:      true,
	}

	// 检查每个字段是否有效
	for _, field := range fields {
		if !validFields[field] {
			return errors.New("无效的更新字段")
		}
	}

	// 如果包含UpdateFieldAll，不应包含其他字段
	containsAll := false
	for _, field := range fields {
		if field == UpdateFieldAll {
			containsAll = true
			break
		}
	}

	if containsAll && len(fields) > 1 {
		return errors.New("不能同时指定'all'和其他具体字段")
	}

	return nil
}

// isValidUpdateStrategy 检查更新策略是否有效
func (v *mediaUpdateValidator) isValidUpdateStrategy(strategy UpdateStrategy) bool {
	validStrategies := map[UpdateStrategy]bool{
		StrategyForceUpdate:    true,
		StrategyIncremental:    true,
		StrategyOnlyOutdated:   true,
		StrategyOnlyMissing:    true,
	}
	return validStrategies[strategy]
}

// isValidTaskStatus 检查任务状态是否有效
func (v *mediaUpdateValidator) isValidTaskStatus(status string) bool {
	validStatuses := map[string]bool{
		UpdateStatusPending:   true,
		UpdateStatusInProgress: true,
		UpdateStatusCompleted: true,
		UpdateStatusFailed:    true,
		UpdateStatusCancelled: true,
	}
	return validStatuses[status]
}
