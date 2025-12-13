package cache

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"
)

// CachedOptions 缓存装饰器选项
// 原Python: @cached装饰器的参数
type CachedOptions struct {
	Region    string        // 缓存区域
	MaxSize   int           // 缓存最大大小
	TTL       time.Duration // 缓存有效期
	SkipNone  bool          // 跳过None值
	SkipEmpty bool          // 跳过空值
}

// defaultCachedOptions 默认缓存选项
var defaultCachedOptions = CachedOptions{
	Region:    DefaultCacheRegion,
	MaxSize:   DefaultCacheSize,
	TTL:       0, // 默认无过期时间（LRU缓存）
	SkipNone:  true,
	SkipEmpty: false,
}

// CachedFunction 封装了缓存函数和相关方法的结构体
// 原Python: 装饰器返回的函数，带有cache_region属性和cache_clear方法
type CachedFunction struct {
	OriginalFunc any           // 原始函数
	CacheRegion  string        // 缓存区域
	CacheBackend CacheBackend  // 缓存后端
	CacheOptions CachedOptions // 缓存选项
}

// CacheClear 清理缓存区域
func (cf *CachedFunction) CacheClear() error {
	return cf.CacheBackend.Clear(cf.CacheRegion)
}

// generateCacheKey 生成缓存键
// 原Python: __get_cache_key函数
func generateCacheKey(fn any, args []reflect.Value) string {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	// 函数名称
	fnName := fnType.String()

	// 提取参数值，按照函数签名顺序
	var argValues []interface{}
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		paramName := paramType.Name()

		// 跳过第一个参数，如果是self或cls
		if i == 0 && (paramName == "self" || paramName == "cls") {
			continue
		}

		// 确保参数索引在args范围内
		if i < len(args) {
			argValues = append(argValues, args[i].Interface())
		}
	}

	// 计算参数值的哈希值
	hash := md5.Sum([]byte(fmt.Sprintf("%v", argValues)))
	hashStr := hex.EncodeToString(hash[:])

	// 生成缓存键：函数名_哈希值
	return fmt.Sprintf("%s_%s", fnName, hashStr)
}

// shouldCache 判断是否应该缓存结果
// 原Python: should_cache函数
func shouldCache(value interface{}, skipNone bool, skipEmpty bool) bool {
	// 检查是否为None
	if skipNone && value == nil {
		return false
	}

	// 检查是否为空值
	if skipEmpty {
		// 对于nil值，直接返回false
		if value == nil {
			return false
		}

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String:
			return v.String() != ""
		case reflect.Slice, reflect.Array, reflect.Map:
			return v.Len() > 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return v.Int() != 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return v.Uint() != 0
		case reflect.Float32, reflect.Float64:
			return v.Float() != 0
		case reflect.Bool:
			return v.Bool()
		case reflect.Struct:
			// 结构体总是返回true，除非是特殊的空结构体
			return true
		case reflect.Interface, reflect.Ptr:
			// 对于接口和指针，检查它们指向的值
			if v.IsNil() {
				return false
			}
			return shouldCache(v.Elem().Interface(), false, skipEmpty)
		default:
			return true
		}
	}

	return true
}

// isValidCacheValue 判断指定的值是否为一个有效的缓存值
// 原Python: is_valid_cache_value函数
func isValidCacheValue(cacheBackend CacheBackend, cacheKey string, cachedValue interface{}, cacheRegion string) bool {
	// 如果cachedValue为nil，检查缓存是否实际存在
	if cachedValue == nil {
		exists, _ := cacheBackend.Exists(cacheKey, cacheRegion)
		return exists
	}
	return true
}

// getCacheRegion 获取缓存区域
// 原Python: 自动生成缓存区域的逻辑
func getCacheRegion(fn any, region string) string {
	if region != "" {
		return region
	}

	// 自动生成缓存区域：基于函数类型和名称
	// 在Go中无法直接获取函数的模块名称，使用函数类型字符串作为替代
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()
	return fmt.Sprintf("%s", fnType.String())
}

