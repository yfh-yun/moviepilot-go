// Package implementations 提供具体动作实现
package implementations

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/workflows/base"
	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/types"

	"go.uber.org/zap"
)

// DownloadAction 下载动作
// 负责处理下载任务的创建和管理
type DownloadAction struct {
	*base.BaseAction

	// 下载服务
	downloadService interfaces.DownloadService

	// 配置
	config *DownloadConfig

	// 验证器
	validator *DownloadValidator
}

// DownloadConfig 下载动作配置
type DownloadConfig struct {
	// 下载器配置
	Downloader string `json:"downloader"` // qbittorrent, transmission, etc.
	SavePath   string `json:"save_path"`
	Category   string `json:"category"`
	Label      string `json:"label"`

	// 种子配置
	Paused      bool     `json:"paused"`
	AutoStart   bool     `json:"auto_start"`
	Priority    int      `json:"priority"`    // 1-10
	Sequential  bool     `json:"sequential"`  // 顺序下载

	// 限制配置
	MaxSpeed    int64    `json:"max_speed"`    // 最大下载速度（字节/秒）
	MaxConnections int    `json:"max_connections"` // 最大连接数

	// 高级配置
	SkipHashCheck bool     `json:"skip_hash_check"`
	SkipFileCheck bool     `json:"skip_file_check"`
	UploadLimit   int64    `json:"upload_limit"`   // 上传限制（字节/秒）
	RatioLimit    float64  `json:"ratio_limit"`    // 分享率限制
	SeedTime      int64    `json:"seed_time"`      // 做种时间（秒）

	// 过滤配置
	MinSize       int64    `json:"min_size"`       // 最小文件大小（字节）
	MaxSize       int64    `json:"max_size"`       // 最大文件大小（字节）
	IncludeRegex  string   `json:"include_regex"`  // 包含正则表达式
	ExcludeRegex  string   `json:"exclude_regex"`  // 排除正则表达式
	AllowedTypes  []string `json:"allowed_types"`  // 允许的文件类型
	BlockedTypes  []string `json:"blocked_types"`  // 阻止的文件类型

	// 通知配置
	NotifyOnComplete bool `json:"notify_on_complete"`
	NotifyOnError    bool `json:"notify_on_error"`

	// 重试配置
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	RetryBackoff    string        `json:"retry_backoff"` // fixed, linear, exponential
}

// DownloadValidator 下载动作验证器
type DownloadValidator struct {
	// 验证规则
	rules map[string]ValidationRule
}

// ValidationRule 验证规则
type ValidationRule struct {
	Required bool
	Type     string
	Pattern  string
	Min      interface{}
	Max      interface{}
}

// NewDownloadAction 创建下载动作
func NewDownloadAction(actionID string, downloadService interfaces.DownloadService, cache interfaces.Cache) *DownloadAction {
	baseAction := base.NewBaseAction(actionID, cache)
	
	config := &DownloadConfig{
		Downloader:     "qbittorrent",
		AutoStart:      true,
		Priority:       5,
		MaxRetries:     3,
		RetryDelay:     5 * time.Second,
		RetryBackoff:   "exponential",
		NotifyOnComplete: true,
		NotifyOnError:    true,
	}

	validator := &DownloadValidator{
		rules: map[string]ValidationRule{
			"url":         {Required: true, Type: "string"},
			"title":       {Required: true, Type: "string"},
			"save_path":   {Required: true, Type: "string"},
			"downloader":  {Required: false, Type: "string", Pattern: "^(qbittorrent|transmission)$"},
			"priority":    {Required: false, Type: "int", Min: 1, Max: 10},
			"max_speed":   {Required: false, Type: "int", Min: 0},
			"max_size":    {Required: false, Type: "int", Min: 0},
			"min_size":    {Required: false, Type: "int", Min: 0},
		},
	}

	action := &DownloadAction{
		BaseAction:     baseAction,
		downloadService: downloadService,
		config:         config,
		validator:      validator,
	}

	// 设置配置和验证器
	baseAction.SetConfig(configToMap(config))
	baseAction.SetValidator(validator)

	return action
}

// Name 获取动作名称
func (da *DownloadAction) Name() string {
	return "DownloadAction"
}

// Description 获取动作描述
func (da *DownloadAction) Description() string {
	return "创建和管理下载任务"
}

// Data 获取动作数据
func (da *DownloadAction) Data() map[string]interface{} {
	data := da.BaseAction.Data()
	data["config"] = da.config
	data["download_service"] = da.downloadService != nil
	return data
}

