package params

// ScanFileParams 文件扫描动作参数
type ScanFileParams struct {
	BaseParamsStruct

	// ScanPath 扫描路径
	ScanPath string `json:"scan_path" mapstructure:"scan_path"`

	// IncludePatterns 包含的文件模式列表
	IncludePatterns []string `json:"include_patterns" mapstructure:"include_patterns"`

	// ExcludePatterns 排除的文件模式列表
	ExcludePatterns []string `json:"exclude_patterns" mapstructure:"exclude_patterns"`

	// Recursive 是否递归扫描
	Recursive bool `json:"recursive" mapstructure:"recursive"`

	// MaxDepth 最大扫描深度，0表示无限制
	MaxDepth int `json:"max_depth" mapstructure:"max_depth"`
}

// Validate 验证文件扫描参数
func (p *ScanFileParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证扫描路径不能为空
	if p.ScanPath == "" {
		return ErrScanPathEmpty
	}

	return nil
}

// NewScanFileParams 创建新的文件扫描参数实例
func NewScanFileParams() *ScanFileParams {
	return &ScanFileParams{}
}
