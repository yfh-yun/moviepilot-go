package storages

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/core/cache"
	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/helper/progress"
	"moviepilot-go/internal/helper/storage"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/crypto"
	"moviepilot-go/internal/utils/httpclient"
	"moviepilot-go/internal/utils/urlutils"
	"moviepilot-go/internal/utils/singleton"
)

// Alist AList相关操作
type Alist struct {
	BaseStorage
}

// NewAlist 创建AList实例
func NewAlist() *Alist {
	return &Alist{
		BaseStorage: *NewBaseStorage(),
	}
}

// Schema 获取存储模式
func (a *Alist) Schema() *StorageSchema {
	return &StorageSchema{Value: string(types.StorageSchemaAlist)}
}

// InitStorage 初始�?func (a *Alist) InitStorage() {
	// 清除缓存
	a.generateToken.CacheClear()
}

// delayGetItem 自动延迟重试 get_item 模块
func (a *Alist) delayGetItem(path string) *schemas.FileItem {
	for i := 0; i < 2; i++ {
		time.Sleep(2 * time.Second)
		fileitem := a.GetItem(path)
		if fileitem != nil {
			return fileitem
		}
	}
	return nil
}

// getBaseURL 获取基础URL
func (a *Alist) getBaseURL() string {
	conf := a.GetConf()
	urlVal, exists := conf["url"].(string)
	if !exists || urlVal == "" {
		return ""
	}
	return urlutils.StandardizeBaseURL(urlVal)
}

// getAPIUrl 获取API URL
func (a *Alist) getAPIUrl(path string) string {
	return urlutils.AdaptRequestURL(a.getBaseURL(), path)
}

// getValuableToken 获取一个可用的token
func (a *Alist) getValuableToken() string {
	return a.generateToken()
}

// generateTokenCache 用于缓存token的结�?type generateTokenCache struct {
	token string
	expireTime time.Time
}

var tokenCache *generateTokenCache
var tokenCacheMutex sync.Mutex

// generateToken 如果设置永久令牌则返回永久令牌，否则使用账号密码生成一个临�?token
// 缓存2天，提前5分钟更新
func (a *Alist) generateToken() string {
	tokenCacheMutex.Lock()
	defer tokenCacheMutex.Unlock()
	
	// 检查缓�?	if tokenCache != nil && time.Now().Before(tokenCache.expireTime) {
		return tokenCache.token
	}
	
	conf := a.GetConf()
	token, tokenExists := conf["token"].(string)
	if tokenExists && token != "" {
		// 缓存永久令牌2�?		tokenCache = &generateTokenCache{
			token: token,
			expireTime: time.Now().Add(48 * time.Hour - 5 * time.Minute),
		}
		return token
	}
	
	// 使用账号密码生成临时令牌
	username, usernameExists := conf["username"].(string)
	password, passwordExists := conf["password"].(string)
	
	if !usernameExists || !passwordExists {
		logger.Warning("【OpenList】用户名或密码未设置")
		return ""
	}
	
	payload := map[string]interface{}{
		"username": username,
		"password": password,
	}
	
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/auth/login"), payload)
	if err != nil {
		logger.Warning("【OpenList】请求登录失败，无法连接alist服务")
		return ""
	}
	
	if resp.StatusCode != 200 {
		logger.Warningf("【OpenList】更新令牌请求发送失败，状态码�?d", resp.StatusCode)
		return ""
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warning("【OpenList】读取响应失�?)
		return ""
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warning("【OpenList】解析响应失�?)
		return ""
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Criticalf("【OpenList】更新令牌，错误信息�?s", message)
		return ""
	}
	
	data, dataExists := result["data"].(map[string]interface{})
	if !dataExists {
		logger.Warning("【OpenList】响应中缺少data字段")
		return ""
	}
	
	newToken, tokenExists := data["token"].(string)
	if !tokenExists {
		logger.Warning("【OpenList】响应中缺少token字段")
		return ""
	}
	
	logger.Debug("【OpenList】AList获取令牌成功")
	
	// 缓存临时令牌2天，提前5分钟更新
	tokenCache = &generateTokenCache{
		token: newToken,
		expireTime: time.Now().Add(48 * time.Hour - 5 * time.Minute),
	}
	
	return newToken
}

// getHeaderWithToken 获取带有token的header
func (a *Alist) getHeaderWithToken() map[string]string {
	return map[string]string{"Authorization": a.getValuableToken()}
}

// Check 检查存储是否可�?func (a *Alist) Check() bool {
	return a.generateToken() != ""
}

