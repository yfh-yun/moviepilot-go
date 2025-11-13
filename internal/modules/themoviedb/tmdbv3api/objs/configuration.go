package objs

import (
	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Configuration 配置对象
type Configuration struct {
	tmdb *tmdbv3api.TMDb
}

// NewConfiguration 创建Configuration实例
func NewConfiguration(tmdb *tmdbv3api.TMDb) *Configuration {
	return &Configuration{
		tmdb: tmdb,
	}
}

// APIConfiguration 获取系统范围的配置信�?/*
Get the system wide configuration info.
*/
func (c *Configuration) APIConfiguration() (map[string]interface{}, error) {
	action := "/configuration"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Countries 获取TMDb中使用的国家列表(ISO 3166-1标签)
/*
Get the list of countries (ISO 3166-1 tags) used throughout TMDb.
*/
func (c *Configuration) Countries() ([]interface{}, error) {
	action := "/configuration/countries"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Jobs 获取TMDb中使用的工作和部门列�?/*
Get a list of the jobs and departments we use on TMDb.
*/
func (c *Configuration) Jobs() ([]interface{}, error) {
	action := "/configuration/jobs"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Languages 获取TMDb中使用的语言列表(ISO 639-1标签)
/*
Get the list of languages (ISO 639-1 tags) used throughout TMDb.
*/
func (c *Configuration) Languages() ([]interface{}, error) {
	action := "/configuration/languages"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// PrimaryTranslations 获取TMDb官方支持的翻译列�?/*
Get a list of the officially supported translations on TMDb.
*/
func (c *Configuration) PrimaryTranslations() ([]interface{}, error) {
	action := "/configuration/primary_translations"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Timezones 获取TMDb中使用的时区列表
/*
Get the list of timezones used throughout TMDb.
*/
func (c *Configuration) Timezones() ([]interface{}, error) {
	action := "/configuration/timezones"
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
