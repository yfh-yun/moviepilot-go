package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"moviepilot-go/pkg/cache"
)

// pkcs7Padding PKCS#7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpadding PKCS#7 解填充
func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("解填充失败: 数据长度为 0")
	}

	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, fmt.Errorf("无效的填充")
	}

	return data[:length-padding], nil
}

// EncryptionManager 加密管理器接口
type EncryptionManager interface {
	// EncryptMessage 使用 Fernet 加密消息
	EncryptMessage(message string, key []byte) (string, error)
	// Decrypt 使用 Fernet 解密数据
	Decrypt(data []byte, key []byte) ([]byte, error)
	// HashSHA256 计算 SHA256 哈希
	HashSHA256(message string) string
	// AESDecrypt AES 解密
	AESDecrypt(data string, key string) (string, error)
	// AESEncrypt AES 加密
	AESEncrypt(data string, key string) (string, error)
	// NexusPHPEncrypt NexusPHP 加密
	NexusPHPEncrypt(dataStr string, key []byte) (string, error)
}

// encryptionManager 加密管理器实现
type encryptionManager struct {
	cacheBackend cache.CacheBackend // 缓存后端
	cacheTTL     time.Duration      // 缓存过期时间
}

// NewEncryptionManager 创建加密管理器
func NewEncryptionManager() EncryptionManager {
	return &encryptionManager{
		cacheBackend: nil, // 默认不使用缓存
		cacheTTL:     0,
	}
}

// NewEncryptionManagerWithCache 创建带缓存的加密管理器
func NewEncryptionManagerWithCache(cacheBackend cache.CacheBackend, cacheTTL time.Duration) EncryptionManager {
	return &encryptionManager{
		cacheBackend: cacheBackend,
		cacheTTL:     cacheTTL,
	}
}

// EncryptMessage 使用 AES 加密消息（带缓存）
func (em *encryptionManager) EncryptMessage(message string, key []byte) (string, error) {
	// 缓存键：使用 SHA256 哈希避免存储明文消息和密钥
	cacheKey := fmt.Sprintf("encrypt_message:%x:%x", sha256.Sum256([]byte(message)), sha256.Sum256(key))
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if em.cacheBackend != nil {
		cachedResult, hit, err := em.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if encryptedMessage, ok := cachedResult.(string); ok {
				return encryptedMessage, nil
			}
		}
	}

	// 缓存未命中，执行实际加密
	// 生成随机 IV
	iv := make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	if err != nil {
		return "", fmt.Errorf("生成 IV 失败: %w", err)
	}

	// 创建 AES 加密器
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return "", fmt.Errorf("创建 AES 加密器失败: %w", err)
	}

	// 填充数据
	paddedData := pkcs7Padding([]byte(message), aes.BlockSize)

	// 加密数据
	encrypted := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, paddedData)

	// 组合 IV 和密文
	combined := append(iv, encrypted...)

	// 编码为 base64
	result := base64.StdEncoding.EncodeToString(combined)

	// 如果缓存已初始化，将加密结果存入缓存
	if em.cacheBackend != nil {
		err = em.cacheBackend.Set(cacheKey, result, em.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响加密结果
			fmt.Printf("缓存加密结果失败: %v\n", err)
		}
	}

	return result, nil
}

// Decrypt 使用 AES 解密数据（带缓存）
func (em *encryptionManager) Decrypt(data []byte, key []byte) ([]byte, error) {
	// 缓存键：使用 SHA256 哈希避免存储密文和密钥
	cacheKey := fmt.Sprintf("decrypt:%x:%x", sha256.Sum256(data), sha256.Sum256(key))
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if em.cacheBackend != nil {
		cachedResult, hit, err := em.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if decryptedData, ok := cachedResult.([]byte); ok {
				return decryptedData, nil
			}
		}
	}

	// 缓存未命中，执行实际解密
	// 解码 base64 数据
	decodedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("base64 解码失败: %w", err)
	}

	if len(decodedData) < aes.BlockSize {
		return nil, fmt.Errorf("解密数据长度不足")
	}

	// 分离 IV 和密文
	iv := decodedData[:aes.BlockSize]
	encrypted := decodedData[aes.BlockSize:]

	// 创建 AES 解密器
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("创建 AES 解密器失败: %w", err)
	}

	// 解密数据
	dst := make([]byte, len(encrypted))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, encrypted)

	// 去除填充
	padding := int(dst[len(dst)-1])
	if padding < 1 || padding > aes.BlockSize {
		return nil, fmt.Errorf("无效的填充")
	}

	result := dst[:len(dst)-padding]

	// 如果缓存已初始化，将解密结果存入缓存
	if em.cacheBackend != nil {
		err = em.cacheBackend.Set(cacheKey, result, em.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响解密结果
			fmt.Printf("缓存解密结果失败: %v\n", err)
		}
	}

	return result, nil
}

// HashSHA256 计算 SHA256 哈希（带缓存）
func (em *encryptionManager) HashSHA256(message string) string {
	// 缓存键：使用 SHA256 哈希避免存储明文消息
	cacheKey := fmt.Sprintf("hash_sha256:%x", sha256.Sum256([]byte(message)))
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if em.cacheBackend != nil {
		cachedResult, hit, err := em.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if hashedMessage, ok := cachedResult.(string); ok {
				return hashedMessage
			}
		}
	}

	// 缓存未命中，执行实际哈希计算
	result := fmt.Sprintf("%x", sha256.Sum256([]byte(message)))

	// 如果缓存已初始化，将哈希结果存入缓存
	if em.cacheBackend != nil {
		err := em.cacheBackend.Set(cacheKey, result, em.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响哈希结果
			fmt.Printf("缓存哈希结果失败: %v\n", err)
		}
	}

	return result
}

