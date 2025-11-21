package actions

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// MediaSyncValidator 媒体同步验证器接口
type MediaSyncValidator interface {
	// ValidateSyncRequest 验证同步请求
	ValidateSyncRequest(request MediaSyncRequest) error
	// ValidateBatchSyncRequest 验证批量同步请求
	ValidateBatchSyncRequest(request MediaSyncRequest) error
	// ValidateTaskID 验证任务ID
	ValidateTaskID(taskID string) error
	// ValidateTaskQuery 验证任务查询参数
	ValidateTaskQuery(query MediaSyncTaskQuery) error
	// ValidateConflictResolution 验证冲突解决方案
	ValidateConflictResolution(resolution SyncConflictResolution) error
	// RegisterValidators 注册自定义验证器
	RegisterValidators() error
}

// mediaSyncValidator 媒体同步验证器实现
type mediaSyncValidator struct {}

// NewMediaSyncValidator 创建媒体同步验证器实例
func NewMediaSyncValidator() MediaSyncValidator {
	return &mediaSyncValidator{}
}

// ValidateSyncRequest 验证同步请求
func (v *mediaSyncValidator) ValidateSyncRequest(request MediaSyncRequest) error {
	// 验证同步源
	if err := v.validateSyncSource(request.Source); err != nil {
		return err
	}
	
	// 验证同步策略
	if err := v.validateSyncStrategy(request.Strategy); err != nil {
		return err
	}
	
	// 验证同步字段
	if err := v.validateSyncFields(request.SyncFields); err != nil {
		return err
	}
	
	// 验证并发数
	if request.Concurrency < 0 {
		return fmt.Errorf("并发数不能为负数")
	}
	if request.Concurrency > 100 {
		return fmt.Errorf("并发数不能超过100")
	}
	
	// 验证超时时间
	if request.Timeout < 0 {
		return fmt.Errorf("超时时间不能为负数")
	}
	if request.Timeout > 86400 { // 24小时
		return fmt.Errorf("超时时间不能超过24小时")
	}
	
	// 验证最大重试次数
	if request.MaxRetries < 0 {
		return fmt.Errorf("最大重试次数不能为负数")
	}
	if request.MaxRetries > 10 {
		return fmt.Errorf("最大重试次数不能超过10")
	}
	
	// 验证媒体ID列表（单个同步时必须提供）
	if len(request.MediaIDs) == 0 && request.Source != SyncSourceLocal && request.Source != SyncSourceRemote {
		return fmt.Errorf("单个同步时必须提供媒体ID列表")
	}
	
	// 根据不同的同步源验证特定参数
	switch request.Source {
	case SyncSourceLocal:
		if request.SourcePath == "" {
			return fmt.Errorf("本地同步时必须提供源路径")
		}
	case SyncSourceRemote:
		if request.RemoteServerID == "" {
			return fmt.Errorf("远程同步时必须提供服务器ID")
		}
	}
	
	return nil
}

// ValidateBatchSyncRequest 验证批量同步请求
func (v *mediaSyncValidator) ValidateBatchSyncRequest(request MediaSyncRequest) error {
	// 先调用基本同步请求验证
	if err := v.ValidateSyncRequest(request); err != nil {
		return err
	}
	
	// 批量同步特有验证
	if request.Concurrency <= 0 {
		// 批量同步默认并发数设置为5
		request.Concurrency = 5
	}
	
	// 验证过滤条件
	if len(request.Filter) > 0 {
		if err := v.validateFilter(request.Filter); err != nil {
			return err
		}
	}
	
	return nil
}

// ValidateTaskID 验证任务ID
func (v *mediaSyncValidator) ValidateTaskID(taskID string) error {
	if taskID == "" {
		return fmt.Errorf("任务ID不能为空")
	}
	
	// 验证任务ID格式
	regex := regexp.MustCompile(`^sync_[a-z0-9_]+_\d+$`)
	if !regex.MatchString(taskID) {
		return fmt.Errorf("任务ID格式无效，应为'sync_<source>_<timestamp>'格式")
	}
	
	return nil
}

