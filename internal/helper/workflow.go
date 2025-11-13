package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/core/cache"
	"moviepilot-go/internal/db"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
)

// WorkflowHelper 工作流分享等
type WorkflowHelper struct {
	workflowShare     string
	workflowShares    string
	workflowFork      string
	sharesCacheRegion string
	shareUserID       string
}

// 使用弱单例模�?var (
	workflowHelperInstance *WorkflowHelper
	once                   sync.Once
)

// GetWorkflowHelper 获取工作流帮助类实例
func GetWorkflowHelper() *WorkflowHelper {
	once.Do(func() {
		workflowHelperInstance = &WorkflowHelper{
			workflowShare:     fmt.Sprintf("%s/workflow/share", config.GlobalConfig.MPServerHost),
			workflowShares:    fmt.Sprintf("%s/workflow/shares", config.GlobalConfig.MPServerHost),
			workflowFork:      fmt.Sprintf("%s/workflow/fork/%%s", config.GlobalConfig.MPServerHost),
			sharesCacheRegion: "workflow_share",
			shareUserID:       "",
		}
		workflowHelperInstance.GetUserUUID()
	})
	return workflowHelperInstance
}

// CheckWorkflowShareEnabled 检查工作流分享功能是否开�?func (w *WorkflowHelper) CheckWorkflowShareEnabled() (bool, string) {
	if !config.GlobalConfig.WorkflowStatisticShare {
		return false, "当前没有开启工作流数据共享功能"
	}
	return true, ""
}

// ValidateWorkflow 验证工作流是否可以分�?func (w *WorkflowHelper) ValidateWorkflow(workflow *db.Workflow) (bool, string) {
	if workflow == nil {
		return false, "工作流不存在"
	}

	if len(workflow.Actions) == 0 || len(workflow.Flows) == 0 {
		return false, "请分享有动作和流程的工作�?
	}

	return true, ""
}

// PrepareWorkflowData 准备工作流分享数�?func (w *WorkflowHelper) PrepareWorkflowData(workflow *db.Workflow) map[string]interface{} {
	workflowDict := workflow.ToMap()
	delete(workflowDict, "id")
	delete(workflowDict, "context")
	
	// 序列化actions和flows
	if actions, ok := workflowDict["actions"].([]interface{}); ok {
		actionsJSON, _ := json.Marshal(actions)
		workflowDict["actions"] = string(actionsJSON)
	}
	
	if flows, ok := workflowDict["flows"].([]interface{}); ok {
		flowsJSON, _ := json.Marshal(flows)
		workflowDict["flows"] = string(flowsJSON)
	}
	
	return workflowDict
}

// BuildSharePayload 构建分享请求载荷
func (w *WorkflowHelper) BuildSharePayload(shareTitle, shareComment, shareUser string, workflowDict map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{})
	payload["share_title"] = shareTitle
	payload["share_comment"] = shareComment
	payload["share_user"] = shareUser
	payload["share_uid"] = w.shareUserID
	
	// 合并workflowDict到payload
	for k, v := range workflowDict {
		payload[k] = v
	}
	
	return payload
}

// handleResponse 处理HTTP响应
func (w *WorkflowHelper) handleResponse(res *http.Response, clearCache bool) (bool, string) {
	if res == nil {
		return false, "连接MoviePilot服务器失�?
	}

	// 检查响应状�?	success := res.StatusCode == http.StatusOK

	if success {
		// 清除缓存
		if clearCache {
			w.getSharesCacheClear()
		}
		return true, ""
	} else {
		var errorMsg string
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Errorf("工作流响应JSON解析失败: %v", err)
			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)
			errorMsg = fmt.Sprintf("响应解析失败: %s...", buf.String()[:100])
		} else {
			if msg, ok := result["message"].(string); ok {
				errorMsg = msg
			} else {
				errorMsg = "未知错误"
			}
		}
		return false, errorMsg
	}
}

// handleListResponse 处理返回List的HTTP响应
func (w *WorkflowHelper) handleListResponse(res *http.Response) []map[string]interface{} {
	if res != nil && res.StatusCode == http.StatusOK {
		var result []map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			logger.Errorf("工作流列表响应JSON解析失败: %v", err)
			return []map[string]interface{}{}
		}
		return result
	}
	return []map[string]interface{}{}
}