// List 浏览文件
func (a *Alist) List(fileitem *schemas.FileItem, password string, page int, perPage int, refresh bool) []*schemas.FileItem {
	if fileitem.Type == "file" {
		item := a.GetItem(fileitem.Path)
		if item != nil {
			return []*schemas.FileItem{item}
		}
		return []*schemas.FileItem{}
	}
	
	payload := map[string]interface{}{
		"path":      fileitem.Path,
		"password":  password,
		"page":      page,
		"per_page":  perPage,
		"refresh":   refresh,
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/list"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求获取目�?%s 的文件列表失败，无法连接alist服务", fileitem.Path)
		return []*schemas.FileItem{}
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求获取目�?%s 的文件列表失败，状态码�?d", fileitem.Path, resp.StatusCode)
		return []*schemas.FileItem{}
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return []*schemas.FileItem{}
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return []*schemas.FileItem{}
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Warnf("【OpenList】获取目�?%s 的文件列表失败，错误信息�?s", fileitem.Path, message)
		return []*schemas.FileItem{}
	}
	
	data, dataExists := result["data"].(map[string]interface{})
	if !dataExists {
		logger.Warn("【OpenList】响应中缺少data字段")
		return []*schemas.FileItem{}
	}
	
	content, contentExists := data["content"].([]interface{})
	if !contentExists {
		return []*schemas.FileItem{}
	}
	
	var items []*schemas.FileItem
	for _, item := range content {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		
		name, _ := itemMap["name"].(string)
		size, _ := itemMap["size"].(float64)
		isDir, _ := itemMap["is_dir"].(bool)
		modified, _ := itemMap["modified"].(string)
		thumb, _ := itemMap["thumb"].(string)
		
		var itemType string
		var extension *string
		var basename string
		
		if isDir {
			itemType = "dir"
			extension = nil
			basename = name
		} else {
			itemType = "file"
			ext := filepath.Ext(name)
			if ext != "" {
				ext = ext[1:] // 移除点号
			}
			extension = &ext
			basename = strings.TrimSuffix(name, filepath.Ext(name))
		}
		
		path := filepath.Join(fileitem.Path, name)
		if isDir {
			path += "/"
		}
		
		items = append(items, &schemas.FileItem{
			Storage:    string(types.StorageSchemaAlist),
			Type:       itemType,
			Path:       path,
			Name:       name,
			Basename:   basename,
			Extension:  extension,
			Size:       int64(size),
			ModifyTime: a.parseTimestamp(modified),
			Thumbnail:  thumb,
		})
	}
	
	return items
}

// CreateFolder 创建目录
func (a *Alist) CreateFolder(fileitem *schemas.FileItem, name string) *schemas.FileItem {
	path := filepath.Join(fileitem.Path, name)
	
	payload := map[string]interface{}{
		"path": path,
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/mkdir"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求创建目�?%s 失败，无法连接alist服务", path)
		return nil
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求创建目�?%s 失败，状态码�?d", path, resp.StatusCode)
		return nil
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return nil
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return nil
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Warnf("【OpenList】创建目�?%s 失败，错误信息：%s", path, message)
		return nil
	}
	
	return a.delayGetItem(path)
}

// GetFolder 获取目录，如目录不存在则创建
func (a *Alist) GetFolder(path string) *schemas.FileItem {
	folder := a.GetItem(path)
	if folder != nil {
		return folder
	}
	
	// 创建目录
	parentPath := filepath.Dir(path)
	dirName := filepath.Base(path)
	
	parentItem := &schemas.FileItem{
		Storage:  string(types.StorageSchemaAlist),
		Type:     "dir",
		Path:     parentPath,
		Name:     dirName,
		Basename: strings.TrimSuffix(dirName, filepath.Ext(dirName)),
	}
	
	if !strings.HasSuffix(parentPath, "/") {
		parentPath += "/"
	}
	
	parentItem.Path = parentPath
	
	folder = a.CreateFolder(parentItem, dirName)
	return folder
}

