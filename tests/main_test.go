package tests

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestMain 测试主入口
func TestMain(m *testing.M) {
	// 设置测试环境
	setupTestEnvironment()

	// 运行测试
	code := m.Run()

	// 清理测试环境
	cleanupTestEnvironment()

	os.Exit(code)
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() {
	log.Println("设置测试环境...")
	
	// 设置环境变量
	os.Setenv("GIN_MODE", "test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_NAME", "moviepilot_test")
	os.Setenv("LOG_LEVEL", "debug")
	
	// 其他初始化逻辑...
}

// cleanupTestEnvironment 清理测试环境
func cleanupTestEnvironment() {
	log.Println("清理测试环境...")
	// 清理逻辑...
}

// BaseTestSuite 基础测试套件
type BaseTestSuite struct {
	suite.Suite
}

// SetupSuite 测试套件设置
func (suite *BaseTestSuite) SetupSuite() {
	// 套件级别的初始化
}

// TearDownSuite 测试套件清理
func (suite *BaseTestSuite) TearDownSuite() {
	// 套件级别的清理
}

// SetupTest 每个测试前的设置
func (suite *BaseTestSuite) SetupTest() {
	// 每个测试前的初始化
}

// TearDownTest 每个测试后的清理
func (suite *BaseTestSuite) TearDownTest() {
	// 每个测试后的清理
}