package storages

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/core/context"
	"moviepilot-go/internal/helper/progress"
	"moviepilot-go/internal/helper/storage"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/crypto"
	"moviepilot-go/internal/utils/httpclient"
	"moviepilot-go/internal/utils/stringutils"
)

// AliPan 阿里云盘存储实现
type AliPan struct {
	BaseStorage
	session       *http.Client
	authState     map[string]string
	chunkSize     int64
	baseURL       string
}

// NoCheckInException 未登录异�?type NoCheckInException struct {
	Message string
}

func (e *NoCheckInException) Error() string {
	return e.Message
}

// SessionInvalidException 会话无效异常
type SessionInvalidException struct {
	Message string
}

func (e *SessionInvalidException) Error() string {
	return e.Message
}

// NewAliPan 创建阿里云盘实例
func NewAliPan() *AliPan {
	return &AliPan{
		BaseStorage: *NewBaseStorage(),
		session:     &http.Client{},
		authState:   make(map[string]string),
		chunkSize:   10 * 1024 * 1024, // 10MB
		baseURL:     "https://openapi.alipan.com",
	}
}

// Schema 获取存储模式
func (a *AliPan) Schema() *StorageSchema {
	return &StorageSchema{Value: string(types.StorageSchemaAlipan)}
}

// InitStorage 初始化存�?func (a *AliPan) InitStorage() {
	// 空实�?}

// GenerateQrcode 生成二维�?func (a *AliPan) GenerateQrcode(args ...interface{}) (*map[string]interface{}, *string) {
	// 生成PKCE参数
	codeVerifier := a.generateCodeVerifier()
	
	// 请求设备�?	payload := map[string]interface{}{
		"client_id":             config.Settings.ALIPAN_APP_ID,
		"scopes":                []string{"user:base", "file:all:read", "file:all:write", "file:share:write"},
		"code_challenge":        codeVerifier,
		"code_challenge_method": "plain",
	}
	
	resp, err := a.requestAPI("POST", "/oauth/authorize/qrcode", payload)
	if err != nil {
		errMsg := "网络错误"
		return nil, &errMsg
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		errMsg := "响应格式错误"
		return nil, &errMsg
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		return nil, &message
	}
	
	// 持久化验证参�?	a.authState["sid"], _ = result["sid"].(string)
	a.authState["code_verifier"] = codeVerifier
	
	// 生成二维码内�?	qrCodeUrl, _ := result["qrCodeUrl"].(string)
	qrCodeData := map[string]interface{}{
		"codeUrl": qrCodeUrl,
	}
	
	return &qrCodeData, nil
}

// generateCodeVerifier 生成PKCE code verifier
func (a *AliPan) generateCodeVerifier() string {
	// 生成随机字符�?	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 96)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:128]
}

// CheckLogin 检查登录状�?func (a *AliPan) CheckLogin(args ...interface{}) *map[string]string {
	statusText := map[string]string{
		"WaitLogin":     "等待登录",
		"ScanSuccess":   "扫码成功",
		"LoginSuccess":  "登录成功",
		"QRCodeExpired": "二维码过�?,
	}
	
	if len(a.authState) == 0 {
		result := map[string]string{"": "生成二维码失�?}
		return &result
	}
	
	url := fmt.Sprintf("%s/oauth/qrcode/%s/status", a.baseURL, a.authState["sid"])
	resp, err := a.session.Get(url)
	if err != nil {
		errMsg := err.Error()
		result := map[string]string{"": errMsg}
		return &result
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		errMsg := err.Error()
		res := map[string]string{"": errMsg}
		return &res
	}
	
	// 扫码结果
	status, _ := result["status"].(string)
	if status == "LoginSuccess" {
		authCode, _ := result["authCode"].(string)
		a.authState["authCode"] = authCode
		tokens, err := a.getAccessToken()
		if err != nil {
			errMsg := err.Error()
			res := map[string]string{"": errMsg}
			return &res
		}
		
		conf := make(map[string]interface{})
		conf["refresh_time"] = time.Now().Unix()
		for k, v := range tokens {
			conf[k] = v
		}
		a.SetConfig(conf)
		a.getDriveId()
	}
	
	res := map[string]string{
		"status": status,
		"tip":    statusText[status],
	}
	return &res
}

