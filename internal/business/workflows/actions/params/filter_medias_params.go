package params

// FilterMediasParams 过滤媒体动作参数
type FilterMediasParams struct {
	BaseParamsStruct

	// FilterRules 过滤规则列表
	FilterRules []FilterRule `json:"filter_rules" mapstructure:"filter_rules"`

	// AndMode 是否使用AND模式，true表示所有规则都必须满足，false表示只要满足一个
	AndMode bool `json:"and_mode" mapstructure:"and_mode"`

	// MediaType 媒体类型，movie或tv，为空表示不限制
	MediaType string `json:"media_type" mapstructure:"media_type"`
}

// Validate 验证过滤媒体参数
func (p *FilterMediasParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证过滤规则列表不能为空
	if len(p.FilterRules) == 0 {
		return ErrFilterRulesEmpty
	}

	// 验证每个过滤规则
	for _, rule := range p.FilterRules {
		if rule.Field == "" {
			return ErrFilterTypeInvalid
		}

		if rule.Operator == "" {
			return ErrFilterTypeInvalid
		}
	}

	// 验证媒体类型
	if p.MediaType != "" && p.MediaType != "movie" && p.MediaType != "tv" {
		return ErrInvalidOperation
	}

	return nil
}

// NewFilterMediasParams 创建新的过滤媒体参数实例
func NewFilterMediasParams() *FilterMediasParams {
	return &FilterMediasParams{}
}
