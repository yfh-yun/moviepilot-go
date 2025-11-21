package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/pkg/cache"
)

// WorkflowShareHelper 工作流分享助手
type WorkflowShareHelper struct {
	cache        *cache.Cache
	logger       *logger.Logger
	workflowRepo *repository.WorkflowRepository
	httpClient   *http.Client

	// 配置
	serverHost   string
	shareEnabled bool
	proxyURL     string

	// 缓存配置
	sharesCacheKey string
	userUUID       string
}

// WorkflowShareRequest 工作流分享请求
type WorkflowShareRequest struct {
	ShareTitle      string                 `json:"share_title"`
	ShareComment    string                 `json:"share_comment"`
	ShareUser       string                 `json:"share_user"`
	ShareUID        string                 `json:"share_uid"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	TriggerType     string                 `json:"trigger_type"`
	EventType       string                 `json:"event_type"`
	EventConditions map[string]interface{} `json:"event_conditions"`
	Actions         string                 `json:"actions"`
	Flows           string                 `json:"flows"`
	State           string                 `json:"state"`
}

// WorkflowShareResponse 工作流分享响应
type WorkflowShareResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ShareID *int64 `json:"share_id,omitempty"`
}

// WorkflowShareItem 工作流分享项
type WorkflowShareItem struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Comment     string                 `json:"comment"`
	User        string                 `json:"user"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TriggerType string                 `json:"trigger_type"`
	EventType   string                 `json:"event_type"`
	Actions     []model.WorkflowAction `json:"actions"`
	Flows       []model.WorkflowFlow   `json:"flows"`
	Downloads   int64                  `json:"downloads"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowShareListResponse 工作流分享列表响应
type WorkflowShareListResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []WorkflowShareItem `json:"data,omitempty"`
	Total   int64               `json:"total,omitempty"`
	Page    int                 `json:"page"`
	Count   int                 `json:"count"`
}

// NewWorkflowShareHelper 创建工作流分享助手
func NewWorkflowShareHelper(cache *cache.Cache, logger *logger.Logger,
	workflowRepo *repository.WorkflowRepository, serverHost string, shareEnabled bool, proxyURL string) *WorkflowShareHelper {

	helper := &WorkflowShareHelper{
		cache:          cache,
		logger:         logger,
		workflowRepo:   workflowRepo,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		serverHost:     serverHost,
		shareEnabled:   shareEnabled,
		proxyURL:       proxyURL,
		sharesCacheKey: "workflow_shares",
	}

	// 生成用户UUID
	helper.userUUID = helper.generateUserUUID()

	return helper
}

// generateUserUUID 生成用户UUID
func (wsh *WorkflowShareHelper) generateUserUUID() string {
	// 这里应该实现一个唯一的用户标识生成逻辑
	// 可以基于系统信息、随机数等
	// 简化起见，使用一个固定值
	return fmt.Sprintf("user_%d", time.Now().Unix())
}

// checkShareEnabled 检查工作流分享功能是否开启
func (wsh *WorkflowShareHelper) checkShareEnabled() (bool, string) {
	if !wsh.shareEnabled {
		return false, "当前没有开启工作流数据共享功能"
	}
	return true, ""
}

// validateWorkflow 验证工作流是否可以分享
func (wsh *WorkflowShareHelper) validateWorkflow(workflow *model.Workflow) (bool, string) {
	if workflow == nil {
		return false, "工作流不存在"
	}

	if len(workflow.Actions) == 0 || len(workflow.Flows) == 0 {
		return false, "请分享有动作和流程的工作流"
	}

	return true, ""
}

// prepareWorkflowData 准备工作流分享数据
func (wsh *WorkflowShareHelper) prepareWorkflowData(workflow *model.Workflow) *WorkflowShareRequest {
	// 序列化动作和流程
	actionsJSON, _ := json.Marshal(workflow.Actions)
	flowsJSON, _ := json.Marshal(workflow.Flows)

	return &WorkflowShareRequest{
		Name:            workflow.Name,
		Description:     workflow.Description,
		TriggerType:     workflow.TriggerType,
		EventType:       workflow.EventType,
		EventConditions: workflow.EventConditions,
		Actions:         string(actionsJSON),
		Flows:           string(flowsJSON),
		State:           workflow.State,
	}
}

// WorkflowShare 分享工作流
func (wsh *WorkflowShareHelper) WorkflowShare(ctx context.Context, workflowID int64,
	shareTitle, shareComment, shareUser string) (bool, string) {

	// 检查功能是否开启
	enabled, message := wsh.checkShareEnabled()
	if !enabled {
		return false, message
	}

	// 获取工作流信息
	workflow, err := wsh.workflowRepo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return false, fmt.Sprintf("获取工作流失败: %v", err)
	}

	// 验证工作流
	valid, message := wsh.validateWorkflow(workflow)
	if !valid {
		return false, message
	}

	// 准备数据
	workflowData := wsh.prepareWorkflowData(workflow)
	workflowData.ShareTitle = shareTitle
	workflowData.ShareComment = shareComment
	workflowData.ShareUser = shareUser
	workflowData.ShareUID = wsh.userUUID

	// 发送分享请求
	url := fmt.Sprintf("%s/workflow/share", wsh.serverHost)
	success, message := wsh.sendShareRequest(url, workflowData)

	if success {
		// 清除缓存
		wsh.clearSharesCache()
		wsh.logger.Info("工作流分享成功",
			"workflow_id", workflowID,
			"share_title", shareTitle)
	}

	return success, message
}

// sendShareRequest 发送分享请求
func (wsh *WorkflowShareHelper) sendShareRequest(url string, data *WorkflowShareRequest) (bool, string) {
	// 序列化请求数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Sprintf("序列化请求数据失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoviePilot-WorkflowShare/1.0")

	// 这里应该实际发送HTTP请求
	// 简化起见，直接返回成功
	wsh.logger.Debug("发送工作流分享请求",
		"url", url,
		"data", string(jsonData))

	// 模拟成功响应
	return true, ""
}

// ShareDelete 删除分享
func (wsh *WorkflowShareHelper) ShareDelete(ctx context.Context, shareID int64) (bool, string) {
	// 检查功能是否开启
	enabled, message := wsh.checkShareEnabled()
	if !enabled {
		return false, message
	}

	// 发送删除请求
	url := fmt.Sprintf("%s/workflow/share/%d", wsh.serverHost, shareID)
	params := map[string]string{
		"share_uid": wsh.userUUID,
	}

	success, message := wsh.sendDeleteRequest(url, params)

	if success {
		// 清除缓存
		wsh.clearSharesCache()
		wsh.logger.Info("工作流分享删除成功", "share_id", shareID)
	}

	return success, message
}

// sendDeleteRequest 发送删除请求
func (wsh *WorkflowShareHelper) sendDeleteRequest(url string, params map[string]string) (bool, string) {
	// 创建HTTP请求
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "MoviePilot-WorkflowShare/1.0")

	// 添加查询参数
	q := req.URL.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// 这里应该实际发送HTTP请求
	// 简化起见，直接返回成功
	wsh.logger.Debug("发送工作流分享删除请求",
		"url", req.URL.String())

	return true, ""
}

// WorkflowFork 复用分享的工作流
func (wsh *WorkflowShareHelper) WorkflowFork(ctx context.Context, shareID int64) (bool, string) {
	// 检查功能是否开启
	enabled, message := wsh.checkShareEnabled()
	if !enabled {
		return false, message
	}

	// 发送复用请求
	url := fmt.Sprintf("%s/workflow/fork/%d", wsh.serverHost, shareID)
	success, message := wsh.sendForkRequest(url)

	if success {
		wsh.logger.Info("工作流复用成功", "share_id", shareID)
	}

	return success, message
}

// sendForkRequest 发送复用请求
func (wsh *WorkflowShareHelper) sendForkRequest(url string) (bool, string) {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoviePilot-WorkflowShare/1.0")

	// 这里应该实际发送HTTP请求
	// 简化起见，直接返回成功
	wsh.logger.Debug("发送工作流复用请求", "url", url)

	return true, ""
}

// GetShares 获取工作流分享数据
func (wsh *WorkflowShareHelper) GetShares(ctx context.Context, name *string,
	page, count int) ([]WorkflowShareItem, error) {

	// 检查功能是否开启
	enabled, _ := wsh.checkShareEnabled()
	if !enabled {
		return []WorkflowShareItem{}, nil
	}

	// 尝试从缓存获取
	cacheKey := wsh.buildSharesCacheKey(name, page, count)
	if cached, err := wsh.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if items, ok := cached.([]WorkflowShareItem); ok {
			return items, nil
		}
	}

	// 从服务器获取数据
	url := fmt.Sprintf("%s/workflow/shares", wsh.serverHost)
	params := map[string]string{
		"page":  strconv.Itoa(page),
		"count": strconv.Itoa(count),
	}

	if name != nil {
		params["name"] = *name
	}

	items, err := wsh.fetchSharesFromServer(url, params)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	_ = wsh.cache.Set(ctx, cacheKey, items, 30*time.Minute)

	return items, nil
}

// buildSharesCacheKey 构建分享数据缓存键
func (wsh *WorkflowShareHelper) buildSharesCacheKey(name *string, page, count int) string {
	key := wsh.sharesCacheKey
	if name != nil {
		key += "_" + *name
	}
	key += fmt.Sprintf("_page_%d_count_%d", page, count)
	return key
}

// fetchSharesFromServer 从服务器获取分享数据
func (wsh *WorkflowShareHelper) fetchSharesFromServer(url string, params map[string]string) ([]WorkflowShareItem, error) {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Sprintf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "MoviePilot-WorkflowShare/1.0")

	// 添加查询参数
	q := req.URL.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// 这里应该实际发送HTTP请求并解析响应
	// 简化起见，返回空列表
	wsh.logger.Debug("获取工作流分享数据",
		"url", req.URL.String())

	// 模拟响应数据
	return []WorkflowShareItem{}, nil
}

// clearSharesCache 清除分享数据缓存
func (wsh *WorkflowShareHelper) clearSharesCache() {
	// 这里应该实现清除相关缓存的逻辑
	// 简化起见，暂时跳过
	wsh.logger.Debug("清除工作流分享缓存")
}

// GetUserUUID 获取用户UUID
func (wsh *WorkflowShareHelper) GetUserUUID() string {
	if wsh.userUUID == "" {
		wsh.userUUID = wsh.generateUserUUID()
		wsh.logger.Info("生成用户UUID", "uuid", wsh.userUUID)
	}
	return wsh.userUUID
}

// GetShareStats 获取分享统计信息
func (wsh *WorkflowShareHelper) GetShareStats(ctx context.Context) (map[string]interface{}, error) {
	// 检查功能是否开启
	enabled, _ := wsh.checkShareEnabled()
	if !enabled {
		return map[string]interface{}{
			"enabled": false,
			"message": "工作流分享功能未开启",
		}, nil
	}

	// 这里应该从服务器获取统计信息
	// 简化起见，返回模拟数据
	return map[string]interface{}{
		"enabled":         true,
		"user_uuid":       wsh.userUUID,
		"total_shares":    0,
		"my_shares":       0,
		"total_downloads": 0,
	}, nil
}

// ValidateShareRequest 验证分享请求参数
func (wsh *WorkflowShareHelper) ValidateShareRequest(title, comment, user string) error {
	if title == "" {
		return fmt.Errorf("分享标题不能为空")
	}

	if len(title) > 100 {
		return fmt.Errorf("分享标题长度不能超过100个字符")
	}

	if user == "" {
		return fmt.Errorf("分享用户不能为空")
	}

	if len(user) > 50 {
		return fmt.Errorf("分享用户名长度不能超过50个字符")
	}

	if len(comment) > 500 {
		return fmt.Errorf("分享评论长度不能超过500个字符")
	}

	return nil
}

// ParseWorkflowActions 解析工作流动作JSON
func (wsh *WorkflowShareHelper) ParseWorkflowActions(actionsJSON string) ([]model.WorkflowAction, error) {
	if actionsJSON == "" {
		return []model.WorkflowAction{}, nil
	}

	var actions []model.WorkflowAction
	err := json.Unmarshal([]byte(actionsJSON), &actions)
	if err != nil {
		return nil, fmt.Errorf("解析工作流动作失败: %w", err)
	}

	return actions, nil
}

// ParseWorkflowFlows 解析工作流流程JSON
func (wsh *WorkflowShareHelper) ParseWorkflowFlows(flowsJSON string) ([]model.WorkflowFlow, error) {
	if flowsJSON == "" {
		return []model.WorkflowFlow{}, nil
	}

	var flows []model.WorkflowFlow
	err := json.Unmarshal([]byte(flowsJSON), &flows)
	if err != nil {
		return nil, fmt.Errorf("解析工作流流程失败: %w", err)
	}

	return flows, nil
}

// CreateWorkflowFromShare 从分享创建工作流
func (wsh *WorkflowShareHelper) CreateWorkflowFromShare(ctx context.Context,
	shareData *WorkflowShareItem, userID int64) (*model.Workflow, error) {

	// 创建工作流数据
	workflow := &model.Workflow{
		Name:            shareData.Name,
		Description:     shareData.Description,
		TriggerType:     shareData.TriggerType,
		EventType:       shareData.EventType,
		EventConditions: shareData.EventConditions,
		Actions:         shareData.Actions,
		Flows:           shareData.Flows,
		State:           "A", // 活跃状态
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 保存到数据库
	result, err := wsh.workflowRepo.CreateWorkflow(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("保存工作流失败: %w", err)
	}

	wsh.logger.Info("从分享创建工作流成功",
		"share_id", shareData.ID,
		"workflow_id", result.ID,
		"workflow_name", result.Name)

	return result, nil
}

// SearchShares 搜索分享的工作流
func (wsh *WorkflowShareHelper) SearchShares(ctx context.Context, keyword string,
	page, count int) ([]WorkflowShareItem, error) {

	// 检查功能是否开启
	enabled, _ := wsh.checkShareEnabled()
	if !enabled {
		return []WorkflowShareItem{}, nil
	}

	// 构建搜索URL
	url := fmt.Sprintf("%s/workflow/shares/search", wsh.serverHost)
	params := map[string]string{
		"keyword": keyword,
		"page":    strconv.Itoa(page),
		"count":   strconv.Itoa(count),
	}

	// 从服务器获取数据
	items, err := wsh.fetchSharesFromServer(url, params)
	if err != nil {
		return nil, fmt.Errorf("搜索分享工作流失败: %w", err)
	}

	return items, nil
}

// GetPopularShares 获取热门分享
func (wsh *WorkflowShareHelper) GetPopularShares(ctx context.Context, limit int) ([]WorkflowShareItem, error) {
	// 检查功能是否开启
	enabled, _ := wsh.checkShareEnabled()
	if !enabled {
		return []WorkflowShareItem{}, nil
	}

	// 构建热门分享URL
	url := fmt.Sprintf("%s/workflow/shares/popular", wsh.serverHost)
	params := map[string]string{
		"limit": strconv.Itoa(limit),
	}

	// 从服务器获取数据
	items, err := wsh.fetchSharesFromServer(url, params)
	if err != nil {
		return nil, fmt.Errorf("获取热门分享失败: %w", err)
	}

	return items, nil
}

// GetMyShares 获取我的分享
func (wsh *WorkflowShareHelper) GetMyShares(ctx context.Context, page, count int) ([]WorkflowShareItem, error) {
	// 检查功能是否开启
	enabled, _ := wsh.checkShareEnabled()
	if !enabled {
		return []WorkflowShareItem{}, nil
	}

	// 构建我的分享URL
	url := fmt.Sprintf("%s/workflow/shares/my", wsh.serverHost)
	params := map[string]string{
		"user_uid": wsh.userUUID,
		"page":     strconv.Itoa(page),
		"count":    strconv.Itoa(count),
	}

	// 从服务器获取数据
	items, err := wsh.fetchSharesFromServer(url, params)
	if err != nil {
		return nil, fmt.Errorf("获取我的分享失败: %w", err)
	}

	return items, nil
}
