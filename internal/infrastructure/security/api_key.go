package security

import (
	"errors"
)

var (
	// ErrAPIKeyInvalid API密钥无效
	ErrAPIKeyInvalid = errors.New("api key is invalid")
	// ErrAPITokenInvalid API令牌无效
	ErrAPITokenInvalid = errors.New("api token is invalid")
)

// VerifyAPIToken 验证API令牌
func VerifyAPIToken(providedToken, expectedToken string) error {
	if providedToken == "" || providedToken != expectedToken {
		return ErrAPITokenInvalid
	}
	return nil
}

// VerifyAPIKey 验证API密钥
func VerifyAPIKey(providedKey, expectedKey string) error {
	if providedKey == "" || providedKey != expectedKey {
		return ErrAPIKeyInvalid
	}
	return nil
}

// ExtractAPIKeyFromHeader 从请求头提取API密钥
func ExtractAPIKeyFromHeader(headerValue string) (string, error) {
	if headerValue == "" {
		return "", ErrAPIKeyInvalid
	}
	return headerValue, nil
}

// ExtractAPIKeyFromQuery 从查询参数提取API密钥
func ExtractAPIKeyFromQuery(queryValue string) (string, error) {
	if queryValue == "" {
		return "", ErrAPIKeyInvalid
	}
	return queryValue, nil
}

// ExtractAPITokenFromQuery 从查询参数提取API令牌
func ExtractAPITokenFromQuery(queryValue string) (string, error) {
	if queryValue == "" {
		return "", ErrAPITokenInvalid
	}
	return queryValue, nil
}
