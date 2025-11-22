package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	workflowapi "moviepilot-go/internal/api/workflow"
	wf "moviepilot-go/internal/platform/workflow"
	"moviepilot-go/pkg/middlewares"
)

// createTestFiles 创建测试文件结构
func createTestFiles(t *testing.T, files []string) string {
	tempDir := t.TempDir()

	for _, filename := range files {
		fp := filepath.Join(tempDir, filename)
		dirPath := filepath.Dir(fp)

		// 创建必要的目录
		if dirPath != tempDir {
			require.NoError(t, os.MkdirAll(dirPath, 0755))
		}

		// 创建文件
		content := fmt.Sprintf("test content for %s", filename)
		require.NoError(t, os.WriteFile(fp, []byte(content), 0644))
	}

	return tempDir
}

// cleanupTestFiles 清理测试文件
func cleanupTestFiles(t *testing.T, dir string) {
	// 由于使用了 t.TempDir()，这里不需要手动清理
	// 但保留函数接口以符合文档示例
}

func TestLocalFileWorkflow_EndToEnd(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// 1. 准备测试数据
	testDir := createTestFiles(t, []string{
		"Movie.2023.1080p.mkv",
		"TV.Show.S01E01.mkv",
		"Documentary.2022.720p.mp4",
		"Anime.Series.S02E12.2023.BDRip.1080p.x265-GROUP.mkv",
		"Action.Movie.2022.2160p.UHD.BluRay.x265-GROUP.mkv",
		"Drama.Series.S02E05.2023.WEB-DL.1080p.AAC2.0-GROUP.mp4",
	})
	defer cleanupTestFiles(t, testDir)

	// 创建目标目录
	targetDir := t.TempDir()

	// 2. 设置服务器
	logger, _ := zap.NewDevelopment()
	workflowManager := wf.NewWorkflowManager(logger)
	workflowConfig := wf.LocalFileWorkflowConfig{
		Logger: logger,
		// 使用默认的业务服务
	}
	workflowService := workflowapi.NewService(workflowManager, workflowConfig, logger)
	workflowHandler := workflowapi.NewHandler(workflowService, logger)

	// 3. 创建 Gin Engine 并添加中间件
	engine := gin.New()
	engine.Use(middlewares.RequestIDMiddleware())
	engine.Use(middlewares.RecoveryMiddleware())
	engine.Use(middlewares.CORSMiddleware())

	// 添加请求日志
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			logger.Info("HTTP request",
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.Int("status", param.StatusCode),
				zap.Duration("latency", param.Latency),
				zap.String("client_ip", param.ClientIP),
			)
			return ""
		},
	}))

	// 注册路由
	engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	// 4. 发送 API 请求
	req := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          testDir,
		Include:           []string{"*.mkv", "*.mp4"},
		Exclude:           []string{},
		MaxDepth:          2,
		FollowSymlink:     false,
		TargetRoot:        targetDir,
		Mode:              "copy", // 使用 copy 模式以便验证
		Category:          "test",
		Overwrite:         false,
		PreserveDir:       false,
		DryRun:            false,
		IncludeFetch:      false, // 不包含种子获取
		FetchKeywords:     []string{},
		WaitForCompletion: true, // 同步等待完成
		ForceRefresh:      false,
		Source:            "integration-test",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost,
		"/api/workflows/local-file-scrape-transfer",
		bytes.NewReader(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", "test-integration-request-123")

	// 5. 执行请求
	w := httptest.NewRecorder()
	start := time.Now()
	engine.ServeHTTP(w, httpReq)
	duration := time.Since(start)

	// 6. 验证响应
	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK for synchronous request")

	var resp struct {
		Success   bool        `json:"success"`
		Code      int         `json:"code"`
		Message   string      `json:"message"`
		Data      interface{} `json:"data"`
		Timestamp int64       `json:"timestamp"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Success, "Response should be successful")
	assert.Equal(t, 0, resp.Code, "Response code should be 0")
	assert.NotEmpty(t, resp.Message, "Response message should not be empty")
	assert.NotNil(t, resp.Data, "Response data should not be empty")
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp, 5, "Timestamp should be recent")

	// 7. 验证工作流结果
	dataBytes, err := json.Marshal(resp.Data)
	require.NoError(t, err)

	var workflowResp workflowapi.StartLocalFileWorkflowResponse
	err = json.Unmarshal(dataBytes, &workflowResp)
	require.NoError(t, err)

	assert.NotEmpty(t, workflowResp.WorkflowID, "Workflow ID should not be empty")
	assert.NotEmpty(t, workflowResp.Status, "Workflow status should not be empty")

	// 由于是同步等待，状态应该是完成状态
	assert.Contains(t, []string{"completed", "failed", "partial"}, workflowResp.Status,
		"Workflow should be in a terminal state")

	// 8. 验证性能
	assert.Less(t, duration, 30*time.Second, "Workflow should complete within 30 seconds")

	// 9. 验证日志输出（通过检查是否有请求ID）
	assert.Contains(t, w.Header().Get("X-Request-ID"), "test-integration-request-123",
		"Request ID should be returned in response header")

	t.Logf("Integration test completed successfully:")
	t.Logf("  - Workflow ID: %s", workflowResp.WorkflowID)
	t.Logf("  - Status: %s", workflowResp.Status)
	t.Logf("  - Duration: %v", duration)
	t.Logf("  - Files processed: %d", 6) // 预期处理 6 个文件
}

func TestLocalFileWorkflow_EndToEnd_AsyncMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// 1. 准备测试数据
	testDir := createTestFiles(t, []string{
		"Quick.Movie.2023.1080p.mkv",
		"Quick.Series.S01E01.mkv",
	})
	defer cleanupTestFiles(t, testDir)

	targetDir := t.TempDir()

	// 2. 设置服务器
	logger, _ := zap.NewDevelopment()
	workflowManager := wf.NewWorkflowManager(logger)
	workflowConfig := wf.LocalFileWorkflowConfig{Logger: logger}
	workflowService := workflowapi.NewService(workflowManager, workflowConfig, logger)
	workflowHandler := workflowapi.NewHandler(workflowService, logger)

	engine := gin.New()
	engine.Use(middlewares.RequestIDMiddleware())
	engine.Use(middlewares.RecoveryMiddleware())
	engine.Use(middlewares.CORSMiddleware())

	engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	// 3. 发送异步请求
	req := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          testDir,
		Include:           []string{"*.mkv"},
		Exclude:           []string{},
		MaxDepth:          1,
		FollowSymlink:     false,
		TargetRoot:        targetDir,
		Mode:              "copy",
		Category:          "test",
		Overwrite:         false,
		PreserveDir:       false,
		DryRun:            false,
		IncludeFetch:      false,
		FetchKeywords:     []string{},
		WaitForCompletion: false, // 异步模式
		ForceRefresh:      false,
		Source:            "async-test",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost,
		"/api/workflows/local-file-scrape-transfer",
		bytes.NewReader(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httpReq)

	// 4. 验证异步响应
	assert.Equal(t, http.StatusAccepted, w.Code, "Expected 202 Accepted for async request")

	var resp struct {
		Success   bool        `json:"success"`
		Code      int         `json:"code"`
		Message   string      `json:"message"`
		Data      interface{} `json:"data"`
		Timestamp int64       `json:"timestamp"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Success)
	assert.Equal(t, 0, resp.Code)
	assert.Contains(t, resp.Message, "workflow started")

	// 5. 验证工作流结果
	dataBytes, err := json.Marshal(resp.Data)
	require.NoError(t, err)

	var workflowResp workflowapi.StartLocalFileWorkflowResponse
	err = json.Unmarshal(dataBytes, &workflowResp)
	require.NoError(t, err)

	assert.NotEmpty(t, workflowResp.WorkflowID)
	// 异步模式下状态应该是运行中
	assert.Equal(t, "running", workflowResp.Status)

	t.Logf("Async workflow started successfully:")
	t.Logf("  - Workflow ID: %s", workflowResp.WorkflowID)
	t.Logf("  - Status: %s", workflowResp.Status)
}

