// Package main 提供加密工具使用示例
package main

import (
	"fmt"
	
	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== 加密工具使用示例 ===")
	
	// RSA工具示例
	fmt.Println("\n1. RSA工具示例:")
	rsaUtils := &utils.RSAUtils{}
	
	// 生成RSA密钥�?	privateKey, publicKey, err := rsaUtils.GenerateRSAKeyPair(2048)
	if err != nil {
		fmt.Printf("生成RSA密钥对失�? %v\n", err)
		return
	}
	
	fmt.Printf("生成的私钥长�? %d 字符\n", len(privateKey))
	fmt.Printf("生成的公钥长�? %d 字符\n", len(publicKey))
	
	// 验证RSA密钥�?	isValid := rsaUtils.VerifyRSAKeys(privateKey, publicKey)
	fmt.Printf("RSA密钥对验证结�? %t\n", isValid)
	
	// 使用无效密钥对验�?	invalidValid := rsaUtils.VerifyRSAKeys("invalid_private_key", publicKey)
	fmt.Printf("无效密钥对验证结�? %t\n", invalidValid)
	
	// Hash工具示例
	fmt.Println("\n2. Hash工具示例:")
	hashUtils := &utils.HashUtils{}
	
	// MD5哈希示例
	testCases := []interface{}{
		"Hello, World!",
		[]byte("Hello, World!"),
		"这是一条测试消�?,
		12345,
		"",
	}
	
	for i, testCase := range testCases {
		md5Hash := hashUtils.MD5(testCase, "utf-8")
		md5Bytes := hashUtils.MD5Bytes(testCase, "utf-8")
		
		fmt.Printf("测试用例 %d:\n", i+1)
		fmt.Printf("  输入: %v\n", testCase)
		fmt.Printf("  MD5哈希(字符�?: %s\n", md5Hash)
		fmt.Printf("  MD5哈希(字节): %x\n", md5Bytes)
	}
	
	// CryptoJS工具示例
	fmt.Println("\n3. CryptoJS工具示例:")
	cryptoUtils := &utils.CryptoJsUtils{}
	
	// 加密示例
	message := []byte("这是一条需要加密的消息")
	passphrase := []byte("mypassword")
	
	encrypted, err := cryptoUtils.Encrypt(message, passphrase)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	
	fmt.Printf("原始消息: %s\n", string(message))
	fmt.Printf("加密后数据长�? %d 字符\n", len(encrypted))
	
	// 解密示例
	decrypted, err := cryptoUtils.Decrypt(encrypted, passphrase)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	
	fmt.Printf("解密后消�? %s\n", string(decrypted))
	fmt.Printf("加密解密是否成功: %t\n", string(message) == string(decrypted))
	
	// 使用错误密码解密
	_, err = cryptoUtils.Decrypt(encrypted, []byte("wrongpassword"))
	if err != nil {
		fmt.Printf("使用错误密码解密失败（预期）: %v\n", err)
	}
}
