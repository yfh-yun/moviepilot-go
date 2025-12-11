package params

// InvokePluginParams 调用插件动作参数
type InvokePluginParams struct {
	BaseParamsStruct

	// PluginID 插件ID
	PluginID string `json:"plugin_id" mapstructure:"plugin_id"`

	// Method 调用方法
	Method string `json:"method" mapstructure:"method"`

	// Params 调用参数
	Params map[string]any `json:"params" mapstructure:"params"`

	// PluginType 插件类型，如site, indexer, mediaserver, notification等
	PluginType string `json:"plugin_type" mapstructure:"plugin_type"`

	// Timeout 超时时间（秒）
	Timeout int `json:"timeout" mapstructure:"timeout"`
}

// Validate 验证调用插件参数
func (p *InvokePluginParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证插件ID不能为空
	if p.PluginID == "" {
		return ErrPluginIDEmpty
	}

	// 验证调用方法不能为空
	if p.Method == "" {
		return ErrInvalidOperation
	}

	// 设置默认超时时间
	if p.Timeout <= 0 {
		p.Timeout = 30 // 默认30秒
	}

	return nil
}

// NewInvokePluginParams 创建新的调用插件参数实例
func NewInvokePluginParams() *InvokePluginParams {
	return &InvokePluginParams{}
}