// WorkflowShare 分享工作�?func (w *WorkflowHelper) WorkflowShare(workflowID int, shareTitle, shareComment, shareUser string) (bool, string) {
	// 检查功能是否开�?	enabled, message := w.CheckWorkflowShareEnabled()
	if !enabled {
		return false, message
	}

	// 获取工作流信�?	workflowOper := db.GetWorkflowOper()
	workflow := workflowOper.Get(workflowID)

	// 验证工作�?	valid, message := w.ValidateWorkflow(workflow)
	if !valid {
		return false, message
	}

	// 准备数据
	workflowDict := w.PrepareWorkflowData(workflow)
	payload := w.BuildSharePayload(shareTitle, shareComment, shareUser, workflowDict)

	// 发送分享请�?	payloadJSON, _ := json.Marshal(payload)
	httpClient := utils.NewRequestUtils("", "", "", config.GlobalConfig.Proxy, "application/json", 10*time.Second)
	res, err := httpClient.Post(w.workflowShare, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return false, "连接MoviePilot服务器失�?
	}
	defer res.Body.Close()

	return w.handleResponse(res, true)
}

// AsyncWorkflowShare 异步分享工作�?func (w *WorkflowHelper) AsyncWorkflowShare(workflowID int, shareTitle, shareComment, shareUser string) (bool, string) {
	// 检查功能是否开�?	enabled, message := w.CheckWorkflowShareEnabled()
	if !enabled {
		return false, message
	}

	// 获取工作流信�?	workflowOper := db.GetWorkflowOper()
	workflow := workflowOper.Get(workflowID)

	// 验证工作�?	valid, message := w.ValidateWorkflow(workflow)
	if !valid {
		return false, message
	}

	// 准备数据
	workflowDict := w.PrepareWorkflowData(workflow)
	payload := w.BuildSharePayload(shareTitle, shareComment, shareUser, workflowDict)

	// 发送分享请�?	payloadJSON, _ := json.Marshal(payload)
	httpClient := utils.NewAsyncRequestUtils("", "", "", config.GlobalConfig.Proxy, "application/json", 10*time.Second)
	res, err := httpClient.Post(w.workflowShare, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return false, "连接MoviePilot服务器失�?
	}
	defer res.Body.Close()

	return w.handleResponse(res, true)
}

// ShareDelete 删除分享
func (w *WorkflowHelper) ShareDelete(shareID int) (bool, string) {
	// 检查功能是否开�?	enabled, message := w.CheckWorkflowShareEnabled()
	if !enabled {
		return false, message
	}

	// 构建URL和参�?	url := fmt.Sprintf("%s/%d", w.workflowShare, shareID)
	params := fmt.Sprintf("share_uid=%s", w.shareUserID)
	
	// 发送删除请�?	httpClient := utils.NewRequestUtils("", "", "", config.GlobalConfig.Proxy, "", 5*time.Second)
	res, err := httpClient.DeleteRes(fmt.Sprintf("%s?%s", url, params))
	if err != nil {
		return false, "连接MoviePilot服务器失�?
	}
	defer res.Body.Close()

	return w.handleResponse(res, true)
}

// AsyncShareDelete 异步删除分享
func (w *WorkflowHelper) AsyncShareDelete(shareID int) (bool, string) {
	// 检查功能是否开�?	enabled, message := w.CheckWorkflowShareEnabled()
	if !enabled {
		return false, message
	}

	// 构建URL和参�?	url := fmt.Sprintf("%s/%d", w.workflowShare, shareID)
	params := fmt.Sprintf("share_uid=%s", w.shareUserID)
	
	// 发送删除请�?	httpClient := utils.NewAsyncRequestUtils("", "", "", config.GlobalConfig.Proxy, "", 5*time.Second)
	res, err := httpClient.DeleteRes(fmt.Sprintf("%s?%s", url, params))
	if err != nil {
		return false, "连接MoviePilot服务器失�?
	}
	defer res.Body.Close()

	return w.handleResponse(res, true)
}

