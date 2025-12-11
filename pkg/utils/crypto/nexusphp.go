package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
)

// NexusPHPEncrypted NexusPHP加密结果结构
// 兼容Python实现的JSON格式

type NexusPHPEncrypted struct {
	IV    string `json:"iv"`    // Base64编码的IV
	Value string `json:"value"` // Base64编码的密文
	Mac   string `json:"mac"`   // HMAC-SHA256的十六进制表示
	Tag   string `json:"tag"`   // 保留字段，默认为空字符串
}

// NexusPHPEncrypt NexusPHP加密实现
// 加密流程：
// 1. 生成随机16字节IV
// 2. AES-CBC加密 + PKCS#7填充
// 3. 计算MAC: HMAC-SHA256(iv_base64 + ciphertext_base64)
// 4. 构造JSON: {iv, value, mac, tag} 并整体Base64编码
func NexusPHPEncrypt(data string, key []byte) (string, error) {
	// 生成随机IV
	iv := make([]byte, AESBlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	ivB64 := base64.StdEncoding.EncodeToString(iv)

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS#7填充
	paddedData := pkcs7Pad([]byte(data), AESBlockSize)

	// AES-CBC加密
	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	// Base64编码密文
	ciphertextB64 := base64.StdEncoding.EncodeToString(ciphertext)

	// 计算MAC: HMAC-SHA256(iv_base64 + ciphertext_base64)
	macData := ivB64 + ciphertextB64
	h := hmac.New(sha256.New, key)
	h.Write([]byte(macData))
	mac := hex.EncodeToString(h.Sum(nil))

	// 构造加密结果
	encrypted := NexusPHPEncrypted{
		IV:    ivB64,
		Value: ciphertextB64,
		Mac:   mac,
		Tag:   "",
	}

	// JSON序列化
	jsonBytes, err := json.Marshal(encrypted)
	if err != nil {
		return "", err
	}

	// 整体Base64编码
	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}
