package themoviedb

import (
	"fmt"
	"sort"
	"strings"
	
	"moviepilot-go/pkg/config"
	"moviepilot-go/pkg/models"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
)

// TheMovieDbModule TMDB媒体信息匹配模块
type TheMovieDbModule struct {
	cache     *TmdbCache
	tmdb      *TmdbApi
	category  *CategoryHelper
	scraper   *TmdbScraper
}

// NewTheMovieDbModule 创建TheMovieDbModule实例
func NewTheMovieDbModule() *TheMovieDbModule {
	return &TheMovieDbModule{}
}

// InitModule 初始化模�?func (t *TheMovieDbModule) InitModule() error {
	t.cache = NewTmdbCache()
	t.tmdb = NewTmdbApi("")
	t.category = NewCategoryHelper()
	t.scraper = NewTmdbScraper()
	return nil
}

// GetName 获取模块名称
func (t *TheMovieDbModule) GetName() string {
	return "TheMovieDb"
}

// GetType 获取模块类型
func (t *TheMovieDbModule) GetType() models.ModuleType {
	return models.MediaRecognize
}

// GetSubtype 获取模块子类�?func (t *TheMovieDbModule) GetSubtype() models.MediaRecognizeType {
	return models.TMDB
}

// GetPriority 获取模块优先�?func (t *TheMovieDbModule) GetPriority() int {
	return 1
}

// Stop 停止模块
func (t *TheMovieDbModule) Stop() {
	t.cache.Save(false)
	// TMDB API不需要特殊关闭处�?}

// Test 测试模块连接�?func (t *TheMovieDbModule) Test() (bool, string) {
	url := fmt.Sprintf("https://%s/3/movie/550?api_key=%s", config.Config.TMDB_API_DOMAIN, config.Config.TMDB_API_KEY)
	
	response, err := utils.RequestUtils.GetRes(url, nil, nil, 0)
	if err != nil {
		return false, fmt.Sprintf("%s 网络连接失败", config.Config.TMDB_API_DOMAIN)
	}
	
	if response.StatusCode == 200 {
		return true, ""
	}
	
	return false, fmt.Sprintf("无法连接 %s，错误码�?d", config.Config.TMDB_API_DOMAIN, response.StatusCode)
}

// InitSetting 初始化设�?func (t *TheMovieDbModule) InitSetting() (string, interface{}) {
	return "", nil
}

