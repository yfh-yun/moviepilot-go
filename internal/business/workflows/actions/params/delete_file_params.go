package params

// DeleteFileParams 删除文件动作参数
type DeleteFileParams struct {
	BaseParamsStruct

	// FilePath 文件路径
	FilePath string `json:"file_path" mapstructure:"file_path"`

	// Force 是否强制删除
	Force bool `json:"force" mapstructure:"force"`

	// Recursive 是否递归删除目录
	Recursive bool `json:"recursive" mapstructure:"recursive"`
}

// Validate 验证删除文件参数
func (p *DeleteFileParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证文件路径不能为空
	if p.FilePath == "" {
		return ErrSourcePathEmpty
	}

	return nil
}

// NewDeleteFileParams 创建新的删除文件参数实例
func NewDeleteFileParams() *DeleteFileParams {
	return &DeleteFileParams{}
}
