package helper

import (
	"testing"
)

func TestPluginHelper(t *testing.T) {
	// 测试创建PluginHelper实例
	t.Run("创建PluginHelper实例", func(t *testing.T) {
		pluginHelper := NewPluginHelper()
		if pluginHelper == nil {
			t.Error("无法创建PluginHelper实例")
		}
	})

	// 测试GetRepoInfo方法
	t.Run("测试GetRepoInfo方法", func(t *testing.T) {
		pluginHelper := NewPluginHelper()
		
		// 测试正常的repo url
		user, repo := pluginHelper.getRepoInfo("https://github.com/user/repo")
		if user != "user" || repo != "repo" {
			t.Errorf("解析repo url失败，期�? user/repo, 实际: %s/%s", user, repo)
		}
		
		// 测试空的repo url
		user, repo = pluginHelper.getRepoInfo("")
		if user != "" || repo != "" {
			t.Error("空repo url应该返回空字符串")
		}
	})

	// 测试Install方法的基本参数验�?	t.Run("测试Install方法参数验证", func(t *testing.T) {
		pluginHelper := NewPluginHelper()
		
		// 测试空pid
		success, message := pluginHelper.Install("", "https://github.com/user/repo", nil, false)
		if success != false {
			t.Error("空pid应该返回失败")
		}
		if message != "参数错误" {
			t.Error("空pid应该返回参数错误")
		}
		
		// 测试空repo url
		success, message = pluginHelper.Install("test-plugin", "", nil, false)
		if success != false {
			t.Error("空repo url应该返回失败")
		}
		if message != "参数错误" {
			t.Error("空repo url应该返回参数错误")
		}
		
		// 测试无效repo url格式
		success, message = pluginHelper.Install("test-plugin", "invalid-url", nil, false)
		if success != false {
			t.Error("无效repo url应该返回失败")
		}
		if message != "不支持的插件仓库地址格式" {
			t.Error("无效repo url应该返回不支持的插件仓库地址格式")
		}
	})
}
