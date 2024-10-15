package middleware

import (
	"context"

	"github.com/aide-family/moon/pkg/merr"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// CheckRbacFun 权限校验函数
type CheckRbacFun func(ctx context.Context, operation string) error

// Rbac 权限校验中间件
func Rbac(check CheckRbacFun) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			operation, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, merr.ErrorNotification("get operation failed")
			}
			// 判断该用户在该资源是否有权限
			if err = check(ctx, operation.Operation()); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}
	}
}
