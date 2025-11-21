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

// MediaFilterValidator 媒体过滤器验证器接口
type MediaFilterValidator interface {
	// ValidateFilterParams 验证过滤参数
	ValidateFilterParams(ctx context.Context, params *MediaFilterParams) error
	// ValidateFilterField 验证过滤字段
	ValidateFilterField(field MediaFilterField, value interface{}, operator FilterOperator) error
	// ValidateFilterGroup 验证过滤条件组
	ValidateFilterGroup(ctx context.Context, group *FilterGroup) error
	// RegisterValidators 注册自定义验证器
	RegisterValidators() error
	// GetValidationErrors 获取验证错误信息
	GetValidationErrors(err error) map[string]string
}

// mediaFilterValidator 媒体过滤器验证器实现
type mediaFilterValidator struct {
	logger logger.Logger
}

// NewMediaFilterValidator 创建媒体过滤器验证器实例
func NewMediaFilterValidator(logger logger.Logger) MediaFilterValidator {
	return &mediaFilterValidator{
		logger: logger,
	}
}

// ValidateFilterParams 验证过滤参数
func (v *mediaFilterValidator) ValidateFilterParams(ctx context.Context, params *MediaFilterParams) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Validating filter parameters")

	if params == nil {
		return fmt.Errorf("过滤参数不能为空")
	}

	// 验证分页参数
	if params.Limit < 0 || params.Limit > 1000 {
		return fmt.Errorf("limit必须在0到1000之间")
	}

	if params.Offset < 0 {
		return fmt.Errorf("offset必须大于等于0")
	}

	if params.Page < 0 {
		return fmt.Errorf("page必须大于等于0")
	}

	// 验证排序参数
	if err := v.validateSortBy(params.SortBy); err != nil {
		return err
	}

	if err := v.validateSortOrder(params.SortOrder); err != nil {
		return err
	}

	// 验证媒体类型
	if err := v.validateMediaTypes(params.MediaTypes); err != nil {
		return err
	}

	// 验证媒体状态
	if err := v.validateMediaStatus(params.Status); err != nil {
		return err
	}

	// 验证本地状态
	if err := v.validateLocalStatus(params.LocalStatus); err != nil {
		return err
	}

	// 验证订阅状态
	if err := v.validateSubscribeStatus(params.SubscribeStatus); err != nil {
		return err
	}

	// 验证下载状态
	if err := v.validateDownloadStatus(params.DownloadStatus); err != nil {
		return err
	}

	// 验证字幕状态
	if err := v.validateSubtitleStatus(params.SubtitleStatus); err != nil {
		return err
	}

	// 验证年份
	if err := v.validateYears(params.Years); err != nil {
		return err
	}

	// 验证评分范围
	if err := v.validateRatingRange(params.RatingMin, params.RatingMax); err != nil {
		return err
	}

	// 验证投票数范围
	if err := v.validateVotesRange(params.VotesMin); err != nil {
		return err
	}

	// 验证时长范围
	if err := v.validateRuntimeRange(params.RuntimeMin, params.RuntimeMax); err != nil {
		return err
	}

	// 验证文件夹大小范围
	if err := v.validateFolderSizeRange(params.FolderSizeMin, params.FolderSizeMax); err != nil {
		return err
	}

	// 验证高级过滤条件
	if params.Filters != nil {
		if err := v.ValidateFilterGroup(ctx, params.Filters); err != nil {
			return fmt.Errorf("高级过滤条件验证失败: %w", err)
		}
	}

	log.Debug("Filter parameters validation successful")
	return nil
}

// ValidateFilterField 验证过滤字段
func (v *mediaFilterValidator) ValidateFilterField(field MediaFilterField, value interface{}, operator FilterOperator) error {
	// 验证字段名
	if err := v.validateFieldName(field); err != nil {
		return err
	}

	// 验证操作符
	if err := v.validateOperator(operator); err != nil {
		return err
	}

	// 验证值
	if operator != FilterOperatorIsNull && operator != FilterOperatorIsNotNull {
		if err := v.validateFieldValue(field, value, operator); err != nil {
			return err
		}
	}

	return nil
}

