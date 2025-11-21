package actions

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"moviepilot-go/pkg/logger"
)

// TorrentFilterValidator 种子过滤器验证器接口
type TorrentFilterValidator interface {
	// RegisterValidators 注册验证器到Gin框架
	RegisterValidators() error
	// ValidateFilterParams 验证过滤参数
	ValidateFilterParams(ctx context.Context, params *TorrentFilterParams) error
	// ValidateFilterField 验证过滤字段
	ValidateFilterField(ctx context.Context, field TorrentFilterField) error
	// ValidateFilterGroup 验证过滤条件组
	ValidateFilterGroup(ctx context.Context, group *TorrentFilterGroup) error
	// ValidateFilterCondition 验证过滤条件
	ValidateFilterCondition(ctx context.Context, condition *TorrentFilterCondition) error
	// ValidateExportParams 验证导出参数
	ValidateExportParams(ctx context.Context, params *TorrentExportParams) error
}

// torrentFilterValidator 种子过滤器验证器实现
type torrentFilterValidator struct {
	logger logger.Logger
	engine *validator.Validate
}

// NewTorrentFilterValidator 创建种子过滤器验证器实例
func NewTorrentFilterValidator(logger logger.Logger) TorrentFilterValidator {
	return &torrentFilterValidator{
		logger: logger,
	}
}

// RegisterValidators 注册验证器到Gin框架
func (v *torrentFilterValidator) RegisterValidators() error {
	if v.engine = validator.New(); v.engine == nil {
		return fmt.Errorf("创建验证器引擎失败")
	}

	// 获取Gin的验证器
	if bind, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.engine = bind
	} else {
		return fmt.Errorf("无法获取Gin验证器引擎")
	}

	// 注册自定义验证器
	if err := v.registerCustomValidators(); err != nil {
		return fmt.Errorf("注册自定义验证器失败: %w", err)
	}

	v.logger.Info("Torrent filter validators registered successfully")
	return nil
}

// ValidateFilterParams 验证过滤参数
func (v *torrentFilterValidator) ValidateFilterParams(ctx context.Context, params *TorrentFilterParams) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Validating torrent filter parameters")

	if params == nil {
		return fmt.Errorf("过滤参数不能为空")
	}

	// 验证基本字段
	if err := v.validateBasicParams(ctx, params); err != nil {
		return err
	}

	// 验证高级过滤条件
	if params.Filters != nil {
		if err := v.ValidateFilterGroup(ctx, params.Filters); err != nil {
			return fmt.Errorf("高级过滤条件验证失败: %w", err)
		}
	}

	// 验证特殊约束
	if err := v.validateSpecialConstraints(ctx, params); err != nil {
		return err
	}

	log.Debug("Torrent filter parameters validation successful")
	return nil
}

// ValidateFilterField 验证过滤字段
func (v *torrentFilterValidator) ValidateFilterField(ctx context.Context, field TorrentFilterField) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Validating filter field", "field", field)

	// 检查字段是否为有效的种子过滤字段
	validFields := map[TorrentFilterField]bool{
		TorrentFilterFieldID:             true,
		TorrentFilterFieldName:           true,
		TorrentFilterFieldHash:           true,
		TorrentFilterFieldSize:           true,
		TorrentFilterFieldProgress:       true,
		TorrentFilterFieldStatus:         true,
		TorrentFilterFieldType:           true,
		TorrentFilterFieldCategory:       true,
		TorrentFilterFieldTags:           true,
		TorrentFilterFieldTracker:        true,
		TorrentFilterFieldDownloader:     true,
		TorrentFilterFieldDownloadSpeed:  true,
		TorrentFilterFieldUploadSpeed:    true,
		TorrentFilterFieldRatio:          true,
		TorrentFilterFieldSeedingTime:    true,
		TorrentFilterFieldCreateTime:     true,
		TorrentFilterFieldAddTime:        true,
		TorrentFilterFieldCompletedTime:  true,
		TorrentFilterFieldLastActiveTime: true,
		TorrentFilterFieldMediaType:      true,
		TorrentFilterFieldMediaID:        true,
		TorrentFilterFieldQuality:        true,
	}

	if !validFields[field] {
		return fmt.Errorf("无效的过滤字段: %s", field)
	}

	return nil
}

