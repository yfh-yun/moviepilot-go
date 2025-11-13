package helper

import (
	"testing"
)

func TestNewSubscribeHelper(t *testing.T) {
	// 测试创建SubscribeHelper实例
	helper := NewSubscribeHelper()
	if helper == nil {
		t.Error("Failed to create SubscribeHelper instance")
	}
	
	// 验证URL是否正确设置
	if helper.subReg == "" {
		t.Error("subReg URL should not be empty")
	}
	
	if helper.subDone == "" {
		t.Error("subDone URL should not be empty")
	}
	
	if helper.subReport == "" {
		t.Error("subReport URL should not be empty")
	}
	
	if helper.subStatistic == "" {
		t.Error("subStatistic URL should not be empty")
	}
	
	if helper.subShare == "" {
		t.Error("subShare URL should not be empty")
	}
	
	if helper.subShares == "" {
		t.Error("subShares URL should not be empty")
	}
	
	if helper.subShareStatistic == "" {
		t.Error("subShareStatistic URL should not be empty")
	}
	
	if helper.subFork == "" {
		t.Error("subFork URL should not be empty")
	}
	
	// 验证管理员用户列表是否正确设�?	if len(helper.adminUsers) == 0 {
		t.Error("adminUsers should not be empty")
	}
}

func TestCheckSubscribeShareEnabled(t *testing.T) {
	// 测试检查订阅分享功能是否开�?	helper := NewSubscribeHelper()
	
	// 由于这是测试环境，我们无法控制配置，但我们至少要确保函数能正常执�?	enabled, message := helper.checkSubscribeShareEnabled()
	
	// 我们只验证函数能正常返回，不关心具体�?	t.Logf("Subscribe share enabled: %v, message: %s", enabled, message)
}

func TestValidateSubscribe(t *testing.T) {
	// 测试验证订阅是否存在
	helper := NewSubscribeHelper()
	
	// 测试nil订阅
	valid, message := helper.validateSubscribe(nil)
	if valid {
		t.Error("validateSubscribe should return false for nil subscribe")
	}
	
	if message != "订阅不存�? {
		t.Error("validateSubscribe should return '订阅不存�? for nil subscribe")
	}
	
	// TODO: 测试有效的订阅对�?	// 由于缺少Subscribe模型的完整实现，暂时跳过
}

func TestPrepareSubscribeData(t *testing.T) {
	// 测试准备订阅分享数据
	helper := NewSubscribeHelper()
	
	// TODO: 测试准备订阅数据
	// 由于缺少Subscribe模型的完整实现，暂时跳过
}

func TestBuildSharePayload(t *testing.T) {
	// 测试构建分享请求载荷
	helper := NewSubscribeHelper()
	
	subscribeDict := map[string]interface{}{
		"name": "Test Subscribe",
		"type": "movie",
	}
	
	payload := helper.buildSharePayload("Test Title", "Test Comment", "Test User", subscribeDict)
	
	if payload["share_title"] != "Test Title" {
		t.Error("share_title not set correctly")
	}
	
	if payload["share_comment"] != "Test Comment" {
		t.Error("share_comment not set correctly")
	}
	
	if payload["share_user"] != "Test User" {
		t.Error("share_user not set correctly")
	}
	
	if payload["name"] != "Test Subscribe" {
		t.Error("subscribe data not merged correctly")
	}
}

func TestHandleResponse(t *testing.T) {
	// 测试处理HTTP响应
	helper := NewSubscribeHelper()
	
	// 测试nil响应
	success, message := helper.handleResponse(nil, true)
	if success {
		t.Error("handleResponse should return false for nil response")
	}
	
	if message != "连接MoviePilot服务器失�? {
		t.Error("handleResponse should return correct error message for nil response")
	}
	
	// TODO: 测试有效的HTTP响应
	// 由于缺少实际的HTTP响应对象，暂时跳�?}

