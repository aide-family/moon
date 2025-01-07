package datasource

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-kratos/kratos/v2/log"
)

// EndpointDuration 函数用于获取指定endpoint的请求相应时间和状态码
func EndpointDuration(ctx context.Context, url string, method vobj.HTTPMethod, headers map[string]string, body string, timeout time.Duration) map[watch.Indexer]*Point {
	now := time.Now()
	var elapsed time.Duration
	var statusCode int
	var err error
	switch method {
	case vobj.HTTPMethodPost:
		elapsed, statusCode, _ = postResponseTimeAndStatusCode(ctx, url, headers, body, timeout)
	default:
		elapsed, statusCode, err = getResponseTimeAndStatusCode(ctx, url, headers, timeout)
		if err != nil {
			log.Error(err)
		}
	}
	points := make(map[watch.Indexer]*Point)
	labels := label.NewLabels(map[string]string{
		label.StrategyHTTPPath:   url,
		label.StrategyHTTPMethod: method.String(),
	})
	points[labels] = &Point{
		Labels: labels.Map(),
		Values: []*Value{
			// 持续时间为0，表示请求失败
			{
				Value:     float64(elapsed.Milliseconds()),
				Timestamp: now.Unix(),
			},
			// 状态码为0，表示请求失败
			{
				Value:     float64(statusCode),
				Timestamp: now.Unix(),
			},
		},
	}
	return points
}

// getResponseTimeAndStatusCode 函数用于获取指定url的响应时间和状态码
func getResponseTimeAndStatusCode(ctx context.Context, url string, headers map[string]string, timeout time.Duration) (time.Duration, int, error) {
	start := time.Now() // 记录开始时间

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) // 构造请求
	if err != nil {
		return 0, 0, err // 返回错误
	}
	for k, v := range headers {
		req.Header.Set(k, v) // 设置请求头
	}
	client := http.Client{Timeout: timeout} // 构造客户端
	resp, err := client.Do(req)             // 发送请求
	if err != nil {
		return 0, 0, err // 返回错误
	}
	defer resp.Body.Close() // 确保关闭响应体

	elapsed := time.Since(start) // 计算响应时间
	return elapsed, resp.StatusCode, nil
}

// postResponseTimeAndStatusCode 函数用于获取指定url的响应时间和状态码
func postResponseTimeAndStatusCode(ctx context.Context, url string, headers map[string]string, body string, timeout time.Duration) (time.Duration, int, error) {
	start := time.Now() // 记录开始时间

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body)) // 构造请求
	if err != nil {
		return 0, 0, err // 返回错误
	}
	for k, v := range headers {
		req.Header.Set(k, v) // 设置请求头
	}
	client := http.Client{Timeout: timeout} // 构造客户端
	resp, err := client.Do(req)             // 发送请求
	if err != nil {
		return 0, 0, err // 返回错误
	}
	defer resp.Body.Close() // 确保关闭响应体

	elapsed := time.Since(start) // 计算响应时间
	return elapsed, resp.StatusCode, nil
}
