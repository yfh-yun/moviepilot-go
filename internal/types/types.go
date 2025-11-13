package types

// Cookie Cookie结构�?type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HTTPResponse HTTP响应结构�?type HTTPResponse struct {
	Body       []byte            `json:"body"`
	Content    string            `json:"content"`
	StatusCode int               `json:"status_code"`
	Status     int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Cookies    []*Cookie         `json:"cookies"`
}