// ValidateFilterGroup 验证过滤条件组
func (v *torrentFilterValidator) ValidateFilterGroup(ctx context.Context, group *TorrentFilterGroup) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Validating filter group")

	if group == nil {
		return fmt.Errorf("过滤条件组不能为空")
	}

	// 验证逻辑操作符
	if group.Logic != "and" && group.Logic != "or" {
		return fmt.Errorf("过滤条件组逻辑必须是 'and' 或 'or'")
	}

	// 验证条件列表
	if len(group.Conditions) == 0 {
		return fmt.Errorf("过滤条件组必须至少包含一个条件")
	}

	// 验证每个条件
	for i, condition := range group.Conditions {
		if err := v.validateCondition(ctx, condition, i); err != nil {
			return fmt.Errorf("条件 %d 验证失败: %w", i+1, err)
		}
	}

	// 检查递归深度
	if err := v.checkRecursiveDepth(ctx, group, 0); err != nil {
		return err
	}

	return nil
}

// ValidateFilterCondition 验证过滤条件
func (v *torrentFilterValidator) ValidateFilterCondition(ctx context.Context, condition *TorrentFilterCondition) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Validating filter condition", "field", condition.Field, "operator", condition.Operator)

	if condition == nil {
		return fmt.Errorf("过滤条件不能为空")
	}

	// 验证字段
	if err := v.ValidateFilterField(ctx, condition.Field); err != nil {
		return err
	}

	// 验证操作符
	if err := v.validateOperator(ctx, condition.Operator); err != nil {
		return err
	}

	// 根据字段类型和操作符验证值
	if err := v.validateConditionValue(ctx, condition); err != nil {
		return err
	}

	return nil
}

// ValidateExportParams 验证导出参数
func (v *torrentFilterValidator) ValidateExportParams(ctx context.Context, params *TorrentExportParams) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Validating export parameters")

	if params == nil {
		return fmt.Errorf("导出参数不能为空")
	}

	// 验证格式
	if err := v.validateExportFormat(ctx, params.Format); err != nil {
		return err
	}

	// 验证文件名
	if params.FileName != "" {
		if err := v.validateFileName(ctx, params.FileName); err != nil {
			return err
		}
	}

	// 验证过滤参数（如果有）
	if params.Filter != nil {
		if err := v.ValidateFilterParams(ctx, params.Filter); err != nil {
			return fmt.Errorf("导出过滤参数验证失败: %w", err)
		}
	}

	return nil
}

// 注册自定义验证器
func (v *torrentFilterValidator) registerCustomValidators() error {
	// 验证种子状态
	if err := v.engine.RegisterValidation("torrent_status", v.validateTorrentStatus); err != nil {
		return err
	}

	// 验证种子类型
	if err := v.engine.RegisterValidation("torrent_type", v.validateTorrentType); err != nil {
		return err
	}

	// 验证媒体类型
	if err := v.engine.RegisterValidation("media_type", v.validateMediaType); err != nil {
		return err
	}

	// 验证排序字段
	if err := v.engine.RegisterValidation("torrent_sort_field", v.validateSortField); err != nil {
		return err
	}

	// 验证排序顺序
	if err := v.engine.RegisterValidation("sort_order", v.validateSortOrder); err != nil {
		return err
	}

	// 验证过滤操作符
	if err := v.engine.RegisterValidation("filter_operator", v.validateOperatorRule); err != nil {
		return err
	}

	// 验证导出格式
	if err := v.engine.RegisterValidation("export_format", v.validateExportFormatRule); err != nil {
		return err
	}

	// 验证文件名
	if err := v.engine.RegisterValidation("file_name", v.validateFileNameRule); err != nil {
		return err
	}

	return nil
}