// GetItem 获取文件或目录，不存在返回nil
func (a *Alist) GetItem(path string, password string, page int, perPage int, refresh bool) *schemas.FileItem {
	payload := map[string]interface{}{
		"path":      path,
		"password":  password,
		"page":      page,
		"per_page":  perPage,
		"refresh":   refresh,
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/get"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求获取文�?%s 失败，无法连接alist服务", path)
		return nil
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求获取文�?%s 失败，状态码�?d", path, resp.StatusCode)
		return nil
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return nil
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return nil
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Debugf("【OpenList】获取文�?%s 失败，错误信息：%s", path, message)
		return nil
	}
	
	data, dataExists := result["data"].(map[string]interface{})
	if !dataExists {
		logger.Warn("【OpenList】响应中缺少data字段")
		return nil
	}
	
	name, _ := data["name"].(string)
	size, _ := data["size"].(float64)
	isDir, _ := data["is_dir"].(bool)
	modified, _ := data["modified"].(string)
	thumb, _ := data["thumb"].(string)
	
	var itemType string
	var extension string
	var basename string
	
	if isDir {
		itemType = "dir"
		extension = ""
		basename = name
	} else {
		itemType = "file"
		extension = filepath.Ext(name)
		if extension != "" {
			extension = extension[1:] // 移除点号
		}
		basename = strings.TrimSuffix(name, filepath.Ext(name))
	}
	
	itemPath := path
	if isDir && !strings.HasSuffix(itemPath, "/") {
		itemPath += "/"
	}
	
	return &schemas.FileItem{
		Storage:    string(types.StorageSchemaAlist),
		Type:       itemType,
		Path:       itemPath,
		Name:       name,
		Basename:   basename,
		Extension:  &extension,
		Size:       int64(size),
		ModifyTime: a.parseTimestamp(modified),
		Thumbnail:  thumb,
	}
}

// GetParent 获取父目�?func (a *Alist) GetParent(fileitem *schemas.FileItem) *schemas.FileItem {
	parentPath := filepath.Dir(strings.TrimSuffix(fileitem.Path, "/"))
	return a.GetFolder(parentPath)
}

// isEmptyDir 判断目录是否为空
func (a *Alist) isEmptyDir(fileitem *schemas.FileItem) bool {
	if fileitem.Type != "dir" {
		return false
	}
	// 获取目录内容
	items := a.List(fileitem, "", 1, 0, false)
	return len(items) == 0
}

// Delete 删除文件或目录，空目录用专用API
func (a *Alist) Delete(fileitem *schemas.FileItem) bool {
	// 如果是空目录，优先用 remove_empty_directory
	if fileitem.Type == "dir" && a.isEmptyDir(fileitem) {
		payload := map[string]interface{}{
			"src_dir": fileitem.Path,
		}
		
		headers := a.getHeaderWithToken()
		httpClient := httpclient.NewRequestUtils(headers)
		resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/remove_empty_directory"), payload)
		if err != nil {
			logger.Warnf("【OpenList】请求删除空目录 %s 失败，无法连接alist服务", fileitem.Path)
			return false
		}
		
		if resp.StatusCode != 200 {
			logger.Warnf("【OpenList】请求删除空目录 %s 失败，状态码�?d", fileitem.Path, resp.StatusCode)
			return false
		}
		
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
			return false
		}
		
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
			return false
		}
		
		code, codeExists := result["code"].(float64)
		if !codeExists || code != 200 {
			message, _ := result["message"].(string)
			logger.Warnf("【OpenList】删除空目录 %s 失败，错误信息：%s", fileitem.Path, message)
			return false
		}
		
		return true
	}
	
	// 其它情况（文件或非空目录�?	payload := map[string]interface{}{
		"dir":   filepath.Dir(fileitem.Path),
		"names": []string{fileitem.Name},
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/remove"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求删除文�?%s 失败，无法连接alist服务", fileitem.Path)
		return false
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求删除文�?%s 失败，状态码�?d", fileitem.Path, resp.StatusCode)
		return false
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return false
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return false
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Warnf("【OpenList】删除文�?%s 失败，错误信息：%s", fileitem.Path, message)
		return false
	}
	
	return true
}

// Rename 重命名文�?func (a *Alist) Rename(fileitem *schemas.FileItem, name string) bool {
	payload := map[string]interface{}{
		"name": name,
		"path": fileitem.Path,
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/rename"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求重命名文件 %s 失败，无法连接alist服务", fileitem.Path)
		return false
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求重命名文件 %s 失败，状态码�?d", fileitem.Path, resp.StatusCode)
		return false
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return false
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return false
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Warnf("【OpenList】重命名文件 %s 失败，错误信息：%s", fileitem.Path, message)
		return false
	}
	
	return true
}

