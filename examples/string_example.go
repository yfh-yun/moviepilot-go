package main

import (
	"fmt"

	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== 字符串工具示�?===")

	// 创建字符串工具类实例
	stringUtils := utils.NewStringUtils()

	// 测试文件大小转换
	fmt.Println("\n--- 文件大小转换 ---")
	testFileSize(stringUtils)

	// 测试时间转换
	fmt.Println("\n--- 时间转换 ---")
	testTimeConversion(stringUtils)

	// 测试语言检�?	fmt.Println("\n--- 语言检�?---")
	testLanguageDetection(stringUtils)

	// 测试字符串清�?	fmt.Println("\n--- 字符串清�?---")
	testStringClear(stringUtils)

	// 测试URL处理
	fmt.Println("\n--- URL处理 ---")
	testUrlHandling(stringUtils)

	// 测试其他功能
	fmt.Println("\n--- 其他功能 ---")
	testOtherFunctions(stringUtils)
}

func testFileSize(stringUtils *utils.StringUtils) {
	// 测试数字大小转换
	fmt.Printf("NumFilesize(1024): %d\n", stringUtils.NumFilesize(1024))
	fmt.Printf("NumFilesize(\"1.5GB\"): %d\n", stringUtils.NumFilesize("1.5GB"))
	fmt.Printf("NumFilesize(\"2.2TB\"): %d\n", stringUtils.NumFilesize("2.2TB"))
	fmt.Printf("NumFilesize(\"500MB\"): %d\n", stringUtils.NumFilesize("500MB"))
	fmt.Printf("StrFilesize(1048576, 2): %s\n", stringUtils.StrFilesize(1048576, 2))
	fmt.Printf("StrFilesize(1073741824, 2): %s\n", stringUtils.StrFilesize(1073741824, 2))
}

func testTimeConversion(stringUtils *utils.StringUtils) {
	// 测试时间转换
	fmt.Printf("StrTimelong(3661): %s\n", stringUtils.StrTimelong(3661))
	fmt.Printf("StrTimelong(\"7261\"): %s\n", stringUtils.StrTimelong("7261"))
	fmt.Printf("StrSecends(3661): %s\n", stringUtils.StrSecends(3661))
	fmt.Printf("StrTimehours(150): %s\n", stringUtils.StrTimehours(150))
}

func testLanguageDetection(stringUtils *utils.StringUtils) {
	// 测试语言检�?	fmt.Printf("IsChinese(\"你好世界\"): %v\n", stringUtils.IsChinese("你好世界"))
	fmt.Printf("IsChinese(\"Hello World\"): %v\n", stringUtils.IsChinese("Hello World"))
	fmt.Printf("IsJapanese(\"こんにちは\"): %v\n", stringUtils.IsJapanese("こんにち�?))
	fmt.Printf("IsKorean(\"안녕하세요\"): %v\n", stringUtils.IsKorean("안녕하세�?))
	fmt.Printf("IsAllChinese(\"你好 世界\"): %v\n", stringUtils.IsAllChinese("你好 世界"))
	fmt.Printf("IsAllChinese(\"你好 World\"): %v\n", stringUtils.IsAllChinese("你好 World"))
	fmt.Printf("IsEnglishWord(\"Hello\"): %v\n", stringUtils.IsEnglishWord("Hello"))
	fmt.Printf("IsEnglishWord(\"Hello World\"): %v\n", stringUtils.IsEnglishWord("Hello World"))
}

func testStringClear(stringUtils *utils.StringUtils) {
	// 测试字符串清�?	fmt.Printf("Clear(\"Hello, World!\", \"\", false): %v\n", stringUtils.Clear("Hello, World!", "", false))
	fmt.Printf("Clear(\"Hello, World!\", \"\", true): %v\n", stringUtils.Clear("Hello, World!", "", true))
	testStr := "Hello, World!"
	fmt.Printf("ClearUpper(&testStr): %v\n", stringUtils.ClearUpper(&testStr))
	fmt.Printf("StrInt(\"1,234\"): %d\n", stringUtils.StrInt("1,234"))
	fmt.Printf("StrFloat(\"1,234.56\"): %f\n", stringUtils.StrFloat("1,234.56"))
}

func testUrlHandling(stringUtils *utils.StringUtils) {
	// 测试URL处理
	fmt.Printf("UrlEqual(\"http://www.example.com\", \"http://example.com\"): %v\n", 
		stringUtils.UrlEqual("http://www.example.com", "http://example.com"))
	fmt.Printf("GetUrlNetloc(\"https://example.com:8080/path\"): %v\n", 
		func() []string {
			scheme, netloc := stringUtils.GetUrlNetloc("https://example.com:8080/path")
			return []string{scheme, netloc}
		}())
	fmt.Printf("GetUrlDomain(\"https://www.example.com\"): %v\n", 
		stringUtils.GetUrlDomain("https://www.example.com"))
	fmt.Printf("GetUrlSld(\"https://sub.example.com:8080\"): %v\n", 
		stringUtils.GetUrlSld("https://sub.example.com:8080"))
	fmt.Printf("GetBaseUrl(\"https://example.com/path?q=1\"): %v\n", 
		stringUtils.GetBaseUrl("https://example.com/path?q=1"))
	name := "test:file*name?.txt"
	fmt.Printf("ClearFileName(\"test:file*name?.txt\"): %v\n", 
		stringUtils.ClearFileName(name))
}

func testOtherFunctions(stringUtils *utils.StringUtils) {
	// 测试其他功能
	fmt.Printf("GenerateRandomStr(10): %v\n", stringUtils.GenerateRandomStr(10))
	fmt.Printf("Md5Hash(\"hello\"): %v\n", stringUtils.Md5Hash("hello"))
	fmt.Printf("StrAmount(1234567, \"$\"): %v\n", stringUtils.StrAmount(1234567, "$"))
	fmt.Printf("CountWords(\"Hello 你好 World 世界\"): %v\n", stringUtils.CountWords("Hello 你好 World 世界"))
	fmt.Printf("IsNumber(\"123.45\"): %v\n", stringUtils.IsNumber("123.45"))
	fmt.Printf("IsNumber(\"abc\"): %v\n", stringUtils.IsNumber("abc"))
	fmt.Printf("FindCommonPrefix(\"hello\", \"help\"): %v\n", stringUtils.FindCommonPrefix("hello", "help"))
	fmt.Printf("IsLink(\"http://example.com\"): %v\n", stringUtils.IsLink("http://example.com"))
	fmt.Printf("IsLink(\"example.com\"): %v\n", stringUtils.IsLink("example.com"))
	fmt.Printf("IsMagnetLink(\"magnet:?xt=urn:btih:...\"): %v\n", stringUtils.IsMagnetLink("magnet:?xt=urn:btih:..."))
	
	// 测试自然排序�?	testStr := "abc123def456"
	naturalKeys := stringUtils.NaturalSortKey(&testStr)
	fmt.Printf("NaturalSortKey(\"abc123def456\"): %v\n", naturalKeys)
	
	// 测试序列格式�?	series := []int{1, 2, 3, 5, 6, 8, 9, 10}
	fmt.Printf("StrSeries([1,2,3,5,6,8,9,10]): %v\n", stringUtils.StrSeries(series))
	
	episodes := []int{1, 2, 3, 5, 6, 7, 10}
	fmt.Printf("FormatEp([1,2,3,5,6,7,10]): %v\n", stringUtils.FormatEp(episodes))
}