// 验证基本参数
func (v *torrentFilterValidator) validateBasicParams(ctx context.Context, params *TorrentFilterParams) error {
	// 分页参数验证
	if params.Page < 0 {
		return fmt.Errorf("page必须大于等于0")
	}

	if params.Limit < 0 || params.Limit > 1000 {
		return fmt.Errorf("limit必须在0到1000之间")
	}

	if params.Offset < 0 {
		return fmt.Errorf("offset必须大于等于0")
	}

	// 排序参数验证
	if err := v.validateSortField(ctx, params.SortBy); err != nil {
		return err
	}

	if params.SortOrder != "" && params.SortOrder != SortOrderAsc && params.SortOrder != SortOrderDesc {
		return fmt.Errorf("排序顺序必须是 'asc' 或 'desc'")
	}

	// 验证枚举类型
	for _, status := range params.Statuses {
		if err := v.validateTorrentStatus(ctx, string(status)); err != nil {
			return fmt.Errorf("无效的种子状态: %s", status)
		}
	}

	for _, t := range params.Types {
		if err := v.validateTorrentType(ctx, string(t)); err != nil {
			return fmt.Errorf("无效的种子类型: %s", t)
		}
	}

	for _, mediaType := range params.MediaTypes {
		if err := v.validateMediaType(ctx, string(mediaType)); err != nil {
			return fmt.Errorf("无效的媒体类型: %s", mediaType)
		}
	}

	// 验证数值范围
	if err := v.validateNumericRanges(ctx, params); err != nil {
		return err
	}

	// 验证时间范围
	if err := v.validateTimeRanges(ctx, params); err != nil {
		return err
	}

	return nil
}

// 验证特殊约束
func (v *torrentFilterValidator) validateSpecialConstraints(ctx context.Context, params *TorrentFilterParams) error {
	// 检查互斥的特殊过滤条件
	specialFlags := []bool{
		params.OnlyActive,
		params.OnlyCompleted,
		params.OnlyDownloading,
		params.OnlySeeding,
		params.OnlyPaused,
		params.OnlyStalled,
	}

	activeFlags := 0
	for _, flag := range specialFlags {
		if flag {
			activeFlags++
		}
	}

	// 允许最多两个互斥条件（如同时筛选下载中和暂停的）
	if activeFlags > 2 {
		return fmt.Errorf("只能同时使用最多两个特殊状态过滤条件")
	}

	// 检查冲突的条件
	if params.OnlyCompleted && params.OnlyDownloading {
		return fmt.Errorf("'only_completed' 和 'only_downloading' 条件不能同时使用")
	}

	if params.OnlyCompleted && params.OnlyPaused {
		return fmt.Errorf("'only_completed' 和 'only_paused' 条件不能同时使用")
	}

	if params.OnlyActive && params.OnlyPaused {
		return fmt.Errorf("'only_active' 和 'only_paused' 条件不能同时使用")
	}

	return nil
}

// 验证数值范围
func (v *torrentFilterValidator) validateNumericRanges(ctx context.Context, params *TorrentFilterParams) error {
	// 大小范围
	if params.SizeMin != nil && params.SizeMax != nil {
		if *params.SizeMin > *params.SizeMax {
			return fmt.Errorf("size_min 不能大于 size_max")
		}
		if *params.SizeMin < 0 || *params.SizeMax < 0 {
			return fmt.Errorf("大小范围不能为负数")
		}
	}

	// 比率范围
	if params.RatioMin != nil && params.RatioMax != nil {
		if *params.RatioMin > *params.RatioMax {
			return fmt.Errorf("ratio_min 不能大于 ratio_max")
		}
		if *params.RatioMin < 0 {
			return fmt.Errorf("ratio_min 不能为负数")
		}
	}

	// 进度范围
	if params.ProgressMin != nil && params.ProgressMax != nil {
		if *params.ProgressMin > *params.ProgressMax {
			return fmt.Errorf("progress_min 不能大于 progress_max")
		}
		if *params.ProgressMin < 0 || *params.ProgressMax < 0 {
			return fmt.Errorf("进度值不能为负数")
		}
		if *params.ProgressMin > 100 || *params.ProgressMax > 100 {
			return fmt.Errorf("进度值不能超过100")
		}
	}

	// 下载速度范围
	if params.DownloadSpeedMin != nil && params.DownloadSpeedMax != nil {
		if *params.DownloadSpeedMin > *params.DownloadSpeedMax {
			return fmt.Errorf("download_speed_min 不能大于 download_speed_max")
		}
		if *params.DownloadSpeedMin < 0 || *params.DownloadSpeedMax < 0 {
			return fmt.Errorf("下载速度不能为负数")
		}
	}

	// 上传速度范围
	if params.UploadSpeedMin != nil && params.UploadSpeedMax != nil {
		if *params.UploadSpeedMin > *params.UploadSpeedMax {
			return fmt.Errorf("upload_speed_min 不能大于 upload_speed_max")
		}
		if *params.UploadSpeedMin < 0 || *params.UploadSpeedMax < 0 {
			return fmt.Errorf("上传速度不能为负数")
		}
	}

	// 做种时间范围
	if params.SeedingTimeMin != nil && params.SeedingTimeMax != nil {
		if *params.SeedingTimeMin > *params.SeedingTimeMax {
			return fmt.Errorf("seeding_time_min 不能大于 seeding_time_max")
		}
		if *params.SeedingTimeMin < 0 || *params.SeedingTimeMax < 0 {
			return fmt.Errorf("做种时间不能为负数")
		}
	}

	return nil
}