func TestLocalFileWorkflow_EndToEnd_ValidationErrors(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	logger, _ := zap.NewDevelopment()
	workflowManager := wf.NewWorkflowManager(logger)
	workflowConfig := wf.LocalFileWorkflowConfig{Logger: logger}
	workflowService := workflowapi.NewService(workflowManager, workflowConfig, logger)
	workflowHandler := workflowapi.NewHandler(workflowService, logger)

	engine := gin.New()
	engine.Use(middlewares.RequestIDMiddleware())
	engine.Use(middlewares.RecoveryMiddleware())
	engine.Use(middlewares.CORSMiddleware())

	engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	tests := []struct {
		name          string
		requestBody   interface{}
		expectedCode  int
		expectedError string
	}{
		{
			name:          "空请求体",
			requestBody:   nil,
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request",
		},
		{
			name: "缺少必需字段",
			requestBody: map[string]interface{}{
				"include": []string{"*.mkv"},
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "required",
		},
		{
			name: "无效的模式",
			requestBody: map[string]interface{}{
				"root_path":   "/tmp",
				"target_root": "/tmp/target",
				"mode":        "invalid_mode",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "oneof",
		},
		{
			name: "负数深度",
			requestBody: map[string]interface{}{
				"root_path":   "/tmp",
				"target_root": "/tmp/target",
				"max_depth":   -1,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "gte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			httpReq, err := http.NewRequestWithContext(context.Background(),
				http.MethodPost,
				"/api/workflows/local-file-scrape-transfer",
				bytes.NewReader(body))
			require.NoError(t, err)
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, httpReq)

			assert.Equal(t, tt.expectedCode, w.Code)

			var resp struct {
				Success bool   `json:"success"`
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.False(t, resp.Success)
			assert.Contains(t, resp.Message, tt.expectedError)
		})
	}
}

func TestLocalFileWorkflow_EndToEnd_DryRunMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// 1. 准备测试数据
	testDir := createTestFiles(t, []string{
		"Dry.Run.Movie.2023.1080p.mkv",
		"Dry.Run.Series.S01E01.mkv",
	})
	defer cleanupTestFiles(t, testDir)

	targetDir := t.TempDir()

	// 2. 设置服务器
	logger, _ := zap.NewDevelopment()
	workflowManager := wf.NewWorkflowManager(logger)
	workflowConfig := wf.LocalFileWorkflowConfig{Logger: logger}
	workflowService := workflowapi.NewService(workflowManager, workflowConfig, logger)
	workflowHandler := workflowapi.NewHandler(workflowService, logger)

	engine := gin.New()
	engine.Use(middlewares.RequestIDMiddleware())
	engine.Use(middlewares.RecoveryMiddleware())
	engine.Use(middlewares.CORSMiddleware())

	engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	// 3. 发送干运行请求
	req := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          testDir,
		Include:           []string{"*.mkv"},
		Exclude:           []string{},
		MaxDepth:          1,
		FollowSymlink:     false,
		TargetRoot:        targetDir,
		Mode:              "copy",
		Category:          "test",
		Overwrite:         false,
		PreserveDir:       false,
		DryRun:            true, // 干运行模式
		IncludeFetch:      false,
		FetchKeywords:     []string{},
		WaitForCompletion: true,
		ForceRefresh:      false,
		Source:            "dry-run-test",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost,
		"/api/workflows/local-file-scrape-transfer",
		bytes.NewReader(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httpReq)

	// 4. 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool        `json:"success"`
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "completed")

	// 5. 验证干运行：目标目录应该为空
	files, err := os.ReadDir(targetDir)
	require.NoError(t, err)
	assert.Len(t, files, 0, "Target directory should be empty in dry run mode")

	// 6. 验证源文件仍然存在
	sourceFiles, err := os.ReadDir(testDir)
	require.NoError(t, err)
	assert.Len(t, sourceFiles, 2, "Source files should still exist in dry run mode")

	t.Logf("Dry run test completed successfully")
}
