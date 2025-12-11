package rule

// CustomRule 自定义规则项
type CustomRule struct {
	// 规则ID
	ID string `json:"id,omitempty"`
	// 名称
	Name string `json:"name,omitempty"`
	// 包含
	Include string `json:"include,omitempty"`
	// 排除
	Exclude string `json:"exclude,omitempty"`
	// 大小范围（MB）
	SizeRange string `json:"size_range,omitempty"`
	// 最少做种人数
	Seeders string `json:"seeders,omitempty"`
	// 发布时间
	PublishTime string `json:"publish_time,omitempty"`
}

// FilterRuleGroup 过滤规则组
type FilterRuleGroup struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 规则串
	RuleString string `json:"rule_string,omitempty"`
	// 适用媒体类型 None-全部 电影/电视剧
	MediaType string `json:"media_type,omitempty"`
	// 适用媒体类别 None-全部 对应二级分类
	Category string `json:"category,omitempty"`
}
