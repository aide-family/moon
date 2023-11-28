package middler

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/pkg/conn"
)

func RbacServer() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 1. 解析jwt
			authClaims, ok := GetAuthClaims(ctx)
			if !ok {
				return nil, ErrTokenInvalid
			}
			path := GetPath(ctx)
			method := GetMethod(ctx)
			// 2. 校验权限
			enforcer := conn.Enforcer()
			has, err := enforcer.Enforce(authClaims.Role, path, method)
			if err != nil {
				return nil, err
			}
			if !has {
				return nil, status.Error(codes.PermissionDenied, "permission denied")
			}

			return handler(ctx, req)
		}
	}
}
