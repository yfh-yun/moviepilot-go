package models

// Response 响应模型
type Response struct {
	// 状�?	Success bool `json:"success"`
	// 消息文本
	Message string `json:"message,omitempty"`
	// 数据
	Data interface{} `json:"data,omitempty"`
}

// NewResponse 创建一个新�?Response 实例
func NewResponse() *Response {
	return &Response{
		Data: make(map[string]interface{}),
	}
}

// NewResponseWithParams 创建一个带有参数的 Response 实例
func NewResponseWithParams(success bool, message string, data interface{}) *Response {
	return &Response{
		Success: success,
		Message: message,
		Data:    data,
	}
}

// NewSuccessResponse 创建一个成功的 Response 实例
func NewSuccessResponse(message string, data interface{}) *Response {
	return &Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建一个错误的 Response 实例
func NewErrorResponse(message string) *Response {
	return &Response{
		Success: false,
		Message: message,
		Data:    make(map[string]interface{}),
	}
}
