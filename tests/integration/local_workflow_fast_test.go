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

// TestLocalFileWorkflow_Fast 快速集成测试，解决超时问题
func TestLocalFileWorkflow_Fast(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// 1. 准备少量测试数据（减少文件数量）
	testDir := createTestFilesFast(t, []string{
		"Movie.2023.1080p.mkv", // 1个电影文件
		"TV.Show.S01E01.mkv",   // 1个电视剧文件
	})
	defer cleanupTestFilesFast(t, testDir)

	targetDir := t.TempDir()

	// 2. 设置服务器
	logger, _ := zap.NewDevelopment()
	workflowManager := wf.NewWorkflowManager(logger)
	workflowConfig := wf.LocalFileWorkflowConfig{
		Logger: logger,
	}
	workflowService := workflowapi.NewService(workflowManager, workflowConfig, logger)
	workflowHandler := workflowapi.NewHandler(workflowService, logger)

	// 3. 创建 Gin Engine（最小中间件）
	engine := gin.New()
	engine.Use(middlewares.RequestIDMiddleware())
	engine.Use(middlewares.RecoveryMiddleware())

	// 注册路由
	engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	// 4. 发送简化的API请求
	req := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          testDir,
		Include:           []string{"*.mkv"},
		Exclude:           []string{},
		MaxDepth:          1, // 限制深度
		FollowSymlink:     false,
		TargetRoot:        targetDir,
		Mode:              "copy",
		Category:          "test",
		Overwrite:         false,
		PreserveDir:       false,
		DryRun:            true,  // 使用干运行模式，避免实际文件操作
		IncludeFetch:      false, // 跳过种子获取
		FetchKeywords:     []string{},
		WaitForCompletion: true, // 同步等待但时间短
		ForceRefresh:      false,
		Source:            "fast-test",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequestWithContext(
		context.WithValue(context.Background(), "timeout", 10*time.Second),
		http.MethodPost,
		"/api/workflows/local-file-scrape-transfer",
		bytes.NewReader(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", "fast-test-123")

	// 5. 执行请求（设置超时）
	w := httptest.NewRecorder()
	start := time.Now()

	// 使用带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	engine.ServeHTTP(w, httpReq)
	duration := time.Since(start)

	// 6. 验证响应
	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK")

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

	// 7. 验证性能（快速测试应该在5秒内完成）
	assert.Less(t, duration, 5*time.Second, "Fast workflow should complete within 5 seconds")
	assert.Greater(t, duration, 100*time.Millisecond, "Workflow should take some time to process")

	t.Logf("Fast integration test completed successfully:")
	t.Logf("  - Duration: %v", duration)
	t.Logf("  - Files processed: 2")
}

// createTestFilesFast 创建快速测试的少量文件
func createTestFilesFast(t *testing.T, files []string) string {
	tempDir := t.TempDir()

	for _, filename := range files {
		fp := filepath.Join(tempDir, filename)
		dirPath := filepath.Dir(fp)

		// 创建必要的目录
		if dirPath != tempDir {
			require.NoError(t, os.MkdirAll(dirPath, 0755))
		}

		// 创建小文件（减少IO时间）
		content := fmt.Sprintf("fast test content for %s", filename)
		require.NoError(t, os.WriteFile(fp, []byte(content), 0644))
	}

	return tempDir
}

// cleanupTestFilesFast 清理快速测试文件
func cleanupTestFilesFast(t *testing.T, dir string) {
	// 由于使用了 t.TempDir()，这里不需要手动清理
}

// TestLocalFileWorkflow_Fast_AsyncMode 快速异步模式测试
func TestLocalFileWorkflow_Fast_AsyncMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// 准备测试数据
	testDir := createTestFilesFast(t, []string{"Async.Movie.2023.mkv"})
	defer cleanupTestFilesFast(t, testDir)

	targetDir := t.TempDir()

	// 设置服务器
	logger, _ := zap.NewDevelopment()
	workflowManager := wf.NewWorkflowManager(logger)
	workflowConfig := wf.LocalFileWorkflowConfig{
		Logger: logger,
	}
	workflowService := workflowapi.NewService(workflowManager, workflowConfig, logger)
	workflowHandler := workflowapi.NewHandler(workflowService, logger)

	engine := gin.New()
	engine.Use(middlewares.RequestIDMiddleware())
	engine.Use(middlewares.RecoveryMiddleware())
	engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	// 发送异步请求
	req := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          testDir,
		Include:           []string{"*.mkv"},
		MaxDepth:          1,
		TargetRoot:        targetDir,
		Mode:              "copy",
		DryRun:            true,
		WaitForCompletion: false, // 异步模式
		IncludeFetch:      false,
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/workflows/local-file-scrape-transfer",
		bytes.NewReader(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	start := time.Now()
	engine.ServeHTTP(w, httpReq)
	duration := time.Since(start)

	// 验证异步响应
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Less(t, duration, 1*time.Second, "Async response should be very fast")

	var resp struct {
		Success bool        `json:"success"`
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Success)
	assert.Equal(t, 0, resp.Code)

	t.Logf("Fast async test completed in %v", duration)
}

// TestLocalFileWorkflow_Fast_Validation 快速参数验证测试
func TestLocalFileWorkflow_Fast_Validation(t *testing.T) {
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
	engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	// 测试用例：只测试最常见的错误
	testCases := []struct {
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			var err error

			if tc.requestBody != nil {
				body, err = json.Marshal(tc.requestBody)
				require.NoError(t, err)
			}

			httpReq, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/api/workflows/local-file-scrape-transfer",
				bytes.NewReader(body))
			require.NoError(t, err)
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			start := time.Now()
			engine.ServeHTTP(w, httpReq)
			duration := time.Since(start)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.Less(t, duration, 100*time.Millisecond, "Validation should be very fast")

			var resp struct {
				Success bool   `json:"success"`
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.False(t, resp.Success)
			assert.Contains(t, resp.Message, tc.expectedError)
		})
	}
}
