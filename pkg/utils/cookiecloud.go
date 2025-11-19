package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CookieCloudHelper CookieCloud同步助手
type CookieCloudHelper struct {
	server       string
	key          string
	password     string
	enableLocal  bool
	localPath    string
	httpClient   *http.Client
	ignoreCookies []string
}

// CookieCloudData CookieCloud数据结构
type CookieCloudData struct {
	CookieData []Cookie `json:"cookie_data"`
	Updated     string   `json:"updated"`
	UUID        string   `json:"uuid"`
}

// Cookie Cookie数据结构
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires"`
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"http_only"`
	SameSite string `json:"same_site"`
}

// NewCookieCloudHelper 创建CookieCloud助手实例
func NewCookieCloudHelper(server, key, password string, enableLocal bool, localPath string) *CookieCloudHelper {
	return &CookieCloudHelper{
		server:      normalizeBaseURL(server),
		key:         strings.TrimSpace(key),
		password:    strings.TrimSpace(password),
		enableLocal: enableLocal,
		localPath:   localPath,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		ignoreCookies: []string{
			"CookieAutoDeleteBrowsingDataCleanup",
			"CookieAutoDeleteCleaningDiscarded",
		},
	}
}

// Download 从CookieCloud下载数据
func (cch *CookieCloudHelper) Download() (*CookieCloudData, error) {
	// 验证参数
	if err := cch.validateParams(); err != nil {
		return nil, err
	}

	var data *CookieCloudData
	var err error

	if cch.enableLocal {
		// 从本地读取数据
		data, err = cch.loadLocalEncryptData(cch.key)
		if err != nil {
			return nil, fmt.Errorf("failed to load local data: %v", err)
		}
	} else {
		// 从远程服务器下载数据
		data, err = cch.downloadFromServer()
		if err != nil {
			return nil, fmt.Errorf("failed to download from server: %v", err)
		}
	}

	if data == nil {
		return &CookieCloudData{
			CookieData: []Cookie{},
			Updated:    "",
			UUID:       "",
		}, nil
	}

	// 过滤忽略的Cookie
	data.CookieData = cch.filterCookies(data.CookieData)

	return data, nil
}

// Upload 上传数据到CookieCloud
func (cch *CookieCloudHelper) Upload(cookies []Cookie) error {
	if err := cch.validateParams(); err != nil {
		return err
	}

	// 过滤忽略的Cookie
	filteredCookies := cch.filterCookies(cookies)

	data := &CookieCloudData{
		CookieData: filteredCookies,
		Updated:    time.Now().Format(time.RFC3339),
		UUID:       generateUUID(),
	}

	if cch.enableLocal {
		return cch.saveLocalEncryptData(cch.key, data)
	}

	return cch.uploadToServer(data)
}

// validateParams 验证参数
func (cch *CookieCloudHelper) validateParams() error {
	if !cch.enableLocal && cch.server == "" {
		return fmt.Errorf("CookieCloud server URL is required when remote mode is enabled")
	}

	if cch.key == "" {
		return fmt.Errorf("CookieCloud key is required")
	}

	if cch.password == "" {
		return fmt.Errorf("CookieCloud password is required")
	}

	return nil
}

// downloadFromServer 从服务器下载数据
func (cch *CookieCloudHelper) downloadFromServer() (*CookieCloudData, error) {
	url := fmt.Sprintf("%s/get/%s", cch.server, cch.key)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cch.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var response struct {
		Data string `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Data == "" {
		return nil, fmt.Errorf("empty data received from server")
	}

	// 解密数据
	decryptedData, err := cch.decryptData(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	var cookieData CookieCloudData
	if err := json.Unmarshal([]byte(decryptedData), &cookieData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookie data: %v", err)
	}

	return &cookieData, nil
}

// uploadToServer 上传数据到服务器
func (cch *CookieCloudHelper) uploadToServer(data *CookieCloudData) error {
	// 加密数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	encryptedData, err := cch.encryptData(string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %v", err)
	}

	url := fmt.Sprintf("%s/upload/%s", cch.server, cch.key)

	requestData := map[string]string{
		"data": encryptedData,
	}

	jsonPayload, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cch.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

// loadLocalEncryptData 从本地加载加密数据
func (cch *CookieCloudHelper) loadLocalEncryptData(key string) (*CookieCloudData, error) {
	// 这里应该实现从本地文件系统读取数据的逻辑
	// 简化实现
	return nil, fmt.Errorf("local storage not implemented")
}