// ValidateFilterGroup 验证过滤条件组
func (v *mediaFilterValidator) ValidateFilterGroup(ctx context.Context, group *FilterGroup) error {
	log := v.logger.WithContext(ctx)
	log.Debug("Validating filter group")

	if group == nil {
		return fmt.Errorf("过滤条件组不能为空")
	}

	// 验证逻辑类型
	if group.Logic != "and" && group.Logic != "or" {
		return fmt.Errorf("过滤条件组逻辑必须是'and'或'or'")
	}

	// 验证条件数量
	if len(group.Conditions) == 0 {
		return fmt.Errorf("过滤条件组必须至少包含一个条件")
	}

	// 递归验证每个条件
	for i, condition := range group.Conditions {
		if err := v.validateFilterCondition(ctx, condition, i); err != nil {
			return err
		}
	}

	log.Debug("Filter group validation successful")
	return nil
}

// RegisterValidators 注册自定义验证器
func (v *mediaFilterValidator) RegisterValidators() error {
	log := v.logger
	log.Debug("Registering custom validators for media filter")

	// 获取validator实例
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册媒体类型验证器
		if err := v.RegisterValidation("media_type", v.validateMediaTypeFunc); err != nil {
			return fmt.Errorf("注册媒体类型验证器失败: %w", err)
		}

		// 注册媒体状态验证器
		if err := v.RegisterValidation("media_status", v.validateMediaStatusFunc); err != nil {
			return fmt.Errorf("注册媒体状态验证器失败: %w", err)
		}

		// 注册过滤字段验证器
		if err := v.RegisterValidation("filter_field", v.validateFilterFieldFunc); err != nil {
			return fmt.Errorf("注册过滤字段验证器失败: %w", err)
		}

		// 注册过滤操作符验证器
		if err := v.RegisterValidation("filter_operator", v.validateFilterOperatorFunc); err != nil {
			return fmt.Errorf("注册过滤操作符验证器失败: %w", err)
		}

		// 注册排序字段验证器
		if err := v.RegisterValidation("sort_field", v.validateSortFieldFunc); err != nil {
			return fmt.Errorf("注册排序字段验证器失败: %w", err)
		}

		// 注册排序顺序验证器
		if err := v.RegisterValidation("sort_order", v.validateSortOrderFunc); err != nil {
			return fmt.Errorf("注册排序顺序验证器失败: %w", err)
		}

		// 注册年份验证器
		if err := v.RegisterValidation("valid_year", v.validateYearFunc); err != nil {
			return fmt.Errorf("注册年份验证器失败: %w", err)
		}

		// 注册评分验证器
		if err := v.RegisterValidation("valid_rating", v.validateRatingFunc); err != nil {
			return fmt.Errorf("注册评分验证器失败: %w", err)
		}

		// 注册过滤条件组验证器
		if err := v.RegisterValidation("filter_group", v.validateFilterGroupFunc); err != nil {
			return fmt.Errorf("注册过滤条件组验证器失败: %w", err)
		}

		log.Info("Successfully registered all custom validators")
		return nil
	}

	return fmt.Errorf("无法获取validator实例")
}

// GetValidationErrors 获取验证错误信息
func (v *mediaFilterValidator) GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			field := e.Field()
			tag := e.Tag()
			errors[field] = v.getErrorMessage(field, tag, e.Param())
		}
	} else {
		errors["general"] = err.Error()
	}

	return errors
}

