package params

// FetchDownloadsParams 获取下载动作参数
type FetchDownloadsParams struct {
	BaseParamsStruct

	// ClientName 下载客户端名称
	ClientName string `json:"client_name" mapstructure:"client_name"`

	// Status 下载状态过滤，如active, paused, completed等
	Status string `json:"status" mapstructure:"status"`

	// Category 下载分类过滤
	Category string `json:"category" mapstructure:"category"`

	// Limit 返回结果数量限制，0表示无限制
	Limit int `json:"limit" mapstructure:"limit"`

	// SortBy 排序字段，如name, size, date_added等
	SortBy string `json:"sort_by" mapstructure:"sort_by"`

	// SortOrder 排序顺序，asc或desc
	SortOrder string `json:"sort_order" mapstructure:"sort_order"`
}

// Validate 验证获取下载参数
func (p *FetchDownloadsParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证客户端名称不能为空
	if p.ClientName == "" {
		return ErrDownloadClientEmpty
	}

	// 验证排序顺序
	if p.SortOrder != "" && p.SortOrder != "asc" && p.SortOrder != "desc" {
		return ErrInvalidOperation
	}

	return nil
}

// NewFetchDownloadsParams 创建新的获取下载参数实例
func NewFetchDownloadsParams() *FetchDownloadsParams {
	return &FetchDownloadsParams{}
}
