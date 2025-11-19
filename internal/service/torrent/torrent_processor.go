// Package torrent 提供种子文件处理服务
package torrent

import (
	"base64"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"

	"go.uber.org/zap"
)

// TorrentProcessor 种子处理器
type TorrentProcessor struct {
	httpClient *utils.HTTPClient
	logger     *zap.Logger
}

// TorrentURLDecoder 复杂URL解码器
type TorrentURLDecoder struct {
	httpClient *utils.HTTPClient
	logger     *zap.Logger
}

// NewTorrentProcessor 创建种子处理器实例
func NewTorrentProcessor(httpClient *utils.HTTPClient) *TorrentProcessor {
	return &TorrentProcessor{
		httpClient: httpClient,
		logger:     logger.Logger,
	}
}

// NewTorrentURLDecoder 创建URL解码器实例
func NewTorrentURLDecoder(httpClient *utils.HTTPClient) *TorrentURLDecoder {
	return &TorrentURLDecoder{
		httpClient: httpClient,
		logger:     logger.Logger,
	}
}

// ProcessTorrentURL 处理种子URL
func (tp *TorrentProcessor) ProcessTorrentURL(ctx context.Context, torrentURL string, options *ProcessOptions) (*TorrentData, error) {
	tp.logger.Info("开始处理种子URL", "url", torrentURL, "title", options.Title)

	// 检查是否为磁力链接
	if strings.HasPrefix(torrentURL, "magnet:") {
		return &TorrentData{
			Type:  "magnet",
			URL:   torrentURL,
			Title: options.Title,
			Hash:  extractMagnetHash(torrentURL),
			Size:  0, // 磁力链接大小未知
			Files: []TorrentFile{},
		}, nil
	}

	// 处理复杂URL格式
	if strings.HasPrefix(torrentURL, "[") {
		decodedURL, err := tp.decodeComplexURL(ctx, torrentURL, options)
		if err != nil {
			return nil, fmt.Errorf("解码复杂URL失败: %w", err)
		}
		torrentURL = decodedURL
	}

	// 下载种子文件
	torrentData, err := tp.downloadTorrentFile(ctx, torrentURL, options)
	if err != nil {
		return nil, fmt.Errorf("下载种子文件失败: %w", err)
	}

	// 解析种子文件内容
	torrentInfo, err := tp.parseTorrentData(torrentData, options)
	if err != nil {
		return nil, fmt.Errorf("解析种子文件失败: %w", err)
	}

	tp.logger.Info("种子URL处理完成", "title", torrentInfo.Title, "files", len(torrentInfo.Files))
	return torrentInfo, nil
}

// decodeComplexURL 解码复杂URL格式 [base64]url
func (tp *TorrentProcessor) decodeComplexURL(ctx context.Context, complexURL string, options *ProcessOptions) (string, error) {
	tp.logger.Info("解码复杂URL", "url", complexURL)

	// 使用正则表达式匹配格式
	m := regexp.MustCompile(`\[(.*)\](.*)`).FindStringSubmatch(complexURL)
	if len(m) != 3 {
		return "", fmt.Errorf("无效的复杂URL格式: %s", complexURL)
	}

	// 解析参数
	base64Str := m[1]
	targetURL := m[2]

	if base64Str == "" {
		return targetURL, nil
	}

	// 解码参数
	decodedBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %w", err)
	}

	var reqParams URLRequestParams
	if err := json.Unmarshal(decodedBytes, &reqParams); err != nil {
		return "", fmt.Errorf("参数JSON解析失败: %w", err)
	}

	// 构建HTTP请求
	req, err := tp.buildHTTPRequest(targetURL, &reqParams, options)
	if err != nil {
		return "", fmt.Errorf("构建HTTP请求失败: %w", err)
	}

	// 执行请求
	resp, err := tp.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 处理响应结果
	result, err := tp.processResponse(resp, &reqParams)
	if err != nil {
		return "", fmt.Errorf("处理响应失败: %w", err)
	}

	tp.logger.Info("复杂URL解码成功", "result", result)
	return result, nil
}

