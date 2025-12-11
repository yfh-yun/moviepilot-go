package workflowapi

// StartLocalFileWorkflowRequest 定义触发本地链路所需的参数。
type StartLocalFileWorkflowRequest struct {
	RootPath      string   `json:"root_path" binding:"required"`
	Include       []string `json:"include" binding:"omitempty,dive,min=1"`
	Exclude       []string `json:"exclude" binding:"omitempty,dive,min=1"`
	MaxDepth      int      `json:"max_depth" binding:"gte=0"`
	FollowSymlink bool     `json:"follow_symlink"`

	TargetRoot  string `json:"target_root" binding:"required"`
	Mode        string `json:"mode" binding:"omitempty,oneof=move copy link hardlink softlink"`
	Category    string `json:"category" binding:"omitempty,max=64"`
	Overwrite   bool   `json:"overwrite"`
	PreserveDir bool   `json:"preserve_dir"`
	DryRun      bool   `json:"dry_run"`

	IncludeFetch      bool     `json:"include_fetch"`
	FetchKeywords     []string `json:"fetch_keywords" binding:"omitempty,dive,min=1"`
	WaitForCompletion bool     `json:"wait_for_completion"`

	ForceRefresh bool   `json:"force_refresh"`
	Source       string `json:"source" binding:"omitempty,max=32"`
}

// StartLocalFileWorkflowResponse 表示 API 返回的基础信息。
type StartLocalFileWorkflowResponse struct {
	WorkflowID string `json:"workflow_id"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Result     any    `json:"result,omitempty"`
}