// 验证时间范围
func (v *torrentFilterValidator) validateTimeRanges(ctx context.Context, params *TorrentFilterParams) error {
	// 创建时间范围
	if params.CreateTimeFrom != nil && params.CreateTimeTo != nil {
		if params.CreateTimeFrom.After(*params.CreateTimeTo) {
			return fmt.Errorf("create_time_from 不能晚于 create_time_to")
		}
		if params.CreateTimeFrom.After(time.Now()) || params.CreateTimeTo.After(time.Now().AddDate(1, 0, 0)) {
			return fmt.Errorf("创建时间范围无效")
		}
	}

	// 添加时间范围
	if params.AddTimeFrom != nil && params.AddTimeTo != nil {
		if params.AddTimeFrom.After(*params.AddTimeTo) {
			return fmt.Errorf("add_time_from 不能晚于 add_time_to")
		}
		if params.AddTimeFrom.After(time.Now()) || params.AddTimeTo.After(time.Now()) {
			return fmt.Errorf("添加时间范围无效")
		}
	}

	// 完成时间范围
	if params.CompletedTimeFrom != nil && params.CompletedTimeTo != nil {
		if params.CompletedTimeFrom.After(*params.CompletedTimeTo) {
			return fmt.Errorf("completed_time_from 不能晚于 completed_time_to")
		}
		if params.CompletedTimeFrom.After(time.Now()) || params.CompletedTimeTo.After(time.Now()) {
			return fmt.Errorf("完成时间范围无效")
		}
	}

	// 最后活动时间范围
	if params.LastActiveTimeFrom != nil && params.LastActiveTimeTo != nil {
		if params.LastActiveTimeFrom.After(*params.LastActiveTimeTo) {
			return fmt.Errorf("last_active_time_from 不能晚于 last_active_time_to")
		}
		if params.LastActiveTimeFrom.After(time.Now()) || params.LastActiveTimeTo.After(time.Now()) {
			return fmt.Errorf("最后活动时间范围无效")
		}
	}

	return nil
}

// 验证条件
func (v *torrentFilterValidator) validateCondition(ctx context.Context, condition interface{}, index int) error {
	switch cond := condition.(type) {
	case *TorrentFilterCondition:
		return v.ValidateFilterCondition(ctx, cond)
	case *TorrentFilterGroup:
		return v.ValidateFilterGroup(ctx, cond)
	default:
		return fmt.Errorf("条件类型无效: %T", condition)
	}
}

