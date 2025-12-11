package dto

import ()

// Response API响应
type Response struct {
	// 状态
	Success bool `json:"success"`
	// 消息文本
	Message string `json:"message,omitempty"`
	// 数据
	Data any `json:"data,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data any) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// NewSuccessResponseWithMessage 创建带消息的成功响应
func NewSuccessResponseWithMessage(message string, data any) *Response {
	return &Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string) *Response {
	return &Response{
		Success: false,
		Message: message,
	}
}

// NewErrorResponseWithData 创建带数据的错误响应
func NewErrorResponseWithData(message string, data any) *Response {
	return &Response{
		Success: false,
		Message: message,
		Data:    data,
	}
}
