package themoviedb

import (
	"fmt"
	"path"
	"strings"
	
	"moviepilot-go/pkg/config"
	"moviepilot-go/pkg/models"
)

// TmdbScraper TMDB刮削�?type TmdbScraper struct {
	metaTmdb *TmdbApi
	imgTmdb  *TmdbApi
}

// NewTmdbScraper 创建TmdbScraper实例
func NewTmdbScraper() *TmdbScraper {
	return &TmdbScraper{}
}

// DefaultTmdb 获取元数据TMDB Api
func (t *TmdbScraper) DefaultTmdb() *TmdbApi {
	if t.metaTmdb == nil {
		t.metaTmdb = NewTmdbApi(config.Config.TMDB_LOCALE)
	}
	return t.metaTmdb
}

// OriginalTmdb 获取图片TMDB Api
func (t *TmdbScraper) OriginalTmdb(mediainfo *models.MediaInfo) *TmdbApi {
	if config.Config.TMDB_SCRAP_ORIGINAL_IMAGE && mediainfo != nil {
		// 注意：这里简化处理，实际应该根据原始语言创建对应的TmdbApi
		return NewTmdbApi(mediainfo.OriginalLanguage)
	}
	return t.DefaultTmdb()
}

// GetMetadataNfo 获取NFO文件内容文本
func (t *TmdbScraper) GetMetadataNfo(meta *models.MetaBase, mediainfo *models.MediaInfo, season, episode *int) string {
	// 这里应该生成NFO XML内容，但为了简化，我们只返回一个占位符
	return fmt.Sprintf("NFO content for %s", mediainfo.Title)
}

// GetMetadataImg 获取图片名称和url
func (t *TmdbScraper) GetMetadataImg(mediainfo *models.MediaInfo, season, episode *int) map[string]string {
	images := make(map[string]string)
	
	if season != nil {
		// 只需要季集的图片
		if episode != nil {
			// 集的图片
			// TODO: 实现集图片获取逻辑
		} else {
			// 季的图片
			// TODO: 实现季图片获取逻辑
		}
		return images
	} else {
		// 获取媒体信息中原有图�?		val := reflectValue(mediainfo)
		for key, value := range val {
			if strings.HasSuffix(key, "_path") && value != nil {
				if str, ok := value.(string); ok && strings.HasPrefix(str, "http") {
					imageName := strings.Replace(key, "_path", "", -1) + path.Ext(str)
					images[imageName] = str
				}
			}
		}
		
		// 替换原语言Poster
		if config.Config.TMDB_SCRAP_ORIGINAL_IMAGE {
			// TODO: 实现原语言图片获取逻辑
		}
	}
	
	return images
}

// reflectValue 简化反射获取结构体字段�?func reflectValue(obj interface{}) map[string]interface{} {
	// 这里简化处理，实际应该使用反射获取所有字�?	// 为了演示目的，返回空map
	return make(map[string]interface{})
}

// GetSeasonPoster 获取季的海报
func (t *TmdbScraper) GetSeasonPoster(seasoninfo map[string]interface{}, season int) (string, string) {
	// TMDB季poster图片
	seaSeq := fmt.Sprintf("%02d", season)
	
	if posterPath, exists := seasoninfo["poster_path"].(string); exists && posterPath != "" {
		// 后缀
		ext := path.Ext(posterPath)
		// URL
		url := fmt.Sprintf("https://%s/t/p/original%s", config.Config.TMDB_IMAGE_DOMAIN, posterPath)
		// S0海报格式不同
		var imageName string
		if season == 0 {
			imageName = fmt.Sprintf("season-specials-poster%s", ext)
		} else {
			imageName = fmt.Sprintf("season%s-poster%s", seaSeq, ext)
		}
		return imageName, url
	}
	
	return "", ""
}
