// Package main 提供HTTP工具使用示例
package main

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== HTTP工具使用示例 ===")

	// 示例1: Cookie解析
	fmt.Println("\n1. Cookie解析:")
	cookiesStr := "name=value; key1=value1; key2=value2"
	cookieDict := utils.CookieParse(cookiesStr, false)
	fmt.Printf("解析为字�? %v\n", cookieDict)

	cookieArray := utils.CookieParse(cookiesStr, true)
	fmt.Printf("解析为数�? %v\n", cookieArray)

	// 示例2: 同步HTTP请求工具
	fmt.Println("\n2. 同步HTTP请求工具:")
	
	// 创建RequestUtils实例
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}
	
	proxies := map[string]string{
		"http":  "http://proxy.example.com:8080",
		"https": "https://proxy.example.com:8080",
	}
	
	requestUtils := utils.NewRequestUtils(
		headers,
		"",              // ua
		"name=value",    // cookies
		proxies,         // proxies
		nil,             // session
		30,              // timeout
		"https://example.com", // referer
		"",              // content_type
		"application/json", // accept_type
	)
	
	// GET请求示例 (由于网络限制，这里只展示用法)
	fmt.Println("创建HTTP请求工具实例成功")
	fmt.Printf("超时时间: %v\n", requestUtils.Timeout)
	fmt.Printf("代理设置: %s\n", requestUtils.Proxies)
	fmt.Printf("请求�? %v\n", requestUtils.Headers)
	fmt.Printf("Cookies: %v\n", requestUtils.Cookies)

	// 示例3: 异步HTTP请求工具
	fmt.Println("\n3. 异步HTTP请求工具:")
	
	asyncRequestUtils := utils.NewAsyncRequestUtils(
		nil,             // headers
		"Custom UA",     // ua
		nil,             // cookies
		nil,             // proxies
		nil,             // client
		15,              // timeout
		"",              // referer
		"",              // content_type
		"",              // accept_type
	)
	
	fmt.Printf("异步工具超时时间: %v\n", asyncRequestUtils.Timeout)

	// 示例4: 发送GET请求 (演示用法，不实际执行网络请求)
	fmt.Println("\n4. GET请求示例:")
	params := map[string]string{
		"key": "value",
		"page": "1",
	}
	
	fmt.Printf("GET参数: %v\n", params)

	// 示例5: 发送POST请求 (演示用法)
	fmt.Println("\n5. POST请求示例:")
	postData := map[string]interface{}{
		"name": "test",
		"value": 123,
	}
	
	fmt.Printf("POST数据: %v\n", postData)

	// 示例6: JSON请求
	fmt.Println("\n6. JSON请求示例:")
	jsonData := map[string]interface{}{
		"action": "login",
		"username": "user",
		"password": "pass",
	}
	
	fmt.Printf("JSON数据: %v\n", jsonData)

	// 示例7: 缓存头生�?	fmt.Println("\n7. 缓存头生�?")
	etag := "abc123"
	cacheControl, maxAge := "public", 3600
	
	cacheHeaders := requestUtils.GenerateCacheHeaders(&etag, cacheControl, &maxAge)
	fmt.Printf("缓存�? %v\n", cacheHeaders)

	// 示例8: 解析Cache-Control�?	fmt.Println("\n8. 解析Cache-Control�?")
	cacheDirective, maxAgeValue := requestUtils.ParseCacheControl("public, max-age=3600")
	fmt.Printf("缓存指令: %s, 最大年�? %v\n", cacheDirective, maxAgeValue)

	// 示例9: 实际HTTP请求 (需要网络连�?
	fmt.Println("\n9. 实际HTTP请求示例:")
	
	// 创建一个简单的请求工具
	simpleRequest := utils.NewRequestUtils(
		nil,    // headers
		"",     // ua
		nil,    // cookies
		nil,    // proxies
		nil,    // session
		10,     // timeout
		"",     // referer
		"",     // content_type
		"",     // accept_type
	)
	
	// 使用context控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 检查是否能建立连接 (使用一个可靠的URL)
	url := "https://httpbin.org/get"
	fmt.Printf("尝试连接: %s\n", url)
	
	// 发送GET请求
	resp, err := simpleRequest.GetRes(url, nil, nil, nil, true, false)
	if err != nil {
		fmt.Printf("请求出错: %v\n", err)
	} else if resp != nil {
		fmt.Printf("请求成功: %s\n", resp.Status)
		resp.Body.Close()
	} else {
		fmt.Println("请求未返回结�?)
	}

	// 异步请求示例
	fmt.Println("\n10. 异步请求示例:")
	asyncUtils := utils.NewAsyncRequestUtils(
		nil,    // headers
		"",     // ua
		nil,    // cookies
		nil,    // proxies
		nil,    // client
		10,     // timeout
		"",     // referer
		"",     // content_type
		"",     // accept_type
	)
	
	asyncResp, err := asyncUtils.Request(ctx, "GET", "https://httpbin.org/get", false, nil, nil)
	if err != nil {
		fmt.Printf("异步请求出错: %v\n", err)
	} else if asyncResp != nil {
		fmt.Printf("异步请求成功: %s\n", asyncResp.Status)
		asyncResp.Body.Close()
	} else {
		fmt.Println("异步请求未返回结�?)
	}

	// PUT和DELETE请求示例
	fmt.Println("\n11. PUT和DELETE请求示例:")
	
	putResp, err := simpleRequest.Put("https://httpbin.org/put", map[string]interface{}{"key": "value"})
	if err != nil {
		fmt.Printf("PUT请求出错: %v\n", err)
	} else if putResp != nil {
		fmt.Printf("PUT请求成功: %s\n", putResp.Status)
		putResp.Body.Close()
	}
	
	deleteResp, err := simpleRequest.DeleteRes("https://httpbin.org/delete", nil, nil, true, false)
	if err != nil {
		fmt.Printf("DELETE请求出错: %v\n", err)
	} else if deleteResp != nil {
		fmt.Printf("DELETE请求成功: %s\n", deleteResp.Status)
		deleteResp.Body.Close()
	}

	// JSON请求示例
	fmt.Println("\n12. JSON请求示例:")
	
	jsonResult, err := simpleRequest.GetJSON("https://httpbin.org/json", nil)
	if err != nil {
		fmt.Printf("JSON GET请求出错: %v\n", err)
	} else if jsonResult != nil {
		fmt.Printf("JSON GET请求成功: %+v\n", jsonResult)
	}
	
	postJSONResult, err := simpleRequest.PostJSON("https://httpbin.org/post", nil, map[string]interface{}{"test": "data"})
	if err != nil {
		fmt.Printf("JSON POST请求出错: %v\n", err)
	} else if postJSONResult != nil {
		fmt.Printf("JSON POST请求成功\n")
	}
}
