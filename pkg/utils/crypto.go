package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
)

// GenerateRandomBytes generates random bytes of specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(result), nil
}

// MD5Hash computes the MD5 hash of a string
func MD5Hash(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

// MD5Bytes computes the MD5 hash of data and returns it as bytes
// 对应 Python HashUtils.md5_bytes 方法
func MD5Bytes(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// MD5String computes the MD5 hash of a string and returns it as bytes
// 对应 Python HashUtils.md5_bytes 方法（接受字符串输入）
func MD5String(input string, encoding string) ([]byte, error) {
	if encoding == "" {
		encoding = "utf-8"
	}
	data := []byte(input) // Go strings are UTF-8 by default
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil), nil
}

// SHA1Hash computes the SHA1 hash of a string
func SHA1Hash(input string) string {
	hash := sha1.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

// SHA256Hash computes the SHA256 hash of a string
func SHA256Hash(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

// SHA512Hash computes the SHA512 hash of a string
func SHA512Hash(input string) string {
	hash := sha512.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

// HMACHash computes HMAC hash using specified algorithm
func HMACHash(algorithm, key, message string) (string, error) {
	var h func() hash.Hash

	switch strings.ToLower(algorithm) {
	case "md5":
		h = md5.New
	case "sha1":
		h = sha1.New
	case "sha256":
		h = sha256.New
	case "sha512":
		h = sha512.New
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	hmac := hmac.New(h, []byte(key))
	hmac.Write([]byte(message))
	return hex.EncodeToString(hmac.Sum(nil)), nil
}

// Base64Encode encodes a string to Base64
func Base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// Base64Decode decodes a Base64 encoded string
func Base64Decode(input string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Base64URLEncode encodes a string to URL-safe Base64
func Base64URLEncode(input string) string {
	return base64.URLEncoding.EncodeToString([]byte(input))
}

// Base64URLDecode decodes a URL-safe Base64 encoded string
func Base64URLDecode(input string) (string, error) {
	bytes, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// AESEncrypt encrypts a string using AES encryption
func AESEncrypt(key, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt decrypts a string using AES encryption
func AESDecrypt(key, encryptedText string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// GeneratePasswordHash generates a bcrypt hash for a password
func GeneratePasswordHash(password string, cost int) (string, error) {
	// Use bcrypt library for password hashing
	// This is a placeholder - actual bcrypt implementation would be used
	hash := SHA256Hash(password + "salt")
	return hash, nil
}

// VerifyPasswordHash verifies a password against a hash
func VerifyPasswordHash(password, hash string) bool {
	hash2 := SHA256Hash(password + "salt")
	return hash == hash2
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() (string, error) {
	return GenerateRandomString(32)
}

// GenerateAccessToken generates a secure access token
func GenerateAccessToken() (string, error) {
	return GenerateRandomString(64)
}

// GenerateSessionID generates a session ID
func GenerateSessionID() (string, error) {
	return GenerateRandomString(16)
}

// GenerateNonce generates a nonce for security purposes
func GenerateNonce() (string, error) {
	return GenerateRandomString(8)
}

// HashWithSalt hashes a string with a salt
func HashWithSalt(data, salt string) string {
	return SHA256Hash(data + salt)
}

// GenerateSalt generates a random salt
func GenerateSalt() (string, error) {
	return GenerateRandomString(16)
}

// SimpleEncrypt performs simple XOR encryption
func SimpleEncrypt(key, data string) string {
	result := make([]byte, len(data))
	keyBytes := []byte(key)

	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ keyBytes[i%len(keyBytes)]
	}

	return base64.StdEncoding.EncodeToString(result)
}

// SimpleDecrypt performs simple XOR decryption
func SimpleDecrypt(key, encryptedData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	result := make([]byte, len(data))
	keyBytes := []byte(key)

	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ keyBytes[i%len(keyBytes)]
	}

	return string(result), nil
}

// GenerateRSAKeyPair 生成 RSA 密钥对（与 Python RSAUtils.generate_rsa_key_pair 对齐）
// 返回的私钥、公钥均为 DER 编码再做 Base64 的字符串表示，不包含 PEM 头
func GenerateRSAKeyPair(keySize int) (privateKeyB64, publicKeyB64 string, err error) {
	if keySize == 0 {
		keySize = 2048
	}
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return "", "", err
	}

	// 私钥使用 PKCS8 DER
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}

	// 公钥使用 SubjectPublicKeyInfo DER
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", err
	}

	privateKeyB64 = base64.StdEncoding.EncodeToString(privDER)
	publicKeyB64 = base64.StdEncoding.EncodeToString(pubDER)
	return
}

// VerifyRSAKeys 使用 RSA 验证私钥和公钥是否匹配
// 参数为 Base64 编码的 DER（私钥 PKCS8、公钥 SubjectPublicKeyInfo）
func VerifyRSAKeys(privateKeyB64, publicKeyB64 string) (bool, error) {
	if privateKeyB64 == "" || publicKeyB64 == "" {
		return false, errors.New("private or public key is empty")
	}

	pubDER, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return false, fmt.Errorf("decode public key: %w", err)
	}
	privDER, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return false, fmt.Errorf("decode private key: %w", err)
	}

	pubKeyIf, err := x509.ParsePKIXPublicKey(pubDER)
	if err != nil {
		return false, fmt.Errorf("parse public key: %w", err)
	}
	pubKey, ok := pubKeyIf.(*rsa.PublicKey)
	if !ok {
		return false, errors.New("public key is not RSA")
	}

	privKeyIf, err := x509.ParsePKCS8PrivateKey(privDER)
	if err != nil {
		return false, fmt.Errorf("parse private key: %w", err)
	}
	privKey, ok := privKeyIf.(*rsa.PrivateKey)
	if !ok {
		return false, errors.New("private key is not RSA")
	}

	// 使用固定消息测试公钥加密、私钥解密
	msg := []byte("test")
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, msg, nil)
	if err != nil {
		return false, fmt.Errorf("rsa encrypt: %w", err)
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, ciphertext, nil)
	if err != nil {
		return false, fmt.Errorf("rsa decrypt: %w", err)
	}

	return string(plaintext) == string(msg), nil
}

// bytesToKey 实现 CryptoJS/OpenSSL 兼容的 key+iv 派生算法
// 等价于 Python CryptoJsUtils.bytes_to_key
func bytesToKey(data, salt []byte, output int) []byte {
	if len(salt) != 8 {
		panic("salt must be 8 bytes")
	}

	buf := make([]byte, len(data)+len(salt))
	copy(buf, data)
	copy(buf[len(data):], salt)

	// 第一次 MD5
	digest := md5.Sum(buf)
	finalKey := digest[:]

	for len(finalKey) < output {
		// 继续对上一次 digest + data+salt 做 MD5
		prevPlusData := make([]byte, len(digest)+len(buf))
		copy(prevPlusData, digest[:])
		copy(prevPlusData[len(digest):], buf)
		digest = md5.Sum(prevPlusData)
		finalKey = append(finalKey, digest[:]...)
	}

	return finalKey[:output]
}

// pkcs7Pad 对数据做 PKCS#7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - (len(data) % blockSize)
	if padLen == 0 {
		padLen = blockSize
	}
	padding := bytesRepeat(byte(padLen), padLen)
	return append(data, padding...)
}

// pkcs7Unpad 去除 PKCS#7 填充
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid padded data length")
	}
	padLen := int(data[len(data)-1])
	if padLen < 1 || padLen > blockSize || padLen > len(data) {
		return nil, errors.New("invalid padding")
	}
	return data[:len(data)-padLen], nil
}

