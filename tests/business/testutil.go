package business

// 测试辅助函数

// stringPtr 返回string指针的辅助函数
func stringPtr(s string) *string {
	return &s
}

// intPtr 返回int指针的辅助函数
func intPtr(i int) *int {
	return &i
}

// float64Ptr 返回float64指针的辅助函数
func float64Ptr(f float64) *float64 {
	return &f
}
