package workflow

import (
	"fmt"

	"moviepilot-go/internal/actions"
	"moviepilot-go/internal/business/media"
	"moviepilot-go/internal/business/storage"
	"moviepilot-go/internal/business/transfer"

	"go.uber.org/zap"
)

// LocalFileWorkflowConfig 描述构建本地文件刮削转移链路所需的依赖。
type LocalFileWorkflowConfig struct {
	Logger          *zap.Logger
	StorageService  storage.Service
	MediaService    media.Service
	TransferService transfer.Service
}

// LocalFileWorkflowOptions 控制工作流构建的元信息与参数。
type LocalFileWorkflowOptions struct {
	ID           string
	Name         string
	Description  string
	Version      string
	IncludeFetch bool
	ActionParams map[string]any
}

// NewLocalFileScrapeTransferWorkflow 构建包含 Scan -> Scrape -> Transfer 的 Workflow。
func NewLocalFileScrapeTransferWorkflow(cfg LocalFileWorkflowConfig) (*actions.ActionChain, error) {
	if cfg.StorageService == nil {
		return nil, fmt.Errorf("storage service is required")
	}
	if cfg.MediaService == nil {
		return nil, fmt.Errorf("media service is required")
	}
	if cfg.TransferService == nil {
		return nil, fmt.Errorf("transfer service is required")
	}

	scanAction := actions.NewScanFileAction(cfg.Logger, cfg.StorageService)
	scrapeAction := actions.NewScrapeFileAction(cfg.Logger, cfg.MediaService)
	fetchAction := actions.NewFetchTorrentsAction(cfg.Logger)
	transferAction := actions.NewTransferFileAction(cfg.Logger, cfg.TransferService)

	return actions.NewActionChain(true, scanAction, scrapeAction, fetchAction, transferAction), nil
}

func BuildLocalFileScrapeTransferWorkflow(cfg LocalFileWorkflowConfig, opts LocalFileWorkflowOptions) (*Workflow, error) {
	if cfg.StorageService == nil {
		return nil, fmt.Errorf("storage service is required")
	}
	if cfg.MediaService == nil {
		return nil, fmt.Errorf("media service is required")
	}
	if cfg.TransferService == nil {
		return nil, fmt.Errorf("transfer service is required")
	}

	if opts.ID == "" {
		opts.ID = "local_file_scrape_transfer"
	}
	if opts.Name == "" {
		opts.Name = "Local File Scrape Transfer"
	}
	if opts.Description == "" {
		opts.Description = "Scan local files, scrape metadata, and transfer"
	}
	if opts.Version == "" {
		opts.Version = "v1"
	}

	workflow := NewWorkflow(opts.ID, opts.Name, opts.Description, opts.Version)
	if workflow == nil {
		return nil, fmt.Errorf("failed to initialize workflow")
	}

	scanAction := actions.NewScanFileAction(cfg.Logger, cfg.StorageService)
	scrapeAction := actions.NewScrapeFileAction(cfg.Logger, cfg.MediaService)
	transferAction := actions.NewTransferFileAction(cfg.Logger, cfg.TransferService)

	scanTask, err := NewActionTask("task_scan_file", scanAction, cfg.Logger)
	if err != nil {
		return nil, err
	}

	scrapeTask, err := NewActionTask("task_scrape_file", scrapeAction, cfg.Logger)
	if err != nil {
		return nil, err
	}
	scrapeTask.AddDependency(scanTask.ID())

	var fetchTask *ActionTask
	if opts.IncludeFetch {
		fetchAction := actions.NewFetchTorrentsAction(cfg.Logger)
		fetchTask, err = NewActionTask("task_fetch_torrents", fetchAction, cfg.Logger)
		if err != nil {
			return nil, err
		}
		fetchTask.AddDependency(scrapeTask.ID())
	}

	transferTask, err := NewActionTask("task_transfer_file", transferAction, cfg.Logger)
	if err != nil {
		return nil, err
	}
	if opts.IncludeFetch && fetchTask != nil {
		transferTask.AddDependency(fetchTask.ID())
	} else {
		transferTask.AddDependency(scrapeTask.ID())
	}

	if err := workflow.AddTask(scanTask); err != nil {
		return nil, err
	}
	if err := workflow.AddTask(scrapeTask); err != nil {
		return nil, err
	}
	if opts.IncludeFetch && fetchTask != nil {
		if err := workflow.AddTask(fetchTask); err != nil {
			return nil, err
		}
	}
	if err := workflow.AddTask(transferTask); err != nil {
		return nil, err
	}

	if workflow.Variables == nil {
		workflow.Variables = make(map[string]interface{})
	}
	if opts.ActionParams != nil {
		workflow.Variables[actionParamsNamespace] = opts.ActionParams
	}

	return workflow, nil
}