// 辅助方法：验证过滤条件
func (v *mediaFilterValidator) validateFilterCondition(ctx context.Context, condition interface{}, index int) error {
	log := v.logger.WithContext(ctx)

	switch cond := condition.(type) {
	case *FilterCondition:
		if cond == nil {
			return fmt.Errorf("条件 %d 不能为空", index)
		}

		if err := v.ValidateFilterField(cond.Field, cond.Value, cond.Operator); err != nil {
			return fmt.Errorf("条件 %d 验证失败: %w", index, err)
		}

	case *FilterGroup:
		if err := v.ValidateFilterGroup(ctx, cond); err != nil {
			return fmt.Errorf("子条件组 %d 验证失败: %w", index, err)
		}

	default:
		log.Error("Invalid condition type", "type", reflect.TypeOf(condition))
		return fmt.Errorf("条件 %d 类型无效: %T", index, condition)
	}

	return nil
}

// 辅助方法：验证排序字段
func (v *mediaFilterValidator) validateSortBy(sortBy MediaSortField) error {
	if sortBy == "" {
		return nil // 允许空值，使用默认排序
	}

	validSortFields := map[MediaSortField]bool{
		MediaSortFieldID:              true,
		MediaSortFieldTitle:           true,
		MediaSortFieldOriginalTitle:   true,
		MediaSortFieldType:            true,
		MediaSortFieldYear:            true,
		MediaSortFieldRating:          true,
		MediaSortFieldVotes:           true,
		MediaSortFieldRuntime:         true,
		MediaSortFieldSeasonCount:     true,
		MediaSortFieldEpisodeCount:    true,
		MediaSortFieldAirDate:         true,
		MediaSortFieldFirstAirDate:    true,
		MediaSortFieldLastAirDate:     true,
		MediaSortFieldReleaseDate:     true,
		MediaSortFieldCreateTime:      true,
		MediaSortFieldUpdateTime:      true,
		MediaSortFieldSortTitle:       true,
		MediaSortFieldLocalStatus:     true,
		MediaSortFieldSubscribeStatus: true,
		MediaSortFieldDownloadStatus:  true,
		MediaSortFieldFolderSize:      true,
		MediaSortFieldQuality:         true,
		MediaSortFieldResolution:      true,
	}

	if !validSortFields[sortBy] {
		return fmt.Errorf("无效的排序字段: %s", sortBy)
	}

	return nil
}

// 辅助方法：验证排序顺序
func (v *mediaFilterValidator) validateSortOrder(sortOrder SortOrder) error {
	if sortOrder == "" {
		return nil // 允许空值，使用默认排序顺序
	}

	if sortOrder != SortOrderAsc && sortOrder != SortOrderDesc {
		return fmt.Errorf("无效的排序顺序: %s，必须是 'asc' 或 'desc'", sortOrder)
	}

	return nil
}

// 辅助方法：验证媒体类型
func (v *mediaFilterValidator) validateMediaTypes(mediaTypes []MediaType) error {
	for _, mediaType := range mediaTypes {
		if !isValidMediaType(mediaType) {
			return fmt.Errorf("无效的媒体类型: %s", mediaType)
		}
	}

	return nil
}

// 辅助方法：验证媒体状态
func (v *mediaFilterValidator) validateMediaStatus(status []MediaStatus) error {
	for _, s := range status {
		if !isValidMediaStatus(s) {
			return fmt.Errorf("无效的媒体状态: %s", s)
		}
	}

	return nil
}

// 辅助方法：验证本地状态
func (v *mediaFilterValidator) validateLocalStatus(localStatus []LocalMediaStatus) error {
	for _, s := range localStatus {
		if !isValidLocalMediaStatus(s) {
			return fmt.Errorf("无效的本地状态: %s", s)
		}
	}

	return nil
}

// 辅助方法：验证订阅状态
func (v *mediaFilterValidator) validateSubscribeStatus(subscribeStatus []SubscribeStatus) error {
	for _, s := range subscribeStatus {
		if !isValidSubscribeStatus(s) {
			return fmt.Errorf("无效的订阅状态: %s", s)
		}
	}

	return nil
}

