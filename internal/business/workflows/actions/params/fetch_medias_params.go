package params

// FetchMediasParams 获取媒体动作参数
type FetchMediasParams struct {
	BaseParamsStruct

	// ServerName 媒体服务器名称
	ServerName string `json:"server_name" mapstructure:"server_name"`

	// MediaType 媒体类型，movie或tv
	MediaType string `json:"media_type" mapstructure:"media_type"`

	// LibraryName 媒体库名称
	LibraryName string `json:"library_name" mapstructure:"library_name"`

	// RecentlyAdded 是否只获取最近添加的媒体
	RecentlyAdded bool `json:"recently_added" mapstructure:"recently_added"`

	// Limit 返回结果数量限制，0表示无限制
	Limit int `json:"limit" mapstructure:"limit"`

	// SortBy 排序字段，如title, year, date_added等
	SortBy string `json:"sort_by" mapstructure:"sort_by"`

	// SortOrder 排序顺序，asc或desc
	SortOrder string `json:"sort_order" mapstructure:"sort_order"`
}

// Validate 验证获取媒体参数
func (p *FetchMediasParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证服务器名称不能为空
	if p.ServerName == "" {
		return ErrMediaServerEmpty
	}

	// 验证媒体类型
	if p.MediaType != "" && p.MediaType != "movie" && p.MediaType != "tv" {
		return ErrInvalidOperation
	}

	// 验证排序顺序
	if p.SortOrder != "" && p.SortOrder != "asc" && p.SortOrder != "desc" {
		return ErrInvalidOperation
	}

	return nil
}

// NewFetchMediasParams 创建新的获取媒体参数实例
func NewFetchMediasParams() *FetchMediasParams {
	return &FetchMediasParams{}
}
