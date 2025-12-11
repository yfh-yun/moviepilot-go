package params

import (
	"github.com/mitchellh/mapstructure"
)

// BaseParams 定义参数的基础接口
type BaseParams interface {
	// Validate 验证参数是否有效
	Validate() error

	// FromMap 从map转换为参数对象
	FromMap(data map[string]any) error
}

// BaseParamsStruct 定义参数的基础结构体，所有参数都要嵌入这个结构体
type BaseParamsStruct struct {
	// ActionName 动作名称
	ActionName string `json:"action_name" mapstructure:"action_name"`

	// ActionID 动作ID
	ActionID string `json:"action_id" mapstructure:"action_id"`

	// WorkflowID 工作流ID
	WorkflowID string `json:"workflow_id" mapstructure:"workflow_id"`

	// Description 动作描述
	Description string `json:"description" mapstructure:"description"`

	// OutputKey 输出键，用于将动作结果存储到全局上下文中
	OutputKey string `json:"output_key" mapstructure:"output_key"`

	// IsAsync 是否异步执行
	IsAsync bool `json:"is_async" mapstructure:"is_async"`
}

// Validate 验证基础参数
func (p *BaseParamsStruct) Validate() error {
	// 基础参数验证为空，可以在子类中扩展
	return nil
}

// FromMap 从map转换为参数对象
func (p *BaseParamsStruct) FromMap(data map[string]any) error {
	// 使用mapstructure进行转换
	if err := mapstructure.Decode(data, p); err != nil {
		return err
	}
	return nil
}

// NewBaseParams 创建新的基础参数实例
func NewBaseParams() *BaseParamsStruct {
	return &BaseParamsStruct{}
}

// Error 定义参数验证错误
var (
	// 文件处理动作错误
	ErrScanPathEmpty        = NewParamError("scan_path_empty", "Scan path cannot be empty")
	ErrSourcePathEmpty      = NewParamError("source_path_empty", "Source path cannot be empty")
	ErrDestinationPathEmpty = NewParamError("destination_path_empty", "Destination path cannot be empty")
	ErrFileNotFound         = NewParamError("file_not_found", "File not found")
	ErrDirectoryNotFound    = NewParamError("directory_not_found", "Directory not found")
	ErrInvalidFilePath      = NewParamError("invalid_file_path", "Invalid file path")
	ErrInvalidOperation     = NewParamError("invalid_operation", "Invalid operation")

	// 资源获取动作错误
	ErrTorrentClientEmpty  = NewParamError("torrent_client_empty", "Torrent client cannot be empty")
	ErrRSSURLsEmpty        = NewParamError("rss_urls_empty", "RSS URLs cannot be empty")
	ErrMediaServerEmpty    = NewParamError("media_server_empty", "Media server cannot be empty")
	ErrDownloadClientEmpty = NewParamError("download_client_empty", "Download client cannot be empty")

	// 过滤动作错误
	ErrFilterRulesEmpty  = NewParamError("filter_rules_empty", "Filter rules cannot be empty")
	ErrFilterTypeInvalid = NewParamError("filter_type_invalid", "Invalid filter type")

	// 核心业务动作错误
	ErrDownloadURLsEmpty   = NewParamError("download_urls_empty", "Download URLs cannot be empty")
	ErrSubscribeTitleEmpty = NewParamError("subscribe_title_empty", "Subscribe title cannot be empty")
	ErrMessageContentEmpty = NewParamError("message_content_empty", "Message content cannot be empty")
	ErrMessageTypeInvalid  = NewParamError("message_type_invalid", "Invalid message type")

	// 系统功能动作错误
	ErrPluginIDEmpty    = NewParamError("plugin_id_empty", "Plugin ID cannot be empty")
	ErrEventNameEmpty   = NewParamError("event_name_empty", "Event name cannot be empty")
	ErrNoteContentEmpty = NewParamError("note_content_empty", "Note content cannot be empty")
)

// ParamError 定义参数验证错误结构体
type ParamError struct {
	Code    string // 错误代码
	Message string // 错误信息
}

// Error 实现error接口
func (e *ParamError) Error() string {
	return e.Message
}

// GetCode 获取错误代码
func (e *ParamError) GetCode() string {
	return e.Code
}

// NewParamError 创建新的参数验证错误
func NewParamError(code, message string) *ParamError {
	return &ParamError{
		Code:    code,
		Message: message,
	}
}
