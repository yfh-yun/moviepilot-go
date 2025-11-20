package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/template"

	"go.uber.org/zap"
)

// MessageChain 消息处理链
// 负责用户消息的接收、处理、分发和响应

type MessageChain struct {
	messageRepo    repository.MessageRepository
	userRepo       repository.UserRepository
	mediaRepo      repository.MediaRepository
	pluginMgr      PluginManager
	templateHelper *template.MessageTemplateHelper
	logger         *zap.Logger
}

// NewMessageChain 创建消息处理链实例
func NewMessageChain(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	mediaRepo repository.MediaRepository,
	pluginMgr PluginManager,
	templateHelper *template.MessageTemplateHelper,
) *MessageChain {
	return &MessageChain{
		messageRepo:    messageRepo,
		userRepo:       userRepo,
		mediaRepo:      mediaRepo,
		pluginMgr:      pluginMgr,
		templateHelper: templateHelper,
		logger:         logger.GetLogger().With(zap.String("module", "chain.message")),
	}
}

// Execute 执行消息处理链
func (mc *MessageChain) Execute(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	mc.logger.Info("开始执行消息处理链",
		zap.String("user_id", req.UserID),
		zap.String("message_type", req.MessageType))

	// 验证请求参数
	if err := mc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("消息请求验证失败: %w", err)
	}

	// 1. 消息预处理
	processedMessage, err := mc.preprocessMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("消息预处理失败: %w", err)
	}

	// 2. 意图识别
	intent, err := mc.recognizeIntent(ctx, processedMessage)
	if err != nil {
		return nil, fmt.Errorf("意图识别失败: %w", err)
	}

	// 3. 内容解析
	parsedContent, err := mc.parseContent(ctx, processedMessage, intent)
	if err != nil {
		return nil, fmt.Errorf("内容解析失败: %w", err)
	}

	// 4. 业务逻辑处理
	businessResult, err := mc.processBusinessLogic(ctx, parsedContent, intent)
	if err != nil {
		return nil, fmt.Errorf("业务逻辑处理失败: %w", err)
	}

	// 5. 响应生成
	response, err := mc.generateResponse(ctx, businessResult, intent)
	if err != nil {
		return nil, fmt.Errorf("响应生成失败: %w", err)
	}

	// 6. 消息发送
	if err := mc.sendResponse(ctx, response); err != nil {
		return nil, fmt.Errorf("消息发送失败: %w", err)
	}

	// 7. 记录处理结果
	if err := mc.recordProcessingResult(ctx, req, response); err != nil {
		mc.logger.Warn("记录处理结果失败", zap.Error(err))
	}

	mc.logger.Info("消息处理链执行完成",
		zap.String("user_id", req.UserID),
		zap.String("intent", intent.Type),
		zap.Bool("success", response.Success))

	return response, nil
}

