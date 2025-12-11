package params

// TransferFileParams 文件转移动作参数
type TransferFileParams struct {
	BaseParamsStruct

	// SourcePath 源路径
	SourcePath string `json:"source_path" mapstructure:"source_path"`

	// DestinationPath 目标路径
	DestinationPath string `json:"destination_path" mapstructure:"destination_path"`

	// DeleteSource 是否删除源文件
	DeleteSource bool `json:"delete_source" mapstructure:"delete_source"`

	// Overwrite 是否覆盖目标文件
	Overwrite bool `json:"overwrite" mapstructure:"overwrite"`

	// Verify 是否验证文件完整性
	Verify bool `json:"verify" mapstructure:"verify"`
}

// Validate 验证文件转移参数
func (p *TransferFileParams) Validate() error {
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

// NewTransferFileParams 创建新的文件转移参数实例
func NewTransferFileParams() *TransferFileParams {
	return &TransferFileParams{}
}
