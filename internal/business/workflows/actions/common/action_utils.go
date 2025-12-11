package common

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// MergeMaps 合并两个map，将src合并到dst中
func MergeMaps(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any)
	}

	for k, v := range src {
		dst[k] = v
	}

	return dst
}

// DeepCopyMap 深拷贝map
func DeepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	// 使用JSON序列化/反序列化实现深拷贝
	data, err := json.Marshal(src)
	if err != nil {
		// 如果序列化失败，返回浅拷贝
		dst := make(map[string]any)
		for k, v := range src {
			dst[k] = v
		}
		return dst
	}

	var dst map[string]any
	if err := json.Unmarshal(data, &dst); err != nil {
		// 如果反序列化失败，返回浅拷贝
		shallowDst := make(map[string]any)
		for k, v := range src {
			shallowDst[k] = v
		}
		return shallowDst
	}

	return dst
}

// ValidateFilePath 验证文件路径是否有效
func ValidateFilePath(path string) error {
	// 检查路径是否为空
	if path == "" {
		return ErrEmptyFilePath
	}

	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrFilePathNotExist
	}

	return nil
}

// EnsureDirectory 确保目录存在，如果不存在则创建
func EnsureDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// GetAbsoluteFilePath 获取文件的绝对路径
func GetAbsoluteFilePath(path string) (string, error) {
	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(path) {
		return path, nil
	}

	// 获取当前工作目录
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 拼接绝对路径
	return filepath.Join(pwd, path), nil
}

// ExtractFileExtension 提取文件扩展名
func ExtractFileExtension(filename string) string {
	return filepath.Ext(filename)[1:] // 移除点号
}

// IsValidOutputKey 检查输出键是否有效
func IsValidOutputKey(key string) bool {
	return key != "" // 简单检查，实际应用中可能需要更复杂的验证
}

// errors 定义工具函数的错误
var (
	ErrEmptyFilePath    = NewActionError("empty_file_path", "File path is empty")
	ErrFilePathNotExist = NewActionError("file_path_not_exist", "File path does not exist")
)

// ActionError 定义动作错误

type ActionError struct {
	Code    string // 错误代码
	Message string // 错误信息
}

// NewActionError 创建新的动作错误
func NewActionError(code, message string) *ActionError {
	return &ActionError{
		Code:    code,
		Message: message,
	}
}

// Error 实现error接口
func (e *ActionError) Error() string {
	return e.Message
}

// GetCode 获取错误代码
func (e *ActionError) GetCode() string {
	return e.Code
}
