package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileBackend 文件缓存后端
// 原Python: FileBackend in app/core/cache.py
type FileBackend struct {
	BaseCacheBackend
	config Config // 配置
}

// NewFileBackend 创建文件缓存后端
func NewFileBackend(config Config) (CacheBackend, error) {
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
func (f *FileBackend) Set(key string, value interface{}, ttl time.Duration, region string, opts ...Option) error {
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
func (f *FileBackend) Get(key string, region string) (interface{}, bool, error) {
	// 获取文件路径
	filePath := f.getFilepath(region, key)

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 检查TTL
	if ttl := f.config.DefaultTTL; ttl > 0 {
		if time.Since(info.ModTime()) > ttl {
			// 过期，删除文件
			_ = f.Delete(key, region)
			return nil, false, nil
		}
	}

	// 读取文件内容
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false, fmt.Errorf("读取文件失败: %w", err)
	}

	// 反序列化值
	var value interface{}
	if err := json.Unmarshal(bytes, &value); err != nil {
		return nil, false, err
	}

	return value, true, nil
}

// Exists 检查缓存是否存在
func (f *FileBackend) Exists(key string, region string) (bool, error) {
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
			_ = f.Delete(key, region)
			return false, nil
		}
	}

	return true, nil
}

// Delete 删除缓存
func (f *FileBackend) Delete(key string, region string) error {
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

// Items 获取指定区域的所有缓存项
func (f *FileBackend) Items(region string) (map[string]interface{}, error) {
	// 构建区域目录路径
	regionDir := filepath.Join(f.config.FileBaseDir, region)

	// 检查目录是否存在
	if _, err := os.Stat(regionDir); os.IsNotExist(err) {
		// 目录不存在，返回空map
		return make(map[string]interface{}), nil
	}

	// 遍历目录中的所有文件
	entries, err := os.ReadDir(regionDir)
	if err != nil {
		return nil, fmt.Errorf("读取区域目录失败: %w", err)
	}

	result := make(map[string]interface{})
	now := time.Now()

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			// 跳过非文件
			continue
		}

		// 获取文件名（不含扩展名）
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".json") {
			// 跳过非JSON文件
			continue
		}
		hashedKey := strings.TrimSuffix(filename, ".json")

		// 读取文件内容
		filePath := filepath.Join(regionDir, filename)
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("读取缓存文件失败: %w", err)
		}

		// 检查TTL
		if ttl := f.config.DefaultTTL; ttl > 0 {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("获取文件信息失败: %w", err)
			}
			if now.Sub(info.ModTime()) > ttl {
				// 过期，删除文件
				_ = f.Delete(hashedKey, region)
				continue
			}
		}

		// 反序列化值
		var value interface{}
		if err := json.Unmarshal(bytes, &value); err != nil {
			return nil, fmt.Errorf("反序列化缓存值失败: %w", err)
		}

		// 直接使用哈希值作为键（在FileBackend中，我们无法恢复原始键名，因为它被哈希处理了）
		result[hashedKey] = value
	}

	return result, nil
}

// Keys 返回所有缓存键
func (f *FileBackend) Keys(region string) ([]string, error) {
	// 构建区域目录路径
	regionDir := filepath.Join(f.config.FileBaseDir, region)

	// 检查目录是否存在
	if _, err := os.Stat(regionDir); os.IsNotExist(err) {
		// 目录不存在，返回空切片
		return make([]string, 0), nil
	}

	// 遍历目录中的所有文件
	entries, err := os.ReadDir(regionDir)
	if err != nil {
		return nil, fmt.Errorf("读取区域目录失败: %w", err)
	}

	result := make([]string, 0, len(entries))
	now := time.Now()

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			// 跳过非文件
			continue
		}

		// 获取文件名（不含扩展名）
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".json") {
			// 跳过非JSON文件
			continue
		}
		hashedKey := strings.TrimSuffix(filename, ".json")

		// 检查TTL
		if ttl := f.config.DefaultTTL; ttl > 0 {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("获取文件信息失败: %w", err)
			}
			if now.Sub(info.ModTime()) > ttl {
				// 过期，删除文件
				_ = f.Delete(hashedKey, region)
				continue
			}
		}

		result = append(result, hashedKey)
	}

	return result, nil
}

// Values 返回所有缓存值
func (f *FileBackend) Values(region string) ([]interface{}, error) {
	// 构建区域目录路径
	regionDir := filepath.Join(f.config.FileBaseDir, region)

	// 检查目录是否存在
	if _, err := os.Stat(regionDir); os.IsNotExist(err) {
		// 目录不存在，返回空切片
		return make([]interface{}, 0), nil
	}

	// 遍历目录中的所有文件
	entries, err := os.ReadDir(regionDir)
	if err != nil {
		return nil, fmt.Errorf("读取区域目录失败: %w", err)
	}

	result := make([]interface{}, 0, len(entries))
	now := time.Now()

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			// 跳过非文件
			continue
		}

		// 获取文件名（不含扩展名）
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".json") {
			// 跳过非JSON文件
			continue
		}
		hashedKey := strings.TrimSuffix(filename, ".json")

		// 读取文件内容
		filePath := filepath.Join(regionDir, filename)
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("读取缓存文件失败: %w", err)
		}

		// 检查TTL
		if ttl := f.config.DefaultTTL; ttl > 0 {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("获取文件信息失败: %w", err)
			}
			if now.Sub(info.ModTime()) > ttl {
				// 过期，删除文件
				_ = f.Delete(hashedKey, region)
				continue
			}
		}

		// 反序列化值
		var value interface{}
		if err := json.Unmarshal(bytes, &value); err != nil {
			return nil, fmt.Errorf("反序列化缓存值失败: %w", err)
		}

		result = append(result, value)
	}

	return result, nil
}

