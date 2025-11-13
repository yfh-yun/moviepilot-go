package main

import (
	"fmt"
	"log"
	"time"

	"moviepilot-go/internal/helper"
)

func main() {
	// 获取RedisHelper实例
	redisHelper := helper.GetRedisHelper()

	// 测试连接
	if redisHelper.Test() {
		fmt.Println("成功连接到Redis服务�?)
	} else {
		log.Fatal("无法连接到Redis服务�?)
	}

	// 基本字符串操�?	fmt.Println("\n=== 基本操作 ===")
	err := redisHelper.Set("username", "john_doe", 30, "users")
	if err != nil {
		log.Printf("设置键值失�? %v", err)
	} else {
		fmt.Println("成功设置 username = john_doe")
	}

	// 获取�?	value, err := redisHelper.Get("username", "users")
	if err != nil {
		log.Printf("获取值失�? %v", err)
	} else {
		fmt.Printf("获取�?username = %v\n", value)
	}

	// 检查键是否存在
	exists := redisHelper.Exists("username", "users")
	fmt.Printf("username 键是否存�? %v\n", exists)

	// 复杂数据结构操作
	fmt.Println("\n=== 复杂数据结构操作 ===")
	userData := map[string]interface{}{
		"id":       12345,
		"name":     "John Doe",
		"email":    "john@example.com",
		"roles":    []string{"user", "admin"},
		"active":   true,
		"balance":  99.99,
		"metadata": map[string]interface{}{"last_login": time.Now().Unix()},
	}

	err = redisHelper.Set("user:12345", userData, 60, "users")
	if err != nil {
		log.Printf("设置用户数据失败: %v", err)
	} else {
		fmt.Println("成功设置用户数据")
	}

	// 获取复杂数据
	userResult, err := redisHelper.Get("user:12345", "users")
	if err != nil {
		log.Printf("获取用户数据失败: %v", err)
	} else {
		fmt.Printf("获取到用户数�? %+v\n", userResult)
	}

	// 删除�?	err = redisHelper.Delete("username", "users")
	if err != nil {
		log.Printf("删除键失�? %v", err)
	} else {
		fmt.Println("成功删除 username �?)
	}

	// 异步Redis操作示例
	fmt.Println("\n=== 异步Redis操作 ===")
	asyncRedisHelper := helper.GetAsyncRedisHelper()

	// 测试异步连接
	if asyncRedisHelper.Test() {
		fmt.Println("成功连接到Redis服务�?异步)")
	}

	// 异步设置�?	err = asyncRedisHelper.Set("async_key", "async_value", 30, "async_region")
	if err != nil {
		log.Printf("异步设置键值失�? %v", err)
	} else {
		fmt.Println("成功异步设置 async_key = async_value")
	}

	// 异步获取�?	asyncValue, err := asyncRedisHelper.Get("async_key", "async_region")
	if err != nil {
		log.Printf("异步获取值失�? %v", err)
	} else {
		fmt.Printf("异步获取�?async_key = %v\n", asyncValue)
	}

	// 异步检查键是否存在
	asyncExists, err := asyncRedisHelper.Exists("async_key", "async_region")
	if err != nil {
		log.Printf("异步检查键存在性失�? %v", err)
	} else {
		fmt.Printf("async_key 键是否存�? %v\n", asyncExists)
	}

	// 异步删除�?	err = asyncRedisHelper.Delete("async_key", "async_region")
	if err != nil {
		log.Printf("异步删除键失�? %v", err)
	} else {
		fmt.Println("成功异步删除 async_key �?)
	}

	// 清理测试数据
	redisHelper.Clear("users")
	asyncRedisHelper.Clear("async_region")

	fmt.Println("\n所有示例完�?)
}