// doExecute 执行下载动作
func (da *DownloadAction) doExecute(ctx context.Context, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error) {
	da.logger.Info("开始执行下载动作",
		zap.Int64("workflow_id", workflowID),
		zap.Any("params", params))

	// 提取参数
	url, ok := params["url"].(string)
	if !ok {
		return context, fmt.Errorf("缺少必需参数: url")
	}

	title, ok := params["title"].(string)
	if !ok {
		title = "Unknown"
	}

	// 构建下载参数
	downloadParams := interfaces.CreateDownloadParams{
		Title:    title,
		Type:     "torrent",
		URL:      url,
		SavePath: da.config.SavePath,
	}

	// 应用配置
	if da.config.Category != "" {
		downloadParams.Category = da.config.Category
	}

	// 创建下载任务
	task, err := da.downloadService.CreateDownload(ctx, downloadParams)
	if err != nil {
		da.SetError(fmt.Sprintf("创建下载任务失败: %s", err.Error()))
		return context, err
	}

	// 添加到上下文
	context.AddDownload(&types.Download{
		ID:         task.ID,
		Title:      task.Title,
		URL:        task.URL,
		Hash:       "", // 需要从下载器获取
		Type:       task.Type,
		Status:     task.Status,
		Progress:   task.Progress,
		FileSize:   task.FileSize,
		Downloaded: task.Downloaded,
		Speed:      task.Speed,
		ETA:        task.ETA,
		Downloader: da.config.Downloader,
		SavePath:   da.config.SavePath,
		Category:   da.config.Category,
		Label:      da.config.Label,
		Priority:   da.config.Priority,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
	})

	// 添加日志
	context.AddLog("info", fmt.Sprintf("下载任务创建成功: %s", task.ID), "DownloadAction", map[string]interface{}{
		"task_id": task.ID,
		"title":   task.Title,
		"url":     task.URL,
	})

	// 设置完成状态
	da.SetDone(fmt.Sprintf("下载任务创建成功: %s", task.ID))

	// 更新上下文进度
	context.UpdateProgress(100, "下载动作完成")

	da.logger.Info("下载动作执行完成",
		zap.String("task_id", task.ID),
		zap.String("title", task.Title))

	return context, nil
}

// ValidateConfig 验证配置
func (dv *DownloadValidator) ValidateConfig(config map[string]interface{}) error {
	// 验证下载器类型
	if downloader, ok := config["downloader"].(string); ok {
		if downloader != "qbittorrent" && downloader != "transmission" {
			return fmt.Errorf("不支持的下载器类型: %s", downloader)
		}
	}

	// 验证优先级
	if priority, ok := config["priority"].(float64); ok {
		if priority < 1 || priority > 10 {
			return fmt.Errorf("优先级必须在1-10之间: %f", priority)
		}
	}

	// 验证速度限制
	if maxSpeed, ok := config["max_speed"].(float64); ok {
		if maxSpeed < 0 {
			return fmt.Errorf("最大速度不能为负数: %f", maxSpeed)
		}
	}

	return nil
}

// ValidateParams 验证参数
func (dv *DownloadValidator) ValidateParams(params types.ActionParams) error {
	for paramName, rule := range dv.rules {
		value, exists := params[paramName]
		if !exists {
			if rule.Required {
				return fmt.Errorf("缺少必需参数: %s", paramName)
			}
			continue
		}

		// 类型验证
		switch rule.Type {
		case "string":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("参数 %s 必须是字符串", paramName)
			}
		case "int":
			if _, ok := value.(int); !ok {
				if _, ok := value.(float64); !ok {
					return fmt.Errorf("参数 %s 必须是整数", paramName)
				}
			}
		case "bool":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("参数 %s 必须是布尔值", paramName)
			}
		}

		// 范围验证
		if rule.Min != nil || rule.Max != nil {
			var numValue float64
			switch v := value.(type) {
			case int:
				numValue = float64(v)
			case float64:
				numValue = v
			default:
				continue
			}

			if rule.Min != nil && numValue < toFloat64(rule.Min) {
				return fmt.Errorf("参数 %s 不能小于 %v", paramName, rule.Min)
			}
			if rule.Max != nil && numValue > toFloat64(rule.Max) {
				return fmt.Errorf("参数 %s 不能大于 %v", paramName, rule.Max)
			}
		}

		// 正则验证
		if rule.Pattern != "" {
			if strValue, ok := value.(string); ok {
				// 这里应该使用正则表达式验证，简化实现
				if rule.Pattern == "^(qbittorrent|transmission)$" {
					if strValue != "qbittorrent" && strValue != "transmission" {
						return fmt.Errorf("参数 %s 必须是 qbittorrent 或 transmission", paramName)
					}
				}
			}
		}
	}

	return nil
}

// ValidateContext 验证上下文
func (dv *DownloadValidator) ValidateContext(context *types.ActionContext) error {
	if context == nil {
		return fmt.Errorf("上下文不能为空")
	}

	// 验证工作流ID
	if context.WorkflowID <= 0 {
		return fmt.Errorf("无效的工作流ID: %d", context.WorkflowID)
	}

	return nil
}

// 辅助函数
func configToMap(config *DownloadConfig) map[string]interface{} {
	return map[string]interface{}{
		"downloader":        config.Downloader,
		"save_path":         config.SavePath,
		"category":          config.Category,
		"label":             config.Label,
		"paused":            config.Paused,
		"auto_start":        config.AutoStart,
		"priority":          config.Priority,
		"sequential":        config.Sequential,
		"max_speed":         config.MaxSpeed,
		"max_connections":   config.MaxConnections,
		"skip_hash_check":   config.SkipHashCheck,
		"skip_file_check":   config.SkipFileCheck,
		"upload_limit":      config.UploadLimit,
		"ratio_limit":       config.RatioLimit,
		"seed_time":         config.SeedTime,
		"min_size":          config.MinSize,
		"max_size":          config.MaxSize,
		"include_regex":     config.IncludeRegex,
		"exclude_regex":     config.ExcludeRegex,
		"allowed_types":     config.AllowedTypes,
		"blocked_types":     config.BlockedTypes,
		"notify_on_complete": config.NotifyOnComplete,
		"notify_on_error":    config.NotifyOnError,
		"max_retries":        config.MaxRetries,
		"retry_delay":        config.RetryDelay,
		"retry_backoff":      config.RetryBackoff,
	}
}

func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}

// 确保实现接口
var _ interfaces.Action = (*DownloadAction)(nil)
var _ interfaces.ActionValidator = (*DownloadValidator)(nil)