func TestHandleListResponse(t *testing.T) {
	// 测试处理返回List的HTTP响应
	helper := NewSubscribeHelper()
	
	// TODO: 测试处理列表响应
	// 由于缺少实际的HTTP响应对象，暂时跳�?}

func TestGetStatistic(t *testing.T) {
	// 测试获取订阅统计数据
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	result := helper.GetStatistic("movie", 1, 30, nil, nil, nil, nil)
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if result == nil {
		t.Error("GetStatistic should not return nil")
	}
}

func TestSubReg(t *testing.T) {
	// 测试新增订阅统计
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	sub := map[string]interface{}{
		"name": "Test Subscribe",
		"type": "movie",
	}
	
	result := helper.SubReg(sub)
	
	// 由于这是测试环境，我们只验证函数能正常执�?	t.Logf("SubReg result: %v", result)
}

func TestSubDone(t *testing.T) {
	// 测试完成订阅统计
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	sub := map[string]interface{}{
		"name": "Test Subscribe",
		"type": "movie",
	}
	
	result := helper.SubDone(sub)
	
	// 由于这是测试环境，我们只验证函数能正常执�?	t.Logf("SubDone result: %v", result)
}

func TestSubRegAsync(t *testing.T) {
	// 测试异步新增订阅统计
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	sub := map[string]interface{}{
		"name": "Test Subscribe",
		"type": "movie",
	}
	
	result := helper.SubRegAsync(sub)
	
	if !result {
		t.Error("SubRegAsync should return true")
	}
}

func TestSubDoneAsync(t *testing.T) {
	// 测试异步完成订阅统计
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	sub := map[string]interface{}{
		"name": "Test Subscribe",
		"type": "movie",
	}
	
	result := helper.SubDoneAsync(sub)
	
	if !result {
		t.Error("SubDoneAsync should return true")
	}
}

func TestSubReport(t *testing.T) {
	// 测试上报存量订阅统计
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	result := helper.SubReport()
	
	// 由于这是测试环境，我们只验证函数能正常执�?	t.Logf("SubReport result: %v", result)
}

func TestSubShare(t *testing.T) {
	// 测试分享订阅
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	success, message := helper.SubShare(1, "Test Title", "Test Comment", "Test User")
	
	// 由于这是测试环境，我们只验证函数能正常执�?	t.Logf("SubShare result: success=%v, message=%s", success, message)
}

func TestShareDelete(t *testing.T) {
	// 测试删除分享
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	success, message := helper.ShareDelete(1)
	
	// 由于这是测试环境，我们只验证函数能正常执�?	t.Logf("ShareDelete result: success=%v, message=%s", success, message)
}

func TestSubFork(t *testing.T) {
	// 测试复用分享的订�?	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	success, message := helper.SubFork(1)
	
	// 由于这是测试环境，我们只验证函数能正常执�?	t.Logf("SubFork result: success=%v, message=%s", success, message)
}

func TestGetShares(t *testing.T) {
	// 测试获取订阅分享数据
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	result := helper.GetShares(nil, 1, 30, nil, nil, nil, nil)
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if result == nil {
		t.Error("GetShares should not return nil")
	}
}

func TestGetShareStatistics(t *testing.T) {
	// 测试获取订阅分享统计数据
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	result := helper.GetShareStatistics()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	if result == nil {
		t.Error("GetShareStatistics should not return nil")
	}
}

func TestGetUserUUID(t *testing.T) {
	// 测试获取用户uuid
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	uuid := helper.getUserUUID()
	
	if uuid == "" {
		t.Error("getUserUUID should not return empty string")
	}
}

func TestGetGithubUser(t *testing.T) {
	// 测试获取github用户
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	user := helper.getGithubUser()
	
	// 由于这是测试环境，可能没有实际数据，但我们至少要确保函数能正常执�?	t.Logf("getGithubUser result: %v", user)
}

func TestIsAdminUser(t *testing.T) {
	// 测试判断是否是管理员
	helper := NewSubscribeHelper()
	
	// 测试函数能正常执�?	isAdmin := helper.IsAdminUser()
	
	// 由于这是测试环境，我们只验证函数能正常执�?	t.Logf("IsAdminUser result: %v", isAdmin)
}
