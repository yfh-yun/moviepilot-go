package utils

import (
	"fmt"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// OtpUtils OTP工具�?type OtpUtils struct{}

// NewOtpUtils 创建新的OTP工具类实�?func NewOtpUtils() *OtpUtils {
	return &OtpUtils{}
}

// GenerateSecretKey 生成密钥和URI
func (o *OtpUtils) GenerateSecretKey(username string) (string, string) {
	// 生成随机密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      fmt.Sprintf("MoviePilot(%s)", username),
		AccountName: "MoviePilot",
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
		SecretSize:  20,
	})
	
	if err != nil {
		fmt.Println(err.Error())
		return "", ""
	}
	
	secret := key.Secret()
	uri := key.URL()
	
	return secret, uri
}

// IsLegal 校验二次验证是否正确
func (o *OtpUtils) IsLegal(otpURI string, password string) bool {
	// 解析URI获取密钥
	key, err := otp.NewKeyFromURL(otpURI)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	
	// 验证密码
	valid := totp.Validate(password, key.Secret())
	return valid
}

// Check 校验二次验证是否正确
func (o *OtpUtils) Check(secret string, password string) bool {
	// 验证密码
	valid := totp.Validate(password, secret)
	return valid
}

// GetSecret 获取uri中的secret
func (o *OtpUtils) GetSecret(otpURI string) string {
	// 解析URI获取密钥
	key, err := otp.NewKeyFromURL(otpURI)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	
	return key.Secret()
}
