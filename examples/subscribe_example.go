package main

import (
	"fmt"
	
	"moviepilot-go/internal/helper"
)

func main() {
	fmt.Println("Subscribe Helper Example")
	
	// 创建订阅帮助类实�?	subscribeHelper := helper.NewSubscribeHelper()
	
	if subscribeHelper == nil {
		fmt.Println("Failed to create SubscribeHelper")
		return
	}
	
	fmt.Println("SubscribeHelper created successfully")
	
	// 显示订阅帮助类的基本信息
	fmt.Printf("Sub Reg URL: %s\n", subscribeHelper.SubReg)
	fmt.Printf("Sub Done URL: %s\n", subscribeHelper.SubDone)
	fmt.Printf("Sub Report URL: %s\n", subscribeHelper.SubReport)
	fmt.Printf("Sub Statistic URL: %s\n", subscribeHelper.SubStatistic)
	fmt.Printf("Sub Share URL: %s\n", subscribeHelper.SubShare)
	fmt.Printf("Sub Shares URL: %s\n", subscribeHelper.SubShares)
	fmt.Printf("Sub Share Statistic URL: %s\n", subscribeHelper.SubShareStatistic)
	fmt.Printf("Sub Fork URL: %s\n", subscribeHelper.SubFork)
	fmt.Printf("Admin Users Count: %d\n", len(subscribeHelper.AdminUsers))
	
	// 测试获取订阅统计数据
	fmt.Println("\n=== 获取订阅统计数据 ===")
	statistic := subscribeHelper.GetStatistic("movie", 1, 30, nil, nil, nil, nil)
	fmt.Printf("获取�?%d 条订阅统计数据\n", len(statistic))
	
	// 测试新增订阅统计
	fmt.Println("\n=== 新增订阅统计 ===")
	sub := map[string]interface{}{
		"name": "Test Subscribe",
		"type": "movie",
		"year": "2023",
	}
	
	result := subscribeHelper.SubReg(sub)
	fmt.Printf("新增订阅统计结果: %v\n", result)
	
	// 测试异步新增订阅统计
	fmt.Println("\n=== 异步新增订阅统计 ===")
	asyncResult := subscribeHelper.SubRegAsync(sub)
	fmt.Printf("异步新增订阅统计结果: %v\n", asyncResult)
	
	// 测试完成订阅统计
	fmt.Println("\n=== 完成订阅统计 ===")
	doneResult := subscribeHelper.SubDone(sub)
	fmt.Printf("完成订阅统计结果: %v\n", doneResult)
	
	// 测试异步完成订阅统计
	fmt.Println("\n=== 异步完成订阅统计 ===")
	asyncDoneResult := subscribeHelper.SubDoneAsync(sub)
	fmt.Printf("异步完成订阅统计结果: %v\n", asyncDoneResult)
	
	// 测试上报存量订阅统计
	fmt.Println("\n=== 上报存量订阅统计 ===")
	reportResult := subscribeHelper.SubReport()
	fmt.Printf("上报存量订阅统计结果: %v\n", reportResult)
	
	// 测试分享订阅
	fmt.Println("\n=== 分享订阅 ===")
	shareSuccess, shareMessage := subscribeHelper.SubShare(1, "Test Title", "Test Comment", "Test User")
	fmt.Printf("分享订阅结果: success=%v, message=%s\n", shareSuccess, shareMessage)
	
	// 测试删除分享
	fmt.Println("\n=== 删除分享 ===")
	deleteSuccess, deleteMessage := subscribeHelper.ShareDelete(1)
	fmt.Printf("删除分享结果: success=%v, message=%s\n", deleteSuccess, deleteMessage)
	
	// 测试复用分享的订�?	fmt.Println("\n=== 复用分享的订�?===")
	forkSuccess, forkMessage := subscribeHelper.SubFork(1)
	fmt.Printf("复用分享的订阅结�? success=%v, message=%s\n", forkSuccess, forkMessage)
	
	// 测试获取订阅分享数据
	fmt.Println("\n=== 获取订阅分享数据 ===")
	shares := subscribeHelper.GetShares(nil, 1, 30, nil, nil, nil, nil)
	fmt.Printf("获取�?%d 条订阅分享数据\n", len(shares))
	
	// 测试获取订阅分享统计数据
	fmt.Println("\n=== 获取订阅分享统计数据 ===")
	shareStatistics := subscribeHelper.GetShareStatistics()
	fmt.Printf("获取�?%d 条订阅分享统计数据\n", len(shareStatistics))
	
	// 测试获取用户UUID
	fmt.Println("\n=== 获取用户UUID ===")
	uuid := subscribeHelper.GetUserUUID()
	fmt.Printf("用户UUID: %s\n", uuid)
	
	// 测试获取Github用户
	fmt.Println("\n=== 获取Github用户 ===")
	githubUser := subscribeHelper.GetGithubUser()
	if githubUser != nil {
		fmt.Printf("Github用户: %s\n", *githubUser)
	} else {
		fmt.Println("未获取到Github用户")
	}
	
	// 测试判断是否是管理员
	fmt.Println("\n=== 判断是否是管理员 ===")
	isAdmin := subscribeHelper.IsAdminUser()
	fmt.Printf("是否是管理员: %v\n", isAdmin)
	
	fmt.Println("\nExample completed")
}
