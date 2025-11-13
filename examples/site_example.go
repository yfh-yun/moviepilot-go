package main

import (
	"fmt"

	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== 站点工具示例 ===")

	// 创建站点工具类实�?	siteUtils := utils.NewSiteUtils()

	// 测试登录状态检�?	fmt.Println("\n--- 登录状态检�?---")
	testLoginCheck(siteUtils)

	// 测试签到状态检�?	fmt.Println("\n--- 签到状态检�?---")
	testCheckinCheck(siteUtils)
}

func testLoginCheck(siteUtils *utils.SiteUtils) {
	// 模拟未登录的HTML（包含密码输入框�?	unloggedInHTML := `
	<html>
	<head><title>Login</title></head>
	<body>
		<form>
			<input type="text" name="username" />
			<input type="password" name="password" />
			<input type="submit" value="Login" />
		</form>
	</body>
	</html>
	`

	// 模拟已登录的HTML（包含登出链接）
	loggedInHTML := `
	<html>
	<head><title>Dashboard</title></head>
	<body>
		<div class="user-info-side">
			Welcome, User!
		</div>
		<a href="/logout">Logout</a>
		<div>Content...</div>
	</body>
	</html>
	`

	// 模拟已登录的HTML（包含用户控制面板）
	userCPHTML := `
	<html>
	<head><title>User CP</title></head>
	<body>
		<a href="/usercp">User Control Panel</a>
		<a id="myitem" href="/myitems">My Items</a>
		<div>Content...</div>
	</body>
	</html>
	`

	// 测试未登录页�?	isLoggedIn := siteUtils.IsLoggedIn(unloggedInHTML)
	fmt.Printf("未登录页面检查结�? %v\n", isLoggedIn) // 应该�?false

	// 测试已登录页面（登出链接�?	isLoggedIn = siteUtils.IsLoggedIn(loggedInHTML)
	fmt.Printf("已登录页面检查结果（登出链接�? %v\n", isLoggedIn) // 应该�?true

	// 测试已登录页面（用户面板�?	isLoggedIn = siteUtils.IsLoggedIn(userCPHTML)
	fmt.Printf("已登录页面检查结果（用户面板�? %v\n", isLoggedIn) // 应该�?true

	// 测试空HTML
	isLoggedIn = siteUtils.IsLoggedIn("")
	fmt.Printf("空HTML检查结�? %v\n", isLoggedIn) // 应该�?false
}

func testCheckinCheck(siteUtils *utils.SiteUtils) {
	// 模拟未签到的HTML（包含签到链接）
	uncheckedHTML := `
	<html>
	<head><title>Dashboard</title></head>
	<body>
		<a href="/attendance.php">签到</a>
		<div>Content...</div>
	</body>
	</html>
	`

	// 模拟已签到的HTML（不包含签到链接�?	checkedHTML := `
	<html>
	<head><title>Dashboard</title></head>
	<body>
		<div>Welcome back!</div>
		<div>Content...</div>
	</body>
	</html>
	`

	// 模拟未签到的HTML（包含打卡按钮）
	uncheckedButtonHTML := `
	<html>
	<head><title>Dashboard</title></head>
	<body>
		<input class="dt_button" type="button" value="打卡" onclick="do_signin()" />
		<div>Content...</div>
	</body>
	</html>
	`

	// 模拟未签到的HTML（包含特定ID的签到按钮）
	uncheckedIDHTML := `
	<html>
	<head><title>Dashboard</title></head>
	<body>
		<a id="do-attendance" href="/attendance.php">每日签到</a>
		<div>Content...</div>
	</body>
	</html>
	`

	// 测试未签到页�?	isChecked := siteUtils.IsCheckin(uncheckedHTML)
	fmt.Printf("未签到页面检查结�? %v\n", isChecked) // 应该�?false

	// 测试已签到页�?	isChecked = siteUtils.IsCheckin(checkedHTML)
	fmt.Printf("已签到页面检查结�? %v\n", isChecked) // 应该�?true

	// 测试未签到页面（打卡按钮�?	isChecked = siteUtils.IsCheckin(uncheckedButtonHTML)
	fmt.Printf("未签到页面检查结果（打卡按钮�? %v\n", isChecked) // 应该�?false

	// 测试未签到页面（特定ID�?	isChecked = siteUtils.IsCheckin(uncheckedIDHTML)
	fmt.Printf("未签到页面检查结果（特定ID�? %v\n", isChecked) // 应该�?false

	// 测试空HTML
	isChecked = siteUtils.IsCheckin("")
	fmt.Printf("空HTML检查结�? %v\n", isChecked) // 应该�?false
}
