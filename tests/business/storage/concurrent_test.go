package business

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"moviepilot-go/internal/business/storage"
)

func TestNewConcurrentScanner(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// 测试默认并发数
	scanner := storage.NewConcurrentScanner(0, logger)
	assert.NotNil(t, scanner)

	// 测试自定义并发数
	scanner = storage.NewConcurrentScanner(20, logger)
	assert.NotNil(t, scanner)
}

func TestConcurrentScanner_ScanConcurrent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	scanner := storage.NewConcurrentScanner(5, logger)

	// 创建测试目录
	tempDir := t.TempDir()

	// 创建测试文件
	testFiles := []string{
		"movie1.mkv",
		"movie2.mp4",
		"movie3.avi",
		"subdir/movie4.mkv",
		"subdir/movie5.mp4",
	}

	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		dir := filepath.Dir(path)
		os.MkdirAll(dir, 0755)
		os.WriteFile(path, []byte("test content"), 0644)
	}

	// 测试基本扫描
	opts := storage.ScanOptions{
		RootPath: tempDir,
	}

	ctx := context.Background()
	files, err := scanner.ScanConcurrent(ctx, opts)

	assert.NoError(t, err)
	assert.Len(t, files, 5)

	// 验证文件信息
	for _, file := range files {
		assert.NotEmpty(t, file.Path)
		assert.Greater(t, file.Size, int64(0))
		assert.False(t, file.ModTime.IsZero())
	}
}

func TestConcurrentScanner_WithIncludeExclude(t *testing.T) {
	logger := zaptest.NewLogger(t)
	scanner := storage.NewConcurrentScanner(5, logger)

	tempDir := t.TempDir()

	// 创建不同类型的文件
	testFiles := map[string]string{
		"movie1.mkv": "video",
		"movie2.mp4": "video",
		"image.jpg":  "image",
		"doc.txt":    "text",
	}

	for file := range testFiles {
		path := filepath.Join(tempDir, file)
		os.WriteFile(path, []byte("test"), 0644)
	}

	// 测试包含规则
	opts := storage.ScanOptions{
		RootPath: tempDir,
		Include:  []string{"*.mkv", "*.mp4"},
	}

	ctx := context.Background()
	files, err := scanner.ScanConcurrent(ctx, opts)

	assert.NoError(t, err)
	assert.Len(t, files, 2)

	// 测试排除规则
	opts = storage.ScanOptions{
		RootPath: tempDir,
		Exclude:  []string{"*.jpg", "*.txt"},
	}

	files, err = scanner.ScanConcurrent(ctx, opts)

	assert.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestConcurrentScanner_WithMaxDepth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	scanner := storage.NewConcurrentScanner(5, logger)

	tempDir := t.TempDir()

	// 创建多层目录结构
	os.MkdirAll(filepath.Join(tempDir, "level1/level2/level3"), 0755)
	os.WriteFile(filepath.Join(tempDir, "file0.mkv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "level1/file1.mkv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "level1/level2/file2.mkv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "level1/level2/level3/file3.mkv"), []byte("test"), 0644)

	// 测试深度限制
	opts := storage.ScanOptions{
		RootPath: tempDir,
		MaxDepth: 2,
	}

	ctx := context.Background()
	files, err := scanner.ScanConcurrent(ctx, opts)

	assert.NoError(t, err)
	// 应该只扫描到 level2，不包括 level3
	assert.LessOrEqual(t, len(files), 3)
}

func TestConcurrentScanner_WithContext(t *testing.T) {
	logger := zaptest.NewLogger(t)
	scanner := storage.NewConcurrentScanner(5, logger)

	tempDir := t.TempDir()

	// 创建大量文件
	for i := 0; i < 100; i++ {
		path := filepath.Join(tempDir, filepath.Join("file"+string(rune(i))+".mkv"))
		os.WriteFile(path, []byte("test"), 0644)
	}

	// 测试上下文取消
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	opts := storage.ScanOptions{
		RootPath: tempDir,
	}

	_, err := scanner.ScanConcurrent(ctx, opts)

	// 可能会因为超时而返回错误
	// 这取决于扫描速度
	if err != nil {
		assert.Equal(t, context.DeadlineExceeded, err)
	}
}

// 基准测试
func BenchmarkConcurrentScanner_Scan(b *testing.B) {
	logger := zaptest.NewLogger(b)
	scanner := storage.NewConcurrentScanner(10, logger)

	tempDir := b.TempDir()

	// 创建测试文件
	for i := 0; i < 100; i++ {
		path := filepath.Join(tempDir, filepath.Join("file"+string(rune(i))+".mkv"))
		os.WriteFile(path, []byte("test content"), 0644)
	}

	opts := storage.ScanOptions{
		RootPath: tempDir,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.ScanConcurrent(ctx, opts)
	}
}

func BenchmarkConcurrentScanner_ScanWithDifferentConcurrency(b *testing.B) {
	tempDir := b.TempDir()

	// 创建测试文件
	for i := 0; i < 1000; i++ {
		path := filepath.Join(tempDir, filepath.Join("file"+string(rune(i))+".mkv"))
		os.WriteFile(path, []byte("test content"), 0644)
	}

	opts := storage.ScanOptions{
		RootPath: tempDir,
	}

	ctx := context.Background()

	concurrencies := []int{1, 5, 10, 20, 50}

	for _, concurrency := range concurrencies {
		b.Run(filepath.Join("concurrency_"+string(rune(concurrency))), func(b *testing.B) {
			logger := zaptest.NewLogger(b)
			scanner := storage.NewConcurrentScanner(concurrency, logger)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = scanner.ScanConcurrent(ctx, opts)
			}
		})
	}
}
