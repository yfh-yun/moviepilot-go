package params

// FilterTorrentsParams 过滤种子动作参数
type FilterTorrentsParams struct {
	BaseParamsStruct

	// FilterRules 过滤规则列表
	FilterRules []FilterRule `json:"filter_rules" mapstructure:"filter_rules"`

	// AndMode 是否使用AND模式，true表示所有规则都必须满足，false表示只要满足一个
	AndMode bool `json:"and_mode" mapstructure:"and_mode"`

	// IncludeZeroSize 是否包含大小为0的种子
	IncludeZeroSize bool `json:"include_zero_size" mapstructure:"include_zero_size"`
}

// FilterRule 定义过滤规则
type FilterRule struct {
	// Field 字段名称，如name, size, seeders, leechers等
	Field string `json:"field" mapstructure:"field"`

	// Operator 操作符，如eq, ne, gt, gte, lt, lte, contains, not_contains, in, not_in等
	Operator string `json:"operator" mapstructure:"operator"`

	// Value 比较值
	Value any `json:"value" mapstructure:"value"`
}

// Validate 验证过滤种子参数
func (p *FilterTorrentsParams) Validate() error {
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

	return nil
}

// NewFilterTorrentsParams 创建新的过滤种子参数实例
func NewFilterTorrentsParams() *FilterTorrentsParams {
	return &FilterTorrentsParams{}
}
