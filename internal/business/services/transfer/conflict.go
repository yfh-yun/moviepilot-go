package transfer

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// ConflictType 冲突类型
type ConflictType int

const (
	// NoConflict 无冲突
	NoConflict ConflictType = iota
	// SameFile 相同文件
	SameFile
	// DifferentFile 不同文件
	DifferentFile
)

// ConflictStrategy 冲突处理策略
type ConflictStrategy string

const (
	// StrategyOverwrite 覆盖
	StrategyOverwrite ConflictStrategy = "overwrite"
	// StrategySkip 跳过
	StrategySkip ConflictStrategy = "skip"
	// StrategyRename 重命名
	StrategyRename ConflictStrategy = "rename"
	// StrategyAsk 询问用户
	StrategyAsk ConflictStrategy = "ask"
)

// ConflictHandler 冲突处理器
type ConflictHandler struct {
	strategy       ConflictStrategy
	verifyChecksum bool
	logger         *zap.Logger
}

// ConflictHandlerConfig 冲突处理器配置
type ConflictHandlerConfig struct {
	Strategy       ConflictStrategy
	VerifyChecksum bool
}

// NewConflictHandler 创建冲突处理器
func NewConflictHandler(config ConflictHandlerConfig, logger *zap.Logger) *ConflictHandler {
	if config.Strategy == "" {
		config.Strategy = StrategySkip
	}

	return &ConflictHandler{
		strategy:       config.Strategy,
		verifyChecksum: config.VerifyChecksum,
		logger:         logger,
	}
}

// CheckConflict 检查文件冲突
func (h *ConflictHandler) CheckConflict(sourcePath, targetPath string) (ConflictType, error) {
	// 检查目标文件是否存在
	targetInfo, err := os.Stat(targetPath)
	if os.IsNotExist(err) {
		return NoConflict, nil
	}
	if err != nil {
		return NoConflict, fmt.Errorf("failed to stat target file: %w", err)
	}

	// 获取源文件信息
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return NoConflict, fmt.Errorf("failed to stat source file: %w", err)
	}

	// 比较文件大小
	if sourceInfo.Size() != targetInfo.Size() {
		if h.logger != nil {
			h.logger.Debug("file size mismatch",
				zap.String("source", sourcePath),
				zap.String("target", targetPath),
				zap.Int64("source_size", sourceInfo.Size()),
				zap.Int64("target_size", targetInfo.Size()))
		}
		return DifferentFile, nil
	}

	// 如果需要校验和验证
	if h.verifyChecksum {
		same, err := h.compareChecksum(sourcePath, targetPath)
		if err != nil {
			if h.logger != nil {
				h.logger.Warn("failed to compare checksum",
					zap.String("source", sourcePath),
					zap.String("target", targetPath),
					zap.Error(err))
			}
			// 校验失败时，认为是不同文件
			return DifferentFile, nil
		}

		if same {
			return SameFile, nil
		}
		return DifferentFile, nil
	}

	// 如果不验证校验和，大小相同就认为是同一文件
	return SameFile, nil
}

// HandleConflict 处理文件冲突
func (h *ConflictHandler) HandleConflict(sourcePath, targetPath string, conflictType ConflictType) (string, error) {
	if conflictType == NoConflict {
		return targetPath, nil
	}

	if h.logger != nil {
		h.logger.Info("handling file conflict",
			zap.String("source", sourcePath),
			zap.String("target", targetPath),
			zap.String("conflict_type", h.conflictTypeString(conflictType)),
			zap.String("strategy", string(h.strategy)))
	}

	switch h.strategy {
	case StrategyOverwrite:
		return targetPath, nil

	case StrategySkip:
		if conflictType == SameFile {
			// 相同文件，跳过
			return "", ErrSkipped
		}
		// 不同文件，也跳过
		return "", ErrSkipped

	case StrategyRename:
		return h.generateNewName(targetPath), nil

	case StrategyAsk:
		// 在实际应用中，这里应该询问用户
		// 目前默认跳过
		return "", ErrSkipped

	default:
		return "", fmt.Errorf("unknown conflict strategy: %s", h.strategy)
	}
}

// compareChecksum 比较两个文件的MD5校验和
func (h *ConflictHandler) compareChecksum(file1, file2 string) (bool, error) {
	hash1, err := h.calculateMD5(file1)
	if err != nil {
		return false, fmt.Errorf("failed to calculate MD5 for %s: %w", file1, err)
	}

	hash2, err := h.calculateMD5(file2)
	if err != nil {
		return false, fmt.Errorf("failed to calculate MD5 for %s: %w", file2, err)
	}

	return hash1 == hash2, nil
}

// calculateMD5 计算文件的MD5校验和
func (h *ConflictHandler) calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// generateNewName 生成新的文件名
func (h *ConflictHandler) generateNewName(targetPath string) string {
	dir := filepath.Dir(targetPath)
	ext := filepath.Ext(targetPath)
	nameWithoutExt := strings.TrimSuffix(filepath.Base(targetPath), ext)

	// 尝试添加数字后缀
	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s (%d)%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(dir, newName)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			if h.logger != nil {
				h.logger.Debug("generated new name",
					zap.String("original", targetPath),
					zap.String("new", newPath))
			}
			return newPath
		}
	}

	// 如果尝试了1000次还没找到，使用时间戳
	timestamp := fmt.Sprintf("%d", os.Getpid())
	newName := fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
	return filepath.Join(dir, newName)
}

// conflictTypeString 将冲突类型转换为字符串
func (h *ConflictHandler) conflictTypeString(ct ConflictType) string {
	switch ct {
	case NoConflict:
		return "no_conflict"
	case SameFile:
		return "same_file"
	case DifferentFile:
		return "different_file"
	default:
		return "unknown"
	}
}

// VerifyTransfer 验证转移后的文件完整性
func VerifyTransfer(sourcePath, targetPath string, verifyChecksum bool, logger *zap.Logger) error {
	// 检查目标文件是否存在
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("target file does not exist: %w", err)
	}

	// 获取源文件信息
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("source file does not exist: %w", err)
	}

	// 比较文件大小
	if sourceInfo.Size() != targetInfo.Size() {
		return fmt.Errorf("file size mismatch: source=%d, target=%d",
			sourceInfo.Size(), targetInfo.Size())
	}

	// 如果需要校验和验证
	if verifyChecksum {
		handler := &ConflictHandler{logger: logger}
		same, err := handler.compareChecksum(sourcePath, targetPath)
		if err != nil {
			return fmt.Errorf("failed to compare checksum: %w", err)
		}

		if !same {
			return fmt.Errorf("checksum mismatch")
		}

		if logger != nil {
			logger.Info("transfer verified successfully",
				zap.String("source", sourcePath),
				zap.String("target", targetPath),
				zap.Int64("size", sourceInfo.Size()))
		}
	}

	return nil
}

// ErrSkipped 跳过错误
var ErrSkipped = fmt.Errorf("file skipped due to conflict")
