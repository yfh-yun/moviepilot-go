package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hex 计算字符串的SHA256哈希值并返回十六进制表示
func SHA256Hex(message string) string {
	hash := sha256.Sum256([]byte(message))
	return hex.EncodeToString(hash[:])
}

// SHA256 计算字节数组的SHA256哈希值
func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
