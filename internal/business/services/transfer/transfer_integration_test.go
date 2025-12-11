package transfer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/storage"
	"moviepilot-go/internal/models/database"
)

// TestTransferServiceIntegration 集成测试
func TestTransferServiceIntegration(t *testing.T) {
	// 跳过集成测试（需要实际文件系统）
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	storageSvc := storage.NewStorageService(nil, logger)
	service := NewDefaultService(storageSvc, logger)

	t.Run("Execute transfer tasks", func(t *testing.T) {
		// 创建临时测试目录
		tmpDir := t.TempDir()
		sourceDir := filepath.Join(tmpDir, "source")
		targetDir := filepath.Join(tmpDir, "target")

		require.NoError(t, os.MkdirAll(sourceDir, 0755))
		require.NoError(t, os.MkdirAll(targetDir, 0755))

		// 创建测试文件
		testFile := filepath.Join(sourceDir, "test.mkv")
		require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

		// 准备转移任务
		season := 1
		episode := 1
		year := "2024"
		tasks := []Task{
			{
				Media: database.Media{
					Type:    "tv",
					Title:   "Test Show",
					Year:    &year,
					Season:  &season,
					Episode: &episode,
				},
				SourcePath: testFile,
				TargetPath: filepath.Join(targetDir, "Test Show", "Season 01", "Test Show - S01E01.mkv"),
				Mode:       storage.TransferModeCopy,
				Overwrite:  false,
				Category:   "tv",
			},
		}

		// 执行转移
		histories, err := service.Execute(tasks)
		require.NoError(t, err)
		assert.Len(t, histories, 1)
		assert.Equal(t, "Test Show", histories[0].Title)
		assert.Equal(t, "S01", histories[0].Seasons)
		assert.Equal(t, "E01", histories[0].Episodes)
	})
}

// TestNamingStrategy 命名策略测试
func TestNamingStrategy(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Generate path with template", func(t *testing.T) {
		// 测试模板命名策略
		strategy, err := NewTemplateNamingStrategy("{title}/Season {season}/{title} - S{season:02d}E{episode:02d}", logger)
		require.NoError(t, err)
		assert.NotNil(t, strategy)

		season := 1
		episode := 5
		year := "2024"
		media := &database.Media{
			Type:    "tv",
			Title:   "Test Series",
			Year:    &year,
			Season:  &season,
			Episode: &episode,
		}

		metadata := FileMetadata{
			Resolution: "1080p",
			Source:     "BluRay",
			Codec:      "H265",
		}

		path, err := strategy.GeneratePath(media, "/source/test.mkv", metadata)
		require.NoError(t, err)
		assert.Contains(t, path, "Test Series")
		assert.Contains(t, path, "Season 1")
		assert.Contains(t, path, "S01E05")
	})
}

// BenchmarkTransferService 性能基准测试
func BenchmarkTransferService(b *testing.B) {
	logger := zap.NewNop()
	storageSvc := storage.NewStorageService(nil, logger)
	service := NewDefaultService(storageSvc, logger)

	season := 1
	episode := 1
	year := "2024"
	tasks := []Task{
		{
			Media: database.Media{
				Type:    "tv",
				Title:   "Benchmark Show",
				Year:    &year,
				Season:  &season,
				Episode: &episode,
			},
			SourcePath: "/tmp/source.mkv",
			TargetPath: "/tmp/target.mkv",
			Mode:       storage.TransferModeCopy,
			Category:   "tv",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Execute(tasks)
	}
}
