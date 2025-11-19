package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DirectoryHelper 目录辅助工具
type DirectoryHelper struct{}

// NewDirectoryHelper 创建目录辅助工具实例
func NewDirectoryHelper() *DirectoryHelper {
	return &DirectoryHelper{}
}

// CreateDirectory 创建目录
func (d *DirectoryHelper) CreateDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	return nil
}

// DeleteDirectory 删除目录
func (d *DirectoryHelper) DeleteDirectory(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}
	return nil
}

// DirectoryExists 检查目录是否存在
func (d *DirectoryHelper) DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsEmptyDirectory 检查目录是否为空
func (d *DirectoryHelper) IsEmptyDirectory(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("打开目录失败: %w", err)
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil // 目录不为空
	}
	if err.Error() == "EOF" {
		return true, nil // 目录为空
	}
	return false, fmt.Errorf("读取目录失败: %w", err)
}

// ListDirectories 列出目录下的所有子目录
func (d *DirectoryHelper) ListDirectories(path string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(path, func(filePath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filePath != path {
			dirs = append(dirs, filePath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("列出目录失败: %w", err)
	}
	return dirs, nil
}

// ListFiles 列出目录下的所有文件
func (d *DirectoryHelper) ListFiles(path string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(path, func(filePath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, filePath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("列出文件失败: %w", err)
	}
	return files, nil
}

// GetDirectorySize 获取目录大小
func (d *DirectoryHelper) GetDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(filePath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileInfo, err := info.Info()
			if err != nil {
				return err
			}
			size += fileInfo.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("获取目录大小失败: %w", err)
	}
	return size, nil
}

// CopyDirectory 复制目录
func (d *DirectoryHelper) CopyDirectory(src, dst string) error {
	// 创建目标目录
	if err := d.CreateDirectory(dst); err != nil {
		return err
	}

	// 遍历源目录
	return filepath.WalkDir(src, func(filePath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(src, filePath)
		if err != nil {
			return err
		}

		// 构建目标路径
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// 创建目录
			return d.CreateDirectory(dstPath)
		} else {
			// 复制文件
			return d.CopyFile(filePath, dstPath)
		}
	})
}

// MoveDirectory 移动目录
func (d *DirectoryHelper) MoveDirectory(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		// 如果重命名失败，尝试复制后删除
		if err := d.CopyDirectory(src, dst); err != nil {
			return fmt.Errorf("移动目录失败: %w", err)
		}
		return d.DeleteDirectory(src)
	}
	return nil
}

// GetParentDirectory 获取父目录
func (d *DirectoryHelper) GetParentDirectory(path string) string {
	return filepath.Dir(path)
}

// GetDirectoryName 获取目录名
func (d *DirectoryHelper) GetDirectoryName(path string) string {
	return filepath.Base(path)
}

// JoinPath 连接路径
func (d *DirectoryHelper) JoinPath(paths ...string) string {
	return filepath.Join(paths...)
}

// IsAbsolutePath 检查是否为绝对路径
func (d *DirectoryHelper) IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// GetRelativePath 获取相对路径
func (d *DirectoryHelper) GetRelativePath(base, target string) (string, error) {
	relPath, err := filepath.Rel(base, target)
	if err != nil {
		return "", fmt.Errorf("获取相对路径失败: %w", err)
	}
	return relPath, nil
}

// CleanPath 清理路径
func (d *DirectoryHelper) CleanPath(path string) string {
	return filepath.Clean(path)
}

// SplitPath 分割路径
func (d *DirectoryHelper) SplitPath(path string) (dir, file string) {
	return filepath.Split(path)
}

// MatchPattern 检查路径是否匹配模式
func (d *DirectoryHelper) MatchPattern(pattern, path string) (bool, error) {
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err != nil {
		return false, fmt.Errorf("路径匹配失败: %w", err)
	}
	return matched, nil
}

// GetTempDir 获取临时目录
func (d *DirectoryHelper) GetTempDir() string {
	return os.TempDir()
}

// GetWorkingDirectory 获取当前工作目录
func (d *DirectoryHelper) GetWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}
	return wd, nil
}

// SetWorkingDirectory 设置工作目录
func (d *DirectoryHelper) SetWorkingDirectory(path string) error {
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("设置工作目录失败: %w", err)
	}
	return nil
}

// CopyFile 复制文件（内部使用）
func (d *DirectoryHelper) CopyFile(src, dst string) error {
	source, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %w", err)
	}

	if err := os.WriteFile(dst, source, 0644); err != nil {
		return fmt.Errorf("写入目标文件失败: %w", err)
	}

	return nil
}

// FindFilesByExtension 根据扩展名查找文件
func (d *DirectoryHelper) FindFilesByExtension(path, ext string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(path, func(filePath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(filePath), strings.ToLower(ext)) {
			files = append(files, filePath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("查找文件失败: %w", err)
	}
	return files, nil
}

// FindFilesByName 根据名称查找文件
func (d *DirectoryHelper) FindFilesByName(path, name string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(path, func(filePath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.EqualFold(filepath.Base(filePath), name) {
			files = append(files, filePath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("查找文件失败: %w", err)
	}
	return files, nil
}