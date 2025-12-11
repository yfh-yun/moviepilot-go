package actions

import (
	"errors"

	"moviepilot-go/internal/business/workflows/actions/params"
)

// ParamsParser 定义参数解析器接口
type ParamsParser interface {
	// Parse 解析参数
	Parse(data map[string]any) (any, error)

	// Validate 验证参数
	Validate(params any) error
}

// DefaultParamsParser 实现默认的参数解析器
type DefaultParamsParser struct {
	// paramType 参数类型
	paramType string
}

// NewDefaultParamsParser 创建新的参数解析器实例
func NewDefaultParamsParser(paramType string) *DefaultParamsParser {
	return &DefaultParamsParser{
		paramType: paramType,
	}
}

// Parse 解析参数
func (p *DefaultParamsParser) Parse(data map[string]any) (any, error) {
	// 根据参数类型创建对应的参数实例
	paramInstance, err := p.createParamInstance()
	if err != nil {
		return nil, err
	}

	// 转换为params.BaseParams接口
	baseParam, ok := paramInstance.(params.BaseParams)
	if !ok {
		return nil, errors.New("param must implement BaseParams interface")
	}

	// 从map转换为参数对象
	if err := baseParam.FromMap(data); err != nil {
		return nil, err
	}

	return baseParam, nil
}

// Validate 验证参数
func (p *DefaultParamsParser) Validate(param any) error {
	// 转换为params.BaseParams接口
	baseParam, ok := param.(params.BaseParams)
	if !ok {
		return errors.New("param must implement BaseParams interface")
	}

	// 验证参数
	return baseParam.Validate()
}

// createParamInstance 根据参数类型创建对应的参数实例
func (p *DefaultParamsParser) createParamInstance() (any, error) {
	// 根据参数类型创建对应的参数实例
	switch p.paramType {
	// 文件处理动作参数
	case "scan_file_params":
		return params.NewScanFileParams(), nil
	case "scrape_file_params":
		return params.NewScrapeFileParams(), nil
	case "transfer_file_params":
		return params.NewTransferFileParams(), nil
	case "delete_file_params":
		return params.NewDeleteFileParams(), nil
	case "copy_file_params":
		return params.NewCopyFileParams(), nil
	case "move_file_params":
		return params.NewMoveFileParams(), nil

	// 资源获取动作参数
	case "fetch_torrents_params":
		return params.NewFetchTorrentsParams(), nil
	case "fetch_rss_params":
		return params.NewFetchRSSParams(), nil
	case "fetch_medias_params":
		return params.NewFetchMediasParams(), nil
	case "fetch_downloads_params":
		return params.NewFetchDownloadsParams(), nil

	// 过滤动作参数
	case "filter_torrents_params":
		return params.NewFilterTorrentsParams(), nil
	case "filter_medias_params":
		return params.NewFilterMediasParams(), nil

	// 核心业务动作参数
	case "add_download_params":
		return params.NewAddDownloadParams(), nil
	case "add_subscribe_params":
		return params.NewAddSubscribeParams(), nil
	case "send_message_params":
		return params.NewSendMessageParams(), nil

	// 系统功能动作参数
	case "invoke_plugin_params":
		return params.NewInvokePluginParams(), nil
	case "send_event_params":
		return params.NewSendEventParams(), nil
	case "note_params":
		return params.NewNoteParams(), nil

	default:
		// 默认返回基础参数实例
		return params.NewBaseParams(), nil
	}
}

// ParseParams 解析参数
func ParseParams(data map[string]any, paramType string) (any, error) {
	// 创建参数解析器
	parser := NewDefaultParamsParser(paramType)

	// 解析参数
	param, err := parser.Parse(data)
	if err != nil {
		return nil, err
	}

	// 验证参数
	if err := parser.Validate(param); err != nil {
		return nil, err
	}

	return param, nil
}