// 辅助方法：验证下载状态
func (v *mediaFilterValidator) validateDownloadStatus(downloadStatus []DownloadStatus) error {
	for _, s := range downloadStatus {
		if !isValidDownloadStatus(s) {
			return fmt.Errorf("无效的下载状态: %s", s)
		}
	}

	return nil
}

// 辅助方法：验证字幕状态
func (v *mediaFilterValidator) validateSubtitleStatus(subtitleStatus []SubtitleStatus) error {
	for _, s := range subtitleStatus {
		if !isValidSubtitleStatus(s) {
			return fmt.Errorf("无效的字幕状态: %s", s)
		}
	}

	return nil
}

// 辅助方法：验证年份
func (v *mediaFilterValidator) validateYears(years []int) error {
	currentYear := time.Now().Year()
	for _, year := range years {
		if year < 1800 || year > currentYear+1 { // 允许未来一年的预测年份
			return fmt.Errorf("无效的年份: %d，必须在1800到%d之间", year, currentYear+1)
		}
	}

	return nil
}

// 辅助方法：验证评分范围
func (v *mediaFilterValidator) validateRatingRange(min, max *float64) error {
	if min != nil {
		if *min < 0 || *min > 10 {
			return fmt.Errorf("评分最小值必须在0到10之间")
		}
	}

	if max != nil {
		if *max < 0 || *max > 10 {
			return fmt.Errorf("评分最大值必须在0到10之间")
		}
	}

	if min != nil && max != nil {
		if *min > *max {
			return fmt.Errorf("评分最小值不能大于最大值")
		}
	}

	return nil
}

// 辅助方法：验证投票数范围
func (v *mediaFilterValidator) validateVotesRange(min *int) error {
	if min != nil {
		if *min < 0 {
			return fmt.Errorf("投票数最小值不能为负数")
		}
	}

	return nil
}

// 辅助方法：验证时长范围
func (v *mediaFilterValidator) validateRuntimeRange(min, max *int) error {
	if min != nil {
		if *min < 0 {
			return fmt.Errorf("时长最小值不能为负数")
		}
	}

	if max != nil {
		if *max < 0 {
			return fmt.Errorf("时长最大值不能为负数")
		}
		if *max > 10000 { // 防止过大的值
			return fmt.Errorf("时长最大值不能超过10000分钟")
		}
	}

	if min != nil && max != nil {
		if *min > *max {
			return fmt.Errorf("时长最小值不能大于最大值")
		}
	}

	return nil
}

// 辅助方法：验证文件夹大小范围
func (v *mediaFilterValidator) validateFolderSizeRange(min, max *int64) error {
	if min != nil {
		if *min < 0 {
			return fmt.Errorf("文件夹大小最小值不能为负数")
		}
	}

	if max != nil {
		if *max < 0 {
			return fmt.Errorf("文件夹大小最大值不能为负数")
		}
	}

	if min != nil && max != nil {
		if *min > *max {
			return fmt.Errorf("文件夹大小最小值不能大于最大值")
		}
	}

	return nil
}

// 辅助方法：验证字段名
func (v *mediaFilterValidator) validateFieldName(field MediaFilterField) error {
	if field == "" {
		return fmt.Errorf("字段名不能为空")
	}

	validFields := v.getValidFilterFields()
	for _, validField := range validFields {
		if validField == field {
			return nil
		}
	}

	return fmt.Errorf("无效的字段名: %s", field)
}

// 辅助方法：验证操作符
func (v *mediaFilterValidator) validateOperator(operator FilterOperator) error {
	validOperators := []FilterOperator{
		FilterOperatorEq, FilterOperatorNe, FilterOperatorGt, FilterOperatorGte,
		FilterOperatorLt, FilterOperatorLte, FilterOperatorLike, FilterOperatorNotLike,
		FilterOperatorIn, FilterOperatorNotIn, FilterOperatorRegex, FilterOperatorNotRegex,
		FilterOperatorBetween, FilterOperatorIsNull, FilterOperatorIsNotNull,
		FilterOperatorStartsWith, FilterOperatorEndsWith,
	}

	for _, validOp := range validOperators {
		if validOp == operator {
			return nil
		}
	}

	return fmt.Errorf("无效的操作符: %s", operator)
}

