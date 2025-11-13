package objs

import (
	"fmt"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Company 公司对象
type Company struct {
	tmdb *tmdbv3api.TMDb
}

// NewCompany 创建Company实例
func NewCompany(tmdb *tmdbv3api.TMDb) *Company {
	return &Company{
		tmdb: tmdb,
	}
}

// Details 获取公司详情
/*
Get a companies details by id.
*/
func (c *Company) Details(companyID int) (map[string]interface{}, error) {
	action := "/company/" + strconv.Itoa(companyID)
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AlternativeNames 获取公司的备选名�?/*
Get the alternative names of a company.
*/
func (c *Company) AlternativeNames(companyID int) ([]interface{}, error) {
	action := "/company/" + strconv.Itoa(companyID) + "/alternative_names"
	key := "results"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取公司的图�?/*
Get the alternative names of a company.
*/
func (c *Company) Images(companyID int) ([]interface{}, error) {
	action := "/company/" + strconv.Itoa(companyID) + "/images"
	key := "logos"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Movies 获取公司的电�?/*
Get the movies of a company by id.
*/
func (c *Company) Movies(companyID, page int) ([]interface{}, error) {
	action := "/company/" + strconv.Itoa(companyID) + "/movies"
	params := fmt.Sprintf("page=%d", page)
	key := "results"
	
	result, err := c.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
