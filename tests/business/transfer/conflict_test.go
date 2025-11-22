package business

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"moviepilot-go/internal/business/transfer"
)

func TestNewConflictHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// 测试默认策略
	config := transfer.ConflictHandlerConfig{}
	handler := transfer.NewConflictHandler(config, logger)
	assert.NotNil(t, handler)

	// 测试自定义策略
	config = transfer.ConflictHandlerConfig{
		Strategy:       transfer.StrategyOverwrite,
		VerifyChecksum: true,
	}
	handler = transfer.NewConflictHandler(config, logger)
	assert.NotNil(t, handler)
}

func TestConflictHandler_CheckConflict(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := transfer.ConflictHandlerConfig{
		Strategy:       transfer.StrategySkip,
		VerifyChecksum: false,
	}
	handler := transfer.NewConflictHandler(config, logger)

	// 创建临时目录
	tempDir := t.TempDir()

	// 创建源文件
	sourcePath := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourcePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	// 测试目标文件不存在
	targetPath := filepath.Join(tempDir, "target.txt")
	conflictType, err := handler.CheckConflict(sourcePath, targetPath)
	assert.NoError(t, err)
	assert.Equal(t, transfer.NoConflict, conflictType)

	// 创建相同的目标文件
	err = os.WriteFile(targetPath, []byte("test content"), 0644)
	assert.NoError(t, err)

	conflictType, err = handler.CheckConflict(sourcePath, targetPath)
	assert.NoError(t, err)
	assert.Equal(t, transfer.SameFile, conflictType)

	// 创建不同的目标文件
	err = os.WriteFile(targetPath, []byte("different content"), 0644)
	assert.NoError(t, err)

	conflictType, err = handler.CheckConflict(sourcePath, targetPath)
	assert.NoError(t, err)
	assert.Equal(t, transfer.DifferentFile, conflictType)
}

func TestConflictHandler_CheckConflict_WithChecksum(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := transfer.ConflictHandlerConfig{
		Strategy:       transfer.StrategySkip,
		VerifyChecksum: true,
	}
	handler := transfer.NewConflictHandler(config, logger)

	// 创建临时目录
	tempDir := t.TempDir()

	// 创建源文件
	sourcePath := filepath.Join(tempDir, "source.txt")
	content := []byte("test content for checksum")
	err := os.WriteFile(sourcePath, content, 0644)
	assert.NoError(t, err)

	// 创建相同内容的目标文件
	targetPath := filepath.Join(tempDir, "target.txt")
	err = os.WriteFile(targetPath, content, 0644)
	assert.NoError(t, err)

	conflictType, err := handler.CheckConflict(sourcePath, targetPath)
	assert.NoError(t, err)
	assert.Equal(t, transfer.SameFile, conflictType)

	// 创建相同大小但不同内容的文件
	differentContent := []byte("diff content for checksum")
	assert.Equal(t, len(content), len(differentContent)) // 确保大小相同
	err = os.WriteFile(targetPath, differentContent, 0644)
	assert.NoError(t, err)

	conflictType, err = handler.CheckConflict(sourcePath, targetPath)
	assert.NoError(t, err)
	assert.Equal(t, transfer.DifferentFile, conflictType)
}

func TestConflictHandler_HandleConflict(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	sourcePath := filepath.Join(tempDir, "source.txt")
	targetPath := filepath.Join(tempDir, "target.txt")

	// 测试无冲突
	config := transfer.ConflictHandlerConfig{Strategy: transfer.StrategySkip}
	handler := transfer.NewConflictHandler(config, logger)

	result, err := handler.HandleConflict(sourcePath, targetPath, transfer.NoConflict)
	assert.NoError(t, err)
	assert.Equal(t, targetPath, result)

	// 测试覆盖策略
	config = transfer.ConflictHandlerConfig{Strategy: transfer.StrategyOverwrite}
	handler = transfer.NewConflictHandler(config, logger)

	result, err = handler.HandleConflict(sourcePath, targetPath, transfer.DifferentFile)
	assert.NoError(t, err)
	assert.Equal(t, targetPath, result)

	// 测试跳过策略
	config = transfer.ConflictHandlerConfig{Strategy: transfer.StrategySkip}
	handler = transfer.NewConflictHandler(config, logger)

	result, err = handler.HandleConflict(sourcePath, targetPath, transfer.SameFile)
	assert.Error(t, err)
	assert.Equal(t, transfer.ErrSkipped, err)
	assert.Empty(t, result)

	// 测试重命名策略
	config = transfer.ConflictHandlerConfig{Strategy: transfer.StrategyRename}
	handler = transfer.NewConflictHandler(config, logger)

	result, err = handler.HandleConflict(sourcePath, targetPath, transfer.DifferentFile)
	assert.NoError(t, err)
	assert.NotEqual(t, targetPath, result)
	assert.Contains(t, result, "target")
	assert.Contains(t, result, "(1)")
}