// 验证递归深度
func (v *torrentFilterValidator) checkRecursiveDepth(ctx context.Context, group *TorrentFilterGroup, currentDepth int) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Checking recursive depth", "current_depth", currentDepth)

	const maxDepth = 5

	if currentDepth > maxDepth {
		return fmt.Errorf("过滤条件组嵌套深度超过限制（最大%d层）", maxDepth)
	}

	for _, condition := range group.Conditions {
		if subGroup, ok := condition.(*TorrentFilterGroup); ok {
			if err := v.checkRecursiveDepth(ctx, subGroup, currentDepth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// 验证操作符
func (v *torrentFilterValidator) validateOperator(ctx context.Context, operator FilterOperator) error {
	validOperators := map[FilterOperator]bool{
		FilterOperatorEq:           true,
		FilterOperatorNe:           true,
		FilterOperatorGt:           true,
		FilterOperatorGte:          true,
		FilterOperatorLt:           true,
		FilterOperatorLte:          true,
		FilterOperatorLike:         true,
		FilterOperatorNotLike:      true,
		FilterOperatorIn:           true,
		FilterOperatorNotIn:        true,
		FilterOperatorRegex:        true,
		FilterOperatorNotRegex:     true,
		FilterOperatorBetween:      true,
		FilterOperatorIsNull:       true,
		FilterOperatorIsNotNull:    true,
		FilterOperatorStartsWith:   true,
		FilterOperatorEndsWith:     true,
	}

	if !validOperators[operator] {
		return fmt.Errorf("无效的操作符: %s", operator)
	}

	return nil
}

// 验证条件值
func (v *torrentFilterValidator) validateConditionValue(ctx context.Context, condition *TorrentFilterCondition) error {
	// 对于空值检查类操作符，不验证值
	if condition.Operator == FilterOperatorIsNull || condition.Operator == FilterOperatorIsNotNull {
		return nil
	}

	// 其他操作符必须有值
	if condition.Value == nil {
		return fmt.Errorf("操作符 %s 必须提供值", condition.Operator)
	}

	// 根据字段类型验证值
	switch condition.Field {
	case TorrentFilterFieldSize,
		TorrentFilterFieldProgress,
		TorrentFilterFieldDownloadSpeed,
		TorrentFilterFieldUploadSpeed,
		TorrentFilterFieldRatio,
		TorrentFilterFieldSeedingTime:
		if err := v.validateNumericValue(ctx, condition.Value, condition.Operator); err != nil {
			return fmt.Errorf("数值字段 %s 的值验证失败: %w", condition.Field, err)
		}

	case TorrentFilterFieldCreateTime,
		TorrentFilterFieldAddTime,
		TorrentFilterFieldCompletedTime,
		TorrentFilterFieldLastActiveTime:
		if err := v.validateTimeValue(ctx, condition.Value, condition.Operator); err != nil {
			return fmt.Errorf("时间字段 %s 的值验证失败: %w", condition.Field, err)
		}

	case TorrentFilterFieldStatus,
		TorrentFilterFieldType,
		TorrentFilterFieldMediaType:
		if err := v.validateEnumValue(ctx, condition.Value, condition.Operator, condition.Field); err != nil {
			return fmt.Errorf("枚举字段 %s 的值验证失败: %w", condition.Field, err)
		}

	case TorrentFilterFieldHash:
		if err := v.validateHashValue(ctx, condition.Value, condition.Operator); err != nil {
			return fmt.Errorf("哈希字段的值验证失败: %w", err)
		}

	case TorrentFilterFieldID:
		if err := v.validateIDValue(ctx, condition.Value, condition.Operator); err != nil {
			return fmt.Errorf("ID字段的值验证失败: %w", err)
		}

	default:
		// 字符串字段，验证格式
		if err := v.validateStringValue(ctx, condition.Value, condition.Operator); err != nil {
			return fmt.Errorf("字符串字段 %s 的值验证失败: %w", condition.Field, err)
		}
	}

	return nil
}

// 验证数值类型值
func (v *torrentFilterValidator) validateNumericValue(ctx context.Context, value interface{}, operator FilterOperator) error {
	// 对于IN操作符，值应该是数组
	if operator == FilterOperatorIn || operator == FilterOperatorNotIn {
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if err := v.isNumeric(item); err != nil {
					return fmt.Errorf("数组元素 %d 不是有效数值", i)
				}
			}
			return nil
		}
		return fmt.Errorf("IN/NotIN操作符的值必须是数组")
	}

	// 对于BETWEEN操作符，值应该是两个元素的数组
	if operator == FilterOperatorBetween {
		if arr, ok := value.([]interface{}); ok && len(arr) == 2 {
			if err := v.isNumeric(arr[0]); err != nil {
				return fmt.Errorf("范围开始值不是有效数值")
			}
			if err := v.isNumeric(arr[1]); err != nil {
				return fmt.Errorf("范围结束值不是有效数值")
			}
			return nil
		}
		return fmt.Errorf("BETWEEN操作符的值必须是包含两个元素的数组")
	}

	// 对于其他操作符，值应该是单一数值
	return v.isNumeric(value)
}

// 验证时间类型值
func (v *torrentFilterValidator) validateTimeValue(ctx context.Context, value interface{}, operator FilterOperator) error {
	// 对于IN操作符，值应该是时间字符串数组
	if operator == FilterOperatorIn || operator == FilterOperatorNotIn {
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if str, ok := item.(string); ok {
					if err := v.isValidTime(str); err != nil {
						return fmt.Errorf("时间字符串 %d 格式无效", i)
					}
				} else {
					return fmt.Errorf("数组元素 %d 不是字符串类型", i)
				}
			}
			return nil
		}
		return fmt.Errorf("IN/NotIN操作符的值必须是数组")
	}

	// 对于BETWEEN操作符，值应该是两个时间字符串的数组
	if operator == FilterOperatorBetween {
		if arr, ok := value.([]interface{}); ok && len(arr) == 2 {
			for i, item := range arr {
				if str, ok := item.(string); ok {
					if err := v.isValidTime(str); err != nil {
						return fmt.Errorf("时间字符串 %d 格式无效", i)
					}
				} else {
					return fmt.Errorf("时间范围元素 %d 不是字符串类型", i)
				}
			}
			return nil
		}
		return fmt.Errorf("BETWEEN操作符的值必须是包含两个元素的数组")
	}

	// 对于其他操作符，值应该是单一时间字符串
	if str, ok := value.(string); ok {
		return v.isValidTime(str)
	}

	return fmt.Errorf("时间值必须是字符串类型")
}

