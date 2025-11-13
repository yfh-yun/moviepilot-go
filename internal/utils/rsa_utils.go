// Package utils 提供加密相关的工具函�?package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// RSAUtils RSA加密工具�?type RSAUtils struct{}

// GenerateRSAKeyPair 生成RSA密钥�?// keySize: 密钥长度，默�?048
// 返回私钥和公钥（Base64编码，无标识符）
func (r *RSAUtils) GenerateRSAKeyPair(keySize int) (privateKey, publicKey string, err error) {
	if keySize <= 0 {
		keySize = 2048
	}

	// 生成RSA密钥�?	privateKeyObj, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return "", "", fmt.Errorf("生成RSA密钥对失�? %v", err)
	}

	publicKeyObj := &privateKeyObj.PublicKey

	// 导出私钥为DER格式 (使用PKCS8格式以匹配Python版本)
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKeyObj)
	if err != nil {
		return "", "", fmt.Errorf("导出私钥失败: %v", err)
	}

	// 导出公钥为DER格式
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKeyObj)
	if err != nil {
		return "", "", fmt.Errorf("导出公钥失败: %v", err)
	}

	// 将DER格式的密钥编码为Base64
	privateKeyB64 := base64.StdEncoding.EncodeToString(privateKeyDER)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKeyDER)

	return privateKeyB64, publicKeyB64, nil
}

// VerifyRSAKeys 验证RSA私钥和公钥是否匹�?// privateKey: 私钥字符串（Base64编码，无标识符）
// publicKey: 公钥字符串（Base64编码，无标识符）
// 返回匹配结果
func (r *RSAUtils) VerifyRSAKeys(privateKey, publicKey string) bool {
	if privateKey == "" || publicKey == "" {
		return false
	}

	// 解码Base64编码的公钥和私钥
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		fmt.Printf("解码公钥失败: %v\n", err)
		return false
	}

	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		fmt.Printf("解码私钥失败: %v\n", err)
		return false
	}

	// 加载公钥
	publicKeyObj, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		fmt.Printf("加载公钥失败: %v\n", err)
		return false
	}

	// 加载私钥 (使用PKCS8格式以匹配Python版本)
	privateKeyObj, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		fmt.Printf("加载私钥失败: %v\n", err)
		return false
	}

	// 测试加解�?	message := []byte("test")
	encryptedMessage, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKeyObj.(*rsa.PublicKey),
		message,
		nil,
	)
	if err != nil {
		fmt.Printf("加密测试失败: %v\n", err)
		return false
	}

	decryptedMessage, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKeyObj.(*rsa.PrivateKey),
		encryptedMessage,
		nil,
	)
	if err != nil {
		fmt.Printf("解密测试失败: %v\n", err)
		return false
	}

	return string(message) == string(decryptedMessage)
}