// ValidateTaskQuery 验证任务查询参数
func (v *mediaSyncValidator) ValidateTaskQuery(query MediaSyncTaskQuery) error {
	// 验证页码
	if query.Page < 0 {
		return fmt.Errorf("页码不能为负数")
	}
	if query.Page > 1000 {
		return fmt.Errorf("页码不能超过1000")
	}
	
	// 验证每页数量
	if query.PageSize < 0 {
		return fmt.Errorf("每页数量不能为负数")
	}
	if query.PageSize > 1000 {
		return fmt.Errorf("每页数量不能超过1000")
	}
	
	// 验证任务状态
	if query.Status != "" {
		validStatuses := map[string]bool{
			string(SyncStatusPending):           true,
			string(SyncStatusInProgress):        true,
			string(SyncStatusCompleted):         true,
			string(SyncStatusFailed):            true,
			string(SyncStatusCancelled):         true,
			string(SyncStatusPartiallyCompleted): true,
		}
		if !validStatuses[query.Status] {
			return fmt.Errorf("无效的任务状态: %s", query.Status)
		}
	}
	
	// 验证同步源
	if query.Source != "" {
		source := SyncSource(query.Source)
		if err := v.validateSyncSource(source); err != nil {
			return err
		}
	}
	
	// 验证时间格式
	if query.StartTime != "" {
		if _, err := time.Parse(time.RFC3339, query.StartTime); err != nil {
			return fmt.Errorf("开始时间格式无效，应为RFC3339格式")
		}
	}
	
	if query.EndTime != "" {
		if _, err := time.Parse(time.RFC3339, query.EndTime); err != nil {
			return fmt.Errorf("结束时间格式无效，应为RFC3339格式")
		}
	}
	
	// 验证搜索关键词长度
	if len(query.Keyword) > 100 {
		return fmt.Errorf("搜索关键词长度不能超过100个字符")
	}
	
	return nil
}

// ValidateConflictResolution 验证冲突解决方案
func (v *mediaSyncValidator) ValidateConflictResolution(resolution SyncConflictResolution) error {
	// 验证冲突ID
	if resolution.ConflictID == "" {
		return fmt.Errorf("冲突ID不能为空")
	}
	
	// 验证冲突ID格式
	regex := regexp.MustCompile(`^conflict_[a-z0-9_]+_\d+$`)
	if !regex.MatchString(resolution.ConflictID) {
		return fmt.Errorf("冲突ID格式无效，应为'conflict_<media_id>_<timestamp>'格式")
	}
	
	// 验证解决方案类型
	validResolutions := map[string]bool{
		"keep_local":  true,
		"keep_remote": true,
		"merge":       true,
	}
	
	if !validResolutions[resolution.Resolution] {
		return fmt.Errorf("无效的解决方案类型: %s，支持的类型有: keep_local, keep_remote, merge", resolution.Resolution)
	}
	
	// 验证字段映射（仅在merge模式下需要）
	if resolution.Resolution == "merge" {
		if err := v.validateFieldMapping(resolution.FieldMapping); err != nil {
			return err
		}
	}
	
	return nil
}

// RegisterValidators 注册自定义验证器到Gin框架
func (v *mediaSyncValidator) RegisterValidators() error {
	// 获取gin的验证器实例
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册同步源验证器
		if err := v.RegisterValidation("sync_source", v.validateSyncSourceTag); err != nil {
			return fmt.Errorf("注册同步源验证器失败: %w", err)
		}
		
		// 注册同步策略验证器
		if err := v.RegisterValidation("sync_strategy", v.validateSyncStrategyTag); err != nil {
			return fmt.Errorf("注册同步策略验证器失败: %w", err)
		}
		
		// 注册同步字段验证器
		if err := v.RegisterValidation("sync_field", v.validateSyncFieldTag); err != nil {
			return fmt.Errorf("注册同步字段验证器失败: %w", err)
		}
		
		// 注册任务ID验证器
		if err := v.RegisterValidation("task_id", v.validateTaskIDTag); err != nil {
			return fmt.Errorf("注册任务ID验证器失败: %w", err)
		}
	}
	
	return nil
}

// 私有辅助方法

// validateSyncSource 验证同步源
func (v *mediaSyncValidator) validateSyncSource(source SyncSource) error {
	validSources := map[SyncSource]bool{
		SyncSourceLocal:    true,
		SyncSourceRemote:   true,
		SyncSourceDatabase: true,
		SyncSourceAPI:      true,
	}
	
	if !validSources[source] {
		return fmt.Errorf("无效的同步源: %s", source)
	}
	
	return nil
}

