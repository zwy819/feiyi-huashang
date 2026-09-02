package model

// API 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message *string     `json:"message"`
	Data    interface{} `json:"data"`
}

type PaginatedData struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

func NewResponse(code int, message string, data interface{}) Response {
	var msg *string
	if message != "" {
		msg = &message
	}
	return Response{Code: code, Message: msg, Data: data}
}

func SuccessResponse(data interface{}) Response {
	return Response{Code: 0, Message: nil, Data: data}
}

func SuccessMessage(message string, data interface{}) Response {
	return Response{Code: 0, Message: &message, Data: data}
}

func ErrorResponse(code int, message string) Response {
	return Response{Code: code, Message: &message, Data: nil}
}
