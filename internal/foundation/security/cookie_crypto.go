// Package security CookieCloud加密工具模块
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// CryptoJSUtils CryptoJS兼容的工具类
type CryptoJSUtils struct{}

// NewCryptoJSUtils 创建CryptoJS工具实例
func NewCryptoJSUtils() *CryptoJSUtils {
	return &CryptoJSUtils{}
}

// Decrypt CryptoJS兼容的AES解密
// 支持ECB和CBC模式，与CryptoJS保持兼容
func (c *CryptoJSUtils) Decrypt(ciphertext string, key []byte) (string, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("invalid AES key length: %d", len(key))
	}

	// 解码Base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// 尝试CBC模式解密（带IV）
	if len(data) >= aes.BlockSize {
		iv := data[:aes.BlockSize]
		encrypted := data[aes.BlockSize:]
		
		if len(encrypted)%aes.BlockSize == 0 {
			result, err := c.decryptCBC(encrypted, key, iv)
			if err == nil && c.isValidPlaintext(result) {
				return result, nil
			}
		}
	}

	// 尝试ECB模式解密（无IV）
	if len(data)%aes.BlockSize == 0 {
		result, err := c.decryptECB(data, key)
		if err == nil && c.isValidPlaintext(result) {
			return result, nil
		}
	}

	return "", fmt.Errorf("failed to decrypt with both CBC and ECB modes")
}

// Encrypt 加密数据（使用CBC模式）
func (c *CryptoJSUtils) Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("invalid AES key length: %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// PKCS7填充
	plaintext = c.pkcs7Padding(plaintext, aes.BlockSize)
	
	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	// 这里应该使用随机数生成器，为了简化暂时使用固定值
	// 在实际应用中应该使用 crypto/rand
	
	// CBC模式加密
	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plaintext))
	mode.CryptBlocks(encrypted, []byte(plaintext))

	// 组合IV和加密数据
	result := append(iv, encrypted...)
	
	return base64.StdEncoding.EncodeToString(result), nil
}

// decryptCBC CBC模式解密
func (c *CryptoJSUtils) decryptCBC(ciphertext, key, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	// 移除填充
	result, err := c.pkcs7Unpadding(decrypted)
	if err != nil {
		return "", err
	}

	return result, nil
}

// decryptECB ECB模式解密
func (c *CryptoJSUtils) decryptECB(ciphertext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	decrypted := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], ciphertext[i:i+aes.BlockSize])
	}

	// 移除填充
	result, err := c.pkcs7Unpadding(decrypted)
	if err != nil {
		return "", err
	}

	return result, nil
}

// pkcs7Padding PKCS7填充
func (c *CryptoJSUtils) pkcs7Padding(plaintext string, blockSize int) string {
	padding := blockSize - len(plaintext)%blockSize
	padtext := make([]byte, padding)
	for i := 0; i < padding; i++ {
		padtext[i] = byte(padding)
	}
	return plaintext + string(padtext)
}

// pkcs7Unpadding 移除PKCS7填充
func (c *CryptoJSUtils) pkcs7Unpadding(data []byte) (string, error) {
	length := len(data)
	if length == 0 {
		return "", fmt.Errorf("empty data")
	}
	
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return "", fmt.Errorf("invalid padding length: %d", padding)
	}
	
	// 验证填充
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return "", fmt.Errorf("invalid padding")
		}
	}
	
	return string(data[:length-padding]), nil
}

// isValidPlaintext 检查明文是否有效
func (c *CryptoJSUtils) isValidPlaintext(plaintext string) bool {
	if plaintext == "" {
		return false
	}
	
	// 检查是否包含可打印字符
	for _, char := range plaintext {
		if char < 32 && char != 9 && char != 10 && char != 13 {
			// 不是制表符、换行符或回车符的不可打印字符
			return false
		}
	}
	
	return true
}

// ParseKeyFormat 解析密钥格式
// 支持多种密钥格式的转换
func (c *CryptoJSUtils) ParseKeyFormat(keyData interface{}) ([]byte, error) {
	switch k := keyData.(type) {
	case string:
		// 字符串密钥
		data, err := base64.StdEncoding.DecodeString(k)
		if err == nil {
			return data, nil
		}
		// 如果不是base64，直接使用字符串
		return []byte(k), nil
	case []byte:
		// 字节数组
		return k, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %T", keyData)
	}
}

// GenerateKeyFromString 从字符串生成AES密钥
func (c *CryptoJSUtils) GenerateKeyFromString(keyString string) []byte {
	hash := NewHashUtils()
	md5Hash := hash.MD5(keyString)
	
	// 如果MD5长度不够，使用SHA256截取
	if len(md5Hash) < 16 {
		sha256Hash := hash.SHA256(keyString)
		if len(sha256Hash) >= 16 {
			return []byte(sha256Hash[:16])
		}
	}
	
	// 返回前16字节作为AES密钥
	if len(md5Hash) >= 16 {
		return []byte(md5Hash[:16])
	}
	
	return []byte(md5Hash)
}

// DetectAndDecrypt 自动检测并解密
// 尝试多种密钥格式和解密模式
func (c *CryptoJSUtils) DetectAndDecrypt(ciphertext string, password string) (string, error) {
	// 尝试直接密码作为密钥
	key := c.GenerateKeyFromString(password)
	result, err := c.Decrypt(ciphertext, key)
	if err == nil {
		return result, nil
	}
	
	// 尝试base64解码的密码作为密钥
	if decodedKey, err := base64.StdEncoding.DecodeString(password); err == nil {
		key := decodedKey
		if len(key) > 32 {
			key = key[:32]
		}
		result, err := c.Decrypt(ciphertext, key)
		if err == nil {
			return result, nil
		}
	}
	
	// 尝试密码的SHA256哈希作为密钥
	hash := NewHashUtils()
	sha256Hash := hash.SHA256(password)
	if len(sha256Hash) >= 16 {
		key := []byte(sha256Hash[:16])
		result, err := c.Decrypt(ciphertext, key)
		if err == nil {
			return result, nil
		}
	}
	
	return "", fmt.Errorf("failed to decrypt with all available methods")
}

// ValidateEncryptedData 验证加密数据格式
func (c *CryptoJSUtils) ValidateEncryptedData(ciphertext string) error {
	if ciphertext == "" {
		return fmt.Errorf("empty ciphertext")
	}
	
	// 检查Base64格式
	_, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return fmt.Errorf("invalid base64 format: %w", err)
	}
	
	// 检查最小长度
	decoded, _ := base64.StdEncoding.DecodeString(ciphertext)
	if len(decoded) < aes.BlockSize {
		return fmt.Errorf("ciphertext too short")
	}
	
	return nil
}