// 辅助方法：验证字段值
func (v *mediaFilterValidator) validateFieldValue(field MediaFilterField, value interface{}, operator FilterOperator) error {
	if value == nil {
		return fmt.Errorf("字段值不能为空")
	}

	// 根据字段类型验证值
	switch field {
	case MediaFilterFieldID, MediaFilterFieldYear, MediaFilterFieldVotes,
		MediaFilterFieldRuntime, MediaFilterFieldSeasonCount, MediaFilterFieldEpisodeCount,
		MediaFilterFieldFolderSize:
		return v.validateNumericFieldValue(value, operator)

	case MediaFilterFieldRating:
		return v.validateRatingFieldValue(value, operator)

	case MediaFilterFieldAirDate, MediaFilterFieldFirstAirDate, MediaFilterFieldLastAirDate,
		MediaFilterFieldReleaseDate, MediaFilterFieldCreateTime, MediaFilterFieldUpdateTime:
		return v.validateDateFieldValue(value, operator)

	case MediaFilterFieldGenres, MediaFilterFieldTags, MediaFilterFieldCast:
		return v.validateListFieldValue(value, operator)

	default:
		return v.validateStringFieldValue(value, operator)
	}
}

// 辅助方法：验证数字字段值
func (v *mediaFilterValidator) validateNumericFieldValue(value interface{}, operator FilterOperator) error {
	switch val := value.(type) {
	case int, int64, float64:
		// 数字类型直接返回
		return nil
	case string:
		// 尝试将字符串转换为数字
		_, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("无效的数字值: %s", val)
		}
		return nil
	case []interface{}:
		// 处理数组类型（用于in操作符）
		for _, item := range val {
			if err := v.validateNumericFieldValue(item, operator); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("无效的数字值类型: %T", value)
	}
}

// 辅助方法：验证评分字段值
func (v *mediaFilterValidator) validateRatingFieldValue(value interface{}, operator FilterOperator) error {
	switch val := value.(type) {
	case float64:
		if val < 0 || val > 10 {
			return fmt.Errorf("评分值必须在0到10之间: %v", val)
		}
		return nil
	case string:
		rating, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("无效的评分值: %s", val)
		}
		if rating < 0 || rating > 10 {
			return fmt.Errorf("评分值必须在0到10之间: %v", rating)
		}
		return nil
	case []interface{}:
		for _, item := range val {
			if err := v.validateRatingFieldValue(item, operator); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("无效的评分值类型: %T", value)
	}
}