// Download 下载文件，保存到本地，返回本地临时文件地址
func (a *Alist) Download(fileitem *schemas.FileItem, path string, password string) string {
	payload := map[string]interface{}{
		"path":      fileitem.Path,
		"password":  password,
		"page":      1,
		"per_page":  0,
		"refresh":   false,
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/get"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求获取文�?%s 失败，无法连接alist服务", path)
		return ""
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求获取文�?%s 失败，状态码�?d", path, resp.StatusCode)
		return ""
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return ""
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return ""
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Warnf("【OpenList】获取文�?%s 失败，错误信息：%s", path, message)
		return ""
	}
	
	data, dataExists := result["data"].(map[string]interface{})
	if !dataExists {
		logger.Warn("【OpenList】响应中缺少data字段")
		return ""
	}
	
	var downloadURL string
	if rawURL, exists := data["raw_url"].(string); exists && rawURL != "" {
		downloadURL = rawURL
	} else {
		downloadURL = urlutils.AdaptRequestURL(a.getBaseURL(), "/d"+fileitem.Path)
		if sign, exists := data["sign"].(string); exists && sign != "" {
			downloadURL = downloadURL + "?sign=" + sign
		}
	}
	
	var localPath string
	if path == "" {
		localPath = filepath.Join(config.Settings.TEMP_PATH, fileitem.Name)
	} else {
		localPath = filepath.Join(path, fileitem.Name)
	}
	
	// 初始化进度回�?	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(fileitem.Path))
	progressCallback.Start()
	
	// 下载文件
	downloadClient := httpclient.NewRequestUtils(headers)
	downloadResp, err := downloadClient.GetStream(downloadURL, true)
	if err != nil {
		logger.Errorf("【OpenList】下载文件失败：%s", err.Error())
		progressCallback.End()
		return ""
	}
	defer downloadResp.Body.Close()
	
	if downloadResp.StatusCode != 200 {
		logger.Errorf("【OpenList】下载文件失败，状态码�?d", downloadResp.StatusCode)
		progressCallback.End()
		return ""
	}
	
	// 创建本地文件
	outFile, err := os.Create(localPath)
	if err != nil {
		logger.Errorf("【OpenList】创建本地文件失败：%s", err.Error())
		progressCallback.End()
		return ""
	}
	defer outFile.Close()
	
	// 读取并写入文件内�?	buffer := make([]byte, 8192)
	totalRead := int64(0)
	
	for {
		n, err := downloadResp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := outFile.Write(buffer[:n])
			if writeErr != nil {
				logger.Errorf("【OpenList】写入文件失败：%s", writeErr.Error())
				outFile.Close()
				os.Remove(localPath)
				progressCallback.End()
				return ""
			}
			
			totalRead += int64(n)
			if fileitem.Size > 0 {
				percent := float64(totalRead*100) / float64(fileitem.Size)
				progressCallback.Update(percent, fmt.Sprintf("%s 进度�?0.2f%%", fileitem.Path, percent))
			}
		}
		
		if err != nil {
			if err != io.EOF {
				logger.Errorf("【OpenList】下载过程中发生错误�?s", err.Error())
				outFile.Close()
				os.Remove(localPath)
				progressCallback.End()
				return ""
			}
			break
		}
	}
	
	// 完成下载
	progressCallback.Update(100, fmt.Sprintf("%s 进度�?00%%", fileitem.Path))
	progressCallback.End()
	
	return localPath
}

// Upload 上传文件（带进度�?func (a *Alist) Upload(fileitem *schemas.FileItem, path string, newName *string, task bool) *schemas.FileItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【OpenList】上传文件异常：%v", r)
		}
	}()
	
	// 获取文件大小
	targetName := filepath.Base(path)
	if newName != nil {
		targetName = *newName
	}
	
	targetPath := filepath.Join(fileitem.Path, targetName)
	
	// 初始化进度回�?	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(path))
	progressCallback.Start()
	
	// 准备上传请求
	encodedPath := url.QueryEscape(targetPath)
	headers := a.getHeaderWithToken()
	headers["Content-Type"] = "application/octet-stream"
	headers["As-Task"] = strconv.FormatBool(task)
	headers["File-Path"] = encodedPath
	
	// 打开本地文件
	file, err := os.Open(path)
	if err != nil {
		logger.Errorf("【OpenList】打开本地文件失败�?s", err.Error())
		progressCallback.End()
		return nil
	}
	defer file.Close()
	
	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Errorf("【OpenList】获取文件信息失败：%s", err.Error())
		progressCallback.End()
		return nil
	}
	
	fileSize := fileInfo.Size()
	uploadedSize := int64(0)
	
	// 创建带进度报告的读取�?	progressReader := &ProgressReader{
		file:           file,
		progressFunc:   func(readBytes int64) {
			uploadedSize += readBytes
			if fileSize > 0 {
				percent := float64(uploadedSize*100) / float64(fileSize)
				progressCallback.Update(percent, fmt.Sprintf("%s 进度�?0.2f%%", path, percent))
			}
		},
	}
	
	// 执行上传
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PutRes(a.getAPIUrl("/api/fs/put"), progressReader)
	if err != nil {
		logger.Warnf("【OpenList】请求上传文�?%s 失败", path)
		progressCallback.End()
		return nil
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求上传文�?%s 失败，状态码�?d", path, resp.StatusCode)
		progressCallback.End()
		return nil
	}
	
	defer resp.Body.Close()
	
	// 完成上传
	progressCallback.Update(100, fmt.Sprintf("%s 进度�?00%%", path))
	progressCallback.End()
	
	// 获取上传后的文件�?	newItem := a.delayGetItem(targetPath)
	if newItem != nil && newName != nil && *newName != filepath.Base(path) {
		if a.Rename(newItem, *newName) {
			return a.delayGetItem(filepath.Join(filepath.Dir(newItem.Path), *newName))
		}
	}
	
	return newItem
}