// buildHTTPRequest 构建HTTP请求
func (tp *TorrentProcessor) buildHTTPRequest(targetURL string, params *URLRequestParams, options *ProcessOptions) (*http.Request, error) {
	var req *http.Request
	var err error

	// 设置请求头
	headers := params.Headers
	if headers == nil {
		headers = make(map[string]string)
	}

	// 设置User-Agent
	if headers["User-Agent"] == "" {
		headers["User-Agent"] = options.UserAgent
	}

	// 设置代理
	proxy := options.Proxy
	if params.Proxy != "" {
		proxy = params.Proxy
	}

	// 构建请求体
	var body io.Reader
	if params.Method == "post" && params.Params != nil {
		jsonData, err := json.Marshal(params.Params)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(jsonData))
	}

	// 创建请求
	if params.Method == "post" {
		req, err = http.NewRequest("POST", targetURL, body)
	} else {
		req, err = http.NewRequest("GET", targetURL, nil)
	}
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 设置Cookie
	if params.Cookie != "" {
		req.Header.Set("Cookie", params.Cookie)
	}

	// 设置查询参数
	if params.Params != nil && params.Method == "get" {
		values := url.Values{}
		for key, value := range params.Params {
			if strValue, ok := value.(string); ok {
				values.Set(key, strValue)
			}
		}
		req.URL.RawQuery = values.Encode()
	}

	return req, nil
}

// processResponse 处理响应
func (tp *TorrentProcessor) processResponse(resp *http.Response, params *URLRequestParams) (string, error) {
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 如果指定了结果提取路径
	if params.Result != "" {
		var jsonData interface{}
		if err := json.Unmarshal(respBody, &jsonData); err != nil {
			return "", fmt.Errorf("响应JSON解析失败: %w", err)
		}

		// 根据路径提取结果
		result := jsonData
		for _, key := range strings.Split(params.Result, ".") {
			if m, ok := result.(map[string]interface{}); ok {
				var exists bool
				result, exists = m[key]
				if !exists {
					return "", fmt.Errorf("结果路径 '%s' 不存在", params.Result)
				}
			} else {
				return "", fmt.Errorf("无效的结果路径: %s", params.Result)
			}
		}

		if resultStr, ok := result.(string); ok {
			return resultStr, nil
		}
		return "", fmt.Errorf("结果不是字符串类型")
	}

	return string(respBody), nil
}