// 验证枚举类型值
func (v *torrentFilterValidator) validateEnumValue(ctx context.Context, value interface{}, operator FilterOperator, field TorrentFilterField) error {
	// 对于IN操作符，值应该是枚举字符串数组
	if operator == FilterOperatorIn || operator == FilterOperatorNotIn {
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if str, ok := item.(string); ok {
					var err error
					switch field {
					case TorrentFilterFieldStatus:
						err = v.validateTorrentStatus(ctx, str)
					case TorrentFilterFieldType:
						err = v.validateTorrentType(ctx, str)
					case TorrentFilterFieldMediaType:
						err = v.validateMediaType(ctx, str)
					}
					if err != nil {
						return fmt.Errorf("枚举值 %d 无效: %w", i, err)
					}
				} else {
					return fmt.Errorf("数组元素 %d 不是字符串类型", i)
				}
			}
			return nil
		}
		return fmt.Errorf("IN/NotIN操作符的值必须是数组")
	}

	// 对于其他操作符，值应该是单一枚举字符串
	if str, ok := value.(string); ok {
		switch field {
		case TorrentFilterFieldStatus:
			return v.validateTorrentStatus(ctx, str)
		case TorrentFilterFieldType:
			return v.validateTorrentType(ctx, str)
		case TorrentFilterFieldMediaType:
			return v.validateMediaType(ctx, str)
		}
	}

	return fmt.Errorf("枚举值必须是字符串类型")
}

// 验证哈希值
func (v *torrentFilterValidator) validateHashValue(ctx context.Context, value interface{}, operator FilterOperator) error {
	// 对于IN操作符，值应该是哈希字符串数组
	if operator == FilterOperatorIn || operator == FilterOperatorNotIn {
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if str, ok := item.(string); ok {
					if err := v.isValidHash(str); err != nil {
						return fmt.Errorf("哈希值 %d 无效", i)
					}
				} else {
					return fmt.Errorf("数组元素 %d 不是字符串类型", i)
				}
			}
			return nil
		}
		return fmt.Errorf("IN/NotIN操作符的值必须是数组")
	}

	// 对于其他操作符，值应该是单一哈希字符串
	if str, ok := value.(string); ok {
		return v.isValidHash(str)
	}

	return fmt.Errorf("哈希值必须是字符串类型")
}

// 验证ID值
func (v *torrentFilterValidator) validateIDValue(ctx context.Context, value interface{}, operator FilterOperator) error {
	// 对于IN操作符，值应该是ID数组（字符串或数字）
	if operator == FilterOperatorIn || operator == FilterOperatorNotIn {
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if _, ok := item.(string); !ok {
					if _, ok := item.(float64); !ok { // JSON中的数字是float64
						return fmt.Errorf("ID数组元素 %d 不是有效类型", i)
					}
				}
			}
			return nil
		}
		return fmt.Errorf("IN/NotIN操作符的值必须是数组")
	}

	// 对于其他操作符，值可以是字符串或数字
	_, isString := value.(string)
	_, isNumber := value.(float64)
	if !isString && !isNumber {
		return fmt.Errorf("ID值必须是字符串或数字类型")
	}

	return nil
}