// getAccessToken 获取访问令牌
func (a *AliPan) getAccessToken() (map[string]interface{}, error) {
	if len(a.authState) == 0 {
		return nil, &SessionInvalidException{Message: "【阿里云盘】请先生成二维码"}
	}
	
	payload := map[string]interface{}{
		"client_id":     config.Settings.ALIPAN_APP_ID,
		"grant_type":    "authorization_code",
		"code":          a.authState["authCode"],
		"code_verifier": a.authState["code_verifier"],
	}
	
	resp, err := a.requestAPI("POST", "/oauth/access_token", payload)
	if err != nil {
		return nil, &SessionInvalidException{Message: "【阿里云盘】获�?access_token 失败"}
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("【阿里云盘】响应格式错�?)
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		return nil, fmt.Errorf("【阿里云盘�?s - %s�?, code, message)
	}
	
	return result, nil
}

// refreshAccessToken 刷新访问令牌
func (a *AliPan) refreshAccessToken(refreshToken string) (map[string]interface{}, error) {
	if refreshToken == "" {
		return nil, &SessionInvalidException{Message: "【阿里云盘】会话失效，请重新扫码登录！"}
	}
	
	payload := map[string]interface{}{
		"client_id":     config.Settings.ALIPAN_APP_ID,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	
	resp, err := a.requestAPI("POST", "/oauth/access_token", payload)
	if err != nil {
		logger.Errorf("【阿里云盘】刷�?access_token 失败：refresh_token=%s", refreshToken)
		return nil, err
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("【阿里云盘】响应格式错�?)
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Warnf("【阿里云盘】刷�?access_token 失败�?s - %s�?, code, message)
	}
	
	return result, nil
}

// getDriveId 获取默认存储桶ID
func (a *AliPan) getDriveId() error {
	resp, err := a.requestAPI("POST", "/adrive/v1.0/user/getDriveInfo", nil)
	if err != nil {
		logger.Error("获取默认存储桶ID失败")
		return err
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return fmt.Errorf("响应格式错误")
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Warnf("获取默认存储ID失败�?s - %s�?, code, message)
		return nil
	}
	
	// 保存用户参数
	conf := a.GetConf()
	for k, v := range result {
		conf[k] = v
	}
	a.SetConfig(conf)
	
	return nil
}

// checkSession 检查会话是否过�?func (a *AliPan) checkSession() error {
	conf := a.GetConf()
	accessToken, exists := conf["access_token"].(string)
	if !exists || accessToken == "" {
		return &NoCheckInException{Message: "【阿里云盘】请先扫码登录！"}
	}
	
	refreshToken, refreshTokenExists := conf["refresh_token"].(string)
	expiresIn, expiresInExists := conf["expires_in"].(float64)
	refreshTime, refreshTimeExists := conf["refresh_time"].(float64)
	
	if refreshTokenExists && expiresInExists && refreshTimeExists &&
		refreshTime+expiresIn < float64(time.Now().Unix()) {
		tokens, err := a.refreshAccessToken(refreshToken)
		if err != nil {
			return err
		}
		
		newConf := make(map[string]interface{})
		newConf["refresh_time"] = time.Now().Unix()
		for k, v := range tokens {
			newConf[k] = v
		}
		a.SetConfig(newConf)
	}
	
	// 更新请求�?	a.session = &http.Client{}
	
	return nil
}

// defaultDriveId 获取默认存储桶ID
func (a *AliPan) defaultDriveId() (string, error) {
	conf := a.GetConf()
	driveId, exists := conf["resource_drive_id"].(string)
	if !exists || driveId == "" {
		driveId, exists = conf["backup_drive_id"].(string)
	}
	if !exists || driveId == "" {
		driveId, exists = conf["default_drive_id"].(string)
	}
	if !exists || driveId == "" {
		return "", &NoCheckInException{Message: "【阿里云盘】请先扫码登录！"}
	}
	return driveId, nil
}

// requestAPI 带错误处理和速率限制的API请求
func (a *AliPan) requestAPI(method string, endpoint string, payload interface{}) (interface{}, error) {
	// 检查会�?	if err := a.checkSession(); err != nil {
		return nil, err
	}
	
	// 构建请求
	var req *http.Request
	var err error
	
	url := a.baseURL + endpoint
	
	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, url, strings.NewReader(string(jsonData)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		return nil, err
	}
	
	// 设置请求�?	req.Header.Set("Content-Type", "application/json")
	conf := a.GetConf()
	if accessToken, exists := conf["access_token"].(string); exists {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	
	// 发送请�?	resp, err := a.session.Do(req)
	if err != nil {
		logger.Errorf("【阿里云盘�?s 请求 %s 网络错误: %s", method, endpoint, err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	
	// 处理速率限制
	if resp.StatusCode == 429 {
		resetTimeHeader := resp.Header.Get("X-RateLimit-Reset")
		resetTime, err := strconv.Atoi(resetTimeHeader)
		if err != nil {
			resetTime = 60
		}
		time.Sleep(time.Duration(resetTime+5) * time.Second)
		return a.requestAPI(method, endpoint, payload)
	}
	
	// 解析响应
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// 检查错�?	if resultMap, ok := result.(map[string]interface{}); ok {
		if code, exists := resultMap["code"]; exists && code != nil {
			message, _ := resultMap["message"].(string)
			logger.Warnf("【阿里云盘�?s %s 返回�?s %s", method, endpoint, code, message)
		}
	}
	
	return result, nil
}

// getFileItem 获取文件信息
func (a *AliPan) getFileItem(fileInfo map[string]interface{}, parent string) *schemas.FileItem {
	if fileInfo == nil {
		return &schemas.FileItem{}
	}
	
	if !strings.HasSuffix(parent, "/") {
		parent += "/"
	}
	
	fileType, _ := fileInfo["type"].(string)
	fileId, _ := fileInfo["file_id"].(string)
	parentFileId, _ := fileInfo["parent_file_id"].(string)
	name, _ := fileInfo["name"].(string)
	size, _ := fileInfo["size"].(float64)
	updatedAt, _ := fileInfo["updated_at"].(string)
	driveId, _ := fileInfo["drive_id"].(string)
	
	if fileType == "folder" {
		return &schemas.FileItem{
			Storage:       string(types.StorageSchemaAlipan),
			FileId:        fileId,
			ParentFileId:  parentFileId,
			Type:          "dir",
			Path:          parent + name + "/",
			Name:          name,
			Basename:      name,
			Size:          int64(size),
			ModifyTime:    stringutils.StrToTimestamp(updatedAt),
			DriveId:       driveId,
		}
	} else {
		fileExtension, _ := fileInfo["file_extension"].(string)
		thumbnail, _ := fileInfo["thumbnail"].(string)
		
		return &schemas.FileItem{
			Storage:       string(types.StorageSchemaAlipan),
			FileId:        fileId,
			ParentFileId:  parentFileId,
			Type:          "file",
			Path:          parent + name,
			Name:          name,
			Basename:      strings.TrimSuffix(name, filepath.Ext(name)),
			Size:          int64(size),
			Extension:     fileExtension,
			ModifyTime:    stringutils.StrToTimestamp(updatedAt),
			Thumbnail:     thumbnail,
			DriveId:       driveId,
		}
	}
}

// calcSha1 计算文件SHA1（符合阿里云盘规范）
func (a *AliPan) calcSha1(filepath string, size *int64) string {
	file, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer file.Close()
	
	sha1Hash := sha1.New()
	
	if size != nil {
		chunk := make([]byte, *size)
		file.Read(chunk)
		sha1Hash.Write(chunk)
	} else {
		for {
			chunk := make([]byte, 8192)
			n, err := file.Read(chunk)
			if err != nil && err != io.EOF {
				break
			}
			if n == 0 {
				break
			}
			sha1Hash.Write(chunk[:n])
		}
	}
	
	return hex.EncodeToString(sha1Hash.Sum(nil))
}

// List 目录遍历实现
func (a *AliPan) List(fileItem *schemas.FileItem) []*schemas.FileItem {
	if fileItem.Type == "file" {
		item := a.Detail(fileItem)
		if item != nil {
			return []*schemas.FileItem{item}
		}
		return []*schemas.FileItem{}
	}
	
	var parentFileId string
	var driveId string
	var err error
	
	if fileItem.Path == "/" {
		parentFileId = "root"
		driveId, err = a.defaultDriveId()
		if err != nil {
			logger.Errorf("获取默认驱动ID失败: %s", err.Error())
			return []*schemas.FileItem{}
		}
	} else {
		parentFileId = fileItem.FileId
		driveId = fileItem.DriveId
	}
	
	items := []*schemas.FileItem{}
	var nextMarker interface{}
	
	for {
		payload := map[string]interface{}{
			"drive_id":        driveId,
			"limit":           100,
			"marker":          nextMarker,
			"parent_file_id":  parentFileId,
		}
		
		resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/list", payload)
		if err != nil {
			logger.Errorf("【阿里云盘�?s 检索出错！: %s", fileItem.Path, err.Error())
			return []*schemas.FileItem{}
		}
		
		result, ok := resp.(map[string]interface{})
		if !ok {
			break
		}
		
		nextMarker, _ = result["next_marker"]
		itemsList, exists := result["items"].([]interface{})
		if !exists {
			break
		}
		
		for _, item := range itemsList {
			if itemMap, ok := item.(map[string]interface{}); ok {
				items = append(items, a.getFileItem(itemMap, fileItem.Path))
			}
		}
		
		if len(itemsList) < 100 {
			break
		}
	}
	
	return items
}

// delayGetItem 自动延迟重试 get_item 模块
func (a *AliPan) delayGetItem(path string) *schemas.FileItem {
	for i := 0; i < 2; i++ {
		time.Sleep(2 * time.Second)
		fileItem := a.GetItem(path)
		if fileItem != nil {
			return fileItem
		}
	}
	return nil
}

// CreateFolder 创建目录
func (a *AliPan) CreateFolder(parentItem *schemas.FileItem, name string) *schemas.FileItem {
	payload := map[string]interface{}{
		"drive_id":       parentItem.DriveId,
		"parent_file_id": parentItem.FileId,
		"name":           name,
		"type":           "folder",
	}
	
	if parentItem.FileId == "" {
		payload["parent_file_id"] = "root"
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/create", payload)
	if err != nil {
		return nil
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Warnf("【阿里云盘】创建目录失�? %s", message)
		return nil
	}
	
	// 缓存新目�?	newPath := filepath.Join(parentItem.Path, name)
	return a.delayGetItem(newPath)
}

// calculatePreHash 计算文件�?KB的SHA1作为pre_hash
func (a *AliPan) calculatePreHash(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	
	sha1Hash := sha1.New()
	data := make([]byte, 1024)
	n, _ := file.Read(data)
	sha1Hash.Write(data[:n])
	
	return hex.EncodeToString(sha1Hash.Sum(nil))
}

// calculateProofCode 计算秒传所需的proof_code
func (a *AliPan) calculateProofCode(filePath string) string {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return ""
	}
	
	fileSize := fileInfo.Size()
	if fileSize == 0 {
		return ""
	}
	
	// Step 1-3: 计算access_token的MD5并取�?6�?	conf := a.GetConf()
	accessToken, _ := conf["access_token"].(string)
	md5Hash := md5.Sum([]byte(accessToken))
	hexStr := hex.EncodeToString(md5Hash[:])[:16]
	
	// Step 4: 转换为无符号int64
	tmpInt, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil {
		logger.Errorf("【阿里云盘】Invalid hex string for proof code calculation")
		return ""
	}
	
	// Step 5-7: 计算读取范围
	index := tmpInt % uint64(fileSize)
	start := int64(index)
	end := int64(index) + 8
	if end > fileSize {
		end = fileSize
	}
	
	// Step 8: 读取文件范围数据并编�?	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	
	file.Seek(start, 0)
	chunk := make([]byte, end-start)
	file.Read(chunk)
	
	return base64.StdEncoding.EncodeToString(chunk)
}

// calculateContentHash 计算整个文件的SHA1作为content_hash
func (a *AliPan) calculateContentHash(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	
	sha1Hash := sha1.New()
	for {
		chunk := make([]byte, 8192)
		n, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			break
		}
		if n == 0 {
			break
		}
		sha1Hash.Write(chunk[:n])
	}
	
	return hex.EncodeToString(sha1Hash.Sum(nil))
}

// createFile 创建文件请求，尝试秒�?func (a *AliPan) createFile(driveId string, parentFileId string, fileName string, filePath string, checkNameMode string, chunkSize int64) (map[string]interface{}, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	
	fileSize := fileInfo.Size()
	preHash := a.calculatePreHash(filePath)
	numParts := int(math.Ceil(float64(fileSize) / float64(chunkSize)))
	
	// 构建分片信息
	partInfoList := []map[string]interface{}{}
	for i := 0; i < numParts; i++ {
		partInfoList = append(partInfoList, map[string]interface{}{"part_number": i + 1})
	}
	
	// 构建请求数据
	data := map[string]interface{}{
		"drive_id":         driveId,
		"parent_file_id":   parentFileId,
		"name":             fileName,
		"type":             "file",
		"check_name_mode":  checkNameMode,
		"size":             fileSize,
		"pre_hash":         preHash,
		"part_info_list":   partInfoList,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/create", data)
	if err != nil {
		return nil, fmt.Errorf("【阿里云盘】创建文件失败！")
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("【阿里云盘】响应格式错误！")
	}
	
	if code, exists := result["code"]; exists && code == "PreHashMatched" {
		// 可以秒传
		proofCode := a.calculateProofCode(filePath)
		contentHash := a.calculateContentHash(filePath)
		
		delete(data, "pre_hash")
		data["proof_code"] = proofCode
		data["proof_version"] = "v1"
		data["content_hash"] = contentHash
		data["content_hash_name"] = "sha1"
		
		resp, err = a.requestAPI("POST", "/adrive/v1.0/openFile/create", data)
		if err != nil {
			return nil, fmt.Errorf("【阿里云盘】创建文件失败！")
		}
		
		result, ok = resp.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("【阿里云盘】响应格式错误！")
		}
		
		if code, exists := result["code"]; exists && code != nil {
			message, _ := result["message"].(string)
			return nil, fmt.Errorf(message)
		}
	}
	
	return result, nil
}

// refreshUploadUrls 刷新分片上传地址
func (a *AliPan) refreshUploadUrls(driveId string, fileId string, uploadId string, partNumbers []int) ([]map[string]interface{}, error) {
	partInfoList := []map[string]interface{}{}
	for _, num := range partNumbers {
		partInfoList = append(partInfoList, map[string]interface{}{"part_number": num})
	}
	
	data := map[string]interface{}{
		"drive_id":      driveId,
		"file_id":       fileId,
		"upload_id":     uploadId,
		"part_info_list": partInfoList,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/getUploadUrl", data)
	if err != nil {
		return nil, fmt.Errorf("【阿里云盘】刷新分片上传地址失败�?)
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("【阿里云盘】响应格式错误！")
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		return nil, fmt.Errorf(message)
	}
	
	partInfoListResult, _ := result["part_info_list"].([]map[string]interface{})
	return partInfoListResult, nil
}

// uploadPart 上传单个分片
func (a *AliPan) uploadPart(uploadUrl string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest("PUT", uploadUrl, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	
	return a.session.Do(req)
}

// listUploadedParts 获取已上传分片列�?func (a *AliPan) listUploadedParts(driveId string, fileId string, uploadId string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"drive_id":   driveId,
		"file_id":    fileId,
		"upload_id":  uploadId,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/listUploadedParts", data)
	if err != nil {
		return nil, fmt.Errorf("【阿里云盘】获取已上传分片失败�?)
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("【阿里云盘】响应格式错误！")
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		return nil, fmt.Errorf(message)
	}
	
	return result, nil
}

// completeUpload 标记上传完成
func (a *AliPan) completeUpload(driveId string, fileId string, uploadId string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"drive_id":   driveId,
		"file_id":    fileId,
		"upload_id":  uploadId,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/complete", data)
	if err != nil {
		return nil, fmt.Errorf("【阿里云盘】完成上传失败！")
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("【阿里云盘】响应格式错误！")
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		return nil, fmt.Errorf(message)
	}
	
	return result, nil
}

// Upload 文件上传：分片、支持秒�?func (a *AliPan) Upload(targetDir *schemas.FileItem, localPath string, newName *string) *schemas.FileItem {
	targetName := filepath.Base(localPath)
	if newName != nil {
		targetName = *newName
	}
	
	targetPath := filepath.Join(targetDir.Path, targetName)
	
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		logger.Errorf("无法获取文件信息: %s", err.Error())
		return nil
	}
	
	fileSize := fileInfo.Size()
	
	// 1. 创建文件并检查秒�?	chunkSize := int64(10 * 1024 * 1024) // 分片大小 10M
	createRes, err := a.createFile(targetDir.DriveId, targetDir.FileId, targetName, localPath, "refuse", chunkSize)
	if err != nil {
		logger.Errorf("创建文件失败: %s", err.Error())
		return nil
	}
	
	if rapidUpload, exists := createRes["rapid_upload"].(bool); exists && rapidUpload {
		logger.Infof("【阿里云盘�?s 秒传完成�?, targetName)
		return a.delayGetItem(targetPath)
	}
	
	if exist, exists := createRes["exist"].(bool); exists && exist {
		logger.Infof("【阿里云盘�?s 已存�?, targetName)
		return a.GetItem(targetPath)
	}
	
	// 2. 准备分片上传参数
	fileId, fileIdExists := createRes["file_id"].(string)
	if !fileIdExists || fileId == "" {
		logger.Warnf("【阿里云盘】创�?%s 文件失败�?, targetName)
		return nil
	}
	
	uploadId, uploadIdExists := createRes["upload_id"].(string)
	if !uploadIdExists {
		logger.Warnf("【阿里云盘】创�?%s 文件失败，缺少upload_id�?, targetName)
		return nil
	}
	
	partInfoList, partInfoListExists := createRes["part_info_list"].([]interface{})
	if !partInfoListExists {
		logger.Warnf("【阿里云盘】创�?%s 文件失败，缺少part_info_list�?, targetName)
		return nil
	}
	
	uploadedParts := make(map[int]bool)
	
	// 3. 获取已上传分�?	uploadedInfo, err := a.listUploadedParts(targetDir.DriveId, fileId, uploadId)
	if err != nil {
		logger.Warnf("【阿里云盘】获取已上传分片失败: %s", err.Error())
		return nil
	}
	
	if uploadedPartsList, exists := uploadedInfo["uploaded_parts"].([]interface{}); exists {
		for _, part := range uploadedPartsList {
			if partMap, ok := part.(map[string]interface{}); ok {
				if partNumber, exists := partMap["part_number"].(float64); exists {
					uploadedParts[int(partNumber)] = true
				}
			}
		}
	}
	
	// 4. 初始化进度条
	logger.Infof("【阿里云盘】开始上�? %s -> %s，分片数�?d", localPath, targetPath, len(partInfoList))
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(localPath))
	progressCallback.Start()
	
	// 5. 分片上传循环
	uploadedSize := int64(0)
	file, err := os.Open(localPath)
	if err != nil {
		logger.Errorf("无法打开文件: %s", err.Error())
		return nil
	}
	defer file.Close()
	
	for _, partInfo := range partInfoList {
		if partInfoMap, ok := partInfo.(map[string]interface{}); ok {
			partNum := int(partInfoMap["part_number"].(float64))
			
			// 计算分片参数
			start := int64(partNum-1) * chunkSize
			end := start + chunkSize
			if end > fileSize {
				end = fileSize
			}
			currentChunkSize := end - start
			
			// 更新进度条（已存在的分片�?			if _, exists := uploadedParts[partNum]; exists {
				uploadedSize += currentChunkSize
				percent := float64(uploadedSize*100) / float64(fileSize)
				progressCallback.Update(percent, fmt.Sprintf("%s 进度�?0.2f%%", localPath, percent))
				continue
			}
			
			// 准备分片数据
			file.Seek(start, 0)
			data := make([]byte, currentChunkSize)
			file.Read(data)
			
			// 上传分片（带重试逻辑�?			success := false
			for attempt := 0; attempt < 3; attempt++ { // 最大重试次�?				var uploadUrl string
				
				// 获取当前上传地址（可能刷新）
				if attempt > 0 {
					newUrls, err := a.refreshUploadUrls(targetDir.DriveId, fileId, uploadId, []int{partNum})
					if err != nil || len(newUrls) == 0 {
						logger.Warnf("【阿里云盘�?s 分片 %d 刷新上传地址失败: %s", targetName, partNum, err.Error())
						continue
					}
					uploadUrl, _ = newUrls[0]["upload_url"].(string)
				} else {
					uploadUrl, _ = partInfoMap["upload_url"].(string)
				}
				
				// 执行上传
				logger.Infof("【阿里云盘】开�?�?d�?上传 %s 分片 %d ...", attempt+1, targetName, partNum)
				response, err := a.uploadPart(uploadUrl, data)
				if err != nil {
					logger.Warnf("【阿里云盘�?s 分片 %d �?%d 次上传异常：%s�?, targetName, partNum, attempt+1, err.Error())
					continue
				}
				
				if response.StatusCode == 200 {
					success = true
					response.Body.Close()
					break
				} else {
					body, _ := io.ReadAll(response.Body)
					logger.Warnf("【阿里云盘�?s 分片 %d �?%d 次上传失败：%s�?, targetName, partNum, attempt+1, string(body))
					response.Body.Close()
				}
			}
			
			// 处理上传结果
			if success {
				uploadedParts[partNum] = true
				uploadedSize += currentChunkSize
				percent := float64(uploadedSize*100) / float64(fileSize)
				progressCallback.Update(percent, fmt.Sprintf("%s 进度�?0.2f%%", localPath, percent))
			} else {
				logger.Errorf("【阿里云盘�?s 分片 %d 上传失败�?, targetName, partNum)
				progressCallback.End()
				return nil
			}
		}
	}
	
	// 6. 关闭进度�?	progressCallback.Update(100, fmt.Sprintf("%s 进度�?00%%", localPath))
	progressCallback.End()
	
	// 7. 完成上传
	result, err := a.completeUpload(targetDir.DriveId, fileId, uploadId)
	if err != nil {
		logger.Errorf("【阿里云盘】完成上传失�? %s", err.Error())
		return nil
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Warnf("【阿里云盘�?s 上传失败�?s�?, targetName, message)
		return nil
	}
	
	return a.getFileItem(result, targetDir.Path)
}

// Download 带实时进度显示的下载
func (a *AliPan) Download(fileItem *schemas.FileItem, path string) string {
	payload := map[string]interface{}{
		"drive_id": fileItem.DriveId,
		"file_id":  fileItem.FileId,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/getDownloadUrl", payload)
	if err != nil {
		logger.Errorf("【阿里云盘】获取下载链接失�? %s", fileItem.Name)
		return ""
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		logger.Errorf("【阿里云盘】获取下载链接失�? %s", fileItem.Name)
		return ""
	}
	
	downloadUrl, exists := result["url"].(string)
	if !exists || downloadUrl == "" {
		logger.Errorf("【阿里云盘】下载链接为�? %s", fileItem.Name)
		return ""
	}
	
	localPath := path
	if localPath == "" {
		localPath = filepath.Join(config.Settings.TEMP_PATH, fileItem.Name)
	}
	
	// 获取文件大小
	fileSize := fileItem.Size
	
	// 初始化进度条
	logger.Infof("【阿里云盘】开始下�? %s -> %s", fileItem.Name, localPath)
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(fileItem.Path))
	progressCallback.Start()
	
	// 构建请求头，包含必要的认证信�?	headers := map[string]string{
		"User-Agent":      config.Settings.NORMAL_USER_AGENT,
		"Referer":         "https://www.aliyundrive.com/",
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Accept-Encoding": "gzip, deflate, br",
		"Connection":      "keep-alive",
		"Sec-Fetch-Dest":  "empty",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "cross-site",
	}
	
	conf := a.GetConf()
	if accessToken, exists := conf["access_token"].(string); exists {
		headers["Authorization"] = "Bearer " + accessToken
	}
	
	// 执行下载
	httpClient := httpclient.NewRequestUtils(headers)
	httpReq, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		logger.Errorf("【阿里云盘】创建下载请求失�? %s", err.Error())
		return ""
	}
	
	httpResp, err := httpClient.Client.Do(httpReq)
	if err != nil {
		logger.Errorf("【阿里云盘】执行下载请求失�? %s", err.Error())
		return ""
	}
	defer httpResp.Body.Close()
	
	if httpResp.StatusCode != 200 {
		logger.Errorf("【阿里云盘】下载请求失败，状态码: %d", httpResp.StatusCode)
		return ""
	}
	
	// 写入文件
	outFile, err := os.Create(localPath)
	if err != nil {
		logger.Errorf("【阿里云盘】创建本地文件失�? %s", err.Error())
		return ""
	}
	defer outFile.Close()
	
	// 下载数据并更新进�?	downloadedSize := int64(0)
	buffer := make([]byte, 1024*1024) // 1MB buffer
	
	for {
		n, err := httpResp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := outFile.Write(buffer[:n])
			if writeErr != nil {
				logger.Errorf("【阿里云盘】写入文件失�? %s", writeErr.Error())
				outFile.Close()
				os.Remove(localPath)
				return ""
			}
			
			// 更新进度
			downloadedSize += int64(n)
			if fileSize > 0 {
				percent := float64(downloadedSize*100) / float64(fileSize)
				progressCallback.Update(percent, fmt.Sprintf("%s 进度�?0.2f%%", fileItem.Path, percent))
			}
		}
		
		if err != nil {
			if err != io.EOF {
				logger.Errorf("【阿里云盘】下载过程中发生错误: %s", err.Error())
				outFile.Close()
				os.Remove(localPath)
				progressCallback.End()
				return ""
			}
			break
		}
	}
	
	// 完成下载
	progressCallback.Update(100, fmt.Sprintf("%s 进度�?00%%", fileItem.Path))
	progressCallback.End()
	logger.Infof("【阿里云盘】下载完�? %s", fileItem.Name)
	
	return localPath
}

// Check 检查存储是否可�?func (a *AliPan) Check() bool {
	conf := a.GetConf()
	accessToken, exists := conf["access_token"].(string)
	return exists && accessToken != ""
}

// Delete 删除文件/目录
func (a *AliPan) Delete(fileItem *schemas.FileItem) bool {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【阿里云盘】删除文件异�? %v", r)
		}
	}()
	
	payload := map[string]interface{}{
		"drive_id": fileItem.DriveId,
		"file_id":  fileItem.FileId,
	}
	
	_, err := a.requestAPI("POST", "/adrive/v1.0/openFile/recyclebin/trash", payload)
	return err == nil
}

// Rename 重命名文�?目录
func (a *AliPan) Rename(fileItem *schemas.FileItem, name string) bool {
	payload := map[string]interface{}{
		"drive_id": fileItem.DriveId,
		"file_id":  fileItem.FileId,
		"name":     name,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/update", payload)
	if err != nil {
		return false
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return false
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Warnf("【阿里云盘】重命名失败: %s", message)
		return false
	}
	
	return true
}

// GetItem 获取指定路径的文�?目录�?func (a *AliPan) GetItem(path string) *schemas.FileItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Debugf("【阿里云盘】获取文件信息异�? %v", r)
		}
	}()
	
	defaultDriveId, err := a.defaultDriveId()
	if err != nil {
		logger.Debugf("【阿里云盘】获取默认驱动ID失败: %s", err.Error())
		return nil
	}
	
	payload := map[string]interface{}{
		"drive_id":   defaultDriveId,
		"file_path":  path,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/get_by_path", payload)
	if err != nil {
		logger.Debugf("【阿里云盘】获取文件信息失�? %s", err.Error())
		return nil
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Debugf("【阿里云盘】获取文件信息失�? %s", message)
		return nil
	}
	
	parentPath := filepath.Dir(path)
	if parentPath == "." {
		parentPath = "/"
	}
	
	return a.getFileItem(result, parentPath)
}

// GetFolder 获取指定路径的文件夹，如不存在则创建
func (a *AliPan) GetFolder(path string) *schemas.FileItem {
	// 是否已存�?	folder := a.GetItem(path)
	if folder != nil {
		return folder
	}
	
	// 逐级查找和创建目�?	defaultDriveId, _ := a.defaultDriveId()
	fileItem := &schemas.FileItem{
		Storage: string(types.StorageSchemaAlipan),
		Path:    "/",
		DriveId: defaultDriveId,
	}
	
	// 分割路径
	relPath, err := filepath.Rel("/", path)
	if err != nil {
		logger.Warnf("【阿里云盘】路径解析失�? %s", err.Error())
		return nil
	}
	
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		if part == "" {
			continue
		}
		
		dirFile := a.findDir(fileItem, part)
		if dirFile != nil {
			fileItem = dirFile
		} else {
			dirFile = a.CreateFolder(fileItem, part)
			if dirFile == nil {
				logger.Warnf("【阿里云盘】创建目�?%s%s 失败�?, fileItem.Path, part)
				return nil
			}
			fileItem = dirFile
		}
	}
	
	return fileItem
}

// findDir 查找下级目录中匹配名称的目录
func (a *AliPan) findDir(fileItem *schemas.FileItem, name string) *schemas.FileItem {
	for _, subFolder := range a.List(fileItem) {
		if subFolder.Type != "dir" {
			continue
		}
		if subFolder.Name == name {
			return subFolder
		}
	}
	return nil
}

// Detail 获取文件/目录详细信息
func (a *AliPan) Detail(fileItem *schemas.FileItem) *schemas.FileItem {
	return a.GetItem(fileItem.Path)
}

// Copy 复制文件到指定路�?func (a *AliPan) Copy(fileItem *schemas.FileItem, path string, newName string) bool {
	destFileItem := a.GetItem(path)
	if destFileItem == nil || destFileItem.Type != "dir" {
		logger.Warnf("【阿里云盘】目标路�?%s 不存在或不是目录�?, path)
		return false
	}
	
	payload := map[string]interface{}{
		"drive_id":          fileItem.DriveId,
		"file_id":           fileItem.FileId,
		"to_drive_id":       fileItem.DriveId,
		"to_parent_file_id": destFileItem.FileId,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/copy", payload)
	if err != nil {
		return false
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return false
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Warnf("【阿里云盘】复制文件失�? %s", message)
		return false
	}
	
	// 重命�?	newPath := filepath.Join(path, fileItem.Name)
	newFile := a.delayGetItem(newPath)
	if newFile != nil {
		a.Rename(newFile, newName)
	}
	
	return true
}

// Move 移动文件到指定路�?func (a *AliPan) Move(fileItem *schemas.FileItem, path string, newName string) bool {
	targetFileItem := a.GetItem(path)
	if targetFileItem == nil || targetFileItem.Type != "dir" {
		logger.Warnf("【阿里云盘】目标路�?%s 不存在或不是目录�?, path)
		return false
	}
	
	payload := map[string]interface{}{
		"drive_id":       fileItem.DriveId,
		"file_id":        fileItem.FileId,
		"to_parent_file_id": targetFileItem.FileId,
		"new_name":       newName,
	}
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/openFile/move", payload)
	if err != nil {
		return false
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return false
	}
	
	if code, exists := result["code"]; exists && code != nil {
		message, _ := result["message"].(string)
		logger.Warnf("【阿里云盘】移动文件失�? %s", message)
		return false
	}
	
	return true
}

// Link 硬链接文件（阿里云盘不支持）
func (a *AliPan) Link(fileItem *schemas.FileItem, targetFile string) bool {
	// 阿里云盘不支持硬链接
	return false
}

// Softlink 软链接文件（阿里云盘不支持）
func (a *AliPan) Softlink(fileItem *schemas.FileItem, targetFile string) bool {
	// 阿里云盘不支持软链接
	return false
}

// Usage 获取存储使用情况
func (a *AliPan) Usage() *schemas.StorageUsage {
	defer func() {
		if r := recover(); r != nil {
			logger.Debugf("【阿里云盘】获取存储使用情况异�? %v", r)
		}
	}()
	
	resp, err := a.requestAPI("POST", "/adrive/v1.0/user/getSpaceInfo", nil)
	if err != nil {
		return nil
	}
	
	result, ok := resp.(map[string]interface{})
	if !ok {
		return nil
	}
	
	space, exists := result["personal_space_info"].(map[string]interface{})
	if !exists {
		return nil
	}
	
	totalSize, _ := space["total_size"].(float64)
	usedSize, _ := space["used_size"].(float64)
	
	return &schemas.StorageUsage{
		Total:     int64(totalSize),
		Available: int64(totalSize - usedSize),
	}
}
