package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GetLocation 查询 IP 所属地，对应 Python WebUtils.get_location
// 优先使用 getLocation1，失败或空结果时回退到 getLocation2。
func GetLocation(ip string) string {
	if ip == "" {
		return ""
	}
	if loc := getLocation1(ip); loc != "" {
		return loc
	}
	return getLocation2(ip)
}

// getLocation1 调用 https://api.mir6.com/api/ip?ip=...&type=json
func getLocation1(ip string) string {
	url := fmt.Sprintf("https://api.mir6.com/api/ip?ip=%s&type=json", ip)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Location string `json:"location"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ""
	}
	return body.Data.Location
}

// getLocation2 调用 https://whois.pconline.com.cn/ipJson.jsp?json=true&ip=...
func getLocation2(ip string) string {
	url := fmt.Sprintf("https://whois.pconline.com.cn/ipJson.jsp?json=true&ip=%s", ip)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var body struct {
		Addr string `json:"addr"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ""
	}
	return body.Addr
}
