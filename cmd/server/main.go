// Package main MoviePilot Go application entry point
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/infrastructure/bootstrap"
	"moviepilot-go/pkg/logger"
)

const (
	// AppVersion application version
	AppVersion = "2.8.1"
	// DefaultPort default server port
	DefaultPort = "3001"
	// DefaultShutdownTimeout default shutdown timeout in seconds
	DefaultShutdownTimeout = 30
)

// @title MoviePilot API
// @version 2.8.1
// @description MoviePilot Go version - Automated media library management tool
// @termsOfService http://swagger.io/terms/

// @contact.name MoviePilot Team
// @contact.url http://www.moviepilot.com
// @contact.email support@moviepilot.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3001
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 使用bootstrap包统一初始化所有组件
	app, err := bootstrap.Bootstrap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bootstrap application: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	app.Logger.Info("Starting MoviePilot Go server...",
		zap.String("version", AppVersion),
		zap.String("go_version", runtime.Version()),
		zap.String("go_os", runtime.GOOS),
		zap.String("go_arch", runtime.GOARCH),
	)

	// Create HTTP server with proper configuration
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.Config.App.Host, app.Config.App.Port),
		Handler:      app.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		app.Logger.Info("Server starting",
			zap.Int("port", app.Config.App.Port),
			zap.Bool("debug", app.Config.App.Debug),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// The context is used to inform the server it has the configured timeout to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// 关闭所有组件
	if err := app.Shutdown(ctx); err != nil {
		app.Logger.Fatal("Failed to shutdown application gracefully", zap.Error(err))
	}
}
