package params

// MoveFileParams 文件移动动作参数
type MoveFileParams struct {
	BaseParamsStruct

	// SourcePath 源路径
	SourcePath string `json:"source_path" mapstructure:"source_path"`

	// DestinationPath 目标路径
	DestinationPath string `json:"destination_path" mapstructure:"destination_path"`

	// Overwrite 是否覆盖目标文件
	Overwrite bool `json:"overwrite" mapstructure:"overwrite"`

	// Verify 是否验证文件完整性
	Verify bool `json:"verify" mapstructure:"verify"`
}

// Validate 验证文件移动参数
func (p *MoveFileParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证源路径不能为空
	if p.SourcePath == "" {
		return ErrSourcePathEmpty
	}

	// 验证目标路径不能为空
	if p.DestinationPath == "" {
		return ErrDestinationPathEmpty
	}

	return nil
}

// NewMoveFileParams 创建新的文件移动参数实例
func NewMoveFileParams() *MoveFileParams {
	return &MoveFileParams{}
}
