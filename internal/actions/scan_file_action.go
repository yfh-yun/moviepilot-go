package actions

import (
	"errors"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/storage"
)

// ScanFileParams 描述扫描文件所需参数。
type ScanFileParams struct {
	BaseParams
	RootPath       string   `mapstructure:"root_path"`
	Include        []string `mapstructure:"include_patterns"`
	Exclude        []string `mapstructure:"exclude_patterns"`
	MaxDepth       int      `mapstructure:"max_depth"`
	FollowSymlinks bool     `mapstructure:"follow_symlinks"`
}

// Validate 参数合法性。
func (p *ScanFileParams) Validate() error {
	if p.RootPath == "" {
		return errors.New("root_path is required")
	}
	if p.MaxDepth < 0 {
		p.MaxDepth = 0
	}
	return nil
}

// ScanFileAction 负责扫描本地目录。
type ScanFileAction struct {
	BaseActionImpl
	storageService storage.Service
}

// NewScanFileAction 创建实例。
func NewScanFileAction(logger *zap.Logger, storageSvc storage.Service) *ScanFileAction {
	return &ScanFileAction{
		BaseActionImpl: NewBaseActionImpl("scan_file", "Scan local directories", logger),
		storageService: storageSvc,
	}
}

// Execute 扫描并返回文件列表。
func (a *ScanFileAction) Execute(workflowID int, rawParams any, ctx *ActionContext) (*ActionContext, error) {
	if ctx == nil {
		ctx = &ActionContext{}
	}
	ctx.Ensure()

	params := &ScanFileParams{}
	if err := DecodeParams(rawParams, params); err != nil {
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	items, err := a.storageService.Scan(storage.ScanOptions{
		RootPath:       params.RootPath,
		Include:        params.Include,
		Exclude:        params.Exclude,
		MaxDepth:       params.MaxDepth,
		FollowSymlinks: params.FollowSymlinks,
	})
	if err != nil {
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	files := make([]FileItem, 0, len(items))
	for _, item := range items {
		files = append(files, FileItem{
			Path:    item.Path,
			Size:    item.Size,
			ModTime: item.ModTime,
			IsDir:   item.IsDir,
		})
	}

	ctx.Files = files
	ctx.UpdatedAt = time.Now()
	a.SaveCache(workflowID, files)
	a.SetResult(true, "scan complete", files)
	return ctx, nil
}