// 验证字符串值
func (v *torrentFilterValidator) validateStringValue(ctx context.Context, value interface{}, operator FilterOperator) error {
	// 对于IN操作符，值应该是字符串数组
	if operator == FilterOperatorIn || operator == FilterOperatorNotIn {
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if _, ok := item.(string); !ok {
					return fmt.Errorf("数组元素 %d 不是字符串类型", i)
				}
			}
			return nil
		}
		return fmt.Errorf("IN/NotIN操作符的值必须是数组")
	}

	// 对于正则表达式操作符，验证正则表达式语法
	if operator == FilterOperatorRegex || operator == FilterOperatorNotRegex {
		if str, ok := value.(string); ok {
			if _, err := regexp.Compile(str); err != nil {
				return fmt.Errorf("正则表达式语法错误: %w", err)
			}
		}
		return fmt.Errorf("正则表达式操作符的值必须是字符串类型")
	}

	// 对于其他操作符，值应该是单一字符串
	if _, ok := value.(string); !ok {
		return fmt.Errorf("字符串值必须是字符串类型")
	}

	return nil
}

// 验证导出格式
func (v *torrentFilterValidator) validateExportFormat(ctx context.Context, format string) error {
	validFormats := map[string]bool{
		"json":  true,
		"csv":   true,
		"tsv":   true,
		"excel": true,
	}

	if !validFormats[format] {
		return fmt.Errorf("不支持的导出格式: %s", format)
	}

	return nil
}

// 验证文件名
func (v *torrentFilterValidator) validateFileName(ctx context.Context, fileName string) error {
	// 检查文件名长度
	if len(fileName) > 255 {
		return fmt.Errorf("文件名长度不能超过255个字符")
	}

	// 检查文件名是否包含非法字符
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)
	if invalidChars.MatchString(fileName) {
		return fmt.Errorf("文件名包含非法字符")
	}

	// 检查文件名是否为空字符串
	if strings.TrimSpace(fileName) == "" {
		return fmt.Errorf("文件名不能为空")
	}

	return nil
}

// 验证种子状态
func (v *torrentFilterValidator) validateTorrentStatus(ctx context.Context, status string) error {
	validStatuses := map[string]bool{
		string(TorrentStatusDownloading): true,
		string(TorrentStatusSeeding):     true,
		string(TorrentStatusCompleted):   true,
		string(TorrentStatusPaused):      true,
		string(TorrentStatusQueued):      true,
		string(TorrentStatusPending):     true,
		string(TorrentStatusChecking):    true,
		string(TorrentStatusError):       true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("无效的种子状态: %s", status)
	}

	return nil
}

// 验证种子类型
func (v *torrentFilterValidator) validateTorrentType(ctx context.Context, torrentType string) error {
	validTypes := map[string]bool{
		string(TorrentTypeMovie):   true,
		string(TorrentTypeTV):      true,
		string(TorrentTypeMusic):   true,
		string(TorrentTypeGame):    true,
		string(TorrentTypeSoftware): true,
		string(TorrentTypeOther):   true,
	}

	if !validTypes[torrentType] {
		return fmt.Errorf("无效的种子类型: %s", torrentType)
	}

	return nil
}

// 验证媒体类型
func (v *torrentFilterValidator) validateMediaType(ctx context.Context, mediaType string) error {
	validMediaTypes := map[string]bool{
		string(MediaTypeMovie):    true,
		string(MediaTypeTV):       true,
		string(MediaTypeMusic):    true,
		string(MediaTypeGame):     true,
		string(MediaTypeSoftware): true,
		string(MediaTypeBook):     true,
		string(MediaTypeOther):    true,
	}

	if !validMediaTypes[mediaType] {
		return fmt.Errorf("无效的媒体类型: %s", mediaType)
	}

	return nil
}

