package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// CryptoJSUtils CryptoJS工具�?type CryptoJSUtils struct{}

// NewCryptoJSUtils 创建一个新�?CryptoJSUtils 实例
func NewCryptoJSUtils() *CryptoJSUtils {
	return &CryptoJSUtils{}
}

// Decrypt 解密数据
func (c *CryptoJSUtils) Decrypt(encrypted, key []byte, method string) ([]byte, error) {
	switch method {
	case "AES/CBC/PKCS7Padding":
		return c.decryptAESCBCPKCS7(encrypted, key)
	case "AES/CBC/PKCS5Padding":
		return c.decryptAESCBCPKCS7(encrypted, key) // PKCS#5和PKCS#7在AES中相�?	default:
		return nil, fmt.Errorf("不支持的解密方法: %s", method)
	}
}

// decryptAESCBCPKCS7 解密AES/CBC/PKCS7Padding数据
func (c *CryptoJSUtils) decryptAESCBCPKCS7(encrypted, key []byte) ([]byte, error) {
	// AES密钥长度必须�?6�?4�?2字节
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("无效的AES密钥长度: %d", len(key))
	}

	// 加密数据必须包含至少一个块大小和IV
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 检查数据长�?	blockSize := block.BlockSize()
	if len(encrypted) < blockSize {
		return nil, fmt.Errorf("加密数据太短")
	}

	// 提取IV（前16字节）和实际加密数据
	iv := encrypted[:blockSize]
	encryptedData := encrypted[blockSize:]

	// 检查加密数据长度是否是块大小的倍数
	if len(encryptedData)%blockSize != 0 {
		return nil, fmt.Errorf("加密数据长度不是块大小的倍数")
	}

	// 创建CBC模式解密�?	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密数据
	decrypted := make([]byte, len(encryptedData))
	mode.CryptBlocks(decrypted, encryptedData)

	// 去除PKCS#7填充
	decrypted, err = c.unpadPKCS7(decrypted)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// unpadPKCS7 去除PKCS#7填充
func (c *CryptoJSUtils) unpadPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("数据为空")
	}

	// 获取填充字节的�?	padding := int(data[len(data)-1])

	// 检查填充长度是否有�?	if padding == 0 || padding > aes.BlockSize || padding > len(data) {
		return nil, errors.New("无效的PKCS#7填充")
	}

	// 检查填充字节是否都相同
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("无效的PKCS#7填充")
		}
	}

	// 去除填充
	return data[:len(data)-padding], nil
}

// Encrypt 加密数据
func (c *CryptoJSUtils) Encrypt(data, key []byte, method string) ([]byte, error) {
	switch method {
	case "AES/CBC/PKCS7Padding":
		return c.encryptAESCBCPKCS7(data, key)
	case "AES/CBC/PKCS5Padding":
		return c.encryptAESCBCPKCS7(data, key) // PKCS#5和PKCS#7在AES中相�?	default:
		return nil, fmt.Errorf("不支持的加密方法: %s", method)
	}
}

// encryptAESCBCPKCS7 加密AES/CBC/PKCS7Padding数据
func (c *CryptoJSUtils) encryptAESCBCPKCS7(data, key []byte) ([]byte, error) {
	// AES密钥长度必须�?6�?4�?2字节
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("无效的AES密钥长度: %d", len(key))
	}

	// 创建加密�?	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// 添加PKCS#7填充
	paddedData := c.padPKCS7(data, aes.BlockSize)

	// 创建CBC模式加密�?	mode := cipher.NewCBCEncrypter(block, iv)

	// 加密数据
	encrypted := make([]byte, len(paddedData))
	mode.CryptBlocks(encrypted, paddedData)

	// 将IV和加密数据组�?	result := make([]byte, len(iv)+len(encrypted))
	copy(result[:len(iv)], iv)
	copy(result[len(iv):], encrypted)

	return result, nil
}

// padPKCS7 添加PKCS#7填充
func (c *CryptoJSUtils) padPKCS7(data []byte, blockSize int) []byte {
	// 计算需要填充的字节�?	padding := blockSize - len(data)%blockSize

	// 创建填充字节切片
	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	// 返回原数据和填充数据的组�?	return append(data, padText...)
}

// AesPadKey AES密钥填充
func (c *CryptoJSUtils) AesPadKey(key []byte) []byte {
	// 如果密钥长度已经�?6�?4�?2字节，则直接返回
	if len(key) == 16 || len(key) == 24 || len(key) == 32 {
		return key
	}

	// 如果密钥长度大于32字节，则截取�?2字节
	if len(key) > 32 {
		return key[:32]
	}

	// 如果密钥长度小于16字节，则�?填充�?6字节
	if len(key) < 16 {
		padded := make([]byte, 16)
		copy(padded, key)
		return padded
	}

	// 如果密钥长度�?6-24字节之间，则填充�?4字节
	if len(key) < 24 {
		padded := make([]byte, 24)
		copy(padded, key)
		return padded
	}

	// 如果密钥长度�?4-32字节之间，则填充�?2字节
	padded := make([]byte, 32)
	copy(padded, key)
	return padded
}

// Parse 解析加密数据
func (c *CryptoJSUtils) Parse(encrypted []byte, key []byte, method string) ([]byte, error) {
	// 这里简化处理，直接调用Decrypt方法
	return c.Decrypt(encrypted, key, method)
}

// ToHex 将字节切片转换为十六进制字符�?func (c *CryptoJSUtils) ToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// FromHex 将十六进制字符串转换为字节切�?func (c *CryptoJSUtils) FromHex(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// ToBase64 将字节切片转换为Base64字符�?func (c *CryptoJSUtils) ToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// FromBase64 将Base64字符串转换为字节切片
func (c *CryptoJSUtils) FromBase64(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}
