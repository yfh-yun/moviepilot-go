package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"go.uber.org/zap"
)

// CookieCloudHelper CookieCloud助手结构�?type CookieCloudHelper struct {
	server      string
	key         string
	password    string
	enableLocal bool
	localPath   string
	ignoreCookies []string
}

// CookieData Cookie数据结构
type CookieData struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
	Expires  int64  `json:"expires"`
}

// CookieCloudResponse CookieCloud响应结构
type CookieCloudResponse struct {
	Encrypted string `json:"encrypted"`
}

// NewCookieCloudHelper 创建新的CookieCloud助手实例
func NewCookieCloudHelper() *CookieCloudHelper {
	helper := &CookieCloudHelper{
		ignoreCookies: []string{"CookieAutoDeleteBrowsingDataCleanup", "CookieAutoDeleteCleaningDiscarded"},
	}
	helper.syncSetting()
	return helper
}

// syncSetting 同步CookieCloud配置�?func (cch *CookieCloudHelper) syncSetting() {
	/*
		同步CookieCloud配置�?	*/
	settings := config.GetConfig()
	
	urlUtils := NewUrlUtils()
	cch.server = urlUtils.StandardizeBaseURL(settings.CookieCloudHost)
	cch.key = strings.TrimSpace(settings.CookieCloudKey)
	cch.password = strings.TrimSpace(settings.CookieCloudPassword)
	cch.enableLocal = settings.CookieCloudEnableLocal
	cch.localPath = settings.CookiePath
}

// Download 从CookieCloud下载数据
func (cch *CookieCloudHelper) Download() (map[string]string, string) {
	/*
		从CookieCloud下载数据
		:return: Cookie数据、错误信�?	*/
	// 更新为最新设�?	cch.syncSetting()

	if ((cch.server == "" && !cch.enableLocal) || cch.key == "" || cch.password == "") {
		return nil, "CookieCloud参数不正�?
	}

	var result map[string]interface{}
	if cch.enableLocal {
		// 开启本地服务时，从本地直接读取数据
		result = cch.loadLocalEncryptData(cch.key)
		if result == nil {
			return map[string]string{}, "未从本地CookieCloud服务加载到cookie数据，请检查服务器设置、用户KEY及加密密码是否正�?
		}
	} else {
		urlUtils := NewUrlUtils()
		reqUrl := urlUtils.CombineUrl(cch.server, fmt.Sprintf("get/%s", cch.key), nil)
		
		// 发送HTTP请求
		httpUtils := NewHttpUtils()
		response, err := httpUtils.GetRes(reqUrl, map[string]string{"Content-Type": "application/json"}, "", nil, nil)
		if err != nil {
			return nil, "CookieCloud请求失败，请检查服务器地址、用户KEY及加密密码是否正�?
		}
		
		if response.StatusCode == 200 {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return map[string]string{}, fmt.Sprintf("�?s下载cookie数据错误�?s", cch.server, err.Error())
			}
			
			err = json.Unmarshal(body, &result)
			if err != nil {
				return map[string]string{}, fmt.Sprintf("�?s下载cookie数据错误�?s", cch.server, err.Error())
			}
			
			if result == nil {
				return map[string]string{}, fmt.Sprintf("未从%s下载到cookie数据", cch.server)
			}
		} else if response != nil {
			return nil, fmt.Sprintf("远程同步CookieCloud失败，错误码�?d", response.StatusCode)
		} else {
			return nil, "CookieCloud请求失败，请检查服务器地址、用户KEY及加密密码是否正�?
		}
	}

	encrypted, ok := result["encrypted"].(string)
	if !ok || encrypted == "" {
		return map[string]string{}, "未获取到cookie密文"
	} else {
		cryptKey := cch.getCryptKey()
		cryptoUtils := NewCryptoJSUtils()
		
		// 解密数据
		encryptedBytes, err := cryptoUtils.FromBase64(encrypted)
		if err != nil {
			return map[string]string{}, "cookie解密失败�? + err.Error()
		}
		
		decryptedData, err := cryptoUtils.Decrypt(encryptedBytes, cryptKey, "AES/CBC/PKCS7Padding")
		if err != nil {
			return map[string]string{}, "cookie解密失败�? + err.Error()
		}
		
		err = json.Unmarshal(decryptedData, &result)
		if err != nil {
			return map[string]string{}, "cookie解密失败�? + err.Error()
		}
	}

	if result == nil {
		return map[string]string{}, "cookie解密为空"
	}

	var contents map[string][]CookieData
	cookieData, ok := result["cookie_data"]
	if ok {
		contentsBytes, _ := json.Marshal(cookieData)
		json.Unmarshal(contentsBytes, &contents)
	} else {
		contentsBytes, _ := json.Marshal(result)
		json.Unmarshal(contentsBytes, &contents)
	}
	
	// 整理数据,使用domain域名的最后两级作为分组依�?	domainGroups := make(map[string][]CookieData)
	stringUtils := NewStringUtils()
	for site, cookies := range contents {
		for _, cookie := range cookies {
			domainKey := stringUtils.GetUrlDomain(cookie.Domain)
			if domainGroups[domainKey] == nil {
				domainGroups[domainKey] = []CookieData{cookie}
			} else {
				domainGroups[domainKey] = append(domainGroups[domainKey], cookie)
			}
		}
	}
	
	// 返回错误
	retCookies := make(map[string]string)
	// 索引�?	for domain, contentList := range domainGroups {
		if len(contentList) == 0 {
			continue
		}
		// 只有cf的cookie过滤�?		cloudflareCookie := true
		for _, content := range contentList {
			if content.Name != "cf_clearance" {
				cloudflareCookie = false
				break
			}
		}
		if cloudflareCookie {
			continue
		}
		// 站点Cookie
		var cookieParts []string
		for _, content := range contentList {
			// 检查是否在忽略列表�?			ignored := false
			for _, ignore := range cch.ignoreCookies {
				if content.Name == ignore {
					ignored = true
					break
				}
			}
			if content.Name != "" && !ignored {
				cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", content.Name, content.Value))
			}
		}
		if len(cookieParts) > 0 {
			retCookies[domain] = strings.Join(cookieParts, ";")
		}
	}
	return retCookies, ""
}

// getCryptKey 使用UUID和密码生成CookieCloud的加解密密钥
func (cch *CookieCloudHelper) getCryptKey() []byte {
	/*
		使用UUID和密码生成CookieCloud的加解密密钥
	*/
	combinedString := fmt.Sprintf("%s-%s", cch.key, cch.password)
	hashUtils := NewHashUtils()
	md5Hash := hashUtils.MD5([]byte(combinedString), "")
	return []byte(md5Hash)[:16]
}

// loadLocalEncryptData 获取本地CookieCloud数据
func (cch *CookieCloudHelper) loadLocalEncryptData(uuid string) map[string]interface{} {
	/*
		获取本地CookieCloud数据
	*/
	filePath := fmt.Sprintf("%s/%s.json", cch.localPath, uuid)
	
	// 检查文件是否存�?	// 在Go中检查文件是否存�?	if _, err := ioutil.ReadFile(filePath); err != nil {
		logger.GetLoggerManager().Warn(fmt.Sprintf("本地CookieCloud文件不存在：%s", filePath))
		return nil
	}

	// 读取文件
	dataBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.GetLoggerManager().Error("读取本地CookieCloud文件失败", zap.Error(err))
		return nil
	}
	
	var data map[string]interface{}
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		logger.GetLoggerManager().Error("解析本地CookieCloud文件失败", zap.Error(err))
		return nil
	}
	
	return data
}
