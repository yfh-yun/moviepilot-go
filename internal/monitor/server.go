package monitor

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server Prometheus 监控服务器
type Server struct {
	addr   string
	server *http.Server
	logger *zap.Logger
}

// Config 监控服务器配置
type Config struct {
	Addr   string // 监听地址，例如 ":9090"
	Logger *zap.Logger
}

// NewServer 创建监控服务器
func NewServer(config Config) *Server {
	if config.Addr == "" {
		config.Addr = ":9090"
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         config.Addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		addr:   config.Addr,
		server: server,
		logger: config.Logger,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	if s.logger != nil {
		s.logger.Info("starting metrics server", zap.String("addr", s.addr))
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	if s.logger != nil {
		s.logger.Info("stopping metrics server")
	}

	return s.server.Shutdown(ctx)
}

// Addr 返回服务器地址
func (s *Server) Addr() string {
	return s.addr
}
