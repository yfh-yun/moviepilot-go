package themoviedb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	
	"moviepilot-go/pkg/config"
	"moviepilot-go/internal/logger"
	"gopkg.in/yaml.v2"
)

// CategoryHelper 二级分类助手
type CategoryHelper struct {
	categoryPath    string
	categorys       map[string]interface{}
	movieCategorys  map[string]interface{}
	tvCategorys     map[string]interface{}
}

// NewCategoryHelper 创建CategoryHelper实例
func NewCategoryHelper() *CategoryHelper {
	ch := &CategoryHelper{
		categoryPath: filepath.Join(config.Config.CONFIG_PATH, "category.yaml"),
		categorys:    make(map[string]interface{}),
	}
	ch.Init()
	return ch
}

// Init 初始�?func (c *CategoryHelper) Init() {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn(fmt.Sprintf("二级分类策略配置文件格式出现严重错误！请检查：%v", r))
			c.categorys = make(map[string]interface{})
		}
	}()
	
	// 检查配置文件是否存�?	if _, err := os.Stat(c.categoryPath); os.IsNotExist(err) {
		// 复制默认配置文件
		defaultPath := filepath.Join(config.Config.INNER_CONFIG_PATH, "category.yaml")
		input, err := ioutil.ReadFile(defaultPath)
		if err != nil {
			logger.Warn(fmt.Sprintf("无法读取默认二级分类策略配置文件�?v", err))
			return
		}
		
		err = ioutil.WriteFile(c.categoryPath, input, 0644)
		if err != nil {
			logger.Warn(fmt.Sprintf("无法创建二级分类策略配置文件�?v", err))
			return
		}
	}
	
	// 读取配置文件
	data, err := ioutil.ReadFile(c.categoryPath)
	if err != nil {
		logger.Warn(fmt.Sprintf("二级分类策略配置文件加载出错�?v", err))
		return
	}
	
	// 解析YAML
	err = yaml.Unmarshal(data, &c.categorys)
	if err != nil {
		logger.Warn(fmt.Sprintf("二级分类策略配置文件解析出错�?v", err))
		return
	}
	
	// 提取电影和电视剧分类
	if movieCat, ok := c.categorys["movie"].(map[interface{}]interface{}); ok {
		c.movieCategorys = convertMap(movieCat)
	}
	
	if tvCat, ok := c.categorys["tv"].(map[interface{}]interface{}); ok {
		c.tvCategorys = convertMap(tvCat)
	}
	
	logger.Info("已加载二级分类策�?category.yaml")
}

// convertMap 转换map[interface{}]interface{}为map[string]interface{}
func convertMap(input map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range input {
		if strKey, ok := key.(string); ok {
			result[strKey] = value
		}
	}
	return result
}

// IsMovieCategory 获取电影分类标志
func (c *CategoryHelper) IsMovieCategory() bool {
	return len(c.movieCategorys) > 0
}

// IsTVCategory 获取电视剧分类标�?func (c *CategoryHelper) IsTVCategory() bool {
	return len(c.tvCategorys) > 0
}

// MovieCategorys 获取电影分类清单
func (c *CategoryHelper) MovieCategorys() []string {
	if len(c.movieCategorys) == 0 {
		return []string{}
	}
	
	keys := make([]string, 0, len(c.movieCategorys))
	for key := range c.movieCategorys {
		keys = append(keys, key)
	}
	
	sort.Strings(keys)
	return keys
}

// TVCategorys 获取电视剧分类清�?func (c *CategoryHelper) TVCategorys() []string {
	if len(c.tvCategorys) == 0 {
		return []string{}
	}
	
	keys := make([]string, 0, len(c.tvCategorys))
	for key := range c.tvCategorys {
		keys = append(keys, key)
	}
	
	sort.Strings(keys)
	return keys
}

// GetMovieCategory 判断电影的分�?func (c *CategoryHelper) GetMovieCategory(tmdbInfo map[string]interface{}) string {
	return c.getCategory(c.movieCategorys, tmdbInfo)
}

// GetTVCategory 判断电视剧的分类，包括动�?func (c *CategoryHelper) GetTVCategory(tmdbInfo map[string]interface{}) string {
	return c.getCategory(c.tvCategorys, tmdbInfo)
}

// getCategory 根据 TMDB信息与分类配置文件进行比较，确定所属分�?func (c *CategoryHelper) getCategory(categorys map[string]interface{}, tmdbInfo map[string]interface{}) string {
	if len(tmdbInfo) == 0 || len(categorys) == 0 {
		return ""
	}
	
	for key, item := range categorys {
		categoryMap, ok := item.(map[string]interface{})
		if !ok || len(categoryMap) == 0 {
			return key
		}
		
		matchFlag := true
		for attr, value := range categoryMap {
			if value == nil {
				continue
			}
			
			var infoValue interface{}
			if attr == "release_year" {
				// 发行年份
				if releaseDate, exists := tmdbInfo["release_date"]; exists {
					infoValue = releaseDate
				} else if firstAirDate, exists := tmdbInfo["first_air_date"]; exists {
					infoValue = firstAirDate
				}
				
				if infoValue != nil {
					if str, ok := infoValue.(string); ok && len(str) >= 4 {
						infoValue = str[0:4]
					}
				}
			} else {
				infoValue = tmdbInfo[attr]
			}
			
			if infoValue == nil {
				matchFlag = false
				continue
			}
			
			var infoValues []string
			if attr == "production_countries" {
				// 制片国家
				if countries, ok := infoValue.([]interface{}); ok {
					for _, country := range countries {
						if countryMap, ok := country.(map[string]interface{}); ok {
							if isoCode, exists := countryMap["iso_3166_1"]; exists {
								if str, ok := isoCode.(string); ok {
									infoValues = append(infoValues, strings.ToUpper(str))
								}
							}
						}
					}
				}
			} else {
				if arr, ok := infoValue.([]interface{}); ok {
					for _, v := range arr {
						if str, ok := v.(string); ok {
							infoValues = append(infoValues, strings.ToUpper(str))
						}
					}
				} else if str, ok := infoValue.(string); ok {
					infoValues = []string{strings.ToUpper(str)}
				}
			}
			
			var values []string
			var invertValues []string
			
			// 如果�?"," 进行分割
			if str, ok := value.(string); ok {
				values = strings.Split(str, ",")
				for i, v := range values {
					values[i] = strings.TrimSpace(v)
				}
			}
			
			// 处理取反�?			tempValues := []string{}
			tempInvertValues := []string{}
			for _, v := range values {
				if strings.HasPrefix(v, "!") {
					tempInvertValues = append(tempInvertValues, strings.ToUpper(v[1:]))
				} else {
					tempValues = append(tempValues, strings.ToUpper(v))
				}
			}
			
			values = tempValues
			invertValues = tempInvertValues
			
			// 检查匹�?			if len(values) > 0 {
				matchFound := false
				for _, infoVal := range infoValues {
					for _, val := range values {
						if infoVal == val {
							matchFound = true
							break
						}
					}
					if matchFound {
						break
					}
				}
				if !matchFound {
					matchFlag = false
				}
			}
			
			if len(invertValues) > 0 {
				for _, infoVal := range infoValues {
					for _, invVal := range invertValues {
						if infoVal == invVal {
							matchFlag = false
							break
						}
					}
					if !matchFlag {
						break
					}
				}
			}
		}
		
		if matchFlag {
			return key
		}
	}
	
	return ""
}
