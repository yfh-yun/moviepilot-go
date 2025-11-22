package business_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"moviepilot-go/internal/business/storage"
)

func TestStorageService_Scan(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	service := storage.NewLocalService(logger)

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		params        storage.ScanOptions
		expectedCount int
		expectedError string
	}{
		{
			name: "扫描包含多种文件的目录",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()

				files := map[string][]byte{
					"Movie.2023.1080p.mkv":      []byte("movie content"),
					"TV.Show.S01E01.mkv":        []byte("tv episode content"),
					"Documentary.2022.720p.mp4": []byte("documentary content"),
					"subtitle.srt":              []byte("subtitle content"),
					"poster.jpg":                []byte("image content"),
					"readme.txt":                []byte("text content"),
				}

				for filename, content := range files {
					fp := filepath.Join(tempDir, filename)
					require.NoError(t, os.WriteFile(fp, content, 0644))
				}

				return tempDir
			},
			params: storage.ScanOptions{
				RootPath:       "", // 将在 setupFunc 中设置
				Include:        []string{"*.mkv", "*.mp4"},
				Exclude:        []string{},
				MaxDepth:       1,
				FollowSymlinks: false,
			},
			expectedCount: 3, // 只有视频文件
		},
		{
			name: "使用排除规则",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()

				files := map[string][]byte{
					"Movie.2023.1080p.mkv":       []byte("movie content"),
					"Movie.2023.sample.mkv":      []byte("sample content"),
					"TV.Show.S01E01.mkv":         []byte("tv episode content"),
					"TV.Show.S01E01.Trailer.mkv": []byte("trailer content"),
				}

				for filename, content := range files {
					fp := filepath.Join(tempDir, filename)
					require.NoError(t, os.WriteFile(fp, content, 0644))
				}

				return tempDir
			},
			params: storage.ScanOptions{
				RootPath:       "", // 将在 setupFunc 中设置
				Include:        []string{"*.mkv"},
				Exclude:        []string{"*sample*", "*Trailer*"},
				MaxDepth:       1,
				FollowSymlinks: false,
			},
			expectedCount: 2, // 排除 sample 和 trailer 文件
		},
		{
			name: "深度扫描",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()

				// 创建多层目录结构
				structure := map[string][]byte{
					"root.mkv":                             []byte("root level"),
					"level1/sub1.mkv":                      []byte("level 1"),
					"level1/level2/sub2.mkv":               []byte("level 2"),
					"level1/level2/level3/sub3.mkv":        []byte("level 3"),
					"level1/level2/level3/level4/sub4.mkv": []byte("level 4"),
				}

				for relPath, content := range structure {
					fullPath := filepath.Join(tempDir, relPath)
					dirPath := filepath.Dir(fullPath)
					require.NoError(t, os.MkdirAll(dirPath, 0755))
					require.NoError(t, os.WriteFile(fullPath, content, 0644))
				}

				return tempDir
			},
			params: storage.ScanOptions{
				RootPath:       "", // 将在 setupFunc 中设置
				Include:        []string{"*.mkv"},
				Exclude:        []string{},
				MaxDepth:       3, // 最多扫描 3 层
				FollowSymlinks: false,
			},
			expectedCount: 3, // root, level1, level2, level3（不包括 level4，深度计算从根目录开始）
		},
		{
			name: "空目录",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			params: storage.ScanOptions{
				RootPath:       "", // 将在 setupFunc 中设置
				Include:        []string{"*"},
				Exclude:        []string{},
				MaxDepth:       1,
				FollowSymlinks: false,
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置测试目录
			testDir := tt.setupFunc(t)
			tt.params.RootPath = testDir

			result, err := service.Scan(tt.params)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Len(t, result, tt.expectedCount)

			// 验证每个文件项的基本信息
			for _, file := range result {
				assert.NotEmpty(t, file.Path)
				assert.Greater(t, file.Size, int64(0))
				assert.NotZero(t, file.ModTime)
				assert.True(t, filepath.IsAbs(file.Path))
			}
		})
	}
}

func TestStorageService_Scan_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	service := storage.NewLocalService(logger)

	params := storage.ScanOptions{
		RootPath:       "/non/existent/directory",
		Include:        []string{"*"},
		Exclude:        []string{},
		MaxDepth:       1,
		FollowSymlinks: false,
	}

	result, err := service.Scan(params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestStorageService_Scan_InvalidParameters(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	service := storage.NewLocalService(logger)

	tests := []struct {
		name          string
		params        storage.ScanOptions
		expectedError string
	}{
		{
			name: "空根路径",
			params: storage.ScanOptions{
				RootPath:       "",
				Include:        []string{"*"},
				Exclude:        []string{},
				MaxDepth:       1,
				FollowSymlinks: false,
			},
			expectedError: "root path is required",
		},
		{
			name: "负数深度",
			params: storage.ScanOptions{
				RootPath:       "/tmp",
				Include:        []string{"*"},
				Exclude:        []string{},
				MaxDepth:       -1,
				FollowSymlinks: false,
			},
			expectedError: "", // 负数深度在实现中被忽略，不会报错
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Scan(tt.params)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				// 对于负数深度，应该正常执行（被忽略）
				if tt.params.MaxDepth < 0 {
					assert.NoError(t, err)
				}
			}
		})
	}
}