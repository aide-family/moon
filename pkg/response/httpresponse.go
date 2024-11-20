package httpres

import (
	"encoding/json"
	"net/http"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Response 是一个通用的HTTP响应结构体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// writeJSON 将响应写入到HTTP响应中
func writeJSON(w kratoshttp.Context, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Response().WriteHeader(status)
	if err := json.NewEncoder(w.Response()).Encode(response); err != nil {
		// 处理编码错误，可以记录日志或返回一个通用错误响应
		w.Response().WriteHeader(http.StatusInternalServerError)
		// 注意：这里我们不能再次使用 Encode，因为可能会再次失败
		_, _ = w.Response().Write([]byte(`{"code":500,"message":"Internal Server Error"}`))
	}
}

// Success 返回一个成功的响应
func Success(w kratoshttp.Context, message string, data interface{}) {
	if message == "" {
		message = "success"
	}
	response := Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	}
	writeJSON(w, http.StatusOK, response)
}

// Error 返回一个错误的响应
func Error(w kratoshttp.Context, message string, status int) {
	if message == "" {
		message = "error"
	}
	response := Response{
		Code:    status,
		Message: message,
	}
	writeJSON(w, status, response)
}
