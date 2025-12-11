package params

// SendEventParams 发送事件动作参数
type SendEventParams struct {
	BaseParamsStruct

	// EventName 事件名称
	EventName string `json:"event_name" mapstructure:"event_name"`

	// EventData 事件数据
	EventData map[string]any `json:"event_data" mapstructure:"event_data"`

	// EventType 事件类型
	EventType string `json:"event_type" mapstructure:"event_type"`

	// Priority 事件优先级，如low, medium, high
	Priority string `json:"priority" mapstructure:"priority"`

	// IsSync 是否同步发送事件
	IsSync bool `json:"is_sync" mapstructure:"is_sync"`
}

// Validate 验证发送事件参数
func (p *SendEventParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证事件名称不能为空
	if p.EventName == "" {
		return ErrEventNameEmpty
	}

	// 初始化事件数据
	if p.EventData == nil {
		p.EventData = make(map[string]any)
	}

	return nil
}

// NewSendEventParams 创建新的发送事件参数实例
func NewSendEventParams() *SendEventParams {
	return &SendEventParams{}
}