// RecognizeMedia 识别媒体信息
func (t *TheMovieDbModule) RecognizeMedia(meta *models.MetaBase, mtype models.MediaType, tmdbid *int, episodeGroup *string, cache bool, kwargs map[string]interface{}) *models.MediaInfo {
	// 验证参数
	if !t.validateRecognizeParams(meta, tmdbid) {
		return nil
	}
	
	var cacheInfo map[string]interface{}
	if meta == nil {
		// 未提供元数据时，直接使用tmdbid查询，不使用缓存
		cacheInfo = make(map[string]interface{})
	} else {
		// 读取缓存
		if mtype != "" {
			meta.Type = mtype
		}
		if tmdbid != nil {
			meta.Tmdbid = *tmdbid
		}
		cacheInfo = t.cache.Get(meta)
	}
	
	// 查询剧集�?	groupSeasons := []map[string]interface{}{}
	if episodeGroup != nil && *episodeGroup != "" {
		// 这里需要实现获取剧集组的逻辑
	}
	
	// 识别匹配
	var info map[string]interface{}
	if len(cacheInfo) == 0 || !cache {
		// 缓存没有或者强制不使用缓存
		if tmdbid != nil {
			// 直接查询详情
			info = t.tmdb.GetInfo(mtype, *tmdbid)
		}
		
		if len(info) == 0 && meta != nil {
			// 准备搜索名称
			names := t.prepareSearchNames(meta)
			for _, name := range names {
				info = t.searchByName(name, meta, groupSeasons)
				if len(info) == 0 {
					// 从网站查询的逻辑需要实�?				}
				if len(info) > 0 {
					// 查到就退�?					break
				}
			}
			
			// 补充全量信息
			if len(info) > 0 {
				if _, hasGenres := info["genres"]; !hasGenres {
					if mediaType, ok := info["media_type"].(string); ok {
						var mediaTypeEnum models.MediaType
						if mediaType == "movie" {
							mediaTypeEnum = models.MediaTypeMovie
						} else {
							mediaTypeEnum = models.MediaTypeTV
						}
						
						if id, ok := info["id"].(float64); ok {
							info = t.tmdb.GetInfo(mediaTypeEnum, int(id))
						}
					}
				}
			}
		} else if len(info) == 0 {
			logger.Error("识别媒体信息时未提供元数据或唯一且有效的tmdbid")
			return nil
		}
		
		// 保存到缓�?		if meta != nil {
			t.cache.Update(meta, info)
		}
	} else {
		// 使用缓存信息
		if title, ok := cacheInfo["title"].(string); ok && title != "" {
			logger.Info(fmt.Sprintf("%s 使用TMDB识别缓存�?s", meta.Name, title))
			if mediaType, ok := cacheInfo["type"].(string); ok {
				var mediaTypeEnum models.MediaType
				if mediaType == "movie" {
					mediaTypeEnum = models.MediaTypeMovie
				} else {
					mediaTypeEnum = models.MediaTypeTV
				}
				
				if id, ok := cacheInfo["id"].(float64); ok {
					info = t.tmdb.GetInfo(mediaTypeEnum, int(id))
				}
			}
		} else {
			logger.Info(fmt.Sprintf("%s 使用TMDB识别缓存：无法识�?, meta.Name))
			info = make(map[string]interface{})
		}
	}
	
	if len(info) > 0 {
		return t.buildMediaInfoResult(info, meta, tmdbid, episodeGroup, groupSeasons)
	} else {
		if meta != nil {
			logger.Info(fmt.Sprintf("%s 未匹配到TMDB媒体信息", meta.Name))
		} else if tmdbid != nil {
			logger.Info(fmt.Sprintf("%d 未匹配到TMDB媒体信息", *tmdbid))
		}
	}
	
	return nil
}

// validateRecognizeParams 验证识别参数
func (t *TheMovieDbModule) validateRecognizeParams(meta *models.MetaBase, tmdbid *int) bool {
	if tmdbid == nil && meta == nil {
		return false
	}
	
	if meta != nil && tmdbid == nil && config.Config.RECOGNIZE_SOURCE != "themoviedb" {
		return false
	}
	
	if meta != nil && meta.Name == "" {
		logger.Warn("识别媒体信息时未提供元数据名�?)
		return false
	}
	
	return true
}

// prepareSearchNames 准备搜索名称列表
func (t *TheMovieDbModule) prepareSearchNames(meta *models.MetaBase) []string {
	// 使用中英文名分别识别，去重去空，但要保持顺序
	names := []string{}
	
	if meta.CnName != "" {
		names = append(names, meta.CnName)
	}
	
	// 简体名称处理需要实�?	
	if meta.EnName != "" {
		names = append(names, meta.EnName)
	}
	
	// 去重
	uniqueNames := []string{}
	seen := make(map[string]bool)
	for _, name := range names {
		if name != "" && !seen[name] {
			seen[name] = true
			uniqueNames = append(uniqueNames, name)
		}
	}
	
	return uniqueNames
}

// searchByName 根据名称搜索媒体信息
func (t *TheMovieDbModule) searchByName(name string, meta *models.MetaBase, groupSeasons []map[string]interface{}) map[string]interface{} {
	if meta.BeginSeason > 0 {
		logger.Info(fmt.Sprintf("正在识别 %s �?d�?...", name, meta.BeginSeason))
	} else {
		logger.Info(fmt.Sprintf("正在识别 %s ...", name))
	}
	
	if meta.Type == models.MediaTypeUnknown && meta.Year == "" {
		return t.tmdb.SearchMultiis(name)[0] // 简化实现，实际需要更复杂的逻辑
	} else {
		if meta.Type == models.MediaTypeTV {
			// 确定是电�?			info := t.tmdb.Match(name, meta.Type, meta.Year, meta.Year, meta.BeginSeason, groupSeasons)
			if len(info) == 0 {
				// 去掉年份再查一�?				info = t.tmdb.Match(name, meta.Type, "", "", meta.BeginSeason, groupSeasons)
			}
			return info
		} else {
			// 有年份先按电影查
			info := t.tmdb.Match(name, models.MediaTypeMovie, meta.Year, "", 0, nil)
			// 没有再按电视剧查
			if len(info) == 0 {
				info = t.tmdb.Match(name, models.MediaTypeTV, meta.Year, "", 0, groupSeasons)
			}
			if len(info) == 0 {
				// 去掉年份和类型再查一�?				if results := t.tmdb.SearchMultiis(name); len(results) > 0 {
					return results[0]
				}
			}
			return info
		}
	}
}

// buildMediaInfoResult 构建MediaInfo结果
func (t *TheMovieDbModule) buildMediaInfoResult(info map[string]interface{}, meta *models.MetaBase, tmdbid *int, episodeGroup *string, groupSeasons []map[string]interface{}) *models.MediaInfo {
	// 确定二级分类
	var cat string
	if mediaType, ok := info["media_type"].(string); ok && mediaType == "tv" {
		cat = t.category.GetTVCategory(info)
	} else {
		cat = t.category.GetMovieCategory(info)
	}
	
	// 赋值TMDB信息并返�?	mediainfo := models.NewMediaInfo(info)
	mediainfo.SetCategory(cat)
	
	if meta != nil {
		logger.Info(fmt.Sprintf("%s TMDB识别结果�?s %s %v", meta.Name, mediainfo.Type, mediainfo.TitleYear, mediainfo.TmdbID))
	} else if tmdbid != nil {
		logger.Info(fmt.Sprintf("%d TMDB识别结果�?s %s", *tmdbid, mediainfo.Type, mediainfo.TitleYear))
	}
	
	// 处理剧集组信息需要实�?	
	return mediainfo
}

// SearchMedias 搜索媒体信息
func (t *TheMovieDbModule) SearchMedias(meta *models.MetaBase) []*models.MediaInfo {
	if config.Config.SEARCH_SOURCE != "" && !strings.Contains(config.Config.SEARCH_SOURCE, "themoviedb") {
		return nil
	}
	
	if meta.Name == "" {
		return []*models.MediaInfo{}
	}
	
	var results []map[string]interface{}
	
	if meta.Type == models.MediaTypeUnknown && meta.Year == "" {
		results = t.tmdb.SearchMultiis(meta.Name)
	} else {
		if meta.Type == models.MediaTypeUnknown {
			movieResults, _ := t.tmdb.search.SearchMovies(meta.Name, meta.Year)
			tvResults, _ := t.tmdb.search.SearchTVShows(meta.Name, meta.Year)
			results = append(movieResults, tvResults...)
			
			// 组合结果的情况下要排�?			sort.Slice(results, func(i, j int) bool {
				date1 := ""
				if releaseDate, ok := results[i]["release_date"].(string); ok {
					date1 = releaseDate
				} else if firstAirDate, ok := results[i]["first_air_date"].(string); ok {
					date1 = firstAirDate
				}
				
				date2 := ""
				if releaseDate, ok := results[j]["release_date"].(string); ok {
					date2 = releaseDate
				} else if firstAirDate, ok := results[j]["first_air_date"].(string); ok {
					date2 = firstAirDate
				}
				
				return date1 > date2
			})
		} else if meta.Type == models.MediaTypeMovie {
			results, _ = t.tmdb.search.SearchMovies(meta.Name, meta.Year)
		} else {
			results, _ = t.tmdb.search.SearchTVShows(meta.Name, meta.Year)
		}
	}
	
	// 将搜索词中的季写入标题中
	if len(results) > 0 {
		medias := []*models.MediaInfo{}
		for _, info := range results {
			media := models.NewMediaInfo(info)
			medias = append(medias, media)
		}
		
		if meta.BeginSeason > 0 {
			// 季号处理需要实�?		}
		
		return medias
	}
	
	return []*models.MediaInfo{}
}

// TmdbInfo 获取TMDB信息
func (t *TheMovieDbModule) TmdbInfo(tmdbid int, mtype models.MediaType, season *int) map[string]interface{} {
	if season == nil {
		return t.tmdb.GetInfo(mtype, tmdbid)
	} else {
		// 获取季详情需要实�?		return make(map[string]interface{})
	}
}

// MediaCategory 获取媒体分类
func (t *TheMovieDbModule) MediaCategory() map[string][]string {
	return map[string][]string{
		string(models.MediaTypeMovie): t.category.MovieCategorys(),
		string(models.MediaTypeTV):    t.category.TVCategorys(),
	}
}

// ClearCache 清除缓存
func (t *TheMovieDbModule) ClearCache() {
	logger.Info("开始清除TMDB缓存 ...")
	// tmdb.clear_cache() 在Go中不需要实�?	t.cache.Clear()
	logger.Info("TMDB缓存清除完成")
}
