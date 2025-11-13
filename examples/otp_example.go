package main

import (
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"

	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== OTP工具示例 ===")

	// 创建OTP工具类实�?	otpUtils := utils.NewOtpUtils()

	// 测试生成密钥和URI
	fmt.Println("\n--- 生成密钥和URI ---")
	testGenerateSecretKey(otpUtils)

	// 测试校验二次验证
	fmt.Println("\n--- 校验二次验证 ---")
	testVerification(otpUtils)

	// 测试从URI获取密钥
	fmt.Println("\n--- 从URI获取密钥 ---")
	testGetSecret(otpUtils)
}

func testGenerateSecretKey(otpUtils *utils.OtpUtils) {
	// 生成密钥和URI
	username := "testuser"
	secret, uri := otpUtils.GenerateSecretKey(username)
	
	fmt.Printf("用户�? %s\n", username)
	fmt.Printf("生成的密�? %s\n", secret)
	fmt.Printf("生成的URI: %s\n", uri)
	
	if secret != "" && uri != "" {
		fmt.Println("�?密钥和URI生成成功")
	} else {
		fmt.Println("�?密钥和URI生成失败")
	}
}

func testVerification(otpUtils *utils.OtpUtils) {
	// 生成一个测试密�?	username := "testuser"
	secret, uri := otpUtils.GenerateSecretKey(username)
	
	if secret == "" || uri == "" {
		fmt.Println("�?无法生成测试密钥")
		return
	}
	
	// 生成当前的OTP密码
	otpPassword, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		fmt.Printf("�?生成OTP密码失败: %v\n", err)
		return
	}
	
	fmt.Printf("生成的OTP密码: %s\n", otpPassword)
	
	// 使用isLegal方法验证
	isValid1 := otpUtils.IsLegal(uri, otpPassword)
	fmt.Printf("IsLegal验证结果: %v\n", isValid1)
	
	// 使用check方法验证
	isValid2 := otpUtils.Check(secret, otpPassword)
	fmt.Printf("Check验证结果: %v\n", isValid2)
	
	// 测试错误密码
	isValid3 := otpUtils.Check(secret, "123456")
	fmt.Printf("错误密码验证结果: %v\n", isValid3)
	
	if isValid1 && isValid2 && !isValid3 {
		fmt.Println("�?OTP验证功能正常")
	} else {
		fmt.Println("�?OTP验证功能异常")
	}
}

func testGetSecret(otpUtils *utils.OtpUtils) {
	// 生成测试数据
	username := "testuser"
	_, uri := otpUtils.GenerateSecretKey(username)
	
	if uri == "" {
		fmt.Println("�?无法生成测试URI")
		return
	}
	
	// 从URI中提取密�?	extractedSecret := otpUtils.GetSecret(uri)
	fmt.Printf("从URI提取的密�? %s\n", extractedSecret)
	
	// 验证提取的密钥是否有�?	if extractedSecret != "" {
		fmt.Println("�?密钥提取成功")
	} else {
		fmt.Println("�?密钥提取失败")
	}
}
