package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
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