// 验证排序字段
func (v *torrentFilterValidator) validateSortField(ctx context.Context, field TorrentSortField) error {
	if field == "" {
		return nil // 允许空值，使用默认排序
	}

	validFields := map[TorrentSortField]bool{
		TorrentSortFieldID:              true,
		TorrentSortFieldName:            true,
		TorrentSortFieldHash:            true,
		TorrentSortFieldSize:            true,
		TorrentSortFieldProgress:        true,
		TorrentSortFieldStatus:          true,
		TorrentSortFieldType:            true,
		TorrentSortFieldCategory:        true,
		TorrentSortFieldDownloadSpeed:   true,
		TorrentSortFieldUploadSpeed:     true,
		TorrentSortFieldRatio:           true,
		TorrentSortFieldSeedingTime:     true,
		TorrentSortFieldCreateTime:      true,
		TorrentSortFieldAddTime:         true,
		TorrentSortFieldCompletedTime:   true,
		TorrentSortFieldLastActiveTime:  true,
		TorrentSortFieldMediaType:       true,
		TorrentSortFieldMediaID:         true,
		TorrentSortFieldQuality:         true,
	}

	if !validFields[field] {
		return fmt.Errorf("无效的排序字段: %s", field)
	}

	return nil
}

// 验证排序顺序
func (v *torrentFilterValidator) validateSortOrder(ctx context.Context, order SortOrder) error {
	if order == "" {
		return nil // 允许空值，使用默认排序顺序
	}

	if order != SortOrderAsc && order != SortOrderDesc {
		return fmt.Errorf("无效的排序顺序: %s", order)
	}

	return nil
}

// 检查是否为数值
func (v *torrentFilterValidator) isNumeric(value interface{}) error {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
		return nil
	case reflect.String:
		// 尝试将字符串转换为浮点数
		if _, err := strconv.ParseFloat(value.(string), 64); err == nil {
			return nil
		}
		return fmt.Errorf("字符串不是有效数值")
	default:
		return fmt.Errorf("类型不是有效数值")
	}
}

// 检查是否为有效时间
func (v *torrentFilterValidator) isValidTime(timeStr string) error {
	// 尝试解析为时间
	layout := "2006-01-02T15:04:05Z07:00" // RFC3339格式
	_, err := time.Parse(layout, timeStr)
	if err != nil {
		// 尝试其他格式
		layout = "2006-01-02 15:04:05"
		_, err = time.Parse(layout, timeStr)
		if err != nil {
			return fmt.Errorf("时间格式无效")
		}
	}

	return nil
}

// 检查是否为有效哈希
func (v *torrentFilterValidator) isValidHash(hash string) error {
	// 检查SHA-1哈希格式（40个十六进制字符）
	sha1Pattern := regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
	if sha1Pattern.MatchString(hash) {
		return nil
	}

	// 检查SHA-256哈希格式（64个十六进制字符）
	sha256Pattern := regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	if sha256Pattern.MatchString(hash) {
		return nil
	}

	return fmt.Errorf("哈希格式无效，必须是SHA-1或SHA-256格式")
}

// 验证器规则函数（用于Gin验证）

func (v *torrentFilterValidator) validateTorrentStatus(fl validator.FieldLevel) bool {
	return v.validateTorrentStatus(context.Background(), fl.Field().String()) == nil
}

func (v *torrentFilterValidator) validateTorrentType(fl validator.FieldLevel) bool {
	return v.validateTorrentType(context.Background(), fl.Field().String()) == nil
}

func (v *torrentFilterValidator) validateMediaType(fl validator.FieldLevel) bool {
	return v.validateMediaType(context.Background(), fl.Field().String()) == nil
}

func (v *torrentFilterValidator) validateSortField(fl validator.FieldLevel) bool {
	return v.validateSortField(context.Background(), TorrentSortField(fl.Field().String())) == nil
}

func (v *torrentFilterValidator) validateSortOrder(fl validator.FieldLevel) bool {
	return v.validateSortOrder(context.Background(), SortOrder(fl.Field().String())) == nil
}

func (v *torrentFilterValidator) validateOperatorRule(fl validator.FieldLevel) bool {
	return v.validateOperator(context.Background(), FilterOperator(fl.Field().String())) == nil
}

func (v *torrentFilterValidator) validateExportFormatRule(fl validator.FieldLevel) bool {
	return v.validateExportFormat(context.Background(), fl.Field().String()) == nil
}

func (v *torrentFilterValidator) validateFileNameRule(fl validator.FieldLevel) bool {
	return v.validateFileName(context.Background(), fl.Field().String()) == nil
}
