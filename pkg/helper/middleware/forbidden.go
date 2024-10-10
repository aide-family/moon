package middleware

import (
	"context"

	"github.com/aide-family/moon/pkg/merr"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Forbidden 禁用列表API请求， 一般用于临时限制某些API访问， 降低因为数据异常造成的影响
func Forbidden(operations ...string) middleware.Middleware {
	blackMap := make(map[string]struct{})
	for _, op := range operations {
		blackMap[op] = struct{}{}
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			operation, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, merr.ErrorNotification("get operation failed")
			}
			operator := operation.Operation()
			if _, ok := blackMap[operator]; ok {
				return nil, merr.ErrorI18nForbidden(ctx)
			}

			return handler(ctx, req)
		}
	}
}
