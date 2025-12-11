package command

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Service 命令服务接口
type Service interface {
	// Register 注册命令处理器
	Register(cmd Handler) error
	// Execute 执行命令
	Execute(ctx context.Context, input string) error
	// List 列出所有命令
	List() []Info
}

// Handler 命令处理器接口
type Handler interface {
	// Name 命令名称（如 "subscribe"）
	Name() string
	// Description 命令描述
	Description() string
	// Execute 执行命令
	Execute(ctx context.Context, args []string) error
}

// Info 命令信息
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// service 命令服务实现
type service struct {
	logger   *zap.Logger
	handlers map[string]Handler
}

// NewService 创建命令服务
func NewService() Service {
	return &service{
		logger:   logger.GetLogger(),
		handlers: make(map[string]Handler),
	}
}

// Register 注册命令处理器
func (s *service) Register(cmd Handler) error {
	name := cmd.Name()
	if _, exists := s.handlers[name]; exists {
		return fmt.Errorf("command already registered: %s", name)
	}
	s.handlers[name] = cmd
	s.logger.Info("Command registered", zap.String("command", name))
	return nil
}

// Execute 执行命令
func (s *service) Execute(ctx context.Context, input string) error {
	// 解析命令
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return fmt.Errorf("empty command input")
	}

	// 获取命令名称（去掉前缀斜杠）
	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	// 查找命令处理器
	handler, exists := s.handlers[cmdName]
	if !exists {
		return fmt.Errorf("command not found: %s", cmdName)
	}

	// 执行命令
	s.logger.Info("Executing command",
		zap.String("command", cmdName),
		zap.Strings("args", args))

	return handler.Execute(ctx, args)
}

// List 列出所有命令
func (s *service) List() []Info {
	infos := make([]Info, 0, len(s.handlers))
	for _, handler := range s.handlers {
		infos = append(infos, Info{
			Name:        handler.Name(),
			Description: handler.Description(),
		})
	}
	return infos
}
