package actions_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"moviepilot-go/internal/actions"
	"moviepilot-go/internal/business/storage"
)

// mockStorageService 模拟存储服务
type mockStorageService struct {
	scanFunc func(opts storage.ScanOptions) ([]storage.FileItem, error)
}

func (m *mockStorageService) Scan(opts storage.ScanOptions) ([]storage.FileItem, error) {
	if m.scanFunc != nil {
		return m.scanFunc(opts)
	}
	return []storage.FileItem{}, nil
}

func (m *mockStorageService) Transfer(tasks []storage.TransferTask) ([]storage.TransferResult, error) {
	return []storage.TransferResult{}, nil
}

func TestScanFileAction_Execute(t *testing.T) {
	logger := zap.NewNop()
	
	// 创建 mock storage service
	mockStorageService := &mockStorageService{
		scanFunc: func(opts storage.ScanOptions) ([]storage.FileItem, error) {
			return []storage.FileItem{
				{Path: "/test/movie1.mkv", Size: 1024 * 1024 * 1024, ModTime: time.Now(), IsDir: false},
				{Path: "/test/movie2.mkv", Size: 2 * 1024 * 1024 * 1024, ModTime: time.Now(), IsDir: false},
			}, nil
		},
	}

	action := actions.NewScanFileAction(logger, mockStorageService)

	tests := []struct {
		name          string
		params        actions.ScanFileParams
		expectedError string
		expectedCount int
	}{
		{
			name: "successful scan",
			params: actions.ScanFileParams{
				BaseParams: actions.BaseParams{},
				RootPath:   "/test",
				Include:        []string{"*.mkv"},
				Exclude:        []string{},
				MaxDepth:       1,
				FollowSymlinks: false,
			},
			expectedCount: 2,
		},
		{
			name: "invalid path",
			params: actions.ScanFileParams{
				BaseParams: actions.BaseParams{},
				RootPath:   "",
				Include:        []string{"*.mkv"},
				Exclude:        []string{},
				MaxDepth:       1,
				FollowSymlinks: false,
			},
			expectedError: "root_path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := action.Execute(1, tt.params, nil)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			
			// 验证结果
			assert.Equal(t, true, action.Success())
			assert.GreaterOrEqual(t, len(result.Files), tt.expectedCount)
		})
	}
}

func TestScanFileAction_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "scan_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFiles := []string{"movie1.mkv", "movie2.mp4", "subtitle.srt"}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// 创建子目录
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	subFile := filepath.Join(subDir, "movie3.mkv")
	err = os.WriteFile(subFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// 创建真实的 storage service
	mockStorageService := &mockStorageService{
		scanFunc: func(opts storage.ScanOptions) ([]storage.FileItem, error) {
			var items []storage.FileItem
			
			err := filepath.Walk(opts.RootPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				
				// 跳过目录本身，只包含文件
				if info.IsDir() {
					return nil
				}
				
				// 检查深度
				relPath, err := filepath.Rel(opts.RootPath, path)
				if err != nil {
					return err
				}
				
				depth := len(filepath.SplitList(relPath))
				if opts.MaxDepth > 0 && depth > opts.MaxDepth {
					return nil
				}
				
				items = append(items, storage.FileItem{
					Path:    path,
					Size:    info.Size(),
					ModTime: info.ModTime(),
					IsDir:   false,
				})
				
				return nil
			})
			
			return items, err
		},
	}

	action := actions.NewScanFileAction(logger, mockStorageService)

	// 测试扫描
	params := actions.ScanFileParams{
		BaseParams:     actions.BaseParams{},
		RootPath:       tempDir,
		Include:        []string{"*.mkv"},
		Exclude:        []string{},
		MaxDepth:       2,
		FollowSymlinks: false,
	}

	result, err := action.Execute(1, params, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证结果
	assert.Equal(t, true, action.Success())
	assert.GreaterOrEqual(t, len(result.Files), 2) // 至少有2个mkv文件

	// 验证文件信息
	for _, file := range result.Files {
		assert.NotEmpty(t, file.Path)
		assert.Greater(t, file.Size, int64(0))
		assert.False(t, file.IsDir)
	}
}