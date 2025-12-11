package interfaces

// ListOptions 列表查询选项
type ListOptions struct {
	Page     int
	PageSize int
	OrderBy  string
	Order    string // asc, desc
	Filters  map[string]any
}

// GetOffset 获取偏移量
func (o *ListOptions) GetOffset() int {
	if o.Page <= 0 {
		o.Page = 1
	}
	if o.PageSize <= 0 {
		o.PageSize = 20
	}
	return (o.Page - 1) * o.PageSize
}

// GetLimit 获取限制数量
func (o *ListOptions) GetLimit() int {
	if o.PageSize <= 0 {
		o.PageSize = 20
	}
	return o.PageSize
}
