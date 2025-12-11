package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Message 消息
type Message struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Channel   string         `json:"channel"` // 消息渠道: telegram, wechat, etc.
	UserID    string         `json:"user_id"` // 用户ID
	Role      string         `json:"role"`    // 角色: system, plugin, user
	Data      map[string]any `json:"data"`    // 附加数据
	CreatedAt time.Time      `json:"created_at"`
}

// NewMessage 创建消息
func NewMessage(title, content, channel, userID, role string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Title:     title,
		Content:   content,
		Channel:   channel,
		UserID:    userID,
		Role:      role,
		Data:      make(map[string]any),
		CreatedAt: time.Now(),
	}
}

// Operator 消息操作器
// 原Python: MessageOper in app/db/message_oper.py
type Operator struct {
	messages []Message
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewOperator 创建消息操作器
func NewOperator(logger *zap.Logger) *Operator {
	return &Operator{
		messages: make([]Message, 0),
		logger:   logger,
	}
}

// Add 添加消息
func (o *Operator) Add(msg *Message) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.messages = append(o.messages, *msg)

	o.logger.Debug("添加消息",
		zap.String("message_id", msg.ID),
		zap.String("title", msg.Title))

	return nil
}

// Get 获取消息
func (o *Operator) Get(messageID string) (*Message, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	for _, msg := range o.messages {
		if msg.ID == messageID {
			return &msg, nil
		}
	}

	return nil, fmt.Errorf("消息不存在: %s", messageID)
}

// List 获取消息列表
func (o *Operator) List(limit int) []Message {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if limit <= 0 || limit > len(o.messages) {
		limit = len(o.messages)
	}

	// 返回最新的消息
	start := len(o.messages) - limit
	if start < 0 {
		start = 0
	}

	return o.messages[start:]
}

// Delete 删除消息
func (o *Operator) Delete(messageID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	for i, msg := range o.messages {
		if msg.ID == messageID {
			o.messages = append(o.messages[:i], o.messages[i+1:]...)
			o.logger.Debug("删除消息", zap.String("message_id", messageID))
			return nil
		}
	}

	return fmt.Errorf("消息不存在: %s", messageID)
}

// Helper 消息助手
// 原Python: MessageHelper in app/helper/message.py
type Helper struct {
	operator *Operator
	queue    *QueueManager
	logger   *zap.Logger
}

// NewHelper 创建消息助手
func NewHelper(operator *Operator, queue *QueueManager, logger *zap.Logger) *Helper {
	return &Helper{
		operator: operator,
		queue:    queue,
		logger:   logger,
	}
}

// Post 发送消息
// 原Python: put(title, message, role)
func (h *Helper) Post(title, content, channel, userID, role string) error {
	msg := NewMessage(title, content, channel, userID, role)

	// 添加到消息队列
	if err := h.queue.Enqueue(msg); err != nil {
		h.logger.Error("消息入队失败",
			zap.String("title", title),
			zap.Error(err))
		return err
	}

	// 保存到消息记录
	if err := h.operator.Add(msg); err != nil {
		h.logger.Error("保存消息失败",
			zap.String("title", title),
			zap.Error(err))
		return err
	}

	h.logger.Info("发送消息",
		zap.String("title", title),
		zap.String("channel", channel))

	return nil
}

// PostWithData 发送带数据的消息
func (h *Helper) PostWithData(title, content, channel, userID, role string, data map[string]any) error {
	msg := NewMessage(title, content, channel, userID, role)
	msg.Data = data

	if err := h.queue.Enqueue(msg); err != nil {
		return err
	}

	if err := h.operator.Add(msg); err != nil {
		return err
	}

	return nil
}

// QueueManager 消息队列管理器
// 原Python: MessageQueueManager in app/helper/message.py
type QueueManager struct {
	queue    chan *Message
	sendFunc func(*Message) error // 发送回调函数
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewQueueManager 创建消息队列管理器
func NewQueueManager(sendFunc func(*Message) error, logger *zap.Logger) *QueueManager {
	ctx, cancel := context.WithCancel(context.Background())

	qm := &QueueManager{
		queue:    make(chan *Message, 100),
		sendFunc: sendFunc,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	// 启动消息处理goroutine
	qm.start()

	return qm
}

// Enqueue 消息入队
func (qm *QueueManager) Enqueue(msg *Message) error {
	select {
	case qm.queue <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("消息入队超时")
	}
}

// start 启动消息处理
func (qm *QueueManager) start() {
	qm.wg.Add(1)
	go qm.processMessages()
}

// processMessages 处理消息
func (qm *QueueManager) processMessages() {
	defer qm.wg.Done()

	for {
		select {
		case msg := <-qm.queue:
			qm.handleMessage(msg)
		case <-qm.ctx.Done():
			qm.logger.Info("消息队列管理器已停止")
			return
		}
	}
}

// handleMessage 处理单个消息
func (qm *QueueManager) handleMessage(msg *Message) {
	defer func() {
		if r := recover(); r != nil {
			qm.logger.Error("消息处理panic",
				zap.String("message_id", msg.ID),
				zap.Any("panic", r))
		}
	}()

	if qm.sendFunc != nil {
		if err := qm.sendFunc(msg); err != nil {
			qm.logger.Error("发送消息失败",
				zap.String("message_id", msg.ID),
				zap.String("title", msg.Title),
				zap.Error(err))
		} else {
			qm.logger.Debug("消息已发送",
				zap.String("message_id", msg.ID),
				zap.String("title", msg.Title))
		}
	}
}

// Stop 停止消息队列管理器
func (qm *QueueManager) Stop() {
	qm.cancel()
	qm.wg.Wait()
	close(qm.queue)
	qm.logger.Info("消息队列管理器已关闭")
}

// GetQueueSize 获取队列大小
func (qm *QueueManager) GetQueueSize() int {
	return len(qm.queue)
}
