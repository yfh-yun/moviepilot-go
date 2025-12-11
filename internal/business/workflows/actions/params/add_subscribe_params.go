package params

// AddSubscribeParams 添加订阅动作参数
type AddSubscribeParams struct {
	BaseParamsStruct

	// Title 订阅标题
	Title string `json:"title" mapstructure:"title"`

	// Type 订阅类型，如tv, movie, rss等
	Type string `json:"type" mapstructure:"type"`

	// Keyword 订阅关键词
	Keyword string `json:"keyword" mapstructure:"keyword"`

	// FilterRules 过滤规则
	FilterRules []FilterRule `json:"filter_rules" mapstructure:"filter_rules"`

	// NotificationEnabled 是否启用通知
	NotificationEnabled bool `json:"notification_enabled" mapstructure:"notification_enabled"`

	// AutoDownload 是否自动下载
	AutoDownload bool `json:"auto_download" mapstructure:"auto_download"`
}

// Validate 验证添加订阅参数
func (p *AddSubscribeParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证订阅标题不能为空
	if p.Title == "" {
		return ErrSubscribeTitleEmpty
	}

	// 验证订阅类型
	if p.Type == "" {
		p.Type = "tv" // 默认类型为tv
	}

	return nil
}

// NewAddSubscribeParams 创建新的添加订阅参数实例
func NewAddSubscribeParams() *AddSubscribeParams {
	return &AddSubscribeParams{}
}
