package server

import (
	"context"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	"github.com/aide-family/Alioth/app/alioth/internal/biz/auth"
)

func authenticate(uc *auth.Usecase) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			header := tr.RequestHeader()
			authorization := header.Get("Authorization")
			if authorization == "" {
				authorization = header.Get("authorization")
			}
			namespace := header.Get("X-Namespace")
			if namespace == "" {
				namespace = header.Get("x-namespace")
			}
			next, err := uc.Authenticate(ctx, tr.Operation(), authorization, namespace)
			if err != nil {
				return nil, err
			}
			return handler(next, req)
		}
	}
}
