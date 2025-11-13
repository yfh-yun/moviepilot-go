package tmdbv3api

// TMDbException TMDb异常结构�?type TMDbException struct {
	Message string
}

// Error 实现error接口
func (e *TMDbException) Error() string {
	return e.Message
}

// NewTMDbException 创建TMDbException实例
func NewTMDbException(message string) *TMDbException {
	return &TMDbException{
		Message: message,
	}
}
