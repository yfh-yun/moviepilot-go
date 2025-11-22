package actions

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/storage"
	"moviepilot-go/internal/business/transfer"
)

// TransferMode 定义文件转移方式。
type TransferMode string

const (
	TransferModeMove     TransferMode = "move"
	TransferModeCopy     TransferMode = "copy"
	TransferModeHardlink TransferMode = "hardlink"
	TransferModeSoftlink TransferMode = "softlink"
)

// TransferFileParams 控制转移行为。
type TransferFileParams struct {
	BaseParams
	TargetRoot  string       `mapstructure:"target_root"`
	Mode        TransferMode `mapstructure:"mode"`
	Overwrite   bool         `mapstructure:"overwrite"`
	Category    string       `mapstructure:"category"`
	DryRun      bool         `mapstructure:"dry_run"`
	PreserveDir bool         `mapstructure:"preserve_dir"`
}

// Validate 参数检查。
func (p *TransferFileParams) Validate() error {
	if p.TargetRoot == "" {
		return errors.New("target_root is required")
	}
	if p.Mode == "" {
		p.Mode = TransferModeMove
	}
	switch p.Mode {
	case TransferModeMove, TransferModeCopy, TransferModeHardlink, TransferModeSoftlink:
	default:
		return errors.New("unsupported transfer mode")
	}
	return nil
}

// TransferFileAction 模拟文件转移，后续可对接 storage/transfer service。
type TransferFileAction struct {
	BaseActionImpl
	transferService transfer.Service
}

// NewTransferFileAction 创建实例。
func NewTransferFileAction(logger *zap.Logger, transferSvc transfer.Service) *TransferFileAction {
	return &TransferFileAction{
		BaseActionImpl:  NewBaseActionImpl("transfer_file", "Transfer files to media library", logger),
		transferService: transferSvc,
	}
}

// Execute 将 context 中的媒体映射成 TransferHistory 记录（当前为占位实现）。
func (a *TransferFileAction) Execute(workflowID int, rawParams any, ctx *ActionContext) (*ActionContext, error) {
	if ctx == nil {
		ctx = &ActionContext{}
	}
	ctx.Ensure()

	params := &TransferFileParams{}
	if err := DecodeParams(rawParams, params); err != nil {
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	if len(ctx.Medias) == 0 {
		msg := "no medias in context to transfer"
		a.SetResult(false, msg, nil)
		return ctx, errors.New(msg)
	}

	if a.transferService == nil {
		err := errors.New("transfer service is not configured")
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	tasks := make([]transfer.Task, 0, len(ctx.Medias))
	for _, media := range ctx.Medias {
		filePath := matchFileForMedia(ctx.Files, media.Title)
		targetPath := buildTargetPath(params.TargetRoot, media.Title, media.Type)

		storageMode := storage.TransferMode(params.Mode)
		if storageMode == "" {
			storageMode = storage.TransferMode(TransferModeMove)
		}

		tasks = append(tasks, transfer.Task{
			Media:      media,
			SourcePath: filePath,
			TargetPath: targetPath,
			Mode:       storageMode,
			Overwrite:  params.Overwrite,
			Category:   params.Category,
		})
	}

	records, err := a.transferService.Execute(tasks)
	if err != nil {
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	ctx.Transfers = records
	ctx.UpdatedAt = time.Now()
	a.SaveCache(workflowID, records)
	a.SetResult(true, "transfer planned", records)
	return ctx, nil
}

func matchFileForMedia(files []FileItem, title string) string {
	lowerTitle := strings.ToLower(title)
	for _, file := range files {
		if strings.Contains(strings.ToLower(file.Path), lowerTitle) {
			return file.Path
		}
	}
	if len(files) > 0 {
		return files[0].Path
	}
	return ""
}

func buildTargetPath(root, title, mediaType string) string {
	sanitizedTitle := strings.ReplaceAll(title, " ", ".")
	folder := mediaType
	if folder == "" {
		folder = "unknown"
	}
	return filepath.Join(root, folder, sanitizedTitle)
}
