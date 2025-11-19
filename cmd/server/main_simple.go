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
	initConfig()

	// Initialize logger
	logger := initLogger()

	// startTime tracks when the server started
	startTime := time.Now()

	logger.Info("Starting MoviePilot Go server...",
		zap.String("version", AppVersion),
		zap.String("go_version", runtime.Version()),
		zap.String("go_os", runtime.GOOS),
		zap.String("go_arch", runtime.GOARCH),
	)

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

	// Create Gin engine
	engine := gin.New()

	// Add global middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   AppVersion,
			"uptime":    time.Since(startTime).String(),
		})
	})

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
				"time":    time.Now().Unix(),
			})
		})
	}

	// Get server port from configuration
	port := viper.GetString("server.port")
	if port == "" {
		port = DefaultPort
	}

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
		logger.Info("Server starting",
			zap.String("port", port),
			zap.String("env", env),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// The context is used to inform the server it has configured timeout to finish
	// the request it is currently handling
	shutdownTimeout := time.Duration(viper.GetInt("server.shutdown_timeout")) * time.Second
	if shutdownTimeout == 0 {
		shutdownTimeout = time.Duration(DefaultShutdownTimeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully",
		zap.Duration("uptime", time.Since(startTime)),
	)
}

// initConfig initializes the application configuration
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("server.port", DefaultPort)
	viper.SetDefault("server.env", "development")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 60)
	viper.SetDefault("server.shutdown_timeout", DefaultShutdownTimeout)

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults and environment variables
			fmt.Fprintf(os.Stderr, "Config file not found, using defaults and environment variables\n")
		} else {
			fmt.Fprintf(os.Stderr, "Failed to read config file: %v\n", err)
			os.Exit(1)
		}
	}

	// Allow environment variables to override configuration
	viper.AutomaticEnv()
}

// initLogger initializes the zap logger
func initLogger() *zap.Logger {
	var logger *zap.Logger
	var err error

	env := viper.GetString("server.env")
	if env == "production" {
		// Production logger
		logger, err = zap.NewProduction()
	} else {
		// Development logger
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	return logger
}