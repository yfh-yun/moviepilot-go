package params

// ScrapeFileParams 文件刮削动作参数
type ScrapeFileParams struct {
	BaseParamsStruct

	// FilePath 文件路径
	FilePath string `json:"file_path" mapstructure:"file_path"`

	// MediaType 媒体类型，movie或tv
	MediaType string `json:"media_type" mapstructure:"media_type"`

	// Language 刮削语言
	Language string `json:"language" mapstructure:"language"`

	// Overwrite 是否覆盖已有信息
	Overwrite bool `json:"overwrite" mapstructure:"overwrite"`

	// Providers 刮削提供商列表
	Providers []string `json:"providers" mapstructure:"providers"`
}

// Validate 验证文件刮削参数
func (p *ScrapeFileParams) Validate() error {
	// 调用基础参数验证
	if err := p.BaseParamsStruct.Validate(); err != nil {
		return err
	}

	// 验证文件路径不能为空
	if p.FilePath == "" {
		return ErrSourcePathEmpty
	}

	// 验证媒体类型
	if p.MediaType != "" && p.MediaType != "movie" && p.MediaType != "tv" {
		return ErrInvalidOperation
	}

	return nil
}

// NewScrapeFileParams 创建新的文件刮削参数实例
func NewScrapeFileParams() *ScrapeFileParams {
	return &ScrapeFileParams{}
}
