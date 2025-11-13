package douban

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	
	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/cache"
)

// DoubanCache 豆瓣缓存数据
type DoubanCache struct {
	cache           *cache.TTLCache
	maxsize         int
	ttl             int
	region          string
	metaFilepath    string
	mutex           sync.RWMutex
	doubanCacheExpire bool
}

// CacheData 缓存数据结构
type CacheData struct {
	ID         string              `json:"id"`
	Title      string              `json:"title"`
	Year       string              `json:"year"`
	Type       types.MediaType     `json:"type"`
	PosterPath string              `json:"poster_path"`
}

// NewDoubanCache 创建DoubanCache实例
func NewDoubanCache() *DoubanCache {
	dc := &DoubanCache{
		maxsize:         config.Config.CONF.Douban,
		ttl:             config.Config.CONF.Meta,
		region:          "__douban_cache__",
		doubanCacheExpire: true,
	}
	
	dc.metaFilepath = filepath.Join(config.Config.TEMP_PATH, dc.region)
	
	// 初始化缓�?	dc.cache = cache.NewTTLCache(dc.region, dc.maxsize, dc.ttl)
	
	// 非Redis加载本地缓存数据
	if !dc.cache.IsRedis() {
		dc.load(dc.metaFilepath)
	}
	
	return dc
}

// Clear 清空所有豆瓣缓�?func (dc *DoubanCache) Clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	dc.cache.Clear()
}

// getKey 获取缓存KEY
func (dc *DoubanCache) getKey(meta *models.MetaBase) string {
	mediaType := "未知"
	if meta.Type != "" {
		mediaType = string(meta.Type)
	}
	
	return "[" + mediaType + "]" + 
		coalesceString(meta.DoubanID, meta.Name) + "-" + 
		coalesceString(meta.Year, "") + "-" + 
		coalesceString(meta.BeginSeason, "")
}

// Get 根据KEY值获取缓存�?func (dc *DoubanCache) Get(meta *models.MetaBase) *CacheData {
	key := dc.getKey(meta)
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	
	if data, exists := dc.cache.Get(key); exists {
		if cacheData, ok := data.(*CacheData); ok {
			return cacheData
		}
	}
	
	return &CacheData{}
}

// Delete 删除缓存信息
func (dc *DoubanCache) Delete(key string) *CacheData {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	if data, exists := dc.cache.Get(key); exists {
		if cacheData, ok := data.(*CacheData); ok {
			dc.cache.Delete(key)
			return cacheData
		}
	}
	
	return &CacheData{}
}

// Modify 修改缓存信息
func (dc *DoubanCache) Modify(key, title string) *CacheData {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	if data, exists := dc.cache.Get(key); exists {
		if cacheData, ok := data.(*CacheData); ok {
			cacheData.Title = title
			dc.cache.Set(key, cacheData)
			return cacheData
		}
	}
	
	return &CacheData{}
}

// load 从文件中加载缓存
func (dc *DoubanCache) load(path string) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	file, err := os.Open(path)
	if err != nil {
		logger.Debugf("加载豆瓣缓存文件失败: %v", err)
		return
	}
	defer file.Close()
	
	decoder := gob.NewDecoder(file)
	var data map[string]*CacheData
	if err := decoder.Decode(&data); err != nil {
		logger.Errorf("解码豆瓣缓存数据失败: %v", err)
		return
	}
	
	for key, value := range data {
		dc.cache.Set(key, value)
	}
}

// Update 新增或更新缓存条�?func (dc *DoubanCache) Update(meta *models.MetaBase, info map[string]interface{}) {
	if info == nil {
		return
	}
	
	var cacheData *CacheData
	
	if len(info) > 0 {
		// 缓存标题
		cacheTitle := ""
		if title, ok := info["title"].(string); ok {
			cacheTitle = title
		}
		
		// 缓存年份
		cacheYear := ""
		if year, ok := info["year"].(string); ok {
			cacheYear = year
		}
		
		// 类型
		var mtype types.MediaType
		if mediaType, ok := info["media_type"].(types.MediaType); ok {
			mtype = mediaType
		} else if typeStr, ok := info["type"].(string); ok {
			if typeStr == "movie" {
				mtype = types.MediaTypeMovie
			} else {
				mtype = types.MediaTypeTV
			}
		} else {
			// 根据标题判断类型
			if meta.BeginSeason != "" {
				mtype = types.MediaTypeTV
			} else {
				mtype = types.MediaTypeMovie
			}
		}
		
		// 海报
		posterPath := ""
		if pic, ok := info["pic"].(map[string]interface{}); ok {
			if large, ok := pic["large"].(string); ok {
				posterPath = large
			}
		}
		
		if posterPath == "" {
			if coverURL, ok := info["cover_url"].(string); ok {
				posterPath = coverURL
			}
		}
		
		if posterPath == "" {
			if cover, ok := info["cover"].(map[string]interface{}); ok {
				if url, ok := cover["url"].(string); ok {
					posterPath = url
				}
			}
		}
		
		cacheData = &CacheData{
			ID:         getStringValue(info, "id"),
			Type:       mtype,
			Year:       cacheYear,
			Title:      cacheTitle,
			PosterPath: posterPath,
		}
	} else if info != nil {
		// None时不缓存，此时代表网络错误，允许重复请求
		cacheData = &CacheData{
			ID: "0",
		}
	}
	
	if cacheData != nil {
		dc.mutex.Lock()
		defer dc.mutex.Unlock()
		dc.cache.Set(dc.getKey(meta), cacheData)
	}
}

// Save 保存缓存数据到文�?func (dc *DoubanCache) Save(force bool) {
	// Redis不需要保存到本地文件
	if dc.cache.IsRedis() {
		return
	}
	
	// TODO: 实现保存逻辑
	logger.Debug("保存豆瓣缓存到文�?)
}

// coalesceString 返回第一个非空字符串
func coalesceString(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}

// getStringValue 从map中安全获取字符串�?func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
