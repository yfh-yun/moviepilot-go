// Package chain 业务逻辑链基础结构
package chain

import (
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/cache"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/plugin"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/template"

	"go.uber.org/zap"
)

// ChainBase 处理链基类
// 对应Python项目中的ChainBase，提供所有业务链的公共功能
type ChainBase struct {
	// 核心服务
	moduleManager service.ModuleManager
	eventManager  service.EventManager
	
	// 消息相关
	messageService service.MessageService
	messageHelper  *template.MessageTemplateHelper
	
	// 插件管理
	pluginManager plugin.Manager
	
	// 缓存系统
	fileCache     cache.FileCache
	asyncFileCache cache.AsyncFileCache
	
	// 日志记录器
	logger *zap.Logger
}

// NewChainBase 创建链基类实例
func NewChainBase(
	moduleManager service.ModuleManager,
	eventManager service.EventManager,
	messageService service.MessageService,
	messageHelper *template.MessageTemplateTemplateHelper,
	pluginManager plugin.Manager,
	fileCache cache.FileCache,
	asyncFileCache cache.AsyncFileCache,
	logger *zap.Logger,
) *ChainBase {
	return &ChainBase{
		moduleManager:   moduleManager,
		eventManager:    eventManager,
		messageService:  messageService,
		messageHelper:   messageHelper,
		pluginManager:   pluginManager,
		fileCache:       fileCache,
		asyncFileCache:  asyncFileCache,
		logger:          logger,
	}
}

