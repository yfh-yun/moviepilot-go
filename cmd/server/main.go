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
	"github.com/spf13/viper"
	"go.uber.org/zap"
	
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers"
	"github.com/yfh-yun/moviepilot-go/internal/apis/routes"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/config"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/apis/middlewares"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/migrations"
	"github.com/yfh-yun/moviepilot-go/internal/schedulers"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/message"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/plugin"
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
	// Initialize configuration
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

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

	// Initialize database
	if err := database.Init(); err != nil {
		zapLogger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Run database migrations
	if err := runMigrations(); err != nil {
		zapLogger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize cache
	if err := cache.Init(); err != nil {
		zapLogger.Fatal("Failed to initialize cache", zap.Error(err))
	}

	// Set Gin mode based on configuration
	env := viper.GetString("server.env")
	if env == "" {
		env = "development"
	}
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create base handler
	baseHandler := handlers.NewBaseHandler()

	// Create router configuration
	routerConfig := &routes.RouterConfig{
		BaseHandler: baseHandler,
	}

	// Setup routes
	engine := routes.SetupRouter(routerConfig)

	// Add global middleware
	engine.Use(middleware.RequestIDMiddleware())
	engine.Use(middleware.LoggerMiddleware())
	engine.Use(middleware.RecoveryMiddleware())
	engine.Use(middleware.CORSMiddleware())

	// Check if scheduler is enabled
	if viper.GetBool("scheduler.enabled") {
		// Create dependency services
		db := database.GetDB()

		// Create workflow repository
		workflowRepo := repositories.NewWorkflowRepository(db)

		// Create message service dependencies
		messageRepo := repositories.NewMessageRepository(db)
		userRepo := repositories.NewUserRepository(db)

		// Create plugin service dependencies
		pluginRepo := repositories.NewPluginRepository(db)

		// Create message service instance
		messageService := message.NewMessageService(messageRepo, userRepo, zapLogger)

		// Create plugin service instance
		basePath := viper.GetString("app.base_path")
		if basePath == "" {
			basePath = "./"
		}
		pluginService := plugin.NewPluginService(pluginRepo, zapLogger, basePath)

		zapLogger.Info("Creating scheduler dependency services...")

		// Initialize scheduler service
		schedulerService, err := scheduler.NewSchedulerService(
			zapLogger,
			workflowRepo,
			nil, // subscribeService - TODO: implement
			nil, // downloadService - TODO: implement
			messageService,
			pluginService,
		)

		if err != nil {
			zapLogger.Error("Failed to initialize scheduler service", zap.Error(err))
		} else {
			// Start scheduler service
			if err := schedulerService.Start(); err != nil {
				zapLogger.Error("Failed to start scheduler service", zap.Error(err))
			} else {
				zapLogger.Info("Scheduler service started successfully")

				// Stop scheduler service when server shuts down
				defer schedulerService.Stop()
			}
		}
	} else {
		zapLogger.Info("Scheduler service is disabled in configuration")
	}

// startTime tracks when the server started
	startTime := time.Now()

	// Get server port from configuration
	port := viper.GetString("server.port")
	if port == "" {
		port = DefaultPort
	}

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

	// Create HTTP server with proper configuration
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      engine,
		ReadTimeout:  time.Duration(viper.GetInt("server.read_timeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("server.write_timeout")) * time.Second,
		IdleTimeout:  time.Duration(viper.GetInt("server.idle_timeout")) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		zapLogger.Info("Server starting", 
			zap.String("port", port),
			zap.String("env", env),
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
	shutdownTimeout := time.Duration(viper.GetInt("server.shutdown_timeout")) * time.Second
	if shutdownTimeout == 0 {
		shutdownTimeout = time.Duration(DefaultShutdownTimeout) * time.Second
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited gracefully")
}

// runMigrations runs database migrations
func runMigrations() error {
	zapLogger := logger.GetLogger()
	zapLogger.Info("Starting database migrations")
	
	// Get database instance
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Run migrations
	migration := migrations.NewMigration(db)
	if err := migration.Run(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	zapLogger.Info("Database migrations completed successfully")
	return nil
}