// WorkflowFork 复用分享的工作流
func (w *WorkflowHelper) WorkflowFork(shareID int) (bool, string) {
	// 检查功能是否开�?	enabled, message := w.CheckWorkflowShareEnabled()
	if !enabled {
		return false, message
	}

	// 构建URL
	url := fmt.Sprintf(w.workflowFork, shareID)
	
	// 发送请�?	httpClient := utils.NewRequestUtils("", "", "", config.GlobalConfig.Proxy, "application/json", 5*time.Second)
	res, err := httpClient.GetRes(url)
	if err != nil {
		return false, "连接MoviePilot服务器失�?
	}
	defer res.Body.Close()

	return w.handleResponse(res, false)
}

// AsyncWorkflowFork 异步复用分享的工作流
func (w *WorkflowHelper) AsyncWorkflowFork(shareID int) (bool, string) {
	// 检查功能是否开�?	enabled, message := w.CheckWorkflowShareEnabled()
	if !enabled {
		return false, message
	}

	// 构建URL
	url := fmt.Sprintf(w.workflowFork, shareID)
	
	// 发送请�?	httpClient := utils.NewAsyncRequestUtils("", "", "", config.GlobalConfig.Proxy, "application/json", 5*time.Second)
	res, err := httpClient.GetRes(url)
	if err != nil {
		return false, "连接MoviePilot服务器失�?
	}
	defer res.Body.Close()

	return w.handleResponse(res, false)
}

// getSharesCacheClear 清除获取分享缓存
func (w *WorkflowHelper) getSharesCacheClear() {
	cache.Delete(w.sharesCacheRegion)
}

// GetShares 获取工作流分享数�?func (w *WorkflowHelper) GetShares(name string, page, count int) []map[string]interface{} {
	// 检查功能是否开�?	enabled, _ := w.CheckWorkflowShareEnabled()
	if !enabled {
		return []map[string]interface{}{}
	}

	// 检查缓�?	cacheKey := fmt.Sprintf("workflow_shares_%s_%d_%d", name, page, count)
	if val, exists := cache.Get(cacheKey); exists {
		if shares, ok := val.([]map[string]interface{}); ok {
			return shares
		}
	}

	// 构建参数
	params := fmt.Sprintf("name=%s&page=%d&count=%d", name, page, count)

	// 发送请�?	httpClient := utils.NewRequestUtils("", "", "", config.GlobalConfig.Proxy, "", 15*time.Second)
	res, err := httpClient.GetRes(fmt.Sprintf("%s?%s", w.workflowShares, params))
	if err != nil {
		return []map[string]interface{}{}
	}
	defer res.Body.Close()

	shares := w.handleListResponse(res)
	
	// 缓存结果
	cache.Set(cacheKey, shares, 1*time.Hour)
	return shares
}

// AsyncGetShares 异步获取工作流分享数�?func (w *WorkflowHelper) AsyncGetShares(name string, page, count int) []map[string]interface{} {
	// 检查功能是否开�?	enabled, _ := w.CheckWorkflowShareEnabled()
	if !enabled {
		return []map[string]interface{}{}
	}

	// 检查缓�?	cacheKey := fmt.Sprintf("workflow_shares_%s_%d_%d", name, page, count)
	if val, exists := cache.Get(cacheKey); exists {
		if shares, ok := val.([]map[string]interface{}); ok {
			return shares
		}
	}

	// 构建参数
	params := fmt.Sprintf("name=%s&page=%d&count=%d", name, page, count)

	// 发送请�?	httpClient := utils.NewAsyncRequestUtils("", "", "", config.GlobalConfig.Proxy, "", 15*time.Second)
	res, err := httpClient.GetRes(fmt.Sprintf("%s?%s", w.workflowShares, params))
	if err != nil {
		return []map[string]interface{}{}
	}
	defer res.Body.Close()

	shares := w.handleListResponse(res)
	
	// 缓存结果
	cache.Set(cacheKey, shares, 1*time.Hour)
	return shares
}

// GetUserUUID 获取用户uuid
func (w *WorkflowHelper) GetUserUUID() string {
	if w.shareUserID == "" {
		w.shareUserID = utils.SystemUtils{}.GenerateUserUniqueID()
		logger.Infof("当前用户UUID: %s", w.shareUserID)
	}
	return w.shareUserID
}
