package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// TwoFactorAuth 双因素认证类
type TwoFactorAuth struct {
	logger *zap.Logger
	code   string
	secret string
}

// NewTwoFactorAuth 创建双因素认证实例
func NewTwoFactorAuth(codeOrSecret string) *TwoFactorAuth {
	var code, secret string

	// 根据输入长度判断是验证码还是密钥
	if len(codeOrSecret) >= 16 {
		// 长度大于等于16，认为是密钥
		secret = codeOrSecret
	} else {
		// 否则认为是验证码
		code = codeOrSecret
	}

	return &TwoFactorAuth{
		logger: logger.GetLogger(),
		code:   code,
		secret: secret,
	}
}

// Calc 计算动态验证码
func Calc(secret string) string {
	if secret == "" {
		return ""
	}

	// 计算当前时间步长（每30秒一个步长）
	inputTime := time.Now().Unix() / 30

	// 解码密钥
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		logger.GetLogger().Error("解码密钥失败", zap.Error(err))
		return ""
	}

	// 打包时间步长为大端序的8字节数据
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(inputTime))

	// 使用HMAC-SHA1算法生成哈希值
	h := hmac.New(sha1.New, key)
	h.Write(msg)
	hash := h.Sum(nil)

	// 计算偏移量
	offset := hash[19] & 0xf

	// 从偏移量开始取4字节，转换为无符号整数
	code := binary.BigEndian.Uint32(hash[offset : offset+4])

	// 取模1000000，得到6位验证码
	code = code & 0x7fffffff
	code = code % 1000000

	// 格式化为6位字符串，不足前面补0
	return fmt.Sprintf("%06d", code)
}

// GetCode 获取验证码
func (t *TwoFactorAuth) GetCode() string {
	// 如果已有验证码，直接返回
	if t.code != "" {
		return t.code
	}

	// 否则计算动态验证码
	return Calc(t.secret)
}

// VerifyCode 验证验证码是否正确
func (t *TwoFactorAuth) VerifyCode(code string) bool {
	// 计算当前时间步长的验证码
	currentCode := t.GetCode()
	if currentCode == code {
		return true
	}

	// 验证前一个时间步长的验证码（允许30秒误差）
	prevAuth := NewTwoFactorAuth(t.secret)
	prevAuth.code = ""
	// 模拟前一个时间步长
	prevCode := CalcWithTimeOffset(t.secret, -30)
	if prevCode == code {
		return true
	}

	// 验证后一个时间步长的验证码（允许30秒误差）
	nextCode := CalcWithTimeOffset(t.secret, 30)
	if nextCode == code {
		return true
	}

	return false
}

// CalcWithTimeOffset 计算指定时间偏移的验证码
func CalcWithTimeOffset(secret string, offsetSeconds int) string {
	if secret == "" {
		return ""
	}

	// 计算指定时间偏移的时间步长
	inputTime := (time.Now().Unix() + int64(offsetSeconds)) / 30

	// 解码密钥
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		logger.GetLogger().Error("解码密钥失败", zap.Error(err))
		return ""
	}

	// 打包时间步长为大端序的8字节数据
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(inputTime))

	// 使用HMAC-SHA1算法生成哈希值
	h := hmac.New(sha1.New, key)
	h.Write(msg)
	hash := h.Sum(nil)

	// 计算偏移量
	offset := hash[19] & 0xf

	// 从偏移量开始取4字节，转换为无符号整数
	code := binary.BigEndian.Uint32(hash[offset : offset+4])

	// 取模1000000，得到6位验证码
	code = code & 0x7fffffff
	code = code % 1000000

	// 格式化为6位字符串，不足前面补0
	return fmt.Sprintf("%06d", code)
}

// GenerateSecret 生成随机密钥
func GenerateSecret() string {
	// 生成16字节的随机数据
	secretBytes := make([]byte, 10)
	_, err := rand.Read(secretBytes)
	if err != nil {
		logger.GetLogger().Error("生成随机密钥失败", zap.Error(err))
		return ""
	}

	// 编码为Base32字符串
	return base32.StdEncoding.EncodeToString(secretBytes)
}