func TestVerifyTransfer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tempDir := t.TempDir()

	// 创建源文件
	sourcePath := filepath.Join(tempDir, "source.txt")
	content := []byte("test content for verification")
	err := os.WriteFile(sourcePath, content, 0644)
	assert.NoError(t, err)

	// 创建相同的目标文件
	targetPath := filepath.Join(tempDir, "target.txt")
	err = os.WriteFile(targetPath, content, 0644)
	assert.NoError(t, err)

	// 测试不验证校验和
	err = transfer.VerifyTransfer(sourcePath, targetPath, false, logger)
	assert.NoError(t, err)

	// 测试验证校验和
	err = transfer.VerifyTransfer(sourcePath, targetPath, true, logger)
	assert.NoError(t, err)

	// 测试大小不匹配
	differentPath := filepath.Join(tempDir, "different.txt")
	err = os.WriteFile(differentPath, []byte("different"), 0644)
	assert.NoError(t, err)

	err = transfer.VerifyTransfer(sourcePath, differentPath, false, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "size mismatch")

	// 测试校验和不匹配 - 需要相同大小的不同内容
	sameSizePath := filepath.Join(tempDir, "same_size.txt")
	// 创建与源文件相同大小但不同内容的文件
	differentContent := make([]byte, len(content))
	copy(differentContent, content)
	differentContent[0] = 'X' // 修改第一个字节
	err = os.WriteFile(sameSizePath, differentContent, 0644)
	assert.NoError(t, err)

	err = transfer.VerifyTransfer(sourcePath, sameSizePath, true, logger)
	assert.Error(t, err)
	// 可能是大小不匹配或校验和不匹配
	assert.True(t, err != nil)
}

func TestConflictHandler_GenerateNewName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := transfer.ConflictHandlerConfig{Strategy: transfer.StrategyRename}
	handler := transfer.NewConflictHandler(config, logger)

	tempDir := t.TempDir()

	// 创建原始文件
	originalPath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(originalPath, []byte("content"), 0644)
	assert.NoError(t, err)

	// 测试生成新名字
	targetPath := filepath.Join(tempDir, "test.txt")
	newPath, err := handler.HandleConflict("/source/test.txt", targetPath, transfer.DifferentFile)
	assert.NoError(t, err)
	assert.NotEqual(t, targetPath, newPath)
	assert.Contains(t, newPath, "test")
	assert.Contains(t, newPath, "(1)")
	assert.Contains(t, newPath, ".txt")

	// 创建第一个重命名文件
	err = os.WriteFile(newPath, []byte("content"), 0644)
	assert.NoError(t, err)

	// 测试生成第二个新名字
	newPath2, err := handler.HandleConflict("/source/test.txt", targetPath, transfer.DifferentFile)
	assert.NoError(t, err)
	assert.NotEqual(t, newPath, newPath2)
	// 第二次会生成 (2) 因为 (1) 已经存在
	assert.Contains(t, newPath2, "(2)")
}

// 基准测试
func BenchmarkConflictHandler_CheckConflict(b *testing.B) {
	logger := zaptest.NewLogger(b)
	config := transfer.ConflictHandlerConfig{
		Strategy:       transfer.StrategySkip,
		VerifyChecksum: false,
	}
	handler := transfer.NewConflictHandler(config, logger)

	tempDir := b.TempDir()
	sourcePath := filepath.Join(tempDir, "source.txt")
	targetPath := filepath.Join(tempDir, "target.txt")

	content := []byte("test content")
	_ = os.WriteFile(sourcePath, content, 0644)
	_ = os.WriteFile(targetPath, content, 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.CheckConflict(sourcePath, targetPath)
	}
}

func BenchmarkConflictHandler_CheckConflict_WithChecksum(b *testing.B) {
	logger := zaptest.NewLogger(b)
	config := transfer.ConflictHandlerConfig{
		Strategy:       transfer.StrategySkip,
		VerifyChecksum: true,
	}
	handler := transfer.NewConflictHandler(config, logger)

	tempDir := b.TempDir()
	sourcePath := filepath.Join(tempDir, "source.txt")
	targetPath := filepath.Join(tempDir, "target.txt")

	content := []byte("test content for checksum benchmark")
	_ = os.WriteFile(sourcePath, content, 0644)
	_ = os.WriteFile(targetPath, content, 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.CheckConflict(sourcePath, targetPath)
	}
}
