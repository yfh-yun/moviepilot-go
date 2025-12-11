package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// AES-CBC 相关常量
const (
	// AESBlockSize AES块大小
	AESBlockSize = 16
	// AESKeySize AES密钥大小（256位）
	AESKeySize = 32
)

// ErrInvalidKeySize 无效的密钥大小
var ErrInvalidKeySize = errors.New("invalid key size, must be 32 bytes for AES-256")

// ErrInvalidCiphertext 无效的密文
var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// ErrDecryptionFailed 解密失败
var ErrDecryptionFailed = errors.New("decryption failed")

// pkcs7Pad 使用PKCS#7填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpad 移除PKCS#7填充
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, ErrInvalidCiphertext
	}

	padding := int(data[length-1])
	if padding > blockSize || padding == 0 {
		return nil, ErrInvalidCiphertext
	}

	return data[:length-padding], nil
}

// AESEncrypt AES-CBC加密（兼容Python实现）
// 加密流程：随机IV + PKCS#7填充 + AES-CBC加密 + Base64(iv + ciphertext)
func AESEncrypt(data string, key []byte) (string, error) {
	// 检查密钥大小
	if len(key) != AESKeySize {
		return "", ErrInvalidKeySize
	}

	// 生成随机IV
	iv := make([]byte, AESBlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

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

	// 组合IV和密文
	combined := append(iv, ciphertext...)

	// Base64编码
	return base64.StdEncoding.EncodeToString(combined), nil
}

// AESDecrypt AES-CBC解密（兼容Python实现）
// 解密流程：Base64解码 → 拆分IV和密文 → AES-CBC解密 → 移除PKCS#7填充
func AESDecrypt(encrypted string, key []byte) (string, error) {
	// 检查密钥大小
	if len(key) != AESKeySize {
		return "", ErrInvalidKeySize
	}

	// Base64解码
	combined, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	// 检查数据长度
	if len(combined) < AESBlockSize {
		return "", ErrInvalidCiphertext
	}

	// 拆分IV和密文
	iv := combined[:AESBlockSize]
	ciphertext := combined[AESBlockSize:]

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// AES-CBC解密
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	// 移除PKCS#7填充
	unpadded, err := pkcs7Unpad(plaintext, AESBlockSize)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(unpadded), nil
}
