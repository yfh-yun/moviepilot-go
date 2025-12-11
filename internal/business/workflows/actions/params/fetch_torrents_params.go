package params

// FetchTorrentsParams 获取种子动作参数
type FetchTorrentsParams struct {
	BaseParamsStruct

	// ClientName 种子客户端名称
	ClientName string `json:"client_name" mapstructure:"client_name"`

	// Status 种子状态过滤，如active, paused, completed等
	Status string `json:"status" mapstructure:"status"`

	// Category 种子分类过滤
	Category string `json:"category" mapstructure:"category"`

	// Limit 返回结果数量限制，0表示无限制
	Limit int `json:"limit" mapstructure:"limit"`

	// SortBy 排序字段，如name, size, date_added等
	SortBy string `json:"sort_by" mapstructure:"sort_by"`

	// SortOrder 排序顺序，asc或desc
	SortOrder string `json:"sort_order" mapstructure:"sort_order"`
}

// Validate 验证获取种子参数
func (p *FetchTorrentsParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证客户端名称不能为空
	if p.ClientName == "" {
		return ErrTorrentClientEmpty
	}

	// 验证排序顺序
	if p.SortOrder != "" && p.SortOrder != "asc" && p.SortOrder != "desc" {
		return ErrInvalidOperation
	}

	return nil
}

// NewFetchTorrentsParams 创建新的获取种子参数实例
func NewFetchTorrentsParams() *FetchTorrentsParams {
	return &FetchTorrentsParams{}
}
