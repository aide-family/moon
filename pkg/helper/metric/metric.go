package metric

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ http.Handler = &metric{}

type metric struct {
	token string
}

// NewMetricHandler 创建指标处理对象
func NewMetricHandler(token string) http.Handler {
	return &metric{token: strings.TrimSpace(token)}
}

// ServeHTTP 处理HTTP请求
func (m *metric) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 从URL中获取token
	if m.validateToken(request) {
		writer.WriteHeader(http.StatusForbidden)
		return
	}
	promhttp.Handler().ServeHTTP(writer, request)
}

func (m *metric) validateToken(request *http.Request) bool {
	if m.token == "" {
		return false
	}
	token := request.URL.Query().Get("token")
	return token == m.token
}