// validateSyncStrategy 验证同步策略
func (v *mediaSyncValidator) validateSyncStrategy(strategy SyncStrategy) error {
	validStrategies := map[SyncStrategy]bool{
		SyncStrategyFull:        true,
		SyncStrategyIncremental: true,
		SyncStrategyOnlyNew:     true,
		SyncStrategyOnlyUpdate:  true,
	}
	
	if !validStrategies[strategy] {
		return fmt.Errorf("无效的同步策略: %s", strategy)
	}
	
	return nil
}

// validateSyncFields 验证同步字段
func (v *mediaSyncValidator) validateSyncFields(fields []SyncField) error {
	if len(fields) == 0 {
		return nil // 允许空字段列表，将使用默认值
	}
	
	validFields := map[SyncField]bool{
		SyncFieldAll:     true,
		SyncFieldBasic:   true,
		SyncFieldMetadata: true,
		SyncFieldFiles:   true,
		SyncFieldStatus:  true,
		SyncFieldTags:    true,
	}
	
	for _, field := range fields {
		if !validFields[field] {
			return fmt.Errorf("无效的同步字段: %s", field)
		}
	}
	
	// 如果包含SyncFieldAll，不应该包含其他字段
	containsAll := false
	for _, field := range fields {
		if field == SyncFieldAll {
			containsAll = true
			break
		}
	}
	
	if containsAll && len(fields) > 1 {
		return fmt.Errorf("当包含'all'字段时，不应该包含其他字段")
	}
	
	return nil
}

// validateFilter 验证过滤条件
func (v *mediaSyncValidator) validateFilter(filter map[string]interface{}) error {
	if len(filter) > 20 {
		return fmt.Errorf("过滤条件数量不能超过20个")
	}
	
	// 验证过滤条件的键值对
	for key, value := range filter {
		// 验证键名
		if len(key) > 50 {
			return fmt.Errorf("过滤条件键名长度不能超过50个字符")
		}
		
		// 验证值类型
		switch value.(type) {
		case string, int, int64, float64, bool, []string, []int:
			// 允许的类型
		default:
			return fmt.Errorf("过滤条件值类型不支持: %T", value)
		}
	}
	
	return nil
}

// validateFieldMapping 验证字段映射
func (v *mediaSyncValidator) validateFieldMapping(fieldMapping map[string]string) error {
	if len(fieldMapping) > 50 {
		return fmt.Errorf("字段映射数量不能超过50个")
	}
	
	// 验证字段映射的键值对
	for key, value := range fieldMapping {
		if len(key) > 50 || len(value) > 50 {
			return fmt.Errorf("字段映射的键名或值长度不能超过50个字符")
		}
	}
	
	return nil
}

// 标签验证器函数（用于Gin验证器）

// validateSyncSourceTag 验证同步源标签
func (v *mediaSyncValidator) validateSyncSourceTag(fl validator.FieldLevel) bool {
	source := SyncSource(fl.Field().String())
	return v.validateSyncSource(source) == nil
}

// validateSyncStrategyTag 验证同步策略标签
func (v *mediaSyncValidator) validateSyncStrategyTag(fl validator.FieldLevel) bool {
	strategy := SyncStrategy(fl.Field().String())
	return v.validateSyncStrategy(strategy) == nil
}

// validateSyncFieldTag 验证同步字段标签
func (v *mediaSyncValidator) validateSyncFieldTag(fl validator.FieldLevel) bool {
	// 处理字符串字段
	if fl.Field().Kind() == validator.StringKind {
		field := SyncField(fl.Field().String())
		return v.validateSyncFields([]SyncField{field}) == nil
	}
	
	// 处理字符串切片字段
	if fl.Field().Kind() == validator.SliceKind {
		fields := make([]SyncField, 0)
		for i := 0; i < fl.Field().Len(); i++ {
			field := SyncField(fl.Field().Index(i).String())
			fields = append(fields, field)
		}
		return v.validateSyncFields(fields) == nil
	}
	
	return false
}

// validateTaskIDTag 验证任务ID标签
func (v *mediaSyncValidator) validateTaskIDTag(fl validator.FieldLevel) bool {
	taskID := fl.Field().String()
	return v.ValidateTaskID(taskID) == nil
}
