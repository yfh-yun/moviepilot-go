// Package utils 提供加密相关的工具函�?package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// HashUtils 哈希工具�?type HashUtils struct{}

// MD5 生成数据的MD5哈希值，并以字符串形式返�?// data: 输入的数据，类型为字符串或字�?// encoding: 字符串编码类型，默认使用UTF-8
// 返回生成的MD5哈希字符�?func (h *HashUtils) MD5(data interface{}, encoding string) string {
	var byteData []byte
	
	switch v := data.(type) {
	case string:
		if encoding == "" {
			encoding = "utf-8"
		}
		byteData = []byte(v)
	case []byte:
		byteData = v
	default:
		// 其他类型转换为字符串再处�?		str := fmt.Sprintf("%v", v)
		byteData = []byte(str)
	}
	
	// 计算MD5哈希
	hash := md5.Sum(byteData)
	return hex.EncodeToString(hash[:])
}

// MD5Bytes 生成数据的MD5哈希值，并以字节形式返回
// data: 输入的数据，类型为字符串或字�?// encoding: 字符串编码类型，默认使用UTF-8
// 返回生成的MD5哈希二进制数�?func (h *HashUtils) MD5Bytes(data interface{}, encoding string) []byte {
	var byteData []byte
	
	switch v := data.(type) {
	case string:
		if encoding == "" {
			encoding = "utf-8"
		}
		byteData = []byte(v)
	case []byte:
		byteData = v
	default:
		// 其他类型转换为字符串再处�?		str := fmt.Sprintf("%v", v)
		byteData = []byte(str)
	}
	
	// 计算MD5哈希
	hash := md5.Sum(byteData)
	return hash[:]
}
