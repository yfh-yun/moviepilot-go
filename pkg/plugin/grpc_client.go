package plugin

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"moviepilot-go/internal/proto/common"
	"moviepilot-go/internal/proto/plugin"
	"moviepilot-go/pkg/logger"
)

// GRPCClient gRPC客户端，连接Python插件服务
type GRPCClient struct {
	client      plugin.PluginServiceClient
	conn        *grpc.ClientConn
	address     string
	port        int
	logger      *zap.Logger
	isConnected bool
}

// NewGRPCClient 创建gRPC客户端
func NewGRPCClient(address string, port int) *GRPCClient {
	return &GRPCClient{
		address:     address,
		port:        port,
		logger:      logger.GetLogger(),
		isConnected: false,
	}
}

// Connect 连接到Python插件服务
func (c *GRPCClient) Connect() error {
	if c.isConnected {
		return nil
	}

	// 设置gRPC连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(1*1024*1024),  // 1MB
		),
	}

	// 连接到gRPC服务器
	addr := fmt.Sprintf("%s:%d", c.address, c.port)
	c.logger.Info("正在连接Python插件服务", zap.String("address", addr))

	// 使用grpc.NewClient替代grpc.Dial，因为grpc.Dial已被弃用
	clientConn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		c.logger.Error("连接Python插件服务失败", zap.Error(err))
		return err
	}

	// 获取底层的ClientConn
	conn := clientConn

	c.conn = conn
	c.client = plugin.NewPluginServiceClient(conn)
	c.isConnected = true
	c.logger.Info("已连接到Python插件服务", zap.String("address", addr))

	return nil
}

// Disconnect 断开与Python插件服务的连接
func (c *GRPCClient) Disconnect() error {
	if !c.isConnected {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		c.logger.Error("断开Python插件服务连接失败", zap.Error(err))
		return err
	}

	c.isConnected = false
	c.logger.Info("已断开与Python插件服务的连接")
	return nil
}

// LoadPlugin 加载插件
func (c *GRPCClient) LoadPlugin(ctx context.Context, pluginID, pluginPath string, config map[string]any) (*plugin.PluginInfo, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	// 将map[string]any转换为map[string]string
	configStr := make(map[string]string)
	for k, v := range config {
		configStr[k] = fmt.Sprintf("%v", v)
	}

	req := &plugin.LoadPluginRequest{
		PluginId:   pluginID,
		PluginPath: pluginPath,
		Config:     configStr,
	}

	resp, err := c.client.LoadPlugin(ctx, req)
	if err != nil {
		c.logger.Error("加载插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		return nil, err
	}

	return resp.PluginInfo, nil
}

// UnloadPlugin 卸载插件
func (c *GRPCClient) UnloadPlugin(ctx context.Context, pluginID string) error {
	if err := c.Connect(); err != nil {
		return err
	}

	req := &plugin.UnloadPluginRequest{
		PluginId: pluginID,
	}

	_, err := c.client.UnloadPlugin(ctx, req)
	if err != nil {
		c.logger.Error("卸载插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		return err
	}

	return nil
}

// StartPlugin 启动插件
func (c *GRPCClient) StartPlugin(ctx context.Context, pluginID string, params map[string]any) error {
	if err := c.Connect(); err != nil {
		return err
	}

	// 将map[string]any转换为map[string]string
	paramsStr := make(map[string]string)
	for k, v := range params {
		paramsStr[k] = fmt.Sprintf("%v", v)
	}

	req := &plugin.StartPluginRequest{
		PluginId: pluginID,
		Params:   paramsStr,
	}

	_, err := c.client.StartPlugin(ctx, req)
	if err != nil {
		c.logger.Error("启动插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		return err
	}

	return nil
}

// StopPlugin 停止插件
func (c *GRPCClient) StopPlugin(ctx context.Context, pluginID string, force bool) error {
	if err := c.Connect(); err != nil {
		return err
	}

	req := &plugin.StopPluginRequest{
		PluginId: pluginID,
		Force:    force,
	}

	_, err := c.client.StopPlugin(ctx, req)
	if err != nil {
		c.logger.Error("停止插件失败", zap.String("plugin_id", pluginID), zap.Error(err))
		return err
	}

	return nil
}

// GetPluginInfo 获取插件信息
func (c *GRPCClient) GetPluginInfo(ctx context.Context, pluginID string) (*plugin.PluginInfo, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	req := &plugin.GetPluginInfoRequest{
		PluginId: pluginID,
	}

	resp, err := c.client.GetPluginInfo(ctx, req)
	if err != nil {
		c.logger.Error("获取插件信息失败", zap.String("plugin_id", pluginID), zap.Error(err))
		return nil, err
	}

	return resp.PluginInfo, nil
}

// ListPlugins 列出所有插件
func (c *GRPCClient) ListPlugins(ctx context.Context, pluginType plugin.PluginType, status plugin.PluginStatus) ([]*plugin.PluginInfo, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	req := &plugin.ListPluginsRequest{
		Type:   pluginType,
		Status: status,
		Page: &common.PageRequest{
			Page:     1,
			PageSize: 1000, // 获取所有插件
		},
	}

	resp, err := c.client.ListPlugins(ctx, req)
	if err != nil {
		c.logger.Error("列出插件失败", zap.Error(err))
		return nil, err
	}

	return resp.Plugins, nil
}

// ExecutePlugin 执行插件方法
func (c *GRPCClient) ExecutePlugin(ctx context.Context, pluginID, method string, params []byte, timeout int) ([]byte, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	req := &plugin.ExecutePluginRequest{
		PluginId: pluginID,
		Method:   method,
		Params:   params,
		Timeout:  int32(timeout),
	}

	resp, err := c.client.ExecutePlugin(ctx, req)
	if err != nil {
		c.logger.Error("执行插件方法失败", zap.String("plugin_id", pluginID), zap.String("method", method), zap.Error(err))
		return nil, err
	}

	return resp.Result, nil
}

// GetPluginConfig 获取插件配置
func (c *GRPCClient) GetPluginConfig(ctx context.Context, pluginID string) ([]*common.ConfigItem, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	req := &plugin.GetPluginConfigRequest{
		PluginId: pluginID,
	}

	resp, err := c.client.GetPluginConfig(ctx, req)
	if err != nil {
		c.logger.Error("获取插件配置失败", zap.String("plugin_id", pluginID), zap.Error(err))
		return nil, err
	}

	return resp.Config, nil
}

// UpdatePluginConfig 更新插件配置
func (c *GRPCClient) UpdatePluginConfig(ctx context.Context, pluginID string, config []*common.ConfigItem) error {
	if err := c.Connect(); err != nil {
		return err
	}

	req := &plugin.UpdatePluginConfigRequest{
		PluginId: pluginID,
		Config:   config,
	}

	_, err := c.client.UpdatePluginConfig(ctx, req)
	if err != nil {
		c.logger.Error("更新插件配置失败", zap.String("plugin_id", pluginID), zap.Error(err))
		return err
	}

	return nil
}

// HealthCheck 健康检查
func (c *GRPCClient) HealthCheck(ctx context.Context) error {
	if err := c.Connect(); err != nil {
		return err
	}

	req := &common.HealthCheckRequest{
		Service: "plugin_service",
	}

	_, err := c.client.HealthCheck(ctx, req)
	if err != nil {
		c.logger.Error("健康检查失败", zap.Error(err))
		return err
	}

	return nil
}

// IsConnected 检查是否已连接
func (c *GRPCClient) IsConnected() bool {
	return c.isConnected
}


