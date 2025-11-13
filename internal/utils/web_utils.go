package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// WebUtils 提供Web相关的工具函�?type WebUtils struct {
	httpUtils *HTTPUtils
}

// LocationResponse1 第一种API的响应结�?type LocationResponse1 struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		IP       string `json:"ip"`
		Location string `json:"location"`
	} `json:"data"`
}

// LocationResponse2 第二种API的响应结�?type LocationResponse2 struct {
	IP   string `json:"ip"`
	Addr string `json:"addr"`
	Pro  string `json:"pro"`
	City string `json:"city"`
}

// NewWebUtils 创建一个新的WebUtils实例
func NewWebUtils() *WebUtils {
	return &WebUtils{
		httpUtils: NewHTTPUtils(),
	}
}

// GetLocation 查询IP所属地
// 先尝试第一种方式，失败则尝试第二种方式
func (w *WebUtils) GetLocation(ip string) string {
	location := w.GetLocation1(ip)
	if location != "" {
		return location
	}
	
	return w.GetLocation2(ip)
}

// GetLocation1 通过api.mir6.com查询IP所属地
func (w *WebUtils) GetLocation1(ip string) string {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("GetLocation1 error: %v\n", err)
		}
	}()
	
	url := fmt.Sprintf("https://api.mir6.com/api/ip?ip=%s&type=json", ip)
	resp, err := w.httpUtils.Get(url, nil, nil, nil, 0)
	if err != nil || resp == nil {
		return ""
	}
	
	// 解析JSON响应
	var result LocationResponse1
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return ""
	}
	
	// 检查响应码
	if result.Code != 200 {
		return ""
	}
	
	// 返回位置信息
	location := result.Data.Location
	if strings.TrimSpace(location) == "" {
		return ""
	}
	
	return location
}

// GetLocation2 通过whois.pconline.com.cn查询IP所属地
func (w *WebUtils) GetLocation2(ip string) string {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("GetLocation2 error: %v\n", err)
		}
	}()
	
	url := fmt.Sprintf("https://whois.pconline.com.cn/ipJson.jsp?json=true&ip=%s", ip)
	resp, err := w.httpUtils.Get(url, nil, nil, nil, 0)
	if err != nil || resp == nil {
		return ""
	}
	
	// 解析JSON响应
	var result LocationResponse2
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return ""
	}
	
	// 返回地址信息
	addr := result.Addr
	if strings.TrimSpace(addr) == "" {
		return ""
	}
	
	return addr
}