// AESDecrypt AES 解密（带缓存）
func (em *encryptionManager) AESDecrypt(data string, key string) (string, error) {
	// 缓存键：使用 SHA256 哈希避免存储密文和密钥
	cacheKey := fmt.Sprintf("aes_decrypt:%x:%x", sha256.Sum256([]byte(data)), sha256.Sum256([]byte(key)))
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if em.cacheBackend != nil {
		cachedResult, hit, err := em.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if decryptedData, ok := cachedResult.(string); ok {
				return decryptedData, nil
			}
		}
	}

	// 缓存未命中，执行实际解密
	if data == "" {
		return "", nil
	}

	// 解码 base64 数据
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}

	if len(decodedData) < 16 {
		return "", fmt.Errorf("解密数据长度不足")
	}

	// 分离 IV 和密文
	iv := decodedData[:16]
	encrypted := decodedData[16:]

	// 创建 AES 解密器
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("创建 AES 解密器失败: %w", err)
	}

	// 解密数据
	dst := make([]byte, len(encrypted))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, encrypted)

	// 去除填充
	unpaddedData, err := pkcs7Unpadding(dst)
	if err != nil {
		return "", err
	}

	result := string(unpaddedData)

	// 如果缓存已初始化，将解密结果存入缓存
	if em.cacheBackend != nil {
		err = em.cacheBackend.Set(cacheKey, result, em.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响解密结果
			fmt.Printf("缓存 AES 解密结果失败: %v\n", err)
		}
	}

	return result, nil
}

// AESEncrypt AES 加密（带缓存）
func (em *encryptionManager) AESEncrypt(data string, key string) (string, error) {
	// 缓存键：使用 SHA256 哈希避免存储明文数据和密钥
	cacheKey := fmt.Sprintf("aes_encrypt:%x:%x", sha256.Sum256([]byte(data)), sha256.Sum256([]byte(key)))
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if em.cacheBackend != nil {
		cachedResult, hit, err := em.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if encryptedData, ok := cachedResult.(string); ok {
				return encryptedData, nil
			}
		}
	}

	// 缓存未命中，执行实际加密
	if data == "" {
		return "", nil
	}

	// 创建 AES 加密器
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("创建 AES 加密器失败: %w", err)
	}

	// 生成随机 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("生成 IV 失败: %w", err)
	}

	// 填充数据
	paddedData := pkcs7Padding([]byte(data), aes.BlockSize)

	// 加密数据
	encrypted := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, paddedData)

	// 组合 IV 和密文
	combined := append(iv, encrypted...)

	// 编码为 base64
	result := base64.StdEncoding.EncodeToString(combined)

	// 如果缓存已初始化，将加密结果存入缓存
	if em.cacheBackend != nil {
		err = em.cacheBackend.Set(cacheKey, result, em.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响加密结果
			fmt.Printf("缓存 AES 加密结果失败: %v\n", err)
		}
	}

	return result, nil
}

// NexusPHPEncrypt NexusPHP 加密（带缓存）
func (em *encryptionManager) NexusPHPEncrypt(dataStr string, key []byte) (string, error) {
	// 缓存键：使用 SHA256 哈希避免存储明文数据和密钥
	cacheKey := fmt.Sprintf("nexusphp_encrypt:%x:%x", sha256.Sum256([]byte(dataStr)), sha256.Sum256(key))
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if em.cacheBackend != nil {
		cachedResult, hit, err := em.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if encryptedData, ok := cachedResult.(string); ok {
				return encryptedData, nil
			}
		}
	}

	// 缓存未命中，执行实际加密
	// 生成随机 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("生成 IV 失败: %w", err)
	}

	// 编码 IV 为 base64
	ivBase64 := base64.StdEncoding.EncodeToString(iv)

	// 创建 AES 加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES 加密器失败: %w", err)
	}

	// 填充数据
	padding := aes.BlockSize - len(dataStr)%aes.BlockSize
	data := dataStr + string(make([]byte, padding))

	// 加密数据
	encrypted := make([]byte, len(data))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, []byte(data))

	// 编码密文为 base64
	ciphertextBase64 := base64.StdEncoding.EncodeToString(encrypted)

	// 生成 HMAC
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(ivBase64 + ciphertextBase64))
	macHex := fmt.Sprintf("%x", mac.Sum(nil))

	// 构造 JSON 数据
	jsonData := map[string]string{
		"iv":    ivBase64,
		"value": ciphertextBase64,
		"mac":   macHex,
		"tag":   "",
	}

	jsonStr, err := json.Marshal(jsonData)
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}

	// 编码 JSON 数据为 base64
	result := base64.StdEncoding.EncodeToString(jsonStr)

	// 如果缓存已初始化，将加密结果存入缓存
	if em.cacheBackend != nil {
		err = em.cacheBackend.Set(cacheKey, result, em.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响加密结果
			fmt.Printf("缓存 NexusPHP 加密结果失败: %v\n", err)
		}
	}

	return result, nil
}