// 辅助方法：验证日期字段值
func (v *mediaFilterValidator) validateDateFieldValue(value interface{}, operator FilterOperator) error {
	switch val := value.(type) {
	case time.Time:
		// 时间类型直接返回
		return nil
	case string:
		// 尝试解析日期字符串
		formats := []string{
			time.RFC3339,
			"2006-01-02",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}

		for _, format := range formats {
			_, err := time.Parse(format, val)
			if err == nil {
				return nil
			}
		}

		return fmt.Errorf("无效的日期格式: %s", val)
	case []interface{}:
		for _, item := range val {
			if err := v.validateDateFieldValue(item, operator); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("无效的日期值类型: %T", value)
	}
}

// 辅助方法：验证字符串字段值
func (v *mediaFilterValidator) validateStringFieldValue(value interface{}, operator FilterOperator) error {
	switch val := value.(type) {
	case string:
		// 对于regex操作符，验证正则表达式
		if operator == FilterOperatorRegex || operator == FilterOperatorNotRegex {
			_, err := regexp.Compile(val)
			if err != nil {
				return fmt.Errorf("无效的正则表达式: %s, %w", val, err)
			}
		}
		return nil
	case []interface{}:
		for _, item := range val {
			if err := v.validateStringFieldValue(item, operator); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("无效的字符串值类型: %T", value)
	}
}

// 辅助方法：验证列表字段值
func (v *mediaFilterValidator) validateListFieldValue(value interface{}, operator FilterOperator) error {
	switch val := value.(type) {
	case []interface{}:
		for _, item := range val {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("列表项必须是字符串类型: %T", item)
			}
		}
		return nil
	case []string:
		return nil
	case string:
		// 对于单个字符串值，允许以逗号分隔的列表
		return nil
	default:
		return fmt.Errorf("无效的列表值类型: %T", value)
	}
}

// 辅助方法：获取有效的过滤字段
func (v *mediaFilterValidator) getValidFilterFields() []MediaFilterField {
	return []MediaFilterField{
		MediaFilterFieldID, MediaFilterFieldTitle, MediaFilterFieldOriginalTitle,
		MediaFilterFieldType, MediaFilterFieldStatus, MediaFilterFieldYear,
		MediaFilterFieldRating, MediaFilterFieldVotes, MediaFilterFieldRuntime,
		MediaFilterFieldSeasonCount, MediaFilterFieldEpisodeCount,
		MediaFilterFieldAirDate, MediaFilterFieldFirstAirDate,
		MediaFilterFieldLastAirDate, MediaFilterFieldReleaseDate,
		MediaFilterFieldOverview, MediaFilterFieldGenres, MediaFilterFieldTags,
		MediaFilterFieldStudio, MediaFilterFieldDirector, MediaFilterFieldCast,
		MediaFilterFieldWriter, MediaFilterFieldIMDBID, MediaFilterFieldTMDBID,
		MediaFilterFieldTVDBID, MediaFilterFieldSource, MediaFilterFieldCover,
		MediaFilterFieldBackdrop, MediaFilterFieldTrailer, MediaFilterFieldLogo,
		MediaFilterFieldLocalStatus, MediaFilterFieldSubscribeStatus,
		MediaFilterFieldDownloadStatus, MediaFilterFieldCreateTime,
		MediaFilterFieldUpdateTime, MediaFilterFieldSortTitle,
		MediaFilterFieldLanguage, MediaFilterFieldCountry,
		MediaFilterFieldNetwork, MediaFilterFieldCollection,
		MediaFilterFieldQuality, MediaFilterFieldCodec,
		MediaFilterFieldResolution, MediaFilterFieldAudio,
		MediaFilterFieldVideoFormat, MediaFilterFieldFolderSize,
		MediaFilterFieldFilePath, MediaFilterFieldFolderPath,
		MediaFilterFieldMediaServer, MediaFilterFieldSubtitleStatus,
		MediaFilterFieldCustom1, MediaFilterFieldCustom2, MediaFilterFieldCustom3,
	}
}

// 辅助方法：获取错误消息
func (v *mediaFilterValidator) getErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s 字段是必填项", field)
	case "min":
		return fmt.Sprintf("%s 字段值必须大于或等于 %s", field, param)
	case "max":
		return fmt.Sprintf("%s 字段值必须小于或等于 %s", field, param)
	case "media_type":
		return fmt.Sprintf("%s 字段包含无效的媒体类型", field)
	case "media_status":
		return fmt.Sprintf("%s 字段包含无效的媒体状态", field)
	case "filter_field":
		return fmt.Sprintf("%s 字段包含无效的过滤字段名", field)
	case "filter_operator":
		return fmt.Sprintf("%s 字段包含无效的过滤操作符", field)
	case "sort_field":
		return fmt.Sprintf("%s 字段包含无效的排序字段名", field)
	case "sort_order":
		return fmt.Sprintf("%s 字段必须是 'asc' 或 'desc'", field)
	case "valid_year":
		return fmt.Sprintf("%s 字段包含无效的年份值", field)
	case "valid_rating":
		return fmt.Sprintf("%s 字段包含无效的评分值", field)
	case "filter_group":
		return fmt.Sprintf("%s 字段包含无效的过滤条件组", field)
	default:
		return fmt.Sprintf("%s 字段验证失败: %s", field, tag)
	}
}

