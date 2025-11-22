package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	workflowapi "moviepilot-go/internal/api/workflow"
	"moviepilot-go/pkg/response"
)

// MockWorkflowService 是 WorkflowService 的 mock 实现
type MockWorkflowService struct {
	mock.Mock
}

func (m *MockWorkflowService) StartLocalFileWorkflow(ctx context.Context, req workflowapi.StartLocalFileWorkflowRequest) (*workflowapi.StartLocalFileWorkflowResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workflowapi.StartLocalFileWorkflowResponse), args.Error(1)
}

func TestWorkflowHandler_StartLocalFileWorkflow_ValidationErrors(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		requestBody   interface{}
		expectedCode  int
		expectedError string
	}{
		{
			name:          "empty body",
			requestBody:   nil,
			expectedCode:  http.StatusBadRequest,
			expectedError: "请求数据绑定失败",
		},
		{
			name: "missing root_path",
			requestBody: map[string]interface{}{
				"target_root": "/tmp/target",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "required",
		},
		{
			name: "missing target_root",
			requestBody: map[string]interface{}{
				"root_path": "/tmp/source",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "required",
		},
		{
			name: "invalid mode",
			requestBody: map[string]interface{}{
				"root_path":   "/tmp/source",
				"target_root": "/tmp/target",
				"mode":        "invalid_mode",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "oneof",
		},
		{
			name: "negative max_depth",
			requestBody: map[string]interface{}{
				"root_path":   "/tmp/source",
				"target_root": "/tmp/target",
				"max_depth":   -1,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "gte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockWorkflowService)
			logger, _ := zap.NewDevelopment()
			handler := workflowapi.NewHandler(mockService, logger)

			r := gin.New()
			r.POST("/test", handler.StartLocalFileWorkflow)

			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/test", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			var resp response.Response
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Message, tt.expectedError)
		})
	}
}

func TestWorkflowHandler_StartLocalFileWorkflow_ServiceError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockService := new(MockWorkflowService)
	logger, _ := zap.NewDevelopment()
	handler := workflowapi.NewHandler(mockService, logger)

	r := gin.New()
	r.POST("/test", handler.StartLocalFileWorkflow)

	// 模拟服务返回错误
	expectedError := errors.New("service error")
	mockService.On("StartLocalFileWorkflow", mock.Anything, mock.Anything).Return(nil, expectedError)

	body := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:   "/tmp/source",
		TargetRoot: "/tmp/target",
	}

	reqBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/test", bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, response.CodeServerError, resp.Code)

	mockService.AssertExpectations(t)
}

func TestWorkflowHandler_StartLocalFileWorkflow_AsyncMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockService := new(MockWorkflowService)
	logger, _ := zap.NewDevelopment()
	handler := workflowapi.NewHandler(mockService, logger)

	r := gin.New()
	r.POST("/test", handler.StartLocalFileWorkflow)

	// 模拟异步响应
	expectedResponse := &workflowapi.StartLocalFileWorkflowResponse{
		WorkflowID: "test-workflow-id",
		Status:     "running",
		Message:    "workflow started",
	}
	mockService.On("StartLocalFileWorkflow", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	body := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          "/tmp/source",
		TargetRoot:        "/tmp/target",
		WaitForCompletion: false, // 异步模式
	}

	reqBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/test", bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, response.CodeSuccess, resp.Code)
	assert.Contains(t, resp.Message, "workflow started")

	mockService.AssertExpectations(t)
}

func TestWorkflowHandler_StartLocalFileWorkflow_SyncMode(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockService := new(MockWorkflowService)
	logger, _ := zap.NewDevelopment()
	handler := workflowapi.NewHandler(mockService, logger)

	r := gin.New()
	r.POST("/test", handler.StartLocalFileWorkflow)

	// 模拟同步响应
	expectedResponse := &workflowapi.StartLocalFileWorkflowResponse{
		WorkflowID: "test-workflow-id",
		Status:     "completed",
		Message:    "workflow completed",
		Result:     map[string]interface{}{"processed": 10},
	}
	mockService.On("StartLocalFileWorkflow", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	body := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          "/tmp/source",
		TargetRoot:        "/tmp/target",
		WaitForCompletion: true, // 同步模式
	}

	reqBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/test", bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, response.CodeSuccess, resp.Code)
	assert.Contains(t, resp.Message, "workflow completed")

	mockService.AssertExpectations(t)
}

func TestWorkflowHandler_StartLocalFileWorkflow_SyncModeCustomMessage(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockService := new(MockWorkflowService)
	logger, _ := zap.NewDevelopment()
	handler := workflowapi.NewHandler(mockService, logger)

	r := gin.New()
	r.POST("/test", handler.StartLocalFileWorkflow)

	// 模拟同步响应，带有自定义消息
	expectedResponse := &workflowapi.StartLocalFileWorkflowResponse{
		WorkflowID: "test-workflow-id",
		Status:     "failed",
		Message:    "custom error message",
	}
	mockService.On("StartLocalFileWorkflow", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	body := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          "/tmp/source",
		TargetRoot:        "/tmp/target",
		WaitForCompletion: true,
	}

	reqBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/test", bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, response.CodeSuccess, resp.Code)
	assert.Equal(t, "custom error message", resp.Message) // 应该使用服务返回的自定义消息

	mockService.AssertExpectations(t)
}

func TestWorkflowHandler_StartLocalFileWorkflow_ValidParameters(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockService := new(MockWorkflowService)
	logger, _ := zap.NewDevelopment()
	handler := workflowapi.NewHandler(mockService, logger)

	r := gin.New()
	r.POST("/test", handler.StartLocalFileWorkflow)

	expectedResponse := &workflowapi.StartLocalFileWorkflowResponse{
		WorkflowID: "test-workflow-id",
		Status:     "running",
	}
	mockService.On("StartLocalFileWorkflow", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	// 测试所有有效参数
	body := workflowapi.StartLocalFileWorkflowRequest{
		RootPath:          "/tmp/source",
		Include:           []string{"*.mkv", "*.mp4"},
		Exclude:           []string{"*.txt", "*.nfo"},
		MaxDepth:          3,
		FollowSymlink:     true,
		TargetRoot:        "/tmp/target",
		Mode:              "copy",
		Category:          "movies",
		Overwrite:         true,
		PreserveDir:       false,
		DryRun:            true,
		IncludeFetch:      true,
		FetchKeywords:     []string{"movie", "2023"},
		WaitForCompletion: false,
		ForceRefresh:      true,
		Source:            "test",
	}

	reqBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/test", bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// 验证时间戳在合理范围内
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp, 5)

	mockService.AssertExpectations(t)
}
