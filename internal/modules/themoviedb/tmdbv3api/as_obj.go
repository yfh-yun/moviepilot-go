package tmdbv3api

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// AsObj 通用对象结构�?type AsObj struct {
	jsonData      interface{}
	key           *string
	dictKey       bool
	dictKeyName   *string
	objList       []interface{}
	listOnly      bool
	dictData      map[string]interface{}
}

// NewAsObj 创建AsObj实例
func NewAsObj(jsonData interface{}, key *string, dictKey bool, dictKeyName *string) *AsObj {
	obj := &AsObj{
		jsonData:    jsonData,
		key:         key,
		dictKey:     dictKey,
		dictKeyName: dictKeyName,
		objList:     make([]interface{}, 0),
		dictData:    make(map[string]interface{}),
	}
	
	if jsonData == nil {
		obj.jsonData = make(map[string]interface{})
	}
	
	// 根据数据类型进行处理
	obj.processData()
	
	return obj
}

// processData 处理数据
func (a *AsObj) processData() {
	if a.jsonData == nil {
		return
	}
	
	// 检查是否为数组
	val := reflect.ValueOf(a.jsonData)
	if val.Kind() == reflect.Slice {
		a.listOnly = true
		sliceLen := val.Len()
		a.objList = make([]interface{}, sliceLen)
		for i := 0; i < sliceLen; i++ {
			item := val.Index(i).Interface()
			if isMapOrSlice(item) {
				a.objList[i] = NewAsObj(item, nil, false, nil)
			} else {
				a.objList[i] = item
			}
		}
	} else if a.dictKey {
		a.listOnly = true
		// 处理字典�?		if reflect.TypeOf(a.jsonData).Kind() == reflect.Map {
			mapVal := reflect.ValueOf(a.jsonData)
			for _, k := range mapVal.MapKeys() {
				if k.Kind() == reflect.String {
					keyStr := k.String()
					value := mapVal.MapIndex(k).Interface()
					if isMapOrSlice(value) {
						a.objList = append(a.objList, NewAsObj(map[string]interface{}{keyStr: value}, &keyStr, true, a.dictKeyName))
					} else {
						a.objList = append(a.objList, value)
					}
				}
			}
		}
	} else {
		// 处理普通对�?		if reflect.TypeOf(a.jsonData).Kind() == reflect.Map {
			mapVal := reflect.ValueOf(a.jsonData)
			for _, k := range mapVal.MapKeys() {
				if k.Kind() == reflect.String {
					keyStr := k.String()
					value := mapVal.MapIndex(k).Interface()
					var final interface{}
					if isMapOrSlice(value) {
						if a.key != nil && keyStr == *a.key {
							final = NewAsObj(value, nil, isMap(value), &keyStr)
							a.objList = append(a.objList, final)
						} else {
							final = NewAsObj(value, nil, false, nil)
						}
					} else {
						final = value
					}
					if a.dictKeyName != nil {
						a.dictData[*a.dictKeyName] = keyStr
					}
					a.dictData[keyStr] = final
				}
			}
		}
	}
}

// isMapOrSlice 检查是否为map或slice
func isMapOrSlice(data interface{}) bool {
	if data == nil {
		return false
	}
	kind := reflect.TypeOf(data).Kind()
	return kind == reflect.Map || kind == reflect.Slice
}

// isMap 检查是否为map
func isMap(data interface{}) bool {
	if data == nil {
		return false
	}
	return reflect.TypeOf(data).Kind() == reflect.Map
}

// toDict 转换为字�?func (a *AsObj) toDict() map[string]interface{} {
	return a.dictData
}

// ToDict 转换为字�?func (a *AsObj) ToDict() map[string]interface{} {
	return a.toDict()
}

// Delitem 删除项目
func (a *AsObj) Delitem(key string) {
	delete(a.dictData, key)
}

// Getitem 获取项目
func (a *AsObj) Getitem(key interface{}) interface{} {
	switch k := key.(type) {
	case int:
		if a.objList != nil && len(a.objList) > k {
			return a.objList[k]
		}
	case string:
		if val, exists := a.dictData[k]; exists {
			return val
		}
	}
	return nil
}

// Iter 迭代�?func (a *AsObj) Iter() []interface{} {
	if a.objList != nil && len(a.objList) > 0 {
		return a.objList
	}
	
	result := make([]interface{}, 0, len(a.dictData))
	for _, v := range a.dictData {
		result = append(result, v)
	}
	return result
}

// Len 获取长度
func (a *AsObj) Len() int {
	if a.objList != nil && len(a.objList) > 0 {
		return len(a.objList)
	}
	return len(a.dictData)
}

// Setitem 设置项目
func (a *AsObj) Setitem(key string, value interface{}) {
	a.dictData[key] = value
}

// String 字符串表�?func (a *AsObj) String() string {
	if a.listOnly {
		return fmt.Sprintf("%v", a.objList)
	}
	return fmt.Sprintf("%v", a.dictData)
}

// Reversed 反向迭代
func (a *AsObj) Reversed() []interface{} {
	items := a.Iter()
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items
}

// Copy 复制对象
func (a *AsObj) Copy() *AsObj {
	jsonBytes, _ := json.Marshal(a.jsonData)
	var newJSON interface{}
	json.Unmarshal(jsonBytes, &newJSON)
	
	return NewAsObj(newJSON, a.key, a.dictKey, a.dictKeyName)
}

// Get 获取�?func (a *AsObj) Get(key string, defaultValue interface{}) interface{} {
	if val, exists := a.dictData[key]; exists {
		return val
	}
	return defaultValue
}

// Items 获取所有项�?func (a *AsObj) Items() []interface{} {
	items := make([]interface{}, 0, len(a.dictData))
	for k, v := range a.dictData {
		items = append(items, []interface{}{k, v})
	}
	return items
}

// Keys 获取所有键
func (a *AsObj) Keys() []string {
	keys := make([]string, 0, len(a.dictData))
	for k := range a.dictData {
		keys = append(keys, k)
	}
	return keys
}

// Pop 弹出�?func (a *AsObj) Pop(key string, defaultValue interface{}) interface{} {
	if val, exists := a.dictData[key]; exists {
		delete(a.dictData, key)
		return val
	}
	return defaultValue
}

// PopItem 弹出项目
func (a *AsObj) PopItem() interface{} {
	for k, v := range a.dictData {
		delete(a.dictData, k)
		return []interface{}{k, v}
	}
	return nil
}

// SetDefault 设置默认�?func (a *AsObj) SetDefault(key string, defaultValue interface{}) interface{} {
	if _, exists := a.dictData[key]; !exists {
		a.dictData[key] = defaultValue
	}
	return a.dictData[key]
}

// Update 更新
func (a *AsObj) Update(entries map[string]interface{}) {
	for k, v := range entries {
		a.dictData[k] = v
	}
}

// Values 获取所有�?func (a *AsObj) Values() []interface{} {
	values := make([]interface{}, 0, len(a.dictData))
	for _, v := range a.dictData {
		values = append(values, v)
	}
	return values
}