// ProgressReader 带进度报告的读取�?type ProgressReader struct {
	file         *os.File
	progressFunc func(readBytes int64)
}

// Read 实现io.Reader接口
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.file.Read(p)
	if n > 0 && pr.progressFunc != nil {
		pr.progressFunc(int64(n))
	}
	return n, err
}

// Detail 获取文件详情
func (a *Alist) Detail(fileitem *schemas.FileItem) *schemas.FileItem {
	return a.GetItem(fileitem.Path, "", 1, 0, false)
}

// Copy 复制文件
func (a *Alist) Copy(fileitem *schemas.FileItem, path string, newName string) bool {
	payload := map[string]interface{}{
		"src_dir": filepath.Dir(fileitem.Path),
		"dst_dir": path,
		"names":   []string{fileitem.Name},
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/copy"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求复制文�?%s 失败，无法连接alist服务", fileitem.Path)
		return false
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求复制文�?%s 失败，状态码�?d", fileitem.Path, resp.StatusCode)
		return false
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return false
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return false
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Warnf("【OpenList】复制文�?%s 失败，错误信息：%s", fileitem.Path, message)
		return false
	}
	
	// 重命�?	if fileitem.Name != newName {
		newItem := a.delayGetItem(filepath.Join(path, fileitem.Name))
		if newItem != nil {
			a.Rename(newItem, newName)
		}
	}
	
	return true
}

// Move 移动文件
func (a *Alist) Move(fileitem *schemas.FileItem, path string, newName string) bool {
	// 先重命名
	if fileitem.Name != newName {
		a.Rename(fileitem, newName)
	}
	
	payload := map[string]interface{}{
		"src_dir": filepath.Dir(fileitem.Path),
		"dst_dir": path,
		"names":   []string{newName},
	}
	
	headers := a.getHeaderWithToken()
	httpClient := httpclient.NewRequestUtils(headers)
	resp, err := httpClient.PostRes(a.getAPIUrl("/api/fs/move"), payload)
	if err != nil {
		logger.Warnf("【OpenList】请求移动文�?%s 失败，无法连接alist服务", fileitem.Path)
		return false
	}
	
	if resp.StatusCode != 200 {
		logger.Warnf("【OpenList】请求移动文�?%s 失败，状态码�?d", fileitem.Path, resp.StatusCode)
		return false
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("【OpenList】读取响应失败：%s", err.Error())
		return false
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Warnf("【OpenList】解析响应失败：%s", err.Error())
		return false
	}
	
	code, codeExists := result["code"].(float64)
	if !codeExists || code != 200 {
		message, _ := result["message"].(string)
		logger.Warnf("【OpenList】移动文�?%s 失败，错误信息：%s", fileitem.Path, message)
		return false
	}
	
	return true
}

// Link 硬链接文�?func (a *Alist) Link(fileitem *schemas.FileItem, targetFile string) bool {
	// 空实�?	return false
}

// Softlink 软链接文�?func (a *Alist) Softlink(fileitem *schemas.FileItem, targetFile string) bool {
	// 空实�?	return false
}

// Usage 存储使用情况
func (a *Alist) Usage() *schemas.StorageUsage {
	// 空实�?	return nil
}

// parseTimestamp 直接使用 ISO 8601 格式解析时间
func (a *Alist) parseTimestamp(timeStr string) *float64 {
	if timeStr == "" {
		return nil
	}
	
	// 尝试解析时间字符�?	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// 尝试其他常见格式
		t, err = time.Parse("2006-01-02T15:04:05.9999999-07:00", timeStr)
		if err != nil {
			logger.Debugf("【OpenList】解析时间字符串失败�?s", err.Error())
			return nil
		}
	}
	
	timestamp := float64(t.Unix())
	return &timestamp
}