// LoadCache 加载缓存
func (cb *ChainBase) LoadCache(ctx context.Context, filename string) (interface{}, error) {
	content, err := cb.fileCache.Get(ctx, filename)
	if err != nil {
		cb.logger.Error("加载缓存失败", 
			zap.String("filename", filename),
			zap.Error(err))
		return nil, fmt.Errorf("加载缓存失败: %w", err)
	}

	if content == nil {
		return nil, nil
	}

	// 反序列化
	var result interface{}
	decoder := gob.NewDecoder(content)
	if err := decoder.Decode(&result); err != nil {
		cb.logger.Error("反序列化缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
		return nil, fmt.Errorf("反序列化缓存失败: %w", err)
	}

	return result, nil
}

// AsyncLoadCache 异步加载缓存
func (cb *ChainBase) AsyncLoadCache(ctx context.Context, filename string) (interface{}, error) {
	content, err := cb.asyncFileCache.Get(ctx, filename)
	if err != nil {
		cb.logger.Error("异步加载缓存失败", 
			zap.String("filename", filename),
			zap.Error(err))
		return nil, fmt.Errorf("异步加载缓存失败: %w", err)
	}

	if content == nil {
		return nil, nil
	}

	// 反序列化
	var result interface{}
	decoder := gob.NewDecoder(content)
	if err := decoder.Decode(&result); err != nil {
		cb.logger.Error("异步反序列化缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
		return nil, fmt.Errorf("异步反序列化缓存失败: %w", err)
	}

	return result, nil
}

// AsyncSaveCache 异步保存缓存
func (cb *ChainBase) AsyncSaveCache(ctx context.Context, cacheData interface{}, filename string) error {
	// 序列化
	var buffer []byte
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(cacheData); err != nil {
		cb.logger.Error("序列化缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	// 保存到缓存
	if err := cb.asyncFileCache.Set(ctx, filename, buffer); err != nil {
		cb.logger.Error("异步保存缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
		return fmt.Errorf("异步保存缓存失败: %w", err)
	}

	cb.logger.Debug("异步保存缓存成功", zap.String("filename", filename))
	return nil
}

// SaveCache 保存缓存
func (cb *ChainBase) SaveCache(ctx context.Context, cacheData interface{}, filename string) error {
	// 序列化
	var buffer []byte
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(cacheData); err != nil {
		cb.logger.Error("序列化缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	// 保存到缓存
	if err := cb.fileCache.Set(ctx, filename, buffer); err != nil {
		cb.logger.Error("保存缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
		return fmt.Errorf("保存缓存失败: %w", err)
	}

	cb.logger.Debug("保存缓存成功", zap.String("filename", filename))
	return nil
}

// RemoveCache 删除缓存
func (cb *ChainBase) RemoveCache(ctx context.Context, filename string) error {
	// 删除同步缓存
	if err := cb.fileCache.Delete(ctx, filename); err != nil {
		cb.logger.Error("删除同步缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
	}

	// 删除异步缓存
	if err := cb.asyncFileCache.Delete(ctx, filename); err != nil {
		cb.logger.Error("删除异步缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
	}

	cb.logger.Debug("删除缓存成功", zap.String("filename", filename))
	return nil
}

// AsyncRemoveCache 异步删除缓存
func (cb *ChainBase) AsyncRemoveCache(ctx context.Context, filename string) error {
	if err := cb.asyncFileCache.Delete(ctx, filename); err != nil {
		cb.logger.Error("异步删除缓存失败",
			zap.String("filename", filename),
			zap.Error(err))
		return fmt.Errorf("异步删除缓存失败: %w", err)
	}

	cb.logger.Debug("异步删除缓存成功", zap.String("filename", filename))
	return nil
}

// ExecutePluginModules 执行插件模块
func (cb *ChainBase) ExecutePluginModules(ctx context.Context, method string, result interface{}, args ...interface{}) (interface{}, error) {
	pluginModules := cb.pluginManager.GetPluginModules()
	
	for pluginID, moduleDict := range pluginModules {
		if moduleMethod, exists := moduleDict[method]; exists {
			cb.logger.Info("执行插件模块",
				zap.String("plugin_id", pluginID),
				zap.String("method", method))

			// 检查结果是否为空
			if cb.isValidEmpty(result) {
				// 返回None，第一次执行或者需继续执行下一模块
				newResult, err := moduleMethod(ctx, args...)
				if err != nil {
					cb.handlePluginError(err, pluginID, method)
					continue
				}
				result = newResult
			} else if resultSlice, ok := result.([]interface{}); ok {
				// 返回为列表，有多个模块运行结果时进行合并
				tempResult, err := moduleMethod(ctx, args...)
				if err != nil {
					cb.handlePluginError(err, pluginID, method)
					continue
				}
				
				if tempSlice, ok := tempResult.([]interface{}); ok {
					resultSlice = append(resultSlice, tempSlice...)
					result = resultSlice
				}
			} else {
				// 有结果了，停止执行
				break
			}
		}
	}

	return result, nil
}

// AsyncExecutePluginModules 异步执行插件模块
func (cb *ChainBase) AsyncExecutePluginModules(ctx context.Context, method string, result interface{}, args ...interface{}) (interface{}, error) {
	pluginModules := cb.pluginManager.GetPluginModules()
	
	for pluginID, moduleDict := range pluginModules {
		if moduleMethod, exists := moduleDict[method]; exists {
			cb.logger.Info("异步执行插件模块",
				zap.String("plugin_id", pluginID),
				zap.String("method", method))

			// 检查结果是否为空
			if cb.isValidEmpty(result) {
				// 返回None，第一次执行或者需继续执行下一模块
				newResult, err := moduleMethod(ctx, args...)
				if err != nil {
					cb.handlePluginError(err, pluginID, method)
					continue
				}
				result = newResult
			} else if resultSlice, ok := result.([]interface{}); ok {
				// 返回为列表，有多个模块运行结果时进行合并
				tempResult, err := moduleMethod(ctx, args...)
				if err != nil {
					cb.handlePluginError(err, pluginID, method)
					continue
				}
				
				if tempSlice, ok := tempResult.([]interface{}); ok {
					resultSlice = append(resultSlice, tempSlice...)
					result = resultSlice
				}
			} else {
				// 有结果了，停止执行
				break
			}
		}
	}

	return result, nil
}

// ExecuteSystemModules 执行系统模块
func (cb *ChainBase) ExecuteSystemModules(ctx context.Context, moduleID, method string, result interface{}, args ...interface{}) (interface{}, error) {
	systemModules := cb.moduleManager.GetSystemModules()
	
	if moduleMethod, exists := systemModules[moduleID][method]; exists {
		cb.logger.Info("执行系统模块",
			zap.String("module_id", moduleID),
			zap.String("method", method))

		// 检查结果是否为空
		if cb.isValidEmpty(result) {
			// 返回None，第一次执行或者需继续执行下一模块
			newResult, err := moduleMethod(ctx, args...)
			if err != nil {
				cb.handleSystemError(err, moduleID, method)
				return nil, err
			}
			result = newResult
		} else if resultSlice, ok := result.([]interface{}); ok {
			// 返回为列表，有多个模块运行结果时进行合并
			tempResult, err := moduleMethod(ctx, args...)
			if err != nil {
				cb.handleSystemError(err, moduleID, method)
				return nil, err
			}
			
			if tempSlice, ok := tempResult.([]interface{}); ok {
				resultSlice = append(resultSlice, tempSlice...)
				result = resultSlice
			}
		} else {
			// 有结果了，停止执行
		}
	}

	return result, nil
}

// isValidEmpty 判断结果是否为空
func (cb *ChainBase) isValidEmpty(result interface{}) bool {
	if result == nil {
		return true
	}

	if tuple, ok := result.([]interface{}); ok {
		for _, value := range tuple {
			if value != nil {
				return false
			}
		}
		return true
	}

	return false
}

// handlePluginError 处理插件模块执行错误
func (cb *ChainBase) handlePluginError(err error, pluginID, method string) {
	cb.logger.Error("运行插件模块出错",
		zap.String("plugin_id", pluginID),
		zap.String("method", method),
		zap.Error(err))

	// 发送错误消息
	errorMsg := fmt.Sprintf("插件 %s 的 %s 方法执行失败: %s", pluginID, method, err.Error())
	
	// 发送通知到用户
	if cb.messageHelper != nil {
		cb.messageHelper.Put(ctx, &template.Message{
			Title:   fmt.Sprintf("%s 发生了错误", pluginID),
			Message: errorMsg,
			Role:    "plugin",
		})
	}

	// 发送事件
	if cb.eventManager != nil {
		cb.eventManager.SendEvent(ctx, &service.Event{
			Type: "SystemError",
			Data: map[string]interface{}{
				"type":          "plugin",
				"plugin_id":     pluginID,
				"plugin_method": method,
				"error":         err.Error(),
			},
		})
	}
}

// handleSystemError 处理系统模块执行错误
func (cb *ChainBase) handleSystemError(err error, moduleID, method string) {
	cb.logger.Error("运行系统模块出错",
		zap.String("module_id", moduleID),
		zap.String("method", method),
		zap.Error(err))

	// 发送错误消息
	errorMsg := fmt.Sprintf("系统模块 %s 的 %s 方法执行失败: %s", moduleID, method, err.Error())
	
	// 发送通知到用户
	if cb.messageHelper != nil {
		cb.messageHelper.Put(ctx, &template.Message{
			Title:   fmt.Sprintf("%s 发生了错误", moduleID),
			Message: errorMsg,
			Role:    "system",
		})
	}

	// 发送事件
	if cb.eventManager != nil {
		cb.eventManager.SendEvent(ctx, &service.Event{
			Type: "SystemError",
			Data: map[string]interface{}{
				"type":           "module",
				"module_id":      moduleID,
				"module_method":  method,
				"error":          err.Error(),
			},
		})
	}
}