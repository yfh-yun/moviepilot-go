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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/api/middleware"
	"moviepilot-go/internal/api/routes"
	"moviepilot-go/internal/config"
	cacheRedis "moviepilot-go/pkg/cache/redis"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/middlewares"
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
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	zapLogger := logger.GetLogger()
	zapLogger.Info("Starting MoviePilot Go server...",
		zap.String("version", AppVersion),
		zap.String("go_version", runtime.Version()),
		zap.String("go_os", runtime.GOOS),
		zap.String("go_arch", runtime.GOARCH),
	)

	// Load application configuration
	cfgManager, err := config.NewManager(config.Options{})
	if err != nil {
		zapLogger.Fatal("Failed to load configuration", zap.Error(err))
	}
	cfg := cfgManager.Get()

	// Initialize Redis cache client
	if _, err := cacheRedis.Init(cacheRedis.Options{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}); err != nil {
		zapLogger.Fatal("Failed to initialize Redis cache", zap.Error(err))
	}

	// Set Gin mode based on configuration
	if cfg.App.Environment == "production" || !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Gin engine
	engine := gin.New()

	// Add essential middleware
	engine.Use(middlewares.RequestIDMiddleware())
	engine.Use(middlewares.RecoveryMiddleware())
	engine.Use(middlewares.CORSMiddleware())
	engine.Use(middlewares.RateLimitMiddleware())

	// Add request logging middleware
	engine.Use(middleware.RequestLoggingMiddleware())

	// startTime tracks when the server started
	startTime := time.Now()

	// Get server port from configuration
	port := cfg.Server.Port

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   AppVersion,
			"uptime":    time.Since(startTime).String(),
			"port":      port,
		})
	})

	// Register API routes via centralized router
	if err := routes.Register(engine, routes.Config{Logger: zapLogger}); err != nil {
		zapLogger.Fatal("Failed to register routes", zap.Error(err))
	}

	// Create HTTP server with proper configuration
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		zapLogger.Info("Server starting",
			zap.Int("port", port),
			zap.String("env", cfg.App.Environment),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	// The context is used to inform the server it has the configured timeout to finish
	// the request it is currently handling
	shutdownTimeout := cfg.Server.ShutdownTimeout

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited gracefully")
}
