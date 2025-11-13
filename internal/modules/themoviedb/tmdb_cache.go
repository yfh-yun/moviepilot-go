package themoviedb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"moviepilot-go/pkg/config"
	"moviepilot-go/pkg/models"
	"moviepilot-go/internal/logger"
)

// TmdbCache TMDB缓存数据
type TmdbCache struct {
	cache       map[string]CacheItem
	mutex       sync.RWMutex
	metaFilepath string
}

// CacheItem 缓存条目结构
type CacheItem struct {
	ID           int         `json:"id"`
	Title        string      `json:"title"`
	Year         string      `json:"year"`
	Type         string      `json:"type"`
	PosterPath   string      `json:"poster_path"`
	BackdropPath string      `json:"backdrop_path"`
	CreatedAt    time.Time   `json:"created_at"`
}

// NewTmdbCache 创建TmdbCache实例
func NewTmdbCache() *TmdbCache {
	tc := &TmdbCache{
		cache:       make(map[string]CacheItem),
		metaFilepath: filepath.Join(config.Config.TEMP_PATH, "__tmdb_cache__"),
	}
	
	// 加载本地缓存数据
	tc.loadCache()
	
	return tc
}

// loadCache 从文件中加载缓存
func (tc *TmdbCache) loadCache() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	
	if _, err := os.Stat(tc.metaFilepath); os.IsNotExist(err) {
		return
	}
	
	data, err := ioutil.ReadFile(tc.metaFilepath)
	if err != nil {
		logger.Error(fmt.Sprintf("加载TMDB缓存失败: %v", err))
		return
	}
	
	var cache map[string]CacheItem
	if err := json.Unmarshal(data, &cache); err != nil {
		logger.Error(fmt.Sprintf("解析TMDB缓存失败: %v", err))
		return
	}
	
	tc.cache = cache
}

// Clear 清空所有TMDB缓存
func (tc *TmdbCache) Clear() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	
	tc.cache = make(map[string]CacheItem)
}

// getCacheKey 获取缓存KEY
func (tc *TmdbCache) getCacheKey(meta *models.MetaBase) string {
	mediaType := "未知"
	if meta.Type != "" {
		mediaType = string(meta.Type)
	}
	
	return fmt.Sprintf("[%s]%v-%s-%d", mediaType, meta.Tmdbid, meta.Name, meta.Year, meta.BeginSeason)
}

// Get 根据KEY值获取缓存�?func (tc *TmdbCache) Get(meta *models.MetaBase) map[string]interface{} {
	key := tc.getCacheKey(meta)
	
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	
	if item, exists := tc.cache[key]; exists {
		result := make(map[string]interface{})
		result["id"] = item.ID
		result["title"] = item.Title
		result["year"] = item.Year
		result["type"] = item.Type
		result["poster_path"] = item.PosterPath
		result["backdrop_path"] = item.BackdropPath
		return result
	}
	
	return make(map[string]interface{})
}

// Update 新增或更新缓存条�?func (tc *TmdbCache) Update(meta *models.MetaBase, info map[string]interface{}) {
	key := tc.getCacheKey(meta)
	
	if len(info) > 0 {
		// 缓存标题
		var cacheTitle string
		if mediaType, ok := info["media_type"].(string); ok && mediaType == "movie" {
			if title, exists := info["title"].(string); exists {
				cacheTitle = title
			}
		} else {
			if name, exists := info["name"].(string); exists {
				cacheTitle = name
			}
		}
		
		// 缓存年份
		var cacheYear string
		if mediaType, ok := info["media_type"].(string); ok && mediaType == "movie" {
			if releaseDate, exists := info["release_date"].(string); exists && len(releaseDate) >= 4 {
				cacheYear = releaseDate[:4]
			}
		} else {
			if firstAirDate, exists := info["first_air_date"].(string); exists && len(firstAirDate) >= 4 {
				cacheYear = firstAirDate[:4]
			}
		}
		
		tc.mutex.Lock()
		defer tc.mutex.Unlock()
		
		// 缓存数据
		item := CacheItem{
			ID:           0,
			Title:        cacheTitle,
			Year:         cacheYear,
			Type:         "",
			PosterPath:   "",
			BackdropPath: "",
			CreatedAt:    time.Now(),
		}
		
		if id, ok := info["id"].(float64); ok {
			item.ID = int(id)
		}
		
		if mediaType, ok := info["media_type"].(string); ok {
			item.Type = mediaType
		}
		
		if posterPath, ok := info["poster_path"].(string); ok {
			item.PosterPath = posterPath
		}
		
		if backdropPath, ok := info["backdrop_path"].(string); ok {
			item.BackdropPath = backdropPath
		}
		
		tc.cache[key] = item
	} else if info != nil {
		// None时不缓存，此时代表网络错误，允许重复请求
		tc.mutex.Lock()
		defer tc.mutex.Unlock()
		
		item := CacheItem{
			ID:        0,
			CreatedAt: time.Now(),
		}
		
		tc.cache[key] = item
	}
}

// Save 保存缓存数据到文�?func (tc *TmdbCache) Save(force bool) {
	tc.mutex.RLock()
	cacheCopy := make(map[string]CacheItem)
	for k, v := range tc.cache {
		cacheCopy[k] = v
	}
	tc.mutex.RUnlock()
	
	// 过滤无法识别的条�?	newCacheData := make(map[string]CacheItem)
	for k, v := range cacheCopy {
		if v.ID > 0 {
			newCacheData[k] = v
		}
	}
	
	// 读取现有文件数据
	var fileData map[string]CacheItem
	if _, err := os.Stat(tc.metaFilepath); err == nil {
		if data, err := ioutil.ReadFile(tc.metaFilepath); err == nil {
			json.Unmarshal(data, &fileData)
		}
	}
	
	// 如果不强制保存且数据未变化，则不保存
	if !force {
		if len(fileData) == len(newCacheData) {
			same := true
			for k, v := range newCacheData {
				if fileV, exists := fileData[k]; !exists || fileV != v {
					same = false
					break
				}
			}
			if same {
				return
			}
		}
	}
	
	// 保存到文�?	if data, err := json.Marshal(newCacheData); err == nil {
		ioutil.WriteFile(tc.metaFilepath, data, 0644)
	}
}

// Delete 删除缓存信息
func (tc *TmdbCache) Delete(key string) map[string]interface{} {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	
	if item, exists := tc.cache[key]; exists {
		delete(tc.cache, key)
		
		result := make(map[string]interface{})
		result["id"] = item.ID
		result["title"] = item.Title
		result["year"] = item.Year
		result["type"] = item.Type
		result["poster_path"] = item.PosterPath
		result["backdrop_path"] = item.BackdropPath
		return result
	}
	
	return make(map[string]interface{})
}

// Modify 修改缓存信息
func (tc *TmdbCache) Modify(key string, title string) map[string]interface{} {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	
	if item, exists := tc.cache[key]; exists {
		item.Title = title
		tc.cache[key] = item
		
		result := make(map[string]interface{})
		result["id"] = item.ID
		result["title"] = item.Title
		result["year"] = item.Year
		result["type"] = item.Type
		result["poster_path"] = item.PosterPath
		result["backdrop_path"] = item.BackdropPath
		return result
	}
	
	return make(map[string]interface{})
}
