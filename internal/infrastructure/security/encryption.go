// Package security 加密工具模块
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
)

// AESEncryptor AES加密器
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor 创建AES加密器
func NewAESEncryptor(key string) *AESEncryptor {
	hash := sha256.Sum256([]byte(key))
	return &AESEncryptor{
		key: hash[:],
	}
}

// Encrypt AES加密
func (a *AESEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES解密
func (a *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// RSAKeyPair RSA密钥对
type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// GenerateRSAKeyPair 生成RSA密钥对
func GenerateRSAKeyPair(bits int) (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	return &RSAKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// RSAEncryptor RSA加密器
type RSAEncryptor struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewRSAEncryptorFromKeys 从密钥创建RSA加密器
func NewRSAEncryptorFromKeys(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) *RSAEncryptor {
	return &RSAEncryptor{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

// NewRSAEncryptorFromPEM 从PEM文件创建RSA加密器
func NewRSAEncryptorFromPEM(publicKeyPEM, privateKeyPEM []byte) (*RSAEncryptor, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key PEM")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	block, _ = pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &RSAEncryptor{
		publicKey:  rsaPublicKey,
		privateKey: privateKey,
	}, nil
}

// Encrypt RSA加密
func (r *RSAEncryptor) Encrypt(plaintext string) (string, error) {
	if r.publicKey == nil {
		return "", fmt.Errorf("no public key available for encryption")
	}

	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		r.publicKey,
		[]byte(plaintext),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with RSA: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt RSA解密
func (r *RSAEncryptor) Decrypt(ciphertext string) (string, error) {
	if r.privateKey == nil {
		return "", fmt.Errorf("no private key available for decryption")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		r.privateKey,
		data,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt with RSA: %w", err)
	}

	return string(plaintext), nil
}

// ToPEM 转换为PEM格式
func (kp *RSAKeyPair) ToPEM() (publicKeyPEM, privateKeyPEM []byte, err error) {
	// 公钥PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(kp.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// 私钥PEM
	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(kp.PrivateKey),
	})

	return publicKeyPEM, privateKeyPEM, nil
}

// Hash 哈希工具
type Hash struct{}

// NewHash 创建哈希工具
func NewHash() *Hash {
	return &Hash{}
}

// SHA256 SHA256哈希
func (h *Hash) SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// HMACSHA256 HMAC-SHA256签名
func (h *Hash) HMACSHA256(data, secret string) string {
	key := []byte(secret)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSHA256 验证HMAC-SHA256签名
func (h *Hash) VerifyHMACSHA256(data, secret, signature string) bool {
	expectedSignature := h.HMACSHA256(data, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// SecureRandom 安全随机数生成器
type SecureRandom struct{}

// NewSecureRandom 创建安全随机数生成器
func NewSecureRandom() *SecureRandom {
	return &SecureRandom{}
}

// GenerateBytes 生成随机字节
func (sr *SecureRandom) GenerateBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateString 生成随机字符串
func (sr *SecureRandom) GenerateString(length int) (string, error) {
	bytes, err := sr.GenerateBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// CryptoJSEncryptor CryptoJS兼容的AES加密器
type CryptoJSEncryptor struct {
	key []byte
}

// NewCryptoJSEncryptor 创建CryptoJS兼容的AES加密器
func NewCryptoJSEncryptor(key []byte) *CryptoJSEncryptor {
	return &CryptoJSEncryptor{
		key: key,
	}
}

// EncryptAES CryptoJS兼容的AES加密
func (c *CryptoJSEncryptor) EncryptAES(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 使用ECB模式，与CryptoJS保持兼容
	if len(plaintext)%aes.BlockSize != 0 {
		// PKCS7填充
		plaintext = c.pkcs7Padding(plaintext)
	}

	encrypted := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += aes.BlockSize {
		block.Encrypt(encrypted[i:i+aes.BlockSize], []byte(plaintext[i:i+aes.BlockSize]))
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptAES CryptoJS兼容的AES解密
func (c *CryptoJSEncryptor) DecryptAES(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(data)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	decrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}

	// 移除PKCS7填充
	plaintext, err := c.pkcs7Unpadding(decrypted)
	if err != nil {
		return "", fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(plaintext), nil
}

// pkcs7Padding PKCS7填充
func (c *CryptoJSEncryptor) pkcs7Padding(plaintext string) string {
	blockSize := aes.BlockSize
	padding := blockSize - len(plaintext)%blockSize
	padtext := make([]byte, padding)
	for i := 0; i < padding; i++ {
		padtext[i] = byte(padding)
	}
	return plaintext + string(padtext)
}

// pkcs7Unpadding 移除PKCS7填充
func (c *CryptoJSEncryptor) pkcs7Unpadding(data []byte) (string, error) {
	length := len(data)
	if length == 0 {
		return "", fmt.Errorf("empty data")
	}
	
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return "", fmt.Errorf("invalid padding")
	}
	
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return "", fmt.Errorf("invalid padding")
		}
	}
	
	return string(data[:length-padding]), nil
}

// CryptoUtils 综合加密工具类
type CryptoUtils struct {
	aeEncryptor   *AESEncryptor
	cryptoJSEncryptor *CryptoJSEncryptor
	hashUtils     *HashUtils
}

// NewCryptoUtils 创建综合加密工具类
func NewCryptoUtils(key string) *CryptoUtils {
	hash := sha256.Sum256([]byte(key))
	return &CryptoUtils{
		aeEncryptor:   NewAESEncryptor(key),
		cryptoJSEncryptor: NewCryptoJSEncryptor(hash[:16]), // 使用前16字节作为AES密钥
		hashUtils:     NewHashUtils(),
	}
}

// NewCryptoUtilsWithKey 使用指定密钥创建加密工具类
func NewCryptoUtilsWithKey(key []byte) *CryptoUtils {
	return &CryptoUtils{
		aeEncryptor:   NewAESEncryptor(string(key)),
		cryptoJSEncryptor: NewCryptoJSEncryptor(key),
		hashUtils:     NewHashUtils(),
	}
}

// DecryptAES 使用AES解密（兼容CryptoJS）
func (c *CryptoUtils) DecryptAES(ciphertext string, key []byte) (string, error) {
	encryptor := NewCryptoJSEncryptor(key)
	return encryptor.DecryptAES(ciphertext)
}

// EncryptAESECB 使用AES-ECB模式加密（兼容CryptoJS）
func (c *CryptoUtils) EncryptAESECB(plaintext string, key []byte) (string, error) {
	encryptor := NewCryptoJSEncryptor(key)
	return encryptor.EncryptAES(plaintext)
}

// GenerateInt 生成随机整数
func (sr *SecureRandom) GenerateInt(min, max int) (int, error) {
	if min >= max {
		return 0, fmt.Errorf("min must be less than max")
	}

	bytes, err := sr.GenerateBytes(4)
	if err != nil {
		return 0, err
	}

	// 将字节转换为整数
	randomInt := int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	if randomInt < 0 {
		randomInt = -randomInt
	}

	return min + (randomInt % (max - min)), nil
}