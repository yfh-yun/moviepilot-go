package params

// NoteParams 笔记动作参数
type NoteParams struct {
	BaseParamsStruct

	// Content 笔记内容
	Content string `json:"content" mapstructure:"content"`

	// Title 笔记标题
	Title string `json:"title" mapstructure:"title"`

	// Tags 笔记标签列表
	Tags []string `json:"tags" mapstructure:"tags"`

	// Priority 笔记优先级，如low, medium, high
	Priority string `json:"priority" mapstructure:"priority"`

	// Category 笔记分类
	Category string `json:"category" mapstructure:"category"`

	// ExpireAt 过期时间，格式为2006-01-02T15:04:05Z07:00
	ExpireAt string `json:"expire_at" mapstructure:"expire_at"`
}

// Validate 验证笔记参数
func (p *NoteParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证笔记内容不能为空
	if p.Content == "" {
		return ErrNoteContentEmpty
	}

	// 初始化标签列表
	if p.Tags == nil {
		p.Tags = []string{}
	}

	return nil
}

// NewNoteParams 创建新的笔记参数实例
func NewNoteParams() *NoteParams {
	return &NoteParams{}
}