// saveLocalEncryptData 保存加密数据到本地
func (cch *CookieCloudHelper) saveLocalEncryptData(key string, data *CookieCloudData) error {
	// 这里应该实现保存数据到本地文件系统的逻辑
	// 简化实现
	return fmt.Errorf("local storage not implemented")
}

// encryptData 加密数据
func (cch *CookieCloudHelper) encryptData(data string) (string, error) {
	// 生成密钥
	key := cch.deriveKey(cch.key, cch.password)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptData 解密数据
func (cch *CookieCloudHelper) decryptData(encryptedData string) (string, error) {
	// 生成密钥
	key := cch.deriveKey(cch.key, cch.password)

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}

// deriveKey 派生密钥
func (cch *CookieCloudHelper) deriveKey(key, password string) []byte {
	// 简化的密钥派生，实际应该使用更安全的KDF
	combined := key + password
	derived := make([]byte, 32) // AES-256
	
	for i := 0; i < 32; i++ {
		if i < len(combined) {
			derived[i] = combined[i]
		} else {
			derived[i] = byte(i)
		}
	}

	return derived
}

// filterCookies 过滤忽略的Cookie
func (cch *CookieCloudHelper) filterCookies(cookies []Cookie) []Cookie {
	var filtered []Cookie

	for _, cookie := range cookies {
		shouldIgnore := false
		for _, ignore := range cch.ignoreCookies {
			if strings.Contains(cookie.Name, ignore) {
				shouldIgnore = true
				break
			}
		}

		if !shouldIgnore {
			filtered = append(filtered, cookie)
		}
	}

	return filtered
}

// normalizeBaseURL 标准化基础URL
func normalizeBaseURL(url string) string {
	if url == "" {
		return ""
	}

	// 移除末尾的斜杠
	url = strings.TrimSuffix(url, "/")

	// 确保有协议前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	return url
}

// generateUUID 生成UUID
func generateUUID() string {
	// 简化的UUID生成
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		time.Now().UnixNano(),
		time.Now().UnixNano()>>16,
		time.Now().UnixNano()>>32,
		time.Now().UnixNano()>>48,
		time.Now().UnixNano()>>64,
	)
}

// SetServer 设置服务器地址
func (cch *CookieCloudHelper) SetServer(server string) {
	cch.server = normalizeBaseURL(server)
}

// SetKey 设置密钥
func (cch *CookieCloudHelper) SetKey(key string) {
	cch.key = strings.TrimSpace(key)
}

// SetPassword 设置密码
func (cch *CookieCloudHelper) SetPassword(password string) {
	cch.password = strings.TrimSpace(password)
}

// SetLocalMode 设置本地模式
func (cch *CookieCloudHelper) SetLocalMode(enable bool, path string) {
	cch.enableLocal = enable
	cch.localPath = path
}

// SetTimeout 设置HTTP客户端超时
func (cch *CookieCloudHelper) SetTimeout(timeout time.Duration) {
	cch.httpClient.Timeout = timeout
}

// GetServer 获取服务器地址
func (cch *CookieCloudHelper) GetServer() string {
	return cch.server
}

// GetKey 获取密钥
func (cch *CookieCloudHelper) GetKey() string {
	return cch.key
}

// IsLocalMode 检查是否为本地模式
func (cch *CookieCloudHelper) IsLocalMode() bool {
	return cch.enableLocal
}

// TestConnection 测试连接
func (cch *CookieCloudHelper) TestConnection() error {
	if cch.enableLocal {
		// 测试本地存储
		return cch.testLocalConnection()
	}

	// 测试远程服务器连接
	url := fmt.Sprintf("%s/health", cch.server)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := cch.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

// testLocalConnection 测试本地连接
func (cch *CookieCloudHelper) testLocalConnection() error {
	// 简化实现
	if cch.localPath == "" {
		return fmt.Errorf("local path not specified")
	}
	return nil
}

// ExportCookies 导出Cookie为JSON字符串
func (cch *CookieCloudHelper) ExportCookies(cookies []Cookie) (string, error) {
	data := &CookieCloudData{
		CookieData: cookies,
		Updated:    time.Now().Format(time.RFC3339),
		UUID:       generateUUID(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal cookies: %v", err)
	}

	return string(jsonData), nil
}

// ImportCookies 从JSON字符串导入Cookie
func (cch *CookieCloudHelper) ImportCookies(jsonData string) ([]Cookie, error) {
	var data CookieCloudData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookies: %v", err)
	}

	return data.CookieData, nil
}