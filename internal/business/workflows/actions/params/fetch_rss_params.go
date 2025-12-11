package params

// FetchRSSParams 获取RSS动作参数
type FetchRSSParams struct {
	BaseParamsStruct

	// URLs RSS源URL列表
	URLs []string `json:"urls" mapstructure:"urls"`

	// Limit 每个RSS源返回的项目数量限制，0表示无限制
	Limit int `json:"limit" mapstructure:"limit"`

	// FilterKeywords 过滤关键词列表
	FilterKeywords []string `json:"filter_keywords" mapstructure:"filter_keywords"`

	// ExcludeKeywords 排除关键词列表
	ExcludeKeywords []string `json:"exclude_keywords" mapstructure:"exclude_keywords"`

	// IncludeFullContent 是否包含完整内容
	IncludeFullContent bool `json:"include_full_content" mapstructure:"include_full_content"`
}

// Validate 验证获取RSS参数
func (p *FetchRSSParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证URL列表不能为空
	if len(p.URLs) == 0 {
		return ErrRSSURLsEmpty
	}

	return nil
}

// NewFetchRSSParams 创建新的获取RSS参数实例
func NewFetchRSSParams() *FetchRSSParams {
	return &FetchRSSParams{}
}
