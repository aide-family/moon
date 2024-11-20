package vobj

import (
	"strings"
)

// HTTPMethod http 请求方法枚举
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=HTTPMethod -linecomment
type HTTPMethod int

const (
	// HTTPMethodUnknown 未知
	HTTPMethodUnknown HTTPMethod = iota // unknown

	// HTTPMethodGet GET 请求
	HTTPMethodGet // GET

	// HTTPMethodPost POST 请求
	HTTPMethodPost // POST

	// HTTPMethodPut PUT 请求
	HTTPMethodPut // PUT

	// HTTPMethodDelete DELETE 请求
	HTTPMethodDelete // DELETE

	// HTTPMethodHead HEAD 请求
	HTTPMethodHead // HEAD

	// HTTPMethodOptions OPTIONS 请求
	HTTPMethodOptions // OPTIONS

	// HTTPMethodTrace TRACE 请求
	HTTPMethodTrace // TRACE

	// HTTPMethodConnect CONNECT 请求
	HTTPMethodConnect // CONNECT

	// HTTPMethodPatch PATCH 请求
	HTTPMethodPatch // PATCH
)

// ToHTTPMethod 将字符串转换为 HttpMethod 枚举
func ToHTTPMethod(method string) HTTPMethod {
	switch strings.ToUpper(method) {
	case "GET":
		return HTTPMethodGet
	case "POST":
		return HTTPMethodPost
	case "PUT":
		return HTTPMethodPut
	case "DELETE":
		return HTTPMethodDelete
	case "HEAD":
		return HTTPMethodHead
	case "OPTIONS":
		return HTTPMethodOptions
	case "TRACE":
		return HTTPMethodTrace
	case "CONNECT":
		return HTTPMethodConnect
	case "PATCH":
		return HTTPMethodPatch
	default:
		return HTTPMethodUnknown
	}
}
