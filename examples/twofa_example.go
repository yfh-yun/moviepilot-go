package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
)

func main() {
	// 示例1: 使用预设的验证码
	fmt.Println("=== 使用预设的验证码 ===")
	tfa1 := helper.NewTwoFactorAuth("123456")
	code1 := tfa1.GetCode()
	fmt.Printf("预设验证�? %s\n", code1)

	// 示例2: 使用密钥动态计算验证码
	fmt.Println("\n=== 使用密钥动态计算验证码 ===")
	// 注意: 这里使用一个示例密钥，实际使用时应替换为真实的密钥
	tfa2 := helper.NewTwoFactorAuth("JBSWY3DPEHPK3PXP")
	code2 := tfa2.GetCode()
	fmt.Printf("动态验证码: %s\n", code2)

	// 示例3: 使用短密钥（会被当作预设验证码处理）
	fmt.Println("\n=== 使用短密�?===")
	tfa3 := helper.NewTwoFactorAuth("SHORT")
	code3 := tfa3.GetCode()
	fmt.Printf("短密钥处理结�? %s\n", code3)
}
