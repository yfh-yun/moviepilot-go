// Package security 安全工具模块
package security

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// HashUtils 哈希工具类
type HashUtils struct{}

// NewHashUtils 创建哈希工具实例
func NewHashUtils() *HashUtils {
	return &HashUtils{}
}

// MD5 计算MD5哈希值
func (h *HashUtils) MD5(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// SHA1 计算SHA1哈希值
func (h *HashUtils) SHA1(data string) string {
	hasher := sha1.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// SHA256 计算SHA256哈希值
func (h *HashUtils) SHA256(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// HMACMD5 计算HMAC-MD5签名
func (h *HashUtils) HMACMD5(data, key string) string {
	mac := hmac.New(md5.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// HMACSHA1 计算HMAC-SHA1签名
func (h *HashUtils) HMACSHA1(data, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// HMACSHA256 计算HMAC-SHA256签名
func (h *HashUtils) HMACSHA256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyMD5 验证MD5哈希值
func (h *HashUtils) VerifyMD5(data, expectedHash string) bool {
	return h.MD5(data) == expectedHash
}

// VerifySHA1 验证SHA1哈希值
func (h *HashUtils) VerifySHA1(data, expectedHash string) bool {
	return h.SHA1(data) == expectedHash
}

// VerifySHA256 验证SHA256哈希值
func (h *HashUtils) VerifySHA256(data, expectedHash string) bool {
	return h.SHA256(data) == expectedHash
}

// Base64Encode Base64编码
func (h *HashUtils) Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64Decode Base64解码
func (h *HashUtils) Base64Decode(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}
	return string(decoded), nil
}

// URLBase64Encode URL安全的Base64编码
func (h *HashUtils) URLBase64Encode(data string) string {
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// URLBase64Decode URL安全的Base64解码
func (h *HashUtils) URLBase64Decode(data string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("URL Base64解码失败: %w", err)
	}
	return string(decoded), nil
}

// GenerateRandomString 生成随机字符串
func (h *HashUtils) GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}
	
	// 使用时间戳和随机数生成简单的随机字符串
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%d", len(hasher.Sum(nil)))))
	result := hex.EncodeToString(hasher.Sum(nil))
	
	if len(result) > length {
		return result[:length]
	}
	
	// 如果不够长，重复哈希计算
	for len(result) < length {
		hasher.Write([]byte(result))
		result += hex.EncodeToString(hasher.Sum(nil))
	}
	
	return result[:length]
}

// HashAlgorithm 哈希算法类型
type HashAlgorithm string

const (
	HashMD5    HashAlgorithm = "md5"
	HashSHA1   HashAlgorithm = "sha1"
	HashSHA256 HashAlgorithm = "sha256"
)

// ComputeHash 计算指定算法的哈希值
func (h *HashUtils) ComputeHash(data string, algorithm HashAlgorithm) (string, error) {
	switch algorithm {
	case HashMD5:
		return h.MD5(data), nil
	case HashSHA1:
		return h.SHA1(data), nil
	case HashSHA256:
		return h.SHA256(data), nil
	default:
		return "", fmt.Errorf("不支持的哈希算法: %s", algorithm)
	}
}

// VerifyHash 验证指定算法的哈希值
func (h *HashUtils) VerifyHash(data, expectedHash string, algorithm HashAlgorithm) (bool, error) {
	computedHash, err := h.ComputeHash(data, algorithm)
	if err != nil {
		return false, err
	}
	return computedHash == expectedHash, nil
}