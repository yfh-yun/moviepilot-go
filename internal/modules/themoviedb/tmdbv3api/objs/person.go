package objs

import (
	"fmt"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Person 人物对象
type Person struct {
	tmdb *tmdbv3api.TMDb
}

// NewPerson 创建Person实例
func NewPerson(tmdb *tmdbv3api.TMDb) *Person {
	return &Person{
		tmdb: tmdb,
	}
}

// Details 获取人物的主要详�?/*
Get the primary person details by id.
*/
func (p *Person) Details(personID int, appendToResponse string) (map[string]interface{}, error) {
	action := "/person/" + strconv.Itoa(personID)
	params := ""
	if appendToResponse != "" {
		params = "append_to_response=" + appendToResponse
	}
	
	result, err := p.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Changes 获取人物的更改记�?/*
Get the changes for a person. By default only the last 24 hours are returned.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
*/
func (p *Person) Changes(personID int, startDate, endDate string, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/person/%d/changes", personID)
	params := fmt.Sprintf("page=%d", page)
	if startDate != "" {
		params += "&start_date=" + startDate
	}
	if endDate != "" {
		params += "&end_date=" + endDate
	}
	key := "changes"
	
	result, err := p.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// MovieCredits 获取人物的电影作�?/*
Get the movie credits for a person.
*/
func (p *Person) MovieCredits(personID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/person/%d/movie_credits", personID)
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TvCredits 获取人物的电视节目作�?/*
Get the TV show credits for a person.
*/
func (p *Person) TvCredits(personID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/person/%d/tv_credits", personID)
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// CombinedCredits 获取人物的电影和电视节目作品
/*
Get the movie and TV credits together in a single response.
*/
func (p *Person) CombinedCredits(personID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/person/%d/combined_credits", personID)
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// ExternalIDs 获取人物的外部ID
/*
Get the external ids for a person. We currently support the following external sources.
IMDB ID, Facebook, Freebase MID, Freebase ID, Instagram, TVRage ID, and Twitter
*/
func (p *Person) ExternalIDs(personID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/person/%d/external_ids", personID)
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Images 获取人物的图�?/*
Get the images for a person.
*/
func (p *Person) Images(personID int) ([]interface{}, error) {
	action := fmt.Sprintf("/person/%d/images", personID)
	key := "profiles"
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TaggedImages 获取人物被标记的图片
/*
Get the images that this person has been tagged in.
*/
func (p *Person) TaggedImages(personID int, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/person/%d/tagged_images", personID)
	params := fmt.Sprintf("page=%d", page)
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

// Translations 获取人物的翻译列�?/*
Get a list of translations that have been created for a person.
*/
func (p *Person) Translations(personID int) ([]interface{}, error) {
	action := fmt.Sprintf("/person/%d/translations", personID)
	key := "translations"
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Latest 获取最新创建的人物
/*
Get the most newly created person. This is a live response and will continuously change.
*/
func (p *Person) Latest() (map[string]interface{}, error) {
	action := "/person/latest"
	
	result, err := p.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Popular 获取热门人物列表
/*
Get the list of popular people on TMDb. This list updates daily.
*/
func (p *Person) Popular(page int) ([]interface{}, error) {
	action := "/person/popular"
	params := fmt.Sprintf("page=%d", page)
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
