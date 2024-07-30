package httpx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// HTTPX http请求封装
type HTTPX struct {
	http.Client

	headers map[string]string
	user    string
	pass    string
}

// NewHTTPX 创建一个http请求
func NewHTTPX() *HTTPX {
	return &HTTPX{
		Client: http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     50 * time.Second,
			},
		},
	}
}

// SetHeader 设置请求头
func (h *HTTPX) SetHeader(headers map[string]string) *HTTPX {
	h.headers = headers
	return h
}

// SetBasicAuth 将用户名和密码添加到请求头中
//
//	req.SetBasicAuth(username, password)
func (h *HTTPX) SetBasicAuth(username, password string) *HTTPX {
	h.user = username
	h.pass = password
	return h
}

// POST 发起post请求
func (h *HTTPX) POST(url string, data []byte) (*http.Response, error) {
	reader := bytes.NewReader(data)
	// 设置请求头
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return nil, err
	}
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Set(k, v)
		}
	} else {
		// 没有请求头时候, 默认设置请求头json, utf-8
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	if h.user != "" && h.pass != "" {
		req.SetBasicAuth(h.user, h.pass)
	}
	return h.Do(req)
}

// POSTWithContext 发起post请求
func (h *HTTPX) POSTWithContext(ctx context.Context, url string, data []byte) (*http.Response, error) {
	reader := bytes.NewReader(data)
	// 设置请求头
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Set(k, v)
		}
	} else {
		// 没有请求头时候, 默认设置请求头json, utf-8
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	if h.user != "" && h.pass != "" {
		req.SetBasicAuth(h.user, h.pass)
	}
	return h.Do(req)
}

// GET 发起get请求
func (h *HTTPX) GET(u string) (*http.Response, error) {
	// 验证URL是否有效
	if u == "" || !isValidURL(u) {
		return nil, errors.New("invalid URL")
	}
	// 设置请求头
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Set(k, v)
		}
	}
	if h.user != "" && h.pass != "" {
		req.SetBasicAuth(h.user, h.pass)
	}
	return h.Do(req)
}

// GETWithContext 发起get请求
func (h *HTTPX) GETWithContext(ctx context.Context, u string) (*http.Response, error) {
	// 验证URL是否有效
	if u == "" || !isValidURL(u) {
		return nil, errors.New("invalid URL")
	}
	// 设置请求头
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Set(k, v)
		}
	}
	if h.user != "" && h.pass != "" {
		req.SetBasicAuth(h.user, h.pass)
	}
	return h.Do(req)
}

// isValidURL 检查URL是否有效
func isValidURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	return err == nil
}

// ParseQuery 处理结构体转为query参数
func ParseQuery(qr map[string]any) string {
	if qr == nil || len(qr) == 0 {
		return ""
	}
	query := url.Values{}
	for k, v := range qr {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	return query.Encode()
}