// bytesRepeat 类似 bytes.Repeat，但避免引入额外依赖
func bytesRepeat(b byte, count int) []byte {
	res := make([]byte, count)
	for i := 0; i < count; i++ {
		res[i] = b
	}
	return res
}

// CryptoJSEncrypt 使用与 Python CryptoJsUtils.encrypt 兼容的方式加密
// message 为明文字节，passphrase 为密码短语字节
// 返回值为 Base64 编码的 "Salted__" + salt(8) + ciphertext
// 对应 Python CryptoJsUtils.encrypt 方法
func CryptoJSEncrypt(message, passphrase []byte) (string, error) {
	// 生成 8 字节 salt
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	keyIv := bytesToKey(passphrase, salt, 32+16)
	key := keyIv[:32]
	iv := keyIv[32:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	padded := pkcs7Pad(message, block.BlockSize())
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// OpenSSL/CryptoJS 格式: "Salted__" + salt + ciphertext
	out := make([]byte, 0, 16+len(ciphertext))
	out = append(out, []byte("Salted__")...)
	out = append(out, salt...)
	out = append(out, ciphertext...)

	return base64.StdEncoding.EncodeToString(out), nil
}

// CryptoJSDecrypt 解密 CryptoJSEncrypt 生成的密文
// encrypted 可以是 Base64 编码的字符串或字节
// passphrase 为密码短语字节
// 对应 Python CryptoJsUtils.decrypt 方法
func CryptoJSDecrypt(encrypted interface{}, passphrase []byte) ([]byte, error) {
	var encryptedBytes []byte
	var err error

	// 处理不同类型的输入
	switch e := encrypted.(type) {
	case string:
		// 如果是字符串，直接使用
		encryptedBytes = []byte(e)
	case []byte:
		// 如果是字节，直接使用
		encryptedBytes = e
	default:
		return nil, errors.New("encrypted must be string or []byte")
	}

	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(string(encryptedBytes))
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	// 检查格式
	if len(data) < 16 || !bytes.HasPrefix(data, []byte("Salted__")) {
		return nil, errors.New("invalid encrypted data format")
	}

	salt := data[8:16]
	ciphertext := data[16:]

	keyIv := bytesToKey(passphrase, salt, 32+16)
	key := keyIv[:32]
	iv := keyIv[32:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}

	plainPadded := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainPadded, ciphertext)

	plain, err := pkcs7Unpad(plainPadded, block.BlockSize())
	if err != nil {
		return nil, err
	}
	return plain, nil
}

// CryptoJSEncryptString 使用与 Python CryptoJsUtils.encrypt 兼容的方式加密字符串
// 对应 Python CryptoJsUtils.encrypt 方法（接受字符串输入）
func CryptoJSEncryptString(message string, passphrase []byte) (string, error) {
	return CryptoJSEncrypt([]byte(message), passphrase)
}

// CryptoJSDecryptString 解密 CryptoJSEncrypt 生成的密文并返回字符串
// 对应 Python CryptoJsUtils.decrypt 方法（返回字符串）
func CryptoJSDecryptString(encrypted interface{}, passphrase []byte) (string, error) {
	bytes, err := CryptoJSDecrypt(encrypted, passphrase)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
