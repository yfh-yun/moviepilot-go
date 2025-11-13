package objs

import (
	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Provider 提供商对�?type Provider struct {
	tmdb *tmdbv3api.TMDb
}

// NewProvider 创建Provider实例
func NewProvider(tmdb *tmdbv3api.TMDb) *Provider {
	return &Provider{
		tmdb: tmdb,
	}
}

// AvailableRegions 返回所有我们有观看提供�?OTT/流媒�?数据的国家列�?/*
Returns a list of all of the countries we have watch provider (OTT/streaming) data for.
*/
func (p *Provider) AvailableRegions() ([]interface{}, error) {
	action := "/watch/providers/regions"
	key := "results"
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// MovieProviders 返回我们可用的电影观看提供商(OTT/流媒�?数据列表
/*
Returns a list of the watch provider (OTT/streaming) data we have available for movies.
*/
func (p *Provider) MovieProviders(region string) ([]interface{}, error) {
	action := "/watch/providers/movie"
	params := ""
	if region != "" {
		params = "watch_region=" + region
	}
	key := "results"
	
	result, err := p.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TvProviders 返回我们可用的电视节目观看提供商(OTT/流媒�?数据列表
/*
Returns a list of the watch provider (OTT/streaming) data we have available for TV series.
*/
func (p *Provider) TvProviders(region string) ([]interface{}, error) {
	action := "/watch/providers/tv"
	params := ""
	if region != "" {
		params = "watch_region=" + region
	}
	key := "results"
	
	result, err := p.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