// preprocessMessage 消息预处理
func (mc *MessageChain) preprocessMessage(ctx context.Context, req *MessageRequest) (*ProcessedMessage, error) {
	// 文本清洗和标准化
	cleanedText := mc.cleanText(req.Content)

	// 语言检测
	language := mc.detectLanguage(cleanedText)

	// 敏感信息过滤
	safeContent := mc.filterSensitiveContent(cleanedText)

	// 获取用户上下文
	userContext, err := mc.getUserContext(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	processed := &ProcessedMessage{
		OriginalContent: req.Content,
		CleanedContent:  safeContent,
		Language:        language,
		UserContext:     userContext,
		MessageType:     req.MessageType,
		Timestamp:       time.Now(),
	}

	return processed, nil
}

// recognizeIntent 意图识别
func (mc *MessageChain) recognizeIntent(ctx context.Context, message *ProcessedMessage) (*MessageIntent, error) {
	// 基于规则匹配
	ruleBasedIntent := mc.ruleBasedIntentRecognition(message)
	if ruleBasedIntent != nil {
		return ruleBasedIntent, nil
	}

	// 基于机器学习
	mlIntent, err := mc.machineLearningIntentRecognition(ctx, message)
	if err != nil {
		return nil, err
	}

	// 插件意图识别
	pluginIntent, err := mc.pluginIntentRecognition(ctx, message)
	if err != nil {
		mc.logger.Warn("插件意图识别失败", zap.Error(err))
	}

	// 意图融合
	finalIntent := mc.intentFusion(ruleBasedIntent, mlIntent, pluginIntent)

	return finalIntent, nil
}

// parseContent 内容解析
func (mc *MessageChain) parseContent(ctx context.Context, message *ProcessedMessage, intent *MessageIntent) (*ParsedContent, error) {
	var parsed ParsedContent

	switch intent.Type {
	case IntentTypeSearch:
		// 解析搜索关键词
		searchQuery := mc.parseSearchQuery(message.CleanedContent)
		parsed.SearchQuery = searchQuery

	case IntentTypeRecommend:
		// 解析推荐偏好
		preferences := mc.parseRecommendationPreferences(message.CleanedContent)
		parsed.Preferences = preferences

	case IntentTypeDownload:
		// 解析下载请求
		downloadInfo := mc.parseDownloadRequest(message.CleanedContent)
		parsed.DownloadInfo = downloadInfo

	case IntentTypeSubscribe:
		// 解析订阅信息
		subscribeInfo := mc.parseSubscriptionRequest(message.CleanedContent)
		parsed.SubscribeInfo = subscribeInfo

	case IntentTypeQuery:
		// 解析查询条件
		queryInfo := mc.parseQueryConditions(message.CleanedContent)
		parsed.QueryInfo = queryInfo

	default:
		// 通用解析
		entities := mc.extractEntities(message.CleanedContent)
		parsed.Entities = entities
	}

	// 解析情感倾向
	parsed.Sentiment = mc.analyzeSentiment(message.CleanedContent)

	// 解析时效性
	parsed.Urgency = mc.analyzeUrgency(message.CleanedContent)

	return &parsed, nil
}

// processBusinessLogic 业务逻辑处理
func (mc *MessageChain) processBusinessLogic(ctx context.Context, content *ParsedContent, intent *MessageIntent) (*BusinessResult, error) {
	var result BusinessResult

	switch intent.Type {
	case IntentTypeSearch:
		// 执行搜索
		searchResult, err := mc.executeSearch(ctx, content.SearchQuery)
		if err != nil {
			return nil, err
		}
		result.SearchResult = searchResult

	case IntentTypeRecommend:
		// 执行推荐
		recommendationResult, err := mc.executeRecommendation(ctx, content.Preferences)
		if err != nil {
			return nil, err
		}
		result.RecommendationResult = recommendationResult

	case IntentTypeDownload:
		// 执行下载
		downloadResult, err := mc.executeDownload(ctx, content.DownloadInfo)
		if err != nil {
			return nil, err
		}
		result.DownloadResult = downloadResult

	case IntentTypeSubscribe:
		// 执行订阅
		subscribeResult, err := mc.executeSubscription(ctx, content.SubscribeInfo)
		if err != nil {
			return nil, err
		}
		result.SubscribeResult = subscribeResult

	case IntentTypeQuery:
		// 执行查询
		queryResult, err := mc.executeQuery(ctx, content.QueryInfo)
		if err != nil {
			return nil, err
		}
		result.QueryResult = queryResult

	default:
		// 通用业务处理
		generalResult, err := mc.executeGeneralBusiness(ctx, content, intent)
		if err != nil {
			return nil, err
		}
		result.GeneralResult = generalResult
	}

	return &result, nil
}

// generateResponse 响应生成
func (mc *MessageChain) generateResponse(ctx context.Context, result *BusinessResult, intent *MessageIntent) (*MessageResponse, error) {
	var response MessageResponse

	switch intent.Type {
	case IntentTypeSearch:
		response = mc.generateSearchResponse(result.SearchResult)

	case IntentTypeRecommend:
		response = mc.generateRecommendationResponse(result.RecommendationResult)

	case IntentTypeDownload:
		response = mc.generateDownloadResponse(result.DownloadResult)

	case IntentTypeSubscribe:
		response = mc.generateSubscriptionResponse(result.SubscribeResult)

	case IntentTypeQuery:
		response = mc.generateQueryResponse(result.QueryResult)

	default:
		response = mc.generateGeneralResponse(result.GeneralResult)
	}

	// 个性化响应优化
	response = mc.personalizeResponse(ctx, response, intent)

	// 格式化响应内容
	response = mc.formatResponseContent(response)

	return &response, nil
}

// sendResponse 消息发送
func (mc *MessageChain) sendResponse(ctx context.Context, response *MessageResponse) error {
	// 根据用户偏好选择发送渠道
	channel, err := mc.selectDeliveryChannel(ctx, response.UserID)
	if err != nil {
		return err
	}

	// 发送消息
	if err := mc.deliverMessage(ctx, response, channel); err != nil {
		return err
	}

	// 记录发送状态
	if err := mc.recordDeliveryStatus(ctx, response, channel); err != nil {
		mc.logger.Warn("记录发送状态失败", zap.Error(err))
	}

	return nil
}

// recordProcessingResult 记录处理结果
func (mc *MessageChain) recordProcessingResult(ctx context.Context, req *MessageRequest, response *MessageResponse) error {
	processingRecord := &model.MessageProcessingRecord{
		MessageID:    req.MessageID,
		UserID:       req.UserID,
		Content:      req.Content,
		IntentType:   response.IntentType,
		Success:      response.Success,
		ResponseTime: response.ResponseTime,
		ErrorMsg:     response.ErrorMessage,
		CreatedAt:    time.Now(),
	}

	return mc.messageRepo.SaveProcessingRecord(ctx, processingRecord)
}

// validateRequest 验证消息请求
func (mc *MessageChain) validateRequest(req *MessageRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	if req.Content == "" {
		return fmt.Errorf("消息内容不能为空")
	}

	if len(req.Content) > 1000 {
		return fmt.Errorf("消息内容过长")
	}

	return nil
}

// 辅助方法实现...

// MessageRequest 消息请求
type MessageRequest struct {
	MessageID   string                 `json:"message_id"`
	UserID      string                 `json:"user_id" validate:"required"`
	Content     string                 `json:"content" validate:"required"`
	MessageType string                 `json:"message_type"` // text, voice, image, etc.
	Platform    string                 `json:"platform"`     // web, mobile, api, etc.
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	MessageID    string                 `json:"message_id"`
	UserID       string                 `json:"user_id"`
	Content      string                 `json:"content"`
	IntentType   string                 `json:"intent_type"`
	Success      bool                   `json:"success"`
	ResponseTime time.Duration          `json:"response_time"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Suggestions  []string               `json:"suggestions,omitempty"`
	NextActions  []string               `json:"next_actions,omitempty"`
	GeneratedAt  time.Time              `json:"generated_at"`
}

// ProcessedMessage 处理后的消息
type ProcessedMessage struct {
	OriginalContent string                 `json:"original_content"`
	CleanedContent  string                 `json:"cleaned_content"`
	Language        string                 `json:"language"`
	UserContext     *UserContext           `json:"user_context"`
	MessageType     string                 `json:"message_type"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// MessageIntent 消息意图
type MessageIntent struct {
	Type       string                 `json:"type"`       // search, recommend, download, etc.
	Confidence float64                `json:"confidence"` // 0-1的置信度
	Entities   []Entity               `json:"entities"`
	Parameters map[string]interface{} `json:"parameters"`
	Requires   []string               `json:"requires"`    // 需要的额外信息
	FallbackTo string                 `json:"fallback_to"` // 备用意图类型
}

// ParsedContent 解析后的内容
type ParsedContent struct {
	SearchQuery   *SearchQuery               `json:"search_query,omitempty"`
	Preferences   *RecommendationPreferences `json:"preferences,omitempty"`
	DownloadInfo  *DownloadInfo              `json:"download_info,omitempty"`
	SubscribeInfo *SubscriptionInfo          `json:"subscribe_info,omitempty"`
	QueryInfo     *QueryInfo                 `json:"query_info,omitempty"`
	Entities      []Entity                   `json:"entities"`
	Sentiment     string                     `json:"sentiment"` // positive, negative, neutral
	Urgency       string                     `json:"urgency"`   // high, medium, low
}

// BusinessResult 业务逻辑处理结果
type BusinessResult struct {
	SearchResult         *SearchResult         `json:"search_result,omitempty"`
	RecommendationResult *RecommendationResult `json:"recommendation_result,omitempty"`
	DownloadResult       *DownloadResult       `json:"download_result,omitempty"`
	SubscribeResult      *SubscriptionResult   `json:"subscribe_result,omitempty"`
	QueryResult          *QueryResult          `json:"query_result,omitempty"`
	GeneralResult        *GeneralResult        `json:"general_result,omitempty"`
}

// 意图类型常量
const (
	IntentTypeSearch    = "search"
	IntentTypeRecommend = "recommend"
	IntentTypeDownload  = "download"
	IntentTypeSubscribe = "subscribe"
	IntentTypeQuery     = "query"
	IntentTypeHelp      = "help"
	IntentTypeFeedback  = "feedback"
	IntentTypeUnknown   = "unknown"
)

// Entity 实体
type Entity struct {
	Type  string `json:"type"` // person, location, organization, etc.
	Value string `json:"value"`
	Start int    `json:"start"` // 在文本中的起始位置
	End   int    `json:"end"`   // 在文本中的结束位置
}

// UserContext 用户上下文
type UserContext struct {
	UserID      string           `json:"user_id"`
	Preferences *UserPreferences `json:"preferences"`
	History     *UserHistory     `json:"history"`
	SessionInfo *SessionInfo     `json:"session_info"`
	DeviceInfo  *DeviceInfo      `json:"device_info"`
}

// SearchQuery 搜索查询
type SearchQuery struct {
	Keywords   []string               `json:"keywords"`
	Filters    map[string]interface{} `json:"filters"`
	SortBy     string                 `json:"sort_by"`
	PageSize   int                    `json:"page_size"`
	PageNumber int                    `json:"page_number"`
}

// 其他类型定义...

type PluginManager interface {
	CallIntentRecognition(ctx context.Context, message *ProcessedMessage) (*MessageIntent, error)
}
