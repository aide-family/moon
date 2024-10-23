package vobj

import (
	"strings"
)

// HttpMethod http 请求方法枚举
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=HttpMethod -linecomment
type HttpMethod int

const (
	// HttpMethodUnknown 未知
	HttpMethodUnknown HttpMethod = iota // unknown

	// HttpMethodGet GET 请求
	HttpMethodGet // GET

	// HttpMethodPost POST 请求
	HttpMethodPost // POST

	// HttpMethodPut PUT 请求
	HttpMethodPut // PUT

	// HttpMethodDelete DELETE 请求
	HttpMethodDelete // DELETE

	// HttpMethodHead HEAD 请求
	HttpMethodHead // HEAD

	// HttpMethodOptions OPTIONS 请求
	HttpMethodOptions // OPTIONS

	// HttpMethodTrace TRACE 请求
	HttpMethodTrace // TRACE

	// HttpMethodConnect CONNECT 请求
	HttpMethodConnect // CONNECT

	// HttpMethodPatch PATCH 请求
	HttpMethodPatch // PATCH
)

// ToHTTPMethod 将字符串转换为 HttpMethod 枚举
func ToHTTPMethod(method string) HttpMethod {
	switch strings.ToUpper(method) {
	case "GET":
		return HttpMethodGet
	case "POST":
		return HttpMethodPost
	case "PUT":
		return HttpMethodPut
	case "DELETE":
		return HttpMethodDelete
	case "HEAD":
		return HttpMethodHead
	case "OPTIONS":
		return HttpMethodOptions
	case "TRACE":
		return HttpMethodTrace
	case "CONNECT":
		return HttpMethodConnect
	case "PATCH":
		return HttpMethodPatch
	default:
		return HttpMethodUnknown
	}
}