// 自定义验证器函数

// validateMediaTypeFunc 媒体类型验证器
func (v *mediaFilterValidator) validateMediaTypeFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch val := field.(type) {
	case MediaType:
		return isValidMediaType(val)
	case string:
		return isValidMediaType(MediaType(val))
	case []MediaType:
		for _, mt := range val {
			if !isValidMediaType(mt) {
				return false
			}
		}
		return true
	case []string:
		for _, mt := range val {
			if !isValidMediaType(MediaType(mt)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// validateMediaStatusFunc 媒体状态验证器
func (v *mediaFilterValidator) validateMediaStatusFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch val := field.(type) {
	case MediaStatus:
		return isValidMediaStatus(val)
	case string:
		return isValidMediaStatus(MediaStatus(val))
	case []MediaStatus:
		for _, ms := range val {
			if !isValidMediaStatus(ms) {
				return false
			}
		}
		return true
	case []string:
		for _, ms := range val {
			if !isValidMediaStatus(MediaStatus(ms)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// validateFilterFieldFunc 过滤字段验证器
func (v *mediaFilterValidator) validateFilterFieldFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch val := field.(type) {
	case MediaFilterField:
		return v.isValidFilterField(val)
	case string:
		return v.isValidFilterField(MediaFilterField(val))
	case []MediaFilterField:
		for _, ff := range val {
			if !v.isValidFilterField(ff) {
				return false
			}
		}
		return true
	case []string:
		for _, ff := range val {
			if !v.isValidFilterField(MediaFilterField(ff)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// validateFilterOperatorFunc 过滤操作符验证器
func (v *mediaFilterValidator) validateFilterOperatorFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch val := field.(type) {
	case FilterOperator:
		return v.isValidFilterOperator(val)
	case string:
		return v.isValidFilterOperator(FilterOperator(val))
	default:
		return false
	}
}

// validateSortFieldFunc 排序字段验证器
func (v *mediaFilterValidator) validateSortFieldFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch val := field.(type) {
	case MediaSortField:
		return v.isValidSortField(val)
	case string:
		return v.isValidSortField(MediaSortField(val))
	default:
		return false
	}
}

// validateSortOrderFunc 排序顺序验证器
func (v *mediaFilterValidator) validateSortOrderFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch val := field.(type) {
	case SortOrder:
		return val == SortOrderAsc || val == SortOrderDesc || val == ""
	case string:
		return val == string(SortOrderAsc) || val == string(SortOrderDesc) || val == ""
	default:
		return false
	}
}

// validateYearFunc 年份验证器
func (v *mediaFilterValidator) validateYearFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()
	currentYear := time.Now().Year()

	switch val := field.(type) {
	case int:
		return val >= 1800 && val <= currentYear+1
	case int64:
		return val >= 1800 && val <= int64(currentYear+1)
	case string:
		year, err := strconv.Atoi(val)
		if err != nil {
			return false
		}
		return year >= 1800 && year <= currentYear+1
	default:
		return false
	}
}

// validateRatingFunc 评分验证器
func (v *mediaFilterValidator) validateRatingFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch val := field.(type) {
	case float64:
		return val >= 0 && val <= 10
	case string:
		rating, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return false
		}
		return rating >= 0 && rating <= 10
	default:
		return false
	}
}

// validateFilterGroupFunc 过滤条件组验证器
func (v *mediaFilterValidator) validateFilterGroupFunc(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	group, ok := field.(*FilterGroup)
	if !ok {
		return false
	}

	if group == nil {
		return false
	}

	if group.Logic != "and" && group.Logic != "or" {
		return false
	}

	return len(group.Conditions) > 0
}

// 辅助函数：验证枚举值
func isValidMediaType(mediaType MediaType) bool {
	validTypes := []MediaType{MediaTypeMovie, MediaTypeTV, MediaTypeTVShow, MediaTypeSeries, MediaTypeAnime}
	for _, t := range validTypes {
		if t == mediaType {
			return true
		}
	}
	return false
}

func isValidMediaStatus(status MediaStatus) bool {
	validStatuses := []MediaStatus{MediaStatusReleased, MediaStatusPlanned, MediaStatusInProduction, MediaStatusCanceled, MediaStatusEnded, MediaStatusReturningSeries}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func isValidLocalMediaStatus(status LocalMediaStatus) bool {
	validStatuses := []LocalMediaStatus{LocalMediaStatusNotFound, LocalMediaStatusPartial, LocalMediaStatusFull}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func isValidSubscribeStatus(status SubscribeStatus) bool {
	validStatuses := []SubscribeStatus{SubscribeStatusNone, SubscribeStatusNormal, SubscribeStatusIgnore, SubscribeStatusSkip}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func isValidDownloadStatus(status DownloadStatus) bool {
	validStatuses := []DownloadStatus{DownloadStatusNone, DownloadStatusPending, DownloadStatusDownloading, DownloadStatusCompleted, DownloadStatusFailed, DownloadStatusPaused}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func isValidSubtitleStatus(status SubtitleStatus) bool {
	validStatuses := []SubtitleStatus{SubtitleStatusNone, SubtitleStatusPartial, SubtitleStatusFull}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func (v *mediaFilterValidator) isValidFilterField(field MediaFilterField) bool {
	validFields := v.getValidFilterFields()
	for _, f := range validFields {
		if f == field {
			return true
		}
	}
	return false
}

func (v *mediaFilterValidator) isValidFilterOperator(operator FilterOperator) bool {
	validOperators := []FilterOperator{
		FilterOperatorEq, FilterOperatorNe, FilterOperatorGt, FilterOperatorGte,
		FilterOperatorLt, FilterOperatorLte, FilterOperatorLike, FilterOperatorNotLike,
		FilterOperatorIn, FilterOperatorNotIn, FilterOperatorRegex, FilterOperatorNotRegex,
		FilterOperatorBetween, FilterOperatorIsNull, FilterOperatorIsNotNull,
		FilterOperatorStartsWith, FilterOperatorEndsWith,
	}

	for _, op := range validOperators {
		if op == operator {
			return true
		}
	}
	return false
}

func (v *mediaFilterValidator) isValidSortField(field MediaSortField) bool {
	validFields := map[MediaSortField]bool{
		MediaSortFieldID:              true,
		MediaSortFieldTitle:           true,
		MediaSortFieldOriginalTitle:   true,
		MediaSortFieldType:            true,
		MediaSortFieldYear:            true,
		MediaSortFieldRating:          true,
		MediaSortFieldVotes:           true,
		MediaSortFieldRuntime:         true,
		MediaSortFieldSeasonCount:     true,
		MediaSortFieldEpisodeCount:    true,
		MediaSortFieldAirDate:         true,
		MediaSortFieldFirstAirDate:    true,
		MediaSortFieldLastAirDate:     true,
		MediaSortFieldReleaseDate:     true,
		MediaSortFieldCreateTime:      true,
		MediaSortFieldUpdateTime:      true,
		MediaSortFieldSortTitle:       true,
		MediaSortFieldLocalStatus:     true,
		MediaSortFieldSubscribeStatus: true,
		MediaSortFieldDownloadStatus:  true,
		MediaSortFieldFolderSize:      true,
		MediaSortFieldQuality:         true,
		MediaSortFieldResolution:      true,
		"":                           true, // 允许空值
	}

	return validFields[field]
}
