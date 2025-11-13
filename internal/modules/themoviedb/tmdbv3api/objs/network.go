package objs

import (
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Network 电视网络对象
type Network struct {
	tmdb *tmdbv3api.TMDb
}

// NewNetwork 创建Network实例
func NewNetwork(tmdb *tmdbv3api.TMDb) *Network {
	return &Network{
		tmdb: tmdb,
	}
}

// Details 通过ID获取电视网络详情
/*
Get a networks details by id.
*/
func (n *Network) Details(networkID int) (map[string]interface{}, error) {
	action := "/network/" + strconv.Itoa(networkID)
	
	result, err := n.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AlternativeNames 获取电视网络的备选名�?/*
Get the alternative names of a network.
*/
func (n *Network) AlternativeNames(networkID int) ([]interface{}, error) {
	action := "/network/" + strconv.Itoa(networkID) + "/alternative_names"
	key := "results"
	
	result, err := n.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取电视网络的Logo图片
/*
Get the TV network logos by id.
*/
func (n *Network) Images(networkID int) ([]interface{}, error) {
	action := "/network/" + strconv.Itoa(networkID) + "/images"
	key := "logos"
	
	result, err := n.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
