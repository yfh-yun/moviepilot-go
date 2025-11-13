package objs

import (
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Collection 合集对象
type Collection struct {
	tmdb *tmdbv3api.TMDb
}

// NewCollection 创建Collection实例
func NewCollection(tmdb *tmdbv3api.TMDb) *Collection {
	return &Collection{
		tmdb: tmdb,
	}
}

// Details 获取合集详情
/*
Get collection details by id.
*/
func (c *Collection) Details(collectionID int) (map[string]interface{}, error) {
	action := "/collection/" + strconv.Itoa(collectionID)
	key := "parts"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取合集图片
/*
Get the images for a collection by id.
*/
func (c *Collection) Images(collectionID int) (map[string]interface{}, error) {
	action := "/collection/" + strconv.Itoa(collectionID) + "/images"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Translations 获取合集翻译列表
/*
Get the list translations for a collection by id.
*/
func (c *Collection) Translations(collectionID int) ([]interface{}, error) {
	action := "/collection/" + strconv.Itoa(collectionID) + "/translations"
	key := "translations"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
