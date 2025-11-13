package objs

import (
	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Certification 认证对象
type Certification struct {
	tmdb *tmdbv3api.TMDb
}

// NewCertification 创建Certification实例
func NewCertification(tmdb *tmdbv3api.TMDb) *Certification {
	return &Certification{
		tmdb: tmdb,
	}
}

// MovieList 获取最新的电影认证列表
/*
Get an up to date list of the officially supported movie certifications on TMDB.
*/
func (c *Certification) MovieList() (map[string]interface{}, error) {
	key := "certifications"
	result, err := c.tmdb.RequestObj("/certification/movie/list", "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TVList 获取最新的电视节目认证列表
/*
Get an up to date list of the officially supported TV show certifications on TMDB.
*/
func (c *Certification) TVList() (map[string]interface{}, error) {
	key := "certifications"
	result, err := c.tmdb.RequestObj("/certification/tv/list", "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
