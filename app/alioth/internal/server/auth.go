package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

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
			next, err := uc.Authenticate(ctx, tr.Operation(), header.Get("Authorization"), header.Get("X-Namespace"))
			if err != nil {
				return nil, err
			}
			return handler(next, req)
		}
	}
}