// Cached 缓存装饰器
// 原Python: @cached装饰器
// 注意：Go不支持Python风格的装饰器，此函数返回一个包装后的函数或CachedFunction结构体
func Cached(fn any, options CachedOptions) any {
	// 使用默认选项填充缺失值
	if options.Region == "" {
		options.Region = defaultCachedOptions.Region
	}
	if options.MaxSize == 0 {
		options.MaxSize = defaultCachedOptions.MaxSize
	}
	if options.TTL == 0 {
		options.TTL = defaultCachedOptions.TTL
	}

	// 获取缓存区域
	cacheRegion := getCacheRegion(fn, options.Region)

	// 根据TTL值选择缓存类型
	cacheType := "lru"
	if options.TTL > 0 {
		cacheType = "ttl"
	}

	// 创建缓存后端
	cacheBackend := Cache(cacheType, options.MaxSize, int64(options.TTL.Seconds()))

	// 获取函数反射值
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	// 检查是否为函数
	if fnType.Kind() != reflect.Func {
		panic("Cached decorator can only be applied to functions")
	}

	// 创建包装函数
	wrapper := func(args []reflect.Value) []reflect.Value {
		// 生成缓存键
		key := generateCacheKey(fn, args)

		// 尝试从缓存获取
		cachedValue, hit, err := cacheBackend.Get(key, cacheRegion)
		if err == nil && hit {
			// 验证缓存值是否有效
			if shouldCache(cachedValue, options.SkipNone, options.SkipEmpty) && isValidCacheValue(cacheBackend, key, cachedValue, cacheRegion) {
				// 缓存命中，转换为函数返回值类型
				result := reflect.ValueOf(cachedValue)
				if result.IsValid() {
					// 检查是否有错误返回值
					if fnType.NumOut() == 2 && fnType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
						return []reflect.Value{result, reflect.Zero(fnType.Out(1))}
					}
					return []reflect.Value{result}
				}
			}
		}

		// 缓存未命中，执行原始函数
		results := fnVal.Call(args)

		// 检查是否需要缓存结果
		if len(results) > 0 {
			result := results[0].Interface()

			// 判断是否应该缓存
			if shouldCache(result, options.SkipNone, options.SkipEmpty) {
				// 将结果存入缓存
				_ = cacheBackend.Set(key, result, options.TTL, cacheRegion)
			}
		}

		return results
	}

	// 收集函数参数类型
	inTypes := make([]reflect.Type, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		inTypes[i] = fnType.In(i)
	}

	// 收集函数返回值类型
	outTypes := make([]reflect.Type, fnType.NumOut())
	for i := 0; i < fnType.NumOut(); i++ {
		outTypes[i] = fnType.Out(i)
	}

	// 创建新的函数类型
	newFnType := reflect.FuncOf(
		inTypes,
		outTypes,
		fnType.IsVariadic(),
	)

	// 创建函数值
	newFn := reflect.MakeFunc(newFnType, wrapper)

	// 返回封装了缓存函数和相关方法的结构体
	return &CachedFunction{
		OriginalFunc: newFn.Interface(),
		CacheRegion:  cacheRegion,
		CacheBackend: cacheBackend,
		CacheOptions: options,
	}
}

// AsyncCached 异步函数缓存装饰器
// 原Python: @cached装饰器的异步版本
func AsyncCached(fn any, options CachedOptions) any {
	// 使用默认选项填充缺失值
	if options.Region == "" {
		options.Region = defaultCachedOptions.Region
	}
	if options.MaxSize == 0 {
		options.MaxSize = defaultCachedOptions.MaxSize
	}
	if options.TTL == 0 {
		options.TTL = defaultCachedOptions.TTL
	}

	// 获取缓存区域
	cacheRegion := getCacheRegion(fn, options.Region)

	// 根据TTL值选择缓存类型
	cacheType := "lru"
	if options.TTL > 0 {
		cacheType = "ttl"
	}

	// 创建异步缓存后端
	cacheBackend := AsyncCache(cacheType, options.MaxSize, int64(options.TTL.Seconds()))

	// 获取函数反射值
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	// 检查是否为函数
	if fnType.Kind() != reflect.Func {
		panic("AsyncCached decorator can only be applied to functions")
	}

	// 创建包装函数
	wrapper := func(args []reflect.Value) []reflect.Value {
		// 生成缓存键
		key := generateCacheKey(fn, args)

		// 尝试从缓存获取
		cachedValue, hit, err := cacheBackend.Get(key, cacheRegion)
		if err == nil && hit {
			// 验证缓存值是否有效
			if shouldCache(cachedValue, options.SkipNone, options.SkipEmpty) && isValidCacheValue(cacheBackend.(CacheBackend), key, cachedValue, cacheRegion) {
				// 缓存命中，转换为函数返回值类型
				result := reflect.ValueOf(cachedValue)
				if result.IsValid() {
					// 检查是否有错误返回值
					if fnType.NumOut() == 2 && fnType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
						return []reflect.Value{result, reflect.Zero(fnType.Out(1))}
					}
					return []reflect.Value{result}
				}
			}
		}

		// 缓存未命中，执行原始函数
		results := fnVal.Call(args)

		// 检查是否需要缓存结果
		if len(results) > 0 {
			result := results[0].Interface()

			// 判断是否应该缓存
			if shouldCache(result, options.SkipNone, options.SkipEmpty) {
				// 将结果存入缓存
				_ = cacheBackend.Set(key, result, options.TTL, cacheRegion)
			}
		}

		return results
	}

	// 收集函数参数类型
	inTypes := make([]reflect.Type, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		inTypes[i] = fnType.In(i)
	}

	// 收集函数返回值类型
	outTypes := make([]reflect.Type, fnType.NumOut())
	for i := 0; i < fnType.NumOut(); i++ {
		outTypes[i] = fnType.Out(i)
	}

	// 创建新的函数类型
	newFnType := reflect.FuncOf(
		inTypes,
		outTypes,
		fnType.IsVariadic(),
	)

	// 创建函数值
	newFn := reflect.MakeFunc(newFnType, wrapper)

	// 返回封装了缓存函数和相关方法的结构体
	return &CachedFunction{
		OriginalFunc: newFn.Interface(),
		CacheRegion:  cacheRegion,
		CacheBackend: cacheBackend.(CacheBackend),
		CacheOptions: options,
	}
}
