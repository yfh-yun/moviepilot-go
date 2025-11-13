package helper

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"moviepilot-go/internal/logger"
)

// TwoFactorAuth 两步验证�?type TwoFactorAuth struct {
	Code   string
	Secret string
}

// NewTwoFactorAuth 创建TwoFactorAuth实例
// code_or_secret: 验证码或密钥
func NewTwoFactorAuth(codeOrSecret string) *TwoFactorAuth {
	tfa := &TwoFactorAuth{}
	if codeOrSecret != "" && len(codeOrSecret) >= 16 {
		tfa.Code = ""
		tfa.Secret = codeOrSecret
	} else {
		tfa.Code = codeOrSecret
		tfa.Secret = ""
	}
	return tfa
}

// calc 计算动态验证码
// secret_key: 密钥
// 返回: 动态验证码
func (t *TwoFactorAuth) calc(secretKey string) string {
	if secretKey == "" {
		return ""
	}

	// 将密钥转换为大写并去除空�?	secretKey = strings.ToUpper(strings.ReplaceAll(secretKey, " ", ""))
	// 补齐Base32编码需要的等号
	missingPadding := len(secretKey) % 8
	if missingPadding != 0 {
		secretKey += strings.Repeat("=", 8-missingPadding)
	}

	// Base32解码
	key, err := base32.StdEncoding.DecodeString(secretKey)
	if err != nil {
		logger.Errorf("Base32解码失败: %v", err)
		return ""
	}

	// 计算时间�?	inputTime := time.Now().Unix() / 30

	// 打包时间�?	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(inputTime))

	// 使用HMAC-SHA1计算验证�?	h := hmac.New(sha1.New, key)
	h.Write(msg)
	hash := h.Sum(nil)

	// 计算偏移�?	offset := hash[19] & 0x0f

	// 获取4字节数据并计算最终验证码
	truncatedHash := binary.BigEndian.Uint32(hash[offset : offset+4])
	truncatedHash &= 0x7FFFFFFF
	code := truncatedHash % 1000000

	// 格式化为6位数�?	codeStr := fmt.Sprintf("%06d", code)
	return codeStr
}

// GetCode 获取验证�?// 返回: 验证�?func (t *TwoFactorAuth) GetCode() string {
	if t.Code != "" {
		return t.Code
	}
	return t.calc(t.Secret)
}
