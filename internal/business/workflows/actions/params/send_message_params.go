package params

// SendMessageParams 发送消息动作参数
type SendMessageParams struct {
	BaseParamsStruct

	// Client 消息渠道列表
	Client []string `json:"client" mapstructure:"client"`

	// UserID 用户ID
	UserID interface{} `json:"userid" mapstructure:"userid"`
}

// Validate 验证发送消息参数
func (p *SendMessageParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 设置默认值
	if p.Client == nil {
		p.Client = []string{}
	}

	return nil
}

// NewSendMessageParams 创建新的发送消息参数实例
func NewSendMessageParams() *SendMessageParams {
	return &SendMessageParams{
		Client: []string{},
	}
}
