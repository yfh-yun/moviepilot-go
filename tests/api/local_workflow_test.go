package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	workflowapi "moviepilot-go/internal/api/workflow"
	wf "moviepilot-go/internal/platform/workflow"
)

// 集成测试：通过 HTTP API 触发本地文件工作流
func TestStartLocalFileWorkflow_API(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// === 1. 组装依赖（真实 WorkflowManager + 占位业务服务） ===
	logger, _ := zap.NewDevelopment()
	manager := wf.NewWorkflowManager(logger)

	cfg := wf.LocalFileWorkflowConfig{
		Logger: logger,
		// StorageService / MediaService / TransferService 为 nil，
		// 由 NewService 内部使用占位实现自动填充。
	}
	svc := workflowapi.NewService(manager, cfg, logger)
	handler := workflowapi.NewHandler(svc, logger)

	// === 2. 构建 Gin Engine 并注册路由 ===
	r := gin.New()
	r.POST("/api/workflows/local-file-scrape-transfer", handler.StartLocalFileWorkflow)

	// === 3. 构造请求体 ===
	body := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          ".",              // 当前目录，做一个最小可行扫描
		Include:           []string{"*.go"}, // 只扫描 go 文件
		Exclude:           nil,
		MaxDepth:          1,
		FollowSymlink:     false,
		TargetRoot:        "./output", // 占位路径，Transfer 默认是“planned”
		Mode:              "copy",
		Category:          "test",
		Overwrite:         false,
		PreserveDir:       false,
		DryRun:            true,  // 建议测试时先 dry-run
		IncludeFetch:      false, // 先不包含 FetchTorrents
		FetchKeywords:     nil,
		ForceRefresh:      false,
		Source:            "",
		WaitForCompletion: true, // 测试同步等待完整结果
	}

	b, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost,
		"/api/workflows/local-file-scrape-transfer",
		bytes.NewReader(b),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// === 4. 发送请求 ===
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// === 5. 断言响应 ===
	require.Equal(t, http.StatusAccepted, w.Code)

	var resp struct {
		Success   bool        `json:"success"`
		Code      int         `json:"code"`
		Message   string      `json:"message"`
		Data      interface{} `json:"data"`
		Timestamp int64       `json:"timestamp"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.True(t, resp.Success)
	require.Equal(t, 0, resp.Code)
	require.Contains(t, resp.Message, "workflow") // 粗略检查一下文案

	// 进一步可以把 Data 反序列化为 StartLocalFileWorkflowResponse
	dataBytes, err := json.Marshal(resp.Data)
	require.NoError(t, err)

	var wfResp workflowapi.StartLocalFileWorkflowResponse
	err = json.Unmarshal(dataBytes, &wfResp)
	require.NoError(t, err)

	require.NotEmpty(t, wfResp.WorkflowID)
	// WaitForCompletion = true 时，status 应该是已完成（TaskCompleted 或 Failed）
	require.NotEmpty(t, wfResp.Status)
	t.Logf("workflow_id=%s status=%s", wfResp.WorkflowID, wfResp.Status)
	// 如果需要，还可以断言 wfResp.Result 的结构
	_ = wfResp.Result

	// 时间戳大致合理
	require.InDelta(t, time.Now().Unix(), resp.Timestamp, 5)
}
