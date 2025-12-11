package utils

import (
	"fmt"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// GenerateSecretKey 生成二次验证密钥和对应的 TOTP URI，对应 Python OtpUtils.generate_secret_key
func GenerateSecretKey(username string) (string, string) {
	issuer := fmt.Sprintf("MoviePilot(%s)", username)
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: "MoviePilot",
	})
	if err != nil {
		return "", ""
	}
	return key.Secret(), key.URL()
}

// IsLegal 校验基于 URI 的二次验证密码，对应 Python OtpUtils.is_legal
func IsLegal(otpURI, password string) bool {
	key, err := otp.NewKeyFromURL(otpURI)
	if err != nil {
		return false
	}
	return totp.Validate(password, key.Secret())
}

// Check 使用明文 secret 校验二次验证密码，对应 Python OtpUtils.check
func Check(secret, password string) bool {
	return totp.Validate(password, secret)
}

// GetSecret 从 TOTP URI 中解析出 secret，对应 Python OtpUtils.get_secret
func GetSecret(otpURI string) string {
	key, err := otp.NewKeyFromURL(otpURI)
	if err != nil {
		return ""
	}
	return key.Secret()
}