// Update 更新缓存
func (f *FileBackend) Update(other map[string]interface{}, region string, ttl time.Duration, opts ...Option) error {
	// 批量设置缓存
	for key, value := range other {
		if err := f.Set(key, value, ttl, region, opts...); err != nil {
			return err
		}
	}

	return nil
}

// Pop 弹出缓存项
func (f *FileBackend) Pop(key string, region string, defaultValue ...interface{}) (interface{}, error) {
	// 构建文件路径
	filePath := f.getFilepath(region, key)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return nil, nil
	}

	// 读取文件内容
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取缓存文件失败: %w", err)
	}

	// 反序列化值
	var value interface{}
	if err := json.Unmarshal(bytes, &value); err != nil {
		return nil, fmt.Errorf("反序列化缓存值失败: %w", err)
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return nil, fmt.Errorf("删除缓存文件失败: %w", err)
	}

	return value, nil
}

// Popitem 弹出最后一个缓存项
func (f *FileBackend) Popitem(region string) (string, interface{}, error) {
	// 构建区域目录路径
	regionDir := filepath.Join(f.config.FileBaseDir, region)

	// 检查目录是否存在
	if _, err := os.Stat(regionDir); os.IsNotExist(err) {
		// 目录不存在，返回空
		return "", nil, nil
	}

	// 遍历目录中的所有文件
	entries, err := os.ReadDir(regionDir)
	if err != nil {
		return "", nil, fmt.Errorf("读取区域目录失败: %w", err)
	}

	if len(entries) == 0 {
		// 目录为空，返回空
		return "", nil, nil
	}

	// 选择第一个文件
	entry := entries[0]

	// 获取文件名（不含扩展名）
	filename := entry.Name()
	if !strings.HasSuffix(filename, ".json") {
		// 跳过非JSON文件
		return "", nil, nil
	}
	hashedKey := strings.TrimSuffix(filename, ".json")

	// 读取文件内容
	filePath := filepath.Join(regionDir, filename)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("读取缓存文件失败: %w", err)
	}

	// 反序列化值
	var value interface{}
	if err := json.Unmarshal(bytes, &value); err != nil {
		return "", nil, fmt.Errorf("反序列化缓存值失败: %w", err)
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return "", nil, fmt.Errorf("删除缓存文件失败: %w", err)
	}

	return hashedKey, value, nil
}

// Setdefault 设置默认值
func (f *FileBackend) Setdefault(key string, defaultValue interface{}, region string, ttl time.Duration, opts ...Option) (interface{}, error) {
	// 构建文件路径
	filePath := f.getFilepath(region, key)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		// 文件存在，读取并返回值
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("读取缓存文件失败: %w", err)
		}

		// 检查TTL
		if fileTTL := f.config.DefaultTTL; fileTTL > 0 {
			info, err := os.Stat(filePath)
			if err != nil {
				return nil, fmt.Errorf("获取文件信息失败: %w", err)
			}
			if time.Since(info.ModTime()) > fileTTL {
				// 过期，删除文件，继续设置默认值
				_ = os.Remove(filePath)
				goto setDefault
			}
		}

		// 反序列化值
		var value interface{}
		if err := json.Unmarshal(bytes, &value); err != nil {
			return nil, fmt.Errorf("反序列化缓存值失败: %w", err)
		}

		return value, nil
	}

setDefault:
	// 文件不存在，设置默认值
	if err := f.Set(key, defaultValue, ttl, region, opts...); err != nil {
		return nil, err
	}

	return defaultValue, nil
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

// IsRedis 判断当前缓存后端是否为Redis
func (f *FileBackend) IsRedis() bool {
	// 文件缓存不是Redis
	return false
}

// AsyncFileBackend 异步文件缓存后端
// 原Python: AsyncFileBackend in app/core/cache.py
type AsyncFileBackend struct {
	*FileBackend
}

// NewAsyncFileBackend 创建异步文件缓存后端
func NewAsyncFileBackend(config Config) (*AsyncFileBackend, error) {
	backend, err := NewFileBackend(config)
	if err != nil {
		return nil, err
	}

	return &AsyncFileBackend{
		FileBackend: backend.(*FileBackend),
	}, nil
}

// Keys 返回所有缓存键（异步实现）
func (f *AsyncFileBackend) Keys(region string) ([]string, error) {
	return f.FileBackend.Keys(region)
}

// Values 返回所有缓存值（异步实现）
func (f *AsyncFileBackend) Values(region string) ([]interface{}, error) {
	return f.FileBackend.Values(region)
}

// Update 更新缓存（异步实现）
func (f *AsyncFileBackend) Update(other map[string]interface{}, region string, ttl time.Duration, opts ...Option) error {
	return f.FileBackend.Update(other, region, ttl, opts...)
}
