package history

import (
	"github.com/yfh-yun/moviepilot-go/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterHistoryRoutes 注册历史记录API路由
func RegisterHistoryRoutes(
	router *gin.Engine,
	historyService service.HistoryService,
	logger *zap.Logger,
) {
	handler := NewHandler(historyService, logger)
	
	// 注册路由
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)
}

// GetRouter 获取路由组（用于手动配置）
func GetRouter(
	historyService service.HistoryService,
	logger *zap.Logger,
) *Handler {
	return NewHandler(historyService, logger)
}