// downloadTorrentFile 下载种子文件
func (tp *TorrentProcessor) downloadTorrentFile(ctx context.Context, torrentURL string, options *ProcessOptions) ([]byte, error) {
	tp.logger.Info("下载种子文件", "url", torrentURL)

	req, err := http.NewRequestWithContext(ctx, "GET", torrentURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", options.UserAgent)
	if options.Cookie != "" {
		req.Header.Set("Cookie", options.Cookie)
	}

	// 设置代理（如果需要）
	if options.Proxy != "" {
		// 这里可以设置代理
	}

	resp, err := tp.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tp.logger.Info("种子文件下载完成", "size", len(data))
	return data, nil
}

// parseTorrentData 解析种子文件数据
func (tp *TorrentProcessor) parseTorrentData(data []byte, options *ProcessOptions) (*TorrentData, error) {
	// 这里应该调用种子解析库来解析种子文件
	// 暂时返回基础信息

	torrentData := &TorrentData{
		Type:  "file",
		Data:  data,
		URL:   "",
		Title: options.Title,
		Size:  int64(len(data)),
		Files: []TorrentFile{},
	}

	// 尝试从种子数据中提取哈希值
	hash := tp.extractTorrentHash(data)
	if hash != "" {
		torrentData.Hash = hash
	}

	return torrentData, nil
}

// extractTorrentHash 提取种子哈希值
func (tp *TorrentProcessor) extractTorrentHash(data []byte) string {
	// 这里应该实现种子哈希提取逻辑
	// 暂时返回模拟值
	return fmt.Sprintf("hash_%x", time.Now().UnixNano()&0xffffff)
}

// extractMagnetHash 从磁力链接中提取哈希值
func extractMagnetHash(magnetURL string) string {
	// 匹配磁力链接中的哈希值
	re := regexp.MustCompile(`magnet:\?xt=urn:btih:([a-fA-F0-9]{40})`)
	matches := re.FindStringSubmatch(magnetURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// 数据结构定义

// TorrentData 种子数据
type TorrentData struct {
	Type     string        `json:"type"`     // "magnet" 或 "file"
	URL      string        `json:"url"`      // URL或磁力链接
	Data     []byte        `json:"data"`     // 种子文件数据
	Title    string        `json:"title"`    // 种子标题
	Hash     string        `json:"hash"`     // 种子哈希
	Size     int64         `json:"size"`     // 种子大小
	Files    []TorrentFile `json:"files"`    // 文件列表
	Trackers []string      `json:"trackers"` // Tracker列表
}

// TorrentFile 种子文件
type TorrentFile struct {
	Index int    `json:"index"` // 文件索引
	Name  string `json:"name"`  // 文件名
	Size  int64  `json:"size"`  // 文件大小
	Path  string `json:"path"`  // 文件路径
	Hash  string `json:"hash"`  // 文件哈希
}

// ProcessOptions 处理选项
type ProcessOptions struct {
	Title     string `json:"title"`
	UserAgent string `json:"user_agent"`
	Cookie    string `json:"cookie"`
	Proxy     string `json:"proxy"`
	SiteID    string `json:"site_id"`
	SiteName  string `json:"site_name"`
}

// URLRequestParams URL请求参数
type URLRequestParams struct {
	Cookie  string                 `json:"cookie,omitempty"`
	Proxy   string                 `json:"proxy,omitempty"`
	Method  string                 `json:"method,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Result  string                 `json:"result,omitempty"`
}

// ValidateURL 验证种子URL
func (tp *TorrentProcessor) ValidateURL(torrentURL string) error {
	if torrentURL == "" {
		return fmt.Errorf("种子URL不能为空")
	}

	// 检查是否为磁力链接
	if strings.HasPrefix(torrentURL, "magnet:") {
		if !regexp.MustCompile(`^magnet:\?xt=urn:btih:[a-fA-F0-9]{40}`).MatchString(torrentURL) {
			return fmt.Errorf("无效的磁力链接格式")
		}
		return nil
	}

	// 检查是否为复杂URL格式
	if strings.HasPrefix(torrentURL, "[") {
		if !regexp.MustCompile(`\[.*\].*`).MatchString(torrentURL) {
			return fmt.Errorf("无效的复杂URL格式")
		}
		return nil
	}

	// 检查是否为有效的HTTP/HTTPS URL
	if !strings.HasPrefix(torrentURL, "http://") && !strings.HasPrefix(torrentURL, "https://") {
		return fmt.Errorf("无效的URL格式，必须是http://或https://开头")
	}

	// 验证URL格式
	_, err := url.Parse(torrentURL)
	if err != nil {
		return fmt.Errorf("无效的URL格式: %w", err)
	}

	return nil
}

// GetTorrentType 获取种子类型
func (tp *TorrentProcessor) GetTorrentType(torrentURL string) string {
	if strings.HasPrefix(torrentURL, "magnet:") {
		return "magnet"
	}
	if strings.HasPrefix(torrentURL, "[") {
		return "complex"
	}
	return "url"
}

// EstimateTorrentSize 估算种子大小
func (tp *TorrentProcessor) EstimateTorrentSize(ctx context.Context, torrentURL string, options *ProcessOptions) (int64, error) {
	// 磁力链接无法估算大小
	if strings.HasPrefix(torrentURL, "magnet:") {
		return 0, nil
	}

	// 尝试通过HEAD请求获取大小
	req, err := http.NewRequestWithContext(ctx, "HEAD", torrentURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", options.UserAgent)
	if options.Cookie != "" {
		req.Header.Set("Cookie", options.Cookie)
	}

	resp, err := tp.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 从Content-Length获取大小
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := utils.ParseInt64(contentLength); err == nil {
			return size, nil
		}
	}

	return 0, fmt.Errorf("无法获取文件大小")
}
