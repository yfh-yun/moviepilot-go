package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileBackend 文件缓存后端
// 原Python: FileBackend in app/core/cache.py
type FileBackend struct {
	config Config // 配置
}

// NewFileBackend 创建文件缓存后端
func NewFileBackend(config Config) (*FileBackend, error) {
	// 确保基础目录存在
	if err := os.MkdirAll(config.FileBaseDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	return &FileBackend{
		config: config,
	}, nil
}

// getFilepath 获取缓存文件路径
// 原Python: 对应文件缓存的路径生成规则
func (f *FileBackend) getFilepath(region, key string) string {
	// 计算key的MD5哈希，避免文件名包含非法字符
	hash := md5.Sum([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	// 构建文件路径: baseDir/region/keyHash.json
	regionDir := filepath.Join(f.config.FileBaseDir, region)
	return filepath.Join(regionDir, fmt.Sprintf("%s.json", keyHash))
}

// Set 设置缓存
func (f *FileBackend) Set(region, key string, value any, ttlSeconds int64) error {
	// 序列化值
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 获取文件路径
	filePath := f.getFilepath(region, key)

	// 确保区域目录存在
	regionDir := filepath.Dir(filePath)
	if err = os.MkdirAll(regionDir, 0755); err != nil {
		return fmt.Errorf("创建区域目录失败: %w", err)
	}

	// 使用临时文件 + 重命名实现原子写入
	tempFile, err := os.CreateTemp(regionDir, "cache_*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// 写入数据
	if _, err := tempFile.Write(bytes); err != nil {
		tempFile.Close()
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 关闭临时文件
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	// 原子替换
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// Get 获取缓存
func (f *FileBackend) Get(region, key string, dest any) (bool, error) {
	// 获取文件路径
	filePath := f.getFilepath(region, key)

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在
			return false, nil
		}
		return false, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 检查TTL
	if ttl := f.config.DefaultTTL; ttl > 0 {
		if time.Since(info.ModTime()) > ttl {
			// 过期，删除文件
			_ = f.Delete(region, key)
			return false, nil
		}
	}

	// 读取文件内容
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("读取文件失败: %w", err)
	}

	// 反序列化值
	if err := json.Unmarshal(bytes, dest); err != nil {
		return false, err
	}

	return true, nil
}

// Exists 检查缓存是否存在
func (f *FileBackend) Exists(region, key string) (bool, error) {
	// 获取文件路径
	filePath := f.getFilepath(region, key)

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在
			return false, nil
		}
		return false, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 检查TTL
	if ttl := f.config.DefaultTTL; ttl > 0 {
		if time.Since(info.ModTime()) > ttl {
			// 过期，删除文件
			_ = f.Delete(region, key)
			return false, nil
		}
	}

	return true, nil
}

// Delete 删除缓存
func (f *FileBackend) Delete(region, key string) error {
	// 获取文件路径
	filePath := f.getFilepath(region, key)

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，忽略错误
			return nil
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// Clear 清空缓存
func (f *FileBackend) Clear(region string) error {
	var targetDir string
	if region == "" {
		// 清空所有区域
		targetDir = f.config.FileBaseDir
	} else {
		// 清空指定区域
		targetDir = filepath.Join(f.config.FileBaseDir, region)
	}

	// 检查目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// 目录不存在，忽略错误
		return nil
	}

	if region == "" {
		// 清空所有区域：删除所有子目录，保留基础目录
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			return fmt.Errorf("读取缓存目录失败: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(targetDir, entry.Name())
				if err := os.RemoveAll(subDir); err != nil {
					return fmt.Errorf("删除区域目录失败: %w", err)
				}
			}
		}
	} else {
		// 清空指定区域：删除整个区域目录
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("删除区域目录失败: %w", err)
		}
	}

	return nil
}

// Close 关闭文件缓存（无需操作）
func (f *FileBackend) Close() error {
	return nil
}
