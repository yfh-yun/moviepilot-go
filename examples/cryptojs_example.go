// Package main 提供CryptoJS兼容加密工具使用示例
package main

import (
	"fmt"
	
	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== CryptoJS兼容加密工具使用示例 ===")
	
	cryptoUtils := &utils.CryptoJsUtils{}
	
	// 测试加密解密
	testCases := []struct {
		message    string
		passphrase string
	}{
		{"Hello, World!", "password"},
		{"这是一条测试消�?, "mypassword"},
		{"", "emptymessage"},
		{"A", "singlechar"},
	}
	
	for i, tc := range testCases {
		fmt.Printf("\n测试用例 %d:\n", i+1)
		fmt.Printf("消息: '%s'\n", tc.message)
		fmt.Printf("密码: '%s'\n", tc.passphrase)
		
		// 加密
		encrypted, err := cryptoUtils.Encrypt([]byte(tc.message), []byte(tc.passphrase))
		if err != nil {
			fmt.Printf("加密失败: %v\n", err)
			continue
		}
		
		fmt.Printf("加密结果长度: %d 字符\n", len(encrypted))
		
		// 解密
		decrypted, err := cryptoUtils.Decrypt(encrypted, []byte(tc.passphrase))
		if err != nil {
			fmt.Printf("解密失败: %v\n", err)
			continue
		}
		
		fmt.Printf("解密结果: '%s'\n", string(decrypted))
		fmt.Printf("加密解密是否成功: %t\n", tc.message == string(decrypted))
	}
}
