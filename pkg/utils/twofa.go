package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// TwoFactorAuth 双因素认证助手
type TwoFactorAuth struct {
	code   string
	secret string
}

// NewTwoFactorAuth 创建双因素认证实例
func NewTwoFactorAuth(codeOrSecret string) *TwoFactorAuth {
	tfa := &TwoFactorAuth{}
	
	if len(codeOrSecret) >= 16 {
		tfa.secret = strings.ToUpper(strings.TrimSpace(codeOrSecret))
		tfa.code = ""
	} else {
		tfa.code = strings.TrimSpace(codeOrSecret)
		tfa.secret = ""
	}
	
	return tfa
}

// GetCode 获取验证码
func (tfa *TwoFactorAuth) GetCode() string {
	if tfa.code != "" {
		return tfa.code
	}
	
	if tfa.secret == "" {
		return ""
	}
	
	return tfa.calculateCode(tfa.secret)
}

// calculateCode 计算动态验证码
func (tfa *TwoFactorAuth) calculateCode(secretKey string) string {
	if secretKey == "" {
		return ""
	}

	// 解码Base32密钥
	key, err := base32Decode(secretKey)
	if err != nil {
		return ""
	}

	// 获取当前时间步
	inputTime := time.Now().Unix() / 30

	// 将时间步转换为8字节大端序
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(inputTime))

	// 计算HMAC-SHA1
	hash := hmac.New(sha1.New, key)
	hash.Write(msg)
	hashBytes := hash.Sum(nil)

	// 动态截取
	offset := int(hashBytes[len(hashBytes)-1] & 0x0f)
	code := binary.BigEndian.Uint32(hashBytes[offset:offset+4]) & 0x7fffffff

	// 转换为6位数字
	result := code % 1000000

	return fmt.Sprintf("%06d", result)
}

// VerifyCode 验证验证码
func (tfa *TwoFactorAuth) VerifyCode(inputCode string) bool {
	if tfa.secret == "" {
		return false
	}

	// 允许前后30秒的时间窗口
	for i := -1; i <= 1; i++ {
		expectedCode := tfa.calculateCodeAtTime(tfa.secret, time.Now().Unix()/30+int64(i))
		if expectedCode == inputCode {
			return true
		}
	}

	return false
}

// calculateCodeAtTime 计算指定时间的验证码
func (tfa *TwoFactorAuth) calculateCodeAtTime(secretKey string, timeStep int64) string {
	if secretKey == "" {
		return ""
	}

	key, err := base32Decode(secretKey)
	if err != nil {
		return ""
	}

	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(timeStep))

	hash := hmac.New(sha1.New, key)
	hash.Write(msg)
	hashBytes := hash.Sum(nil)

	offset := int(hashBytes[len(hashBytes)-1] & 0x0f)
	code := binary.BigEndian.Uint32(hashBytes[offset:offset+4]) & 0x7fffffff

	result := code % 1000000
	return fmt.Sprintf("%06d", result)
}

// GetRemainingTime 获取当前验证码剩余有效时间
func (tfa *TwoFactorAuth) GetRemainingTime() int {
	currentTime := time.Now().Unix()
	timeStep := currentTime / 30
	nextTimeStep := (timeStep + 1) * 30
	return int(nextTimeStep - currentTime)
}

// GenerateSecret 生成随机密钥
func GenerateSecret() (string, error) {
	// 生成16字节的随机数据
	randomBytes := make([]byte, 16)
	for i := range randomBytes {
		randomBytes[i] = byte(time.Now().UnixNano() % 256)
	}

	// 转换为Base32编码
	return base32Encode(randomBytes), nil
}

// ValidateSecret 验证密钥格式
func ValidateSecret(secret string) bool {
	if secret == "" {
		return false
	}

	// 移除空格并转为大写
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))

	// 检查长度（应该是16的倍数）
	if len(secret)%8 != 0 {
		return false
	}

	// 检查是否只包含Base32字符
	for _, char := range secret {
		if !((char >= 'A' && char <= 'Z') || (char >= '2' && char <= '7') || char == '=') {
			return false
		}
	}

	// 尝试解码
	_, err := base32Decode(secret)
	return err == nil
}

// base32Decode Base32解码
func base32Decode(input string) ([]byte, error) {
	// 移除空格并转为大写
	input = strings.ToUpper(strings.ReplaceAll(input, " ", ""))
	
	// 标准Base32字符集
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	
	// 创建解码映射
	decodeMap := make(map[byte]int)
	for i, char := range alphabet {
		decodeMap[byte(char)] = i
	}
	
	// 移除填充字符
	input = strings.TrimRight(input, "=")
	
	// 计算输出长度
	outputLen := len(input) * 5 / 8
	output := make([]byte, outputLen)
	
	buffer := 0
	bitsLeft := 0
	index := 0
	
	for _, char := range input {
		if char == '=' {
			continue
		}
		
		val, exists := decodeMap[byte(char)]
		if !exists {
			return nil, fmt.Errorf("invalid base32 character: %c", char)
		}
		
		buffer = (buffer << 5) | val
		bitsLeft += 5
		
		if bitsLeft >= 8 {
			output[index] = byte(buffer >> uint(bitsLeft-8))
			index++
			bitsLeft -= 8
		}
	}
	
	return output, nil
}

// base32Encode Base32编码
func base32Encode(input []byte) string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	
	var result strings.Builder
	buffer := 0
	bitsLeft := 0
	
	for _, b := range input {
		buffer = (buffer << 8) | int(b)
		bitsLeft += 8
		
		for bitsLeft >= 5 {
			index := (buffer >> uint(bitsLeft-5)) & 0x1f
			result.WriteByte(alphabet[index])
			bitsLeft -= 5
		}
	}
	
	if bitsLeft > 0 {
		index := (buffer << uint(5-bitsLeft)) & 0x1f
		result.WriteByte(alphabet[index])
	}
	
	// 添加填充字符
	padding := (8 - (len(input) % 8)) % 8
	if padding > 0 {
		for i := 0; i < padding; i++ {
			result.WriteByte('=')
		}
	}
	
	return result.String()
}

// GetQRCodeURL 获取二维码URL（用于Google Authenticator等）
func (tfa *TwoFactorAuth) GetQRCodeURL(accountName, issuer string) string {
	if tfa.secret == "" {
		return ""
	}

	// URL编码
	accountName = strings.ReplaceAll(accountName, " ", "%20")
	issuer = strings.ReplaceAll(issuer, " ", "%20")

	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", 
		issuer, accountName, tfa.secret, issuer)
}

// TimeSyncTest 时间同步测试
func TimeSyncTest(secret string) (bool, error) {
	if !ValidateSecret(secret) {
		return false, fmt.Errorf("invalid secret format")
	}

	tfa := NewTwoFactorAuth(secret)
	
	// 生成当前时间的验证码
	currentCode := tfa.GetCode()
	if currentCode == "" {
		return false, fmt.Errorf("failed to generate code")
	}

	// 验证当前验证码
	if !tfa.VerifyCode(currentCode) {
		return false, fmt.Errorf("code verification failed")
	}

	return true, nil
}

// BackupCodes 生成备用验证码
func GenerateBackupCodes(count int) []string {
	if count <= 0 {
		count = 10
	}

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// 生成8位随机数字
		code := fmt.Sprintf("%08d", int(math.Abs(float64(time.Now().UnixNano()%100000000))))
		codes[i] = code
	}

	return codes
}