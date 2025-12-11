package params

// AddDownloadParams 添加下载动作参数
type AddDownloadParams struct {
	BaseParamsStruct

	// Downloader 下载器名称
	Downloader string `json:"downloader" mapstructure:"downloader"`

	// SavePath 保存路径
	SavePath string `json:"save_path" mapstructure:"save_path"`

	// Labels 标签（,分隔）
	Labels string `json:"labels" mapstructure:"labels"`

	// OnlyLack 仅下载缺失的资源
	OnlyLack bool `json:"only_lack" mapstructure:"only_lack"`
}

// Validate 验证添加下载参数
func (p *AddDownloadParams) Validate() error {
	// 调用基础参数验证
	return p.BaseParamsStruct.Validate()
}

// NewAddDownloadParams 创建新的添加下载参数实例
func NewAddDownloadParams() *AddDownloadParams {
	return &AddDownloadParams{}
}
